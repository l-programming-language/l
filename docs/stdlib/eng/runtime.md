# Runtime

The English `runtime` dictionary supplies L-authored App supervision, error, capture and shutdown policy. It is a compiler implementation dictionary, distinct from ordinary application vocabulary. Its generated native support works with the host and platform mechanisms. Importing it does not create a new runtime or launch a process.

## Import and calling forms

Word authors can name the dictionary explicitly:

```l
Use Runtime from `runtime`.
```

Its Functions require the native types, async context and host facilities of the generated runtime module. Ordinary applications use the [builtin lifecycle words](../../reference/builtins/lifecycle.md) without importing Runtime. A source name such as Runtime.App.Run identifies an implementation Function. It is not an alternate spelling for a top-level Run instruction.

| Task | Application interface | Implementation group |
| --- | --- | --- |
| Check a prebuilt App | Build | [App](#app) |
| Launch an App | Run, or Start to compose Build then Run | [App](#app) |
| Request termination without joining | Stop | [App](#app) |
| Join an invocation and activate its continuation | Await | [App](#app), [Invocation](#invocation) |
| Build, launch and join | Execute | [App](#app), [Invocation](#invocation) |
| Read finalized output or failure | App Output and Error | [App](#app), [Error](#error) |
| Display a supported value | Show | [Output and shutdown](#output-and-shutdown), [Format](#format) |
| Settle process ownership on exit | Exit or normal completion | [Output and shutdown](#output-and-shutdown) |

## Public and private names

The groups below describe native support interfaces, not a public application-library catalog. Public source roots let the compiler compose selected definitions and generate their native symbols. Private helpers retain definition-site scope. Runtime.Host is a private adapter-root namespace. Its selected aliases serve startup and native traits, rather than exposing another application import surface.

Runtime privately imports required Storage words and the canonical Formatting implementation. Authored L policy lives under `runtime/eng/`, native entry, host and OS inputs live under `runtime/rust/`, and derived AML and provenance live under `runtime/generated/`. See [support assembly](../../cli-build.md#source-generation-and-native-builds) for selection and freshness rules.

## App

Coordinate a prebuilt App executable and its current invocation.

| Member | Native inputs and result | Behavior |
| --- | --- | --- |
| Set | `&App`, `String`<br />Returns unit | Set source attribution for diagnostics |
| Build | `&App`<br />Returns `Result<()>` | Validate the prebuilt companion without invoking a compiler |
| Run | `&Arc<App>`<br />Returns `Result<Arc<Invocation>>` | Launch or record a terminal startup failure |
| Stop | `&App`<br />Returns a future of `Result<()>` | Request termination without joining |
| Wait | `&App`<br />Returns a future of `Result<Arc<Invocation>>` | Join the finalized invocation |
| Read | `&App`<br />Returns `Result<Arc<[u8]>>` | Read finalized stdout bytes |
| Inspect | `&App`<br />Returns `Result<Option<Error>>` | Read the selected failure |

These native types belong to the generated module. A launched process, a finalized invocation and a completed continuation are distinct states. Source Await composes joining with continuation ownership. Calling an arbitrary native member does not reproduce that full source operation. [Invocation and result rules](../../cli-build.md#invocation-and-results) define misuse, reruns, cached results and failure selection.

## Capture

Collect a child's output stream and preserve its bytes and capture failure.

Capture.Read accepts an optional async reader, a deadline receiver, host dependencies and a stream identity. It returns a future producing Capture. Capture.Join converts a capture-task result and stream label into `(Arc<[u8]>, Option<Error>)`. Snapshot holds stdout and stderr separately, with an optional Error and exit status.

There is no capture-size cap. The post-exit/cancellation drain bound is two seconds. It is not a normal process-duration timeout. Finalized output includes whitespace, NUL and non-UTF-8 bytes. See [stream and shutdown limits](../../cli-build.md#shutdown-streams-and-limits).

## Error

Represent a shared runtime failure with message bytes and retained causes. Runtime.Result names the generated `std::result::Result<T, Error>` carrier.

| Members | Policy |
| --- | --- |
| Describe, Form | Construct from borrowed text or owned message bytes |
| Import, Explain | Retain an IO error or attach context to a native cause |
| Join | Return None for no causes, preserve a singleton, or construct a joined parent |
| Read, Render, Display | Borrow root message bytes, render diagnostic bytes, or write native formatter output |
| Copy, Identify | Retain the shared error and observe its identity |
| Inspect, Trace | Borrow child errors or the native cause |
| Cancel, Check | Construct cancellation or inspect its cancellation flag |
| Attach | Attach captured stderr |

Display's native text conversion may replace invalid UTF-8. Raw message and captured-output interfaces remain byte-oriented. An Error value is distinct from the App Error member that selects a current operation or invocation failure. Error construction does not itself exit, join or roll back effects.

## Format

Runtime.Format reexports [Formatting.Format](format.md#format), retaining that definition's identity and signature, which accepts `f64` and returns `Vec<u8>`. It adds no lifecycle policy or second number-formatting implementation.

## Invocation

Own one App launch, its terminal result and its continuation claim.

Claim returns `Result<bool>` and admits the continuation once. Finish releases its ownership, Read returns an optional Snapshot, and Complete waits for completion. Allocate, Terminate, Kill, Record, Publish and Supervise support launch, termination and finalization. They require the real generated state, channels, process and host context, not values invented by an ordinary App program.

Snapshot.Copy preserves shared result bytes and error data. Stop can race with natural completion, but finalization publishes one coherent result and each launched direct child has one Wait owner. The [canonical lifecycle contract](../../cli-build.md#app-execution-semantics) owns the detailed ordering rules.

## Output and shutdown

Show accepts a Runtime and borrowed bytes, returning a future of `Result<()>`. It checks terminal admission, writes those exact bytes and flushes without adding LF. Fail, Leave and Request select failure, normal exit or an explicit terminal request. Begin starts shutdown, and Clean performs idempotent cleanup. Finish resolves a source Result to an exit status, while Run drives the source future through cancellation and finalization. Terminal.Copy preserves the selected terminal state.

Cleanup is explicit, and later requests cannot replace the first accepted terminal decision. Fatal output errors, signals, pending children and cleanup failures follow the [shutdown contract](../../cli-build.md#shutdown-streams-and-limits). This is not a promise to clean up after an uncatchable orchestrator kill or host failure, or to own descendants that escaped their process group.

## Startup

Open takes no arguments and returns `Result<Runtime>` using native dependencies and stdout/stderr. Install instead accepts `Arc<dyn host::Dependencies>` and two host Output values. Check tests runtime admission, Copy retains a runtime handle, and Appoint associates a companion path with a new `Arc<App>`.

The native entrypoint owns executor startup and invokes selected L policy. Runtime is not a second L interpreter or a Function-dispatch runtime. Current native execution targets are macOS and Linux. The other-platform adapter retains explicit unsupported behavior.

## Lifecycle example

This complete application uses the runtime through builtin words. Run launches the child. The first Await finalizes its output and activates its continuation, and the second Await reuses the result without repeating that continuation. No Runtime import is required.

<!-- example:runtime-continuation-once -->

For termination, see the [Stop/Await example](../../guide/lifecycle.md). The host witnesses distinguish source compilation, actual process execution, stream capture and cleanup. Generating the page does not launch the example.

## Sources and witnesses

The [Runtime dictionary](../../../runtime/eng/dictionary.l) assembles [errors](../../../runtime/eng/errors.l), [process state](../../../runtime/eng/process.l), [App operations](../../../runtime/eng/app.l), [startup](../../../runtime/eng/start.l), [shutdown](../../../runtime/eng/shutdown.l), [supervision](../../../runtime/eng/supervision.l) and [capture](../../../runtime/eng/capture.l).

[Selection tests](../../../lyger/runtime_library_test.go) and [private host-root tests](../../../lyger/runtime_selection_resources_test.go) check compiler-owned dependencies. [Composition tests](../../../logiceme/runtime_composition_test.go) reject missing/stale resources, and [native layout tests](../../../logiceme/native_runtime_layout_test.go) check selected input paths. [Runtime parity](../../../logiceme/runtime_parity_test.go) and [error-cause tests](../../../logiceme/runtime_error_causes_test.go) observe native behavior. The maintained [integration suite](../../../testdata/integration/README.md) covers lifecycle, capture, faults and shutdown on the executing host. It supplies no automatic cross-platform proof.
