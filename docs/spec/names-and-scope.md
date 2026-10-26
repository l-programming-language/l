# Names, scope, and evaluation state

L resolves names before emitting Rust. Paragraph boundaries determine lexical scope; runtime activations determine when values are initialized and retained.

## Declarations and definitions

Quoted labels declare names. `Create` introduces a category or action, `Define` supplies a value or native definition, and `"name" is value` declares or assigns. The grammatical site and resolved value determine their meaning.

A statement publishes its declarations together, preserving separate coordinated targets and source order. Failure leaves existing bindings intact. Duplicate names in one namespace reject; private scopes may reuse spellings. Dependency retries do not reorder execution or pronouns. Forward local names reserve their spellings before lookup considers outer names or builtins.

<!-- example:binding-forward-member -->

Both indented `Show App's Label` occurrences bind the local Label, including the earlier occurrence. The final Show binds the outer Label. This binding witness does not authorize reading an unassigned runtime value: [definite assignment](#runtime-storage-and-definite-assignment) is checked separately.

## Paragraph scope

An ordinary paragraph uses its program or dictionary's environment. Four spaces before its first code line create a private scope. Parent bindings remain available unless shadowed; siblings cannot access each other's private declarations.

Public/Publicly publish only the selected declaration into the source unit's root, retaining its identity. They do not expose other declarations or a private owner's namespace. Each visibility modifier must target one declaration in its clause.

<!-- example:local-shadowing -->

Each indented paragraph owns its `message`; the final Show reads the outer one. Expected stdout is `inner7truesecondouter`, without a newline.

<!-- example:binding-public-scalars -->

This private paragraph publishes Greeting, Count and Enabled. Later paragraphs in the [complete example](../../demo/paragraphs/public-values.l) can name them. The excerpt omits the subsequent Dictionary and Show statements.

## Pronouns

It, this and that bind the last subject **defined** in their original source paragraph. References within a statement use its incoming subject; coordinated declarations establish the next subject in source order. Mentioning an object or executing an action does not change it. A new paragraph has no inherited antecedent.

<!-- example:binding-pronoun-alias -->

`Use Worker as "Job"` makes Job the latest subject. `Start it` starts that same App; both Await statements address its invocation, and `Job's Output` reads finalized stdout. Expected output is `canonical\n` once.

<!-- example:binding-pronoun-boundary -->

This example rejects: `Await it` begins a paragraph without a subject. Use Worker explicitly and launch it before Await; changing the pronoun alone does not launch it.

Within an action definition, a pronoun may bind a preceding `any` parameter, as in Start's “Building any App and then Running it”. Ambiguous generic antecedents reject. Imports retain their own paragraph origins and cannot replace the caller's antecedent.

## Possessives and member lookup

`Worker's Output` and `it's Output` select a member of the bound owner. Dotted names and imported aliases preserve that ownership: a missing member never falls back to a different outer owner. Members can be words, definitions or values; implicit Output/Error require an App.

<!-- example:file-import-helper -->

`LocalTools's "Greeting"` and `LocalTools's "Display…"` declare vocabulary members. The [importing program](imports.md#explicit-source-file-imports) aliases them before using unqualified names. These are not runtime Map entries.

## Aliases and canonical identity

`Use existing as "Alias"` retains the declaration's App identity, descriptor, members and signature. Imports use this relationship after checking exports. Aliases create no new process slot and confer no Main entry status. Cycles and unresolved targets reject. Dictionary private bindings stay in the dictionary; qualified members require an explicit alias for unqualified use.

## Runtime storage and definite assignment

Assignments write at their source positions. Every path continuing to a read must have assigned the value; exiting branches cannot reach the read. Conditional declarations supply no hidden initial value.

Assignment, arguments, returns and captures snapshot ordinary values and retain explicit Pointer identity. Short reads may borrow storage; [Pointer](values.md#pointer) shares targets without Rust lifetime annotations.

An action or Function declaration retains an environment for its activation. Captured assignments update that environment. Copies and aliases retain it until their last owner is released, even after the original name is rebound. Every capture must be assigned when the declaration executes. Static definitions retain definition-site bindings.

Source-word arguments evaluate once in the caller before entering the callee's environment. List/Map arguments retain their supplied contents. A callee can read an older captured List while receiving Text extracted from the caller's current List.

## Output snapshots

Output requires finalized capture; declaration, Build and assignment do not implicitly Await. Stored Output retains the bytes read from that invocation after reruns. App continuations get fresh local storage per invocation; repeated Await returns the same result without repeating effects. See [execution](execution.md#continuations-and-activation).

<!-- example:repeated-activation -->

Each Run/Await pair activates the indented continuation once, showing its local result twice. The outer snapshot retains the first invocation. Expected stdout is six `Repeat\n` lines; the [counter-based witness](../../testdata/phase1/f08-rerun-keeps-snapshot/case.json) additionally checks differing invocation bytes.

## Diagnostics

Unresolved names, ambiguous inflections, duplicate publication, cycles, inaccessible captures and incompatible assignments retain distinct diagnostics with original spans and expanded call sites. Internal Unknown storage never supplies a fallback value.

## Sources and witnesses

See [scope planning](../../lyger/scope.go), [environments](../../lyger/lexical_scope.go), [member resolution](../../lyger/semantic_resolution.go), [binding](../../lyger/binding.go), [visibility](../../lyger/visibility.go), [publication](../../lyger/publication.go) and [flow analysis](../../lyger/orchestration_flow.go). Tests cover [bindings](../../lyger/paragraph_binding_architecture_test.go), [publication](../../lyger/paragraph_publication_architecture_test.go), [captures](../../lyger/paragraph_capture_architecture_test.go), [activations](../../lyger/paragraph_activation_architecture_test.go), [assignment](../../lyger/orchestration_plan_test.go), [stored output](../../lyger/orchestrator_test.go), [caller arguments](../../lyger/source_action_arguments_test.go) and [last-owner release](../../lyger/runtime_lifetime_test.go).
