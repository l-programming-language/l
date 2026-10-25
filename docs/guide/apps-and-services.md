# Organize Apps and Services

An App is a language value with application structure and process state. Services organize code within App construction; additional constructed Services become library packages. Classes organize modules and Objects represent files.

## Begin with a Basic App

The [first program](first-program.md) creates an App through the Basic descriptor. That descriptor supplies a Main Service and its entry composition. The compiler emits an App binary, while the outer orchestrator owns ordered lifecycle actions.

Put Print actions intended for the child in the App body. Put Run/Await and output handling in the enclosing program or its continuation. Native lifecycle operations in App bodies remain unsupported.

## Run an App

<!-- example:runnable-app -->

Run launches Worker as a separate OS process. Await joins its finalized capture before Show reads Output. The expected stdout is `work completed\n`. Source-generation and native observers are retained in [the CLI tests](../../cmd/l/runnable_service_test.go) and [native suite](../../cmd/l/runnable_service_native_test.go); this documentation edit did not run them.

## Service composition

A standalone Service declaration, including `New Main Service`, remains structural and produces no package or process. Standalone Service body effects, native Build/Run/Stop/Await and implicit Output/Error reject. A Service does not acquire an App continuation. Explicit members and ordinary source-defined actions retain their lexical meaning.

Within App construction, Main/Basic composition and primary Service topology remain unchanged. Additional constructed Services retain their library packages and dependency relationships. This does not provide arbitrary linking of a standalone Service into an App or an in-process Service task lifecycle.

## Organize code without inventing lifetimes

Classes become module/source groups; Objects become files. Neither automatically becomes a process or has a running lifetime. The [source topology](../architecture.md#native-executable-pipeline) keeps executable crates distinct from library-only crates. Published App generations retain version-1 manifests.

Read the [coverage map](../../testdata/service-execution/coverage.md), then continue with [lifecycle and captured output](lifecycle.md).
