// Reviewed meanings and authored paragraphs are the only inputs. This collection
// is transient: it is shared by coverage and rendering, never a second registry.
const parts = new Set(['noun', 'verb', 'adjective', 'adverb', 'article', 'preposition', 'conjunction', 'evaluation']);
const compare = (a, b) => a < b ? -1 : a > b ? 1 : 0;
const escape = (value) => String(value).replaceAll('&', '&amp;').replaceAll('<', '&lt;').replaceAll('>', '&gt;').replaceAll('"', '&quot;');
const text = (node) => node.type === 'html' ? '' : node.value ?? (node.children ?? []).map(text).join('');
const marker = /^<!--\s*word:([^\s<>]+)\s*-->$/;
const explicitAnchor = /^<(?:a|span)\s+id=["'][^"']+["']\s*>\s*<\/(?:a|span)>$/;
const anchorID = (node) => {
  const value = node?.type === 'html' ? node.value : node?.type === 'paragraph' && node.children.every((child) => child.type === 'html' || (child.type === 'text' && !child.value.trim())) ? node.children.map((child) => child.value).join('') : '';
  return explicitAnchor.test(value.trim()) ? value.match(/id=["']([^"']+)/)[1] : null;
};
const forms = (entry) => [...new Set([entry.name.split('.').at(-1), ...(entry.inflections ?? [])])];
const eligible = (entry) => ['builtin', 'grammar-word', 'authoring-word'].includes(entry.kind)
  || (entry.kind === 'type' && entry.support !== 'internal')
  || (entry.kind === 'stdlib-export' && entry.visibility === 'public' && entry.support !== 'internal');

export function deriveWords({ entries, lexical = { entries: {} }, pages, documents }) {
  const byId = new Map(entries.map((entry) => [entry.id, entry]));
  const canonical = (id) => {
    const entry = byId.get(id);
    if (!entry?.word || entry.wordRef) throw new Error(`Missing direct canonical word: ${id}`);
    return entry;
  };
  for (const entry of entries) {
    if (entry.word && entry.wordRef) throw new Error(`Entry has both word and wordRef: ${entry.id}`);
    if (eligible(entry) && !entry.word && !entry.wordRef) throw new Error(`Missing word coverage: ${entry.id}`);
    if (entry.wordRef) canonical(entry.wordRef);
    if (!entry.word) continue;
    const word = entry.word;
    if (!word || typeof word !== 'object' || Array.isArray(word) || Object.keys(word).some((key) => !['partOfSpeech', 'audience', 'formRoles', 'targets'].includes(key))) throw new Error(`Invalid word metadata: ${entry.id}`);
    if (!parts.has(word.partOfSpeech)) throw new Error(`Invalid word part of speech: ${entry.id}`);
    if (word.audience !== undefined && !['application', 'authoring'].includes(word.audience)) throw new Error(`Invalid word audience: ${entry.id}`);
    if (!eligible(entry) || entry.support === 'internal' || (entry.kind === 'authoring-word' && word.audience !== 'authoring')) throw new Error(`Internal or private meaning cannot be promoted: ${entry.id}`);
    if (entry.inflections !== undefined && (!Array.isArray(entry.inflections) || entry.inflections.some((form) => typeof form !== 'string' || !form))) throw new Error(`Invalid declared forms: ${entry.id}`);
    if (word.formRoles !== undefined && (!word.formRoles || typeof word.formRoles !== 'object' || Array.isArray(word.formRoles))) throw new Error(`Invalid form roles: ${entry.id}`);
    for (const [form, role] of Object.entries(word.formRoles ?? {})) {
      if (!forms(entry).includes(form)) throw new Error(`Undeclared form ${form}: ${entry.id}`);
      if (!role || !parts.has(role.partOfSpeech) || typeof role.role !== 'string' || !role.role.trim() || Object.keys(role).some((key) => !['partOfSpeech', 'role'].includes(key))) throw new Error(`Invalid form role ${form}: ${entry.id}`);
    }
    if (word.targets !== undefined && (!Array.isArray(word.targets) || !word.targets.length)) throw new Error(`Invalid modifier targets: ${entry.id}`);
    for (const target of word.targets ?? []) {
      if (!target || typeof target.entry !== 'string' || Object.keys(target).some((key) => !['entry', 'form'].includes(key))) throw new Error(`Invalid modifier target: ${entry.id}`);
      let owner;
      try { owner = canonical(target.entry); } catch { throw new Error(`Unknown canonical modifier target ${target.entry}: ${entry.id}`); }
      if (target.form !== undefined && !forms(owner).includes(target.form)) throw new Error(`Unknown target form ${target.form}: ${entry.id}`);
    }
  }
  for (const [key, review] of Object.entries(lexical.entries)) {
    if (review.wordRef) canonical(review.wordRef);
    else if (review.support !== 'internal' || !review.notes?.trim()) throw new Error(`Missing configured word association: ${key}`);
  }
  const found = new Map();
  for (const page of pages) {
    const document = documents.get(page.source);
    const children = document.tree.children;
    const visit = (node, top = false) => {
      if (node.type === 'html' && /<!--\s*word:/.test(node.value) && (!top || !marker.test(node.value.trim()))) throw new Error(`Word marker must be standalone in ${page.source}:${node.position.start.line}`);
      for (const child of node.children ?? []) visit(child, node.type === 'root');
    };
    visit(document.tree);
    for (let i = 0; i < children.length; i++) {
      const node = children[i];
      const match = node.type === 'html' && node.value.trim().match(marker);
      if (!match) continue;
      const id = match[1]; const entry = byId.get(id);
      const location = `${page.source}:${node.position.start.line} (${id})`;
      if (!entry?.word) throw new Error(`Unknown word marker ${id} at ${location}`);
      if (found.has(id)) throw new Error(`Duplicate word marker: ${location}`);
      let previous = i - 1;
      const allowed = new Set();
      while (previous >= 0 && anchorID(children[previous])) {
        allowed.add(anchorID(children[previous])); previous--;
      }
      const heading = children[previous];
      if (heading?.type !== 'heading') throw new Error(`Word marker must follow immediately after its heading: ${location}`);
      allowed.add(document.headings.get(heading));
      const [source, anchor] = entry.doc.split('#');
      if (source !== page.source || !allowed.has(anchor)) throw new Error(`Word marker at wrong source or anchor: ${location}; expected ${entry.doc}`);
      if (document.anchorCounts.get(anchor) !== 1) throw new Error(`Ambiguous canonical word anchor ${anchor}: ${location}`);
      const definition = children[i + 1];
      const label = definition?.children?.length === 1 && definition.children[0].type === 'strong' && ['Usage', 'Example', 'Behavior', 'Availability', 'Compose with'].includes(text(definition).replace(/:$/, ''));
      if (definition?.type !== 'paragraph' || !text(definition).trim() || label || definition.children?.some((child) => child.type === 'html' && /<!--/.test(child.value))) throw new Error(`Missing definition paragraph: ${location}`);
      let end = i + 2;
      while (end < children.length && !(children[end].type === 'heading' && children[end].depth <= heading.depth)) end++;
      found.set(id, { ...entry, root: entry.name.split('.').at(-1), audience: entry.word.audience ?? 'application', forms: forms(entry), source, anchor, link: page.route + '#' + anchor, gloss: text(definition).replace(/\s+/g, ' ').trim(), marker: node, section: { start: heading.position.start.offset, end: children[end]?.position.start.offset ?? document.text.length } });
    }
  }
  for (const entry of entries) if (entry.word && !found.has(entry.id)) throw new Error(`Missing canonical word marker: ${entry.id} at ${entry.doc}`);
  return [...found.values()].sort((a, b) => compare(a.root.toLowerCase(), b.root.toLowerCase()) || compare(a.name, b.name) || compare(a.dictionary?.lexicon ?? '', b.dictionary?.lexicon ?? '') || compare(a.id, b.id));
}

export function wordFacts(word, words, sourceLink) {
  const owner = word.dictionary;
  const availability = (owner ? `Explicit dictionary: ${owner.lexicon}/${owner.key}` : 'Implicit builtin vocabulary') + (word.audience === 'authoring' ? '; advanced word authoring' : '');
  const listedForms = word.forms.map((form) => {
    const role = word.word.formRoles?.[form];
    return `<code>${escape(form)}</code>${role ? ` (${escape(role.partOfSpeech)}; ${escape(role.role)})` : ''}`;
  }).join(', ');
  const targets = (word.word.targets ?? []).map((target) => {
    const selected = words.find((entry) => entry.id === target.entry);
    return `<a href="${escape(selected.link)}"><code>${escape(selected.name)}${target.form ? ` · ${escape(target.form)}` : ''}</code></a>`;
  });
  const sources = (word.sources ?? []).map((source) => sourceLink(source)).join(', ');
  return `<div class="l-word-facts">\n\n<p><strong>${escape(word.name)}</strong> · ${escape(word.word.partOfSpeech)} · ${escape(word.support)}<br />${escape(availability)}.<br />Declared forms: ${listedForms}.${targets.length ? `<br />Targets: ${targets.join(', ')}.` : ''}</p>\n\n<details data-pagefind-ignore><summary>Source and review evidence</summary><p>${escape(word.evidence)}. This documentation record is not a compiler identity or a runtime certificate.</p><p>${sources}</p></details>\n\n</div>`;
}

export function wordIndex(words, selection = '', dictionaries = []) {
  let selected = words;
  if (selection === 'builtins') selected = words.filter((word) => !word.dictionary);
  else if (selection) {
    const [lexicon, key, extra] = selection.split(':');
    if (extra || !dictionaries.some((dictionary) => dictionary.language === lexicon && dictionary.key === key)) throw new Error(`Unknown word-index library: ${selection}`);
    selected = words.filter((word) => word.dictionary?.lexicon === lexicon && word.dictionary.key === key && word.audience === 'application');
  }
  const groups = selection ? [['Words', selected]] : [['Application vocabulary', selected.filter((word) => word.audience === 'application')], ['Advanced word authoring', selected.filter((word) => word.audience === 'authoring')]];
  return groups.filter(([, items]) => items.length).map(([title, items], group) => {
    const letters = [...new Set(items.map((word) => word.root[0].toUpperCase()))];
    const id = (letter) => `words-${group}-${letter.toLowerCase()}`;
    const alphabet = letters.map((letter) => `<a href="#${escape(id(letter))}">${escape(letter)}</a>`).join(' · ');
    const sections = letters.map((letter) => `<h3 id="${escape(id(letter))}">${escape(letter)}</h3>\n<ul>${items.filter((word) => word.root[0].toUpperCase() === letter).map((word) => `<li><a href="${escape(word.link)}"><strong>${escape(word.root)}</strong> — ${escape(word.name)}</a> <span>(${escape(word.word.partOfSpeech)}; ${escape(word.support)}${word.audience === 'authoring' ? '; authoring' : ''})</span>: ${escape(word.gloss)}</li>`).join('\n')}</ul>`).join('\n');
    return `<nav class="l-word-index" data-pagefind-ignore aria-label="${escape(title)}"><h2>${escape(title)}</h2><p>${alphabet}</p>${sections}</nav>`;
  }).join('\n\n');
}
