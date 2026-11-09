# Primitive Types

| Type | Meaning | Current use |
| --- | --- | --- |
| [Text](#text) | A sequence of bytes, including whitespace. | Write a backtick-delimited literal, such as `` `hello` ``, then store, compare, or print it. |
| [Number](#number) | A finite binary64 number, such as `7` or `-12.5`. | Store, compare, or print it. Integer spellings use the same number type. |
| [Truth](#truth) | A Boolean value: `true` or `false`. | Store, compare, print, or use it as a condition. |
| [Thing](#thing) | A value retaining its concrete runtime type. | Hold different value categories at an explicitly broad boundary. |
| [Byte / Bytes](#byte-and-bytes) | An owned sequence of bytes. | Construct from Text and read numbered bytes. |
| [Pointer](#pointer) | A managed reference to a typed value. | Copy handles to share a target with checked reads and writes. |
| [Null](#null-and-internal-states) | An explicit absence value. | Initialize a value with `Null`, distinct from an unassigned slot. |

Primitive types describe data, while [lexical types](lexical-types.md) describe words. [Collections](collections-and-categories.md) and [Functions](applicational-types.md#functions-properties-and-operations) use the same ownership rules.

## Declarations, assignment, and reading values

A declaration names a value: `Define "count" as 7.` Runtime assignments take effect in source order, and every continuing path must assign a value before reading it. Declaring a type does not promise an implicit initial value.

The public-values program gives Text, Number, and Truth explicit initial values, then displays them with Show:

<!-- example:public-values -->

Assignment, Function arguments/results and captures snapshot ordinary values, recursively through collections. Embedded Pointers retain their shared target. App handles retain process identity. Stored Output keeps its invocation's bytes after reruns. See [capture environments](names-and-scope.md#runtime-storage-and-definite-assignment).

## Text

[Text in the Word Glossary](../reference/word-glossary.md#word-type-text).

## Number

[Number in the Word Glossary](../reference/word-glossary.md#word-type-number).

## Truth

[Truth in the Word Glossary](../reference/word-glossary.md#word-type-truth).

## Thing

[Thing in the Word Glossary](../reference/word-glossary.md#word-type-thing).

## Byte and Bytes

### Byte

[Byte in the Word Glossary](../reference/word-glossary.md#word-type-byte).

### Bytes

[Bytes in the Word Glossary](../reference/word-glossary.md#word-type-bytes).

## Pointer

[Pointer in the Word Glossary](../reference/word-glossary.md#word-type-pointer).

## Null and internal states

### Null

[Null in the Word Glossary](../reference/word-glossary.md#word-type-null).

## Error

[Error in the Word Glossary](../reference/word-glossary.md#word-type-error).

## Conversion and aliasing boundaries

Matching Text, Number, and Truth values can be compared, but ordering is Number-only. No implicit cross-scalar coercion is promised. Definition slots separately check their required argument kinds.

The repeated-activation demo contrasts an App alias with a Text snapshot:

<!-- example:values-output-snapshot -->

`Worker` and `Repeat` refer to the same App. After the first join, `snapshot` holds that invocation's Output. Running the App again does not update `snapshot`: `Show snapshot` reads the stored scalar, while `Show Repeat's Output` reads the later invocation. Both invocations happen to print the same message. This excerpt depends on the demo's earlier declarations.

## Sources and witnesses

Primary sources: [value tags](../../lyger/value_kinds.go), [primitive storage and loading](../../lyger/types.go), [type spellings](../../lyger/utilities.go), [native encoding](../../lyger/literal_atoms.go), [definition encoding](../../lyger/definition_bindings.go), and [typed orchestration](../../lyger/orchestration_plan.go). Existing witnesses: [Number encoding](../../lyger/definition_number_test.go), [numeric resolution](../../lyger/numeric_resolution_test.go), [raw Text](../../lyger/raw_text_test.go), [Text encoding](../../lyger/definition_text_encoding_test.go), [IO numeric formatting](../../lyger/io_numeric_test.go), [stored-value invariants](../../lyger/stored_values_test.go), and [scalar conditions](../../lyger/orchestrator_test.go). Example labels distinguish complete programs from fragments. Website validation does not execute L.
