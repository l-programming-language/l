import assert from 'node:assert/strict';
import fs from 'node:fs/promises';
import os from 'node:os';
import path from 'node:path';
import { spawn } from 'node:child_process';
import { repository, website, walk, safeFile, sha256, stable, atomicWrite } from './lib.mjs';
import { fingerprints, inputInventory } from './provenance.mjs';

// Optional website-only smoke check. Run with the exact .node-version runtime:
// /path/to/node-24.21.0/bin/node scripts/archive-smoke.mjs
// Installed npm dependencies are copied into the temporary archive, excluding
// caches. This validates archive content/build behavior, not a second npm ci.
const pin = (await fs.readFile(path.join(website, '.node-version'), 'utf8')).trim();
if (process.versions.node !== pin) throw new Error(`Archive smoke requires Node ${pin}; current runtime is ${process.versions.node}. Invoke this script with the pinned Node binary.`);
await fs.access(path.join(website, 'node_modules/astro/bin/astro.mjs'));

const excluded = new Set(['.git', '.DS_Store', 'node_modules', 'dist', 'target', '.harness', '.cache', '.tools', '.astro', '.generated', 'reports', 'test-results', 'playwright-report']);
const docsFilter = (_name, entry) => entry.name !== 'assets' && !excluded.has(entry.name);
const websiteFilter = (name, entry) => !excluded.has(entry.name)
  && !name.startsWith('website/src/content/docs/docs/')
  && !name.startsWith('website/src/content/docs/releases/')
  && !['website/src/content/docs/docs', 'website/src/content/docs/releases'].includes(name)
  && !name.startsWith('website/public/content-assets/')
  && !['website/public/content-assets', 'website/public/version.json', 'website/public/robots.txt'].includes(name);

const language = await inputInventory(repository);
const sources = new Set(language);
for (const name of await walk(repository, 'docs', docsFilter)) sources.add(name);
for (const name of await walk(repository, 'website', websiteFilter)) sources.add(name);
for (const dir of ['.github', '.agents', 'internal', 'tools']) for (const name of await walk(repository, dir, (_name, entry) => !excluded.has(entry.name))) sources.add(name);
for (const entry of await fs.readdir(repository, { withFileTypes: true })) {
  if (entry.isFile() && !excluded.has(entry.name)) sources.add(entry.name);
}
const expectedSourceFingerprint = `sha256:${(await fingerprints(repository, language)).fingerprint}`;
const temporary = await fs.mkdtemp(path.join(os.tmpdir(), 'L documentation archive with spaces '));
const archiveRoot = await fs.realpath(temporary);
const archiveSite = path.join(archiveRoot, 'website');
const report = {
  scope: 'Website-only real static build from a Git-less source archive staged in a path containing spaces.',
  node: process.versions.node,
  dependencyMode: 'Existing locked node_modules copied into temporary staging without caches; dependency installation is separately validated by npm ci.',
  copiedFiles: sources.size,
  languageExecution: 'not run',
  commands: [],
  status: 'running',
};

async function run(script, args = []) {
  const command = [process.execPath, script, ...args];
  console.log(`Archive smoke: node ${script}${args.length ? ' ' + args.join(' ') : ''}`);
  const output = await new Promise((resolve, reject) => {
    const child = spawn(command[0], command.slice(1), {
      cwd: archiveSite, detached: process.platform !== 'win32',
      env: { ...process.env, ASTRO_TELEMETRY_DISABLED: '1', DOCS_CHANNEL: 'local' },
      stdio: ['ignore', 'pipe', 'pipe'],
    });
    const chunks = []; let timedOut = false;
    const terminate = () => { try { process.kill(process.platform === 'win32' ? child.pid : -child.pid, 'SIGKILL'); } catch (error) { if (error.code !== 'ESRCH') reject(error); } };
    const timer = setTimeout(() => { timedOut = true; terminate(); }, 180000);
    child.stdout.on('data', (chunk) => chunks.push(chunk));
    child.stderr.on('data', (chunk) => chunks.push(chunk));
    child.on('error', (error) => { clearTimeout(timer); reject(error); });
    child.on('close', (code, signal) => {
      clearTimeout(timer);
      const text = Buffer.concat(chunks).toString('utf8');
      report.commands.push({ command: `node ${script}${args.length ? ' ' + args.join(' ') : ''}`, exitCode: code, signal, timedOut });
      if (code !== 0 || timedOut) reject(new Error(`Archive command failed: ${script}; exit=${code}; signal=${signal}; timeout=${timedOut}\n${text.slice(-16000)}`));
      else resolve(text);
    });
  });
  return output;
}

try {
  const authoredHashes = new Map();
  for (const relative of [...sources].sort()) {
    const original = await safeFile(repository, relative);
    if ((await fs.lstat(original)).isSymbolicLink()) throw new Error(`Archive smoke refuses a canonical source symlink: ${relative}`);
    const bytes = await fs.readFile(original);
    const target = path.join(archiveRoot, relative);
    await fs.mkdir(path.dirname(target), { recursive: true });
    await fs.writeFile(target, bytes);
    authoredHashes.set(relative, sha256(bytes));
  }
  await fs.cp(path.join(website, 'node_modules'), path.join(archiveSite, 'node_modules'), {
    recursive: true, dereference: false, verbatimSymlinks: true,
    filter: (source) => !['.vite', '.cache', '.astro'].includes(path.basename(source)),
  });
  console.log('Archive smoke: temporary dependency copy completed; original node_modules no longer needed.');
  await assert.rejects(fs.access(path.join(archiveRoot, '.git')));
  await run('scripts/check.mjs');
  const owned = JSON.parse(await fs.readFile(path.join(archiveSite, '.generated/owned.json'), 'utf8'));
  const projectedBefore = new Map();
  for (const relative of owned) projectedBefore.set(relative, sha256(await fs.readFile(path.join(archiveSite, relative))));
  await run('scripts/prepare.mjs');
  assert.deepEqual(JSON.parse(await fs.readFile(path.join(archiveSite, '.generated/owned.json'), 'utf8')), owned);
  for (const [relative, hash] of projectedBefore) assert.equal(sha256(await fs.readFile(path.join(archiveSite, relative))), hash, `Nondeterministic projection: ${relative}`);
  const version = JSON.parse(await fs.readFile(path.join(archiveSite, 'public/version.json'), 'utf8'));
  for (const field of ['languageCommit', 'docsCommit', 'dirty', 'repository']) assert.equal(version[field], null, `Archive fabricated ${field}`);
  assert.equal(version.channel, 'local');
  assert.equal(version.sourceFingerprint, expectedSourceFingerprint);
  if (!sources.has('VERSION')) assert.equal(version.languageVersion, null);
  await run('scripts/astro.mjs', ['build']);
  await run('scripts/check-dist.mjs');
  await run('scripts/verify-disabled.mjs');
  for (const [relative, hash] of authoredHashes) assert.equal(sha256(await fs.readFile(path.join(archiveRoot, relative))), hash, `Archive build modified authored input: ${relative}`);
  const output = JSON.parse(await fs.readFile(path.join(archiveSite, 'reports/built-output.json'), 'utf8'));
  report.identity = version;
  report.staticOutput = { routes: output.routes, htmlFiles: output.htmlFiles, totalAssetBytes: output.totalAssetBytes, javascriptBytes: output.javascriptBytes };
  report.deterministicPreparation = 'passed';
  report.authoredInputPreservation = 'passed';
  report.status = 'passed';
  console.log(`Archive smoke passed: ${output.routes} routes, ${output.htmlFiles} static HTML files, null Git identity, deterministic preparation and unchanged authored inputs.`);
} catch (error) {
  report.status = 'failed';
  report.error = error.message;
  process.exitCode = 1;
  console.error(error.stack);
} finally {
  await fs.rm(archiveRoot, { recursive: true, force: true });
  report.temporaryArchiveCleanup = 'completed';
  await atomicWrite(path.join(website, 'reports/archive-smoke.json'), stable(report));
}
