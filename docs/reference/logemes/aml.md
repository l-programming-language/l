# AML, ownership and the source-bundle boundary

## Assembly Markup Language

AML means Assembly Markup Language. It describes source topology, metadata and logeme composition; it is not machine-code assembly. It is a custom framed protocol, not generic XML/HTML. Preserve backticks, delimiters, whitespace and literal native bytes. Rendering AML as escaped code is appropriate; parsing it through a DOM/XML normalizer is not.

| Unit | Meaning |
| --- | --- |
| PROGRAM / STRUCTURE | Ordered program structure. |
| BOOK | One application/source bundle. |
| OUTLINE / JSON | Source membership, locations and metadata. |
| SETTINGS | Generated-application configuration. |
| CONTENT | Ordered source composition. |
| CHAPTER | Service source group. |
| TEXT | Class grouping. |
| SENTENCE | Object/source-file content. |
| W / S | Framed word container and syllabic sequence. |

The actual word/sequence openers are `<|W|>` and `<|S|>`. [The canonical AML guide](../../../logiceme/AML.md#complete-composed-example) provides complete framing. Source-generation code matches outline names and paths to content ownership. An AML Service is a source grouping; it is not automatically a process. Objects remain files and Classes remain module/source groups.

## Two AML views

Lyger's observable AML and its canonical backend AML describe the same graph but differ in framing bytes. Consumers must use the existing handoff rather than inserting backticks with ad hoc replacements. Raw L, cached literals, expanded source, observable AML and backend AML have different purposes. Preserving one view does not authorize normalizing another.

## Evaluate and Executable

```go
service := Logiceme.NewLogicemeService(nil)
bundle, err := service.Evaluate(aml, Logiceme.GenerationOptions{
    IncludeRuntime: false,
})
```

This Go API example assumes `aml` is complete framed input and the Logiceme import is in scope. Nil resources selects embedded assets; a nonnil filesystem is a complete replacement. Evaluate accepts at most one GenerationOptions value, containing only IncludeRuntime. It has no target-language option: Rust is the sole generated language, and source lexicon selection belongs to Lyger.

Executable is an in-memory source bundle: ordered ByteArrays align with MapArrays paths by application, with DependencyArrays and EnvDepArrays. It is not a running program. Errors return a zero-value bundle rather than partial output. Reusing a service must preserve previous bundle bytes and isolate metadata/provider state.

## Native build boundary

The CLI checks safe logical paths, dimensions, collisions and Cargo graph identities before materializing and compiling bundles. A leading slash in a bundle path is a logical root, not an arbitrary host destination. Library-only units have no executable wrapper/settings/start-stop artifacts; nonprivate composed items form their public interface. Raw AML can explicitly compose a separate native consumer of a library; this does not make a standalone L Service runnable.

The direct API can include inline runtime support. The CLI's workspace-entry bridge renders a separate library under the [runtime selection contract](../../cli-build.md#source-generation-and-native-builds). Neither path rewrites generated source to infer ownership. See [Logiceme architecture](../../../logiceme/ARCHITECTURE.md) for implementation entry points.
