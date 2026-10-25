import assert from 'node:assert/strict';
import fs from 'node:fs/promises';
import os from 'node:os';
import path from 'node:path';
import test from 'node:test';
import { parse as parseYAML } from 'yaml';
import { anchors, parseMarkdown, preparePage, validateMap, validateStdlib, stdlibCatalog } from '../scripts/content.mjs';
import { prepare, publishOwned } from '../scripts/prepare.mjs';
import { atomicWrite, contained, safeFile, sha256 } from '../scripts/lib.mjs';

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

test('map validation enforces route, ownership and navigation metadata', () => {
  assert.deepEqual(validateMap({ schemaVersion: 1, pages: [page()] }).pages, [page()]);
  assert.throws(() => validateMap({ pages: [] }), /schemaVersion/);
  assert.throws(() => validateMap({ schemaVersion: 1, pages: [page(), page('docs/other.md')] }), /Duplicate route/);
  assert.throws(() => validateMap({ schemaVersion: 1, pages: [page(), page('docs/start.md', '/docs/other/')] }), /Duplicate canonical/);
  for (const route of ['/docs/guide', '/l/docs/guide/', '/docs/_meta/', '/docs/../guide/', '/']) {
    assert.throws(() => validateMap({ schemaVersion: 1, pages: [page('docs/start.md', route)] }), /route/i);
  }
  for (const override of [{ title: '' }, { order: NaN }, { generated: 'false' }, { role: 'probably-supported' }, { source: 'docs/_meta/private.md' }]) {
    assert.throws(() => validateMap({ schemaVersion: 1, pages: [{ ...page(), ...override }] }));
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

test('source inclusions safely fence braces, backticks and tags and retain bytes', async (t) => {
  const code = '"literal" is `<BOOK>{value}</BOOK>`.\n*** ```` and [not a link](absent.md) ***\n';
  const input = '# Included source\n\n<!-- example:edge -->\n';
  const root = await fixture(t, { 'docs/start.md': input, 'examples/path with spaces.l': code });
  const { output } = await project(root, input, { examples: [{
    id: 'edge', source: 'examples/path with spaces.l', sha256: sha256(code), lang: 'text',
    description: 'Literal source bytes.', classification: 'fragment', validation: 'existing-witness-not-run',
  }] });
  assert.ok(output.includes('`````text title="examples/path with spaces.l"\n' + code + '`````'));
  const tree = parseMarkdown(output.replace(/^---\n[\s\S]*?\n---\n/, ''));
  assert.equal(tree.children.filter((node) => node.type === 'code').length, 1);
  assert.equal(tree.children.find((node) => node.type === 'code').value, code.trimEnd());
  assert.match(output, /existing-witness-not-run/);
  assert.equal(await fs.readFile(path.join(root, 'examples/path with spaces.l'), 'utf8'), code);
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
  assert.ok(output.includes('`````l title="examples/path with spaces.l (lines 2–3)"\n' + selected + '`````'));
  assert.match(output, /Source: `examples\/path with spaces\.l \(lines 2–3\)`/);
  assert.doesNotMatch(output, /not selected/);
  const tree = parseMarkdown(output.replace(/^---\n[\s\S]*?\n---\n/, ''));
  const blocks = tree.children.filter((node) => node.type === 'code');
  assert.equal(blocks.length, 1);
  assert.equal(blocks[0].value, selected.trimEnd());
  assert.equal(await fs.readFile(path.join(root, example.source), 'utf8'), code);

  const { output: finalLine } = await project(root, input, { examples: [{ ...example, lines: { start: 4, end: 4 } }] });
  assert.ok(finalLine.includes('\nnot selected after\n```'));
  assert.match(finalLine, /lines 4–4/);
});

test('excerpt freshness hashes the complete canonical file before selecting lines', async (t) => {
  const code = 'outside excerpt\nShow `inside`.\n';
  const input = '# Excerpt\n\n<!-- example:part -->\n';
  const root = await fixture(t, { 'docs/start.md': input, 'example.l': code });
  const example = { id: 'part', source: 'example.l', sha256: sha256(code), classification: 'fragment', lines: { start: 2, end: 2 } };
  await assert.rejects(project(root, input, { examples: [{ ...example, sha256: sha256('Show `inside`.\n') }] }), /Stale example hash/);
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
  const root = await fixture(t, { 'docs/start.md': text, 'example.l': 'Show `x`.\n' });
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
  assert.match(output, /Refresh the canonical snapshot/);
});

test('a second indexed dictionary requires both a dedicated page and reviewed public exports', () => {
  const index = { format: 1, languages: { eng: {
    io: { entry: 'stdlib/eng/io/dictionary.l', dictionary: 'IO' },
    extra: { entry: 'stdlib/eng/extra/dictionary.l', dictionary: 'Extra' },
  } } };
  const pages = [page('docs/stdlib/io.md', '/docs/stdlib/eng/io/')];
  const entries = [{ kind: 'stdlib-export', visibility: 'public', name: 'IO.Print', sources: ['stdlib/eng/io/dictionary.l'] }];
  assert.throws(() => validateStdlib(index, pages, entries), /Unclassified indexed dictionary eng\/extra/);
  pages.push(page('docs/stdlib/extra.md', '/docs/stdlib/eng/extra/'));
  assert.throws(() => validateStdlib(index, pages, entries), /Missing reviewed export inventory for eng\/extra/);
  entries.push({ kind: 'stdlib-export', visibility: 'public', name: 'Extra.Value', sources: ['stdlib/eng/extra/dictionary.l'] });
  entries.push({ kind: 'stdlib-export', visibility: 'private', name: 'Extra.Hidden', sources: ['stdlib/eng/extra/dictionary.l'] });
  entries.push({ kind: 'stdlib-export', visibility: 'namespace', name: 'Extra', sources: ['stdlib/eng/extra/dictionary.l'] });
  assert.deepEqual(validateStdlib(index, pages, entries).map((row) => [row.key, row.exports]), [
    ['extra', ['Extra.Value']], ['io', ['IO.Print']],
  ]);
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

test('catalog rendering handles additional languages and libraries without IO-specific code', () => {
  const rendered = stdlibCatalog([
    { language: 'eng', key: 'io', dictionary: 'IO', exports: ['IO.Print'], route: '/docs/stdlib/eng/io/' },
    { language: 'eng', key: 'vectors', dictionary: 'Vectors', exports: ['Vectors.Length', 'Vectors.Dot'], route: '/docs/stdlib/eng/vectors/' },
    { language: 'test', key: 'sample', dictionary: 'Sample', exports: ['Sample.Read'], route: '/docs/stdlib/test/sample/' },
  ]);
  assert.match(rendered, /Vectors.Length, Vectors.Dot/);
  assert.match(rendered, /\[Vectors\]\(\/docs\/stdlib\/eng\/vectors\/\)/);
  assert.match(rendered, /\[Sample\]\(\/docs\/stdlib\/test\/sample\/\)/);
});
