import assert from 'node:assert/strict';
import fs from 'node:fs/promises';
import os from 'node:os';
import path from 'node:path';
import test from 'node:test';
import { checkCoverage, generateReferences, references } from '../scripts/references.mjs';
import { sha256 } from '../scripts/lib.mjs';

const jsonText = (value) => JSON.stringify(value, null, 2) + '\n';
const segment = 'builtins/segments/rust/a.def';
const page = (source, route) => ({ source, route, title: source, description: 'Synthetic reference witness.', section: 'Reference', order: 1, role: 'implementation', generated: source.startsWith('docs/_generated/') });

async function write(root, relative, content) {
  const target = path.join(root, relative);
  await fs.mkdir(path.dirname(target), { recursive: true });
  await fs.writeFile(target, content);
}

async function editJSON(root, relative, change) {
  const value = JSON.parse(await fs.readFile(path.join(root, relative), 'utf8'));
  change(value);
  await write(root, relative, jsonText(value));
}

async function fixture(t) {
  const temporary = await fs.mkdtemp(path.join(os.tmpdir(), 'l reference review '));
  const root = await fs.realpath(temporary);
  t.after(() => fs.rm(root, { recursive: true, force: true }));
  const grammar = { statement: { category: 'statement', class: 'action', structure: '<subject> | predicate', cases: { direct: 'verb | object', passive: 'subject | verb' } } };
  const vocabulary = { verbs: { Show: { forms: ['Show', 'Shows'], description: 'Write a value.' } } };
  const index = { format: 1, languages: { eng: { io: { entry: 'stdlib/eng/io/dictionary.l', dictionary: 'IO' } } } };
  const files = {
    'lexicon/eng/grammar.json': jsonText(grammar),
    'lexicon/eng/vocabulary.json': jsonText(vocabulary),
    'stdlib/index.json': jsonText(index),
    'stdlib/eng/io/dictionary.l': 'Create "IO" as a Dictionary.\n',
    'cmd/l/command.go': 'package main\nconst buildUsage = `l build [SOURCE]\nOptions: --help\n`\n',
    'cmd/l/command_test.go': '// Existing CLI witness; fixture is never compiled.\n',
    'lyger/value_kinds.go': 'package Lyger\nvar valueTypeNames = [valueTypeCount]string{\n valueText: "Text",\n}\n',
    'lyger/grammar_terms.go': '// source notation implementation\n',
    'lyger/grammar_match.go': '// matching implementation\n',
    'lyger/grammar_terms_test.go': '// grammar witness\n',
    'lyger/grammar_sequence_test.go': '// sequence witness\n',
    'lyger/grammar_consumption_test.go': '// consumption witness\n',
    'lyger/scanner_test.go': '// vocabulary witness\n',
    'logiceme/lucid_generation_test.go': 'package Logiceme\nvar lucidCases = []lucidCase{\n {name: "sample"},\n {name: "rejected", diagnostic: "E0603"},\n}\n',
    'logiceme/testdata/lucid/templates.sha256.json': jsonText({ [segment]: sha256('native template\n') }),
    [segment]: 'native template\n',
    'logiceme/testdata/lucid/sample.aml.txt': '<PROGRAM>sample</PROGRAM>\n',
    'logiceme/testdata/lucid/sample.rs.txt': 'fn main() {}\n',
    'logiceme/testdata/lucid/rejected.aml.txt': '<PROGRAM>rejected</PROGRAM>\n',
    'logiceme/testdata/lucid/rejected.rs.txt': '// intentional native rejection\n',
    'logiceme/testdata/lucid/forms-invalid/bad.aml.txt': '<PROGRAM>unclosed\n',
    'logiceme/testdata/lucid/README.md': '# Synthetic Lucid witness\n',
    'docs/spec/grammar.md': '# Grammar\n\n## Statement structure\n\nConfigured grammar and its limits.\n',
    'docs/spec/values.md': '# Values\n\n## Text\n\nA source-backed value category.\n',
    'docs/cli/build.md': '# Build\n\n## Command syntax\n\nBuild a source.\n\n## Legacy os option\n\nRejected.\n',
    'docs/reference/logemes/index.md': '# Logemes\n\n## Segment A\n\nBounded backend behavior.\n',
    'docs/stdlib/io.md': '# IO\n\n## Print\n\nPublic output action.\n',
  };
  for (const [name, bytes] of Object.entries(files)) await write(root, name, bytes);
  const review = (source, entries) => ({ schemaVersion: 1, sourceHashes: { [source]: sha256(files[source]) }, entries });
  await write(root, 'docs/_meta/grammar-review.json', jsonText(review('lexicon/eng/grammar.json', {
    statement: { support: 'partial', topic: 'Statement structure', doc: 'docs/spec/grammar.md#statement-structure', notes: 'Recognition is separate from supported lowering.' },
  })));
  await write(root, 'docs/_meta/vocabulary-review.json', jsonText(review('lexicon/eng/vocabulary.json', {
    'verbs.Show': { support: 'supported', topic: 'Statement structure', doc: 'docs/spec/grammar.md#statement-structure', notes: 'Reviewed source action.' },
  })));
  await write(root, 'docs/_meta/cli-review.json', jsonText({
    schemaVersion: 1, sourceHashes: { 'cmd/l/command.go': sha256(files['cmd/l/command.go']) },
    entries: ['build', '-os'].map((name) => ({
      id: `cli:${name}`, kind: 'cli', name, role: 'normative', support: name === '-os' ? 'rejected' : 'supported',
      doc: `docs/cli/build.md#${name === '-os' ? 'legacy-os-option' : 'command-syntax'}`,
      sources: ['cmd/l/command.go'], witnesses: ['cmd/l/command_test.go'], evidence: 'static-inspection; existing-witness-not-run',
    })),
  }));
  const semantic = (id, kind, name, source, doc) => ({
    id, kind, name, role: 'implementation', support: 'supported', doc,
    ...(kind === 'stdlib-export' ? { visibility: 'public' } : {}),
    sources: [source], sourceHashes: { [source]: sha256(files[source]) },
    witnesses: ['lyger/scanner_test.go'], evidence: 'static-inspection; existing-witness-not-run',
  });
  await write(root, 'docs/_meta/semantics.json', jsonText({ schemaVersion: 1, entries: [
    semantic('value:Text', 'value-tag', 'Text', 'lyger/value_kinds.go', 'docs/spec/values.md#text'),
    semantic('stdlib:IO.Print', 'stdlib-export', 'IO.Print', 'stdlib/eng/io/dictionary.l', 'docs/stdlib/io.md#print'),
  ] }));
  await write(root, 'docs/_meta/logemes.json', jsonText({ schemaVersion: 1, entries: [
    semantic('logeme:A', 'logeme', 'A', segment, 'docs/reference/logemes/index.md#segment-a'),
  ] }));
  const pages = [
    page('docs/_generated/grammar.md', '/docs/reference/grammar/'),
    page('docs/_generated/vocabulary.md', '/docs/reference/vocabulary/'),
    page('docs/_generated/cli-help.md', '/docs/cli/help/'),
    page('docs/_generated/lucid.md', '/docs/reference/lucid/'),
    page('docs/spec/grammar.md', '/docs/spec/grammar/'),
    page('docs/spec/values.md', '/docs/spec/values/'),
    page('docs/cli/build.md', '/docs/cli/build/'),
    page('docs/reference/logemes/index.md', '/docs/reference/logemes/'),
    page('docs/stdlib/io.md', '/docs/stdlib/eng/io/'),
  ];
  await write(root, 'website/content-map.json', jsonText({ schemaVersion: 1, pages }));
  return { root, pages };
}

test('mechanical reference inventory is deterministic and preserves evidence-stage distinctions', async (t) => {
  const { root } = await fixture(t);
  const first = await generateReferences(root);
  const second = await generateReferences(root);
  assert.deepEqual(first, second);
  assert.equal(first.inventory.grammarRules, 1);
  assert.equal(first.inventory.grammarCases, 2);
  assert.equal(first.inventory.vocabularyEntries, 1);
  assert.deepEqual(first.inventory.valueNames, ['Text']);
  assert.equal(first.inventory.lucidCases.length, 2);
  assert.deepEqual(first.inventory.malformedInputs, ['logiceme/testdata/lucid/forms-invalid/bad.aml.txt']);
  assert.equal(first.inventory.entries.find((entry) => entry.id === 'cli:-os').support, 'rejected');
  assert.equal(first.inventory.entries.find((entry) => entry.id === 'lucid:rejected').expectedRustDiagnostic, 'E0603');
  assert.match(first.files['docs/_generated/grammar.md'], /&lt;subject&gt; &#124; predicate/);
  assert.match(first.files['docs/_generated/cli-help.md'], /source-derived.*, not runtime-captured/);
  assert.ok(first.files['docs/_generated/cli-help.md'].includes('l build [SOURCE]\nOptions: --help\n'));
  assert.match(first.files['docs/_generated/lucid.md'], /does not execute them/);
  assert.match(first.files['docs/_generated/lucid.md'], /Intentional rejection: E0603/);
  assert.ok(Object.values(first.inventory.inputs).every((hash) => /^[a-f0-9]{64}$/.test(hash)));
});

test('fresh snapshots and complete meaningful coverage pass from a no-Git fixture', async (t) => {
  const { root } = await fixture(t);
  const generated = await references({ root, write: true });
  const checked = await references({ root });
  assert.deepEqual(checked, generated);
  const coverage = JSON.parse(await fs.readFile(path.join(root, 'docs/_meta/coverage.json'), 'utf8'));
  assert.ok(coverage.entries.some((entry) => entry.id === 'value:Text'));
  assert.ok(coverage.entries.some((entry) => entry.id === 'grammar:statement:passive'));
  assert.ok(coverage.entries.some((entry) => entry.id === 'lucid-invalid:forms-invalid/bad.aml.txt'));
  assert.match(coverage.description, /no language execution/);
});

test('changing a reviewed grammar or vocabulary dependency fails before snapshot refresh', async (t) => {
  for (const source of ['lexicon/eng/grammar.json', 'lexicon/eng/vocabulary.json']) {
    await t.test(source, async (subtest) => {
      const { root } = await fixture(subtest);
      await fs.appendFile(path.join(root, source), '\n');
      await assert.rejects(generateReferences(root), /Reviewed grammar\/vocabulary dependency changed/);
    });
  }
});

test('new grammar and vocabulary inventory needs reviewed semantic classification even after hash review', async (t) => {
  await t.test('grammar rule', async (subtest) => {
    const { root } = await fixture(subtest);
    await editJSON(root, 'lexicon/eng/grammar.json', (value) => { value.newRule = { structure: 'new form' }; });
    const hash = sha256(await fs.readFile(path.join(root, 'lexicon/eng/grammar.json')));
    await editJSON(root, 'docs/_meta/grammar-review.json', (value) => { value.sourceHashes['lexicon/eng/grammar.json'] = hash; });
    await assert.rejects(generateReferences(root), /New unclassified grammar rule: newRule/);
  });
  await t.test('vocabulary entry', async (subtest) => {
    const { root } = await fixture(subtest);
    await editJSON(root, 'lexicon/eng/vocabulary.json', (value) => { value.verbs.NewWord = { forms: ['NewWord'] }; });
    const hash = sha256(await fs.readFile(path.join(root, 'lexicon/eng/vocabulary.json')));
    await editJSON(root, 'docs/_meta/vocabulary-review.json', (value) => { value.sourceHashes['lexicon/eng/vocabulary.json'] = hash; });
    await assert.rejects(generateReferences(root), /New unclassified vocabulary entry: verbs.NewWord/);
  });
});

test('removed source entries leave detectable stale reviewed mappings', async (t) => {
  const { root } = await fixture(t);
  await editJSON(root, 'lexicon/eng/vocabulary.json', (value) => { delete value.verbs.Show; });
  const hash = sha256(await fs.readFile(path.join(root, 'lexicon/eng/vocabulary.json')));
  await editJSON(root, 'docs/_meta/vocabulary-review.json', (value) => { value.sourceHashes['lexicon/eng/vocabulary.json'] = hash; });
  await assert.rejects(generateReferences(root), /Stale reviewed vocabulary entry: verbs.Show/);
});

test('changed CLI implementation cannot be approved by mechanical snapshot generation', async (t) => {
  const { root } = await fixture(t);
  await references({ root, write: true });
  const previous = await fs.readFile(path.join(root, 'docs/_generated/cli-help.md'), 'utf8');
  const reviewBytes = await fs.readFile(path.join(root, 'docs/_meta/cli-review.json'));
  await fs.appendFile(path.join(root, 'cmd/l/command.go'), 'flags.StringVar(&command.NewOption, "new-option", "", "new public behavior")\n');
  await assert.rejects(generateReferences(root), /Reviewed CLI dependency changed: cmd\/l\/command.go/);
  await assert.rejects(references({ root, write: true }), /snapshot generation cannot approve this change/);
  assert.equal(await fs.readFile(path.join(root, 'docs/_generated/cli-help.md'), 'utf8'), previous);
  assert.deepEqual(await fs.readFile(path.join(root, 'docs/_meta/cli-review.json')), reviewBytes);
});

test('CLI review cannot omit its approval hash or substitute unrelated inventory entries', async (t) => {
  await t.test('missing hash', async (subtest) => {
    const { root } = await fixture(subtest);
    await editJSON(root, 'docs/_meta/cli-review.json', (value) => { delete value.sourceHashes['cmd/l/command.go']; });
    await assert.rejects(generateReferences(root), /reviewed command.go hash/);
  });
  await t.test('wrong kind', async (subtest) => {
    const { root } = await fixture(subtest);
    await editJSON(root, 'docs/_meta/cli-review.json', (value) => { value.entries[0].kind = 'value-tag'; });
    await assert.rejects(generateReferences(root), /Invalid reviewed CLI entry/);
  });
});

test('mechanical snapshots cannot be edited independently of their source', async (t) => {
  const { root } = await fixture(t);
  await references({ root, write: true });
  await fs.appendFile(path.join(root, 'docs/_generated/cli-help.md'), '\nInvented runtime-captured flag.\n');
  await assert.rejects(references({ root }), /Stale mechanical snapshot: docs\/_generated\/cli-help.md/);
});

test('coverage snapshot drift is reported without overwriting it in check mode', async (t) => {
  const { root } = await fixture(t);
  await references({ root, write: true });
  const forged = '{"schemaVersion":1,"entries":[]}\n';
  await write(root, 'docs/_meta/coverage.json', forged);
  await assert.rejects(references({ root }), /Coverage snapshot is stale/);
  assert.equal(await fs.readFile(path.join(root, 'docs/_meta/coverage.json'), 'utf8'), forged);
});

test('semantic source hashes require review of changed library and segment meanings', async (t) => {
  for (const source of ['stdlib/eng/io/dictionary.l', segment]) {
    await t.test(source, async (subtest) => {
      const { root, pages } = await fixture(subtest);
      const inventory = await references({ root, write: true });
      await fs.appendFile(path.join(root, source), 'changed meaning\n');
      await assert.rejects(checkCoverage(root, pages, inventory), /Reviewed semantic dependency changed/);
    });
  }
});

test('coverage rejects missing witness files and missing or unpublished documentation anchors', async (t) => {
  await t.test('missing witness', async (subtest) => {
    const { root, pages } = await fixture(subtest);
    const inventory = await references({ root, write: true });
    await fs.rm(path.join(root, 'lyger/scanner_test.go'));
    await assert.rejects(checkCoverage(root, pages, inventory), /ENOENT.*scanner_test.go/);
  });
  await t.test('missing anchor', async (subtest) => {
    const { root, pages } = await fixture(subtest);
    const inventory = await references({ root, write: true });
    await editJSON(root, 'docs/_meta/semantics.json', (value) => { value.entries[0].doc = 'docs/spec/values.md#does-not-exist'; });
    await assert.rejects(checkCoverage(root, pages, inventory), /Coverage anchor does not exist/);
  });
  await t.test('unpublished source', async (subtest) => {
    const { root, pages } = await fixture(subtest);
    const inventory = await references({ root, write: true });
    await editJSON(root, 'docs/_meta/semantics.json', (value) => { value.entries[0].doc = 'docs/private-review.md#text'; });
    await assert.rejects(checkCoverage(root, pages, inventory), /Coverage points to unpublished source/);
  });
});

test('new internal value tags and backend segments fail until explicitly classified', async (t) => {
  const { root, pages } = await fixture(t);
  const inventory = await references({ root, write: true });
  await assert.rejects(checkCoverage(root, pages, { ...inventory, valueNames: [...inventory.valueNames, 'NewInternalState'] }), /Unclassified internal value tag: NewInternalState/);
  await assert.rejects(checkCoverage(root, pages, { ...inventory, segments: { ...inventory.segments, 'builtins/segments/rust/new.def': 'unknown' } }), /Unclassified backend segment/);
});

test('coverage rejects duplicate IDs and invented support classifications', async (t) => {
  await t.test('duplicate ID', async (subtest) => {
    const { root, pages } = await fixture(subtest);
    const inventory = await references({ root, write: true });
    await editJSON(root, 'docs/_meta/semantics.json', (value) => { value.entries.push(value.entries[0]); });
    await assert.rejects(checkCoverage(root, pages, inventory), /Duplicate or missing coverage ID/);
  });
  await t.test('invalid support', async (subtest) => {
    const { root, pages } = await fixture(subtest);
    const inventory = await references({ root, write: true });
    await editJSON(root, 'docs/_meta/semantics.json', (value) => { value.entries[0].support = 'probably-fine'; });
    await assert.rejects(checkCoverage(root, pages, inventory), /Invalid support/);
  });
});

test('registered case source additions require both actual AML and expected native witness files', async (t) => {
  const { root } = await fixture(t);
  const current = await fs.readFile(path.join(root, 'logiceme/lucid_generation_test.go'), 'utf8');
  await write(root, 'logiceme/lucid_generation_test.go', current.replace('\n}', '\n {name: "new-case"},\n}'));
  await assert.rejects(generateReferences(root), /ENOENT.*new-case.aml.txt/);
  await write(root, 'logiceme/testdata/lucid/new-case.aml.txt', '<PROGRAM>new</PROGRAM>');
  await assert.rejects(generateReferences(root), /ENOENT.*new-case.rs.txt/);
});

test('coverage requires an evidence classification separate from support and role', async (t) => {
  const { root } = await fixture(t);
  await references({ root, write: true });
  await editJSON(root, 'docs/_meta/semantics.json', (value) => { delete value.entries[0].evidence; });
  await assert.rejects(references({ root, write: true }), /independent evidence/);
});
