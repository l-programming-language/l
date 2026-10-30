# Install L

Check [GitHub Releases](https://github.com/l-programming-language/l/releases) for published binaries, or use the [source-build fallback](#source-build-installation). The commands below pin `v0.0.1` and become available when that release is published.

## Prerequisites

Building L programs needs native C/linker tools: use Xcode Command Line Tools on macOS or your Linux distribution's development tools. Initial setup requires internet access. The binary installer can prepare [rustup](https://rust-lang.org/tools/install/), Rust 1.98.1 and the locked dependencies; L invokes `cargo +1.98.1`, so standalone Cargo is insufficient.

Native builds target the current host. Running a published L program requires neither L nor Go/Rust toolchains.

## Install a released binary

Download the versioned installer for macOS or Linux on ARM64 or x86-64:

```sh
curl --fail --location --proto '=https' --tlsv1.2 \
  https://github.com/l-programming-language/l/releases/download/v0.0.1/install-release.sh \
  --output install-release.sh
```

Inspect the script, then run it:

```sh
bash ./install-release.sh --version 0.0.1 --with-toolchain
export PATH="$HOME/.local/bin:$HOME/.cargo/bin:$PATH"
l --version
```

The installer selects the host archive, verifies its SHA-256 checksum, prepares dependencies, and installs `l` into `~/.local/bin`. Language resources are embedded; binary installation needs no Go or compiler source checkout. Add the `PATH` export to your shell configuration. The installer never edits profiles, uses sudo, or invokes a system package manager.

| Option | Effect |
| --- | --- |
| `--with-toolchain` | Permit installing missing Rust tools. Otherwise interactive setup asks; noninteractive setup requires them to exist. |
| `--binary-only` | Install only `l`, deferring native prerequisites. |
| `--prefix DIRECTORY` | Choose the executable directory. |
| `--version VERSION` | Pin a release; omission selects the latest normal release once for all assets. |
| `--repository OWNER/REPO` | Intentionally select another repository's release. |

Setup fetches locked dependencies and verifies frozen resolution without changing the default Rust toolchain. For a manually extracted archive, run `bash setup-release.sh --with-toolchain` with its setup files still alongside it.

Upgrades replace the binary only after verification and requested setup succeed; failure preserves the existing installation. Read the [compatibility policy](../releases/index.md#versioning-and-compatibility) before upgrading.

macOS releases below `v1.0.0` have no Developer ID signature or Apple notarization. macOS may block them or require approval, depending on download method and security settings. Follow [Apple's guidance](https://support.apple.com/102445) only after checking that you trust the source. Production macOS releases require signing and notarization from `v1.0.0` onward.

## First native build

After dependency setup, [build your first program](../guide/first-program.md). The walkthrough covers saving a source file, compiling it, and running the result without a compiler source checkout.

## Source-build installation

Source installation also needs Git and [Go](https://go.dev/doc/install) 1.27.3, pinned in [.go-version](../../.go-version).

```sh
git clone https://github.com/l-programming-language/l.git
cd l
bash ./install.sh
export PATH="$PWD/bin:$PATH"
l --version
```

Check out the release's tag before building when you need a specific version. [install.sh](../../install.sh) verifies Go dependencies and builds the self-contained `bin/l` with `-mod=readonly`; it does not prepare Rust.

## Rust dependency preparation

For source installation, run `CGO_ENABLED=1 scripts/l-ci bootstrap` from the matching checkout to download pinned tools and locked dependencies and compile their caches. This can take several minutes; see the [development policy](../development-harness.md#temporary-validation-pause).

L's native compilation uses frozen, offline Cargo resolution. Installing Rust alone does not populate its dependency cache; preserve the retained locks rather than removing `--frozen`.
