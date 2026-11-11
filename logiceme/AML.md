# Assembly Markup Language and the logeme handoff

AML carries application structure, output layout, settings, metadata and composed
logeme sequences into `LogicemeService.Evaluate`. Read
[LOGEMES.md](../builtins/LOGEMES.md) for letter meanings and
[CONTRACTS.md](../builtins/CONTRACTS.md) for assembly contracts and future design constraints. This guide documents
framing and logical obligations. It does not make proposed syllables operational.
The [Rust suite](testdata/lucid/README.md) preserves the original flat composed
baseline and adds scoped selection, iteration, declaration, invocation, callable,
trait, empty-composition, keyed-pattern and generic/lifetime coverage. Its
inventory and dated results distinguish generation, native compilation, and
bounded execution of three future cases. Source-graph schemas and preparation
branches are documented separately. Current CLI execution is limited to core
Print in a standalone App.

## Preserve the representation boundary

Lyger's observable AML and its canonical backend AML describe the same graph but
are not interchangeable byte strings. Backend syllabic sequences are framed with
backticks and tags. Use the existing frontend-to-backend handoff. Do not repair
test inputs with ad hoc replacements or change protected observable output.
ASM/JIT field names do not mean this is machine code.

The markup is a configured, framed protocol, not general XML or HTML. Use its
scanner and literal boundaries. Do not use substring searches or a generic XML
parser that rewrites native bytes, whitespace, quoting or unknown tag syntax.

## Structural hierarchy

| Unit | Role |
| --- | --- |
| PROGRAM / STRUCTURE | Document and ordered application structure |
| BOOK | One application/source bundle |
| OUTLINE / JSON | Application/component metadata and output-file structure |
| SETTINGS | Configuration values for the generated application |
| CONTENT | Ordered generated source content |
| CHAPTER | Service |
| TEXT | Class |
| SENTENCE | Object/source-file content associated with the outline |
| W | Word container for syllabic sequences |
| S | One framed syllabic sequence |

Word and sequence tags are spelled `<|W|>` and `<|S|>` respectively, as shown below.
Class grouping is optional for a small application.
Outline names and locations must agree with content ownership. Book identity
does not replace the per-file opening A declaration.

## Complete composed example

This Rust example illustrates existing framing and composition. It is not an
oracle for exact renderer whitespace. Calculation and definition remain composed.
Native names/literals are metadata atoms.

```text
<PROGRAM><STRUCTURE><BOOK "Example"><OUTLINE>
<JSON>{"AppMeta":{"AppGenDir":"/"}}</JSON>
<JSON>{"Services":[{"Name":"main","Location":"/","GenericObjects":["main"]}]}</JSON>
<JSON>{"Objects":[{"Name":"main","Type":"root"}]}</JSON>
</OUTLINE><SETTINGS></SETTINGS>
<CONTENT><CHAPTER "main"><SENTENCE "main"><|W|>
<|S|>`A`</S|>
<|S|>`FW{|Key: ~main~, Outputs: ~&[()]&~|}^W^(DOB{|Keys: ~&[result]&~|}^DE{|Ops: ~&[Op{Type: Calculate, Operator: +, Operands: &[1, 2]&}Op]&~|}*BOD^E{|Statement: ~::std::println!("{}", result);~|})`</S|>
</W|></SENTENCE></CHAPTER></CONTENT>
</BOOK></STRUCTURE></PROGRAM>
```

The intended source opens a source unit, defines main, calculates and binds
result, then prints it. Native fragments remain explicit Rust source. They are
not translated from another language. A remains the ordinary source-unit opener.
Frozen Rust A renders a package comment and `use` imports. The assembler supplies
graph-owned module/safety items separately at the compiled A boundary. The
template interface remains unchanged. See [topology and libraries](../builtins/README.md#frontend-topology-and-library-packages).

## Services, objects, and source layout

The outline and content describe two connected views of an application. The
outline supplies membership and file placement. Content supplies the logeme
programs that become those files. This separation follows the disclosure's
file-system scaffolding and content-tag structures: pages 30-37, particularly
linking stages 602-608 and Figure 7's service/object definitions. See
[technical provenance](../builtins/LOGEMES.md#technical-provenance).

An AML **service** groups related source objects within one application. It is
not an L App, independently launched process, network service, or Rust task.
An AML **object** is a named source-file unit, not automatically an object-oriented
class, Rust struct, or runtime value. Those language constructs are expressed by
the syllables inside its Sentence.

### Reading the minimal outline

```text
<JSON>{"Services":[{"Name":"main","Location":"/","GenericObjects":["main"]}]}</JSON>
<JSON>{"Objects":[{"Name":"main","Type":"root"}]}</JSON>
```

These records have distinct responsibilities:

1. `Services` declares a service named `main`, corresponding to
   `<CHAPTER "main">` in CONTENT.
2. Its `GenericObjects` list references the object named `main` in the book's
   `Objects` catalog and places it at the service's `Location`.
3. The matching `<SENTENCE "main">` supplies that object's Words and syllables.
4. Ordinary source assembly combines the selected location, object name, and
   Rust suffix: `/` + `main` + `.rs` gives `/main.rs`.

Names match exactly. The service and object happen to share `main` here, but
they belong to different structural roles and need not have identical names.
The Sentence name must match its referenced object. A function's `Key` need not
match either name. A single object may contain several functions or type items.

`"Type":"root"` is an outline classification used by existing producers. It does
not create a main function, choose a native type, select a suffix, or direct
runtime wrapping. In particular, changing an object's Type does not turn its
Sentence into an application entry point. Source logic, target support, and the
build boundary must agree on the entry contract.

### Service fields and membership

The backend reads service records into `OutlineEntry` in
[compiler.go](compiler.go). Classes use the same record shape, with the versioned rules below.

| Field | Role and current behavior |
| --- | --- |
| `Name` | Service identity looked up from the Chapter name. Use an unambiguous name within the book's service catalog |
| `Location` | Logical output-directory prefix for this entry's GenericObjects. Author `/`, `/src/`, or `/src/tools/` with a trailing slash |
| `GenericObjects` | Names of catalog objects placed at this entry's Location. This overrides each referenced object's own Location |
| `DynamicObjects` | Names of catalog objects that retain their individual object-level Locations |
| `Dimensions` | In StructureVersion 2, direct child Class identities for both Services and Classes |
| `Parent` | In StructureVersion 2, immediate enclosing Class identity; omit for top-level Classes. It never denotes a process owner |
| `Type`, `EntryType` | Accepted descriptive fields. The current assembly path does not use them to select segment templates or create target-language constructs |

Here **generic** means placement inherited from the containing outline entry.
It has no connection to Rust generic parameters. **Dynamic** means retaining
object-specific placement, not dynamic dispatch, dynamic loading, or execution.
List an object in one of these membership lists for its owner, not both.

The source hierarchy nests Classes and Objects recursively within a Service. Classes
organize files, not runtime values. Populated Classes use `StructureVersion: "2"`.
Flat Service/Object inputs use the flat assembly path.

### Object definitions and content

| Field | Role and current behavior |
| --- | --- |
| `Name` | Catalog identity and the matching Sentence's output basename; ordinarily omit the target extension |
| `Type` | Descriptive classification such as `root`, distinct from a syllable's Type metadata |
| `Location` | Explicit directory prefix used through DynamicObjects; GenericObjects substitutes its containing entry's Location |
| `Package`, `Private`, `Meta`, `Inputs` | Fields present on the backend Object record, but not a fully implemented outline-to-template configuration interface |
| `Sentence`, `Bytes` | Also present on the internal serializable record; ordinary AML authoring supplies code in CONTENT, not through these fields |

In StructureVersion 2, the validated catalog record supplies each Object's
placement, Module role and ordinary Meta, while the Sentence supplies its code.
Local metadata overrides inherited ordinary generation data. Imports, paths,
root identity, target policy and modules are prepared separately and are not
inherited as ordinary shared data. Object.Package remains descriptive. It does
not establish a Rust module or Cargo package relationship.

Outline entries never synthesize content. Every version-2 catalog Object and
Class must have exactly one matching CONTENT owner. Missing, duplicated,
conflicting and cyclic relationships produce classified errors and no bundle.

### Placement example: shared and individual locations

This layout-only excerpt illustrates a Rust project. It is not a complete AML
application or a compiled example:

```text
<JSON>{"AppMeta":{"AppGenDir":"/src"}}</JSON>
<JSON>{"Services":[{"Name":"main","Location":"/src/","GenericObjects":["main","helper"],"DynamicObjects":["worker"]}]}</JSON>
<JSON>{"Objects":[{"Name":"main","Type":"root"},{"Name":"helper","Location":"/ignored/"},{"Name":"worker","Location":"/src/tools/"}]}</JSON>
```

Supply the corresponding composed content as three Sentences within
`<CHAPTER "main">`, each opening with A:

| Sentence | Location selection | Ordinary Rust output |
| --- | --- | --- |
| `main` | GenericObjects inherits `/src/` | `/src/main.rs` |
| `helper` | GenericObjects overrides `/ignored/` with `/src/` | `/src/helper.rs` |
| `worker` | DynamicObjects retains `/src/tools/` | `/src/tools/worker.rs` |

File placement alone does not generate imports, exports or App calls. Explicit version-2 Module/Target metadata
and Object privacy drive the separate Rust module/item visibility assembly.
Those relationships must be expressed or prepared through their own contracts.

The leading `/` denotes the root of the returned application's logical bundle,
not the host filesystem root. Object assembly currently concatenates Location
and Name. It does not insert a missing slash between them. With a blank Location
it uses `/` plus Name. Supply directory prefixes explicitly. `/src` as an object
prefix would produce `/srcmain.rs`, not `/src/main.rs`. Final Rust paths are
validated, including duplicate and reserved-support collisions.

`AppMeta.AppGenDir` is separate from object placement: it does not relocate the
objects assembled from Services/Objects. Flat Rust support looks for
`main.rs` beneath AppGenDir, defaulting to `/src` when it is missing or empty,
and puts configuration source support beside it. Keep this value aligned
with the entry object's resolved path. To use the minimal outline's `/main.rs`,
set AppGenDir to `/`. An empty value selects `/src/main.rs` for Rust. Cargo's
binary path is then rendered relative to the manifest. A `root` Type cannot
repair a mismatch between these locations. See [bundle_support.go](bundle_support.go).

### Ordering, validation, and authoring rules

- Use separate JSON records for AppMeta, Services, Classes, and Objects.
  Put all entries of a category in one array. `ScanOutline` classifies each record
  once. Version 2 rejects repeated category records, while the flat reader
  replaces the previous array rather than appending. Combining Services and
  Objects into one JSON record is not a supported substitute. See [scanner.go](scanner.go).
- Version-2 CONTENT supplies depth-first, interleaved generation order. A child
  Class is traversed where its TEXT occurs among direct Object Sentences.
  GenericObjects/DynamicObjects and Dimensions are membership lists, not a
  separate ordering program. Flat Services retain their declared order.
- Require unique catalog identities, resolvable references, consistent ownership,
  acyclic parent relationships, and unique final output paths. Validate both
  names and paths: different service names do not prevent file collisions, and
  renaming a function does not rename its containing Object.
- Do not rely on flat first-match/default-location behavior for version-2
  source graphs. All references and final paths are validated before publication.

## Classes as optional source groups

A Class recursively groups child Classes and Object file wrappers. It does not
introduce a runtime activation, Rust lifetime, L paragraph, or cleanup scope.
L pronouns, source paragraphs, continuations and possible Sections retain their
separate semantic contracts.

Version 2 is selected by `<JSON>{"AppMeta":{"StructureVersion":"2"}}</JSON>`.
Services and Classes use `Dimensions` for direct child Classes. A Class's Parent
must match its immediate enclosing Class. Top-level Classes omit Parent. Content
nesting, outline membership and Parent must agree. Catalog identities are unique
within their kind across the Book. Module names are unique within a parent module.

### Rust module descriptors

The following schema drives Rust source-graph validation and structural header
assembly. Module declarations belong to the assembler, not the frozen A template.
Topology/library regressions supplement the flat Lucid corpus.

Each version-2 Service supplies `Target` with `Kind: "bin"` or `"lib"`, target
`Name`, `Root` Object identity and Boolean `Primary`. Each Class supplies its
`Root` Object and a `Module` descriptor. Every nonroot Object has a Module.
A root is a direct Object owned by that Service or Class, not an inferred filename.

Module descriptors contain `Name` and optional `Visibility`. Omitted or `private`
is private. `public`, `crate`, and `super` prepare `pub`, `pub(crate)`, and
`pub(super)` module declarations. Item visibility is independent. One Class root
cannot be silently flattened into another file, and no automatic re-exports are
emitted.

```text
<JSON>{"AppMeta":{"StructureVersion":"2","AppGenDir":"/src","Deps":"[]"}}</JSON>
<JSON>{"Services":[{"Name":"main","Location":"/src/","GenericObjects":["main"],"Dimensions":["Models"],"Target":{"Kind":"bin","Name":"example","Root":"main","Primary":true}}]}</JSON>
<JSON>{"Classes":[{"Name":"Models","Location":"/src/","GenericObjects":["models"],"Dimensions":["Details"],"Root":"models","Module":{"Name":"models","Visibility":"public"}},{"Name":"Details","Parent":"Models","Location":"/src/models/","GenericObjects":["details"],"Root":"details","Module":{"Name":"details"}}]}</JSON>
<JSON>{"Objects":[{"Name":"main"},{"Name":"models"},{"Name":"details"}]}</JSON>
```

The matching Chapter contains main's Sentence and Models' TEXT. Models contains
its root Sentence and Details' TEXT. Details contains its root Sentence. Each
composed Sentence starts with A. The [Lucid suite](testdata/lucid/README.md) covers
flat compositions. Independent backend topology tests check recursive modules
and native library use without the removed L libraries.

The compiler emits crate-root safety attributes, child module declarations and
relative `#[path]` attributes from the validated graph. It inserts these structural
items at the known compiled A boundary without reparsing completed source.
Nonprivate Objects expose composed top-level FW/HY/MY items as `pub(crate)` in
executable packages or `pub` in library targets. Module visibility remains a
separate boundary. Private Objects keep private items. No native EA declarations
are rewritten.

One Book produces one Cargo package. A package containing binaries has at most
one library and exactly one Primary binary. Alternatively, a sole library target
with `Primary: false` returns only declared sources and Cargo.toml.
Supplied executable SETTINGS reject. Explicit targets disable automatic bin, lib,
example, test and benchmark discovery. Additional binaries are source artifacts,
not new L App identities. Flat inputs keep one binary at AppGenDir/main.rs.
EA remains an opaque whole-file exception and supplies its own literal header.
It cannot appear beside A or be used to bypass composed module testing.

### Class metadata and unsupported Features

Class `Meta` is an ordinary JSON object of generation metadata. Ordinary local
keys override ancestor keys without modifying parent/sibling scopes. Rust
`Lifetimes` and `GenericTypes` are exceptions: they belong to individual glyphs
and are rejected whenever present in effective inherited Class/Object file
metadata, including blank or sentinel-valued entries. A local glyph override
does not rescue invalid inherited placement, and the check also applies to EA
files. Package metadata does not turn source grouping into lifetime ownership.
Reusable shared Attributes are attached explicitly to Object compositions in
their reviewed order.

Features and PARAGRAPH are unsupported in canonical AML and active assembly.
An outline Features key, including an empty array, or structural PARAGRAPH is a
classified error. The reader never silently discards their content.
Express file membership with recursive Classes, inherited ordinary data with
Class Meta, and shared logical behavior with explicit compositions. Native source
containing these words remains literal data. L source paragraphs and runtime
activations are separate concepts.

## Syllables, integration and data

Within a backtick-framed sequence:

```text
FW{...} ^ W ^ (DOB{...} ^ DE{...} * BOM ^ KY{...})
```

Braces and ellipses above are schematic. Actual metadata uses `{|...|}`, as in
the complete example. `^` integrates a child, `*` orders siblings and `()` groups
composition. Use explicit parentheses to make the intended tree unambiguous.
Letter codas qualify a syllable. They are not an alternative nesting operator.

The parent consumes children through declared roles such as
`Conjoinments.Calculation` or `Conjoinments.Operations`. Singular roles require
the declared cardinality. Plural collections preserve order. Do not drop a child
because a template forwards the wrong collection. A broad category alone does
not prove that the native type/expression/statement/item/pattern role is legal.

## Metadata framing

Existing forms use named values enclosed by `{|` and `|}`, with native payloads
framed using `~...~`. Sequence and record examples include `&[...]&`,
`Prop{...}Prop`, `Func{...}Func`, and `Op{...}Op`. Scalar and collection types are
defined by the selected segment's Inputs contract. The existing implementation
also has record/receiver framing. Preserve those contracts rather than inventing
a second serialization format.

Preserve Meta, Funcs, Props, Ops and Conjoinments names. Func/Prop/Op group inputs
but do not dictate logical ownership. Keep missing, explicitly empty and false
values distinct wherever their contract distinguishes them. Existing `Void`
absence handling is not interchangeable with every record's empty-string field.

Respect framed source data: native strings, comments, backticks, generic type
commas, AML-like text and Lucid-like markers are not permission to split or
interpret the contents. Use supported framing. Do not invent undocumented escape
syntax. Escaping Rust literals, TOML entries and shell arguments are separate
target-preparation responsibilities.

Rust's template inputs, conditions and forwarding are frozen. Preparation fields
and validation entries do not independently establish an authoring interface.
`Mutable` supplies no general mutability contract. CV does consume glyph-local
`Capture: Infer|Move` under the [future contract](../builtins/README.md#future-construction-awaiting-and-value-bindings);
FW/BW reject supplied Capture. Other native mutability and async scaffolding
remain explicit where required.

### Rust lifetime and generic inputs

`Lifetimes` and `GenericTypes` are glyph-local String metadata beside signature
records. These framed excerpts contain native list interiors, first for a
declaration and then for an application:

```text
{|Lifetimes: ~'a, 'b~, GenericTypes: ~T, U: Clone + 'a~|}
{|Lifetimes: ~'_~, GenericTypes: ~String, Vec<(u8, u16)>~|}
```

FW/HY/MY declare parameters; Y and named KY apply arguments. The
[Rust input contract](../builtins/README.md#lifetime-and-generic-inputs) owns
placement, absence, provenance, punctuation and native-validation rules. These
values are neither record fields nor inherited file metadata.

## Source units, imports and modules

Each ordinarily composed Sentence/Object starts with A. Collect imports from all
its nested syllables before rendering the header, with no leakage into siblings
or another book. Rust preparation provides A with an ordered, deduplicated Imports
collection of bare payloads. Its existing Header loop supplies `use` punctuation.
An empty collection emits no imports. This path remains distinct from structural
attributes/module assembly described above. Do not substitute a
whole-file EA for the feature under composition.

An explicit source/module graph determines which file declares a module. An App
book is a separate project. File-level inner
attributes and item-level outer attributes have different attachment rules.
The source assembler supplies configuration declarations independently of A.
Plain composed entries with SETTINGS receive initialization before user effects.

Native source suffixes are fixed. Outline/Object metadata controls paths.
Validate final paths, duplicate outputs and reserved support collisions. Do not
derive filenames from native fragment text or from unrelated later Type fields.

<a id="evaluation-options-and-application-bundles"></a>

## Evaluation and application bundles

```go
bundle, err := service.Evaluate(completeAML)
```

Rust output is independent of Lyger's language selection. The backend has no
runtime-injection option.
Input and earlier returned bundles remain unchanged after every call.

Rust evaluates all books in source order, returning one project per application.
The current CLI selects one standalone App. Raw backend targets are a separate
source-generation contract.
See [ARCHITECTURE.md](ARCHITECTURE.md).

Each Rust package has a Cargo manifest. Executable packages receive applicable
support artifacts, while library-only packages omit them. No orchestrator or
process runtime is supplied by the current distribution.

The existing Rust support boundary uses App.Name, App.MainLocation and App.Deps.
AppMeta.Deps contains a JSON-encoded ordered array of normalized TOML dependency
rows. It is decoded in support assembly, not generic metadata parsing. The
assembler does not resolve versions or infer dependencies from use statements. Configuration keeps source order
so duplicate keys have last-entry-wins behavior, and values use target-appropriate
escaping. Resolution, downloads and compilation retries are build-coordinator
work. Tests use retained locks/pins without fallback.

## Failure boundaries and tests

Malformed framing, unsupported forms, wrong child roles/cardinalities and invalid
output paths are generator errors. Type/ownership/exhaustiveness constraints that
require Rust compilation belong to native diagnostic tests after valid generation.
An unrelated syntax error, missing dependency or unrelated generator rejection cannot satisfy
an intentional Rust semantic-negative case. Preserve causes and source context.
The [current suite](testdata/lucid/README.md) records complete successful-build
and intended native-negative scenarios, bounded malformed-composition cases,
and separate Expected/Generated compilation contracts. Its inventory defines
the current obligations.

For end-to-end tests, pass a stored complete document to Evaluate and compare
every returned artifact, not only main.rs. Verify ordered application/path
inventories, dependency metadata, exact bytes and no unexpected Go files for a
Rust target. Compile the actual returned project when that stage is required.
Never substitute expected code or formatting repairs. See the
[minimum composed corpus](../builtins/CONTRACTS.md#verification-contract).
Three Lucid future cases have bounded execution checks. Deployment scripts remain
unexecuted.

Native fragments are appropriate for atoms and external calls. They must not
replace the feature under test. EA is the emergency standalone whole-file
exception. Its placement and verbatim behavior are described once in
[the shared contract](../builtins/CONTRACTS.md#literal-file-exception). Ordinary
logeme examples and parity work should concentrate on composed programs.
