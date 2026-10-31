<img src="website/public/assets/l/brand/logo-tile.svg" alt="L logo" width="112" height="112">

# L

L is a human-readable programming language with applications as values. It
compiles to native programs for macOS and Linux; Apps run in separate OS processes.

**Early development:** backward compatibility is not guaranteed before `v1.0.0`.
Releases are published manually; merging into main does not release L.

## A first program

The [basic demo](demo/basic/main.l) creates an App, runs it, and displays its output:

```l
Use IO's Print from `io` as "Print".

"message" is `Hello World!`.
Create "Hello" as a New Basic App.
The App Prints message.
Execute Hello.
Show Hello's Output.
```

See the [first-program walkthrough](docs/guide/first-program.md) for an explanation
of the syntax and execution steps.

## Try L

Check [GitHub Releases](https://github.com/l-programming-language/l/releases) for
Apple Silicon macOS (ARM64) and Linux (ARM64 and x86-64) binaries. Once `v0.0.1`
is published, download its installer:

```sh
curl --fail --location --proto '=https' --tlsv1.2 \
  https://github.com/l-programming-language/l/releases/download/v0.0.1/install-release.sh \
  --output install-release.sh
```

Inspect the script, then install L and its pinned Rust dependencies:

```sh
bash ./install-release.sh --version 0.0.1 --with-toolchain
export PATH="$HOME/.local/bin:$HOME/.cargo/bin:$PATH"
l --version
```

Keep that `PATH` in your shell configuration. Native C/linker tools are required;
pre-1.0 macOS binaries are unsigned. The [installation guide](docs/cli/installation.md)
covers prerequisites, macOS approval and building L from source when a binary
release is unavailable. Binary installation needs neither Go nor a source checkout.

Save the example as `main.l`, then build and run it:

```sh
l build main.l
./main
```

Output is `Hello World!` followed by a newline. Keep `main` and `main.apps/`
together when moving the program. Running the result requires neither L nor the
Go/Rust toolchains. See the [CLI guide](docs/cli/build.md) for build options.

## Learn more

- [Learn L](docs/guide/index.md): syntax, values, and Apps.
- [Language specification](docs/spec/index.md): the implemented language contract.
- [CLI reference](docs/cli/index.md): installation, commands, and runtime behavior.
- [Architecture](docs/architecture.md): compiler and process design.

## Contributing

Start with [CONTRIBUTING.md](CONTRIBUTING.md) and the
[development guide](docs/development-harness.md). Pull requests run Linux CI;
expensive local validation remains subject to the
[validation policy](docs/development-harness.md#temporary-validation-pause).

Maintainers can find publication prerequisites and recovery steps in the
[release activation guide](website/DEPLOYMENT.md).

## License

L is licensed under [Apache-2.0](LICENSE).
