# Rust Lucid suite and current verification scope

This maintained Rust end-to-end suite uses frozen templates and independent
expected bundles. [The result ledger](FAILURES.md) records verification for a
specific source identity and host.

## Current inventory

The current runner contains the following obligations:

| Boundary | Required coverage |
| --- | --- |
| Complete AML generation | 59 applications, with all 295 returned artifacts compared exactly. |
| Malformed composition | 192 stored inputs: 62 callable, 35 selection, 50 iteration, five empty-composition, four Deferred, two invocation, 14 keyed-GI and 20 generic/lifetime cases. |
| Native compilation | 59 Expected and 59 Generated contracts; each phase requires 45 successful builds and 14 specific intentional Rust errors. |
| Frozen definitions | All 18 selected Rust segment assets match their recorded hashes. |
| Focused generic/lifetime checks | 50 exact primitive profiles, plus lookup, preparation, placement and resource/isolation checks. |

Shared interpreter, provenance, primitive, preparation and resource/isolation
tests supplement these complete-application obligations. The 50 profiles above
are the generic/lifetime additions, not the total number of low-level suite tests.
Native checks compile the Expected and Generated contracts; three positive
future cases also execute generated assertions. Deployment scripts remain unexecuted.

This guide owns the current inventory and commands. The [dated ledger](FAILURES.md)
owns observed generation, native compilation and acceptance results. The sections
below describe each family's composition, rejection and isolation obligations.

## Frozen templates and retained shared tests

[templates.sha256.json](templates.sha256.json) records the repository-relative
paths and SHA-256 hashes of the 18 Rust segment files. The freeze check compares
the selected resource files against those hashes; there is no automatic update
mode. The [Rust reference](../../../builtins/README.md) defines the supported
forms, excluded Deferred interface and integration limits. Changes to frozen
assets or independent expectations require explicit finite authorization.

Shared parser, interpreter, conjoinment, source-fragment, communication-operand
and native `Void` provenance regressions remain. Shared low-level interpreter
cases retain Rust and generic systemless settings.

## Core compositions

| Case | Composed obligation |
| --- | --- |
| `grouping` | I/GE preserves the nested `2 * (1 + 2)` expression and result binding. |
| `statements` | Three ordered native statement children declare explicit `let mut`, assign an increment, and print. |
| `fields` | Native record declaration plus Y-composed construction preserves both fields and member use. |
| `calculations` | Both Op records retain their order, grouping, prefixes, postfixes, and combined calculation. |
| `communication` | CE composes bounded Tokio send and receive expressions; native E scaffolding creates the runtime and handles `Result`/`Option`. |
| `conditionals` | The composed false/true/fallback chain retains ordered, distinct branch bodies. |

Each `<case>.aml.txt` is a complete AML document defining the application
`lucid_nested_rust` with source location `/src/main.rs`. It uses an ordinary
composed FW entry point with explicit unit output. Native scaffolding never
replaces the composed feature with EA. Communication supplies its native `async`
block explicitly. The current CE path validates operands; Cargo checks the
emitted async context. This case does not establish general compiler inference
or validation of asynchronous execution contexts.

The public call is:

```go
bundle, err := service.Evaluate(input, GenerationOptions{
    IncludeRuntime: false,
})
```

The generation runner checks input immutability, one returned application,
ordered file paths, dependency metadata, duplicate/unexpected paths, and exact
bytes for all five artifacts:

```text
/src/main.rs
/Cargo.toml
/src/config.rs
/start.sh
/stop.sh
```

Expected Rust sources and shared support artifacts are independently authored,
checked-in oracles. Compare them using `bytes.Equal`; do not trim, format, reorder,
normalize, regenerate from actual output, or replace actual files with expected
ones. Common immutable support expectations are shared across case mappings.

## Futures

Five complete pairs use the existing plain manifest and lock. `future-composed`
checks CV construction, CO assignment, COB binding, DV native/composed operands,
and ordered single evaluation. `future-move` returns an owned non-Copy future
with a borrowed parameter lifetime; `future-infer` covers nested construction
and inferred capture. `future-borrow-escape` requires E0373, while
`future-move-reuse` requires E0382. The three positive generated programs also
have a bounded execution check of their assertions.

`lucid_future_test.go` checks malformed compositions, source/resource isolation,
and async-context boundaries. Existing complete pairs and their expected bytes
are unchanged.

## Selection

Six positive complete-AML/source pairs use the same application and
five-file bundle shape:

| Case | Witnesses |
| --- | --- |
| select-communication | DOC-composed ordered receives and send, non-Copy payload, absent/populated bodies. |
| select-bindings | Caller-named/raw DOC results, explicit outer O assignments, caller-supplied `_`, explicit disposal, singular/plural OC inputs and DOC with E. |
| select-default-order | Explicit OC defaults authored first/middle/last; stable ordinary branch order. |
| select-degenerate | Explicit OC default-only and empty selections, empty bodies, subsequent statement and singular/plural OC unit values. |
| select-nesting-hygiene | Nested selection, preserved caller helper-like name, explicit repeated assignments, body once. |
| select-closed-receive | Explicit Option handling and scope across subsequent use. |

Three more complete-AML/source pairs require successful exact generation and
intentional Cargo rejection: select-sync-context (E0728), select-borrow-conflict
(E0499), select-moved-send (E0382). The runner requires the intended code and a
primary source/expansion span, rejects unrelated compiler errors, and never treats
dependency/setup failures as an expected negative.

`selection-invalid/` contains 35 complete AML documents for invalid CI/BI/OC/DOC
compositions. They cover DO/DOB and bare guard rejection, Definition metadata and
cardinality, placement and unsupported codas as well as existing CI/BI rules.
Checks assert the classified underlying semantic cause separately
from application/file context, no panic, unchanged input, and zero returned bundle.
Selection isolation reuses complete fixtures for A/B/A, previous-bundle stability
and equivalent explicit/embedded resources.

The positive projects use ordinary FW and native async scaffolding. CI/BI/CE and
OC/DOC remain composed; ordinary O performs caller-authored outer assignments.
Every BI needs a caller-supplied Definition, including defaults. Default metadata
orders the branch and never supplies a wildcard or ready value. No compiler
temporary or drop statement is inserted. The original communication fixture
continues to use ordinary DOB outside BI, unchanged.
See the [selection contract](../../../builtins/README.md#cibi-selection-contract)
for priority, defaults, result handling, cancellation and binding limits.

## Iteration

Rust uses the Void sentinel for absent named forwarded singular children and requires
all supplied conjunction clauses to be true. Empty plural collections and
explicit keyless records retain their distinct behavior.

Six positive applications cover unconditional X, Calculation-controlled
X, the existing X/I adapter, OG/DOG conditional bindings, all four iterator
binding shapes, caller-supplied iterator adapters, explicit borrowed matching,
async receive conditions and direct PO labels. Nested labels exercise targeted
break/continue. Ordered bodies retain their scope and appear exactly once.
Native statements establish setup and async context; X, conditional Definition
and label composition remain in the logemes under test.

| Case | Witnesses or intended native diagnostic |
| --- | --- |
| iteration-conditions | Unconditional repetition, direct Calculation and retained ordered I/Evaluation conditions. |
| iteration-bindings | OG singular/plural metadata, DOG composition, nested punctuation and caller-explicit borrowed patterns. |
| iteration-async | DOG with an explicit awaited receive and caller-supplied Tokio async context. |
| iteration-traversal | Wildcard, single, discarded-pair and named-pair bindings, with caller-supplied iterable expressions. |
| iteration-labels | PO directly labels unconditional, conditional, definitional and iterator loops. |
| iteration-nesting-scope | Nested labels, targeted break/continue, body-local scope and exactly-once ordered body emission. |
| iteration-nonboolean | E0308: a non-Boolean Calculation used as a while condition. |
| iteration-binding-scope | E0425: a conditional binding used outside its loop body. |
| iteration-sync-await | E0728: an awaited condition outside async. |

The last three applications require exact generation followed by their intended
native diagnostics. Native pattern/type/ownership rules remain Rust's
responsibility; successful compilation
does not prove runtime order, termination or channel behavior.

`iteration-invalid/` contains 50 complete AML documents covering conflicting
conditions, iterator/condition mixtures, three-part composition, body and
child cardinality, ordinary Definitions used as conditions, OG/DOG metadata and
placement, and PO child/label/coda rules. They use the same classified-cause,
zero-bundle, immutability and no-panic requirements as the 35 selection-invalid
cases. Iteration isolation covers A/B/A reuse, earlier bundles and explicit versus
embedded resources without changing selection isolation obligations.

Eight stored `primitive/iterate-*.rs.txt` expectations render the actual Iterate
Default case for Definition, Calculation, Evaluation and unconditional branches,
plus wildcard, single,
discarded-pair and named-pair traversal. These low-level checks supplement complete
AML; they do not bypass composition validation as evidence of application support.

Iteration uses the plain and Tokio rt/sync manifests and retained locks. The
[iteration contract](../../../builtins/README.md#iteration-and-definition-contract)
defines the four categories, OG/DOG fragments, direct PO child forwarding and
rejection of initializer/condition/step composition.

## Scheduling and Deferred rejection

Execute and Execution exclude Deferred inputs, forwarding and branches.
Scheduling expressions require explicit caller-supplied futures and context.

`schedule-direct` exercises V's native Key and `schedule-composed` exercises
MV's composed Instantiation without supplying Deferred. The four documents in
`scheduling-invalid/` cover Deferred True and False on both routes, asserting
the unchanged `Unsupported` cause, zero bundle, input immutability and no panic.
`TestLucidDeferredRejected` checks these failures, and
`TestLucidSchedulingIsolation` checks reuse and selected resource providers.
These tests do not expand the accepted execution child roles or certify other
async forms. Caller-supplied futures and Tokio setup remain explicit.

Expected and Generated projects use the existing manifests and retained locks.
These scheduling checks compile without executing applications or deployment
scripts.

See the [execution contract](../../../builtins/README.md#execution-without-deferred)
for the distinction between scheduling and resource cleanup through RAII. Removing
the last unsupported markers does not establish complete Rust support or verified
runtime behavior.

## Run and retain evidence

The [validation policy](../../../docs/development-harness.md#temporary-validation-pause)
governs the local pause and authorized hosted runs. Run full local validation only
when explicitly requested. The focused command below is diagnostic evidence, not
the ordinary repository gate. From the repository root, select a new absolute
run directory so previous artifacts remain intact:

```sh
export L_LUCID_ARTIFACT_DIR="$PWD/.harness/runs/lucid-native-$(date -u +%Y%m%dT%H%M%SZ)"
GOMAXPROCS=4 GOCACHE="$PWD/.harness/cache/go-build" GOENV=off GOFLAGS= GOWORK=off GOTOOLCHAIN=local \
  go test -mod=readonly -count=1 -run '^TestLucid' -timeout=60s \
  -json ./logiceme
```

For explicitly requested full validation, run the separate required native stage:

```sh
GOMAXPROCS=4 GOCACHE="$PWD/.harness/cache/go-build" GOENV=off GOFLAGS= GOWORK=off GOTOOLCHAIN=local \
  go test -mod=readonly -count=1 -tags=nativecompile \
  -run '^(TestLucidNativeCompile|TestNativeMetadataArrowsCompile)$' -timeout=30m -json ./logiceme
```

The native stage has Expected and Generated sub-stages for all 54 exact-generation
cases, including twelve intentional native negatives per phase.
It compiles independently authored expected projects to detect authoring mistakes,
then materializes and compiles actual returned projects without repair. A byte
mismatch does not block compilation of materializable actual output; generation
failure records an explicit blocked native phase.

The same command also runs the inline native-metadata arrow regression's
independent Expected and Generated phases. It reuses the retained plain manifest
and lock under `lucid_nested_rust`; it does not add a frozen-corpus fixture.

Use installed Rust **1.98.1**, edition **2024**, and each fixture's retained initial
lockfile. Communication uses Tokio **=1.53.1** with `rt` and `sync`; selection adds
`macros` in a separate manifest and initial lockfile. The original locks remain unchanged. Initial locks are resolved from the independently authored manifests;
verification never regenerates them or performs version fallback. Native builds
use the generated manifest:

```sh
cargo +1.98.1 build --frozen --message-format=json --color=never \
  --manifest-path <materialized-project>/Cargo.toml
```

Dependencies must already be available for frozen builds. The subprocess limit
is three minutes and the separate stage limit is thirty minutes. Missing tooling,
cache, dependencies, or timeouts are infrastructure failures. Never execute
Rust binaries or generated start/stop scripts. Successful compilation does not
establish runtime behavior or parity.

For requested full validation, also retain shared regression results and run the
repository gate:

```sh
scripts/l-harness doctor
scripts/l-harness verify
```

Focused checks and the native stage do not replace acceptance. When full validation
is requested, run `scripts/l-check` before claiming it passed; it includes workflow/Go linting,
unfiltered acceptance and this complete native stage. Preserve existing limits
and assertions; report each stage's result separately.

## Declaration adapters

Five complete applications under this directory exercise Declaration adapters.

| Case | Routing and forwarding obligation |
| --- | --- |
| declaration-operation | KA forwards one named FW with two ordered statements and deduplicated nested imports. |
| declaration-description | HA forwards bare U's helper and main functions once, in order; main calls helper natively. |
| declaration-evaluation | FOB receives GA's composed I/GE grouped multiplication; the result is subsequently used. |
| declaration-instantiation | FOB receives MA's Y construction with fields 11 and 22; both fields are read. |
| declaration-iteration | PA forwards X/I/W's conditional loop, ordered body and following statement between native function-boundary fragments. |

The vowel selectors use consonant action names; Declare's nominal selectors and
singular children are unchanged. Rust FW signatures remain parameterless with
explicit unit results, Y has two fields, and PA does not introduce a new W child
role. These choices isolate adapter routing from unrelated integration limits.
No EA, native replacement loop, dispatcher alias, or compiler repair is involved.

Rust expectations are authored from literal template spans and compared without
normalization. The applications are not executed. The [ledger](FAILURES.md)
records generation, compilation and acceptance evidence.

## Named invocation

The [invocation contract](../../../builtins/README.md#named-invocation-contract)
prepares canonical absence for omitted KY Receiver and distinguishes prepared
native `Void` from the bare absence sentinel. A supplied Receiver retains its
original bytes. Explicitly
empty or whitespace-only Receiver is rejected as `InvalidInput` with the exact
leaf cause `Rust KY: Receiver must be nonempty when supplied`. Quoted `"Void"`
comparisons retain their textual meaning.

| Case | Invocation obligation |
| --- | --- |
| invocation-named | A composed parameterless FW helper, receiverless KY with zero or supplied arguments, a qualified callable, an explicit receiver, and receiverless KY under MV with explicit Tokio scaffolding. |
| invocation-void | Composed KY calls with a native Key or Receiver spelled `Void`; native helper declarations isolate the invocation route from Func-field interpolation. |

Both pairs use independently authored exact source expectations and the existing
five-artifact bundle. `invocation-named` reuses the Tokio rt/sync manifest and
retained lock; `invocation-void` reuses the plain manifest and lock.

`invocation-invalid/` contains the two complete malformed Receiver documents.
`TestLucidInvocationInvalid` checks their exact classified leaf causes, zero
bundles, input immutability, no panic and selected Instantiate/Instantiation assets.
`TestLucidInvocationIsolation` evaluates `invocation-named`, `schedule-composed`
and `invocation-named` again using one service under filesystem and embedded
providers, checking exact bundles and previous-result stability. Additional
Rust interpreter tests compare bare/quoted equality and inequality without
mutating metadata or parsed source. Preparation tests distinguish omitted Receiver
from native `Void`, including inherited equal-byte metadata provenance.

Native helper declarations isolate KY from Func-field interpolation. These
invocation witnesses do not establish BW, general signatures or KYC behavior.

## Native callable-name provenance

The [callable-name contract](../../../builtins/README.md#native-callable-name-provenance)
preserves native names spelled `Void` during exact `Meta.Func.Key` interpolation.
The helper uses interpolation's existing Func decoder, keeping the decoded record
authoritative over a differing root `Meta.Key`. Output types and indexed Funcs
have separate provenance contracts below.

| Case | Callable-name obligation |
| --- | --- |
| function-void-flat | Declare `Void` through composed FW with flat Key, empty Inputs and unit Outputs; retain W/BOD/E body and a composed KY call through MOB. |
| function-void-record | Declare the same function through FW with the explicit framed Func record; retain the same composed body and invocation. |

The two independently authored source expectations reuse the existing plain
manifest, support expectations and lockfile. Every returned artifact remains an
exact comparison.

`TestLucidFuncKeyProvenance` covers exact interpolation and bare/quoted equality
and inequality under Rust, differing record/root names, ordinary
names, empty and absent controls, and repeated interpretation without metadata,
working-variable or parsed-source mutation. Empty Key compares unequal to bare `Void`, while the explicit
prepared `Func = Void` sentinel compares equal. A completely unset Func entry
retains its empty-record fallback; it also compares unequal and is not promoted
to native-name provenance.

These callable-name witnesses do not establish general signatures or runtime
behavior. Generated Rust applications and scripts remain unexecuted.

## Typed functions and closures

The [callable contract](../../../builtins/README.md#typed-functions-and-anonymous-callables)
defines typed FW functions, bounded BW placement and KY invocation. Omitted or
empty Outputs omit the return annotation; native Rust checks supplied types.

| Case | Composed obligation |
| --- | --- |
| callable-typed-flat | One/multiple typed FW inputs with flat metadata, explicit output and KY calls. |
| callable-typed-record | Explicit Func, tuple/raw-identifier parameters in order, native name Void and nested import propagation. |
| callable-outputs | Named FW omitted, empty, unit and nonunit Outputs; absent annotations mean unit. |
| closure-inline | KY/BW through flat and explicit Func records; zero/multiple typed inputs and inferred/explicit output. |
| closure-stored-capture | KOB binding, KO assignment into caller-declared uninitialized local, named KY calls, borrowed String capture and owner reuse. |
| callable-unit-mismatch | Exact generation followed by E0308 for returning an integer from a named function without Outputs. |

These independently authored source expectations follow literal template spans
and use the plain manifest/support/lock variant. The explicit BW Func records
omit Key, which decodes as an empty name; the preserved record parser does not
accept the literal empty-field spelling `Key: ,`.

`callable-invalid/` contains 62 complete AML cases. Fifty-one cover required
names, bounded BW placement, KY children/metadata, Func/flat exclusivity,
unsupported Capture/headers, strict nested fields and framing errors; eleven
cover MY signatures as described below. Tests assert exact classified leaf causes,
zero bundles, no panic and input immutability. Interpreter controls retain
existing Prop rendering and repeated-interpretation state.
A/B/A checks use filesystem and embedded providers with import isolation and
Operate/Operation/Instantiate/Instantiation resource provenance.

All native checks compile only; closure invocation, capture behavior and owner
reuse are not executed. No Rust runtime parity is claimed.

## Native output-type provenance

`callable-output-void` supplies the nonunit alias `type Void = u8;` as the explicit
return type of composed FW functions and BW closures, each using flat metadata
and an explicit Func record. All four callables are invoked through KY and their
results have typed uses. The independently authored source retains the frozen
`-> (Void)` punctuation. It uses the plain manifest, shared support expectations
and retained lock.

`TestLucidFuncOutputProvenance` checks joined and parenthesized output
interpolation, bare/quoted comparisons and raw-list presence under Rust. Missing/empty Func and Outputs, unit/ordinary/multiple outputs, and repeated
interpretation without state mutation remain explicit controls. Only exact
`Meta.Func.Outputs~` gains Rust native-source provenance for a nonempty output
collection. Indexed Funcs have the separate MY contract below.

Generated Rust applications and deployment scripts remain unexecuted. The
[ledger](FAILURES.md) records generation, compilation and acceptance evidence.

## MY required signatures

The [MY contract](../../../builtins/README.md#my-required-trait-signatures)
defines required receiverless signatures with optional return annotations.

`trait-signature-outputs` composes MY from eight ordered Func records: missing and
empty Outputs, explicit unit/scalar/multiple/tuple outputs, nested-comma parameter
types and a native method/output type named `Void`. An inline native type alias
and composed A/FW/W main provide the application scaffold. The trait declarations
are not native scaffolding. The independent source expectation follows the
template and uses the plain manifest and retained lock.

Focused indexed Funcs tests exercise numeric and Count/Add access, exact names
and joined output text, raw-list presence, bare/quoted Void comparisons, nested
parameters, bounds and repeated interpretation without mutation.
The direct Rust decoder returns no partial records on malformed input; complete
AML rejects malformed records before rendering with classified causes.

Eleven of the 62 callable validation cases cover supplied/empty Receiver, missing
method names, unknown/duplicate/malformed Func fields and invalid input Props.
Filesystem/embedded A/B/A checks compare complete bundles and verify input,
prior-bundle and resource isolation.

These witnesses cover receiverless required signatures. Empty traits and local
generic/lifetime parameters have separate controls below. Composed methods,
per-method parameters and implementations remain outside this contract.
Generated Rust applications and deployment scripts are not executed.

## Empty Y/HY/MY

The [empty-composition contract](../../../builtins/README.md#empty-record-and-trait-compositions)
defines braced Y/HY/MY output for empty effective collections.

`empty-instantiations` covers omitted and explicit empty Props/Funcs, braced
HY records and Y constructions through MOB, empty MY traits, a populated HY/Y
control, typed uses and trait references. The feature under test is composed;
native fragments supply ordinary uses rather than replacement declarations.
The independent source expectation reuses the existing plain manifest and lock.

Six exact primitive cases cover the empty and nonempty routes. Twenty-one
preparation cases cover canonical missing collections, explicit empty lists,
inherited/local precedence and malformed local/inherited records. Repeated
preparation compares effective metadata by key/value/type; original inputs,
resource data, parsed tags and source bytes remain exact and immutable.

Five stored malformed inputs in `empty-instantiation-invalid` reject empty
native collections with unrendered headers/methods, and empty HY with Shape Sum.
They assert classified leaf causes, zero bundles, no mutation and no panic.
A/B/A service reuse under filesystem and embedded resources verifies exact
bundles and selected Instantiate/Instantiation provenance.

Empty declaration support does not implement composed members, composed
generic-header children or sum types. Local String parameters have separate
controls below. Generated Rust applications and scripts remain unexecuted.

## Keyed GI

The [keyed GI contract](../../../builtins/README.md#keyed-gi-pattern-selection)
requires every supplied pattern/body pair and only explicit terminal fallbacks.

Three complete applications use the exact-bundle/native runner:

| Case | Obligation |
| --- | --- |
| keyed-conditionals | Single and multiple arms, preserved final bindings, exhaustive Boolean arms, literal/tuple/constructor/range patterns, direct/composed terminal fallback, fallback-only, caller wildcard, quoted/qualified/parenthesized Void, nested bodies and separately grouped category ordering. |
| keyed-nonexhaustive | Two supplied integer patterns remain intact; Expected and Generated projects produce E0004 rather than acquiring a fabricated wildcard. |
| keyed-call-pattern | A native call-shaped pattern remains intact and produces E0164; no synthesized equality guard. |

Fourteen stored `keyed-invalid` documents cover cardinality, empty Key/patterns,
unexpected children and direct/composed duplicate or nonterminal fallbacks.
They assert exact classified leaf causes, zero bundles, immutable input and no
panic. Five exact Evaluate primitive cases and thirteen preparation cases cover
one-shell rendering, repeated interpretation, Count restoration, inherited/local
Key precedence and native Void provenance. Filesystem/embedded A/B/A reuses
`keyed-conditionals`, existing `conditionals`, then `keyed-conditionals`, checking
all artifacts, prior bundles and Evaluate/Evaluation resource provenance.

The independent oracles use the plain manifest/lock and exact byte comparisons.
Runtime branch selection is not executed or claimed.

## Lifetime and generic inputs

The [local-input contract](../../../builtins/README.md#lifetime-and-generic-inputs)
defines glyph-local `Lifetimes` and `GenericTypes` String inputs on FW/HY/MY
declarations and Y/named KY applications, with explicit forwarding and no
additional records or child categories.

| Case | Composed obligation |
| --- | --- |
| generics-functions | Flat and record FW; lifetime-only, type-only and combined parameters; inline bounds; nested BW using its surrounding type without receiving a suffix; nongeneric neighbors. |
| generics-records | Borrowed single/multiple-field HY; Y with explicit arguments and inference; typed uses and later owner reuse. |
| generics-traits | Empty and populated MY; trait parameters emitted once and referenced by receiverless signatures; nongeneric neighbor. |
| generics-calls | Named and receiver KY; explicit early-bound lifetime; final-path-component arguments; nested comma/function-pointer native list contents; supplied type Void and nongeneric neighbors. |
| generics-late-bound-lifetime | Explicit call-site lifetime remains supplied and produces E0794. |
| generics-unused-parameters | Empty generic HY/Y retain their requested syntax and produce E0392, without hidden fields. |
| generics-argument-arity | Excess caller-supplied type arguments remain intact and produce E0107. |

All seven independent source oracles follow intended composition and literal
template layout. The Expected stage compiles those projects before Generated
output is considered. Each uses the existing
plain manifest, support expectations and retained lock. Tests compare all five
returned artifacts byte-for-byte; no output normalization or oracle-update mode
is provided.

Twenty complete AML inputs in `generics-invalid/` cover blank values, unsupported
forms, anonymous calls, inputs incorrectly placed inside Func records, duplicates
and inherited Class/Object metadata. An inherited-input EA case exercises the
file-boundary guard. Assertions require the classified leaf cause, zero-value
bundle, unchanged input and no panic.

Fifty exact primitive profiles cover all ten suffix sites with absent inputs,
lifetime-only, type-only, both inputs and native type Void. Additional preparation
and lookup controls cover canonical absence, local provenance, blanks on every
supported form, inherited sentinel/blank/override rejection, repeated immutable
interpretation and generic systemless lookup. Filesystem/embedded service reuse
checks A/B/A output, earlier bundles and provenance for all four selected assets.

Consult the [dated ledger](FAILURES.md) for observed gate outcomes. Generated Rust
applications and deployment scripts remain unexecuted;
compilation establishes neither runtime behavior nor ownership-inference parity.

See the [input contract](../../../builtins/README.md#lifetime-and-generic-inputs)
for native String framing, glyph-local scope and declaration/application placement.
Trait-level inputs do not add method-specific lists; dedicated const-generic,
independent where-clause and generic implementation interfaces remain separate work.

## Recording evidence

[FAILURES.md](FAILURES.md) records the latest dated evidence. Record generation
errors/panics, every actual bundle, expected references, byte-difference evidence,
expected/actual Cargo results, and blocked phases. Classify causes as production generation/validation,
support assembly, fixture/harness, infrastructure, or a reviewed expectation-only
mismatch. Compilation success alone cannot justify expectation-only: review
literal contents, operation order, scopes, and all support artifacts first.

The ledger records whether the current generation and native obligations passed
at a given revision. A documentation update alone establishes no new behavioral
result. For subsequent behavior changes, use an existing exposing regression
when available and repair the responsible production layer within the approved
scope. Keep independent expectations and frozen hashes; changing either requires
a concrete, separately scoped amendment, never an automatic update from actual
output. Consult the [current integration limits](../../../builtins/README.md#current-integration-limits)
before treating the current corpus as support for additional compositions.
