// Release authority and GitHub publication. No command publishes without the
// committed policy switch, trusted workflow identity, main ref and human roles.
import assert from 'node:assert/strict';
import { createHash } from 'node:crypto';
import { readFile, writeFile, appendFile } from 'node:fs/promises';
import { execFileSync } from 'node:child_process';
import { resolve, basename } from 'node:path';
import { fileURLToPath } from 'node:url';

const REPO = /^[A-Za-z0-9_.-]+\/[A-Za-z0-9_.-]+$/;
const SHA = /^[a-f0-9]{40}$/;
const VERSION = /^(0|[1-9]\d*)\.(0|[1-9]\d*)\.(0|[1-9]\d*)$/;
const TARGETS = ['darwin-amd64', 'darwin-arm64', 'linux-amd64', 'linux-arm64'];
const PATHS = ['.github/workflows/release.yml', '.github/workflows/docs-release-pages.yml', '.github/workflows/docs-recover-pages.yml'];
const digest = (bytes) => createHash('sha256').update(bytes).digest('hex');
const git = (...args) => execFileSync('git', args, { encoding: 'utf8' }).trim();

export function requiresMacOSSigning(version) {
  assert(VERSION.test(version), 'Stable VERSION required for signing policy');
  return version.split('.')[0] !== '0';
}

export function verifyPublicationSigning(manifest) {
  const required = requiresMacOSSigning(manifest.languageVersion);
  for (const target of TARGETS) {
    const artifact = manifest.artifacts.find((item) => item.target === target);
    assert(artifact, `Missing native target ${target}`);
    if (target.startsWith('darwin-')) {
      assert(['unsigned', 'developer-id'].includes(artifact.signing), `Unsupported macOS signing status: ${target}`);
      if (required) assert(artifact.signing === 'developer-id', `Signing and notarization are required from 1.0.0: ${target}`);
    }
  }
}

export async function githubAPI(path, { method = 'GET', body, token = process.env.GITHUB_TOKEN, allow404 = false } = {}) {
  assert(path.startsWith('/') && !path.startsWith('//'), 'Expected GitHub API path');
  assert(token, 'Missing GitHub credential');
  const response = await fetch(`https://api.github.com${path}`, {
    method, headers: { Authorization: `Bearer ${token}`, Accept: 'application/vnd.github+json', 'X-GitHub-Api-Version': '2022-11-28', ...(body === undefined ? {} : { 'Content-Type': 'application/json' }) },
    body: body === undefined ? undefined : JSON.stringify(body), redirect: 'error', signal: AbortSignal.timeout(60000),
  });
  if (allow404 && response.status === 404) return null;
  // Do not echo response bodies: third-party responses can include credentials.
  assert(response.ok, `GitHub ${method} ${path} failed (${response.status})`);
  return response.status === 204 ? null : response.json();
}

export function checkContext(policy, env, publication) {
  assert(REPO.test(policy.repository) && env.GITHUB_REPOSITORY === policy.repository, 'Wrong source repository');
  assert(policy.productionBranch === 'main' && env.GITHUB_REF === 'refs/heads/main', 'Release requires main');
  assert(env.GITHUB_EVENT_NAME === 'workflow_dispatch', 'Only manual dispatch is trusted');
  assert(SHA.test(env.GITHUB_SHA), 'Missing immutable dispatch SHA');
  for (const key of ['GITHUB_RUN_ID', 'GITHUB_RUN_ATTEMPT']) assert(/^[1-9]\d*$/.test(env[key]), `Missing ${key}`);
  for (const key of ['GITHUB_ACTOR', 'GITHUB_TRIGGERING_ACTOR']) assert(/^[A-Za-z0-9-]+$/.test(env[key]), `Missing human ${key}`);
  assert.deepEqual([...policy.targets].sort(), TARGETS, 'Release target policy changed');
  if (publication) {
    assert(policy.publicationEnabled === true, 'Publication is disabled by committed policy');
    assert(Number.isSafeInteger(policy.nativePublisher?.workflowId) && policy.nativePublisher.workflowId > 0 && policy.nativePublisher.workflowPath === PATHS[0], 'Native publisher is not registered');
  }
}

export async function guard({ policy, publication = true, env = process.env, api = githubAPI }) {
  checkContext(policy, env, publication);
  const users = [...new Set([env.GITHUB_ACTOR, env.GITHUB_TRIGGERING_ACTOR])];
  for (const user of users) {
    const access = await api(`/repos/${policy.repository}/collaborators/${encodeURIComponent(user)}/permission`);
    assert(['maintain', 'admin'].includes(access.role_name) && access.user?.type === 'User', `Release requires Maintain or Admin access: ${user}`);
  }
  const run = await api(`/repos/${policy.repository}/actions/runs/${env.GITHUB_RUN_ID}/attempts/${env.GITHUB_RUN_ATTEMPT}`);
  assert(run.event === 'workflow_dispatch' && run.head_branch === 'main' && run.head_sha === env.GITHUB_SHA && run.run_attempt === Number(env.GITHUB_RUN_ATTEMPT), 'Run identity differs from dispatch');
  assert(run.repository?.full_name === policy.repository && PATHS.includes(run.path), 'Unapproved release workflow');
  if (publication && run.path === PATHS[0]) assert(run.workflow_id === policy.nativePublisher.workflowId, 'Publisher workflow ID mismatch');
  if (publication) {
    const repository = await api(`/repos/${policy.repository}`);
    assert(repository.full_name === policy.repository && repository.private === false, 'Release publication requires this source repository to be public');
  }
  return run;
}

export async function currentJobId(policy, name, api = githubAPI) {
  const matches = [];
  for (let page = 1; ; page++) {
    const response = await api(`/repos/${policy.repository}/actions/runs/${process.env.GITHUB_RUN_ID}/attempts/${process.env.GITHUB_RUN_ATTEMPT}/jobs?per_page=100&page=${page}`);
    matches.push(...response.jobs.filter((job) => job.name === name && job.status === 'in_progress'));
    if (response.jobs.length < 100) break;
  }
  assert(matches.length === 1 && Number.isSafeInteger(matches[0].id), 'Cannot uniquely identify current publishing job');
  return matches[0].id;
}

export async function ensureTag(repo, tag, commit, token, api = githubAPI) {
  const base = `/repos/${repo}`;
  let ref = await api(`${base}/git/ref/tags/${encodeURIComponent(tag)}`, { token, allow404: true });
  if (!ref) ref = await api(`${base}/git/refs`, { method: 'POST', token, body: { ref: `refs/tags/${tag}`, sha: commit } });
  let object = ref.object;
  for (let depth = 0; object.type === 'tag' && depth < 8; depth++) object = (await api(`${base}/git/tags/${object.sha}`, { token })).object;
  assert(object.type === 'commit' && object.sha === commit, 'Existing source tag points to different bytes');
}

async function downloadAuthenticated(url, token) {
  assert(url.startsWith('https://api.github.com/'), 'Unexpected authenticated download URL');
  let response = await fetch(url, { headers: { Authorization: `Bearer ${token}`, Accept: 'application/octet-stream' }, redirect: 'manual', signal: AbortSignal.timeout(60000) });
  if ([301, 302, 303, 307, 308].includes(response.status)) {
    const target = new URL(response.headers.get('location'));
    assert(target.protocol === 'https:' && (target.hostname.endsWith('.blob.core.windows.net') || target.hostname === 'release-assets.githubusercontent.com' || target.hostname.endsWith('.actions.githubusercontent.com')), 'Unexpected artifact storage redirect');
    response = await fetch(target, { redirect: 'error', signal: AbortSignal.timeout(60000) });
  }
  assert(response.ok, `Artifact download failed (${response.status})`);
  return Buffer.from(await response.arrayBuffer());
}

async function ensureDraft(repo, tag, commit, token, api = githubAPI) {
  const base = `/repos/${repo}`;
  let release = await api(`${base}/releases/tags/${encodeURIComponent(tag)}`, { token, allow404: true });
  // The tag endpoint can omit drafts; list drafts before creating another one.
  if (!release) for (let page = 1; ; page++) {
    const releases = await api(`${base}/releases?per_page=100&page=${page}`, { token });
    release = releases.find((item) => item.tag_name === tag);
    if (release || releases.length < 100) break;
  }
  if (!release) release = await api(`${base}/releases`, { method: 'POST', token, body: { tag_name: tag, target_commitish: commit, name: `L ${tag}`, draft: true, prerelease: false, body: `L ${tag}. See the attached installation instructions, checksums, and release manifest.` } });
  assert(!release.prerelease, 'Existing release is a prerelease');
  return release;
}

export async function uploadOnce(repo, release, name, bytes, token, api = githubAPI, upload = fetch) {
  assert(name === basename(name) && /^[A-Za-z0-9_.-]+$/.test(name), 'Invalid release asset name');
  const assets = [];
  for (let page = 1; ; page++) {
    const batch = await api(`/repos/${repo}/releases/${release.id}/assets?per_page=100&page=${page}`, { token });
    assets.push(...batch); if (batch.length < 100) break;
  }
  const expected = `sha256:${digest(bytes)}`, existing = assets.find((asset) => asset.name === name);
  if (existing) {
    assert(existing.state === 'uploaded' && existing.digest === expected, `Existing asset differs: ${name}`);
    return;
  }
  assert(release.draft, `Cannot append missing native asset to a published release: ${name}`);
  const url = `https://uploads.github.com/repos/${repo}/releases/${release.id}/assets?name=${encodeURIComponent(name)}`;
  const response = await upload(url, { method: 'POST', headers: { Authorization: `Bearer ${token}`, 'Content-Type': 'application/octet-stream' }, body: bytes, redirect: 'error', signal: AbortSignal.timeout(120000) });
  assert(response.ok, `Asset upload failed: ${name} (${response.status})`);
  const asset = await response.json();
  assert(asset.state === 'uploaded' && asset.digest === expected, `Uploaded asset digest mismatch: ${name}`);
}

// The caller has authenticated the producer and verified every file against the
// retained proof. Keep the remote transaction separate so retries can be tested
// without publishing, downloading credentials, or replacing existing asset bytes.
export async function publishVerifiedAssets(repo, manifest, files, { token = process.env.GITHUB_TOKEN, api = githubAPI, upload = fetch } = {}) {
  assert(repo === manifest.repository && SHA.test(manifest.languageCommit) && VERSION.test(manifest.languageVersion) && manifest.tag === `v${manifest.languageVersion}`, 'Release destination or source identity differs');
  const repository = await api(`/repos/${repo}`, { token });
  assert(repository.full_name === repo && repository.private === false, 'Release publication requires this source repository to be public');
  await ensureTag(repo, manifest.tag, manifest.languageCommit, token, api);
  const release = await ensureDraft(repo, manifest.tag, manifest.languageCommit, token, api);
  for (const [name, bytes] of files) await uploadOnce(repo, release, name, bytes, token, api, upload);
  // Complete one draft before exposure. Retries reuse its original matching
  // assets; neither a source tag nor a published native asset is overwritten.
  if (release.draft) await api(`/repos/${repo}/releases/${release.id}`, { method: 'PATCH', token, body: { draft: false, make_latest: 'true' } });
  return release;
}

async function publish(policy, options) {
  await guard({ policy });
  const { verifyNativeArchive, verifyNativeEvidence } = await import('../../website/scripts/release.mjs');
  const nativeDir = resolve(options.native), manifestPath = resolve(nativeDir, 'native-release.json');
  const manifest = JSON.parse(await readFile(manifestPath));
  assert(manifest.schemaVersion === 2 && manifest.repository === policy.repository && manifest.languageCommit === process.env.GITHUB_SHA, 'Native identity mismatch');
  assert(manifest.producer.runId === Number(process.env.GITHUB_RUN_ID) && manifest.producer.runAttempt <= Number(process.env.GITHUB_RUN_ATTEMPT), 'Native producer attempt mismatch');
  // A failed publication can reuse a successful sealing job from the original
  // attempt. Never rewrite its producer identity or repackage its archive.
  const producer = await githubAPI(`/repos/${policy.repository}/actions/jobs/${manifest.producer.jobId}`);
  assert(producer.name === 'Seal native release' && producer.run_id === manifest.producer.runId && producer.run_attempt === manifest.producer.runAttempt && producer.head_sha === manifest.languageCommit && producer.status === 'completed' && producer.conclusion === 'success', 'Native sealing job did not complete successfully');
  assert(VERSION.test(manifest.languageVersion) && manifest.tag === `v${manifest.languageVersion}`, 'Stable VERSION required');
  verifyPublicationSigning(manifest);
  await verifyNativeArchive(options.archive, nativeDir, manifestPath);
  await verifyNativeEvidence(manifest, nativeDir);
  const artifact = await githubAPI(`/repos/${policy.repository}/actions/artifacts/${options['artifact-id']}`);
  const archive = await readFile(options.archive);
  assert(artifact.name === `l-native-release-${manifest.languageCommit}` && artifact.workflow_run?.id === manifest.producer.runId && artifact.workflow_run.head_sha === manifest.languageCommit && artifact.digest === `sha256:${digest(archive)}`, 'Native proof archive authority mismatch');
  const publicBytes = await readFile(resolve(nativeDir, 'public-release.json')), publicManifest = JSON.parse(publicBytes);
  assert(publicManifest.sourceRepository === policy.repository && publicManifest.sourceCommit === manifest.languageCommit && publicManifest.version === manifest.languageVersion && publicManifest.tag === manifest.tag, 'Public provenance mismatch');
  assert(publicManifest.artifacts.length === 5, 'Public release must contain four packages and installer');
  const distributables = [...publicManifest.artifacts, ...['public-release.json', 'checksums.txt'].map((name) => {
    const item = manifest.artifacts.find((entry) => entry.path === name); assert(item, `Missing public asset ${name}`); return { name, sha256: item.sha256 };
  })];
  const files = new Map();
  for (const item of distributables) {
    assert(item.name === basename(item.name) && !files.has(item.name), 'Unsafe or duplicate public asset');
    const bytes = await readFile(resolve(nativeDir, item.name));
    assert(digest(bytes) === item.sha256 && manifest.artifacts.some((entry) => entry.path === item.name && entry.sha256 === item.sha256), 'Public asset is not bound to native proof');
    files.set(item.name, bytes);
  }
  assert(manifest.artifacts.length === files.size && manifest.artifacts.every((item) => item.path === item.releaseAsset && files.has(item.path)), 'Release asset inventory differs from verified downloads');
  for (const [name, bytes] of [[`${artifact.name}.json`, await readFile(manifestPath)], [`${artifact.name}.zip`, archive]]) {
    assert(!files.has(name), 'Native proof asset collides with a download');
    files.set(name, bytes);
  }
  await publishVerifiedAssets(policy.repository, manifest, files);
}

function args(values) {
  const result = {};
  for (let i = 0; i < values.length; i += 2) {
    assert(values[i].startsWith('--') && values[i + 1] !== undefined && !values[i + 1].startsWith('--'), 'Expected named arguments');
    assert(!Object.hasOwn(result, values[i].slice(2)), 'Duplicate argument');
    result[values[i].slice(2)] = values[i + 1];
  }
  return result;
}

async function main() {
  const [command, ...values] = process.argv.slice(2), options = args(values);
  const policy = JSON.parse(await readFile(options.policy || 'website/ci-templates/release-policy.json'));
  if (command === 'guard' || command === 'preflight') {
    assert(['true', 'false'].includes(options.publication ?? 'true'), 'publication must be true or false');
    await guard({ policy, publication: options.publication !== 'false' });
    assert(git('rev-parse', 'HEAD') === process.env.GITHUB_SHA && git('status', '--porcelain') === '', 'Checkout must be clean and match dispatch SHA');
    if (command === 'preflight') {
      const version = (await readFile('VERSION', 'utf8')).trim();
      assert(VERSION.test(version) && options.version === version, 'Version confirmation must match committed stable VERSION');
      assert(!options.publication || options.publication === 'false' || policy.nativePublisher, 'Publisher not registered');
      const publication = options.publication !== 'false';
      const signMacOS = publication && requiresMacOSSigning(version);
      if (process.env.GITHUB_OUTPUT) await appendFile(process.env.GITHUB_OUTPUT, `version=${version}\nsha=${process.env.GITHUB_SHA}\npublish=${publication}\nsign_macos=${signMacOS}\n`);
    }
    console.log(options.publication === 'false' ? 'Authorized manual dry run; publication remains disabled.' : 'Authorized manual production operation.');
  } else if (command === 'job-id') {
    assert(['true', 'false'].includes(options.publication), 'publication must be true or false');
    await guard({ policy, publication: options.publication === 'true' });
    const id = await currentJobId(policy, 'Seal native release');
    assert(process.env.GITHUB_OUTPUT, 'Missing Actions output file');
    await appendFile(process.env.GITHUB_OUTPUT, `job_id=${id}\n`);
  } else if (command === 'download-proof') {
    await guard({ policy });
    assert(/^[1-9]\d*$/.test(options['artifact-id']), 'Invalid artifact ID');
    const artifact = await githubAPI(`/repos/${policy.repository}/actions/artifacts/${options['artifact-id']}`);
    const bytes = await downloadAuthenticated(`https://api.github.com/repos/${policy.repository}/actions/artifacts/${options['artifact-id']}/zip`, process.env.GITHUB_TOKEN);
    assert(artifact.digest === `sha256:${digest(bytes)}`, 'Downloaded Actions artifact digest mismatch');
    await writeFile(options.output, bytes, { flag: 'wx' });
  } else if (command === 'publish') await publish(policy, options);
  else throw new Error('Expected guard, preflight, job-id, download-proof, or publish');
}

if (process.argv[1] && resolve(process.argv[1]) === fileURLToPath(import.meta.url)) main().catch((error) => { console.error(`Release refused: ${error.message}`); process.exitCode = 1; });
