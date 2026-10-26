import fs from 'node:fs/promises';
import path from 'node:path';
import { repository, website, json, stable, safeFile, contained, atomicWrite, exists, withLock, isMain } from './lib.mjs';
import { identity } from './provenance.mjs';
import { validateMap, validateStdlib, preparePage } from './content.mjs';
import { references, semanticEntries } from './references.mjs';

export async function publishOwned(site, outputs) {
  const manifest = path.join(site, '.generated/owned.json');
  const old = await exists(manifest) ? await json(manifest) : [];
  for (const name of [...old, ...outputs.keys()]) {
    if (name.split('/').some((part) => part === '..' || part === '.' || !part)) throw new Error(`Refusing noncanonical owned path: ${name}`);
    if (!/^(src\/content\/docs\/(docs|releases)\/|public\/(content-assets\/|version\.json$|robots\.txt$)|\.generated\/(version|sidebar)\.json$)/.test(name)) throw new Error(`Refusing to manage unowned output path: ${name}`);
    const file = contained(site, name);
    let inspect = file;
    while (inspect !== site) {
      const stat = await fs.lstat(inspect).catch((e) => { if (e.code === 'ENOENT') return null; throw e; });
      if (stat?.isSymbolicLink()) throw new Error(`Refusing symlink in generated output: ${name}`);
      inspect = path.dirname(inspect);
    }
    if (outputs.has(name) && await exists(file) && !old.includes(name)) throw new Error(`Refusing to replace authored/unowned file: ${name}`);
  }
  for (const [name, bytes] of outputs) await atomicWrite(contained(site, name), bytes);
  for (const name of old) if (!outputs.has(name)) await fs.rm(contained(site, name), { force: true });
  await atomicWrite(manifest, stable([...outputs.keys()].sort()));
}

export async function prepare({ root = repository, site = website, verifyReferences = true } = {}) {
  return withLock(path.join(site, '.generated'), async () => {
    const map = validateMap(await json(path.join(site, 'content-map.json')));
    if (verifyReferences) await references({ root });
    const dictionaries = verifyReferences ? validateStdlib(await json(path.join(root, 'stdlib/index.json')), map.pages, await semanticEntries(root)) : [];
    const version = await identity(root, site, { production: process.env.DOCS_CHANNEL === 'production' });
    const examples = (await json(path.join(root, 'docs/_meta/examples.json'))).examples;
    const sourceTexts = new Map();
    for (const p of map.pages) sourceTexts.set(p.source, await fs.readFile(await safeFile(root, p.source), 'utf8'));
    const assets = new Map(); const outputs = new Map();
    const problems = [];
    for (const p of map.pages) {
      try {
        const content = await preparePage({ root, page: p, pages: map.pages, identity: version, examples, sourceTexts, assets, dictionaries });
        const ext = p.source.endsWith('.mdx') ? 'mdx' : 'md';
        outputs.set(`src/content/docs${p.route}index.${ext}`, content);
      } catch (e) { problems.push(e.message); }
    }
    if (problems.length) throw new Error(`Canonical content errors:\n${problems.map((p) => `- ${p}`).join('\n')}`);
    for (const [name, bytes] of assets) outputs.set(`public/content-assets/${name}`, bytes);
    const groups = [...new Set(map.pages.map((p) => p.section))];
    const sidebar = groups.map((label) => ({ label, collapsed: !['Start here', 'Guide'].includes(label), items: map.pages.filter((p) => p.section === label).sort((a, b) => a.order - b.order || a.route.localeCompare(b.route)).map((p) => ({ label: p.title, link: p.route })) }));
    outputs.set('.generated/sidebar.json', stable(sidebar));
    outputs.set('.generated/version.json', stable(version));
    outputs.set('public/version.json', stable(version));
    outputs.set('public/robots.txt', version.channel === 'production' ? 'User-agent: *\nAllow: /\nSitemap: https://ldocs.org/sitemap-index.xml\n' : 'User-agent: *\nDisallow: /\n');
    await publishOwned(site, outputs);
    return { pages: map.pages.length, assets: assets.size, version };
  });
}
if (isMain(import.meta.url)) prepare().then((r) => console.log(`Prepared ${r.pages} canonical pages (${r.version.channel}; ${r.version.dirty === null ? 'Git unavailable' : r.version.dirty ? 'dirty checkout' : 'clean checkout'}).`)).catch((e) => { console.error(e.stack); process.exitCode = 1; });
