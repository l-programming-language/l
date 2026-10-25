# Build your first program

The maintained basic demo introduces the whole source-to-native path without adding a project bootstrap. Builtins are part of the compiler's selected distribution; `IO.Print` is an explicit library import.

## Read the program

<!-- example:basic -->

The first sentence imports the `Print` member of `IO` from the indexed `io` dictionary and gives it the local name `Print`. Backticks delimit Text. Double quotes introduce names: `"message"` names a value, and `"Hello"` names the App being constructed.

`New Basic App` uses the implicit Basic descriptor. Basic composes the App's entry structure from Main and Process definitions. `The App Prints message` puts the printing action in the App body. `Execute Hello` expands the source-defined App action into Start followed by Await; Start itself performs Build followed by Run.

The final `Show` writes the App's captured stdout. It does not launch the App, wait for it, or add another newline.

## Build and run

After completing the [compiler setup](../cli/index.md), run from the repository root:

```sh
./bin/l build demo/basic/main.l
./demo/basic/main
```

The first command generates Rust and invokes Cargo, then prints the absolute path to the published executable. It does not run the program. The second command runs the orchestrator, which executes the source actions and manages the App process.

The existing test witness expects these stdout bytes from the generated program:

```text
Hello World!
```

The exact expected stream is `Hello World!` followed by one LF byte. Stderr is empty and the expected exit status is 0. This is the retained assertion in [the distribution CLI tests](../../acceptance/distribution/cli/cli_test.go), not a claim that native verification ran during this documentation pass.

## Keep the output together

Building `main.l` publishes `main` beside it and a companion directory named `main.apps`. Move or distribute both together. The executable uses its adjacent immutable generation, so copying only `main` can prevent its App from launching.

To choose a separate publication directory, put the option before the filename:

```sh
./bin/l build --output-dir ./build/hello demo/basic/main.l
./build/hello/main
```

The output-directory option is resolved from the command's working directory. The compiler does not change that directory to locate its own resources. See [publication and companion layout](../cli-build.md#output-and-publication) for generation details and failed-build behavior.

## Recognize common first-program failures

Without the explicit import, `Print` is not supplied by implicit builtins. `Run Hello` starts a process but does not wait; use `Await Hello` before an action that requires its finalized output. Comparing `message` with `Hello's Output` also needs care: the App prints a newline, so the two Text values are not byte-equal.

The [paragraph negative demo](../../demo/paragraphs/main.l) intentionally exposes that newline difference. Do not trim output or edit its expectation to make it pass.

Next, learn how [values and output](values.md) are represented.
