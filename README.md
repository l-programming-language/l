<img src="website/public/assets/l/brand/logo-tile.svg" alt="L logo" width="112" height="112">

# L

L is a human-readable programming language with applications as values. It
compiles to native programs for macOS and Linux; Apps run in separate OS processes.

**Early development:** binary releases are not available yet, and publication
remains disabled. Backward compatibility is not guaranteed before `v1.0.0`.

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

Until binary releases are available, build the compiler from source using
Go **1.27.3**, pinned in [.go-version](.go-version). Follow the
[installation guide](docs/cli/installation.md) to clone the repository and prepare
the native build prerequisites: C/linker tools, rustup with Rust **1.98.1**, and
the locked Rust dependency cache. Initial setup requires internet access.

`install.sh` requires Bash and builds only the L compiler; it does not prepare
Rust or its dependencies. After completing prerequisite setup, run from the
repository root:

```sh
bash ./install.sh
./bin/l build ./demo/basic/main.l
./demo/basic/main
```

Output:

```text
Hello World!
```

Building and running are separate steps. Keep `main` and its `main.apps/`
directory together when moving the program. Running the generated program
requires neither L nor the Go/Rust toolchains.

The compiler embeds its language resources, so `bin/l` can be moved independently
or placed on your `PATH`. See the [CLI guide](docs/cli/build.md) for build options
and output locations.

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
