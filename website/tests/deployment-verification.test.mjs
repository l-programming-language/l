import test from 'node:test';
import assert from 'node:assert/strict';
import { mkdtemp, readFile, rm, writeFile } from 'node:fs/promises';
import { tmpdir } from 'node:os';
import { join } from 'node:path';
import { fileURLToPath } from 'node:url';
import { spawnSync } from 'node:child_process';
import * as release from '../scripts/release.mjs';

const repository = 'l-programming-language/l', base = `/repos/${repository}`;
const origin = 'https://ldocs.org', commit = 'a'.repeat(40), recoveryCommit = 'b'.repeat(40);
const script = fileURLToPath(new URL('../scripts/release.mjs', import.meta.url));
const { sha256, stableJSON } = release;

function fixture(t, { recovery = false } = {}) {
  const deploymentSha = recovery ? recoveryCommit : commit;
  const policy = { publicationEnabled: true, repository, origin, productionBranch: 'main', targets: release.nativeTargets,
    nativePublisher: { workflowId: 42, workflowPath: '.github/workflows/release.yml' } };
  const version = { languageVersion: '0.0.1', languageCommit: commit, docsCommit: commit, docsRevision: commit,
    channel: 'production', dirty: false, sourceFingerprint: `sha256:${'c'.repeat(64)}`,
    releaseFingerprint: 'd'.repeat(64), repository: `https://github.com/${repository}` };
  const versionBytes = stableJSON(version);
  const bodies = { '/': '<html>Released home</html>\n', '/docs/spec/': '<html>Released spec</html>\n',
    '/docs/cli/': '<html>Released CLI</html>\n', '/pagefind/pagefind.js': 'export const released = true;\n' };
  const files = [{ path: 'version.json', sha256: sha256(versionBytes) },
    ...Object.entries(bodies).map(([path, bytes]) => ({ path: path.endsWith('/') ? `${path.slice(1)}index.html` : path.slice(1), sha256: sha256(bytes) }))];
  const promotion = { schemaVersion: 1, mode: 'release', ...version, nativeManifestSha256: 'e'.repeat(64),
    siteArchiveSha256: 'f'.repeat(64), versionSha256: sha256(versionBytes), expectedCurrentDigest: null, files };
  const promotionURL = `https://api.github.com${base}/releases/assets/20`;
  const asset = { name: `docs-${commit}.promotion.json`, url: promotionURL, state: 'uploaded', digest: `sha256:${sha256(stableJSON(promotion))}` };
  const archive = { tag_name: 'v0.0.1', draft: false, prerelease: false, immutable: true, assets: [asset] };
  const deployment = { id: 30, sha: deploymentSha, ref: 'main', environment: 'github-pages', task: 'deploy',
    created_at: '2026-10-31T07:51:02Z', performed_via_github_app: { slug: 'github-actions' } };
  const statuses = [{ state: 'failure', log_url: `https://github.com/${repository}/actions/runs/10/job/11`,
    created_at: '2026-10-31T07:51:33Z' }];
  const job = { id: 11, run_id: 10, run_attempt: 1, head_sha: deploymentSha, status: 'completed', conclusion: 'failure',
    steps: [{ name: 'Authorize manual publication', status: 'completed', conclusion: 'success', number: 4 },
      { name: `Run actions/deploy-pages@${'1'.repeat(40)}`, status: 'completed', conclusion: 'success', number: 8 }] };
  const run = { id: 10, run_attempt: 1, workflow_id: recovery ? 43 : 42,
    path: recovery ? '.github/workflows/docs-recover-pages.yml' : '.github/workflows/release.yml',
    repository: { full_name: repository }, head_branch: 'main', head_sha: deploymentSha,
    event: 'workflow_dispatch', status: 'completed', conclusion: 'failure' };
  const pages = { status: 'succeed' };
  const routes = new Map([
    [`${base}/releases/tags/v0.0.1`, () => ({ json: archive })],
    [`${base}/git/ref/tags/v0.0.1`, () => ({ json: { object: { type: 'commit', sha: commit } } })],
    [`${base}/releases/assets/20`, () => ({ bytes: stableJSON(promotion) })],
    [`${base}/deployments/30/statuses`, () => ({ json: statuses })],
    [`${base}/actions/jobs/11`, () => ({ json: job })],
    [`${base}/actions/runs/10/attempts/1`, () => ({ json: run })],
    [`${base}/pages/deployments/${deploymentSha}`, () => ({ json: pages })],
  ]);
  const deploymentPages = new Map([[1, [deployment]]]);
  const live = new Map([['/version.json', [{ bytes: versionBytes }]], ...Object.entries(bodies).map(([path, bytes]) => [path, [{ bytes }]])]);
  const calls = [], counts = new Map();
  const previous = { token: process.env.GITHUB_TOKEN, sha: process.env.GITHUB_SHA };
  process.env.GITHUB_TOKEN = 'deployment-fixture-token'; process.env.GITHUB_SHA = deploymentSha;
  t.after(() => {
    for (const [key, value] of [['GITHUB_TOKEN', previous.token], ['GITHUB_SHA', previous.sha]]) {
      if (value === undefined) delete process.env[key]; else process.env[key] = value;
    }
  });
  const fetch = async (url, options = {}) => {
    const target = new URL(url), headers = new Headers(options.headers);
    assert.equal(options.method || 'GET', 'GET', 'Verification must remain read-only');
    calls.push(target.href);
    let entry;
    if (target.origin === origin) {
      assert.equal(headers.has('Authorization'), false, 'Public requests must not receive GitHub credentials');
      assert.equal(options.cache, 'no-store');
      assert.equal(options.redirect, 'error');
      const sequence = live.get(target.pathname);
      assert(sequence, `Unexpected public request: ${target.href}`);
      const count = counts.get(target.pathname) || 0; counts.set(target.pathname, count + 1);
      entry = sequence[Math.min(count, sequence.length - 1)];
    } else {
      assert.equal(target.origin, 'https://api.github.com');
      assert.equal(headers.get('Authorization'), 'Bearer deployment-fixture-token');
      if (target.pathname === `${base}/deployments`) {
        assert.equal(target.searchParams.get('environment'), 'github-pages');
        entry = { json: deploymentPages.get(Number(target.searchParams.get('page') || 1)) || [] };
      } else {
        assert(routes.has(target.pathname), `Unexpected GitHub request: ${target.href}`);
        entry = routes.get(target.pathname)();
      }
    }
    if (entry.error) throw new TypeError(entry.error);
    return new Response(entry.bytes ?? JSON.stringify(entry.json ?? {}), { status: entry.status ?? 200 });
  };
  t.mock.method(globalThis, 'fetch', fetch);
  return { policy, version, versionBytes, promotion, promotionURL, archive, asset, deployment, deploymentPages,
    statuses, job, run, pages, routes, live, calls, counts, deploymentSha,
    seal: () => { asset.digest = `sha256:${sha256(stableJSON(promotion))}`; } };
}

const verify = (f, options = {}) => release.verifyDeployment(f.policy, f.promotion,
  { deploymentSha: f.deploymentSha, attempts: 3, intervalMs: 0, ...options });

test('first postdeployment verification succeeds before the environment job completes', async (t) => {
  const f = fixture(t); f.statuses[0].state = 'in_progress'; f.job.status = 'in_progress'; f.job.conclusion = null;
  await verify(f);
  assert(f.calls.some((url) => url.endsWith(`/pages/deployments/${commit}`)));
  assert(!f.calls.some((url) => new URL(url).pathname === `${base}/deployments`), 'Current deployment must not depend on its own terminal job status');
  for (const path of ['/version.json', '/', '/docs/spec/', '/docs/cli/', '/pagefind/pagefind.js']) assert(f.counts.has(path), path);
});

test('postdeploy CLI verifies the first deployment using the running workflow SHA', async (t) => {
  const f = fixture(t), root = await mkdtemp(join(tmpdir(), 'l-postdeploy-cli-'));
  t.after(() => rm(root, { recursive: true, force: true }));
  const policyPath = join(root, 'policy.json'), manifestPath = join(root, 'promotion.json'), mockPath = join(root, 'fetch.mjs');
  await writeFile(policyPath, stableJSON(f.policy)); await writeFile(manifestPath, stableJSON(f.promotion));
  const responses = {
    [`https://api.github.com${base}/deployments?environment=github-pages&per_page=30`]: [f.deployment],
    [`https://api.github.com${base}/deployments/30/statuses`]: [{ state: 'in_progress' }],
    [`https://api.github.com${base}/pages/deployments/${commit}`]: f.pages,
    [`https://api.github.com${base}/releases/tags/v0.0.1`]: f.archive,
    [`https://api.github.com${base}/git/ref/tags/v0.0.1`]: { object: { type: 'commit', sha: commit } },
    [f.promotionURL]: stableJSON(f.promotion),
    ...Object.fromEntries([...f.live].map(([path, value]) => [origin + path, value[0].bytes])),
  };
  await writeFile(mockPath, `import assert from 'node:assert/strict';\nconst responses = ${JSON.stringify(responses)};\nglobalThis.fetch = async (url) => { assert(Object.hasOwn(responses,url),'Unexpected URL: '+url); const body=responses[url]; return new Response(typeof body==='string' ? body : JSON.stringify(body)); };\n`);
  const result = spawnSync(process.execPath, ['--import', mockPath, script, 'postdeploy', '--policy', policyPath,
    '--manifest', manifestPath, '--deployment-sha', commit], { encoding: 'utf8', timeout: 10000,
    env: { ...process.env, NODE_OPTIONS: '', GITHUB_TOKEN: 'deployment-fixture-token', GITHUB_SHA: commit } });
  assert.equal(result.error, undefined, result.error?.message);
  assert.equal(result.status, 0, result.stderr);
  assert.match(result.stdout, /verified/);
});

for (const state of ['in_progress', 'failure', 'success']) test(`production discovery accepts a successful Pages receipt while environment is ${state}`, async (t) => {
  const f = fixture(t); f.statuses[0].state = state;
  assert.deepEqual(await release.fetchCurrent(f.policy), f.promotion);
  assert(f.calls.some((url) => url.endsWith(`/pages/deployments/${commit}`)));
  assert(f.calls.some((url) => url.endsWith('/actions/runs/10/attempts/1')));
});

test('recovery uses its workflow SHA receipt while retaining an older archived docs identity', async (t) => {
  const f = fixture(t, { recovery: true });
  assert.deepEqual(await release.fetchCurrent(f.policy), f.promotion);
  assert(f.calls.some((url) => url.endsWith(`/pages/deployments/${recoveryCommit}`)));
  assert(!f.calls.some((url) => url.endsWith(`/pages/deployments/${commit}`)));
  await verify(f, { deploymentSha: undefined });
});

test('named deployment step is accepted during recovery', async (t) => {
  const f = fixture(t); f.job.steps[1].name = 'Deploy verified Pages artifact';
  assert.deepEqual(await release.fetchCurrent(f.policy), f.promotion);
});

test('production discovery paginates past deployments that did not publish Pages', async (t) => {
  const f = fixture(t);
  f.deploymentPages.set(1, Array.from({ length: 100 }, (_, i) => ({ ...f.deployment, id: 100 + i })));
  f.deploymentPages.set(2, [f.deployment]);
  for (let i = 0; i < 100; i++) f.routes.set(`${base}/deployments/${100 + i}/statuses`, () => ({ json: [{ state: 'queued' }] }));
  assert.deepEqual(await release.fetchCurrent(f.policy), f.promotion);
  assert(f.calls.some((url) => new URL(url).searchParams.get('page') === '2'));
});

test('a genuine empty production baseline requires HTTP 404 and no successful deployment', async (t) => {
  const f = fixture(t); f.deploymentPages.set(1, []); f.live.set('/version.json', [{ status: 404 }]);
  assert.equal(await release.fetchCurrent(f.policy), null);
});

for (const value of [null, false, 0, []]) test(`HTTP 200 with ${JSON.stringify(value)} cannot authorize an empty production baseline`, async (t) => {
  const f = fixture(t); f.deploymentPages.set(1, []);
  f.live.set('/version.json', [{ bytes: stableJSON(value) }]);
  await assert.rejects(release.fetchCurrent(f.policy), /Production identity must be an object/);
});

test('a previously successful deployment cannot be reclassified as empty production', async (t) => {
  const f = fixture(t); f.statuses[0].state = 'success'; f.live.set('/version.json', [{ status: 404 }]);
  await assert.rejects(release.fetchCurrent(f.policy));
});

test('a live site without authenticated deployment evidence is rejected', async (t) => {
  const f = fixture(t); f.deploymentPages.set(1, []);
  await assert.rejects(release.fetchCurrent(f.policy));
});

test('untrusted or failed deployment evidence never authorizes a production baseline', async (t) => {
  for (const mutate of [
    (f) => { f.pages.status = 'deployment_failed'; },
    (f) => { f.job.steps[1].conclusion = 'failure'; },
    (f) => { f.job.steps[1].conclusion = 'skipped'; },
    (f) => { f.job.head_sha = recoveryCommit; },
    (f) => { f.job.run_id = 99; },
    (f) => { f.run.repository.full_name = 'someone/else'; },
    (f) => { f.run.head_branch = 'feature'; },
    (f) => { f.run.event = 'pull_request'; },
    (f) => { f.run.path = '.github/workflows/untrusted.yml'; },
    (f) => { f.run.workflow_id = 99; },
    (f) => { f.statuses[0].log_url = 'https://github.com/someone/else/actions/runs/10/job/11'; },
  ]) await t.test(mutate.toString(), async (sub) => {
    const f = fixture(sub); mutate(f);
    await assert.rejects(release.fetchCurrent(f.policy));
  });
});

test('ordinary publication cannot use an unrelated workflow SHA as recovery evidence', async (t) => {
  const f = fixture(t, { recovery: true }); f.run.path = '.github/workflows/release.yml'; f.run.workflow_id = 42;
  await assert.rejects(release.fetchCurrent(f.policy));
});

test('postdeployment verification rejects a caller-supplied SHA unrelated to the current workflow', async (t) => {
  const f = fixture(t);
  await assert.rejects(verify(f, { deploymentSha: recoveryCommit }));
});

test('a failed Pages receipt rejects postdeployment verification even when HTTP is healthy', async (t) => {
  const f = fixture(t); f.pages.status = 'deployment_failed';
  await assert.rejects(verify(f));
});

test('immutable archive authority and exact live identity are enforced', async (t) => {
  for (const mutate of [
    (f) => { f.archive.immutable = false; },
    (f) => { f.archive.draft = true; },
    (f) => { f.archive.prerelease = true; },
    (f) => { f.asset.digest = `sha256:${'0'.repeat(64)}`; },
    (f) => { f.archive.assets.push({ ...f.asset }); },
    (f) => { f.asset.state = 'new'; },
    (f) => { f.asset.url = 'https://other.example/promotion.json'; },
    (f) => { f.routes.set(`${base}/git/ref/tags/v0.0.1`, () => ({ json: { object: { type: 'commit', sha: recoveryCommit } } })); },
    (f) => { f.live.set('/version.json', [{ bytes: stableJSON({ ...f.version, dirty: true }) }]); },
    (f) => { f.promotion.siteArchiveSha256 = 'not-a-digest'; f.seal(); },
  ]) await t.test(mutate.toString(), async (sub) => {
    const f = fixture(sub); mutate(f);
    await assert.rejects(release.fetchCurrent(f.policy));
  });
});

test('postdeploy compares the supplied manifest with the authenticated published manifest', async (t) => {
  const f = fixture(t), expected = { ...f.promotion, siteArchiveSha256: '0'.repeat(64) };
  await assert.rejects(release.verifyDeployment(f.policy, expected, { deploymentSha: commit, attempts: 1, intervalMs: 0 }));
});

test('postdeploy retries temporary HTTP failures and stale version bytes before succeeding', async (t) => {
  const f = fixture(t);
  f.live.set('/version.json', [{ status: 404 }, { status: 503 }, { bytes: stableJSON({ ...f.version, docsCommit: recoveryCommit }) }, { bytes: f.versionBytes }]);
  await verify(f, { attempts: 4 });
  assert.equal(f.counts.get('/version.json'), 4);
});

test('postdeploy retries transient network errors and delayed representative routes', async (t) => {
  const f = fixture(t);
  f.live.set('/version.json', [{ error: 'temporary DNS error' }, { bytes: f.versionBytes }]);
  f.live.set('/docs/cli/', [{ status: 503 }, { bytes: '<html>Released CLI</html>\n' }]);
  await verify(f, { attempts: 4 });
  assert.equal(f.counts.get('/docs/cli/'), 2);
});

test('postdeploy retries stale representative route bytes but never accepts HTTP 200 alone', async (t) => {
  const f = fixture(t);
  f.live.set('/docs/spec/', [{ bytes: '<html>Old spec</html>\n' }]);
  await assert.rejects(verify(f, { attempts: 2 }), /spec|digest|hash|identity|timed out|attempt/i);
  assert.equal(f.counts.get('/docs/spec/'), 2);
});

test('postdeploy fails after its bounded attempts when live identity never converges', async (t) => {
  const f = fixture(t); f.live.set('/version.json', [{ bytes: stableJSON({ ...f.version, docsCommit: recoveryCommit }) }]);
  await assert.rejects(verify(f, { attempts: 2 }), /version|identity|digest|hash|timed out|attempt/i);
  assert.equal(f.counts.get('/version.json'), 2);
});

test('authentication failure is not retried as HTTP propagation', async (t) => {
  const f = fixture(t); f.archive.immutable = false;
  await assert.rejects(verify(f));
  assert.equal(f.calls.filter((url) => url.endsWith('/releases/tags/v0.0.1')).length, 1);
});

test('promotion CLI persists the checked recovery action and emits no plan on a stale baseline', async (t) => {
  const root = await mkdtemp(join(tmpdir(), 'l-promotion-output-'));
  t.after(() => rm(root, { recursive: true, force: true }));
  const manifest = { siteArchiveSha256: 'f'.repeat(64), docsCommit: commit, expectedCurrentDigest: null };
  const manifestPath = join(root, 'promotion.json'), currentPath = join(root, 'current.json');
  await writeFile(manifestPath, stableJSON(manifest));
  for (const [label, current, action] of [
    ['empty', null, 'deploy'],
    ['already-live', manifest, 'verify'],
    ['stale', { ...manifest, siteArchiveSha256: 'e'.repeat(64) }, null],
  ]) {
    const output = join(root, `${label}.json`);
    await writeFile(currentPath, stableJSON(current));
    const result = spawnSync(process.execPath, [script, 'guard-promotion', '--manifest', manifestPath,
      '--current', currentPath, '--mode', 'retry', '--expected-current', 'null', '--root', root, '--output', output],
    { encoding: 'utf8', timeout: 10000, env: { ...process.env, NODE_OPTIONS: '' } });
    assert.equal(result.error, undefined, result.error?.message);
    if (action) {
      assert.equal(result.status, 0, result.stderr);
      const plan = JSON.parse(await readFile(output, 'utf8'));
      assert.equal(plan.action, action);
      assert.deepEqual(plan, JSON.parse(result.stdout));
    } else {
      assert.equal(result.status, 1);
      assert.match(result.stderr, /Stale promotion/);
      await assert.rejects(readFile(output), { code: 'ENOENT' });
    }
  }
});
