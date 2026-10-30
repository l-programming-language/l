import test from 'node:test';
import assert from 'node:assert/strict';
import { readFile } from 'node:fs/promises';
import YAML from 'yaml';
import { verifyReleaseWorkflow } from '../scripts/verify-disabled.mjs';

const policy = JSON.parse(await readFile(new URL('../ci-templates/release-policy.json', import.meta.url)));
const readWorkflow = async (name) => YAML.parse(await readFile(new URL(`../../.github/workflows/${name}`, import.meta.url), 'utf8'));
const native = await readWorkflow('release.yml'), docs = await readWorkflow('docs-release-pages.yml');

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
