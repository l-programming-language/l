// Release packaging is deliberately separate from compiler execution and publication.
// Only workflow callers run the canonical gate; these helpers retain its actual bytes.
import { readFile, writeFile, mkdir, readdir, lstat, realpath, copyFile, chmod, mkdtemp, rm } from 'node:fs/promises';
import { resolve, join, relative, dirname, sep } from 'node:path';
import { tmpdir } from 'node:os';
import { execFileSync } from 'node:child_process';
import { fileURLToPath } from 'node:url';
import { sha256, stableJSON, sourceFingerprintAt, verifyNativeEvidence } from '../../website/scripts/release.mjs';
import { inputInventory, fingerprints } from '../../website/scripts/provenance.mjs';

export const targets = ['darwin-amd64', 'darwin-arm64', 'linux-amd64', 'linux-arm64'];
export const archiveMembers = ['l', 'LICENSE', 'README.md', 'setup-release.sh', 'toolchain-version', 'dependencies/Cargo.toml', 'dependencies/Cargo.lock', 'dependencies/src/main.rs'];
const stages = ['lint', 'acceptance', 'native'];
const requiredCommands = { lint: ['gofmt', 'lint-config', 'golangci-ordinary', 'golangci-nativecompile', 'actionlint'], acceptance: ['doctor', 'verify'], native: ['nativecompile'] };
const assert = (condition, message) => { if (!condition) throw new Error(message); };
const json = async (path) => JSON.parse(await readFile(path, 'utf8'));
const git = (root, ...args) => execFileSync('git', args, { cwd: root, encoding: 'utf8' }).trim();

async function regular(path) {
  assert((await lstat(path)).isFile(), `Expected a regular file: ${path}`);
  return readFile(path);
}

function within(root, path) {
  const rel = relative(resolve(root), resolve(path));
  assert(rel && !rel.startsWith(`..${sep}`) && rel !== '..' && !rel.startsWith(sep), `Path escapes evidence root: ${path}`);
  return rel.split(sep).join('/');
}

async function retainedFile(root, path) {
  const name = within(root, path);
  assert(await realpath(path) === join(await realpath(root), name), `Evidence path contains a symlink: ${path}`);
  return { name, bytes: await regular(path) };
}

// Generate only a local seed package. External package records remain byte-for-byte
// copies of the compiler's lock fragment; the production dependency rows own versions.
export function dependencySeed(runtimeSource, retainedLock) {
  const start = runtimeSource.indexOf('func runtimeDependencies(');
  assert(start >= 0, 'Missing production runtime dependency contract');
  const body = runtimeSource.slice(start).split('\n\tresult :=')[0];
  const rows = [...body.matchAll(/`((?:tokio|nix) = \{[^`]+\})`/g)].map((match) => match[1]);
  assert(rows.length === 2 && new Set(rows.map((row) => row.split(' ')[0])).size === 2, 'Ambiguous runtime dependency contract');
  for (const row of rows) {
    const [, name, version] = row.match(/^(tokio|nix) = \{ version = "=([0-9.]+)"/) || [];
    assert(name && retainedLock.includes(`name = "${name}"\nversion = "${version}"\nsource = `), `Runtime dependency is absent from retained lock: ${row}`);
  }
  assert(retainedLock.startsWith('[[package]]\n') && !retainedLock.includes('name = "l-release-dependencies"'), 'Unexpected compiler lock fragment');
  return {
    manifest: '[package]\nname = "l-release-dependencies"\nversion = "0.0.0"\nedition = "2024"\npublish = false\n\n[dependencies]\n' + rows.join('\n') + '\n',
    lock: '# Release-local root; external records are retained by L.\nversion = 4\n\n[[package]]\nname = "l-release-dependencies"\nversion = "0.0.0"\ndependencies = [\n "nix",\n "tokio",\n]\n\n' + retainedLock,
  };
}

export function validateStageReport(report, stage, target, revision, since, now = Date.now()) {
  assert(stages.includes(stage) && targets.includes(target), 'Unknown validation stage or target');
  assert(report.stage === stage && report.revision === revision && report.dirty === false && report.passed === true && !report.error, `Failed or mismatched ${target}/${stage} evidence`);
  assert(`${report.goos}-${report.goarch}` === target, `Validation did not run natively on ${target}`);
  assert(/^[a-f0-9]{64}$/.test(report.source_identity), 'Missing canonical source identity');
  const begin = Date.parse(report.started_at), end = Date.parse(report.finished_at), boundary = Date.parse(since);
  assert(Number.isFinite(boundary) && Number.isFinite(begin) && begin >= boundary && end >= begin && end <= now + 60000, 'Stale or incomplete validation timestamps');
  assert(Array.isArray(report.commands) && report.commands.length > 0 && report.commands.every((command) => command.exit_code === 0 && command.timed_out === false), 'Incomplete validation commands');
  const names = new Set(report.commands.map((command) => command.name));
  assert(requiredCommands[stage].every((name) => names.has(name)), `Missing canonical ${stage} commands`);
  return report;
}

// Retain reports, command streams, native result records and text snapshots, never
// build caches or executable products. Original report bytes/absolute paths survive.
async function copyEvidenceTree(sourceRoot, source, destinationRoot, prefix, inventory, seen) {
  for (const item of await readdir(source, { withFileTypes: true })) {
    const path = join(source, item.name);
    assert(!item.isSymbolicLink(), `Symlink in validation evidence: ${path}`);
    if (item.isDirectory()) {
      if (!['target', '.git', 'node_modules'].includes(item.name)) await copyEvidenceTree(sourceRoot, path, destinationRoot, prefix, inventory, seen);
    } else if (item.isFile() && /\.(?:json|jsonl|log|stdout|stderr|txt|l)$/.test(item.name)) {
      const { name, bytes } = await retainedFile(sourceRoot, path), dest = `${prefix}/${name}`;
      if (seen.has(dest)) continue;
      seen.add(dest);
      await mkdir(dirname(join(destinationRoot, dest)), { recursive: true });
      await writeFile(join(destinationRoot, dest), bytes, { flag: 'wx' });
      inventory.push({ path: dest, sha256: sha256(bytes) });
    }
  }
}

export async function collectEvidence({ runs, output, target, revision, since }) {
  const selected = [], identities = new Set(), inventory = [], seen = new Set();
  for (const stage of stages) {
    const candidates = [];
    for (const item of await readdir(runs, { withFileTypes: true })) {
      if (!item.name.startsWith(`ci-${stage}-`)) continue;
      assert(item.isDirectory() && !item.isSymbolicLink(), 'Invalid CI evidence directory');
      const path = join(runs, item.name, 'stage.json'), report = await json(path);
      if (Date.parse(report.started_at) >= Date.parse(since)) candidates.push({ path, report });
    }
    assert(candidates.length === 1, `Expected exactly one fresh ${target}/${stage} report, found ${candidates.length}`);
    const { path, report } = candidates[0];
    validateStageReport(report, stage, target, revision, since);
    identities.add(report.source_identity);
    for (const command of report.commands) {
      for (const field of ['stdout_path', 'stderr_path']) {
        assert(typeof command[field] === 'string', 'Missing raw command output');
        within(dirname(path), command[field]);
        await retainedFile(runs, command[field]);
      }
    }
    await copyEvidenceTree(runs, dirname(path), output, `evidence/${target}`, inventory, seen);
    // The acceptance stage retains checked original harness reports whose command
    // streams live in a sibling run directory, not inside the CI wrapper directory.
    if (stage === 'acceptance') for (const command of ['doctor', 'verify']) {
      const harness = await json(join(dirname(path), `${command}.report.json`));
      assert(harness.command === command && harness.current_revision === revision && harness.git_dirty === false && harness.exit_code === 0, 'Invalid retained harness report');
      if (command === 'verify') assert(harness.acceptance_passed === true, 'Acceptance did not pass');
      const source = resolve(harness.run_directory);
      within(runs, source);
      assert(await realpath(source) === join(await realpath(runs), within(runs, source)), 'Harness evidence directory is a symlink');
      await copyEvidenceTree(runs, source, output, `evidence/${target}`, inventory, seen);
    }
    selected.push({ stage, path: `evidence/${target}/${within(runs, path)}` });
  }
  assert(identities.size === 1, 'Validation source identities disagree');
  return { stages: selected, evidence: inventory.sort((a, b) => a.path.localeCompare(b.path)), sourceIdentity: [...identities][0] };
}

export async function pack({ root, target, binary, output, runs, since, signing = 'unsigned', receipt }) {
  assert(targets.includes(target), 'Unsupported release target');
  assert(git(root, 'status', '--porcelain') === '', 'Packaging requires a clean source checkout');
  const revision = git(root, 'rev-parse', 'HEAD'), version = (await readFile(join(root, 'VERSION'), 'utf8')).trim();
  assert(/^\d+\.\d+\.\d+$/.test(version), 'Stable release requires numeric VERSION');
  const pins = await json(join(root, '.ci-tools.json'));
  const buildInfo = execFileSync('go', ['version', '-m', binary], { encoding: 'utf8' });
  const [goos, goarch] = target.split('-');
  assert(buildInfo.split('\n')[0].endsWith(`: go${pins.go}`) && buildInfo.includes(`GOOS=${goos}`) && buildInfo.includes(`GOARCH=${goarch}`) && buildInfo.includes('CGO_ENABLED=0'), 'Binary toolchain or target differs from release policy');
  assert(execFileSync(binary, ['--version'], { encoding: 'utf8', timeout: 10000 }).trim() === version, 'Packaged binary version mismatch');
  assert(['unsigned', 'developer-id'].includes(signing), 'Unsupported signing status');
  const binaryBytes = await regular(binary);
  let signingRecord;
  if (signing === 'developer-id') {
    assert(goos === 'darwin' && receipt, 'Developer ID receipt is required on macOS');
    signingRecord = await json(receipt);
    assert(signingRecord.status === 'Accepted' && signingRecord.binarySha256 === sha256(binaryBytes) && signingRecord.notarizationId && signingRecord.identity, 'Notarization receipt differs from binary');
    execFileSync('codesign', ['--verify', '--strict', '--verbose=2', binary], { stdio: 'pipe' });
  }
  await mkdir(output); // A rerun must not silently replace a different package.
  const evidence = await collectEvidence({ runs, output, target, revision, since });
  const stage = await mkdtemp(join(tmpdir(), 'l-release-package-'));
  const asset = `l_${version}_${goos}_${goarch}.tar.gz`;
  try {
    await mkdir(join(stage, 'dependencies/src'), { recursive: true });
    await writeFile(join(stage, 'l'), binaryBytes, { mode: 0o755 });
    await copyFile(join(root, 'LICENSE'), join(stage, 'LICENSE'));
    await copyFile(join(root, 'scripts/setup-release.sh'), join(stage, 'setup-release.sh'));
    await chmod(join(stage, 'setup-release.sh'), 0o755);
    await writeFile(join(stage, 'toolchain-version'), pins.rust + '\n');
    const seed = dependencySeed(await readFile(join(root, 'logiceme/bundle_support.go'), 'utf8'), await readFile(join(root, 'cmd/l/cargo_dependencies.lock'), 'utf8'));
    await writeFile(join(stage, 'dependencies/Cargo.toml'), seed.manifest);
    await writeFile(join(stage, 'dependencies/Cargo.lock'), seed.lock);
    await writeFile(join(stage, 'dependencies/src/main.rs'), 'fn main() {}\n');
    const signingNotice = signing === 'developer-id' ? 'Developer ID signed and notarized.' : goos === 'darwin'
      ? 'Not Developer ID signed or notarized. macOS may block downloaded executables or require explicit approval; see https://support.apple.com/102445.'
      : 'Unsigned.';
    await writeFile(join(stage, 'README.md'), `# L ${version}\n\nHost: ${target}. Language resources are embedded in l.\n\nRun bash setup-release.sh to check native tools and prepare pinned Rust ${pins.rust}.\nUse --with-toolchain to permit Rustup/toolchain installation noninteractively.\nCopy l to a directory on PATH, then run l --version or l build main.l.\nGo and the repository checkout are not required.\n\n${signingNotice}\n`);
    execFileSync('tar', ['-czf', join(output, asset), '-C', stage, ...archiveMembers], { env: { ...process.env, COPYFILE_DISABLE: '1' } });
  } finally { await rm(stage, { recursive: true, force: true }); }
  const archiveHash = sha256(await readFile(join(output, asset)));
  await copyFile(join(root, 'scripts/install-release.sh'), join(output, 'install-release.sh'));
  await writeFile(join(output, 'checksums.txt'), `${archiveHash}  ${asset}\n${sha256(await readFile(join(output, 'install-release.sh')))}  install-release.sh\n`);
  if (signingRecord) {
    const path = `evidence/${target}/signing.json`, bytes = await regular(receipt);
    await writeFile(join(output, path), bytes);
    evidence.evidence.push({ path, sha256: sha256(bytes) });
  }
  const record = { schemaVersion: 1, target, revision, version, artifact: { path: asset, releaseAsset: asset, sha256: archiveHash, target, compilerVersion: `go${pins.go}`, signing, ...(signingRecord ? { signingReceipt: `evidence/${target}/signing.json` } : {}) }, ...evidence };
  await writeFile(join(output, 'target.json'), stableJSON(record));
  assert(git(root, 'status', '--porcelain') === '', 'Packaging changed the source checkout');
  return record;
}

export function validateSmoke(smoke, record) {
  assert(smoke.schemaVersion === 1 && smoke.passed === true && smoke.target === record.target && smoke.version === record.version && smoke.archiveSha256 === record.artifact.sha256, 'Installer smoke evidence does not match the packaged binary');
  assert(smoke.freshCargoHome === true && smoke.freshRustupHome === true && smoke.offlineBuild === true, 'Installer smoke did not cover isolated setup and offline compilation');
  assert(Date.parse(smoke.finishedAt) >= Date.parse(smoke.startedAt), 'Installer smoke timing is incomplete');
}

export async function assemble({ root, input, output, jobId, repository = process.env.GITHUB_REPOSITORY, runId = process.env.GITHUB_RUN_ID, runAttempt = process.env.GITHUB_RUN_ATTEMPT }) {
  assert(git(root, 'status', '--porcelain') === '', 'Release assembly requires a clean checkout');
  assert(/^[\w.-]+\/[\w.-]+$/.test(repository || ''), 'Missing source repository');
  const producer = { runId: Number(runId), runAttempt: Number(runAttempt), jobId: Number(jobId) };
  assert(Object.values(producer).every((value) => Number.isSafeInteger(value) && value > 0), 'Missing producer identity');
  const revision = git(root, 'rev-parse', 'HEAD'), version = (await readFile(join(root, 'VERSION'), 'utf8')).trim(), pins = await json(join(root, '.ci-tools.json'));
  const policy = await json(join(root, 'website/ci-templates/release-policy.json'));
  assert(Array.isArray(policy.targets) && stableJSON([...policy.targets].sort()) === stableJSON(targets), 'Configured release target inventory differs from packaging contract');
  await mkdir(output);
  const manifest = { schemaVersion: 2, repository, languageVersion: version, tag: `v${version}`, languageCommit: revision, sourceFingerprint: `sha256:${(await fingerprints(root, await inputInventory(root))).fingerprint}`, releaseFingerprint: sourceFingerprintAt(root, revision), dirty: false, producer, artifacts: [], evidence: [], validation: { command: 'scripts/l-check', targets: [] } };
  const installer = await regular(join(root, 'scripts/install-release.sh'));
  const sourceIdentities = new Set();
  for (const target of targets) {
    const directory = join(input, target), record = await json(join(directory, 'target.json'));
    assert(record.schemaVersion === 1 && record.target === target && record.revision === revision && record.version === version, 'Target artifact source identity differs');
    assert(record.artifact.target === target && record.artifact.path === `l_${version}_${target.replace('-', '_')}.tar.gz`, 'Unexpected native asset name');
    assert(record.artifact.compilerVersion === `go${pins.go}` && /^[a-f0-9]{64}$/.test(record.sourceIdentity), 'Invalid compiler/evidence identity');
    sourceIdentities.add(record.sourceIdentity);
    const archive = await regular(join(directory, record.artifact.path));
    assert(sha256(archive) === record.artifact.sha256, 'Native archive changed after packaging');
    assert((await regular(join(directory, 'install-release.sh'))).equals(installer), 'Installer exercised by target smoke differs from release source');
    await writeFile(join(output, record.artifact.path), archive, { flag: 'wx' });
    manifest.artifacts.push(record.artifact);
    assert(record.stages.length === 3 && new Set(record.stages.map((stage) => stage.stage)).size === 3 && stages.every((stage) => record.stages.some((entry) => entry.stage === stage)), 'Target lacks canonical stages');
    for (const evidence of record.evidence) {
      assert(evidence.path.startsWith(`evidence/${target}/`), 'Evidence belongs to another target');
      const { bytes } = await retainedFile(directory, join(directory, evidence.path));
      assert(sha256(bytes) === evidence.sha256, 'Target evidence changed after packaging');
      await mkdir(dirname(join(output, evidence.path)), { recursive: true });
      await writeFile(join(output, evidence.path), bytes, { flag: 'wx' });
      manifest.evidence.push(evidence);
    }
    const smoke = await json(join(directory, 'smoke.json'));
    validateSmoke(smoke, record);
    for (const name of ['smoke.json', 'smoke.stdout.log', 'smoke.stderr.log']) {
      const bytes = await regular(join(directory, name)), path = `evidence/${target}/${name}`;
      await writeFile(join(output, path), bytes, { flag: 'wx' });
      manifest.evidence.push({ path, sha256: sha256(bytes) });
    }
    manifest.validation.targets.push({ target, stages: record.stages, smoke: `evidence/${target}/smoke.json` });
  }
  assert(sourceIdentities.size === 1, 'Native targets validated different source trees');
  await writeFile(join(output, 'install-release.sh'), installer, { mode: 0o755 });
  const all = (name, bytes) => ({ path: name, releaseAsset: name, sha256: sha256(bytes), target: 'all', compilerVersion: 'not-applicable' });
  manifest.artifacts.push(all('install-release.sh', installer));
  const publicManifest = { schemaVersion: 1, sourceRepository: repository, sourceCommit: revision, version, tag: `v${version}`, toolchains: { go: pins.go, rust: pins.rust }, artifacts: manifest.artifacts.map((asset) => ({ name: asset.path, sha256: asset.sha256 })) };
  const publicBytes = stableJSON(publicManifest);
  await writeFile(join(output, 'public-release.json'), publicBytes);
  manifest.artifacts.push(all('public-release.json', publicBytes));
  const checksums = manifest.artifacts.map((asset) => `${asset.sha256}  ${asset.path}\n`).join('');
  await writeFile(join(output, 'checksums.txt'), checksums);
  manifest.artifacts.push(all('checksums.txt', checksums));
  await verifyNativeEvidence(manifest, output);
  // The downstream verifier's bounded ZIP reader permits at most one GiB of
  // expanded proof. Fail clearly instead of publishing an unusable proof bundle.
  let totalBytes = Buffer.byteLength(stableJSON(manifest));
  for (const item of [...manifest.artifacts, ...manifest.evidence]) totalBytes += (await lstat(join(output, item.path))).size;
  assert(totalBytes <= 1024 ** 3, 'Native proof exceeds the retained one-GiB verification bound');
  await writeFile(join(output, 'native-release.json'), stableJSON(manifest));
  assert(git(root, 'status', '--porcelain') === '', 'Release assembly changed the source checkout');
  return manifest;
}

function options(args) {
  const result = {};
  for (let i = 0; i < args.length; i += 2) {
    assert(/^--[a-z-]+$/.test(args[i]) && args[i + 1] && !args[i + 1].startsWith('--'), 'Expected --name value arguments');
    const key = args[i].slice(2).replace(/-([a-z])/g, (_, char) => char.toUpperCase());
    assert(!Object.hasOwn(result, key), `Duplicate argument: ${args[i]}`);
    result[key] = args[i + 1];
  }
  for (const key of ['root', 'input', 'output', 'binary', 'runs', 'receipt']) if (result[key]) result[key] = resolve(result[key]);
  return result;
}

if (process.argv[1] && resolve(process.argv[1]) === fileURLToPath(import.meta.url)) {
  const [command, ...args] = process.argv.slice(2);
  const action = command === 'pack' ? pack : command === 'assemble' ? assemble : null;
  Promise.resolve().then(() => { assert(action, 'Usage: package.mjs pack|assemble --root DIR --output DIR ...'); return action(options(args)); }).then(() => console.log(`${command}: retained verified release artifacts`)).catch((error) => { console.error(`Release packaging refused: ${error.message}`); process.exitCode = 1; });
}
