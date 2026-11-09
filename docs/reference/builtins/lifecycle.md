# Lifecycle and observation builtins

Build, Run, Stop, and Await act on checked process targets. Show and Exit act in the outer generated program. The full maintained contract is [Building and running L programs](../../cli-build.md#app-execution-semantics).

## Build

<!-- word:builtin:build -->

Check that an App's prebuilt executable is available, recording the operation result without launching it.

**Usage and example.** `Build Worker` in the [explicit lifecycle example](../../spec/execution.md#build-run-stop-and-await) precedes Run. It is not the CLI compilation command.

Use an explicit Error check when handling the result. Build creates no invocation, and building while one remains outstanding is misuse. Run revalidates the artifact, including after Build failure. See [invocation and result rules](../../cli-build.md#invocation-and-results).

## Run

<!-- word:builtin:run -->

Launch an App invocation in a separate operating-system process without joining it.

**Usage and example.** Start in the [two-App program](../../../demo/phase1/main.l) composes Run for each App before either Await. Use [Await](#await) for finalized output.

Run returns after process creation acknowledgment or recorded startup failure. Starting several Apps before Await permits overlap. A startup failure is still awaitable; running an outstanding target again rejects. The [launch contract](../../cli-build.md#launch-and-language-forms) and [invocation rules](../../cli-build.md#invocation-and-results) define acknowledgment, aliases and reruns.

## Stop

<!-- word:builtin:stop -->

Request termination of an App's owned process group without waiting for finalized completion.

**Usage and example.** The [Stop/Await witness](../../spec/execution.md#build-run-stop-and-await) shows `stop-requested|` before joining. A natural exit may win the race, so always Await when completion is needed.

Stop before Run is misuse; after finalized completion it is an idempotent no-op. Signal-delivery failure selects a recoverable operation Error and preserves ownership. Use Await for finalized output and cleanup state. The [shutdown contract](../../cli-build.md#shutdown-streams-and-limits) defines group signaling and its race with natural exit.

## Await

<!-- word:builtin:await -->

Join an App invocation's finalized result and activate its continuation once.

**Usage and example.** The [continuation example](../../spec/execution.md#continuations-and-activation) awaits Worker twice, but only the first join runs the paragraph. `Worker is awaited` is the supported outer passive form.

Await joins reaping, captured streams and owned-group cleanup. Failed launches and stopped Apps still have awaitable results. The [invocation contract](../../cli-build.md#invocation-and-results) owns continuation claims, cached/self-Await, nested Await and slot release.

Imperative, outer passive and source-action applicative forms are supported. Native App-body Await rejects ownership, and standalone Service body effects reject. See [supported language forms](../../cli-build.md#launch-and-language-forms).

## Start and Execute

These source-defined actions share the App-only lifecycle contract.

### Start

<!-- word:builtin:start -->

Build an App and then launch its native process without waiting for completion.

**Usage and example.** `Start First and Second` applies the whole Build/Run composition to each App in order. The [two-App program](../../../demo/phase1/main.l) starts both before joining. Start does not promise application readiness.

### Execute

<!-- word:builtin:execute -->

Start an App and await its finalized result and first continuation before proceeding.

**Usage and example.** `Execute Hello` in the [first program](../../guide/first-program.md#read-the-program) supplies the completion boundary before `Show Hello's Output`. Its source parameter remains `any App`, and no explicit Runtime import is needed.

A source-defined action with coordinated Apps performs its complete body for the first App, then the next. Thus coordinating Execute has a different overlap pattern from coordinating Run followed by coordinating Await. That ordering follows source expansion, not parallel implicit execution.

## Output and Error

These members refer to a selected App, unlike the [IO.Output handle](../../stdlib/eng/io.md#output).

### Output

<!-- word:builtin:output -->

Read the finalized raw standard-output bytes of an App invocation as Text.

**Usage and example.** `Show Hello's Output` in the [first program](../../guide/first-program.md#read-the-program) follows Execute. Assigning Output makes a snapshot, demonstrated by the [repeated-activation example](../../../demo/paragraphs/repeated-activation.l). Function signature `Outputs` names result types instead. See [Function](definitions.md#function).

Reading Output before finalization is fatal misuse. No hidden join occurs, and an already finalized read does not activate the continuation. The [result contract](../../cli-build.md#invocation-and-results) covers availability, separate stderr capture and snapshots across reruns.

### Error

<!-- word:builtin:error -->

Select an App's current operation or invocation failure for inspection.

**Usage and example.** `it has an Error` in the [conditional member example](../../../lyger/test/member-resolution.l) checks the declared HelloWorldApp after Execute. Extracting an absent Error rejects. This member is distinct from the [Error value type](../../spec/values.md#error).

Build/Stop can select operation errors, while Await reselects its immutable invocation result. `target has an Error` checks matching owner identity. See [error selection](../../cli-build.md#invocation-and-results).

## Show

<!-- word:builtin:show -->

Write a supported runtime value to the outer program's output without adding a newline.

**Usage and example.** The [public-values program](../../../demo/paragraphs/public-values.l) shows Text, Number and Truth as adjacent bytes. The [first program](../../guide/first-program.md#read-the-program) shows finalized App Output.

Show preserves Text bytes and formats supported runtime values. Direct App and Error display rejects, and other targets' streams are not replayed. Imported [IO.Print](../../stdlib/eng/io.md#print) supplies the LF in the first program's `Hello World!` output.

## Exit

<!-- word:builtin:exit -->

Terminate the whole generated L program, cleaning up its owned processes.

**Usage and example.** The [bare Exit example](../../spec/control-flow.md#exit-and-return) ends execution inside a continuation. Later actions do not run. [Function.Return](definitions.md#functionreturn) instead completes one invocation.

Bare Exit succeeds if cleanup succeeds. Exit with a supported explicit failure/Error operand selects failure, including inside a continuation or source action. Cleanup does not activate unawaited continuations. The [shutdown contract](../../cli-build.md#shutdown-streams-and-limits) specifies status codes, signal handling and cleanup failures.

## Sources and witnesses

Sources: [typed operations](../../../lyger/builtin_operations.go), [instructions](../../../lyger/instructions.go), [runtime plan](../../../lyger/orchestration_plan.go), [core source composition](../../../builtins/eng/core.l), and [runtime contract](../../../testdata/phase1/runtime/contract.md). Witnesses: [phase-one language](../../../lyger/phase1_language_test.go), [nested Await](../../../lyger/phase1_nested_await_test.go), [paragraph examples](../../../demo/paragraphs/README.md), and [Stop without joining](../../../cmd/l/runnable_service_stop_test.go).
