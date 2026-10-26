# Rust Lyger parity fixtures

These canonical Rust tests retain the original frontend parity obligations. They call current Lyger APIs and pass actual, unmodified AML to Logiceme. Fixtures retain their Rust resource identities. Test helpers do not lower L, replace source bundles, or implement application lifecycle behavior.

The maintained frontend tests preserve the original language, ownership, source-order, literal, table-row and runtime obligations. Backend generation, source observations, compilation and execution are separate phases. A failure at a prerequisite leaves later phases blocked; it does not count as a passing native test.

The bounded source observer reads actual Rust tokens and paired delimiters, function signatures, calls, byte/string literals and assignment scopes. It is not a general Rust parser; frozen Cargo builds are the native legality check. Tests compile the actual returned source and Cargo.toml without alteration. The source-bundle API does not return locks: `builds/` retains eight independently authored manifest/lock pairs for the existing native handoff identities. A retained lock is attached only after the actual manifest matches its independent expectation exactly. No test resolves dependencies or rewrites package identities. The CLI workspace observers separately require actual workspace metadata and retained build locks.

## Native observers

`native/` is an independently authored test-only Cargo workspace with retained dependency resolution (Rust 1.98.1, Tokio 1.53.1, nix 0.31.3). Its helper preserves the original real-process/lifetime-lock protocol and snapshot marker behavior. It is not generated orchestration.

The shutdown adapter supplies a Runtime through the private `new_observed` dependency/output observer. It requires these test-only generated seams, which delegate to the actual generated implementation:

- `#[cfg(test)] pub(crate) async fn l_entry_with_runtime(runtime: &runtime::Runtime) -> u8`, delegating the same actual generated entry/l_main path. The adapter provides no replacement source or lifecycle implementation.
- `runtime::test_support::terminal_status(&Runtime) -> Option<u8>`, reading committed terminal status under the runtime lock. `terminal_entered` remains a distinct pre-acceptance callback.
- The generated runtime and process modules must be visible within the isolated test crate. This grants no public application API.

The adapter preserves 2 completion modes × 6 fault/signal modes, raw output, readiness/PGID/lifetime ownership, exactly-one wait, signal-acceptance ordering, terminal status and leak assertions. Native Rust panics use exit 101 and native panic markers (including a non-string panic payload); No sanitizer-equivalence claim is made. Existing concurrency/barrier assertions remain required.

## Current coverage

The canonical tests retain independent inputs, expected bytes and semantic assertions across these boundaries:

| Boundary | Maintained tests |
| --- | --- |
| Imports, builtin resources and source ownership | [Configured imports](../configured_import_test.go), [distribution](../distribution_frontend_test.go), [source-unit origins](../source_unit_origins_test.go) |
| Definition binding, native framing and exact literals | [Typed definitions](../definition_test.go), [text encoding](../definition_text_encoding_test.go), [literal framing](../literal_framing_regression_test.go) |
| Paragraph binding, captures and activation ownership | [Bindings](../paragraph_binding_architecture_test.go), [captures](../paragraph_capture_architecture_test.go), [activations](../paragraph_activation_architecture_test.go) |
| Recursive source construction and AML | [Recursive Classes](../recursive_classes_test.go), [AML rendering](../aml_render_test.go) |
| Typed orchestration and ordered language effects | [Orchestrator](../orchestrator_test.go), [Phase 1 language](../phase1_language_test.go) |
| Native lifecycle, continuation and shutdown | [Nested Await](../phase1_nested_await_test.go), [shutdown](../phase1_shutdown_test.go) |
| Selected typed builtins, App lifecycle and structural Service boundaries | [Builtin catalog](../builtin_operations_contract_test.go), [Service boundaries](../runnable_service_test.go), [App process artifacts](../service_process_units_test.go), [coverage boundaries](../../testdata/service-execution/coverage.md) |

Shared frontend tests also load these fixture paths directly. Source-origin expectations retain the same Rust paths rather than aliases for another target.

## Verification

The [local validation pause](../../docs/development-harness.md#temporary-validation-pause)
requires an explicit request before full package or completion checks. Automatic
Linux PR validation is authorized for the documented PR events; manual dispatch
requires explicit authorization. When requested, package-level frontend checks
use `GOMAXPROCS=4 go test -mod=readonly -count=1 -timeout=10m -json ./lyger`.
Go runs the compiler tests; generated Apps and their orchestrator execute as Rust binaries.
For full validation, follow the [development guide](../../docs/development-harness.md#full-validation-when-requested):
`scripts/l-check` includes workflow lint, both Go lint profiles, unfiltered
acceptance and complete tagged native compilation. The
[implementation ledger](../../testdata/integration/IMPLEMENTATION.md) records
revision-specific generation, compilation and runtime evidence separately.
