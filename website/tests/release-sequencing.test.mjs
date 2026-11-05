import test from 'node:test';
import assert from 'node:assert/strict';
import { readFile } from 'node:fs/promises';
import YAML from 'yaml';
import { verifyReleaseWorkflow } from '../scripts/verify-disabled.mjs';

const policy = JSON.parse(await readFile(new URL('../ci-templates/release-policy.json', import.meta.url)));
const readWorkflow = async (name) => YAML.parse(await readFile(new URL(`../../.github/workflows/${name}`, import.meta.url), 'utf8'));
const native = await readWorkflow('release.yml'), docs = await readWorkflow('docs-release-pages.yml');
const recovery = await readWorkflow('docs-recover-pages.yml');

test('manual release seals native and docs bytes before its sole publication transaction', () => {
  verifyReleaseWorkflow(native, 'release.yml', policy);
  verifyReleaseWorkflow(docs, 'docs-release-pages.yml', policy);
  assert.equal(native.jobs.publish, undefined);
  assert.deepEqual(native.jobs.documentation.needs, ['preflight', 'seal']);
  assert.equal(docs.jobs.publish.environment, "${{ inputs.mode == 'release' && 'release-publication' || 'documentation-publication' }}");
  const commands = docs.jobs.publish.steps.map((step) => step.run || '').join('\n');
  const publication = commands.indexOf('node scripts/release/github.mjs publish ');
  assert(publication > commands.indexOf('node scripts/release.mjs verify-promotion '));
  assert(publication > commands.indexOf('node scripts/release.mjs guard-promotion '));
  assert.match(commands, /github\.mjs publish[^\n]+--docs "\$RUNNER_TEMP\/sealed" --dist "\$RUNNER_TEMP\/site"/);
  assert.match(commands, /release\.mjs publish-docs[^\n]+--dist "\$RUNNER_TEMP\/site"/);
  assert.match(commands, /download-native[^\n]+--mode release/);
  const prepare = docs.jobs.prepare.steps.map((step) => step.run || '').join('\n');
  assert.match(prepare, /download-native[^\n]+--mode "\$DOCS_RELEASE_MODE"/);
  assert.match(prepare, /verify-native[^\n]+--mode "\$DOCS_RELEASE_MODE"/);
  assert.deepEqual(docs.on.workflow_dispatch.inputs.mode.options, ['docs-correction']);
});

test('policy rejects early publication and bypassed seal dependencies', () => {
  for (const mutate of [
    (w) => { w.jobs.documentation.needs = ['preflight']; },
    (w) => { delete w.jobs.documentation.if; },
    (w) => { w.jobs.documentation.with.native_artifact_id = 'untrusted'; },
    (w) => { w.jobs.publish = { permissions: { contents: 'write' }, steps: [] }; },
  ]) {
    const changed = structuredClone(native); mutate(changed);
    assert.throws(() => verifyReleaseWorkflow(changed, 'release.yml', policy));
  }
  for (const mutate of [
    (w) => { delete w.jobs.publish.needs; },
    (w) => { w.jobs.deploy.needs = ['prepare']; },
    (w) => { w.jobs.prepare.permissions.contents = 'write'; },
  ]) {
    const changed = structuredClone(docs); mutate(changed);
    assert.throws(() => verifyReleaseWorkflow(changed, 'docs-release-pages.yml', policy));
  }
});

test('deployment consumes the published immutable archive and compares its sealed identity', () => {
  const steps = docs.jobs.deploy.steps;
  const verify = steps.findIndex((step) => step.run?.includes('download-promotion '));
  const upload = steps.findIndex((step) => step.uses?.startsWith('actions/upload-pages-artifact@'));
  assert(verify >= 0 && upload > verify);
  assert.match(steps[verify].run, /cmp "\$RUNNER_TEMP\/sealed\/promotion\.json" "\$RUNNER_TEMP\/published\/promotion\.json"/);
  assert.match(steps[verify].run, /tar -xzf "\$RUNNER_TEMP\/published\/site\.tar\.gz"/);
  assert.match(steps[verify].run, /verify-promotion[^\n]+--manifest "\$RUNNER_TEMP\/published\/promotion\.json"/);
});

test('postdeployment verification uses the current Pages deployment without waiting for its environment job', () => {
  const postdeploy = docs.jobs.deploy.steps.find((step) => step.run?.includes('release.mjs postdeploy '));
  assert.match(postdeploy.run, /--deployment-sha "\$GITHUB_SHA"/);
  for (const job of [docs.jobs.prepare, docs.jobs.publish, docs.jobs.deploy, recovery.jobs.recover]) {
    assert(['read', 'write'].includes(job.permissions.pages), 'Production reconciliation needs Pages API access');
  }
});

test('already-live recovery verifies without uploading or deploying again', () => {
  verifyReleaseWorkflow(recovery, 'docs-recover-pages.yml', policy);
  const job = recovery.jobs.recover, steps = job.steps;
  const plan = steps.find((step) => step.id === 'plan');
  assert(plan, 'Recovery must publish its checked promotion decision');
  assert.match(plan.run, /guard-promotion[^\n]+--output "\$RUNNER_TEMP\/promotion-plan\.json"/);
  assert.match(plan.run, /GITHUB_OUTPUT/);
  for (const action of ['upload-pages-artifact', 'deploy-pages']) {
    const step = steps.find((item) => item.uses?.startsWith(`actions/${action}@`));
    assert.equal(step.if, "steps.plan.outputs.action == 'deploy'");
    assert(steps.indexOf(step) > steps.indexOf(plan));
  }
  const verify = steps.find((step) => step.run?.includes('release.mjs postdeploy '));
  assert.equal(verify.if, undefined, 'Live verification is required for both plan outcomes');
  assert.equal(verify.env.PROMOTION_ACTION, '${{ steps.plan.outputs.action }}');
  assert.match(verify.run, /if \[\[ "\$PROMOTION_ACTION" == deploy \]\]; then/);
  assert.match(verify.run, /--deployment-sha "\$GITHUB_SHA"/);
  assert.equal(job.environment.url, 'https://ldocs.org');
});

test('workflow policy rejects recovery that bypasses its promotion decision or live verification', () => {
  for (const mutate of [
    (w) => { delete w.jobs.recover.steps.find((step) => step.uses?.startsWith('actions/upload-pages-artifact@')).if; },
    (w) => { delete w.jobs.recover.steps.find((step) => step.uses?.startsWith('actions/deploy-pages@')).if; },
    (w) => { w.jobs.recover.steps.find((step) => step.run?.includes('release.mjs postdeploy ')).if = "steps.plan.outputs.action == 'deploy'"; },
    (w) => { delete w.jobs.recover.steps.find((step) => step.id === 'plan')?.id; },
    (w) => { delete w.jobs.recover.permissions.pages; },
  ]) {
    const changed = structuredClone(recovery); mutate(changed);
    assert.throws(() => verifyReleaseWorkflow(changed, 'docs-recover-pages.yml', policy));
  }
});
