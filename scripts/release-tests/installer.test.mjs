import test from 'node:test';
import assert from 'node:assert/strict';
import { mkdtemp, mkdir, writeFile, readFile, chmod, copyFile, rm, symlink } from 'node:fs/promises';
import { tmpdir } from 'node:os';
import { join, resolve } from 'node:path';
import { fileURLToPath } from 'node:url';
import { createHash } from 'node:crypto';
import { spawnSync } from 'node:child_process';

const root = fileURLToPath(new URL('../../', import.meta.url));
const installer = resolve(root, 'scripts/install-release.sh');
const setup = resolve(root, 'scripts/setup-release.sh');
const files = ['l', 'LICENSE', 'README.md', 'setup-release.sh', 'toolchain-version', 'dependencies/Cargo.toml', 'dependencies/Cargo.lock', 'dependencies/src/main.rs'];
const digest = (bytes) => createHash('sha256').update(bytes).digest('hex');

async function fixture(t, { version = '0.0.1', binaryVersion = version, host = 'Linux', arch = 'x86_64' } = {}) {
  const directory = await mkdtemp(join(tmpdir(), 'l release installer '));
  t.after(() => rm(directory, { recursive: true, force: true }));
  const packageDir = join(directory, 'package'), assets = join(directory, 'assets'), tools = join(directory, 'tools'), prefix = join(directory, 'installed bin');
  await Promise.all([mkdir(join(packageDir, 'dependencies/src'), { recursive: true }), mkdir(assets), mkdir(tools), mkdir(prefix)]);
  const executable = async (name, text) => { await writeFile(name, text); await chmod(name, 0o755); };
  await executable(join(packageDir, 'l'), `#!/bin/sh\n[ "$1" = --version ] || exit 2\nprintf '%s\\n' '${binaryVersion}'\n`);
  await copyFile(setup, join(packageDir, 'setup-release.sh'));
  await writeFile(join(packageDir, 'LICENSE'), 'test license\n');
  await writeFile(join(packageDir, 'README.md'), 'fixture; no compiler is executed\n');
  await writeFile(join(packageDir, 'toolchain-version'), '1.98.1\n');
  await writeFile(join(packageDir, 'dependencies/Cargo.toml'), '[package]\nname="fixture"\nversion="0.0.1"\n');
  await writeFile(join(packageDir, 'dependencies/Cargo.lock'), '# fixture retained lock\n');
  await writeFile(join(packageDir, 'dependencies/src/main.rs'), 'fn main() {}\n');
  await executable(join(tools, 'uname'), `#!/bin/sh\ncase "$1" in -s) echo '${host}';; -m) echo '${arch}';; *) exit 2;; esac\n`);
  await executable(join(tools, 'cc'), '#!/bin/sh\nexit 0\n');
  await executable(join(tools, 'rustup'), `#!/bin/sh\nprintf '%s\\n' "rustup $*" >> "$INSTALLER_TEST_LOG"\nif [ "$1" = run ] && [ "\${INSTALLER_MISSING_RUST:-}" = true ] && [ ! -e "$INSTALLER_TEST_RUST_INSTALLED" ]; then exit 1; fi\nif [ "$1" = toolchain ] && [ "$2" = install ]; then : > "$INSTALLER_TEST_RUST_INSTALLED"; fi\nexit 0\n`);
  await executable(join(tools, 'cargo'), `#!/bin/sh\nprintf '%s\\n' "cargo $* offline=\${CARGO_NET_OFFLINE:-unset}" >> "$INSTALLER_TEST_LOG"\nif [ "\${INSTALLER_BAD_CARGO:-}" = true ]; then exit 1; fi\nif [ "\${INSTALLER_CHANGE_LOCK:-}" = true ]; then\n  while [ "$#" -gt 0 ]; do if [ "$1" = --manifest-path ]; then shift; printf changed > "\$(dirname "$1")/Cargo.lock"; break; fi; shift; done\nfi\nexit 0\n`);
  const state = { directory, packageDir, assets, tools, prefix, version, host, arch, executable };
  state.env = { ...process.env, HOME: join(directory, 'home'), CARGO_HOME: join(directory, 'cargo home'), RUSTUP_HOME: join(directory, 'rustup home'), PATH: `${tools}:${process.env.PATH}`, INSTALLER_TEST_LOG: join(directory, 'tool-log'), INSTALLER_TEST_RUST_INSTALLED: join(directory, 'rust-installed') };
  state.asset = `l_${version}_${host === 'Darwin' ? 'darwin' : 'linux'}_${['arm64', 'aarch64'].includes(arch) ? 'arm64' : 'amd64'}.tar.gz`;
  state.pack = async (entries = files) => {
    const result = spawnSync('tar', ['-czf', join(assets, state.asset), '-C', packageDir, ...entries], { encoding: 'utf8' });
    assert.equal(result.status, 0, result.stderr);
    await writeFile(join(assets, 'checksums.txt'), `${digest(await readFile(join(assets, state.asset)))}  ${state.asset}\n`);
  };
  state.install = (extra = [], env = {}) => spawnSync('bash', [installer, '--from-dir', assets, '--version', version, '--prefix', prefix, ...extra], { env: { ...state.env, ...env }, encoding: 'utf8', timeout: 10000 });
  await state.pack();
  return state;
}

test('binary-only installation handles spaces, repeat installs and versioned executable', async (t) => {
  const f = await fixture(t);
  const first = f.install(['--binary-only']);
  assert.equal(first.status, 0, first.stderr);
  assert.match(first.stdout, /Installed L 0.0.1/);
  assert.equal(spawnSync(join(f.prefix, 'l'), ['--version'], { encoding: 'utf8' }).stdout, '0.0.1\n');
  assert.equal(f.install(['--binary-only']).status, 0);
  assert.equal(await readFile(join(f.packageDir, 'dependencies/Cargo.lock'), 'utf8'), '# fixture retained lock\n');
});

test('checksum mismatch and missing/duplicate checksum entries preserve the old install', async (t) => {
  const f = await fixture(t);
  await writeFile(join(f.prefix, 'l'), 'old installation');
  await writeFile(join(f.assets, 'checksums.txt'), `${'0'.repeat(64)}  ${f.asset}\n`);
  assert.match(f.install(['--binary-only']).stderr, /checksum does not match/);
  await f.pack();
  const checksums = await readFile(join(f.assets, 'checksums.txt'), 'utf8');
  await writeFile(join(f.assets, 'checksums.txt'), checksums + checksums);
  assert.match(f.install(['--binary-only']).stderr, /duplicates/);
  await writeFile(join(f.assets, 'checksums.txt'), `${'0'.repeat(64)}  other.tar.gz\n`);
  assert.match(f.install(['--binary-only']).stderr, /missing/);
  assert.equal(await readFile(join(f.prefix, 'l'), 'utf8'), 'old installation');
});

test('regular-file inventory rejects unexpected entries, duplicates and symbolic links', async (t) => {
  const f = await fixture(t);
  await writeFile(join(f.packageDir, 'extra'), 'not allowed');
  await f.pack([...files, 'extra']);
  assert.match(f.install(['--binary-only']).stderr, /unexpected paths/);
  await f.pack([...files, 'l']);
  assert.match(f.install(['--binary-only']).stderr, /duplicate/);
  await rm(join(f.packageDir, 'l'));
  await symlink('README.md', join(f.packageDir, 'l'));
  await f.pack();
  assert.match(f.install(['--binary-only']).stderr, /non-regular files/);
});

test('wrong binary version and unsupported platform fail without replacement', async (t) => {
  const f = await fixture(t, { binaryVersion: '9.9.9' });
  await writeFile(join(f.prefix, 'l'), 'old installation');
  assert.match(f.install(['--binary-only']).stderr, /Binary version does not match/);
  await f.executable(join(f.tools, 'uname'), '#!/bin/sh\necho FreeBSD\n');
  assert.match(f.install(['--binary-only']).stderr, /Supported systems/);
  assert.equal(await readFile(join(f.prefix, 'l'), 'utf8'), 'old installation');
});

test('toolchain setup fetches locked then verifies frozen, without changing source lock', async (t) => {
  const f = await fixture(t, { host: 'Darwin', arch: 'arm64' });
  const result = f.install(['--with-toolchain'], { INSTALLER_MISSING_RUST: 'true' });
  assert.equal(result.status, 0, result.stderr);
  const log = await readFile(f.env.INSTALLER_TEST_LOG, 'utf8');
  assert.match(log, /rustup toolchain install 1\.98\.1 --profile minimal --no-self-update/);
  assert.match(log, /cargo \+1\.98\.1 fetch --locked --target host-tuple .*offline=false/);
  assert.match(log, /cargo \+1\.98\.1 fetch --frozen --target host-tuple .*offline=true/);
  assert.equal(await readFile(join(f.packageDir, 'dependencies/Cargo.lock'), 'utf8'), '# fixture retained lock\n');
});

test('Intel macOS is rejected before downloading or changing an existing installation', async (t) => {
  const f = await fixture(t, { host: 'Darwin', arch: 'x86_64' });
  await writeFile(join(f.prefix, 'l'), 'old installation');
  await f.executable(join(f.tools, 'curl'), '#!/bin/sh\nprintf download >> "$INSTALLER_TEST_LOG"\nexit 1\n');
  const result = f.install(['--binary-only']);
  assert.notEqual(result.status, 0);
  assert.match(result.stderr, /Intel Macs are not supported/);
  const download = spawnSync('bash', [installer, '--version', f.version, '--prefix', f.prefix, '--binary-only'], { env: f.env, encoding: 'utf8', timeout: 10000 });
  assert.notEqual(download.status, 0);
  assert.match(download.stderr, /Intel Macs are not supported/);
  assert.equal(await readFile(join(f.prefix, 'l'), 'utf8'), 'old installation');
  await assert.rejects(readFile(f.env.INSTALLER_TEST_LOG), { code: 'ENOENT' });
});

test('noninteractive missing Rust and failed prerequisite setup leave existing binary intact', async (t) => {
  const f = await fixture(t);
  await writeFile(join(f.prefix, 'l'), 'old installation');
  assert.match(f.install([], { INSTALLER_MISSING_RUST: 'true' }).stderr, /--with-toolchain/);
  assert.match(f.install(['--with-toolchain'], { INSTALLER_BAD_CARGO: 'true' }).stderr, /Cargo must be the rustup proxy/);
  assert.match(f.install(['--with-toolchain'], { INSTALLER_CHANGE_LOCK: 'true' }).stderr, /changed the retained lock/);
  assert.equal(await readFile(join(f.prefix, 'l'), 'utf8'), 'old installation');
});

test('missing rustup uses the opt-in bootstrap without modifying profiles or selecting a default toolchain', async (t) => {
  const f = await fixture(t), isolated = join(f.directory, 'isolated tools');
  await mkdir(isolated);
  // Deliberately omit every host rustup/cargo path, including hosted runner tools.
  for (const name of ['cat', 'dirname', 'mkdir', 'mktemp', 'cp', 'chmod', 'cmp', 'rm', 'sh']) {
    const command = spawnSync('/bin/sh', ['-c', `command -v ${name}`], { encoding: 'utf8' });
    assert.equal(command.status, 0, `Missing fixture tool ${name}`);
    await symlink(command.stdout.trim(), join(isolated, name));
  }
  for (const name of ['cc', 'uname']) await symlink(join(f.tools, name), join(isolated, name));
  await f.executable(join(isolated, 'curl'), `#!/bin/bash\nset -eu\nout=\nurl=\nwhile [ "$#" -gt 0 ]; do\n  case "$1" in --output) out=$2; shift 2;; --proto|--proto-redir|--retry) shift 2;; https://*) url=$1; shift;; *) shift;; esac\ndone\n[ "$url" = https://sh.rustup.rs ] || exit 1\ncat > "$out" <<'BOOTSTRAP'\n#!/bin/sh\nset -eu\nprintf '%s\\n' "bootstrap $*" >> "$INSTALLER_TEST_LOG"\nmkdir -p "$CARGO_HOME/bin"\ncp "$INSTALLER_TEST_RUSTUP_TEMPLATE" "$CARGO_HOME/bin/rustup"\ncp "$INSTALLER_TEST_CARGO_TEMPLATE" "$CARGO_HOME/bin/cargo"\nchmod 755 "$CARGO_HOME/bin/rustup" "$CARGO_HOME/bin/cargo"\nBOOTSTRAP\n`);
  const result = spawnSync('/bin/bash', [join(f.packageDir, 'setup-release.sh'), '--with-toolchain'], {
    env: { ...f.env, PATH: isolated, INSTALLER_MISSING_RUST: 'true', INSTALLER_TEST_RUSTUP_TEMPLATE: join(f.tools, 'rustup'), INSTALLER_TEST_CARGO_TEMPLATE: join(f.tools, 'cargo') }, encoding: 'utf8', timeout: 10000,
  });
  assert.equal(result.status, 0, result.stderr);
  const log = await readFile(f.env.INSTALLER_TEST_LOG, 'utf8');
  assert.match(log, /bootstrap -y --no-modify-path --profile minimal --default-toolchain none/);
  assert.match(log, /rustup toolchain install 1\.98\.1 --profile minimal --no-self-update/);
  assert.match(log, /cargo \+1\.98\.1 fetch --locked/);
  assert.match(log, /cargo \+1\.98\.1 fetch --frozen/);
});

test('latest uses the source repository by default or an explicit override, resolving one immutable version', async (t) => {
  for (const override of [null, 'example/releases']) {
    const repository = override || 'l-programming-language/l';
    const f = await fixture(t);
    await f.executable(join(f.tools, 'curl'), `#!/bin/bash\nset -eu\nprintf '%s\\n' "$*" >> "$INSTALLER_TEST_DOWNLOAD_LOG"\nout=\nurl=\nwhile [ "$#" -gt 0 ]; do\n  case "$1" in --output) out=$2; shift 2;; --write-out|--proto|--proto-redir|--retry) shift 2;; https://*) url=$1; shift;; *) shift;; esac\ndone\ncase "$url" in\n  https://github.com/${repository}/releases/latest) printf '%s' 'https://github.com/${repository}/releases/tag/v0.0.1';;\n  https://github.com/${repository}/releases/download/v0.0.1/*) cp "$INSTALLER_TEST_ASSETS/\${url##*/}" "$out";;\n  *) exit 1;;\nesac\n`);
    const result = spawnSync('bash', [installer, ...(override ? ['--repository', override] : []), '--prefix', f.prefix, '--binary-only'], { env: { ...f.env, INSTALLER_TEST_DOWNLOAD_LOG: join(f.directory, 'downloads'), INSTALLER_TEST_ASSETS: f.assets }, encoding: 'utf8', timeout: 10000 });
    assert.equal(result.status, 0, result.stderr);
    const log = await readFile(join(f.directory, 'downloads'), 'utf8');
    assert.equal(log.match(/releases\/latest/g)?.length, 1);
    assert.equal(log.match(/releases\/download\/v0\.0\.1/g)?.length, 2);
  }
});

test('local assets require explicit version and symlink install destinations reject', async (t) => {
  const f = await fixture(t);
  const result = spawnSync('bash', [installer, '--from-dir', f.assets, '--binary-only'], { env: f.env, encoding: 'utf8' });
  assert.match(result.stderr, /explicit --version/);
  await writeFile(join(f.directory, 'unrelated'), 'unrelated');
  await symlink(join(f.directory, 'unrelated'), join(f.prefix, 'l'));
  assert.match(f.install(['--binary-only']).stderr, /symbolic link/);
  assert.equal(await readFile(join(f.directory, 'unrelated'), 'utf8'), 'unrelated');
});
