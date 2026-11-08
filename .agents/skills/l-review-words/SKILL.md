---
name: l-review-words
description: Review L word and dictionary changes for semantic factoring, grammatical roles, canonical targets, typed binding hygiene, evaluation, ownership, failures, native boundaries, contract evidence, and removable complexity. Use for vocabulary audits and PR reviews. Report findings by default; apply changes only when requested.
---

# Review L words

Resolve the current repository root, read applicable instructions and establish the requested behavior and review scope. Inspect the exact revision/diff or selected definitions, qualified consumers and tests independently. Treat the author's explanation as context, not proof. Reconstruct the selected composition and its ownership before accepting conclusions. Historical layouts and earlier green reports do not describe the current checkout automatically.

Review-only means no implementation, documentation, metadata, generated-file, trust-policy or protected-input edits, including cleanup and formatting. Authorized diagnostics may create disposable outputs. Compare reviewed inputs before and after and report any evidence boundary. Apply fixes only when the user requests them, then use [word authoring](../l-author-word/SKILL.md) and [acceptance](../l-acceptance/SKILL.md) within that scope.

## Inspect the relevant semantics

Use [definitions](../../../docs/guide/definitions.md), actual selected source, its contracts and maintained tests. Cover the dimensions affected by the change without imposing a boilerplate report for every leaf:

- **Meaning and factoring:** Check coherent single-word roots, POS/inflections, actor identity, dictionary ownership and visibility. Identify genuinely independent policies and exact component roles; spelling alone does not prove poor factoring. Preserve useful high-level compositions and private helpers.
- **Dataflow and demand:** Trace typed arguments/results, explicit Process inputs, intermediate bindings, once-only evaluation, order, inactive branches, callback multiplicity/escape and repeated or nested scope. Check provenance and definition-site lookup rather than generated-name coincidences. Distinguish callable returns from program exits and recognized grammar from implemented support.
- **Ownership and effects:** Check shared/exclusive borrows, moves, retention, copying, destruction, replacement and release. Reject silent clones or blanket lifetime/representation changes introduced to make a wrapper compile. Identify deliberate propagation, replacement, discard, panic and partial completion.
- **Modifiers and identity:** Inspect the exact canonical target and replaced role, retained base, original operands/resource, wrapper order, repeated construction and unrelated preserved structure. A [canonical alias](../../../docs/spec/names-and-scope.md#aliases-and-canonical-identity) is different from a same-spelled foreign target. Different modifier names or facets do not prove commutativity.
- **Placement and reduction:** Check the appropriate source owner and whether public vocabulary is necessary. Search qualified definitions, calls, aliases, descriptors and generated/evidence dependencies before calling a word unused. Preserve foreign homonyms, distinct policies and tests of different boundaries.

Check changed public meanings against their [canonical entries](../../../docs/contributing/documentation.md#maintain-word-entries), reviewed associations and meaningful examples; flag stale roles, targets, support claims or missing coverage without promoting private helpers.

The [Storage semantics](../../../docs/internals/storage.md) distinguish Require's conditional error construction from eager message arguments, Accept's old-error destruction, continuation-dependent traversal, byte preservation and different retention laws. The [IO contract](../../../docs/stdlib/eng/io.md#explicit-byte-output) distinguishes returned failure from rollback: payload or newline failure can leave partial output. Release alone establishes no flush, durability, cancellation or App completion. Apply the relevant distinction rather than inferring behavior from a word's English name.

## Inspect native boundaries and evidence

Follow the [native guide](../../../docs/guide/native-definitions.md). Accept justified bounded calls, operators, constructors, borrows/projections, patterns and returns. A `Statement` occurrence is not itself a defect, and a native library's internal loop does not require a new L implementation. Inspect new helpers and closures behind calls: a short leaf can hide branching, iteration, resource choice or failure policy that belongs in L. Reject spelling-specific compiler handling and hidden executable bodies, not legitimate typed native atoms.

Check that claimed reusable interfaces have meaningful independent composition evidence and accurate target/ownership rejection. Do not demand invented public consumers for every private continuation. Route compiler diagnosis to [frontend tracing](../l-frontend-trace/SKILL.md) or [logeme handoff](../l-logeme-handoff/SKILL.md); neither a parser pass nor native compilation establishes all runtime semantics.

Use the current compiler/assets and [contract reports](../../../docs/contracts.md). Separate structural assurance, declarations/inferred facts, implementation evidence, environment observations, trusted assumptions and unknowns. Structural certificates and successful compilation do not establish runtime no-allocation, no-blocking, panic or callback guarantees. Include implicit argument/capture/destruction work and outcome-specific behavior in the requested scope.

For [PreservesOnFailure](../../../docs/contracts.md#scoped-checks-and-failure-footprints), identify the actual published resource location, observation and failure outcomes. An error return is not evidence that the resource stayed unchanged or that retry is safe. Missing/stale evidence leaves a required guarantee unresolved; do not manufacture trust, suppress unknowns, narrow scope or substitute a different meaning to make enforcement pass.

Offline inspection, checking, explanation and dictionary verification must not replay recorded commands, query native tools, fetch dependencies or silently capture evidence. Separately explicit builds/tests/capture retain their own authorization and evidence. Preserve protected inputs and expectations during diagnosis; current [acceptance policy](../l-acceptance/SKILL.md) owns validation requirements rather than this skill adding a universal native gate to reviews.

## Report concrete findings

Distinguish these categories; add severity when it helps prioritization:

| Category | What establishes it |
| --- | --- |
| Behavioral defect | A triggering composition and incorrect observable value, bytes, identity, ownership, failure or lifecycle behavior. |
| Unsupported guarantee | A claimed proposition with missing, stale, inapplicable or contradictory evidence at its required scope. |
| Removable complexity | A simpler implementation whose removal conditions preserve required behavior, public meaning and distinct coverage. |
| Optional design preference | A naming or organization choice without a demonstrated correctness defect. |

For each actionable finding, provide the location and qualified word, triggering composition, consequence or unmet obligation, observed evidence or explicitly proposed exposing test, smallest coherent correction, and safe-removal conditions where applicable. Distinguish tests actually run from suggestions. A regex match, adjective label or author's assurance is not sufficient evidence.

Do not force a line-count reduction, inline useful L compositions, publish private helpers or delete behavioral coverage merely to shorten a change. Authorized clean breaks migrate the selected meaning's callers before deletion and leave no compatibility route; they do not authorize unrelated vocabulary retirement.

It is valid to find no blocking defect. State what was inspected, the useful controls, remaining uncertainty and unrun checks. Report generation, compilation, runtime and acceptance evidence separately; successful language tests alone do not prove that every reviewed guarantee holds.
