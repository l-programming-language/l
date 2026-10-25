---
name: l-frontend-trace
description: Trace an existing failing L frontend case through source units, imports, tokens, grammar, bindings and AML. Use for lyger or cmd/l compiler diagnosis; ordinary setup and documentation edits do not need this workflow.
---

# Trace the earliest frontend failure

Read [frontend guidance](../../../lyger/AGENTS.md), [architecture](../../../docs/architecture.md) and the failing test/fixture. Use [harness diagnostics](../../../docs/development-harness.md), including `scripts/l-harness diagnose --frontend` when private state is needed.

Follow only necessary boundaries:

- Resources/source: selected filesystem, implicit builtin and dictionary units, typed process-operation definitions and source-defined construction descriptors, package-local fixtures, literal pool, typed origins and program-source projection. Adjacent builtin.l is never loaded implicitly.
- Imports: raw spans/literal identity versus normalized path/alias values; provider ownership and canonical dictionary exports/private helpers.
- Tokens: lexical role, morphology/possession, declaration/reference status, storage and value types before/after binding.
- Grammar: selected case, consumed spans and remaining tokens; failed alternatives cannot count as successful consumption.
- Semantics: records, descriptor/member/alias bindings, App identity and structural Service declarations, paragraph provenance and source order independently of retry order.
- Lowering: ordered instructions, typed Rust storage and process targets, App-only continuation ownership and App-constructed library projection and separate observable/backend views.

Use existing tests or diagnostic overlays without editing protected inputs or routing verify through overlays. Normalize UUIDs/map order only in diagnostic views; retain raw bytes and instruction order. When authorized, repair shared behavior while preserving source-defined vocabulary. Do not restore mirrors, bootstrap fallback, source-rewrite helpers or demo special cases.

Return failing stage/source location, decisive transition, affected implementation and test evidence. Use the acceptance skill for completion.
