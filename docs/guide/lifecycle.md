# Manage lifecycle and captured output

Apps run in separate processes. Use Run to start one, Await to wait for completion, and Show to display its captured Output.

<!-- diagram:runtime-lifecycle -->

## Choose the operation by its effect

| Form | Effect |
| --- | --- |
| Build | Check an already compiled companion; no runtime compilation |
| Run | Start an App without waiting for completion |
| Stop | Request forceful termination without waiting for completion |
| Await | Join the finalized result; activate its continuation once |
| Start | Source-defined Build then Run |
| Execute | Source-defined Start then Await |
| Output | Read finalized raw stdout without waiting |
| Error | Inspect an operation or App failure |

Start and Execute accept `any App`.

## Keep an invocation outstanding until Await finishes

Wait for the first Await continuation to finish before running the App again. Repeated Await returns the same result without repeating the continuation. Aliases refer to the same App and share its current run.

<!-- example:repeated-activation -->

Worker aliases Repeat. Each run gets fresh continuation locals. The final two Shows read a saved snapshot and the latest Output separately. Together, the program writes six `Repeat\n` lines.

## Output and errors are different streams

Output preserves stdout bytes and newlines. Stderr is captured separately and is not forwarded automatically. An App failure becomes a stored result that the program can inspect.

This [example](../../testdata/phase1/language/lang-10.l) gives Worker a continuation that shows `error|` or `success|`, then performs:

<!-- example:runtime-stop-await -->

Stop does not wait: Await still joins the result and runs the continuation. A short-lived App can finish before Stop reaches it, so either success or termination is possible.

Reading Output does not wait. Use Await first to ensure completion, and remember that Print's trailing newline remains in Output.

## Nested work belongs to the continuation

<!-- example:nested-continuation -->

A continuation can Run/Await another target, but cannot rerun its own active target. Inline ordering produces `Other\nInner\nfinisheddone`.

## Termination and cleanup

Normal program completion and bare Exit clean up owned processes without running unawaited continuations. See the [execution reference](../spec/execution.md) for failure, signal and cleanup rules.

Lifecycle operations require an App. This example is rejected before execution:

<!-- example:invalid-await-text -->

Text cannot be Awaited. The compiler reports TypeMismatch and publishes no new executable.
