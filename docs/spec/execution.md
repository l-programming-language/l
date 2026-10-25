# Execution and lifecycle

L applications execute as separate native OS processes owned by an orchestrator. This chapter summarizes the maintained [CLI/runtime contract](../cli-build.md#app-execution-semantics); that contract owns the detailed runtime guarantees.

## Compilation is not execution

Lyger compiles source into AML/logeme compositions. Logiceme Evaluate returns an Executable source bundle; the name does not mean a running application. The native coordinator validates and builds those sources. `l build` publishes output but does not run Apps or generated deployment scripts.

After publication, invoking the generated entry executable executes the source's outer instructions. Runtime Build checks the prebuilt process artifact; it does not invoke Cargo or another compiler. Keep this language operation distinct from the CLI command with the same English word.

## Targets and process ownership

An App is a process target. Standalone Service declarations remain structural and produce no package or process. Native lifecycle operations and implicit runtime members require an App; lexical source actions and explicit members retain precedence. Classes and Objects organize sources without process lifetimes.

Children inherit the caller's environment and CWD, receive EOF on stdin, and receive no synthesized runtime arguments. Runtime argument transport and process return-value transport are outside the current contract. The process result carries raw captured output and error state instead.

**App lifecycle example.**

<!-- example:runnable-app -->

The Basic App declaration supplies Worker's executable body through ordinary Main/Basic construction. Run launches its process, and Await joins it before Show reads Output. Expected stdout is `work completed\n`. Standalone Service declarations and bodies do not supply executable targets.

## Build, Run, Stop, and Await

| Operation | Completion boundary |
| --- | --- |
| Build | Validate the selected prebuilt executable and select its operation error |
| Run | Acknowledge process creation or record terminal startup failure; do not join |
| Stop | Request termination of the owned process group or observe completion; do not join |
| Await | Join reaping, captured streams, and owned-group cleanup; return immutable result and activate continuation once |

Run does not wait for application readiness, normal completion, output finalization, or continuation effects. Stop requests SIGKILL of the owned group and returns after request submission/known completion. The process leader exiting while output drain or group cleanup remains pending is not finalized completion.

Start and Execute are source-defined App actions. Start composes Build then Run; Execute composes Start then Await. Their `any App` signatures remain intentional.

**Supported explicit lifecycle — excerpt, with a declared Worker required.** The App [lifecycle witness](../../cmd/l/test/runnable_service/basic.l) contains these ordered operations:

<!-- example:runtime-build-run -->

Build checks Worker's already compiled companion. Run starts it; Await is the completion boundary; Show observes its final output. This excerpt omits the witness's declarations and test-owned native printing probe, so it is not an independent source file. The complete canonical-IO example above demonstrates the public printing path.

**Stop is a request, then Await joins — excerpt from the existing race-aware witness.** Its preceding paragraph declares Worker and supplies a continuation that inspects Error:

<!-- example:runtime-stop-await -->

`stop-requested|` is shown after Stop returns and before the first Await. Only that Await activates the continuation; the second Await does not repeat it. A short-lived Worker can finish before the termination request wins, so the [native witness](../../acceptance/phase1/language/language_test.go) permits either `stop-requested|continuation|error|joined|` or `stop-requested|continuation|success|joined|`. The excerpt is not a promise of one deterministic Stop outcome and was not executed here.

## Invocations and ordering

Each canonical target has one outstanding invocation slot. Completion alone does not release that slot; the first Await continuation must finish. Duplicate Run, Build while outstanding, Await/Stop before Run, and rerunning a target during its own active continuation are lifecycle misuse.

A startup failure still creates an awaitable failed invocation; an unstarted command has no OS Wait. Ordinary child nonzero exit is stored as a result and does not automatically Exit the orchestrator. An alias shares its target's invocation slot.

A later accepted Run creates a new invocation. Previously captured results and assigned scalar Output snapshots remain unchanged. A Stop captures its target invocation before signaling so a race cannot redirect the signal to a replacement invocation. A completed Stop is an idempotent no-op.

## Continuations and activation

An ordinary encapsulated paragraph executes when source flow reaches it. An App-led paragraph associates its continuation with that target. The first Await of an invocation claims and executes the continuation inline exactly once, including failed launch, nonzero exit, or per-target Stop.

Repeated/cached Await observes the same result and cannot claim the continuation again. Self-Await while the claim is active likewise cannot repeat it. Nested Await of a different target is allowed. Fresh accepted invocations receive fresh continuation locals; lexical paragraph identity does not suppress another target's claim.

A continuation Exit terminates the whole program. Global shutdown aborts source flow instead of returning from public Await into a continuation. Fallthrough cleanup does not execute continuations for unawaited invocations.

**Supported first-Await activation.**

<!-- example:runtime-continuation-once -->

Indentation starts the App-led source paragraph; the blank line before Run ends it. The App's Print runs in its child process. The paragraph's two Shows are continuation effects: they run inline at the first Await, between `launched|` and `joined|`. The final Await reads the cached result without repeating `continued|`. The retained native assertion is `launched|continued|child\njoined|again|`.

This paragraph illustrates one invocation. The [snapshot example](control-flow.md#assignment) shows an accepted later Run, and the [repeated-activation example](../guide/lifecycle.md#keep-an-invocation-outstanding-until-await-finishes) demonstrates fresh continuation locals on each invocation.

## Output and Error

Output is raw stdout, including whitespace, newlines, NUL, and arbitrary bytes. Stderr is captured separately and is not automatically replayed. Stderr alone does not indicate failure. Print inside a child commonly adds LF; Show in the orchestrator preserves the captured result as-is.

Reading Output before capture finalization is fatal misuse, even when assigning it to an otherwise unused variable. There is no implicit Await. If an invocation has already finalized, Output can be read before Await without activating its continuation. Programs requiring deterministic completion should explicitly Await.

Error selects the relevant operation/invocation error. Build selects its operation error; a Stop signal-delivery failure selects an error while retaining the invocation and ownership. A subsequent Await, including a cached Await, reselects that invocation's immutable result. The typed condition `target has an Error` checks this state; it does not expose the Go compiler's lerrors package as an L exception API.

## Shutdown and cleanup

Normal fallthrough and bare Exit clean owned processes and return 0 if cleanup succeeds. Explicit Error/fatal misuse cleans up and returns 1. SIGINT/SIGTERM abort source flow, settle ownership, and use 130/143. A cleanup failure on otherwise successful exit produces failure; explicit primary errors are retained and cleanup errors are appended once in launch order.

Broken stdout or stderr pipes require cleanup and failure. Error reporting must not recursively retry a failed diagnostic writer. One terminal request selects the primary diagnostic/status when fatal and signal requests race; cleanup does not turn a native panic into success.

Each launched direct child has one Wait owner. Same-group descendants are owned even after closing inherited pipes. There is a bounded two-second post-exit/cancellation pipe drain, but no normal process-duration timeout, process-count cap, or capture-size cap. Deliberately escaped process groups, uncatchable SIGKILL of the orchestrator, and host/power loss remain outside the cleanup guarantee.

## Platform and concurrency limits

The supported native hosts are macOS and Linux. Multiple launched processes can overlap while the orchestrator executes ordered instructions. Phase 2 in-code concurrency, general async source functions, cross-compilation, remote deployment, containers, and unrestricted third-party installation are outside the contract.

Tokio is internal runtime implementation, not a public L concurrency standard library. Rust Deferred metadata is rejected; process cleanup is explicit runtime behavior and must not be explained as automatic Deferred execution.

## Sources and witnesses

Normative maintained source: [CLI/runtime contract](../cli-build.md). Implementation: [typed process instructions](../../lyger/instructions.go), [builtin process operations](../../lyger/builtin_operations.go), [paragraph instructions](../../lyger/paragraph_instructions.go), and [orchestration flow](../../lyger/orchestration_flow.go). Witnesses: [phase-one suite](../../lyger/phase1_language_test.go), [paragraph examples](../../demo/paragraphs/README.md), [Service coverage](../../testdata/service-execution/coverage.md), and [runtime contract](../../testdata/phase1/runtime/contract.md). Recorded runtime outcomes apply to their tested revisions and hosts.
