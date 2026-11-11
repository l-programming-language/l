# Applicational Types

These categories describe source and generated-code organization. The supported executable is one standalone App with a core Print body.

| Category | Meaning |
| --- | --- |
| App | Application source selected for standalone native compilation |
| Service | Source structure; core Main supplies the App's entry composition |
| Class | Group of generated modules |
| Object | Source-file owner |
| Function | Typed callable composition |
| Attribute, Definition | Compiler construction and implementation records |
| Property, Operation | Native representation records |

Internal categories do not automatically supply source constructors or a runtime.

## Apps

[App](../reference/word-glossary.md#word-type-app) is the standalone compilation target.

### Create and execute an App

Follow the [first program](../guide/first-program.md). Build it with `l build`, then invoke the resulting executable. Do not add source-level lifecycle actions.

## Services

[Service](../reference/word-glossary.md#word-type-service) participates in source organization. The current CLI rejects additional Service libraries; a retained Service declaration does not create a process.

## Classes and Objects

### Class

[Class](../reference/word-glossary.md#word-type-class) groups source modules without introducing a running lifetime.

### Object

[Object](../reference/word-glossary.md#word-type-object) owns file content and its metadata.

### Compose an Object inside nested Classes

Preserve canonical ownership through construction and native projection. Classes cannot inject defaults into file-owned Object metadata. The current tutorial does not claim arbitrary nested compositions are covered by the temporary gate.

## Attributes and Definitions

### Attribute

[Attribute](../reference/word-glossary.md#word-type-attribute) is a retained construction category.

### Definition

[Definition](../reference/word-glossary.md#word-type-definition) records a selected implementation. Its identity is distinct from its word's grammatical role.

## Functions, Properties, and Operations

Core [Function](../reference/word-glossary.md#word-builtin-function) supplies callable metadata and a Process body. Preserve ordered typed inputs/results, canonical invocation identity and once-only evaluation. Return belongs to its selected Function and is distinct from whole-program Exit.

Native Func/Prop/Op records follow the [logeme contracts](../../builtins/README.md). They do not supply source-level capture, process transport or automatic lifetime inference.

### Construct a Function in an Object

Core Main composes a Function inside an Object inside a Service, while Basic composes the App. Its Function has no inputs or outputs and receives the selected Process body. See [core](../../builtins/eng/core.l) and [definitions](definitions.md).

## Sources and witnesses

See [application structures](../../lyger/application_types.go), [construction](../../lyger/construction.go), [workspace projection](../../lyger/workspace_units.go) and the [temporary hello-world gate](../development-harness.md).
