import fs from 'node:fs/promises';
import path from 'node:path';
import { unified } from 'unified';
import remarkParse from 'remark-parse';
import remarkGfm from 'remark-gfm';
import remarkMdx from 'remark-mdx';
import { visit } from 'unist-util-visit';
import GithubSlugger from 'github-slugger';
import { parse as parseYAML, stringify as yaml } from 'yaml';
import { safeFile, sha256 } from './lib.mjs';
import { sourceURL } from './provenance.mjs';
import { wordFacts, wordIndex } from './words.mjs';

const html = (value) => String(value).replaceAll('&', '&amp;').replaceAll('<', '&lt;').replaceAll('>', '&gt;').replaceAll('"', '&quot;');
const sourceCode = (code, language, label) => `<pre tabindex="0" role="region" aria-label="${html(label)}"><code class="language-${html(language)}">${html(code)}</code></pre>`;

export function parseMarkdown(text, mdx = false) {
  const parser = unified().use(remarkParse).use(remarkGfm);
  if (mdx) parser.use(remarkMdx);
  return parser.parse(text);
}
export function plain(node) { return node.value ?? (node.children ?? []).map(plain).join(''); }
export function parseDocument(source, mdx = false) {
  const fm = source.match(/^---\r?\n([\s\S]*?)\r?\n---\r?\n/);
  const text = fm ? source.slice(fm[0].length) : source;
  const authored = fm ? parseYAML(fm[1]) ?? {} : {};
  const tree = parseMarkdown(text, mdx);
  const slugger = new GithubSlugger();
  const found = new Set(), headings = new Map(), anchorCounts = new Map();
  const add = (id) => { found.add(id); anchorCounts.set(id, (anchorCounts.get(id) ?? 0) + 1); };
  visit(tree, 'heading', (node) => { const id = slugger.slug(plain(node)); add(id); headings.set(node, id); });
  visit(tree, 'html', (node) => { for (const m of node.value.matchAll(/<(?:a|span)\s+id=["']([^"']+)["'][^>]*>/g)) add(m[1]); });
  return { text, authored, tree, anchors: found, anchorCounts, headings };
}
export function anchors(text, mdx = false) { return parseDocument(text, mdx).anchors; }

function sourceLines(code, start, end, label) {
  // Keep original terminators; a final newline does not create an extra line.
  const lines = code.match(/[^\n]*\n|[^\n]+$/g) ?? [];
  if (end > lines.length) throw new Error(`${label} lines out of bounds: ${start}-${end}, source has ${lines.length} lines.`);
  return lines.slice(start - 1, end).join('');
}

export function validateMap(map) {
  if (map.schemaVersion !== 1 || !Array.isArray(map.pages)) throw new Error('content-map.json requires schemaVersion 1 and a pages array.');
  const routes = new Set(['/']); const sources = new Set();
  for (const p of map.pages) {
    for (const key of ['source', 'route', 'title', 'description', 'section', 'role']) if (typeof p[key] !== 'string' || !p[key].trim()) throw new Error(`Invalid content metadata ${key}: ${p.source}`);
    if (!/^\/(?:docs(?:\/[a-z0-9-]+)*|releases(?:\/[a-z0-9-]+)*)\/$/.test(p.route)) throw new Error(`Invalid public route: ${p.route}`);
    if (routes.has(p.route)) throw new Error(`Duplicate route: ${p.route}`);
    if (sources.has(p.source)) throw new Error(`Duplicate canonical source: ${p.source}`);
    if (!Number.isFinite(p.order) || typeof p.generated !== 'boolean' || (p.sidebar !== undefined && typeof p.sidebar !== 'boolean')) throw new Error(`Invalid navigation metadata: ${p.source}`);
    if (!['normative', 'explanatory', 'implementation', 'future'].includes(p.role)) throw new Error(`Invalid document role: ${p.role}`);
    if (p.source.includes('/_meta/')) throw new Error(`Private metadata cannot be a public page: ${p.source}`);
    routes.add(p.route); sources.add(p.source);
  }
  const pages = new Map(map.pages.map((p) => [p.route, p]));
  for (const p of map.pages) {
    const seen = new Set([p.route]);
    let child = p;
    while (child.parent !== undefined) {
      const parent = pages.get(child.parent);
      if (!parent || parent.sidebar === false || child.sidebar === false || parent.section !== child.section) throw new Error(`Invalid sidebar parent: ${child.route}`);
      if (seen.has(parent.route)) throw new Error(`Cyclic sidebar parent: ${p.route}`);
      seen.add(parent.route); child = parent;
    }
  }
  return map;
}

export function validateStdlib(index, pages, semanticEntries) {
  if (index.format !== 1 || !index.languages || typeof index.languages !== 'object') throw new Error('Unsupported dictionary registry format.');
  const exports = semanticEntries.filter((entry) => entry.kind === 'stdlib-export');
  for (const entry of semanticEntries.filter((entry) => entry.dictionary || ['stdlib-export', 'authoring-word'].includes(entry.kind))) {
    const owner = entry.dictionary;
    if (!owner || !Object.hasOwn(index.languages, owner.lexicon) || !Object.hasOwn(index.languages[owner.lexicon], owner.key)) throw new Error(`Invalid reviewed dictionary owner: ${entry.id ?? entry.name}`);
    if (!Array.isArray(entry.sources) || !entry.sources.includes(index.languages[owner.lexicon][owner.key].entry)) throw new Error(`Missing reviewed dictionary source: ${entry.id ?? entry.name}`);
  }
  const dictionaries = [];
  for (const [language, entries] of Object.entries(index.languages).sort()) {
    for (const [key, entry] of Object.entries(entries).sort()) {
      if (!/^[a-z0-9-]+$/.test(language) || !/^[a-z0-9-]+$/.test(key) || !entry.entry || !entry.dictionary) throw new Error(`Invalid indexed dictionary: ${language}/${key}`);
      const reviewed = exports.filter((x) => x.dictionary.lexicon === language && x.dictionary.key === key);
      if (reviewed.some((x) => !['namespace', 'public', 'private'].includes(x.visibility))) throw new Error(`Missing reviewed export visibility for ${language}/${key}.`);
      const internal = reviewed.find((x) => x.visibility === 'namespace' && x.support === 'internal');
      if (internal) {
        if (reviewed.some((x) => x.support !== 'internal')) throw new Error(`Internal dictionary ${language}/${key} has conflicting public documentation classifications.`);
        const page = pages.find((p) => p.source === internal.doc?.split('#')[0]);
        if (!page) throw new Error(`Missing implementation reference for ${language}/${key}.`);
        dictionaries.push({ language, key, ...entry, route: page.route, source: page.source, description: page.description, internal: true, exports: [] });
        continue;
      }
      const route = `/docs/stdlib/${language}/${key}/`;
      const page = pages.find((p) => p.route === route);
      if (!page) throw new Error(`Unclassified indexed dictionary ${language}/${key}: author its documentation and add ${route} to content-map.json.`);
      if (!reviewed.length) throw new Error(`Missing reviewed export inventory for ${language}/${key}.`);
      if (!reviewed.some((entry) => entry.visibility === 'public')) throw new Error(`Missing reviewed public export inventory for ${language}/${key}.`);
      dictionaries.push({ language, key, ...entry, route, source: page.source, description: page.description, internal: false, exports: reviewed.filter((x) => x.visibility === 'public').map((x) => x.name) });
    }
  }
  return dictionaries;
}

export function stdlibCatalog(dictionaries) {
  const escape = (value) => String(value).replaceAll('&', '&amp;').replaceAll('<', '&lt;').replaceAll('|', '&#124;').replaceAll('\n', ' ');
  return '| Dictionary | Import key (lexicon) | Audience | Purpose |\n| --- | --- | --- | --- |\n' + dictionaries.map((item) => `| [${escape(item.dictionary)}](${item.route}) | ${escape(item.key)} (${escape(item.language)}) | ${item.internal ? 'Implementation' : 'Application'} | ${escape(item.description)} |`).join('\n');
}

export async function preparePage({ root, page, pages, identity, examples, sourceTexts, assets, dictionaries = [], documents = new Map(), words = [] }) {
  const sourcePath = await safeFile(root, page.source);
  const document = documents.get(page.source) ?? parseDocument(sourceTexts.get(page.source) ?? await fs.readFile(sourcePath, 'utf8'), page.source.endsWith('.mdx'));
  const { authored, tree } = document;
  let { text } = document;
  const sourceAnchors = (source) => documents.get(source)?.anchors ?? anchors(sourceTexts.get(source), source.endsWith('.mdx'));
  const changes = []; const routeBySource = new Map(pages.map((p) => [p.source, p]));
  const defs = new Map(); visit(tree, 'definition', (node) => defs.set(node.identifier, node));
  const add = (node, replacement) => changes.push({ start: node.position.start.offset, end: node.position.end.offset, replacement });
  const first = tree.children[0];
  if (first?.type === 'heading' && first.depth === 1) add(first, `<a id="${new GithubSlugger().slug(plain(first))}"></a>`);
  async function resolve(url, image = false) {
    if (/^(https?:|mailto:)/i.test(url)) return { url };
    if (/^[A-Za-z][\w+.-]*:/.test(url) || url.startsWith('//')) throw new Error(`Disallowed URL in ${page.source}: ${url}`);
    const [beforeFragment, ...hashParts] = url.split('#'); const hash = hashParts.join('#');
    const [pathname, ...queryParts] = beforeFragment.split('?'); const query = queryParts.length ? `?${queryParts.join('?')}` : '';
    if (!pathname) {
      if (hash && !document.anchors.has(decodeURIComponent(hash))) throw new Error(`Missing fragment ${page.source}#${hash}`);
      return { url };
    }
    if (pathname.startsWith('/')) {
      if (pathname.startsWith('/assets/l/')) return { url };
      const mapped = pages.find((p) => p.route === pathname);
      if (pathname !== '/' && !mapped) throw new Error(`Unknown site route in ${page.source}: ${url}`);
      if (hash && mapped && !sourceAnchors(mapped.source).has(decodeURIComponent(hash))) throw new Error(`Missing fragment: ${url}`);
      return { url };
    }
    const relative = path.posix.normalize(path.posix.join(path.posix.dirname(page.source), decodeURIComponent(pathname)));
    let file;
    try { file = await safeFile(root, relative); }
    catch (e) {
      // Directory references in maintained repository guides are honest source links.
      const candidate = path.resolve(root, relative);
      if (!candidate.startsWith(root + path.sep)) throw e;
      const real = await fs.realpath(candidate).catch(() => null);
      const realRoot = await fs.realpath(root);
      if (real && real !== realRoot && !real.startsWith(realRoot + path.sep)) throw new Error(`Escaped source directory link in ${page.source}: ${url}`);
      const stat = await fs.stat(candidate).catch(() => null);
      if (!stat?.isDirectory()) throw new Error(`Unresolved local link in ${page.source}: ${url} (${e.message})`);
      const remote = sourceURL(root, relative, identity);
      return remote ? { url: remote.replace('/blob/', '/tree/') + query + (hash ? `#${hash}` : '') } : { source: relative };
    }
    if (image) {
      const bytes = await fs.readFile(file); const ext = path.extname(relative).toLowerCase();
      if (!['.svg', '.png', '.jpg', '.jpeg', '.gif', '.webp', '.avif'].includes(ext)) throw new Error(`Unsupported local image ${relative}`);
      const asset = `${sha256(bytes).slice(0, 24)}${ext}`;
      assets.set(asset, bytes);
      return { url: `/content-assets/${asset}` };
    }
    const mapped = routeBySource.get(relative);
    if (mapped) {
      if (hash && !sourceAnchors(relative).has(decodeURIComponent(hash))) throw new Error(`Missing document fragment ${relative}#${hash} linked by ${page.source}`);
      return { url: mapped.route + query + (hash ? `#${hash}` : '') };
    }
    const remote = sourceURL(root, relative, identity);
    return remote ? { url: remote + query + (hash ? `#${hash}` : '') } : { source: relative };
  }
  const links = []; visit(tree, (node) => { if (['link', 'image', 'definition', 'linkReference', 'imageReference', 'html'].includes(node.type)) links.push(node); });
  let indices = 0;
  for (const node of links) {
    if (node.type === 'html') {
      const word = node.value.trim().match(/^<!--\s*word:([^\s<>]+)\s*-->$/);
      if (word) {
        const selected = words.find((entry) => entry.id === word[1] && entry.source === page.source);
        if (!selected) throw new Error(`Unknown word marker ${word[1]} in ${page.source}`);
        add(node, wordFacts(selected, words, (source) => {
          const url = sourceURL(root, source, identity);
          return url ? `<a href="${html(url)}"><code>${html(source)}</code></a>` : `<code>${html(source)}</code>`;
        }));
        continue;
      }
      const definition = node.value.trim().match(/^<!--\s*word-definition:([^\s<>]+)\s*-->$/);
      if (/<!--\s*word-definition\b/.test(node.value) && (!definition || !tree.children.includes(node))) throw new Error(`Invalid word-definition directive in ${page.source}:${node.position.start.line}`);
      if (definition) {
        const selected = words.find((entry) => entry.id === definition[1] && entry.source === page.source);
        if (!selected?.word.definitions) throw new Error(`Unknown word-definition marker ${definition[1]} in ${page.source}`);
        if (selected.definitionMarker?.position.start.offset !== node.position.start.offset) throw new Error(`Word-definition marker outside its canonical word section: ${definition[1]} in ${page.source}`);
        const excerpts = [];
        for (const reference of selected.word.definitions) {
          const bytes = await fs.readFile(await safeFile(root, reference.source));
          let code;
          try { code = new TextDecoder('utf-8', { fatal: true, ignoreBOM: true }).decode(bytes); }
          catch { throw new Error(`Invalid UTF-8 in word definition: ${selected.id} -> ${reference.source}`); }
          const [start, end] = reference.lines;
          code = sourceLines(code, start, end, `Word definition ${selected.id}`);
          const label = `${reference.source} (lines ${start}–${end})`;
          const link = sourceURL(root, reference.source, identity);
          const source = link ? `<a href="${html(link)}#L${start}-L${end}"><code>${html(label)}</code></a>` : `<code>${html(label)}</code>`;
          const language = path.posix.extname(reference.source).slice(1);
          const kind = { l: 'L source', json: 'Configured JSON', go: 'Compiler definition' }[language];
          excerpts.push(`<p>${kind}: ${source}.</p>\n${sourceCode(code, language, `${selected.name} source definition`)}`);
        }
        add(node, `<details class="l-word-definition" data-pagefind-ignore><summary>Source definition</summary>\n${excerpts.join('\n')}\n</details>`);
        continue;
      }
      const index = node.value.trim().match(/^<!--\s*word-index(?::([a-z0-9:-]+))?\s*-->$/);
      if (/<!--\s*word-index\b/.test(node.value) && (!index || !tree.children.includes(node))) throw new Error(`Invalid word-index directive in ${page.source}:${node.position.start.line}`);
      if (index) {
        if (++indices > 1) throw new Error(`Multiple word indices in ${page.source}`);
        const rendered = wordIndex(words, index[1] ?? '', dictionaries);
        for (const [, id] of rendered.matchAll(/\bid="([^"]+)"/g)) if (document.anchors.has(id)) throw new Error(`Word index conflicts with authored anchor ${id} in ${page.source}`);
        add(node, rendered); continue;
      }
      if (/^<!--\s*stdlib-catalog\s*-->$/.test(node.value.trim())) { add(node, stdlibCatalog(dictionaries)); continue; }
      const illustration = node.value.trim().match(/^<!--\s*illustration:(source-to-native|dictionaries)\s*-->$/);
      if (illustration) {
        const name = illustration[1];
        const content = name === 'source-to-native'
          ? '<img class="dark:sl-block light:sl-hidden" src="/assets/l/illustrations/hero-source-to-native-dark.svg" alt="Source passes through a compiler to an App process and an additional App-constructed Service library" width="1000" height="680" loading="lazy" /><img class="dark:sl-hidden light:sl-block" src="/assets/l/illustrations/hero-source-to-native-light.svg" alt="Source passes through a compiler to an App process and an additional App-constructed Service library" width="1000" height="680" loading="lazy" />'
          : '<img src="/assets/l/illustrations/dictionaries.svg" alt="Vocabulary dictionaries group names that a program can explicitly import" loading="lazy" />';
        add(node, `<figure class="l-section-art">${content}</figure>`);
        continue;
      }
      const diagram = node.value.trim().match(/^<!--\s*diagram:(compiler-pipeline|runtime-lifecycle|process-model)\s*-->$/);
      if (diagram) {
        const labels = { 'compiler-pipeline': 'L source through Lyger, AML, Logiceme Rust bundles and Cargo native compilation', 'runtime-lifecycle': 'Build checks readiness, Run launches, Stop requests termination, and Await joins final output', 'process-model': 'Orchestrator-owned App processes, App-constructed Service libraries, and structural standalone Services' };
        const name = diagram[1];
        add(node, `<figure class="l-diagram"><img class="dark:sl-block light:sl-hidden" src="/assets/l/diagrams/${name}-dark.svg" alt="${labels[name]}" loading="lazy" /><img class="dark:sl-hidden light:sl-block" src="/assets/l/diagrams/${name}-light.svg" alt="${labels[name]}" loading="lazy" /><figcaption><a href="/assets/l/diagrams/${name}-light.svg">Open full-size diagram</a>. The surrounding prose explains the boundaries.</figcaption></figure>`);
        continue;
      }
      const match = node.value.trim().match(/^<!--\s*example:([A-Za-z0-9_-]+)\s*-->$/);
      if (!match) continue;
      const example = examples.find((e) => e.id === match[1]);
      if (!example) throw new Error(`Unknown example ${match[1]} in ${page.source}`);
      const bytes = await fs.readFile(await safeFile(root, example.source));
      if (sha256(bytes) !== example.sha256) throw new Error(`Stale example hash: ${example.id}`);
      const excerpt = Object.hasOwn(example, 'lines');
      if (excerpt && (example.classification !== 'fragment' || !example.lines || Array.isArray(example.lines) ||
        !Number.isSafeInteger(example.lines.start) || !Number.isSafeInteger(example.lines.end) ||
        example.lines.start < 1 || example.lines.end < example.lines.start ||
        Object.keys(example.lines).some((key) => !['start', 'end'].includes(key)))) {
        throw new Error(`Invalid example lines for ${example.id}: excerpts require classification fragment and positive inclusive start/end integers.`);
      }
      let code; let displayNote = '';
      try { code = new TextDecoder('utf-8', { fatal: true, ignoreBOM: excerpt }).decode(bytes); }
      catch {
        if (excerpt) throw new Error(`Invalid UTF-8 in excerpt source: ${example.id}. Use a whole-file escaped-byte inclusion instead.`);
        code = [...bytes].map((b) => b >= 32 && b < 127 && b !== 92 ? String.fromCharCode(b) : `\\x${b.toString(16).padStart(2, '0')}`).join(''); displayNote = ' Display uses escaped bytes because the source is not valid UTF-8.';
      }
      let sourceLabel = example.source;
      let sourceFragment = '';
      if (excerpt) {
        const { start, end } = example.lines;
        code = sourceLines(code, start, end, `Example ${example.id}`);
        sourceLabel += ` (lines ${start}–${end})`;
        sourceFragment = `#L${start}-L${end}`;
      }
      const fence = '`'.repeat(Math.max(3, ...[...code.matchAll(/`+/g)].map((m) => m[0].length + 1)));
      const language = example.lang ?? example.language ?? 'l';
      const block = example.classification === 'fragment' ? sourceCode(code, language, sourceLabel)
        : `${fence}${language} title="${sourceLabel}"\n${code}${code.endsWith('\n') ? '' : '\n'}${fence}`;
      const link = sourceURL(root, example.source, identity);
      const runtimeFailure = example.classification === 'runnable' && Number.isInteger(example.expectedExit) && example.expectedExit !== 0;
      const label = runtimeFailure ? 'Complete program · expected runtime failure' : { runnable: 'Complete program', 'compile-only': 'Compile-only example', invalid: 'Expected compilation failure', fragment: 'Fragment' }[example.classification];
      add(node, `${block}\n\n<div data-pagefind-ignore>\n\n**${label}.** Source: ${link ? `[${sourceLabel}](${link}${sourceFragment})` : `\`${sourceLabel}\``}.${displayNote}\n\n<details>\n<summary>Example evidence</summary>\n\n${example.description} **${example.validation}**. [Evidence guide](/docs/contributing/evidence/).\n\n</details>\n\n</div>`);
      continue;
    }
    if (node.type.endsWith('Reference')) {
      const def = defs.get(node.identifier); if (!def) throw new Error(`Missing reference definition ${node.identifier}`);
      const resolved = await resolve(def.url, node.type === 'imageReference');
      if (resolved.source) add(node, `${node.type === 'imageReference' ? node.alt : plain(node)} (source: \`${resolved.source}\`)`);
      else if (node.type === 'imageReference') add(node, `![${node.alt}](<${resolved.url}>)`);
      continue;
    }
    const result = await resolve(node.url, node.type === 'image');
    if (node.type === 'definition') {
      if (result.source) add(node, '');
      else add(node, `[${node.label ?? node.identifier}]: <${result.url}>${node.title ? ` ${JSON.stringify(node.title)}` : ''}`);
    } else if (result.source) add(node, `${plain(node)} (source: \`${result.source}\`)`);
    else if (result.url !== node.url) {
      const raw = text.slice(node.position.start.offset, node.position.end.offset);
      const labelEnd = raw.lastIndexOf('](');
      if (labelEnd < 0) throw new Error(`Unsupported inline link syntax in ${page.source}`);
      add(node, raw.slice(0, labelEnd + 2) + `<${result.url}>${node.title ? ` ${JSON.stringify(node.title)}` : ''})`);
    }
  }
  changes.sort((a, b) => b.start - a.start);
  let boundary = text.length + 1;
  for (const change of changes) {
    if (change.end > boundary) throw new Error(`Overlapping Markdown transformations in ${page.source}`);
    text = text.slice(0, change.start) + change.replacement + text.slice(change.end); boundary = change.start;
  }
  const source = sourceURL(root, page.source, identity);
  const edit = source && !page.generated && identity.channel === 'production' ? source.replace(`/blob/${identity.docsCommit}/`, '/edit/main/') : false;
  const metadata = { ...authored, title: page.title, description: page.description, slug: page.route.slice(1, -1), editUrl: edit, sourcePath: page.source, documentRole: page.role, generatedReference: page.generated };
  const attribution = `\n\n<div data-pagefind-ignore>\n\n---\n\n${page.generated ? '**Generated reference:**' : '**Canonical source:**'} ${source ? `[${page.source}](${source})` : `\`${page.source}\``}.${identity.dirty && source ? ' Source links show the committed base. This local checkout contains changes.' : ''}\n\n</div>\n`;
  return `---\n${yaml(metadata)}---\n\n${text.trim()}${attribution}`;
}
