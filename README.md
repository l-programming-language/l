<img src="website/public/assets/l/brand/logo-tile.svg" alt="L logo" width="112" height="112">

# L

L is a human-readable programming language implemented in Go. Lyger resolves
source meaning into AML, and Logiceme composes native Rust through logeme templates.

## Current development boundary

This checkout supports one standalone App using core Print. The standard
libraries and Runtime have been removed for reconstruction in L. Process
supervision, capture and runtime-dependent words are unavailable. Publication
is disabled, and existing releases describe earlier capabilities.

## A first program

The [basic demo](demo/basic/main.l) uses the implicit core vocabulary:

```l
"message" is `Hello World!`.
Create "Hello" as a New Basic App.
The App Prints message.
```

Print writes Text followed by a newline. Build and execute it with this checkout:

```sh
scripts/l-go run ./cmd/l build --output-dir .harness/hello demo/basic/main.l
.harness/hello/main
```

See the [first-program guide](docs/guide/first-program.md) and
[build guide](docs/cli/build.md) for the current boundary.

## Local setup

Use the pinned Go and Rust tools described in the
[development guide](docs/development-harness.md). `scripts/l-go` selects the
repository's Go version. Explicit `scripts/l-ci bootstrap` prepares pinned tools
and locked dependencies; build commands do not fetch missing tools implicitly.

Run `scripts/l-check` for the temporary hello-world gate. It checks actual source
compilation and execution, selected core authority, rejection of unsupported
runtime effects and frozen logeme assets. This is not full language acceptance.

## Contributing

Read [CONTRIBUTING.md](CONTRIBUTING.md), [architecture](docs/architecture.md), and
[documentation development](website/README.md). Keep compiler coverage useful
without retaining deleted vocabulary as a requirement. Do not change the frozen
logeme templates or restore hardcoded runtime implementations.

L is licensed under [Apache-2.0](LICENSE).
