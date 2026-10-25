# Develop the documentation locally

The website is an independent private npm package under `website/`. Its exact Node patch is recorded in `.node-version`; dependency resolution is recorded in `package-lock.json`. Use those pins rather than changing a global language toolchain to satisfy the site.

## Start the development server

```sh
cd website
npm ci
npm run dev
```

Open the localhost URL printed by Astro. Content preparation watches canonical documentation, mapped examples and reference inputs outside `website/` as well as presentation files. Changes, additions, renames and deletions update the generated projection. Stop the supervisor with Ctrl-C so its watcher and Astro child exit together.

Do not open generated HTML with `file://`: nested navigation, browser modules and search require an HTTP server. Normal local navigation must remain on localhost rather than jumping to the future production domain.

## Check the built site

```sh
npm run check
npm run references:check
npm test
npm run build
npm run preview
```

The build writes the complete static site and Pagefind search index. Preview serves that built result. Test search using terms such as Number, Await, IO.Print and logeme; dev-server behavior does not validate the final search index.

Browser checks use the built site:

```sh
npm run test:e2e
npm run ci:verify-disabled
```

Use `npm run verify:website` for the aggregate website-only verification. The [website README](../../website/README.md) records its stages and exact local addresses. Build once for groups that share the same output. Unit and content checks must inspect real invariants rather than succeed regardless of input.

## Troubleshoot the right layer

A dependency installation may need network access; an ordinary content build should not fetch fonts, release data or remote documentation. An unsupported Node version is an environment problem, not permission to change repository Go/Rust pins. Missing browser binaries affect browser verification, not the integrity of a completed static check.

A missing route usually means a content-map or projection problem. A broken heading link requires the canonical target's actual anchor. A stale example fingerprint requires review of that source and its evidence. Do not delete authored pages to clear generated output.

If Git metadata is unavailable, local content still builds with null/unknown commit identity and a source fingerprint. That is an expected source-archive mode. If a remote is unavailable, source attribution uses repository paths; the site must not invent an owner or commit URL.

## Keep native work opt-in

Website setup and validation do not authorize local compiler bootstrap, full-module Go lint, `go test ./...`, full ordinary/native suites, `scripts/l-check` or `scripts/l-harness verify`. The [temporary validation pause](../development-harness.md#temporary-validation-pause) remains in effect.

The optional example verifier requires an explicit invocation and an existing compiler path. Its bounded example checks remain separate from the authoritative acceptance gate. No npm lifecycle hook should execute language code or activate deployment automation.

Website CI/CD is prepared but disabled. No deployment has been performed.
