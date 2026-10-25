# Import dictionaries and project source

L distinguishes implicit language definitions, explicitly imported standard dictionaries and project-owned `.l` sources. This separation makes import ownership predictable and prevents nearby files from silently replacing a standard dictionary.

## Import the public IO member

The first sentence of the [basic program](first-program.md) imports `IO's Print` from the indexed `io` dictionary and assigns the local word `Print`. The current standard-library index contains `eng/io`; its public printing member is IO.Print. The private Log/Logger implementation is not a separately documented public API.

The compiler ships resource bytes, but shipment is not an implicit import. You do not need to copy core definitions into a project; you do need the IO import when using Print. A local `io` or `io.l` does not shadow the indexed dictionary path.

## Import a project-owned member

<!-- example:file-import -->

The companion source owns the imported declarations:

<!-- example:file-import-helper -->

The first import selects `Greeting` from `helpers.l`; that file is resolved relative to the importing source. The next import aliases the loaded `Display` member. Display is an ordinary source action around Show, so the retained expected stdout is `local hello|` without a newline.

Keep these two files together. Running the CLI from a different directory does not make `helpers.l` relative to the compiler executable. The [distribution CLI tests](../../acceptance/distribution/cli/cli_test.go) include both explicit file imports and a different invocation directory.

## Preserve source-unit boundaries

Imported program files expand at their import positions. The builtin unit and dictionary units retain their own scopes instead of becoming a textual paste into the caller's paragraph. Dictionary units inherit builtins but do not inherit arbitrary caller locals.

Aliases retain canonical ownership and member identity. A qualified name can resolve a public member without making every private helper in the dictionary public. The resolver checks complete qualified bindings and then canonical member traversal; it does not fall back to unrelated strings or filesystem paths after a failed member lookup.

Active-stack import cycles are rejected. Missing files and missing members are distinct failures. The source attribution should identify the original file and statement rather than an invented generated filename.

## Use a complete custom resource bundle

`--assets ROOT` replaces the entire selected resource filesystem: lexicon, builtins, stdlib and backend assets must be complete. It is a development/resource-selection facility, not a merge of a few local overrides with embedded defaults. Relative asset paths resolve against invocation CWD.

Only the bundled English lexicon is documented as supplied. Selecting another code requires suitable assets and still encounters the implementation's partial localization boundary. It does not translate a program automatically.

For exact resolution rules see [the CLI resource contract](../cli-build.md#invocation-and-assets), [the architecture](../architecture.md#embedded-builtins-and-explicit-dictionaries), and [stdlib](../stdlib/index.md). Next, organize [Apps and Services](apps-and-services.md).
