import fs from 'node:fs/promises';
import { existsSync } from 'node:fs';
import path from 'node:path';
import { execFileSync } from 'node:child_process';
import { exists, walk, safeFile, sha256, stable } from './lib.mjs';

export function git(root, ...args) {
  // An unpacked source tree must not inherit identity from a parent repository.
  if (!existsSync(path.join(root, '.git'))) return null;
  try { return execFileSync('git', ['-C', root, ...args], { encoding: 'utf8', stdio: ['ignore', 'pipe', 'ignore'] }).trim(); }
  catch { return null; }
}

export function repositoryURL(root) {
  const remote = git(root, 'remote', 'get-url', 'origin');
  if (!remote) return null;
  const match = remote.match(/^(?:https:\/\/github\.com\/|git@github\.com:)([A-Za-z0-9_.-]+\/[A-Za-z0-9_.-]+?)(?:\.git)?$/);
  return match ? `https://github.com/${match[1]}` : null;
}

const sourceLinks = new WeakMap();
export function sourceURL(root, relative, identity) {
  if (!identity.repository || !identity.docsCommit) return null;
  let links = sourceLinks.get(identity);
  if (!links) { links = new Map(); sourceLinks.set(identity, links); }
  const key = `${root}\0${relative}`;
  if (!links.has(key)) links.set(key, git(root, 'cat-file', '-e', `${identity.docsCommit}:${relative}`) === null ? null : `${identity.repository}/blob/${identity.docsCommit}/${relative.split('/').map(encodeURIComponent).join('/')}`);
  return links.get(key);
}

export const languageRoots = ['cmd', 'lyger', 'logiceme', 'builtins', 'stdlib', 'lexicon', 'lerrors', 'utilities', 'resources', 'internal', 'tools', 'acceptance', 'testdata', 'demo', 'scripts', 'harness', '.ci'];
export const languageFiles = ['go.mod', 'go.sum', '.go-version', '.ci-tools.json', '.golangci.yml', 'install.sh', 'VERSION'];
export async function inputInventory(root) {
  const files = [];
  for (const entry of await fs.readdir(root, { withFileTypes: true })) if (entry.isFile() && entry.name.endsWith('.go')) files.push(entry.name);
  for (const dir of languageRoots) files.push(...await walk(root, dir, (rel, item) => !['.DS_Store', 'target', 'node_modules', '.cache', '.git'].includes(item.name) && !rel.endsWith('.tmp')));
  for (const file of languageFiles) if (await exists(path.join(root, file))) files.push(file);
  return [...new Set(files)].sort();
}

export async function fingerprints(root, files) {
  const entries = [];
  for (const name of [...new Set(files)].sort()) entries.push([name, sha256(await fs.readFile(await safeFile(root, name)))]);
  return { fingerprint: sha256(stable(entries)), entries };
}

export async function identity(root, siteRoot, { production = false } = {}) {
  const head = git(root, 'rev-parse', '--verify', 'HEAD');
  const dirty = head ? git(root, 'status', '--porcelain=v1', '--untracked-files=normal') !== '' : null;
  const versionFile = path.join(root, 'VERSION');
  const languageVersion = await exists(versionFile) ? (await fs.readFile(versionFile, 'utf8')).trim() : null;
  if (languageVersion !== null && !/^\d+\.\d+\.\d+(?:-[0-9A-Za-z.-]+)?$/.test(languageVersion)) throw new Error('VERSION must contain a valid release version.');
  const source = await fingerprints(root, await inputInventory(root));
  const docs = await walk(root, 'docs', (rel, item) => item.name !== 'assets' && !item.name.startsWith('.'));
  const siteRelative = path.relative(root, siteRoot).split(path.sep).join('/');
  const presentation = await walk(root, siteRelative, (rel, item) => !['node_modules', 'dist', '.astro', '.cache', '.tools', '.generated', 'reports', 'test-results', 'playwright-report'].includes(item.name) && !rel.startsWith(`${siteRelative}/src/content/docs/docs`) && !rel.startsWith(`${siteRelative}/src/content/docs/releases`) && ![`${siteRelative}/public/version.json`, `${siteRelative}/public/robots.txt`].includes(rel) && !rel.startsWith(`${siteRelative}/public/content-assets/`));
  const docsDigest = await fingerprints(root, [...docs, ...presentation]);
  const value = { languageVersion, languageCommit: head, docsCommit: head, docsRevision: `sha256:${docsDigest.fingerprint}`, channel: 'local', dirty, sourceFingerprint: `sha256:${source.fingerprint}`, repository: repositoryURL(root) };
  if (production) {
    const { authenticatedReleaseIdentity } = await import('./release.mjs');
    return authenticatedReleaseIdentity({ root, policyPath: path.join(siteRoot, 'ci-templates/release-policy.json'), manifestPath: process.env.DOCS_RELEASE_MANIFEST, artifacts: process.env.DOCS_RELEASE_ARTIFACTS, artifactId: process.env.DOCS_NATIVE_ARTIFACT_ID, archive: process.env.DOCS_NATIVE_ARCHIVE, mode: process.env.DOCS_RELEASE_MODE, sourceFingerprint: value.sourceFingerprint });
  }
  return value;
}
