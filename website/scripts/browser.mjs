import path from 'node:path';
import { spawn } from 'node:child_process';
import { website } from './lib.mjs';

const child = spawn(process.execPath, [path.join(website, 'node_modules/@playwright/test/cli.js'), ...process.argv.slice(2)], {
  cwd: website, stdio: 'inherit', env: { ...process.env, PLAYWRIGHT_BROWSERS_PATH: process.env.PLAYWRIGHT_BROWSERS_PATH ?? path.join(website, '.cache/ms-playwright') },
});
process.on('SIGINT', () => child.kill('SIGINT'));
process.on('SIGTERM', () => child.kill('SIGTERM'));
child.on('error', (error) => { console.error(error); process.exitCode = 1; });
child.on('exit', (code) => { process.exitCode = code ?? 1; });
