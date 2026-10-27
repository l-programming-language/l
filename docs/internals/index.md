# Inside the compiler

Lyger resolves source meaning, Logiceme generates Rust, and the CLI builds native programs. Running the resulting program starts its App processes.

## Source to native output

<!-- diagram:compiler-pipeline -->

Lyger resolves grammar, bindings, and ordered instructions into AML (Assembly Markup Language). Logiceme's `Evaluate` returns a source bundle; the CLI owns native compilation and publication.

- [Architecture](../architecture.md): responsibilities and source ownership.
- [AML protocol](../../logiceme/AML.md): the frontend/backend handoff.
- [Logeme reference](../reference/logemes/index.md): template composition.
- [Native definitions](../guide/native-definitions.md): writing native-backed L vocabulary.
- [Build/runtime contract](../cli-build.md): native output and execution.

## Program organization and processes

<!-- diagram:process-model -->

Apps run in separate OS processes. Their constructed Services supply libraries; standalone Services create no package or process. Classes group modules and Objects group files, without introducing runtime scopes.

## Evidence at each boundary

The [Lucid suite](../../logiceme/testdata/lucid/README.md) checks exact generation and native compilation, plus bounded execution of three future cases. The [frontend/runtime ledger](../../testdata/integration/IMPLEMENTATION.md) records separate native observations. Results apply to their recorded revisions and hosts; see [evidence and review](../contributing/evidence.md).

## Where to investigate

Use [compiler error categories](../error-model.md) to locate the failing stage, and follow the [development policy](../development-harness.md) for validation.
