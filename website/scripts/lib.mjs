import fs from 'node:fs/promises';
import path from 'node:path';
import { fileURLToPath } from 'node:url';
import { createHash } from 'node:crypto';

export const website = path.resolve(fileURLToPath(new URL('..', import.meta.url)));
export const repository = path.dirname(website);
export const sha256 = (bytes) => createHash('sha256').update(bytes).digest('hex');
export const json = async (file) => JSON.parse(await fs.readFile(file, 'utf8'));
export const stable = (value) => JSON.stringify(value, null, 2) + '\n';
export const exists = async (file) => fs.access(file).then(() => true, () => false);
export const isMain = (url) => process.argv[1] && path.resolve(process.argv[1]) === fileURLToPath(url);

export function contained(root, relative) {
  if (typeof relative !== 'string' || relative.includes('\\') || path.isAbsolute(relative)) throw new Error(`Expected a portable relative path: ${relative}`);
  const result = path.resolve(root, relative);
  if (result !== root && !result.startsWith(root + path.sep)) throw new Error(`Path escapes its root: ${relative}`);
  return result;
}

export async function safeFile(root, relative) {
  const file = contained(root, relative);
  const realRoot = await fs.realpath(root);
  const real = await fs.realpath(file);
  if (real !== realRoot && !real.startsWith(realRoot + path.sep)) throw new Error(`Symlink escapes its root: ${relative}`);
  if (!(await fs.stat(real)).isFile()) throw new Error(`Expected a file: ${relative}`);
  return file;
}

export async function walk(root, relative = '', filter = () => true) {
  const result = [];
  if (!await exists(path.join(root, relative))) return result;
  for (const item of (await fs.readdir(path.join(root, relative), { withFileTypes: true })).sort((a, b) => a.name.localeCompare(b.name, 'en'))) {
    const rel = path.posix.join(relative, item.name);
    if (!filter(rel, item)) continue;
    if (item.isDirectory()) result.push(...await walk(root, rel, filter));
    else if (item.isFile() || item.isSymbolicLink()) result.push(rel);
  }
  return result;
}

export async function atomicWrite(file, data) {
  const bytes = Buffer.isBuffer(data) ? data : Buffer.from(data);
  if (await exists(file) && (await fs.readFile(file)).equals(bytes)) return false;
  await fs.mkdir(path.dirname(file), { recursive: true });
  const temp = `${file}.${process.pid}.tmp`;
  try { await fs.writeFile(temp, bytes); await fs.rename(temp, file); }
  finally { await fs.rm(temp, { force: true }); }
  return true;
}

export async function withLock(directory, action) {
  await fs.mkdir(directory, { recursive: true });
  const file = path.join(directory, 'prepare.lock');
  let handle;
  for (let attempt = 0; attempt < 100; attempt++) {
    try { handle = await fs.open(file, 'wx'); break; }
    catch (error) {
      if (error.code !== 'EEXIST') throw error;
      const owner = Number(await fs.readFile(file, 'utf8').catch(() => ''));
      if (owner > 0) {
        try { process.kill(owner, 0); }
        catch (e) { if (e.code === 'ESRCH') { await fs.rm(file, { force: true }); continue; } }
      }
      await new Promise((resolve) => setTimeout(resolve, 100));
    }
  }
  if (!handle) throw new Error('Another content preparation is still running. Retry after it finishes.');
  await handle.writeFile(String(process.pid));
  try { return await action(); }
  finally { await handle.close(); await fs.rm(file, { force: true }); }
}
