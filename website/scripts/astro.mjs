import path from 'node:path';
import { spawn } from 'node:child_process';
import { website } from './lib.mjs';

// Keep website commands local: no telemetry or global Astro preference writes.
const child = spawn(process.execPath, [path.join(website, 'node_modules/astro/bin/astro.mjs'), ...process.argv.slice(2)], {
  cwd: website, stdio: 'inherit', env: { ...process.env, ASTRO_TELEMETRY_DISABLED: '1' },
});
let stopping = false;
process.on('SIGINT', () => { stopping = true; child.kill('SIGINT'); });
process.on('SIGTERM', () => { stopping = true; child.kill('SIGTERM'); });
child.on('error', (error) => { console.error(error); process.exitCode = 1; });
child.on('exit', (code) => { process.exitCode = code ?? (stopping ? 0 : 1); });
