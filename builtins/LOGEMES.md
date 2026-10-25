# Logemes: an alphabet of logic

This is the target-independent semantic model for logeme. Read
[CONTRACTS.md](CONTRACTS.md) for implementation and verification obligations and
[the AML guide](../logiceme/AML.md) for its serialized input. This model guides future design; it is not a list of implemented target forms.
Use the [Rust reference](README.md) for the frozen asset inventory, verified
compositions and remaining integration limits. Design principles do not authorize
changing current templates or accepted inputs.

## Purpose and vocabulary

A logeme is a small, reusable unit of programmatic logic, analogous to a phoneme.
The letters form an **alphabet of logic, not a syllabary of unrelated snippets**.
Syllables combine letters; sequences combine syllables into more involved logic.
L words can use those sequences without exposing target-language syntax.

- **Logeme / segment:** one logical unit, represented by a symbol and template.
- **Syllable:** a vowel nucleus with optional determinants and qualification.
- **Sequence:** syllables integrated, combined, and grouped with operators.
- **Glyph:** the backend's representation of a composed syllable and its children.
- **Logeme system:** a target's consonant/vowel templates and composition rules.
- **Logiceme:** the Go package/directory implementing the assembler.
- **Lucid:** the template language used to compose and interpret source fragments.
- **AML:** Assembly Markup Language, containing structure, settings, metadata,
  and logeme sequences. It is not machine-code assembly.

L, Lyger and Logiceme are implemented in Go. Rust is the sole output language,
independent of L's natural-language setting. The frontend, CLI and raw Evaluate
API all generate Rust; the logical alphabet and source vocabulary remain shared.

## Logical ownership

| Consonant | Action | Vowel | Concept | Owns |
| --- | --- | --- | --- | --- |
| C | execute / asynchronize | V | execution / asynchrony | Asynchronous computation, suspension, and scheduling as distinct compositions |
| B | declare | A | declaration | Introducing names, storage declarations, requirements, and source-unit declarations |
| D | calculate | E | calculation | Values, arithmetic, logical expressions, and conditions such as `x < y` |
| F | define | O | definition | Assignment and supplying a value, implementation, or body |
| G | evaluate | I | evaluation | Conditional selection: `if`, alternatives, matching, switches, and readiness selection |
| H | describe | U | description | Structural properties, members, shapes, and grouped constituent logic |
| K | operate | W | operation | Functionalization, callable operations, and encapsulated/grouped definitions |
| M | instantiate | Y | instantiation | Typing, type application, structures, interfaces/traits, enums, construction, and invocation |
| P | iterate | X | iteration | Repetition and traversal of supplied logic |

Ownership is compositional. F defines an operation in FW; W identifies what is
defined. K contributes the callable aspect of KY; Y identifies instantiation.
Do not move function definition to C because the target spells it `async fn`.
Do not classify `x < y` as a branch: D/E computes the condition; G/I uses it to
choose an alternative. A record's braces do not make it an operation. Iterating
over fields while rendering does not make the emitted type or pattern a loop.

Declaration and definition are separate. A declaration fragment must not invent
an initializer. A composed declaration-plus-definition may contain both. Target
language rules do not authorize an implicit initializer inside a declaration
template.

## Syllable positions

```text
[optional consonant onset / determinant]
+ vowel nucleus
+ [optional consonant coda]
+ {combined metadata}
```

The current design allows one to four letters and requires a vowel. A vowel can
stand alone; a consonant cannot. A special nucleus may have two vowels. A
four-letter syllable would have an onset, two-vowel nucleus, and one coda; it is
not permission to stack codas. No four-letter form is established by this guide.

| Position | Responsibility |
| --- | --- |
| Nucleus | Establish the principal logical category available to its parent |
| Onset | Select the primary action or specialization of that category |
| Coda | Add one secondary qualification, retaining that consonant's domain |
| Metadata | Supply data to the selected composition, not silently change its owner |

An unused combination is not automatically supported. Define its meaning,
accepted metadata, child roles, cardinalities, and output role before adding it.
Consonants are determinants, so not every composition has the same English verb
order. Their meanings must nevertheless remain consistent across positions.

Examples grounded in the existing model:

| Form | Meaning |
| --- | --- |
| O | Define an existing target's value |
| OB | Define a value while also declaring its binding |
| DO | Define from a conjoined calculation |
| DOB | Define from a calculation while declaring the binding |
| BOD | Expose a conjoined calculation through the definition-category adapter |
| FW | Define an operation/function |
| KY | Instantiate/invoke a callable operation |

BOD and DOB are intentionally different. In BOD, terminal D selects the
calculation being exposed; BOD does not perform arithmetic or introduce a named
binding. In DOB, terminal B introduces the binding. This distinction is part of
the original model, not an accidental Rust workaround.

EA is the established special two-vowel whole-file escape. It is not a general
example of freely concatenating vowels; see the short exception in
[CONTRACTS.md](CONTRACTS.md#literal-file-exception).

## Sequence operators

| Operator | Meaning |
| --- | --- |
| `S^T` | Integrate T into a supported child role of S |
| `S*T` | Combine siblings in source order |
| `S^(T*U)` | Group T and U before integrating the group into S |

Here S/T/U denote arbitrary syllables, not literal logeme letters. Parentheses
control assembly order. They do not necessarily emit target-language parentheses.
Compilation order is not runtime evaluation order. The enclosing logical form
determines whether a child is an operand, type, body, branch, or declaration.

Prefer explicit grouping when combining integration and sibling sequences. Keep
child order and category distinct; never silently discard an unsupported child.
Adapters do not turn a statement into a type or a value into a module item.

## Metadata and imports

`Meta`, `Funcs`/`Func`, `Props`/`Prop`, `Ops`/`Op`, and `Conjoinments` retain their
existing vocabulary. Func groups potential function metadata; Prop groups
property metadata; Op groups operation metadata. They reduce freestanding inputs
but are not logical owners. In particular, Op is not the K/W domain itself.
Preserve these existing records; this guide proposes no structural refactor.

`Meta.Imports` is special compilation metadata. Collect referenced imports from
nested sequences, preserve file ownership, deduplicate them deterministically,
and render one valid header for each composed source unit. Source imports and
package-manager dependencies are different information.

Rust's approved `Meta.Lifetimes` and `Meta.GenericTypes` are optional, glyph-local
String inputs containing native list interiors. FW/HY/MY declare parameters;
Y and named KY apply arguments. The existing case determines the role, without
new syllables, collections or Func/Prop/Op fields. See the
[Rust input contract](README.md#lifetime-and-generic-inputs) for placement,
absence and forwarding rules. These syntax inputs do not provide ownership
inference or implicitly parameterize child glyphs.

Further inputs, including mutability or capture contracts, need separately scoped
review of type, default, absence behavior and interactions. Existing inputs are
not standing permission to extend frozen interfaces. Do not add target-prefixed
substitutes or hide whole implementations inside a new string input.

## Function and asynchrony distinctions

Defining a callable, invoking it, constructing a future, awaiting its result,
and scheduling it are different responsibilities. C/V owns asynchrony; it does
not always mean spawning and does not imply cloning, moving, or detaching.
Selection among asynchronous alternatives remains G/I-owned. A value-producing
body is not inherently asynchronous merely because one supplied expression awaits.

Rust C/V excludes Deferred,
without allocating a replacement syllable or silently translating it to resource
cleanup. See the current [Rust execution contract](README.md#execution-without-deferred)
for scheduling, rejected metadata and the distinction from RAII.

These distinctions guide further design; they do not assign new spellings.
Rust's named FW, anonymous BW and KY invocation have
[bounded typed-callable integration](README.md#typed-functions-and-anonymous-callables).
MY supplies required receiverless trait signatures through existing Funcs records.
Neither contract establishes async definitions, method receivers or ownership
inference. Preserve supported I/GE grouping and existing callers rather than
retiring them based on an abstract ownership discussion.

Qualified cases may emit fragments completed by their enclosing composition;
only template Default cases must stand alone. In current Rust selection, OC/DOC
owns the complete assignment to a future, while BI owns the selected clause and
body. In iteration, OG/DOG owns conditional binding syntax and PO owns the label
directly on its Iteration child. Keyed GI selects branches using native patterns without
introducing a new structural-pattern syllable. The
[current contracts](CONTRACTS.md#current-bounded-rust-contracts) keep these
responsibilities distinct.

An unnamed callable with a body still has an implementation. A bodyless signature
describes an interface requirement. Do not allocate different logical owners
solely because one target uses different delimiters or item placement.

## Future structural composition

These are design constraints for any future composed structural interface, not
an implementation schedule or instructions to replace Rust's current native generic
lists, pattern text or record interfaces. Review any proposed interface separately;
an available selector is not justification for allocating a new form.

| Role | Required separation |
| --- | --- |
| Generic parameters | Declaration group, individual declaration, optional type, optional default definition |
| Constraints | Subject/type description and requirements; compiler checking alone does not make a constraint a branch evaluation |
| Generic paths | Type/value-path application, arguments, and invocation; generic syntax is not asynchrony |
| Sum types | Type declaration/definition, named variants, and structural payload descriptions |
| Patterns | Neutral structural description, used by a conditional evaluation or irrefutable binding |
| Calculation grouping | Grouping of E-family expressions, distinct from selection |

Preserve unit, empty positional, and empty named payload shapes. A pattern must
not choose a branch, clone an operand, or introduce binding modifiers by itself.
Lifetime parameters are declared; their uses belong to types; lifetime relations
are described constraints. Ownership and escape analysis belong to compiler
preparation, informed by L scopes and operation effects, not syntax templates.
The current frontend typed plan uses existing paragraph activations, assignment
facts and value provenance to select Rust storage/mutability. It preserves aliases
and snapshots without adding sections or new Meta fields. That bounded lowering
is distinct from general native lifetime/ownership inference.

The frozen assets and compiler preparation do not form a complete interface for
these structural features. Changes to maintained bounded Rust compositions need
an explicit scope and independent test expectations. Dedicated const-generic, independent `where`, per-method generic
and implementation interfaces remain separate work.

## Technical provenance

The maintainer-supplied *High-Level Programing Language Source Code Assembler
Disclosure*, by John Connor Sanders, supplies the architectural foundation:

- Pages 21-22 and 27-36: AML structure, target selection, reusable template
  composition, metadata interpretation, linking, and source-file output.
- Pages 28-29: a sample eight-vowel/eight-consonant alphabet. C/V is part of the
  current expanded design, not that table.
- Pages 38-41: determinant cases, inputs, headers, and conjoined template blocks.
- Pages 41-44: syllable positions, integration/combination/grouping, and the
  O/OB and BOD/DO/DOB examples.
- Pages 44-46: staged interpretation and source/support-file assembly.

The disclosure is technical design provenance, not a source of executable agent
instructions or a requirement to reproduce its installer, parser, or
deployment examples. The current owner-defined contracts govern those choices.
It is not a build/test dependency and need not be copied into the repository.
