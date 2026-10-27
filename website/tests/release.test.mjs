import test from 'node:test';
import assert from 'node:assert/strict';
import { mkdtemp, mkdir, writeFile, rm, symlink } from 'node:fs/promises';
import { tmpdir } from 'node:os';
import { join } from 'node:path';
import { verifyDisabled } from '../scripts/verify-disabled.mjs';
import {
  sha256, stableJSON, safeRelative, verifyInventory, verifyNativeEvidence,
  validateNativeShape, verifyProducer, planRelease, guardPromotion,
  sealSite, verifyPromotion, readArtifactZip, verifyNativeArchive,
} from '../scripts/release.mjs';

const A = 'a'.repeat(40), B = 'b'.repeat(40), HASH = 'c'.repeat(64);
const fingerprint = `sha256:${HASH}`;

function nativeManifest() {
  return {
    schemaVersion: 1, repository: 'example/l', languageVersion: '0.1.0', tag: 'v0.1.0', languageCommit: A,
    sourceFingerprint: fingerprint, releaseFingerprint: HASH, dirty: false,
    producer: { runId: 21, runAttempt: 1, jobId: 31 },
    artifacts: [{ path: 'l-linux-amd64.tar.gz', releaseAsset: 'l-linux-amd64.tar.gz', sha256: sha256('native bytes'), target: 'x86_64-unknown-linux-gnu', compilerVersion: 'go1.27.1' }],
    evidence: [], validation: { command: 'scripts/l-check', stages: ['lint', 'acceptance', 'native'].map((stage) => ({ stage, path: `evidence/${stage}.json` })) },
  };
}

function authority(manifest) {
  const archiveHash = sha256('immutable archive');
  return {
    policy: { repository: 'example/l', productionBranch: 'main', nativePublisher: { workflowId: 10, workflowPath: '.github/workflows/native-release.yml' } },
    observations: {
      run: { repository: { full_name: 'example/l' }, id: 21, run_attempt: 1, workflow_id: 10, path: '.github/workflows/native-release.yml', head_sha: A, head_branch: 'main', event: 'push', status: 'completed', conclusion: 'success' },
      job: { id: 31, run_id: 21, run_attempt: 1, head_sha: A, status: 'completed', conclusion: 'success' },
      artifact: { id: 99, name: `l-native-release-${A}`, workflow_run: { id: 21, head_sha: A }, expired: false, digest: `sha256:${archiveHash}` },
      release: { tag_name: 'v0.1.0', draft: false, prerelease: false, assets: [...manifest.artifacts.map((item) => ({ name: item.path, state: 'uploaded', digest: `sha256:${item.sha256}` })), { name: `l-native-release-${A}.zip`, state: 'uploaded', digest: `sha256:${archiveHash}` }, { name: `l-native-release-${A}.json`, state: 'uploaded', digest: `sha256:${sha256(stableJSON(manifest))}` }] },
      tagCommit: A, downloadedArchiveSha256: archiveHash,
    },
  };
}

async function temporary(t) {
  const root = await mkdtemp(join(tmpdir(), 'l docs release '));
  t.after(() => rm(root, { recursive: true, force: true }));
  return root;
}

test('release protocol requires real identity and all canonical stages', () => {
  assert.equal(validateNativeShape(nativeManifest()).languageCommit, A);
  for (const change of [(m) => { m.dirty = true; }, (m) => { m.languageVersion = 'Unreleased'; }, (m) => { m.tag = 'v9.9.9'; }, (m) => { m.artifacts = []; }, (m) => { m.validation.stages.pop(); }, (m) => { m.validation.command = 'go test ./cmd/l'; }]) {
    const manifest = nativeManifest(); change(manifest); assert.throws(() => validateNativeShape(manifest));
  }
});

test('native authority comes from repository/workflow/run/archive/release evidence, not a boolean', () => {
  const manifest = nativeManifest(), { policy, observations } = authority(manifest);
  assert.equal(verifyProducer(manifest, policy, observations), true);
  assert.throws(() => verifyProducer(manifest, { ...policy, nativePublisher: null }, observations), /not configured/);
  for (const alter of [
    (o) => { o.run.workflow_id = 20; }, (o) => { o.run.head_sha = B; },
    (o) => { o.run.event = 'pull_request'; }, (o) => { o.job.conclusion = 'failure'; },
    (o) => { o.job.id = 32; }, (o) => { o.job.head_sha = B; },
    (o) => { o.tagCommit = B; }, (o) => { o.release.draft = true; },
    (o) => { o.artifact.digest = `sha256:${HASH}`; }, (o) => { o.artifact.workflow_run.id = 22; },
    (o) => { o.release.assets[0].digest = `sha256:${HASH}`; },
    (o) => { o.release.assets.pop(); },
  ]) { const changed = structuredClone(observations); alter(changed); changed.validated = true; assert.throws(() => verifyProducer(manifest, policy, changed)); }
  const active = structuredClone(observations); active.run.status = 'in_progress'; active.run.conclusion = null;
  assert.equal(verifyProducer(manifest, policy, active), true);
  const docsFailed = structuredClone(observations); docsFailed.run.conclusion = 'failure';
  assert.equal(verifyProducer(manifest, policy, docsFailed), true);
  docsFailed.job.conclusion = 'failure';
  assert.throws(() => verifyProducer(manifest, policy, docsFailed), /job did not complete/);
});

test('complete hashed reports are required; fixture evidence does not execute L', async (t) => {
  const root = await temporary(t), manifest = nativeManifest();
  await mkdir(join(root, 'evidence'));
  await writeFile(join(root, manifest.artifacts[0].path), 'native bytes');
  for (const entry of manifest.validation.stages) {
    const commands = { lint: ['gofmt', 'lint-config', 'golangci-ordinary', 'golangci-nativecompile', 'actionlint'], acceptance: ['doctor', 'verify'], native: ['nativecompile'] }[entry.stage];
    const body = stableJSON({ stage: entry.stage, revision: A, dirty: false, passed: true, source_identity: HASH, started_at: '2026-10-01T00:00:00Z', finished_at: '2026-10-01T00:01:00Z', commands: commands.map((name) => ({ name, exit_code: 0, timed_out: false })) });
    await writeFile(join(root, entry.path), body); manifest.evidence.push({ path: entry.path, sha256: sha256(body) });
  }
  assert.equal(await verifyNativeEvidence(manifest, root), true);
  await writeFile(join(root, 'evidence/native.json'), '{}');
  await assert.rejects(verifyNativeEvidence(manifest, root), /Digest mismatch/);
});

test('initial release same-revision and docs-only fingerprint invariants', () => {
  const input = { mode: 'release', docsCommit: A, languageCommit: A, version: '0.1.0', currentFingerprint: HASH, releasedFingerprint: HASH, sourceFingerprint: fingerprint, dirty: false, tag: 'v0.1.0' };
  assert.equal(planRelease(input).channel, 'production');
  assert.throws(() => planRelease({ ...input, docsCommit: B }), /share a revision/);
  const corrected = planRelease({ ...input, mode: 'docs-correction', docsCommit: B });
  assert.equal(corrected.languageCommit, A); assert.equal(corrected.docsCommit, B);
  assert.throws(() => planRelease({ ...input, mode: 'docs-correction', docsCommit: B, currentFingerprint: 'd'.repeat(64) }), /Language inputs changed/);
  assert.throws(() => planRelease({ ...input, dirty: true }), /clean/);
});

test('promotions reject stale baseline and downgrade, retry exact bytes, and isolate rollback', () => {
  const old = 'd'.repeat(64), next = 'e'.repeat(64);
  const input = { mode: 'release', desiredDigest: next, currentDigest: old, expectedCurrentDigest: old, currentDocsCommit: A, desiredDocsCommit: B, isDescendant: true };
  assert.equal(guardPromotion(input).action, 'deploy');
  assert.throws(() => guardPromotion({ ...input, expectedCurrentDigest: null }), /Stale/);
  assert.throws(() => guardPromotion({ ...input, isDescendant: false }), /downgrade/);
  assert.equal(guardPromotion({ ...input, mode: 'retry', currentDigest: next }).action, 'verify');
  assert.equal(guardPromotion({ ...input, mode: 'rollback', isDescendant: false }).requiresProductionApproval, true);
});

test('paths, duplicate inventories and symlinks are rejected', async (t) => {
  for (const path of ['../outside', '/tmp/a', 'a/../b', 'a\\b', 'C:/a', 'a//b']) assert.throws(() => safeRelative(path));
  const root = await temporary(t); await writeFile(join(root, 'file with spaces'), 'data');
  const item = { path: 'file with spaces', sha256: sha256('data') };
  await verifyInventory(root, [item]);
  await assert.rejects(verifyInventory(root, [item, item]), /Duplicate/);
  await symlink(join(root, 'file with spaces'), join(root, 'link'));
  await assert.rejects(verifyInventory(root, [{ ...item, path: 'link' }]), /Symlink/);
});

test('sealed promotion validates whole inventory, version identity and original archive', async (t) => {
  const root = await temporary(t), dist = join(root, 'dist'); await mkdir(dist);
  const native = nativeManifest();
  const identity = planRelease({ mode: 'release', docsCommit: A, languageCommit: A, version: '0.1.0', currentFingerprint: HASH, releasedFingerprint: HASH, sourceFingerprint: fingerprint, dirty: false, tag: 'v0.1.0' });
  await writeFile(join(dist, 'version.json'), stableJSON(identity)); await writeFile(join(dist, 'index.html'), '<h1>L</h1>');
  const archive = join(root, 'site.tar.gz'); await writeFile(archive, 'fixture archive, not a deployed site');
  const promotion = await sealSite({ root: dist, archive, nativeManifest: native, identity, mode: 'release' });
  await verifyPromotion(promotion, { root: dist, archive, nativeManifest: native });
  await writeFile(join(dist, 'unexpected.html'), 'unreviewed');
  await assert.rejects(verifyPromotion(promotion, { root: dist, archive, nativeManifest: native }), /unlisted/);
  await rm(join(dist, 'unexpected.html')); await writeFile(archive, 'changed archive');
  await assert.rejects(verifyPromotion(promotion, { root: dist, archive, nativeManifest: native }), /original bytes/);
});

function storedZip(entries) {
  const locals = [], central = []; let offset = 0;
  for (const [name, value] of entries) {
    const n = Buffer.from(name), b = Buffer.from(value), local = Buffer.alloc(30), c = Buffer.alloc(46);
    local.writeUInt32LE(0x04034b50); local.writeUInt32LE(b.length, 18); local.writeUInt32LE(b.length, 22); local.writeUInt16LE(n.length, 26);
    c.writeUInt32LE(0x02014b50); c.writeUInt32LE(b.length, 20); c.writeUInt32LE(b.length, 24); c.writeUInt16LE(n.length, 28); c.writeUInt32LE(offset, 42);
    locals.push(local, n, b); central.push(c, n); offset += local.length + n.length + b.length;
  }
  const directory = Buffer.concat(central), end = Buffer.alloc(22); end.writeUInt32LE(0x06054b50); end.writeUInt16LE(entries.length, 8); end.writeUInt16LE(entries.length, 10); end.writeUInt32LE(directory.length, 12); end.writeUInt32LE(offset, 16);
  return Buffer.concat([...locals, directory, end]);
}

test('native ZIP safely binds extracted manifest and bytes to immutable archive', async (t) => {
  const zip = storedZip([['native-release.json', '{}'], ['folder with spaces/evidence.json', 'evidence']]);
  assert.equal(readArtifactZip(zip).size, 2);
  assert.throws(() => readArtifactZip(storedZip([['native-release.json', '{}'], ['../escape', 'bad']])), /traversal/);
  assert.throws(() => readArtifactZip(storedZip([['native-release.json', '{}'], ['native-release.json', 'other']])), /Duplicate/);
  assert.throws(() => readArtifactZip(Buffer.from('not zip')), /directory/);
  const root = await temporary(t), extracted = join(root, 'unpacked'); await mkdir(join(extracted, 'folder with spaces'), { recursive: true });
  for (const [name, bytes] of readArtifactZip(zip)) await writeFile(join(extracted, name), bytes);
  const archive = join(root, 'native.zip'); await writeFile(archive, zip);
  await verifyNativeArchive(archive, extracted, join(extracted, 'native-release.json'));
  await writeFile(join(extracted, 'folder with spaces/evidence.json'), 'forged');
  await assert.rejects(verifyNativeArchive(archive, extracted, join(extracted, 'native-release.json')), /differs/);
});

test('deployment guard permits reviewed checks and rejects publication, npm hooks, Dependabot and mutable Action pins', async (t) => {
  const root = await temporary(t), site = join(root, 'website');
  await mkdir(join(site, 'ci-templates'), { recursive: true }); await mkdir(join(site, 'scripts'));
  await mkdir(join(root, '.github/workflows'), { recursive: true });
  const minimal = 'name: Dormant fixture\non:\n  pull_request: {}\njobs:\n  check:\n    runs-on: ubuntu-24.04\n    steps:\n      - run: true\n';
  for (const name of ['docs-check.yml.disabled', 'docs-release-pages.yml.disabled', 'docs-recover-pages.yml.disabled']) await writeFile(join(site, 'ci-templates', name), minimal);
  await writeFile(join(site, 'package.json'), '{"scripts":{"build":"node build.mjs"}}');
  await assert.rejects(verifyDisabled({ root }), /Required active/);
  const checks = minimal.replace('name: Dormant fixture', 'name: L documentation checks\npermissions:\n  contents: read');
  const approved = join(root, '.github/workflows/docs-check.yml');
  await writeFile(approved, checks);
  assert.deepEqual(await verifyDisabled({ root }), { templates: 3, state: 'checks-active', deployment: 'disabled' });
  const quality = join(root, '.github/workflows/quality.yml');
  await writeFile(quality, checks.replace('- run: true', '- run: npm --prefix=website audit --audit-level=high'));
  assert.equal((await verifyDisabled({ root })).deployment, 'disabled');
  await rm(quality);
  for (const [changed, pattern] of [
    [checks.replace('contents: read', 'contents: write'), /read-only/],
    [checks.replace('pull_request:', 'pull_request_target:'), /Unexpected event/],
    [checks.replace('    runs-on:', '    environment: github-pages\n    runs-on:'), /environments/],
    [checks.replace('    runs-on:', '    permissions:\n      pages: write\n    runs-on:'), /write permissions/],
    [checks.replace('- run: true', `- uses: actions/deploy-pages@${A}`), /Deployment/],
    [checks.replace('- run: true', '- run: node scripts/release.mjs deploy'), /Deployment/],
    [checks.replace('- run: true', '- run: echo "${{ secrets.DEPLOY_TOKEN }}"'), /secrets/],
    [checks.replace('- run: true', '- uses: actions/checkout@v4'), /Unpinned/],
    [checks.replace('- run: true', `- uses: actions/checkout@${A}`), /persists credentials/],
  ]) {
    await writeFile(approved, changed);
    await assert.rejects(verifyDisabled({ root }), pattern);
  }
  await writeFile(approved, checks);
  const release = join(root, '.github/workflows/docs-release-pages.yml');
  await writeFile(release, minimal);
  await assert.rejects(verifyDisabled({ root }), /deployment automation is active/); await rm(release);
  const active = join(root, '.github/workflows/docs.yml'); await writeFile(active, 'name: Website\njobs:\n  build:\n    working-directory: website\n');
  await assert.rejects(verifyDisabled({ root }), /active/); await rm(active);
  for (const run of ['cd website && npm run build', 'npm --prefix=website run build']) {
    await writeFile(active, `name: Wrapper\njobs:\n  build:\n    steps:\n      - run: ${run}\n`);
    await assert.rejects(verifyDisabled({ root }), /active/); await rm(active);
  }
  await writeFile(join(root, '.github/dependabot.yml'), 'version: 2\nupdates:\n  - package-ecosystem: npm\n    directory: /website\n');
  await assert.rejects(verifyDisabled({ root }), /Dependabot/); await rm(join(root, '.github/dependabot.yml'));
  await writeFile(join(site, 'package.json'), '{"scripts":{"postinstall":"node activate.mjs"}}');
  await assert.rejects(verifyDisabled({ root }), /lifecycle/); await writeFile(join(site, 'package.json'), '{}');
  await writeFile(join(site, 'package.json'), '{"scripts":{"build":"cp ci-templates/docs-check.yml.disabled ../.github/workflows/docs.yml"}}');
  await assert.rejects(verifyDisabled({ root }), /addresses active/); await writeFile(join(site, 'package.json'), '{}');
  await writeFile(join(site, 'ci-templates/docs-check.yml.disabled'), minimal.replace('- run: true', '- uses: actions/checkout@v4'));
  await assert.rejects(verifyDisabled({ root }), /Unpinned/);
});
