# Working on builtin logeme assets

Read [LOGEMES.md](LOGEMES.md), [CONTRACTS.md](CONTRACTS.md), and the
[backend instructions](../logiceme/AGENTS.md) before changing composition.
Root [ownership and validation rules](../AGENTS.md#current-reset-boundary) apply here;
this file adds the asset-specific boundaries.

## Source and asset ownership

Only `eng/core.l` remains as authored builtin L. It loads directly and includes Print, composed through Process and a private single-call native output leaf. Storage and prelude are deleted; all standard-library authored/native/generated resources are deleted. Do not restore them or implement their policies in Go/Rust. The existing explicit embedding list selects only retained resources.

Core obeys the same [complete-family](../docs/spec/definitions.md#inflection-declarations)
and [authored-name/alias](../docs/spec/names-and-scope.md#possessives-and-member-lookup)
contracts as imported and user source, including private leaves. Native definition
bodies and template slots keep their exact bytes; their internal dotted names are
not authored L names. Keep configured JSON vocabulary separate from source families.

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
inventory and separate native-compilation stage. Preserve independent expectations, shared interpreter tests,
pins, locks, deadlines and exposing regressions. Follow root acceptance instructions
for the temporary hello-world gate; compilation and documentation checks do not establish execution.

Sections, general ownership inference and new mutability metadata remain separate
work. Do not introduce unsafe generated implementation code.

Runtime is absent pending reconstruction in L. No lifecycle or cleanup support is implied by the retained native syntax templates. Generated installer/start/stop scripts remain artifacts, never local validation commands.
