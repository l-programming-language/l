---
name: l-frontend-trace
description: Trace an existing failing L frontend case through source units, imports, tokens, grammar, bindings and AML. Use for lyger or cmd/l compiler diagnosis; ordinary setup and documentation edits do not need this workflow.
---

# Trace the earliest frontend failure

Read [frontend guidance](../../../lyger/AGENTS.md), [architecture](../../../docs/architecture.md) and the failing test/fixture. Use [harness diagnostics](../../../docs/development-harness.md), including `scripts/l-harness diagnose --frontend` when private state is needed.

Follow only necessary boundaries:

- Resources/source: selected filesystem and root `dictionaries.json`, direct core builtin unit (the bundled dictionary registry is empty), source-defined construction descriptors, package-local fixtures, literal pool, typed origins and program-source projection. Includes retain their owning unit and resource root. Adjacent builtin.l is never loaded implicitly.
- Imports: raw spans/literal identity versus normalized path/alias values; provider ownership and canonical dictionary exports/private helpers.
- Tokens: lexical role, morphology/possession, declaration/reference status, storage and value types before/after binding.
- Grammar: selected case, consumed spans and remaining tokens; failed alternatives cannot count as successful consumption.
- Semantics: records, descriptor/member/alias bindings, paragraph provenance and source order independently of retry order.
- Lowering: ordered instructions, typed operands and separate observable/backend views. Preserve generic AML composition for grammar, bindings and control flow. Old library-specific adapters are not requirements to restore deleted words. The CLI executable boundary remains core Print in one standalone App.

For a role-sensitive failure, trace the declaring grammar case, registered family/POS, inflection role, selected caller grammar and canonical target/binding before AML. Preserve the [distinction from Function/value storage](../../../docs/spec/definitions.md#declaration-roles-and-identity), including aliases and same-spelled modifiers. Route naming, ownership-depth and factoring judgments to [word review](../l-review-words/SKILL.md) or [authoring](../l-author-word/SKILL.md); a questionable source design is not permission for an English-spelling compiler exception.

Use existing tests or diagnostic overlays without editing frozen assets or substituting diagnostics for the temporary gate. Normalize UUIDs/map order only in diagnostic views; retain raw bytes and instruction order. When authorized, repair shared behavior while preserving source-defined vocabulary. Do not restore mirrors, bootstrap fallback, source-rewrite helpers or demo special cases.

Return failing stage/source location, decisive transition, affected implementation and test evidence. Use the acceptance skill for completion.
