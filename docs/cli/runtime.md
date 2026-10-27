# Running a published program

## The executable is the orchestrator

After `l build` succeeds, run the printed executable path. This orchestrator executes your program's actions in order and starts its Apps as separate processes.

Keep the executable and adjacent `.apps` directory together. Neither L nor the Go/Rust toolchains are needed to run it. Apps inherit the invocation's working directory and environment; they receive EOF on stdin and no synthesized arguments.

## Lifecycle and output

Use [Run, Await and Show](../guide/lifecycle.md) to start Apps, wait for their results and display captured stdout. Output preserves its bytes and newlines. Captured stderr is separate and is not automatically forwarded.

## Exit and limits

An App failure is a result your program handles; it does not automatically determine the orchestrator's exit status. See [execution and lifecycle](../spec/execution.md) for termination and cleanup rules, and [CLI/runtime semantics](../cli-build.md) for the complete runtime contract.
