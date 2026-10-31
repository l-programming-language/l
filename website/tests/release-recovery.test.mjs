import test from 'node:test';
import assert from 'node:assert/strict';
import { mkdtemp, writeFile, readFile, rm } from 'node:fs/promises';
import { join } from 'node:path';
import { tmpdir } from 'node:os';
import { fileURLToPath } from 'node:url';
import { spawnSync } from 'node:child_process';
import { nativeTargets, sha256, stableJSON } from '../scripts/release.mjs';

const repository = 'l-programming-language/l', base = `/repos/${repository}`;
const languageCommit = 'a'.repeat(40), otherCommit = 'b'.repeat(40), hash = 'c'.repeat(64);
const script = fileURLToPath(new URL('../scripts/release.mjs', import.meta.url));

async function fixture(t, correction = false) {
  const root = await mkdtemp(join(tmpdir(), 'l-release-recovery-'));
  t.after(() => rm(root, { recursive: true, force: true }));
  const docsCommit = correction ? otherCommit : languageCommit;
  const policy = { publicationEnabled: true, repository, productionBranch: 'main', targets: nativeTargets,
    nativePublisher: { workflowId: 42, workflowPath: '.github/workflows/release.yml' } };
  const native = { schemaVersion: 2, repository, languageVersion: '0.0.1', tag: 'v0.0.1', languageCommit,
    sourceFingerprint: `sha256:${hash}`, releaseFingerprint: hash, dirty: false,
    producer: { runId: 10, runAttempt: 1, jobId: 11 },
    artifacts: nativeTargets.map((target) => ({ path: `${target}.tar.gz`, releaseAsset: `${target}.tar.gz`, target, compilerVersion: 'go1.27.3', sha256: sha256(target) })),
    validation: { command: 'scripts/l-check', targets: nativeTargets.map((target) => ({ target, smoke: `evidence/${target}/smoke.json`,
      stages: ['lint', 'acceptance', 'native'].map((stage) => ({ stage, path: `evidence/${target}/${stage}.json` })) })) } };
  const archive = Buffer.from('sealed documentation archive'), proof = Buffer.from('durable native proof archive');
  const promotion = { docsCommit, languageCommit, nativeManifestSha256: sha256(stableJSON(native)), siteArchiveSha256: sha256(archive) };
  const routes = {}, asset = (name, bytes) => {
    const url = `https://api.github.com${base}/releases/assets/${Object.keys(routes).length + 1}`;
    routes[new URL(url).pathname] = { bytes: Buffer.from(bytes).toString('base64') };
    return { name, url, state: 'uploaded', digest: `sha256:${sha256(bytes)}` };
  };
  const identity = asset(`l-native-release-${languageCommit}.json`, stableJSON(native));
  const durable = asset(`l-native-release-${languageCommit}.zip`, proof);
  const nativeRelease = { tag_name: native.tag, draft: false, prerelease: false, immutable: true,
    assets: [...native.artifacts.map((item) => ({ name: item.path, state: 'uploaded', digest: `sha256:${item.sha256}` })), identity, durable] };
  const docsAssets = [asset(`docs-${docsCommit}.tar.gz`, archive), asset(`docs-${docsCommit}.promotion.json`, stableJSON(promotion)), asset(`docs-${docsCommit}.native.json`, stableJSON(native))];
  const docsTag = correction ? `docs-v0.0.1-${docsCommit}` : native.tag;
  const docsRelease = correction ? { tag_name: docsTag, draft: false, prerelease: true, immutable: true, assets: docsAssets } : nativeRelease;
  if (!correction) nativeRelease.assets.push(...docsAssets);
  routes[`${base}/releases/tags/${native.tag}`] = { json: nativeRelease };
  routes[`${base}/releases/tags/${docsTag}`] = { json: docsRelease };
  routes[`${base}/git/ref/tags/${native.tag}`] = { json: { object: { type: 'commit', sha: languageCommit } } };
  routes[`${base}/git/ref/tags/${docsTag}`] = { json: { object: { type: 'commit', sha: docsCommit } } };
  routes[`${base}/actions/runs/10/attempts/1`] = { json: { id: 10, run_attempt: 1, workflow_id: 42, path: policy.nativePublisher.workflowPath,
    head_sha: languageCommit, head_branch: 'main', event: 'workflow_dispatch', status: 'completed' } };
  routes[`${base}/actions/jobs/11`] = { json: { id: 11, run_id: 10, run_attempt: 1, head_sha: languageCommit, status: 'completed', conclusion: 'success' } };
  const artifact = { name: `l-native-release-${languageCommit}`, digest: durable.digest, expired: false };
  routes[`${base}/actions/artifacts/20`] = { json: artifact };
  routes[`${base}/actions/artifacts/20/zip`] = { bytes: proof.toString('base64'), accept: 'application/vnd.github+json' };
  await writeFile(join(root, 'policy.json'), stableJSON(policy));
  await writeFile(join(root, 'VERSION'), '0.0.1\n');
  await writeFile(join(root, 'mock-fetch.mjs'), `
import assert from 'node:assert/strict';
import { readFileSync, appendFileSync } from 'node:fs';
const routes = JSON.parse(readFileSync(process.env.RELEASE_TEST_ROUTES));
globalThis.fetch = async (url, options = {}) => {
  const target = new URL(url), headers = new Headers(options.headers);
  assert.equal(target.origin, 'https://api.github.com');
  assert.equal(options.method || 'GET', 'GET');
  assert.equal(headers.get('Authorization'), 'Bearer recovery-fixture-token');
  appendFileSync(process.env.RELEASE_TEST_CALLS, JSON.stringify(target.pathname + target.search) + '\\n');
  const entry = routes[target.pathname + target.search];
  assert(entry, 'Unexpected API request: ' + target.pathname + target.search);
  assert.equal(headers.get('Accept'), entry.accept ?? (entry.bytes === undefined ? 'application/vnd.github+json' : 'application/octet-stream'));
  return new Response(entry.bytes === undefined ? JSON.stringify(entry.json) : Buffer.from(entry.bytes, 'base64'));
};
`);
  async function run(command, args) {
    const attempt = await mkdtemp(join(root, 'attempt-')), output = join(attempt, 'download');
    const callsPath = join(attempt, 'calls.jsonl'), routesPath = join(attempt, 'routes.json');
    await writeFile(routesPath, stableJSON(routes));
    await writeFile(callsPath, '');
    const result = spawnSync(process.execPath, ['--import', join(root, 'mock-fetch.mjs'), script, command,
      '--policy', join(root, 'policy.json'), '--output', output, ...args], {
      cwd: root, encoding: 'utf8', timeout: 10000,
      env: { ...process.env, NODE_OPTIONS: '', GITHUB_TOKEN: 'recovery-fixture-token', RELEASE_TEST_ROUTES: routesPath, RELEASE_TEST_CALLS: callsPath },
    });
    assert.equal(result.error, undefined, result.error?.message);
    const calls = (await readFile(callsPath, 'utf8')).trim().split('\n').filter(Boolean).map(JSON.parse);
    return { ...result, calls, output };
  }
  return { root, run, routes, native, docsCommit, docsTag, nativeRelease, docsRelease, identity, durable, docsAssets, archive, proof, artifact };
}

for (const correction of [false, true]) test(`recovery downloads ${correction ? 'snapshot' : 'initial'} docs and verifies the separate native authority`, async (t) => {
  const f = await fixture(t, correction);
  const result = await f.run('download-promotion', ['--tag', 'v0.0.1', '--docs-commit', f.docsCommit]);
  assert.equal(result.status, 0, result.stderr);
  assert.deepEqual(await readFile(join(result.output, 'site.tar.gz')), f.archive);
  assert.deepEqual(JSON.parse(await readFile(join(result.output, 'native-release.json'))), f.native);
  for (const path of [`${base}/releases/tags/v0.0.1`, `${base}/releases/tags/${f.docsTag}`, `${base}/git/ref/tags/v0.0.1`, `${base}/git/ref/tags/${f.docsTag}`, `${base}/actions/runs/10/attempts/1`, `${base}/actions/jobs/11`]) assert(result.calls.includes(path), path);
  for (const asset of f.docsAssets) assert(result.calls.includes(new URL(asset.url).pathname));
  assert(!result.calls.some((path) => path.includes('/actions/artifacts/')));
});

test('recovery rejects mutable archives, wrong tags, changed hashes and invalid native authority', async (t) => {
  for (const [mutate, expected] of [
    [(f) => { f.docsRelease.immutable = false; }, /published immutable archive release/],
    [(f) => { f.nativeRelease.immutable = false; }, /published immutable language release/],
    [(f) => { f.routes[`${base}/git/ref/tags/${f.docsTag}`].json.object.sha = languageCommit; }, /archive tag identifies a different revision/],
    [(f) => { f.routes[new URL(f.docsAssets[0].url).pathname].bytes = Buffer.from('changed archive').toString('base64'); }, /Recovery asset digest mismatch/],
    [(f) => { f.identity.digest = `sha256:${'0'.repeat(64)}`; }, /Native manifest is not the published release identity/],
    [(f) => { f.routes[`${base}/actions/runs/10/attempts/1`].json.workflow_id = 99; }, /Native producer is untrusted/],
    [(f) => { f.routes[`${base}/git/ref/tags/v0.0.1`].json.object.sha = 'd'.repeat(40); }, /Recovery identity does not match requested release/],
  ]) {
    const f = await fixture(t, true); mutate(f);
    const result = await f.run('download-promotion', ['--tag', 'v0.0.1', '--docs-commit', f.docsCommit]);
    assert.equal(result.status, 1);
    assert.match(result.stderr, expected);
  }
});

test('docs correction downloads durable native proof even when the Actions artifact has expired', async (t) => {
  const f = await fixture(t, true); f.artifact.expired = true;
  const result = await f.run('download-native', ['--artifact-id', '20', '--mode', 'docs-correction', '--root', f.root]);
  assert.equal(result.status, 0, result.stderr);
  assert.deepEqual(await readFile(result.output), f.proof);
  assert.deepEqual(result.calls, [`${base}/actions/artifacts/20`, `${base}/releases/tags/v0.0.1`, new URL(f.durable.url).pathname]);
});

test('pending native download requires an unexpired artifact and never queries a public release', async (t) => {
  const f = await fixture(t);
  const args = ['--artifact-id', '20', '--mode', 'release'];
  const fresh = await f.run('download-native', args);
  assert.equal(fresh.status, 0, fresh.stderr);
  assert.deepEqual(await readFile(fresh.output), f.proof);
  assert.deepEqual(fresh.calls, [`${base}/actions/artifacts/20`, `${base}/actions/artifacts/20/zip`]);
  f.artifact.expired = true;
  const expired = await f.run('download-native', args);
  assert.equal(expired.status, 1);
  assert.match(expired.stderr, /Native proof artifact has expired/);
  assert.deepEqual(expired.calls, [`${base}/actions/artifacts/20`]);
});

test('durable native downloads reject mutable releases and mismatched proof digests', async (t) => {
  for (const [mutate, expected] of [
    [(f) => { f.nativeRelease.immutable = false; }, /published immutable native release/],
    [(f) => { f.durable.digest = `sha256:${'0'.repeat(64)}`; }, /Durable native archive differs/],
    [(f) => { f.routes[new URL(f.durable.url).pathname].bytes = Buffer.from('changed proof').toString('base64'); }, /Downloaded native archive digest mismatch/],
  ]) {
    const f = await fixture(t); mutate(f);
    const result = await f.run('download-native', ['--artifact-id', '20', '--mode', 'docs-correction', '--root', f.root]);
    assert.equal(result.status, 1);
    assert.match(result.stderr, expected);
    assert(!result.calls.includes(`${base}/actions/artifacts/20/zip`));
  }
});
