# Installation options

Start with [Install L](installation.md) for the recommended versioned installer. This page covers platform requirements, custom locations, unattended installation, upgrades, and building the compiler from source.

## Supported platforms

Published binaries support Apple Silicon macOS (ARM64) and Linux (ARM64 and x86-64). Intel macOS is not supported. Contributions toward future support are welcome. Native builds target the current host.

## Native build tools

L needs a C compiler and linker to build native programs:

- **macOS:** run `xcode-select --install` and finish Apple's Xcode Command Line Tools installation prompt.
- **Linux:** install your distribution's C compiler and linker development packages.

Initial setup requires internet access. The installer prepares Rust tools and dependencies without `sudo` or a system package manager.

<a id="download-and-inspect-the-release-installer"></a>

## Download and inspect

Each [GitHub release](https://github.com/l-programming-language/l/releases) retains its own installer. To inspect the L 0.0.1 script before running it:

```sh
curl --fail --location --proto '=https' --tlsv1.2 \
  https://github.com/l-programming-language/l/releases/download/v0.0.1/install-release.sh \
  --output install-release.sh &&
less install-release.sh
```

After reviewing the successfully downloaded script:

```sh
bash ./install-release.sh --version 0.0.1 --with-toolchain &&
export PATH="$HOME/.local/bin:${CARGO_HOME:-$HOME/.cargo}/bin:$PATH" &&
l --version
```

The installer verifies the archive checksum and binary version before replacing L. For another pinned release, change both the download URL and `--version`.

## Installer options

Pass these options to `bash ./install-release.sh`:

| Option | Effect |
| --- | --- |
| `--version VERSION` | Pin a release. Omission selects the latest normal release once for all assets. |
| `--prefix DIRECTORY` | Choose the executable directory instead of `~/.local/bin`. |
| `--with-toolchain` | Permit installing missing Rust tools without a prompt. Otherwise, interactive setup asks and noninteractive setup requires them to exist. |
| `--binary-only` | Install only `l`, deferring native prerequisites. |
| `--repository OWNER/REPO` | Intentionally select another repository's release. |
| `--help` | Show installer options without installing anything. |

## Shell setup

The release installer does not create an activation file or edit shell startup files. The PATH export above activates the current Bash or Zsh terminal. Add it to your chosen startup file for new terminals; other shells need their own PATH syntax. With `--prefix`, replace `$HOME/.local/bin` in the export with the selected directory.

Export custom `CARGO_HOME` and `RUSTUP_HOME` values before installation and retain those exports in later sessions. Keep the rustup Cargo proxy on PATH; standalone Cargo does not support L's pinned-toolchain invocation.

## Unattended installation

Use the same pinned download and install commands with `--with-toolchain`. No `--yes` option is needed. Native system tools must already be present. The `&&` chain stops after a failed download or installation instead of activating an older binary.

## Upgrades and repair

Read the [compatibility policy](../releases/index.md#versioning-and-compatibility), select a published release, and download its matching installer and `--version` as above. Rerunning the pinned 0.0.1 command reinstalls that version. To repair missing Rust tools or cached dependencies, rerun the selected installer with `--with-toolchain`, including after a binary-only installation.

Download verification and native prerequisite setup complete before the compiler binary is replaced. Failures at those stages preserve the existing compiler. Language resources stay embedded, and native setup preserves dependency locks. If installation succeeds but `l` is unavailable, check your PATH and selected executable directory.

## Uninstall

Remove only the `l` executable from the selected directory, normally `~/.local/bin/l`, and any L-specific shell configuration you added. Keep shared Rust tools, Cargo caches, and directories used by other projects.

If an earlier bootstrap created `~/.local/share/l/env` and profile blocks between `# >>> L environment >>>` and `# <<< L environment <<<`, remove those L-owned entries too. Do not delete entire shell startup files or shared executable/toolchain directories.

## Manual archives

Download your platform's archive and `checksums.txt` from the same [GitHub release](https://github.com/l-programming-language/l/releases). Verify its SHA-256 checksum before extraction. Keep the setup files together and run `bash setup-release.sh --with-toolchain` from the extracted directory. Put `l` on PATH alongside the rustup Cargo proxy as described in [shell setup](#shell-setup).

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

The tag selects released compiler source; use another release tag when needed. [install.sh](../../install.sh) verifies Go dependencies and builds the self-contained `bin/l` with `-mod=readonly`. It does not prepare Rust. Persist this checkout's absolute `bin` path in your shell configuration.

## Rust dependency preparation

L invokes `cargo +1.98.1` through rustup. Release setup installs the minimal pinned toolchain, fetches locked dependencies, and verifies offline resolution without changing the lockfile. On a fresh rustup installation, Rust 1.98.1 becomes the default toolchain.

For binary installations, rerun the versioned installer or extracted archive's setup script with `--with-toolchain`. For source installations, run `CGO_ENABLED=1 scripts/l-ci bootstrap` from the matching checkout. This also prepares contributor tools and validation caches and can take several minutes. See the [development policy](../development-harness.md#validation-authorization).

Installing Rust alone does not populate L's dependency cache. Native compilation uses frozen, offline Cargo resolution. Preserve the retained locks rather than removing `--frozen`. See [troubleshooting](troubleshooting.md#cargo-and-publication) if a build still fails after setup.
