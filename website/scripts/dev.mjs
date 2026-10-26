import { spawn } from 'node:child_process';
import path from 'node:path';
import chokidar from 'chokidar';
import { prepare } from './prepare.mjs';
import { repository, website } from './lib.mjs';
import { languageRoots, languageFiles } from './provenance.mjs';

await prepare();
const child = spawn(process.execPath, [path.join(website, 'scripts/astro.mjs'), 'dev', '--ignore-lock', '--host', '127.0.0.1', '--port', '4321'], { cwd: website, stdio: 'inherit' });
let running = false; let pending = false; let timer; let closing = false;
async function refresh() {
  if (closing) return;
  if (running) { pending = true; return; }
  running = true;
  try { const result = await prepare(); console.log(`[docs] Refreshed ${result.pages} pages.`); }
  catch (error) { console.error(`[docs] ${error.message}\nFix the canonical input; the last complete projection is retained.`); }
  finally { running = false; if (pending) { pending = false; await refresh(); } }
}
const watcher = chokidar.watch([...new Set(['docs', ...languageRoots, ...languageFiles, 'website/content-map.json'])].map((p) => path.join(repository, p)), { ignoreInitial: true, awaitWriteFinish: { stabilityThreshold: 200, pollInterval: 50 }, ignored: (p) => /(?:^|\/)(target|\.git|\.harness|node_modules)(?:\/|$)/.test(p) });
const rootWatcher = chokidar.watch(repository, { depth: 0, ignoreInitial: true });
function queue() { clearTimeout(timer); timer = setTimeout(refresh, 150); }
watcher.on('all', queue);
rootWatcher.on('all', (_event, file) => { if (file.endsWith('.go')) queue(); });
async function close(signal = 'SIGTERM') {
  if (closing) return; closing = true; clearTimeout(timer); await Promise.all([watcher.close(), rootWatcher.close()]); child.kill(signal);
  const force = setTimeout(() => child.kill('SIGKILL'), 4000); force.unref();
}
process.on('SIGINT', () => close('SIGINT')); process.on('SIGTERM', () => close());
child.on('error', (error) => { console.error(error); close(); process.exitCode = 1; });
child.on('exit', async (code) => { await close(); process.exitCode = code ?? 0; });
