import fs from 'node:fs/promises';
import path from 'node:path';
import { prepare } from './prepare.mjs';
import { repository, website, json, sha256, safeFile, walk } from './lib.mjs';

try {
  const result = await prepare();
  const examples = (await json(path.join(repository, 'docs/_meta/examples.json'))).examples;
  const ids = new Set();
  for (const example of examples) {
    if (!example.id || ids.has(example.id)) throw new Error(`Duplicate example ID: ${example.id}`); ids.add(example.id);
    if (!['runnable', 'compile-only', 'invalid', 'fragment'].includes(example.classification)) throw new Error(`Invalid example classification: ${example.id}`);
    if (sha256(await fs.readFile(await safeFile(repository, example.source))) !== example.sha256) throw new Error(`Stale example: ${example.id}`);
    for (const witness of example.witnesses ?? []) await safeFile(repository, witness);
    if (!example.validation) throw new Error(`Example has no truthful validation status: ${example.id}`);
  }
  const map = await json(path.join(website, 'content-map.json'));
  const authored = await walk(repository, 'docs', (p, item) => !['_meta', '_generated', 'assets'].includes(item.name));
  for (const file of authored.filter((p) => /\.mdx?$/.test(p))) if (!map.pages.some((p) => p.source === file)) throw new Error(`Unmapped authored document: ${file}`);
  console.log(`Content integrity passed: ${result.pages} mapped pages, ${examples.length} examples; canonical inputs were not executed.`);
} catch (error) { console.error(error.stack); process.exitCode = 1; }
