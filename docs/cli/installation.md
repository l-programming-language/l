# Install L

L is available for Apple Silicon macOS (ARM64) and Linux (ARM64 and x86-64). Setup installs the latest stable release and prepares the Rust tools needed to build L programs. You do not need Go or a compiler source checkout.

## Install a released binary

Run this command in a Bash or Zsh terminal:

```sh
(set -o pipefail; curl --proto '=https' --tlsv1.2 -fsSL https://ldocs.org/install.sh | sh) && . "$HOME/.local/share/l/env"
```

Setup shows the selected release and installation locations, then asks once before making changes. It verifies the download, installs `l` into `~/.local/bin`, prepares its pinned Rust dependencies, and configures Bash and Zsh for new terminals. The final part of the command activates L in your current terminal. Setup does not use `sudo` or install system packages.

To review the script before running it, use [download and inspect](installation-options.md#download-and-inspect) or the [versioned release installer](installation-options.md#download-and-inspect-the-release-installer).

## Prerequisites

You need internet access and your system's [native build tools](installation-options.md#native-build-tools). If setup reports a missing C compiler or linker, follow those platform instructions and rerun the command.

## Activate and check L

Check that L is available:

```sh
l --version
```

It should print the version selected during setup. If you need to activate an existing installation in another Bash or Zsh session, run `. "$HOME/.local/share/l/env"`. See [shell setup](installation-options.md#shell-setup) for custom locations and other shells. If macOS blocks the binary, see [macOS approval](installation-options.md#macos-approval).

## First native build

Next, [build your first program](../guide/first-program.md). The walkthrough takes you from a source file to a running program and confirms that native compilation works. Running the built program requires neither L nor Go/Rust tools.

## More installation options

See [installation options](installation-options.md) for version selection, custom locations, unattended setup, upgrades, and removal.

<a id="source-build-installation"></a>

To compile the L compiler yourself, follow [source-build installation](installation-options.md#source-build-installation).

<a id="rust-dependency-preparation"></a>

For missing Rust tools or cached dependencies, see [dependency preparation](installation-options.md#rust-dependency-preparation).
