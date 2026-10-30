# L Language Specification

L expresses values, program structure and behavior through grammatical sentences. This specification describes the current language and its Rust target; start with the [first program](../guide/first-program.md) to see them together.

## Scope and conformance

The maintained repository contracts govern source bytes, ownership, declaration identity and execution. Lyger translates L into AML/logeme compositions; Logiceme generates Rust source; Cargo builds the native program. Running that program starts the orchestrator, which controls separate App processes. Generating source alone does not run it.

## Reading the specification

The type reference answers two questions: **what does this value represent?** and **what role does this word play in a sentence?**

| Family | What it describes | Examples |
| --- | --- | --- |
| [Primitive Types](values.md) | Individual data values and managed references. | Text, Number, Truth, Thing, Bytes, Pointer, Null |
| [Iterative Types](collections-and-categories.md) | Collections and named-member namespaces. | List, Map, Dictionary |
| [Lexical Types](lexical-types.md) | Grammatical roles of words. | Nouns, Verbs, Adjectives, Adverbs, Articles, Prepositions |
| [Applicational Types](applicational-types.md) | Application structure and behavior. | App, Service, Class, Object, Function |

These are related dimensions: `App` is a noun grammatically and names an applicational type. Each page starts with the complete category table, including names with limited support.

Read [lexical structure](lexical-structure.md) for literals and paragraphs, [names and scope](names-and-scope.md) for binding and visibility, and [grammar](grammar.md) for sentence forms. [Definitions](definitions.md), [control flow](control-flow.md), [imports](imports.md), and [execution](execution.md) explain how those pieces become a program. Consult [errors](errors.md) and the [glossary](glossary.md) for diagnostics and terminology.

The [builtin reference](../reference/builtins/index.md) documents implicit words; the [standard library](../stdlib/index.md) documents explicit imports. Native definitions use the separate [logeme model](../../builtins/LOGEMES.md) and [Rust target contracts](../../builtins/README.md).

## Role, support, and evidence

Tables distinguish executable behavior, recognized but incomplete forms, and internal representations. The [evidence guide](../contributing/evidence.md) explains example labels and verification stages.

Code blocks marked `l` contain L; `rust` contains native fragments; `text` may contain grammar notation or AML. Native Rust syntax inside a definition does not become ordinary L syntax.

## Language guarantees and target limits

Source ownership, binding and raw captured output follow the language/runtime contracts. Cargo layouts, Rust encoding and native library visibility describe the current target. Native macOS/Linux execution is supported; cross-compilation, in-code concurrency, general async source functions and automatic Deferred execution are outside this release. L manages ownership for its typed runtime values; opaque native definitions retain explicit Rust lifetime and generic contracts. Releases embed English resources and follow the [pre-1.0 compatibility policy](../releases/index.md#versioning-and-compatibility).

## Sources and review

See [architecture](../architecture.md), [CLI/runtime semantics](../cli-build.md), [validation policy](../development-harness.md), and the [error model](../error-model.md). [Semantic mappings](../_meta/semantics.json) and the [specification audit](../_meta/specification-audit.md) record source evidence and unresolved support boundaries.
