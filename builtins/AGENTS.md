# Working on builtin logeme assets

Read [LOGEMES.md](LOGEMES.md), [CONTRACTS.md](CONTRACTS.md), and the
[backend instructions](../logiceme/AGENTS.md) before changing composition.
Root [ownership and validation rules](../AGENTS.md#current-contract) apply here;
this file adds the asset-specific boundaries.

## Source and asset ownership

Follow the root layout and source-ownership rules: one registry/prelude, one
Storage dictionary, a replaceable Runtime standard library with separate
authored/native/generated resources, and explicit
Formatting without supervision imports. Use the selected `fs.FS` and explicit
[embedding list](../distribution.go); definitions and source compositions own
behavior, with no Statement/EA algorithm escape or separate frontend implementation.

Letters retain their abstract owners: F/O defines, B/A declares, D/E calculates,
G/I selects and C/V handles asynchrony. Before composing a form, state its letter
meanings, input/child cardinalities, output role and existing pairings. An unused
selector does not establish semantic fit; do not repurpose forms or discard children.

## Frozen native contracts

Do not edit the 18 Rust `segments/`, non-runtime `templates/`, their cases,
conditions, forwarding, interfaces or frozen hash expectations. Preserve
Meta/Funcs/Props/Ops/Conjoinments terminology and Func/Prop/Op records.
Only Default cases must stand alone; qualified fragments can require their parent.
The [current limits](README.md#current-integration-limits) qualify all coverage claims.

Read the applicable canonical contract before changing preparation or composition:

- [Lifetimes/GenericTypes](README.md#lifetime-and-generic-inputs), including local
  String framing, missing/blank/Void behavior and the separately bounded CV Capture input.
- [Callables](README.md#typed-functions-and-anonymous-callables),
  [receiverless MY](README.md#my-required-trait-signatures), and
  [empty records/traits](README.md#empty-record-and-trait-compositions).
- [Selection](README.md#cibi-selection-contract),
  [iteration](README.md#iteration-and-definition-contract), and
  [keyed GI](README.md#keyed-gi-pattern-selection), including branch/binding ownership.
- [Futures](README.md#future-construction-awaiting-and-value-bindings) and
  [execution](README.md#execution-without-deferred), including Deferred rejection.

Preserve Declare's no-initialization contract and assembler-owned topology from
backend guidance. EA remains a standalone literal file with no A sibling; it cannot
replace the composition being tested.

## Verification and support

Use complete AML through Evaluate and compare exact artifact bytes, ordered paths
and dependencies. The [Rust suite](../logiceme/testdata/lucid/README.md) owns the
inventory and separate native-compilation stage; its [ledger](../logiceme/testdata/lucid/FAILURES.md)
is historical evidence. Preserve independent expectations, shared interpreter tests,
pins, locks, deadlines and exposing regressions. Follow root acceptance instructions
for the full gate; compilation and documentation checks do not establish execution.

Sections, general ownership inference and new mutability metadata remain separate
work. Do not introduce unsafe generated implementation code.

L owns runtime supervision, capture and cleanup policy; native inputs retain host,
executor and OS mechanisms. Use the [supervised native regressions](../testdata/integration/IMPLEMENTATION.md)
for runtime evidence. Sole-library bundles omit executable support and export
nonprivate composed items. The coordinator owns dependency/build policy. Generated
installer/start/stop scripts remain artifacts, never local validation commands.
