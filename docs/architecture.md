# Architecture of `l`

This map describes the current compiler. [Build and execution semantics](cli-build.md) define the native runtime contract; [the development guide](development-harness.md) defines acceptance and current-input protection.

## Language and compiler responsibilities

`l` describes both applications and the actions of an enclosing program. An application is a language value with structure, compilation data, status, output, and error. Internally it is an Applicative alongside Service, Class, Object, Attribute, Definition, Function, Property, and Operation. [Application structures](../lyger/application_types.go) retain that hierarchy; [stored values](../lyger/types.go) distinguish its payloads from scalars, collections, and lexical definitions.

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
| `builtins/eng` and `stdlib/eng` | English L source compositions, small native leaves and construction descriptors |
| `builtins/operations/rust` | Typed builtin process-operation compositions selected by the frontend |
| `builtins/segments` and `builtins/templates` | Frozen Rust templates with reviewed composition amendments and retained Cargo/configuration/runtime support assets; the maintained Lucid suite defines verified scope |

Logeme is implemented in `logiceme/`. Frontend fields named ASM or AML hold structured markup rather than machine assembly. The backend's `Executable` holds source bundles; the caller owns persistence, building, and execution. The AML and template protocol is described below.

## Production boundaries and retained APIs

Frontend checks use the same checked grammar/lexicon loaders, source preparation,
import classifier and typed Rust planner/emitter as compilation. Tests observe
actual storage, instruction reachability and emitted AML. Generic/systemless
Lucid interpretation and the exercised host-file AML scanner remain distinct
supported contracts, even where their repository callers are tests.

`ApplicationUnits` and `OrchestratorUnit` retain their App-only contracts;
the [workspace pipeline](#native-executable-pipeline) also projects additional
App-constructed Service libraries. Supported metadata/scanner APIs and semantic queries/validators
remain protected by the [current integrity contract](development-harness.md#current-integrity-and-amendments);
an absence of production callers alone does not make a tested contract disposable.

## Embedded builtins and explicit dictionaries

The root distribution package embeds canonical resource files and
exposes them through `fs.FS`. Lyger and logeme use one selected filesystem.
Embedded resources are the default; an explicit complete directory
override is useful for development. Resource files stay in their native JSON, L
and template formats, and generated programs do not depend on them at runtime.

`builtins/<configured-language>/core.l` supplies the implicit builtin unit.
The [core](../builtins/eng/core.l) supplies foundational words and imports the
[Storage dictionary](../stdlib/eng/storage/dictionary.l) privately. Programs and
dictionaries inherit core words; they do not inherit Storage's implementation
namespace. Ordinary source Functions, locals, descriptors and invocation words
compose storage behavior. Selected declarations carry only their required source
dependencies into AML.

The [dictionary index](../stdlib/index.json) also selects IO and private support
dictionaries. IO.Print composes typed rendering with byte output. IO and Runtime
select the same [Formatting Function](../builtins/runtime/eng/number.l). Runtime lifecycle, error and capture policy lives in
[builtins/runtime/eng](../builtins/runtime/eng); its generated, fingerprinted AML
is rendered by Logiceme with the same frozen logemes. Host adapters retain OS and
async access. Adding an indexed dictionary never makes its words implicit.

Build, Run, Stop and Await retain their semantic operation identities. The
[private builtin catalog](../lyger/builtin_operations.go) loads their process
implementations, including continuation finishing, from
`builtins/operations/rust/process-*.def` in the same selected filesystem.
Compiler-typed receiver, invocation and continuation bindings instantiate these
logeme compositions after planning. Lexical resolution still allows a source
action named Run to shadow the primitive; it does not replace a catalog entry.
The catalog belongs to one compilation,
validates its complete inputs, and adds no public definition-slot or glyph field.
Start and Execute still accept `any App` through their existing source signatures.


The [basic demo](../demo/basic/main.l) imports IO.Print, creates a Basic App, executes it, and shows its raw output. Tests execute that source directly. Controlled process fixtures use explicit test-owned companion definitions without copying the language foundation. Basic construction and backend generation retain their existing semantic assertions.

Definitions can embed logeme expressions in text marked with `<DEF>`. Placeholders such as `<@Output|Thing@>` carry substitution information. Labels such as `Log|-s|-ged|-ging|-ger` define related grammatical forms. Resolving one of those words from memory must preserve its grammatical identity, not replace it with the storage label `lexeme`.

Number arguments in definitions and orchestrator primitives share one finite
binary64 encoder. It emits the shortest round-trip decimal representation with
an explicit `_f64` suffix, preserving negative zero and the parsed value's
rounding. Definition binding encodes the parsed value rather than copying its
spelling: integral, leading-plus and hexadecimal floating-point inputs therefore
reach Rust as valid floating-point expressions. Number arguments supplied to
Thing use the same rule; native fragments and native integers remain opaque.
Malformed or overflowing definition arguments retain their contextual Syntax
errors and native parse causes; stored nonfinite primitives remain Unsupported.

## Preserve three representations

Keep these concepts distinct even where they share implementation structures:

| Representation | Purpose | Information to retain |
| --- | --- | --- |
| Literal pool | Preserve source text and embedded definitions while parsing surrounding syntax | Exact literal bytes and their source provenance; UUIDs are internal identities |
| Working/semantic memory | Resolve named values, definitions, descriptors, members, and aliases | Value/storage types, lexical classification, qualified bindings, and links to definitions |
| Ordered instructions | Represent what the source says and in what context | Subjects, predicates, clauses, conditions, source order, paragraph identity, and spans |

A single global map cannot by itself represent ordered execution or paragraph-local references. Conversely, a grammar tree without bindings cannot explain what dynamically defined Main, Basic, or Print means. Lowering combines those representations into concrete application structures and the required output view.

## Internal types and ownership

Closed internal domains use checked `iota` enums: instruction/operand kinds, native roles/members, storage/value tags, paragraph scope, supported condition operators, grammar machinery, definition argument interpretation, and fixed parser modes. Their indexed `String()` methods preserve diagnostic names; text marshaling preserves existing JSON strings and empty snapshots. Unknown integers fail checked operations. There is no generic enum registry or cast between unrelated domains.

Relationships belong to the enum that supplies them. A native role can select an instruction or native member; a member supplies its result type; readable value types map to storage, planned Rust representations, or supported definition arguments. These are partial mappings. They do not resolve aliases, select a process target, fetch Output, or determine initialization. Configured grammar/case names, lexical POS/type paths, source-defined nominal names, and exported definition argument/slot strings remain open text. Recognizing a concrete value tag does not add it to the narrower accepted source-type vocabulary.

[Stored-value readers](../lyger/types.go) return `(val, error)` and verify exactly one matching outer wrapper plus a nonnil selected application/collection payload. Valid nil Thing and Bytes remain readable. Reads preserve canonical App/Dictionary pointers. Unresolved ID/None/Unknown states can exist during construction; checked resolved reads do not turn failures into ordinary values. Runtime provenance is independent of storage. [Lexical receivers](../lyger/lexical_types.go) bind per-use arguments, load ordered actions, attach descriptors, and relabel directly; no generic update-map path remains.

The [grammar-term decoder](../lyger/grammar_terms.go) is shared by configuration validation, lookup, matching, and node dispatch. [Grammar matching](../lyger/grammar_match.go) owns candidate copies, case selection, prefix consumption, and repeated-item lookahead. [Node types/building](../lyger/grammar_types.go) preserve configured structures; [extraction](../lyger/grammar_extraction.go) returns declaration-token/record pairs. The borrowed [registration sites](../lyger/registration.go) follow those grammar nodes for declarations, aliases, descriptor refinements, and modifier ownership. They retain original tokens and separate independent clauses from definition bodies. Registration prepares every declaration and descriptor update before the [atomic publisher](../lyger/publication.go) validates the complete statement and commits its relationships; a rejected batch publishes nothing. Descriptor refinements clone existing targets/rules, and template discovery follows each canonical record back to its own value site. Instruction lowering skips registration-only clauses while preserving following assignments and effects.

Literal restoration visits each original statement token once and checks exact pool keys in its original backtick-delimited spans using the shared quote scanner. A token may contain several such spans. It skips quoted identifiers, preserves intervening bytes and escapes, and never rescans inserted content. Each cached fragment and the assembled Text are checked before committing token bytes and primitive storage together; grammar nodes share these token pointers. Display rendering uses one deterministic replacement pass, so inserted literal bytes cannot become another cache lookup. The import builder owns one flattened accumulator and copies each prepared source's local literals into its own pool. Compilation adopts complete outputs only after expansion succeeds. A preparation-local position index avoids rescanning source prefixes for every span.

## Source meaning and output views

[PrepareSource](../lyger/source.go) splits period-terminated sentences and whitespace-delimited tokens outside quoted literals/comments. It records original byte/line spans, paragraph identity, literal bytes and source order before caching or stamping. The [builder](../lyger/builder.go) loads the implicit builtin unit, canonical dictionary units and importer-relative file inclusions, detects active-stack cycles, and preserves typed source origins alongside the expanded diagnostic bytes.

The builder retains `SourcePaths` as a physical-dependency projection, including local empty, comment-only and import-only files. Embedded logical origins are recorded in the source-unit inventory and never treated as host paths. The working Compilation carries this inventory to the executable coordinator, which applies its canonical path and filesystem identity checks before staging or publication. Repeated imports keep their original occurrence order and expansion behavior; statement spans continue to supply diagnostic and paragraph provenance.

Pronouns resolve from source order within their paragraph, independently of the order in which unresolved statements become known. Possessives select members of explicit or resolved values. A source paragraph is unrelated to the source-file hierarchy: Classes group files, and Objects own files. AML does not define a Feature/PARAGRAPH layer.

The [semantic resolver](../lyger/semantic_resolution.go) starts lookup at each token's lexical scope and borrows the source-scope plan and canonical semantic maps. The program root scope owns `Compilation.Memory` and inherits builtin declarations; dictionary unit roots inherit builtins without seeing caller locals; the resolver has no second memory map or fallback for unscoped tokens. Source scopes retain declaration tokens; semantic `Subjects` connects each token to its registered record, including aliases and qualified declarations. Each lexical layer checks the complete qualified query first, preserving explicitly Public qualified bindings with private owners. Otherwise the longest locally visible owner prefix commits the remainder to canonical member tables; a reserved owner or missing member stays pending instead of reopening outer lookup. Possessive static lookup uses the same member traversal. Checked alias traversal selects a canonical owner before member lookup. Alias records remain addressable under their source names, while a shared dictionary retains its canonical ID and member namespace. Missing dependencies wait for the existing no-progress-checked retry loop.

The [paragraph scope model](../lyger/lexical_scope.go) separates physical source
positions, pronoun paragraphs, expansion occurrences, and lexical environments.
A four-space opening prefix creates an encapsulation; continuation lines remain
inside until a physical blank separator outside comments/literals. Removing a comment-only line does not create a paragraph boundary. Inline sentences retain byte spans.
Imported declarations inherit their import-site environment, while explicit
paragraphs in an imported occurrence create children. Local declarations reserve
all source-defined forms before retries, so unresolved locals cannot bind outer
names temporarily. Configured vocabulary remains shared and immutable.

`Public` and `Publicly` carry the same configured visibility role as adjective and
adverb respectively. Grammar associates the modifier with exactly one declaration;
visibility is excluded from value descriptors and names. [Publication](../lyger/publication.go)
validates a complete declaration batch before committing canonical records,
reservations, member entries and root/local indexes. Root memory remains the
program cache; it is not a flattened copy of builtin, dictionary or paragraph
locals. Public targets the declaring unit root, so dictionary-local publication
cannot leak into the program.
Descriptors, their target sets/rules, members, aliases and instantiated nominal
definitions use canonical declaration identity. Descriptor native targets use
the existing value/POS enums; an alias keeps its target identity and a distinct
same-named local definition cannot match an outer descriptor. Private Apps may
share a source name: only their generated artifact IDs
are disambiguated deterministically, in source declaration order. Existing unique
and root/Public App names remain unchanged.

Qualified static members take precedence over native runtime members. Concrete owners use their canonical member tables; native type namespaces use qualified lookup in the token's lexical environment, including unresolved local reservations. Native reads require an actual App and the configured lexical role (`functional.deflective` or `functional.error`); a bare type namespace permits static definitions but does not create an invocation. Each native use retains its lexical metadata and a fresh runtime reference. Deferred Output Text takes that annotation from its actual RHS operand. Instruction lowering uses the same ownership rules; local action arguments never fall back to paragraph/global symbols.

Definition framing is structural: one complete outer marker pair, optionally inside complete backticks, surrounds an encoded body. The shared [framing helper](../lyger/definition_bindings.go) does not decode or validate template glyphs. Source Text requires the quoted form to classify as a definition; ordinary Text with interior/unmatched markers remains exact. Expanded display and explicit template parsing decode once at their respective boundaries, while lexeme storage preserves encoding and quote presence. Explicit parsing still accepts markerless internal templates and rejects malformed/empty/nested definitions. When a Text scalar becomes a Rust expression, shared quoting protects backend delimiters, metadata separators, repeated spaces, and Lucid openers with Rust escapes. Definition arguments (including Text in Thing slots) and orchestration use that boundary; literal storage, source rendering, raw definitions, and completed AML are not rewritten. Orchestration retains its additional glyph escapes. L slot labels and type labels accept dotted identifiers whose components start with a Unicode letter or underscore and continue with Unicode letters, digits or underscores. Native language keywords do not restrict these L labels; Rust identifiers embedded in native definitions use their separate native validation boundary.

Grammar includes value/subject/object roles, phrases, clauses, conditions, predicates, and statements. Its JSON notation expresses optional/repeated items and alternative cases; Go code implements their interpretation. Vocabulary localization remains partial because grammar loading and semantic type names retain English-specific logic. Import operators follow the loaded root-form lexical roles: associative verb, objective preposition, and declarative preposition. Import aliases remove only the original file clause, retaining configured operator spelling, owner punctuation, identifier quoting, and source attribution. Possessive suffix recognition is case-insensitive without changing owner bytes.

The expanded-source view preserves declaration names, descriptor targets, template references, typed parameters, source order and ordinary literal bytes. Program-file inclusions expand at their import positions; implicit builtin and dictionary units retain their own ownership rather than masquerading as copied program text. A member import retains its local alias statement. Resolution scheduling never reorders the source view. Display rendering removes comments, source stamps and literal-cache identities, strips definition framing and decodes its display escaping once; ordinary literals retain their escaping. It emits deterministic period/newline separators without a final newline and preserves raw main-source bytes separately. It does not widen Function to Definition, substitute Noun.Process or erase Syllables annotations.

The current CLI tests distinguish five views:

| View | Current obligation |
| --- | --- |
| `mainBytes` / INITTEST | Exact raw main-source bytes |
| COMPTEST | Originating-user-source projection: trim each main-path statement's Raw bytes and join in order; comments remain. The current constant equals INITTEST. |
| `compiledBytes` / `CompilationResult.ExpandedSource` | Full expanded display, including resolved source units; it is not COMPTEST |
| `asmBytes` / `CompilationResult.AML` | Exact observable application AML required by AMLTEST |
| `CompilationResult.BackendAML` | Canonical framed generator input, including syllable backticks |

`TestCompilerInitialize` compares the complete compiler struct, including nil/empty state. `TestCompilerCompile` checks raw source, exact AML, unchanged initialized compilation/options, equality to the returned expanded view, and originating-source preservation. The returned working result owns the actual semantic graph. Both AML views are separately allocated renderings of that graph. Preserve exact expected bytes and nil-versus-empty distinctions; never copy fixture constants into compilation or discard the graph to satisfy a tested view.

## Application structure and execution

The structural mapping is Program to books, App to Book, Service to Chapter, recursive Class to Text, and Object/source file to Sentence. [Construction nodes](../lyger/construction.go), [typed definition instantiation](../lyger/definition_bindings.go), and [lowering](../lyger/lowering.go) retain source-defined descriptors. [Source hierarchy lowering](../lyger/source_hierarchy.go) traverses each ordered Class/Object child exactly once; [program.go](../lyger/program.go) publishes the complete graph only after both AML views render. Populated Classes select StructureVersion 2 with direct child Dimensions and immediate Class Parent. Ordinary flat construction retains its source order; Rust workspace topology is a separate projection. Class membership never creates a lexical scope or runtime activation. Feature is not a built-in L value type or an AML hierarchy level.

Application defaults are `rust`, version `0.0.0`, directory `/src`, and Host `http://127.0.0.1`. Source declarations determine names and bodies. Descriptors retain typed parameters and reject expansion cycles. A Function's embedded `Func` structure identifies its constructor; lowering must instantiate that selected definition. Renaming its source word must not select another implementation.

Private application generation stages fresh metadata, settings, services, construction bodies, and books while resolving against original canonical App identities. It renders the complete staged program before publishing owned fields. Failure retains any prior completed graph; repeated success replaces generated services. Runtime/user fields and canonical pointers are preserved. Ordinary applications and the synthetic orchestrator share small fresh-map default constructors, with their caller-specific settings retained.

The orchestrator is the `l` main method. Run launches an App and Await joins its invocation. Output preserves exact raw bytes.

Standalone Service declarations, including `New Main Service`, remain structural values and produce no workspace packages. Standalone body effects, native lifecycle operations and implicit process Output/Error reject. Main and Basic remain ordinary source-defined construction descriptors used in Apps. App construction preserves its primary Service and additional Service libraries; this introduces no independent Service execution or in-process task lifecycle.

## Frontend diagnostics and results

Binding preserves lexical forms and storage categories, resolves qualified members without global fallbacks, and isolates mutable definition uses. [Source scopes](../lyger/scope.go) bind paragraph pronouns before retry scheduling; [instructions](../lyger/instructions.go) retain conditional branches, runtime member reads, and ordered Build/Run instructions.

[Statement diagnostics](../lyger/diagnostics.go) report the original file, statement line/column, phase and grammar family/case. Failed alternatives identify the remaining token or end of statement. Diagnostic messages restore cached literals without mutating compiler state and retain their underlying error cause; the first failure stops subsequent phases. Lifecycle and actual source-generation regressions assert that compiler diagnostics do not enter stdout. The shared [error model](error-model.md) classifies returned compiler/build failures with `lerrors.Kind`, while receiver-owned context and ordinary `%w` wrappers retain native causes. Categories do not control parser fallback, dependency readiness, CLI statuses or application execution.

The checked grammar loader validates [configured rule references](../lyger/grammar_validation.go), including optional targets, case selectors, modifiers and dynamic grammar families. All configured entries are checked, including coordinated clauses and Await forms. Grammar and vocabulary each have one canonical protected file under `lexicon/eng`; there are no package-local mirrors. This validation checks reference consistency; vocabulary subtype availability, general grammar termination and semantic coverage of unused language constructs remain outside its claim.

[CompilationResult](../lyger/result.go) owns the working Compilation, restored source view, observable AML, and canonical backend AML. The CLI's tested lifecycle field stays empty while its returned result owns the real state. [The renderer](../lyger/aml_render.go) emits both AML views from the same graph; the observable view matches AMLTEST exactly and the backend view includes syllable backticks.

[The integration test](../cmd/l/source_integration_test.go) compiles actual demo source through lyger and logeme, checks the actual Rust source structure and literal, and builds main/configuration source in a temporary directory. Changed application/literal and renamed source-vocabulary cases also build. That source-generation test does not execute generated scripts or binaries. Separate executable tests run controlled Apps and orchestrators; the [Service coverage ledger](../testdata/service-execution/coverage.md) distinguishes exact generation, independent expected-Rust compilation and actual execution.

## Native executable pipeline

[Application units](../lyger/application_units.go) preserve the source graph: each App owns a binary, its additional Services become libraries, Classes become modules, and Objects remain files. Standalone Services create no package, process or continuation. Raw AML retains independent library/bin targets. Programs without Apps still use the orchestrator with an empty generation manifest.

The [typed plan](../lyger/orchestration_plan.go) resolves storage, captures, definite assignment and transfers before AML generation. Functions bind invocation verbs and compose their selected construction, invocation and Return definitions with private source storage/suspension words. Lyger supplies resolved signature/value atoms; Logiceme renders callable syntax. Generated bodies stay composed, and authored native leaves remain opaque.

A Function stores its retained environment and a typed native function pointer. Its boxed future borrows the program context and returns a typed result tuple or program Exit. Invocation preserves the callable; arguments/results snapshot ordinary values, share Pointer targets and box concrete values when Thing is required. Function types include canonical constructor identity: incompatible contracts reject, including extraction from erased storage. Aliases preserve that identity. A composed Return must belong to the enclosing Function's constructor.

[Scoped instructions](../lyger/paragraph_instructions.go) retain child paragraphs inside their enclosing Function or App continuation. A nested Function's header executes in its enclosing activation. [Activation frames](../lyger/orchestration_scope.go) provide fresh private storage and retain published bindings; captured environments survive through their owners. Source-word arguments evaluate once before entering a retained environment. Uninitialized reads, invalid borrows and ownership cycles reject.

[CLI/runtime semantics](cli-build.md) owns process supervision, continuation claiming, exact Output bytes, termination, cleanup bounds and publication. The CLI builds complete workspaces with pinned Rust and frozen offline dependencies, verifies artifacts and publishes companions before the orchestrator. It never runs generated deployment scripts.

Both compiler services accept the same resource filesystem; explicit providers have no embedded fallback. Lyger's AssetRoot records physical provenance, not resource selection. See [developer defaults](development-harness.md#compiler-defaults) for API options. Host AML reads remain host-file operations; resource errors preserve phase and cause.

## AML and logeme protocol

`Logiceme.NewLogicemeService(resources).Evaluate(documentBytes)` returns an `Executable` with `ByteArrays`, `MapArrays`, `DependencyArrays` and `EnvDepArrays`. Evaluation scans tagged input, compiles document structure, assembles glyph/template output, then links source files and imports. Every AML book becomes a separate Cargo project; the caller owns building and execution.

`Evaluate` generates Rust and accepts only the `IncludeRuntime` option. L, Lyger and Logiceme remain Go implementations. This is an AML source-generation API; the frontend/CLI separately owns building and execution. The 18 Rust segments remain frozen after reviewed amendments; the [Rust reference](../builtins/README.md) distinguishes their implemented composition contracts from retained but incomplete integration.

The Rust renderer separates initialization records from executable cases, dispatches parsed primitive calls, prepares existing Func/Prop/Op records and forwards ordered category collections through explicit aliases. A renders its package comment and collected `use` imports. Standalone EA retains exact file contents. The [Lucid suite](../logiceme/testdata/lucid/README.md) preserves exact composition scenarios covering scoped selection, iteration, declaration, invocation, callable, trait, empty-composition, keyed-pattern and generic/lifetime witnesses. It compares every returned artifact exactly and compiles both expected and actual projects; its [dated ledger](../logiceme/testdata/lucid/FAILURES.md) records counts and results separately from runtime claims. Three future-composition cases also execute their generated programs; broader runtime behavior belongs to supervised frontend/runtime tests. Generated deployment scripts remain unexecuted.

Rust supports optional glyph-local `Lifetimes` and `GenericTypes` String inputs through existing cases: FW/HY/MY declare parameters, while Y and named KY apply arguments. Templates own the conditional angle-bracket or turbofish punctuation. Callers supply native list interiors, and Rust checks their legality. These inputs are not inherited Class/Object defaults or fields inside Func/Prop/Op records; see the [input contract](../builtins/README.md#lifetime-and-generic-inputs).

Version-2 recursive Classes, Object roots, module descriptors and Service crate targets are integrated through compiler-owned structural header and workspace assembly. A retains its frozen package-comment/import convention; the compiler supplies module, visibility and startup boundaries without adding template interfaces. Library-only bundles omit executable support and reject executable configuration/runtime options. Focused tests compile actual generated libraries against independent consumers. See [current integration limits](../builtins/README.md#current-integration-limits). Leftover advanced preparation/validation entries are not the frozen template inventory, and neither their presence nor old matrix results certify a feature.

| Tagged unit | Meaning |
| --- | --- |
| Book | Application |
| Chapter | Service |
| Text | Recursive Class; source-file membership, not runtime scope |
| Sentence | Object/source file |
| Word | Collection of syllables |
| Syllable | Composed glyph expression |

The `PROGRAM`, `STRUCTURE`, `OUTLINE`, `JSON`, `SETTINGS` and `CONTENT` records must agree on hierarchy and output locations. This is a custom tagged format, not general Markdown or XML. JSON supplies application/component metadata; settings carry key/value/type records. Syllables require literal backticks, for example:

```text
<|W|>
<|S|>`A`</S|>
<|S|>`FW{|Key: ~main~|}^W^BOD^E{|Statement: ~::std::println!("Hello World!");~|}`</S|>
</W|>
```

That fragment still needs the surrounding book, outline, service and sentence. `{|...|}` contains metadata; `^` composes glyphs, parentheses group expressions and `*` separates grouped siblings. Func, Prop and Op encode structured signatures, properties and operators.

| Content category | Action |
| --- | --- |
| A Declaration | B Declare |
| E Calculation | D Calculate |
| I Evaluation | G Evaluate |
| O Definition | F Define |
| U Description | H Describe |
| V Execution | C Execute |
| W Operation | K Operate |
| X Iteration | P Iterate |
| Y Instantiation | M Instantiate |

A fixed typed backend table registers the 18 segment templates, including Execute/Execution; their presence in backend templates does not add L concurrency syntax. `EA` supports literal file contents. Child glyphs compile first, and parent templates consume their output by category. Linking gathers dependency/import metadata; recursive source composition is the purpose of the backend's scalar/vector/matrix terminology.

Lucid uses conditions, repeated regions and substitutions:

```text
{|/|-{tags: Meta.Tags}--{if Type == "Default"}--{template body}-|/|}
(/={tags: Meta.Tags}=={for Operation in Conjoinments.Operations}=={[|Operation|]}=/)
[|Meta.Key|]
[|Conjoinments.Operation|]
```

The shared scanner in `logiceme/lucid.go` recognizes a subtag at `(/`, optional space/tab/CR/newline, then `={`. Each tag needs three complete `={...}=` expressions and closing `/)`; nested tags count their own expressions. Original slices preserve Unicode, whitespace and ordinary delimiter fragments. Incomplete recognized tags return contextual asset errors through Evaluate. The extraction-only API retains its slice-only signature and returns no slices on a scanner error. Complete openers remain reserved; there is no additional native quote parser or escape convention.

Backend linking can add Rust configuration modules, initialization calls, Cargo manifests and setup/start/stop scripts. These scripts are generated artifacts; native build and acceptance never execute them. The typed generation option adds runtime support only for the synthetic orchestrator. Public Evaluate retains default source-bundle behavior. Host-file AML reads remain host-file operations, independently of embedded resource loading.

## Runtime test boundary

The generated runtime uses per-runtime dependencies, never mutable global injection. `Runtime::new()` initializes real OS operations and streams. Synchronized invocation snapshots expose immutable result state rather than a live capture buffer. Concurrent internal observers share one result, one OS Wait owner and one continuation claim; this adds no L concurrency syntax.

Native drivers compile the actual generated runtime and host adapters, wrapping real Start/Wait/process-group signaling/capture only at named fault or barrier points. Checkpoints run outside locks held across callbacks: `pre-start` follows allocation and precedes readiness/admission, including refused attempts; `successful-start` follows Start success and secured ownership; `captured-stop-target` follows stable invocation selection and precedes signaling; `pre-publication` follows reaping/capture/group cleanup and precedes final publication. The [test-local receiver contract](../testdata/phase1/runtime/contract.md) defines the private interface. Fixture expectations never substitute for production runtime code.
