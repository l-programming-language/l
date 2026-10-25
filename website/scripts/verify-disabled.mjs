import { readFile, readdir, lstat } from 'node:fs/promises';
import { resolve, join } from 'node:path';
import { fileURLToPath } from 'node:url';
import YAML from 'yaml';

const script = fileURLToPath(import.meta.url);
const fail = (message) => { throw new Error(message); };
async function exists(path) { try { await lstat(path); return true; } catch (e) { if (e.code === 'ENOENT') return false; throw e; } }

export async function verifyDisabled({ root = resolve(script, '../../..') } = {}) {
  const site = join(root, 'website'), templateDir = join(site, 'ci-templates');
  const names = await readdir(templateDir);
  const templates = names.filter((n) => /\.ya?ml\.disabled$/.test(n));
  if (!templates.includes('docs-check.yml.disabled') || !templates.includes('docs-release-pages.yml.disabled')) fail('Required dormant website workflow templates are missing');
  for (const name of names) if (/\.ya?ml$/.test(name)) fail(`Website template is not physically disabled: ${name}`);
  for (const name of templates) {
    if ((await lstat(join(templateDir, name))).isSymbolicLink()) fail(`Workflow template must be a regular owned file: ${name}`);
    const value = YAML.parse(await readFile(join(templateDir, name), 'utf8'), { uniqueKeys: true });
    if (!value || typeof value.name !== 'string' || !value.on || !value.jobs) fail(`Invalid workflow template: ${name}`);
    if (Object.hasOwn(value.on, 'pull_request_target')) fail('Website templates must never use privileged pull_request_target');
    for (const job of Object.values(value.jobs)) {
      for (const step of job.steps || []) {
        if (step.uses && !step.uses.startsWith('./') && !/^[\w.-]+\/[\w./-]+@[a-f0-9]{40}$/.test(step.uses)) fail(`Unpinned action in ${name}: ${step.uses}`);
        if (step.uses?.startsWith('actions/checkout@') && step.with?.['persist-credentials'] !== false) fail(`Checkout persists credentials: ${name}`);
      }
    }
  }
  const workflows = join(root, '.github/workflows');
  if (await exists(workflows)) for (const name of await readdir(workflows)) {
    if (!/\.ya?ml$/.test(name)) continue;
    const text = await readFile(join(workflows, name), 'utf8');
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
  return { templates: templates.length, state: 'disabled', deployment: 'not performed by this check' };
}

if (process.argv[1] && resolve(process.argv[1]) === script) verifyDisabled().then((result) => console.log(`Website CI/CD is prepared but disabled. Checked ${result.templates} dormant templates; no deployment performed.`)).catch((error) => { console.error(error.message); process.exitCode = 1; });
