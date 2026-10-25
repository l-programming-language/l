# Manage lifecycle and captured output

The orchestrator owns separate App processes. Await makes completion and continuation effects explicit.

<!-- diagram:runtime-lifecycle -->

## Choose the operation by its effect

| Form | Effect |
| --- | --- |
| Build | Check an already compiled companion; no runtime compilation |
| Run | Launch or record startup failure; return after launch acknowledgment |
| Stop | Request forceful process-group termination without joining |
| Await | Join the finalized result; activate its continuation once |
| Start | Source-defined Build then Run |
| Execute | Source-defined Start then Await |
| Output | Read finalized raw stdout without waiting |
| Error | Inspect the selected operation/invocation error |

Start and Execute accept `any App`.

## Keep an invocation outstanding until Await finishes

Process exit alone does not permit another Run: the first Await continuation must finish. Duplicate Run, Build while outstanding, Await/Stop before Run and unavailable Output reads are fatal misuse. Repeated Await returns the cached immutable result without repeating continuation effects. Aliases share the target and invocation slot.

<!-- example:repeated-activation -->

Worker aliases Repeat. The first Await runs the continuation's two Shows; the second invocation gets fresh locals and repeats them. The final two Shows read the earlier snapshot and latest Output separately. The retained assertion is six `Repeat\n` lines.

## Output and errors are different streams

Output preserves raw stdout, including NUL, whitespace and Unicode. Stderr is captured separately; it is neither forwarded automatically nor itself evidence of failure. Child failure becomes a stored result rather than an implicit orchestrator Exit.

The [Stop witness](../../testdata/phase1/language/lang-10.l) declares Worker and gives its continuation this condition: ``If Worker has an Error, show `error|`, else show `success|`.`` It then performs:

<!-- example:runtime-stop-await -->

Stop returns before Await joins and activates that condition. A short-lived child may finish first, so either success or termination can win; repeated Await does not repeat the check. A Stop delivery error and the invocation result are distinct: Await reselects the invocation's immutable result.

Finalized Output can be read before Await without activating the continuation. Use Await for deterministic completion. Print's newline is retained; the [newline-mismatch demo](../../demo/paragraphs/main.l) shows why trimming changes comparisons.

## Nested work belongs to the continuation

<!-- example:nested-continuation -->

A continuation can Run/Await another target, but cannot rerun its own active target. Inline ordering produces `Other\nInner\nfinisheddone`.

## Termination and cleanup

Stop requests SIGKILL; Await still joins. Fallthrough and bare Exit clean owned processes without running unawaited continuations. Bare Exit succeeds if cleanup succeeds; explicit errors/fatal misuse exit 1. SIGINT/SIGTERM abort source flow and select 130/143.

There is no normal duration timeout, process-count limit or capture cap. The two-second post-exit/cancellation pipe-drain bound is a cleanup rule. Escaped process groups and uncatchable orchestrator termination are outside the guarantee.

The [runtime contract](../cli-build.md#app-execution-semantics) details startup, error selection and cleanup precedence. Operand types are checked earlier:

<!-- example:invalid-await-text -->

Text cannot be Awaited; the [negative fixture](../../testdata/phase1/n01-await-text/case.json) requires TypeMismatch and no new executable publication.
