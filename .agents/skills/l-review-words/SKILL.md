---
name: l-review-words
description: Review L word and dictionary changes for semantic factoring, grammatical roles, canonical targets, typed binding hygiene, evaluation, ownership, failures, native boundaries, contract evidence, and removable complexity. Use for vocabulary audits and PR reviews. Report findings by default; apply changes only when requested.
---

# Review L words

Resolve the current repository root, read applicable instructions and establish the requested behavior and review scope. Inspect the exact revision/diff or selected definitions, qualified consumers and tests independently. Treat the author's explanation as context, not proof. Reconstruct the selected composition and its ownership before accepting conclusions. Historical layouts and earlier green reports do not describe the current checkout automatically.

Review-only means no implementation, documentation, metadata, generated-file, trust-policy or protected-input edits, including cleanup and formatting. Authorized diagnostics may create disposable outputs. Compare reviewed inputs before and after and report any evidence boundary. Apply fixes only when the user requests them, then use [word authoring](../l-author-word/SKILL.md) and [acceptance](../l-acceptance/SKILL.md) within that scope.

## Inspect the relevant semantics

State the inspected scope and omissions. Follow relevant public words, qualified implementation members, private unit roots, aliases, descriptor targets and compiler-selected definitions; the glossary intentionally omits some of these. Inspect actual selected source, contracts and consumers before accepting a source inventory as a design review.

Use the [word design convention](../../../docs/guide/definitions.md#organize-words-by-meaning-and-grammatical-role) and [registration semantics](../../../docs/spec/definitions.md#declaration-roles-and-identity). For each candidate concern, establish its declaring form, registered family, use-site role, owner, inputs/results and real callers. Distinguish an action, actor construction, modifier, predicate, value and native pattern. A noun label may accurately transcribe a questionable source design; changing the label alone does not repair it.

Assess the semantic reason for deeper declaration ownership under the guide's convention. Authored L uses grammatical possessives, not dotted canonical names. Separate declaration design from ordinary member access; unjustified depth is a design concern, not proof of a runtime defect.

Check the [complete-family contract](../../../docs/spec/definitions.md#inflection-declarations) and [authored-name/alias rules](../../../docs/spec/names-and-scope.md#possessives-and-member-lookup) as requirements, including private, core and imported declarations. Distinguish ordinary labels and Function-value storage from lexical declarations and newly named invocations. Trace alias chains to their canonical root and selected role before accepting inherited forms. Preserve dots in data, paths, native/template contents and internal identities.

Investigate the guide's actor and stage-noun warning signs while checking its valid controls. Expand every declared form with the actual suffix rules and inspect intended spelling, role and canonical caller identity. A missing or invalid required family is a contract violation; a complete family can still merit semantic review. Passing source/metadata consistency can preserve bad morphology; use the registration rules to distinguish passive forms from descriptors.

When a suspect noun resembles another word's inflection, show that family's expanded forms and distinguish their canonical records in the report. Give each nominated role/family concern a disposition, including designs retained with a reason; do not leave that identity comparison implicit in a general semantics summary.

Cover the dimensions affected by the change without imposing a boilerplate report for every leaf:

- **Meaning and factoring:** Check coherent single-word roots, POS/inflections, actor identity, dictionary ownership and visibility. Identify genuinely independent policies and exact component roles; spelling alone does not prove poor factoring. Preserve useful high-level compositions and private helpers.
- **Dataflow and demand:** Trace typed arguments/results, explicit Process inputs, intermediate bindings, once-only evaluation, order, inactive branches, callback multiplicity/escape and repeated or nested scope. Check provenance and definition-site lookup rather than generated-name coincidences. Distinguish callable returns from program exits and recognized grammar from implemented support.
- **Ownership and effects:** Check shared/exclusive borrows, moves, retention, copying, destruction, replacement and release. Reject silent clones or blanket lifetime/representation changes introduced to make a wrapper compile. Identify deliberate propagation, replacement, discard, panic and partial completion.
- **Modifiers and identity:** Inspect the exact canonical target and replaced role, retained base, original operands/resource, wrapper order, repeated construction and unrelated preserved structure. A [canonical alias](../../../docs/spec/names-and-scope.md#aliases-and-canonical-identity) is different from a same-spelled foreign target. Different modifier names or facets do not prove commutativity.
- **Placement and reduction:** Check the appropriate source owner and whether public vocabulary is necessary. Search qualified definitions, calls, aliases, descriptors and compiler/native/generated/evidence dependencies before calling a word unused. Compare types/targets, mutability, ownership, demand, return/discard and failure behavior before proposing reuse. Trace fixed intermediate names between stages. Similar bodies do not establish safe substitution, and a single caller does not establish uselessness. Preserve foreign homonyms, distinct policies and tests of different boundaries.

Check changed public meanings against their [canonical entries](../../../docs/contributing/documentation.md#maintain-word-entries), reviewed associations and meaningful examples; flag stale roles, targets, support claims or missing coverage without promoting private helpers.

Distinguish conditional work from eager arguments, error replacement from propagation, and returned failure from rollback. Partial output may remain after failure. Release alone establishes no flush, durability, cancellation or completion. Establish each obligation from current source and requirements; removed library definitions are not the specification.

## Inspect native boundaries and evidence

Follow the [native guide](../../../docs/guide/native-definitions.md). Accept justified bounded calls, operators, constructors, borrows/projections, patterns and returns. A `Statement` occurrence is not itself a defect, and a native library's internal loop does not require a new L implementation. Inspect new helpers and closures behind calls for concealed library algorithms. Preserve compiler-owned grammar, binding and control-flow lowering through frozen AML interfaces. Reject spelling-specific compiler handling and hidden executable bodies, not legitimate typed native atoms.

Check that claimed reusable interfaces have meaningful independent composition evidence and accurate target/ownership rejection. Do not demand invented public consumers for every private continuation. Route compiler diagnosis to [frontend tracing](../l-frontend-trace/SKILL.md) or [logeme handoff](../l-logeme-handoff/SKILL.md); neither a parser pass nor native compilation establishes all runtime semantics.

Use the current compiler/assets and actual source, test and [native build evidence](../../../docs/cli-build.md#optional-native-build-evidence). Separate source inspection, implementation observations, assumptions and unknowns. Successful compilation does not establish runtime no-allocation, no-blocking, panic or callback guarantees. Include implicit argument/capture/destruction work and outcome-specific behavior in the requested scope.

For failure-preservation claims, identify the actual resource, observed state and failure outcomes. An error return is not evidence that the resource stayed unchanged or that retry is safe. Missing or stale evidence leaves a required guarantee unresolved; do not manufacture trust, suppress unknowns, narrow scope or substitute a different meaning.

Do not replay recorded commands, query native tools or fetch dependencies merely to inspect historical evidence. Explicit builds and tests retain their own authorization and evidence. Preserve protected inputs and expectations during diagnosis; current [acceptance policy](../l-acceptance/SKILL.md) owns validation requirements rather than this skill adding a universal native gate to reviews.

## Report concrete findings

Distinguish these categories; add severity when it helps prioritization:

| Category | What establishes it |
| --- | --- |
| Behavioral defect | A triggering composition and incorrect observable value, bytes, identity, ownership, failure or lifecycle behavior. |
| Unsupported guarantee | A claimed proposition with missing, stale, inapplicable or contradictory evidence at its required scope. |
| Removable complexity | A simpler implementation whose removal conditions preserve required behavior, public meaning and distinct coverage. |
| Optional design preference | A naming or organization choice without a demonstrated correctness defect. |

For each actionable finding, provide the location and qualified word, triggering composition, consequence or unmet obligation, observed evidence or explicitly proposed exposing test, smallest coherent correction, and safe-removal conditions where applicable. A role/family correction must account for source declarations, forms, callers, aliases, descriptor targets and selected-definition/provenance dependencies, not just metadata. Distinguish tests actually run from suggestions. A regex match, adjective label or author's assurance is not sufficient evidence.

Do not force a line-count reduction, inline useful L compositions, publish private helpers or delete behavioral coverage merely to shorten a change. Authorized clean breaks migrate the selected meaning's callers before deletion and leave no compatibility route; they do not authorize unrelated vocabulary retirement.

It is valid to find no blocking defect. State what was inspected, the useful controls, remaining uncertainty and unrun checks. Report generation, compilation, runtime and acceptance evidence separately; successful language tests alone do not prove that every reviewed guarantee holds.
