# Organize Apps and Services

Apps own executable processes. Services compose code within Apps. Classes group modules, and Objects represent files.

## Begin with a Basic App

`New Basic App` supplies a Main Service and its entry composition. `The App Prints ...` adds child-process work. Run/Await and output handling belong to the enclosing program or its continuation. Native lifecycle operations in App bodies are unsupported.

## Run an App

<!-- example:runnable-app -->

Run launches Worker, Await joins its finalized capture, and Show reads Output. The program writes `work completed\n`.

## Service composition

The [workspace fixture](../../lyger/test/workspace/services.l) constructs additional Services through descriptors. This excerpt requires its preceding IO.Print import:

<!-- example:constructed-service-composition -->

Left and Right each construct a Function inside an Object inside a Service. Grouped constructs an App containing Main, Left and Right Services, and Create applies it to First. Main supplies the executable entry, while Left and Right become separate libraries. `using any Process` passes the App's body composition into those constructions.

Nonprivate composed items form the libraries' native interface. The [Service coverage ledger](../../testdata/service-execution/coverage.md) records the workspace and native-caller checks.

A standalone `New Main Service` creates no package or process. Its body effects, native lifecycle operations and implicit Output/Error reject. Explicit members and source-defined actions retain lexical meaning. See the [Service contract](../cli-build.md#service-composition) for that structural boundary.

## Organize code without inventing lifetimes

Classes and Objects have no running lifetime. [Source topology](../architecture.md#native-executable-pipeline) explains their generated files and modules.

See [Service coverage](../../testdata/service-execution/coverage.md), then [lifecycle and output](lifecycle.md).
