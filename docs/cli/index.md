# The L command line

L builds native macOS/Linux programs. Run the published executable directly; see the [runtime guide](runtime.md).

## Available commands

| Invocation | Behavior |
| --- | --- |
| `l build [main.l]` | Compile and publish; source defaults to `main.l`. |
| `l --help`, `l -h`, `l -help` | Show usage. Use alone. |
| `l build --help` | Show build usage before a filename; also accepts `-h`, `-help`, `--h`. |
| `l --version`, `l -v` | Print `0.0.1` plus a newline. Use alone. |

Build options: [`--output-dir` and `--`](build.md#options). Compiler contributors have [`--assets` and `--language` overrides](../development-harness.md#compiler-defaults). See [installation](installation.md) and [troubleshooting](troubleshooting.md).

## Evidence and authority

The [parser](../../cmd/l/command.go), [command tests](../../cmd/l/command_test.go), [version tests](../../cmd/l/command_version_test.go) and [runtime contract](../cli-build.md) define these interfaces.
