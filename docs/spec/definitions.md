# Definitions and extensibility

L extends vocabulary through declarations, source actions, descriptors and native-backed definitions. Typed arguments and source ownership constrain each mechanism. Arbitrary English is not executable.

## Create, Define, and Describe

Create introduces a supported value or lexical action. Define supplies a value or native-backed definition. Describe associates an adjective with nominal targets and constructor/wrapper rules. Active/passive grammar and resolved ownership determine the declaration site.

Declarations retain tokens, formal identities, bindings, descriptor targets and source spans. Publication validates relationships before committing. Expansion creates use-specific instances without mutating prototypes.

### Create

[Create in the Word Glossary](../reference/word-glossary.md#word-builtin-create).

### Define

[Define in the Word Glossary](../reference/word-glossary.md#word-builtin-define).

### Describe

[Describe in the Word Glossary](../reference/word-glossary.md#word-builtin-describe).

### Declaration roles and identity

In the maintained English grammar, `Define "Word" as` followed by a complete backtick-delimited `<DEF>…<DEF>` literal declares a noun lexical definition. `Define "Word" to` with such a literal declares a verb lexical definition. `Create "Word" for ...` declares a source action. These are specific declaration forms: `as` also participates in value bindings and Function construction, so it does not universally mean noun registration. Complete framing alone does not establish that the native body is valid.

A Function declaration binds its invocation verb, including through a selected source constructor. Ordinary source Functions retain Function value storage while exposing that verb's morphology; native/generic Functions use lexical definition records. The invocation word and value share canonical identity. Explicit adjective descriptors register their nominal targets and construction rules separately.

Keep lexical family, grammatical role at use, storage category and canonical identity distinct. A verb's objective form serves as an actor noun. A noun or alias defined from that actor keeps its nominal role without changing the verb. Root aliases retain the verb family. Passive forms do not become descriptors or actor constructors. New/Generic construction requires a nominal role, a Function's root constructor, or a root reference with an applied descriptor targeting its canonical actor. Arbitrary modifiers do not supply that role.

See [registration](../../lyger/registration.go), [Function binding](../../lyger/binding.go), and the [Function identity](../../lyger/function_invocation_words_test.go) and [modifier homonym](../../lyger/word_library_pos_test.go) witnesses. The [authoring convention](../guide/definitions.md#organize-words-by-meaning-and-grammatical-role) governs design choices without changing these rules.

## Source actions

Core Print is a source action composed through Process and its private Log action. Typed operands and definition-site identity survive expansion. Start and Execute remain declared, but their runtime effects reject.

Calls check argument count and nominal type. A one-argument source action with coordinated operands expands its complete body for each operand in order. Definition-site references retain their bindings. Actual arguments and diagnostics belong to the call. Source-action expansion cycles reject. [Function invocations](applicational-types.md#functions-properties-and-operations) have their own callable semantics.

## Inflection declarations

Quoted labels list bar-separated forms. Nouns use root/plural/descriptive/active positions, while verbs use root/active/passive/applicative/objective positions. One leading hyphen appends a suffix. Additional hyphens remove trailing root bytes first. See [lexical rules](lexical-structure.md#vocabulary-and-inflection).

`Start|-s|-ed|-ing|-er` produces Start, Starts, Started, Starting and Starter. `Execute|-s|-d|--ing|--or` trims the final e for Executing/Executor. Source-action and Function clauses require the requested declared form on their selected identity. Configured native operations retain their grammar-owned forms, including passive `Worker is run`. Inflection does not publish private helpers or change ownership.

## Descriptors and constructors

When the selected construction header starts with an unscoped Label, that slot names its result, regardless of spelling. Transparent source forwarding inherits this contract, including zero-operand constructors and descriptor replacements. Multi-stage compositions require explicit results. Scoped pattern and iteration labels keep their binding roles; forwarding preserves App capture rejection.

Descriptors bind nominal identities to ordered construction rules. A source noun can bind typed formals and fixed operands to another constructor, with explicit fixed operands. A replacement constructs directly from those operands. A wrapper explicitly names the base result and retains its evaluation. Matching rules apply in authored order. Same-named private declarations remain distinct, and aliases preserve identity. Missing constructors and recursive expansion reject.

The [authoring guide](../guide/definitions.md#compose-a-source-action) describes source actions and grammatical design. New selects construction, Generic participates in source composition and Public controls visibility. None infers Rust lifetimes or generic arguments.

Main constructs a Function inside an Object inside a Service. Basic supplies that Main Service to an App. These rules define `New Basic App`.

### New

[New in the Word Glossary](../reference/word-glossary.md#word-builtin-new).

### Generic

[Generic in the Word Glossary](../reference/word-glossary.md#word-builtin-generic).

## Process, Function, and Return

<!-- example:definitions-callable-core -->

Process composes ordered Syllables through W. Function supplies FW/Func metadata: Name, typed Inputs, output types and an instantiated Process. Source Function declarations use this selected definition too. Function-local Return binds its builtin member after checking source results.

A structured Process can declare typed Inputs for a selected pattern or iteration body. The [Process contract](../reference/word-glossary.md#word-builtin-process) specifies scope, cardinality and native type/borrow checking. The [native guide](../guide/native-definitions.md#keep-branch-data-inside-its-process) explains their ownership. A Process has no callable outputs or capture runtime.

Native source Function constructors accept explicit generic parameter atoms before their signature. Descriptors on derived invocation nouns retain the base result and forward checked operands once, in source order, including through transparent value definitions. Structured bodies retain their execution placement. One selected signature determines operand boundaries and result bindings: wrappers preserve it, replacements supply it, and source forwarding inherits its remaining parameters. Each declaration owns its body and formal inputs, including declarations sharing a sentence. Native move/borrow checks still apply, and forwarding introduces no clone. The [native guide](../guide/native-definitions.md#compose-a-wrapper) states wrapper requirements.

<!-- example:main-descriptor -->

Each `It describes` refines Main. App construction selects its Service, Object and Function rules. Main supplies no inputs or outputs. [Ordinary Functions](applicational-types.md#functions-properties-and-operations) supply their declared signatures and bodies. A standalone Main Service remains structural. Methods, overloads and variadic signatures are not introduced.

## Framing and escaping

Source definitions use backtick Text with one complete outer `<DEF>…<DEF>` pair. Both markers are spelled `<DEF>`, and neither is an XML closing tag. Interior/incomplete markers remain exact ordinary Text. The [framing helper](../../lyger/definition_bindings.go) inspects quoted/unquoted frames without decoding bodies or validating glyphs. Explicit internal template parsing accepts markerless interiors and rejects malformed, empty or nested definitions. Those internal interfaces do not remove source backticks.

Core's private Log definition demonstrates a complete frame and one bounded output call. That native body is not a pattern for embedding orchestration or error policy.

A body may be one expression, a JSON string array representing a word, or an array of word arrays. Grouping survives instantiation. Expanded display and explicit template parsing decode quote, backslash and backtick escapes once at their respective boundaries. Lexeme storage preserves encoding and quote presence. Native escapes such as `\n` survive, and arguments are not decoded again.

## Typed slots

`<@Name|Type@>` declares a typed slot, and `<@Name@>` references an instantiated definition. Compatible repetitions share one ordered formal. Conflicting declarations reject. Native slot labels use the [internal identifier validator](lexical-structure.md#names-possessives-and-qualification). Its internal dotted forms do not permit dot-qualified authored L names. Target-language identifier validation is separate.

| Binding kind | Data |
| --- | --- |
| Text | Unquoted scalar bytes, safely encoded for native use |
| Number | Finite binary64 with native `f64` typing |
| Truth | Boolean scalar |
| Null | Explicit native display atom, not a null pointer |
| Syllable | Instantiated expression |
| Syllables | Ordered expressions with preserved cardinality |
| Props | Typed native property records |
| Texts | Ordered native type/output text collection |
| Type | Native type atom or selected source type declaration, retaining its dependency |
| Atoms | Ordered scalar/name/reference arguments, never an executable body |
| Label | Validated result identifier, distinct from an assignment target |
| Definition | Instantiated definition with identity and grouping |

Native Thing slots accept Text, Number, Truth and Null without converting them to strings. Core Log requires Text for its single output call. Preserve explicit [value ownership](values.md#conversion-and-aliasing-boundaries); a generic slot does not provide removed runtime values.

Type, Atoms and Label are frontend roles using existing AML metadata. Nested constructors bind results once, in operand order, before the consuming word. Composition slots retain their selected definitions. Native and ordinary Functions share body lowering and bind their own invocation words. Dependencies retain canonical source ownership.

Definition arguments use a finite binary64 encoder.
It emits the shortest round-trip decimal with an explicit `_f64` suffix,
preserving negative zero and the parsed value's rounding. Integral, leading-plus
and hexadecimal floating-point spellings therefore become valid Rust expressions.
Number in a Thing slot uses the same encoding. Native fragments and integers
remain opaque. Malformed/overflowing definition arguments retain contextual
Syntax errors and native parse causes; stored nonfinite primitives remain Unsupported.

Text scalar encoding protects backend delimiters, metadata separators, repeated
spaces and Lucid openers with Rust escapes. Definition arguments, including Text
in Thing slots, use that boundary. Literal storage, source rendering, raw definitions and
completed AML are not rewritten. Unknown kinds, incompatible types, missing
bindings, invalid native names and malformed compositions reject. Explicit native
fragments retain target-dependent meaning.

## Composition and source ownership

`^` integrates a child, `*` combines siblings and parentheses group compositions. Assembly order does not promise universal runtime evaluation order. Func/Prop/Op records, imports and children retain their source-file owner. Classes cannot inject defaults into Objects.

Glyph-local Lifetimes/GenericTypes accept native String-list interiors only on FW/HY/MY declarations and Y/named-KY applications, with their explicit absence/blank rules. They are not L generics, Func/Prop/Op fields, inferred ownership or inherited settings. Rust rejects Deferred metadata, even false. Scheduling implies neither scope-exit cleanup nor hidden continuations.

Core Objective preserves native-composition grouping. This is not an L collection literal or a general runtime-order guarantee.

## Sources and witnesses

Sources: [core](../../builtins/eng/core.l), [parser/encoder](../../lyger/definition_bindings.go), [argument kinds](../../lyger/definition_kinds.go), [construction](../../lyger/construction.go), [actions](../../lyger/instructions.go), [lexical forms](../../lyger/lexical_types.go), [Rust inputs](../../builtins/README.md). Witnesses: [composition](../../lyger/definition_test.go), [framing](../../lyger/definition_framing_test.go), [slot names](../../lyger/definition_slot_names_test.go), [actions](../../lyger/action_definition_test.go), [descriptor identity](../../lyger/paragraph_descriptor_identity_test.go).
