import { test, expect } from '@playwright/test';
import AxeBuilder from '@axe-core/playwright';
import fs from 'node:fs/promises';

const pages = [
  ['home', '/'],
  ['docs', '/docs/'],
  ['guide', '/docs/guide/first-program/'],
  ['apps-services', '/docs/guide/apps-and-services/'],
  ['lifecycle', '/docs/guide/lifecycle/'],
  ['internals', '/docs/internals/'],
  ['primitives', '/docs/spec/values/'],
  ['grammar', '/docs/spec/grammar/'],
  ['logemes', '/docs/reference/logemes/alphabet/'],
  ['cli', '/docs/cli/build/'],
  ['stdlib', '/docs/stdlib/eng/io/'],
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
      await page.screenshot({ path: `reports/screenshots/${name}-${width}-${theme}.png`, fullPage: true });
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
  for (const query of ['Number', 'Await', 'IO.Print', 'logeme']) {
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
  expect(await page.evaluate(() => navigator.clipboard.readText())).toContain('Use IO');
  await page.locator('main a[href="/docs/cli/"]').first().click();
  await expect(page).toHaveURL('http://127.0.0.1:4321/docs/cli/');
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

test('rendered source preserves L, shell, JSON, Rust and literal AML', async ({ page }) => {
  for (const [route, fixture, start, end] of [
    ['/docs/guide/apps-and-services/', '../cmd/l/test/runnable_service/canonical-io.l', 0, undefined],
    ['/docs/guide/apps-and-services/', '../lyger/test/workspace/services.l', 1, 12],
    ['/docs/guide/native-definitions/', '../cmd/l/test/runnable_service/probe.l', 0, 3],
    ['/docs/guide/native-definitions/', '../cmd/l/test/runnable_service/basic.l', 0, 4],
    ['/docs/guide/first-program/', '../demo/basic/main.l', 0, undefined],
  ] as const) {
    await page.goto(route);
    const source = (await fs.readFile(fixture, 'utf8')).split('\n').slice(start, end).join('\n').trimEnd();
    // Expressive Code uses divs per line and a newline sentinel for empty lines.
    const blocks = await page.locator('pre[data-language="l"]').evaluateAll((pres) => pres.map((pre) => [...pre.querySelectorAll('.ec-line .code')].map((line) => line.textContent?.replace(/\n$/, '')).join('\n').trimEnd()));
    expect(blocks).toContain(source);
  }
  await expect(page.locator('pre[data-language="sh"]').first()).toContainText(/^l build main\.l/);
  await page.goto('/docs/stdlib/');
  expect(JSON.parse((await page.locator('pre[data-language="json"] code').textContent())!)).toEqual(JSON.parse(await fs.readFile('../stdlib/index.json', 'utf8')));
  await page.goto('/docs/reference/logemes/rust-contract/');
  await expect(page.locator('pre[data-language="rust"]')).toContainText("'outer: loop {");
  await page.goto('/docs/internals/aml-protocol/');
  const reference = await fs.readFile('../logiceme/AML.md', 'utf8');
  const aml = reference.match(/## Complete composed example\n[\s\S]*?```text\n([\s\S]*?)\n```/);
  expect(aml).not.toBeNull();
  const blocks = await page.locator('pre[data-language="text"]').evaluateAll((pres) => pres.map((pre) => [...pre.querySelectorAll('.ec-line .code')].map((line) => line.textContent?.replace(/\n$/, '')).join('\n')));
  expect(blocks.map((x) => x.trimEnd())).toContain(aml![1].trimEnd());
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
  const nested = await request.get('/docs/stdlib/eng/io/'); expect(nested.status()).toBe(200);
  const version = await (await request.get('/version.json')).json();
  const expected = JSON.parse(await fs.readFile('dist/version.json', 'utf8'));
  expect(version).toEqual(expected);
  expect(version.sourceFingerprint).toMatch(/^sha256:[a-f0-9]{64}$/);
  expect(await (await request.get('/robots.txt')).text()).toContain(expected.channel === 'production' ? 'Allow: /' : 'Disallow: /');
});
