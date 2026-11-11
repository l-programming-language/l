import assert from 'node:assert/strict';
import fs from 'node:fs/promises';
import os from 'node:os';
import path from 'node:path';
import { spawnSync } from 'node:child_process';
import test from 'node:test';
import { parseFragment } from 'parse5';
import { parse as parseYAML } from 'yaml';
import { anchors, parseDocument, parseMarkdown, preparePage, validateMap, validateStdlib, stdlibCatalog } from '../scripts/content.mjs';
import { deriveWords } from '../scripts/words.mjs';
import { prepare, publishOwned } from '../scripts/prepare.mjs';
import { atomicWrite, contained, safeFile, sha256 } from '../scripts/lib.mjs';
import { checkDist } from '../scripts/check-dist.mjs';

const page = (source = 'docs/start.md', route = '/docs/guide/') => ({
  source, route, title: 'Start here', description: 'A useful page.',
  section: 'Guide', order: 1, role: 'explanatory', generated: false,
});

async function fixture(t, files = {}) {
  const temporary = await fs.mkdtemp(path.join(os.tmpdir(), 'l docs content '));
  const root = await fs.realpath(temporary);
  t.after(() => fs.rm(root, { recursive: true, force: true }));
  for (const [relative, content] of Object.entries(files)) {
    const file = path.join(root, relative);
    await fs.mkdir(path.dirname(file), { recursive: true });
    await fs.writeFile(file, content);
  }
  return root;
}

async function project(root, text, options = {}) {
  const current = options.page ?? page();
  const pages = options.pages ?? [current];
  const assets = new Map();
  const output = await preparePage({
    root, page: current, pages,
    identity: { repository: null, docsCommit: null, dirty: null },
    examples: [], sourceTexts: new Map([[current.source, text]]), assets,
    ...options,
  });
  return { output, assets };
}

function definitionWords(input, definitions) {
  return deriveWords({
    entries: [{ id: 'meaning', name: 'Meaning', kind: 'builtin', support: 'supported', doc: 'docs/start.md#meaning', sources: definitions.map((definition) => definition.source), evidence: 'static-inspection', word: { partOfSpeech: 'verb', definitions } }],
    pages: [page()], documents: new Map([['docs/start.md', parseDocument(input)]]),
  });
}

test('map validation enforces route, ownership and navigation metadata', () => {
  assert.deepEqual(validateMap({ schemaVersion: 1, pages: [page()] }).pages, [page()]);
  assert.throws(() => validateMap({ pages: [] }), /schemaVersion/);
  assert.throws(() => validateMap({ schemaVersion: 1, pages: [page(), page('docs/other.md')] }), /Duplicate route/);
  assert.throws(() => validateMap({ schemaVersion: 1, pages: [page(), page('docs/start.md', '/docs/other/')] }), /Duplicate canonical/);
  for (const route of ['/docs/guide', '/l/docs/guide/', '/docs/_meta/', '/docs/../guide/', '/']) {
    assert.throws(() => validateMap({ schemaVersion: 1, pages: [page('docs/start.md', route)] }), /route/i);
  }
  for (const override of [{ title: '' }, { order: NaN }, { generated: 'false' }, { sidebar: 'false' }, { role: 'probably-supported' }, { source: 'docs/_meta/private.md' }]) {
    assert.throws(() => validateMap({ schemaVersion: 1, pages: [{ ...page(), ...override }] }));
  }
});

test('static output rejects incomplete Pagefind assets before claiming a usable search build', async (t) => {
  const identity = JSON.stringify({ channel: 'local', sourceFingerprint: `sha256:${'a'.repeat(64)}` });
  const html = '<!doctype html><html><head><title>Docs</title><meta name="description" content="Reference"><meta name="robots" content="noindex"></head><body><h1>Docs</h1></body></html>';
  const entry = { version: '1.5.2', languages: { en: { hash: 'en_reviewed', wasm: 'en', page_count: 1 } } };
  const files = {
    'content-map.json': JSON.stringify({ pages: [] }), '.generated/version.json': identity,
    'public/install.sh': '#!/bin/sh\n', 'dist/install.sh': '#!/bin/sh\n',
    'dist/version.json': identity, 'dist/index.html': html, 'dist/404.html': html,
    'dist/pagefind/pagefind.js': 'export const search = () => {};',
    'dist/pagefind/pagefind-worker.js': 'self.onmessage = () => {};',
    'dist/pagefind/pagefind-entry.json': JSON.stringify(entry),
    'dist/pagefind/pagefind.en_reviewed.pf_meta': 'metadata',
    'dist/pagefind/wasm.en.pagefind': 'wasm',
  };
  const valid = await fixture(t, files);
  assert.equal((await checkDist(valid)).routes, 1);
  for (const [name, file, bytes] of [
    ['empty entry', 'pagefind-entry.json', ''],
    ['malformed entry', 'pagefind-entry.json', '{'],
    ['missing language inventory', 'pagefind-entry.json', JSON.stringify({ version: '1.5.2', languages: {} })],
    ['missing selected metadata', 'pagefind-entry.json', JSON.stringify({ ...entry, languages: { en: { ...entry.languages.en, hash: 'en_missing' } } })],
    ['invalid selected wasm', 'pagefind-entry.json', JSON.stringify({ ...entry, languages: { en: { ...entry.languages.en, wasm: '../elsewhere' } } })],
    ['empty worker', 'pagefind-worker.js', ''],
    ['empty selected metadata', 'pagefind.en_reviewed.pf_meta', ''],
    ['empty selected wasm', 'wasm.en.pagefind', ''],
  ]) {
    await t.test(name, async (t) => {
      const root = await fixture(t, { ...files, [`dist/pagefind/${file}`]: bytes });
      await assert.rejects(checkDist(root), /Pagefind/i);
    });
  }
});

test('anchors account for repeated headings and explicit anchors without parsing fenced examples', () => {
  const input = '# Example\n\n## Number\n\n## Number\n\n<a id="literal-anchor"></a>\n\n```md\n## Not an anchor\n<a id="not-real"></a>\n```\n';
  assert.deepEqual([...anchors(input)], ['example', 'number', 'number-1', 'literal-anchor']);
});

test('AST projection rewrites inline/reference links and images from the canonical source directory', async (t) => {
  const start = '# Start\n\n[Next](next.md?mode=read#part-two) and [reference][chapter].\n\n[chapter]: next.md#part-two "Next chapter"\n\n![A diagram](assets/image%20with%20space.svg)\n\n![Reference diagram][art]\n\n[art]: assets/image%20with%20space.svg\n\n[Source](../../source/native.rs)\n';
  const next = '# Next\n\n## Part two\n\nDescription.\n';
  const illustration = '<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 1 1"><path d="M0 0h1v1z"/></svg>';
  const root = await fixture(t, {
    'docs/imported/start.md': start, 'docs/imported/next.md': next,
    'docs/imported/assets/image with space.svg': illustration, 'source/native.rs': 'fn main() {}\n',
  });
  const current = page('docs/imported/start.md', '/docs/imported/');
  const nextPage = page('docs/imported/next.md', '/docs/imported/next/');
  const { output, assets } = await project(root, start, {
    page: current, pages: [current, nextPage],
    sourceTexts: new Map([[current.source, start], [nextPage.source, next]]),
  });
  assert.match(output, /\[Next\]\(<\/docs\/imported\/next\/\?mode=read#part-two>\)/);
  assert.match(output, /\[chapter\]: <\/docs\/imported\/next\/#part-two> "Next chapter"/);
  assert.match(output, /Source \(source: `source\/native\.rs`\)/);
  const assetName = `${sha256(illustration).slice(0, 24)}.svg`;
  assert.deepEqual([...assets.keys()], [assetName]);
  assert.equal(assets.get(assetName).toString(), illustration);
  assert.match(output, new RegExp(`!\\[A diagram\\]\\(<\\/content-assets\\/${assetName}>\\)`));
  assert.match(output, new RegExp(`!\\[Reference diagram\\]\\(<\\/content-assets\\/${assetName}>\\)`));
  assert.doesNotMatch(output, /https:\/\/github\.com|\/Users\//);
});

test('literal Markdown instructions, AML and inline code are not rewritten', async (t) => {
  const literal = '```markdown\n[keep](missing.md#not-real)\n<!-- example:not-an-inclusion -->\n```\n\n```text\n<PROGRAM>{body} `[literal](also-missing.md)`</PROGRAM>\n```\n\n`[inline](missing.md)`\n';
  const input = `# Literal source\n\n${literal}`;
  const root = await fixture(t, { 'docs/start.md': input });
  const { output } = await project(root, input);
  assert.ok(output.includes(literal.trim()));
  assert.ok(output.includes('<PROGRAM>{body} `[literal](also-missing.md)`</PROGRAM>'));
});

test('real word markers reject unknown identities while literal word instructions remain untouched', async (t) => {
  const input = '# Words\n\n## Missing\n\n<!-- word:missing:meaning -->\n\nAn unreviewed meaning.\n';
  const root = await fixture(t, { 'docs/start.md': input });
  await assert.rejects(project(root, input), /Unknown word.*missing:meaning/);
  const literal = '# Words\n\n```markdown\n<!-- word:missing:meaning -->\n<!-- word-index -->\n```\n\n`<!-- word:missing:meaning -->`\n';
  assert.ok((await project(root, literal)).output.includes(literal.slice(literal.indexOf('```'))));
});

test('a page cannot emit conflicting word indices', async (t) => {
  const input = '# Words\n\n<!-- word-index -->\n\n<!-- word-index:builtins -->\n';
  const root = await fixture(t, { 'docs/start.md': input });
  await assert.rejects(project(root, input), /Multiple word indices/);
});

test('malformed or inline word indices do not silently disappear', async (t) => {
  const root = await fixture(t, { 'docs/start.md': '# Words\n' });
  for (const directive of ['<!-- word-index:eng:Fixture -->', '<!-- word-index:eng/fixture -->', 'Text <!-- word-index --> after.']) {
    await assert.rejects(project(root, '# Words\n\n' + directive + '\n'), /Invalid word-index directive/);
  }
});

test('word definitions display exact L, configured JSON and compiler source lines in collapsed escaped blocks', async (t) => {
  const input = '# Words\n\n## Meaning\n\n<!-- word:meaning -->\n\nA visible definition.\n\n**Usage:** A visible usage.\n\n<!-- word-definition:meaning -->\n';
  const selected = '"literal" is `<BOOK>{value}</BOOK>`.\r\n*** ```` and </code></pre></details> ***\r\n';
  const json = '"it": {"value": "<IT>&value", "pos": "noun"}';
  const go = 'case "bytes":\n\treturn Bytes';
  const root = await fixture(t, { 'docs/start.md': input, 'source.l': 'before\r\n' + selected + 'after', 'vocabulary.json': json, 'types.go': go });
  const words = definitionWords(input, [{ source: 'source.l', lines: [2, 3] }, { source: 'vocabulary.json', lines: [1, 1] }, { source: 'types.go', lines: [1, 2] }]);
  const { output } = await project(root, input, { words });
  assert.match(output, /<details class="l-word-definition" data-pagefind-ignore><summary>Source definition<\/summary>/);
  assert.doesNotMatch(output, /<details[^>]*\bopen\b/);
  assert.match(output, /L source.*source\.l.*lines 2–3/);
  assert.match(output, /Configured JSON.*vocabulary\.json.*lines 1–1/);
  assert.match(output, /Compiler definition.*types\.go.*lines 1–2/);
  assert.equal([...output.matchAll(/<pre tabindex="0" role="region" aria-label="Meaning source definition">/g)].length, 3);
  assert.ok(output.includes('<code class="language-l">&quot;literal&quot; is `&lt;BOOK&gt;{value}&lt;/BOOK&gt;`.\r\n*** ```` and &lt;/code&gt;&lt;/pre&gt;&lt;/details&gt; ***\r\n</code></pre>'));
  assert.ok(output.includes('<code class="language-json">&quot;it&quot;: {&quot;value&quot;: &quot;&lt;IT&gt;&amp;value&quot;, &quot;pos&quot;: &quot;noun&quot;}</code></pre>'));
  assert.ok(output.includes('<code class="language-go">case &quot;bytes&quot;:\n\treturn Bytes</code></pre>'));
  assert.doesNotMatch(output, /before|after|word-definition:meaning/);
  assert.ok(output.indexOf('A visible definition.') < output.indexOf('class="l-word-definition"'));
  assert.ok(output.indexOf('A visible usage.') < output.indexOf('class="l-word-definition"'));
  assert.equal(await fs.readFile(path.join(root, 'source.l'), 'utf8'), 'before\r\n' + selected + 'after');
});

test('word definitions read actual source files and reject missing, out-of-range and invalid UTF-8 input', async (t) => {
  const input = '# Words\n\n## Meaning\n\n<!-- word:meaning -->\n\nA meaning.\n\n<!-- word-definition:meaning -->\n';
  const root = await fixture(t, { 'docs/start.md': input, 'source.l': 'first\nlast' });
  const words = definitionWords(input, [{ source: 'source.l', lines: [2, 2] }]);
  const { output } = await project(root, input, { words, sourceTexts: new Map([['docs/start.md', input], ['source.l', 'fabricated\ncopy']]) });
  assert.ok(output.includes('<code class="language-l">last</code>'));
  assert.doesNotMatch(output, /fabricated|copy/);
  await assert.rejects(project(root, input, { words: definitionWords(input, [{ source: 'source.l', lines: [2, 3] }]) }), /out of bounds/);
  await assert.rejects(project(root, input, { words: definitionWords(input, [{ source: 'missing.l', lines: [1, 1] }]) }), /ENOENT/);
  await fs.writeFile(path.join(root, 'source.l'), Buffer.from([0x61, 0x0a, 0xff]));
  await assert.rejects(project(root, input, { words }), /Invalid UTF-8.*word definition/);
  const outside = await fixture(t, { 'private.l': 'private' });
  await fs.symlink(path.join(outside, 'private.l'), path.join(root, 'escape.l'));
  await assert.rejects(project(root, input, { words: definitionWords(input, [{ source: 'escape.l', lines: [1, 1] }]) }), /Symlink escapes/);
});

test('word-definition projection rejects unowned directives and preserves literal instructions', async (t) => {
  const literal = '# Words\n\n```markdown\n<!-- word-definition:missing -->\n```\n\n`<!-- word-definition:missing -->`\n';
  const root = await fixture(t, { 'docs/start.md': literal });
  assert.ok((await project(root, literal)).output.includes(literal.slice(literal.indexOf('```'))));
  for (const directive of ['<!-- word-definition:missing -->', 'Text <!-- word-definition:missing --> after.', '<!-- word-definition -->']) {
    await assert.rejects(project(root, '# Words\n\n' + directive + '\n'), /word-definition/);
  }
  const original = '# Words\n\n## Meaning\n\n<!-- word:meaning -->\n\nA meaning.\n\n<!-- word-definition:meaning -->\n';
  const words = definitionWords(original, [{ source: 'source.l', lines: [1, 1] }]);
  await assert.rejects(project(root, original.replace('<!-- word-definition:meaning -->', '## Other\n\n<!-- word-definition:meaning -->'), { words }), /canonical word section/);
});

test('authoring dictionaries require the same real owner and source as public library records', () => {
  const index = { format: 1, languages: { eng: { workshop: { entry: 'fixtures/workshop/dictionary.l', dictionary: 'Workshop' } } } };
  const namespace = { id: 'workshop', kind: 'stdlib-export', dictionary: { lexicon: 'eng', key: 'workshop' }, visibility: 'namespace', support: 'internal', doc: 'docs/workshop.md#details', sources: ['fixtures/workshop/dictionary.l'] };
  const author = { id: 'author', kind: 'authoring-word', dictionary: namespace.dictionary, sources: namespace.sources };
  assert.deepEqual(validateStdlib(index, [page('docs/workshop.md', '/docs/internals/workshop/')], [namespace, author]).map((row) => [row.key, row.internal, row.exports]), [['workshop', true, []]]);
  assert.throws(() => validateStdlib(index, [], [namespace, { ...author, dictionary: { lexicon: 'eng', key: 'missing' } }]), /dictionary owner/);
  assert.throws(() => validateStdlib(index, [], [namespace, { ...author, sources: ['fixtures/workshop/other.l'] }]), /dictionary source/);
  assert.throws(() => validateStdlib(index, [], [namespace, { ...author, dictionary: undefined }]), /dictionary owner/);
});

test('an application dictionary namespace alone is not reviewed export coverage', () => {
  const index = { format: 1, languages: { fra: { math: { entry: 'stdlib/fra/math/dictionary.l', dictionary: 'Math' } } } };
  const namespace = { id: 'math', kind: 'stdlib-export', dictionary: { lexicon: 'fra', key: 'math' }, visibility: 'namespace', support: 'supported', sources: ['stdlib/fra/math/dictionary.l'] };
  assert.throws(() => validateStdlib(index, [page('docs/math.md', '/docs/stdlib/fra/math/')], [namespace]), /public export inventory/);
});

test('fragment resolution rejects missing local and mapped headings', async (t) => {
  const root = await fixture(t, { 'docs/start.md': '# Start\n', 'docs/next.md': '# Next\n' });
  const pages = [page(), page('docs/next.md', '/docs/next/')];
  for (const link of ['[bad](#missing)', '[bad](next.md#missing)', '[bad](/docs/next/#missing)']) {
    const input = `# Start\n\n${link}`;
    await assert.rejects(project(root, input, {
      pages, sourceTexts: new Map([['docs/start.md', input], ['docs/next.md', '# Next\n']]),
    }), /fragment/i);
  }
});

function fragmentCode(output) {
  const elements = (node) => [node, ...(node.childNodes ?? []).flatMap(elements)];
  const blocks = elements(parseFragment(output)).filter((node) => node.tagName === 'pre');
  assert.equal(blocks.length, 1);
  assert.ok(blocks[0].attrs.some((attr) => attr.name === 'tabindex' && attr.value === '0'));
  assert.ok(blocks[0].attrs.some((attr) => attr.name === 'aria-label' && attr.value));
  const code = blocks[0].childNodes[0];
  assert.equal(code.tagName, 'code');
  assert.ok(code.childNodes.every((node) => node.nodeName === '#text'), 'fragments need no syntax-highlighter subtree');
  return code.childNodes.map((node) => node.value).join('');
}

test('source fragments preserve literal code without per-token presentation markup', async (t) => {
  const code = '"literal" is `<BOOK>{value}</BOOK>`.\n*** ```` and [not a link](absent.md) ***\n';
  const input = '# Included source\n\n<!-- example:edge -->\n';
  const root = await fixture(t, { 'docs/start.md': input, 'examples/path with spaces.l': code });
  const { output } = await project(root, input, { examples: [{
    id: 'edge', source: 'examples/path with spaces.l', sha256: sha256(code), lang: 'text',
    description: 'Literal source bytes.', classification: 'fragment', validation: 'existing-witness-not-run',
  }] });
  assert.equal(fragmentCode(output), code);
  assert.match(output, /<div data-pagefind-ignore>[\s\S]*existing-witness-not-run[\s\S]*<\/div>/);
  assert.match(output, /\*\*Fragment\.\*\* Source:/);
  assert.match(output, /<summary>Example evidence<\/summary>[\s\S]*existing-witness-not-run[\s\S]*<\/details>/);
  assert.equal(await fs.readFile(path.join(root, 'examples/path with spaces.l'), 'utf8'), code);
});

test('intentional runtime failures remain visible outside collapsed example evidence', async (t) => {
  const source = 'Exit.\n';
  const input = '# Program\n\n<!-- example:program -->\n';
  const root = await fixture(t, { 'docs/start.md': input, 'program.l': source });
  for (const expectedExit of [0, 1]) {
    const { output } = await project(root, input, { examples: [{
      id: 'program', source: 'program.l', sha256: sha256(source), classification: 'runnable',
      description: 'A fixture for presentation only.', validation: 'existing-witness-not-run', expectedExit,
    }] });
    const visible = output.split('<details>')[0];
    if (expectedExit) assert.match(visible, /Complete program · expected runtime failure/);
    else assert.match(visible, /\*\*Complete program\.\*\*/);
  }
});

test('canonical excerpts retain selected line bytes, escape literal content and identify source lines', async (t) => {
  const selected = '"literal" is `<BOOK>{value}</BOOK>`.\r\n*** ```` and [not a link](absent.md) ***\r\n';
  const code = `not selected before\r\n${selected}not selected after`;
  const input = '# Excerpt\n\n<!-- example:part -->\n';
  const root = await fixture(t, { 'docs/start.md': input, 'examples/path with spaces.l': code });
  const example = {
    id: 'part', source: 'examples/path with spaces.l', sha256: sha256(code), lang: 'l',
    lines: { start: 2, end: 3 }, description: 'Selected canonical definition.',
    classification: 'fragment', validation: 'existing-witness-not-run',
  };
  const { output } = await project(root, input, { examples: [example] });
  assert.ok(output.includes('&quot;literal&quot; is `&lt;BOOK&gt;{value}&lt;/BOOK&gt;`.\r\n*** ```` and [not a link](absent.md) ***\r\n</code></pre>'));
  assert.match(output, /Source: `examples\/path with spaces\.l \(lines 2–3\)`/);
  assert.doesNotMatch(output, /not selected/);
  assert.equal(fragmentCode(output), selected.replaceAll('\r\n', '\n'));
  assert.equal(await fs.readFile(path.join(root, example.source), 'utf8'), code);

  const { output: finalLine } = await project(root, input, { examples: [{ ...example, lines: { start: 4, end: 4 } }] });
  assert.equal(fragmentCode(finalLine), 'not selected after');
  assert.match(finalLine, /lines 4–4/);
});

test('excerpt freshness hashes the complete canonical file before selecting lines', async (t) => {
  const code = '*** outside excerpt ***\n"value" is `inside`.\n';
  const input = '# Excerpt\n\n<!-- example:part -->\n';
  const root = await fixture(t, { 'docs/start.md': input, 'example.l': code });
  const example = { id: 'part', source: 'example.l', sha256: sha256(code), classification: 'fragment', lines: { start: 2, end: 2 } };
  await assert.rejects(project(root, input, { examples: [{ ...example, sha256: sha256('"value" is `inside`.\n') }] }), /Stale example hash/);
  await fs.writeFile(path.join(root, 'example.l'), code.replace('outside excerpt', 'changed outside excerpt'));
  await assert.rejects(project(root, input, { examples: [example] }), /Stale example hash/);
});

test('excerpts reject invalid ranges, nonfragment classifications and invalid UTF-8', async (t) => {
  const code = 'first\nsecond\n';
  const input = '# Excerpt\n\n<!-- example:part -->\n';
  const root = await fixture(t, { 'docs/start.md': input, 'example.l': code });
  const example = { id: 'part', source: 'example.l', sha256: sha256(code), classification: 'fragment', lines: { start: 1, end: 1 } };
  for (const lines of [null, [], '1-2', {}, { start: 0, end: 1 }, { start: 2, end: 1 }, { start: 1.5, end: 2 }, { start: 1, end: '2' }, { start: 1, end: 2, extra: true }, { start: 1, end: 3 }]) {
    await assert.rejects(project(root, input, { examples: [{ ...example, lines }] }), /Invalid example lines|out of bounds/);
  }
  for (const classification of ['runnable', 'compile-only', 'invalid', undefined]) {
    await assert.rejects(project(root, input, { examples: [{ ...example, classification }] }), /excerpts require classification fragment/);
  }
  // Invalid bytes outside the selected line still make the entire source unsuitable for excerpts.
  const invalid = Buffer.concat([Buffer.from('first\n'), Buffer.from([0xff])]);
  await fs.writeFile(path.join(root, 'example.l'), invalid);
  await assert.rejects(project(root, input, { examples: [{ ...example, sha256: sha256(invalid) }] }), /Invalid UTF-8 in excerpt source/);
  await fs.writeFile(path.join(root, 'example.l'), '');
  await assert.rejects(project(root, input, { examples: [{ ...example, sha256: sha256('') }] }), /source has 0 lines/);
});

test('invalid UTF-8 inclusion is explicitly escaped instead of silently replaced', async (t) => {
  const source = Buffer.from([0x61, 0xff, 0x5c, 0x00, 0x0a]);
  const input = '# Bytes\n\n<!-- example:bytes -->\n';
  const root = await fixture(t, { 'docs/start.md': input, 'examples/bytes.l': source });
  const { output } = await project(root, input, { examples: [{ id: 'bytes', source: 'examples/bytes.l', sha256: sha256(source), lang: 'text', description: 'Raw bytes.', classification: 'fragment', validation: 'existing-witness-not-run' }] });
  assert.ok(output.includes('a\\xff\\x5c\\x00\\x0a'));
  assert.match(output, /escaped bytes because the source is not valid UTF-8/);
  assert.ok(!output.includes('\ufffd'));
  assert.deepEqual(await fs.readFile(path.join(root, 'examples/bytes.l')), source);
});

test('inclusions reject stale hashes and unknown IDs', async (t) => {
  const text = '# Example\n\n<!-- example:x -->\n';
  const root = await fixture(t, { 'docs/start.md': text, 'example.l': '"value" is `x`.\n' });
  await assert.rejects(project(root, text), /Unknown example x/);
  await assert.rejects(project(root, text, { examples: [{ id: 'x', source: 'example.l', sha256: '0'.repeat(64) }] }), /Stale example hash/);
});

test('untrusted URL schemes and source traversal do not become navigation', async (t) => {
  const root = await fixture(t, { 'docs/start.md': '# Start\n' });
  for (const url of ['javascript:alert', '//remote.example/page', '../../outside.md', '%2e%2e/%2e%2e/outside.md']) {
    await assert.rejects(project(root, `# Start\n\n[unsafe](${url})`));
  }
});

test('frontmatter source ownership is canonical and generated references have no edit invitation', async (t) => {
  const input = '---\ntitle: Wrong title\neditUrl: https://example.com/wrong\n---\n# Source\n\nText.\n';
  const root = await fixture(t, { 'docs/start.md': input });
  const { output } = await project(root, input, { page: { ...page(), generated: true } });
  const metadata = parseYAML(output.match(/^---\n([\s\S]*?)\n---/)[1]);
  assert.equal(metadata.title, 'Start here');
  assert.equal(metadata.sourcePath, 'docs/start.md');
  assert.equal(metadata.editUrl, false);
  assert.equal(metadata.generatedReference, true);
  assert.match(output, /<div data-pagefind-ignore>[\s\S]*Generated reference:[\s\S]*docs\/start.md[\s\S]*<\/div>/);
});

test('a second indexed dictionary requires both a dedicated page and reviewed public exports', () => {
  const index = { format: 1, languages: { eng: {
    fixture: { entry: 'fixtures/fixture/dictionary.l', dictionary: 'Fixture' },
    extra: { entry: 'stdlib/eng/extra/dictionary.l', dictionary: 'Extra' },
  } } };
  const pages = [page('docs/dictionaries/fixture.md', '/docs/stdlib/eng/fixture/')];
  const entries = [{ kind: 'stdlib-export', dictionary: { lexicon: 'eng', key: 'fixture' }, visibility: 'public', name: 'Fixture.Draw', sources: ['fixtures/fixture/dictionary.l'] }];
  assert.throws(() => validateStdlib(index, pages, entries), /Unclassified indexed dictionary eng\/extra/);
  pages.push(page('docs/stdlib/extra.md', '/docs/stdlib/eng/extra/'));
  assert.throws(() => validateStdlib(index, pages, entries), /Missing reviewed export inventory for eng\/extra/);
  entries.push({ kind: 'stdlib-export', dictionary: { lexicon: 'eng', key: 'extra' }, visibility: 'public', name: 'Extra.Value', sources: ['stdlib/eng/extra/dictionary.l'] });
  entries.push({ kind: 'stdlib-export', dictionary: { lexicon: 'eng', key: 'extra' }, visibility: 'private', name: 'Extra.Hidden', sources: ['stdlib/eng/extra/dictionary.l'] });
  entries.push({ kind: 'stdlib-export', dictionary: { lexicon: 'eng', key: 'extra' }, visibility: 'namespace', name: 'Extra', sources: ['stdlib/eng/extra/dictionary.l'] });
  assert.deepEqual(validateStdlib(index, pages, entries).map((row) => [row.key, row.exports]), [
    ['extra', ['Extra.Value']], ['fixture', ['Fixture.Draw']],
  ]);
});

test('catalog includes implementation references without claiming public exports', () => {
  for (const [key, dictionary, entry] of [
    ['support', 'Support', 'stdlib/eng/support.l'],
  ]) {
    const index = { format: 1, languages: { test: { [key]: { entry, dictionary } } } };
    const entries = [{ kind: 'stdlib-export', dictionary: { lexicon: 'test', key }, visibility: 'namespace', support: 'internal', doc: 'docs/support.md#details', sources: [entry] }];
    const reference = page('docs/support.md', '/docs/internals/support/');
    const [row] = validateStdlib(index, [reference], entries);
    assert.ok(row, 'Implementation dictionary must appear in the combined catalog.');
    assert.deepEqual([row.language, row.key, row.internal, row.exports, row.route, row.description], ['test', key, true, [], reference.route, reference.description]);
    assert.throws(() => validateStdlib(index, [], entries), /Missing implementation reference/);
    entries.push({ kind: 'stdlib-export', dictionary: { lexicon: 'test', key }, visibility: 'public', support: 'supported', sources: [entry] });
    assert.throws(() => validateStdlib(index, [], entries), /conflicting public documentation classifications/);
  }
});

test('export ownership is distinct from reviewed source dependencies', () => {
  const index = { format: 1, languages: { eng: {
    fixture: { entry: 'fixtures/fixture/dictionary.l', dictionary: 'Fixture' },
    support: { entry: 'stdlib/eng/support.l', dictionary: 'Support' },
  } } };
  const pages = [page('docs/dictionaries/fixture.md', '/docs/stdlib/eng/fixture/'), page('docs/support.md', '/docs/internals/support/')];
  const owner = { lexicon: 'eng', key: 'fixture' };
  const entries = [
    { kind: 'stdlib-export', dictionary: owner, visibility: 'public', support: 'supported', name: 'Fixture.Bold', sources: ['fixtures/fixture/dictionary.l', 'stdlib/eng/support.l'] },
    { kind: 'stdlib-export', dictionary: { lexicon: 'eng', key: 'support' }, visibility: 'namespace', support: 'internal', doc: 'docs/support.md', sources: ['stdlib/eng/support.l'] },
  ];
  assert.deepEqual(validateStdlib(index, pages, entries).filter((row) => !row.internal).map((row) => [row.key, row.exports]), [['fixture', ['Fixture.Bold']]]);
  for (const dictionary of [undefined, { lexicon: 'eng', key: 'unknown' }]) {
    assert.throws(() => validateStdlib(index, pages, [{ ...entries[0], dictionary }]), /Invalid reviewed dictionary owner/);
  }
  assert.throws(() => validateStdlib(index, pages, [{ ...entries[0], sources: ['stdlib/eng/support.l'] }]), /Missing reviewed dictionary source/);
});

test('portable containment and safe files reject traversal and external symlinks', async (t) => {
  const root = await fixture(t, { 'inside.txt': 'inside' });
  const outside = await fixture(t, { 'private.txt': 'outside' });
  for (const name of ['../private.txt', '/private.txt', 'folder\\file']) assert.throws(() => contained(root, name));
  assert.equal(await safeFile(root, 'inside.txt'), path.join(root, 'inside.txt'));
  await fs.symlink(path.join(outside, 'private.txt'), path.join(root, 'link.txt'));
  await assert.rejects(safeFile(root, 'link.txt'), /Symlink escapes/);
  await assert.rejects(safeFile(root, '.'), /Expected a file/);
});

test('directory source links reject symlinks escaping the canonical repository', async (t) => {
  const root = await fixture(t, { 'docs/start.md': '# Start\n' });
  const outside = await fixture(t, { 'private.l': 'outside data' });
  await fs.symlink(outside, path.join(root, 'docs/linked-directory'));
  await assert.rejects(project(root, '# Start\n\n[Source directory](linked-directory/)'), /Escaped source directory link/);
});

test('atomic writes avoid rewriting equal bytes and leave no temporary files', async (t) => {
  const root = await fixture(t);
  const target = path.join(root, 'nested/path with spaces.md');
  assert.equal(await atomicWrite(target, 'first\n'), true);
  const fixedTime = new Date('2001-01-01T00:00:00.000Z');
  await fs.utimes(target, fixedTime, fixedTime);
  assert.equal(await atomicWrite(target, Buffer.from('first\n')), false);
  assert.equal((await fs.stat(target)).mtimeMs, fixedTime.getTime());
  assert.equal(await atomicWrite(target, 'second\n'), true);
  assert.equal(await fs.readFile(target, 'utf8'), 'second\n');
  assert.deepEqual(await fs.readdir(path.dirname(target)), ['path with spaces.md']);
});

test('owned projection removes renamed/deleted outputs and preserves authored files', async (t) => {
  const root = await fixture(t, { 'src/content/docs/index.mdx': 'authored landing', 'src/content/docs/docs/manual.md': 'authored article' });
  const first = 'src/content/docs/docs/old/index.md';
  const renamed = 'src/content/docs/docs/new/index.md';
  await publishOwned(root, new Map([[first, 'old'], ['public/version.json', '{"channel":"local"}\n']]));
  await publishOwned(root, new Map([[renamed, 'new'], ['public/version.json', '{"channel":"local"}\n']]));
  await assert.rejects(fs.access(path.join(root, first)));
  assert.equal(await fs.readFile(path.join(root, renamed), 'utf8'), 'new');
  assert.equal(await fs.readFile(path.join(root, 'src/content/docs/index.mdx'), 'utf8'), 'authored landing');
  assert.equal(await fs.readFile(path.join(root, 'src/content/docs/docs/manual.md'), 'utf8'), 'authored article');
  assert.deepEqual(JSON.parse(await fs.readFile(path.join(root, '.generated/owned.json'), 'utf8')), ['public/version.json', renamed]);
});

test('owned projection validates the whole batch before replacing authored files', async (t) => {
  const root = await fixture(t, { 'src/content/docs/docs/authored.md': 'original' });
  await assert.rejects(publishOwned(root, new Map([
    ['public/version.json', '{}'], ['src/content/docs/docs/authored.md', 'replacement'],
  ])), /authored\/unowned/);
  await assert.rejects(fs.access(path.join(root, 'public/version.json')));
  assert.equal(await fs.readFile(path.join(root, 'src/content/docs/docs/authored.md'), 'utf8'), 'original');
  await assert.rejects(publishOwned(root, new Map([['src/content/docs/index.mdx', 'replacement']])), /unowned output/);
});

test('owned projection rejects normalized traversal through an allowed prefix', async (t) => {
  const root = await fixture(t, { 'README.md': 'authored' });
  await assert.rejects(publishOwned(root, new Map([['src/content/docs/docs/../../../../README.md', 'unsafe']])));
  assert.equal(await fs.readFile(path.join(root, 'README.md'), 'utf8'), 'authored');
});

test('owned projection rejects a parent symlink escaping the website', async (t) => {
  const root = await fixture(t);
  const outside = await fixture(t, { 'index.md': 'outside authored bytes' });
  await fs.mkdir(path.join(root, 'src/content/docs/docs'), { recursive: true });
  await fs.symlink(outside, path.join(root, 'src/content/docs/docs/escaped'));
  await assert.rejects(publishOwned(root, new Map([['src/content/docs/docs/escaped/index.md', 'unsafe']])));
  assert.equal(await fs.readFile(path.join(outside, 'index.md'), 'utf8'), 'outside authored bytes');
});

test('a forged previous ownership record cannot delete a file outside its managed directory', async (t) => {
  const root = await fixture(t, {
    'README.md': 'authored',
    '.generated/owned.json': JSON.stringify(['src/content/docs/docs/../../../../README.md']),
  });
  await assert.rejects(publishOwned(root, new Map()));
  assert.equal(await fs.readFile(path.join(root, 'README.md'), 'utf8'), 'authored');
});

test('clean archive preparation is deterministic and reconciles canonical renames', async (t) => {
  const mapped = { schemaVersion: 1, pages: [page()] };
  const root = await fixture(t, {
    'docs/start.md': '# Start\n\nSource archive content.\n',
    'docs/_meta/examples.json': '{"schemaVersion":1,"examples":[]}',
    'website/content-map.json': JSON.stringify(mapped),
    'website/src/content/docs/index.mdx': '# Authored landing\n',
  });
  const site = path.join(root, 'website');
  const first = await prepare({ root, site, verifyReferences: false });
  assert.equal(first.pages, 1);
  assert.equal(first.version.docsCommit, null);
  assert.equal(first.version.channel, 'local');
  const old = path.join(site, 'src/content/docs/docs/guide/index.md');
  const before = await fs.readFile(old);
  const second = await prepare({ root, site, verifyReferences: false });
  assert.deepEqual(second, first);
  assert.deepEqual(await fs.readFile(old), before);
  await fs.rename(path.join(root, 'docs/start.md'), path.join(root, 'docs/renamed.md'));
  mapped.pages = [page('docs/renamed.md', '/docs/renamed/')];
  await fs.writeFile(path.join(site, 'content-map.json'), JSON.stringify(mapped));
  await prepare({ root, site, verifyReferences: false });
  await assert.rejects(fs.access(old));
  assert.match(await fs.readFile(path.join(site, 'src/content/docs/docs/renamed/index.md'), 'utf8'), /Source archive content/);
  assert.equal(await fs.readFile(path.join(site, 'src/content/docs/index.mdx'), 'utf8'), '# Authored landing\n');
});

test('pages omitted from the sidebar retain their routes, fragments and source links', async (t) => {
  const start = page('docs/start.md', '/docs/start/');
  const reference = { ...page('docs/reference.md', '/docs/reference/'), sidebar: false };
  const root = await fixture(t, {
    'docs/start.md': '# Start\n\n[Reference](reference.md#details)\n',
    'docs/reference.md': '# Reference\n\n## Details\n\nReference content.\n',
    'docs/_meta/examples.json': '{"schemaVersion":1,"examples":[]}',
    'website/content-map.json': JSON.stringify({ schemaVersion: 1, pages: [start, reference] }),
  });
  const site = path.join(root, 'website');
  assert.equal((await prepare({ root, site, verifyReferences: false })).pages, 2);
  const sidebar = JSON.parse(await fs.readFile(path.join(site, '.generated/sidebar.json'), 'utf8'));
  assert.deepEqual(sidebar.flatMap((group) => group.items.map((item) => item.link)), ['/docs/start/']);
  assert.match(await fs.readFile(path.join(site, 'src/content/docs/docs/start/index.md'), 'utf8'), /\/docs\/reference\/#details/);
  const content = await fs.readFile(path.join(site, 'src/content/docs/docs/reference/index.md'), 'utf8');
  assert.match(content, /## Details/);
  assert.match(content, /Canonical source:[\s\S]*docs\/reference.md/);
});

test('catalog rendering handles additional languages and libraries without dictionary-specific code', () => {
  const rendered = stdlibCatalog([
    { language: 'eng', key: 'fixture', dictionary: 'Fixture', exports: ['Fixture.Draw'], route: '/docs/stdlib/eng/fixture/', description: 'Test-owned shapes.' },
    { language: 'eng', key: 'vectors', dictionary: 'Vectors', exports: ['Vectors.Length', 'Vectors.Dot'], route: '/docs/stdlib/eng/vectors/', description: 'Vector arithmetic.' },
    { language: 'test', key: 'sample', dictionary: 'Sample', exports: [], internal: true, route: '/docs/internals/sample/', description: 'Internal A | B support.' },
  ]);
  assert.match(rendered, /Application \| Vector arithmetic\./);
  assert.match(rendered, /Implementation \| Internal A &#124; B support\./);
  assert.match(rendered, /\[Vectors\]\(\/docs\/stdlib\/eng\/vectors\/\)/);
  assert.match(rendered, /\[Sample\]\(\/docs\/internals\/sample\/\)/);
});

test('example verification replaces prior success even when the current run fails', async (t) => {
  const source = 'Create "Fixture" as a New Basic App.\nThe App Prints `checked`.\n';
  for (const scenario of [
    { name: 'invalid invocation', stage: 'configuration', message: /provide --compiler/ },
    { name: 'invalid manifest', stage: 'configuration', message: /JSON/ },
    { name: 'changed source', stage: 'source-validation', message: /Source changed/ },
    { name: 'build failure', stage: 'native-compilation', message: /build failed/ },
    { name: 'runtime failure', stage: 'native-runtime', message: /runtime output\/status differs/ },
    { name: 'success', stage: 'native-runtime' },
  ]) {
    await t.test(scenario.name, async (t) => {
      const manifest = { examples: [{ id: 'current', source: 'example.l', classification: 'runnable', sha256: sha256(source), expectedExit: 0, expectedStdoutBase64: Buffer.from('checked\n').toString('base64') }] };
      if (scenario.name === 'changed source') manifest.examples[0].sha256 = '0'.repeat(64);
      const root = await fixture(t, {
        'website/scripts/examples.mjs': await fs.readFile(new URL('../scripts/examples.mjs', import.meta.url)),
        'website/scripts/lib.mjs': await fs.readFile(new URL('../scripts/lib.mjs', import.meta.url)),
        'website/reports/examples-verification.json': JSON.stringify({ compiler: 'old', outcomes: [{ id: 'old', stage: 'native-runtime', status: 'passed' }] }),
        'docs/_meta/examples.json': scenario.name === 'invalid manifest' ? '{' : JSON.stringify(manifest),
        'example.l': source,
      });
      const compiler = path.join(root, 'compiler.mjs');
      await fs.writeFile(compiler, `#!${process.execPath}\nimport fs from 'node:fs';\nimport path from 'node:path';\nif (${scenario.name === 'build failure'}) process.exit(2);\nfs.writeFileSync(path.join(process.argv[4], path.basename(process.argv[5], '.l')), ${JSON.stringify('#!/bin/sh\nprintf "checked\\n"\nexit ' + (scenario.name === 'runtime failure' ? 3 : 0) + '\n')}, {mode:0o755});\n`, { mode: 0o755 });
      const args = scenario.name === 'invalid invocation' ? [] : ['--compiler', compiler];
      const result = spawnSync(process.execPath, [path.join(root, 'website/scripts/examples.mjs'), ...args], { encoding: 'utf8', timeout: 10000 });
      assert.ifError(result.error);
      assert.equal(result.status, scenario.message ? 1 : 0, result.stderr);
      if (scenario.message) assert.match(result.stderr, scenario.message);
      const report = JSON.parse(await fs.readFile(path.join(root, 'website/reports/examples-verification.json'), 'utf8'));
      assert.ok(report.outcomes.every((outcome) => outcome.id !== 'old'), 'failed run left the prior successful report current');
      assert.equal(report.status, scenario.message ? 'failed' : 'passed');
      assert.equal(report.exitCode, result.status);
      const outcome = report.outcomes.at(-1);
      assert.equal(outcome.stage, scenario.stage);
      assert.equal(outcome.status, report.status);
      if (scenario.message) assert.match(outcome.error, scenario.message);
    });
  }
});
