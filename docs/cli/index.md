# The L command line

The `l` compiler produces native executables on macOS or Linux. `l build` publishes a program; invoking that executable runs it.

## Available commands

Use `build` to compile, `--help` for usage, and [`--version`](build.md#version) to identify the installed compiler. Start with [installation](installation.md), then the [build reference](build.md) and [runtime guide](runtime.md).

```sh
l --version
l build ./demo/basic/main.l
./demo/basic/main
```

## Evidence and authority

[command.go](../../cmd/l/command.go) defines the CLI; [command tests](../../cmd/l/command_test.go) and [version tests](../../cmd/l/command_version_test.go) verify it. The [CLI/runtime contract](../cli-build.md) defines publication and process semantics. See [troubleshooting](troubleshooting.md) for failures.
