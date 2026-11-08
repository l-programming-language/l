# Lexical Types

| Lexical type | Role in a sentence | Examples |
| --- | --- | --- |
| Nouns | Name values, categories, members, or references. | `Text`, `App`, `Output`, a declared name, `it` |
| Verbs | Express actions and relationships. | `Create`, `Show`, `Run`, `Execute`, imported `Print` |
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
| Builtin definitions | The [source-defined foundation](../reference/builtins/index.md#source-defined-facilities) is available without `Use`. |
| User-defined words | A program defines words by composing existing actions, construction rules, or logeme sequences. They follow the declaring scope. |
| Imported dictionary words | A `Use` statement introduces an exported word, such as IO.Print. Bundling the standard library does not make its exports implicit. |

A word's origin does not determine its part of speech. Execute is a builtin verb written using other verbs, and a user can compose a verb the same way. Generic definitions extend vocabulary through these supported mechanisms. They do not introduce unrestricted English interpretation or generic type parameters.

## Attributes, Definitions, and lexical categories

Noun, Verb, Adjective, and Adverb are recognized lexical declaration categories. Articles and prepositions participate in grammar. They are not additional public constructors to use as `New Article` or `New Preposition`.

Attribute and Definition are [applicational categories](applicational-types.md#attributes-and-definitions). A source-defined word can carry a native definition, but the word's grammatical role remains distinct from that implementation. Similarly, an adjective descriptor supplies construction rules rather than an ordinary object field.

### Noun

<!-- word:type:noun -->

Classify a word that names a value, category, reference or constructed definition.

**Usage and example.** Process in the [core construction example](definitions.md#process-function-and-return) is a declared noun. Its Processing form is an applicative verb derived from the noun’s active slot. The [first program](../guide/first-program.md#read-the-program) uses the App and Text categories as nouns. Noun is a lexical role, not a scalar runtime type.

### Verb

<!-- word:type:verb -->

Classify a word that expresses an action or relationship with explicit operands.

**Usage and example.** Launch in the [source-action example](../guide/definitions.md#compose-a-source-action) composes Build, Run and Await and declares their invocation forms. A verb's actor noun constructs that same selected meaning. It is not a new public synonym.

### Adjective

<!-- word:type:adjective -->

Classify a word that qualifies a noun or describes construction for exact nominal targets.

**Usage and example.** Main below declares Service, Object and Function targets. The [recursive fixture](../../lyger/test/classes/recursive.l) applies Main's Object rule through further descriptors. An unrelated same-spelled target does not satisfy that declaration.

### Adverb

<!-- word:type:adverb -->

Classify a word that modifies an action, declaration, comparison or source sequence.

**Usage.** Publicly on the Identity declaration in the [Function example](../examples/composed-functions.l) publishes that Function, while Then in a [source action](../guide/definitions.md#compose-a-source-action) separates composed actions.

**Partial support.** The grammar recognizes adverbs, and configured roles such as visibility, condition negation, comparison and progression have specific handling. Declaring a new Adverb does not provide a general mechanism for applying its authored body to an action. Adverb bodies have no counterpart to the source-defined adjective constructor rules. Use the implemented configured modifiers in their documented contexts.

## Describe with an adjective

<!-- example:main-descriptor -->

The first sentence declares Main as an Adjective targeting Service, Object, or Function. Each `It describes` adds a rule to that same adjective. The Function rule supplies its name and signature, the Object rule wraps that Function, and the Service rule wraps the Object. `New Main Service` applies the construction. `Generic` describes the pattern, while `any Process` accepts the supplied body. These builtin declarations are already loaded.

## Define and use a word

The [Launch example](../guide/definitions.md#compose-a-source-action) defines a verb from Build, Run, and Await. `Launch First` calls that sequence for an App. Its pipe-delimited label declares related spellings, as explained under [word forms](../guide/values.md#read-pipe-delimited-word-forms).

A [native-backed definition](../guide/native-definitions.md#start-from-the-existing-foundation) instead uses logeme sequences inside `<DEF>` framing. Typed slots bind the caller's arguments. Both approaches define vocabulary that participates in L grammar. The native body additionally obeys its target's contracts.

The [English vocabulary](../../lexicon/eng/vocabulary.json), [builtin core](../../builtins/eng/core.l), and [lexical representation](../../lyger/lexical_types.go) provide the corresponding source definitions.
