# Documentation architecture

## Package and framework

The website is independent of the Go module. Astro and Starlight build static HTML; [package-lock.json](package-lock.json) owns dependency versions and [.node-version](.node-version) pins Node. Starlight provides search, themes, mobile navigation, heading links, contents, pagination, and code blocks.

## Canonical ownership and projection

Prose lives under `docs/` and in explicitly mapped compiler references. [content-map.json](content-map.json) assigns routes, descriptions, navigation, and document roles. The landing page and Astro components own site presentation.

Preparation validates inputs before writing. Markdown-aware transformations rewrite mapped links to site routes, check heading anchors, and include approved example bytes. Other repository links identify committed sources or show their paths when Git metadata is unavailable. Images are checked for containment and copied by content hash; diagrams use supplied theme variants. Fenced source and literal directive examples are preserved.

Generated pages under `src/content/docs/docs/` and `releases/` carry explicit slugs, source attribution, and edit targets. Preparation is serialized and writes atomically per file. An ownership manifest limits stale-output deletion; unowned files and symlinked output paths reject.

The development supervisor watches canonical inputs outside the package and serializes refreshes. Failed preparation retains the last complete projection; Astro watches presentation files. Shutdown terminates the owned server.

## Evidence and future dictionaries

JSON inputs and checked source regions supply grammar, vocabulary, CLI help, and Lucid inventories. Reviewed manifests provide semantic classifications, source hashes, and witness paths. Mechanical generation refreshes snapshots, not semantic review hashes. Unclassified entries and stale dependencies fail freshness checks. [Evidence and review](../docs/contributing/evidence.md) explains the distinction between source review and executed validation.

The standard-library catalog joins index language/key identities to authored pages and reviewed public exports. Routes use `/docs/stdlib/<lexicon>/<dictionary>/`; new dictionaries require both documentation and reviewed coverage. Catalog rendering is independent of any one dictionary.

## Identity and release boundary

`/version.json` records language and documentation identities, channel, dirty status, and source fingerprint. Local previews use root `VERSION`; source archives retain null Git fields.

[scripts/provenance.mjs](scripts/provenance.mjs) hashes sorted source paths and bytes, excluding generated output and caches. Documentation has its own revision hash. Release tooling additionally records a Git-tree fingerprint. Production identity requires authenticated native-release evidence, not an environment flag alone. [DEPLOYMENT.md](DEPLOYMENT.md) owns activation and recovery procedures.

## Presentation and validation

Brand colors map to Starlight tokens using system fonts. SVG diagrams have nearby explanations. Keyboard-focus hooks keep code and tables accessible; the Search wrapper handles Pagefind's initial focus timing. Neither reading nor mobile navigation requires JavaScript.

Website checks cover canonical content, types, transformations, provenance, emitted HTML/links/assets, search, browser interaction, and accessibility. They do not execute L examples or native validation. See [local development](../docs/contributing/local-development.md) for commands.

Pull-request documentation and dependency checks use read-only permissions. Manual release and recovery workflows live in `.github/workflows/`; committed policy currently blocks publication. The release guard checks approved manual paths, authorization, action pins, and read-only PR workflows.
