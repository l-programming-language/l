---
name: l-logeme-handoff
description: Diagnose the boundary between AML and Rust logeme source bundles, including the Rust frontend handoff and Rust-only API, when framing, paths, templates or source emission block a compiler task.
---

# Inspect the logeme handoff

The [temporary validation pause](../../../docs/development-harness.md#temporary-validation-pause) governs this workflow. Expensive local validation requires explicit user authorization; configured pull-request events already authorize hosted CI. Follow that canonical policy for requested runs and report unrun stages accurately.

Read [backend guidance](../../../logiceme/AGENTS.md), [AML protocol](../../../docs/architecture.md#aml-and-logeme-protocol) and the relevant fixed expectation. Use maintained `TestDemoGeneratesBuildableSource` and `TestMemberSourceGeneratesBuildableSource` in cmd/l for the Rust frontend handoff. Distinguish App runtime observers from `TestRunnableServiceComposition`, which retains raw AML library and explicit native-consumer controls. The [coverage ledger](../../../testdata/service-execution/coverage.md) records these boundaries. For raw Evaluate (Rust by default), use the [maintained complete-AML corpus](../../../logiceme/testdata/lucid/README.md). Follow the [documented environment](../../../docs/development-harness.md).

1. Preserve exact input and identify its representation: observable frontend view, canonical backend markup or diagnostic adaptation.
2. Confirm the Rust source boundary and selected `fs.FS`, including typed builtin process definitions, then trace tags, framing, metadata and composed glyphs to the parser/evaluator. Keep source binding failures separate from backend boundaries. Rust initialization records are not executable cases; forwarded aliases must preserve category, cardinality and source order. For Lifetimes/GenericTypes, check supported glyph-local placement and explicit forwarding: inherited file values reject, including EA, and parent parameters do not become child-call arguments. Consult the [input contract](../../../builtins/README.md#lifetime-and-generic-inputs) before proposing a change.
3. Record any temporary in-memory adaptation exactly; it never becomes a golden replacement or permission to update expectations.
4. Inspect source bytes, paths, imports and dependencies, including App-constructed Service libraries and explicit raw AML library/bin targets. Executable is a source bundle; distinguish parsing, generation, independent expected-source compilation, generated-source compilation and execution evidence.
5. Establish an exposing regression, reusing an existing failing case when available. Repair the responsible generator layer when authorized; do not replace compilation with a fixture result or enlarge templates to conceal an interpreter defect.

Use selected `fs.FS` and actual production runtime templates. Generated deployment scripts remain artifacts, never validation commands. Preserve raw application stdout and separate compiler diagnostics. The native macOS/Linux CLI compiles Rust workspaces and uses the actual Rust runtime. Establish runtime parity through the supervised mapped tests; backend compilation alone does not prove lifecycle behavior. Containers and Phase 2 in-code concurrency remain outside that CLI scope.

For Rust, compare every returned artifact exactly and run the separate Expected/Generated compilation stage with retained locks. Do not repair or format actual projects before Cargo. The 18 segment definitions remain frozen unless an explicitly scoped task changes them. Read the [current integration limits](../../../builtins/README.md#current-integration-limits); template routes, residual preparation entries and verified compositions are different evidence.

Return input identity, first failing boundary, relevant outputs and evidence limits. Report generation, native compilation, acceptance and runtime verification separately. Use the acceptance skill for the full gate.
