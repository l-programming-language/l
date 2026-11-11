# Architecture of `l`

L, Lyger and Logiceme are implemented in Go. Lyger resolves source semantics and emits AML; Logiceme renders Rust from selected source definitions and frozen logeme assets. The current CLI compiles one standalone App with core Print. Runtime and authored standard libraries are absent.

## Language and compiler responsibilities

```text
source + one selected resource provider
  -> source preparation and imports
  -> tokens, grammar and semantic bindings
  -> standalone App and AML
  -> Rust source bundle
  -> supervised Cargo build
  -> published App executable
```

| Area | Owns |
| --- | --- |
| `cmd/l` | Arguments, source validation, standalone-unit selection, Cargo supervision and artifact publication |
| `lexicon/eng` | Configured roots, inflections, grammar structures and cases |
| `lyger` | Source preparation, vocabulary, grammar, bindings, types, source ownership and AML composition |
| `logiceme` | AML/glyph evaluation, source assembly and Rust bundles |
| `builtins/eng/core.l` | Retained source constructions and Text Print |
| `builtins/segments/` and `builtins/templates/` | Frozen native syntax and configuration templates |

Do not replace source algorithms with an interpreter or generated Go/Rust helpers. Core Print's private native leaf performs one output call. Sequencing, selection, iteration, error policy and resource choice belong in L.

## Production boundaries and retained APIs

Compiler APIs and internal type tags can describe more than the current executable supports. They do not establish that deleted dictionaries, orchestration, capture or process lifecycle are available. Unsupported effects reject explicitly.

The [build contract](cli-build.md) defines the current executable boundary. The [temporary hello-world gate](development-harness.md) defines current completion evidence; broader retained suites are historical regression material.

## Embedded builtins and explicit dictionaries

The distribution exposes canonical resources through one `fs.FS`. Embedded assets are the default; `--assets` replaces the complete provider without changing CWD or falling back to embedded resources.

`builtins/eng/core.l` loads directly. The root `dictionaries.json` is empty. There is no prelude or bundled Storage, IO, Formatting or Runtime. Explicit project `.l` includes resolve beside their importer. Dictionary-owned imports stay within the selected provider/root. No neighboring-file search, mirrored resources or second resolver is permitted.

## Preserve three representations

| Representation | Information retained |
| --- | --- |
| Literal pool | Exact literal bytes and source provenance |
| Semantic memory | Value/storage types, lexical roles, canonical bindings and definition identity |
| Ordered instructions | Subjects, predicates, clauses, source order, paragraph identity and spans |

A map of names cannot replace ordered instructions or paragraph-local references. A grammar tree without bindings cannot identify the selected meaning of Print. Raw source, expanded source, cached source and AML remain distinct representations.

## Internal types and ownership

Closed internal domains use checked enums; configured grammar names, lexical paths and source-defined names remain open text. A recognized internal tag does not make a value constructible in source.

Stored-value readers validate their wrappers and retain canonical App/Dictionary identity. Resolution and publication must not turn unknown state or failed lookup into a normal value. Registration preserves declaration tokens and commits a complete valid batch atomically.

Literal restoration scans original tokens once, preserves escapes and intervening bytes, and does not rescan inserted contents. Import expansion owns its accumulator and adopts outputs only after successful preparation.

## Source meaning and output views

[Source preparation](../lyger/source.go) records byte/line spans, paragraph identity and source order before caching or stamping. The [builder](../lyger/builder.go) expands imports, detects active-stack cycles and keeps typed origins. Embedded logical origins are not host paths.

Program and dictionary roots inherit core independently. Dictionaries cannot inherit caller-local bindings. The [semantic resolver](../lyger/semantic_resolution.go) starts at each token's lexical scope, checks complete bindings and follows canonical member ownership without an unscoped fallback.

Four-space encapsulation and physical blank separators determine lexical scope. Paragraph pronouns follow the last subject defined in original source order. Imported occurrences retain their own paragraph identity. Public/Publicly publish the selected declaration without exposing private siblings or changing its identity.

Dot-qualified names are not valid authored L. Use grammatical possessives or explicit aliases. Internal qualified identity strings remain separate. See [names and scope](spec/names-and-scope.md) and [lexical structure](spec/lexical-structure.md).

## Application structure and execution

Basic constructs an App around the Main Service/Object/Function composition supplied by core. The CLI validates one standalone App and compiles that App's actual source bundle. Multiple Apps, additional Service libraries and runtime-dependent top-level effects reject.

Retained Service, Class, Object and callable structures describe ownership and native topology; they do not introduce a runtime or a second process. The App executable runs directly.

## Frontend diagnostics and results

Diagnostics retain original source spans, expansion call sites, phase and native causes. Preserve tested observable source/AML views without replacing the real compilation graph with fixture constants. The [error model](error-model.md) describes classified failures.

## Native executable pipeline

Logiceme returns Rust source bundles. The CLI checks logical paths, dimensions, target/dependency relationships and actual Cargo artifact identity. It does not rewrite generated source to infer ownership or synthesize an orchestration program.

Builds use pinned tools, frozen offline dependency resolution, existing deadlines and private staging. Artifact publication remains atomic. The retained generation directory contains a copy of the standalone executable, not a separate child-process implementation. See [publication](cli-build.md#output-and-publication).

## AML and logeme protocol

AML is structured markup, not machine assembly. The [AML protocol](../logiceme/AML.md) and [logeme contracts](../builtins/LOGEMES.md) define framing, books/pages, composition and ownership. `Evaluate` returns a source bundle; its caller owns persistence, compilation and execution.

`^` integrates children, `*` combines siblings and parentheses preserve grouping. File-owned metadata, imports and callable declarations retain their source owner. Native lifetime/generic inputs remain glyph-local and do not add global inference or new metadata fields. Deferred metadata rejects.

Only the eighteen Rust segments and four templates are protected during the reset. Preserve every byte, interface and frozen hash expectation. Retained syntax for selection, iteration and futures does not supply deleted source words or lifecycle behavior.

## Runtime test boundary

There is no Runtime implementation in the current distribution. The temporary gate observes standalone Print and rejection of runtime effects. Historical runtime tests and their old passing results are not current acceptance claims.
