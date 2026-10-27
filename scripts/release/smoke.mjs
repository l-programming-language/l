// Run only on explicitly dispatched fresh release runners. Never part of npm's
// website checks or local acceptance-by-default.
import { readFile, writeFile, mkdir, mkdtemp, rm } from 'node:fs/promises';
import { createWriteStream } from 'node:fs';
import { join, resolve } from 'node:path';
import { tmpdir } from 'node:os';
import { spawn } from 'node:child_process';
import { once } from 'node:events';
import { sha256, stableJSON } from '../../website/scripts/release.mjs';

const directory = resolve(process.argv[2] || '');
if (process.argv.length !== 3) throw new Error('Usage: node scripts/release/smoke.mjs TARGET_ARTIFACT_DIRECTORY');
const record = JSON.parse(await readFile(join(directory, 'target.json'), 'utf8'));
const host = `${process.platform}-${process.arch === 'x64' ? 'amd64' : process.arch}`;
if (host !== record.target) throw new Error('Installer smoke must run on the packaged native target');
if (sha256(await readFile(join(directory, record.artifact.path))) !== record.artifact.sha256) throw new Error('Packaged archive changed');
const sandbox = await mkdtemp(join(tmpdir(), 'l-install-smoke-'));
const stdout = createWriteStream(join(directory, 'smoke.stdout.log'), { flags: 'wx' });
const stderr = createWriteStream(join(directory, 'smoke.stderr.log'), { flags: 'wx' });
const startedAt = new Date().toISOString();
const env = { ...process.env, CARGO_HOME: join(sandbox, 'cargo'), RUSTUP_HOME: join(sandbox, 'rustup'), XDG_CACHE_HOME: join(sandbox, 'cache'), RUSTUP_AUTO_INSTALL: '0' };
delete env.GITHUB_TOKEN;
delete env.GH_TOKEN;
delete env.APPLE_APP_PASSWORD;
delete env.MACOS_CERTIFICATE_P12;
delete env.MACOS_CERTIFICATE_PASSWORD;

async function run(command, args, timeout, extra = {}) {
  stdout.write(`\n$ ${command} ${args.join(' ')}\n`);
  const child = spawn(command, args, { cwd: sandbox, env: { ...env, ...extra }, detached: true, stdio: ['ignore', 'pipe', 'pipe'] });
  let output = '', error, timedOut = false;
  child.stdout.on('data', (data) => { output += data; stdout.write(data); });
  child.stderr.pipe(stderr, { end: false });
  child.on('error', (value) => { error = value; });
  const timer = setTimeout(() => {
    timedOut = true;
    try { process.kill(-child.pid, 'SIGKILL'); } catch { /* The process may already have exited. */ }
  }, timeout);
  const [code, signal] = await once(child, 'close').finally(() => clearTimeout(timer));
  if (error || timedOut || code !== 0) throw new Error(`${command} failed: ${error?.message || (timedOut ? 'watchdog expired' : signal || code)}`);
  return output;
}

let passed = false, failure;
try {
  await mkdir(join(sandbox, 'sources'));
  // This is a release-user example, independent of repository fixtures and Go.
  await writeFile(join(sandbox, 'sources/main.l'), 'Use IO\'s Print from `io` as "Print".\n\n"message" is `Hello World!`.\nCreate "Hello" as a New Basic App.\nThe App Prints message.\nExecute Hello.\nShow Hello\'s Output.\n');
  await run('bash', [join(directory, 'install-release.sh'), '--from-dir', directory, '--version', record.version, '--prefix', join(sandbox, 'bin'), '--with-toolchain'], 15 * 60 * 1000);
  env.PATH = `${join(sandbox, 'bin')}:${join(env.CARGO_HOME, 'bin')}:${process.env.PATH}`;
  const binary = join(sandbox, 'bin/l');
  if ((await run(binary, ['--version'], 10000)).trim() !== record.version) throw new Error('Installed CLI version differs');
  await run(binary, ['--help'], 10000);
  await run(binary, ['build', '--output-dir', join(sandbox, 'out'), join(sandbox, 'sources/main.l')], 4 * 60 * 1000, { CARGO_NET_OFFLINE: 'true' });
  if (await run(join(sandbox, 'out/main'), [], 30000, { CARGO_NET_OFFLINE: 'true' }) !== 'Hello World!\n') throw new Error('Installed compiler produced unexpected program output');
  passed = true;
} catch (error) { failure = error.message; stderr.write(`${error.stack}\n`); }
finally {
  stdout.end(); stderr.end();
  await Promise.all([once(stdout, 'finish'), once(stderr, 'finish')]);
  await writeFile(join(directory, 'smoke.json'), stableJSON({ schemaVersion: 1, target: record.target, version: record.version, archiveSha256: record.artifact.sha256, startedAt, finishedAt: new Date().toISOString(), freshCargoHome: true, freshRustupHome: true, offlineBuild: passed, passed, ...(failure ? { error: failure } : {}) }), { flag: 'wx' });
  await rm(sandbox, { recursive: true, force: true });
}
if (!passed) { console.error(`Installer smoke failed: ${failure}`); process.exitCode = 1; }
else console.log(`Installer smoke passed: ${record.target}`);
