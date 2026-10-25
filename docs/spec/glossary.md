# Glossary

These terms describe L and its implementation. They should not be replaced automatically with superficially similar terms from other programming languages.

## Source and meaning

| Term | Meaning |
| --- | --- |
| Lexicon | A selected natural-language grammar and configured vocabulary; English `eng` is the bundled choice |
| Lexeme | A word definition with grammatical role, inflections, and possibly source/native behavior |
| Token | One source occurrence, retaining spelling, role, provenance, and binding |
| Sentence / statement | Source grammatical unit terminated by the implemented punctuation rule |
| Clause | An action and its owned subject/object/modifiers within a statement |
| Phrase | A grammatical relation, commonly introduced by a preposition |
| Subject | The source entity being declared or acted upon; pronoun antecedents use defined subjects |
| Predicate | The action/condition portion associated with a statement |
| Object | An operand or related entity in a grammatical action |
| Paragraph | Physical source grouping separated by blank lines outside literals/comments |
| Lexical scope | An environment owning name bindings; distinct from physical layout and activation |
| Activation | A particular runtime execution context with its own planned local storage |
| Descriptor | A source adjective definition constraining nominal targets and supplying construction/wrapper rules |
| Alias | A binding edge preserving canonical identity of another declaration |
| Source unit | A builtin, program, or dictionary ownership boundary |
| Source occurrence | A particular original or included file occurrence with import ancestry |
| Dictionary | Either a collection/namespace value or an indexed vocabulary source unit; context distinguishes them |

For these terms in actual L, read the [annotated basic program](grammar.md#read-a-complete-sentence-in-context). In `The App Prints message`, App is the grammatical subject and message is the grammatical object. A capitalized **Object** source category instead denotes the file-owning construction shown in [the Class/Object example](collections-and-categories.md#compose-an-object-inside-nested-classes). A [paragraph example](names-and-scope.md#paragraph-scope) shows why physical paragraphs and lexical scope must also be distinguished.

## Application organization

| Term | Meaning |
| --- | --- |
| App | Application value and executable process target |
| Service | Source organization within App construction; standalone declarations create no package or process |
| Class | Recursive module/source group, not a process or paragraph |
| Object | Source-file unit, not an automatic runtime object scope |
| Main | Canonical builtin descriptor selecting the bounded runnable entry construction |
| Basic | Builtin descriptor composing an App around a Main Service |
| Process | Source-defined composition of supplied operation/body syllables |
| Function | Source-defined callable composition; distinct from native Func metadata |
| Orchestrator | Generated entry executable controlling target processes and outer source instructions |
| Invocation | One accepted launch attempt and its immutable final result/continuation ownership |
| Continuation | Paragraph-associated effects activated once by the first Await of an invocation |
| Output | Finalized raw captured stdout of the selected invocation |
| Error | Selected process operation/invocation failure, separate from compiler lerrors |

The [App example](collections-and-categories.md#create-and-execute-an-app) puts Basic, Execute, and Output in context. Compare it with the [explicit App lifecycle example](collections-and-categories.md#run-an-app-and-read-its-output) and [execution chapter](execution.md): standalone Service declarations create no package or process, and awaiting a launched target owns its finalized result and continuation.

## Compiler and backend

| Term | Meaning |
| --- | --- |
| L | The source language and its Go implementation/tooling |
| Lyger | Frontend compiling L through bindings and typed plans into AML/logeme markup |
| AML | Assembly Markup Language; structure/settings/content markup, not machine-code assembly |
| Observable AML | Frontend inspection view retained alongside backend-ready framing |
| Canonical backend AML | Framed input supplied to the backend without post-hoc source rewriting |
| Logiceme / logeme | Go assembler/source-generation implementation in `logiceme/` |
| Executable | Backend API source bundle; the object is not a running application |
| Lucid | Template language interpreting native source fragments |
| Logeme / segment | One alphabet unit with logical ownership and target template |
| Syllable | A vowel-centered composition with permitted optional determinants/qualification |
| Sequence | Syllables combined/integrated/grouped using `*`, `^`, and parentheses |
| Glyph | Backend representation of a selected syllable and its children |
| Logeme system | One target's segment templates and supported composition rules |
| Meta | Metadata supplied to a glyph; allowed fields depend on the concrete form |
| Func / Prop / Op | Native metadata records for functions/properties/operations, not new L types by implication |
| Conjoinment | Typed child-role input to a native composition |
| Generation | Producing source/artifacts; does not establish native compilation |
| Native compilation | Building generated sources; does not establish runtime behavior |
| Runtime witness | Actual supervised execution with observed output/lifecycle behavior |
| Acceptance | The repository's unfiltered gate and protected-input checks for one revision/host |

## Logical alphabet

The target-independent model pairs B/A declare/declaration, D/E calculate/calculation, G/I evaluate/evaluation, F/O define/definition, H/U describe/description, K/W operate/operation, P/X iterate/iteration, M/Y instantiate/instantiation, and C/V execute-asynchronize/execution-asynchrony. These meanings guide composition; only the concrete frozen forms listed in the Rust reference are operationally supported.

## Sources

Definitions follow [architecture](../architecture.md), [source contracts](../../lyger/AGENTS.md), [logeme model](../../builtins/LOGEMES.md), [AML](../../logiceme/AML.md), and [Rust reference](../../builtins/README.md). The glossary is explanatory; its short definitions do not broaden a supported interface.
