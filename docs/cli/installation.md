# Build L from this checkout

Use the current checkout's compiler for the core-only standalone App/Print path. Previously released binaries can embed removed dictionaries and runtime behavior.

## Prerequisites

Use native macOS or Linux with a C compiler/linker and the pinned Go and Rust tools recorded in [.ci-tools.json](../../.ci-tools.json). Native App builds use frozen, offline Cargo resolution, so retained dependencies must already be available. See [native build tools](installation-options.md#native-build-tools) for host setup.

## Build the compiler

From the repository root:

```sh
scripts/l-go build -mod=readonly -o /tmp/l ./cmd/l
/tmp/l --version
/tmp/l build demo/basic/main.l
./demo/basic/main
```

The last command prints `Hello World!` followed by one newline. The compiler embeds the current English resources and does not need a project resource copy.

<a id="activate-and-check-l"></a>
<a id="source-build-installation"></a>

## Use your compiler

Invoke the built compiler by its path or place it in a directory on your PATH. The [first-program guide](../guide/first-program.md) uses `l` to refer to that current compiler. Building an App and running it are separate actions.

<a id="rust-dependency-preparation"></a>

## Tools and validation

A missing tool or offline dependency is a setup failure. Preserve pins and locks; do not remove frozen/offline resolution or restore old library resources to make a build succeed. Follow [development and acceptance](../development-harness.md) for the temporary hello-world gate.

[Installation options](installation-options.md) describe retained release tooling. A historical release is not a substitute for this checkout when checking the reset.
