# Define vocabulary and behavior

Give each word a coherent purpose and explicit inputs, results and effects. Decompose behavior where a step has a useful independent meaning, not to meet a line limit or add forwarding layers.

## Compose a source action

<!-- example:local-definitions -->

`Create ... for ...` defines Launch. `any App` declares its argument, and `it` refers to that argument. `Launch First` builds, runs and awaits First. Coordinated operands each receive the complete body in order. Source actions retain their [definition-site bindings](../spec/names-and-scope.md#runtime-storage-and-definite-assignment).

`"Launch|-es|-ed|-ing"` declares Launch, Launches, Launched and Launching. See [word forms](values.md#read-pipe-delimited-word-forms). Each paragraph owns its Launch action and Compact descriptor, so sibling definitions remain independent.

## Give a result-bearing operation a Function

<!-- example:composed-functions -->

Identity returns its Text input. Relay names that call's result Nested and returns it. The outer call names Result before Show reads it. The [first-program commands](first-program.md#build-and-run) build and run this complete example, which writes exactly `hello`, without a newline. Relay makes the dataflow visible here. A forwarding word is useful in real code when it adds a meaningful boundary.

Each Function's four-space paragraph owns its inputs and body and ends at the blank line. Publicly exposes the Function, not its local names. A Function invocation has fresh parameters and locals. A source action such as Launch expands its ordered actions. A Process supplies a body to another composition, and an App owns a separate OS process. Choose the form that matches the work.

## Organize words by meaning and grammatical role

Choose a meaning before its declaration syntax. Distinguish an action, its derived actor constructor, a modifier of an exact target, a predicate computing a result, and a noun denoting a value, resource, field, layout or native pattern. Inspect what callers actually receive and what changes. A name that sounds like a property can conceal a mutation or conversion. [Registered roles](../spec/definitions.md#declaration-roles-and-identity) describe how the compiler interprets the chosen form, not whether the design is useful.

Prefer one possessive in a vocabulary declaration: `Dictionary.Word`. Count resolved ownership edges, including dotted and mixed spelling. Each deeper level needs a reason, such as a nominal target, representation-owned field, compiler-selected member or owner-specific policy. This is a design convention, not an implemented compiler warning or a grammar restriction. Ordinary nested member access is not a declaration. Do not flatten useful typed members, invent compound roots, expose private helpers or expand global vocabulary merely to shorten qualification.

Use single-word roots and keep helpers private unless callers need their interface. [Project dictionary imports](imports.md#import-a-project-owned-member) show scoped ownership. A reusable mechanism should not depend on an unrelated domain merely because its first caller lived there. Conversely, one caller does not make a helper unnecessary, and a coherent high-level L composition need not be split into one word per step.

A verb describes an action, and its actor noun constructs that action's definition. IO's `"Send|-s|--t|-ing|-er"` declares Send, Sends, Sent, Sending and Sender as forms of one identity. Lined modifies the canonical Sender target. Reuse that family rather than declaring a second Sender noun. Expand every declared form with the actual [inflection rules](../spec/lexical-structure.md#vocabulary-and-inflection), then check its intended spelling, role and caller identity. Source-to-documentation agreement can faithfully reproduce a malformed family. Supply additional forms only where their meaning and grammar are supported.

Unexplained single-form operations, orphan actor nouns and adjective- or participle-shaped stage nouns warrant review. They are signals, not spelling-based verdicts: concrete nouns, projections, bounded patterns and useful private continuations remain valid. A proposed modifier needs an actual base and canonical target with compatible operands and results. Similar English names do not establish that relationship, and relabeling documentation cannot create it.

Before sharing or removing a mechanism, compare its operand/result types, target identity, mutability, moves, borrows, retention and destruction. Trace demand and evaluation order, returned versus discarded results, and failure policy. Inspect the native callee as well as the leaf: similar bodies can differ in writable storage or ownership, while one short call can conceal an entire policy. Trace intermediate bindings too. Independent words should accept explicit inputs and result labels rather than communicate through fixed generated names.

Exported interfaces that claim generality need a meaningfully different composition witness. Private helpers need no invented public consumers. Preserve distinct policies and behavioral observers when reducing duplication. For an authorized role or ownership change, migrate qualified callers, aliases, descriptor targets and selected-definition dependencies together; retain accurate documentation until the actual source changes.

## Keep branch data inside its Process

A Process supplies a body to a selected branch or iteration. Its inputs and locals stay within that body. The advanced [Require walkthrough](native-definitions.md#keep-branch-data-inside-its-process) shows typed branch inputs and conditional error construction.

## Describe construction

In the Launch example, Compact describes a Generic App with a Main Service. `using any Process` passes the supplied body into that entry structure. `New Compact App` applies the pattern, much like Basic.

Descriptor targets are canonical identities: an App rule cannot describe a Service, and a same-named private declaration is a different target. Generic here selects a construction pattern. It does not declare type parameters.

A wrapper composes around its retained base, while a replacement supplies a new construction. Checked value operands evaluate once in source order, while Process bodies stay within their branch or iteration. Reusing a value does not clone it or erase its borrow contract. The [native guide's Lined example](native-definitions.md#compose-a-wrapper) shows a source wrapper that uses the base result and original writer.

## Know the implicit foundation

The implicit foundation includes [Start/Execute](../reference/word-glossary.md#context-builtin-lifecycle--start-and-execute) and the [core constructions](../reference/word-glossary.md) behind Basic. Extend existing words before adding a native leaf. The [native guide](native-definitions.md) explains that boundary.

## Diagnose a definition failure

Check the selected word, its argument count and types, and the original source location. Return belongs to a Function. Exit terminates the program. Every continuing result-bearing Function path must return. Expansion cycles, malformed native framing and unsupported target compositions are distinct failures.

Continue with [control flow](control-flow.md), [imports](imports.md), and [lifecycle](lifecycle.md) before the advanced [native-definition guide](native-definitions.md).
