# Install L

## Install a released binary

**Release publication is currently disabled; no public binary download is available yet.**
The implemented release process targets macOS and Linux on ARM64 and x86-64.
Binary releases and the installer will be hosted in `l-programming-language/l`
after this repository becomes public and publication is explicitly enabled.

After a release is published, download its `install-release.sh` asset and inspect
it before running it. The installer selects the host archive, checks its SHA-256
checksum, prepares dependencies, and installs `l` into `~/.local/bin`. Language
resources are embedded; no separate compiler bundle or Go installation is needed.

macOS releases below `v1.0.0` are distributed without Developer ID signing or
Apple notarization. Depending on how the binary is downloaded and your security
settings, macOS may block it or require explicit approval. Follow
[Apple's guidance for opening downloaded software](https://support.apple.com/102445)
only after checking that you trust its source. Developer ID signing and
notarization are required for production macOS releases starting with `v1.0.0`.

For the planned first release:

```sh
bash ./install-release.sh --version 0.0.1 --with-toolchain
export PATH="$HOME/.local/bin:$HOME/.cargo/bin:$PATH"
l --version
```

Add that `PATH` export to your shell configuration. The installer never edits a
shell profile or uses sudo. `--with-toolchain` permits installing missing Rust
tools; without it, interactive setup asks first, and noninteractive setup requires
the tools to exist already. `--binary-only` installs only `l` and defers native
build prerequisites. `--prefix DIRECTORY` selects the executable directory.

Omitting `--version` selects the latest normal release once, then downloads every
asset from that version in `l-programming-language/l`. `--repository OWNER/REPO`
overrides that source when intentionally installing another repository's release.
Repeating installation or upgrading replaces the binary
only after verification and the requested setup complete; failure preserves an
existing installation. Read the [compatibility policy](../releases/index.md#versioning-and-compatibility)
before upgrading.

## Prerequisites

Building L programs requires native C/linker tools and [rustup](https://rust-lang.org/tools/install/) with its Cargo proxy on `PATH`. L invokes `cargo +1.98.1`; standalone Cargo is insufficient. Use Xcode Command Line Tools on macOS or your Linux distribution's development tools. Native builds target the current host.

Running a published L program requires neither the L compiler nor Go/Rust toolchains. Go is needed only for the source-build and repository bootstrap steps below.

## Rust dependency preparation

The release installer includes `setup-release.sh`, the pinned toolchain version,
and the retained dependency manifest/lock. It fetches locked dependencies and
then verifies frozen resolution. L's subsequent compilation remains frozen and
offline; installing Rust alone does not populate the cache. Setup preserves the
existing Rust default toolchain and never requires Go, linters or the repository.

For an archive installed manually, run its `bash setup-release.sh --with-toolchain`
before the first native build, with the setup files still alongside it. Missing
C/linker tools require your host's normal installation procedure; the installer
does not invoke a system package manager.

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

After completing dependency setup, save this program as `main.l` in a new
directory; no compiler source checkout is needed:

```l
Use IO's Print from `io` as "Print".

"message" is `Hello World!`.
Create "Hello" as a New Basic App.
The App Prints message.
Execute Hello.
Show Hello's Output.
```

```sh
l build --output-dir ./out ./main.l
./out/main
```

The first command prints the executable path; the second runs it. Keep `out/main.apps/` beside `out/main` when moving the program. See [published files](build.md#published-files).
