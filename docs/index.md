# L documentation

L is a human-readable language with applications as values. It generates Rust source and uses the Rust toolchain to build native artifacts for your current macOS or Linux host. Building and running are separate steps: run the generated executable to start your program.

This site documents the current repository checkout. It presents an **Unreleased / Local** documentation view when no authoritative release identity is available. New specification prose is source-backed documentation, not a claim of independent standards approval or maintainer ratification.

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

Document role, implementation support, and verification are separate. A guide explains a workflow; a specification chapter states maintained contracts and clearly delimited observations; an implementation reference explains internal mechanisms. A grammar entry may be recognized without having an executable implementation.

An existing test is a witness to the intended assertion. It is not a result from this documentation build. Source generation, native compilation, runtime execution and full acceptance establish different things. The [evidence guide](contributing/evidence.md) explains these distinctions; the [integration ledger](../testdata/integration/IMPLEMENTATION.md) records revision-specific historical runs.

## Maintained technical guides

The four existing repository guides remain at their canonical paths: [architecture](architecture.md), [CLI and runtime semantics](cli-build.md), [development and acceptance](development-harness.md), and [compiler errors](error-model.md). The website renders these sources instead of maintaining independent copies.

Website CI/CD is prepared but disabled. No deployment has been performed.
