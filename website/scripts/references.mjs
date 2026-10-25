import fs from 'node:fs/promises';
import path from 'node:path';
import { repository, website, json, sha256, stable, safeFile, atomicWrite, walk, exists, isMain } from './lib.mjs';
import { anchors, validateStdlib } from './content.mjs';

const cell = (v) => String(v ?? '').replaceAll('|', '&#124;').replaceAll('\n', ' ');
const code = (v) => `<code>${String(v ?? '').replaceAll('&', '&amp;').replaceAll('<', '&lt;').replaceAll('>', '&gt;').replaceAll('|', '&#124;').replaceAll('\n', ' ')}</code>`;
export const referenceInputs = ['lexicon/eng/grammar.json', 'lexicon/eng/vocabulary.json', 'stdlib/index.json', 'cmd/l/command.go', 'lyger/value_kinds.go', 'logiceme/lucid_generation_test.go', 'logiceme/testdata/lucid/templates.sha256.json'];

export async function generateReferences(root) {
  const files = {};
  const inputs = {};
  for (const name of referenceInputs) inputs[name] = sha256(await fs.readFile(await safeFile(root, name)));
  const grammar = await json(path.join(root, referenceInputs[0]));
  const vocabulary = await json(path.join(root, referenceInputs[1]));
  const stdlib = await json(path.join(root, 'stdlib/index.json'));
  const grammarReview = await json(path.join(root, 'docs/_meta/grammar-review.json'));
  const lexicalReview = await json(path.join(root, 'docs/_meta/vocabulary-review.json'));
  for (const review of [grammarReview, lexicalReview]) for (const [file, hash] of Object.entries(review.sourceHashes)) if (sha256(await fs.readFile(await safeFile(root, file))) !== hash) throw new Error(`Reviewed grammar/vocabulary dependency changed: ${file}. Review classifications before updating hashes.`);
  const entries = [];
  let g = '# Grammar rule inventory\n\nThis mechanical reference preserves the configured English recognizers. It does not claim all recognized forms can be lowered or executed. Read [grammar notation](../spec/grammar.md) before interpreting these tables: ` | ` is sequencing, not an alternative operator. Support notes are separately reviewed.\n\n';
  for (const [name, rule] of Object.entries(grammar)) {
    const review = grammarReview.entries[name];
    if (!review) throw new Error(`New unclassified grammar rule: ${name}`);
    const id = `grammar-${name.replaceAll('.', '-')}`;
    g += `## ${name}\n\n<a id="${id}"></a>\n\n${review.notes}\n\nSemantic context: [${review.topic}](${path.posix.relative('docs/_generated', review.doc.split('#')[0])}${review.doc.includes('#') ? '#' + review.doc.split('#')[1] : ''}).\n\n| Property | Configured value |\n| --- | --- |\n| Category | ${code(rule.category)} |\n| Class | ${code(rule.class)} |\n| Structure | ${code(rule.structure)} |\n| Support | ${review.support} |\n\n`;
    const cases = Object.entries(rule.cases ?? {});
    if (cases.length) {
      g += '| Named case | Sequence | Semantic context |\n| --- | --- | --- |\n';
      for (const [caseName, sequence] of cases) {
        const detail = { ...review, ...review.cases?.[caseName] };
        const [file, fragment] = detail.doc.split('#');
        g += `| ${code(caseName)} | ${code(sequence)} | [${detail.topic}](${path.posix.relative('docs/_generated', file)}#${fragment}) · ${detail.support} |\n`;
      }
      g += '\n';
    } else g += 'This rule has no named case alternatives; its structure is used directly.\n\n';
    const base = { kind: 'grammar', name, role: 'implementation', support: review.support, doc: `docs/_generated/grammar.md#${id}`, semanticDoc: review.doc, sources: [referenceInputs[0], 'lyger/grammar_terms.go', 'lyger/grammar_match.go'], witnesses: ['lyger/grammar_terms_test.go', 'lyger/grammar_sequence_test.go', 'lyger/grammar_consumption_test.go'], evidence: 'static-inspection; existing-witness-not-run', notes: review.notes };
    entries.push({ id: `grammar:${name}`, ...base });
    for (const [caseName, sequence] of cases) {
      const detail = { ...review, ...review.cases?.[caseName] };
      entries.push({ id: `grammar:${name}:${caseName}`, ...base, kind: 'grammar-case', name: `${name}.${caseName}`, sequence, semanticDoc: detail.doc, support: detail.support, notes: detail.notes });
    }
    for (const caseName of Object.keys(review.cases ?? {})) if (!Object.hasOwn(rule.cases ?? {}, caseName)) throw new Error(`Stale reviewed grammar case: ${name}.${caseName}`);
  }
  for (const name of Object.keys(grammarReview.entries)) if (!grammar[name]) throw new Error(`Stale reviewed grammar rule: ${name}`);
  files['docs/_generated/grammar.md'] = g.trimEnd() + '\n';
  let v = '# Vocabulary inventory\n\nThe configured lexicon supplies roots, inflection transformations, grammatical categories and handler descriptions. These tables are exact configuration data, not a claim that every handler has native instruction support. See [grammar and vocabulary](../spec/grammar.md) and the linked semantic subjects. Source-defined words, including builtins and dictionary exports, are documented separately.\n\n';
  for (const [section, words] of Object.entries(vocabulary)) {
    v += `## ${section}\n\n`;
    for (const [name, word] of Object.entries(words)) {
      const key = `${section}.${name}`; const review = lexicalReview.entries[key];
      if (!review) throw new Error(`New unclassified vocabulary entry: ${key}`);
      const id = `vocabulary-${section}-${name}`;
      v += `### ${name}\n\n<a id="${id}"></a>\n\n${review.notes}\n\nSemantic context: [${review.topic}](${path.posix.relative('docs/_generated', review.doc.split('#')[0])}${review.doc.includes('#') ? '#' + review.doc.split('#')[1] : ''}).\n\n| Field | Value |\n| --- | --- |\n`;
      for (const [field, value] of Object.entries(word)) v += `| ${cell(field)} | ${code(typeof value === 'string' ? value : JSON.stringify(value))} |\n`;
      v += '\n';
      entries.push({ id: `vocabulary:${key}`, kind: 'vocabulary', name, role: 'implementation', support: review.support, doc: `docs/_generated/vocabulary.md#${id}`, semanticDoc: review.doc, sources: [referenceInputs[1]], witnesses: ['lyger/scanner_test.go'], evidence: 'static-inspection; existing-witness-not-run', notes: review.notes });
    }
  }
  for (const key of Object.keys(lexicalReview.entries)) { const dot = key.indexOf('.'); const section = key.slice(0, dot); const name = key.slice(dot + 1); if (!vocabulary[section]?.[name]) throw new Error(`Stale reviewed vocabulary entry: ${key}`); }
  files['docs/_generated/vocabulary.md'] = v.trimEnd() + '\n';
  const cli = await fs.readFile(path.join(root, 'cmd/l/command.go'), 'utf8');
  // The command parser is unstructured source, so its public surface is a
  // reviewed mapping. Mechanical refresh must never approve a changed parser.
  const cliReview = await json(path.join(root, 'docs/_meta/cli-review.json'));
  if (cliReview.schemaVersion !== 1 || !Array.isArray(cliReview.entries) || !cliReview.entries.length || !cliReview.sourceHashes?.['cmd/l/command.go']) throw new Error('CLI review requires schemaVersion 1, a reviewed command.go hash and surface entries.');
  for (const [file, hash] of Object.entries(cliReview.sourceHashes)) if (sha256(await fs.readFile(await safeFile(root, file))) !== hash) throw new Error(`Reviewed CLI dependency changed: ${file}. Review command/option semantics and mappings before updating its hash; snapshot generation cannot approve this change.`);
  for (const entry of cliReview.entries) {
    if (entry.kind !== 'cli' || !entry.name || entry.id !== `cli:${entry.name}` || !entry.sources?.includes('cmd/l/command.go')) throw new Error(`Invalid reviewed CLI entry: ${entry.id}`);
    entries.push(entry);
  }
  const help = cli.match(/const buildUsage = `([^`]*)`/);
  if (!help) throw new Error('CLI help extraction no longer matches its reviewed raw-string constant. Review command.go.');
  files['docs/_generated/cli-help.md'] = `# Source-derived CLI help\n\nExtracted mechanically from the actual \`buildUsage\` constant in [the command parser](../../cmd/l/command.go). This is **source-derived**, not runtime-captured.\n\n\`\`\`text\n${help[1]}\`\`\`\n`;
  // This is extraction of a fixed string table, not inference of Go or L type semantics.
  const kinds = await fs.readFile(path.join(root, 'lyger/value_kinds.go'), 'utf8');
  const namesBlock = kinds.match(/var valueTypeNames = \[valueTypeCount\]string\{([\s\S]*?)\n\}/);
  if (!namesBlock) throw new Error('Value-name string table changed shape; review extraction.');
  const valueNames = [...namesBlock[1].matchAll(/value\w+: "(\w+)"/g)].map((m) => m[1]);
  const segments = await json(path.join(root, 'logiceme/testdata/lucid/templates.sha256.json'));
  for (const name of Object.keys(segments)) inputs[name] = sha256(await fs.readFile(path.join(root, name)));
  const lucidSource = await fs.readFile(path.join(root, 'logiceme/lucid_generation_test.go'), 'utf8');
  const block = lucidSource.match(/var lucidCases = \[\]lucidCase\{([\s\S]*?)\n\}/);
  if (!block) throw new Error('Lucid registered-case table changed shape; review extraction.');
  const cases = [...block[1].matchAll(/\{name: "([^"]+)"([^}]*)\}/g)].map((m) => ({ name: m[1], expectedRustDiagnostic: m[2].match(/diagnostic: "([^"]+)"/)?.[1] ?? null }));
  const malformed = (await walk(root, 'logiceme/testdata/lucid')).filter((x) => x.split('/').some((part) => part.endsWith('-invalid')) && x.endsWith('.aml.txt'));
  for (const file of malformed) inputs[file] = sha256(await fs.readFile(path.join(root, file)));
  let l = '# Lucid case inventory\n\nThese registered backend cases witness exact source generation. The native stage separately compiles both expected and generated bundles; it does not execute them. No stage was run for this website task. [Read the maintained suite contract](../../logiceme/testdata/lucid/README.md).\n\n| Case | Native expectation |\n| --- | --- |\n';
  for (const c of cases) {
    const source = `logiceme/testdata/lucid/${c.name}.aml.txt`;
    const output = `logiceme/testdata/lucid/${c.name}.rs.txt`;
    for (const name of [source, output]) inputs[name] = sha256(await fs.readFile(await safeFile(root, name)));
    l += `| <a id="lucid-${c.name}"></a>[${c.name}](${path.posix.relative('docs/_generated', source)}) | ${c.expectedRustDiagnostic ? `Intentional rejection: ${c.expectedRustDiagnostic}` : 'Compile-only positive'} |\n`;
    entries.push({ id: `lucid:${c.name}`, kind: 'lucid-case', name: c.name, role: 'implementation', support: c.expectedRustDiagnostic ? 'rejected' : 'supported', doc: `docs/_generated/lucid.md#lucid-${c.name}`, sources: [source, output, 'logiceme/lucid_generation_test.go'], witnesses: ['logiceme/lucid_generation_test.go'], evidence: 'existing-generation-and-native-compilation-witness-not-run', expectedRustDiagnostic: c.expectedRustDiagnostic });
  }
  l += '\n## Malformed-input inventory\n\nEvery listed input is expected to be rejected by the bounded backend interface, before a successful source bundle. These paths are discovered mechanically.\n\n';
  for (const file of malformed) {
    const label = file.split('/').slice(-2).join('/'); const anchor = `invalid-${label.replaceAll('/', '-').replaceAll('.', '-')}`;
    l += `- <a id="${anchor}"></a>[${label}](${path.posix.relative('docs/_generated', file)})\n`;
    entries.push({ id: `lucid-invalid:${label}`, kind: 'lucid-rejection', name: label, role: 'implementation', support: 'rejected', doc: `docs/_generated/lucid.md#${anchor}`, sources: [file], witnesses: ['logiceme/lucid_generation_test.go'], evidence: 'existing-backend-rejection-witness-not-run' });
  }
  files['docs/_generated/lucid.md'] = l;
  const inventory = { schemaVersion: 1, inputs, grammarRules: Object.keys(grammar).length, grammarCases: Object.values(grammar).reduce((sum, r) => sum + Object.keys(r.cases ?? {}).length, 0), vocabularyEntries: Object.values(vocabulary).reduce((sum, r) => sum + Object.keys(r).length, 0), valueNames, stdlib, segments, lucidCases: cases, malformedInputs: malformed, entries };
  files['docs/_generated/inventory.json'] = stable(inventory);
  return { files, inventory };
}

export async function semanticEntries(root) {
  const entries = [];
  for (const name of ['semantics', 'logemes']) entries.push(...(await json(path.join(root, `docs/_meta/${name}.json`))).entries);
  return entries;
}

export async function checkCoverage(root, pages, inventory) {
  const semantic = await semanticEntries(root);
  const entries = [...inventory.entries, ...semantic];
  const ids = new Set();
  const checkedSources = new Map();
  const docAnchors = new Map();
  for (const page of pages) docAnchors.set(page.source, anchors(await fs.readFile(await safeFile(root, page.source), 'utf8'), page.source.endsWith('.mdx')));
  for (const entry of entries) {
    if (!entry.id || ids.has(entry.id)) throw new Error(`Duplicate or missing coverage ID: ${entry.id}`); ids.add(entry.id);
    if (!['normative', 'explanatory', 'implementation', 'future'].includes(entry.role)) throw new Error(`Invalid role: ${entry.id}`);
    if (!['supported', 'partial', 'rejected', 'design-only', 'internal', 'unresolved'].includes(entry.support)) throw new Error(`Invalid support: ${entry.id}`);
    if (typeof entry.evidence !== 'string' || !entry.evidence.trim()) throw new Error(`Missing independent evidence classification: ${entry.id}`);
    if (!entry.sources?.length) throw new Error(`Missing source evidence: ${entry.id}`);
    for (const p of [...entry.sources, ...(entry.witnesses ?? [])]) if (!checkedSources.has(p)) checkedSources.set(p, sha256(await fs.readFile(await safeFile(root, p))));
    if (semantic.includes(entry)) {
      for (const p of entry.sources) if (checkedSources.get(p) !== entry.sourceHashes?.[p]) throw new Error(`Reviewed semantic dependency changed: ${entry.id} -> ${p}. Review the meaning before updating its hash.`);
    }
    for (const target of [entry.doc, entry.semanticDoc].filter(Boolean)) {
      const [file, fragment] = target.split('#');
      if (!pages.find((p) => p.source === file)) throw new Error(`Coverage points to unpublished source ${target} (${entry.id})`);
      if (!fragment || !docAnchors.get(file)?.has(fragment)) throw new Error(`Coverage anchor does not exist: ${target} (${entry.id})`);
    }
  }
  for (const name of inventory.valueNames) if (!semantic.some((e) => e.kind === 'value-tag' && e.name.toLowerCase() === name.toLowerCase())) throw new Error(`Unclassified internal value tag: ${name}`);
  for (const file of Object.keys(inventory.segments)) if (!semantic.some((e) => e.kind === 'logeme' && e.sources.includes(file))) throw new Error(`Unclassified backend segment: ${file}`);
  validateStdlib(inventory.stdlib, pages, semantic);
  return { schemaVersion: 1, description: 'Static coverage and existing witnesses; no language execution during website validation.', entries };
}

export async function references({ root = repository, write = false, coverage = true } = {}) {
  const { files, inventory } = await generateReferences(root);
  for (const [name, bytes] of Object.entries(files)) {
    if (write) await atomicWrite(path.join(root, name), bytes);
    else if (!await exists(path.join(root, name)) || await fs.readFile(path.join(root, name), 'utf8') !== bytes) throw new Error(`Stale mechanical snapshot: ${name}. Run npm run references:generate and review the diff.`);
  }
  if (coverage) {
    const map = await json(path.join(root, 'website/content-map.json'));
    const result = await checkCoverage(root, map.pages, inventory);
    const file = path.join(root, 'docs/_meta/coverage.json');
    if (write) await atomicWrite(file, stable(result));
    else if (!await exists(file) || await fs.readFile(file, 'utf8') !== stable(result)) throw new Error('Coverage snapshot is stale. Review mappings then run references:generate.');
  }
  return inventory;
}
if (isMain(import.meta.url)) references({ write: process.argv.includes('--write') }).then((i) => console.log(`References current: ${i.grammarRules} grammar rules, ${i.grammarCases} cases, ${i.vocabularyEntries} vocabulary entries, ${Object.keys(i.segments).length} segments.`)).catch((e) => { console.error(e.message); process.exitCode = 1; });
