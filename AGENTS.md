# Working on `l`

`l` is a human-readable, extensible language with application values. L, Lyger and Logiceme are implemented in Go. Lyger compiles source to AML/logeme markup; logeme in `logiceme/` generates Rust source bundles. Applications and their orchestrator are Rust. `LogicemeService.Evaluate` generates Rust; its optional settings control runtime support. An `Executable` is a source bundle, not a running application.

Read [architecture](docs/architecture.md), [CLI/runtime semantics](docs/cli-build.md) and [development/acceptance](docs/development-harness.md). Read scoped instructions before changing these areas:

- [cmd/l/AGENTS.md](cmd/l/AGENTS.md): CLI, publication and tested compiler state.
- [lyger/AGENTS.md](lyger/AGENTS.md): frontend, grammar and bindings.
- [logiceme/AGENTS.md](logiceme/AGENTS.md): backend and `builtins/` integration.

## Temporary validation pause

Follow the [validation policy](docs/development-harness.md#temporary-validation-pause): expensive local validation remains paused unless explicitly requested; configured pull-request events authorize automatic Linux CI. This overrides automatic local completion commands elsewhere in repository documentation and skills. Use lightweight inspection appropriate to the change and report unrun validation accurately. Previous green runs are revision-specific evidence. Preserve tests, assertions, protected inputs and tool pins.

## Current contract

Preserve current behavioral inputs, assertions, literal expected bytes, expected-value helpers and canonical resources protected by [harness/contract.json](harness/contract.json). Do not skip tests, return hardcoded demo artifacts or regenerate locks to accommodate compiler changes. A protected-fixture amendment requires explicit author approval of its scope; earlier exceptions grant no standing permission. The development guide defines the current integrity and amendment policy.

The [Rust frontend/CLI suite](testdata/integration/README.md) runs in ordinary discovery. Its retained observers cover frontend, publication and native runtime behavior. Read the [implementation evidence](testdata/integration/IMPLEMENTATION.md) for current outcomes and unresolved work; an earlier red or passing snapshot is not the status of this checkout.

Preserve the finite [CI budget policy](docs/development-harness.md#validation-budgets), both unfiltered test phases, runtime watchdogs and concurrency. Further deadline changes require explicit scope; separate compilation from runtime and preserve cumulative runtime allowances across bounded rebuilding.

When the user explicitly requests full validation during the pause, run `scripts/l-check` and resolve every lint warning/error before claiming the gate passed. It runs workflow linting, both full-module Go lint profiles, harness doctor/verify, and the complete tagged native-compilation stage. Use `scripts/l-ci bootstrap` to provision the pinned tools and dependencies first. The unfiltered `scripts/l-harness verify` remains the authoritative acceptance gate; focused tests, diagnostic overlays and explicit vet are investigative evidence, not substitutes for the completion command. Do not filter lint to changed lines, weaken checks, add blanket exclusions, or skip native validation. Narrow suppressions require a specific linter and an explanation of the preserved invariant or intentional best-effort behavior. Historical designs, closed plans, migration records and retained run artifacts are provenance; they are not current instructions or test dependencies. Do not recreate archive trees, bootstrap fallback or copied configuration/lexicon mirrors.

The 18 Rust segments are frozen by the [Lucid suite](logiceme/testdata/lucid/README.md). Its exact generation and separate native compilation stages establish those compositions only; see the [current limitations](builtins/README.md#current-integration-limits). Changing frozen definitions or interfaces requires an explicitly scoped task. Old plans, leftover compiler tables and historical passing matrices do not authorize restoring discarded forms or expanding metadata. Reuse an existing exposing regression for TDD when available. Report generation, compilation, acceptance and runtime verification separately.

Follow the Rust reference's current form contracts. Its approved [Lifetimes and GenericTypes inputs](builtins/README.md#lifetime-and-generic-inputs) are glyph-local native String interiors on FW/HY/MY/Y/named KY, not Func/Prop/Op fields or inherited Class/Object defaults. They introduce no ownership or lifetime inference at the logeme boundary. The frontend's typed Rust plan separately tracks paragraph activations, definite assignment, storage and required mutability from resolved instructions. It adds no new glyph metadata, hidden L-value initialization, sections or blanket cloning. The Lucid suite guide owns backend coverage; the migration ledger owns frontend/runtime outcomes.

Rust excludes Deferred from its logeme templates and rejects supplied Deferred metadata. Preserve this [target contract](builtins/README.md#execution-without-deferred); do not introduce deferred-call semantics or infer runtime cleanup from scheduling.

Apps execute in separate OS processes owned by the orchestrator. Run launches; Stop requests termination without joining; Await joins finalized output and activates its continuation once. Output is raw stdout, including whitespace and newlines. Builtin logeme implementations use compiler-typed process bindings. Start/Execute retain their `any App` source signatures. Pronouns bind the last subject defined in their source paragraph; possessives perform member access. Native macOS/Linux execution is supported; Phase 2 in-code concurrency and cross-compilation are outside scope.

Version-2 source topology and the build coordinator separate App executable crates from their additional constructed Service libraries. Standalone Service declarations remain structural: they produce no package or process, and standalone Service bodies and native lifecycle/runtime-member use reject. Only Apps own process lifecycle and continuations. Preserve Main/Basic construction, primary Service topology, raw AML library/bin targets, and version-1 App publication. Objects remain files and Classes remain module/source groups, not runtime scopes.

The CLI uses embedded resources by default or one complete `--assets` filesystem, without changing CWD. Builtins are implicit; IO is imported explicitly. Explicit `.l` imports remain relative to their source provider. Tests retain package-local fixture loading. Preserve source-unit ownership, descriptor targets, typed parameters and source order after expansion.

Keep compiler diagnostics separate from application stdout. Normalize incidental UUIDs only in diagnostic views. Use repository skills for acceptance checks, frontend tracing and logeme handoff diagnosis; keep changes tied to the current task.
