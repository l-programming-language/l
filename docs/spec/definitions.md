# Definitions and extensibility

L extends vocabulary through declarations, source actions, descriptors and native-backed definitions. Typed arguments and source ownership constrain each mechanism. Arbitrary English is not executable.

## Create, Define, and Describe

Create introduces a supported value or lexical action. Define supplies a scalar, runtime value or native-backed definition. Describe associates an adjective with nominal targets and constructor/wrapper rules. Active/passive grammar and resolved ownership determine the declaration site.

Declarations retain tokens, formal identities, bindings, descriptor targets and source spans. Publication validates relationships before committing. Expansion creates use-specific instances without mutating prototypes.

### Create

<!-- word:builtin:create -->

Introduce a named supported value, construction or source action in the current lexical scope.

**Usage and example.** `Create "Hello" as a New Basic App` in the [first program](../guide/first-program.md#read-the-program) creates an App. `Create "Launch..." for ...` in the [action example](../guide/definitions.md#compose-a-source-action) creates a reusable composition. Its parameter has the declared App meaning.

### Define

<!-- word:builtin:define -->

Associate a declaration with a value or an explicitly supplied native-backed definition.

**Usage and example.** The [public-values program](../../demo/paragraphs/public-values.l) defines Text, Number and Truth. [Framed definitions](#framing-and-escaping) bind native source through typed slots. Define does not bypass ownership, type or visibility checks.

### Describe

<!-- word:builtin:describe -->

Associate a descriptor with its canonical target and ordered construction refinements.

**Usage and example.** Main's `It describes` rules in the [descriptor example](lexical-types.md#describe-with-an-adjective) construct different Function/Object/Service targets. The [recursive caller](../../lyger/test/classes/recursive.l) applies the selected Object rule. Description does not make unrelated homonyms equivalent.

## Source actions

<!-- example:lexical-action-inflections -->

`any App` declares a typed argument, and `it` refers to that argument. Start expands Build/Run, and Execute expands Start/Await. Neither is a primitive instruction root.

Calls check argument count and nominal type. A one-argument source action with coordinated operands expands its complete body for each operand in order. Definition-site references retain their bindings. Actual arguments and diagnostics belong to the call. Source-action expansion cycles reject. [Function invocations](applicational-types.md#functions-properties-and-operations) have their own callable semantics.

## Inflection declarations

Quoted labels list bar-separated forms. Nouns use root/plural/descriptive/active positions, while verbs use root/active/passive/applicative/objective positions. One leading hyphen appends a suffix. Additional hyphens remove trailing root bytes first. See [lexical rules](lexical-structure.md#vocabulary-and-inflection).

`Start|-s|-ed|-ing|-er` produces Start, Starts, Started, Starting and Starter. `Execute|-s|-d|--ing|--or` trims the final e for Executing/Executor. Grammatical roles can change while ownership remains canonical. Inflection does not publish private helpers.

## Descriptors and constructors

Descriptors bind nominal identities to ordered construction rules. A source noun can bind typed formals and fixed operands to another constructor, as Formatting's Integer specializes Storage.Integral. A replacement constructs directly from those operands. A wrapper explicitly names the base result and retains its evaluation. Matching rules apply in authored order. Same-named private declarations remain distinct, and aliases preserve identity. Missing constructors and recursive expansion reject.

The [authoring guide](../guide/definitions.md#compose-a-source-action) demonstrates separate paragraph-local actions and descriptors. New selects construction, Generic participates in source composition and Public controls visibility. None infers Rust lifetimes or generic arguments.

Main constructs a Function inside an Object inside a Service. Basic supplies that Main Service to an App. These rules define `New Basic App`.

### New

<!-- word:builtin:new -->

Request construction through the selected noun and its matching descriptor rules.

**Usage and example.** `New Basic App` in the [first program](../guide/first-program.md#read-the-program) applies Basic and Main to the App body. `New List of Number` in the [value program](../examples/word-values.l) selects checked List construction. New supplies no universal constructor, clone or initialization fallback.

### Generic

<!-- word:builtin:generic -->

Describe a reusable source construction pattern whose explicit operands are supplied when applied.

**Usage and example.** Compact in the [source-action example](../../demo/paragraphs/local-definitions.l) describes a Generic App using any Process. `New Compact App` supplies its body. Generic alone does not infer native Rust type parameters, lifetimes or unconstrained constructor support.

**Supported boundary.** The selected source constructor declares the operands it accepts. Its callers and refinements retain that signature and canonical target. Generic describes this reusable construction mechanism, not a promise of inferred polymorphism.

## Process, Function, and Return

<!-- example:definitions-callable-core -->

Process composes ordered Syllables through W. Function supplies FW/Func metadata: Name, typed Inputs, output types and an instantiated Process. Source Function declarations use this selected definition too. Function-local Return binds its builtin member after checking source results.

A structured Process can declare typed Inputs for a selected pattern or iteration body. The [Process contract](../reference/builtins/definitions.md#process) specifies scope, cardinality and native type/borrow checking. The [Require walkthrough](../guide/definitions.md#keep-branch-data-inside-its-process) shows their use. A Process has no callable outputs or capture runtime.

Native source Function constructors accept explicit generic parameter atoms before their signature. Descriptors on derived invocation nouns retain the base result and forward checked operands once, in source order, including through transparent value definitions. Structured bodies retain their execution placement. One selected signature determines operand boundaries and result bindings: wrappers preserve it, replacements supply it, and source forwarding inherits its remaining parameters. Each declaration owns its body and formal inputs, including declarations sharing a sentence. Native move/borrow checks still apply, and forwarding introduces no clone. [Lined](../guide/native-definitions.md#compose-a-wrapper) demonstrates a source wrapper.

<!-- example:main-descriptor -->

Each `It describes` refines Main. App construction selects its Service, Object and Function rules. Main supplies no inputs or outputs. [Ordinary Functions](applicational-types.md#functions-properties-and-operations) supply their declared signatures and bodies. A standalone Main Service remains structural. Methods, overloads and variadic signatures are not introduced.

## Framing and escaping

Source definitions use backtick Text with one complete outer `<DEF>…<DEF>` pair. Both markers are spelled `<DEF>`, and neither is an XML closing tag. Interior/incomplete markers remain ordinary Text. Empty or nested definition bodies reject. The framing helper can inspect quoted/unquoted frames. Explicit internal template parsing also accepts markerless interiors. Those internal interfaces do not remove source backticks.

<!-- example:native-dictionary-definition -->

This retained ProbeIO fixture witnesses framing and escaping, not the recommended decomposition of IO policy. `BOD^E` integrates the Calculation, `{|Statement: …|}` supplies metadata and `~…~` encloses native statement text. Its [caller](../../cmd/l/test/runnable_service/basic.l) supplies Output. For compositional authoring, follow the [Send/Lined walkthrough](../guide/native-definitions.md#start-from-the-existing-foundation).

A body may be one expression, a JSON string array representing a word, or an array of word arrays. Grouping survives instantiation. Quote, backslash and backtick escapes decode once. Native escapes such as `\n` survive, and arguments are not decoded again.

## Typed slots

`<@Name|Type@>` declares a typed slot, and `<@Name@>` references an instantiated definition. Compatible repetitions share one ordered formal. Conflicting declarations reject. Dotted labels follow L identifier rules, independently of Go/Rust keywords and native identifier validation.

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

Native Thing slots accept Text, Number, Truth and Null without converting them to strings. This Log body requires Text for `.as_bytes()`. Ordinary source actions also accept supported collection and Pointer arguments, preserving [value ownership](values.md#conversion-and-aliasing-boundaries).

Type, Atoms and Label are frontend roles using existing AML metadata. Nested constructors bind results once, in operand order, before the consuming word. Composition slots retain their selected definitions. Native and ordinary Functions share body lowering and bind their own invocation words. Dependencies retain canonical source ownership.

Unknown kinds, incompatible types, missing bindings, invalid native names and malformed compositions reject. Text receives protected native quoting around AML/logeme delimiters. Explicit native fragments retain target-dependent meaning.

## Composition and source ownership

`^` integrates a child, `*` combines siblings and parentheses group compositions. Assembly order does not promise universal runtime evaluation order. Func/Prop/Op records, imports and children retain their source-file owner. Classes cannot inject defaults into Objects.

Glyph-local Lifetimes/GenericTypes accept native String-list interiors only on FW/HY/MY declarations and Y/named-KY applications, with their explicit absence/blank rules. They are not L generics, Func/Prop/Op fields, inferred ownership or inherited settings. Rust rejects Deferred metadata, even false. Scheduling implies neither scope-exit cleanup nor hidden continuations.

<!-- example:definitions-objective -->

Objective's nested array preserves one word containing A and Function. It is native-composition grouping, not an L collection literal. [Grouping tests](../../lyger/definition_test.go) check independent instances, not arbitrary runtime ordering.

## Sources and witnesses

Sources: [core](../../builtins/eng/core.l), [parser/encoder](../../lyger/definition_bindings.go), [argument kinds](../../lyger/definition_kinds.go), [construction](../../lyger/construction.go), [actions](../../lyger/instructions.go), [lexical forms](../../lyger/lexical_types.go), [Rust inputs](../../builtins/README.md). Witnesses: [composition](../../lyger/definition_test.go), [framing](../../lyger/definition_framing_test.go), [slot names](../../lyger/definition_slot_names_test.go), [actions](../../lyger/action_definition_test.go), [descriptor identity](../../lyger/paragraph_descriptor_identity_test.go), [local definitions](../../demo/paragraphs/local-definitions.l).
