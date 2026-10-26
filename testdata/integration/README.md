# Frontend and native runtime integration

This suite exercises the actual Rust frontend, CLI workspace/publication pipeline
and runtime. Source generation, native compilation, runtime behavior and full
acceptance are separate obligations. The [implementation evidence](IMPLEMENTATION.md)
records current boundaries and revision-specific verification; a recorded result
does not establish the status of a later checkout. The
[current integrity and amendment policy](../../docs/development-harness.md#current-integrity-and-amendments)
protects scenarios, assertions and expected bytes.

## Fixture layout

- [Lyger](../../lyger/test): source composition and frontend runtime inputs.
- [CLI](../../cmd/l/test): compiler result and native build inputs.
- [Phase 1](../phase1): lifecycle, continuation, language and shutdown inputs.
- [Distribution](../distribution): resource/API and source-observer inputs.
- [App processes and Service boundaries](../service-execution/coverage.md): typed
  builtins, structural Service rejection, App-constructed libraries and raw AML controls.

Go test runners remain in their packages to inspect the Go-implemented compiler
and supervise native processes. Shared tests, demos and the
[backend corpus](../../logiceme/testdata/lucid/README.md) exercise their own
boundaries. There is no fallback fixture search.

## Test boundaries

Frontend tests call the actual `LygerService.Compile`. Handoff tests send
its unmodified AML to `LogicemeService.Evaluate`. CLI source
tests separately call the existing `compiler.process()` without selecting a
backend on its behalf. Installed-CLI tests build L first, then deny child Go
tool invocations while recording and delegating actual Cargo commands.

Cargo's [metadata and compiler-artifact interface](https://doc.rust-lang.org/cargo/reference/external-tools.html)
establishes package/target identities and executable provenance. The observer
retains artifacts before the CLI can remove its build directory, compares
published hashes, checks native architecture, and rejects Go build information.
It never manufactures a generated file or executable. Manifests, locks and
source returned by production are not repaired or normalized. Direct backend
bundles retain their existing artifact inventory. Native tests attach a retained
test lock only after comparing the actual manifest with its independently
authored expectation; they never resolve dependencies during a test. Runtime-only
standalone checks remove compiler tools from `PATH`.

The Rust runtime driver compiles beside the actual canonical runtime/process
assets. Its private `runtime::test_support` contract observes real production
operations and specified fault injection; it is not a substitute runtime. Interface failures block the
corresponding behavioral observations and remain failures. External Go
controllers continue to supervise real processes, readiness barriers and
lifetime locks. Native helper validity is checked independently; that success
does not establish runtime behavior.

Apps own process lifecycle and continuations. Standalone Service declarations
produce no packages; their body effects and native lifecycle/runtime members
reject. App-constructed Service libraries retain their existing topology. The
[coverage map](../service-execution/coverage.md) separates App runtime observers,
structural Service rejection and raw AML library/native-consumer controls.

All 98 runtime registry scenarios and the six additional concurrency scenarios
remain mandatory. Ordinary Rust tests do not establish race-detector or sanitizer
coverage: safe Rust prevents data races but not all
[lifecycle races or deadlocks](https://doc.rust-lang.org/nomicon/races.html).
Rust-native string and `panic_any(())` panics exercise propagation and cleanup
obligations. The Go compiler/API/provenance tests retain their own host race checks.

## Evidence and verification

Ordinary bundle records use a unique harness-run namespace (or a standalone
process namespace), write each record once, and log its SHA-256. The
`L_HARNESS_EVIDENCE_DIR` override selects the storage root, not a reusable
filename. Evidence is diagnostic material, never an expected artifact or compiler
input.

For controlled rebuild investigations, set `L_HARNESS_CARGO_EVIDENCE_DIR` to a
private ignored directory. It retains exact staged inputs, identities and Cargo
results before temporary cleanup. Incomplete captures cannot support a rebuild
conclusion. Leave global diagnostic capture and fingerprint logging unset for
acceptance; tests that require artifact provenance still enable private captures.
See the [development guide](../../docs/development-harness.md) for cache ownership
and diagnostic boundaries.

The [local validation pause](../../docs/development-harness.md#temporary-validation-pause)
requires an explicit user request for full suites. Automatic Linux PR
validation is authorized for the documented PR events; manual dispatch requires
explicit authorization. For requested full validation, follow the
[development guide](../../docs/development-harness.md#full-validation-when-requested)
and run `scripts/l-check`: workflow lint, both Go lint profiles, doctor, both
unfiltered package phases and the separate tagged native matrix. The harness
assigns these frontend/runtime packages to `tests-native`; preserve the
[validation budgets](../../docs/development-harness.md#validation-budgets).
Focused commands remain investigative evidence, not substitutes.

Test-owned native crates use Rust 1.98.1, edition 2024 and retained locks; tests
use frozen resolution. Tokio 1.53.1 and nix 0.31.3 remain pinned. Original
watchdogs, exact streams, source identities and status/cleanup assertions are
retained. No sanitizer framework or additional language scenario is introduced.
Generated deployment scripts are never executed. Supervised runtime execution
here is limited to the maintained source and runtime scenarios.

## Helper and cache layout

Fixtures live directly beneath each package or scenario root. Runners and
helpers use responsibility-based names, with `internal/cargocache` for staging
and `internal/testsupport/native` (imported as `nativetest`) for independent
Cargo/provenance observers. No test results or source expectations are cached.

The [current API and cache layout](IMPLEMENTATION.md#current-apis-and-cache-layout)
describes the public entry points and retained physical cache suffix.
