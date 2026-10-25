# Logeme assembler architecture

This is the navigation map and intended separation of concerns for the Go-written
assembler. [LOGEMES.md](../builtins/LOGEMES.md) defines the logical alphabet;
[CONTRACTS.md](../builtins/CONTRACTS.md) defines invariants; [AML.md](AML.md)
defines the input boundary. This map distinguishes existing entry points from
required architecture: a template route is not evidence that every contract is met.
The frontend, CLI and raw API generate Rust. Runtime support is the only
generation option; no output-target selector remains. See the [implementation record](../testdata/integration/IMPLEMENTATION.md)
for dated generation, native execution and acceptance evidence.

## Position in L

```text
L source / developer definitions
             |
             v
Lyger: parsing, bindings, scopes, semantic instructions, AML
             |
             v
LogicemeService.Evaluate(complete AML, generation options)
             |
             selected fs.FS
             |
             v
framed tag tree + outline/file structure + scoped metadata
             |
             v
syllable/operator composition -> compiled Lucid template blocks
             |
             v
metadata interpretation -> ordered source fragments
             |
             v
file headers + linking + support assembly
             |
             v
Executable: ordered application/file bytes, paths, dependency metadata
             |
             v
caller/build coordinator: materialize, validate, compile, publish, run
```

Source assembly is the product of Logiceme, not native compilation or execution.
The compiler implementation remains Go even when the generated project is Rust.
Each App remains a separate native process; tasks inside a generated app do not
replace L's App lifecycle.

## Code map

| Area | Responsibility and entry points |
| --- | --- |
| [logiceme_service.go](logiceme_service.go) | Public Evaluate/options validation and evaluation coordination |
| [scanner.go](scanner.go), [book.go](book.go) | Framed AML, tag hierarchy, books/words/syllables and associated metadata |
| [segments.go](segments.go), [segmentals.go](segmentals.go) | Resource loading, glyph identities, determinant/coda selection, inputs and conjoinment compilation |
| [meta.go](meta.go) | Existing metadata representations and scope operations |
| [lucid.go](lucid.go) | Template framing and parsed literal/control structure |
| [lucid_interpreter.go](lucid_interpreter.go) | Ordered interpretation, substitutions, loops and conditions |
| [compiler.go](compiler.go), [bundle_support.go](bundle_support.go) | Compilation/assembly/linking orchestration and Executable construction |
| [glyph_roles.go](glyph_roles.go), [glyph_metadata_validation.go](glyph_metadata_validation.go) | Composition, context and metadata validation |
| [inline_fragments.go](inline_fragments.go) | Native-fragment boundary handling |
| [native_framing.go](native_framing.go), [source_scan.go](source_scan.go) | Native framed payloads and ordered generation-document scanning |
| [source_file.go](source_file.go) | Object compilation/assembly and the limited literal-file path |
| [support_templates.go](support_templates.go), [bundle_support.go](bundle_support.go) | Support loading/rendering, linking, path validation and dependencies |
| [workspace_runtime.go](workspace_runtime.go) | Separate workspace entry and runtime-library bundles |
| [glyph_render.go](glyph_render.go) | Rust executable-case selection, primitive dispatch and explicit child forwarding |
| [glyph_prepare.go](glyph_prepare.go), [native_metadata.go](native_metadata.go), [glyph_metadata_validation.go](glyph_metadata_validation.go), [glyph_roles.go](glyph_roles.go) | Rust record preparation, local input validation and role checks; additional branches remain beyond the bounded Lucid scope |
| [source_graph.go](source_graph.go), [source_headers.go](source_headers.go) | Recursive source ownership, explicit binary/library targets, module declarations and structural header assembly separate from frozen A |
| [builtins](../builtins) | Target segment primitives, vowel forwarding and dedicated support templates |
| [distribution.go](../distribution.go) | Embedded asset filesystem |
| [cmd/l](../cmd/l), [lyger](../lyger) | Native build/publication and frontend semantics; not responsibilities of segment templates |

These paths identify ownership, not a requirement to introduce a parallel generic
compiler IR or duplicate the implementation per target.

## Public boundary and interpretation isolation

`NewLogicemeService(resources fs.FS)` selects one complete asset provider; nil
uses embedded resources. A private typed table registers the 18 glyphs with fresh
metadata per evaluation. `Evaluate(input []byte, options ...GenerationOptions)`
accepts at most one options value containing only `IncludeRuntime`. Invalid
option counts return errors and a zero-value Executable. Natural-language
selection belongs to `Lyger.Options`.

An Executable has ByteArrays, MapArrays, DependencyArrays and EnvDepArrays indexed
by application; each byte array aligns with its path. Preserve order, bytes and
independent ownership of results. Reusing a service must not mutate a previous
result or leak metadata. Actual native compilation is a separate boundary.

Current integration detail: Rust assembles every book in a document in order.
The coordinator consumes explicit application/library units and builds a Cargo
workspace. All generated entries and companion applications are native Rust.

## Frontend lowering boundary

Lyger's typed Rust plan operates on resolved semantic instructions before AML
emission. It records activation-local storage, definite assignment, mutation,
effects and value provenance. Repeated paragraph activations preserve separate
private slots; canonical App aliases and published snapshots keep their identity.
Runtime-dependent presence uses explicit bookkeeping rather than a fabricated L
value. Native leaves remain opaque; the plan adds no new generic/mutability Meta
fields, sections, per-paragraph tasks or general ownership inference.

The emitter composes existing FW/W/GI and declaration/definition forms, with
native atoms for supported runtime calls and explicit typed storage punctuation.
It preserves raw bytes and finite scalar formatting at the typed boundary.
Source topology and compiler support remain separate from that execution plan.

## Layer contracts

### Scan and preserve structure

Read the selected resources and frame the AML before interpreting its contents.
Retain the tag tree, outline ownership, source positions and native literal data.
Recognize syntax from parsed structure, not substring matches inside source text.
Do not treat markup-like native bytes as application boundaries or comments.

### Compose templates

Select the vowel's determinant case and any qualified consonant case, gather
their declared inputs, and integrate child templates by logical category.
The operators preserve grouping and sibling order. Compilation composes template
blocks; it does not execute generated application logic.

Keep native roles in addition to broad logical categories: a Y child may be a
type, constructed value, or invocation. A generic adapter cannot erase that
distinction. Validation should report missing/conflicting children, unsupported
forms and invalid placement at the responsible source composition.

The Rust renderer selects an executable case independently of `Inputs` and
`Headers` initialization records. It dispatches a framed primitive call by its
parsed kind and preserves surrounding template whitespace. Forwarded children
retain their logical category, singular/plural form and order, but are visible
to the primitive only under their explicit argument names. The existing
`Conjoinment` alias mechanism supplies values such as Definition initializers;
private preparation projections are not additional template categories.

Rust iteration preparation validates one body and mutually exclusive iterator,
Calculation, I/Evaluation or OG/DOG condition inputs before rendering. It does
not interpret a Calculation as a step or a Definition as initialization. OG/DOG
retain native pattern/value metadata through the existing Definition interface;
PO forwards its one Iteration child directly, so its label targets the loop rather
than a generated wrapper block. Three-part composition is rejected, not
expanded. Caller-owned iterator expressions and explicit await remain native
inputs; no helper loop or iterator adaptation is prepared.

### Interpret metadata and control regions

Interpret composed templates with scoped inputs and conjoinments. Preserve each
body as ordered literal spans and direct child occurrences. Every loop iteration
renders the entire body under its binding, then restores the previous binding.
Do not globally replace identical control-tag strings or rescan inserted text.
Whitespace belongs to the templates and literal source, not cleanup heuristics.

Keep composition validation and record preparation separate from template-owned
punctuation. Existing Func/Prop/Op records retain their order and native-source
provenance, including supplied callable names and output types spelled `Void`.
Rust parameter formatting and ordered Ops use the existing decoders; Prefix/Postfix
framing is removed exactly once. Native type, lifetime, pattern and ownership
legality remains with Rust. Neither iteration nor generic preparation invents
borrowing, cloning, parameter names or iterator adapters.

The optional Rust `Lifetimes` and `GenericTypes` inputs remain glyph-local String
metadata beside existing records. FW/HY/MY declare parameters; Y and named KY
apply arguments through small conditional template suffixes and explicit wrapper
forwarding. Preparation preserves native list interiors, rejects supplied blanks
and unsupported destinations, and installs String/`Void` absence for missing
inputs. Explicit native `Void` retains SourceString provenance. Effective inherited
file metadata rejects both names before composed/EA selection and during direct
glyph preparation, even when a local override is present. See the
[authoring contract](../builtins/README.md#lifetime-and-generic-inputs).

Direct primitive interpretation can bypass AML preparation. A read-only Rust-only
`metaSwitch` fallback returns String/`Void` for exactly these two missing names; it
does not mutate metadata, promote explicit native `Void` to absence, or change
generic systemless lookup. Shared ordered rendering and binding restoration remain unchanged.
Frozen inputs and the verified corpus define the supported boundary; preparation
tables and diagnostics do not establish general ownership or structural inference.

### Link files and assemble support

Use outline/object identity to determine output paths. Rust A currently renders
its package comment and stable, deduplicated file-owned imports, with the template
supplying `use` punctuation. The source assembler emits graph-owned attributes
and module declarations independently at the compiled A boundary; it never
searches or rewrites completed source. Structural version-2 roots receive safety
attributes, and explicit Module descriptors determine paths/visibility. Nonprivate
composed FW/HY/MY items are crate-visible in executable packages and public in
library targets; native EA contents remain opaque. See the
[topology contract](../builtins/README.md#frontend-topology-and-library-packages).

Validate normalized final paths, duplicates and reserved support collisions.
Do not infer cross-App modules. Keep source bundles atomic on failure. Support
templates may contain complete runtime implementations; segment templates may not.

Each Rust package has a Cargo manifest. Explicit topology supports binary targets
and a sole library-only package. A library-only package has declared sources and
Cargo.toml, without executable config/runtime/start/stop artifacts; IncludeRuntime
and supplied SETTINGS reject. Mixed packages retain one Primary binary and at
most one library. The CLI builds validated packages together as a workspace.

Runtime assembly wraps the first/orchestrator book and consumes the future
returned by its composed function, whose generated signature is:

```rust
fn l_main(runtime: &runtime::Runtime) -> impl std::future::Future<Output = runtime::Result<()>> + '_
```

Companion entries remain ordinary entry points. IncludeRuntime installs the
known pinned runtime dependency rows, rejecting conflicting supplied rows. The
assembler does not resolve crate versions, retry builds, translate Go source or
infer dependencies from `use` statements.

The CLI's `EvaluateWorkspaceEntry` coordinator bridge returns the composed
entry and a separate compiler-owned runtime library. It selects the runtime
import before rendering the entry header, then links the existing entry wrapper
at its known assembly boundary. The library copies the selected runtime/process
assets unchanged. The coordinator allocates its package identity and relative
dependency path; neither becomes AML metadata. Ordinary `Evaluate` and
`IncludeRuntime` retain their inline-runtime bundle contract.

The runtime wrapper supplies configuration startup before creating Tokio and
executes the composed future through Runtime supervision. Plain composed entries
with SETTINGS receive a prepared initialization operation before user effects.
EA remains an opaque whole-file path with its distinct support handling.

The unchanged Lucid cases use `IncludeRuntime: false` and compile without running.
Separate frontend, topology, library and runtime regressions compile and exercise
actual outputs; their [implementation evidence](../testdata/integration/IMPLEMENTATION.md)
must be consulted independently of the Lucid ledger. These focused checks do not
establish a new full-acceptance result.

## Cross-cutting invariants

- The selected `fs.FS` is authoritative; explicit providers are complete.
  Frontend `Lyger.Options.AssetRoot` records provenance, never provider selection.
- Shared ordered-body rendering and loop scoping remain shared with generic systemless interpretation.
  Narrow Rust hooks handle existing records, forwarded aliases and A's Header
  loop; they do not replace the shared interpreter or normalize completed source.
- Preserve L aliases, value snapshots and App identity. Missing lifetime or
  mutability information does not authorize cloning or implicit value changes.
- Keep diagnostics separate from application streams, preserve error causes,
  and return failures instead of process exits or partial bundles.
- Keep raw process output byte-preserving, including non-UTF-8 data, in runtime
  support. A Rust String substitution is not a blanket mapping for raw output.
- Configuration runs before Tokio resources exist. The Rust support design uses
  startup re-execution rather than unsafe environment mutation; explicit cleanup
  remains necessary for App processes and finalized captures.

See [CLI/runtime semantics](../docs/cli-build.md) for observable lifecycle details
and [Rust support documentation](../builtins/README.md) for the bounded
support contracts. Actual supervised tests, not asset presence or target selection,
provide runtime evidence; their passing scope does not prove every platform or
unrelated native form.

## Navigation for changes and evidence

| Work | Begin with | Evidence required |
| --- | --- | --- |
| Alphabet/classification | LOGEMES and target segments | Semantic reading, child roles and explicit change scope |
| AML framing/paths | AML, scanner, compiler | Complete documents, preserved literals, ordered full bundles |
| Nested rendering | Lucid parser/interpreter | Exact shared interpreter regressions and composed Rust programs |
| Target support | Service, selected assets, target support | Every artifact/path, resource errors, configuration/dependency isolation |
| Ownership/signatures | Prepared metadata and target boundaries | Legal programs plus intended native diagnostic cases |

The [Rust Lucid suite](testdata/lucid/README.md) retains six baseline
compositions and adds scoped selection, iteration,
declaration, invocation, callable, trait, empty-composition, keyed-pattern and
generic/lifetime coverage. It owns the current inventory and expected native
contracts. Shared low-level interpreter regressions retain Rust and systemless
controls. Passing reports from other matrices do not validate these templates.

The [acceptance guide](../docs/development-harness.md) owns the repository gate;
the Rust suite owns its separate Expected/Generated native compilation stage.
The [current ledger](testdata/lucid/FAILURES.md) records dated exact-generation,
validation, native-compilation and acceptance evidence. The 18 template hashes
remain frozen after finite reviewed amendments, linked from the
[Rust reference](../builtins/README.md). Those records authorize no future
interface expansion. That corpus remains compile-only. The separate runtime
regressions exercise generated native applications and topology; deployment
scripts remain unexecuted. Report each evidence layer and any outstanding
acceptance failures separately.

## Documentation design and provenance

The disclosure's stages (technical pages 27-46, summarized in LOGEMES) motivate
the separation of structure, composition, interpretation and source assembly.
Its example deployment scripts and alphabet do not override current
contracts. Keep this map about responsibilities and stable entry points, not
temporary failures or a full list of functions.

The documentation layout follows the emphasis on concise, scoped agent guidance
in [AGENTS.md](https://agents.md/) and
[OpenAI's guidance](https://learn.chatgpt.com/docs/agent-configuration/agents-md),
with component/data-flow navigation informed by
[ARCHITECTURE.md](https://architecture.md/). Long semantic references live in
linked documents rather than being duplicated in every AGENTS.md.
