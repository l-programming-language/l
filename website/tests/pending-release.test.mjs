import test from 'node:test';
import assert from 'node:assert/strict';
import { mkdtemp, writeFile, rm } from 'node:fs/promises';
import { tmpdir } from 'node:os';
import { join } from 'node:path';
import { verifyPendingProducer, verifyProducer, nativeTargets, documentationTag, documentationFiles, planRelease, sha256, stableJSON } from '../scripts/release.mjs';

const SHA = 'a'.repeat(40), OTHER = 'b'.repeat(40), HASH = 'c'.repeat(64);
function pending() {
  const manifest = {
    schemaVersion: 2, repository: 'l-programming-language/l', languageVersion: '0.0.1', tag: 'v0.0.1', languageCommit: SHA,
    sourceFingerprint: `sha256:${HASH}`, releaseFingerprint: HASH, dirty: false,
    producer: { runId: 12, runAttempt: 1, jobId: 22 }, evidence: [],
    artifacts: nativeTargets.map((target) => ({ path: `${target}.tgz`, releaseAsset: `${target}.tgz`, target, compilerVersion: 'go1.27.3', sha256: HASH })),
    validation: { command: 'scripts/l-check', targets: nativeTargets.map((target) => ({ target, smoke: `${target}/smoke.json`, stages: ['lint', 'acceptance', 'native'].map((stage) => ({ stage, path: `${target}/${stage}.json` })) })) },
  };
  const policy = { repository: manifest.repository, productionBranch: 'main', publicationEnabled: true, targets: nativeTargets, nativePublisher: { workflowId: 42, workflowPath: '.github/workflows/release.yml' } };
  const observations = {
    run: { id: 12, run_attempt: 1, head_sha: SHA, head_branch: 'main', event: 'workflow_dispatch', status: 'in_progress', path: policy.nativePublisher.workflowPath, workflow_id: 42, repository: { full_name: manifest.repository } },
    job: { id: 22, run_id: 12, run_attempt: 1, head_sha: SHA, name: 'Seal native release', status: 'completed', conclusion: 'success' },
    artifact: { name: `l-native-release-${SHA}`, workflow_run: { id: 12, head_sha: SHA }, expired: false, digest: `sha256:${HASH}` }, downloadedArchiveSha256: HASH,
  };
  const env = { GITHUB_REPOSITORY: manifest.repository, GITHUB_REF: 'refs/heads/main', GITHUB_EVENT_NAME: 'workflow_dispatch', GITHUB_SHA: SHA, GITHUB_RUN_ID: '12', GITHUB_RUN_ATTEMPT: '2' };
  return { manifest, policy, observations, env };
}

test('initial docs authenticate successful sealed proof from the same manual dispatch without a published release', () => {
  const f = pending();
  assert.equal(verifyPendingProducer(f.manifest, f.policy, f.observations, f.env), true);
  assert.throws(() => verifyProducer(f.manifest, f.policy, f.observations));
  for (const change of [{ GITHUB_RUN_ID: '13' }, { GITHUB_RUN_ATTEMPT: '0' }, { GITHUB_SHA: OTHER }, { GITHUB_REPOSITORY: 'fork/l' }, { GITHUB_REF: 'refs/heads/topic' }, { GITHUB_EVENT_NAME: 'push' }]) {
    assert.throws(() => verifyPendingProducer(f.manifest, f.policy, f.observations, { ...f.env, ...change }));
  }
  for (const mutate of [
    (o) => { o.run.workflow_id++; }, (o) => { o.run.path = '.github/workflows/docs-release-pages.yml'; },
    (o) => { o.job.conclusion = 'failure'; }, (o) => { o.job.name = 'Other job'; },
    (o) => { o.artifact.expired = true; }, (o) => { o.artifact.workflow_run.id++; },
    (o) => { o.artifact.digest = `sha256:${'d'.repeat(64)}`; },
  ]) { const changed = structuredClone(f.observations); mutate(changed); assert.throws(() => verifyPendingProducer(f.manifest, f.policy, changed, f.env)); }
});

test('initial docs use the language release; corrections use an exact separate snapshot identity', () => {
  const identity = { languageVersion: '0.0.1', languageCommit: SHA, docsCommit: SHA };
  assert.equal(documentationTag(identity), 'v0.0.1');
  assert.equal(documentationTag({ ...identity, docsCommit: OTHER }), `docs-v0.0.1-${OTHER}`);
  for (const change of [{ docsCommit: 'main' }, { languageCommit: '' }, { languageVersion: '../bad' }]) assert.throws(() => documentationTag({ ...identity, ...change }));
  assert.throws(() => planRelease({ mode: 'docs-correction', docsCommit: SHA, languageCommit: SHA, version: '0.0.1', tag: 'v0.0.1', dirty: false, sourceFingerprint: `sha256:${HASH}`, currentFingerprint: HASH, releasedFingerprint: HASH }), /distinct documentation revision/);
});

test('documentation publication binds all three original sealed files and rejects changed bytes', async (t) => {
  const dir = await mkdtemp(join(tmpdir(), 'l-docs-seal-'));
  t.after(() => rm(dir, { recursive: true, force: true }));
  const native = stableJSON(pending().manifest), archive = Buffer.from('sealed archive');
  const manifest = { channel: 'production', docsCommit: SHA, nativeManifestSha256: sha256(native), siteArchiveSha256: sha256(archive) };
  const paths = ['site.tar.gz', 'native.json', 'promotion.json'].map((name) => join(dir, name));
  const bytes = [archive, native, stableJSON(manifest)];
  for (let i = 0; i < paths.length; i++) await writeFile(paths[i], bytes[i]);
  const files = await documentationFiles(manifest, ...paths);
  assert.deepEqual([...files.keys()], [`docs-${SHA}.tar.gz`, `docs-${SHA}.promotion.json`, `docs-${SHA}.native.json`]);
  for (let i = 0; i < paths.length; i++) {
    await writeFile(paths[i], 'changed');
    await assert.rejects(documentationFiles(manifest, ...paths), /Sealed documentation changed/);
    await writeFile(paths[i], bytes[i]);
  }
});
