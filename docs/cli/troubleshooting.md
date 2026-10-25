# Diagnose a build or program failure

## Invalid invocation

Status 2 with usage usually means a missing `build` command, an unknown command/option, an unimplemented OS selector, too many filenames, or options after the source. Move options before the filename and use `--` for a filename beginning with `-`. A non-macOS/non-Linux host also returns 2. `l --help` checks invocation without compiling.

## Missing source or resources

Check source paths relative to the invocation directory. Released binaries embed English resources. If testing modified compiler resources, see [developer overrides](../development-harness.md#compiler-defaults).

Bare imports resolve through the selected [stdlib index](../../stdlib/index.json). An adjacent `io.l` does not override the indexed `io` dictionary. Explicit `.l` paths resolve relative to the importing source.

## Syntax, bindings and types

Read the complete diagnostic and source location. A phrase can parse successfully but fail name resolution, type checking or native generation. The [error model](../error-model.md) explains these compiler error categories.

For source-backed definitions, check typed placeholders, literal delimiters and declaration scope. Consult the [logeme reference](../reference/logemes/index.md) for supported native forms.

## Cargo and publication

Check `cargo +1.98.1 --version`, then follow [dependency preparation](installation.md#rust-dependency-preparation) for missing offline packages. Preserve the retained locks. A Rust diagnostic after successful L parsing can indicate an invalid native fragment; source generation alone does not establish a successful build.

Publication rejects unrelated existing destinations and symlinked companion directories. Choose a new output directory or inspect ownership; do not delete an unknown directory to make a build succeed. Keep a working executable and its companion generations together. The compiler intentionally retains previous generations for running processes.

## Program failure after a successful build

Successful compilation does not execute the program. Reproduce by invoking the published path separately and record its status/stdout/stderr. Await before relying on finalized Output. Stop does not join. Building while an invocation is outstanding, duplicate Run, Await/Stop before Run, or reading unavailable Output are lifecycle misuses.

See the [runtime contract](../cli-build.md#app-execution-semantics) for cleanup and failure selection.
