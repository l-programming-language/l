# CLI and compiler coordination

Read root [instructions](../../AGENTS.md), [architecture](../../docs/architecture.md),
[build semantics](../../docs/cli-build.md) and the
[development guide](../../docs/development-harness.md).

The CLI validates sources, builds native Rust and publishes local build artifacts.
Lyger resolves source meaning into AML; Logiceme renders native source from
canonical compositions. The CLI must not lower Functions, repair returned Rust
or implement a replacement runtime.

## Current build boundary

`compile()` and `process()` remain source-only. The executable coordinator
consumes their result once, validates a sole standalone App, and builds its real
source bundle through pinned Cargo. Core Print
is the supported body. Reject runtime-dependent outer effects, continuations and
multiple Apps; never discard them or synthesize a runner. Runtime will be rebuilt
in L later.

Retain Cargo artifact provenance, offline bounded subprocess supervision and
atomic publication. Match build artifacts to their manifest/target identity,
retain immutable generation metadata and preserve a working executable on
failure. Cache locks protect source/artifact identity through private snapshots.
Keep compiler diagnostics separate from application stdout. Generated deployment
scripts are artifacts, never local validation commands.

## Compiler observations

`TestCompilerInitialize` compares the complete compiler struct.
`TestCompilerCompile` checks raw source, observable AML, unchanged initialized
state, expanded view and originating-user-source projection. `helpers.go` owns
the expected constructors and INITTEST/COMPTEST. Preserve the three-field cliData,
nine-field compiler, positional field order and nil-versus-empty distinctions.
The working result owns the semantic graph; do not substitute fixture constants
or discard it to satisfy an expected view.

Tests run with cmd/l as CWD and use package-local fixtures. `--assets` selects
one complete provider without changing CWD; default resources are embedded.
Do not restore configuration mirrors, bootstrap fallback or deleted libraries.
Historical build-evidence records do not authorize command replay or tool probes.

Run the root temporary hello-world gate with its existing pins, watchdogs and
concurrency. Only frozen logeme assets and their expectations remain protected.
Report generation, compilation and native execution separately. Publication
remains disabled.
