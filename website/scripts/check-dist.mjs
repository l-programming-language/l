import fs from 'node:fs/promises';
import path from 'node:path';
import { parse } from 'parse5';
import { gzipSync } from 'node:zlib';
import { website, json, walk, exists, contained, atomicWrite, stable, isMain } from './lib.mjs';

function elements(node, found = []) { if (node.tagName) found.push(node); for (const child of node.childNodes ?? []) elements(child, found); return found; }
const attrs = (node) => Object.fromEntries((node.attrs ?? []).map((a) => [a.name, a.value]));
export async function checkDist(site = website) {
  const dist = path.join(site, 'dist'); const files = await walk(dist);
  const map = await json(path.join(site, 'content-map.json'));
  const expectedIdentity = await json(path.join(site, '.generated/version.json'));
  for (const route of ['/', ...map.pages.map((p) => p.route)]) if (!files.includes(route.slice(1) + 'index.html')) throw new Error(`Missing real static route: ${route}`);
  if (!files.includes('404.html') || !files.some((p) => p.startsWith('pagefind/') && p.endsWith('.js'))) throw new Error('Missing static 404 or built Pagefind index.');
  if (files.some((p) => /(?:^|\/)(?:_meta|reports|ci-templates|node_modules)(?:\/|$)/.test(p))) throw new Error('Private maintenance content leaked into dist.');
  const pages = new Map();
  for (const file of files.filter((p) => p.endsWith('.html'))) {
    const html = await fs.readFile(path.join(dist, file), 'utf8');
    if (html.includes('/Users/') || html.includes('TODO') || html.includes('lorem ipsum')) throw new Error(`Unresolved/private content in ${file}`);
    const nodes = elements(parse(html)); const ids = new Set();
    for (const node of nodes) { const a = attrs(node); if (a.id) { if (ids.has(a.id)) throw new Error(`Duplicate rendered ID ${a.id} in ${file}`); ids.add(a.id); } }
    pages.set(file, { html, nodes, ids });
  }
  const failures = [];
  for (const [file, { nodes }] of pages) {
    const pageURL = new URL('/' + file.replace(/index\.html$/, ''), 'http://localhost');
    const title = nodes.find((n) => n.tagName === 'title');
    if (!title?.childNodes?.some((n) => n.value?.trim())) failures.push(`${file}: empty page title`);
    const metas = nodes.filter((n) => n.tagName === 'meta').map(attrs);
    if (!metas.some((a) => a.name === 'description' && a.content)) failures.push(`${file}: missing description`);
    if (expectedIdentity.channel !== 'production' && !metas.some((a) => a.name === 'robots' && a.content.includes('noindex'))) failures.push(`${file}: local build is indexable`);
    for (const node of nodes) {
      const a = attrs(node);
      if (node.tagName === 'img' && !Object.hasOwn(a, 'alt')) failures.push(`${file}: image without alt`);
      for (const name of node.tagName === 'a' ? ['href'] : ['src', ...(node.tagName === 'link' ? ['href'] : [])]) {
        const value = a[name]; if (!value || /^(mailto:|data:|javascript:)/.test(value)) continue;
        if (/^https:\/\/ldocs\.org\//.test(value) && node.tagName === 'a') failures.push(`${file}: production-domain internal navigation ${value}`);
        const url = new URL(value, pageURL); if (url.origin !== pageURL.origin) continue;
        const decoded = decodeURIComponent(url.pathname);
        let destination = decoded.endsWith('/') ? decoded.slice(1) + 'index.html' : decoded.slice(1);
        if (!files.includes(destination) && !path.posix.extname(destination)) destination += '/index.html';
        try { contained(dist, destination); } catch { failures.push(`${file}: escaped output link ${value}`); continue; }
        if (!files.includes(destination)) { failures.push(`${file}: missing local file ${value}`); continue; }
        if (url.hash && pages.has(destination) && !pages.get(destination).ids.has(decodeURIComponent(url.hash.slice(1)))) failures.push(`${file}: missing rendered fragment ${value}`);
      }
    }
  }
  if (failures.length) throw new Error(failures.slice(0, 60).join('\n'));
  const version = await json(path.join(dist, 'version.json'));
  if (JSON.stringify(version) !== JSON.stringify(expectedIdentity) || !/^sha256:[a-f0-9]{64}$/.test(version.sourceFingerprint)) throw new Error('Invalid provenance output.');
  const assets = [];
  for (const file of files.filter((p) => /\.(?:js|css|svg|png|ico)$/.test(p))) { const bytes = await fs.readFile(path.join(dist, file)); assets.push({ path: file, bytes: bytes.length, gzipBytes: gzipSync(bytes).length }); }
  assets.sort((a, b) => b.bytes - a.bytes);
  const report = { scope: 'static website output only', routes: map.pages.length + 1, htmlFiles: pages.size, assets, totalAssetBytes: assets.reduce((n, a) => n + a.bytes, 0), javascriptBytes: assets.filter((a) => a.path.endsWith('.js')).reduce((n, a) => n + a.bytes, 0) };
  await atomicWrite(path.join(site, 'reports/built-output.json'), stable(report));
  return report;
}
if (isMain(import.meta.url)) checkDist().then((r) => console.log(`Static output verified: ${r.routes} routes, ${r.htmlFiles} HTML files; ${r.javascriptBytes} JS bytes including optional search assets.`)).catch((e) => { console.error(e.stack); process.exitCode = 1; });
