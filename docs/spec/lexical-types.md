# Lexical Types

| Lexical type | Role in a sentence | Examples |
| --- | --- | --- |
| Nouns | Name values, categories, members, or references. | `Text`, `App`, a declared name, `it` |
| Verbs | Express actions and relationships. | `Create`, `Define`, core `Print` |
| Adjectives | Qualify nouns or describe how to construct them. | `Public`, `New`, `Generic`, `Basic`, `Main` |
| Adverbs | Modify an action, declaration, comparison, or sequence. | `Publicly`, `Not`, `Less`, `Greater`, `Then` |
| Articles | Mark a reference, declaration, argument, or absence. | `the`, `a`/`an`, `any`, `no` |
| Prepositions | Connect a phrase to its value, definition, owner, or source. | `as`, `of`, `with`, `from`, `to`, `for` |

These are grammatical roles. [Primitive](values.md), [iterative](collections-and-categories.md), and [applicational](applicational-types.md) types describe what values represent. A noun can name a Text value, a Dictionary, or an App. In `The App Prints message`, App is the subject, Prints is the verb, and message supplies its operand. The value of message may be Text.

Pronouns such as `it` are nouns, not another value type. The configured vocabulary also includes conjunctions (`and`, `or`) and evaluation words (`if`, `else`, `otherwise`, `while`). They connect actions or select control-flow forms. Recognition alone does not make every combination executable. See [grammar](grammar.md) for those rules and the complete preposition inventory.

## Where words come from

| Source | How the word becomes available |
| --- | --- |
| Language roots | The embedded English lexicon supplies words such as Create, App, New, and As, with their grammatical roles and inflections. |
| Builtin definitions | The [core constructions and callables](../reference/word-glossary.md) are available without `Use`. |
| User-defined words | A program defines words by composing existing actions, construction rules, or logeme sequences. They follow the declaring scope. |
| Imported dictionary words | A `Use` statement introduces an explicitly selected export. The root registry currently contains no bundled dictionaries. |

A word's origin does not determine its part of speech. Core Print is a verb composed in L. Start and Execute remain declared but their lifecycle effects are unavailable. Generic definitions extend vocabulary through these supported mechanisms. They do not introduce unrestricted English interpretation or generic type parameters.

## Attributes, Definitions, and lexical categories

Noun, Verb, Adjective, and Adverb are recognized lexical declaration categories. Articles and prepositions participate in grammar. They are not additional public constructors to use as `New Article` or `New Preposition`.

Attribute and Definition are [applicational categories](applicational-types.md#attributes-and-definitions). A source-defined word can carry a native definition, but the word's grammatical role remains distinct from that implementation. Similarly, an adjective descriptor supplies construction rules rather than an ordinary object field.

### Noun

[Noun in the Word Glossary](../reference/word-glossary.md#word-type-noun).

### Verb

[Verb in the Word Glossary](../reference/word-glossary.md#word-type-verb).

### Adjective

[Adjective in the Word Glossary](../reference/word-glossary.md#word-type-adjective).

### Adverb

[Adverb in the Word Glossary](../reference/word-glossary.md#word-type-adverb).

## Describe with an adjective

<!-- example:main-descriptor -->

The first sentence declares Main as an Adjective targeting Service, Object, or Function. Each `It describes` adds a rule to that same adjective. The Function rule supplies its name and signature, the Object rule wraps that Function, and the Service rule wraps the Object. `New Main Service` applies the construction. `Generic` describes the pattern, while `any Process` accepts the supplied body. These builtin declarations are already loaded.

## Define and use a word

The [authoring guide](../guide/definitions.md#compose-a-source-action) explains typed source actions and [word forms](lexical-structure.md#vocabulary-and-inflection). Choose a meaning and grammatical role before declaration syntax. The current CLI supports one standalone App with core Print.
