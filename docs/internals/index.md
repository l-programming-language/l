# Inside the compiler

L's toolchain has three distinct boundaries: source meaning, source generation, and native program execution. Follow those boundaries when reading diagnostics or evaluating implementation evidence.

## Source to native output

<!-- diagram:compiler-pipeline -->

Lyger prepares source units and imports, tokenizes configured English vocabulary, matches grammatical structures, and resolves bindings. It constructs ordered instructions and application structures, then emits AML. AML means Assembly Markup Language: it is not machine-code assembly and is not ordinary XML.

Logiceme's logeme assembler composes glyphs and templates into Rust source bundles. `Evaluate` returns an `Executable` source bundle; it does not run Cargo or launch a process. The CLI build coordinator plans crates, runs native compilation, and publishes an executable generation. The generated orchestrator performs the program's ordered actions only when a user runs it.

Start with the maintained [architecture map](../architecture.md), then the [AML protocol](../../logiceme/AML.md), [logeme reference](../reference/logemes/index.md), and [native build/runtime contract](../cli-build.md).

## Program organization and processes

<!-- diagram:process-model -->

Apps run in separate OS processes owned by the orchestrator. Their additional constructed Services supply library packages. Standalone Service declarations remain structural and produce no packages, process adapters or continuation ownership. Classes group modules/sources; Objects group files. Source organization does not create independent runtime scopes.

The source topology preserves ownership, descriptor targets, typed parameters and source order through definition expansion. These are compiler invariants, not reasons to infer a general dynamic object system or automatic process creation.

## Evidence at each boundary

The [Lucid suite](../../logiceme/testdata/lucid/README.md) checks frozen backend generation and separately checks native compilation. Its generated cases do not demonstrate runtime behavior. The [frontend/runtime ledger](../../testdata/integration/IMPLEMENTATION.md) records revision-specific native observations. Neither is a substitute for fresh execution at a changed checkout.

This website's checks validate documentation structure, references and rendering. They do not run the compiler or native applications. Errors cross separate [compiler/build categories](../error-model.md), while the generated process runtime has its own failure behavior.

## Where to investigate

Inspect source ownership and imports before tokenization; inspect token roles and grammar consumption before binding; inspect ordered instructions before AML; inspect source bundles before native builds. Preserve raw literal and output bytes. Do not normalize UUIDs in executable inputs or expected output—normalization belongs only in diagnostic views.

The [development policy](../development-harness.md) controls language validation. The website's local Node-only commands remain separate from that gate.
