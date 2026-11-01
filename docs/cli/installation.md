# Install L

[L 0.0.1 is available](https://github.com/l-programming-language/l/releases/tag/v0.0.1) for Apple Silicon macOS (ARM64) and Linux (ARM64 and x86-64). The installer selects your platform and prepares the Rust tools needed to build L programs. You do not need Go or a compiler source checkout.

## Prerequisites

You need internet access and your system's [native build tools](installation-options.md#native-build-tools). If setup reports a missing C compiler or linker, follow those platform instructions and rerun the installer.

## Install a released binary

In a Bash or Zsh terminal, download the installer:

```sh
curl --fail --location --proto '=https' --tlsv1.2 \
  https://github.com/l-programming-language/l/releases/download/v0.0.1/install-release.sh \
  --output install-release.sh
```

Inspect the downloaded script, then install L and its build dependencies:

```sh
bash ./install-release.sh --version 0.0.1 --with-toolchain
```

The installer verifies the download and installs `l` into `~/.local/bin`. It does not use `sudo` or edit your shell configuration.

## Activate and check L

Make L and its build tools available in this terminal:

```sh
export PATH="$HOME/.local/bin:${CARGO_HOME:-$HOME/.cargo}/bin:$PATH"
l --version
```

The version should be `0.0.1`. Add the export line to your [shell startup file](installation-options.md#shell-setup) to keep L available in new terminals. If macOS blocks the binary, see [macOS approval](installation-options.md#macos-approval).

## First native build

Next, [build your first program](../guide/first-program.md). The walkthrough takes you from a source file to a running program and confirms that native compilation works. Running the built program requires neither L nor Go/Rust tools.

## More installation options

See [installation options](installation-options.md) for custom locations, version selection, upgrades, manual downloads, and supported platforms.

<a id="source-build-installation"></a>

To compile the L compiler yourself, follow [source-build installation](installation-options.md#source-build-installation).

<a id="rust-dependency-preparation"></a>

For missing Rust tools or cached dependencies, see [dependency preparation](installation-options.md#rust-dependency-preparation).
