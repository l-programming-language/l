# Build and install the compiler

## Prerequisites

The L compiler is a Go program. The module declares Go 1.20 as its minimum language version; repository validation pins Go 1.27.1 in [.go-version](../../.go-version). Native L program builds require macOS or Linux, Rust/Cargo 1.98.1, and the retained Cargo dependencies already available locally. Running a published program requires neither Go, Rust, Cargo, nor an installed L compiler.

The versions serve different purposes. The Go module declaration describes the compiler's source language; the repository tool pin describes reproducible development validation. Rust's native host is the output target. The command does not cross-compile or accept an operating-system selector.

## Source-build installation

Read the short [installation script](../../install.sh), then run it from the checkout:

```sh
bash ./install.sh
./bin/l --help
```

The script requires Go on `PATH`, downloads and verifies Go modules, and builds the ignored `bin/l` binary with `-mod=readonly`. It does not install Go or Rust, change global settings, or provision the Rust dependency cache. It needs network access for uncached Go dependencies. Run it; do not source it into your shell.

The compiler embeds the lexicon, builtins, standard dictionaries, and native generation assets. Copy `bin/l` to a directory of your choice, or use its absolute path. If you maintain a personal executable directory, add that directory to `PATH` using your shell's normal configuration. The binary does not require the repository beside it and does not discover assets from its current working directory.

There is no authoritative released version, binary download channel, package-manager installation, or version command established by this checkout. The website therefore labels this documentation Local/Unreleased. Its [release page](../releases/index.md) explains provenance instead of inventing a download.

## Rust dependency preparation

The CLI deliberately uses frozen, offline Cargo resolution. Merely installing `rustc` does not populate its dependencies. The repository's [development guide](../development-harness.md) describes `scripts/l-ci bootstrap`, the pinned tools, retained dependency locks, and the preparation process. Bootstrap includes expensive compilation and tool installation; it is an explicitly authorized development operation, not part of website setup or an automatic documentation build.

This documentation task does not run bootstrap or the compiler. A missing pinned toolchain or offline dependency is a prerequisite failure, not permission to remove `--frozen`, update locks, or install arbitrary dependencies. A configured complete asset bundle must retain compatible support templates and locked dependency contracts.

## First native build

From the repository root after prerequisites have been prepared:

```sh
./bin/l build --output-dir ./out ./demo/basic/main.l
./out/main
```

The first command prints the absolute path of `out/main` when publication succeeds. The second executes it. Preserve `out/main.apps/` beside it when moving or distributing the result. The [build reference](build.md#published-files) describes the immutable companion generations and the difference between build output and application stdout.

The [basic demo](../../demo/basic/main.l) imports `IO.Print` explicitly and uses implicit Basic/Execute definitions. Its source is the canonical example; generated Rust and historical outputs must not be substituted for compiling that source.
