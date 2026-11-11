# Execution

The executable is one standalone App. Core Print accepts Text and writes it followed by one newline. Runtime-dependent orchestration is unavailable.

## Compilation is not execution

Lyger resolves source, Logiceme generates Rust and Cargo builds the App. `l build` publishes the resulting executable without running it. Invoke that executable separately.

## Targets and process ownership

The CLI accepts one App and no additional Service libraries. It builds the App's own source bundle, with no generated launcher or process-supervision runtime. See the [build contract](../cli-build.md).

## Build, Run, Stop, and Await

Source-level lifecycle effects reject during the reset. Start and Execute remain declared in core, but their Build/Run/Await composition is unavailable. Recognition or declaration does not establish executable support.

## Invocations and ordering

The supported program runs its Print body directly. There is no enclosing invocation scheduler. The compiler must preserve source ordering and reject unsupported effects instead of dropping them.

## Continuations and activation

Runtime continuations and capture are unavailable. A lexical paragraph remains a source-scope concept, not a runtime activation or concurrent task.

## Output and Error

Print writes to the process's standard output. There is no source-level captured Output or runtime Error inspection. Observe stdout and process status through the invoking shell or test harness.

## Shutdown and cleanup

There is no L process-supervision or lifecycle-cleanup contract in the current distribution. Retained backend scheduling templates do not supply one.

## Platform and concurrency limits

Builds target the native macOS or Linux host. Cross-compilation, multiple Apps, runtime lifecycle and source concurrency are outside the current executable boundary.

## Sources and witnesses

[Core](../../builtins/eng/core.l), [standalone compiler coordination](../../cmd/l/compiler.go) and [hello-world tests](../../cmd/l/core_hello_world_test.go) define the selected path. The [temporary hello-world gate](../development-harness.md) supplies current validation.
