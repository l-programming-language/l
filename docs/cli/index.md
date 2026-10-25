# The L command line

The `l` program compiles L source into native Rust executables on macOS or Linux. It is implemented in Go. Compilation and execution are separate: `l build` publishes a program; invoking the resulting executable runs it.

## Available commands

The current parser implements `build` and help. It has no `run`, `test`, `init`, `fmt`, `docs`, or `--version` command. Use the [installation guide](installation.md), then the [build reference](build.md). The [runtime guide](runtime.md) explains how the published program differs from its compiler.

```sh
./bin/l --help
./bin/l build ./demo/basic/main.l
./demo/basic/main
```

## Evidence and authority

The parser and exact help text are in [command.go](../../cmd/l/command.go). [Parser and stream tests](../../cmd/l/command_test.go) cover invocation, resource selection, help and errors. The maintained [CLI/runtime contract](../cli-build.md) owns publication and process semantics. [Troubleshooting](troubleshooting.md) separates invalid invocations from compiler, native tool and program failures.

L's `--language` selects a source lexicon. The English website interface is independent of this option. Only `eng` is bundled in this checkout; an option that accepts a language identifier is not evidence of additional supported source languages.
