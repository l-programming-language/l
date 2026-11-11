# Logeme implementation contracts

This document describes current assembly and verification contracts. It does not
certify every template route or authorize changing frozen assets.
Use [LOGEMES.md](LOGEMES.md) for semantics, [the backend architecture](../logiceme/ARCHITECTURE.md)
for ownership boundaries, and [AML.md](../logiceme/AML.md) for input structure.

## Authority and change scope

Follow the current user-approved scope. These contracts do not authorize
replacing templates, adding inputs or changing callers. The
18 Rust segments remain frozen. Keep `segments/`, `templates/`, their interfaces
and frozen hash expectations unchanged. Consult the [current Rust reference](README.md).
Other fixtures and tests may change within authorized reset work. Keep useful
independent behavioral assertions; never update expected bytes merely to match a bug.

Proposed syllables are not operational until parsing, template cases, forwarding,
validation, documentation, and end-to-end tests agree. Preserve established
pairings until an explicit change addresses their callers. Do not grant every
letter permutation meaning or silently reuse an occupied spelling.

## Logical ownership

1. **Declaration does not define.** The current Rust declaration primitive emits
   typed `let` without an initializer or automatic mutability. `Default::default()`,
   zero values and invented clones do not belong in this primitive.
   Definition/assignment is F/O-owned.
2. **Definition does not silently declare.** Introducing a binding requires a
   declared composition, such as terminal B or the scoped OG/DOG conditional
   Definition fragment. Existing adapters have explicit
   meanings. Do not infer assignment or binding from category alone.
3. **Mutability follows typed semantics.** The frontend may select mutable storage
   from resolved assignment paths without a new glyph input. Do not introduce a
   template mutability interface. Do not
   make every local mutable, add `mut` to `_`, or infer arbitrary native captures.
4. **Calculation is not conditional selection.** Comparisons produce conditions.
   Evaluations choose alternatives. Type constraints and structural patterns
   must not acquire G/I merely because the compiler checks them.
5. **Grouping is not iteration or functionalization by punctuation.** P/X owns
   repeated program logic. K/W owns operations, not every pair of parentheses.
6. **Asynchrony is not a lifecycle shortcut.** Invocation, future construction,
   awaiting, scheduling, joining, cancellation, and selection have separate
   contracts. No implicit detach, await, spawn, or shutdown side effect.
7. **Description preserves structure.** Field/pattern descriptions do not perform
   hidden control flow, binding, mutation, or ownership conversion.

For example, target syntax may express declaration and definition together as
`let x: i32 = 3;`, but the logeme composition must account for both responsibilities.
Keep declarations free of synthesized initialization and keep each primitive
focused on its logical responsibility.

## Small templates, explicit composition

A segment contributes one logical fragment, its local punctuation, and its
declared child slots. Vowel cases select/forward compositions. Consonant cases
render their focused primitive. Keep the nine logical pairs and existing names
`Meta`, `Conjoinments`, `Funcs`, `Props`, and `Ops`.

- Source L owns library algorithms, callback policy, error policy and resource
  selection. Lyger lowers grammar and control flow through frozen AML. Native leaves perform one call, operator, constructor,
  borrow/projection or bounded pattern/binding. Do not hide algorithms in native
  helpers, closures, Statement or EA. Reuse source compositions and preserve typed
  input/intermediate binding identity, scope and provenance.
- Local loops over fields or children are appropriate. Dependency resolution,
  ownership inference, process supervision, and substantial validation are not.
- Do not parse rendered child strings to recover lost roles. Preserve roles and
  source provenance through preparation and composition.
- Specify singular/plural forwarding, child order and cardinality explicitly.
  Missing, extra, conflicting, or wrong-category children must not be discarded.
- An adapter preserves its child's expression/statement/item/type/pattern role.
  Current W appends explicit template newlines. Definitions and native leaves
  supply their own statement punctuation. Do not add duplicate terminators.
- Leave unknown or invalid inputs to classified diagnostics. Do not select an
  unrelated default case to make compilation proceed.

## Metadata, values, and ownership

Func/Prop/Op are existing metadata groupings, not logical classifications. Their
redesign is out of scope. Do not add template metadata or change its interfaces.
Compose the existing forms with resolved typed bindings.
Do not introduce target-prefixed duplicates or opaque whole-program inputs to
bypass composition.

The [Rust lifetime/generic contract](README.md#lifetime-and-generic-inputs)
owns the glyph-local String interfaces and their placement, absence and provenance
rules. Preserve native interiors and explicit forwarding without record-field
extensions, inherited defaults, argument inference or propagation to child glyphs.

L's types, aliases and value ownership do not change with the target.
Map them to Rust or accept deliberately supplied native types in developer
definitions. Prepare target-specific parameter, result, receiver, literal, and
iterator formatting before the tiny renderer needs it.

Preserve typed bindings, once-only evaluation and source provenance. Missing
lifetime metadata is not permission to clone, invent a value or widen its lifetime.
Native ownership and lifetime constraints remain subject to Rust validation.

## Source-unit declaration and headers

Each ordinarily composed source file begins with A in its AML content, including
helpers and companion application entry files. A declares the source unit/header.
It is not merely an import printer. A Book identifies the application, while its
Sentence/Object and outline identify each file.

The compiler prepares header needs from the complete file's compositions and
the selected support plan before header emission. A consumes its declared header inputs.
It must not scan directories, resolve packages, infer a module tree from source
strings, or initialize a runtime.

Frozen Rust A emits a package comment
and `use` statements through the existing Headers loop. Compiler preparation
collects file-owned imports, including imports from nested compositions. The
maintained suite checks header bytes and import isolation. Ordinary composed
Rust files do not currently receive crate safety attributes or module declarations
from A.

The source assembler now emits safety attributes and explicit module items from
the source graph at the compiled A boundary. Configuration initialization precedes user effects. Keep these operations
out of source rewriting, namespace flattening and segment templates. An App book
is not an implicit Rust module. Imports, modules and Cargo dependencies remain
separate data. Sole-library targets emit no executable support, export nonprivate
composed items and reject executable settings/runtime requests. See the
[Rust topology contract](README.md#frontend-topology-and-library-packages).

## Interpretation and assembly integrity

- Preserve literal prefix, interstitial, and suffix bytes around nested Lucid
  tags. Render child occurrences in order, including identical siblings.
- Render the complete loop body for every iteration. Evaluate the iterable before
  binding. Restore any shadowed binding in full afterward.
- Templates own their separators. Do not add synthetic loop newlines or strip
  whitespace/comments from completed source to hide an interpreter defect.
- Metadata and generated child text are data, not new template instructions to
  rescan. Preserve established scoped substitution and literal framing.
- Keep input bytes and prior returned bundles immutable across calls. Target,
  imports, settings, bindings, and file state must not leak between evaluations.
- Resolve assets through the selected filesystem. An explicit provider is
  complete. Missing assets must not fall back to unrelated host files.
- Preserve ordered books/files, validate final paths and support collisions,
  and return errors with causes and a zero-value bundle rather than partial output.

## Literal-file exception

EA is an emergency developer escape for one complete literal file, not the normal
way to implement logeme programs. Its Sentence/Object supplies the path. Contents
is required and copied verbatim, including explicitly empty content. It stands
alone, without A or composed siblings. Preserve this boundary, but do not use
whole-file native contents to replace the feature under test or expand EA testing
at the expense of ordinary composition coverage.

## Current bounded Rust contracts

The [Rust reference](README.md) owns detailed form and metadata contracts.
The following boundaries guide changes without treating every retained compiler
entry as supported:

- **Callables and records:** FW/BW use existing typed Func/Prop inputs, and KY
  invokes named callables or its permitted anonymous BW child. MY renders ordered,
  receiverless required signatures from Funcs. Missing/empty outputs omit return
  annotations. Supplied native names and output types, including `Void`, retain
  their supported field provenance. Function/method-signature receiver emission, async definitions,
  per-method parameters and implementations remain separate work. See
  [callables](README.md#typed-functions-and-anonymous-callables) and
  [MY signatures](README.md#my-required-trait-signatures).
- **Empty records and traits:** Y/HY/MY emit braced empty constructors, records
  and traits for genuinely missing or explicitly empty native collections.
  Preparation preserves inherited collections and local overrides. Malformed
  metadata must remain an error. Empty HY/MY must not discard composed children,
  and empty HY cannot turn `Shape: Sum` into a struct. See
  [empty compositions](README.md#empty-record-and-trait-compositions).
- **Selection:** OC/DOC owns the complete caller-supplied `left = async { value }`
  fragment. BI consumes that Definition and adds the selected body. CI owns
  biased selection. Callers supply names, wildcards, ready values and any outer
  assignment. Default orders branches only. Preparation adds no temporary or
  drop. OC/DOC remains BI-only. See the
  [selection contract](README.md#cibi-selection-contract).
- **Iteration:** X owns traversal, unconditional repetition, Boolean repetition
  and conditional-definition repetition, with one Operation body. I is the
  retained Boolean adapter with ordered Calculations. OG/DOG owns
  `let pattern = value` only as X's direct condition. PO owns a label directly
  on its Iteration child, without an extra block. The initializer,
  condition and step composition is rejected, not lowered. No iterator adapter,
  hidden mutation, borrowing or async wrapper is introduced. See the
  [iteration contract](README.md#iteration-and-definition-contract).
- **Keyed GI:** Every Operation has an explicit corresponding native Calculation
  pattern, preserving category order, the single arm and the final arm. At most
  one explicit `Void` fallback carrier may appear, and it must be last. Rust
  checks pattern legality, binding and exhaustiveness. Logiceme never invents
  equality guards or a wildcard from arm position. Unkeyed GI retains Boolean
  conditions. See [keyed GI](README.md#keyed-gi-pattern-selection).
- **Lifetime/generic lists:** Preserve the declaration/application ownership and
  native-validation boundary in the [local-input contract](README.md#lifetime-and-generic-inputs).

Only template Default cases must stand alone. Qualified Definition fragments are
completed by their enclosing compositions, and definitions/native leaves retain
their own punctuation. Existing category/record interfaces retain these bounded
contracts. The suite checks exact
generation and native compilation. The temporary hello-world gate executes core
Print; its result does not broaden the logeme interfaces.

## Verification contract

Tests must exercise complete AML through `LogicemeService.Evaluate`, returning
the complete application's source/support bundle. For multiple Apps, verify every
book's ordered inventory and artifacts. Segment and interpreter tests supplement,
not replace, that boundary. A source bundle is not proof of compilation/execution.

The [Rust suite](../logiceme/testdata/lucid/README.md) owns the current case
inventory, composed baseline and explicitly scoped extensions. The 18 Rust
segment definitions, interfaces and hash expectations remain frozen. Add source
regressions within task scope without replacing the maintained corpus.

The [Rust reference](README.md#verified-compositions) maps the baseline cases to
the exact composed features they exercise. Preserve those obligations. Native
scaffolding does not establish support for unrelated forms.

Use explicit Rust syntax while preserving the logical obligation. Verify branch
composition, not a handwritten native implementation of the whole conditional.
EA must not replace the composed feature under test. The maintained cases define
a bounded Rust scope, not evidence of complete language, Class, ownership,
concurrency or runtime support. Preserve shared parser/interpreter regressions
for Rust and generic systemless interpretation. Broader coverage requires a separately reviewed follow-up.

For exact Rust goldens, compare every artifact's bytes, path inventory, and
dependency metadata. Do not trim, rustfmt, normalize UUIDs, reorder imports, or
replace actual output with an oracle. Expected output is independently authored.
Distinguish malformed AML from intentional native type/ownership failures.

Use TDD for behavior changes: record an exposing failure using the applicable
existing test, or add a focused regression when that is within the approved task.
Repair the responsible layer and check L/Lyger integration. Never fix shared
Lucid defects by enlarging templates. Compilation success alone does not prove
an exact-byte mismatch is expectation-only. Review literal data, evaluation
order, scopes, and all support files before making that classification.

The [acceptance guide](../docs/development-harness.md) owns commands and deadlines.
Follow its [validation policy](../docs/development-harness.md#validation-authorization):
task-authorized checks proceed without repeated permission requests.
`scripts/l-check` runs the temporary hello-world gate. The complete Lucid
generation and native-compilation suites are separate checks. Record exactly
which stages ran; documentation consistency establishes no native result.
Generated deployment scripts remain artifacts.

## Dependency and runtime boundaries

Compiler-side target preparation and dedicated support templates own manifests
and configuration. Segment templates do not resolve versions,
retry compilation, or supervise processes. The CLI coordinator uses pinned
Rust/Cargo and dependency resolution with frozen builds, and validates workspace
package/path identities before compilation and publication. No automatic version
fallback is implied. Verification retains its reviewed pins and lockfiles.

Rust Execute/Execution exclude Deferred from their inputs, forwarding and
branches. Explicit Deferred metadata remains unsupported even when False. V/MV
render their supplied scheduling expressions without synthesized scope-exit
behavior, RAII guards or cleanup callbacks.
The [Rust execution contract](README.md#execution-without-deferred) owns the
detailed boundary. Removing unsupported markers does not complete other forms.

The current distribution supplies no lifecycle, capture or process-supervision
runtime. The [build contract](../docs/cli-build.md) supports one standalone App
with core Print; scheduling templates do not widen that boundary.

## Review checklist

For each change, identify the owning letters, exact input/child contracts, emitted
role, target preparation needed, and complete-AML witness. Reject unrelated
letter allocation, hidden initializers/mutability/ownership changes, discarded
children, excessive primitive bodies, and golden updates without semantic review.
Mark proposed, implemented, compiled, and behaviorally verified claims separately.
This document does not authorize template changes or restoring deleted libraries.
