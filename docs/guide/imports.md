# Import project source

Core loads directly from the selected resource provider. The root [dictionary registry](../../dictionaries.json) is empty: IO, Formatting, Storage and Runtime are not bundled dictionaries.

## Import a project-owned member

An explicit `.l` include resolves relative to the source that imports it. Preserve both files when moving a project. Imported declarations retain their original source spans and import occurrence; aliases retain the selected declaration's canonical identity.

Use a grammatical possessive to select a member and an explicit alias for unqualified use. Dot-qualified authored names are invalid. A file include does not automatically create a public namespace.

## Preserve source-unit boundaries

Explicit dictionary imports, when supplied by a complete custom provider, resolve only through that provider's registry and resource root. Their private includes cannot escape to unrelated host files. The compiler does not search the current directory for replacements or fall back to embedded resources when an override is incomplete.

A program's `.l` includes resolve beside the importer. Cycles, missing files and missing members reject with source attribution. See the [import specification](../spec/imports.md).

## Use a complete custom resource bundle

[Developer overrides](../development-harness.md#compiler-defaults) replace the entire provider. They do not restore removed libraries or widen the standalone executable boundary.
