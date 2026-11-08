# Words A–Z

Find a word by its root, then choose its qualified meaning. Entries explain the
declared forms, grammatical role, supported context, and examples. The same
spelling can denote different meanings: App Output and IO.Output, for example,
have different owners and behavior.

Builtin vocabulary is available implicitly. [Public libraries](../stdlib/index.md)
require an explicit import. Advanced Storage compositions appear separately below
and require their documented native authoring context.

The index covers the reviewed public inventory reconciled with current source.
It is not a compiler-generated export list or a runtime certification. For
language concepts rather than individual words, use the [terminology glossary](../spec/glossary.md).

## How to read the labels

The first label is the word's **grammatical role**, such as noun, verb or article.
The second describes implementation support for the linked meaning:

| Label | Meaning |
| --- | --- |
| **Supported** | The documented meaning is implemented within its stated contexts and operand rules. |
| **Partial** | Some of the documented forms or stages work, but a specific capability is missing. The entry's **Partial support** paragraph names what works and the exact gap. |
| **Rejected** | The documented operation is unavailable. Recognizing its spelling or grammar does not make that operation executable. |

Requiring an App, an exact descriptor target or Number operands is an intentional
usage rule, not by itself a reason to mark a word partial. These labels describe
reviewed implementation coverage. [Test execution and other evidence](../contributing/evidence.md#verification-stage)
are recorded separately.

**Authoring** identifies advanced word-definition or native integration context.
It is an audience label, independent of implementation support.

<!-- word-index -->
