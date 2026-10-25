# L Language Specification

L is a human-readable, extensible language with application values. This specification describes the language implemented by the current checkout and the boundary of its Rust target. It is a source-backed account, not a claim of independent maintainer ratification or standards approval.

## Scope and conformance

The maintained repository contracts take precedence over this explanatory consolidation. A conforming description preserves source bytes, source ownership, declaration identity, and the distinction between generation and execution. Where implementation and intent disagree, this specification identifies the boundary instead of declaring the discrepancy a new language guarantee.

L source is processed by Lyger into AML and logeme compositions. Logiceme interprets those compositions into Rust source bundles. A separate native build coordinator invokes Cargo and publishes the executable and companion artifacts. Running that executable starts the orchestrator; source lifecycle instructions control separate App processes. Merely parsing source, producing AML, or returning an `Executable` source bundle does not execute an application.

The implementation tools are Go programs. Rust is the generated target. English (`eng`) is the only bundled natural-language lexicon. The website interface language, source lexicon, and native output target are three different choices; an English website does not constrain a future source lexicon, and a `--language` option does not establish another implemented lexicon.

## Reading the specification

The chapters form a continuous description:

1. [Lexical structure and source organization](lexical-structure.md): bytes, names, literals, statements, and paragraphs.
2. [Primitive values](values.md): Text, Number, Truth, and the bounded status of other value categories.
3. [Collections and application categories](collections-and-categories.md): Lists, Dictionaries, Apps, Services, source groups, and callable representations.
4. [Names, scope, and state](names-and-scope.md): declarations, pronouns, aliases, visibility, and definite assignment.
5. [Grammar and vocabulary](grammar.md): the exact configuration notation, grammatical families, and support boundaries.
6. [Definitions and extensibility](definitions.md): actions, descriptors, inflections, typed slots, and native definitions.
7. [Operations and control flow](control-flow.md): sequencing, comparison, selection, termination, and recognized unsupported forms.
8. [Imports and program structure](imports.md): dictionaries, explicit source files, ownership, and dependency boundaries.
9. [Execution and lifecycle](execution.md): Build, Run, Stop, Await, continuations, and output.
10. [Errors and compatibility](errors.md): failure stages and diagnostic boundaries.
11. [Glossary](glossary.md): the language and compiler's own terms.

Use the [builtin reference](../reference/builtins/index.md) for implicit definitions, the [standard library](../stdlib/index.md) for explicitly imported dictionaries, and the maintained [architecture](../architecture.md) for implementation structure. The [logeme model](../../builtins/LOGEMES.md) and [Rust target reference](../../builtins/README.md) describe composition below the source language.

## Role, support, and evidence

Three independent labels are used by the coverage manifest. A role describes what a document is for: normative contract, explanatory guide, implementation detail, or future design. Support describes a feature: supported within a stated boundary, partial, rejected, design-only, internal, or unresolved. Evidence describes the observation: static source inspection, an existing test witness, or a verification actually executed at a named stage.

A grammar rule establishes a recognizable form, not working native behavior. A test file is a witness of an assertion the repository maintains, not a report that the test passed today. Generation, native compilation, runtime observation, and full acceptance are separate stages. This documentation pass uses static inspection and existing witnesses; language execution was not performed under the local validation pause.

Code blocks marked `l` contain L source; `rust` blocks are target-specific native fragments; `text` blocks may show grammar notation or AML. Explanatory fragments are not implicitly executable programs. The complete basic example below is included from the maintained repository fixture; its expected behavior is an existing test contract rather than a new runtime observation.

Read each example's classification and explanation together. A **runnable** example may deliberately exit with an error; **compile-only** supplies no runtime claim; **invalid** demonstrates an expected compiler rejection. A **fragment** may omit surrounding declarations or illustrate a form reviewed only through static source inspection. Unsupported source forms are labeled at the example itself, and internal-only categories are not given invented construction syntax.

<!-- example:basic -->

## Language guarantees and target limits

Source paragraph identity, binding, explicit import ownership, and raw captured output are maintained language/runtime contracts. Cargo layouts, Rust `f64` encoding, Tokio supervision, and native library visibility are implementation details that explain the supported target. This distinction matters when reading native-backed definitions: Rust syntax accepted inside a deliberately native fragment is not automatically L syntax.

Native macOS and Linux execution is supported by the maintained contract. Cross-compilation, Phase 2 in-code concurrency, inferred ownership/lifetimes, general asynchronous source functions, and automatic Deferred execution are outside this release boundary. The current site describes an unreleased/local checkout unless authoritative release identity is supplied.

## Sources and review

Primary sources are [architecture](../architecture.md), [CLI/runtime semantics](../cli-build.md), [validation policy](../development-harness.md), [error model](../error-model.md), [Lyger](../../lyger/lyger_service.go), and the [integration evidence ledger](../../testdata/integration/IMPLEMENTATION.md). The machine-readable review record is [semantics.json](../_meta/semantics.json); the [specification audit](../_meta/specification-audit.md) records unresolved discrepancies. These maintenance files are source artifacts, not language APIs.
