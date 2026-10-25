# Define vocabulary and behavior

L's extensibility is part of ordinary source compilation. A definition can introduce a word, its grammatical forms, typed arguments and an ordered composition. A descriptor can describe how a value is constructed. These mechanisms are not interchangeable with an unrestricted native macro system.

## Compose a source action

<!-- example:local-definitions -->

`Launch` is introduced with `Create ... for ...`. Its argument is `any App`; the later `it` refers to the bound App. Building, Running and Awaiting retain their process-operation meanings. The action expands its whole body for one operand before processing a coordinated next operand.

The label `Launch|-es|-ed|-ing` declares related forms. Inflection suffixes belong to the language's word-declaration protocol; they are not a regular expression to be interpreted by a docs generator. The grammar uses the resulting lexical roles to recognize forms such as a noun, verb or participle.

Each paragraph owns an independent Launch definition and Compact descriptor. Reusing these names in sibling paragraphs does not overwrite the enclosing program. The exact expected output is `First local\nSecond local\ndone`, witnessed by the [paragraph test table](../../cmd/l/paragraph_integration_test.go).

## Describe construction

`"Compact" describes a Generic App with a New Main Service using any Process` is a descriptor. Its composition resembles the implicit Basic descriptor. When an App is constructed as New Compact, the compiler retains the descriptor's typed targets and source ownership while expanding it.

Do not replace a descriptor target with a nearby type merely because both have similar implementation structures. Source ownership, declared names, typed parameters and order are part of the observable expansion contract.

## Know the implicit foundation

The [canonical core](../../builtins/eng/core.l) defines Start, Execute, Process, Function, Return, Objective, Main and Basic. Start and Execute are source actions accepting App arguments; they are not separate primitive runtime opcodes and do not become generic Service actions.

Process wraps a sequence of syllables. Function provides a native-backed callable definition with typed inputs and outputs. Return is a definition belonging to Function. Objective and Main describe the structural entry composition used by Basic Apps. These definitions do not make standalone Services executable.

The presence of a backend callable form does not establish ordinary L closures, ownership inference, language-level generics or async functions. [Builtins](../reference/builtins/index.md) explains the public foundation; the [logeme reference](../reference/logemes/index.md) separates native template facilities from source-language features.

## Diagnose a definition failure

An unresolved typed argument is a binding problem, not permission to replace it with a String. Expansion cycles are rejected. A malformed framed native body is distinct from a well-framed but unsupported Rust template composition. Keep the source span and the original error while investigating.

Before writing native fragments, complete [control flow](control-flow.md), [imports](imports.md), and [lifecycle](lifecycle.md). The advanced [native-definition guide](native-definitions.md) then introduces the backend boundary.
