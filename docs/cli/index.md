# The L command line

L builds native macOS/Linux programs. Run the published executable directly. See the [runtime guide](runtime.md).

## Available commands

| Invocation | Behavior |
| --- | --- |
| `l build [main.l]` | Compile and publish. Source defaults to `main.l`. |
| `l contracts inspect INPUT` | Inspect bindings, typed targets, and contract obligations without building or running Apps. |
| `l contracts check --profile NAME INPUT` | Enforce selected contract profiles. |
| `l contracts explain --binding SELECTOR INPUT` | Explain a selected binding. |
| `l contracts verify-dictionary DICTIONARY` | Verify a dictionary. Focused symbolic adjective verification is explicit. |
| `l contracts --help` | Show contracts usage. Also accepts `-h`. |
| `l --help`, `l -h`, `l -help` | Show usage. Use alone. |
| `l build --help` | Show build usage before a filename. Also accepts `-h`, `-help`, `--h`. |
| `l --version`, `l -v` | Print `0.0.1` plus a newline. Use alone. |

See the [contracts reference](../contracts.md#commands) for options, scoped symbolic verification, reports, and exit statuses.

Build options: [`--output-dir` and `--`](build.md#options). Compiler contributors have [`--assets` and `--language` overrides](../development-harness.md#compiler-defaults). See [installation](installation.md) and [troubleshooting](troubleshooting.md).

## Evidence and authority

The [parser](../../cmd/l/command.go), [command tests](../../cmd/l/command_test.go), [version tests](../../cmd/l/command_version_test.go), [contracts parser](../../cmd/l/contracts.go), [contracts tests](../../cmd/l/contracts_test.go), [source verification tests](../../cmd/l/contracts_source_test.go), and [runtime contract](../cli-build.md) define these interfaces.
