# Learn L

L uses a configured grammar and an extensible vocabulary. Sentences describe values, definitions, application structure and ordered actions. This is a precise programming language: ordinary English that resembles an example is not automatically valid L.

The learning path starts with a complete maintained program and adds one concept at a time. Examples are included from repository files. Their expected results come from documented contracts and existing tests; the website build does not execute them.

## A practical learning path

1. [Build your first program](first-program.md): import Print, construct an App, execute it and observe raw output.
2. [Work with values](values.md): Text, Number and Truth; assignment and output boundaries.
3. [Understand names and paragraphs](names-and-paragraphs.md): scope, pronouns, public declarations and continuation activation.
4. [Define vocabulary and behavior](definitions.md): typed action arguments, inflections, descriptors and Functions.
5. [Make decisions and sequence work](control-flow.md): conditions, source ordering, Exit and iteration limits.
6. [Import dictionaries and source files](imports.md): explicit IO, project members and source-relative ownership.
7. [Organize Apps and Services](apps-and-services.md): application values, App-constructed libraries and source organization.
8. [Manage lifecycle and output](lifecycle.md): Run, Stop, Await and invocation results.
9. [Write native-backed definitions](native-definitions.md): the advanced boundary between L, AML and Rust templates.

## What you need

The CLI generates Rust source and builds for your current macOS or Linux host. Follow the [installation prerequisites](../cli/installation.md#prerequisites), including the pinned Rust/Cargo toolchain and dependencies, then return to [the first program](first-program.md). A website-only contributor needs the independent Node package, not the native toolchain.

The source language bundled here is English (`eng`). The site's English interface is a separate concern. A language-selection CLI flag does not establish that another source lexicon is available or fully localized.

## How examples are presented

Complete sources carry a canonical source path and an example classification. Runnable examples have an expected byte stream; compile-only examples stop before execution; intentionally invalid examples illustrate a specific failure; fragments explain a smaller mechanism. A fragment must not be pasted into the CLI as if it were a complete program.

`Print` adds a newline. `Show` does not. Output blocks in prose are for reading; the [example manifest](../_meta/examples.json) stores exact bytes separately, so a trailing newline is never inferred from how a browser renders a block.

For the full supported surface and precise boundaries, use the [specification](../spec/index.md). For implementation detail, use the [architecture](../architecture.md) and [logeme reference](../reference/logemes/index.md).
