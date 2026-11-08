# Iterative Types

| Type | Meaning | Operations |
| --- | --- | --- |
| [List](#lists) | Ordered owned values with an element type. | Construct, append, read and replace by zero-based index. |
| [Map](#maps) | Owned typed values keyed by Text bytes. | Insert, read and replace; keys have deterministic byte order. |
| [Dictionary](#dictionaries) | A vocabulary namespace with its own owner. | Declare and import words; use Map for runtime key/value data. |

Assignment, arguments, results and captures snapshot ordinary elements while retaining nested [Pointer](values.md#pointer) identity. These operations introduce no implicit traversal, membership or removal API.

## Lists

<!-- word:type:list -->

Own an ordered sequence of values with a checked element type.

**Usage and example.** The [value program](../examples/word-values.l) declares a List of Number, appends 7, replaces it with 9 and reads the checked element. The following excerpt contains that sequence.

### Declare a List

<!-- example:word-values-list -->

`of Number` declares the element type. The read states its expected result type; incompatible reads or updates reject. Indices must be nonnegative integers within bounds. A List constructed from differently typed values infers Thing elements and preserves their individual kinds. Typed reads verify complete element types and Function signatures. A List owns its elements; an L List does not become native `Props` or `Texts` metadata.

## Maps

<!-- word:type:map -->

Own values indexed by exact Text keys with a checked value type.

**Usage and example.** The [value program](../examples/word-values.l) stores Ada under the Text key first and reads it back. A Map is runtime data, unlike a vocabulary Dictionary.

<!-- example:word-values-map -->

`as` supplies the Text key; `at` selects it. Reading a missing key reports an error. No default value is inserted. Text keys preserve their bytes rather than applying Unicode normalization.

## Dictionaries

<!-- word:type:dictionary -->

Own a namespace of lexical declarations and their canonical member identities.

**Usage and example.** The [public-values program](../../demo/paragraphs/public-values.l) declares Shared and its Display action. The [file-import example](imports.md#explicit-source-file-imports) imports a project-owned member without changing its owner.

A Dictionary owns named vocabulary members. `Create "Shared" as a New Public Dictionary` declares the owner; `Shared's "Display|-s|-ed|-ing|-er"` declares a member word. Aliases preserve the owner and member identities. See the [public-values example](../guide/values.md#text-number-and-truth-together).

An imported dictionary has its own source scope and exports. [Indexed imports](imports.md#indexed-dictionaries) select those exports; computed Map keys do not participate in lexical lookup.

## Put

<!-- word:builtin:put -->

Insert a value into a supported runtime destination with its required element or target type.

**Usage and example.** `Put 7 in items` appends to a List; the [value program](../examples/word-values.l) also supplies a Map key and updates a Pointer target. Put does not mutate vocabulary Dictionaries.

## Place

<!-- word:builtin:place -->

Replace an existing indexed List element or supported keyed runtime field.

**Usage and example.** `Place 9 in items at 0` in the [value program](../examples/word-values.l) retains the List's Number element type. The index is checked; Place does not append past the end.

## Sources and witnesses

[Runtime values](../../lyger/runtime_types.go), [source operations](../../lyger/runtime_values_source.go), and [value regressions](../../lyger/runtime_values_test.go) define runtime collections. [Member-resolution tests](../../lyger/member_resolution_regression_test.go) cover vocabulary ownership.
