# App process and Service composition coverage

## Current contract

Only Apps own native Build/Run/Stop/Await, implicit Output/Error and process
continuations. Run launches without joining; Stop requests termination without
joining; Await joins finalized capture and claims its continuation once. Explicit
source actions and static members retain lexical lookup and shadowing.

Standalone Service declarations, including Main aliases and composed descriptors,
remain structural and produce no workspace packages. Standalone Service body
effects and native process operations reject. Main/Basic construction, primary
Service topology and additional App-constructed Service libraries remain intact.

## Retained coverage

- Frontend controls include `TestServiceDeclarationsAreStructural`,
  `TestStandaloneServiceBodiesAreUnsupported`,
  `TestServiceLifecycleAndImplicitMembersReject`,
  `TestServiceUnsupportedTargetsAndOwners`, `TestServiceSourceActionAliasAndParameter`,
  `TestServiceOrdinaryMembersAndActionsRemainLexical`, `TestRunnableAppContinuations`
  and the `TestRetainedApp*` activation/storage regressions.
- CLI Service-named tests exercise App process behavior with the retained raw
  output, misuse, Stop/Await, cleanup, provenance and publication obligations;
  structural Service inputs have explicit no-package/rejection observers.
- `cmd/l/service_workspace_test.go` retains additional App-constructed libraries,
  their exact emitted bytes and an independent native consumer.
- `lyger/builtin_operations_contract_test.go` retains all five typed process
  definitions, selected-filesystem provenance, malformed/missing resource
  diagnostics, isolated receivers, renamed construction descriptors and App parity.
- `logiceme/runnable_service_composition_test.go` retains independently authored
  raw AML library/bin compositions, exact artifacts, separate Expected/Generated
  compilation and private-access rejection. These are backend controls; their
  historical names do not expose an L Service execution feature.

<a id="private-entry-coverage"></a>

Raw AML private-function and private-module controls still require E0603 at the
explicit native consumer call. They do not imply an L `Private` entry syntax.

The [finite amendment](../../docs/_meta/service-executable-amendment.json) records
exact protected scope. All sixteen original required outcomes, frozen segments,
raw backend oracles, native runtime bytes, tool pins and watchdogs remain.
Generation, compilation, runtime and full acceptance are separate obligations.

## October 15, 2026 validation

The [macOS ARM64 report](validation-2026-10-15.md) records successful canonical
acceptance: all 4,383 named outcomes in both unfiltered phases, all sixteen
original required outcomes, and 508 integrity checks before and after execution,
including 386 protected hashes and eleven required-absent paths. Both publication
regressions reached their final preservation/runtime assertions. Workflow lint,
both complete Go lint profiles, generation, native runtime and the separate tagged
compilation stage passed. This documentation update did not rerun those stages.

<a id="stop-without-joining"></a>

## Direct App Stop observer

`TestRunnableServiceStopWithoutJoining` in `cmd/l/runnable_service_stop_test.go`
uses `cmd/l/test/runnable_service/stop_pipe.l`. The CLI helpers build and run the
actual generated and published Apps and orchestrator.

The observer transfers a duplicate of Worker's captured stdout pipe to the Go
controller through a test helper and a Unix socket. The helper exits and is reaped
before Worker readiness; the controller owns the retained writer. A separate
gated App lets the controller establish Worker readiness before the orchestrator
reaches Stop. The orchestrator emits `stop-returned|`, awaits Worker, shows its
Output and emits `joined|`.

Only after observing `stop-returned|` does the controller write `late-capture\n`
through the held writer and close it. Success requires exact final stdout
`stop-returned|late-capture\njoined|`, a successful late write, exit status zero
and empty stderr. Stop therefore returned before the capture pipe reached EOF,
while Await's finalized Output contains the later bytes. A timed-out capture
cannot contribute the required token. The controller observes process cleanup
before fallback cleanup and retains the execution watchdog.

## Verification policy

The [local validation pause](../../docs/development-harness.md#temporary-validation-pause)
requires an explicit request for local bootstrap, full lint, full ordinary/native
suites and `scripts/l-check`. Automatic Linux PR validation is authorized on
creation, new commits and reopening; manual dispatch requires explicit
authorization. Follow the [development guide](../../docs/development-harness.md#full-validation-when-requested)
for requested full validation. Focused feature selectors include
`^(TestBuiltinProcessLogeme|TestService|TestStandaloneService|TestRunnableApp|TestRetainedApp|TestRunnableService|TestConstructedServicesCompileAsLibraries)` and the
separate distribution selector `^TestRelease(Inventory|SourceResources)$`.
Focused results do not replace unfiltered acceptance or tagged native compilation.
