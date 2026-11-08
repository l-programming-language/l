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
  assert.match(wordFacts(send, words, String), /Sender<\/code> \(noun; actor\)/);
  assert.match(wordFacts(lined, words, String), /href="\/docs\/test\/#send".*IO.Send · Sender/);
  assert.doesNotMatch(wordFacts(send, words, String), /Sended/);
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

test('advanced builtin audience does not change implicit availability or remove builtin navigation', () => {
  const words = derive([entry('retain', 'Retain', 'retain', { word: { partOfSpeech: 'verb', audience: 'authoring' } })], '# Words\n\n## Retain\n\n<!-- word:retain -->\n\nRetain an Arc handle.\n');
  const facts = wordFacts(words[0], words, String);
  assert.match(facts, /Implicit builtin/);
  assert.doesNotMatch(facts, /explicit dictionary import/i);
  assert.match(wordIndex(words, 'builtins'), /Retain.*authoring/s);
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
