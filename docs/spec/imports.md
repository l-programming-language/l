# Imports, dictionaries, and program structure

Resolution uses one selected resource provider and original source ownership. It never searches neighboring files as replacements for indexed resources.

## Implicit builtins

The compiler loads `builtins/eng/core.l` directly into the builtin unit. Program and dictionary roots inherit core independently. No prelude or foundational dictionary is loaded. The root [registry](../../dictionaries.json) has no bundled dictionaries.

## Indexed dictionaries

An indexed import resolves through the selected provider's registry. It does not make every dictionary word implicit. A complete custom provider must supply its own selected resources; missing entries cannot fall back to historical or embedded libraries. [Use](../reference/word-glossary.md#word-builtin-use) introduces explicit aliases.

## Explicit source-file imports

A `.l` include resolves beside its importing source. It retains original spans, physical paragraphs and its expansion occurrence. Expansion adds occurrences to the same source unit without inventing a public namespace.

Dictionary-owned includes stay inside the selected dictionary root/provider. Absolute paths, backslash paths and traversal escaping that root reject. Program-relative includes and provider-relative includes have separate ownership; neither uses a CWD fallback.

## Visibility and aliases

Public/Publicly publish only selected declarations into their source-unit root. An alias preserves the target's canonical identity, signature and members. Private dependencies remain private. Use grammatical possessives, not dot-qualified authored names or aliases. Imported declarations obey the same [complete-family rules](definitions.md#inflection-declarations) as other authored source. Only qualification removal under an unchanged canonical root inherits a validated family; renames and actor selections follow the explicit [alias-family rules](names-and-scope.md#aliases-and-canonical-identity).

## Cycles and errors

Active-stack import cycles reject. Missing files, missing registry entries and unresolved members retain their owning phase and source attribution. Expansion does not silently omit failed imports.

## Resource independence

Embedded assets are the default. `--assets` selects one complete replacement without changing CWD. See [compiler defaults](../development-harness.md#compiler-defaults).

## Sources and witnesses

See [source preparation](../../lyger/source.go), [builder](../../lyger/builder.go) and [semantic resolution](../../lyger/semantic_resolution.go). Import support does not widen the [standalone execution boundary](execution.md).
