# Architecture of `l`

> **Core-only reset:** The project is in a core-only reset. `core.l` loads directly and supplies Print. The authored libraries, Storage, prelude, and native/generated Runtime assets have been removed. The CLI directly compiles one standalone App; no orchestrator or runtime is linked. The detailed architecture below describes the previous library/runtime implementation and is retained as reconstruction context, not a current capability claim.

This map describes the current compiler. [Build and execution semantics](cli-build.md) define the native runtime contract. [The development guide](development-harness.md) defines acceptance and current-input protection.

## Language and compiler responsibilities

`l` describes both applications and the actions of an enclosing program. An application is a language value with structure, compilation data, status, output, and error. Internally it is an Applicative alongside Service, Class, Object, Attribute, Definition, Function, Property, and Operation. [Application structures](../lyger/application_types.go) retain that hierarchy. [Stored values](../lyger/types.go) distinguish its payloads from scalars, collections, and lexical definitions.

The compilation flow is:

```text
program sources + selected distribution + configured grammar/vocabulary
  -> source preparation and imports
  -> tokens, grammatical structures, and semantic bindings
  -> concrete application structures + ordered program actions
  -> required frontend output artifacts and AML
  -> logeme source bundles
  -> native orchestrator, App binaries and App-constructed Service libraries
```

Apps run as separate OS processes while the orchestrator preserves ordered source execution. Containers, cross-compilation and Phase 2 in-code concurrency are outside the supported build command.

| Area | Owns |
| --- | --- |
| `cmd/l` | CLI arguments, initialization, service coordination, tested result view, source validation, Cargo workspace planning, native Rust builds and generation publication |
| `lexicon/eng` | Configured lexical roots, inflections, grammatical structures and cases |
| `lyger` | Source preparation, vocabulary, grammar, bindings, types, ownership/activation analysis and AML composition |
| `logiceme` | AML interpretation, glyph/template evaluation, callable declaration rendering, support assembly and linking |
| `builtins/eng` | Source bootstrap, implicit core definitions and foundational Storage compositions |
| `stdlib/eng` | IO, Formatting and the default Runtime dictionary with their private helpers |
| `stdlib/runtime/rust` and `stdlib/runtime/generated` | Runtime's native host/platform inputs and derived AML/provenance |
| `builtins/operations/rust` | Typed builtin process-operation compositions selected by the frontend |
| `builtins/segments` and `builtins/templates` | Frozen Rust syntax and retained Cargo/configuration/start/stop templates. The maintained Lucid suite defines verified scope |

Logeme is implemented in `logiceme/`. Frontend fields named ASM or AML hold structured markup rather than machine assembly. The backend's `Executable` holds source bundles. The caller owns persistence, building, and execution. The AML and template protocol is described below.

## Production boundaries and retained APIs

Frontend checks use the same checked grammar/lexicon loaders, source preparation,
import classifier and typed Rust planner/emitter as compilation. Tests observe
actual storage, instruction reachability and emitted AML. Generic/systemless
Lucid interpretation and the exercised host-file AML scanner remain distinct
supported contracts, even where their repository callers are tests.

`ApplicationUnits` and `OrchestratorUnit` retain their App-only contracts.
The [workspace pipeline](#native-executable-pipeline) also projects additional
App-constructed Service libraries. Supported metadata/scanner APIs and semantic queries/validators
remain protected by the [current integrity contract](development-harness.md#current-integrity-and-amendments).
An absence of production callers alone does not make a tested contract disposable.

## Embedded builtins and explicit dictionaries

The root distribution package embeds canonical resource files and
exposes them through `fs.FS`. Lyger and logeme use one selected filesystem.
Embedded resources are the default. An explicit complete directory
override selects a complete distribution, including its Runtime implementation. Resource files stay in their native JSON, L
and template formats, and generated programs do not depend on them at runtime.

`builtins/<configured-language>/prelude.l` supplies the implicit builtin unit
and includes `core.l` in that same unit.
The [core reference](reference/words.md) describes its implicit vocabulary.
Prelude privately imports [Storage](internals/storage.md). That implementation namespace
requires an explicit import at a source use. The [dictionary index](../dictionaries.json)
also selects [IO](stdlib/eng/io.md), [Formatting](stdlib/eng/format.md) and Runtime. Adding an
indexed dictionary does not make its words implicit.

The default Runtime standard library composes lifecycle, errors and captures in
[L source](../stdlib/eng/runtime). One complete provider can replace its policy
through the existing registry while preserving the
[compiler/runtime contract](stdlib/eng/runtime.md#compiler-runtime-contract).
Required runtime linkage needs no source import and does not make Runtime's
words implicit. Ordinary imports and user homonyms cannot change that selection.
Its [generated support](../stdlib/runtime/generated) is rendered with the same frozen
logemes. [Native inputs](../stdlib/runtime/rust) retain executor, host, OS and async access.
Formatting has one implementation under `stdlib/<language>/format/`, shared by
IO and Runtime without source-loading process supervision. Storage's same-unit
includes retain one canonical dictionary and lexical ownership. A native
`runtime::` symbol reference inside Storage is a linking dependency, not a source
import of Runtime. Selected source declarations carry their dependencies into
AML without an unconditional library aggregate.

Build, Run, Stop and Await retain their semantic operation identities. The
[private builtin catalog](../lyger/builtin_operations.go) loads their process
implementations, including continuation finishing, from
`builtins/operations/rust/process-*.def` in the same selected filesystem.
Compiler-typed receiver, invocation and continuation bindings instantiate these
logeme compositions after planning. Lexical resolution still allows a source
action named Run to shadow the primitive. It does not replace a catalog entry.
The catalog belongs to one compilation,
validates its complete inputs, and adds no public definition-slot or glyph field.
Start and Execute still accept `any App` through their existing source signatures.


For source authoring, see [composable definitions](guide/definitions.md) and the
[native boundary](guide/native-definitions.md). Those guides own invocation forms,
inflections and examples. This map covers compiler responsibilities.

Definition arguments and orchestrator primitives share native Text/Number
encoding. The [typed-slot contract](spec/definitions.md#typed-slots) owns its byte,
numeric and diagnostic rules.

## Preserve three representations

Keep these concepts distinct even where they share implementation structures:

| Representation | Purpose | Information to retain |
| --- | --- | --- |
| Literal pool | Preserve source text and embedded definitions while parsing surrounding syntax | Exact literal bytes and their source provenance; UUIDs are internal identities |
| Working/semantic memory | Resolve named values, definitions, descriptors, members, and aliases | Value/storage types, lexical classification, qualified bindings, and links to definitions |
| Ordered instructions | Represent what the source says and in what context | Subjects, predicates, clauses, conditions, source order, paragraph identity, and spans |

A single global map cannot by itself represent ordered execution or paragraph-local references. Conversely, a grammar tree without bindings cannot explain what dynamically defined Main, Basic, or Print means. Lowering combines those representations into concrete application structures and the required output view.

## Internal types and ownership

Closed internal domains use checked `iota` enums: instruction/operand kinds, native roles/members, storage/value tags, paragraph scope, supported condition operators, grammar machinery, definition argument interpretation, and fixed parser modes. Their indexed `String()` methods preserve diagnostic names. Text marshaling preserves existing JSON strings and empty snapshots. Unknown integers fail checked operations. There is no generic enum registry or cast between unrelated domains.

Relationships belong to the enum that supplies them. A native role can select an instruction or native member. A member supplies its result type. Readable value types map to storage, planned Rust representations, or supported definition arguments. These are partial mappings. They do not resolve aliases, select a process target, fetch Output, or determine initialization. Configured grammar/case names, lexical POS/type paths, source-defined nominal names, and exported definition argument/slot strings remain open text. Recognizing a concrete value tag does not add it to the narrower accepted source-type vocabulary.

[Stored-value readers](../lyger/types.go) return `(val, error)` and verify exactly one matching outer wrapper plus a nonnil selected application/collection payload. Valid nil Thing and Bytes remain readable. Reads preserve canonical App/Dictionary pointers. Unresolved ID/None/Unknown states can exist during construction. Checked resolved reads do not turn failures into ordinary values. Runtime provenance is independent of storage. [Lexical receivers](../lyger/lexical_types.go) bind per-use arguments, load ordered actions, attach descriptors, and relabel directly. No generic update-map path remains.

The shared [grammar-term decoder](../lyger/grammar_terms.go) serves configuration
validation, lookup, matching and dispatch. [Matching](../lyger/grammar_match.go)
owns candidate copies, case selection, prefix consumption and repeated lookahead;
[nodes](../lyger/grammar_types.go) preserve configured structures.
[Extraction](../lyger/grammar_extraction.go) pairs declaration tokens with records.
[Registration](../lyger/registration.go) follows those nodes for declarations,
aliases, refinements and modifiers, retaining original tokens and separating
independent clauses from definition bodies. It clones refinements and prepares
the complete statement before [atomic publication](../lyger/publication.go).
Rejection publishes nothing. Template discovery follows canonical records to
their own value sites; lowering skips registration-only clauses and retains later
assignments and effects.

Literal restoration uses the shared quote scanner on each original token's
backtick spans once, checking
exact pool keys, skipping quoted identifiers and preserving intervening bytes
and escapes. It validates fragments and assembled Text before committing token
bytes and primitive storage together; grammar nodes share those token pointers.
Inserted content is never rescanned, including during deterministic display
replacement. The import builder owns one flattened accumulator and copies local
literals into its pool. Compilation adopts outputs only after expansion succeeds.
A preparation-local position index avoids repeated source-prefix scans.

## Source meaning and output views

[PrepareSource](../lyger/source.go) splits period-terminated sentences and whitespace-delimited tokens outside quoted literals/comments. It records original byte/line spans, paragraph identity, literal bytes and source order before caching or stamping. The [builder](../lyger/builder.go) loads the implicit builtin unit, canonical dictionary units and importer-relative file inclusions, detects active-stack cycles, and preserves typed source origins alongside the expanded diagnostic bytes.

The builder retains `SourcePaths` as a physical-dependency projection, including local empty, comment-only and import-only files. Embedded logical origins are recorded in the source-unit inventory and never treated as host paths. The working Compilation carries this inventory to the executable coordinator, which applies its canonical path and filesystem identity checks before staging or publication. Repeated imports keep their original occurrence order and expansion behavior. Statement spans continue to supply diagnostic and paragraph provenance.

Pronouns resolve from source order within their paragraph, independently of the order in which unresolved statements become known. Possessives select members of explicit or resolved values. A source paragraph is unrelated to the source-file hierarchy: Classes group files, and Objects own files. AML does not define a Feature/PARAGRAPH layer.

The [semantic resolver](../lyger/semantic_resolution.go) borrows the source-scope
plan and canonical maps, starting at each token's lexical scope. It has no second
memory map or unscoped fallback. Program-root `Compilation.Memory` inherits
builtins; dictionary roots inherit builtins without caller locals. Scopes retain
declaration tokens, and `Subjects` links uses to registered records, including
aliases and qualified declarations.

Each layer first checks the complete qualified query, preserving Public bindings
with private owners. Otherwise its longest visible owner prefix commits lookup
to canonical member tables. Reserved owners/missing members stay pending through
the retry loop's no-progress check. Possessives share that traversal; aliases resolve their
canonical owner while remaining addressable by source name. Shared dictionaries
retain their canonical ID and member namespace. See [source lookup rules](spec/names-and-scope.md#possessives-and-member-lookup).

The [paragraph scope model](../lyger/lexical_scope.go) separates physical source
positions, pronoun paragraphs, expansion occurrences, and lexical environments.
The [lexical contract](spec/lexical-structure.md#whitespace-sentences-and-paragraphs)
defines four-space encapsulation and physical blank separators. Removing a
comment-only line creates no boundary; inline sentences retain byte spans.
Imported declarations inherit their import-site environment, while explicit
paragraphs in an imported occurrence create children. Local declarations reserve
all source-defined forms before retries, so unresolved locals cannot bind outer
names temporarily. Configured vocabulary remains shared and immutable.

`Public` and `Publicly` carry the same configured visibility role as adjective and
adverb respectively. Grammar associates the modifier with exactly one declaration.
Visibility is excluded from value descriptors and names. [Publication](../lyger/publication.go)
validates a complete declaration batch before committing canonical records,
reservations, member entries and root/local indexes. Root memory remains the
program cache. It is not a flattened copy of builtin, dictionary or paragraph
locals. Public targets the declaring unit root, so dictionary-local publication
cannot leak into the program.
Descriptors, their target sets/rules, members, aliases and instantiated nominal
definitions use canonical declaration identity. Descriptor native targets use
the existing value/POS enums. An alias keeps its target identity and a distinct
same-named local definition cannot match an outer descriptor. Private Apps may
share a source name: only their generated artifact IDs
are disambiguated deterministically, in source declaration order. Existing unique
and root/Public App names remain unchanged.

Qualified static members take precedence over native runtime members. Concrete owners use their canonical member tables. Native type namespaces use qualified lookup in the token's lexical environment, including unresolved local reservations. Native reads require an actual App and the configured lexical role (`functional.deflective` or `functional.error`). A bare type namespace permits static definitions but does not create an invocation. Each native use retains its lexical metadata and a fresh runtime reference. Deferred Output Text takes that annotation from its actual RHS operand. Instruction lowering uses the same ownership rules. Local action arguments never fall back to paragraph/global symbols.

The shared [definition-binding implementation](../lyger/definition_bindings.go)
owns structural framing and typed native encoding. The definition reference owns
[framing/escaping](spec/definitions.md#framing-and-escaping) and
[typed slots](spec/definitions.md#typed-slots); [lexical rules](spec/lexical-structure.md#names-possessives-and-qualification)
distinguish L slot/type labels from native identifiers.

Grammar includes value/subject/object roles, phrases, clauses, conditions, predicates, and statements. Its JSON notation expresses optional/repeated items and alternative cases. Go code implements their interpretation. Vocabulary localization remains partial because grammar loading and semantic type names retain English-specific logic. Import operators follow the loaded root-form lexical roles: associative verb, objective preposition, and declarative preposition. Import aliases remove only the original file clause, retaining configured operator spelling, owner punctuation, identifier quoting, and source attribution. Possessive suffix recognition is case-insensitive without changing owner bytes.

The expanded-source view preserves declaration names, descriptor targets, template references, typed parameters, source order and ordinary literal bytes. Program-file inclusions expand at their import positions. Implicit builtin and dictionary units retain their own ownership rather than masquerading as copied program text. A member import retains its local alias statement. Resolution scheduling never reorders the source view. Display rendering removes comments, source stamps and literal-cache identities, strips definition framing and decodes its display escaping once. Ordinary literals retain their escaping. It emits deterministic period/newline separators without a final newline and preserves raw main-source bytes separately. It does not widen Function to Definition, substitute Noun.Process or erase Syllables annotations.

Raw input, originating-source projection, expanded display and the two AML views
remain distinct. The [development guide](development-harness.md#investigating-failures)
maps them to exact CLI test obligations. The returned working result owns the
semantic graph; both AML views are separately allocated renderings of it.

## Application structure and execution

The structural mapping is Program to books, App to Book, Service to Chapter, recursive Class to Text, and Object/source file to Sentence. [Construction nodes](../lyger/construction.go), [typed definition instantiation](../lyger/definition_bindings.go), and [lowering](../lyger/lowering.go) retain source-defined descriptors. [Source hierarchy lowering](../lyger/source_hierarchy.go) traverses each ordered Class/Object child exactly once. [program.go](../lyger/program.go) publishes the complete graph only after both AML views render. Populated Classes select StructureVersion 2 with direct child Dimensions and immediate Class Parent. Ordinary flat construction retains its source order. Rust workspace topology is a separate projection. Class membership never creates a lexical scope or runtime activation. Feature is not a built-in L value type or an AML hierarchy level.

Application defaults are `rust`, version `0.0.0`, directory `/src`, and Host `http://127.0.0.1`. Source declarations determine names and bodies. Descriptors retain typed parameters and reject expansion cycles. A Function's embedded `Func` structure identifies its constructor. Lowering must instantiate that selected definition. Renaming its source word must not select another implementation.

Private application generation stages fresh metadata, settings, services, construction bodies, and books while resolving against original canonical App identities. It renders the complete staged program before publishing owned fields. Failure retains any prior completed graph. Repeated success replaces generated services. Runtime/user fields and canonical pointers are preserved. Ordinary applications and the synthetic orchestrator share small fresh-map default constructors, with their caller-specific settings retained.

The orchestrator is the `l` main method. Run launches an App and Await joins its invocation. Output preserves exact raw bytes.

Standalone Service declarations, including `New Main Service`, remain structural values and produce no workspace packages. Standalone body effects, native lifecycle operations and implicit process Output/Error reject. Main and Basic remain ordinary source-defined construction descriptors used in Apps. App construction preserves its primary Service and additional Service libraries. This introduces no independent Service execution or in-process task lifecycle.

## Frontend diagnostics and results

Binding preserves lexical forms and storage categories, resolves qualified members without global fallbacks, and isolates mutable definition uses. [Source scopes](../lyger/scope.go) bind paragraph pronouns before retry scheduling. [Instructions](../lyger/instructions.go) retain conditional branches, runtime member reads, and ordered Build/Run instructions.

[Statement diagnostics](../lyger/diagnostics.go) report the original file, statement line/column, phase and grammar family/case. Failed alternatives identify the remaining token or end of statement. Diagnostic messages restore cached literals without mutating compiler state and retain their underlying error cause. The first failure stops subsequent phases. Lifecycle and actual source-generation regressions assert that compiler diagnostics do not enter stdout. The shared [error model](error-model.md) classifies returned compiler/build failures with `lerrors.Kind`, while receiver-owned context and ordinary `%w` wrappers retain native causes. Categories do not control parser fallback, dependency readiness, CLI statuses or application execution.

The checked grammar loader validates [configured rule references](../lyger/grammar_validation.go), including optional targets, case selectors, modifiers and dynamic grammar families. All configured entries are checked, including coordinated clauses and Await forms. Grammar and vocabulary each have one canonical protected file under `lexicon/eng`. There are no package-local mirrors. This validation checks reference consistency. Vocabulary subtype availability, general grammar termination and semantic coverage of unused language constructs remain outside its claim.

[CompilationResult](../lyger/result.go) owns the working Compilation, restored source view, observable AML, and canonical backend AML. The CLI's tested lifecycle field stays empty while its returned result owns the real state. [The renderer](../lyger/aml_render.go) emits both AML views from the same graph. The observable view matches AMLTEST exactly and the backend view includes syllable backticks.

[The integration test](../cmd/l/source_integration_test.go) compiles actual demo source through lyger and logeme, checks the actual Rust source structure and literal, and builds main/configuration source in a temporary directory. Changed application/literal and renamed source-vocabulary cases also build. That source-generation test does not execute generated scripts or binaries. Separate executable tests run controlled Apps and orchestrators. The [Service coverage ledger](../testdata/service-execution/coverage.md) distinguishes exact generation, independent expected-Rust compilation and actual execution.

## Native executable pipeline

[Workspace units](../lyger/workspace_units.go) project the source graph: each App owns a binary, its additional Services become libraries, Classes become modules, and Objects remain files. Standalone Services create no package, process or continuation. Raw AML retains independent library/bin targets. Programs without Apps still use the orchestrator with an empty generation manifest.

The [typed plan](../lyger/orchestration_plan.go) resolves storage, captures, definite assignment and transfers before AML generation. Functions bind invocation verbs and compose their selected construction, invocation and Return definitions with private source storage/suspension words. Lyger supplies resolved signature/value atoms. Logiceme renders callable syntax. Generated bodies stay composed, and authored native leaves remain opaque.

A Function stores its retained environment and a typed native function pointer. Its boxed future borrows the program context and returns a typed result tuple or program Exit. Invocation preserves the callable. Arguments/results snapshot ordinary values, share Pointer targets and box concrete values when Thing is required. Function types include canonical constructor identity: incompatible contracts reject, including extraction from erased storage. Aliases preserve that identity. A composed Return must belong to the enclosing Function's constructor.

[Scoped instructions](../lyger/paragraph_instructions.go) retain child paragraphs inside their enclosing Function or App continuation. A nested Function's header executes in its enclosing activation. [Activation frames](../lyger/orchestration_scope.go) provide fresh private storage and retain published bindings. Captured environments survive through their owners. Source-word arguments evaluate once before entering a retained environment. Uninitialized reads, invalid borrows and ownership cycles reject.

[CLI/runtime semantics](cli-build.md) owns process supervision, continuation claiming, exact Output bytes, termination, cleanup bounds and publication. The CLI builds complete workspaces with pinned Rust and frozen offline dependencies, verifies artifacts and publishes companions before the orchestrator. It never runs generated deployment scripts.

Both compiler services accept the same resource filesystem. Explicit providers have no embedded fallback. Lyger's AssetRoot records physical provenance, not resource selection. See [developer defaults](development-harness.md#compiler-defaults) for API options. Host AML reads remain host-file operations. Resource errors preserve phase and cause.

## AML and logeme protocol

`Logiceme.NewLogicemeService(resources).Evaluate(documentBytes)` returns an
`Executable` containing ordered source bytes, paths and dependency metadata.
Each AML book becomes a Cargo project. Building and execution belong to the caller.
`Evaluate` generates Rust and accepts only `IncludeRuntime`. The CLI's separate
workspace bridge follows the [runtime selection contract](cli-build.md#source-generation-and-native-builds).

The [AML guide](../logiceme/AML.md) owns framing, topology, metadata, library
visibility and support assembly. AML is a custom protocol: preserve backticks,
literal bytes and source ownership. Host-file AML reads remain independent of
resource-provider selection. The [logical alphabet](../builtins/LOGEMES.md)
defines the nine letter pairs and composition operators. The [Rust reference](../builtins/README.md)
defines their bounded implementation contracts, including glyph-local lifetime
and generic inputs. The eighteen segment templates remain frozen.

The Lucid scanner recognizes `(/`, optional space/tab/CR/newline, then `={`.
Each subtag requires three complete `={...}=` expressions and closing `/)`.
Nested tags count separately. Original slices preserve literal bytes. Incomplete
recognized tags produce contextual asset errors through `Evaluate`. The
extraction-only API returns no slices. Complete openers remain reserved, without
an additional native quote parser or escape convention.

The [backend architecture](../logiceme/ARCHITECTURE.md) locates scanning,
preparation, interpretation and linking. The [Lucid suite](../logiceme/testdata/lucid/README.md)
records exact-generation and native-compilation obligations plus three bounded
future-execution checks. Its dated results do not establish broader runtime
behavior. Supervised process evidence belongs to the [runtime ledger](../testdata/integration/IMPLEMENTATION.md).
Generated deployment scripts remain unexecuted.

## Runtime test boundary

The generated runtime uses per-runtime dependencies, never mutable global injection. `Runtime::new()` initializes real OS operations and streams. Synchronized invocation snapshots expose immutable result state rather than a live capture buffer. Concurrent internal observers share one result, one OS Wait owner and one continuation claim. This adds no L concurrency syntax.

Native drivers compile the actual generated runtime and host adapters, wrapping real Start/Wait/process-group signaling/capture only at named fault or barrier points. Checkpoints run outside locks held across callbacks: `pre-start` follows allocation and precedes readiness/admission, including refused attempts; `successful-start` follows Start success and secured ownership; `captured-stop-target` follows stable invocation selection and precedes signaling; `pre-publication` follows reaping/capture/group cleanup and precedes final publication. The [test-local receiver contract](../testdata/phase1/runtime/contract.md) defines the private interface. Fixture expectations never substitute for production runtime code.
