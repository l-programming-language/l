# Build your first program

After [installing L and its build prerequisites](../cli/installation.md), save the following program as `main.l` in a new directory. You do not need the compiler's source checkout.

## Read the program

<!-- example:basic -->

``Use IO's Print from `io` as "Print"`` imports a printing action. Backticks delimit Text; double quotes introduce names. Here, `message` names Text and `Hello` names an App.

`New Basic App` creates the App. `The App` refers to Hello, and `Prints` is a form of Print that puts the action in its body. See [word forms](values.md#read-pipe-delimited-word-forms).

`Execute Hello` starts the App and waits for it. The possessive `'s` selects its captured `Output`; Show writes those bytes without adding a newline.

## Build and run

Run from the directory containing `main.l`:

```sh
l build main.l
./main
```

The first command builds the program and prints the executable path. The second runs it. Building alone does not execute your program.

Expected output:

```text
Hello World!
```

Print supplies one trailing newline. The program exits successfully with no stderr.

## Keep the output together

Building `main.l` creates `main` and a companion directory `main.apps` beside it. Keep both together when moving the program. See the [CLI reference](../cli/index.md) for build options and [published files](../cli/build.md#published-files) for the layout.

## Recognize common first-program failures

Keep the Print import. Use Execute to run and wait before reading Output. Remember that Print adds a newline, so `message` and `Hello's Output` differ by that final byte. See [troubleshooting](../cli/troubleshooting.md) for build failures.

Next, learn how [values and output](values.md) are represented.
