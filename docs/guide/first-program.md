# Build your first program

The basic demo imports a printing action, constructs an App, runs it and displays its output. The compiler supplies builtins; IO is imported explicitly.

## Read the program

<!-- example:basic -->

``Use IO's Print from `io` as "Print"`` imports a dictionary member under the local name Print. The possessive `'s` selects a member: IO's Print is a printing action; Hello's Output is the App's captured stdout. Backticks delimit Text, while double quotes introduce names: `"message"` names a value and `"Hello"` names an App.

`New Basic App` constructs an App with the standard entry structure supplied by Basic, Main and Process. `The App` refers to Hello just declared. `Prints` is the active form of Print, so `The App Prints message` places that action in Hello's body. See [word forms](values.md#read-pipe-delimited-word-forms).

`Execute Hello` performs Start followed by Await; Start performs Build followed by Run. The final Show writes the captured output without launching, waiting or adding a newline.

## Build and run

After [compiler setup](../cli/index.md), run from the repository root:

```sh
l build demo/basic/main.l
./demo/basic/main
```

The first command generates Rust, invokes Cargo and prints the published executable's absolute path. The second runs the orchestrator, which executes source actions and manages the App process. Building alone does not run the program.

Expected stdout is `Hello World!` followed by one LF byte; stderr is empty and exit status is 0. These exact results are asserted by [the distribution CLI tests](../../acceptance/distribution/cli/cli_test.go).

## Keep the output together

Building `main.l` publishes `main` beside it and a companion directory `main.apps`. Distribute both together: the executable needs its adjacent immutable companion generation to launch Apps.

To choose a publication directory, put the option before the filename:

```sh
l build --output-dir ./build/hello demo/basic/main.l
./build/hello/main
```

Relative output paths use the command's working directory. The compiler does not change it to find resources. See [publication and companion layout](../cli-build.md#output-and-publication) for generation details and failed-build behavior.

## Recognize common first-program failures

Print requires its import. Run starts an App without waiting; Await joins its finalized result. Print adds a newline, so `message` and `Hello's Output` are not byte-equal. The [paragraph negative demo](../../demo/paragraphs/main.l) intentionally demonstrates this difference; trimming its output would change the assertion.

Next, learn how [values and output](values.md) are represented.
