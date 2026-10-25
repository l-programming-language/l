# Working on builtin logeme assets

Read [LOGEMES.md](LOGEMES.md) for the alphabet and
[CONTRACTS.md](CONTRACTS.md) for required behavior before changing segments.
Read [backend guidance](../logiceme/AGENTS.md) when changing integration.
The repository root instructions remain applicable.

## Scope and ownership

- `segments/` and `templates/` contain native syntax and support assets.
  `eng/core.l` contains L-level source definitions; the matching dictionaries
  live under `../stdlib/eng`. Keep source libraries distinct from native
  backend assets. Rust is the sole generated language.
- Preserve the Rust L-source vocabulary,
  typed placeholders, declaration order and logeme composition. The bounded
  [Rust source differences](README.md#l-source-library-counterparts) do not
  authorize additional vocabulary or library behavior. The current Rust frontend
  handles native literal/type encoding and selects these resources explicitly.
- L, Lyger, and Logiceme remain Go implementations. Rust generation preserves
  L's value/type semantics. The typed frontend Rust plan owns
  storage, effect and activation analysis; templates do not infer ownership.
- Letters own abstract logic. F/O defines; B/A declares; D/E calculates;
  G/I selects alternatives; C/V handles asynchrony. Target syntax cannot reassign
  those owners. Preserve focused template granularity.
- Keep primitives short. Compose involved behavior through syllables. Metadata
  preparation, ownership inference, dependency resolution, and process lifecycle
  belong outside segment templates.
- Preserve Meta/Funcs/Props/Ops/Conjoinments terminology and Func/Prop/Op records.
  The current Rust template interfaces are frozen, including the approved
  [Lifetimes/GenericTypes inputs](README.md#lifetime-and-generic-inputs).
  These local String inputs qualify existing forms; they are not record fields.
  Additional inputs or mutability/capture interfaces need a separately scoped
  change with explicit type/default/absence rules.
- Rust Execute/Execution exclude Deferred. No replacement
  syllable or implicit cleanup is introduced by that removal; see the
  [Rust execution contract](README.md#execution-without-deferred).
- Use composed examples and tests. EA is an emergency whole-file escape, not the
  default authoring path or a substitute for exercising a supported syllable.

## Before changing a form

State its letter meanings, input and child cardinalities, output role, and
interaction with existing pairings. Distinguish supported syntax from design
proposals. An unused selector is not evidence of semantic fit. Never silently
rename or repurpose a form or discard an unsupported child.

Keep target resources canonical; use the selected `fs.FS` and the
explicit embedding list in [distribution.go](../distribution.go). Do not restore
resource mirrors or add host-path fallback. Generated installer/start/stop files
are artifacts, not commands to run while working on templates.

## Verification

Follow the [validation policy](../docs/development-harness.md#temporary-validation-pause)
for the local pause and authorized hosted runs. Run local bootstrap, full lint,
full ordinary/native suites or `scripts/l-check` only on explicit user request.
For requested full validation, follow [acceptance](../docs/development-harness.md),
resolve every lint finding, and require unfiltered acceptance and complete tagged
native compilation.

Use complete AML through Evaluate and compare every artifact's exact bytes,
ordered paths and dependencies. Preserve scenario identities, independent
expectations and shared interpreter tests. Record an exposing failure before an
authorized repair; review any finite expectation amendment separately. Compile
actual outputs unchanged in the `nativecompile` stage, preserving pins, locks
and deadlines. Compilation does not establish runtime parity.

The [Rust suite](../logiceme/testdata/lucid/README.md) owns the maintained inventory;
the [ledger](../logiceme/testdata/lucid/FAILURES.md) records dated results.
Documentation-only work requires local link checks and diff review. State which
validation was not run; documentation checks establish no new generation,
compilation, runtime or acceptance result.

## Native asset boundaries

The 18 Rust segments are frozen assets with explicitly reviewed amendments and
bounded compiler integration. The suite retains exact bundle comparisons,
Expected/Generated native compilation and a template hash check. Consult the
[current integration limits](README.md#current-integration-limits); old preparation
code and the retired broad matrix do not certify other forms.

- Preserve the frozen conditions, selectors, inputs, primitive forwarding and
  template bytes unless the user explicitly scopes a later asset change. Do not
  revive abandoned forms or add metadata/conjoinment types to bypass a compiler
  integration defect.
- Keep Meta/Funcs/Props/Ops/Conjoinments and the existing Func/Prop/Op records.
  More involved behavior belongs in composition or compiler preparation, not
  larger primitives. Only Default cases must stand alone; qualified fragments
  may require their enclosing composition to complete the syntax.
- Preserve the approved [Lifetimes/GenericTypes contract](README.md#lifetime-and-generic-inputs):
  local String list interiors on FW/HY/MY declarations and Y/named KY applications.
  Missing inputs emit nothing; supplied blanks reject; native `Void` stays text.
  These are not record fields or inherited Class/Object/file defaults. Further
  interfaces, including mutability/capture policy, require separate review.
- Declare currently renders typed `let` through `Meta.Keys` and `Meta.Type`, with
  no initializer or automatic `mut`. Definitions own their supplied values.
  Missing lifetime/capture data is not permission to clone, move, or invent one.
- A currently emits a package comment and the existing Headers/use loop. The
  compiler collects file-owned imports. The source assembler adds topology-owned
  safety/module items at the known A boundary and composes configuration startup
  before user effects. This does not expand the frozen declaration template.
- Preserve the bounded [FW/BW and KY callable contracts](README.md#typed-functions-and-anonymous-callables),
  native name/output provenance, [receiverless MY signatures](README.md#my-required-trait-signatures),
  and [empty Y/HY/MY compositions](README.md#empty-record-and-trait-compositions).
  Function receivers, generic closures, composed methods and implementations
  remain outside these contracts.
- Follow the [selection](README.md#cibi-selection-contract),
  [iteration](README.md#iteration-and-definition-contract) and
  [keyed GI](README.md#keyed-gi-pattern-selection) contracts. OC/DOC own selection
  assignments, OG/DOG own conditional binding, and PO labels its Iteration
  directly. BI Default orders branches only; legacy three-part X rejects.
  Native patterns and caller-supplied async scaffolding do not establish general
  structural-pattern, enum or async-function interfaces.
- Tokio supplies Rust runtime process supervision. Execute/Execution have
  no Deferred input, forwarding or branch. Supplied Deferred remains a classified
  rejection, including False. Follow the [execution contract](README.md#execution-without-deferred);
  do not translate it into RAII guards, scope-exit callbacks or implicit cleanup.
- EA remains a standalone emergency literal file, not the normal way to author
  an app or make a composition test pass. Do not add an A sibling to an EA file.

## Preparation and support

Keep target-specific metadata preparation outside segments and preserve L's
value, alias and paragraph semantics and the lifecycle of Apps. The frontend typed plan
tracks activation-local storage, definite assignment, effects and required mutation.
It borrows inspection values and preserves owned snapshots where L requires them;
it does not parse opaque native leaves or infer arbitrary user lifetimes. Sections,
general ownership inference and new mutability metadata remain separate work.
Do not introduce unsafe generated implementation code as part of Rust development.

Dedicated support templates own Cargo files, startup configuration, process
supervision, stream capture and cleanup. The supervised runtime and frontend
parity tests exercise actual native binaries; read their [implementation evidence](../testdata/integration/IMPLEMENTATION.md)
separately from the compile-only Lucid suite. Sole-library packages omit executable
support and export nonprivate composed items. Dependency/build policy belongs to
the coordinator. Run native applications only through the applicable bounded
regression tests; deployment start/stop scripts remain artifacts.
