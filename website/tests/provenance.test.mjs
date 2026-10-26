import assert from 'node:assert/strict';
import fs from 'node:fs/promises';
import os from 'node:os';
import path from 'node:path';
import test from 'node:test';
import { fingerprints, identity, inputInventory, repositoryURL, sourceURL } from '../scripts/provenance.mjs';
import { website } from '../scripts/lib.mjs';

test('an archive root without .git cannot borrow the surrounding checkout identity', () => {
  assert.equal(repositoryURL(website), null);
});

async function fixture(t, files = {}) {
  const temporary = await fs.mkdtemp(path.join(os.tmpdir(), 'l docs source archive '));
  const root = await fs.realpath(temporary);
  t.after(() => fs.rm(root, { recursive: true, force: true }));
  for (const [relative, content] of Object.entries(files)) {
    const file = path.join(root, relative);
    await fs.mkdir(path.dirname(file), { recursive: true });
    await fs.writeFile(file, content);
  }
  return root;
}

test('fingerprints are ordered, deduplicated and independent of absolute archive location', async (t) => {
  const files = { 'source/with spaces.l': 'Show `spaces`.\n', 'source/other.l': 'Show `other`.\n' };
  const first = await fixture(t, files);
  const second = await fixture(t, files);
  const a = await fingerprints(first, ['source/with spaces.l', 'source/other.l', 'source/other.l']);
  const b = await fingerprints(second, ['source/other.l', 'source/with spaces.l']);
  assert.deepEqual(a, b);
  assert.deepEqual(a.entries.map(([name]) => name), ['source/other.l', 'source/with spaces.l']);
  await fs.writeFile(path.join(second, 'source/with spaces.l'), 'changed\n');
  assert.notEqual((await fingerprints(second, Object.keys(files))).fingerprint, a.fingerprint);
});

test('fingerprints reject missing sources and files escaping through symlinks', async (t) => {
  const root = await fixture(t);
  const outside = await fixture(t, { 'private.l': 'outside\n' });
  await assert.rejects(fingerprints(root, ['missing.l']), /ENOENT/);
  await fs.symlink(path.join(outside, 'private.l'), path.join(root, 'escape.l'));
  await assert.rejects(fingerprints(root, ['escape.l']), /Symlink escapes/);
});

test('no-Git source archives have truthful null identity and deterministic source revisions', async (t) => {
  const files = {
    'lyger/compiler.go': 'package Lyger\n', 'go.mod': 'module example.test/l\n',
    'docs/guide.md': '# Guide\n', 'website/package.json': '{"private":true}\n',
    'website/src/styles/l-theme.css': ':root { color: black; }\n',
  };
  const first = await fixture(t, files);
  const second = await fixture(t, files);
  const a = await identity(first, path.join(first, 'website'));
  const b = await identity(second, path.join(second, 'website'));
  assert.deepEqual(a, b);
  assert.equal(a.languageVersion, null);
  assert.equal(a.languageCommit, null);
  assert.equal(a.docsCommit, null);
  assert.equal(a.dirty, null);
  assert.equal(a.repository, null);
  assert.equal(a.channel, 'local');
  assert.match(a.docsRevision, /^sha256:[a-f0-9]{64}$/);
  assert.match(a.sourceFingerprint, /^sha256:[a-f0-9]{64}$/);
  assert.ok(!JSON.stringify(a).includes(first));
  assert.ok(!Object.keys(a).some((key) => /timestamp|createdAt|builtAt/i.test(key)));
  assert.equal(repositoryURL(first), null);
  assert.equal(sourceURL(first, 'docs/guide.md', a), null);
});

test('generated projections and transient artifacts do not change documentation identity', async (t) => {
  const root = await fixture(t, { 'docs/guide.md': '# Guide\n', 'website/package.json': '{}\n' });
  const site = path.join(root, 'website');
  const initial = await identity(root, site);
  const ignored = {
    'website/src/content/docs/docs/guide/index.md': 'projection',
    'website/src/content/docs/releases/index.md': 'projection',
    'website/.generated/version.json': '{}', 'website/public/version.json': '{}',
    'website/public/robots.txt': 'User-agent: *',
    'website/public/content-assets/image.svg': '<svg/>',
    'website/node_modules/package/index.js': 'cache',
    'website/dist/index.html': 'built output', 'website/reports/screen.png': 'screen',
  };
  for (const [relative, content] of Object.entries(ignored)) {
    await fs.mkdir(path.dirname(path.join(root, relative)), { recursive: true });
    await fs.writeFile(path.join(root, relative), content);
  }
  assert.deepEqual(await identity(root, site), initial);
  await fs.writeFile(path.join(root, 'docs/guide.md'), '# Guide\n\nCorrection.\n');
  const changed = await identity(root, site);
  assert.notEqual(changed.docsRevision, initial.docsRevision);
  assert.equal(changed.sourceFingerprint, initial.sourceFingerprint);
});

test('changes to language inputs affect source identity independently of a docs correction', async (t) => {
  const root = await fixture(t, { 'lyger/compiler.go': 'before', 'docs/guide.md': '# Guide', 'website/package.json': '{}' });
  const site = path.join(root, 'website');
  const initial = await identity(root, site);
  await fs.writeFile(path.join(root, 'lyger/compiler.go'), 'after');
  const changed = await identity(root, site);
  assert.notEqual(changed.sourceFingerprint, initial.sourceFingerprint);
  assert.equal(changed.docsRevision, initial.docsRevision);
});

test('the root embedded distribution implementation is a release-relevant input', async (t) => {
  const root = await fixture(t, { 'distribution.go': 'package distribution\n', 'docs/guide.md': '# Guide', 'website/package.json': '{}' });
  const site = path.join(root, 'website');
  assert.ok((await inputInventory(root)).includes('distribution.go'), 'distribution.go must participate in language identity');
  const initial = await identity(root, site);
  await fs.writeFile(path.join(root, 'distribution.go'), 'package distribution\n// changed embedded selection\n');
  assert.notEqual((await identity(root, site)).sourceFingerprint, initial.sourceFingerprint);
});

test('VERSION support does not convert an archive into a validated production release', async (t) => {
  const root = await fixture(t, { VERSION: '2.4.1\n', 'docs/guide.md': '# Guide', 'website/package.json': '{}' });
  const site = path.join(root, 'website');
  const value = await identity(root, site);
  assert.equal(value.languageVersion, '2.4.1');
  assert.equal(value.languageCommit, null);
  assert.equal(value.channel, 'local');
  await assert.rejects(identity(root, site, { production: true }), /production|release|ENOENT/i);
  await fs.writeFile(path.join(root, 'VERSION'), 'official latest\n');
  await assert.rejects(identity(root, site), /VERSION/);
});
