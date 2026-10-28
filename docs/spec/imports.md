# Imports, dictionaries, and program structure

Builtins are implicit; optional library vocabulary is imported explicitly. Resolution uses the selected resource provider and original source ownership, not a search of neighboring files with convenient names.

## Implicit builtins

Released compilers load the embedded English `builtins/eng/core.l`. The builtin unit owns its canonical definitions and supplies an inherited lexical environment to the program and each imported dictionary. No project bootstrap or lexicon copy is needed; [resource overrides](../development-harness.md#compiler-defaults) are for compiler development.

## Indexed dictionaries

Bare dictionary names resolve through the selected `stdlib/index.json`. The current format is version 1, with language keys containing dictionary keys. Each entry names a repository/resource-relative `.l` entry and the registered Dictionary it exports.

For example:

<!-- example:imports-indexed-member -->

This imports the public member from the registered IO dictionary and binds a local alias. It does not search for a neighboring `io` directory or `io.l`, nor does it expose private rendering and writing helpers. Importing the dictionary itself retains a namespace instead of flattening all members into the program.

The index validates duplicate JSON keys, safe resource paths, format, actual entry files, and consistent dictionary identity. An unknown key is an unresolved dictionary; an unsupported provider-style name reports its bounded unsupported-provider diagnostic. There is no package download or unrestricted remote dependency resolver hidden in Use.

Each indexed dictionary owns a private lexical root inheriting builtins. Its public export must resolve to the registered Dictionary. Imported member aliases resolve in that dictionary's environment; helper definitions do not capture caller-added members. Repeated dictionary dependencies reuse canonical dictionary identity while preserving import-site attribution.

## Explicit source-file imports

A dependency ending in `.l` is an explicit source include. Program includes resolve relative to the importing source file, even when the compiler is invoked from another CWD. Quoted paths can contain spaces. Included statements retain original spans, physical paragraph identity, and the import occurrence that introduced them.

An include adds occurrences to its importing source unit; it does not automatically create a new public namespace. Imported aliases retain the caller's declaration location while resolving the included target. Interleaved source order survives expansion.

Dictionary-owned includes remain inside that dictionary's resource root and provider. Absolute paths, backslash paths, and traversal escaping that root reject. They cannot fall back to unrelated host files merely because those files exist beside the current process.

**Supported example — a project-owned vocabulary and its caller:**

The canonical `demo/dictionaries/file-import/main.l` contains:

<!-- example:file-import -->

The companion `helpers.l` contains:

<!-- example:file-import-helper -->

The first import resolves `helpers.l` beside `main.l`, loads its declarations, and aliases the selected Greeting member. The second sentence aliases Display from the now-known LocalTools namespace. `Display Greeting` expands to Show and emits `local hello|` without adding a newline. The existing CLI witness records those exact bytes; it has not been rerun here. Keep both source files together when trying the complete example.

An include can also occur inside an App-led continuation without discarding its owning paragraph:

<!-- example:imports-continuation -->

The included `continuation.l` supplies:

<!-- example:imports-continuation-body -->

At the include position, `snapshot` still names the caller's local output snapshot. The included statements show it and the imported `marker`, before the caller's following Show. The source origins and physical paragraph identities of the two files remain distinct; a pronoun in the included file cannot borrow an antecedent from the caller. This complete source pair has frontend and native-runtime witnesses. Its retained expected stdout is `before-await|imported\nfrom-import|imported\nafter-await|`, not a result measured during the documentation update.

## Cycles, duplicates, and visibility

Active source/import identity detects cycles and reports the dependency chain. Aliases and descriptor/action expansion have separate cycle checks. Missing files preserve filesystem errors; missing exports are reference/type errors rather than silent empty dictionaries.

Two imports that attempt incompatible declarations in one scope can collide. Case-insensitive bindings and inflections participate in that check. A private helper is not made public by being needed internally by a public export. A member's canonical owner and any private captures remain intact after aliasing.

**Intentionally invalid explanatory fragment — a selected import does not expose implementation helpers:**

<!-- example:imports-private-helper -->

The Print import is valid, but the final name is unresolved: Write belongs to the dictionary's private lexical root. The [existing distribution test](../../lyger/distribution_frontend_test.go) checks this rejection for Write and Discard after a complete App program. This shortened documentation fragment is inspected against that test; it is not a separately executed negative fixture. Import the documented public member instead of depending on a helper's implementation spelling.

## Source units and generated layout

The compiler distinguishes builtin, program, and dictionary source units from individual file occurrences. An explicit file include remains in its importing unit. Source origins carry identity and content hashes for diagnostics and ownership; absolute local paths are implementation provenance and should not be invented in portable site output.

Apps project executable units and their additional constructed Services project libraries. Standalone Service declarations produce no workspace units. Objects are files and Classes are recursive module/source groups; this layout does not introduce runtime lifetimes or lexical paragraph scopes.

App publication retains version-1 manifests. App-constructed library targets have no executable runtime wrapper or start/stop artifacts. Standalone Service declarations create no packages.

## Adding a standard library

The public standard-library section follows every actual language/key entry in `stdlib/index.json`. The current IO documentation is one member of that catalog, not a special-case architecture. A future indexed dictionary requires reviewed export/visibility evidence and authored reference content before it is classified as supported.

The index can mechanically establish available dictionary keys and source entry paths. It cannot establish semantic exports, overloads, effects, or private-helper visibility. Those are reviewed mappings tied to actual frontend resolution witnesses and source hashes. See the [standard-library catalog](../stdlib/index.md#catalog-maintenance) for the documentation contract.

## Sources and witnesses

Implementations: [builder expansion](../../lyger/builder.go), [source units](../../lyger/source_units.go), [index validation](../../lyger/dictionary_index.go), [lexical roots](../../lyger/lexical_scope.go), [member resolution](../../lyger/semantic_resolution.go), and [index](../../stdlib/index.json). Existing witnesses: [configured imports](../../lyger/configured_import_test.go), [case-insensitive imports](../../lyger/dictionary_case_test.go), [collisions](../../lyger/dictionary_collision_test.go), [private ownership](../../lyger/dictionary_ownership_test.go), [dependency capture](../../lyger/dictionary_dependency_capture_test.go), and [source-unit origins](../../lyger/source_unit_origins_test.go).
