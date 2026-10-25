# Names, scope, and evaluation state

L binds source occurrences to declarations before emitting native code. Physical paragraph identity, lexical scope, source-unit ownership, and runtime activation are related but distinct concepts.

## Declarations and definitions

A quoted label introduces a declaration. `Create` constructs or introduces a source category/action; `Define` supplies a value or native definition; passive `"name" is value` is an assignment/declaration form. Their meanings depend on the grammatical site and resolved value, not only the first word.

The semantic pass prepares a complete statement before publishing its relationships. Coordinated declarations retain separate targets and source order. A failed batch must not partially replace existing bindings. Duplicate declarations are publication errors where the same scope/namespace already owns the name; different private scopes may legitimately use the same spelling.

The compiler can retry statements while dependencies resolve. This worklist does not reorder executable instructions or change pronoun antecedents. A forward local declaration reserves its forms before lookup falls back to a builtin or outer name. An unresolved reserved name therefore cannot accidentally bind to an outer definition during an earlier retry.

**Binding example — frontend evidence only.** This source is reproduced from the maintained native-type-member binding regression:

<!-- example:binding-forward-member -->

Both `Show App's Label` occurrences in the indented paragraph bind the local declaration of `Label`, including the occurrence before that declaration. The final occurrence belongs to the outer paragraph and binds the outer declaration. The test checks canonical bindings and source locations; it does **not** establish that a runtime read before assignment succeeds. Name resolution and [definite assignment](#runtime-storage-and-definite-assignment) are separate checks.

## Paragraph scope

An ordinary paragraph participates in the enclosing program or dictionary environment. A paragraph beginning with four spaces at its first code line receives its own lexical scope. Private declarations stay within that scope. Parent bindings are available unless shadowed by a local declaration; sibling private bindings do not leak.

Public and Publicly publish the selected declaration into the containing environment while retaining canonical identity. They do not publish every declaration in the paragraph or expose a private owner's namespace implicitly. Each visibility modifier must target exactly one declaration in its grammatical clause.

The maintained local-shadowing example demonstrates independent private values and an outer value:

<!-- example:local-shadowing -->

The first line establishes the outer `message`. Each four-space paragraph introduces its own `message`; the final `Show message` resolves to the outer value. The existing native test expects `inner7truesecondouter`, with no newline added by Show. This witness has not been executed during the documentation update.

Publication deliberately changes that visibility boundary:

<!-- example:binding-public-scalars -->

This excerpt starts a private paragraph but publishes just the three scalar declarations. Later paragraphs in the [complete public-values example](../../demo/paragraphs/public-values.l) can name `Greeting`, `Count`, and `Enabled`. The four spaces before its first line are significant; the excerpt omits the following Dictionary and Show statements.

## Pronouns

Configured pronouns such as it, this, and that bind the last subject **defined** in the same original source paragraph. This is not the last object mentioned, the last value evaluated, or a dynamic runtime subject stack.

All references within a statement see its incoming paragraph subject. The statement's declarations establish the subject for following statements. When declarations are coordinated, their source order determines the final declared subject. A new paragraph has no inherited physical-paragraph antecedent merely because an outer lexical binding exists.

**Supported example — existing frontend and native-runtime witnesses, not rerun here:**

<!-- example:binding-pronoun-alias -->

`Use Worker as "Job"` declares the latest subject in this paragraph. `Start it` therefore starts `Job`, which retains Worker's canonical App identity. Starting the App is an action, not a new declaration; it does not replace the antecedent. Both Await statements address the same invocation, and `Job's Output` reads that App's finalized output. The retained native test expects the bytes `canonical\n` once.

**Intentionally invalid example — a blank line removes the antecedent:**

<!-- example:binding-pronoun-boundary -->

`Await it` begins a new source paragraph with no preceding subject. The negative fixture expects an unresolved paragraph-pronoun diagnostic before publication. Use the explicit `Worker` name at this boundary, with the appropriate Run/Start before Await; replacing the pronoun alone does not launch an App.

Source-defined action bodies can introduce an `any` parameter. A later pronoun inside that directive binds its preceding generic argument. Multiple preceding generic action arguments can make that pronoun ambiguous; the compiler rejects rather than guessing. This is how Start's “Building any App and then Running it” maintains a typed App target.

Imported file occurrences and dictionary units retain their own paragraph origins. An imported declaration or descriptor refinement must not accidentally replace the caller's pronoun antecedent. Retry order is deliberately irrelevant.

## Possessives and member lookup

A possessive selects a member of its bound owner. `Worker's Output` and the repository's `it's Output` retain the same owner identity through native lowering. A member access is not an independent global lookup for Output.

Dotted qualified names and imported aliases preserve namespace ownership. Once a local owner is selected, a missing member does not fall back to an outer object with the same spelling. A source-defined member can be a lexical action, definition, or value; Output and Error additionally have compiler-typed process-member roles.

The [pronoun example](#pronouns) ends with `Show Job's Output`: `Job` selects Worker and the possessive selects its Output member. For an ordinary user-defined member, the project-owned dictionary source is:

<!-- example:file-import-helper -->

`LocalTools's "Greeting"` declares a member of LocalTools, and `LocalTools's "Display…"` declares an action in that same namespace. The [importing program](imports.md#explicit-source-file-imports) explicitly aliases those members before using their unqualified names. These are vocabulary members, not an arbitrary runtime key/value lookup API.

## Aliases and canonical identity

`Use existing as "Alias"` creates an alias edge to the resolved declaration. Imports use the same relationship after validating the selected dictionary export. Aliases retain canonical application identity, descriptor identity, member ownership, and source signatures. They do not create a second process slot or grant Main entry status to an unrelated definition.

Alias cycles and unresolved targets reject. Private dictionaries cannot capture members later injected into the caller's namespace. A qualified dictionary member stays qualified until an explicit alias makes it available unqualified.

## Runtime storage and definite assignment

The typed Rust plan separates a symbol's declaration from its runtime producer. Assignment writes stable storage at the source instruction position. A runtime symbol may be read only when every path that continues to the read has assigned it. Exiting branches are excluded from that intersection, because they cannot reach the read.

A conditional assignment is not a promise of a hidden initial value. Text storage is owned when language assignment requires a snapshot; comparisons and observation may borrow it. This implementation does not expose a new public ownership or mutability annotation syntax.

Runtime captures must be reachable from the active owning paragraph. Publishing a runtime-dependent definition cannot make private inactive storage available outside its activation. Static definitions and literal captures preserve their original bindings even when the caller shadows the same name.

## Output snapshots

Output becomes available when capture has finalized, not because a variable is declared or Build succeeds. Reading an unfinished target's Output is a lifecycle misuse; assignment does not implicitly Await. A stored Output value keeps the bytes of the invocation read at that point. Rerunning the target later leaves the stored value unchanged.

App-led paragraph continuations receive fresh local storage for each accepted invocation. Cached Await returns the same invocation result without repeating continuation effects. See [execution](execution.md#continuations-and-activation) for slot and cleanup rules.

**Supported example — two activations with separate local and retained outer values:**

<!-- example:repeated-activation -->

The indented App paragraph owns the continuation's `result`. Each outer Run/Await pair starts and joins an invocation, activating that continuation once and showing its local result twice. The outer `snapshot` is assigned after the first Await; the second invocation does not rewrite it. The retained test expects six `Repeat\n` lines: two from each continuation and two from the final Show statements. Because both App invocations print the same text, the separate [counter-based snapshot witness](../../testdata/phase1/f08-rerun-keeps-snapshot/case.json) checks differing invocation bytes.

## Diagnostics

Unresolved names, ambiguous inflection matches, duplicate publication, cycles, inaccessible captures, and incompatible assignment operands retain separate error boundaries. Original source spans and expanded-action call sites remain available. Avoid treating a compiler storage tag such as Unknown as a valid fallback value after an unresolved reference.

## Sources and witnesses

Implementations: [scope planning](../../lyger/scope.go), [lexical environments](../../lyger/lexical_scope.go), [member resolution](../../lyger/semantic_resolution.go), [binding](../../lyger/binding.go), [visibility](../../lyger/visibility.go), [publication](../../lyger/publication.go), and [flow analysis](../../lyger/orchestration_flow.go). Witnesses include [paragraph binding](../../lyger/paragraph_binding_architecture_test.go), [publication](../../lyger/paragraph_publication_architecture_test.go), [capture](../../lyger/paragraph_capture_architecture_test.go), [activation](../../lyger/paragraph_activation_architecture_test.go), [definite assignment](../../lyger/orchestration_plan_test.go), and [stored output](../../lyger/orchestrator_test.go).
