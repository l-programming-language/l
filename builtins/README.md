# Rust logeme assets and current integration

The eighteen frozen Rust segments support selection, iteration, callable
signatures, empty records/traits, keyed matching and local lifetime/generic lists
through bounded Rust compiler preparation. L, Lyger and Logiceme are implemented
in Go. The current CLI generates one standalone Rust App. Raw Evaluate calls return
Rust source bundles; removed runtime support is unavailable.

The templates retain `Meta`, `Funcs`, `Props`, `Ops` and the existing `Conjoinments`
categories. Qualified cases can supply fragments completed by their parent.
OC/DOC supplies selection assignments. OG/DOG supplies conditional bindings.
PO defines a label directly on its Iteration. `Lifetimes` and `GenericTypes` are
an approved Rust-specific interface on the existing named declaration/application
cases. The contracts below define their bounded integration. Retained compiler
entries and unused template routes do not establish additional support.

[templates.sha256.json](../logiceme/testdata/lucid/templates.sha256.json)
freezes the current asset bytes. The
[Lucid suite guide](../logiceme/testdata/lucid/README.md) owns the current
case inventory and validation commands. Its future cases also execute generated
programs. Deployment scripts remain unexecuted.

Calculation, Description and Iteration retain their frozen composition wrappers.
Declaration renders the file header. Evaluation forwards BI's
Definitions, Operations and Default. Definition forwards PO's Iteration directly.
Execution excludes Deferred. Operation and Instantiation explicitly forward the
two local generic/lifetime inputs through their approved cases. The
[Declaration composition witnesses](../logiceme/testdata/lucid/README.md#declaration-adapters)
exercise the outer action selectors and singular child forwarding. Source-library
composition does not widen frozen assets, their interfaces or hash expectations.

## L source counterparts

[Core](eng/core.l) loads directly and supplies the implicit vocabulary, including
Print. The root [dictionary registry](../dictionaries.json) is empty. Storage,
IO, Formatting and Runtime have been removed for reconstruction in L; their
former words and generated/native support are not available.

Native source definitions bind typed operands through existing AML fields.
Source-level recognition alone does not establish executable support. The
[current build boundary](../docs/cli-build.md) is a standalone App with core
Print. Existing process-operation `.def` resources remain compiler inputs but
cannot execute without a future L-owned Runtime. Their presence does not widen
that boundary or authorize a hardcoded substitute.

## Verified compositions

The [Lucid suite guide](../logiceme/testdata/lucid/README.md) owns the complete-AML
case inventory, exact-bundle assertions, native diagnostics and execution scope.
Its [core compositions](../logiceme/testdata/lucid/README.md#core-compositions)
and [Declaration adapters](../logiceme/testdata/lucid/README.md#declaration-adapters)
record the baseline cases; later families exercise the bounded contracts below.
The inventory reports no validation run.

## Segment pairs

| Vowel / category | Consonant / primitive |
| --- | --- |
| [A / Declaration](segments/vowels/declaration.lu) | [B / Declare](segments/consonants/declare.lu) |
| [E / Calculation](segments/vowels/calculation.lu) | [D / Calculate](segments/consonants/calculate.lu) |
| [I / Evaluation](segments/vowels/evaluation.lu) | [G / Evaluate](segments/consonants/evaluate.lu) |
| [O / Definition](segments/vowels/definition.lu) | [F / Define](segments/consonants/define.lu) |
| [U / Description](segments/vowels/description.lu) | [H / Describe](segments/consonants/describe.lu) |
| [W / Operation](segments/vowels/operation.lu) | [K / Operate](segments/consonants/operate.lu) |
| [X / Iteration](segments/vowels/iteration.lu) | [P / Iterate](segments/consonants/iterate.lu) |
| [Y / Instantiation](segments/vowels/instantiation.lu) | [M / Instantiate](segments/consonants/instantiate.lu) |
| [V / Execution](segments/vowels/execution.lu) | [C / Execute](segments/consonants/execute.lu) |

## Copied forms and forwarding

This table inventories template routes, not proven end-to-end Rust support.
Singular and plural child categories retain their established meanings. No arbitrary letter
combinations or codas are implied by a row.

| Forms | Copied responsibility / inputs and children |
| --- | --- |
| A | File header: package metadata and the `Headers` import loop. |
| EA | Existing complete-file `Contents` route. It remains a standalone literal file through the existing assembler, not an inline composition workaround. |
| BA | Declare `Keys` with `Type`, without an explicit value. |
| DA, KA, HA, GA, MA, PA | Forward one Calculation, Operation, Description, Evaluation, Instantiation, or Iteration through Declare, respectively. |
| E | Emit `Meta.Statement`. |
| DE | Calculate from `Meta.Ops`. Retain the existing `Op.Type == "Calculate"` grouping condition. |
| GE | Forward Calculations and Key to Evaluate. |
| CE | Calculate communication from the existing Op operand model: one operand receives, two send. |
| I | Forward Calculations and Key to Evaluate. Retain its existing grouping selector. |
| GI | Unkeyed Boolean selection; with Key, ordered native pattern/body pairs and an optional explicit terminal Void fallback. |
| CI | Forward ordered BI Evaluations to biased Tokio selection; preparation places the sole explicit default last. |
| BI | Forward one OC/DOC Definition, zero/one W Operation and Default to one selection alternative. Default controls ordering only. |
| O, OB | Define from native Values, with ordinary assignment or declaration-assignment. |
| OC | Define `Keys = async { Values }` from exactly one caller-supplied name/wildcard and one explicit nonempty native value, without children; a BI-only fragment. |
| OG | Define `let pattern = value` from one Key/Keys pattern and one nonempty explicit Value/Values expression, without children; an X-only condition fragment. |
| FO, FOB | Define from a singular Declaration child. |
| KO, KOB | Define from a singular Operation child. |
| DO, DOB | Define from a singular Calculation child. |
| DOC | Define `Keys = async { Calculation }` from exactly one caller-supplied name/wildcard and one Calculation, without native Values; a BI-only fragment. |
| DOG | Define `let pattern = Calculation` from one Key/Keys pattern and one Calculation, without native Values; an X-only condition fragment. |
| GO, GOB | Define from a singular Evaluation child. |
| MO, MOB | Define from a singular Instantiation child. |
| PO | Define one caller-supplied label directly on one Iteration child, without native Values, an extra block or a semicolon. |
| BO | Route to Declare with Keys and Type. No explicit initializer. |
| BOD, BOK, BOH, BOG, BOM, BOP, BOC | Expose a singular Calculation, Operation, Description, Evaluation, Instantiation, Iteration, or Execution through Declare, respectively. |
| U | Describe ordered Operations. |
| MU | Describe updates using the existing Props loop. |
| W | Emit ordered Definitions. |
| BW | Anonymous callable syntax using Func and ordered Operations. |
| FW | Named function syntax using Func and ordered Operations; local Lifetimes/GenericTypes declare parameters. |
| X | Iterator traversal, unconditional repetition, Calculation-controlled repetition, or OG/DOG-controlled repetition; one Operation body. I remains a Boolean condition carrier. |
| Y | Construct a named value from Props; an empty collection emits `Key {}`. Local Lifetimes/GenericTypes supply constructor arguments. |
| KY | Invoke a named callable, an optional Receiver member, or a singular anonymous Operation, using Inputs. Only named calls accept local Lifetimes/GenericTypes arguments. |
| MY | Receiverless required trait signatures from Funcs; an empty collection emits `trait Key {}`. Local Lifetimes/GenericTypes qualify the trait once. Omitted/empty method Outputs omit the return annotation. |
| HY | Struct-shaped syntax from Props; an empty Record emits `struct Key {}`. Local Lifetimes/GenericTypes declare parameters. |
| V | Schedule the invocation text in Key through `tokio::spawn`. |
| MV | Schedule a singular Instantiation child through `tokio::spawn`. |
| CV | Construct an unscheduled future from one W body; local Capture selects Infer or Move. |
| DV | Await one native or composed future operand inside an asynchronous body. |
| CO / COB | Assign / declare one binding from one CV or DV value. |

The two local generic/lifetime inputs below qualify existing named declarations
and applications without adding structural-parameter syllables or generic-path
roles. CV/DV and CO/COB have the [future contract](#future-construction-awaiting-and-value-bindings) below. Enum, general structural-pattern,
mutability and implementation extensions remain outside these contracts.
OC/DOC and OG/DOG do not establish those interfaces.

## Lifetime and generic inputs

Optional Rust-only `Lifetimes` and `GenericTypes` String inputs contain native
comma-separated list interiors without enclosing angle brackets:

```text
Lifetimes: ~'a, 'b~
GenericTypes: ~T, U: Clone + 'a~
```

| Forms | Meaning | Example |
| --- | --- | --- |
| FW | Declare function parameters, beside flat metadata or an existing Func record | `fn borrow<'a, T>(...)` |
| HY | Declare record parameters | `struct Borrowed<'a, T> { ... }` |
| MY | Declare trait parameters once, outside its Funcs signatures | `trait Reader<'a, T> { ... }` |
| Y | Apply constructor arguments | `Borrowed::<'_, u8> { ... }` |
| Named KY | Apply arguments to the final named function/member component | `borrow::<u8>(...)`, `receiver.method::<T>(...)` |

The conditional suffixes emit nothing when both inputs are absent. If both are
supplied, Lifetimes precedes GenericTypes with one template-owned `, ` between
the inputs. Callers own separators within each input. Supplied text is preserved
without generic-list parsing, additional unframing, whitespace normalization or
name extraction. Empty or whitespace-only supplied values receive `InvalidInput`.
Native `GenericTypes: ~Void~` remains a supplied type rather than absence.

These inputs are glyph-local. Effective inherited Class/Object/file metadata is
rejected, including blank or sentinel-valued entries and files using standalone
EA. A local override does not rescue inherited placement. Parent parameters may
appear in child native types, but no parameter list propagates into child calls
or neighboring declarations. Func/Prop/Op records remain unchanged. These
inputs are not record fields. FW/MY reject these names inside Func records.
BW, anonymous KY and other forms do not accept these inputs. Named KY requires
a locally supplied Key.

Arguments attach to KY's final named component: Key `Vec::new` plus GenericTypes
`u8` spells `Vec::new::<u8>()`, not `Vec::<u8>::new()`. Callers can supply the
latter complete native path as Key without new metadata. Trait-level parameters
may be referenced by existing MY method input/output types. They are not repeated
after method names and do not define per-method parameter lists.

Rust checks native syntax, argument arity, bounds, ownership and lifetime rules.
Declaring a lifetime does not imply that callers may explicitly supply it: the
native suite retains E0794 for a late-bound function lifetime. Empty generic
structs retain their declared fields exactly and may receive E0392 for unused
parameters. The compiler inserts no `PhantomData`. E0107 verifies that excess
call arguments are not removed. No lifetime inference, cloning, borrowing,
`'static`, capture policy, argument inference or implicit default is introduced.

The four assets use conditional suffixes, explicit forwarding and Inputs
declarations. Preparation installs canonical
absence and native-source provenance. A Rust-only direct-primitive lookup fallback
handles missing values for these two names alone. Generic systemless lookup and shared ordered rendering remain unchanged. Dedicated const-generic, independent `where`, generic
implementation and per-method interfaces remain separate work.

The [suite guide](../logiceme/testdata/lucid/README.md#lifetime-and-generic-inputs)
records the witnesses, focused controls and compile-only evidence boundary.

## Literal syntax translations

- Declare uses typed `let` syntax. It adds no initializer, automatic `mut`,
  `Default::default()`, or clone. Declaration-assignment uses `let ... = ...`.
  Ordinary assignment remains `... = ...`. Statement definitions, declarations,
  and MU updates have semicolons. OC/DOC and OG/DOG fragments do not. PO supplies
  the label colon and its Iteration child without another body boundary. W's copied Definition
  loop is unchanged.
- Functions and anonymous callables use Rust `fn`/closure punctuation with the
  original Func interpolation slots. The templates add no implementation wrapper, signature
  preparation, `move`, or async qualification. The compiler restores existing
  Func records. Rust's `Func.Inputs~` formatter supplies `name: type` parameters.
  Omitted/empty Outputs omit the whole return annotation. Explicit Outputs retain
  their existing punctuation. FW keeps its opaque
  Receiver interpolation inside the parameter list, without adding conditions
  to format separators or receiver arity.
- Interface and record shells use `trait` and `struct`. Fields use colons and
  commas. Method signatures use Rust punctuation. StructTags text is retained as
  a comment rather than interpreted as Rust attributes.
- Calls and named-field construction retain their existing child composition.
  If chains remain if chains. Keyed switches use `match` punctuation, with no
  generated selector binding, comparison guard, or ownership decision.
- Conditional and infinite loops use `while` and `loop`. Range branches use
  `for ... in ...`. No iterator adapter, enumeration, borrowing, or channel-drain
  implementation is inserted.
- CE's receive/send leaves use `.recv().await` and `.send(...).await`. They do not
  unwrap, bind, or otherwise interpret `Option`/`Result`, or check async context.
- V/MV use `tokio::spawn(invocation)`, without generating an async wrapper. The
  supplied invocation must already return a suitable future.
- A preserves package metadata as a comment and uses the existing Headers loop
  for `use` statements. It prepares no safety attributes, modules, or imports.

Interpolation values remain opaque. The templates do not repair incorrectly formatted
parameters, receiver/result lists, type spellings, quoted imports, operators,
multiple-key bindings, or loop inputs. Unsupported tuple bindings and non-pattern match cases may produce invalid or
incomplete Rust. Optional result slots and empty structure collections have the
bounded contracts documented below. Raw multi-line StructTags text is not normalized into a safe
multi-line comment. These are documented draft limitations, not promises of
implicit conversion or validation.

## Named invocation contract

KY's named-call route consumes its supplied Key and Inputs. Omitting Receiver
prepares the ordinary absence sentinel and selects `Key(Inputs)`. A supplied
Receiver retains its exact bytes and selects `Receiver.Key(Inputs)`. Explicitly
empty or whitespace-only Receiver is rejected with `InvalidInput`:

```text
Rust KY: Receiver must be nonempty when supplied
```

Rust distinguishes native identifiers spelled `Void` from the absent-value
sentinel. Rust prepared `SourceString` values do not equal an unquoted `Void`
in template conditions. Quoted `"Void"` comparisons remain textual. Absent Receiver replaces inherited local provenance even when its
bytes already equal `Void`, without mutating file-owned metadata.

The [invocation witnesses](../logiceme/testdata/lucid/README.md#named-invocation)
isolate these calls from general signature generation. They do not establish BW
or KYC semantics. Callable-name provenance has the separate contract below.

## Native callable-name provenance

Exact `Meta.Func.Key` interpolation preserves a nonempty native name, including
`Void`. The Rust provenance helper uses the same Func decoder as interpolation.
The Func record remains authoritative
when root `Meta.Key` differs. Bare `Void` compares unequal to a supplied native
name, while quoted `"Void"` remains an ordinary textual comparison.

Output provenance has a [separate contract](#native-output-type-provenance).
Empty Key and empty Inputs/Outputs retain their existing behavior, as does absent
Receiver. An explicitly prepared
`Func = Void` sentinel compares equal to bare `Void`. An unset Func dictionary
entry falls back to an empty Func whose Key compares unequal. Neither absence
representation gains native-name provenance.

The [callable-name witnesses](../logiceme/testdata/lucid/README.md#native-callable-name-provenance)
cover flat and record FW names. Other Func fields and indexed provenance require
their own contracts.

## Typed functions and anonymous callables

Func/Prop records supply callable signatures. FW requires a nonempty name.
BW creates a synchronous anonymous closure and is allowed only
as KY's direct callable child or KO/KOB's direct value child within an enclosing
callable. Named KY accepts no callable child. Unnamed KY requires exactly one BW
and excludes Receiver. KO/KOB retain ordinary assignment/binding punctuation.

Both FW and BW require a name and type for each input Prop. Only Key/Type are
accepted there. Supplied Value, StructTags and unknown fields are rejected.
Flat and explicit Func signature metadata are exclusive. Func records accept
only Key/Inputs/Outputs/Receiver, and unknown Receiver-Prop fields cannot silently
become an empty receiver. BW excludes flat Key/Receiver and rejects a nonempty
decoded record name or receiver. A Func record may omit Key, decoding to an empty
name. The existing parser still rejects the literal empty-field spelling
`Key: ,`.

Omitted or empty Outputs omit the whole annotation: named functions return unit,
while closures use Rust output inference. Explicit output punctuation, including
unit parentheses, is unchanged. Closure parameter types are still required.
Native Rust infers captures. No `move`, clone or capture policy is synthesized.
Supplied Capture and generic BA/U header children are rejected on FW/BW because
the current templates do not consume those interfaces. Other advanced callable
forms retain their existing paths without new support claims.

The [callable witnesses](../logiceme/testdata/lucid/README.md#typed-functions-and-closures)
cover signatures, optional outputs, closures and the intentional unit mismatch.
Receiver methods, FWC/KYC and general ownership analysis remain outside this
contract. Supported output and indexed provenance fields are specified below.

## Future construction, awaiting, and value bindings

`CV` constructs an unscheduled future from exactly one `W` body. Its optional,
glyph-local `Capture` is `Infer` (the default) or `Move`. These select Rust's
inferred capture or ownership transfer. Neither form clones captures or starts a
task. The body's final expression supplies the future's result.

`DV` awaits one future within a `CV` body. Supply either a nonempty native `Key`
or one `E`, `KY`, `CV`, or `DV` operand, never both. The operand is evaluated once.
Rust checks its Future implementation and result type. Awaiting does not unwrap
errors, select a result representation, or imply process supervision.

`CO` assigns, and `COB` declares and initializes, one simple `Key` (or one-element
`Keys`) from exactly one `CV` or `DV` child. Raw identifiers and `_` are accepted.
Type, mutability, native Value fields, and scheduling statements are excluded.
For mutation, declare the storage explicitly before assigning through `CO`.

`FW`, `HY`, and `KY` retain definition, structural-data, and invocation ownership.
Existing glyph-local `Lifetimes`/`GenericTypes` express signature relationships.
CV does not infer `'static` or introduce lifetime metadata. Nested synchronous
callables reset the await context. Existing V/MV scheduling and OC/DOC selection
forms retain their separate contracts.

The complete `future-*` AML/Rust pairs cover composed invocation and native
operands, single evaluation, inferred and owned captures, borrowed signatures,
and Rust rejection of escaping borrows and moved-value reuse. They add no L
syntax or Function implementation.

## Native output-type provenance

Exact `Meta.Func.Outputs~` interpolation preserves supplied Rust type text,
including a sole type named `Void`. The existing provenance helper reads the
same decoded Func as interpolation and recognizes only nonempty Outputs.
The frozen FW/BW punctuation therefore remains `-> (Void)`. Missing and empty
outputs and the explicit Func absence sentinel retain their existing behavior.
Bare `Void` remains an absence comparison. Quoted `"Void"` compares text.

The [output-type witness](../logiceme/testdata/lucid/README.md#native-output-type-provenance)
covers flat/record FW and BW. Indexed Funcs provenance has the separate MY
contract below.

## MY required trait signatures

MY renders ordered receiverless signatures from its existing `Funcs` records.
Omitted and explicitly empty Outputs emit no annotation. Unit, scalar and multiple
supplied outputs keep the parenthesized punctuation and terminating semicolon.

Rust decodes each Func through the existing balanced record machinery and uses
the current `name: type` parameter formatter. Nested commas in tuples, generics
and function-pointer types retain their source meaning. Exact indexed `Key` and
`Outputs~` fields preserve nonempty native names/types, including `Void`, using
the selected record and current Count/Add binding. Other field provenance remains
unchanged.

MY validates named signatures and named/typed input Props, rejects malformed or
unsupported record fields, and rejects any supplied Receiver field, even empty.
Its frozen template does not consume Receiver. Empty trait emission has the
bounded contract below. Trait-level generic/lifetime inputs have the contract
above. Composed/default methods, independent method parameters and implementations
remain separate work. Retained preparation entries are not proof of those routes.

The [MY witnesses](../logiceme/testdata/lucid/README.md#my-required-signatures)
cover required signatures, native provenance, malformed records and isolation.

## Empty record and trait compositions

Y with empty Props emits `Key {}`, HY with empty Props emits `struct Key {}`, and
MY with empty Funcs emits `trait Key {}`. Omitted collections have the same result.
Record braces are preserved. HY does not substitute a unit-struct declaration.
The existing nonempty loops, selectors, inputs and forwarding remain unchanged.

Rust preparation supplies canonical `&[]&` only when the effective collection is
missing. It preserves inherited metadata and explicit local overrides, and uses
the existing record readers to reject malformed data rather than treating it as
empty. This is a Rust preparation contract. Shared Lucid absence behavior is unchanged.

An empty HY/MY native collection cannot accompany composed children, including
member and generic-header children: the frozen wrappers do not render those
collections. Such input receives a classified rejection rather than losing its
composition. Empty HY accepts absent Shape or `Shape: Record`. `Shape: Sum` is
rejected rather than emitted as a struct. This does not implement composed
fields/methods, composed generic-header children or sum types, nor certify their
nonempty paths. The separate local String inputs above also apply to empty shells.

The [empty-composition witnesses](../logiceme/testdata/lucid/README.md#empty-yhymy)
cover omitted/empty collections, populated controls, preparation and rejection.

## CI/BI selection contract

CI owns the `tokio::select!` shell. OC/DOC own the complete `left = async { value }`
Definition, and BI supplies only `=>` and the selected body. CE owns its awaited
channel expression. An ordinary O owns any explicit assignment inside W. No
public BI Key, extra metadata, record field or logical category is introduced.

CI accepts only BI children, at most one default, inside a callable. Every BI,
including `Default: True`, requires exactly one OC/DOC Definition and zero/one W
body. Bare Calculations and DO/DOB guards are rejected. OC/DOC are allowed
only as direct BI children. They are not standalone statements or general future
adapters. Native async scaffolding remains explicit. Cargo checks actual placement.

OC consumes exactly one Key/Keys entry and one explicit, nonempty Value/Values
entry, without children. DOC consumes exactly one Key/Keys entry and exactly one
Calculation, without Value/Values. DOC accepts an otherwise valid Calculation,
including E, CE send and CE receive. The async block returns that supplied value.
It does not automatically await a future returned by it. Both allow Imports and
Tags. The caller must choose a simple/raw identifier or `_`. Composite patterns,
assignment places and Type/Mutable/Parameter/Capture qualifiers are rejected.

| Source alternative | Emitted behavior |
| --- | --- |
| DOC name with CE receive | `name = async { CE }`. The selected body receives the explicit Option result. |
| DOC `_` with CE send/receive | `_ = async { CE }`. The caller explicitly discards the returned result. |
| DOC name with another Calculation | `name = async { Calculation }`. The supplied expression remains the tail. |
| OC with Value `()` | `name_or_wildcard = async { () }`; an explicitly supplied ready alternative. |
| Default True | Place the supplied Definition last, preserving other branch order; invent no name, value or readiness. |
| Empty CI | Existing I/E composition emits `{ ::core::future::pending::<()>().await }`. |

A named selection pattern creates a branch-local binding. To update an outer
variable, the caller supplies the branch name and an ordinary O inside W, such
as `selected = async { receiver.recv().await } => { existing = selected; }`.
The compiler does not derive a BI Key or introduce temporary bindings, assignments,
drop calls, cloning, moving or unwrapping. `_` follows Rust/Tokio wildcard behavior.
Callers needing explicit disposal before later statements supply a named result
and their own drop statement. Destructor timing has no runtime verification.
Ordinary O/OB/DO/DOB statement punctuation remains unchanged.

`biased;` gives earlier ready branches priority and can starve later branches.
With a caller-supplied ready final alternative, fallback occurs when preceding
futures report Pending during polling. This is not an atomic readiness test or
Tokio's all-disabled `else`. `Default: True` alone gives no readiness guarantee.
Selection runs on the current task. Async-block operands execute when polled,
without promising eager evaluation of every operand. See
[Tokio selection](https://docs.rs/tokio/1.53.1/tokio/macro.select.html).

Closed receive results remain None for explicit handling. Cancelled receives do
not consume a message. A cancelled send can drop its unsent payload. The caller
chooses whether to bind or discard the send Result. Reservation/permit-based
lossless sending is outside this contract. See
[Tokio send cancellation](https://docs.rs/tokio/latest/tokio/sync/mpsc/struct.Sender.html#method.send).
Verification of these selection witnesses remains generation/compilation only.

## Iteration and Definition contract

Iterate owns repetition syntax and the body boundary. Every X has exactly one
Operation body and one of four abstract behaviors:

| Behavior | Supplied inputs/children | Emitted shape |
| --- | --- | --- |
| Iterator traversal | Iterable with the existing Key, Index and RangeArity metadata. No condition child | `for pattern in iterable { body }` |
| Unconditional repetition | No Iterable or condition child | `loop { body }` |
| Conditional repetition | One Calculation, or one I/Evaluation adapter containing ordered Calculations | `while condition { body }` |
| Definitional repetition | One OG or DOG Definition | `while let pattern = value { body }` |

The I route represents the same Boolean condition category. It is not a fifth
loop mode. Conflicting condition categories, repeated conditions, iterator/condition
mixtures and missing or extra bodies are rejected before rendering. An ordinary
O/OB/DO/DOB is not a conditional Definition. Initializer/Evaluation/step composition is rejected. X provides no three-part
lowering, hidden initializer, first-iteration flag or automatic post-step.

OG consumes exactly one nonempty Key/Keys pattern and one explicit, nonempty
Value/Values expression with no children. DOG consumes exactly one nonempty
Key/Keys pattern and one Calculation with no Value/Values metadata. Both are
permitted only as X's direct conditional Definition. Define supplies `let` and
`=` without a semicolon. Callers supply the pattern, not an entire let expression.
Nested tuple/record punctuation and caller-written binding modifiers remain
native data. Framed pattern commas do not create extra Keys. Rust checks native
pattern, type and ownership legality. The compiler checks composition and
cardinality. No structural-pattern forms or pattern parser are introduced.

PO consumes one nonempty Key/Keys label and exactly one Iteration child, without
Value/Values or a coda. It is a Definition statement in W within a callable.
The caller supplies the native spelling, such as `'outer`. Define supplies the
colon and forwards `Conjoinments.Iteration` directly, never through Meta.Values:

```rust
'outer: loop {
    continue 'outer;
}
```

There is no labeled wrapper block. Labels, nested break/continue targets and body
scope are preserved without generating or renaming identifiers.

RangeArity 0 emits `_`, 1 emits Key, and 2 (also the existing absent-arity default)
emits `(Index, Key)` or `(_, Key)` when Index is absent. The iterable must already
yield the required items. Caller-supplied ranges, `enumerate()`, reversal,
filtering and other library expressions need no extra loop category. The compiler
adds no iterator adapter, borrowing or cloning. DOG can consume an explicitly
awaited Calculation with caller-supplied async scaffolding. There is no inferred
async mode, channel-draining primitive or async wrapper.

All ordinary O/OB/DO/DOB and OC/DOC contracts remain unchanged. The supplied
template interpolation modifiers own layout. Native multiline-literal formatting
is not broadened. Complete-AML and primitive fixtures
cover the approved branches, with intended native diagnostic cases for a non-Boolean
condition, an out-of-scope binding and awaiting outside async. They compile only:
execution order, termination and Rust runtime parity are not established.

## Execution without Deferred

Execute/Execution have no Deferred input, forwarding or conditional branch. V
schedules its supplied Key expression, and MV schedules its supplied Instantiation
through `tokio::spawn`. These templates add no invocation, async wrapper, join,
clone, ownership conversion or cleanup operation. The caller supplies a suitable
future and execution context. Other execution-role and callable integration gaps
remain separate follow-up work.

Supplying Deferred in Rust AML is still rejected before rendering with the
existing `Unsupported` classification and diagnostic, including False:

```text
Rust <form>: Deferred is unsupported; Rust has no deferred execution operation
```

Omitting Deferred is the supported interface. No scope-exit syllable or facility
is defined. The 18 Rust segment assets contain no `compile_error!` marker. This
is an asset inventory fact, not a claim that all template routes or compiler
preparation branches are integrated.

Rust resource cleanup can use ownership and destructors (RAII), governed by
Rust's [drop scopes](https://doc.rust-lang.org/reference/destructors.html#drop-scopes).
The borrow checker checks ownership and borrowing constraints at compile time.
It does not execute cleanup. Ordinary [`Drop::drop`](https://doc.rust-lang.org/std/ops/trait.Drop.html)
is synchronous. Callers own any explicit guards, cleanup requiring error handling,
and asynchronous shutdown. Releasing a resource does not establish a successful
flush, completed transaction or completed shutdown.

RAII does not imply deferred-call argument evaluation or function-exit ordering. The logemes synthesize no guard or Drop implementation.
[`std::process::exit`](https://doc.rust-lang.org/std/process/fn.exit.html) does not
run stack destructors. The scheduling witnesses remain compile-only and establish
neither destructor timing nor shutdown behavior. The suite's separate
[future execution checks](../logiceme/testdata/lucid/README.md#futures) exercise
construction, captures and awaiting.

The [scheduling witnesses](../logiceme/testdata/lucid/README.md#scheduling-and-deferred-rejection)
cover V/MV and Deferred rejection. [Named-invocation witnesses](../logiceme/testdata/lucid/README.md#named-invocation)
separately cover receiverless KY.

## Keyed GI pattern selection

Evaluate's keyed GI branch emits one `match` shell and preserves
all supplied Calculation patterns and Operation bodies in their category order,
including a single arm and the final arm. Children need not alternate in AML.
Every body requires an explicit pattern. At least one pair is required.

`Key` supplies the native subject. Effective inherited metadata and local
overrides are honored, including native `Void` provenance. An empty or
whitespace-only subject is invalid. Unkeyed GI remains the existing Boolean
condition route. A bare identifier in a keyed arm follows Rust's binding rules,
not an inferred equality comparison. Rust checks native pattern syntax, types,
borrow/move rules and exhaustiveness. No equality guards, clones or structural
pattern forms are generated.

A Calculation that compares equal to the existing standalone `Void` carrier
after composition supplies the fallback. This includes `GE^E{Statement: Void}`.
There may be one fallback, and it must be terminal. Missing patterns, empty
pattern output, duplicate/nonterminal carriers and unexpected child categories
are classified composition errors. No child is silently discarded. The compiler
checks compiled children with the same indexed comparison used by the template
without rewriting their source.

Quoted `"Void"`, qualified names such as `Choice::Void` and parenthesized native
patterns remain literal patterns. Use qualified spelling for a native constant
whose bare name would collide with the fallback carrier. Caller-written `_`
remains ordinary Rust pattern syntax. The compiler does not analyze its coverage
or invent a fallback from the final body's position.

The [keyed-GI witnesses](../logiceme/testdata/lucid/README.md#keyed-gi)
cover ordered matching and native rejection of missing coverage or call-shaped
patterns without invented wildcard/equality repairs.

## Current integration limits

Rust iteration uses the existing Void sentinel for absent named forwarded singular
children, and conjunctions require all supplied clauses to be true. Empty plural
collections and explicit keyless records retain their behavior.

A template case, a parser/validation entry, successful generation and a compiled
application are separate levels of support. The remaining mismatches below are
follow-up work. This reference does not authorize repairs or interface changes.

| Area | Current state and limitation |
| --- | --- |
| Named invocation and callable names | KY's omitted Receiver/native scalar `Void` handling and exact `Meta.Func.Key` provenance have the bounded contracts above. Singular output and indexed MY Key/Outputs~ provenance are covered separately; other Func fields, general signatures and KYC semantics remain outside these contracts. |
| Typed and anonymous callables | FW/BW and named/anonymous KY have the bounded contract above. BW is restricted to KY or KO/KOB inside callable scope; inferred parameter types, function receiver methods and generic closure declarations remain unsupported. FW parameters have the separate local-input contract above. |
| Deferred | Absent from Execute/Execution inputs and forwarding; any supplied Deferred, including False, retains its existing compiler rejection. No deferred-execution operation is implemented. |
| Selection | CI/BI supports the [bounded selection contract](#cibi-selection-contract). No fairness, cancellation, destructor timing or Rust runtime behavior is established by compilation. |
| Three-part loops | Unsupported by X and rejected before rendering. Iterate has no three-part marker or compiler expansion. |
| Future composition | CV/DV construct and await futures; CO/COB bind their values under the [bounded contract](#future-construction-awaiting-and-value-bindings). FW remains synchronous and can return a future through its existing output signature. Function receivers, FWC/KYC and generic closures remain outside these contracts. |
| Output-field provenance | Exact `Meta.Func.Outputs~` preserves supplied native types, including `Void`, for flat/record FW and BW. Empty-output behavior is preserved. Indexed Funcs Key and Outputs~ have the bounded MY contract above; other indexed fields are not broadened. |
| MY trait signatures | Receiverless Funcs records and optional outputs are covered above. Empty traits have the bounded empty-composition contract; supplied Receiver metadata is rejected. Trait-level parameter lists use the local-input contract above. Composed methods, per-method parameters and implementations remain outside these contracts. |
| Generics, enums and patterns | Local String parameters/arguments are supported only on FW/HY/MY/Y/named KY as described above. Const-generic, where-clause, per-method and implementation interfaces remain separate work. OG/DOG do not establish general compositional patterns or enums. |
| Header/module/runtime coordination | A retains its package comment and Headers/use loop. The source assembler emits graph-owned safety/module declarations separately at the known A boundary, and configuration initializes before user effects. Explicit library-only targets omit executable support. Runtime resources are absent during the reset. |
| Ownership and mutability | The standalone Lucid mutation witnesses supply native text explicitly. No general source ownership/lifetime inference or runtime snapshot behavior follows from these template contracts. |
| Iteration and keyed GI | X has the bounded four-category contract above. Keyed GI preserves native patterns and explicit fallback carriers as described above. Iterator adaptation and native pattern legality stay caller-owned/rustc-checked. These standalone branch/loop witnesses remain compile-only. |

The compiler separates Inputs/Headers initialization from executable case
selection, dispatches parsed primitive calls while preserving wrapper whitespace,
prepares the existing metadata records and exposes only explicitly forwarded
child aliases. These boundaries do not establish support for unrelated preparation paths.

Templates own whitespace and punctuation. The current source expectations were
independently derived from their literal spans, including the explicit `%` and
`&` interpolation modifiers, and compared without formatting or normalization.

## Frontend topology and library packages

The source assembler uses StructureVersion 2 Target/Module descriptors, not
completed-source inspection. Nonprivate composed FW/HY/MY items in a library
Object are public. Executable-package items retain their existing crate visibility.
Module.Visibility separately controls module exports. Objects remain files and
Classes remain source/module groups without runtime scopes.

A sole `Kind: "lib"`, `Primary: false` target is a library-only Cargo package. It
returns declared sources and Cargo.toml, with no config/runtime/start/stop files.
Executable SETTINGS reject on that package. A library cannot
be Primary. Mixed binary/library packages still require exactly one Primary
binary and at most one library. An independent path-dependent binary checks
public interfaces and shared nominal types in the library regression. This is
separate from the [Lucid inventory](../logiceme/testdata/lucid/README.md#current-inventory).

Raw AML library targets and explicit native consumers retain this library-only contract. These backend targets do not imply current CLI support for L Service workspaces. See [CLI semantics](../docs/cli-build.md).

## Support assets and verification

Support resources have responsibilities separate from the eighteen segment templates:

| Resource | Responsibility |
| --- | --- |
| [config/config.lu](templates/config/config.lu) | Startup configuration. |
| [install/cargo.lu](templates/install/cargo.lu) | Cargo manifest. |
| [operate/start.lu](templates/operate/start.lu) | Native build and launch script. |
| [operate/stop.lu](templates/operate/stop.lu) | Explicit-PID stop script. |

The six core nonruntime bundles verify the common manifest, config and start/stop
artifact bytes. They do not execute deployment artifacts or establish process
supervision. Core Print execution is checked by the temporary hello-world gate.

The [suite guide](../logiceme/testdata/lucid/README.md#run-and-retain-evidence)
owns commands, pins, locks and artifact reporting. Compile generated files without
repair or substitution. Preserve the [validation budgets](../docs/development-harness.md#validation-budgets).

Preserve shared interpreter regressions and L/Lyger integration checks.
The [acceptance guide](../docs/development-harness.md) owns the current temporary
gate. Do not infer a new gate result from documentation edits, broaden or replace the suite incidentally, or
update template hashes to accommodate production changes.
