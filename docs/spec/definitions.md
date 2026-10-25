# Definitions and extensibility

L extends its source vocabulary through declarations, source actions, descriptors, and explicitly native-backed definitions. Extension preserves typed arguments and canonical source ownership; it does not make arbitrary English sentences executable.

## Create, Define, and Describe

Create introduces a supported value or a lexical action. Define supplies a scalar, runtime value, or native-backed definition. Describe associates an adjective descriptor with a nominal target and a constructor/wrapper rule. Passive declarations and active descriptor forms can express related operations, but grammar and resolved ownership determine the exact site.

An existing declaration is not merely a macro name for text substitution. The compiler retains source tokens, formal argument identities, bindings, descriptor targets, and the original source span. It prepares relationships before publication and expands use-specific instances without mutating the prototype.

## Source actions

The implicit Start and Execute declarations are actual source:

<!-- example:lexical-action-inflections -->

This included canonical source describes the builtin definitions in `builtins/eng/core.l`; the site does not run them as examples. `any App` establishes a typed formal argument. The pronoun it refers to that action argument. Start expands to Build then Run. Execute expands to Start then Await. Neither is a primitive lexical instruction root.

Source action calls check argument count and nominal type. A single-argument action applied to coordinated operands expands the complete body for each operand in source order. Recursive action expansion is a Cycle failure; this mechanism does not establish general runtime recursion. The action's static references remain bound to its definition, while actual arguments and call-site diagnostics belong to the invocation.

## Inflection declarations

Quoted labels may contain bar-separated forms. Nouns use root/plural/descriptive/active positions; verbs use root/active/passive/applicative/objective positions. A leading hyphen appends a suffix, and additional hyphens trim trailing root bytes before appending. The detailed lexical rule is in [lexical structure](lexical-structure.md#vocabulary-and-inflection).

Inflection can change grammatical role without changing canonical ownership: Process → Processing; Function → Functional; Log → Logger. An imported alias is a deliberate namespace operation; a private helper's inflection does not become a public export by resembling a useful word.

In the two declarations above, `Start|-s|-ed|-ing|-er` records `Start`, `Starts`, `Started`, `Starting`, and `Starter`. `Execute|-s|-d|--ing|--or` uses the trim form to produce `Executing` and `Executor`, rather than appending those endings to the final `e`. The source action uses the resulting participles `Starting` and `Awaiting`; they still resolve to their canonical actions.

## Descriptors and constructors

A descriptor records which nominal identities it describes and the ordered rules supplying child constructions or wrappers. A rule targeting one private nominal declaration does not accidentally target a different declaration with the same spelling in a sibling scope. Aliases preserve the canonical descriptor and target identities.

Main is declared as an adjective for Service, Object, or Function. Its rules construct a `main` Function from a Process with No Inputs and No Outputs, a `main` Object around an Objective Main Function, and a `main` Service around a Main Object. Basic describes a Generic App with a Main Service. This is the source origin of the common `New Basic App` form.

The maintained local-definitions program demonstrates two private Launch actions and Compact descriptors without letting the second paragraph capture the first paragraph's message or definitions:

<!-- example:local-definitions -->

This is an existing runnable witness, not a new execution result. A descriptor without a constructor for the requested target rejects. Recursive descriptor expansion rejects. New, Generic, and Public affect different concerns: construction, source composition, and publication. None creates an inference system for Rust lifetimes or generic arguments.

## Process, Function, and Return

Process composes supplied Syllables into operation W. Function composes FW with Func metadata containing Name, Inputs, Outputs, and a Process body. Objective wraps a Function into its source grouping. Function.Return emits a native return statement through BOD/E.

Main's demonstrated source function signature has no inputs and no outputs. A nonempty signature requires explicit property values; ordinary text spelling “with Inputs” is not sufficient. Typed native Func inputs/outputs have their own bounded reference. General closures, methods, overloaded runtime dispatch, async-function declarations, or variadic source functions are not implied by Function's existence.

Return's Statement slot is typed Text at the native-definition boundary. Its target behavior must be understood from the emitted native fragment, not assumed to provide a universal L expression-return API. The maintained tests cover concrete compositions, and broader callable behavior requires a separate supported source form.

**Implicit definitions — reference excerpt, not a standalone program:**

<!-- example:definitions-callable-core -->

`<@Definitions|Syllables@>` is Process's ordered collection of composed expressions. Function binds a textual Name, typed property Inputs, native output Texts, and an already instantiated Process. Its Return member supplies the Rust `return …;` statement. The next descriptor excerpt shows the actual no-input/no-output construction used for a Main entry:

<!-- example:main-descriptor -->

Here the later `It describes` sentences refine Main, the adjective declared by the first sentence. Within App construction, Main selects the Service rule and constructs its Object and Function. A standalone `New Main Service` remains a structural declaration without generated contents. This does not demonstrate an arbitrary nonempty callable signature or ordinary L method dispatch.

## Framing and escaping

A native-backed source definition uses a backtick literal with paired `<DEF>` markers. The Process declaration above encloses `W^<@Definitions|Syllables@>` in that frame. The retained ProbeIO fixture shows the same framing around a native-backed operation used by App bodies:

<!-- example:native-dictionary-definition -->

`Log` binds an Output argument through the Thing slot. ProbeIO.Print composes a Logger from that definition, and App body actions supply the value. The `BOD^E` composition and Rust stdout write are explicitly native content. This fragment defines vocabulary; an App using it supplies the executable body.

Both markers are spelled `<DEF>`; the final marker is not an XML closing tag. Nested markers and empty definitions reject. This text is parsed by L's definition and logeme framing rules, not by a browser or generic XML parser.

A body may be one composed expression, a JSON array of strings representing a word, or a JSON array of such arrays representing grouped words. Explicit grouping survives into the DefinitionInstance. Source quote/backslash escapes for quote, backslash, and backtick are decoded once in this boundary. Native escapes such as `\n` remain native text; argument values are not decoded again.

## Typed slots

Slots use `<@Name|Type@>`; an untyped `<@Name@>` references an already instantiated definition rather than defaulting to Text. Repeated compatible slots keep one ordered formal declaration; conflicting declarations reject. Dotted slot names follow L identifier rules and are not filtered by Go/Rust keyword lists.

| Binding kind | Boundary and data |
| --- | --- |
| Text | Unquoted scalar bytes, safely encoded for native use |
| Number | Parsed finite binary64, emitted with native `f64` typing |
| Truth | Typed boolean scalar |
| Null | Explicit local native display atom, not a null pointer |
| Syllable | An already instantiated expression |
| Syllables | Ordered expression collection with preserved cardinality |
| Props | Typed native property records |
| Texts | Ordered native type/output text collection |
| Definition | A previously instantiated definition with identity and grouping |

Thing slots accept supported scalar argument kinds without coercing them all to strings. Unknown kinds, incompatible types, missing bindings, invalid native names, and malformed compositions reject. The public Go DefinitionArgument interface supports more binding forms than ordinary stored L scalars; current stored source-call arguments lower Text, Number, and Truth.

The emitter escapes native literal content around AML/logeme delimiters. It does not concatenate untrusted Text as native source. Deliberately native fragments remain opaque leaves whose validity depends on their selected target and complete composition.

In the Emit example, `<@Message|Text@>` declares the slot's type; the backtick literal at the call site is its value. Contrast this with `<@Process@>` in Function: that untyped-looking reference expects the already instantiated Process definition. Changing `Text` to `Props` would request a different binding interface, not parse an ordinary Text into properties. Nonempty native signature records remain a bounded advanced interface; the empty Main signature is the supported construction demonstrated here.

## Composition and source ownership

Logeme integration `^`, sibling combination `*`, and parentheses describe assembly. Their ordering is not a blanket runtime evaluation order. Func/Prop/Op records, imports, and child roles retain the owning source file; a Class cannot silently inject defaults into every Object.

Glyph-local Lifetimes and GenericTypes inputs exist only on approved FW/HY/MY declaration and Y/named-KY application forms. They contain native String list interiors, with explicit absence/blank rules. They are not L generic declarations, Func/Prop/Op fields, inferred ownership, or inherited class settings.

Rust excludes Deferred and rejects supplied Deferred metadata, including false. Native scheduling does not imply scope-exit cleanup or a hidden continuation language. See [the concrete target contract](../../builtins/README.md#execution-without-deferred).

The builtin Objective descriptor supplies a grouped word rather than an unstructured string of glyphs:

<!-- example:definitions-objective -->

Its nested array retains one word containing `A` and the supplied Function definition. This is L source declaring a native composition, not an L collection literal or an executable JSON value. The [definition grouping tests](../../lyger/definition_test.go) check retained grouping and independent instantiations; they do not assign runtime evaluation order to every possible composition.

## Sources and witnesses

Sources: [builtin definitions](../../builtins/eng/core.l), [definition parser and encoder](../../lyger/definition.go), [argument kinds](../../lyger/definition_kinds.go), [construction](../../lyger/construction.go), [action expansion](../../lyger/instructions.go), [lexical forms](../../lyger/lexical_types.go), and [Rust target inputs](../../builtins/README.md). Witnesses: [definition composition](../../lyger/definition_test.go), [framing](../../lyger/definition_framing_test.go), [slot names](../../lyger/definition_slot_names_test.go), [action definitions](../../lyger/action_definition_test.go), [descriptor identity](../../lyger/paragraph_descriptor_identity_test.go), and [independent local definitions](../../demo/paragraphs/local-definitions.l).
