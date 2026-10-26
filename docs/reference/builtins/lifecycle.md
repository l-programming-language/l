# Lifecycle and observation builtins

Build, Run, Stop, and Await act on checked process targets. Show and Exit act in the outer generated program. The full maintained contract is [Building and running L programs](../../cli-build.md#app-execution-semantics).

## Build

Build accepts an App. At runtime it checks the already published executable and selects its operation error. It does not invoke Cargo. CLI `l build` is the separate compilation/publication command.

Build creates no invocation. Building while an invocation remains outstanding is misuse. A later Run revalidates its artifact, including after a Build failure. Use an explicit Error check when handling an operation result rather than assuming Build has launched a process.

## Run

Run launches the selected target in a separate OS process and returns after process creation acknowledgment or recorded terminal startup failure. It waits for neither readiness nor exit/capture completion. Each target has one canonical outstanding invocation slot, shared by aliases.

Run before Await permits overlap between targets. Running a target twice while outstanding, or during its own claimed continuation, rejects. A startup failure remains an awaitable result. Children inherit environment and CWD, get EOF on stdin, and receive no synthesized arguments.

## Stop

Stop requests SIGKILL of the owned process group and returns without joining. A natural exit racing Stop must still produce one coherent result. Stop before Run is misuse; Stop after finalized completion is an idempotent no-op. A signal-delivery failure selects a recoverable operation error and preserves ownership for subsequent Await/cleanup.

Use Await to obtain finalized output and cleanup state after Stop. Do not teach Stop as a synchronous cancellation-and-join primitive. The retained Stop observer exercises this distinction with an App.

## Await

Await joins process reaping, raw stream finalization, and owned-group cleanup. The first Await claims the invocation continuation and executes it inline once. Failed launch, nonzero child exit, and per-target Stop still permit that continuation. Repeated/self-Await returns the cached result without repeating effects.

The invocation slot remains occupied until the first continuation finishes. A nested Await of another target is allowed; Run of the currently claimed target is not. Global shutdown aborts source flow instead of reentering a continuation. The imperative, outer passive, and source-action applicative forms are supported; native App-body Await rejects ownership; standalone Service body effects reject.

## Start and Execute

Start expands to Building any App and then Running it. Execute expands to Starting any App and then Awaiting it. Both retain a source `any App` parameter. Native lifecycle operations are also App-only.

A source-defined action with coordinated Apps performs its complete body for the first App, then the next. Thus coordinating Execute has a different overlap pattern from coordinating Run followed by coordinating Await. That ordering follows source expansion, not parallel implicit execution.

## Output and Error

Output is raw finalized stdout. Reading it before finalization is fatal misuse; no hidden join occurs. Captured stderr stays separate and does not automatically mean failure. Output may be available before Await if the invocation has already finalized, but reading it does not activate its continuation.

Error identifies the selected operation or invocation failure. Build/Stop can select operation errors; Await reselects its immutable invocation result. `target has an Error` checks matching owner identity. Stored Text snapshots survive later runs unchanged.

## Show

Show writes Text exactly and formats supported runtime values without adding LF. Direct App and Error display rejects. Show a target's Output only when it is available. It does not replay other targets' streams. This distinguishes it from imported IO.Print, which appends one LF inside its native action.

The maintained basic program uses Execute then Show:

<!-- example:basic -->

Its expected stdout is `Hello World!` followed by LF, because Print supplied that newline. The linked fixture records its expected output.

## Exit

Bare Exit cleans all owned processes and exits successfully if cleanup succeeds. Exit with a supported explicit failure/Error operand selects failure. It terminates the entire L program even inside a continuation or source-expanded action.

Fallthrough/Exit cleanup does not activate unawaited continuations. Broken outer output writers trigger cleanup and failure. Signal and cleanup details, including 130/143 and bounded post-exit pipe draining, remain in the maintained runtime contract.

## Sources and witnesses

Sources: [typed operations](../../../lyger/builtin_operations.go), [instructions](../../../lyger/instructions.go), [runtime plan](../../../lyger/orchestration_plan.go), [core source composition](../../../builtins/eng/core.l), and [runtime contract](../../../testdata/phase1/runtime/contract.md). Witnesses: [phase-one language](../../../lyger/phase1_language_test.go), [nested Await](../../../lyger/phase1_nested_await_test.go), [paragraph examples](../../../demo/paragraphs/README.md), and [Stop without joining](../../../cmd/l/runnable_service_stop_test.go).
