# Building and running L programs

The current build supports one standalone App with a core Print body. Print accepts Text and writes it followed by one newline. The CLI compiles the actual App source bundle directly; no orchestrator or runtime is linked.

```sh
l build main.l
./main
```

Build and execution are separate. Building prints the absolute executable path after successful native compilation and publication. It never runs the program or generated deployment scripts. See the [first program](guide/first-program.md).

## Invocation and assets

[Build options](cli/build.md) precede the optional source filename, which defaults to `main.l`. Use `--` for a filename starting with a dash. Builds target the native macOS or Linux host.

Core loads directly from the embedded English resources. The root dictionary registry is empty. An explicit `--assets` override selects one complete provider without changing CWD or falling back to embedded files. Program `.l` includes resolve beside their importer; dictionary-owned resources remain inside their selected provider/root. See [compiler defaults](development-harness.md#compiler-defaults).

## Supported source boundary

Use one App and core Print with Text. Multiple Apps, additional Service libraries and runtime-dependent top-level effects reject. Source-level Build, Run, Stop, Await, Start/Execute, Show, Error/Output inspection, capture and continuations are unavailable. Retained declarations and recognized grammar do not authorize those effects.

Keep unsupported effects explicit: the compiler must reject them rather than silently omit them or synthesize a runner. Runtime will be rebuilt in L.

## Optional native build evidence

```sh
l build --native-evidence-output build-environment.json main.l
```

This option records tool/environment observations, the actual Cargo build and artifact identity in a separate JSON document with schema `2.0`, analyzer `l-contracts-2.0/1` and build-record version `2`. Earlier schemas and removed fields reject. Choose a destination outside inputs, custom assets and build outputs. Evidence is written before executable publication; failure to write it preserves the previous published output.

The record identifies selected source and build inputs, tool executables, the native host, Cargo invocation, output artifact hashes and companion generation. The entry and generation copy contain the same compiled artifact. A failed Cargo invocation remains recorded as a failure, and a successful compilation remains recorded if later publication fails.

Evidence collection does not run the App or generated deployment scripts. It does not approve native behavior, establish runtime properties, or certify a different checkout. The [build-evidence observers](../cmd/l/native_environment_native_test.go) cover actual Cargo records, artifact identity and failure preservation; their presence is not a claim that they ran for the current checkout.

## Output and publication

For `main.l`, the published executable is `main`. The existing publication machinery also retains `main.apps/.l-owner` and an immutable generation containing a manifest and `app-000`. The entry and generation copy contain the same standalone Cargo artifact. The executable does not launch that copy or depend on a runtime companion.

All native artifacts are staged and verified before the executable is atomically replaced. Failed builds preserve working output. Unrelated destinations and symlinked companion/generation directories reject. The compiler removes only its own staging; it does not delete unknown files to make publication succeed.

## Source generation and native builds

Lyger retains source units, lexical scope, typed arguments, once-only evaluation and provenance while producing AML. Logiceme renders the selected App definitions through the frozen logemes. The CLI validates bundle dimensions, safe logical paths, target identities and dependencies before invoking Cargo. A leading slash in a bundle is a logical root, not a host absolute destination.

Builds use the pinned Rust toolchain, frozen offline dependency resolution and existing subprocess supervision. Cargo messages must identify the actual manifest and binary target. Cached products are reusable only with current input and artifact identity; successful generation or an old cache entry is not a successful build.

## Build implementation ownership

[Compiler coordination](../cmd/l/compiler.go) selects the standalone unit. [Source bundles](../cmd/l/source_bundle.go) validate returned files; Cargo planning and supervision live in `cmd/l/cargo_*.go`. [Publication](../cmd/l/build.go) stages and atomically replaces verified artifacts.

See [architecture](architecture.md), [compiler errors](error-model.md) and the [temporary hello-world gate](development-harness.md).
