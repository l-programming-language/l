# Grammar and vocabulary

The exact English grammar lives in `lexicon/eng/grammar.json`, and the exact configured words live in `lexicon/eng/vocabulary.json`. This chapter explains their notation and semantic groupings. The generated inventory preserves every rule, named case, and vocabulary entry without treating every table combination as executable.

## Inventory and authority

The inspected English grammar has 99 rules and 307 named cases. Its categories are three NOUN rules, thirteen PHRASE rules, sixty-nine CLAUSE rules, one CONDITION rule, four CONDITIONAL rules, five PREDICATE rules, and four STATEMENT rules. The vocabulary has 72 entries across eight parts of speech. These counts describe the reviewed source snapshot. Freshness checks must catch later changes.

The [complete rule and case inventory](../_generated/grammar.md) provides a stable entry for every configured recognizer, and the [vocabulary inventory](../_generated/vocabulary.md) retains each root, inflection, role, and handler description. These tables supplement the semantic explanations below.

The grammar controls recognition. Bound source declarations, descriptor construction, typed instruction lowering, and backend support independently determine whether a recognized sentence can be built. A retained grammar entry may be partial or explicitly rejected in executable contexts.

## Configuration notation

The notation is specialized and is not EBNF:

| Notation | Actual parser meaning |
| --- | --- |
| `A \| B` | Sequence: match A, then B |
| `TERM*` | Optional TERM. Failure leaves the prior state intact |
| `[TERM]` | Multiple association of TERM, using the grammar's coordinated-sequence handling |
| `[TERM]*` | Optional multiple association |
| `CASE` | Select a named case for this rule |
| `ID` | A resolved/reference identifier position |
| `"ID"` | A declaration identifier position |
| `FAMILY.kind.case` | Qualified grammar or lexical role |
| backtick `~` pattern | Text token category |
| `%~%`, `?~?`, `!~!` | Number, Truth, Null token categories |
| `#~#` | Native type token category |
| `&~&`, `{~}` | List and Dictionary type categories |

Square brackets do not mean optionality here, and a vertical bar does not mean alternation. Named cases are alternatives. Symbols such as `%~%` describe already-typed tokens rather than source text the reader should type around a literal.

The decoder validates family names, qualification depth, type patterns, and declared identifier markers. PRONOUN is not admitted as a grammar rule element even though lexical pronouns participate in source binding.

## Matching and ambiguity

Nested grammar consumes a contiguous source prefix. Every candidate case begins from the same state. The parser chooses the successful candidate consuming the longest prefix. Equal lengths use sorted case names, so map iteration order does not decide the parse. Optional elements commit only on success. A full statement rejects leftover unassociated tokens.

A parse choice is different from semantic ambiguity. If a spelling resolves to more than one valid lexical declaration, binding can report AmbiguousReference even after a grammar shape has been recognized. Coordinated subjects/objects retain occurrence identity. Identical token spelling is not token identity.

## Values, subjects, and objects

The value rule recognizes definitive references and Text/Number/Truth/Null/type/List/Dictionary patterns. Subject cases describe declarations, references, definitions, directives, utilization, and typed/collection declarations. Object cases additionally distinguish named objects, dependencies, definitive references, and scalar operands.

Articles and adjectives modify these positions. `a`/`an` is indefinite, `the` definite, `any` identifies variable arguments in supported action signatures, and `no` expresses absence or a supported negative condition. These roles are not interchangeable decoration.

A subject identifies what is being defined or acted upon, while an object supplies the action operand. That distinction matters for `The App Prints message`: the App owns the native body effect, while message supplies Print's argument.

### Read a complete sentence in context

The [complete first program](../guide/first-program.md#the-source) gives those grammatical roles a concrete meaning:

| Source phrase | Role in this program |
| --- | --- |
| `"message" is` followed by Text | Declares a name and assigns its literal value |
| `Create "Hello" as a New Basic App` | Imperative declaration. The word `as` introduces the construction |
| `The App Prints message` | Active statement: App is the subject, Prints is the action, and message is its operand |

`The App` associates Print with the App's generated body. The built standalone App performs that output directly. Grammatical form does not erase that ownership distinction.

### Any

[Any in the Word Glossary](../reference/word-glossary.md#word-grammar-word-any).

### No

[No in the Word Glossary](../reference/word-glossary.md#word-grammar-word-no).

### A

[A in the Word Glossary](../reference/word-glossary.md#word-grammar-word-a).

### The

[The in the Word Glossary](../reference/word-glossary.md#word-grammar-word-the).

## Phrases

The thirteen phrase families are generic, directive, declarative, locative, objective, inflective, associative, indicative, utilization, dependency, signature, arguments, and results.

Declarative `as` relates a subject to a type, value, definition, or alias. Directive `for` introduces action composition. Utilization `to` connects a native-backed definition or imperative use. Associative `with` composes described child structures. The word `using` is the applicative form of configured Use and supplies parameters. Dependency `from` selects a dictionary or explicit source provider in supported import forms.

Locative `at` and inflective `in` participate in [collection updates and typed reads](collections-and-categories.md). A phrase's meaning follows its owning action, not the preposition alone.

Signature phrases declare ordered Function inputs and outputs, argument phrases bind invocation operands, and result phrases name invocation outputs. These forms retain the [typed Function checks](applicational-types.md#functions-properties-and-operations) for counts, types, and assignment after invocation.

### Compose an action with `for`

`for` introduces a source action body. Core Print uses this form to compose its private output leaf through Process. Typed arguments, inflections and pronouns retain their declared meaning; `and then` is source sequencing rather than the grammar JSON's notation.

The [Main descriptor](lexical-types.md#describe-with-an-adjective) uses `of`, `with` and `using` in Function/Object/Service construction. Core is already loaded.

### As

[As in the Word Glossary](../reference/word-glossary.md#word-grammar-word-as).

### Of

[Of in the Word Glossary](../reference/word-glossary.md#word-grammar-word-of).

### At

[At in the Word Glossary](../reference/word-glossary.md#word-grammar-word-at).

### With

[With in the Word Glossary](../reference/word-glossary.md#word-grammar-word-with).

### In

[In in the Word Glossary](../reference/word-glossary.md#word-grammar-word-in).

### Out

[Out in the Word Glossary](../reference/word-glossary.md#word-grammar-word-out).

### By

[By in the Word Glossary](../reference/word-glossary.md#word-grammar-word-by).

### From

[From in the Word Glossary](../reference/word-glossary.md#word-grammar-word-from).

### To

[To in the Word Glossary](../reference/word-glossary.md#word-grammar-word-to).

### Than

[Than in the Word Glossary](../reference/word-glossary.md#word-grammar-word-than).

### For

[For in the Word Glossary](../reference/word-glossary.md#word-grammar-word-for).

## Clauses and grammatical forms

Most action families have imperative, active, passive, and applicative clauses. The configured verb root acts as the imperative. Active is typically the `-s` form, passive the past form, and applicative the `-ing` form. Source inflection declarations determine the actual spelling.

| Action family | Root | Current semantic boundary |
| --- | --- | --- |
| Declarative | Create | Named values, actions, and supported construction |
| Definitive | Define | Value and native-backed definitions |
| Descriptive | Describe | Source descriptors and bounded construction refinements |
| Associative | Use | Aliases and explicit imports |
| Generative | Build | Unavailable runtime effect |
| Executive | Run | Unavailable runtime effect |
| Suspensive | Await | Unavailable runtime effect |
| Interruptive | Stop | Unavailable runtime effect |
| Observative | Show | Unavailable runtime observation |
| Terminative | Exit | Unavailable runtime effect |
| Expansive / locative | Put / Place | Runtime updates outside the standalone boundary |
| Receptive / reductive | Get / Take | Recognized collection forms with no established general execution |
| Inflective / deflective | Import / Export | Configured roots that do not replace supported Use import syntax |
| Generic | Source-defined word | Depends on bound typed definition/action |

Grammatical voice does not widen the current executable boundary. Build, Run, Await, Stop, Show and Exit reject as runtime-dependent effects. Standalone applicative clauses do not become executable statements merely because action bodies use them.

<a id="supported-passive-await"></a>

### Passive Await

The grammar recognizes passive Await, but lifecycle execution is unavailable. No outer-program workaround supplies the removed Runtime.

### Recognized but rejected active Await

App-body Await is also unavailable. Unsupported effects must reject without being moved, omitted or replaced by a generated runner.

### Get

[Get in the Word Glossary](../reference/word-glossary.md#word-builtin-get).

### Take

[Take in the Word Glossary](../reference/word-glossary.md#word-builtin-take).

### Import

[Import in the Word Glossary](../reference/word-glossary.md#word-builtin-import).

### Export

[Export in the Word Glossary](../reference/word-glossary.md#word-builtin-export).

## Conditions

The condition rule recognizes truth, has, equality, inequality and ordering forms. Their exact spellings come from configured vocabulary. A branch separator comma is source punctuation, distinct from the grammar configuration's `|` sequence notation.

Recognition does not supply runtime selection, error inspection, boolean short-circuiting or general loops. Keep [control-flow limits](control-flow.md) distinct from this grammar inventory.

## Predicates and statements

Predicates assemble clauses in imperative, active, passive, applicative or conditional form. Statements choose objective or subjective organization. In the supported example, the App subject owns its Print body. Source order, ownership and paragraph provenance survive expansion.

## Vocabulary reference

Nouns include pronouns, Input/Output/Error, List/Map/Dictionary, Null/Thing/Text/Number/Byte/Truth/Pointer, native Func/Prop/Op concepts, Definition/Attribute, and Object/Class/Service/App. The adjective entries are Public, New, and Generic. Adverbs are Publicly, Not, Less, Greater, and Then. Prepositions are as/of/at/with/in/out/by/from/to/than/for. Articles are any/no/a/the, conjunctions and/or, and evaluations if/else/otherwise/while.

The [builtin vocabulary](../reference/word-glossary.md) includes configured words and source-defined core words. Core Print is implicit; no standard-library dictionary is bundled. New words compose existing definitions. Vocabulary recognition alone does not implement an action.

## Sources and witnesses

Exact data: [grammar](../../lexicon/eng/grammar.json), [vocabulary](../../lexicon/eng/vocabulary.json). Parser: [grammar terms](../../lyger/grammar_terms.go), [matching](../../lyger/grammar_match.go), [grammar types](../../lyger/grammar_types.go), [sequence parsing](../../lyger/grammar.go). Lowering: [instructions](../../lyger/instructions.go), [native roles](../../lyger/ir_kinds.go). Witnesses: [term decoding](../../lyger/grammar_terms_test.go), [consumption](../../lyger/grammar_consumption_test.go), [sequence ownership](../../lyger/grammar_sequence_test.go), [paragraph grammar](../../lyger/paragraph_grammar_architecture_test.go), and [phase-one language](../../lyger/phase1_language_test.go).
