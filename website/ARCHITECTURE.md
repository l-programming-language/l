# Documentation architecture

## Package and framework

This private npm package is independent of the Go module. Astro 7.3.8 and Starlight 0.42.6 build static HTML; Node 24.21.0/npm 11.19.0 are the local/CI pin. Astro Check 0.9.10 accepts TypeScript 6.0.3; TypeScript 7 is deliberately outside this dependency set. Exact direct dependencies and the transitive graph are captured in package-lock.json.

The selected Starlight loader is `docsLoader()` from `@astrojs/starlight/loaders`, paired with `docsSchema()`. No invented loader root/glob options are used. Astro's documented Unified processor (`@astrojs/markdown-remark` 7.3.2) keeps the Markdown-aware tooling consistent. Starlight owns Pagefind, theme selection, native mobile popover navigation, heading links, on-page contents, pagination and Expressive Code. Its source was checked against the selected package version.

## Canonical ownership and projection

Authored prose lives under `docs/`. The four original guides keep their paths. An explicit allowlist also projects maintained AML, builtin and Lucid contract documents. `content-map.json` provides one route/title/description/section/order/role/generated record per canonical source. The landing MDX page and small Astro components are site-owned presentation.

Preparation loads and validates all canonical inputs before writing. Remark parses links, reference definitions, heading anchors and example directives. Positional edits preserve fenced code and untouched literal content. Mapped document links become root-relative routes; other repository files link to identifiable committed sources or display their source path. A changed local file's link is explicitly labeled as its committed base. New/unpublished files do not get fabricated remote edit URLs.

Images referenced in Markdown are checked for containment and copied by content hash. `<!-- example:ID -->` includes only manifest-approved source bytes. `<!-- diagram:compiler-pipeline -->`, `runtime-lifecycle` and `process-model` insert the reviewed supplied theme variants. Ordinary fenced examples of these directives remain literal.

The pipeline writes ignored `src/content/docs/docs/` and `releases/` files with documented Starlight frontmatter, explicit slugs, canonical edit targets and attribution. It publishes atomically per file and serializes preparation. A private ownership manifest governs stale deletion; unowned files, noncanonical paths and symlinked output paths reject. Starlight's public directories do not begin with underscores.

The development supervisor prepares once, starts Astro, and watches source directories and reference inputs outside the package. Events are debounced and serialized; events arriving during a pass schedule another pass. Failed preparation keeps the last complete projection. Astro watches presentation files. Shutdown terminates the owned child cleanly.

## Evidence and future dictionaries

Structured JSON supplies grammar, vocabulary and dictionary index records. Fixed, checked source regions supply the CLI help constant, named value table and Lucid case registrations; these extract names/data only. They are not alternative semantic parsers. Reviewed manifests supply visibility, type meaning, builtin/export and bounded backend classifications with hashes and witness paths.

Role, support and evidence remain independent. Existing test paths establish that a witness exists, not that it ran for this checkout. Node freshness checks compare exact snapshots and dependencies. A newly discovered rule, dictionary, value tag or segment without reviewed coverage fails. Semantic review hashes are never automatically refreshed by mechanical generation.

Standard-library navigation comes from index language/key identities joined to authored map records and reviewed exports. Routes use `/docs/stdlib/<lexicon>/<dictionary>/`; the site has no IO-specific renderer. Unknown indexed libraries cannot silently become empty pages. Tests add a synthetic second dictionary to verify this contract.

## Identity and release boundary

`/version.json` exposes nullable language version/commit, docs commit/revision, channel, dirty status and source fingerprint. No root VERSION is created. Local builds use the Unreleased/Local label; Git-less archives retain null Git fields. The dirty flag prevents HEAD from being mistaken for the complete current content.

The source fingerprint hashes sorted portable paths and exact bytes from root Go files; compiler, backend, resource, stdlib, lexicon, error, utility, acceptance, test, demo, script and harness directories; and relevant root tool/module pins. Generated target/cache directories are excluded. The exact implementation is `inputInventory` in scripts/provenance.mjs. Additions and deletions change the digest. Docs revision hashes canonical docs and website inputs, excluding generated output, caches and reports. There are no wall-clock timestamps in generated identity.

Release tooling additionally uses a conservative Git-tree release fingerprint. Production identity must be authenticated through the dormant release pipeline; setting an environment flag alone fails. See DEPLOYMENT.md for the native release prerequisite and exact-artifact promotion contract.

## Presentation and validation

The supplied SVG mark geometry is preserved. A thin CSS layer maps brand colors to Starlight tokens using system fonts. Source diagrams are SVG images with nearby explanation; raster social cards are metadata assets. Large diagrams load lazily. Pack navigation/theme/highlighter scripts are excluded.

`check` validates canonical content plus Astro types. Node tests exercise transformations, ownership, inventory changes and provenance. Build checks inspect actual emitted HTML, links, fragments, images, 404 and search assets. Playwright tests operate on `build` plus `preview`, including keyboard search, themes, no-JavaScript reading, mobile navigation, asset/network errors, screenshots and axe checks. Neither build nor tests run L examples or language validation.

All new website automation is physically inactive under ci-templates. Existing language CI is unchanged. Initial/final workflow byte comparisons are task evidence, not permanent website policy locks.

The L grammar and keyboard-focus hook are configured in `ec.config.mjs`, as required for non-JSON Expressive Code options shared with the Code component. Code blocks and wide tables remain keyboard accessible without JavaScript. A small Search wrapper handles the initial Pagefind idle-load focus race while retaining Starlight's dialog and controls. Astro telemetry is disabled per command, and foreground `--ignore-lock` mode lets the supervisor and browser tests own server lifetime.

The public stdlib catalog is generated from `<!-- stdlib-catalog -->`, using the same index/authored-page/export join as coverage. It does not require an IO-specific template update when another reviewed dictionary is added. The JSON registry itself is included by the example manifest.
