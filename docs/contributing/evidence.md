# Evidence and documentation review

A useful specification says both what readers can rely on and how that claim is supported. Three independent dimensions prevent a grammar inventory or an old green report from becoming a broader guarantee than its evidence permits.

## Document role

| Role | What the reader should expect |
| --- | --- |
| Normative contract | Maintained behavioral requirements, with clear implementation boundary |
| Explanatory guide | A practical path through supported concepts and canonical examples |
| Implementation reference | Internal representations, algorithms, generation or host details |
| Future design | A proposal that does not claim present operational support |

Newly authored specification chapters consolidate maintained contracts and source observations. They have not thereby received independent maintainer ratification or standards-body approval.

## Implementation support

| Disposition | Meaning |
| --- | --- |
| Supported | Accepted and implemented within the documented boundary |
| Partial | Some specific forms/stages work; the omitted boundary is named |
| Unsupported | Recognized/rejected or unavailable |
| Internal-only | An implementation representation, not a constructible public facility |
| Design-only | A conceptual model with no present support claim |
| Unresolved | Conflicting evidence requires further investigation |

A single item can have distinct frontend and backend support. For example, a callable template may be accepted by the Rust backend without proving a public L closure form. A raw AML library visibility check at native Rust compilation does not establish an L frontend entry syntax.

## Verification stage

Static inspection establishes what source and contracts currently say. An existing test witness establishes the intended assertion and exercised path. Neither means that the test ran for this checkout.

Source-generation success establishes returned source bytes. Independent expected-Rust compilation establishes feasibility of an oracle. Compiling generated Rust establishes that returned artifacts compile. Supervised runtime checks establish behavior on their actual host and inputs. Full acceptance adds the unfiltered inventory, integrity and original-outcome requirements.

The [implementation ledger](../../testdata/integration/IMPLEMENTATION.md) identifies the tested source, host and validation stages. It does not validate a later checkout. The [Lucid guide](../../logiceme/testdata/lucid/README.md) owns compile-only backend coverage; the [Service ledger](../../testdata/service-execution/coverage.md) separates generation, native and runtime witnesses.

## Resolve a discrepancy without changing semantics

Record the exact source path, semantic question, related contract, existing witness and downstream effect. Prefer the narrow statement that the evidence establishes. If a source enum contains a value not accepted by source construction, document the enum as internal rather than inventing a public type. If a grammar case parses but lowering rejects it, explain both stages.

Leave a precise unresolved item when the evidence conflicts. A documentation task does not authorize changing compiler behavior, protected expected bytes, frozen templates or harness settings. Do not trim output, hide a failure, or call parser acceptance runtime support.

## What a documentation review checks

Every discovered public surface item needs a meaningful section and evidence. Generated inventory complements prose; it does not replace definitions, usage, failure boundaries and examples. Review terminology, imports, source ordering, literal bytes, visibility and target assumptions together.

Website tests also have boundaries. Automated accessibility checks identify the rules they execute; they are not exhaustive certification. Screenshots require visual inspection. A build alone does not establish search behavior, keyboard access, mobile layout or no-JavaScript readability. Report those results separately from language verification.
