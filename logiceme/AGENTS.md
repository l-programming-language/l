# Working on the logeme assembler

Read [ARCHITECTURE.md](ARCHITECTURE.md), [AML.md](AML.md), and the shared
[semantics](../builtins/LOGEMES.md) and [contracts](../builtins/CONTRACTS.md).
When editing assets, also read the [builtin instructions](../builtins/AGENTS.md).
Root instructions, [runtime semantics](../docs/cli-build.md), and
[acceptance protections](../docs/development-harness.md) remain applicable.

## Boundaries to preserve

- `LogicemeService.Evaluate` accepts complete framed AML and returns an in-memory
  source bundle. An Executable is not a built/running application. Keep scanning,
  syllable composition, Lucid interpretation, linking and support assembly distinct.
- Rust is the only generated language. Evaluate accepts at most one options value,
  containing only IncludeRuntime; no output-target selector remains.
  Natural-language selection belongs to Lyger. Process all books without
  contaminating reused services or prior bundles.
- `NewLogicemeService(fs.FS)` selects a complete provider; nil uses embedded
  resources. Fixed glyph definitions allocate fresh metadata per evaluation.
  Explicit host AML inputs remain host-file operations.
- Preserve original input, source order, scoped metadata, prior bundles, exact
  literal bytes, and error causes. Return classified errors and no partial bundle;
  do not panic, print compiler diagnostics to stdout, or substitute demo artifacts.
- Keep metadata/signature/ownership preparation out of tiny segment templates.
  Preserve Func/Prop/Op and terminology. Rust's template interfaces,
  cases and forwarding are frozen; changing them or adding metadata requires an
  explicitly reviewed task. Leftover preparation fields are not authoring contracts.
- Rust's [Lifetimes/GenericTypes inputs](../builtins/README.md#lifetime-and-generic-inputs)
  are local native String interiors on FW/HY/MY/Y/named KY. Preserve their bytes,
  declaration/application roles and explicit forwarding. Inherited file metadata
  rejects either name, even with a local override or an EA file. Do not turn them
  into record fields, infer arguments or parse a new generic parameter structure.
- Validate logical child roles and execution contexts before rendering. Preserve
  Rust checks and the generic systemless interpreter's independent contracts.
- Rust rejects supplied Deferred metadata before rendering, including False.
  Preserve its Unsupported diagnostic; do not widen execution roles or
  synthesize scope-exit behavior; see the
  [execution contract](../builtins/README.md#execution-without-deferred).
- Shared Lucid fixes must preserve literal spans, child occurrences, loop
  iteration/binding scope, and template-owned separators for Rust and generic systemless interpretation.
- Rust A renders a package comment and collected file-owned imports. The source
  assembler adds graph-owned safety/module items separately at the compiled A
  boundary and composes configuration initialization before user effects. Sole
  library targets return sources/Cargo only, reject executable settings/runtime
  and export nonprivate composed items. Keep all of this out of frozen A; see
  [topology and libraries](../builtins/README.md#frontend-topology-and-library-packages).
  Preserve EA's standalone literal-file boundary without using it to replace
  the composition under test.
- The CLI workspace-entry bridge places canonical runtime/process assets in a
  dedicated library through typed header/support assembly. Keep its allocated
  package/path separate from AML metadata. Ordinary Evaluate/IncludeRuntime
  callers retain their inline-runtime bundles; do not rewrite completed source
  or expose the private runtime placement as a glyph input.

## Workflow and verification

Use the repository logeme-handoff skill for boundary diagnosis and the acceptance
skill for baseline/gate work. Fix the earliest responsible layer; do not compensate
for an interpreter defect in larger templates or normalize completed output.

Establish a failing regression before a behavioral repair; reuse an existing
exposing case when available. Exercise complete AML through
Evaluate, every source/support artifact, target/resource isolation, and relevant
native compilation. The maintained Rust corpus preserves its composed baseline
and scoped extensions. Follow the
[Rust stage instructions](testdata/lucid/README.md) for exact bundle checks
and separate compilation. The 18 Rust segments are frozen by that suite; do not
change them outside an explicitly scoped amendment or substitute a different matrix
to make the current tests pass.
The [result ledger](testdata/lucid/FAILURES.md) records exact generation,
expected/actual native compilation and acceptance separately; it does not establish
correctness of every compiler branch. Runtime behavior is checked
separately by supervised native regressions and the [implementation record](../testdata/integration/IMPLEMENTATION.md).
Retain shared backend controls and the required downstream test identities;
their native observers check the Rust frontend/CLI obligations.

Do not change protected input bytes, expected-value helpers, locks or canonical
resources outside an explicit finite amendment. Generated install/start/stop
scripts are artifacts, never local validation commands. Keep process supervision
and raw stdout/continuation semantics in dedicated support, not glyphs.

Follow the [validation policy](../docs/development-harness.md#temporary-validation-pause)
for the local pause and authorized hosted runs. Run local bootstrap, full lint,
full ordinary/native suites or `scripts/l-check` only on explicit user request.
For explicitly requested full validation, run `scripts/l-check`, resolve every
lint finding, and require workflow lint, ordinary unfiltered acceptance and the
complete tagged native-compilation stage. Focused checks do not establish full
validation. Check documentation links, consistency and diff scope, and state
which validation was not run during the pause.
Report proposed, implemented, generated, compiled and runtime-verified separately.
