import test from 'node:test';
import assert from 'node:assert/strict';
import { mkdtemp, mkdir, writeFile, readFile, rm, chmod, symlink, lstat, stat, realpath } from 'node:fs/promises';
import { existsSync } from 'node:fs';
import { tmpdir } from 'node:os';
import { join, dirname } from 'node:path';
import { fileURLToPath } from 'node:url';
import { spawn, spawnSync } from 'node:child_process';

const bootstrap = fileURLToPath(new URL('../public/install.sh', import.meta.url));
const environmentHeader = '# L shell environment (managed by the L installer)';
const systemPath = '/usr/bin:/bin:/usr/sbin:/sbin';
const quote = (value) => `'${value.replaceAll("'", "'\\''")}'`;

// Every network request and the downloaded release installer are fixtures. The
// real bootstrap still verifies bytes, delegates through Bash, and edits actual
// temporary files; no compiler, package manager, or real user profile is used.
const fixtureHandler = String.raw`
import fs from 'node:fs';
import path from 'node:path';
import { createHash } from 'node:crypto';
import { spawnSync } from 'node:child_process';
const root = process.env.L_TEST_ROOT;
const [mode, ...args] = process.argv.slice(2);
const record = (value) => fs.appendFileSync(path.join(root, 'events.jsonl'), JSON.stringify(value) + '\n');
const executable = (file, source) => { fs.mkdirSync(path.dirname(file), { recursive: true }); fs.writeFileSync(file, source, { mode: 0o755 }); };
const value = (flag) => args[args.indexOf(flag) + 1];
if (mode === 'curl') {
  const url = args.find((arg) => /^https?:\/\//.test(arg));
  record({ mode, args, url });
  if (process.env.L_TEST_FAIL_DOWNLOAD && url?.endsWith(process.env.L_TEST_FAIL_DOWNLOAD)) process.exit(22);
  let body;
  if (url === 'https://ldocs.org/install.sh') body = fs.readFileSync(process.env.L_TEST_BOOTSTRAP);
  else if (url === 'https://github.com/l-programming-language/l/releases/latest') {
    const previous = fs.readFileSync(path.join(root, 'events.jsonl'), 'utf8').trim().split('\n').map(JSON.parse).filter((event) => event.url === url).length;
    body = process.env.L_TEST_LATEST_URL || 'https://github.com/l-programming-language/l/releases/tag/v' + (previous === 1 ? '0.0.1' : '0.0.2');
    process.stdout.write(body); process.exit(0);
  } else if (/^https:\/\/github\.com\/l-programming-language\/l\/releases\/download\/v[0-9]+\.[0-9]+\.[0-9]+\/(checksums\.txt|install-release\.sh)$/.test(url ?? '')) {
    const installer = fs.readFileSync(path.join(root, 'release-installer.sh'));
    if (url.endsWith('/install-release.sh')) body = installer;
    else {
      const digest = createHash('sha256').update(installer).digest('hex');
      const checksum = digest + '  install-release.sh\n';
      switch (process.env.L_TEST_CHECKSUM) {
        case 'mismatch': body = '0'.repeat(64) + '  install-release.sh\n'; break;
        case 'missing': body = digest + '  other.sh\n'; break;
        case 'duplicate': body = checksum + checksum; break;
        case 'malformed': body = 'not-a-sha256  install-release.sh\n'; break;
        default: body = checksum;
      }
    }
  } else { console.error('Fixture rejected unexpected network URL: ' + url); process.exit(91); }
  const outputIndex = args.findIndex((arg) => arg === '--output' || arg === '-o');
  if (process.env.L_TEST_PARTIAL_DOWNLOAD && url.endsWith(process.env.L_TEST_PARTIAL_DOWNLOAD)) {
    const partial = body.subarray ? body.subarray(0, Math.floor(body.length / 2)) : body.slice(0, Math.floor(body.length / 2));
    if (outputIndex >= 0) fs.writeFileSync(args[outputIndex + 1], partial);
    else process.stdout.write(partial);
    process.exit(18);
  }
  if (outputIndex >= 0) fs.writeFileSync(args[outputIndex + 1], body);
  else process.stdout.write(body);
} else if (mode === 'delegate') {
  record({ mode, args, bash: process.env.L_TEST_DELEGATE_BASH });
  if (process.env.L_TEST_DELEGATE_FAIL) process.exit(42);
  const prefix = value('--prefix'), version = value('--version');
  if (!prefix || !version || !args.includes('--with-toolchain')) process.exit(92);
  executable(path.join(prefix, 'l'), '#!/bin/sh\n[ "$1" = --version ] || exit 2\nprintf "%s\\n" "' + version + '"\n');
  const cargo = process.env.CARGO_HOME || path.join(process.env.HOME, '.cargo');
  executable(path.join(cargo, 'bin/cargo'), '#!/bin/sh\nprintf "%s\\n" "cargo 1.98.1 (fixture)"\n');
  executable(path.join(cargo, 'bin/rustup'), '#!/bin/sh\nexit 0\n');
  if (process.env.L_TEST_MUTATE_PROFILE) fs.appendFileSync(process.env.L_TEST_MUTATE_PROFILE, '# concurrent user edit\n');
  if (process.env.L_TEST_CHMOD_PROFILE) fs.chmodSync(process.env.L_TEST_CHMOD_PROFILE, 0o600);
} else if (mode === 'mv') {
  const destination = args.at(-1);
  const failed = path.join(root, 'rename-failed');
  if (destination === process.env.L_TEST_FAIL_RENAME && !fs.existsSync(failed)) {
    if (process.env.L_TEST_ROLLBACK_CHMOD) fs.chmodSync(process.env.L_TEST_ROLLBACK_CHMOD, 0o600);
    fs.writeFileSync(failed, 'failed once'); record({ mode, destination, failed: true }); process.exit(1);
  }
  const result = spawnSync('/bin/mv', args, { stdio: 'inherit' });
  process.exit(result.status ?? 1);
} else process.exit(93);
`;

async function fixture(t) {
  const root = await realpath(await mkdtemp(join(tmpdir(), "L bootstrap's fixture ")));
  t.after(() => rm(root, { recursive: true, force: true }));
  const home = join(root, 'home'), tools = join(root, 'tools'), temporary = join(root, 'tmp');
  await Promise.all([mkdir(home), mkdir(tools), mkdir(temporary)]);
  const handler = join(root, 'handler.mjs');
  await writeFile(handler, fixtureHandler);
  const executable = async (file, source) => {
    await mkdir(dirname(file), { recursive: true });
    await writeFile(file, source, { mode: 0o755 });
  };
  for (const command of ['curl', 'mv']) await executable(join(tools, command), `#!/bin/sh\nexec ${quote(process.execPath)} ${quote(handler)} ${command} "$@"\n`);
  await executable(join(tools, 'cc'), '#!/bin/sh\n[ "${L_TEST_NO_CC:-}" != 1 ] || exit 1\nwhile [ "$#" -gt 0 ]; do\n  if [ "$1" = -o ]; then shift; printf "#!/bin/sh\\nexit 0\\n" > "$1"; chmod +x "$1"; fi\n  shift\ndone\n');
  await executable(join(tools, 'uname'), '#!/bin/sh\ncase "$1" in -s) printf "%s\\n" "${L_TEST_OS:-Linux}";; -m) printf "%s\\n" "${L_TEST_ARCH:-x86_64}";; *) exit 1;; esac\n');
  await executable(join(root, 'release-installer.sh'), `#!/usr/bin/env bash\nset -eu\nexport L_TEST_DELEGATE_BASH="\u0024{BASH_VERSION:?Bash is required}"\nexec ${quote(process.execPath)} ${quote(handler)} delegate "$@"\n`);
  const env = {
    HOME: home, PATH: `${tools}:${systemPath}`, TMPDIR: temporary,
    LC_ALL: 'C', TERM: 'dumb', SHELL: '/bin/bash',
    L_TEST_ROOT: root, L_TEST_BOOTSTRAP: bootstrap,
  };
  const state = { root, home, tools, env, executable, prefix: join(home, '.local/bin'), envFile: join(home, '.local/share/l/env') };
  state.run = (command, args, extra = {}) => spawnSync(command, args, { cwd: root, env: { ...env, ...extra }, encoding: 'utf8', timeout: 15000 });
  state.install = async (args = ['--yes'], extra = {}) => spawnSync('/bin/sh', ['-s', '--', ...args], { cwd: root, env: { ...env, ...extra }, input: await readFile(bootstrap, 'utf8'), encoding: 'utf8', timeout: 15000 });
  state.events = async () => (await readFile(join(root, 'events.jsonl'), 'utf8').catch((error) => { if (error.code === 'ENOENT') return ''; throw error; })).trim().split('\n').filter(Boolean).map(JSON.parse);
  return state;
}

function passed(result) { assert.equal(result.error, undefined); assert.equal(result.status, 0, result.stderr || result.stdout); }
function failed(result) { assert.equal(result.error, undefined); assert.notEqual(result.status, 0, result.stdout); }
async function absent(file) { await assert.rejects(lstat(file), { code: 'ENOENT' }); }
async function ownedEnvironment(f, contents = 'export L_TEST_OLD_ENV=preserved\n') {
  await mkdir(dirname(f.envFile), { recursive: true });
  await writeFile(f.envFile, `${environmentHeader}\n${contents}`);
  return readFile(f.envFile, 'utf8');
}

test('bootstrap resolves latest once, verifies its installer, and delegates through Bash with an explicit version', async (t) => {
  const f = await fixture(t);
  passed(await f.install());
  const events = await f.events(), requests = events.filter((event) => event.mode === 'curl');
  assert.equal(requests.filter((event) => event.url.endsWith('/releases/latest')).length, 1);
  assert.ok(requests.some((event) => event.url.endsWith('/v0.0.1/checksums.txt')));
  assert.ok(requests.some((event) => event.url.endsWith('/v0.0.1/install-release.sh')));
  assert.ok(requests.every((event) => event.url.startsWith('https://github.com/l-programming-language/l/')));
  for (const request of requests) {
    assert.ok(request.args.includes('--proto') && request.args.includes('=https'));
    assert.ok(request.args.includes('--proto-redir'));
  }
  const delegates = events.filter((event) => event.mode === 'delegate');
  assert.equal(delegates.length, 1);
  assert.ok(delegates[0].bash);
  assert.equal(delegates[0].args[delegates[0].args.indexOf('--version') + 1], '0.0.1');
  assert.ok(delegates[0].args.includes('--with-toolchain'));
  assert.ok((await readFile(f.envFile, 'utf8')).startsWith(`${environmentHeader}\n`));
});

test('published one-line shape activates the current shell and new Bash and Zsh sessions', async (t) => {
  const f = await fixture(t);
  const command = '(set -o pipefail; curl -fsSL https://ldocs.org/install.sh | sh -s -- --yes) && . "$HOME/.local/share/l/env" && command -v l && l --version';
  const immediate = f.run('/bin/bash', ['--noprofile', '--norc', '-c', command]);
  passed(immediate);
  assert.ok(immediate.stdout.endsWith(`${f.prefix}/l\n0.0.1\n`), immediate.stdout);
  for (const args of [['--noprofile', '--rcfile', join(f.home, '.bashrc'), '-ic', 'command -v l; l --version'], ['--login', '-c', 'command -v l; l --version']]) {
    const result = f.run('/bin/bash', args); passed(result);
    assert.equal(result.stdout, `${f.prefix}/l\n0.0.1\n`);
  }
  for (const profile of ['.zshrc', '.zprofile']) {
    const result = f.run('/bin/sh', ['-c', `. ${quote(join(f.home, profile))}; command -v l; l --version`]);
    passed(result); assert.equal(result.stdout, `${f.prefix}/l\n0.0.1\n`);
  }
  const zsh = ['/bin/zsh', '/usr/bin/zsh'].find(existsSync);
  if (zsh) for (const args of [['-ic', 'command -v l; l --version'], ['-lic', 'command -v l; l --version']]) {
    const result = f.run(zsh, args, { ZDOTDIR: f.home }); passed(result);
    assert.equal(result.stdout, `${f.prefix}/l\n0.0.1\n`);
  }
});

test('explicit versions and custom prefix, Cargo, and Rustup homes survive new-shell activation', async (t) => {
  const f = await fixture(t), prefix = join(f.root, "chosen bin's directory"), cargo = join(f.root, "Cargo home's directory"), rustup = join(f.root, "Rustup home's directory");
  passed(await f.install(['--version', '0.2.3', '--prefix', prefix, '--yes'], { CARGO_HOME: cargo, RUSTUP_HOME: rustup }));
  assert.equal((await f.events()).filter((event) => event.url?.endsWith('/releases/latest')).length, 0);
  const result = f.run('/bin/bash', ['--noprofile', '--rcfile', join(f.home, '.bashrc'), '-ic', 'command -v l; command -v cargo; l --version; printf "%s\\n" "$CARGO_HOME" "$RUSTUP_HOME"']);
  passed(result);
  assert.equal(result.stdout, `${prefix}/l\n${cargo}/bin/cargo\n0.2.3\n${cargo}\n${rustup}\n`);
});

test('activation respects explicit Cargo and Rustup overrides and preserves empty PATH fields when repeated', async (t) => {
  const f = await fixture(t), installedCargo = join(f.root, 'installed Cargo'), installedRustup = join(f.root, 'installed Rustup');
  const chosenCargo = join(f.root, 'override Cargo'), chosenRustup = join(f.root, 'override Rustup');
  passed(await f.install(['--yes'], { CARGO_HOME: installedCargo, RUSTUP_HOME: installedRustup }));
  const initialPath = `:${f.prefix}::${chosenCargo}/bin:${f.tools}:${systemPath}:`;
  const result = f.run('/bin/sh', ['-c', '. "$HOME/.local/share/l/env"; . "$HOME/.local/share/l/env"; printf "%s\\n" "$CARGO_HOME" "$RUSTUP_HOME" "$PATH"'], { CARGO_HOME: chosenCargo, RUSTUP_HOME: chosenRustup, PATH: initialPath });
  passed(result);
  assert.equal(result.stdout, `${chosenCargo}\n${chosenRustup}\n${f.prefix}:${chosenCargo}/bin:::${f.tools}:${systemPath}:\n`);
});

test('repeat installation and repeated activation preserve profiles and prefer the selected compiler without PATH duplicates', async (t) => {
  const f = await fixture(t), competing = join(f.root, 'other bin');
  await f.executable(join(competing, 'l'), '#!/bin/sh\nprintf "wrong compiler\\n"\n');
  await writeFile(join(f.home, '.bashrc'), '# user configuration without a trailing newline');
  passed(await f.install(['--yes', '--version', '0.0.1']));
  const files = ['.bashrc', '.profile', '.zshrc', '.zprofile'].map((name) => join(f.home, name));
  const before = await Promise.all(files.map((file) => readFile(file, 'utf8')));
  assert.ok(before[0].startsWith('# user configuration without a trailing newline'));
  passed(await f.install(['--yes', '--version', '0.0.1']));
  assert.deepEqual(await Promise.all(files.map((file) => readFile(file, 'utf8'))), before);
  const result = f.run('/bin/sh', ['-c', '. "$HOME/.local/share/l/env"; . "$HOME/.local/share/l/env"; command -v l; l --version; printf "%s\\n" "$PATH"'], { PATH: `${competing}:${f.prefix}:${f.tools}:${systemPath}` });
  passed(result);
  const lines = result.stdout.trim().split('\n');
  assert.deepEqual(lines.slice(0, 2), [`${f.prefix}/l`, '0.0.1']);
  const entries = lines[2].split(':');
  assert.equal(entries.filter((entry) => entry === f.prefix).length, 1);
  assert.equal(entries.filter((entry) => entry === join(f.home, '.cargo/bin')).length, 1);
  assert.ok(entries.indexOf(f.prefix) < entries.indexOf(competing));
});

test('Bash login integration respects the first existing startup file without shadowing others', async (t) => {
  for (const selected of ['.bash_profile', '.bash_login', '.profile']) await t.test(selected, async (t) => {
    const f = await fixture(t), candidates = ['.bash_profile', '.bash_login', '.profile'];
    const selectedIndex = candidates.indexOf(selected);
    for (const name of candidates.slice(selectedIndex)) await writeFile(join(f.home, name), `# existing ${name}\n`);
    passed(await f.install());
    for (const name of candidates.slice(0, selectedIndex)) await absent(join(f.home, name));
    assert.match(await readFile(join(f.home, selected), 'utf8'), /L environment/);
    for (const name of candidates.slice(selectedIndex + 1)) assert.equal(await readFile(join(f.home, name), 'utf8'), `# existing ${name}\n`);
    const result = f.run('/bin/bash', ['--login', '-c', 'command -v l; l --version']); passed(result);
    assert.equal(result.stdout, `${f.prefix}/l\n0.0.1\n`);
  });
});

test('exported ZDOTDIR directs both Zsh hooks without creating competing home profiles', async (t) => {
  const f = await fixture(t), zdotdir = join(f.root, "Zsh user's configuration");
  await mkdir(zdotdir);
  passed(await f.install(['--yes'], { ZDOTDIR: zdotdir }));
  for (const name of ['.zshrc', '.zprofile']) {
    await absent(join(f.home, name));
    const contents = await readFile(join(zdotdir, name), 'utf8'); assert.match(contents, /L environment/);
    const result = f.run('/bin/sh', ['-c', `. ${quote(join(zdotdir, name))}; l --version`]);
    passed(result); assert.equal(result.stdout, '0.0.1\n');
  }
});

test('valid profile symlinks retain their identity and update the regular target without losing user bytes or mode', async (t) => {
  for (const relative of [false, true]) await t.test(relative ? 'relative target' : 'absolute target', async (t) => {
    const f = await fixture(t), target = join(f.root, "dotfiles' bashrc");
    await writeFile(target, '# user-owned dotfile\nexport KEEP_ME=yes\n');
    await chmod(target, 0o640);
    await symlink(relative ? "../dotfiles' bashrc" : target, join(f.home, '.bashrc'));
    passed(await f.install());
    assert.equal((await lstat(join(f.home, '.bashrc'))).isSymbolicLink(), true);
    const contents = await readFile(target, 'utf8');
    assert.ok(contents.startsWith('# user-owned dotfile\nexport KEEP_ME=yes\n'));
    assert.match(contents, /L environment/);
    assert.equal((await stat(target)).mode & 0o777, 0o640);
  });
});

test('dangling, cyclic, and nonregular profile destinations reject without installing', async (t) => {
  for (const kind of ['dangling', 'cycle', 'directory']) await t.test(kind, async (t) => {
    const f = await fixture(t), profile = join(f.home, '.bashrc');
    if (kind === 'dangling') await symlink(join(f.root, 'missing'), profile);
    else if (kind === 'cycle') { await symlink(join(f.home, '.loop'), profile); await symlink(profile, join(f.home, '.loop')); }
    else await mkdir(profile);
    failed(await f.install());
    assert.equal((await f.events()).filter((event) => event.mode === 'delegate').length, 0);
    await absent(f.envFile); await absent(join(f.prefix, 'l'));
  });
});

test('no-modify-path installs and writes activation without changing or creating shell startup files', async (t) => {
  const f = await fixture(t);
  await writeFile(join(f.home, '.bashrc'), '# leave this alone\n');
  passed(await f.install(['--yes', '--no-modify-path']));
  assert.equal(await readFile(join(f.home, '.bashrc'), 'utf8'), '# leave this alone\n');
  for (const name of ['.profile', '.zshrc', '.zprofile']) await absent(join(f.home, name));
  const result = f.run('/bin/sh', ['-c', '. "$HOME/.local/share/l/env"; l --version']);
  passed(result); assert.equal(result.stdout, '0.0.1\n');
});

test('help has no installation effects and noninteractive provisioning requires explicit yes', async (t) => {
  const f = await fixture(t);
  const help = await f.install(['--help']); passed(help); assert.match(help.stdout, /--no-modify-path/);
  assert.deepEqual(await f.events(), []);
  failed(await f.install([]));
  assert.equal((await f.events()).filter((event) => event.mode === 'delegate').length, 0);
  await absent(f.envFile); await absent(join(f.prefix, 'l'));
});

test('a terminal confirmation prompts once and either installs or leaves the user files unchanged', async (t) => {
  const script = ['/usr/bin/script', '/bin/script'].find(existsSync);
  if (!script) { t.skip('The script terminal utility is unavailable on this host.'); return; }
  for (const reply of ['y', 'n']) await t.test(reply === 'y' ? 'accept' : 'decline', async (t) => {
    const f = await fixture(t), profile = join(f.home, '.bashrc');
    await writeFile(profile, '# original user configuration\n');
    const command = ['/bin/sh', bootstrap, '--version', '0.0.1'];
    const args = process.platform === 'darwin'
      ? ['-q', '/dev/null', ...command]
      : ['-q', '-e', '-c', command.map(quote).join(' '), '/dev/null'];
    // cat gives BSD script a real pipe instead of Node's socketpair. Answer only
    // after the prompt and retain stdin until completion, avoiding an early EOF.
    const result = await new Promise((resolve) => {
      const child = spawn('/bin/sh', ['-c', `cat | ${[script, ...args].map(quote).join(' ')}`], { cwd: f.root, env: f.env, detached: true });
      let stdout = '', stderr = '', answered = false, error;
      const timer = setTimeout(() => {
        error = new Error('Terminal fixture exceeded its 10-second deadline.');
        process.kill(-child.pid, 'SIGKILL');
      }, 10000);
      child.stdout.setEncoding('utf8').on('data', (chunk) => {
        stdout += chunk;
        if (!answered && stdout.includes('Continue? [Y/n]')) { answered = true; child.stdin.write(`${reply}\n`); }
        if (/L 0\.0\.1 is ready\.|Installation cancelled|L setup failed:/.test(stdout)) child.stdin.end();
      });
      child.stderr.setEncoding('utf8').on('data', (chunk) => { stderr += chunk; });
      child.on('error', (caught) => { error = caught; });
      child.on('close', (status) => { clearTimeout(timer); resolve({ status, stdout, stderr, error }); });
    });
    assert.equal((result.stdout.match(/Continue\? \[Y\/n\]/g) || []).length, 1, result.stdout + result.stderr);
    if (reply === 'y') {
      passed(result);
      assert.equal((await f.events()).filter((event) => event.mode === 'delegate').length, 1);
      assert.ok((await readFile(profile, 'utf8')).startsWith('# original user configuration\n'));
      const activated = f.run('/bin/sh', ['-c', '. "$HOME/.local/share/l/env"; l --version']);
      passed(activated); assert.equal(activated.stdout, '0.0.1\n');
    } else {
      failed(result);
      assert.match(result.stdout, /Installation cancelled/);
      assert.equal((await f.events()).filter((event) => event.mode === 'delegate').length, 0);
      assert.equal(await readFile(profile, 'utf8'), '# original user configuration\n');
      await absent(f.envFile); await absent(join(f.prefix, 'l'));
      for (const name of ['.profile', '.zshrc', '.zprofile']) await absent(join(f.home, name));
    }
  });
});

test('invalid flags, versions, and PATH-incompatible prefixes reject before delegation', async (t) => {
  for (const args of [['--unknown'], ['--version'], ['--version', '../bad'], ['--version', '0.0.1\n'], ['--version', '0.0.1\n0.0.2'], ['--version', '0.0.1\r'], ['--prefix', '/tmp/colon:path'], ['--prefix', '/tmp/new\nline'], ['--prefix', '/tmp/new\rline']]) await t.test(JSON.stringify(args), async (t) => {
    const f = await fixture(t); failed(await f.install(['--yes', ...args]));
    assert.deepEqual(await f.events(), []);
    await absent(f.envFile);
  });
});

test('unsupported hosts and unavailable native C tools fail before downloading the release installer', async (t) => {
  for (const extra of [{ L_TEST_OS: 'FreeBSD' }, { L_TEST_OS: 'Darwin', L_TEST_ARCH: 'x86_64' }, { L_TEST_NO_CC: '1' }]) await t.test(JSON.stringify(extra), async (t) => {
    const f = await fixture(t); failed(await f.install(['--yes'], extra));
    assert.equal((await f.events()).filter((event) => event.mode === 'delegate' || event.url?.endsWith('/install-release.sh')).length, 0);
    await absent(f.envFile);
  });
});

test('missing, duplicated, malformed, or mismatched installer checksums prevent execution', async (t) => {
  for (const checksum of ['missing', 'duplicate', 'malformed', 'mismatch']) await t.test(checksum, async (t) => {
    const f = await fixture(t); failed(await f.install(['--yes'], { L_TEST_CHECKSUM: checksum }));
    assert.equal((await f.events()).filter((event) => event.mode === 'delegate').length, 0);
    await absent(f.envFile); await absent(join(f.prefix, 'l'));
  });
});

test('latest resolution cannot redirect installation to another repository or a nonrelease URL', async (t) => {
  for (const url of ['https://github.com/other/project/releases/tag/v0.0.1', 'http://github.com/l-programming-language/l/releases/tag/v0.0.1', 'https://github.com/l-programming-language/l/releases/tag/main', 'https://github.com/l-programming-language/l/releases/tag/vbad/v0.0.1', 'https://github.com/l-programming-language/l/releases/tag/v0.0.1\nv0.0.2']) await t.test(url, async (t) => {
    const f = await fixture(t);
    failed(await f.install(['--yes'], { L_TEST_LATEST_URL: url }));
    assert.equal((await f.events()).filter((event) => event.mode === 'delegate').length, 0);
    assert.equal((await f.events()).filter((event) => event.mode === 'curl').length, 1);
    await absent(f.envFile); await absent(join(f.prefix, 'l'));
  });
});

test('release download failures preserve existing environment, profiles, and compiler', async (t) => {
  for (const extra of [{ L_TEST_FAIL_DOWNLOAD: '/checksums.txt' }, { L_TEST_FAIL_DOWNLOAD: '/install-release.sh' }, { L_TEST_PARTIAL_DOWNLOAD: '/install-release.sh' }]) await t.test(JSON.stringify(extra), async (t) => {
    const f = await fixture(t), previousEnv = await ownedEnvironment(f);
    await writeFile(join(f.home, '.bashrc'), '# original profile\n');
    await f.executable(join(f.prefix, 'l'), '#!/bin/sh\nprintf "old\\n"\n');
    failed(await f.install(['--yes'], extra));
    assert.equal(await readFile(f.envFile, 'utf8'), previousEnv);
    assert.equal(await readFile(join(f.home, '.bashrc'), 'utf8'), '# original profile\n');
    assert.equal(f.run(join(f.prefix, 'l'), ['--version']).stdout, 'old\n');
    assert.equal((await f.events()).filter((event) => event.mode === 'delegate').length, 0);
  });
});

test('a failed bootstrap HTTP download cannot activate a stale environment through the documented pipeline', async (t) => {
  for (const extra of [{ L_TEST_FAIL_DOWNLOAD: '/install.sh' }, { L_TEST_PARTIAL_DOWNLOAD: '/install.sh' }]) await t.test(JSON.stringify(extra), async (t) => {
    const f = await fixture(t);
    await ownedEnvironment(f, 'printf "STALE ENV ACTIVATED\\n"\n');
    const result = f.run('/bin/bash', ['--noprofile', '--norc', '-c', '(set -o pipefail; curl -fsSL https://ldocs.org/install.sh | sh -s -- --yes) && . "$HOME/.local/share/l/env"'], extra);
    failed(result); assert.ok(!result.stdout.includes('STALE ENV ACTIVATED'));
    assert.equal((await f.events()).filter((event) => event.mode === 'delegate').length, 0);
  });
});

test('failed delegated installation never publishes shell integration or replaces an existing compiler', async (t) => {
  const f = await fixture(t), previousEnv = await ownedEnvironment(f);
  await writeFile(join(f.home, '.bashrc'), '# original profile\n');
  await f.executable(join(f.prefix, 'l'), '#!/bin/sh\nprintf "old\\n"\n');
  failed(await f.install(['--yes'], { L_TEST_DELEGATE_FAIL: '1' }));
  assert.equal(await readFile(f.envFile, 'utf8'), previousEnv);
  assert.equal(await readFile(join(f.home, '.bashrc'), 'utf8'), '# original profile\n');
  assert.equal(f.run(join(f.prefix, 'l'), ['--version']).stdout, 'old\n');
});

test('unowned environment files and malformed managed profile blocks are not overwritten', async (t) => {
  for (const kind of ['unowned-env', 'unterminated', 'duplicate']) await t.test(kind, async (t) => {
    const f = await fixture(t);
    let file, original;
    if (kind === 'unowned-env') { file = f.envFile; original = '# unrelated user environment\n'; }
    else {
      file = join(f.home, '.bashrc');
      const block = '# >>> L environment >>>\n# original\n# <<< L environment <<<\n';
      original = kind === 'unterminated' ? '# >>> L environment >>>\n# original\n' : block + block;
    }
    await mkdir(dirname(file), { recursive: true }); await writeFile(file, original);
    failed(await f.install());
    assert.equal(await readFile(file, 'utf8'), original);
    assert.equal((await f.events()).filter((event) => event.mode === 'delegate').length, 0);
  });
});

test('concurrent profile edits survive configuration conflict, with old environment retained and partial install reported', async (t) => {
  const f = await fixture(t), previousEnv = await ownedEnvironment(f), profile = join(f.home, '.bashrc');
  await writeFile(profile, '# original profile\n');
  const result = await f.install(['--yes'], { L_TEST_MUTATE_PROFILE: profile });
  failed(result); assert.match(result.stderr + result.stdout, /installed|partial|shell setup/i);
  assert.equal(await readFile(profile, 'utf8'), '# original profile\n# concurrent user edit\n');
  assert.equal(await readFile(f.envFile, 'utf8'), previousEnv);
  passed(f.run(join(f.prefix, 'l'), ['--version']));
});

test('a concurrent profile permission change is a conflict and preserves the user mode', async (t) => {
  const f = await fixture(t), previousEnv = await ownedEnvironment(f), profile = join(f.home, '.bashrc');
  await writeFile(profile, '# original profile\n'); await chmod(profile, 0o644);
  const result = await f.install(['--yes'], { L_TEST_CHMOD_PROFILE: profile });
  failed(result); assert.match(result.stderr + result.stdout, /installed|partial|shell setup/i);
  assert.equal(await readFile(profile, 'utf8'), '# original profile\n');
  assert.equal((await stat(profile)).mode & 0o777, 0o600);
  assert.equal(await readFile(f.envFile, 'utf8'), previousEnv);
  passed(f.run(join(f.prefix, 'l'), ['--version']));
});

test('failed configuration publication restores already-updated files without rolling back installed Rust or the compiler', async (t) => {
  const f = await fixture(t), previousEnv = await ownedEnvironment(f);
  const profiles = ['.bashrc', '.profile', '.zshrc', '.zprofile'];
  for (const name of profiles) await writeFile(join(f.home, name), `# original ${name}\n`);
  const result = await f.install(['--yes'], { L_TEST_FAIL_RENAME: join(f.home, '.zprofile') });
  failed(result); assert.match(result.stderr + result.stdout, /installed|partial|shell setup/i);
  assert.equal(await readFile(f.envFile, 'utf8'), previousEnv);
  for (const name of profiles) assert.equal(await readFile(join(f.home, name), 'utf8'), `# original ${name}\n`);
  assert.ok(existsSync(join(f.root, 'rename-failed')));
  passed(f.run(join(f.prefix, 'l'), ['--version']));
});

test('configuration rollback removes files first created by the failed setup', async (t) => {
  const f = await fixture(t);
  await writeFile(join(f.home, '.bashrc'), '# original profile\n');
  const result = await f.install(['--yes'], { L_TEST_FAIL_RENAME: join(f.home, '.zprofile') });
  failed(result);
  assert.ok(existsSync(join(f.root, 'rename-failed')));
  assert.equal(await readFile(join(f.home, '.bashrc'), 'utf8'), '# original profile\n');
  for (const file of [f.envFile, ...['.profile', '.zshrc', '.zprofile'].map((name) => join(f.home, name))]) await absent(file);
  passed(f.run(join(f.prefix, 'l'), ['--version']));
});

test('rollback preserves a concurrent permission edit to an already-published profile', async (t) => {
  const f = await fixture(t), previousEnv = await ownedEnvironment(f), profile = join(f.home, '.bashrc');
  for (const name of ['.bashrc', '.profile', '.zshrc', '.zprofile']) {
    await writeFile(join(f.home, name), `# original ${name}\n`); await chmod(join(f.home, name), 0o644);
  }
  const result = await f.install(['--yes'], { L_TEST_FAIL_RENAME: join(f.home, '.zprofile'), L_TEST_ROLLBACK_CHMOD: profile });
  failed(result); assert.match(result.stderr + result.stdout, /Preserved a concurrent edit/);
  const contents = await readFile(profile, 'utf8');
  assert.ok(contents.startsWith('# original .bashrc\n'));
  assert.match(contents, /# >>> L environment >>>/);
  assert.equal((await stat(profile)).mode & 0o777, 0o600);
  assert.equal(await readFile(f.envFile, 'utf8'), previousEnv);
  for (const name of ['.profile', '.zshrc', '.zprofile']) {
    assert.equal(await readFile(join(f.home, name), 'utf8'), `# original ${name}\n`);
    assert.equal((await stat(join(f.home, name))).mode & 0o777, 0o644);
  }
});
