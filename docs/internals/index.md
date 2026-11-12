# Inside the compiler

Lyger resolves L source into AML. Logiceme renders Rust from selected definitions and frozen logeme assets. The CLI compiles the actual standalone App bundle and publishes its executable.

- [Architecture](../architecture.md): responsibilities, ownership and representations.
- [AML protocol](../../logiceme/AML.md): frontend/backend handoff.
- [Logeme reference](../reference/logemes/index.md): retained native composition.
- [Native definitions](../guide/native-definitions.md): source and native boundaries.
- [Build contract](../cli-build.md): standalone compilation and publication.
- [Compiler errors](../error-model.md): failure categories and source attribution.

Runtime and standard-library implementations are absent. Retained compiler structures and native templates do not supply process lifecycle, capture or cleanup. The [temporary hello-world gate](../development-harness.md) defines current validation.
