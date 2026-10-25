# l build reference

## Command syntax

```text
l build [--language LANGUAGE] [--assets ROOT] [--output-dir DIRECTORY] [main.l]
```

Build options follow `build` and precede the source filename. With no source argument, `main.l` is resolved relative to the invocation directory. Exactly one positional filename is accepted. `--` ends option parsing, including for a filename beginning with a dash.

```sh
l build
l build --language eng --output-dir './output directory' './source directory/main.l'
l build -- -example.l
```

Options accept single or double hyphens and either a separate value or `=value`. Prefer the double-hyphen forms above. Empty values are rejected; whitespace-only or unavailable language identifiers fail initialization. If an option repeats, its last value wins.

## Options

| Option | Default | Meaning |
| --- | --- | --- |
| `--language LANGUAGE` | `eng` | Select the source lexicon and matching source resources. Only English is bundled. |
| `--assets ROOT` | Embedded resources | Select one complete filesystem containing `lexicon`, `builtins`, and `stdlib`. |
| `--output-dir DIRECTORY` | Source directory | Put the published executable and companion directory together here. |
| `--` | None | Stop flag parsing before the optional filename. |

Relative source, asset and output paths use caller CWD. Resolution does not change CWD. An explicit asset provider has no embedded fallback and no compiler configuration JSON. Bare dictionary imports use that provider's standard-library index; explicit `.l` imports remain relative to the importing source's provider.

## Legacy OS option

The legacy `-os`/`--os` option is rejected. Native builds target the host operating system and architecture.

## Help

`--help`, `-h`, and `-help` show [CLI usage](../../cmd/l/command.go) alone or after `build`, before a filename. Help writes to stdout and returns 0, or 1 if writing fails. After a filename, a help flag is another positional argument.

## Version

`l --version` writes the embedded `major.minor.patch` version and one newline to stdout: `0.0.1` for release `v0.0.1`. Stderr stays empty; success returns 0 and an output-write failure returns 1. Supply it alone: extra arguments, `l build --version`, and aliases such as `-v` are invalid (status 2). No checkout, assets, or Rust toolchain is needed.

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
