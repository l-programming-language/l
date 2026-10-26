import assert from 'node:assert/strict';
import fs from 'node:fs/promises';
import os from 'node:os';
import path from 'node:path';
import test from 'node:test';
import { checkCoverage, generateReferences, references, referenceEntries } from '../scripts/references.mjs';
import { sha256 } from '../scripts/lib.mjs';
import { anchors, parseMarkdown } from '../scripts/content.mjs';
import GithubSlugger from 'github-slugger';

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
    'docs/cli/build.md': '# Build\n\n## Command syntax\n\nBuild a source.\n\n## Target OS\n\nRejected.\n',
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
      doc: `docs/cli/build.md#${name === '-os' ? 'target-os' : 'command-syntax'}`,
      sources: ['cmd/l/command.go'], witnesses: ['cmd/l/command_test.go'], evidence: 'static-inspection; existing-witness-not-run',
    })),
  }));
  const semantic = (id, kind, name, source, doc) => ({
    id, kind, name, role: 'implementation', support: 'supported', doc,
    ...(kind === 'stdlib-export' ? { visibility: 'public' } : {}),
    sources: [source],
    witnesses: ['lyger/scanner_test.go'], evidence: 'static-inspection; existing-witness-not-run',
  });
  await write(root, 'docs/_meta/semantics.json', jsonText({ schemaVersion: 3, checkedStages: ['static-source-inspection'], executedStages: [], sourceHashes: { 'lyger/value_kinds.go': sha256(files['lyger/value_kinds.go']), 'stdlib/eng/io/dictionary.l': sha256(files['stdlib/eng/io/dictionary.l']) }, entries: [
    semantic('value:Text', 'value-tag', 'Text', 'lyger/value_kinds.go', 'docs/spec/values.md#text'),
    semantic('stdlib:IO.Print', 'stdlib-export', 'IO.Print', 'stdlib/eng/io/dictionary.l', 'docs/stdlib/io.md#print'),
  ] }));
  await write(root, 'docs/_meta/logemes.json', jsonText({ schemaVersion: 3, checkedStages: ['static-source-inspection'], executedStages: [], sourceHashes: { [segment]: sha256(files[segment]) }, entries: [
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
  assert.equal(referenceEntries(first.inventory).find((entry) => entry.id === 'cli:-os').support, 'rejected');
  assert.equal(first.inventory.entries.find((entry) => entry.id === 'lucid:rejected').expectedRustDiagnostic, 'E0603');
  assert.match(first.files['docs/_generated/grammar.md'], /&lt;subject&gt; &#124; predicate/);
  assert.match(first.files['docs/_generated/cli-help.md'], /source-derived.*, not runtime-captured/);
  assert.ok(first.files['docs/_generated/cli-help.md'].includes('l build [SOURCE]\nOptions: --help\n'));
  assert.match(first.files['docs/_generated/lucid.md'], /Runtime execution is a separate witness\. This inventory reports no validation run/);
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


test('semantic review manifests reject missing, competing, and unreferenced hash records', async (t) => {
  for (const [name, mutate, expected] of [
    ['missing', (v) => { delete v.sourceHashes['lyger/value_kinds.go']; }, /Missing reviewed semantic hash/],
    ['entry override', (v) => { v.entries[0].sourceHashes = {}; }, /Entry-specific hashes are not allowed/],
    ['missing stages', (v) => { delete v.checkedStages; }, /Invalid semantic review manifest/],
    ['invalid stages', (v) => { v.executedStages = [false]; }, /Invalid semantic review manifest/],
    ['stage override', (v) => { v.entries[0].executedStages = ['native']; }, /Entry-specific stages are not allowed/],
    ['unused', (v) => { v.sourceHashes['unused.go'] = '0'.repeat(64); }, /Unused semantic review hash/],
    ['wrong schema', (v) => { v.schemaVersion = 1; }, /Invalid semantic review manifest/],
  ]) await t.test(name, async (subtest) => {
    const { root, pages } = await fixture(subtest);
    const inventory = await references({ root, write: true });
    await editJSON(root, 'docs/_meta/semantics.json', mutate);
    await assert.rejects(checkCoverage(root, pages, inventory), expected);
  });
  await t.test('cross-manifest conflict', async (subtest) => {
    const { root, pages } = await fixture(subtest);
    const inventory = await references({ root, write: true });
    await editJSON(root, 'docs/_meta/logemes.json', (v) => {
      v.entries[0].sources.push('lyger/value_kinds.go');
      v.sourceHashes['lyger/value_kinds.go'] = '0'.repeat(64);
    });
    await assert.rejects(checkCoverage(root, pages, inventory), /Conflicting semantic review hash/);
  });
});


test('snapshot note catalog preserves every reviewed explanation and rejects dangling references', async (t) => {
  const { root, pages } = await fixture(t);
  const inventory = await references({ root, write: true });
  const rule = referenceEntries(inventory).find((entry) => entry.id === 'grammar:statement');
  assert.equal(rule.notes, 'Recognition is separate from supported lowering.');
  assert.equal(new Set(inventory.notes).size, inventory.notes.length);
  const coverage = await checkCoverage(root, pages, inventory);
  for (const entry of referenceEntries(coverage).filter((entry) => entry.checkedStages)) {
    assert.deepEqual(entry.checkedStages, ['static-source-inspection']);
    assert.deepEqual(entry.executedStages, []);
  }
  for (const entry of referenceEntries(inventory)) assert.deepEqual(referenceEntries(coverage).find((item) => item.id === entry.id), entry);
  const broken = structuredClone(inventory);
  broken.entries[0].note = broken.notes.length;
  assert.throws(() => referenceEntries(broken), /Missing reference note/);
  const inline = structuredClone(inventory);
  inline.entries[0].notes = 'unreviewed';
  assert.throws(() => referenceEntries(inline), /Inline snapshot note/);
});


test('compact inventories preserve configured fields, anchors and deduplicated review notes', async (t) => {
  const { root } = await fixture(t);
  for (const [kind, duplicate] of [['grammar', (data) => { data.duplicate = { ...data.statement }; }], ['vocabulary', (data) => { data.verbs.Tell = { ...data.verbs.Show, description: 'Preserve <literal> | punctuation & bytes.' }; }]]) {
    const source = `lexicon/eng/${kind}.json`;
    await editJSON(root, source, duplicate);
    const hash = sha256(await fs.readFile(path.join(root, source)));
    await editJSON(root, `docs/_meta/${kind}-review.json`, (review) => {
      review.sourceHashes[source] = hash;
      if (kind === 'grammar') review.entries.duplicate = { ...review.entries.statement };
      else review.entries['verbs.Tell'] = { ...review.entries['verbs.Show'] };
    });
  }
  const { files } = await generateReferences(root);
  const text = (node) => node.type === 'html' ? '' : node.value ?? (node.children ?? []).map(text).join('');
  for (const kind of ['grammar', 'vocabulary']) {
    const source = JSON.parse(await fs.readFile(path.join(root, `lexicon/eng/${kind}.json`), 'utf8'));
    const review = JSON.parse(await fs.readFile(path.join(root, `docs/_meta/${kind}-review.json`), 'utf8'));
    const markdown = files[`docs/_generated/${kind}.md`];
    const ids = anchors(markdown);
    const tables = new Map();
    let section; let name;
    for (const node of parseMarkdown(markdown).children) {
      if (node.type === 'heading' && node.depth === 2) { section = text(node); name = kind === 'grammar' ? section : null; }
      if (node.type === 'heading' && node.depth === 3) name = text(node);
      if (node.type === 'table') {
        const [header, ...rows] = node.children;
        if (kind === 'vocabulary' && text(header.children[0]) === 'Word') {
          for (const row of rows) {
            const name = text(row.children[0]);
            tables.set(name, { ...tables.get(name), ...Object.fromEntries(header.children.map((cell, index) => [text(cell), text(row.children[index])])) });
          }
        } else if (kind === 'grammar' && name && !tables.has(name)) {
          tables.set(name, Object.fromEntries(header.children.map((cell, index) => [text(cell), text(rows[0].children[index])])));
        }
      }
    }
    const entries = kind === 'grammar' ? Object.entries(source).map(([name, value]) => ['', name, value]) : Object.entries(source).flatMap(([section, words]) => Object.entries(words).map(([name, value]) => [section, name, value]));
    for (const [section, name, configured] of entries) {
      assert.ok(ids.has(new GithubSlugger().slug(name)));
      assert.ok(ids.has(kind === 'grammar' ? `grammar-${name.replaceAll('.', '-')}` : `vocabulary-${section}-${name}`));
      const row = tables.get(name);
      for (const [field, value] of Object.entries(configured)) {
        if (kind === 'grammar' && field === 'cases') continue;
        const key = kind === 'grammar' ? field[0].toUpperCase() + field.slice(1) : field;
        assert.equal(row[key], typeof value === 'string' ? value : JSON.stringify(value));
      }
      const detail = review.entries[kind === 'grammar' ? name : `${section}.${name}`];
      assert.equal(row.Support, detail.support);
      assert.equal(row['Semantic context'], detail.topic);
      assert.match(row.Notes, /^Note [0-9]+$/);
    }
    for (const note of new Set(Object.values(review.entries).map((entry) => entry.notes))) {
      assert.equal(markdown.split(note).length - 1, 1, 'reviewed explanation must appear once, without rewriting');
    }
  }
});


test('shared snapshot evidence preserves attribution and rejects overrides', async (t) => {
  const { root } = await fixture(t);
  const { inventory } = await generateReferences(root);
  const expanded = referenceEntries(inventory);
  assert.ok(inventory.evidenceSets.length < inventory.entries.length);
  assert.ok(expanded.every((entry) => Array.isArray(entry.sources) && Array.isArray(entry.witnesses) && entry.evidence));
  const missing = structuredClone(inventory);
  missing.entries[0].evidenceSet = missing.evidenceSets.length;
  assert.throws(() => referenceEntries(missing), /Missing or invalid reference evidence/);
  const override = structuredClone(inventory);
  override.entries[0].sources = [];
  assert.throws(() => referenceEntries(override), /Inline snapshot evidence/);
  const invalid = structuredClone(inventory);
  invalid.evidenceSets[0].id = 'override';
  assert.throws(() => referenceEntries(invalid), /Missing or invalid reference evidence/);
});


test('snapshot classifications preserve review fields and reject invalid references or overrides', async (t) => {
  const { root, pages } = await fixture(t);
  const inventory = await references({ root, write: true });
  const coverage = await checkCoverage(root, pages, inventory);
  assert.equal(inventory.schemaVersion, 3);
  assert.ok(inventory.classifications.length < inventory.entries.length);
  assert.equal(new Set(inventory.classifications.map((value) => JSON.stringify(value))).size, inventory.classifications.length);
  const expanded = referenceEntries(inventory);
  for (const entry of expanded) {
    assert.deepEqual(referenceEntries(coverage).find((value) => value.id === entry.id), entry);
    assert.equal(typeof entry.kind, 'string');
    assert.ok(['implementation', 'normative'].includes(entry.role));
    assert.ok(['supported', 'partial', 'rejected'].includes(entry.support));
  }
  const category = (id) => { const { kind, role, support } = expanded.find((entry) => entry.id === id); return { kind, role, support }; };
  assert.deepEqual(category('grammar:statement'), { kind: 'grammar', role: 'implementation', support: 'partial' });
  assert.deepEqual(category('cli:-os'), { kind: 'cli', role: 'normative', support: 'rejected' });
  for (const index of [-1, 0.5, null, '0', inventory.classifications.length]) {
    const invalid = structuredClone(inventory);
    invalid.entries[0].classification = index;
    assert.throws(() => referenceEntries(invalid), /Missing or invalid reference classification/);
  }
  for (const key of ['kind', 'role', 'support']) {
    const invalid = structuredClone(inventory);
    invalid.entries[0][key] = 'override';
    assert.throws(() => referenceEntries(invalid), /Inline snapshot classification/);
    delete invalid.entries[0][key];
    invalid.classifications[invalid.entries[0].classification][key] = '';
    assert.throws(() => referenceEntries(invalid), /Missing or invalid reference classification/);
  }
  for (const value of [null, [], { kind: 'grammar', role: 'implementation' }, { kind: 'grammar', role: 'implementation', support: 'partial', id: 'override' }]) {
    const invalid = structuredClone(inventory);
    invalid.classifications[invalid.entries[0].classification] = value;
    assert.throws(() => referenceEntries(invalid), /Missing or invalid reference classification/);
  }
  const wrongSchema = structuredClone(inventory);
  wrongSchema.schemaVersion = 2;
  assert.throws(() => referenceEntries(wrongSchema), /Invalid reference snapshot/);
});
