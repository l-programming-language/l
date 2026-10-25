import { spawn } from 'node:child_process';
import fs from 'node:fs/promises';
import path from 'node:path';
import { website, stable, atomicWrite } from './lib.mjs';

const steps = ['ci:verify-disabled', 'references:check', 'check', 'test', 'build', 'test:e2e'];
const results = [];
for (const step of steps) {
  console.log(`\nWebsite verification: npm run ${step}`);
  const started = Date.now();
  const code = await new Promise((resolve, reject) => {
    const child = spawn('npm', ['run', step], { cwd: website, stdio: 'inherit', env: process.env });
    child.on('error', reject); child.on('exit', (code) => resolve(code ?? 1));
  });
  results.push({ command: `npm run ${step}`, exitCode: code, durationMs: Date.now() - started });
  await atomicWrite(path.join(website, 'reports/website-verification.json'), stable({ scope: 'website-only; no L execution', results }));
  if (code) { process.exitCode = code; break; }
}
