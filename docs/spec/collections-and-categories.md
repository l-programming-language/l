# Iterative Types

| Type | What it represents | Current use in L |
| --- | --- | --- |
| [List](#lists) | An ordered collection of values. | Recognized declarations and collection grammar; general element operations and iteration are not implemented in the executable source path. |
| [Dictionary](#dictionaries) | Named entries sharing one owner. | Supported vocabulary namespaces and member lookup; arbitrary runtime map operations and iteration are not implemented. |

Iterative types group values. Their names describe collection structure; they do not by themselves provide a loop or traversal API. For individual values, see [Primitive Types](values.md); for word roles and program structure, see [Lexical Types](lexical-types.md) and [Applicational Types](applicational-types.md).

## Lists

### Declare a List

<!-- example:list-declaration -->

`"items"` names the List and `as a List` selects its category. This fragment illustrates recognized grammar and storage; it has no compilation or runtime witness. Source element literals, insertion, indexing, update, membership, traversal and removal are not established APIs. The configured Put, Place, Get and Take word families do not supply those runtime operations on their own.

Internally, a List stores ordered values and their types. That representation does not specify element homogeneity, bounds errors or mutation aliasing. Native definition inputs such as `Props` and `Texts` have their own contracts; an L List does not automatically become one of those metadata arrays.

## Dictionaries

A Dictionary provides a namespace: named members belong to a shared owner. The public-values example creates `Shared` and defines its `Display` action:

<!-- example:public-values -->

`Create "Shared" as a New Public Dictionary` declares the owner. `Shared's "Display|-s|-ed|-ing|-er"` defines a member word; the possessive selects that owner. The member is a vocabulary definition, not an arbitrary runtime map entry. Aliases and qualified member references preserve the Dictionary's identity.

An imported vocabulary dictionary is a source unit with its own scope and selected exports. Its public Dictionary declaration exposes the namespace; [indexed imports](imports.md#indexed-dictionaries) explain how to use it. Internal key/value maps do not establish computed-key access, missing-key defaults, removal or iteration order for L programs.

## Sources and witnesses

See [collection storage](../../lyger/types.go), [type recognition](../../lyger/utilities.go), [collection grammar](../_generated/grammar.md), the [public-values demo](../../demo/paragraphs/public-values.l), and [member-resolution tests](../../lyger/member_resolution_regression_test.go). Example labels distinguish executable evidence from explanatory fragments.
