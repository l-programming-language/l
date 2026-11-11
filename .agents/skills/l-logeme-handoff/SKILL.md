---
name: l-logeme-handoff
description: Diagnose the boundary between AML and Rust logeme source bundles, including the Rust frontend handoff and Rust-only API, when framing, paths, templates or source emission block a compiler task.
---

# Inspect the logeme handoff

The [validation policy](../../../docs/development-harness.md#validation-authorization) governs this workflow. Explicit task authorization covers its bounded baseline, regression, native and completion checks without repeated permission requests. Configured pull-request events authorize hosted CI; manual dispatch requires explicit authorization. Report unrun stages accurately.

Read [backend guidance](../../../logiceme/AGENTS.md), [AML protocol](../../../docs/architecture.md#aml-and-logeme-protocol) and the failing fixture. Use `TestCoreHelloWorld` in cmd/l for the supported frontend handoff and the [complete-AML corpus](../../../logiceme/testdata/lucid/README.md) for independent backend compositions. Follow the [documented environment](../../../docs/development-harness.md).

1. Preserve exact input and identify its representation: observable frontend view, canonical backend markup or diagnostic adaptation.
2. Confirm the Rust source boundary and selected `fs.FS`, then trace tags, framing, metadata and composed glyphs to the parser/evaluator. Keep source binding failures separate from backend boundaries. Rust initialization records are not executable cases; forwarded aliases must preserve category, cardinality and source order. For Lifetimes/GenericTypes, check supported glyph-local placement and explicit forwarding: inherited file values reject, including EA, and parent parameters do not become child-call arguments. Consult the [input contract](../../../builtins/README.md#lifetime-and-generic-inputs) before proposing a change.
3. Record any temporary in-memory adaptation exactly; it never becomes a golden replacement or permission to update expectations.
4. Inspect source bytes, paths, imports and dependencies, including explicit raw AML library/bin targets. Executable is a source bundle; distinguish parsing, generation, independent expected-source compilation, generated-source compilation and execution evidence.
5. Establish an exposing regression, reusing an existing failing case when available. Repair the responsible generator layer when authorized; do not replace compilation with a fixture result or enlarge templates to conceal an interpreter defect.

Use the selected `fs.FS` and retained core/logeme resources. Runtime assets and their assembly APIs are removed. The CLI compiles one standalone App. Preserve generic grammar lowering and AML rendering; do not rebuild library policies in Go/Rust or restore them from caches. Generated deployment scripts remain artifacts, never validation commands. Preserve raw application stdout and separate compiler diagnostics.

Keep source registration and descriptor-target failures upstream with [frontend tracing](../l-frontend-trace/SKILL.md). When stages depend on fixed generated names, trace their authored inputs/results and selected definitions before changing emission. Inspect the native callee behind a leaf for concealed policy, while retaining legitimate bounded operations. Route vocabulary factoring to [word review](../l-review-words/SKILL.md); neither a misleading word role nor tightly coupled stages justify a Rust algorithm, rendered-name recovery or an AML/logeme interface change.

[Authored-name](../../../docs/spec/names-and-scope.md#possessives-and-member-lookup) and [complete-family](../../../docs/spec/definitions.md#inflection-declarations) validation belongs upstream. Preserve native code, AML, `<@Owner.Member@>` template slots and internal identities exactly; dots there do not justify changing backend syntax or repairing a rejected source name or alias.

For Rust, compare every returned artifact exactly and run the separate Expected/Generated compilation stage with retained locks. Do not repair or format actual projects before Cargo. Do not edit `builtins/segments/`, `builtins/templates/`, their interfaces or frozen hash expectations. Read the [current integration limits](../../../builtins/README.md#current-integration-limits); template routes, residual preparation entries and verified compositions are different evidence.

Return input identity, first failing boundary, relevant outputs and evidence limits. Report generation, native compilation, acceptance and runtime verification separately. Use the acceptance skill for the current gate.
