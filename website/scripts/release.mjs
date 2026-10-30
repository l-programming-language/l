import { createHash } from 'node:crypto';
import { readFile, writeFile, lstat, realpath, readdir, mkdir } from 'node:fs/promises';
import { resolve, relative, sep, dirname } from 'node:path';
import { fileURLToPath } from 'node:url';
import { execFileSync } from 'node:child_process';
import { inflateRawSync } from 'node:zlib';

const SHA = /^[a-f0-9]{40}$/;
const DIGEST = /^[a-f0-9]{64}$/;
const FINGERPRINT = /^sha256:[a-f0-9]{64}$/;
const VERSION = /^(0|[1-9]\d*)\.(0|[1-9]\d*)\.(0|[1-9]\d*)(?:-[0-9A-Za-z]+(?:[.-][0-9A-Za-z]+)*)?$/;
const assert = (ok, message) => { if (!ok) throw new Error(message); };
export const sha256 = (bytes) => createHash('sha256').update(bytes).digest('hex');
export const stableJSON = (value) => JSON.stringify(value, null, 2) + '\n';

export const nativeTargets = ['darwin-arm64', 'darwin-amd64', 'linux-arm64', 'linux-amd64'];

// Local preparation is useful while publication is disabled. Every production
// entry point, including direct helper invocation, must cross this boundary.
export function assertPublicationPolicy(policy) {
  assert(policy.publicationEnabled === true, 'Publication is disabled by the committed release policy');
  assert(policy.productionBranch === 'main' && policy.repository === 'l-programming-language/l', 'Invalid production repository or branch');
  assert(Number.isSafeInteger(policy.nativePublisher?.workflowId) && policy.nativePublisher.workflowId > 0 && policy.nativePublisher.workflowPath === '.github/workflows/release.yml', 'Native release publisher is not configured; production activation is blocked');
  assert(Array.isArray(policy.targets) && [...policy.targets].sort().join(',') === [...nativeTargets].sort().join(','), 'Production requires the complete native target matrix');
  return policy;
}

function validationGroups(manifest) {
  return manifest.schemaVersion === 1 ? [{ stages: manifest.validation?.stages }] : manifest.validation?.targets;
}

function validateTargetPolicy(manifest, policy) {
  if (Object.hasOwn(policy, 'publicationEnabled')) assertPublicationPolicy(policy);
  if (manifest.schemaVersion === 2 || policy.targets) {
    assert(manifest.schemaVersion === 2, 'Production requires per-target native validation');
    assert(Array.isArray(policy.targets) && [...policy.targets].sort().join(',') === manifest.validation.targets.map((x) => x.target).sort().join(','), 'Native validation targets do not match release policy');
  }
}

// GitHub artifact ZIPs only: bounded, regular UTF-8 files with stored/deflate
// compression. ZIP64, encrypted entries, links and ambiguous paths fail closed.
export function readArtifactZip(bytes) {
  let end = -1;
  for (let i = bytes.length - 22; i >= Math.max(0, bytes.length - 65557); i--) if (bytes.readUInt32LE(i) === 0x06054b50 && i + 22 + bytes.readUInt16LE(i + 20) === bytes.length) { end = i; break; }
  assert(end >= 0, 'Artifact ZIP has no complete directory');
  const count = bytes.readUInt16LE(end + 10), size = bytes.readUInt32LE(end + 12), offset = bytes.readUInt32LE(end + 16);
  assert(bytes.readUInt32LE(end + 4) === 0 && bytes.readUInt16LE(end + 8) === count && count !== 65535 && offset + size === end, 'Multipart/ZIP64/ambiguous artifact ZIP rejected');
  const result = new Map(); let cursor = offset, total = 0;
  for (let i = 0; i < count; i++) {
    assert(cursor + 46 <= end && bytes.readUInt32LE(cursor) === 0x02014b50, 'Malformed ZIP central directory');
    const flags = bytes.readUInt16LE(cursor + 8), method = bytes.readUInt16LE(cursor + 10), packed = bytes.readUInt32LE(cursor + 20), unpacked = bytes.readUInt32LE(cursor + 24);
    const n = bytes.readUInt16LE(cursor + 28), extra = bytes.readUInt16LE(cursor + 30), comment = bytes.readUInt16LE(cursor + 32), attrs = bytes.readUInt32LE(cursor + 38), start = bytes.readUInt32LE(cursor + 42);
    assert(cursor + 46 + n + extra + comment <= end && !(flags & 1) && [0, 8].includes(method), 'Unsupported or truncated ZIP entry');
    const raw = bytes.subarray(cursor + 46, cursor + 46 + n), name = raw.toString('utf8');
    assert(Buffer.from(name).equals(raw) && ((attrs >>> 16) & 0o170000) !== 0o120000, 'Non-UTF-8 filename or symlink rejected');
    const directory = name.endsWith('/'); safeRelative(directory ? name.slice(0, -1) : name);
    assert(!result.has(name), 'Duplicate ZIP entry');
    assert(start + 30 <= offset && bytes.readUInt32LE(start) === 0x04034b50, 'Invalid ZIP local entry');
    const localName = bytes.readUInt16LE(start + 26), localExtra = bytes.readUInt16LE(start + 28), data = start + 30 + localName + localExtra;
    assert(bytes.subarray(start + 30, start + 30 + localName).equals(raw) && data + packed <= offset, 'ZIP local/central identity mismatch');
    total += unpacked;
    assert(total <= 1024 ** 3 && unpacked !== 0xffffffff && packed !== 0xffffffff, 'Artifact expands beyond 1 GiB or requires ZIP64');
    const body = method === 0 ? bytes.subarray(data, data + packed) : inflateRawSync(bytes.subarray(data, data + packed), { maxOutputLength: Math.max(1, unpacked) });
    assert(body.length === unpacked, 'ZIP size mismatch');
    if (!directory) result.set(name, body);
    cursor += 46 + n + extra + comment;
  }
  assert(cursor === end && result.has('native-release.json'), 'ZIP directory is incomplete or lacks native-release.json');
  return result;
}

export async function verifyNativeArchive(archive, root, manifestPath) {
  const files = readArtifactZip(await readFile(archive));
  assert((await readFile(manifestPath)).equals(files.get('native-release.json')), 'Native manifest is not from the verified archive');
  for (const [name, bytes] of files) assert((await readOwned(root, name)).equals(bytes), `Extracted archive differs: ${name}`);
  const actual = await directoryInventory(root);
  assert(actual.length === files.size, 'Unlisted files in native artifact directory');
}

export function safeRelative(value) {
  assert(typeof value === 'string' && value.length > 0 && !value.includes('\\') && !value.includes('\0') && !value.startsWith('/') && !/^[A-Za-z]:/.test(value), 'Expected a safe relative path');
  assert(value.split('/').every((part) => part !== '' && part !== '.' && part !== '..'), 'Path traversal or empty component rejected');
  return value;
}

export async function readOwned(root, name) {
  safeRelative(name);
  const base = await realpath(root);
  let current = base;
  for (const part of name.split('/')) {
    current = resolve(current, part);
    const info = await lstat(current);
    assert(!info.isSymbolicLink(), `Symlink rejected: ${name}`);
  }
  assert((await lstat(current)).isFile(), `Not a regular file: ${name}`);
  assert(!relative(base, await realpath(current)).startsWith(`..${sep}`), 'Escaped artifact root');
  return readFile(current);
}

export async function verifyInventory(root, inventory) {
  assert(Array.isArray(inventory) && inventory.length > 0, 'Artifact/evidence inventory must not be empty');
  const seen = new Set();
  for (const item of inventory) {
    safeRelative(item.path);
    assert(!seen.has(item.path), `Duplicate artifact: ${item.path}`);
    seen.add(item.path);
    assert(DIGEST.test(item.sha256), `Invalid digest: ${item.path}`);
    assert(sha256(await readOwned(root, item.path)) === item.sha256, `Digest mismatch: ${item.path}`);
  }
  return seen;
}

export function validateNativeShape(manifest) {
  assert([1, 2].includes(manifest.schemaVersion), 'Unsupported native manifest schema');
  assert(/^[\w.-]+\/[\w.-]+$/.test(manifest.repository), 'Invalid repository');
  assert(VERSION.test(manifest.languageVersion) && manifest.tag === `v${manifest.languageVersion}`, 'A real VERSION and matching release tag are required');
  assert(SHA.test(manifest.languageCommit) && FINGERPRINT.test(manifest.sourceFingerprint) && DIGEST.test(manifest.releaseFingerprint), 'Invalid native source identity');
  assert(manifest.dirty === false, 'Native release must identify a clean checkout');
  assert(['runId', 'runAttempt', 'jobId'].every((key) => Number.isSafeInteger(manifest.producer?.[key]) && manifest.producer[key] > 0), 'Missing native producer identities');
  assert(Array.isArray(manifest.artifacts) && manifest.artifacts.length > 0, 'No published native artifacts');
  for (const item of manifest.artifacts) {
    assert(/^[\w-]+$/.test(item.target) && typeof item.compilerVersion === 'string' && item.compilerVersion.length > 0, 'Native artifact requires target/compiler provenance');
    assert(item.releaseAsset === item.path && !item.path.includes('/'), 'Native release assets must have exact basename identities');
  }
  assert(manifest.validation?.command === 'scripts/l-check', 'Canonical language validation is required');
  const groups = validationGroups(manifest);
  assert(Array.isArray(groups) && groups.length > 0, 'Missing native target validation');
  const paths = new Set();
  for (const group of groups) {
    assert(Array.isArray(group.stages) && group.stages.length === 3, 'All three canonical validation stages are required');
    assert(new Set(group.stages.map((x) => x.stage)).size === 3 && ['lint', 'acceptance', 'native'].every((s) => group.stages.some((x) => x.stage === s)), 'Missing or duplicate language validation stage');
    for (const entry of group.stages) {
      safeRelative(entry.path);
      assert(!paths.has(entry.path), 'Validation stage evidence cannot be shared between targets');
      paths.add(entry.path);
    }
    if (manifest.schemaVersion === 2) {
      safeRelative(group.smoke);
      assert(!paths.has(group.smoke), 'Installer smoke evidence cannot be shared between targets or stages');
      paths.add(group.smoke);
    }
  }
  if (manifest.schemaVersion === 2) {
    assert(groups.length === nativeTargets.length && groups.map((x) => x.target).sort().join(',') === [...nativeTargets].sort().join(','), 'All four distinct native targets are required');
    for (const target of nativeTargets) assert(manifest.artifacts.filter((x) => x.target === target).length === 1, `Exactly one native artifact is required: ${target}`);
    assert(manifest.artifacts.every((x) => x.target === 'all' || nativeTargets.includes(x.target)), 'Unknown native artifact target');
  }
  return manifest;
}

// These observations must come from authenticated GitHub API calls in production.
// They are injectable only to unit-test policy, never as CLI JSON authority.
function verifyNativeProducer(manifest, policy, observations) {
  validateNativeShape(manifest);
  validateTargetPolicy(manifest, policy);
  const publisher = policy.nativePublisher;
  assert(publisher && Number.isSafeInteger(publisher.workflowId) && publisher.workflowPath, 'Native release publisher is not configured; production activation is blocked');
  const { run, job, artifact, downloadedArchiveSha256 } = observations;
  assert(run.repository?.full_name === manifest.repository && manifest.repository === policy.repository, 'Producer repository mismatch');
  assert(run.workflow_id === publisher.workflowId && run.path === publisher.workflowPath, 'Native producer is not the approved workflow identity');
  assert(run.id === manifest.producer.runId && run.run_attempt === manifest.producer.runAttempt && run.head_sha === manifest.languageCommit && run.head_branch === policy.productionBranch, 'Native producer revision/ref mismatch');
  assert((manifest.schemaVersion === 2 ? ['workflow_dispatch'] : ['push', 'workflow_dispatch']).includes(run.event), 'Untrusted native producer event');
  verifyProducerJob(manifest, run, job);
  assert(artifact.name === `l-native-release-${manifest.languageCommit}` && artifact.workflow_run?.id === run.id && artifact.workflow_run?.head_sha === manifest.languageCommit, 'Native artifact provenance is missing or stale');
  assert(DIGEST.test(downloadedArchiveSha256) && artifact.digest === `sha256:${downloadedArchiveSha256}`, 'Native artifact archive identity mismatch');
}

export function verifyPendingProducer(manifest, policy, observations, env) {
  verifyNativeProducer(manifest, policy, observations);
  assert(manifest.schemaVersion === 2 && observations.job.name === 'Seal native release', 'Pending native proof requires the canonical sealing job');
  assert(env.GITHUB_REPOSITORY === policy.repository && env.GITHUB_REF === 'refs/heads/main' && env.GITHUB_EVENT_NAME === 'workflow_dispatch', 'Pending native proof requires a manual main release');
  assert(manifest.producer.runId === Number(env.GITHUB_RUN_ID) && manifest.producer.runAttempt <= Number(env.GITHUB_RUN_ATTEMPT) && manifest.languageCommit === env.GITHUB_SHA, 'Pending native proof must belong to this release dispatch');
  assert(observations.run.path === '.github/workflows/release.yml' && observations.artifact.expired === false, 'Pending native proof must come from the registered live release artifact');
  return true;
}

export function verifyProducer(manifest, policy, observations) {
  verifyNativeProducer(manifest, policy, observations);
  const { artifact, release, tagCommit } = observations;
  assert(release.tag_name === manifest.tag && release.draft === false && tagCommit === manifest.languageCommit, 'Release/tag does not identify the native source revision');
  assert(release.prerelease === false, 'Production requires a non-prerelease language release');
  if (manifest.schemaVersion === 2) assert(release.immutable === true, 'Production requires an immutable language release');
  const proof = release.assets?.find((x) => x.name === `${artifact.name}.zip`);
  assert(proof?.state === 'uploaded' && proof.digest === artifact.digest, 'Durable native validation archive is missing from release');
  const identity = release.assets?.find((x) => x.name === `${artifact.name}.json`);
  assert(identity?.state === 'uploaded' && identity.digest === `sha256:${sha256(stableJSON(manifest))}`, 'Published native manifest identity mismatch');
  for (const item of manifest.artifacts) {
    const asset = release.assets?.find((x) => x.name === item.releaseAsset);
    assert(asset && asset.state === 'uploaded' && asset.digest === `sha256:${item.sha256}`, `Published native asset mismatch: ${item.path}`);
  }
  return true;
}

export function verifyProducerJob(manifest, run, job) {
  assert(['in_progress', 'completed'].includes(run.status), 'Native producer run has not reached execution');
  assert(job?.id === manifest.producer.jobId && job.run_id === manifest.producer.runId && job.run_attempt === manifest.producer.runAttempt && job.head_sha === manifest.languageCommit && job.status === 'completed' && job.conclusion === 'success', 'Native producer job did not complete successfully at the release revision');
  // The whole orchestration may fail later in docs publication/deployment.
  // A verified successful native job, its hashed real l-check reports and
  // published native artifact identities survive that downstream failure.
  return true;
}

export async function verifyNativeEvidence(manifest, root) {
  validateNativeShape(manifest);
  await verifyInventory(root, manifest.artifacts);
  const known = await verifyInventory(root, manifest.evidence);
  const identity = new Set();
  for (const group of validationGroups(manifest)) for (const entry of group.stages) {
    assert(known.has(entry.path), `Unhashed stage evidence: ${entry.path}`);
    const report = JSON.parse(await readOwned(root, entry.path));
    assert(report.stage === entry.stage && report.revision === manifest.languageCommit && report.dirty === false && report.passed === true && !report.error, `Failed/mismatched ${entry.stage} report`);
    if (manifest.schemaVersion === 2) assert(`${report.goos}-${report.goarch}` === group.target, `Validation report host differs from ${group.target}`);
    assert(DIGEST.test(report.source_identity), 'Missing canonical validation source identity');
    identity.add(report.source_identity);
    const start = Date.parse(report.started_at), finish = Date.parse(report.finished_at);
    assert(Number.isFinite(start) && Number.isFinite(finish) && finish >= start, 'Incomplete stage timing');
    assert(Array.isArray(report.commands) && report.commands.length > 0 && report.commands.every((x) => x.exit_code === 0 && !x.timed_out), 'Failed or missing canonical stage commands');
    const names = new Set(report.commands.map((x) => x.name));
    const required = { lint: ['gofmt', 'lint-config', 'golangci-ordinary', 'golangci-nativecompile', 'actionlint'], acceptance: ['doctor', 'verify'], native: ['nativecompile'] }[entry.stage];
    assert(required.every((name) => names.has(name)), `Incomplete ${entry.stage} command inventory`);
  }
  assert(identity.size === 1, 'Language stage source identities disagree');
  if (manifest.schemaVersion === 2) for (const group of manifest.validation.targets) {
    assert(known.has(group.smoke), `Unhashed installer smoke evidence: ${group.smoke}`);
    const smoke = JSON.parse(await readOwned(root, group.smoke));
    const artifact = manifest.artifacts.find((item) => item.target === group.target);
    assert(smoke.schemaVersion === 1 && smoke.passed === true && !smoke.error && smoke.target === group.target && smoke.version === manifest.languageVersion && smoke.archiveSha256 === artifact.sha256, `Installer smoke does not match the released package: ${group.target}`);
    assert(smoke.freshCargoHome === true && smoke.freshRustupHome === true && smoke.offlineBuild === true, `Installer smoke omitted isolated setup or offline compilation: ${group.target}`);
    const start = Date.parse(smoke.startedAt), finish = Date.parse(smoke.finishedAt);
    assert(Number.isFinite(start) && Number.isFinite(finish) && finish >= start, `Incomplete installer smoke timing: ${group.target}`);
  }
  // Detailed command/outcome enforcement remains in scripts/l-check itself;
  // trusted producer + raw hashed reports prevents substituting a boolean claim.
  return true;
}

export function planRelease({ mode, docsCommit, languageCommit, version, currentFingerprint, releasedFingerprint, sourceFingerprint, dirty, tag }) {
  assert(['release', 'docs-correction'].includes(mode), 'Planning supports release or docs-correction');
  assert(SHA.test(docsCommit) && SHA.test(languageCommit) && VERSION.test(version) && tag === `v${version}`, 'Release planning requires immutable commits, VERSION and tag');
  assert(dirty === false && DIGEST.test(currentFingerprint) && FINGERPRINT.test(sourceFingerprint), 'Production source must be clean and fingerprinted');
  if (mode === 'release') assert(docsCommit === languageCommit, 'Initial release docs and language must share a revision');
  else {
    assert(docsCommit !== languageCommit, 'A docs correction requires a distinct documentation revision');
    assert(currentFingerprint === releasedFingerprint, 'Language inputs changed; a docs-only correction requires a new language release decision');
  }
  return { languageVersion: version, languageCommit, docsCommit, docsRevision: docsCommit, channel: 'production', dirty: false, sourceFingerprint, releaseFingerprint: currentFingerprint };
}

export function guardPromotion({ mode, desiredDigest, currentDigest, expectedCurrentDigest, currentDocsCommit, desiredDocsCommit, isDescendant }) {
  assert(['release', 'docs-correction', 'retry', 'rollback'].includes(mode), 'Invalid promotion mode');
  assert(DIGEST.test(desiredDigest), 'Invalid desired promotion digest');
  assert(currentDigest === null || DIGEST.test(currentDigest), 'Invalid current production identity');
  if (desiredDigest === currentDigest) return { action: 'verify', reason: 'Already deployed exact artifact' };
  assert(currentDigest === expectedCurrentDigest, 'Stale promotion: production changed after preparation');
  if (mode !== 'rollback' && currentDocsCommit) {
    assert(SHA.test(currentDocsCommit) && SHA.test(desiredDocsCommit) && (currentDocsCommit === desiredDocsCommit || isDescendant === true), 'Stale promotion or source downgrade rejected');
  }
  return { action: 'deploy', requiresProductionApproval: true };
}

export async function directoryInventory(root) {
  const files = [];
  async function walk(dir, prefix = '') {
    for (const item of (await readdir(dir, { withFileTypes: true })).sort((a, b) => a.name.localeCompare(b.name, 'en'))) {
      const name = prefix + item.name;
      assert(!item.isSymbolicLink(), `Symlink rejected: ${name}`);
      if (item.isDirectory()) await walk(resolve(dir, item.name), `${name}/`);
      else { assert(item.isFile(), `Unsupported artifact entry: ${name}`); files.push({ path: name, sha256: sha256(await readOwned(root, name)) }); }
    }
  }
  await walk(root);
  return files;
}

export async function sealSite({ root, archive, nativeManifest, identity, mode, expectedCurrentDigest = null }) {
  assert(['release', 'docs-correction'].includes(mode), 'Only a new validated build can be sealed');
  validateNativeShape(nativeManifest);
  assert(identity.languageVersion === nativeManifest.languageVersion && identity.languageCommit === nativeManifest.languageCommit && identity.sourceFingerprint === nativeManifest.sourceFingerprint, 'Site and native release identities differ');
  assert(identity.dirty === false && identity.channel === 'production' && SHA.test(identity.docsCommit) && identity.docsRevision === identity.docsCommit, 'Invalid production documentation identity');
  const version = JSON.parse(await readOwned(root, 'version.json'));
  for (const key of ['languageVersion', 'languageCommit', 'docsCommit', 'docsRevision', 'channel', 'dirty', 'sourceFingerprint']) assert(version[key] === identity[key], `version.json differs: ${key}`);
  const inventory = await directoryInventory(root);
  return { schemaVersion: 1, mode, ...identity, nativeManifestSha256: sha256(stableJSON(nativeManifest)), siteArchiveSha256: sha256(await readFile(archive)), versionSha256: sha256(await readOwned(root, 'version.json')), expectedCurrentDigest, files: inventory };
}

export async function verifyPromotion(manifest, { root, archive, nativeManifest }) {
  assert(manifest.schemaVersion === 1 && DIGEST.test(manifest.siteArchiveSha256) && DIGEST.test(manifest.versionSha256), 'Invalid promotion manifest');
  assert(sha256(await readFile(archive)) === manifest.siteArchiveSha256, 'Site archive digest mismatch; retry/rollback must use original bytes');
  assert(sha256(stableJSON(nativeManifest)) === manifest.nativeManifestSha256, 'Native release manifest changed');
  assert(manifest.languageVersion === nativeManifest.languageVersion && manifest.languageCommit === nativeManifest.languageCommit && manifest.sourceFingerprint === nativeManifest.sourceFingerprint, 'Promotion changed language identity');
  await verifyInventory(root, manifest.files);
  assert(stableJSON(await directoryInventory(root)) === stableJSON(manifest.files), 'Site archive contains missing or unlisted files');
  const versionBytes = await readOwned(root, 'version.json');
  assert(sha256(versionBytes) === manifest.versionSha256, 'Site version file mismatch');
  const version = JSON.parse(versionBytes);
  for (const key of ['languageVersion', 'languageCommit', 'docsCommit', 'docsRevision', 'channel', 'dirty', 'sourceFingerprint']) assert(version[key] === manifest[key], `Promotion version differs: ${key}`);
  return true;
}

export function sourceFingerprintAt(root, commit) {
  assert(SHA.test(commit), 'Fingerprint requires a full commit SHA');
  const files = execFileSync('git', ['ls-tree', '-r', '-z', '--full-tree', commit], { cwd: root, encoding: 'utf8' }).split('\0').filter(Boolean);
  // Fail closed: everything outside explicitly documentation-owned paths matters.
  const entries = files.filter((entry) => { const path = entry.slice(entry.indexOf('\t') + 1); return !path.startsWith('docs/') && !path.startsWith('website/'); });
  return sha256(entries.sort().join('\0') + '\0');
}

async function github(path) {
  assert(process.env.GITHUB_TOKEN, 'Authenticated GitHub API verification requires GITHUB_TOKEN');
  assert(path.startsWith('/repos/'), 'Unexpected API path');
  const response = await fetch(`https://api.github.com${path}`, { headers: { Authorization: `Bearer ${process.env.GITHUB_TOKEN}`, Accept: 'application/vnd.github+json', 'X-GitHub-Api-Version': '2022-11-28' } });
  assert(response.ok, `GitHub verification failed: ${response.status} ${path}`);
  return response.json();
}

async function download(url, output) {
  assert(url.startsWith('https://api.github.com/repos/') && process.env.GITHUB_TOKEN, 'Authenticated repository download required');
  const response = await fetch(url, { headers: { Authorization: `Bearer ${process.env.GITHUB_TOKEN}`, Accept: 'application/octet-stream' } });
  assert(response.ok, `Artifact download failed: ${response.status}`);
  const bytes = Buffer.from(await response.arrayBuffer());
  assert(bytes.length <= 1024 ** 3, 'Artifact archive exceeds 1 GiB');
  await writeFile(output, bytes, { flag: 'wx' });
  return bytes;
}

export function documentationTag({ languageVersion, languageCommit, docsCommit }) {
  assert(VERSION.test(languageVersion) && SHA.test(languageCommit) && SHA.test(docsCommit), 'Documentation storage requires exact source identities');
  return docsCommit === languageCommit ? `v${languageVersion}` : `docs-v${languageVersion}-${docsCommit}`;
}

async function tagCommit(repository, tag) {
  let object = (await github(`/repos/${repository}/git/ref/tags/${encodeURIComponent(tag)}`)).object;
  for (let depth = 0; object.type === 'tag' && depth < 8; depth++) object = (await github(`/repos/${repository}/git/tags/${object.sha}`)).object;
  assert(object.type === 'commit', 'Release tag does not resolve to a commit');
  return object.sha;
}

async function documentationRelease(policy, identity) {
  const tag = documentationTag(identity), release = await github(`/repos/${policy.repository}/releases/tags/${tag}`);
  assert(release.tag_name === tag && release.draft === false && release.immutable === true && release.prerelease === (identity.docsCommit !== identity.languageCommit), 'Documentation requires its published immutable archive release');
  assert(await tagCommit(policy.repository, tag) === identity.docsCommit, 'Documentation archive tag identifies a different revision');
  return release;
}

async function fetchCurrent(policy) {
  assertPublicationPolicy(policy);
  const deployments = await github(`/repos/${policy.repository}/deployments?environment=github-pages&per_page=30`);
  let deployed = null;
  for (const item of deployments) {
    const states = await github(`/repos/${policy.repository}/deployments/${item.id}/statuses`);
    if (states[0]?.state === 'success') { deployed = item; break; }
  }
  const response = await fetch(`${policy.origin}/version.json`, { cache: 'no-store', redirect: 'error' });
  if (!deployed && response.status === 404) return null;
  assert(deployed && response.ok, 'Production HTTP and GitHub deployment records disagree; first deployment requires no prior success and HTTP 404');
  const version = await response.json();
  // Recovery runs can execute trusted workflow code from a newer commit while
  // promoting an older exact artifact; the deployed content owns docsCommit.
  assert(SHA.test(version.docsCommit) && VERSION.test(version.languageVersion), 'Production identity lacks immutable source identity');
  const release = await documentationRelease(policy, version);
  const asset = release.assets?.find((x) => x.name === `docs-${version.docsCommit}.promotion.json`);
  assert(asset && /^sha256:[a-f0-9]{64}$/.test(asset.digest), 'Production promotion manifest is missing from its release');
  const raw = await fetch(asset.url, { headers: { Authorization: `Bearer ${process.env.GITHUB_TOKEN}`, Accept: 'application/octet-stream' } });
  assert(raw.ok, 'Cannot read production promotion manifest');
  const bytes = Buffer.from(await raw.arrayBuffer());
  assert(`sha256:${sha256(bytes)}` === asset.digest, 'Published promotion manifest digest mismatch');
  const current = JSON.parse(bytes);
  assert(current.docsCommit === version.docsCommit && current.languageCommit === version.languageCommit, 'Production release manifest and live identity disagree');
  return current;
}

export async function documentationFiles(manifest, archive, nativePath, manifestPath) {
  assert(manifest.channel === 'production' && SHA.test(manifest.docsCommit), 'Only sealed production artifacts can be published');
  const files = new Map(), prefix = `docs-${manifest.docsCommit}`;
  for (const [suffix, file, expected] of [
    ['tar.gz', archive, manifest.siteArchiveSha256],
    ['promotion.json', manifestPath, sha256(stableJSON(manifest))],
    ['native.json', nativePath, manifest.nativeManifestSha256],
  ]) {
    const bytes = await readFile(file);
    assert(sha256(bytes) === expected, `Sealed documentation changed: ${suffix}`);
    files.set(`${prefix}.${suffix}`, bytes);
  }
  return files;
}

async function publishDocs(policy, manifest, archive, nativePath, manifestPath, root) {
  assertPublicationPolicy(policy);
  const { guard, publishAssetSet } = await import('../../scripts/release/github.mjs');
  await guard({ policy });
  assert(root && manifest.mode === 'docs-correction' && manifest.docsCommit === process.env.GITHUB_SHA && manifest.docsCommit !== manifest.languageCommit, 'Docs correction requires a distinct captured documentation revision');
  const native = JSON.parse(await readFile(nativePath, 'utf8'));
  await verifyPromotion(manifest, { root, archive, nativeManifest: native });
  const release = await github(`/repos/${policy.repository}/releases/tags/${native.tag}`);
  await verifyReleasedAuthority(policy, native, release);
  assert(sourceFingerprintAt(process.env.GITHUB_WORKSPACE, manifest.docsCommit) === native.releaseFingerprint && manifest.releaseFingerprint === native.releaseFingerprint, 'Docs correction changed release-relevant source');
  await publishAssetSet(policy.repository, { tag: documentationTag(manifest), commit: manifest.docsCommit,
    name: `Documentation for ${native.tag} (${manifest.docsCommit})`, prerelease: true, latest: false,
    body: `Immutable documentation snapshot for L ${native.tag}. Compiler downloads remain on the language release.` }, await documentationFiles(manifest, archive, nativePath, manifestPath));
}

async function verifyReleasedAuthority(policy, native, release) {
  validateNativeShape(native);
  validateTargetPolicy(native, policy);
  assert(policy.nativePublisher && native.repository === policy.repository, 'Native publisher/repository is not configured');
  assert(release.tag_name === native.tag && release.draft === false && release.prerelease === false && release.immutable === true, 'Native authority requires a published immutable language release');
  const identity = release.assets.find((x) => x.name === `l-native-release-${native.languageCommit}.json`);
  assert(identity?.state === 'uploaded' && identity.digest === `sha256:${sha256(stableJSON(native))}`, 'Native manifest is not the published release identity');
  const [run, job] = await Promise.all([github(`/repos/${policy.repository}/actions/runs/${native.producer.runId}/attempts/${native.producer.runAttempt}`), github(`/repos/${policy.repository}/actions/jobs/${native.producer.jobId}`)]);
  assert(run.workflow_id === policy.nativePublisher.workflowId && run.path === policy.nativePublisher.workflowPath && run.head_sha === native.languageCommit && (native.schemaVersion === 2 ? ['workflow_dispatch'] : ['push', 'workflow_dispatch']).includes(run.event) && run.head_branch === policy.productionBranch, 'Native producer is untrusted or failed');
  verifyProducerJob(native, run, job);
  assert(await tagCommit(policy.repository, native.tag) === native.languageCommit, 'Native language tag moved');
  for (const item of native.artifacts) assert(release.assets.some((x) => x.name === item.releaseAsset && x.digest === `sha256:${item.sha256}` && x.state === 'uploaded'), 'Native artifact identity changed');
}

async function downloadPromotion(policy, tag, docsCommit, output) {
  assertPublicationPolicy(policy);
  assert(policy.nativePublisher, 'Native release publisher is not configured; production activation is blocked');
  assert(SHA.test(docsCommit) && tag.startsWith('v') && VERSION.test(tag.slice(1)), 'Recovery requires an immutable docs SHA and real release tag');
  const nativeRelease = await github(`/repos/${policy.repository}/releases/tags/${tag}`);
  const languageCommit = await tagCommit(policy.repository, tag);
  const release = await documentationRelease(policy, { languageVersion: tag.slice(1), languageCommit, docsCommit });
  await mkdir(output);
  for (const [suffix, filename] of [['.tar.gz', 'site.tar.gz'], ['.promotion.json', 'promotion.json'], ['.native.json', 'native-release.json']]) {
    const asset = release.assets.find((x) => x.name === `docs-${docsCommit}${suffix}`);
    assert(asset?.state === 'uploaded' && /^sha256:[a-f0-9]{64}$/.test(asset.digest), `Missing archived recovery asset: ${suffix}`);
    const bytes = await download(asset.url, resolve(output, filename));
    assert(`sha256:${sha256(bytes)}` === asset.digest, `Recovery asset digest mismatch: ${suffix}`);
  }
  const manifest = JSON.parse(await readFile(resolve(output, 'promotion.json'))), native = JSON.parse(await readFile(resolve(output, 'native-release.json')));
  validateNativeShape(native);
  assert(manifest.docsCommit === docsCommit && manifest.languageCommit === languageCommit && native.tag === tag && native.repository === policy.repository && sha256(stableJSON(native)) === manifest.nativeManifestSha256, 'Recovery identity does not match requested release');
  await verifyReleasedAuthority(policy, native, nativeRelease);
  assert(sha256(await readFile(resolve(output, 'site.tar.gz'))) === manifest.siteArchiveSha256, 'Recovery site archive differs from sealed artifact');
}

async function verifyNativeAuthority(manifest, policy, artifactId, archive, mode) {
  assert(['release', 'docs-correction'].includes(mode), 'Native authority requires an explicit release mode');
  assert(/^\d+$/.test(String(artifactId)), 'An immutable native artifact ID is required');
  const base = `/repos/${manifest.repository}`;
  const [run, job, artifact] = await Promise.all([
    github(`${base}/actions/runs/${manifest.producer.runId}/attempts/${manifest.producer.runAttempt}`),
    github(`${base}/actions/jobs/${manifest.producer.jobId}`),
    github(`${base}/actions/artifacts/${artifactId}`),
  ]);
  const observations = { run, job, artifact, downloadedArchiveSha256: sha256(await readFile(archive)) };
  if (mode === 'release') {
    const { guard } = await import('../../scripts/release/github.mjs');
    const current = await guard({ policy });
    assert(current.path === policy.nativePublisher.workflowPath, 'Initial documentation must run inside the registered native workflow');
    return verifyPendingProducer(manifest, policy, observations, process.env);
  }
  observations.release = await github(`${base}/releases/tags/${encodeURIComponent(manifest.tag)}`);
  observations.tagCommit = await tagCommit(manifest.repository, manifest.tag);
  return verifyProducer(manifest, policy, observations);
}

export async function authenticatedReleaseIdentity({ root, policyPath, manifestPath, artifacts, artifactId, archive, mode, sourceFingerprint }) {
  const policy = JSON.parse(await readFile(policyPath, 'utf8'));
  assertPublicationPolicy(policy);
  assert(policy.nativePublisher, 'Native release publisher is not configured; production activation is blocked');
  const manifest = JSON.parse(await readFile(manifestPath, 'utf8'));
  await verifyNativeArchive(archive, artifacts, manifestPath);
  await verifyNativeEvidence(manifest, artifacts);
  await verifyNativeAuthority(manifest, policy, artifactId, archive, mode);
  const head = execFileSync('git', ['rev-parse', 'HEAD'], { cwd: root, encoding: 'utf8' }).trim();
  const version = (await readFile(resolve(root, 'VERSION'), 'utf8')).trim();
  const dirty = execFileSync('git', ['status', '--porcelain'], { cwd: root, encoding: 'utf8' }).trim() !== '';
  const released = sourceFingerprintAt(root, manifest.languageCommit);
  assert(released === manifest.releaseFingerprint, 'Released source inventory disagrees with native manifest');
  assert(sourceFingerprint === manifest.sourceFingerprint, 'Website reference inputs do not match released native inputs');
  return { ...planRelease({ mode, docsCommit: head, languageCommit: manifest.languageCommit, version, currentFingerprint: sourceFingerprintAt(root, head), releasedFingerprint: released, sourceFingerprint, dirty, tag: manifest.tag }), repository: `https://github.com/${policy.repository}` };
}

function parseArgs(args) {
  const options = {};
  for (let i = 0; i < args.length; i += 2) { assert(args[i].startsWith('--') && args[i + 1] && !args[i + 1].startsWith('--'), `Missing argument: ${args[i]}`); options[args[i].slice(2)] = args[i + 1]; }
  return options;
}

async function main() {
  const [command, ...args] = process.argv.slice(2), options = parseArgs(args);
  const json = async (path) => { assert(path, 'Missing required file argument'); return JSON.parse(await readFile(path, 'utf8')); };
  if (command === 'download-native') {
    const policy = assertPublicationPolicy(await json(options.policy)); assert(/^\d+$/.test(options['artifact-id']), 'Invalid native artifact ID');
    assert(['release', 'docs-correction'].includes(options.mode), 'Native download requires an explicit release mode');
    const artifact = await github(`/repos/${policy.repository}/actions/artifacts/${options['artifact-id']}`);
    assert(/^l-native-release-[a-f0-9]{40}$/.test(artifact.name), 'Artifact is not a native release proof');
    let url;
    if (options.mode === 'release') {
      assert(artifact.expired === false, 'Native proof artifact has expired');
      url = `https://api.github.com/repos/${policy.repository}/actions/artifacts/${options['artifact-id']}/zip`;
    } else {
      const version = (await readFile(resolve(options.root || '..', 'VERSION'), 'utf8')).trim();
      assert(VERSION.test(version), 'Native archive requires real VERSION');
      const release = await github(`/repos/${policy.repository}/releases/tags/v${version}`);
      assert(release.draft === false && release.prerelease === false && release.immutable === true, 'Docs correction requires a published immutable native release');
      const proof = release.assets?.find((x) => x.name === `${artifact.name}.zip`);
      assert(proof?.state === 'uploaded' && proof.digest === artifact.digest, 'Durable native archive differs from the original workflow artifact');
      url = proof.url;
    }
    const bytes = await download(url, options.output);
    assert(`sha256:${sha256(bytes)}` === artifact.digest, 'Downloaded native archive digest mismatch');
  } else if (command === 'download-promotion') {
    await downloadPromotion(await json(options.policy), options.tag, options['docs-commit'], options.output);
  } else if (command === 'verify-native') {
    const manifest = await json(options.manifest), policy = assertPublicationPolicy(await json(options.policy));
    assert(policy.nativePublisher, 'Native release publisher is not configured; production activation is blocked');
    await verifyNativeArchive(options.archive, options.artifacts, options.manifest);
    await verifyNativeEvidence(manifest, options.artifacts);
    await verifyNativeAuthority(manifest, policy, options['artifact-id'], options.archive, options.mode);
    console.log('Verified native producer, archive authority and canonical validation evidence');
  } else if (command === 'extract-native') {
    const files = readArtifactZip(await readFile(options.archive));
    await mkdir(options.output); // Must be new: never overwrite another artifact tree.
    for (const [name, bytes] of files) { const path = resolve(options.output, name); await mkdir(dirname(path), { recursive: true }); await writeFile(path, bytes, { flag: 'wx' }); }
  } else if (command === 'plan') {
    const native = await json(options.native), root = resolve(options.root || '..');
    const head = execFileSync('git', ['rev-parse', 'HEAD'], { cwd: root, encoding: 'utf8' }).trim();
    const version = (await readFile(resolve(root, 'VERSION'), 'utf8')).trim();
    const dirty = execFileSync('git', ['status', '--porcelain'], { cwd: root, encoding: 'utf8' }).trim() !== '';
    const released = sourceFingerprintAt(root, native.languageCommit);
    assert(released === native.releaseFingerprint, 'Released source inventory disagrees with native manifest');
    const plan = planRelease({ mode: options.mode, docsCommit: head, languageCommit: native.languageCommit, version, currentFingerprint: sourceFingerprintAt(root, head), releasedFingerprint: released, sourceFingerprint: native.sourceFingerprint, dirty, tag: native.tag });
    await writeFile(options.output, stableJSON(plan));
  } else if (command === 'seal-site') {
    const result = await sealSite({ root: options.dist, archive: options.archive, nativeManifest: await json(options.native), identity: await json(options.identity), mode: options.mode, expectedCurrentDigest: options['expected-current'] === 'null' ? null : options['expected-current'] });
    await writeFile(options.output, stableJSON(result));
  } else if (command === 'verify-promotion') {
    await verifyPromotion(await json(options.manifest), { root: options.dist, archive: options.archive, nativeManifest: await json(options.native) });
    console.log('Verified exact site archive, complete file inventory and native identity');
  } else if (command === 'guard-promotion') {
    const manifest = await json(options.manifest), current = options.current === 'null' ? null : await json(options.current);
    const root = resolve(options.root || '..');
    let descendant = false;
    if (current) { try { execFileSync('git', ['merge-base', '--is-ancestor', current.docsCommit, manifest.docsCommit], { cwd: root, stdio: 'ignore' }); descendant = true; } catch { /* Ancestry failure rejects promotion below. */ } }
    const expected = ['retry', 'rollback'].includes(options.mode) && options['expected-current'] ? options['expected-current'] : manifest.expectedCurrentDigest ?? null;
    // Dispatch inputs are strings; preserve an explicit empty-production baseline
    // instead of falling back to the retained manifest's original digest.
    const expectedDigest = expected === 'null' ? null : expected;
    console.log(stableJSON(guardPromotion({ mode: options.mode, desiredDigest: manifest.siteArchiveSha256, currentDigest: current?.siteArchiveSha256 ?? null, expectedCurrentDigest: expectedDigest, currentDocsCommit: current?.docsCommit, desiredDocsCommit: manifest.docsCommit, isDescendant: descendant })));
  } else if (command === 'fetch-current') {
    const policy = await json(options.policy);
    await writeFile(options.output, stableJSON(await fetchCurrent(policy)));
  } else if (command === 'publish-docs') {
    await publishDocs(await json(options.policy), await json(options.manifest), options.archive, options.native, options.manifest, options.dist);
  } else if (command === 'postdeploy') {
    const policy = await json(options.policy), expected = await json(options.manifest);
    const current = await fetchCurrent(policy);
    assert(current?.siteArchiveSha256 === expected.siteArchiveSha256 && current.docsCommit === expected.docsCommit, 'Deployed identity is not the validated artifact');
    for (const path of ['/', '/docs/spec/', '/docs/cli/', '/pagefind/pagefind.js']) { const response = await fetch(policy.origin + path, { redirect: 'error', cache: 'no-store' }); assert(response.ok, `Production route failed: ${path}`); }
    console.log('Production version and representative routes verified');
  } else throw new Error('Usage: release.mjs verify-native|plan|seal-site|verify-promotion|guard-promotion|fetch-current with explicit file arguments; see DEPLOYMENT.md');
}

if (process.argv[1] && resolve(process.argv[1]) === fileURLToPath(import.meta.url)) main().catch((error) => { console.error(`Release refused: ${error.message}`); process.exitCode = 1; });
