# Lifecycle and observation builtins

Build, Run, Stop, and Await act on checked process targets. Show and Exit act in the outer generated program. The full maintained contract is [Building and running L programs](../../cli-build.md#app-execution-semantics).

## Build

<!-- word:builtin:build -->

Check that an App's prebuilt executable is available, recording the operation result without launching it.

**Usage and example.** `Build Worker` in the [explicit lifecycle example](../../spec/execution.md#build-run-stop-and-await) precedes Run; it is not the CLI compilation command.

Build accepts an App. At runtime it checks the already published executable and selects its operation error. It does not invoke Cargo. CLI `l build` is the separate compilation/publication command.

Build creates no invocation. Building while an invocation remains outstanding is misuse. A later Run revalidates its artifact, including after a Build failure. Use an explicit Error check when handling an operation result rather than assuming Build has launched a process.

## Run

<!-- word:builtin:run -->

Launch an App invocation in a separate operating-system process without joining it.

**Usage and example.** Start in the [two-App program](../../../demo/phase1/main.l) composes Run for each App before either Await. Use [Await](#await) for finalized output.

Run launches the selected target in a separate OS process and returns after process creation acknowledgment or recorded terminal startup failure. It waits for neither readiness nor exit/capture completion. Each target has one canonical outstanding invocation slot, shared by aliases.

Run before Await permits overlap between targets. Running a target twice while outstanding, or during its own claimed continuation, rejects. A startup failure remains an awaitable result. Children inherit environment and CWD, get EOF on stdin, and receive no synthesized arguments.

## Stop

<!-- word:builtin:stop -->

Request termination of an App's owned process group without waiting for finalized completion.

**Usage and example.** The [Stop/Await witness](../../spec/execution.md#build-run-stop-and-await) shows `stop-requested|` before joining. A natural exit may win the race; always Await when completion is needed.

Stop requests SIGKILL of the owned process group and returns without joining. A natural exit racing Stop must still produce one coherent result. Stop before Run is misuse; Stop after finalized completion is an idempotent no-op. A signal-delivery failure selects a recoverable operation error and preserves ownership for subsequent Await/cleanup.

Use Await to obtain finalized output and cleanup state after Stop. Do not teach Stop as a synchronous cancellation-and-join primitive. The retained Stop observer exercises this distinction with an App.

## Await

<!-- word:builtin:await -->

Join an App invocation's finalized result and activate its continuation once.

**Usage and example.** The [continuation example](../../spec/execution.md#continuations-and-activation) awaits Worker twice; only the first join runs the paragraph. `Worker is awaited` is the supported outer passive form.

Await joins process reaping, raw stream finalization, and owned-group cleanup. The first Await claims the invocation continuation and executes it inline once. Failed launch, nonzero child exit, and per-target Stop still permit that continuation. Repeated/self-Await returns the cached result without repeating effects.

The invocation slot remains occupied until the first continuation finishes. A nested Await of another target is allowed; Run of the currently claimed target is not. Global shutdown aborts source flow instead of reentering a continuation. The imperative, outer passive, and source-action applicative forms are supported; native App-body Await rejects ownership; standalone Service body effects reject.

## Start and Execute

These source-defined actions share the App-only lifecycle contract.

### Start

<!-- word:builtin:start -->

Build an App and then launch its native process without waiting for completion.

**Usage and example.** `Start First and Second` applies the whole Build/Run composition to each App in order. The [two-App program](../../../demo/phase1/main.l) starts both before joining; Start does not promise application readiness.

### Execute

<!-- word:builtin:execute -->

Start an App and await its finalized result and first continuation before proceeding.

**Usage and example.** `Execute Hello` in the [first program](../../guide/first-program.md#read-the-program) supplies the completion boundary before `Show Hello's Output`. Its source parameter remains `any App`; no explicit Runtime import is needed.

A source-defined action with coordinated Apps performs its complete body for the first App, then the next. Thus coordinating Execute has a different overlap pattern from coordinating Run followed by coordinating Await. That ordering follows source expansion, not parallel implicit execution.

## Output and Error

These members refer to a selected App, unlike the [IO.Output handle](../../stdlib/eng/io.md#output).

### Output

<!-- word:builtin:output -->

Read the finalized raw standard-output bytes of an App invocation as Text.

**Usage and example.** `Show Hello's Output` in the [first program](../../guide/first-program.md#read-the-program) follows Execute. Assigning Output makes a snapshot, demonstrated by the [repeated-activation example](../../../demo/paragraphs/repeated-activation.l). Function signature `Outputs` names result types instead; see [Function](definitions.md#function).

Output is raw finalized stdout. Reading it before finalization is fatal misuse; no hidden join occurs. Captured stderr stays separate and does not automatically mean failure. Output may be available before Await if the invocation has already finalized, but reading it does not activate its continuation.

### Error

<!-- word:builtin:error -->

Select an App's current operation or invocation failure for inspection.

**Usage and example.** `it has an Error` in the [conditional member example](../../../lyger/test/member-resolution.l) checks the declared HelloWorldApp after Execute. Extracting an absent Error rejects; this member is distinct from the [Error value type](../../spec/values.md#error).

Error identifies the selected operation or invocation failure. Build/Stop can select operation errors; Await reselects its immutable invocation result. `target has an Error` checks matching owner identity. Stored Text snapshots survive later runs unchanged.

## Show

<!-- word:builtin:show -->

Write a supported runtime value to the outer program's output without adding a newline.

**Usage and example.** The [public-values program](../../../demo/paragraphs/public-values.l) shows Text, Number and Truth as adjacent bytes. The [first program](../../guide/first-program.md#read-the-program) shows finalized App Output.

Show writes Text exactly and formats supported runtime values without adding LF. Direct App and Error display rejects. Show a target's Output only when it is available. It does not replay other targets' streams. This distinguishes it from imported IO.Print, whose source composition appends one LF after the payload succeeds.

The [first program](../../guide/first-program.md) uses Execute then Show. Its stdout is `Hello World!` followed by LF, supplied by Print.

## Exit

<!-- word:builtin:exit -->

Terminate the whole generated L program, cleaning up its owned processes.

**Usage and example.** The [bare Exit example](../../spec/control-flow.md#exit-and-return) ends execution inside a continuation; later actions do not run. [Function.Return](definitions.md#functionreturn) instead completes one invocation.

Bare Exit cleans all owned processes and exits successfully if cleanup succeeds. Exit with a supported explicit failure/Error operand selects failure. It terminates the entire L program even inside a continuation or source-expanded action.

Fallthrough/Exit cleanup does not activate unawaited continuations. Broken outer output writers trigger cleanup and failure. Signal and cleanup details, including 130/143 and bounded post-exit pipe draining, remain in the maintained runtime contract.

## Sources and witnesses

Sources: [typed operations](../../../lyger/builtin_operations.go), [instructions](../../../lyger/instructions.go), [runtime plan](../../../lyger/orchestration_plan.go), [core source composition](../../../builtins/eng/core.l), and [runtime contract](../../../testdata/phase1/runtime/contract.md). Witnesses: [phase-one language](../../../lyger/phase1_language_test.go), [nested Await](../../../lyger/phase1_nested_await_test.go), [paragraph examples](../../../demo/paragraphs/README.md), and [Stop without joining](../../../cmd/l/runnable_service_stop_test.go).
