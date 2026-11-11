# Build your first program

Core Print accepts Text and writes it followed by one newline. It is loaded implicitly.

## The source

The complete [basic program](../../demo/basic/main.l) is:

<!-- example:basic -->

The first sentence binds Text to `message`. `New Basic App` constructs Hello with its Main entry. `The App` refers to that App, and `Prints` adds the output operation to its body.

## Build and run

From the repository root, using a compiler built from this checkout:

```sh
l build demo/basic/main.l
./demo/basic/main
```

The build prints the absolute executable path. Running that executable prints `Hello World!` followed by one newline. Building does not run the program.

The executable is the App itself. There is no runtime launcher or enclosing orchestration program. Do not append Run, Await, Execute, Show or Output inspection: runtime-dependent effects reject during the reset. See the [build contract](../cli-build.md).
