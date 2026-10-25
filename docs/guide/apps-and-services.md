# Organize Apps and Services

Apps own executable processes. Services compose code within Apps; Classes group modules and Objects represent files.

## Begin with a Basic App

`New Basic App` supplies a Main Service and its entry composition. `The App Prints ...` adds child-process work. Run/Await and output handling belong to the enclosing program or its continuation; native lifecycle operations in App bodies are unsupported.

## Run an App

<!-- example:runnable-app -->

Run launches Worker, Await joins its finalized capture, and Show reads Output. Expected stdout is `work completed\n`; [CLI](../../cmd/l/runnable_service_test.go) and [native](../../cmd/l/runnable_service_native_test.go) witnesses retain that assertion.

## Service composition

The [workspace fixture](../../lyger/test/workspace/services.l) constructs additional Services through descriptors. This excerpt requires its preceding IO.Print import:

<!-- example:constructed-service-composition -->

Left and Right each construct a Function inside an Object inside a Service. Grouped constructs an App containing Main, Left and Right Services; Create applies it to First. Main supplies the executable entry; Left and Right become separate libraries. `using any Process` passes the App's body composition into those constructions.

The [workspace test](../../cmd/l/service_workspace_test.go) compiles the two Apps in the complete fixture, four additional libraries and a native library caller. Nonprivate composed items form the libraries' native interface.

A standalone `New Main Service` remains structural: it creates no package or process. Its body effects, native Build/Run/Stop/Await and implicit Output/Error reject; it has no App continuation. Explicit members and source-defined actions retain lexical meaning. This does not provide arbitrary linking of standalone Services or an in-process Service lifecycle.

## Organize code without inventing lifetimes

Classes and Objects have no running lifetime. [Source topology](../architecture.md#native-executable-pipeline) separates App executable crates from constructed Service libraries; published App generations retain version-1 manifests.

See [Service coverage](../../testdata/service-execution/coverage.md), then [lifecycle and output](lifecycle.md).
