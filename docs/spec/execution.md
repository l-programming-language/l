# Execution and lifecycle

L applications execute as separate native OS processes owned by an orchestrator. This chapter summarizes the maintained [CLI/runtime contract](../cli-build.md#app-execution-semantics), which owns the detailed runtime guarantees.

## Compilation is not execution

Lyger compiles source into AML/logeme compositions. Logiceme Evaluate returns an Executable source bundle. The name does not mean a running application. The native coordinator validates and builds those sources. `l build` publishes output but does not run Apps or generated deployment scripts.

After publication, invoking the generated entry executable executes the source's outer instructions. Runtime Build checks the prebuilt process artifact. It does not invoke Cargo or another compiler. Keep this language operation distinct from the CLI command with the same English word.

## Targets and process ownership

An App is a process target. Standalone Service declarations remain structural and produce no package or process. Native lifecycle operations and implicit runtime members require an App. Lexical source actions and explicit members retain precedence. Classes and Objects organize sources without process lifetimes.

The process result carries captured output and error state. Runtime argument and process return-value transport are outside the contract. See [child environment, streams and limits](../cli-build.md#shutdown-streams-and-limits).

The [complete App example](../guide/apps-and-services.md#run-an-app) supplies Worker's body through Basic construction, then runs and awaits it before reading Output. Standalone Service declarations and bodies do not supply executable targets.

## Build, Run, Stop, and Await

| Operation | Completion boundary |
| --- | --- |
| Build | Validate the selected prebuilt executable and select its operation error |
| Run | Acknowledge process creation or record terminal startup failure without joining |
| Stop | Request termination of the owned process group or observe completion without joining |
| Await | Join reaping, captured streams, and owned-group cleanup, then return immutable result and activate continuation once |

Run acknowledges launch; Await establishes completion. Stop requests termination without joining. The [launch](../cli-build.md#launch-and-language-forms) and [shutdown](../cli-build.md#shutdown-streams-and-limits) contracts define the native boundaries.

Start and Execute are source-defined App actions. Start composes Build then Run, and Execute composes Start then Await. Their `any App` signatures remain intentional.

**Supported explicit lifecycle excerpt, with a declared Worker required.** The App [lifecycle witness](../../cmd/l/test/runnable_service/basic.l) contains these ordered operations:

<!-- example:runtime-build-run -->

Build checks Worker's already compiled companion. Run starts it, Await is the completion boundary, and Show observes its final output. This excerpt omits the witness's declarations and test-owned native printing probe. Use the linked App example for the complete public IO path.

**Stop is a request, then Await joins.** This excerpt is from the existing race-aware witness. Its preceding paragraph declares Worker and supplies a continuation that inspects Error:

<!-- example:runtime-stop-await -->

`stop-requested|` is shown before the first Await, which activates the continuation. The second Await does not repeat it. Worker can finish before termination wins, so the [native witness](../../acceptance/phase1/language/language_test.go) permits either `stop-requested|continuation|error|joined|` or `stop-requested|continuation|success|joined|`.

## Invocations and ordering

Each canonical App, including its aliases, has one outstanding invocation until its first Await continuation finishes. A startup failure is still awaitable. Later accepted runs create fresh invocations without changing saved Output snapshots. The [invocation contract](../cli-build.md#invocation-and-results) defines invalid operation ordering, failed launches, reruns and Stop races.

## Continuations and activation

An ordinary encapsulated paragraph executes when source flow reaches it. An App-led paragraph associates its continuation with that target. The first Await of an invocation claims and executes the continuation inline exactly once, including failed launch, nonzero exit, or per-target Stop.

Repeated Await reuses the result without repeating effects. Fresh invocations receive fresh continuation locals. [Continuation ownership](../cli-build.md#invocation-and-results) also governs self-Await, nested Await, Exit and shutdown.

**Supported first-Await activation.**

<!-- example:runtime-continuation-once -->

Indentation starts the App-led source paragraph, and the blank line before Run ends it. The App's Print runs in its child process. The paragraph's two Shows are continuation effects: they run inline at the first Await, between `launched|` and `joined|`. The final Await reads the cached result without repeating `continued|`. The retained native assertion is `launched|continued|child\njoined|again|`.

This paragraph illustrates one invocation. The [snapshot example](control-flow.md#assignment) shows an accepted later Run, and the [repeated-activation example](../guide/lifecycle.md#keep-an-invocation-outstanding-until-await-finishes) demonstrates fresh continuation locals on each invocation.

## Output and Error

Output preserves raw stdout, including whitespace, newlines, NUL and arbitrary bytes. Await explicitly before reading it when completion must be deterministic. Print adds LF in the child; Show preserves the captured bytes. Error selects the current operation/invocation failure, tested by `target has an Error`. The [result contract](../cli-build.md#invocation-and-results) owns availability checks, separate stderr capture, error selection and stored snapshots. Compiler diagnostics use a separate [error model](errors.md).

## Shutdown and cleanup

Fallthrough and Exit settle owned processes without activating unawaited continuations. The [shutdown contract](../cli-build.md#shutdown-streams-and-limits) defines exit statuses, signal and output failures, process-group ownership, the two-second drain bound and limits of cleanup. It imposes no normal process-duration timeout, process-count cap or capture-size cap.

## Platform and concurrency limits

The supported native hosts are macOS and Linux. Multiple launched processes can overlap while the orchestrator executes ordered instructions. Phase 2 in-code concurrency, general async source functions, cross-compilation, remote deployment, containers, and unrestricted third-party installation are outside the contract.

Tokio is internal runtime implementation, not a public L concurrency standard library. Rust Deferred metadata is rejected. Process cleanup is explicit runtime behavior and must not be explained as automatic Deferred execution.

## Sources and witnesses

Normative maintained source: [CLI/runtime contract](../cli-build.md). Implementation: [typed process instructions](../../lyger/instructions.go), [builtin process operations](../../lyger/builtin_operations.go), [paragraph instructions](../../lyger/paragraph_instructions.go), and [orchestration flow](../../lyger/orchestration_flow.go). Witnesses: [phase-one suite](../../lyger/phase1_language_test.go), [paragraph examples](../../demo/paragraphs/README.md), [Service coverage](../../testdata/service-execution/coverage.md), and [runtime contract](../../testdata/phase1/runtime/contract.md). Recorded runtime outcomes apply to their tested revisions and hosts.
