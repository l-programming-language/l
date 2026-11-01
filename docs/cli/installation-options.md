# Installation options

Start with [Install L](installation.md) for the recommended setup. This page covers platform requirements, custom locations, unattended installation, upgrades, and building the compiler from source.

## Supported platforms

Published binaries support Apple Silicon macOS (ARM64) and Linux (ARM64 and x86-64). Intel macOS is not supported; contributions toward future support are welcome. Native builds target the current host.

## Native build tools

L needs a C compiler and linker to build native programs. Install your system's tools if setup reports that they are missing:

- **macOS:** use Xcode Command Line Tools. Run `xcode-select --install` and finish Apple's installation prompt.
- **Linux:** install your distribution's C compiler and linker development packages.

Initial setup requires internet access. L setup prepares Rust tools and dependencies, but it does not use `sudo` or invoke your system package manager.

## Download and inspect

The [quick-install command](installation.md#install-a-released-binary) downloads a small setup script. To read it before running it:

```sh
curl --proto '=https' --tlsv1.2 -fsSL https://ldocs.org/install.sh --output install-l.sh
less install-l.sh
```

After reviewing it, run:

```sh
sh ./install-l.sh && . "$HOME/.local/share/l/env"
l --version
```

The script resolves the latest stable release once and uses that release's versioned installer and assets. Setup prepares Rust build prerequisites, including missing tools, after one confirmation in your terminal. It does not build the L compiler from source.

## Installer options

Pass these options to `sh ./install-l.sh`, or to `sh -s --` when using the quick-install command:

| Option | Effect |
| --- | --- |
| `--version VERSION` | Install a particular release instead of the latest stable release. |
| `--prefix DIRECTORY` | Choose the executable directory instead of `~/.local/bin`. |
| `--no-modify-path` | Leave shell startup files unchanged. The activation file is still written. |
| `--yes` | Accept the proposed installation without an interactive prompt. Required when there is no terminal. |
| `--help` | Show setup options without installing anything. |

For example, install a specific version without editing shell startup files:

```sh
sh ./install-l.sh --version 0.0.1 --no-modify-path && . "$HOME/.local/share/l/env"
```

The quick installer always prepares Rust and locked dependencies. For binary-only installation or a different release repository, use the [versioned release installer](#download-and-inspect-the-release-installer).

## Shell setup

Setup writes `~/.local/share/l/env`, which adds the selected L executable directory and rustup's Cargo directory to PATH. Activate it in the current Bash or Zsh terminal with:

```sh
. "$HOME/.local/share/l/env"
```

By default, setup adds a marked block to Bash's `~/.bashrc` and its first existing login file (`~/.bash_profile`, `~/.bash_login`, or `~/.profile`). If none of those login files exists, it creates `~/.profile`. It also configures Zsh's `.zshrc` and `.zprofile` under the exported `ZDOTDIR`, or under `HOME` when `ZDOTDIR` is unset. Existing valid profile symlinks are preserved. Repeating setup updates the L block rather than adding another copy.

Use `--no-modify-path` to manage shell integration yourself. You can source the activation file manually or add the command above to your chosen startup file. Other shells need their own PATH syntax.

Export a custom `CARGO_HOME`, `RUSTUP_HOME`, or `ZDOTDIR` before running setup. The activation file remembers the Cargo and Rustup locations selected during installation while preserving values you explicitly export later. With `--prefix`, it also uses your chosen executable directory.

If you use `ZDOTDIR`, make sure new Zsh sessions receive that setting too. Setup writes profiles in that directory but does not change how Zsh locates it.

## Unattended installation

In Bash or Zsh, use an explicit version and accept setup with `--yes`. Use `--no-modify-path` when the environment manages its own shell configuration:

```sh
(set -o pipefail; curl --proto '=https' --tlsv1.2 -fsSL https://ldocs.org/install.sh | sh -s -- --yes --no-modify-path --version 0.0.1) && . "$HOME/.local/share/l/env"
l --version
```

Native system tools must already be present. Rust tools and locked dependencies are prepared in the selected user locations. A failed download makes the command fail instead of activating an older installation.

## Upgrades and repair

Rerun the [quick-install command](installation.md#install-a-released-binary) to install the latest stable release, or pass `--version` to select a particular release. Read the [compatibility policy](../releases/index.md#versioning-and-compatibility) first. Repeating setup also repairs missing Rust tools and cached dependencies.

The release installer replaces the compiler binary only after download verification and native prerequisite setup succeed. Those failures preserve the existing compiler. Language resources are embedded in the binary; native setup preserves the retained dependency locks.

If terminal configuration fails after the binary is installed, setup reports that partial result. It attempts to restore its shell changes while preserving concurrent edits; the installed L binary and prepared Rust tools remain. Resolve the reported file or permission problem, then rerun setup to repair activation.

Successful setup saves its installation log as `~/.local/share/l/install.*`. If a failure occurs after release installation starts, setup prints the temporary directory containing `install.log` and configuration backups. Keep those files when reporting the failure.

## Uninstall

Remove only the `l` executable from the directory selected during installation, normally `~/.local/bin/l`. Remove the L-owned blocks between `# >>> L environment >>>` and `# <<< L environment <<<` from the shell startup files configured above, then remove `~/.local/share/l/env`. If you configured PATH manually, remove only the L-specific entry you added.

Keep shared Rust tools, Cargo caches, and other files used by other projects. Do not delete `~/.local/bin`, Cargo or Rustup directories, or entire shell startup files to remove L.

## Download and inspect the release installer

Each [GitHub release](https://github.com/l-programming-language/l/releases) retains its own installer. For L 0.0.1, download the versioned script:

```sh
curl --fail --location --proto '=https' --tlsv1.2 \
  https://github.com/l-programming-language/l/releases/download/v0.0.1/install-release.sh \
  --output install-release.sh
```

Inspect the script, then install that version and its build prerequisites:

```sh
bash ./install-release.sh --version 0.0.1 --with-toolchain
export PATH="$HOME/.local/bin:${CARGO_HOME:-$HOME/.cargo}/bin:$PATH"
l --version
```

This release installer does not create the quick installer's activation file or modify shell startup files. Add the PATH export to your chosen startup file to make this installation available in new terminals. Its flags differ from those of `install-l.sh`:

| Option | Effect |
| --- | --- |
| `--with-toolchain` | Permit installing missing Rust tools. Otherwise interactive setup asks; noninteractive setup requires them to exist. |
| `--binary-only` | Install only `l`, deferring native prerequisites. |
| `--prefix DIRECTORY` | Choose the executable directory instead of `~/.local/bin`. |
| `--version VERSION` | Pin a release; omission selects the latest normal release once for all assets. |
| `--repository OWNER/REPO` | Intentionally select another repository's release. |

For a different pinned version, change both the release URL and `--version`. To repair a binary-only installation, rerun its versioned installer with `--with-toolchain`.

## Manual archives

Download the archive for your platform and `checksums.txt` from the same [GitHub release](https://github.com/l-programming-language/l/releases). Verify its SHA-256 checksum before extracting it. Keep the setup files together, then run `bash setup-release.sh --with-toolchain` from the extracted directory. Put `l` in a directory on your PATH and keep the rustup Cargo proxy available using the PATH export above.

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

For binary installations, rerun quick setup or use the versioned installer or extracted archive's setup script as described above. For source installations, the current preparation path is `CGO_ENABLED=1 scripts/l-ci bootstrap` from the matching checkout. This also prepares contributor tools and validation caches and can take several minutes; see the [development policy](../development-harness.md#temporary-validation-pause).

Installing Rust alone does not populate L's dependency cache. Native compilation uses frozen, offline Cargo resolution; preserve the retained locks rather than removing `--frozen`. See [troubleshooting](troubleshooting.md#cargo-and-publication) if a build still fails after setup.
