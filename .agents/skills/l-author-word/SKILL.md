---
name: l-author-word
description: Design, implement, revise, or retire L word families and dictionary compositions. Use for vocabulary changes requiring explicit grammatical roles, reusable source composition, typed Process bindings, narrow native leaves, contracts, and tests-first validation. Route isolated compiler diagnosis and validation-only tasks to the existing repository skills.
---

# Author L words

Work from the current checkout and selected definitions, not a historical layout or an installed compiler's embedded vocabulary. Resolve the repository root before following paths or running checks, including when invoked from a nested directory. Read applicable repository instructions and inspect the relevant diff before editing. Preserve unrelated work and the user's requested scope.

## Establish the meaning and owner

Search existing qualified words and their consumers by behavior before choosing syntax. Decide whether to reuse a capability, improve its definition, or introduce a genuinely new meaning. Reuse can completely satisfy a request; an unfamiliar proposed name does not require a new synonym.

Write a proportional semantic brief in task notes or the existing change description:

- Identify the owning dictionary, qualified root, POS, actual inflections, visibility, and concrete observable meaning.
- Identify the canonical target. For a modifier, specify the component or role it requires, introduces, changes, replaces, and preserves.
- State arguments, results, dataflow and relevant borrow, move, retention, copying and release behavior.
- State evaluation order and multiplicity, conditional demand, Process/callback invocation and escape, and failure, partial-effect and cleanup behavior.
- Name reused definitions, justify new abstraction boundaries, and distinguish required guarantees from available evidence and unknowns.

A trivial leaf needs only relevant decisions, not a permanent per-word manifest. Use [definitions](../../../docs/guide/definitions.md) for supported forms and [inflection](../../../docs/spec/lexical-structure.md#vocabulary-and-inflection) for word families. Choose coherent single-word roots and dictionary ownership. Factor independent policies by meaning and target, not a CamelCase splitter. Actor nouns belong to the existing verb family; do not independently redefine that identity. Keep helpers private unless their interface has a real external use. Several source steps can form a useful high-level word.

Choose the supported source action, Function, Process, descriptor or native leaf that matches the work. A callable Return is not program Exit. Grammatical recognition alone does not establish lowering or execution support for an adverb, attachment or wrapper. Consult [Process, Function and Return](../../../docs/spec/definitions.md#process-function-and-return) and [construction](../../../docs/spec/definitions.md#descriptors-and-constructors) before extending a form.

## Establish an exposing test

Reuse an existing exposing regression or add the smallest independent observer before changing the implementation. Exercise the selected production definition, or the task-owned definition under development, through the real compiler. Cover the changed semantics: exact values/bytes, inactive work, order, nested scope, target identity, ownership/destruction, partial failure or callback behavior as applicable.

For repairs, observe the incorrect behavior before fixing it. For a new capability, a failure at that precise missing-definition/support boundary is availability RED; label it separately from runtime RED. Missing tools, unrelated syntax errors, broken fixtures and stale resources are setup failures. Never manufacture RED by breaking production code or weakening an oracle.

Keep native observers limited to observation, such as counting calls, recording writes or destruction. They must not implement the algorithm being tested. Preserve protected fixtures, literal expectations, tool pins and gates. Record commands and actual outcomes, including initial failures.

## Compose and preserve identity

Keep new sequencing, branching, iteration, callback policy, error policy and resource selection in L. Use explicit typed Process inputs and intermediate labels. Preserve definition-site lookup, canonical targets, paragraph/source order, lexical scope and source provenance through repeated and nested uses. Do not recover bindings from rendered Rust, invent hidden result names, use global temporary counters or add English-spelling compiler cases.

Evaluate operands once in their intended order and forward the original target/resource. Preserve wrapper order and repeated construction: different names or facets do not establish commutativity or justify deduplication. A [canonical alias](../../../docs/spec/names-and-scope.md#aliases-and-canonical-identity) retains applicability; a foreign same-spelled word does not. Do not silently clone consumed inputs, erase ownership, impose blanket lifetimes or change representation to make a wrapper compile.

Follow the [native boundary guide](../../../docs/guide/native-definitions.md). Before adding a leaf, identify the missing target operation, why existing words do not supply it and which policy remains in L. One call, operator, constructor, borrow/projection or bounded pattern/binding can be appropriate. A small `Statement` return is not automatically an escape hatch. Conversely, one call to a new Rust helper can conceal the whole algorithm; inspect that helper. An established native operation such as `write_all` can remain a legitimate boundary despite its internal loop.

Use approved typed AML/name/type/lifetime inputs and existing logeme interfaces. Do not hide executable bodies in Statement, EA, opaque closures or a second interpreter. Route a genuine binding/lowering failure to [frontend tracing](../l-frontend-trace/SKILL.md), or an AML/native source failure to [logeme handoff](../l-logeme-handoff/SKILL.md). A compiler correction must be authorized, generic and covered by an independent-word regression; unsupported forms are not permission to enlarge frozen interfaces.

For an exported modifier or higher-order interface claimed to be general, exercise a structurally different consumer: an independent action, different writer, nested composition, canonical alias or meaningful ownership variation. Renaming a copied fixture proves little. Private leaves need no invented public consumers. Preserve accurate rejection of invalid targets and consumed-value reuse.

## Check semantics and assurance

Use current source and maintained contracts as the semantic authority. The [Storage guide](../../../docs/internals/storage.md) distinguishes conditional error construction from eager argument evaluation, replacement-error destruction, traversal result/continuation policies, exact byte preservation, and distinct Arc/Rc, snapshot, Pointer and final-owner behavior. Preserve those distinctions when relevant. Fewer buffers do not prove a fixed allocation count. Release does not imply flush, commit, durability, cancellation or App completion. The [IO contract](../../../docs/stdlib/eng/io.md#explicit-byte-output) owns Send/Lined partial output and their separate newline, locking, flush and panic policies.

Use the edited checkout's compiler and complete selected assets. Read current [contract commands and reports](../../../docs/contracts.md); obtain real occurrence, target, scope and resource IDs rather than guessing them from spelling. Keep structural certificates, inferred summaries, declared facts, implementation evidence, environment observations, trusted assumptions and unknowns distinct. Compilation and environment capture do not approve native behavior; structural adjective assurance does not prove runtime allocation, blocking, panic or callback claims.

Include argument construction, captures, destruction and outcome-specific work within the requested scope. [Failure preservation](../../../docs/contracts.md#scoped-checks-and-failure-footprints) requires the exact published footprint, observation and explicit failure outcomes. Do not invent witnesses, suppress unknowns, narrow required scope or select another meaning to pass a profile. An unresolved required guarantee remains incomplete.

Offline inspect/check/explain and dictionary verification must not replay recorded commands, query native versions, fetch dependencies or silently capture evidence. Explicit builds, tests and evidence capture are separate authorized operations. Regenerate or invalidate affected evidence through its actual dependencies; never copy a previous green status onto changed definitions.

## Complete the change

For an authorized retirement, inspect qualified definitions, consumers, imports, aliases, descriptors, docs, tests and generated-name/evidence dependencies. Migrate callers and delete the retired meaning, preserving unrelated homonyms and distinct policies. Leave no compatibility alias, shim, second spelling, duplicate layout or fallback unless the user explicitly requires one. Check a forwarding helper's callers and purpose before removing it.

Maintain the changed meaning’s [canonical word entry](../../../docs/contributing/documentation.md#maintain-word-entries), meaningful examples and reviewed associations without publishing private helpers or duplicating definitions. Use [acceptance](../l-acceptance/SKILL.md) for validation and completion; it owns commands, protections, budgets and stage requirements. Do not impose additional native runs on unrelated documentation or read-only tasks.

Report deleted, modified and new qualified words with POS and definitions; reused hierarchy and native boundaries; observed RED/GREEN tests; scope, evidence, assumptions and unknowns; and unrun checks or limitations. If reuse required no new word, say so. Separate generation, native compilation, runtime behavior and full acceptance, and do not claim guarantees beyond the observed evidence.
