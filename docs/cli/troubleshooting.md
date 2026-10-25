# Diagnose a build or program failure

## Invalid invocation

Status 2 with usage usually means a missing `build` command, an unknown command/option, a legacy OS selector, too many filenames, or options after the source. Move options before the filename and use `--` for a filename beginning with `-`. A non-macOS/non-Linux host also returns 2. `l --help` is a useful parser-only check.

## Missing source or resources

Check the path relative to the invocation directory. `--assets` is a complete replacement, not an overlay; a missing vocabulary, grammar, builtin, dictionary or native asset fails without embedded fallback. A resource directory from an older checkout may no longer satisfy the selected compiler. Ordinary users should prefer the embedded bundle.

Bare imports resolve through the selected [stdlib index](../../stdlib/index.json). An adjacent `io.l` does not override the indexed `io` dictionary. Explicit `.l` paths resolve relative to the importing source. Do not change CWD inside a build helper to conceal ownership errors.

## Syntax, bindings and types

Preserve the complete diagnostic, source location and phase. A grammar-recognized phrase may fail later because its ownership or Rust lowering is unsupported. Distinguish unresolved names, ambiguous references, duplicate declarations, type mismatches and import/expansion cycles. The [error model](../error-model.md) defines Go compiler error categories; it is not an L exception-handling library.

For source-backed definitions, check typed placeholders, literal framing and source ownership before editing native templates. The [logeme reference](../reference/logemes/index.md) documents a bounded backend; unused table entries are not feature promises. Documentation discrepancies should be recorded, not repaired by changing protected fixtures or compiler behavior during documentation work.

## Cargo and publication

Confirm pinned Rust/Cargo 1.98.1 and the retained offline dependencies are present. Frozen-resolution failures are not fixed by casually updating dependency locks. Native compiler diagnostics can expose an invalid supplied Rust fragment even when L parsing and source generation succeeded. A generated source file is not proof of a compiled executable.

Publication rejects unrelated existing destinations and symlinked companion directories. Choose a new output directory or inspect ownership; do not delete an unknown directory to make a build succeed. Keep a working executable and its companion generations together. The compiler intentionally retains previous generations for running processes.

## Program failure after a successful build

Successful compilation does not execute the program. Reproduce by invoking the published path separately and record its status/stdout/stderr. Await before relying on finalized Output. Stop does not join. Building while an invocation is outstanding, duplicate Run, Await/Stop before Run, or reading unavailable Output are lifecycle misuses.

See the [runtime contract](../cli-build.md#app-execution-semantics) for cleanup and failure selection. Website validation does not run the expensive compiler/native suites. The [development policy](../development-harness.md#temporary-validation-pause) governs separately authorized investigation and acceptance.
