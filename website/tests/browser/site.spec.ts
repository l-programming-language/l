import { test, expect } from '@playwright/test';
import AxeBuilder from '@axe-core/playwright';
import fs from 'node:fs/promises';
import { readFileSync } from 'node:fs';

type Word = {
  id: string; name: string; doc: string; inflections?: string[];
  dictionary?: { lexicon: string; key: string };
  word?: { partOfSpeech: string; audience?: string; definitions?: { source: string; lines: [number, number] }[] };
};
const semantics = JSON.parse(readFileSync(new URL('../../../docs/_meta/semantics.json', import.meta.url), 'utf8')) as { entries: Word[] };
const contentMap = JSON.parse(readFileSync(new URL('../../content-map.json', import.meta.url), 'utf8')) as { pages: { source: string; route: string }[] };
const wordRecord = (id: string) => {
  const word = semantics.entries.find((entry) => entry.id === id && entry.word);
  if (!word) throw new Error(`Missing canonical word record: ${id}`);
  return word;
};
const pageRoute = (source: string) => {
  const page = contentMap.pages.find((entry) => entry.source === source);
  if (!page) throw new Error(`Missing canonical page route: ${source}`);
  return page.route;
};
const wordURL = (id: string) => {
  const [source, anchor] = wordRecord(id).doc.split('#');
  if (!anchor) throw new Error(`Missing canonical word anchor: ${id}`);
  return `${pageRoute(source)}#${anchor}`;
};
const wordAnchor = (id: string) => `[id=${JSON.stringify(wordRecord(id).doc.split('#')[1])}]`;
const wordLink = (id: string) => `a[href=${JSON.stringify(wordURL(id))}]`;
const glossaryRoute = pageRoute('docs/reference/word-glossary.md');
const nativeWordsRoute = pageRoute('docs/reference/words.md');
const printID = 'builtin:print';
const returnID = 'builtin:function.return';

const pages = [
  ['home', '/'],
  ['docs', '/docs/'],
  ['guide', '/docs/guide/first-program/'],
  ['internals', '/docs/internals/'],
  ['definitions', '/docs/guide/definitions/'],
  ['words', '/docs/words/'],
  ['word-glossary', glossaryRoute],
  ['primitives', '/docs/spec/values/'],
  ['grammar', '/docs/spec/grammar/'],
  ['logemes', '/docs/reference/logemes/alphabet/'],
  ['cli', '/docs/cli/build/'],
] as const;

for (const width of [390, 768, 1440]) for (const theme of ['light', 'dark']) {
  test(`${width}px ${theme}: rendered pages, assets, accessibility and screenshots`, async ({ page }) => {
    test.setTimeout(120000);
    await page.setViewportSize({ width, height: 1000 });
    await page.addInitScript((value) => localStorage.setItem('starlight-theme', value), theme);
    const errors: string[] = []; const requests: string[] = [];
    page.on('pageerror', (e) => errors.push(e.message));
    page.on('console', (m) => { if (m.type() === 'error') errors.push(m.text()); });
    page.on('requestfailed', (r) => errors.push(`${r.url()}: ${r.failure()?.errorText}`));
    page.on('request', (r) => { if (!r.url().startsWith('http://127.0.0.1:4321/') && !r.url().startsWith('data:')) requests.push(r.url()); });
    await fs.mkdir('reports/screenshots', { recursive: true });
    for (const [name, route] of pages) {
      const response = await page.goto(route);
      expect(response?.status(), route).toBe(200);
      await expect(page.locator('html')).toHaveAttribute('data-theme', theme);
      await expect(page.locator('h1').filter({ visible: true })).toHaveCount(1);
      expect(await page.evaluate(() => document.documentElement.scrollWidth <= window.innerWidth + 1), `${route} horizontal overflow`).toBe(true);
      for (const image of await page.locator('img').all()) {
        if (await image.isVisible()) await image.scrollIntoViewIfNeeded();
        else if (await image.getAttribute('loading') === 'lazy') continue;
        await expect.poll(() => image.evaluate((img) => img instanceof HTMLImageElement && img.complete && img.naturalWidth > 0), { message: `${route}: ${await image.getAttribute('src')}` }).toBe(true);
      }
      await page.evaluate(() => window.scrollTo(0, 0));
      const axe = await new AxeBuilder({ page }).withTags(['wcag2a', 'wcag2aa', 'wcag21aa']).analyze();
      await fs.writeFile(`reports/axe-${name}-${width}-${theme}.json`, JSON.stringify(axe, null, 2));
      expect(axe.violations.map((v) => ({ id: v.id, impact: v.impact, nodes: v.nodes.map((n) => n.target) })), route).toEqual([]);
      await page.screenshot({ path: `reports/screenshots/${name}-${width}-${theme}.png`, fullPage: name !== 'word-glossary' });
      if (name === 'word-glossary') for (const [sample, id] of [['main', 'builtin:main'], ['print', printID]]) {
        await page.locator(wordAnchor(id)).scrollIntoViewIfNeeded();
        await page.screenshot({ path: `reports/screenshots/word-glossary-${sample}-${width}-${theme}.png`, fullPage: false });
      }
    }
    expect(errors).toEqual([]); expect(requests).toEqual([]);
  });
}

test('built Pagefind searches, no-results behavior and keyboard interaction', async ({ page }) => {
  await page.goto('/docs/');
  await page.keyboard.press('Control+k');
  const dialog = page.getByRole('dialog', { name: 'Search' });
  await expect(dialog).toBeVisible();
  const input = dialog.getByRole('textbox', { name: 'Search', exact: true });
  await expect(input).toBeFocused();
  for (const query of ['Text', 'Print', 'logeme']) {
    await input.fill(query);
    await expect(dialog.locator('.pagefind-ui__result-link').first()).toBeVisible();
    await expect(dialog.locator('.pagefind-ui__message')).not.toContainText('No results');
  }
  await input.fill('zznonexistentlqueryzz');
  await expect(dialog.locator('.pagefind-ui__message')).toContainText(/No results/i);
  await input.fill('Await');
  await expect(dialog.locator('.pagefind-ui__result-link').first()).toBeVisible();
  await page.keyboard.press('Tab');
  const focusedInDialog = await page.evaluate(() => Boolean(document.activeElement?.closest('dialog')));
  expect(focusedInDialog).toBe(true);
  await page.keyboard.press('Escape');
  await expect(dialog).not.toBeVisible();
});

const wordSearches = [
  { query: 'Print', destinations: [{ url: wordURL(printID), context: /Print/ }] },
  { query: 'Text', destinations: [{ url: wordURL('type:text'), context: /Text/ }] },
  { query: 'While', destinations: [{ url: wordURL('grammar-word:while'), context: /While/i }] },
] as const;

for (const { query, destinations } of wordSearches) {
  test(`word dictionary search: ${query} reaches its canonical meanings`, async ({ page }, testInfo) => {
    await page.goto('/docs/');
    // Read the actual built index, including subresults the stock UI may thin
    // to three per page. Ranking order is not a semantic contract.
    const matches = await page.evaluate(async (term) => {
      const modulePath = '/pagefind/pagefind.js';
      const pagefind = await import(modulePath);
      type Result = { url: string; excerpt: string; meta: { title: string }; sub_results: { url: string; title: string; excerpt: string }[] };
      const found = await pagefind.search(term) as { results: { data(): Promise<Result> }[] };
      const data = await Promise.all(found.results.map((result) => result.data()));
      const text = (html: string) => new DOMParser().parseFromString(html, 'text/html').body.textContent ?? '';
      return data.flatMap((result) => [
        { url: result.url, context: text(`${result.meta.title} ${result.excerpt}`) },
        ...result.sub_results.map((sub) => ({ url: sub.url, context: text(`${result.meta.title} ${sub.title} ${sub.excerpt}`) })),
      ]).map((match) => {
        const url = new URL(match.url, location.origin);
        return { ...match, url: url.pathname + url.hash };
      });
    }, query);
    await testInfo.attach('built-search-results', { body: JSON.stringify(matches, null, 2), contentType: 'application/json' });
    for (const expected of destinations) {
      const canonical = matches.filter((match) => match.url === expected.url);
      expect(canonical.length, `${query}: missing ${expected.url}; complete built results are attached`).toBeGreaterThan(0);
      expect(canonical.map((match) => match.context).join(' '), `${query}: readable meaning context`).toMatch(expected.context);
    }
    if (query === 'Print') {
      expect(matches.filter((match) => /#words-/.test(match.url)), 'generated word navigation must stay outside the search index').toEqual([]);
    }

    await page.keyboard.press('Control+k');
    const dialog = page.getByRole('dialog', { name: 'Search' });
    await expect(dialog).toBeVisible();
    await dialog.getByRole('textbox', { name: 'Search', exact: true }).fill(query);
    await expect(dialog.locator('.pagefind-ui__result-link').first()).toBeVisible();
    await expect(dialog.locator('.pagefind-ui__loading')).toHaveCount(0);
    // The reader must be offered the actual sections, not just a page whose
    // source happens to contain the query.
    for (const expected of destinations) {
      const link = dialog.locator(`.pagefind-ui__result-link[href="${expected.url}"]`);
      const more = dialog.locator('.pagefind-ui__button');
      while (await link.count() === 0 && await more.isVisible()) {
        const count = await dialog.locator('.pagefind-ui__result').count();
        await more.click();
        await expect.poll(() => dialog.locator('.pagefind-ui__result').count()).toBeGreaterThan(count);
        await expect(dialog.locator('.pagefind-ui__loading')).toHaveCount(0);
      }
      await expect(link).toBeVisible();
      const result = link.locator('xpath=ancestor::*[contains(concat(" ", normalize-space(@class), " "), " pagefind-ui__result ")][1]');
      await expect(result).toContainText(expected.context);
    }
    await dialog.locator(`.pagefind-ui__result-link[href="${destinations[0].url}"]`).click();
    await expect(page).toHaveURL(`http://127.0.0.1:4321${destinations[0].url}`);
    await page.waitForLoadState('domcontentloaded');
    const anchor = destinations[0].url.slice(destinations[0].url.indexOf('#'));
    await expect(page.locator(anchor)).toBeVisible();
    const role = query === 'Text' ? /noun/i : query === 'While' ? /rejected|unsupported/i : null;
    if (role) {
      const section = await page.locator(anchor).evaluate((heading) => {
        const headings = [...document.querySelectorAll('main :is(h1,h2,h3,h4,h5,h6)')];
        const next = headings.slice(headings.indexOf(heading) + 1).find((candidate) => candidate.tagName <= heading.tagName);
        const range = document.createRange();
        range.setStartBefore(heading);
        if (next) range.setEndBefore(next); else range.setEndAfter(document.querySelector('main')!);
        return range.toString();
      });
      expect(section).toMatch(role);
    }
  });
}

test('word dictionary browsing works without JavaScript', async ({ browser }) => {
  const context = await browser.newContext({ javaScriptEnabled: false, viewport: { width: 390, height: 844 } });
  const page = await context.newPage();
  await page.goto(`http://127.0.0.1:4321${nativeWordsRoute}`);
  await expect(page.getByRole('heading', { level: 1 })).toHaveText('Native Words');
  const alphabet = page.locator('.l-word-index a[href^="#"]').first();
  const fragment = await alphabet.getAttribute('href');
  expect(fragment).toMatch(/^#[^#]+$/);
  await alphabet.click();
  await expect(page.locator(fragment!)).toBeVisible();
  const nativeDestinations = await page.locator('.l-word-index li > a').evaluateAll((links) => links.map((link) => link.getAttribute('href')).sort());
  expect(nativeDestinations).toEqual(semantics.entries.filter((entry) => entry.word && !entry.dictionary).map((entry) => wordURL(entry.id)).sort());
  await page.locator(`.l-word-index ${wordLink(returnID)}`).click();
  await expect(page).toHaveURL(`http://127.0.0.1:4321${wordURL(returnID)}`);
  await expect(page.locator(wordAnchor(returnID))).toHaveCount(1);
  await expect(page.getByRole('heading', { name: "Function's Return", exact: true })).toBeVisible();
  expect(await page.evaluate(() => document.documentElement.scrollWidth <= innerWidth + 1)).toBe(true);
  await context.close();
});

test('word navigation and source disclosures support keyboard focus', async ({ page }) => {
  await page.setViewportSize({ width: 390, height: 844 });
  await page.goto(nativeWordsRoute);
  const print = page.locator(`.l-word-index ${wordLink(printID)}`);
  await print.focus();
  await expect(print).toBeFocused();
  expect(await print.evaluate((element) => getComputedStyle(element).outlineStyle)).not.toBe('none');
  await page.keyboard.press('Enter');
  await page.waitForURL((url) => url.pathname === glossaryRoute, { waitUntil: 'domcontentloaded' });
  await expect(page.locator(wordAnchor(printID))).toBeVisible();
  const facts = page.locator(wordAnchor(printID)).locator('xpath=following::div[contains(concat(" ", normalize-space(@class), " "), " l-word-facts ")][1]');
  await expect(facts.locator('ol > li')).toHaveCount(wordRecord(printID).inflections!.length);
  await expect(facts.locator('ol')).toContainText('Printer (noun, actor)');
  const disclosure = facts.locator('details');
  await disclosure.locator('summary').focus();
  await page.keyboard.press('Enter');
  await expect(disclosure).toHaveAttribute('open', '');
  await expect(disclosure.locator('a').first()).toBeVisible();
  await page.keyboard.press('Enter');
  await expect(disclosure).not.toHaveAttribute('open', '');
});

for (const id of ['builtin:main', printID]) {
  test(`glossary source definition preserves exact selected bytes: ${id}`, async ({ page }) => {
    const word = wordRecord(id);
    const references = word.word?.definitions;
    expect(references?.length, `${id}: reviewed source definitions`).toBeGreaterThan(0);
    await page.setViewportSize({ width: 390, height: 844 });
    await page.goto(wordURL(id));
    const heading = page.locator(wordAnchor(id));
    await expect(heading).toBeVisible();
    const definition = heading.locator('xpath=following::details[contains(concat(" ", normalize-space(@class), " "), " l-word-definition ")][1]');
    await expect(definition).not.toHaveAttribute('open', '');
    const blocks = definition.locator('pre code');
    await expect(blocks).toHaveCount(references!.length);
    await expect(blocks.first()).not.toBeVisible();
    await definition.getByText('Source definition', { exact: true }).click();
    await expect(definition).toHaveAttribute('open', '');
    for (const [index, reference] of references!.entries()) {
      const [start, end] = reference.lines;
      const source = await fs.readFile(new URL(`../../../${reference.source}`, import.meta.url), 'utf8');
      const selected = (source.match(/[^\n]*\n|[^\n]+$/g) ?? []).slice(start - 1, end).join('');
      await expect(blocks.nth(index)).toBeVisible();
      expect(await blocks.nth(index).textContent()).toBe(selected);
      await expect(definition).toContainText(`${reference.source} (lines ${start}–${end})`);
      await expect(definition.locator(`a[href$=${JSON.stringify(`${reference.source}#L${start}-L${end}`)}]`)).toBeVisible();
      const pre = blocks.nth(index).locator('..');
      if (await pre.evaluate((element) => element.scrollWidth > element.clientWidth)) {
        await pre.focus();
        await expect(pre).toBeFocused();
        await page.keyboard.press('ArrowRight');
        await expect.poll(() => pre.evaluate((element) => element.scrollLeft)).toBeGreaterThan(0);
      }
    }
    expect(await page.evaluate(() => document.documentElement.scrollWidth <= innerWidth + 1)).toBe(true);
  });
}

test('theme persistence, correct logos, copy controls and local cross-links', async ({ page, context }) => {
  await context.grantPermissions(['clipboard-read', 'clipboard-write']);
  await page.goto('/docs/guide/first-program/');
  const select = page.locator('starlight-theme-select select').first();
  await select.selectOption('light');
  await expect(page.locator('html')).toHaveAttribute('data-theme', 'light');
  await page.reload();
  await expect(page.locator('html')).toHaveAttribute('data-theme', 'light');
  const lightLogo = await page.locator('.site-title img:visible').getAttribute('src');
  await select.selectOption('dark');
  const darkLogo = await page.locator('.site-title img:visible').getAttribute('src');
  expect(darkLogo).not.toBe(lightLogo);
  const copy = page.locator('.expressive-code .copy button').first();
  await copy.click();
  expect(await page.evaluate(() => navigator.clipboard.readText())).toContain('The App Prints message.');
  await page.locator(`main a[href="${pageRoute('docs/cli-build.md')}"]`).first().click();
  await expect(page).toHaveURL(`http://127.0.0.1:4321${pageRoute('docs/cli-build.md')}`);
});

test('checkout build command copies intact and the retained bootstrap is served unchanged', async ({ page, context, request }) => {
  await context.grantPermissions(['clipboard-read', 'clipboard-write']);
  await page.setViewportSize({ width: 390, height: 844 });
  await page.goto('/docs/cli/installation/');
  const command = [
    'scripts/l-go build -mod=readonly -o /tmp/l ./cmd/l',
    '/tmp/l --version',
    '/tmp/l build demo/basic/main.l',
    './demo/basic/main',
  ].join('\n');
  await page.locator('.expressive-code .copy button').first().click();
  expect((await page.evaluate(() => navigator.clipboard.readText())).trimEnd()).toBe(command);
  expect(await page.evaluate(() => document.documentElement.scrollWidth <= window.innerWidth + 1)).toBe(true);
  const script = await request.get('/install.sh');
  expect(script.status()).toBe(200);
  expect(await script.body()).toEqual(await fs.readFile('public/install.sh'));
  await page.getByRole('link', { name: 'Installation options', exact: true }).click();
  await expect(page).toHaveURL(/\/docs\/cli\/installation-options\/$/);
});

test('mobile navigation and visible keyboard focus', async ({ page }) => {
  await page.setViewportSize({ width: 390, height: 844 });
  await page.goto('/docs/spec/values/');
  await page.keyboard.press('Tab');
  await expect(page.getByRole('link', { name: 'Skip to content' })).toBeFocused();
  expect(await page.getByRole('link', { name: 'Skip to content' }).evaluate((el) => getComputedStyle(el).outlineStyle)).not.toBe('none');
  await page.getByRole('button', { name: 'Menu', exact: true }).click();
  await expect(page.locator('#starlight__sidebar')).toBeVisible();
  await page.locator('#starlight__sidebar a[href="/docs/spec/grammar/"]').click();
  await expect(page).toHaveURL(/\/docs\/spec\/grammar\/$/);
  await page.goto('/docs/reference/logemes/alphabet/');
  const table = page.locator('main table').first();
  await table.focus();
  await page.keyboard.press('ArrowRight');
  await expect.poll(() => table.evaluate((el) => el.scrollLeft)).toBeGreaterThan(0);
});

test('rendered source preserves L, shell, Rust and literal AML', async ({ page }) => {
  const sourceBlocks = (language: string) => page.locator(`pre[data-language="${language}"], pre:has(> code.language-${language})`).evaluateAll((pres) => pres.map((pre) => {
    // Complete examples have Expressive Code lines; fragments retain one text node.
    const lines = [...pre.querySelectorAll('.ec-line .code')];
    return (lines.length ? lines.map((line) => line.textContent?.replace(/\n$/, '')).join('\n') : pre.querySelector('code')?.textContent ?? '').trimEnd();
  }));
  for (const [route, fixture, start, end] of [
    ['/docs/spec/definitions/', '../builtins/eng/core.l', 2, 5],
    ['/docs/guide/first-program/', '../demo/basic/main.l', 0, undefined],
  ] as const) {
    await page.goto(route);
    const source = (await fs.readFile(fixture, 'utf8')).split('\n').slice(start, end).join('\n').trimEnd();
    expect(await sourceBlocks('l')).toContain(source);
  }
  await expect(page.locator('pre[data-language="sh"]').first()).toContainText(/^l build demo\/basic\/main\.l/);
  await page.goto('/docs/reference/logemes/rust-contract/');
  await expect(page.locator('pre[data-language="rust"]')).toContainText("'outer: loop {");
  await page.goto('/docs/internals/aml-protocol/');
  const reference = await fs.readFile('../logiceme/AML.md', 'utf8');
  const aml = reference.match(/## Complete composed example\n[\s\S]*?```text\n([\s\S]*?)\n```/);
  expect(aml).not.toBeNull();
  expect(await sourceBlocks('text')).toContain(aml![1].trimEnd());
  expect(await page.locator('main app, main logeme, main outline').count()).toBe(0);
});

test('no-JavaScript reading, native menu and reduced motion', async ({ browser }) => {
  const context = await browser.newContext({ javaScriptEnabled: false, viewport: { width: 390, height: 844 }, reducedMotion: 'reduce' });
  const page = await context.newPage();
  await page.goto('http://127.0.0.1:4321/docs/spec/values/');
  await expect(page.getByRole('heading', { name: 'Primitive Types', exact: true })).toBeVisible();
  await page.getByRole('button', { name: 'Menu', exact: true }).click();
  await expect(page.locator('#starlight__sidebar')).toBeVisible();
  await page.locator('#starlight__sidebar a[href="/docs/spec/grammar/"]').click();
  await expect(page.getByRole('heading', { level: 1 })).toBeVisible();
  await context.close();
});

test('static 404 and truthful build provenance', async ({ request }) => {
  const missing = await request.get('/does-not-exist/deep-page/');
  expect(missing.status()).toBe(404);
  const missingHTML = await missing.text();
  expect(missingHTML).toContain('404');
  expect(missingHTML).toContain('Browse documentation');
  expect(missingHTML).toContain('Return home');
  expect(missingHTML).toMatch(/href="\/docs\/"/);
  expect(missingHTML).toMatch(/href="\/"/);
  expect((await request.get('/docs/')).status()).toBe(200);
  expect((await request.get('/')).status()).toBe(200);
  const nested = await request.get('/docs/reference/logemes/alphabet/'); expect(nested.status()).toBe(200);
  const version = await (await request.get('/version.json')).json();
  const expected = JSON.parse(await fs.readFile('dist/version.json', 'utf8'));
  expect(version).toEqual(expected);
  expect(version.sourceFingerprint).toMatch(/^sha256:[a-f0-9]{64}$/);
  expect(await (await request.get('/robots.txt')).text()).toContain(expected.channel === 'production' ? 'Allow: /' : 'Disallow: /');
});
