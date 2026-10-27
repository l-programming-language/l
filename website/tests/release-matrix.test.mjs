import test from 'node:test';
import assert from 'node:assert/strict';
import { mkdtemp, mkdir, writeFile, rm } from 'node:fs/promises';
import { join } from 'node:path';
import { tmpdir } from 'node:os';
import { assertPublicationPolicy, nativeTargets, sha256, stableJSON, validateNativeShape, verifyNativeEvidence, verifyProducer } from '../scripts/release.mjs';
import { verifyReleaseWorkflow } from '../scripts/verify-disabled.mjs';

const SHA = 'a'.repeat(40), HASH = 'b'.repeat(64);
function policy() {
  return { repository: 'l-programming-language/l', productionBranch: 'main', publicationEnabled: true, targets: [...nativeTargets], nativePublisher: { workflowId: 8, workflowPath: '.github/workflows/release.yml' } };
}
function manifest() {
  return {
    schemaVersion: 2, repository: 'l-programming-language/l', languageVersion: '0.0.1', tag: 'v0.0.1', languageCommit: SHA,
    sourceFingerprint: `sha256:${HASH}`, releaseFingerprint: HASH, dirty: false,
    producer: { runId: 1, runAttempt: 1, jobId: 2 }, evidence: [],
    artifacts: nativeTargets.map((target) => ({ path: `l-${target}.tar.gz`, releaseAsset: `l-${target}.tar.gz`, sha256: sha256(target), target, compilerVersion: 'go1.27.3' })),
    validation: { command: 'scripts/l-check', targets: nativeTargets.map((target) => ({ target, smoke: `evidence/${target}/smoke.json`, stages: ['lint', 'acceptance', 'native'].map((stage) => ({ stage, path: `evidence/${target}/${stage}.json` })) })) },
  };
}

test('production policy stays blocked without all activation decisions', () => {
  assert.equal(assertPublicationPolicy(policy()).publicationEnabled, true);
  for (const changed of [
    { publicationEnabled: false }, { repository: 'other/repository' }, { nativePublisher: null },
    { productionBranch: 'other' }, { targets: nativeTargets.slice(1) },
  ]) assert.throws(() => assertPublicationPolicy({ ...policy(), ...changed }));
});

test('schema two requires distinct complete evidence and binaries for every target', () => {
  validateNativeShape(manifest());
  for (const mutate of [
    (m) => { m.validation.targets.pop(); },
    (m) => { m.validation.targets[0].target = m.validation.targets[1].target; },
    (m) => { m.validation.targets[1].stages[0].path = m.validation.targets[0].stages[0].path; },
    (m) => { m.validation.targets[0].stages.pop(); },
    (m) => { m.artifacts.pop(); },
    (m) => { delete m.validation.targets[0].smoke; },
    (m) => { m.validation.targets[1].smoke = m.validation.targets[0].smoke; },
  ]) { const m = manifest(); mutate(m); assert.throws(() => validateNativeShape(m)); }
});

test('target evidence binds the actual native host and rejects reused reports', async (t) => {
  const root = await mkdtemp(join(tmpdir(), 'l-release-matrix-'));
  t.after(() => rm(root, { recursive: true, force: true }));
  const m = manifest();
  for (const artifact of m.artifacts) await writeFile(join(root, artifact.path), artifact.target);
  for (const group of m.validation.targets) for (const entry of group.stages) {
    const [goos, goarch] = group.target.split('-');
    const commands = { lint: ['gofmt', 'lint-config', 'golangci-ordinary', 'golangci-nativecompile', 'actionlint'], acceptance: ['doctor', 'verify'], native: ['nativecompile'] }[entry.stage];
    const report = { stage: entry.stage, goos, goarch, revision: SHA, dirty: false, passed: true, source_identity: HASH, started_at: '2026-10-26T00:00:00Z', finished_at: '2026-10-26T00:01:00Z', commands: commands.map((name) => ({ name, exit_code: 0, timed_out: false })) };
    await mkdir(join(root, 'evidence', group.target), { recursive: true });
    await writeFile(join(root, entry.path), stableJSON(report));
    m.evidence.push({ path: entry.path, sha256: sha256(stableJSON(report)) });
  }
  for (const group of m.validation.targets) {
    const smoke = { schemaVersion: 1, target: group.target, version: m.languageVersion, archiveSha256: m.artifacts.find((x) => x.target === group.target).sha256, passed: true, freshCargoHome: true, freshRustupHome: true, offlineBuild: true, startedAt: '2026-10-26T00:01:00Z', finishedAt: '2026-10-26T00:02:00Z' };
    await writeFile(join(root, group.smoke), stableJSON(smoke));
    m.evidence.push({ path: group.smoke, sha256: sha256(stableJSON(smoke)) });
  }
  assert.equal(await verifyNativeEvidence(m, root), true);
  const group = m.validation.targets[0], smokePath = join(root, group.smoke);
  const originalSmoke = { schemaVersion: 1, target: group.target, version: m.languageVersion, archiveSha256: m.artifacts.find((x) => x.target === group.target).sha256, passed: true, freshCargoHome: true, freshRustupHome: true, offlineBuild: true, startedAt: '2026-10-26T00:01:00Z', finishedAt: '2026-10-26T00:02:00Z' };
  for (const changed of [{ passed: false }, { freshCargoHome: false }, { freshRustupHome: false }, { offlineBuild: false }, { archiveSha256: HASH }, { version: '0.0.2' }, { target: 'linux-amd64' }, { finishedAt: 'invalid' }]) {
    const bytes = stableJSON({ ...originalSmoke, ...changed });
    await writeFile(smokePath, bytes);
    m.evidence.find((x) => x.path === group.smoke).sha256 = sha256(bytes);
    await assert.rejects(verifyNativeEvidence(m, root), /Installer smoke|installer smoke/);
  }
  await writeFile(smokePath, stableJSON(originalSmoke));
  m.evidence.find((x) => x.path === group.smoke).sha256 = sha256(stableJSON(originalSmoke));
  const smokeEvidence = m.evidence.find((x) => x.path === group.smoke);
  m.evidence = m.evidence.filter((x) => x.path !== group.smoke);
  await assert.rejects(verifyNativeEvidence(m, root), /Unhashed installer smoke/);
  m.evidence.push(smokeEvidence);
  const wrongHost = { stage: 'lint', goos: 'linux', goarch: 'amd64', revision: SHA, dirty: false, passed: true };
  const entry = m.validation.targets[0].stages[0];
  await writeFile(join(root, entry.path), stableJSON(wrongHost));
  m.evidence.find((x) => x.path === entry.path).sha256 = sha256(stableJSON(wrongHost));
  await assert.rejects(verifyNativeEvidence(m, root), /host differs/);
});

test('schema two cannot trust a push publisher or a partial target policy', () => {
  const m = manifest(), p = policy();
  const observations = { run: { repository: { full_name: m.repository }, id: 1, run_attempt: 1, workflow_id: 8, path: p.nativePublisher.workflowPath, head_sha: SHA, head_branch: 'main', event: 'push' } };
  assert.throws(() => verifyProducer(m, p, observations), /Untrusted native producer event/);
  assert.throws(() => verifyProducer(m, { ...p, targets: nativeTargets.slice(1) }, observations), /complete native target matrix/);
  assert.throws(() => verifyProducer(m, { ...p, publicationEnabled: false }, observations), /disabled/);
});

function workflow() {
  return { on: { workflow_dispatch: {} }, permissions: { contents: 'read' }, concurrency: { group: 'docs-production', 'cancel-in-progress': false }, jobs: { deploy: {
    permissions: { pages: 'write', 'id-token': 'write' }, steps: [
      { uses: `actions/checkout@${SHA}`, with: { 'persist-credentials': false } },
      { uses: `actions/setup-node@${SHA}` },
      { run: 'node scripts/release/github.mjs guard --policy website/ci-templates/release-policy.json --publication true' },
      { uses: `actions/deploy-pages@${SHA}` },
    ],
  } } };
}

test('approved manual workflows cannot bypass guards or publish before authorization', () => {
  verifyReleaseWorkflow(workflow(), 'docs-recover-pages.yml', { ...policy(), publicationEnabled: false, nativePublisher: null });
  for (const mutate of [
    (w) => { w.on.push = {}; },
    (w) => { w.concurrency['cancel-in-progress'] = true; },
    (w) => { w.jobs.deploy.steps[2].if = 'false'; },
    (w) => { w.jobs.deploy.steps[2]['continue-on-error'] = true; },
    (w) => { w.jobs.deploy.steps[2].run = 'echo disabled'; },
    (w) => { w.jobs.deploy.steps.unshift({ run: 'gh release create bad' }); },
    (w) => { w.jobs.deploy.steps[2]['working-directory'] = 'website'; },
  ]) { const w = workflow(); mutate(w); assert.throws(() => verifyReleaseWorkflow(w, 'docs-recover-pages.yml', policy())); }
});
