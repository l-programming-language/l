---
name: l-author-word
description: Design, implement, revise, or retire L word families and dictionary compositions. Use for vocabulary changes requiring explicit grammatical roles, reusable source composition, typed Process bindings, narrow native leaves, contracts, and tests-first validation. Route isolated compiler diagnosis and validation-only tasks to the existing repository skills.
---

# Author L words

Work from the current checkout and selected definitions, not a historical layout or an installed compiler's embedded vocabulary. Resolve the repository root before following paths or running checks, including when invoked from a nested directory. Read applicable repository instructions and inspect the relevant diff before editing. Preserve unrelated work and the user's requested scope.

## Establish the meaning and owner

Search existing qualified words and their consumers by behavior before choosing syntax. Include private unit roots, implementation members, aliases, descriptor targets and compiler-selected definitions relevant to the task; the public glossary is not the whole vocabulary. Decide whether to reuse a capability, improve its definition, or introduce a genuinely new meaning. Reuse can completely satisfy a request; an unfamiliar proposed name does not require a new synonym.

Write a proportional semantic brief in task notes or the existing change description:

- Identify the owning dictionary, qualified root, registered POS, actual inflections, visibility, and concrete observable meaning. Classify action, actor construction, modifier, predicate, value or native pattern before selecting syntax.
- Identify the canonical target. For a modifier, specify the component or role it requires, introduces, changes, replaces, and preserves.
- State arguments, results, dataflow and relevant borrow, move, retention, copying and release behavior.
- State evaluation order and multiplicity, conditional demand, Process/callback invocation and escape, and failure, partial-effect and cleanup behavior.
- Name reused definitions, justify new abstraction boundaries, and distinguish required guarantees from available evidence and unknowns.

A trivial leaf needs only relevant decisions, not a permanent per-word manifest. Apply the [word design convention](../../../docs/guide/definitions.md#organize-words-by-meaning-and-grammatical-role) in that brief: justify deeper declaration ownership and explain each public interface. Use grammatical possessives in authored L; dotted canonical names are not source syntax. Factor independent policies while retaining useful private helpers and high-level compositions.

Trace the [actual declaration role](../../../docs/spec/definitions.md#declaration-roles-and-identity), not just the English name or glossary label. Investigate the guide's operation, actor and stage-noun warning signs through actual behavior and callers. For a modifier, demonstrate a compatible base and canonical target rather than inferring them from related spellings.

Expand every declared form through the current [inflection rules](../../../docs/spec/lexical-structure.md#vocabulary-and-inflection). Check intended spelling, grammatical role and representative caller identity, including actors in the same family. Resolve incorrect forms in authorized source changes, not metadata-only relabeling, and do not manufacture unsupported cases.

Before reusing or deleting a similar mechanism, compare actual operands/results, target, mutability, moves/borrows, retention/destruction, demand/order, return/discard and failure behavior. Inspect its native callee and consumers. A writable buffer and an immutable empty value, or Rc and Arc retention, are not interchangeable merely because their names or bodies resemble each other.

Choose the supported source action, Function, Process, descriptor or native leaf that matches the work. A callable Return is not program Exit. Grammatical recognition alone does not establish lowering or execution support for an adverb, attachment or wrapper. Consult [Process, Function and Return](../../../docs/spec/definitions.md#process-function-and-return) and [construction](../../../docs/spec/definitions.md#descriptors-and-constructors) before extending a form.

## Establish an exposing test

Reuse an existing exposing regression or add the smallest independent observer before changing the implementation. Exercise the selected production definition, or the task-owned definition under development, through the real compiler. Cover the changed semantics: exact values/bytes, inactive work, order, nested scope, target identity, ownership/destruction, partial failure or callback behavior as applicable.

For repairs, observe the incorrect behavior before fixing it. For a new capability, a failure at that precise missing-definition/support boundary is availability RED; label it separately from runtime RED. Missing tools, unrelated syntax errors, broken fixtures and stale resources are setup failures. Never manufacture RED by breaking production code or weakening an oracle.

Keep native observers limited to observation, such as counting calls, recording writes or destruction. They must not implement the algorithm being tested. Preserve frozen assets, meaningful literal expectations, tool pins and the current gate. Record commands and actual outcomes, including initial failures.

## Compose and preserve identity

Keep library algorithms in L. Compiler-owned grammar, binding and control flow lower through frozen AML interfaces; they do not require library words. Use explicit typed Process inputs and intermediate labels. Trace producer/consumer bindings before introducing stages: independent words must not communicate through fixed generated names. Preserve definition-site lookup, canonical targets, paragraph/source order, lexical scope and source provenance through repeated and nested uses. Do not recover bindings from rendered Rust, invent hidden result names, use global temporary counters or add English-spelling compiler cases.

Evaluate operands once in their intended order and forward the original target/resource. Preserve wrapper order and repeated construction: different names or facets do not establish commutativity or justify deduplication. A [canonical alias](../../../docs/spec/names-and-scope.md#aliases-and-canonical-identity) retains applicability; a foreign same-spelled word does not. Do not silently clone consumed inputs, erase ownership, impose blanket lifetimes or change representation to make a wrapper compile.

Follow the [native boundary guide](../../../docs/guide/native-definitions.md). Before adding a leaf, identify the missing target operation, why existing words do not supply it and which policy remains in L. One call, operator, constructor, borrow/projection or bounded pattern/binding can be appropriate. A small `Statement` return is not automatically an escape hatch. Conversely, one call to a new Rust helper can conceal the whole algorithm; inspect that helper. An established native operation such as `write_all` can remain a legitimate boundary despite its internal loop.

Use approved typed AML/name/type/lifetime inputs and existing logeme interfaces. Do not hide executable bodies in Statement, EA, opaque closures or a second interpreter. Route a genuine binding/lowering failure to [frontend tracing](../l-frontend-trace/SKILL.md), or an AML/native source failure to [logeme handoff](../l-logeme-handoff/SKILL.md). A compiler correction must be authorized, generic and covered by an independent-word regression; unsupported forms are not permission to enlarge frozen interfaces.

For an exported modifier or higher-order interface claimed to be general, exercise a structurally different consumer: an independent action, different writer, nested composition, canonical alias or meaningful ownership variation. Renaming a copied fixture proves little. Private leaves need no invented public consumers. Preserve accurate rejection of invalid targets and consumed-value reuse.

## Check semantics and assurance

Use current source and maintained contracts as the semantic authority. Compare conditional work with eager argument evaluation, replacement-error destruction, traversal policy, exact bytes and distinct ownership laws where relevant. An error return does not imply rollback, and release does not imply flush, durability, cancellation or completion. Removed library definitions are not available building blocks or semantic requirements.

Use the edited checkout's compiler and complete selected assets. Trace actual canonical targets and resource ownership from source and compiler observations. Keep source inspection, implementation tests, [native build evidence](../../../docs/cli-build.md#optional-native-build-evidence), assumptions and unknowns distinct. Successful compilation or a toolchain record does not establish runtime allocation, blocking, panic or callback guarantees.

Include argument construction, captures, destruction and outcome-specific work within the requested scope. A failure-preservation claim needs the actual resource, observed state and explicit failure outcomes. Do not invent evidence or narrow the required scope to claim success. An unresolved required guarantee remains incomplete.

Reading historical evidence does not authorize replaying its commands, probing tools or fetching dependencies. Builds and tests are explicit operations. Invalidate affected evidence through its actual dependencies; never copy a previous green status onto changed definitions.

## Complete the change

For an authorized role, family or ownership change, migrate the declaring source, forms/actor uses, qualified consumers, imports, aliases, descriptor targets and compiler/native selection together. Review generated support, provenance, documentation and evidence dependencies. For retirement, delete the old meaning after migration, preserving unrelated homonyms and distinct policies. Leave no compatibility alias, shim, second spelling, duplicate layout or fallback unless the user explicitly requires one. Check a forwarding helper's callers and purpose before removing it. A review or guidance task does not authorize this production migration.

Maintain the changed meaning’s [canonical word entry](../../../docs/contributing/documentation.md#maintain-word-entries), meaningful examples and reviewed associations without publishing private helpers or duplicating definitions. Use [acceptance](../l-acceptance/SKILL.md) for validation and completion; it owns commands, protections, budgets and stage requirements. Do not impose additional native runs on unrelated documentation or read-only tasks.

Report deleted, modified and new qualified words with POS and definitions; reused hierarchy and native boundaries; observed RED/GREEN tests; scope, evidence, assumptions and unknowns; and unrun checks or limitations. If reuse required no new word, say so. Separate generation, native compilation, runtime behavior and the temporary hello-world gate, and do not claim guarantees beyond the observed evidence.
