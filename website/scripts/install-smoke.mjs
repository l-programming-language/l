// Explicit opt-in only: downloads the published release, installs its Rust
// toolchain in a disposable user home, and compiles/runs the canonical demo.
// Never include this command in ordinary website verification.
import fs from 'node:fs/promises';
import os from 'node:os';
import path from 'node:path';
import { spawn } from 'node:child_process';
import { repository, website, sha256, stable } from './lib.mjs';

const previewURL = 'http://127.0.0.1:4321/install.sh';
const systemPath = '/usr/bin:/bin:/usr/sbin:/sbin';
const usage = 'Usage: npm run test:install\nRequires the built site running with npm run preview at http://127.0.0.1:4321.\nChecks the retained bootstrap and documented release installer using disposable homes, then builds and runs the canonical demo.\n';
if (process.argv.length > 2) {
  if (process.argv.length === 3 && process.argv[2] === '--help') console.log(usage);
  else { console.error(usage); process.exitCode = 1; }
} else {
  await smoke().catch((error) => { console.error(error.stack); process.exitCode ||= 1; });
}

async function smoke() {
  const target = `${process.platform}-${process.arch === 'x64' ? 'amd64' : process.arch}`;
  if (!['darwin-arm64', 'linux-arm64', 'linux-amd64'].includes(target)) throw new Error(`Unsupported installation smoke host: ${target}`);
  const version = (await fs.readFile(path.join(repository, 'VERSION'), 'utf8')).trim();
  const toolchain = JSON.parse(await fs.readFile(path.join(repository, '.ci-tools.json'), 'utf8')).rust;
  const bootstrap = await fs.readFile(path.join(website, 'public/install.sh'));
  const source = await fs.readFile(path.join(repository, 'demo/basic/main.l'));
  const releaseCommand = [
    "curl --fail --location --proto '=https' --tlsv1.2 \\",
    `  https://github.com/l-programming-language/l/releases/download/v${version}/install-release.sh \\`,
    '  --output install-release.sh &&',
    `bash ./install-release.sh --version ${version} --with-toolchain &&`,
    'export PATH="$HOME/.local/bin:${CARGO_HOME:-$HOME/.cargo}/bin:$PATH" &&',
    'l --version',
  ].join('\n');
  const installationGuide = await fs.readFile(path.join(repository, 'docs/cli/installation.md'), 'utf8');
  if (installationGuide.match(/```sh\n([\s\S]*?)\n```/)?.[1] !== releaseCommand) throw new Error('Review the documented release installation command before running this smoke.');
  if (!(await fs.readFile(path.join(website, 'dist/install.sh'))).equals(bootstrap)) throw new Error('Build the current website before running installation smoke.');
  const reports = path.join(website, 'reports');
  await fs.mkdir(reports, { recursive: true });
  const evidence = await fs.mkdtemp(path.join(reports, 'install-smoke-'));
  const sandbox = await fs.realpath(await fs.mkdtemp(path.join(os.tmpdir(), 'l-website-install-')));
  const testHome = path.join(sandbox, 'home');
  const prefix = path.join(sandbox, 'custom bin');
  const cargoHome = path.join(sandbox, 'toolchains/cargo');
  const rustupHome = path.join(sandbox, 'toolchains/rustup');
  const temp = path.join(sandbox, 'tmp');
  const stages = [];
  const report = {
    schemaVersion: 1, scope: 'explicit real-release installation and native runtime smoke',
    target, version, versionSelection: 'explicit --version from repository VERSION', expectedRustVersion: toolchain,
    previewURL, bootstrapSha256: sha256(bootstrap), source: 'demo/basic/main.l', sourceSha256: sha256(source),
    freshHome: true, freshCargoHome: true, freshRustupHome: true, freshNativeCache: true,
    startedAt: new Date().toISOString(), stages, passed: false,
  };
  // Deliberately construct the child environment: host credentials, startup
  // overrides, toolchain settings, and caches must not reach the test user.
  const baseEnv = {
    HOME: testHome, PATH: systemPath, TMPDIR: temp, SHELL: '/bin/bash', TERM: 'dumb', LC_ALL: 'C',
    XDG_CACHE_HOME: path.join(testHome, '.cache'), XDG_CONFIG_HOME: path.join(testHome, '.config'),
    RUSTUP_AUTO_INSTALL: '0',
  };
  let stopActive, interrupted;
  const interrupt = (signal) => { interrupted = signal; stopActive?.(); };
  const onInterrupt = () => interrupt('SIGINT');
  const onTerminate = () => interrupt('SIGTERM');
  process.on('SIGINT', onInterrupt); process.on('SIGTERM', onTerminate);

  async function run(name, command, args, timeoutMs, extraEnv = {}) {
    if (interrupted) throw new Error(`Installation smoke interrupted by ${interrupted}`);
    console.log(`Installation smoke: ${name}`);
    const startedAt = new Date().toISOString();
    const result = await new Promise((resolve) => {
      const child = spawn(command, args, { cwd: sandbox, env: { ...baseEnv, ...extraEnv }, detached: true, stdio: ['ignore', 'pipe', 'pipe'] });
      const stdout = [], stderr = []; let error, timedOut = false;
      function stop() {
        if (!child.pid) return;
        try { process.kill(-child.pid, 'SIGKILL'); }
        catch (failure) { if (failure.code !== 'ESRCH') error = failure; }
      }
      stopActive = stop;
      const timer = setTimeout(() => { timedOut = true; stop(); }, timeoutMs);
      child.stdout.on('data', (bytes) => stdout.push(bytes));
      child.stderr.on('data', (bytes) => stderr.push(bytes));
      child.on('error', (failure) => { error = failure; });
      child.on('close', (code, signal) => {
        clearTimeout(timer); stop(); stopActive = undefined;
        resolve({ code, signal, timedOut, ...(error ? { error: error.message } : {}), stdout: Buffer.concat(stdout), stderr: Buffer.concat(stderr) });
      });
    });
    const stdoutPath = `${name}.stdout.log`, stderrPath = `${name}.stderr.log`;
    await Promise.all([fs.writeFile(path.join(evidence, stdoutPath), result.stdout), fs.writeFile(path.join(evidence, stderrPath), result.stderr)]);
    stages.push({ name, command, args, timeoutMs, startedAt, finishedAt: new Date().toISOString(), exitCode: result.code, signal: result.signal, timedOut: result.timedOut, stdoutPath, stderrPath, ...(result.error ? { error: result.error } : {}) });
    if (interrupted || result.error || result.timedOut || result.code !== 0) throw new Error(`${name} failed: ${interrupted || result.error || (result.timedOut ? 'watchdog expired' : result.signal || result.code)}; see ${evidence}`);
    return result;
  }

  const resolved = 'set -eu\n[ "$(command -v l)" = "$1/l" ]\n[ "$(command -v cargo)" = "$3/bin/cargo" ]\n[ "$(command -v rustup)" = "$3/bin/rustup" ]\n[ "${CARGO_HOME-}" = "$3" ]\n[ "${RUSTUP_HOME-}" = "$4" ]\ncase $(cargo "+$5" --version) in "cargo $5 "*) ;; *) exit 1 ;; esac\nl --version\n';
  const assertVersion = (result, name) => {
    if (!result.stdout.equals(Buffer.from(`${version}\n`))) throw new Error(`${name}: installed version output differs from ${version}`);
  };
  try {
    await Promise.all([fs.mkdir(testHome, { recursive: true }), fs.mkdir(temp), fs.mkdir(path.join(sandbox, 'source'))]);
    await fs.writeFile(path.join(sandbox, 'source/main.l'), source);
    // Only the initial preview fetch permits loopback HTTP. The unchanged
    // bootstrap itself downloads the installer and release assets over HTTPS.
    const response = await fetch(previewURL, { redirect: 'error', signal: AbortSignal.timeout(30000) });
    if (!response.ok || !Buffer.from(await response.arrayBuffer()).equals(bootstrap)) throw new Error('Preview is not serving the exact current installation bootstrap.');
    const captured = path.join(sandbox, 'executed-install.sh');
    const cargoVersionPath = path.join(sandbox, 'activated-cargo-version');
    const installation = await run('install-and-activate', '/bin/bash', ['--noprofile', '--norc', '-c',
      'set -euo pipefail\ncurl --fail --silent --show-error --max-time 30 --proto "=http" "$1" | tee "$3" | sh -s -- --yes --version "$5" --prefix "$2"\n. "$HOME/.local/share/l/env"\n[ "$(command -v l)" = "$2/l" ]\n[ "$(command -v cargo)" = "$7/bin/cargo" ]\n[ "$(command -v rustup)" = "$7/bin/rustup" ]\n[ "$CARGO_HOME" = "$7" ]\n[ "$RUSTUP_HOME" = "$8" ]\nl --version > "$4"\ncargo "+$6" --version > "$9"\n',
      'l-install-smoke', previewURL, prefix, captured, path.join(sandbox, 'activated-version'), version, toolchain, cargoHome, rustupHome, cargoVersionPath], 15 * 60 * 1000, { CARGO_HOME: cargoHome, RUSTUP_HOME: rustupHome });
    if (!(await fs.readFile(captured)).equals(bootstrap)) throw new Error('Executed bootstrap differs from the reviewed website bytes.');
    if (!(await fs.readFile(path.join(sandbox, 'activated-version'))).equals(Buffer.from(`${version}\n`))) throw new Error('Immediate shell activation did not resolve the released L version.');
    report.cargoVersion = (await fs.readFile(cargoVersionPath, 'utf8')).trim();
    if (!report.cargoVersion.startsWith(`cargo ${toolchain} `)) throw new Error(`Immediate shell activation did not select pinned Cargo ${toolchain}.`);
    report.immediateActivation = installation.code === 0;
    // These shells start from the minimal PATH with no inherited Cargo/Rustup
    // homes: startup files must supply the persistent environment themselves.
    for (const [name, shell, args] of [
      ['bash-interactive', '/bin/bash', ['-i', '-c']],
      ['bash-login', '/bin/bash', ['--login', '-i', '-c']],
      ['zsh-interactive', '/bin/zsh', ['-i', '-c']],
      ['zsh-login', '/bin/zsh', ['-l', '-i', '-c']],
    ]) {
      if (shell === '/bin/zsh' && !await fs.access(shell, fs.constants.X_OK).then(() => true, () => false)) {
        if (process.platform === 'darwin') throw new Error('Supported macOS host is missing /bin/zsh.');
        stages.push({ name, status: 'unavailable', reason: 'Zsh is not installed on this Linux host' });
        continue;
      }
      assertVersion(await run(name, shell, [...args, resolved, 'l-shell-smoke', prefix, version, cargoHome, rustupHome, toolchain], 10000), name);
    }
    const output = path.join(sandbox, 'out');
    await run('offline-build', '/bin/bash', ['--noprofile', '--norc', '-c',
      'set -eu\n. "$HOME/.local/share/l/env"\nl build --output-dir "$1" "$2"\n',
      'l-build-smoke', output, path.join(sandbox, 'source/main.l')], 4 * 60 * 1000, { CARGO_NET_OFFLINE: 'true' });
    report.offlineBuild = true;
    const runtime = await run('runtime', path.join(output, 'main'), [], 30000, { CARGO_NET_OFFLINE: 'true' });
    if (!runtime.stdout.equals(Buffer.from('Hello World!\n')) || runtime.stderr.length !== 0) throw new Error('Runtime output differs from exact Hello World! newline with empty stderr.');
    report.runtimeVerified = true;
    // Preserve the bootstrap's clean-home observations above. The direct route
    // gets a separate home/cache but reuses only the prepared Rust dependencies.
    const directHome = path.join(sandbox, 'direct-home');
    await fs.mkdir(directHome);
    const directEnv = {
      HOME: directHome, CARGO_HOME: cargoHome, RUSTUP_HOME: rustupHome,
      XDG_CACHE_HOME: path.join(directHome, '.cache'), XDG_CONFIG_HOME: path.join(directHome, '.config'),
    };
    const direct = report.directInstaller = { freshHome: true, freshNativeCache: true, dependencySetup: 'reuses isolated bootstrap Cargo/Rustup homes', passed: false };
    await run('direct-install-and-path', '/bin/bash', ['--noprofile', '--norc', '-c',
      `set -eu\n${releaseCommand}\n[ "$(command -v l)" = "$HOME/.local/bin/l" ]\n[ "$(command -v cargo)" = "$CARGO_HOME/bin/cargo" ]\ncargo "+$1" --version\n`,
      'l-direct-install-smoke', toolchain], 15 * 60 * 1000, directEnv);
    direct.installerSha256 = sha256(await fs.readFile(path.join(sandbox, 'install-release.sh')));
    const directBinary = path.join(directHome, '.local/bin/l');
    const directBytes = await fs.readFile(directBinary);
    if (!directBytes.equals(await fs.readFile(path.join(prefix, 'l')))) throw new Error('Direct and bootstrap installations selected different compiler bytes.');
    direct.binarySha256 = sha256(directBytes);
    if (await fs.access(path.join(directHome, '.local/share/l/env')).then(() => true, () => false)) throw new Error('The direct release installer unexpectedly created a bootstrap activation file.');
    directEnv.PATH = `${path.dirname(directBinary)}:${path.join(cargoHome, 'bin')}:${systemPath}`;
    assertVersion(await run('direct-version', directBinary, ['--version'], 10000, directEnv), 'direct-version');
    const directOutput = path.join(sandbox, 'direct-out');
    await run('direct-offline-build', directBinary, ['build', '--output-dir', directOutput, path.join(sandbox, 'source/main.l')], 4 * 60 * 1000, { ...directEnv, CARGO_NET_OFFLINE: 'true' });
    const directRuntime = await run('direct-runtime', path.join(directOutput, 'main'), [], 30000, { ...directEnv, CARGO_NET_OFFLINE: 'true' });
    if (!directRuntime.stdout.equals(Buffer.from('Hello World!\n')) || directRuntime.stderr.length !== 0) throw new Error('Direct installation runtime output differs from exact Hello World! newline with empty stderr.');
    direct.offlineBuild = true;
    direct.runtimeVerified = true;
    direct.passed = true;
    report.passed = true;
  } catch (error) {
    report.error = error.message;
    process.exitCode = interrupted === 'SIGINT' ? 130 : interrupted ? 143 : 1;
  } finally {
    stopActive?.();
    process.off('SIGINT', onInterrupt); process.off('SIGTERM', onTerminate);
    try { await fs.rm(sandbox, { recursive: true, force: true }); report.cleanedUp = true; }
    catch (error) { report.cleanedUp = false; report.passed = false; report.cleanupError = error.message; process.exitCode = 1; }
    report.finishedAt = new Date().toISOString();
    await fs.writeFile(path.join(evidence, 'report.json'), stable(report));
  }
  if (!report.passed) throw new Error(`Installation smoke failed: ${report.error || report.cleanupError}; report: ${path.join(evidence, 'report.json')}`);
  console.log(`Installation smoke passed (${target}): ${path.join(evidence, 'report.json')}`);
}
