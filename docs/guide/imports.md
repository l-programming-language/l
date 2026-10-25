# Import dictionaries and project source

Builtins are implicit, standard dictionaries are imported explicitly, and project files use `.l` includes. Neighboring files cannot silently replace a standard dictionary.

## Import the public IO member

The [basic program](first-program.md) imports IO's Print from `io` under the local name Print. The standard-library index maps `eng/io` to the IO dictionary, whose public member is IO.Print; Log/Logger remain private helpers. Builtins need no project copy, but Print requires its import. A local `io` or `io.l` does not shadow the indexed `io` import.

## Import a project-owned member

<!-- example:file-import -->

The companion file supplies the declarations:

<!-- example:file-import-helper -->

The first sentence loads the declarations in `helpers.l`, including LocalTools and its members, then gives Greeting a local alias. Display is now known, so the second sentence only creates its short alias; it needs no file clause. `Display Greeting` calls the Show action and produces `local hello|` without a newline.

Keep both files together. `helpers.l` resolves beside the importing source, regardless of the CLI's working directory or executable location. The [distribution CLI tests](../../acceptance/distribution/cli/cli_test.go) cover this behavior and the exact output.

## Preserve source-unit boundaries

Project files expand at their import positions. Builtins and indexed dictionaries retain separate scopes; dictionaries inherit builtins, not arbitrary caller locals. An alias preserves its member's canonical owner and identity without exposing private helpers.

Lookup checks a complete qualified binding before traversing members of its selected owner. A missing member does not trigger a search through unrelated names or filesystem paths. Import cycles reject; missing files and members produce distinct errors attributed to the original source.

## Use a complete custom resource bundle

`--assets ROOT` selects a complete replacement filesystem containing lexicon, builtins, stdlib and backend assets. It does not merge local overrides with embedded defaults. Relative asset paths use the invocation's working directory.

Only English (`eng`) is bundled. Another source-language code needs suitable assets and still faces partial localization support; selecting it does not translate a program.

See [CLI resources](../cli-build.md#invocation-and-assets), [resource architecture](../architecture.md#embedded-builtins-and-explicit-dictionaries), and [stdlib](../stdlib/index.md). Next, organize [Apps and Services](apps-and-services.md).
