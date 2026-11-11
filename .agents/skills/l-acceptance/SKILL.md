---
name: l-acceptance
description: Run and interpret the current repository gate, frozen logeme checks and exact expected output. Use for baseline checks, progress verification and completion claims.
---

# Check acceptance

Read [current root instructions](../../../AGENTS.md#temporary-acceptance) and the [development guide](../../../docs/development-harness.md). During the core-only reset, `scripts/l-check` runs the temporary hello-world gate through the existing supervised runner. Do not run or claim the retired full gate as a completion requirement.

The selected checks cover original hello-world source/compiler behavior, real standalone App compilation and execution, explicit rejection of runtime effects, selected core Print authority, and frozen logeme assets. Require complete named outcomes, exact expected output, and successful subprocess exit. An empty selection, skipped test, malformed report, stale artifact or unavailable tool is not a pass.

Only `builtins/segments/` and `builtins/templates/` remain protected. Their bytes, interfaces and hash expectations must stay unchanged. Other fixtures can change within the authorized reset. Replace assertions that enforce retired words or invalid L; retain useful runtime observations through valid replacement callers. Do not regenerate an oracle just to hide a defect.

Use pinned Go/Rust tools and existing supervision, watchdogs and concurrency. No hidden toolchain installation, runtime restoration, command replay or publication. Report revision/dirty state, platform, command/report locations and actual failures. Separate source generation, native compilation, execution and the temporary gate. The historical full-module, runtime, website and tagged-native suites are unrun unless separately executed; they cannot be reported as passing.

Use [word review](../l-review-words/SKILL.md) for morphology, roles and factoring. Passing behavior or metadata consistency does not approve vocabulary design. Establish exposing RED before a repair and preserve it separately from GREEN. Runtime rebuilding is a subsequent source-word task, not permission to add hardcoded Rust support to make the gate pass.

For source name/family changes, interpret focused checks against the [authored-name and alias contract](../../../docs/spec/names-and-scope.md#possessives-and-member-lookup) and [complete-family contract](../../../docs/spec/definitions.md#inflection-declarations). Positive callers must use valid families; intentional rejection cases must identify their expected boundary. Preserve ordinary-label, same-spelled-form, canonical-root-alias and data/native-byte controls. Focused regressions supplement the temporary hello-world gate. Review affected documentation/provenance dependencies before refreshing any fingerprint; a regenerated snapshot is not semantic approval.
