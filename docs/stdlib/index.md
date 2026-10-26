# Standard library

L's standard library supplies explicitly imported vocabulary dictionaries. English releases embed the library and its `stdlib/index.json`; builtins remain implicit.

<!-- illustration:dictionaries -->

## Current catalog

<!-- stdlib-catalog -->

The catalog joins each indexed dictionary to its reference and reviewed public exports. See [the source index](../../stdlib/index.json), [IO reference](eng/io.md), and [site provenance](../releases/index.md#inspect-this-sites-provenance) for the documented revision.

## Import model

Use selects a dictionary or its member and can bind a local alias. The common current import is:

```l
Use IO's Print from `io` as "Print".
```

The bare `io` key resolves only through the selected index. A neighboring file cannot shadow it. An explicit `.l` dependency follows source-relative inclusion rules instead. Dictionary helpers retain their private source-unit lexical environment and can depend on implicit builtins without becoming public API.

## Public exports and evidence

A registered Dictionary and its owned public members define the supported import surface. The source index tells us where the dictionary lives and which namespace to resolve; it does not describe exported signatures, effects, overloads, visibility, or native correctness.

Reference pages therefore pair each documented export with reviewed definitions and frontend ownership/visibility witnesses. An existing test is identified as an existing witness. A matching content hash proves freshness of that review dependency, not that the export compiled or executed during a website build.

## Catalog maintenance

The index uses `(language, key)` as dictionary identity; released entries use `eng`. Entry paths and exported namespaces are reviewed properties, not inferred from keys. Dictionary pages follow `docs/stdlib/<language>/<key>.md` or explicit content-map routes for nested keys.

Each new dictionary requires reviewed export names, authored anchors, source/include dependencies, visibility evidence, support status, and deterministic hashes. The current mapping includes IO.Print and records private Log/Logger machinery.

An index addition, rename, removal, entry-file change, or relevant included-source change must make classification/freshness checks fail until its mapping and prose are reviewed. New private helpers also require a review of whether public behavior or visibility changed. Removing an export must update its documentation and release notes rather than leave a stale searchable API page.

Mechanical extraction may enumerate index entries and source dependencies. It must not infer exports by regex over Create/Define text: source-defined expansion, canonical aliases, dictionary roots, and publication rules determine semantic visibility. Where a future dictionary exceeds what explicit reviewed mappings can safely establish, an opt-in exporter should reuse the real frontend rather than introduce a competing L parser.

## Sources and witnesses

Authoritative inventory: [stdlib index](../../stdlib/index.json). Resolution: [index reader](../../lyger/dictionary_index.go), [source units and export checks](../../lyger/source_units.go), [source expansion](../../lyger/builder.go), and [lexical roots](../../lyger/lexical_scope.go). Witnesses: [index validation](../../lyger/dictionary_index_test.go), [configured imports](../../lyger/configured_import_test.go), [ownership](../../lyger/dictionary_ownership_test.go), and [dependency capture](../../lyger/dictionary_dependency_capture_test.go).

## Index format

Each source lexicon owns a map from import keys to an entry file and its exported dictionary namespace. The current registry is included directly below; new entries must have an authored reference and reviewed export inventory before website checks pass.

<!-- example:stdlib-index -->
