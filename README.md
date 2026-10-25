# L

`l` is a human-readable programming language with applications as language values.
Lyger compiles `l` into AML/logeme markup using Rust source definitions.
The CLI builds a Rust workspace containing an orchestrator, companion Apps,
App-constructed Service libraries.

L's documentation site is configured for [ldocs.org](https://ldocs.org).
See the [website guide](website/README.md) to build and preview it locally.

L, Lyger and Logiceme remain implemented in Go; generated programs use Rust.
The [Rust reference](builtins/README.md) describes the unchanged frozen
logemes and their bounded interfaces. The [Lucid suite](logiceme/testdata/lucid/README.md)
owns backend generation/compilation coverage; the [frontend/runtime implementation
evidence](testdata/integration/IMPLEMENTATION.md) records revision-specific
verification. The raw Logiceme API returns Rust source bundles; its optional
settings control runtime support.

The compiler embeds its grammar, primitive vocabulary, source-defined builtins,
standard dictionaries and Rust backend assets. Projects contain their
own source and explicit local imports; no adjacent `builtin.l` is required.

## Local setup

Install Go and make sure `go version` works in your terminal. The module declares
Go 1.20 as its minimum version. Native executable builds support macOS and Linux.
Bash is required for the setup script, and the first setup needs internet access
to download the pinned Go dependency. Building L programs additionally requires
Rust/Cargo 1.98.1 and the pinned Rust dependencies; the compiler itself is still
built with Go.

From the repository root:

```bash
bash ./install.sh
./bin/l --help
```

The script verifies dependencies and builds the ignored `bin/l` executable.
Invoke it by path from any directory; run it rather than sourcing it.

Copy `bin/l` wherever you want to run it. Embedded defaults do not depend on the
executable location or current directory, and `go run ./cmd/l` uses the same
resources. Compiling generated programs uses Cargo; running a published program
requires neither compiler toolchain nor L installation.

## Build and run

From the repository root:

```bash
./bin/l build ./demo/phase1/main.l
./demo/phase1/main
```

`l build` writes the orchestrator beside the source as `main` and companion
App binaries under `main.apps/`. A successful build prints the absolute
orchestrator path. Building does not run the program; execute `main` to run
the source program. Keep the executable and its `.apps` directory together when
moving the output. Both the entry and companions are native Rust binaries;
running them does not require L, Go, Cargo or compiler assets.

The [basic demo](demo/basic/main.l) prints Hello World. The [paragraph examples](demo/paragraphs/README.md) illustrate scoped bindings, continuations and intentional error exits. Each source imports `io` explicitly.

The command syntax is:

```text
l build [--language LANGUAGE] [--assets ROOT] [--output-dir DIRECTORY] [main.l]
```

- The source defaults to `main.l` in the calling directory.
- Options follow `build` and precede the source filename. Use `--` before a
  filename that begins with a dash.
- `--language LANGUAGE` defaults to `eng`, the only bundled language.
  Alternate languages require suitable assets; [localization remains partial](docs/architecture.md#source-meaning-and-output-views).
- `--assets ROOT` selects a complete resource bundle containing `lexicon`,
  `builtins`, and `stdlib`, with no compiler configuration file or embedded fallback.
- `--output-dir DIRECTORY` writes both outputs there, preserving the source's
  filename stem. Relative paths use the calling directory.
- Builds target the current macOS or Linux host and architecture. The legacy
  `-os` option and cross-compilation are unsupported.

For example:

```bash
./bin/l build --output-dir ./dist/build ./demo/phase1/main.l
go run ./cmd/l build --output-dir ./dist/build ./demo/phase1/main.l
```

An existing output produced by `l` can be rebuilt. Unrelated files or companion
directories at the destination are rejected. Rebuilds retain earlier companion
generations so already-running orchestrators keep their companion binaries.

CLI exit status is 0 for success or help, 2 for invalid invocation or an
unsupported host, and 1 for asset, initialization, compilation, build, or output-write failures.
Compiler diagnostics go to stderr. These statuses are separate from the built
program's own exit behavior.

See the [CLI build guide](docs/cli-build.md) for output layout and runtime details.

## Execution model

The orchestrator executes source instructions in order. Apps
provide `Build`, `Run`, `Stop`, `Await`, `Output`, and `Error`. `Run` launches a separate
OS process and returns after launch acknowledgment. `Await` joins the finalized
result and executes that invocation's continuation inline, once. `Stop` requests
forceful termination of the target's process group; use Await to join it. A new Run
is allowed after the first Await continuation finishes. Aliases share the same
target and invocation.

Executables are built ahead of time; runtime `Build` checks artifact readiness.
`Output` is finalized raw stdout, preserving whitespace, newlines and partial
failed-run output. Reading it before completion fails; it does not implicitly
wait. Stderr is captured separately. `Show` adds no newline. Normal exit cleans
unawaited processes without running their continuations.

Start and Execute remain source-defined App actions: Start composes Build/Run,
and Execute composes Start/Await. An App can also use the primitives directly:

```l
Use IO's Print from `io` as "Print".

Create "Worker" as a New Basic App.
The App Prints `work completed`.

Run Worker.
Await Worker.
Show Worker's Output.
```

The program above prints `work completed` followed by one newline.
Write Await wherever later instructions need a Run result, and place a same-target
rerun outside its active continuation. See [execution semantics](docs/cli-build.md#app-execution-semantics).

Native builds use Cargo workspace members, pinned runtime dependencies and
validated local library dependencies. Runtime support uses Tokio for process
supervision. The typed frontend plan
preserves paragraph activations, assignment checks and value snapshots without
adding sections or requiring users to write ownership annotations. General native
ownership inference, Phase 2 in-code concurrency, container packaging and
cross-compilation remain outside this build command.

## Validation and development

Start with [AGENTS.md](AGENTS.md) and the [development guide](docs/development-harness.md).

**Expensive local validation remains paused.** Pull requests automatically run
the full Linux gate when opened, updated with commits, or reopened. Manual
dispatch also supports macOS. Full local checks require an explicit request;
see the [validation policy](docs/development-harness.md#temporary-validation-pause).
For an authorized local run:

```bash
./scripts/l-ci bootstrap
GOMAXPROCS=4 ./scripts/l-check
```

Bootstrap installs pinned tools and prepares dependency compiler caches without
running acceptance tests or generated Apps. When requested, `l-check`
requires zero findings from workflow linting and both
full-module Go lint profiles, harness `doctor` and unfiltered `verify`, and the
complete tagged native-compilation suite. Protected fixtures and expected bytes
remain checked. Fresh logs and structured reports are written under ignored
`.harness/runs/` directories. Focused diagnostic probes do not replace completion.
Linux CI success does not establish macOS validation. The development guide owns
the current workflow and branch-protection policy.

Maintained references: [architecture](docs/architecture.md), [CLI and runtime](docs/cli-build.md), [development and acceptance](docs/development-harness.md), and [compiler errors](docs/error-model.md).
