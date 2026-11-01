# Installation options

Start with [Install L](installation.md) for the recommended binary installation. This page covers platform requirements, custom setup, upgrades, and building the compiler from source.

## Supported platforms

Published binaries support Apple Silicon macOS (ARM64) and Linux (ARM64 and x86-64). Intel macOS is not supported; contributions toward future support are welcome. Native builds target the current host.

## Native build tools

L needs a C compiler and linker to build native programs. Install your system's tools before running the L installer:

- **macOS:** use Xcode Command Line Tools. If they are missing, run `xcode-select --install` and finish Apple's installation prompt.
- **Linux:** install your distribution's C compiler and linker development packages.

Initial setup requires internet access. The L installer can prepare Rust tools and dependencies, but it does not use `sudo` or invoke your system package manager.

## Shell setup

For the default installation, add this line to your shell's startup file:

```sh
export PATH="$HOME/.local/bin:${CARGO_HOME:-$HOME/.cargo}/bin:$PATH"
```

For Zsh, use `~/.zshrc`. For interactive Bash, use `~/.bashrc`, or `~/.bash_profile` for login shells that do not read it. Open a new terminal and run `l --version` to check the result. The installer does not modify these files.

If you use `--prefix`, replace `$HOME/.local/bin` with that directory. If you use a custom `CARGO_HOME`, define it before the export line. Other shells need their own PATH syntax.

## Installer options

Pass these options to the [downloaded installer](installation.md#install-a-released-binary):

| Option | Effect |
| --- | --- |
| `--with-toolchain` | Permit installing missing Rust tools. Otherwise interactive setup asks; noninteractive setup requires them to exist. |
| `--binary-only` | Install only `l`, deferring native prerequisites. |
| `--prefix DIRECTORY` | Choose the executable directory instead of `~/.local/bin`. |
| `--version VERSION` | Pin a release; omission selects the latest normal release once for all assets. |
| `--repository OWNER/REPO` | Intentionally select another repository's release. |

Language resources are embedded in the binary. The installer verifies the archive's SHA-256 checksum and installs it only after verification and requested setup succeed.

## Upgrades and repair

To upgrade, follow the [installation steps](installation.md), replacing the version in both the download URL and `--version` with the [release you want](https://github.com/l-programming-language/l/releases). Read the [compatibility policy](../releases/index.md#versioning-and-compatibility) first. A failed download, verification, or prerequisite setup preserves the existing compiler binary.

To prepare missing Rust tools or dependencies after a binary-only installation, rerun the installer with an explicit version and `--with-toolchain`. For L 0.0.1, use `bash ./install-release.sh --version 0.0.1 --with-toolchain`. This also reinstalls the selected version of `l`.

To remove a default installation, delete only `~/.local/bin/l` and remove its PATH entry if you no longer need it. Keep shared Rust tools and directories used by other projects.

## Manual archives

Download the archive for your platform and `checksums.txt` from the same [GitHub release](https://github.com/l-programming-language/l/releases). Verify its SHA-256 checksum before extracting it. Keep the setup files together, then run `bash setup-release.sh --with-toolchain` from the extracted directory. Put `l` in a directory on your PATH and keep the rustup Cargo proxy available as described in [shell setup](#shell-setup).

## macOS approval

macOS releases below `v1.0.0` have no Developer ID signature or Apple notarization. macOS may block them or require approval, depending on download method and security settings. Follow [Apple's guidance](https://support.apple.com/102445) only after checking that you trust the source. Production macOS releases require signing and notarization from `v1.0.0` onward.

## Source-build installation

Building the L compiler from source also needs Git and [Go](https://go.dev/doc/install) 1.27.3, pinned in [.go-version](../../.go-version).

```sh
git clone https://github.com/l-programming-language/l.git
cd l
git checkout v0.0.1
bash ./install.sh
export PATH="$PWD/bin:${CARGO_HOME:-$HOME/.cargo}/bin:$PATH"
l --version
```

The tag selects the released compiler source; use another release tag when needed. [install.sh](../../install.sh) verifies Go dependencies and builds the self-contained `bin/l` with `-mod=readonly`. It does not prepare Rust. Keep the absolute path to this checkout's `bin` directory in your shell configuration if you use this installation in new terminals.

## Rust dependency preparation

L uses rustup, Rust 1.98.1, and the retained Cargo dependencies. It invokes `cargo +1.98.1`, so standalone Cargo is insufficient. Release setup installs the minimal Rust toolchain, fetches locked dependencies, and checks that they resolve offline without changing the lockfile. On a fresh rustup installation, release setup selects Rust 1.98.1 as the default toolchain.

For binary installations, use the installer or extracted archive's setup script as described above. For source installations, the current preparation path is `CGO_ENABLED=1 scripts/l-ci bootstrap` from the matching checkout. This also prepares contributor tools and validation caches and can take several minutes; see the [development policy](../development-harness.md#temporary-validation-pause).

Installing Rust alone does not populate L's dependency cache. Native compilation uses frozen, offline Cargo resolution; preserve the retained locks rather than removing `--frozen`. See [troubleshooting](troubleshooting.md#cargo-and-publication) if a build still fails after setup.
