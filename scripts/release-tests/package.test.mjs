import test from 'node:test';
import assert from 'node:assert/strict';
import { mkdtemp, mkdir, writeFile, readFile, rm, symlink } from 'node:fs/promises';
import { join } from 'node:path';
import { tmpdir } from 'node:os';
import { execFileSync } from 'node:child_process';
import { dependencySeed, validateStageReport, collectEvidence, validateSmoke, archiveMembers, assemble, targets } from '../release/package.mjs';
import { sha256 } from '../../website/scripts/release.mjs';

const SHA = 'a'.repeat(40), source = 'b'.repeat(64), since = '2026-10-01T00:00:00Z';
const commands = { lint: ['gofmt', 'lint-config', 'golangci-ordinary', 'golangci-nativecompile', 'actionlint'], acceptance: ['doctor', 'verify'], native: ['nativecompile'] };
function report(stage = 'lint') {
  return { stage, revision: SHA, source_identity: source, dirty: false, passed: true, goos: 'linux', goarch: 'amd64', started_at: since, finished_at: '2026-10-01T00:01:00Z', commands: commands[stage].map((name) => ({ name, exit_code: 0, timed_out: false })) };
}
async function temporary(t) {
  const dir = await mkdtemp(join(tmpdir(), 'l release package '));
  t.after(() => rm(dir, { recursive: true, force: true }));
  return dir;
}

test('release dependency seed preserves canonical external lock bytes', async () => {
  const runtime = await readFile(new URL('../../logiceme/bundle_support.go', import.meta.url), 'utf8');
  const lock = await readFile(new URL('../../cmd/l/cargo_dependencies.lock', import.meta.url), 'utf8');
  const seed = dependencySeed(runtime, lock);
  assert.ok(seed.lock.endsWith(lock));
  assert.match(seed.manifest, /tokio = \{ version = "=1\.53\.1"/);
  assert.match(seed.manifest, /nix = \{ version = "=0\.31\.3"/);
  assert.doesNotMatch(seed.manifest, /path =|git =|test|observer|helper/);
  assert.throws(() => dependencySeed(runtime.replace('=1.53.1', '=9.0.0'), lock), /absent/);
  assert.throws(() => dependencySeed('', lock), /Missing/);
});

test('packaging rejects wrong hosts, revisions, stale or incomplete canonical evidence', () => {
  validateStageReport(report(), 'lint', 'linux-amd64', SHA, since);
  for (const change of [
    (r) => { r.goarch = 'arm64'; }, (r) => { r.revision = 'c'.repeat(40); },
    (r) => { r.started_at = '2026-09-01T00:00:00Z'; }, (r) => { r.passed = false; },
    (r) => { r.dirty = true; }, (r) => { r.commands.pop(); },
    (r) => { r.commands[0].timed_out = true; }, (r) => { r.source_identity = ''; },
  ]) {
    const value = report(); change(value);
    assert.throws(() => validateStageReport(value, 'lint', 'linux-amd64', SHA, since));
  }
});

async function evidenceFixture(t) {
  const root = await temporary(t), runs = join(root, 'runs'), output = join(root, 'output');
  await mkdir(runs); await mkdir(output);
  for (const stage of ['lint', 'acceptance', 'native']) {
    const dir = join(runs, `ci-${stage}-fresh`); await mkdir(dir);
    const value = report(stage);
    for (const command of value.commands) for (const field of ['stdout_path', 'stderr_path']) {
      command[field] = join(dir, `${command.name}.${field}.log`);
      await writeFile(command[field], `${stage} raw command bytes\n`);
    }
    await writeFile(join(dir, 'stage.json'), JSON.stringify(value));
    if (stage === 'acceptance') for (const name of ['doctor', 'verify']) {
      const harness = join(runs, `${name}-fresh`); await mkdir(harness);
      await writeFile(join(harness, 'original.log'), 'original harness stream\n');
      await writeFile(join(dir, `${name}.report.json`), JSON.stringify({ command: name, current_revision: SHA, git_dirty: false, exit_code: 0, acceptance_passed: name === 'verify' ? true : null, run_directory: harness }));
    }
  }
  return { runs, output, target: 'linux-amd64', revision: SHA, since };
}

test('evidence collection retains raw reports and sibling harness streams', async (t) => {
  const args = await evidenceFixture(t), collected = await collectEvidence(args);
  assert.equal(collected.stages.length, 3);
  assert.equal(collected.sourceIdentity, source);
  assert.ok(collected.evidence.some((entry) => entry.path === 'evidence/linux-amd64/verify-fresh/original.log'));
  const bytes = await readFile(join(args.runs, 'ci-lint-fresh/stage.json'));
  assert.deepEqual(await readFile(join(args.output, collected.stages[0].path)), bytes);
});

test('evidence collection refuses ambiguous fresh stage runs and symlinked streams', async (t) => {
  const args = await evidenceFixture(t);
  await mkdir(join(args.runs, 'ci-lint-second'));
  await writeFile(join(args.runs, 'ci-lint-second/stage.json'), JSON.stringify(report()));
  await assert.rejects(collectEvidence(args), /exactly one fresh/);
  await rm(join(args.runs, 'ci-lint-second'), { recursive: true });
  const stream = join(args.runs, 'ci-lint-fresh/gofmt.stdout_path.log');
  await rm(stream); await symlink(join(args.runs, 'ci-lint-fresh/stage.json'), stream);
  await assert.rejects(collectEvidence(args), /symlink/i);
});

test('installer smoke must cover matching bytes, native target and isolated offline build', () => {
  const record = { target: 'linux-amd64', version: '0.0.1', artifact: { sha256: source } };
  const smoke = { schemaVersion: 1, passed: true, target: record.target, version: record.version, archiveSha256: source, freshCargoHome: true, freshRustupHome: true, offlineBuild: true, startedAt: since, finishedAt: since };
  validateSmoke(smoke, record);
  for (const property of ['passed', 'freshCargoHome', 'freshRustupHome', 'offlineBuild']) assert.throws(() => validateSmoke({ ...smoke, [property]: false }, record));
  assert.throws(() => validateSmoke({ ...smoke, archiveSha256: 'c'.repeat(64) }, record));
  assert.deepEqual(archiveMembers, ['l', 'LICENSE', 'README.md', 'setup-release.sh', 'toolchain-version', 'dependencies/Cargo.toml', 'dependencies/Cargo.lock', 'dependencies/src/main.rs']);
});

test('assembly binds all native reports and installer smoke while excluding private evidence from public metadata', async (t) => {
  const temporaryRoot = await temporary(t), root = join(temporaryRoot, 'source'), input = join(temporaryRoot, 'targets'), output = join(temporaryRoot, 'native');
  await mkdir(join(root, 'website/ci-templates'), { recursive: true });
  await mkdir(join(root, 'scripts'));
  await writeFile(join(root, 'VERSION'), '0.0.1\n');
  await writeFile(join(root, '.ci-tools.json'), JSON.stringify({ go: '1.27.3', rust: '1.98.1' }));
  await writeFile(join(root, 'scripts/install-release.sh'), '# fixture installer\n');
  await writeFile(join(root, 'website/ci-templates/release-policy.json'), JSON.stringify({ targets }));
  const git = (...args) => execFileSync('git', args, { cwd: root, encoding: 'utf8', stdio: ['ignore', 'pipe', 'pipe'] }).trim();
  git('init', '-q'); git('add', '.');
  git('-c', 'user.name=Release fixture', '-c', 'user.email=fixture@example.invalid', 'commit', '-qm', 'fixture');
  const revision = git('rev-parse', 'HEAD');
  for (const target of targets) {
    const directory = join(input, target); await mkdir(join(directory, `evidence/${target}`), { recursive: true });
    const asset = `l_0.0.1_${target.replace('-', '_')}.tar.gz`, bytes = `fixture binary archive ${target}`;
    await writeFile(join(directory, asset), bytes);
    await writeFile(join(directory, 'install-release.sh'), '# fixture installer\n');
    const record = { schemaVersion: 1, target, revision, version: '0.0.1', sourceIdentity: source, artifact: { path: asset, releaseAsset: asset, target, compilerVersion: 'go1.27.3', sha256: sha256(bytes), signing: 'unsigned' }, stages: [], evidence: [] };
    for (const stage of ['lint', 'acceptance', 'native']) {
      const value = { ...report(stage), revision, goos: target.split('-')[0], goarch: target.split('-')[1] };
      const body = JSON.stringify(value), path = `evidence/${target}/${stage}.json`;
      await writeFile(join(directory, path), body);
      record.stages.push({ stage, path }); record.evidence.push({ path, sha256: sha256(body) });
    }
    await writeFile(join(directory, 'target.json'), JSON.stringify(record));
    await writeFile(join(directory, 'smoke.json'), JSON.stringify({ schemaVersion: 1, target, version: '0.0.1', archiveSha256: sha256(bytes), passed: true, freshCargoHome: true, freshRustupHome: true, offlineBuild: true, startedAt: since, finishedAt: since }));
    await writeFile(join(directory, 'smoke.stdout.log'), 'private smoke fixture log\n');
    await writeFile(join(directory, 'smoke.stderr.log'), '');
  }
  const manifest = await assemble({ root, input, output, jobId: '33', runId: '22', runAttempt: '1', repository: 'example/private-source' });
  assert.equal(manifest.schemaVersion, 2);
  assert.equal(manifest.validation.targets.length, 3);
  assert.equal(manifest.artifacts.length, 6);
  assert.equal(manifest.evidence.length, 18);
  const publicManifest = JSON.parse(await readFile(join(output, 'public-release.json')));
  assert.equal(publicManifest.sourceCommit, revision);
  assert.equal(publicManifest.artifacts.length, 4);
  assert.doesNotMatch(JSON.stringify(publicManifest), /evidence|private smoke|runId/);
  const checksumLines = (await readFile(join(output, 'checksums.txt'), 'utf8')).trim().split('\n');
  assert.equal(checksumLines.length, 5);
  assert.equal(git('status', '--porcelain'), '');
  await writeFile(join(input, targets[0], 'install-release.sh'), '# untested replacement\n');
  await assert.rejects(assemble({ root, input, output: join(temporaryRoot, 'refused'), jobId: '33', runId: '22', runAttempt: '1', repository: 'example/private-source' }), /Installer exercised/);
});
