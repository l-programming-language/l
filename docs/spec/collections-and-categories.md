# Iterative Types

| Type | Meaning | Operations |
| --- | --- | --- |
| [List](#lists) | Ordered owned values with an element type. | Construct, append, read and replace by zero-based index. |
| [Map](#maps) | Owned typed values keyed by Text bytes. | Insert, read and replace. Keys have deterministic byte order. |
| [Dictionary](#dictionaries) | A vocabulary namespace with its own owner. | Declare and import words. Use Map for runtime key/value data. |

Assignment, arguments, results and captures snapshot ordinary elements while retaining nested [Pointer](values.md#pointer) identity. These operations introduce no implicit traversal, membership or removal API.

## Lists

[List in the Word Glossary](../reference/word-glossary.md#word-type-list).

### Declare a List

[Declare a List in the Word Glossary](../reference/word-glossary.md#word-type-list--declare-a-list).

## Maps

[Map in the Word Glossary](../reference/word-glossary.md#word-type-map).

## Dictionaries

[Dictionary in the Word Glossary](../reference/word-glossary.md#word-type-dictionary).

## Put

[Put in the Word Glossary](../reference/word-glossary.md#word-builtin-put).

## Place

[Place in the Word Glossary](../reference/word-glossary.md#word-builtin-place).

## Sources and witnesses

[Runtime values](../../lyger/runtime_types.go), [source operations](../../lyger/runtime_values_source.go), and [value regressions](../../lyger/runtime_values_test.go) define runtime collections. [Member-resolution tests](../../lyger/member_resolution_regression_test.go) cover vocabulary ownership.
