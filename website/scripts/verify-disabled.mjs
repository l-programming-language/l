import { readFile, readdir, lstat } from 'node:fs/promises';
import { resolve, join } from 'node:path';
import { fileURLToPath } from 'node:url';
import YAML from 'yaml';
import { nativeTargets } from './release.mjs';

const script = fileURLToPath(import.meta.url);
const fail = (message) => { throw new Error(message); };
async function exists(path) { try { await lstat(path); return true; } catch (e) { if (e.code === 'ENOENT') return false; throw e; } }

function verifyActions(value, name) {
  for (const job of Object.values(value.jobs)) {
    for (const step of job.steps || []) {
      if (step.uses && !step.uses.startsWith('./') && !/^[\w.-]+\/[\w./-]+@[a-f0-9]{40}$/.test(step.uses)) fail(`Unpinned action in ${name}: ${step.uses}`);
      if (step.uses?.startsWith('actions/checkout@') && step.with?.['persist-credentials'] !== false) fail(`Checkout persists credentials: ${name}`);
    }
  }
}

function verifyReadOnlyChecks(value, text, name) {
  if (!value || !value.on || !value.jobs) fail(`Invalid active check workflow: ${name}`);
  const events = typeof value.on === 'string' ? [value.on] : Array.isArray(value.on) ? value.on : Object.keys(value.on);
  if (!events.length || events.some((event) => !['pull_request', 'workflow_dispatch'].includes(event))) fail(`Unexpected event in website checks: ${name}`);
  const readOnly = (permissions) => permissions && typeof permissions === 'object' && !Array.isArray(permissions)
    && Object.values(permissions).every((permission) => ['read', 'none'].includes(permission));
  if (!readOnly(value.permissions)) fail(`Website checks require explicit read-only permissions: ${name}`);
  for (const job of Object.values(value.jobs)) {
    if (job.permissions && !readOnly(job.permissions)) fail(`Website check job has write permissions: ${name}`);
    if (job.environment || job.uses || job.secrets) fail(`Website checks cannot use environments, reusable workflows, or secrets: ${name}`);
  }
  if (/secrets\s*[.\[]|actions\/(?:configure-pages|upload-pages-artifact|deploy-pages)@|(?:docs-release-pages|docs-recover-pages)\.ya?ml|scripts\/release\.mjs/i.test(text)) fail(`Deployment or secrets are not permitted in website checks: ${name}`);
  verifyActions(value, name);
}

const releaseWorkflows = ['release.yml', 'docs-release-pages.yml', 'docs-recover-pages.yml'];
const publicationGuard = 'node scripts/release/github.mjs guard --policy website/ci-templates/release-policy.json --publication true';
const needs = (job) => [job?.needs].flat().filter(Boolean).sort().join(',');

export function verifyReleaseWorkflow(value, name, policy) {
  if (!policy || !value?.on || !value.jobs) fail(`Website deployment automation is active without a reviewed policy: ${name}`);
  const events = typeof value.on === 'string' ? [value.on] : Array.isArray(value.on) ? value.on : Object.keys(value.on);
  const allowed = name === 'docs-release-pages.yml' ? ['workflow_dispatch', 'workflow_call'] : ['workflow_dispatch'];
  if (!events.length || !events.includes('workflow_dispatch') || events.some((event) => !allowed.includes(event))) fail(`Release workflow must be manual-only: ${name}`);
  if (typeof policy.publicationEnabled !== 'boolean' || policy.productionBranch !== 'main' || policy.repository !== 'l-programming-language/l') fail('Invalid committed release policy');
  if (!Array.isArray(policy.targets) || [...policy.targets].sort().join(',') !== [...nativeTargets].sort().join(',')) fail('Release policy must retain all four native targets');
  if (policy.publicationEnabled && (!Number.isSafeInteger(policy.nativePublisher?.workflowId) || policy.nativePublisher.workflowId < 1 || policy.nativePublisher.workflowPath !== '.github/workflows/release.yml')) fail('Enabled publication lacks registered publisher');
  if (!value.permissions || Object.values(value.permissions).some((p) => !['read', 'none'].includes(p))) fail(`Release default permissions must be read-only: ${name}`);
  if (value.concurrency?.['cancel-in-progress'] !== false || !value.concurrency.group) fail(`Release concurrency must preserve active publication: ${name}`);
  if (name === 'release.yml') {
    const docs = value.jobs.documentation;
    if (docs?.uses !== './.github/workflows/docs-release-pages.yml' || needs(docs) !== 'preflight,seal' || docs?.if !== "needs.preflight.outputs.publish == 'true'" || docs?.with?.mode !== 'release' || docs?.with?.native_artifact_id !== '${{ needs.seal.outputs.artifact_id }}') fail('Native release must pass sealed proof to documentation before publication');
  } else if (name === 'docs-release-pages.yml') {
    if (needs(value.jobs.publish) !== 'prepare' || needs(value.jobs.deploy) !== 'prepare,publish') fail('Documentation must seal before publication and publish before deployment');
    if (Object.values(value.jobs.prepare?.permissions || {}).includes('write')) fail('Documentation preparation cannot publish');
  }
  for (const [id, job] of Object.entries(value.jobs)) {
    if (job.uses) {
      if (name !== 'release.yml' || id !== 'documentation' || job.uses !== './.github/workflows/docs-release-pages.yml') fail(`Unexpected reusable release job: ${name}/${id}`);
      continue;
    }
    const writes = Object.values(job.permissions || {}).some((permission) => permission === 'write');
    if (name === 'release.yml' && writes) fail('Native publication belongs after documentation sealing in the reusable workflow');
    if (name !== 'release.yml' || writes) {
      const steps = job.steps || [], guardIndex = steps.findIndex((step) => step.run?.trim() === publicationGuard);
      if (guardIndex < 0) fail(`Missing publication authorization: ${name}/${id}`);
      const guard = steps[guardIndex];
      if (guard.if !== undefined || guard['continue-on-error'] || (guard['working-directory'] || job.defaults?.run?.['working-directory'] || '.') !== '.') fail(`Publication authorization can be bypassed: ${name}/${id}`);
      if (steps.slice(0, guardIndex).some((step) => !/^actions\/(checkout|setup-node)@[a-f0-9]{40}$/.test(step.uses || ''))) fail(`Publication work precedes authorization: ${name}/${id}`);
    }
  }
  verifyActions(value, name);
}

export async function verifyDisabled({ root = resolve(script, '../../..') } = {}) {
  const site = join(root, 'website'), templateDir = join(site, 'ci-templates');
  const names = await readdir(templateDir);
  const templates = names.filter((n) => /\.ya?ml\.disabled$/.test(n));
  if (!['docs-check.yml.disabled', 'docs-release-pages.yml.disabled', 'docs-recover-pages.yml.disabled'].every((name) => templates.includes(name))) fail('Required dormant website workflow templates are missing');
  for (const name of names) if (/\.ya?ml$/.test(name)) fail(`Website template is not physically disabled: ${name}`);
  for (const name of templates) {
    if ((await lstat(join(templateDir, name))).isSymbolicLink()) fail(`Workflow template must be a regular owned file: ${name}`);
    const value = YAML.parse(await readFile(join(templateDir, name), 'utf8'), { uniqueKeys: true });
    if (!value || typeof value.name !== 'string' || !value.on || !value.jobs) fail(`Invalid workflow template: ${name}`);
    if (Object.hasOwn(value.on, 'pull_request_target')) fail('Website templates must never use privileged pull_request_target');
    verifyActions(value, name);
  }
  const workflows = join(root, '.github/workflows');
  const policyPath = join(templateDir, 'release-policy.json');
  const policy = await exists(policyPath) ? JSON.parse(await readFile(policyPath, 'utf8')) : null;
  if (!await exists(join(workflows, 'docs-check.yml'))) fail('Required active documentation checks are missing');
  if (await exists(workflows)) for (const name of await readdir(workflows)) {
    if (!/\.ya?ml$/.test(name)) continue;
    if ((await lstat(join(workflows, name))).isSymbolicLink()) fail(`Workflow must be a regular owned file: ${name}`);
    const text = await readFile(join(workflows, name), 'utf8');
    if (['docs-check.yml', 'quality.yml'].includes(name)) {
      verifyReadOnlyChecks(YAML.parse(text, { uniqueKeys: true }), text, name);
      continue;
    }
    if (releaseWorkflows.includes(name)) {
      verifyReleaseWorkflow(YAML.parse(text, { uniqueKeys: true }), name, policy);
      continue;
    }
    if (/^docs-(?:release|recover)-pages\.ya?ml$/.test(name)) fail(`Website deployment automation is active in .github/workflows/${name}`);
    if (/website\/|working-directory:\s*['"]?\.?\/?website\b|\b(?:cd|pushd)\s+['"]?(?:\.\/)?website\b|--prefix(?:=|\s+)['"]?(?:\.\/)?website\b|(?:docs-check|docs-release-pages|docs-recover-pages)\.ya?ml|L documentation (?:checks|release|recovery)/i.test(text)) fail(`Website automation is active in .github/workflows/${name}`);
  }
  const dependabot = join(root, '.github/dependabot.yml');
  if (await exists(dependabot)) {
    const value = YAML.parse(await readFile(dependabot, 'utf8'));
    if ((value.updates || []).some((item) => item['package-ecosystem'] === 'npm' && [item.directory, ...(item.directories || [])].some((dir) => dir === '/website' || dir === '/website/' || dir === '/**'))) fail('Website npm Dependabot automation is active');
  }
  const pkg = JSON.parse(await readFile(join(site, 'package.json'), 'utf8'));
  for (const hook of ['preinstall', 'install', 'postinstall', 'prepare', 'prepublish', 'prepublishOnly', 'publish', 'postpublish']) if (pkg.scripts?.[hook]) fail(`Automatic npm lifecycle hook is not permitted: ${hook}`);
  for (const [name, command] of Object.entries(pkg.scripts || {})) if (/\.github[\\/]workflows/.test(command)) fail(`Website npm script addresses active workflows: ${name}`);
  for (const name of await readdir(join(site, 'scripts'))) {
    if (!/\.[cm]?js$/.test(name) || name === 'verify-disabled.mjs') continue;
    const text = await readFile(join(site, 'scripts', name), 'utf8');
    if (/(?:copyFile|rename|writeFile|cp\s|mv\s)[^\n]{0,200}\.github\/workflows/.test(text)) fail(`Possible workflow activation helper: ${name}`);
  }
  return { templates: templates.length, state: 'checks-active', deployment: policy?.publicationEnabled ? 'manual-only' : 'disabled' };
}

if (process.argv[1] && resolve(process.argv[1]) === script) verifyDisabled().then((result) => console.log(`Website checks are active; publication policy: ${result.deployment}. Checked ${result.templates} retained templates; no deployment performed.`)).catch((error) => { console.error(error.message); process.exitCode = 1; });
