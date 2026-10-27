# L documentation website

The independent npm package renders canonical repository documentation as a static Astro/Starlight site. Its `private: true` package flag prevents npm publication. It does not execute the L compiler during normal setup, development, checks, builds or browser tests.

The production origin is **https://ldocs.org/**. Local development and preview keep navigation on localhost.

**[Documentation checks](../.github/workflows/docs-check.yml) run on pull requests. Manual release workflows are implemented, but publication remains disabled by committed policy.**

Native binaries, installer assets, validation evidence and docs recovery archives
will share releases in `l-programming-language/l` after this repository becomes
public. Publication uses GitHub's automatic `GITHUB_TOKEN`; no separate
distribution repository or GitHub App credential is required.

## Setup and local use

Use Node **24.21.0 LTS** and its npm **11.19.0**. `.node-version` records the shared local/CI pin. No global Astro, TypeScript or browser-test installation is required. Existing later compatible Node versions may work, but the recorded validation uses the pin. On this task's workstation the verified official runtime is in ignored `.tools/node-v24.21.0-darwin-arm64/bin`; it is not part of a clean clone. Select the pinned Node using your usual Node installation/version manager, or a checksum-verified local runtime.

```sh
cd website
npm ci
npm run dev
```

Open **http://localhost:4321/**. Development prepares content and watches canonical files outside this package. Source changes that invalidate reviewed evidence produce a useful error and leave the last complete projection in place. Correct the review record and regenerate mechanical references before continuing.

```sh
npm run check
npm run build
npm run preview
```

Preview serves the built static output at **http://localhost:4321/**. Stop the dev server first. Search must be tested with build plus preview; the development server does not have the production Pagefind index. Internal navigation remains on localhost even though canonical metadata records the future production origin.

## Website verification

```sh
npm test
npm run references:check
npm run ci:verify-disabled
npm run test:browsers:install
npm run test:e2e
npm run verify:website
```

`test:browsers:install` explicitly downloads Playwright's Chromium browser; it is not an npm install hook. Set `PLAYWRIGHT_BROWSERS_PATH="$PWD/.cache/ms-playwright"` for a project-local browser cache. The Playwright configuration uses that location by default. Browser tests start and stop their own preview server; close other servers on port 4321. Install browser dependencies through your normal host setup if required; no privileged installer runs automatically.

`verify:website` checks the manual release policy (currently disabled), then runs reference freshness, content/type checks, unit tests, **one** build with static output checks, and browser tests. The legacy-named `ci:verify-disabled` command permits only the reviewed manual release paths and verifies publication authorization, action pins and read-only PR checks. It never invokes Go/Rust, native samples, bootstrap, `scripts/l-check`, or the language harness. Tests exercise real assertions and fail on incomplete evidence or output. Reports and screenshots live under ignored `reports/`; the maintained [validation summary](reports/validation.md) records actual outcomes.

The documentation workflow runs this aggregate on Ubuntu for every PR opening, update, and reopening, and supports manual dispatch. It installs the locked dependencies and Chromium with its Linux dependencies, then retains reports and failure traces for seven days. No path filter excludes changes to canonical sources outside `website/`. The repository's [local validation pause](../docs/development-harness.md#temporary-validation-pause) still applies to expensive local checks.

## Content and references

Edit language documentation under `../docs/`, not generated projections under `src/content/docs/docs/`. Site presentation lives in this package. `content-map.json` explicitly maps canonical sources to routes, navigation and edit/source attribution. Maintenance records are never public routes.

```sh
npm run references:generate
npm run references:check
```

Generation mechanically refreshes grammar, vocabulary, index, help and Lucid snapshots. It does not approve changed semantics. Source hashes in the reviewed manifests must be updated only after reviewing the changed behavior and affected documentation. A changed dictionary index requires new or updated dictionary documentation and reviewed exports, not a change to an IO-specific generator.

See [authoring documentation](../docs/contributing/documentation.md), [architecture](ARCHITECTURE.md), and the [specification audit](../docs/_meta/specification-audit.md).

## Explicitly opt-in language examples

The repository's [validation pause](../docs/development-harness.md#temporary-validation-pause) governs compiler execution. The following requires separate authorization and is not part of website verification:

```sh
npm run examples:verify -- --compiler /absolute/path/to/l --id basic
```

The supplied compiler must already exist with its documented native prerequisites. The verifier uses only manifest-approved complete examples, builds into temporary output directories, bounds compilation/runtime separately, compares exact bytes/statuses and cleans owned processes/files. It does not execute Markdown fences, install compiler tools or replace the authoritative acceptance gate. No compiler execution is needed for mechanical snapshot regeneration.

## Troubleshooting and archives

- Missing dependency: run `npm ci` in this directory with the pinned runtime. Initial installation requires network access; content builds use local inputs only.
- Missing browser: run `npm run test:browsers:install`; no browser PASS is claimed until tests execute.
- Stale evidence: review the named source/semantic mapping, then regenerate snapshots. Do not simply change expected behavior to pass a check.
- Missing or renamed document: update its explicit map entry and links. The next preparation deletes only generator-owned stale output.
- No Git metadata: builds still work. Unknown commits/dirty state become null; source paths are shown without invented remote URLs.
- Production mode: intentionally fails closed until this repository is public, the native workflow is registered, and the committed policy is explicitly enabled. Pre-1.0 releases need no custom secrets; macOS signing secrets are required from `v1.0.0` onward. See [deployment](DEPLOYMENT.md).

Keep `.tools`, `.cache`, `node_modules`, `.astro`, `.generated`, projected content, test traces, and `dist` out of Git. The package has no deployment credentials, activation hooks, hosted search, external fonts, analytics, service worker, server adapter or browser L execution.

For an additional website-only source-archive smoke test, run `npm run test:archive` with the pinned Node. It stages canonical inputs in a temporary path containing spaces, copies the existing locked dependency tree, removes Git identity, checks deterministic preparation, builds the full site and cleans staging. It does not substitute for `npm ci`, which should also be tested in a clean setup.
