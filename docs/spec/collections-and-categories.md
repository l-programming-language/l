# Iterative Types

| Type | Meaning | Operations |
| --- | --- | --- |
| [List](#lists) | Ordered owned values with an element type. | Construct, append, read and replace by zero-based index. |
| [Map](#maps) | Owned typed values keyed by Text bytes. | Insert, read and replace; keys have deterministic byte order. |
| [Dictionary](#dictionaries) | A vocabulary namespace with its own owner. | Declare and import words; use Map for runtime key/value data. |

Assignment, arguments, results and captures snapshot ordinary elements while retaining nested [Pointer](values.md#pointer) identity. These operations introduce no implicit traversal, membership or removal API.

## Lists

### Declare a List

```l
Define "items" as a New List of Number.
Put 7 in items.
Place 9 in items at 0.
Define "first" as a New Number from items at 0.
Show first.
```

`of Number` declares the element type. The read states its expected result type; incompatible reads or updates reject. Indices must be nonnegative integers within bounds. A List constructed from differently typed values infers Thing elements and preserves their individual kinds. Typed reads verify complete element types and Function signatures. A List owns its elements; an L List does not become native `Props` or `Texts` metadata.

## Maps

```l
Define "names" as a New Map of Text.
Put `Ada` in names as `first`.
Define "name" as a New Text from names at `first`.
Show name.
```

`as` supplies the Text key; `at` selects it. Reading a missing key reports an error. No default value is inserted. Text keys preserve their bytes rather than applying Unicode normalization.

## Dictionaries

A Dictionary owns named vocabulary members. `Create "Shared" as a New Public Dictionary` declares the owner; `Shared's "Display|-s|-ed|-ing|-er"` declares a member word. Aliases preserve the owner and member identities. See the [public-values example](../guide/values.md#text-number-and-truth-together).

An imported dictionary has its own source scope and exports. [Indexed imports](imports.md#indexed-dictionaries) select those exports; computed Map keys do not participate in lexical lookup.

## Sources and witnesses

[Runtime values](../../lyger/runtime_types.go), [source operations](../../lyger/runtime_values_source.go), and [value regressions](../../lyger/runtime_values_test.go) define runtime collections. [Member-resolution tests](../../lyger/member_resolution_regression_test.go) cover vocabulary ownership.
