# Install L

L is available for Apple Silicon macOS (ARM64) and Linux (ARM64 and x86-64). The released installer prepares L and its Rust build dependencies without Go or a compiler source checkout.

## Install a released binary

Install L 0.0.1 in a Bash or Zsh terminal:

```sh
curl --fail --location --proto '=https' --tlsv1.2 \
  https://github.com/l-programming-language/l/releases/download/v0.0.1/install-release.sh \
  --output install-release.sh &&
bash ./install-release.sh --version 0.0.1 --with-toolchain &&
export PATH="$HOME/.local/bin:${CARGO_HOME:-$HOME/.cargo}/bin:$PATH" &&
l --version
```

The installer verifies the release archive and installs `l` into `~/.local/bin`. `--with-toolchain` permits installation of missing Rust tools without another prompt. The PATH export activates this terminal; `l --version` should print `0.0.1`. Setup does not use `sudo`, install system packages, or edit shell startup files.

To review the script before running it, use [download and inspect](installation-options.md#download-and-inspect).

## Prerequisites

You need internet access and your system's [native build tools](installation-options.md#native-build-tools). If setup reports a missing C compiler or linker, follow those platform instructions and rerun the command.

<a id="activate-and-check-l"></a>

## Keep L on PATH

Add this export to your chosen Bash or Zsh startup file for new terminals:

```sh
export PATH="$HOME/.local/bin:${CARGO_HOME:-$HOME/.cargo}/bin:$PATH"
```

See [shell setup](installation-options.md#shell-setup) for custom locations and other shells. If macOS blocks the binary, see [macOS approval](installation-options.md#macos-approval).

## First native build

Next, [build your first program](../guide/first-program.md). The walkthrough takes you from a source file to a running program and confirms that native compilation works. Running the built program requires neither L nor Go/Rust tools.

## More installation options

See [installation options](installation-options.md) for version selection, custom locations, unattended setup, upgrades, and removal.

<a id="source-build-installation"></a>

To compile the L compiler yourself, follow [source-build installation](installation-options.md#source-build-installation).

<a id="rust-dependency-preparation"></a>

For missing Rust tools or cached dependencies, see [dependency preparation](installation-options.md#rust-dependency-preparation).
