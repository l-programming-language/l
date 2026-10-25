# Build and install the compiler

## Prerequisites

Use macOS or Linux with:

- Git and a native C compiler/linker (Xcode Command Line Tools on macOS, or your Linux distribution's development tools).
- [Go](https://go.dev/doc/install) 1.27.1 on `PATH`, matching [.go-version](../../.go-version). The module's Go 1.20 declaration specifies its source language version.
- [rustup](https://rust-lang.org/tools/install/) and its Cargo proxy on `PATH`, normally in `~/.cargo/bin`. L invokes `cargo +1.98.1`; a standalone Cargo installation is insufficient. The setup below installs that pinned toolchain.

Native builds target the current host. Published programs need neither the L compiler nor Go/Rust toolchains.

## Source-build installation

The organization import currently lives on `initial`. Clone that branch, then build the compiler:

```sh
git clone --branch initial https://github.com/l-programming-language/l.git
cd l
bash ./install.sh
./bin/l --help
```

[install.sh](../../install.sh) downloads and verifies Go dependencies, then builds `bin/l` with `-mod=readonly`. It does not prepare Rust. The compiler embeds its language resources, so you can move `bin/l` independently or add its directory to `PATH`.

This checkout has no official binary release, package-manager installation or version command. See [release status](../releases/index.md).

## Rust dependency preparation

Before the first native build, explicitly prepare the pinned toolchain and locked dependency caches:

```sh
CGO_ENABLED=1 scripts/l-ci bootstrap
cargo +1.98.1 --version
```

Bootstrap requires network access and the native C tools above. It installs pinned Rust and lint tools, fetches locked dependencies, and compiles Go/Rust dependency caches. It can take several minutes; it does not run the acceptance tests. Repository work remains subject to the [validation policy](../development-harness.md#temporary-validation-pause).

L builds use frozen, offline Cargo resolution. Installing Rust alone is insufficient: keep the prepared dependency cache and retained locks. For missing dependencies, complete bootstrap rather than removing `--frozen` or updating locks. Custom asset bundles must preserve compatible native templates and locked dependencies.

## First native build

From the repository root:

```sh
./bin/l build --output-dir ./out ./demo/basic/main.l
./out/main
```

The first command prints the absolute executable path; the second runs it. Keep `out/main.apps/` beside `out/main` when moving the program. See [published files](build.md#published-files).

The [basic demo](../../demo/basic/main.l) imports `IO.Print`, constructs a Basic App, executes it and shows its output.
