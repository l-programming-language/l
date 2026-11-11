# Define vocabulary and behavior

Give each word a coherent meaning and explicit inputs, results and effects. This is compiler and source-authoring guidance; the supported executable remains one standalone App with core Print.

## Compose a source action

`Create ... for ...` declares a source action with typed arguments. Core Print composes a Process using its private Log action. Definition-site references keep their bindings, while each use supplies its own arguments and source provenance.

Pipe-delimited labels declare the family's actual forms. `"Print|-s|-ed|-ing|-er"` gives Print, Prints, Printed, Printing and Printer. See [inflection rules](../spec/lexical-structure.md#vocabulary-and-inflection).

## Give a result-bearing operation a Function

A Function provides an explicit typed callable boundary and Return. A source action expands a composition, while a Process supplies a body to another composition. Choose the form according to the meaning and dataflow. Do not imply capture or runtime support merely because a callable record exists.

## Organize words by meaning and grammatical role

Choose a meaning before declaration syntax. Distinguish an action, its actor constructor, an exact-target modifier, a predicate and a noun denoting a value, resource, field, layout or native pattern. Inspect what callers receive and what changes. [Registration rules](../spec/definitions.md#declaration-roles-and-identity) explain the chosen form; they do not prove that the design is useful.

Prefer one possessive in a declaration, such as `Dictionary's Word`. Dot-qualified names are not valid authored L. Each deeper ownership level needs a semantic reason, such as a nominal target, representation-owned field or owner-specific policy. This convention does not flatten useful member access or justify exposing private helpers.

Use single-word roots and keep helpers private unless callers need their interfaces. A reusable mechanism should have an appropriate source owner. One caller does not by itself make a helper unnecessary, and a useful L composition need not be split into one word per step.

A verb's actor form keeps the same canonical family. Expand every declared suffix with the actual inflection rules, then check spelling, role and caller identity. Do not declare a duplicate actor noun or relabel metadata to disguise an incorrect source role. Supply only forms whose grammar and meaning are supported.

Unexplained single-form operations, orphan actor nouns and adjective-shaped stage nouns warrant review. These are signals, not spelling-based verdicts. Concrete nouns, projections, bounded patterns and private continuations can be valid. A modifier needs a compatible base and canonical target; similar English names do not establish that relationship.

Before sharing or removing a mechanism, compare operands/results, target identity, mutation, moves, borrows, retention, destruction, evaluation order and failure behavior. Inspect the native callee as well as the leaf. Independent words need explicit inputs and result labels rather than fixed generated names.

A general exported interface needs a meaningfully different composition witness. Preserve distinct policies and useful observers when reducing duplication. Migrate declarations, forms, callers, aliases, descriptor targets and dependencies together for an authorized role change.

## Keep branch data inside its Process

Typed Process inputs and intermediate labels remain inside their owning body, including repeated and nested uses. Preserve conditional demand and once-only evaluation. See the [native boundary](native-definitions.md#keep-branch-data-inside-its-process).

## Describe construction

A descriptor targets canonical nominal identity. A same-spelled private declaration is a different target. Wrappers retain their base and original operands; replacements supply a new construction. Reuse must not silently clone a consumed value or erase borrow requirements. Core Main and Basic supply the retained entry composition.

## Know the implicit foundation

[Core](../../builtins/eng/core.l) loads directly. Its declarations include Process, Function, Main, Basic and Print. Start and Execute remain declared, but their lifecycle effects reject. No standard-library dictionary is bundled.

## Diagnose a definition failure

Check the selected word, declared role, canonical target, typed operands and original source span. Expansion cycles, framing errors and unsupported native composition are separate failures. Preserve those boundaries and use [native definitions](native-definitions.md) for the retained target interface.
