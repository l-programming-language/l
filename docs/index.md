# L documentation

L is a human-readable language with applications as values. It generates Rust source and uses the Rust toolchain to build native artifacts for your current macOS or Linux host. Building and running are separate steps: run the generated executable to start your program.

Install L from [GitHub Releases](https://github.com/l-programming-language/l/releases/latest). Read the [versioning policy](releases/index.md#versioning-and-compatibility) before upgrading: backward compatibility is not guaranteed before `v1.0.0`. Example labels distinguish supported behavior from fragments and expected failures.

<!-- illustration:source-to-native -->

## Choose a starting point

| Your task | Start here |
| --- | --- |
| Write and build a first program | [Learn L](guide/index.md) |
| Install the compiler or understand a build command | [CLI reference](cli/index.md) |
| Read the implemented language contract | [Language specification](spec/index.md) |
| Look up implicit facilities | [Builtins](reference/builtins/index.md) |
| Import a library dictionary | [Standard library](stdlib/index.md) |
| Understand a logeme composition | [Logeme reference](reference/logemes/index.md) |
| Understand source ownership, generation or processes | [Architecture](architecture.md) |
| Improve documentation or compiler development | [Contributing](contributing/index.md) |
| Understand version identity and release limits | [Release information](releases/index.md) |

## Read evidence with the claim

Example classifications describe their intended use, not a fresh test result. The [evidence guide](contributing/evidence.md) defines support and verification stages; the [integration ledger](../testdata/integration/IMPLEMENTATION.md) records results for specific revisions and hosts.

## Maintained technical guides

The four existing repository guides remain at their canonical paths: [architecture](architecture.md), [CLI and runtime semantics](cli-build.md), [development and acceptance](development-harness.md), and [compiler errors](error-model.md). The website renders these sources instead of maintaining independent copies.
