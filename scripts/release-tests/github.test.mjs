import test from 'node:test';
import assert from 'node:assert/strict';
import { createHash } from 'node:crypto';
import { checkContext, guard, ensureTag, uploadOnce, requiresMacOSSigning, verifyPublicationSigning, publishVerifiedAssets, publishAssetSet } from '../release/github.mjs';

const sha = 'a'.repeat(40);
const policy = { repository: 'example/l', productionBranch: 'main', publicationEnabled: true, nativePublisher: { workflowId: 42, workflowPath: '.github/workflows/release.yml' }, targets: ['darwin-arm64', 'darwin-amd64', 'linux-arm64', 'linux-amd64'] };
const env = { GITHUB_REPOSITORY: 'example/l', GITHUB_REF: 'refs/heads/main', GITHUB_SHA: sha, GITHUB_EVENT_NAME: 'workflow_dispatch', GITHUB_RUN_ID: '12', GITHUB_RUN_ATTEMPT: '2', GITHUB_ACTOR: 'maintainer', GITHUB_TRIGGERING_ACTOR: 'owner' };
const run = { repository: { full_name: 'example/l' }, event: 'workflow_dispatch', head_branch: 'main', head_sha: sha, run_attempt: 2, workflow_id: 42, path: '.github/workflows/release.yml' };

test('Apple signing starts at stable 1.0.0 and remains required for later major versions', () => {
  for (const version of ['0.0.1', '0.9.9', '0.99.999']) assert.equal(requiresMacOSSigning(version), false);
  for (const version of ['1.0.0', '1.0.1', '1.9.0', '2.0.0', '10.0.0']) assert.equal(requiresMacOSSigning(version), true);
  for (const version of ['', 'v1.0.0', '01.0.0', '1.0', '1.0.0-rc.1']) assert.throws(() => requiresMacOSSigning(version), /Stable VERSION/);
});

test('publication permits unsigned pre-1.0 macOS packages but rejects them from 1.0.0', () => {
  const manifest = (version) => ({ languageVersion: version, artifacts: policy.targets.map((target) => ({ target, signing: 'unsigned' })) });
  assert.doesNotThrow(() => verifyPublicationSigning(manifest('0.0.1')));
  assert.doesNotThrow(() => verifyPublicationSigning(manifest('0.99.999')));
  for (const version of ['1.0.0', '1.0.1', '2.0.0', '10.0.0']) {
    const signed = manifest(version);
    for (const item of signed.artifacts) if (item.target.startsWith('darwin-')) item.signing = 'developer-id';
    assert.doesNotThrow(() => verifyPublicationSigning(signed));
    for (const target of ['darwin-arm64', 'darwin-amd64']) {
      const unsigned = structuredClone(signed);
      unsigned.artifacts.find((item) => item.target === target).signing = 'unsigned';
      assert.throws(() => verifyPublicationSigning(unsigned), /Signing and notarization are required from 1.0.0/);
    }
  }
  for (const signing of [undefined, 'ad-hoc']) {
    const invalid = manifest('0.0.1');
    invalid.artifacts.find((item) => item.target === 'darwin-arm64').signing = signing;
    assert.throws(() => verifyPublicationSigning(invalid), /Unsupported macOS signing status/);
  }
  const incomplete = manifest('0.0.1');
  incomplete.artifacts.pop();
  assert.throws(() => verifyPublicationSigning(incomplete), /Missing native target/);
});

test('committed disabled policy refuses production before any API or credential use', async () => {
  let calls = 0;
  await assert.rejects(guard({ policy: { ...policy, publicationEnabled: false, nativePublisher: null }, env, api: async () => { calls++; } }), /Publication is disabled/);
  assert.equal(calls, 0);
  assert.doesNotThrow(() => checkContext({ ...policy, publicationEnabled: false, nativePublisher: null }, env, false));
  assert.doesNotThrow(() => checkContext(policy, env, true));
  for (const change of [{ nativePublisher: null }, { targets: policy.targets.slice(1) }]) {
    assert.throws(() => checkContext({ ...policy, ...change }, env, true));
  }
});

test('release rejects tags, other branches, automatic events, forks and missing rerun actors', () => {
  for (const change of [{ GITHUB_REF: 'refs/heads/topic' }, { GITHUB_REF: 'refs/tags/main' }, { GITHUB_EVENT_NAME: 'push' }, { GITHUB_EVENT_NAME: 'pull_request' }, { GITHUB_REPOSITORY: 'fork/l' }, { GITHUB_TRIGGERING_ACTOR: '' }, { GITHUB_SHA: 'main' }]) {
    assert.throws(() => checkContext(policy, { ...env, ...change }, true));
  }
});

test('authorization distinguishes Maintain from Write and rechecks the rerun user', async () => {
  const calls = [];
  const api = async (path) => { calls.push(path); return path.includes('/collaborators/') ? { role_name: path.includes('/owner/') ? 'admin' : 'maintain', permission: 'write', user: { type: 'User' } } : path === '/repos/example/l' ? { full_name: 'example/l', private: false } : run; };
  await guard({ policy, env, api });
  assert.equal(calls.filter((path) => path.includes('/permission')).length, 2);
  await assert.rejects(guard({ policy, env, api: async (path) => path.includes('/owner/') ? { role_name: 'write', permission: 'write', user: { type: 'User' } } : api(path) }), /Maintain or Admin/);
  await assert.rejects(guard({ policy, env, api: async (path) => path.includes('/permission') ? { role_name: 'admin', user: { type: 'Bot' } } : run }), /Maintain or Admin/);
});

test('production requires the source repository to be public while dry runs remain available privately', async () => {
  const calls = [];
  const api = async (path) => {
    calls.push(path);
    if (path.includes('/collaborators/')) return { role_name: 'maintain', user: { type: 'User' } };
    if (path === '/repos/example/l') return { full_name: 'example/l', private: true };
    return run;
  };
  await assert.rejects(guard({ policy, env, api }), /source repository to be public/);
  assert(calls.includes('/repos/example/l'));
  calls.length = 0;
  await guard({ policy: { ...policy, publicationEnabled: false, nativePublisher: null }, publication: false, env, api });
  assert(!calls.includes('/repos/example/l'));
});

test('authorization binds actual run SHA, event, workflow ID and path', async () => {
  for (const change of [{ head_sha: 'b'.repeat(40) }, { head_branch: 'other' }, { event: 'push' }, { run_attempt: 1 }, { workflow_id: 43 }, { path: '.github/workflows/unreviewed.yml' }]) {
    await assert.rejects(guard({ policy, env, api: async (path) => path.includes('/permission') ? { role_name: 'maintain', user: { type: 'User' } } : { ...run, ...change } }));
  }
});

test('source tags are never moved and annotated tags must resolve to exact source', async () => {
  const calls = [];
  await ensureTag('example/l', 'v0.0.1', sha, 'fixture', async (path, options) => {
    calls.push([path, options]);
    return path.endsWith('/git/ref/tags/v0.0.1') ? { object: { type: 'tag', sha: 'b'.repeat(40) } } : { object: { type: 'commit', sha } };
  });
  assert(calls.every(([, options]) => !options.method));
  await assert.rejects(ensureTag('example/l', 'v0.0.1', sha, 'fixture', async () => ({ object: { type: 'commit', sha: 'c'.repeat(40) } })), /different bytes/);
});

test('asset retries accept matching bytes but never replace or append to published native release', async () => {
  const bytes = Buffer.from('real fixture bytes'), hash = createHash('sha256').update(bytes).digest('hex');
  let uploaded = false;
  const upload = async () => { uploaded = true; throw Error('must not upload'); };
  await uploadOnce('example/l', { id: 1, draft: false }, 'l.tar.gz', bytes, 'fixture', async () => [{ name: 'l.tar.gz', digest: `sha256:${hash}`, state: 'uploaded' }], upload);
  await assert.rejects(uploadOnce('example/l', { id: 1, draft: true }, 'l.tar.gz', bytes, 'fixture', async () => [{ name: 'l.tar.gz', digest: `sha256:${'0'.repeat(64)}`, state: 'uploaded' }], upload), /differs/);
  await assert.rejects(uploadOnce('example/l', { id: 1, draft: false }, 'l.tar.gz', bytes, 'fixture', async () => [], upload), /Cannot append/);
  assert.equal(uploaded, false);
});

function publicationFixture(snapshot = false) {
  const manifest = { repository: 'example/l', languageCommit: sha, languageVersion: '0.0.1', tag: 'v0.0.1' };
  const docsCommit = snapshot ? 'b'.repeat(40) : sha;
  const identity = snapshot
    ? { tag: `docs-v0.0.1-${docsCommit}`, commit: docsCommit, name: 'Documentation snapshot', prerelease: true, latest: false, body: 'Documentation for L v0.0.1.' }
    : { tag: manifest.tag, commit: sha, prerelease: false, latest: true };
  const names = [
    ...(snapshot ? [] : [...policy.targets.map((target) => `l_0.0.1_${target.replace('-', '_')}.tar.gz`),
      'install-release.sh', 'checksums.txt', 'public-release.json', `l-native-release-${sha}.json`, `l-native-release-${sha}.zip`]),
    ...['tar.gz', 'promotion.json', 'native.json'].map((suffix) => `docs-${docsCommit}.${suffix}`),
  ];
  const files = new Map(names.map((name) => [name, Buffer.from(`verified fixture ${name}`)]));
  const state = { tag: null, release: null, assets: [], calls: [], uploaded: [], failUpload: null, private: false, immutableEnabled: true };
  const api = async (path, options = {}) => {
    assert.equal(options.token, 'automatic-token-fixture');
    assert(path.startsWith('/repos/example/l'), `Unexpected cross-repository or App request: ${path}`);
    state.calls.push({ path, ...options });
    const method = options.method || 'GET';
    if (path === '/repos/example/l') return { full_name: 'example/l', private: state.private };
    if (path === `/repos/example/l/git/ref/tags/${identity.tag}`) return state.tag;
    if (path === '/repos/example/l/git/refs' && method === 'POST') {
      assert.deepEqual(options.body, { ref: `refs/tags/${identity.tag}`, sha: identity.commit });
      assert.equal(state.tag, null);
      return state.tag = { object: { type: 'commit', sha: identity.commit } };
    }
    if (path === `/repos/example/l/releases/tags/${identity.tag}`) return state.release;
    if (path === '/repos/example/l/releases?per_page=100&page=1') return state.release ? [state.release] : [];
    if (path === '/repos/example/l/releases' && method === 'POST') {
      assert.equal(options.body.target_commitish, identity.commit);
      assert.equal(options.body.tag_name, identity.tag);
      assert.equal(options.body.draft, true);
      assert.equal(options.body.prerelease, identity.prerelease);
      assert.equal(state.release, null);
      return state.release = { id: 9, draft: true, prerelease: identity.prerelease, tag_name: identity.tag };
    }
    if (path === '/repos/example/l/releases/9/assets?per_page=100&page=1') return [...state.assets];
    if (path === '/repos/example/l/releases/9' && method === 'PATCH') {
      assert.deepEqual(options.body, { draft: false, make_latest: String(identity.latest) });
      assert.deepEqual(state.assets.map((asset) => asset.name).sort(), names.toSorted());
      state.release.draft = false;
      state.release.immutable = state.immutableEnabled;
      return state.release;
    }
    assert.fail(`Unexpected GitHub operation: ${method} ${path}`);
  };
  const upload = async (url, options) => {
    const parsed = new URL(url), name = parsed.searchParams.get('name');
    assert.equal(parsed.origin + parsed.pathname, 'https://uploads.github.com/repos/example/l/releases/9/assets');
    assert.equal(options.headers.Authorization, 'Bearer automatic-token-fixture');
    assert.equal(state.release.draft, true);
    if (name === state.failUpload) return { ok: false, status: 503 };
    assert.deepEqual(options.body, files.get(name));
    assert(!state.assets.some((asset) => asset.name === name));
    const asset = { name, state: 'uploaded', digest: `sha256:${createHash('sha256').update(options.body).digest('hex')}` };
    state.assets.push(asset);
    state.uploaded.push(name);
    return { ok: true, json: async () => asset };
  };
  return { manifest, identity, files, state, options: { token: 'automatic-token-fixture', api, upload } };
}

test('one source release receives every verified asset before publication using only its workflow token', async () => {
  const f = publicationFixture();
  await publishVerifiedAssets('example/l', f.manifest, f.files, f.options);
  assert.equal(f.state.release.draft, false);
  assert.equal(f.state.calls.filter((call) => call.path === '/repos/example/l/git/refs').length, 1);
  assert.equal(f.state.calls.filter((call) => call.path === '/repos/example/l/releases').length, 1);
  assert.equal(f.state.calls.filter((call) => call.method === 'PATCH').length, 1);
  assert.equal(f.state.uploaded.length, 12);
  const writes = f.state.calls.filter((call) => call.method).length;
  await publishVerifiedAssets('example/l', f.manifest, f.files, f.options);
  assert.equal(f.state.calls.filter((call) => call.method).length, writes);
  assert.equal(f.state.uploaded.length, 12);
});

test('a partial upload remains a draft and retries reuse matching original assets', async () => {
  const f = publicationFixture();
  f.state.failUpload = 'checksums.txt';
  await assert.rejects(publishVerifiedAssets('example/l', f.manifest, f.files, f.options), /Asset upload failed/);
  assert.equal(f.state.release.draft, true);
  assert.equal(f.state.calls.filter((call) => call.method === 'PATCH').length, 0);
  f.state.failUpload = null;
  await publishVerifiedAssets('example/l', f.manifest, f.files, f.options);
  assert.equal(f.state.release.draft, false);
  assert.equal(f.state.uploaded.length, 12);
  assert.equal(new Set(f.state.uploaded).size, 12);
});

test('publication rejects private repositories, moved source tags and changed retry bytes', async () => {
  const privateRepo = publicationFixture();
  privateRepo.state.private = true;
  await assert.rejects(publishVerifiedAssets('example/l', privateRepo.manifest, privateRepo.files, privateRepo.options), /source repository to be public/);
  assert(!privateRepo.state.calls.some((call) => call.method));
  const moved = publicationFixture();
  moved.state.tag = { object: { type: 'commit', sha: 'b'.repeat(40) } };
  await assert.rejects(publishVerifiedAssets('example/l', moved.manifest, moved.files, moved.options), /different bytes/);
  assert(!moved.state.calls.some((call) => call.method));
  const changed = publicationFixture();
  await publishVerifiedAssets('example/l', changed.manifest, changed.files, changed.options);
  changed.files.set('install-release.sh', Buffer.from('changed retry bytes'));
  await assert.rejects(publishVerifiedAssets('example/l', changed.manifest, changed.files, changed.options), /Existing asset differs/);
  assert.equal(changed.state.uploaded.length, 12);
});


test('initial publication refuses missing documentation before contacting GitHub', async () => {
  for (const suffix of ['tar.gz', 'promotion.json', 'native.json']) {
    const f = publicationFixture();
    f.files.delete(`docs-${sha}.${suffix}`);
    await assert.rejects(publishVerifiedAssets('example/l', f.manifest, f.files, f.options), /requires all sealed documentation assets/);
    assert.equal(f.state.calls.length, 0);
    assert.equal(f.state.uploaded.length, 0);
  }
});

test('a failed documentation upload keeps the combined native release private until retry completes', async () => {
  const f = publicationFixture();
  f.state.failUpload = `docs-${sha}.promotion.json`;
  await assert.rejects(publishVerifiedAssets('example/l', f.manifest, f.files, f.options), /Asset upload failed/);
  assert.equal(f.state.release.draft, true);
  assert.equal(f.state.uploaded.length, 10);
  assert.equal(f.state.calls.filter((call) => call.method === 'PATCH').length, 0);
  f.state.failUpload = null;
  await publishVerifiedAssets('example/l', f.manifest, f.files, f.options);
  assert.equal(f.state.release.draft, false);
  assert.equal(f.state.release.immutable, true);
  assert.equal(f.state.uploaded.length, 12);
  assert.equal(new Set(f.state.uploaded).size, 12);
});

test('unexpected draft assets prevent publication rather than being silently retained or deleted', async () => {
  const f = publicationFixture();
  f.state.tag = { object: { type: 'commit', sha } };
  f.state.release = { id: 9, draft: true, prerelease: false, tag_name: 'v0.0.1' };
  f.state.assets.push({ name: 'unexpected.txt', state: 'uploaded', digest: `sha256:${'0'.repeat(64)}` });
  await assert.rejects(publishVerifiedAssets('example/l', f.manifest, f.files, f.options), /Draft asset inventory differs/);
  assert.equal(f.state.release.draft, true);
  assert(!f.state.calls.some((call) => ['PATCH', 'DELETE'].includes(call.method)));
});

test('publication verifies GitHub immutability without administrator access and rejects mutable releases', async () => {
  const disabled = publicationFixture();
  disabled.state.immutableEnabled = false;
  await assert.rejects(publishVerifiedAssets('example/l', disabled.manifest, disabled.files, disabled.options), /Published release must be immutable/);
  assert(!disabled.state.calls.some((call) => call.path.endsWith('/immutable-releases')));
  const mutable = publicationFixture();
  await publishVerifiedAssets('example/l', mutable.manifest, mutable.files, mutable.options);
  mutable.state.release.immutable = false;
  const writes = mutable.state.calls.filter((call) => call.method).length;
  await assert.rejects(publishVerifiedAssets('example/l', mutable.manifest, mutable.files, mutable.options), /Existing published release is not immutable/);
  assert.equal(mutable.state.calls.filter((call) => call.method).length, writes);
});

test('documentation snapshots are immutable non-latest prereleases without native release mutation', async () => {
  const f = publicationFixture(true);
  await publishAssetSet('example/l', f.identity, f.files, f.options);
  assert.equal(f.state.release.tag_name, `docs-v0.0.1-${'b'.repeat(40)}`);
  assert.equal(f.state.tag.object.sha, 'b'.repeat(40));
  assert.equal(f.state.release.prerelease, true);
  assert.equal(f.state.release.draft, false);
  assert.equal(f.state.release.immutable, true);
  assert.equal(f.state.uploaded.length, 3);
  assert(f.state.uploaded.every((name) => name.startsWith(`docs-${'b'.repeat(40)}.`)));
  const publication = f.state.calls.filter((call) => call.method === 'PATCH');
  assert.equal(publication.length, 1);
  assert.deepEqual(publication[0].body, { draft: false, make_latest: 'false' });
  assert(!f.state.calls.some((call) => call.path.endsWith('/tags/v0.0.1')));
  assert.deepEqual(f.state.calls.filter((call) => call.path.endsWith('/git/refs')).map((call) => call.body), [{ ref: `refs/tags/${f.identity.tag}`, sha: f.identity.commit }]);
  const writes = f.state.calls.filter((call) => call.method).length;
  await publishAssetSet('example/l', f.identity, f.files, f.options);
  assert.equal(f.state.calls.filter((call) => call.method).length, writes);
  assert.equal(f.state.uploaded.length, 3);
});
