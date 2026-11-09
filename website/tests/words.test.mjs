import assert from 'node:assert/strict';
import test from 'node:test';
import fs from 'node:fs/promises';
import { parseDocument } from '../scripts/content.mjs';
import { deriveWords, wordFacts, wordIndex } from '../scripts/words.mjs';

const page = { source: 'docs/words.md', route: '/docs/test/' };
const entry = (id, name, anchor, extra = {}) => ({ id, name, kind: 'builtin', support: 'supported', doc: `${page.source}#${anchor}`, word: { partOfSpeech: 'verb' }, evidence: 'static-inspection', sources: ['source.l'], ...extra });
const derive = (entries, source, extra = {}) => deriveWords({ entries, pages: [page], documents: new Map([[page.source, parseDocument(source)]]), ...extra });

test('nested senses keep their own definitions and literal directives never register', () => {
  const source = '# Reference\n\n## Option.Select\n\n<!-- word:option -->\n\nSelect `Some` or [None][none] exactly once.\n\n### Result.Select\n\n<!-- word:result -->\n\nSelect the Result branch.\n\n**Example**\n\n```md\n<!-- word:option -->\n```\n\n`<!-- word:result -->`\n\n[none]: #optionselect\n';
  const words = derive([entry('option', 'Storage.Option.Select', 'optionselect'), entry('result', 'Storage.Result.Select', 'resultselect')], source);
  assert.deepEqual(words.map((word) => word.gloss), ['Select Some or None exactly once.', 'Select the Result branch.']);
  assert.notEqual(words[0].link, words[1].link);
  assert.equal(words[0].root, 'Select');
  assert.ok(words[0].section.end >= words[1].section.end);
});

test('real inline markers and headings cannot substitute for a definition', () => {
  const item = entry('x', 'Meaning', 'meaning');
  for (const source of [
    '# Words\n\n## Meaning\n\nText <!-- word:x --> more text.\n',
    '# Words\n\n## Meaning\n\n<!-- word:x -->\n\n### Usage\n\nAnother section.\n',
  ]) assert.throws(() => derive([item], source), /standalone|definition paragraph/);
});

test('a generated label or example directive is not a substantive definition', () => {
  for (const following of ['**Usage**', '**Example**', '<!-- example:sample -->', '<!-- word:y -->']) {
    assert.throws(() => derive([entry('x', 'Meaning', 'meaning')], `# Words\n\n## Meaning\n\n<!-- word:x -->\n\n${following}\n`), /definition paragraph/);
  }
});

test('explicit canonical anchors are accepted without generating another ID', () => {
  const words = derive([entry('x', 'Meaning', 'kept-anchor')], '# Words\n\n## Meaning\n\n<a id="kept-anchor"></a>\n\n<!-- word:x -->\n\nA retained meaning.\n');
  const facts = wordFacts(words[0], words, (source) => source);
  assert.equal(words[0].link, '/docs/test/#kept-anchor');
  assert.doesNotMatch(facts, / id=/);
});

test('canonical associations are direct and internal tags cannot become application entries', () => {
  const source = '# Words\n\n## Meaning\n\n<!-- word:x -->\n\nA meaning.\n';
  const canonical = entry('x', 'Meaning', 'meaning');
  const alias = { ...entry('review', 'Meaning', 'meaning'), word: undefined, wordRef: 'x' };
  assert.equal(derive([canonical, alias], source).length, 1);
  assert.throws(() => derive([canonical, alias, { ...alias, id: 'chain', wordRef: 'review' }], source), /direct canonical/);
  assert.throws(() => derive([{ ...canonical, kind: 'value-tag' }], source), /cannot be promoted/);
  assert.throws(() => derive([canonical], source, { lexical: { entries: { root: { support: 'supported', notes: 'Reviewed.' } } } }), /configured word association/);
  assert.equal(derive([canonical], source, { lexical: { entries: { native: { support: 'internal', notes: 'Native metadata, not a source constructor.' } } } }).length, 1);
});

test('an application dictionary namespace is a documented noun without publishing implementation namespaces', () => {
  const namespace = entry('io', 'IO', 'io', { kind: 'stdlib-export', visibility: 'namespace', dictionary: { lexicon: 'eng', key: 'io' }, word: { partOfSpeech: 'noun' } });
  const source = '# Words\n\n## IO\n\n<!-- word:io -->\n\nName the explicitly imported output dictionary.\n';
  const words = derive([namespace], source);
  assert.equal(words[0].audience, 'application');
  assert.match(wordIndex(words, 'eng:io', [{ language: 'eng', key: 'io', dictionary: 'IO' }]), /IO.*noun/);
  assert.doesNotMatch(wordIndex(words, 'builtins'), /IO/);
  assert.throws(() => derive([{ ...namespace, word: undefined }], source), /Missing word coverage/);
  assert.throws(() => derive([{ ...namespace, support: 'internal' }], source), /cannot be promoted/);
});

test('associated reviews cannot give the same canonical meaning conflicting support labels', () => {
  const source = '# Words\n\n## Meaning\n\n<!-- word:x -->\n\nA meaning.\n';
  const canonical = entry('x', 'Meaning', 'meaning');
  const alias = { ...entry('review', 'Meaning', 'meaning'), word: undefined, wordRef: 'x', support: 'partial' };
  assert.throws(() => derive([canonical, alias], source), /Conflicting word support.*review.*x/);
  const lexical = { entries: { root: { wordRef: 'x', support: 'partial', notes: 'Reviewed.' } } };
  assert.throws(() => derive([canonical], source, { lexical }), /Conflicting word support.*root.*x/);
  alias.support = 'supported';
  lexical.entries.root.support = 'supported';
  assert.equal(derive([canonical, alias], source, { lexical }).length, 1);
});

test('irregular and actor forms retain qualified modifier targets', () => {
  const words = derive([
    entry('send', 'IO.Send', 'send', { inflections: ['Send', 'Sends', 'Sent', 'Sending', 'Sender'], word: { partOfSpeech: 'verb', formRoles: { Sender: { partOfSpeech: 'noun', role: 'actor' } } } }),
    entry('lined', 'IO.Lined', 'lined', { word: { partOfSpeech: 'adjective', targets: [{ entry: 'send', form: 'Sender' }] } }),
  ], '# IO\n\n## Send\n\n<!-- word:send -->\n\nSend exact bytes.\n\n## Lined\n\n<!-- word:lined -->\n\nAppend LF after successful payload.\n');
  const send = words.find((word) => word.id === 'send'), lined = words.find((word) => word.id === 'lined');
  assert.match(wordFacts(send, words, String), /Sender<\/code> \(noun, actor\)/);
  assert.match(wordFacts(lined, words, String), /href="\/docs\/test\/#send".*IO.Send · Sender/);
  assert.doesNotMatch(wordFacts(send, words, String), /Sended/);
});

test('every declared form has an explicit part of speech in its numbered list', () => {
  const words = derive([
    entry('send', 'IO.Send', 'send', { inflections: ['Send', 'Sends', 'Sent', 'Sending', 'Sender'], word: { partOfSpeech: 'verb', formRoles: { Sender: { partOfSpeech: 'noun', role: 'actor' } } } }),
    entry('function', 'Function', 'function', { inflections: ['Function', 'Functions', 'Functioning'], word: { partOfSpeech: 'noun', formRoles: { Functioning: { partOfSpeech: 'verb', role: 'applicative (noun active form)' } } } }),
    entry('lined', 'IO.Lined', 'lined', { word: { partOfSpeech: 'adjective', targets: [{ entry: 'send', form: 'Sender' }] } }),
  ], '# Words\n\n## Send\n\n<!-- word:send -->\n\nSend exact bytes.\n\n## Function\n\n<!-- word:function -->\n\nDefine a callable word.\n\n## Lined\n\n<!-- word:lined -->\n\nAppend LF after successful payload.\n');
  const expected = {
    send: ['<code>Send</code> (verb)', '<code>Sends</code> (verb)', '<code>Sent</code> (verb)', '<code>Sending</code> (verb)', '<code>Sender</code> (noun, actor)'],
    function: ['<code>Function</code> (noun)', '<code>Functions</code> (noun)', '<code>Functioning</code> (verb, applicative (noun active form))'],
    lined: ['<code>Lined</code> (adjective)'],
  };
  for (const word of words) {
    const facts = wordFacts(word, words, (source) => `<a href="/${source}">${source}</a>`);
    const lists = [...facts.matchAll(/<ol>([\s\S]*?)<\/ol>/g)];
    assert.equal(lists.length, 1, `${word.name} has one numbered forms list`);
    assert.deepEqual([...lists[0][1].matchAll(/<li>([\s\S]*?)<\/li>/g)].map((match) => match[1]), expected[word.id]);
    assert.match(facts, /Source and review evidence/);
    assert.match(facts, /static-inspection/);
    assert.match(facts, /<a href="\/source\.l">source\.l<\/a>/);
    if (word.id === 'lined') assert.match(facts, /href="\/docs\/test\/#send".*IO.Send · Sender/);
  }
});

test('word definitions reference reviewed source lines without adding authored code', () => {
  const source = '# Words\n\n## Meaning\n\n<!-- word:x -->\n\nA meaning.\n\n<!-- word-definition:x -->\n';
  const item = entry('x', 'Meaning', 'meaning', { word: { partOfSpeech: 'verb', definitions: [{ source: 'source.l', lines: [1, 3] }] } });
  const [word] = derive([item], source);
  assert.equal(word.definitionMarker.value, '<!-- word-definition:x -->');
  for (const definitions of [[], {}, [null], [{ source: 'source.l' }], [{ source: 'source.l', lines: [0, 2] }], [{ source: 'source.l', lines: [2, 1] }], [{ source: 'source.l', lines: [1, 2.5] }], [{ source: 'source.l', lines: [1, 2, 3] }], [{ source: 'source.l', lines: [1, 2], code: 'fabricated' }], [{ source: 'unreviewed.l', lines: [1, 2] }], [{ source: 'source.l', key: 'verbs.meaning' }]]) {
    assert.throws(() => derive([{ ...item, word: { ...item.word, definitions } }], source), /Invalid word definition|Unreviewed word definition source/);
  }
});

test('word-definition directives occur once inside their nearest canonical word section', () => {
  const item = entry('x', 'Meaning', 'meaning', { word: { partOfSpeech: 'verb', definitions: [{ source: 'source.l', lines: [1, 1] }] } });
  const start = '# Words\n\n## Meaning\n\n<!-- word:x -->\n\nA meaning.\n';
  const marker = '\n<!-- word-definition:x -->\n';
  const literal = '\n```markdown\n<!-- word-definition:x -->\n```\n\n`<!-- word-definition:x -->`\n';
  assert.equal(derive([item], start + marker + literal).length, 1);
  assert.throws(() => derive([item], start + literal), /Missing word-definition marker/);
  assert.throws(() => derive([item], start + marker + marker), /Duplicate word-definition marker/);
  assert.throws(() => derive([item], start + '\n## Other\n' + marker), /canonical word section/);
  assert.throws(() => derive([item], start + '\nText <!-- word-definition:x --> after.\n'), /standalone/);
  assert.throws(() => derive([item], start + '\n> <!-- word-definition:x -->\n'), /standalone/);
  assert.throws(() => derive([entry('x', 'Meaning', 'meaning')], start + marker), /without reviewed definitions/);
  assert.throws(() => derive([item], start + '\n<!-- word-definition:unknown -->\n'), /Unknown word-definition/);
  const child = entry('y', 'Nested', 'nested');
  assert.throws(() => derive([item, child], start + '\n### Nested\n\n<!-- word:y -->\n\nA nested meaning.\n' + marker), /canonical word section/);
});

test('one derived collection renders generic library and global navigation with stable ownership ordering', () => {
  const library = { lexicon: 'fra', key: 'math' };
  const entries = [
    entry('z', 'Z.Output', 'z', { dictionary: library }),
    entry('a', 'Output', 'a'),
    entry('b', 'Storage.Option.Select', 'b', { kind: 'authoring-word', dictionary: { lexicon: 'eng', key: 'storage' }, word: { partOfSpeech: 'verb', audience: 'authoring' } }),
  ];
  const source = '# Words\n\n## Z\n\n<!-- word:z -->\n\nAn output.\n\n## A\n\n<!-- word:a -->\n\nThe other output.\n\n## B\n\n<!-- word:b -->\n\nA selection.\n';
  const words = derive(entries, source);
  assert.deepEqual(words.map((word) => word.id), ['a', 'z', 'b']);
  assert.deepEqual(derive([...entries].reverse(), source).map((word) => word.id), ['a', 'z', 'b']);
  const global = wordIndex(words);
  assert.match(global, /Advanced word authoring/);
  assert.match(global, /data-pagefind-ignore/);
  assert.doesNotMatch(global, /href="#words-0-a"/);
  const local = wordIndex(words, 'fra:math', [{ language: 'fra', key: 'math' }]);
  assert.match(local, /Z.Output/);
  assert.doesNotMatch(local, /other output|Option.Select/);
  assert.throws(() => wordIndex(words, 'fra:unknown', []), /Unknown word-index library/);
  assert.throws(() => wordIndex(words, 'fra:math', [{ language: 'fra', key: 'math', internal: true }]), /Unknown word-index library/);
  assert.doesNotMatch(wordIndex(words, 'builtins'), /Z.Output|Option.Select/);
});

test('internal dictionary indexes include only reviewed authoring words and preserve their audience', () => {
  const dictionary = { lexicon: 'eng', key: 'storage' };
  const entries = [
    entry('select', 'Storage.Option.Select', 'select', { kind: 'authoring-word', dictionary, word: { partOfSpeech: 'verb', audience: 'authoring' } }),
    entry('foreign', 'Foreign.Require', 'foreign', { kind: 'authoring-word', dictionary: { lexicon: 'fra', key: 'storage' }, word: { partOfSpeech: 'verb', audience: 'authoring' } }),
    entry('application', 'Storage.Application', 'application', { dictionary }),
    entry('private', 'Storage.Private', 'private', { kind: 'stdlib-export', dictionary, visibility: 'private', word: undefined }),
    entry('internal', 'Storage.Internal', 'internal', { kind: 'value-tag', dictionary, support: 'internal', word: undefined }),
  ];
  const source = '# Words\n\n## Select\n\n<!-- word:select -->\n\nSelect the matching branch.\n\n## Foreign\n\n<!-- word:foreign -->\n\nRequire another dictionary.\n\n## Application\n\n<!-- word:application -->\n\nAn application meaning.\n';
  const words = derive(entries, source);
  const registries = [{ language: 'eng', key: 'storage', internal: true }];
  const local = wordIndex(words, 'eng:storage', registries);
  assert.match(local, /Storage\.Option\.Select.*\(verb, supported, authoring\)/);
  assert.doesNotMatch(local, /Foreign|Application|Private|Internal/);
  const global = wordIndex(words);
  const application = global.slice(0, global.indexOf('<nav', global.indexOf('</nav>')));
  assert.doesNotMatch(application, /Storage\.Option\.Select|Foreign\.Require/);
  assert.match(global, /Advanced word authoring[\s\S]*Storage\.Option\.Select/);
  assert.doesNotMatch(wordIndex(words, 'builtins'), /Storage\.|Foreign\./);
  assert.equal(words.find((word) => word.id === 'select').audience, 'authoring');
  assert.match(wordFacts(words.find((word) => word.id === 'select'), words, String), /Explicit dictionary: eng\/storage, advanced word authoring/);
  assert.throws(() => wordIndex(words.filter((word) => word.kind !== 'authoring-word'), 'eng:storage', registries), /Unknown word-index library/);
  assert.throws(() => wordIndex([], 'eng:storage', registries), /Unknown word-index library/);
  assert.throws(() => derive([{ ...entries[0], word: { partOfSpeech: 'verb', audience: 'application' } }], source), /cannot be promoted/);
  assert.throws(() => derive([{ ...entries[3], word: { partOfSpeech: 'verb', audience: 'authoring' } }], source), /cannot be promoted/);
  assert.throws(() => derive([{ ...entries[0], word: undefined }], source), /Missing word coverage/);
});

test('library aliases link to canonical meanings without changing native or global membership', () => {
  const words = derive([
    entry('format', 'Formatting.Format', 'format', { dictionary: { lexicon: 'eng', key: 'format' }, inflections: ['Format', 'Formats'], word: { partOfSpeech: 'verb', aliases: ['Runtime.Format'] } }),
    entry('return', 'Function.Return', 'return', { inflections: ['Return', 'Returns'], word: { partOfSpeech: 'verb', aliases: ['Storage.Routine.Return'] } }),
    entry('zebra', 'Runtime.Zebra', 'zebra', { kind: 'authoring-word', dictionary: { lexicon: 'eng', key: 'runtime' }, word: { partOfSpeech: 'noun', audience: 'authoring' } }),
  ], '# Words\n\n## Format\n\n<!-- word:format -->\n\nFormat a value.\n\n## Return\n\n<!-- word:return -->\n\nReturn from the selected Function.\n\n## Zebra\n\n<!-- word:zebra -->\n\nA runtime meaning.\n');
  const before = structuredClone(words);
  const registries = [
    { language: 'eng', key: 'format', dictionary: 'Formatting' },
    { language: 'eng', key: 'runtime', dictionary: 'Runtime', internal: true },
    { language: 'eng', key: 'storage', dictionary: 'Storage', internal: true },
    { language: 'eng', key: 'runtime-extra', dictionary: 'RuntimeExtra', internal: true },
  ];
  const formatting = wordIndex(words, 'eng:format', registries);
  const runtime = wordIndex(words, 'eng:runtime', registries);
  const storage = wordIndex(words, 'eng:storage', registries);
  assert.match(formatting, /href="\/docs\/test\/#format"[^\n]*Formatting\.Format[^\n]*\(verb, supported\)<\/span>: Format a value/);
  assert.match(runtime, /href="\/docs\/test\/#format"[^\n]*Runtime\.Format[^\n]*\(verb, supported, authoring\)<\/span>: Format a value/);
  assert.match(storage, /href="\/docs\/test\/#return"[^\n]*Storage\.Routine\.Return[^\n]*\(verb, supported, authoring\)<\/span>: Return from the selected Function/);
  assert.ok(runtime.indexOf('Runtime.Format') < runtime.indexOf('Runtime.Zebra'));
  assert.doesNotMatch(runtime, /Formatting\.Format|Storage\.Routine\.Return/);
  assert.throws(() => wordIndex(words, 'eng:runtime-extra', registries), /Unknown word-index library/);
  for (const selection of ['', 'builtins']) {
    const index = wordIndex(words, selection);
    assert.doesNotMatch(index, /Runtime\.Format|Storage\.Routine\.Return/);
    assert.equal([...index.matchAll(/href="\/docs\/test\/#return"/g)].length, 1);
    assert.match(index, /Function\.Return[^\n]*\(verb, supported\)/);
  }
  const native = wordIndex(words, 'builtins');
  assert.doesNotMatch(native, /Formatting\.Format|Runtime\.Zebra/);
  const canonical = words.find((word) => word.id === 'format');
  assert.match(wordFacts(canonical, words, String), /Aliases: <code>Runtime\.Format<\/code>/);
  assert.deepEqual(words, before, 'index rows must not mutate canonical inventory, audience or forms');
});

test('reviewed aliases require distinct nonempty qualified names', () => {
  const source = '# Words\n\n## Meaning\n\n<!-- word:x -->\n\nA meaning.\n';
  for (const aliases of [[], {}, [''], [' '], ['Name'], ['Runtime.'], ['.Name'], ['Runtime..Name'], ['Runtime.Name '], ['Runtime.Name', 'Runtime.Name'], ['Runtime.Name', 'runtime.name'], ['Function.Return']]) {
    assert.throws(() => derive([entry('x', 'Function.Return', 'meaning', { word: { partOfSpeech: 'verb', aliases } })], source), /Invalid word aliases/);
  }
});

test('advanced builtin audience does not change implicit availability or remove builtin navigation', () => {
  const words = derive([entry('retain', 'Retain', 'retain', { word: { partOfSpeech: 'verb', audience: 'authoring' } })], '# Words\n\n## Retain\n\n<!-- word:retain -->\n\nRetain an Arc handle.\n');
  const facts = wordFacts(words[0], words, String);
  assert.match(facts, /Implicit builtin/);
  assert.doesNotMatch(facts, /explicit dictionary import/i);
  assert.match(wordIndex(words, 'builtins'), /Retain.*authoring/s);
});

test('ordinary Function.Return remains discoverable in application vocabulary', async () => {
  const { entries } = JSON.parse(await fs.readFile(new URL('../../docs/_meta/semantics.json', import.meta.url), 'utf8'));
  const record = entries.find((entry) => entry.id === 'builtin:function.return');
  const source = record.doc.split('#')[0];
  const document = parseDocument(await fs.readFile(new URL(`../../${source}`, import.meta.url), 'utf8'));
  const children = document.tree.children;
  const marker = children.findIndex((node) => node.type === 'html' && node.value.trim() === `<!-- word:${record.id} -->`);
  assert.ok(marker >= 0, 'Function.Return retains its canonical source entry');
  let start = marker - 1;
  while (start >= 0 && children[start].type !== 'heading') start--;
  assert.ok(start >= 0, 'Function.Return has an owning heading');
  let end = marker + 1;
  while (end < children.length && !(children[end].type === 'heading' && children[end].depth <= children[start].depth)) end++;
  const section = document.text.slice(children[start].position.start.offset, children[end]?.position.start.offset ?? document.text.length);
  const { pages } = JSON.parse(await fs.readFile(new URL('../content-map.json', import.meta.url), 'utf8'));
  const owner = pages.find((page) => page.source === source);
  assert.ok(owner, 'Function.Return canonical source has a published route');
  const words = deriveWords({ entries: [record], pages: [owner], documents: new Map([[source, parseDocument(section)]]) });
  assert.equal(words[0].link, owner.route + '#' + record.doc.split('#')[1]);
  assert.equal(words[0].audience, 'application');
  assert.match(wordIndex(words), /Application vocabulary[\s\S]*Function\.Return/);
  assert.doesNotMatch(wordIndex(words), /Advanced word authoring/);
});

test('two meanings cannot claim the same rendered canonical anchor', () => {
  const source = '# Words\n\n## First\n\n<a id="same"></a>\n\n<!-- word:x -->\n\nFirst meaning.\n\n## Second\n\n<a id="same"></a>\n\n<!-- word:y -->\n\nSecond meaning.\n';
  assert.throws(() => derive([entry('x', 'First', 'same'), entry('y', 'Second', 'same')], source), /Ambiguous.*anchor/);
});

test('reviewed IO.Lined applies to the selected IO.Send actor, not a same-spelled foreign actor', async () => {
  const { entries } = JSON.parse(await fs.readFile(new URL('../../docs/_meta/semantics.json', import.meta.url), 'utf8'));
  const lined = entries.find((record) => record.name === 'IO.Lined');
  const send = entries.find((record) => record.name === 'IO.Send');
  assert.deepEqual(lined.word.targets, [{ entry: send.id, form: 'Sender' }]);
  assert.deepEqual(send.word.formRoles.Sender, { partOfSpeech: 'noun', role: 'actor' });
  assert.ok(lined.sources.includes('stdlib/eng/io/dictionary.l'));
});

test('noun active forms expose their actual applicative verb role', async () => {
  const { entries } = JSON.parse(await fs.readFile(new URL('../../docs/_meta/semantics.json', import.meta.url), 'utf8'));
  for (const [name, form] of [['Process', 'Processing'], ['Function', 'Functioning']]) {
    const record = entries.find((entry) => entry.kind === 'builtin' && entry.name === name);
    assert.deepEqual(record.word.formRoles[form], { partOfSpeech: 'verb', role: 'applicative (noun active form)' });
  }
});

test('gloss and fact output escape HTML and keep evidence strings inert', () => {
  const words = derive([entry('x', 'Meaning', 'meaning', { evidence: 'touch /tmp/not-a-command; <script>alert(1)</script>' })], '# Words\n\n## Meaning\n\n<!-- word:x -->\n\nKeep `<script>` & bytes.\n');
  assert.match(wordIndex(words), /&lt;script&gt;.*&amp; bytes/);
  const facts = wordFacts(words[0], words, String);
  assert.doesNotMatch(facts, /<script>/);
  assert.match(facts, /touch \/tmp\/not-a-command/);
});
