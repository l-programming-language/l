# l build reference

## Command syntax

```text
l build [--output-dir DIRECTORY] [main.l]
```

English grammar, vocabulary, builtins, standard library and backend resources are embedded in `l`; no resource selection or separate bundle is needed. Build options precede the single source filename, which defaults to `main.l` in the invocation directory. `--` ends option parsing for filenames beginning with a dash.

```sh
l build main.l
l build --output-dir './output directory' './source directory/main.l'
l build -- -example.l
```

Options accept single or double hyphens and either a separate value or `=value`. Prefer double hyphens. Empty values are rejected; repeated options use the last value.

## Options

| Option | Default | Meaning |
| --- | --- | --- |
| `--output-dir DIRECTORY` | Source directory | Put the published executable and companion directory together here. |
| `--` | None | Stop flag parsing before the optional filename. |

Relative source and output paths use the invocation directory without changing it. Compiler contributors can use [developer resource overrides](../development-harness.md#compiler-defaults); ordinary builds need neither override.

## Target OS

In `v0.0.1`, builds target the host OS and architecture; `-os`/`--os` is rejected. OS selection through Rust target configuration is planned after this release.

## Help

`--help`, `-h`, and `-help` show [source-derived CLI help](../_generated/cli-help.md) alone or after `build`, before a filename. Help writes to stdout and returns 0, or 1 if writing fails. After a filename, a help flag is another positional argument.

## Version

`l --version` or `l -v` prints the embedded `0.0.1` plus a newline for release `v0.0.1`, without compiler initialization or external resources. Stderr stays empty; status is 0, or 1 if stdout fails. Use either alone: `l -v extra`, `l build --version` and `l build -v` return 2.

## Published files

![Cargo builds the orchestrator, App binaries and additional App-constructed Service libraries. Standalone Service declarations produce no artifacts.](../../website/public/assets/l/illustrations/native-builds.svg)

For `main.l`, successful publication yields:

```text
main
main.apps/
  .l-owner
  generations/
    <generation-id>/
      manifest.json
      app-000
```

The version-1 manifest lists App binaries under `apps`. App-constructed Service libraries are build dependencies, not executable companions; standalone Service declarations add no packages.

Builds privately compile and validate all artifacts before publishing an immutable companion generation and atomically replacing the orchestrator. Failures preserve working output; old generations remain for running programs. Unrelated outputs and symlinked companion/generation directories are rejected.

Move the executable and its `.apps` directory together: it finds companions relative to itself. The source filename stem determines the output basename. The [publication contract](../cli-build.md#output-and-publication) covers failure and concurrent-build details.

## Streams and statuses

| Status | Boundary |
| --- | --- |
| 0 | Help, version output, or a successful native build. |
| 2 | Invalid invocation or unsupported host. |
| 1 | Resolution, asset initialization, compilation, generation, Cargo, publication, or output-write failure. |

Successful builds print only the absolute executable path and a newline; diagnostics go to stderr. Building never runs source instructions, Apps or deployment scripts. The separately invoked program owns its status and stdout.

## Verification witnesses

[Version tests](../../cmd/l/command_version_test.go) check exact output and invalid combinations; [command tests](../../cmd/l/command_test.go) cover help, defaults, option order and resource diagnostics. [Publication tests](../../cmd/l/build_publish_test.go) cover preserved prior outputs, matched concurrent generations and unrelated destinations. [Cargo identity tests](../../cmd/l/cargo_identity_integration_test.go) inspect genuine native artifact identity.
