# Standard library

L's standard library supplies explicitly imported vocabulary dictionaries. It is separate from implicit builtins, compiler implementation packages, and native logeme templates. Availability is determined by the selected complete distribution's `stdlib/index.json`.

<!-- illustration:dictionaries -->

## Current catalog

<!-- stdlib-catalog -->

The rendered catalog joins every language/dictionary entry in the source index to its authored reference and reviewed public exports. Repository readers can inspect [the source index](../../stdlib/index.json) and the current [IO reference](eng/io.md) directly.

This is the current inspected catalog, not a permanent limit on the standard library. More indexed dictionaries can be added without changing the site's directory model or import resolver documentation. No math, filesystem, networking, collections, or concurrency module is invented here.

## Import model

Use selects a dictionary or its member and can bind a local alias. The common current import is:

```l
Use IO's Print from `io` as "Print".
```

The bare `io` key resolves only through the selected index. A neighboring file cannot shadow it. An explicit `.l` dependency follows source-relative inclusion rules instead. Dictionary helpers retain their private source-unit lexical environment and can depend on implicit builtins without becoming public API.

English is the only bundled source lexicon; the catalog language column concerns L source, not the site's interface locale. An alternate `--language` selection requires corresponding resources rather than automatic translation of these imports.

## Public exports and evidence

A registered Dictionary and its owned public members define the supported import surface. The source index tells us where the dictionary lives and which namespace to resolve; it does not describe exported signatures, effects, overloads, visibility, or native correctness.

Reference pages therefore pair each documented export with reviewed definitions and frontend ownership/visibility witnesses. An existing test is identified as an existing witness. A matching content hash proves freshness of that review dependency, not that the export compiled or executed during a website build.

No authoritative release version is assigned to the current catalog. Availability is stated for the inspected local/unreleased source. Future release metadata can identify the exact catalog snapshot when an actual release exists.

## Catalog maintenance

The documentation pipeline treats the index as an open language-to-dictionary map. Dictionary inventory identity is the pair `(language, key)`; the entry path and exported namespace are reviewed properties, not assumed from the key. A future dictionary page follows `docs/stdlib/<language>/<key>.md` or an explicit content-map route for a nested key.

Each new indexed dictionary requires a reviewed export mapping with canonical export names, authored anchors, source entry/include dependencies, visibility evidence, support status, and deterministic hashes. Export identities include language and dictionary key, so two lexicons can document distinct APIs without collision. The current mapping includes IO.Print and explicitly records private Log/Logger machinery.

An index addition, rename, removal, entry-file change, or relevant included-source change must make classification/freshness checks fail until its mapping and prose are reviewed. New private helpers also require a review of whether public behavior or visibility changed. Removing an export must update its documentation and compatibility note rather than leave a stale searchable API page.

Mechanical extraction may enumerate index entries and source dependencies. It must not infer exports by regex over Create/Define text: source-defined expansion, canonical aliases, dictionary roots, and publication rules determine semantic visibility. Where a future dictionary exceeds what explicit reviewed mappings can safely establish, an opt-in exporter should reuse the real frontend rather than introduce a competing L parser.

## Sources and witnesses

Authoritative inventory: [stdlib index](../../stdlib/index.json). Resolution: [index reader](../../lyger/dictionary_index.go), [source units and export checks](../../lyger/source_units.go), [source expansion](../../lyger/builder.go), and [lexical roots](../../lyger/lexical_scope.go). Witnesses: [index validation](../../lyger/dictionary_index_test.go), [configured imports](../../lyger/configured_import_test.go), [ownership](../../lyger/dictionary_ownership_test.go), and [dependency capture](../../lyger/dictionary_dependency_capture_test.go).

## Index format

Each source lexicon owns a map from import keys to an entry file and its exported dictionary namespace. The current registry is included directly below; new entries must have an authored reference and reviewed export inventory before website checks pass.

<!-- example:stdlib-index -->
