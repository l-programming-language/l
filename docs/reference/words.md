# Native Words

<a id="words-az"></a>

Find the implicit builtin, core and configured grammatical vocabulary by root.
Every link opens its complete [Word Glossary](word-glossary.md) entry, including
the source definition, forms, part of speech, description and usage examples.

These words require no dictionary import. For imported words, browse
[IO](../stdlib/eng/io.md), [Formatting](../stdlib/eng/format.md),
[Runtime](../stdlib/eng/runtime.md) or [Storage](../internals/storage.md).
The [terminology glossary](../spec/glossary.md) explains language concepts.

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

<!-- word-index:builtins -->
