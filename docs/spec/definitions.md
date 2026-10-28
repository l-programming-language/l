# Definitions and extensibility

L extends vocabulary through declarations, source actions, descriptors and native-backed definitions. Typed arguments and source ownership constrain each mechanism; arbitrary English is not executable.

## Create, Define, and Describe

Create introduces a supported value or lexical action. Define supplies a scalar, runtime value or native-backed definition. Describe associates an adjective with nominal targets and constructor/wrapper rules. Active/passive grammar and resolved ownership determine the declaration site.

Declarations retain tokens, formal identities, bindings, descriptor targets and source spans. Publication validates relationships before committing; expansion creates use-specific instances without mutating prototypes.

## Source actions

<!-- example:lexical-action-inflections -->

`any App` declares a typed argument; `it` refers to that argument. Start expands Build/Run; Execute expands Start/Await. Neither is a primitive instruction root.

Calls check argument count and nominal type. Coordinated operands expand the complete body for each operand in order. Definition-site references retain their bindings; actual arguments and diagnostics belong to the call. Expansion cycles reject rather than supplying runtime recursion.

## Inflection declarations

Quoted labels list bar-separated forms. Nouns use root/plural/descriptive/active positions; verbs use root/active/passive/applicative/objective positions. One leading hyphen appends a suffix; additional hyphens remove trailing root bytes first. See [lexical rules](lexical-structure.md#vocabulary-and-inflection).

`Start|-s|-ed|-ing|-er` produces Start, Starts, Started, Starting and Starter. `Execute|-s|-d|--ing|--or` trims the final e for Executing/Executor. Grammatical roles can change while ownership remains canonical; inflection does not publish private helpers.

## Descriptors and constructors

Descriptors bind nominal identities to ordered child-construction or wrapper rules. Same-named private declarations remain distinct; aliases preserve identity. Missing target constructors and recursive expansion reject.

<!-- example:local-definitions -->

The two paragraphs own separate Launch actions, Compact descriptors and captured messages. New selects construction, Generic participates in source composition and Public controls visibility; none infers Rust lifetimes or generic arguments.

Main constructs a Function inside an Object inside a Service; Basic supplies that Main Service to an App. These rules define `New Basic App`.

## Process, Function, and Return

<!-- example:definitions-callable-core -->

Process composes ordered Syllables through W. Function supplies FW/Func metadata: Name, typed Inputs, output types and an instantiated Process. Source Function declarations use this selected definition too; Function-local Return binds its builtin member after checking source results.

<!-- example:main-descriptor -->

Each `It describes` refines Main. App construction selects its Service, Object and Function rules. Main supplies no inputs or outputs; [ordinary Functions](applicational-types.md#functions-properties-and-operations) supply their declared signatures and bodies. A standalone Main Service remains structural; methods, overloads and variadic signatures are not introduced.

## Framing and escaping

Source definitions use backtick Text with one complete outer `<DEF>…<DEF>` pair. Both markers are spelled `<DEF>`; neither is an XML closing tag. Interior/incomplete markers remain ordinary Text. Empty or nested definition bodies reject. The framing helper can inspect quoted/unquoted frames; explicit internal template parsing also accepts markerless interiors. Those internal interfaces do not remove source backticks.

<!-- example:native-dictionary-definition -->

Log defines a native body; ProbeIO.Print composes its Logger form. App-body callers supply Output. `BOD^E` integrates the Calculation, `{|Statement: …|}` supplies metadata and `~…~` encloses native statement text. See the [native-definition guide](../guide/native-definitions.md) for its caller.

A body may be one expression, a JSON string array representing a word, or an array of word arrays. Grouping survives instantiation. Quote, backslash and backtick escapes decode once; native escapes such as `\n` survive, and arguments are not decoded again.

## Typed slots

`<@Name|Type@>` declares a typed slot; `<@Name@>` references an instantiated definition. Compatible repetitions share one ordered formal; conflicting declarations reject. Dotted labels follow L identifier rules, independently of Go/Rust keywords and native identifier validation.

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
| Definition | Instantiated definition with identity and grouping |

Thing slots bind supported scalars without converting everything to strings; this Log body specifically requires Text for its Rust `.as_bytes()` call. Changing Thing to Props would request property records, not parse Text into them. Ordinary stored source-call arguments currently lower Text, Number and Truth; the public Go DefinitionArgument interface supports the wider table.

Unknown kinds, incompatible types, missing bindings, invalid native names and malformed compositions reject. Text receives protected native quoting around AML/logeme delimiters; explicit native fragments retain target-dependent meaning.

## Composition and source ownership

`^` integrates a child, `*` combines siblings and parentheses group compositions; assembly order does not promise universal runtime evaluation order. Func/Prop/Op records, imports and children retain their source-file owner. Classes cannot inject defaults into Objects.

Glyph-local Lifetimes/GenericTypes accept native String-list interiors only on FW/HY/MY declarations and Y/named-KY applications, with their explicit absence/blank rules. They are not L generics, Func/Prop/Op fields, inferred ownership or inherited settings. Rust rejects Deferred metadata, even false; scheduling implies neither scope-exit cleanup nor hidden continuations.

<!-- example:definitions-objective -->

Objective's nested array preserves one word containing A and Function. It is native-composition grouping, not an L collection literal. [Grouping tests](../../lyger/definition_test.go) check independent instances, not arbitrary runtime ordering.

## Sources and witnesses

Sources: [core](../../builtins/eng/core.l), [parser/encoder](../../lyger/definition_bindings.go), [argument kinds](../../lyger/definition_kinds.go), [construction](../../lyger/construction.go), [actions](../../lyger/instructions.go), [lexical forms](../../lyger/lexical_types.go), [Rust inputs](../../builtins/README.md). Witnesses: [composition](../../lyger/definition_test.go), [framing](../../lyger/definition_framing_test.go), [slot names](../../lyger/definition_slot_names_test.go), [actions](../../lyger/action_definition_test.go), [descriptor identity](../../lyger/paragraph_descriptor_identity_test.go), [local definitions](../../demo/paragraphs/local-definitions.l).
