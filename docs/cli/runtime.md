# Running a published program

## The executable is the orchestrator

After `l build` succeeds, invoke the printed executable path. Its source instructions execute in order. Apps run as separate OS processes owned by that orchestrator. App-constructed Service libraries retain their dependency roles; a standalone Service declaration produces no package or process.

Keep the executable and adjacent `.apps` directory together. Children inherit the environment and invocation CWD, receive EOF on stdin, and receive no synthesized arguments. No compiler installation is needed at runtime. [Runtime Build](../cli-build.md#app-execution-semantics) checks the prebuilt artifact; it does not invoke Cargo and is different from the `l build` command.

## Lifecycle and output

| Source operation | Effect |
| --- | --- |
| Build | Check the target's prebuilt artifact readiness. |
| Run | Launch, returning after launch acknowledgment or terminal startup failure. |
| Stop | Request process-group termination without joining. |
| Await | Join finalized output and cleanup; activate the invocation's continuation once. |
| Start | Source-defined App action: Build, then Run. |
| Execute | Source-defined App action: Start, then Await. |

Start and Execute retain App signatures. A completed process still occupies its invocation slot until its first Await continuation finishes. A second Run while that invocation is outstanding is an error. Aliases share the target and invocation, rather than launching a copied process.

Output is finalized raw stdout, preserving whitespace, newlines, NUL and partial failed-run output. Reading it before completion fails; the read does not wait. Await makes completion available and runs the continuation. Show writes without adding a newline. Stderr is captured separately and is not automatically forwarded; stderr alone does not imply child failure.

## Exit and limits

Normal fallthrough and bare Exit clean owned processes without running unawaited continuations. Fatal misuse or explicit Error exits unsuccessfully. SIGINT/SIGTERM abort source flow and clean up with statuses 130/143. Broken output pipes also trigger cleanup. Application failures are distinct from the compiler's 0/1/2 statuses.

There is no normal process-duration, process-count or capture-size cap in this contract. Post-exit/cancellation pipe draining has a two-second bound. Deliberately escaped process groups and uncatchable orchestrator termination are outside the cleanup guarantee. Native macOS/Linux execution is supported; Phase 2 in-code concurrency, cross-compilation and runtime argument transport are outside scope.

The complete authoritative details—including idempotent Stop, saved snapshots, launch failure, nested Await, terminal races and retained failure ordering—remain in [Building and running L programs](../cli-build.md). The [Service coverage map](../../testdata/service-execution/coverage.md) and [runtime integration ledger](../../testdata/integration/IMPLEMENTATION.md) separate supported contracts from recorded generation, compilation and runtime results.
