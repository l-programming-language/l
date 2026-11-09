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
| Native compilation | 59 Expected and 59 Generated contracts. Each phase requires 45 successful builds and 14 specific intentional Rust errors. |
| Frozen definitions | All 18 selected Rust segment assets match their recorded hashes. |
| Focused generic/lifetime checks | 50 exact primitive profiles, plus lookup, preparation, placement and resource/isolation checks. |

Shared interpreter, provenance, primitive, preparation and resource/isolation
tests supplement these complete-application obligations. The 50 profiles above
are the generic/lifetime additions, not the total number of low-level suite tests.
Native checks compile the Expected and Generated contracts. A separate tagged
check executes assertions in the three positive [future cases](#futures).
All other listed applications are compile-only. Deployment scripts remain
unexecuted. Compilation establishes neither runtime behavior nor ownership-inference parity.

This guide owns the current inventory and commands. The [dated ledger](FAILURES.md)
owns observed generation, native compilation and acceptance results. The sections
below describe each family's composition, rejection and isolation obligations.

## Frozen templates and retained shared tests

[templates.sha256.json](templates.sha256.json) records the repository-relative
paths and SHA-256 hashes of the 18 Rust segment files. The freeze check compares
the selected resource files against those hashes. There is no automatic update
mode. The [Rust reference](../../../builtins/README.md) defines the supported
forms, excluded Deferred interface and integration limits. Keep frozen assets,
interfaces and hash expectations unchanged. Independent fixture/oracle amendments
require explicit finite authorization.

Shared parser, interpreter, conjoinment, source-fragment, communication-operand
and native `Void` provenance regressions remain. Shared low-level interpreter
cases retain Rust and generic systemless settings.

## Core compositions

| Case | Composed obligation |
| --- | --- |
| `grouping` | I/GE preserves `2 * (1 + 2)` through GOB binding. |
| `statements` | Three BOD/E statements remain ordered within W/FW: explicit native `let mut`, increment assignment, then printing. This adds no mutability metadata. |
| `fields` | DA/E native declaration plus Y/MOB two-field construction preserves fields and member use. It does not test HY's struct shell. |
| `calculations` | DE/DOB preserves both Ops, grouping, prefixes and postfixes; operand syntax remains native data. |
| `communication` | CE/DOB preserves send/receive order and bindings. Native E scaffolding creates Tokio, the async block and channel, and handles `Result`/`Option`. It tests neither V/MV scheduling nor CI/BI selection. |
| `conditionals` | GI/BOG preserves false, true and fallback bodies in an unkeyed chain, without testing general pattern matching. |

Each `<case>.aml.txt` is a complete AML document defining the application
`lucid_nested_rust` with source location `/src/main.rs`. It uses an ordinary
composed FW entry point with empty parameters and explicit unit output. The empty
Header loop, restored Func/Props/Ops records, singular/plural forwarding and
ordered child categories are part of their integration. Native scaffolding never
replaces the composed feature with EA. Communication supplies its native `async`
block explicitly. The current CE path validates operands. Cargo checks the
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
checked-in oracles. Compare them using `bytes.Equal`. Do not trim, format, reorder,
normalize, regenerate from actual output, or replace actual files with expected
ones. Common immutable support expectations are shared across case mappings.

The named family checks below share these guarantees where applicable:

| Check | Common assertions |
| --- | --- |
| Malformed composition | Exact classified semantic cause, separate from application/file context; unchanged input, no panic and a zero-value bundle. |
| Resource/isolation | Explicit and embedded providers, A/B/A service reuse, stable earlier bundles and file-owned imports, with provenance for the selected assets. |
| Interpreter/preparation | Preserve metadata, bindings, parsed input and literal bytes across repeated use; family-specific controls below supplement these guarantees. |
| Native negative | Expected and Generated projects require the intended diagnostic code and primary source/expansion span. Unrelated errors or dependency/setup failures cannot satisfy it. |

## Futures

Five complete pairs use the existing plain manifest and lock. `future-composed`
checks CV construction, CO assignment, COB binding, DV native/composed operands,
and ordered single evaluation. `future-move` returns an owned non-Copy future
with a borrowed parameter lifetime. `future-infer` covers nested construction
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

Three more pairs require exact generation and intentional Cargo rejection:
select-sync-context (E0728), select-borrow-conflict (E0499), select-moved-send (E0382).

`selection-invalid/` contains 35 complete AML documents for invalid CI/BI/OC/DOC
compositions. They cover DO/DOB and bare guard rejection, Definition metadata and
cardinality, placement and unsupported codas as well as existing CI/BI rules.

The positive projects use ordinary FW and native async scaffolding. CI/BI/CE and
OC/DOC remain composed. Ordinary O performs caller-authored outer assignments.
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
Native statements establish setup and async context. X, conditional Definition
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
responsibility. Successful compilation
does not prove runtime order, termination or channel behavior.

`iteration-invalid/` contains 50 complete AML documents covering conflicting
conditions, iterator/condition mixtures, three-part composition, body and
child cardinality, ordinary Definitions used as conditions, OG/DOG metadata and
placement, and PO child/label/coda rules. Iteration and selection retain independent
isolation checks under the common requirements above.

Eight stored `primitive/iterate-*.rs.txt` expectations render the actual Iterate
Default case for Definition, Calculation, Evaluation and unconditional branches,
plus wildcard, single,
discarded-pair and named-pair traversal. These low-level checks supplement complete
AML. They do not bypass composition validation as evidence of application support.

Iteration uses the plain and Tokio rt/sync manifests and retained locks. The
[iteration contract](../../../builtins/README.md#iteration-and-definition-contract)
defines the four categories, OG/DOG fragments, direct PO child forwarding and
rejection of initializer/condition/step composition.

## Scheduling and Deferred rejection

Execute and Execution exclude Deferred inputs, forwarding and branches.
Scheduling expressions require explicit caller-supplied futures and context.

`schedule-direct` exercises V's native Key and `schedule-composed` exercises
MV with KY's receiver-method call `ready.into_future()`, without Deferred.
The four documents in `scheduling-invalid/` cover Deferred True and False on both
routes, retaining the `Unsupported` cause.
`TestLucidDeferredRejected` checks these failures, and
`TestLucidSchedulingIsolation` checks reuse and selected resource providers.
These tests do not expand the accepted execution child roles or certify other
async forms. Caller-supplied futures and Tokio setup remain explicit.

Expected and Generated projects use the existing manifests and retained locks.

See the [execution contract](../../../builtins/README.md#execution-without-deferred)
for the distinction between scheduling and resource cleanup through RAII. Removing
the last unsupported markers does not establish complete Rust support or verified
runtime behavior.

## Run and retain evidence

Follow the [validation policy](../../../docs/development-harness.md#validation-authorization).
Task-authorized checks proceed without repeated permission requests. Focused
checks establish only their stated boundary. The [development guide](../../../docs/development-harness.md)
owns tool provisioning, environment and validation budgets.

For an authorized native-compilation check, run:

```sh
scripts/l-ci native
```

This records fresh evidence for Expected and Generated compilation of the
[current inventory](#current-inventory), including intentional native negatives,
and the native-metadata arrow regression. It compiles independently authored
expected projects and actual returned projects without repair. A byte mismatch
does not block compilation of materializable output. Generation failure records
a blocked phase. The arrow regression reuses the retained plain manifest and
lock rather than adding a corpus fixture.

The configured native stage compiles only. The three [future cases](#futures)
have a separate tagged execution check. When that check is requested, run it on
macOS or Linux with the pinned toolchain and retained dependencies:

```sh
L_LUCID_ARTIFACT_DIR="$PWD/.harness/runs/lucid-futures-$(date -u +%Y%m%dT%H%M%SZ)" \
  scripts/l-go test -mod=readonly -count=1 -tags=nativecompile \
  -run '^TestLucidFutureExecution$' -timeout=30m -json ./logiceme
```

Retain each fixture's initial lockfile and frozen dependency resolution. Missing
tooling, dependencies or expired deadlines are failures. Generated start/stop
scripts remain artifacts. Compilation alone does not establish execution behavior.

For requested full repository validation, run `scripts/l-check`. It includes
linting, unfiltered acceptance and the configured native-compilation stage.
Report any separately requested future-execution results explicitly. Preserve
assertions and budgets. Do not substitute focused checks for the gate.

Record generation errors/panics, actual bundles and expected references, byte
differences, Expected/Generated Cargo results, and blocked phases in the
[dated ledger](FAILURES.md). Classify causes as production generation/validation,
support assembly, fixture/harness, infrastructure, or a reviewed expectation-only
mismatch. Compilation success alone cannot justify the last classification:
review literal contents, operation order, scopes and every support artifact.
Use an existing exposing regression when available and repair the responsible
production layer. Independent expectation amendments require concrete approval,
never an automatic update from actual output. Keep frozen hashes unchanged.
Documentation updates establish no new behavioral result; consult the
[integration limits](../../../builtins/README.md#current-integration-limits)
before extending a support claim.

## Declaration adapters

Five complete applications under this directory exercise Declaration adapters.

| Case | Routing and forwarding obligation |
| --- | --- |
| declaration-operation | KA forwards one named FW with two ordered statements and deduplicated nested imports. |
| declaration-description | HA forwards bare U's helper and main functions once, in order; main calls helper natively. |
| declaration-evaluation | FOB receives GA's composed I/GE grouped multiplication. The result is subsequently used. |
| declaration-instantiation | FOB receives MA's Y construction with fields 11 and 22; both fields are read. |
| declaration-iteration | PA forwards X/I/W's conditional loop, ordered body and following statement between native function-boundary fragments. |

The outer vowel selectors use `Operate`, `Describe`, `Evaluate`, `Instantiate`
and `Iterate`; Declare uses the category nouns `Operation`, `Description`,
`Evaluation`, `Instantiation` and `Iteration`. Each forwards exactly one child.
These witnesses do not establish every child form, signature or execution context.
Rust FW signatures remain parameterless with
explicit unit results, Y has two fields, and PA does not introduce a new W child
role. These choices isolate adapter routing from unrelated integration limits.
No EA, native replacement loop, dispatcher alias, or compiler repair is involved.

Rust expectations are authored from literal template spans.

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
retained lock. `invocation-void` reuses the plain manifest and lock.

`invocation-invalid/` contains two malformed Receiver documents, checked by
`TestLucidInvocationInvalid` with Instantiate/Instantiation provenance.
`TestLucidInvocationIsolation` uses `invocation-named`, `schedule-composed`, then
`invocation-named`. Interpreter controls compare bare/quoted equality and
inequality. Preparation distinguishes omitted Receiver from native `Void`,
including inherited equal-byte metadata provenance.

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
| function-void-flat | Declare `Void` through composed FW with flat Key, empty Inputs and unit Outputs. Retain W/BOD/E body and a composed KY call through MOB. |
| function-void-record | Declare the same function through FW with the explicit framed Func record. Retain the same composed body and invocation. |

The two independently authored source expectations reuse the existing plain
manifest, support expectations and lockfile. Every returned artifact remains an
exact comparison.

`TestLucidFuncKeyProvenance` covers exact interpolation and bare/quoted equality
and inequality under Rust, differing record/root names, ordinary
names, empty and absent controls, and repeated interpretation without metadata,
working-variable or parsed-source mutation. Empty Key compares unequal to bare `Void`, while the explicit
prepared `Func = Void` sentinel compares equal. A completely unset Func entry
retains its empty-record fallback. It also compares unequal and is not promoted
to native-name provenance.

These callable-name witnesses do not establish general signatures.

## Typed functions and closures

The [callable contract](../../../builtins/README.md#typed-functions-and-anonymous-callables)
defines typed FW functions, bounded BW placement and KY invocation. Omitted or
empty Outputs omit the return annotation. Native Rust checks supplied types.

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
omit Key, which decodes as an empty name. The preserved record parser does not
accept the literal empty-field spelling `Key: ,`.

`callable-invalid/` contains 62 complete AML cases. Fifty-one cover required
names, bounded BW placement, KY children/metadata, Func/flat exclusivity,
unsupported Capture/headers, strict nested fields and framing errors. Eleven
cover MY signatures as described below. Interpreter controls retain existing Prop
rendering. Isolation checks cover Operate/Operation/Instantiate/Instantiation
provenance. Closure invocation, capture behavior and owner reuse remain compile-only.

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
The direct Rust decoder returns no partial records on malformed input. Complete
AML rejects malformed records before rendering with classified causes.

Eleven of the 62 callable validation cases cover supplied/empty Receiver, missing
method names, unknown/duplicate/malformed Func fields and invalid input Props.

These witnesses cover receiverless required signatures. Empty traits and local
generic/lifetime parameters have separate controls below. Composed methods,
per-method parameters and implementations remain outside this contract.

## Empty Y/HY/MY

The [empty-composition contract](../../../builtins/README.md#empty-record-and-trait-compositions)
defines braced Y/HY/MY output for empty effective collections.

`empty-instantiations` covers omitted and explicit empty Props/Funcs, braced
HY records and Y constructions through MOB, empty MY traits, a populated HY/Y
control, typed uses and trait references. The feature under test is composed.
Native fragments supply ordinary uses rather than replacement declarations.
The independent source expectation reuses the existing plain manifest and lock.

Six exact primitive cases cover the empty and nonempty routes. Twenty-one
preparation cases cover canonical missing collections, explicit empty lists,
inherited/local precedence and malformed local/inherited records. Repeated
preparation compares effective metadata by key/value/type. Original inputs,
resource data, parsed tags and source bytes remain exact and immutable.

Five stored malformed inputs in `empty-instantiation-invalid` reject empty
native collections with unrendered headers/methods, and empty HY with Shape Sum.
Isolation checks retain Instantiate/Instantiation provenance.

Empty declaration support does not implement composed members, composed
generic-header children or sum types. Local String parameters have separate
controls below.

## Keyed GI

The [keyed GI contract](../../../builtins/README.md#keyed-gi-pattern-selection)
requires every supplied pattern/body pair and only explicit terminal fallbacks.

Three complete applications use the exact-bundle/native runner:

| Case | Obligation |
| --- | --- |
| keyed-conditionals | Single and multiple arms, preserved final bindings, exhaustive Boolean arms, literal/tuple/constructor/range patterns, direct/composed terminal fallback, fallback-only, caller wildcard, quoted/qualified/parenthesized Void, nested bodies and separately grouped category ordering. |
| keyed-nonexhaustive | Two supplied integer patterns remain intact; Expected and Generated projects produce E0004 rather than acquiring a fabricated wildcard. |
| keyed-call-pattern | A native call-shaped pattern remains intact and produces E0164. No synthesized equality guard. |

Fourteen stored `keyed-invalid` documents cover cardinality, empty Key/patterns,
unexpected children and direct/composed duplicate or nonterminal fallbacks.
Five exact Evaluate primitive cases and thirteen preparation cases cover
one-shell rendering, repeated interpretation, Count restoration, inherited/local
Key precedence and native Void provenance. Filesystem/embedded A/B/A reuses
`keyed-conditionals`, existing `conditionals`, then `keyed-conditionals`, with
Evaluate/Evaluation resource provenance.

The independent oracles use the plain manifest/lock.

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
returned artifacts byte-for-byte. No output normalization or oracle-update mode
is provided.

Twenty complete AML inputs in `generics-invalid/` cover blank values, unsupported
forms, anonymous calls, inputs incorrectly placed inside Func records, duplicates
and inherited Class/Object metadata. An inherited-input EA case exercises the
file-boundary guard.

Fifty exact primitive profiles cover all ten suffix sites with absent inputs,
lifetime-only, type-only, both inputs and native type Void. Additional preparation
and lookup controls cover canonical absence, local provenance, blanks on every
supported form, inherited sentinel/blank/override rejection, repeated immutable
interpretation and generic systemless lookup. Filesystem/embedded service reuse
checks A/B/A output, earlier bundles and provenance for all four selected assets.

See the [input contract](../../../builtins/README.md#lifetime-and-generic-inputs)
for native String framing, glyph-local scope and declaration/application placement.
Trait-level inputs do not add method-specific lists. Dedicated const-generic,
independent where-clause and generic implementation interfaces remain separate work.

## Recording evidence

Follow [Run and retain evidence](#run-and-retain-evidence) for commands, failure
classification and reporting. [FAILURES.md](FAILURES.md) records the latest dated
results for their tested revision and host.
