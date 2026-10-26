import fs from 'node:fs/promises';
import os from 'node:os';
import path from 'node:path';
import { spawn } from 'node:child_process';
import { repository, website, json, sha256, safeFile, atomicWrite, stable, isMain } from './lib.mjs';

export function runBounded(command, args, { cwd, timeoutMs = 30000 } = {}) {
  return new Promise((resolve, reject) => {
    const child = spawn(command, args, { cwd, stdio: ['ignore', 'pipe', 'pipe'], detached: process.platform !== 'win32' });
    const stdout = []; const stderr = []; let timedOut = false;
    function stop() { try { process.kill(process.platform === 'win32' ? child.pid : -child.pid, 'SIGKILL'); } catch (e) { if (e.code !== 'ESRCH') reject(e); } }
    const timer = setTimeout(() => { timedOut = true; stop(); }, timeoutMs);
    child.stdout.on('data', (b) => stdout.push(b)); child.stderr.on('data', (b) => stderr.push(b));
    child.on('error', (error) => { clearTimeout(timer); reject(error); });
    child.on('close', (code, signal) => { clearTimeout(timer); stop(); resolve({ code, signal, timedOut, stdout: Buffer.concat(stdout), stderr: Buffer.concat(stderr) }); });
  });
}
export async function verifyExamples(args) {
  const at = args.indexOf('--compiler');
  if (at < 0 || !args[at + 1] || !path.isAbsolute(args[at + 1])) throw new Error('Opt-in language execution: provide --compiler /absolute/path/to/l. This requires separately authorized compiler/native validation under the repository pause. No toolchains are installed by this command.');
  const compiler = args[at + 1]; const idAt = args.indexOf('--id'); const selectedId = idAt >= 0 ? args[idAt + 1] : null;
  const all = (await json(path.join(repository, 'docs/_meta/examples.json'))).examples;
  const selected = all.filter((e) => e.classification !== 'fragment' && (!selectedId || e.id === selectedId));
  if (!selected.length) throw new Error('No complete examples matched.');
  const outcomes = [];
  for (const example of selected) {
    const source = await safeFile(repository, example.source);
    if (sha256(await fs.readFile(source)) !== example.sha256) throw new Error(`Source changed: ${example.id}`);
    if (example.classification === 'runnable' && (example.expectedStdoutBase64 === undefined || example.expectedExit === undefined)) throw new Error(`No reviewed exact output expectation for ${example.id}; select a fully specified example with --id.`);
    const temporary = await fs.mkdtemp(path.join(os.tmpdir(), 'l-docs-example-'));
    try {
      const build = await runBounded(compiler, ['build', '--output-dir', temporary, source], { cwd: repository, timeoutMs: 40000 });
      if (build.timedOut) throw new Error(`${example.id}: compilation timed out`);
      if (example.classification === 'invalid') {
        const tokens = Array.isArray(example.diagnosticIncludes) ? example.diagnosticIncludes : [example.diagnosticIncludes];
        if (build.code === 0 || (example.expectedExit !== undefined && build.code !== example.expectedExit) || !tokens.length || tokens.some((token) => !token || !build.stderr.toString('utf8').toLowerCase().includes(token.toLowerCase()))) throw new Error(`${example.id}: expected compile diagnostic was not observed`);
        outcomes.push({ id: example.id, stage: 'compile-rejection', status: 'passed' }); continue;
      }
      if (build.code !== 0) throw new Error(`${example.id}: build failed: ${build.stderr.toString('utf8')}`);
      if (example.classification === 'runnable') {
        const binary = path.join(temporary, path.basename(source, '.l'));
        const run = await runBounded(binary, [], { cwd: temporary, timeoutMs: 15000 });
        if (run.timedOut || run.code !== example.expectedExit || !run.stdout.equals(Buffer.from(example.expectedStdoutBase64, 'base64')) || !run.stderr.equals(Buffer.from(example.expectedStderrBase64 ?? '', 'base64'))) throw new Error(`${example.id}: runtime output/status differs from its reviewed exact expectation`);
      }
      outcomes.push({ id: example.id, stage: example.classification === 'runnable' ? 'native-runtime' : 'native-compilation', status: 'passed' });
    } finally { await fs.rm(temporary, { recursive: true, force: true }); }
  }
  await atomicWrite(path.join(website, 'reports/examples-verification.json'), stable({ compiler, outcomes }));
  return outcomes;
}
if (isMain(import.meta.url)) verifyExamples(process.argv.slice(2)).then((r) => console.log(stable(r))).catch((e) => { console.error(e.message); process.exitCode = 1; });
