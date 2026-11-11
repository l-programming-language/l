<img src="website/public/assets/l/brand/logo-tile.svg" alt="L logo" width="112" height="112">

> **Development reset:** This checkout currently supports a standalone App with core Print. Standard libraries and Runtime are removed pending reconstruction in L. Run `scripts/l-check` for the temporary hello-world gate. Broader examples and documentation describe the prior implementation and are outside this gate. Publication is disabled.

Build and run the current hello-world example with the checkout's compiler:

```sh
scripts/l-go run ./cmd/l build --output-dir .harness/hello demo/basic/main.l
.harness/hello/main
```

It prints `Hello World!` followed by a newline. Source-level process supervision is unavailable until Runtime is rebuilt.


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

[Install L](docs/cli/installation.md) for Apple Silicon macOS (ARM64) or Linux
(ARM64 and x86-64). The versioned release installer prepares Rust dependencies;
native C/linker tools are required. Binary installation needs neither Go nor a
source checkout. The guide covers shell setup and unsigned pre-1.0 macOS binaries.

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
[validation policy](docs/development-harness.md#validation-authorization).

Maintainers can find publication prerequisites and recovery steps in the
[release activation guide](website/DEPLOYMENT.md).

## License

L is licensed under [Apache-2.0](LICENSE).
