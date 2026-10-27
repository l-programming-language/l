# Install L

## Install a released binary

Download the English L binary for your macOS or Linux host and CPU architecture from [GitHub Releases](https://github.com/l-programming-language/l/releases/latest). Language resources are embedded; no separate bundle is required. Extract it, then run from the directory containing `l`:

```sh
mkdir -p "$HOME/.local/bin"
chmod +x ./l
mv ./l "$HOME/.local/bin/l"
export PATH="$HOME/.local/bin:$PATH"
l --version
```

Add the `PATH` export to your shell configuration. `l --version` reports `0.0.1` for the first release. Read the [compatibility policy](../releases/index.md#versioning-and-compatibility) before upgrading.

## Prerequisites

Building L programs requires native C/linker tools and [rustup](https://rust-lang.org/tools/install/) with its Cargo proxy on `PATH`. L invokes `cargo +1.98.1`; standalone Cargo is insufficient. Use Xcode Command Line Tools on macOS or your Linux distribution's development tools. Native builds target the current host.

Running a published L program requires neither the L compiler nor Go/Rust toolchains. Go is needed only for the source-build and repository bootstrap steps below.

## Rust dependency preparation

Follow your release's toolchain and dependency setup instructions before the first native build. L uses pinned, frozen, offline Cargo dependencies; installing Rust alone does not populate that cache.

For repository-based setup, install Go 1.27.3 as pinned in [.go-version](../../.go-version), then run `CGO_ENABLED=1 scripts/l-ci bootstrap` from the matching source checkout. This downloads pinned tools and locked dependencies and compiles their caches; it can take several minutes. Preserve the retained locks rather than removing `--frozen`. See the [development policy](../development-harness.md#temporary-validation-pause).

## Source-build installation

To build the compiler yourself, install [Go](https://go.dev/doc/install) 1.27.3 and Git, then:

```sh
git clone https://github.com/l-programming-language/l.git
cd l
bash ./install.sh
export PATH="$PWD/bin:$PATH"
l --version
```

Check out the release's tag before building when you need a specific version. [install.sh](../../install.sh) verifies Go dependencies and builds the self-contained `bin/l` with `-mod=readonly`; it does not prepare Rust.

## First native build

From a source checkout containing the [basic demo](../../demo/basic/main.l), after completing dependency setup:

```sh
l build --output-dir ./out ./demo/basic/main.l
./out/main
```

The first command prints the executable path; the second runs it. Keep `out/main.apps/` beside `out/main` when moving the program. See [published files](build.md#published-files).
