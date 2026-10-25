# Manage lifecycle and captured output

The orchestrator owns separate OS processes for Apps. Ordered operations make launch, completion and continuation effects explicit. This chapter explains the maintained runtime contract; it does not claim that native verification ran during website construction.

<!-- diagram:runtime-lifecycle -->

## Choose the operation by its effect

| Form | Effect |
| --- | --- |
| Build | Check readiness of an already compiled companion; no compilation at runtime |
| Run | Launch or record a terminal startup failure; return after launch acknowledgment |
| Stop | Request forceful termination of the owned process group; do not join |
| Await | Join the finalized result and claim the continuation once |
| Start | Source-defined App action: Build then Run |
| Execute | Source-defined App action: Start then Await |
| Output | Read finalized raw stdout; no implicit wait |
| Error | Inspect the selected operation/invocation error |

Start and Execute retain `any App` signatures.

## Keep an invocation outstanding until Await finishes

One invocation remains outstanding until its first Await continuation completes. Process exit alone does not free the target for another Run. Duplicate Run, Build while an invocation is outstanding, Await/Stop before Run, and reading unavailable Output are fatal lifecycle misuse.

Repeated Await returns the cached immutable result and does not repeat continuation effects. A new accepted invocation receives fresh continuation locals. Aliases share the same target and invocation slot, so renaming a value does not create another independently running instance.

The maintained repeated-activation example demonstrates both retained snapshots and a later invocation:

<!-- example:repeated-activation -->

The existing paragraph witness expects six `Repeat\n` lines. Each Show reads the particular stored or current value expressed by the source; the compiler does not rewrite all snapshots to the latest Output.

## Output and errors are different streams

Output preserves raw stdout, including NUL, whitespace and Unicode. Stderr is captured separately and is not forwarded automatically. Stderr alone does not imply failure. Ordinary child failure is a stored result, not an implicit Exit of the orchestrator.

Reading finalized Output before Await is possible once capture and cleanup are finished, but it does not run the continuation. Use Await when the next operation requires completion rather than trying to predict a race. The [intentional newline-mismatch demo](../../demo/paragraphs/main.l) demonstrates why captured text must not be trimmed for comparisons.

## Nested work belongs to the continuation

<!-- example:nested-continuation -->

A continuation may Run and Await another target. It cannot Run the same target whose continuation is still active. The retained expected output here is `Other\nInner\nfinisheddone`, showing inline continuation ordering rather than a background callback.

## Termination and cleanup

Stop requests SIGKILL and still requires Await to join. Normal fallthrough and bare Exit clean up owned processes without running unawaited continuations. Explicit error/fatal misuse exits 1 after cleanup; SIGINT/SIGTERM abort source flow and select 130/143. A successful bare Exit selects 0 if cleanup succeeds.

The runtime imposes no normal duration timeout, process-count limit or output-capture cap. Its two-second post-exit/cancellation pipe-drain bound is a cleanup rule. Deliberately escaped process groups and uncatchable orchestrator termination are outside the guarantee.

The [full runtime guide](../cli-build.md#app-execution-semantics) defines operation-error selection, startup failure, races and cleanup precedence. For a deliberate compiler error example:

<!-- example:invalid-await-text -->

That Text is not an App; the [negative fixture](../../testdata/phase1/n01-await-text/case.json) requires a TypeMismatch and no new executable publication.
