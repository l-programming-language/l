# l build reference

## Command syntax

```text
l build [--output-dir DIRECTORY] [main.l]
```

English grammar, vocabulary, core and backend resources are embedded in `l`, so no resource selection or separate bundle is needed. Build options precede the single source filename, which defaults to `main.l` in the invocation directory. `--` ends option parsing for filenames beginning with a dash.

```sh
l build main.l
l build --output-dir './output directory' './source directory/main.l'
l build -- -example.l
```

Options accept single or double hyphens and either a separate value or `=value`. Prefer double hyphens. Empty values are rejected. Repeated options use the last value.

## Options

| Option | Default | Meaning |
| --- | --- | --- |
| `--output-dir DIRECTORY` | Source directory | Put the published executable and companion directory together here. |
| `--native-evidence-output PATH` | Disabled | Record explicit native-environment observations, actual Cargo results and artifact identities in a separate JSON document. See [build evidence](../cli-build.md#optional-native-build-evidence). |
| `--` | None | Stop flag parsing before the optional filename. |

Relative source and output paths use the invocation directory without changing it. [Resource overrides](../development-harness.md#compiler-defaults) support compiler development. Ordinary builds need neither override.

## Target OS

In `v0.0.1`, builds target the host OS and architecture. `-os`/`--os` is rejected. OS selection through Rust target configuration is planned after this release.

## Help

`--help`, `-h`, and `-help` show [source-derived CLI help](../_generated/cli-help.md) alone or after `build`, before a filename. Help writes to stdout and returns 0, or 1 if writing fails. After a filename, a help flag is another positional argument.

## Version

`l --version` or `l -v` prints the embedded `0.0.1` plus a newline for release `v0.0.1`, without compiler initialization or external resources. Stderr stays empty. Status is 0, or 1 if stdout fails. Use either alone: `l -v extra`, `l build --version` and `l build -v` return 2.

## Published files

For `main.l`, the executable is `main`. The retained publication machinery also writes `main.apps/` with ownership metadata and an immutable generation containing a manifest and `app-000`. The executable and generation copy contain the same standalone App artifact; there is no separate launcher or child-process runtime.

Artifacts finish in private staging before the executable is atomically replaced. Failed builds preserve working output. Unrelated destinations and symlinked generation directories reject. The [publication contract](../cli-build.md#output-and-publication) explains this retained layout.

## Streams and statuses

| Status | Boundary |
| --- | --- |
| 0 | Help, version output, or a successful native build. |
| 2 | Invalid invocation or unsupported host. |
| 1 | Resolution, asset initialization, compilation, generation, Cargo, publication, or output-write failure. |

Successful builds print only the absolute executable path and a newline. Diagnostics go to stderr. Building never runs source instructions, Apps or deployment scripts. The separately invoked program owns its status and stdout.

## Verification witnesses

[Version tests](../../cmd/l/command_version_test.go) check exact output and invalid combinations. The [command tests](../../cmd/l/command_test.go) cover help, defaults, option order and resource diagnostics. [Publication tests](../../cmd/l/build_publish_test.go) cover preserved prior outputs, matched concurrent generations and unrelated destinations. [Cargo identity tests](../../cmd/l/cargo_identity_integration_test.go) inspect genuine native artifact identity.
