# Runtime

Runtime is the default standard library for L execution. The compiler links it
automatically, so ordinary App programs need no Runtime import. Its words are
advanced native authoring interfaces. Import them explicitly when writing a
source composition:

```l
Use Runtime from `runtime`.
```

An import exposes the selected dictionary's vocabulary. It does not choose a
different implementation for that source file or App.

## Select a replacement

Use one complete `--assets` provider from the same compiler revision, containing
its grammar, vocabulary, builtins, standard libraries and native assets. In its
root `dictionaries.json`, retarget `languages.eng.runtime.entry` to the replacement
L entry and retain `"dictionary": "Runtime"`. The entry must declare Runtime and
export Runtime.Host for startup and native trait adaptation. Source files stay
under a permitted library root, and relative includes remain confined to the
entry's directory. Other indexed imports resolve through that provider too.

The bundled source lives in `stdlib/eng/runtime/`. Shared native inputs live in
`stdlib/runtime/rust/`, and explicit generation writes the selected provider's
`stdlib/runtime/generated/support.aml` and `support.json`. From the matching
compiler checkout:

```sh
scripts/l-go run ./tools/composition -assets /absolute/provider
scripts/l-go run ./tools/composition -assets /absolute/provider -check
l build --assets /absolute/provider main.l
```

The complete provider must already contain the output directory. Generation
writes only the two support files. Check mode writes nothing, and neither
operation changes the caller's working directory. A build checks that external
support matches its current L sources. Missing or stale resources fail without
regeneration or fallback. Offline contract inspection never runs the generator,
native compiler or toolchain probes.

One provider selects one runtime for the whole workspace. Ordinary imports,
neighboring files and same-spelled user words cannot replace it. This replacement
boundary supports L policy under the current Rust/Tokio protocol and pinned
dependencies. Different executors, platforms or native representations require
separate compiler work.

## Compiler runtime contract

The generated `runtime` module supplies `Runtime`, `Error`, `App`, `Invocation`,
`Snapshot`, `Terminal` and `Result<T>`. `Result<T>` is
`std::result::Result<T, Error>`, including propagation through `?`. App
handles use `std::sync::Arc<App>`. The unchanged native host adapter implements
Clone for Runtime, Error, Snapshot and Terminal, and Debug, Display,
`std::error::Error` and `From<std::io::Error>` for Error through the functions
below. Private struct fields and the bundled implementation's test hooks are
not replacement interfaces.

These signatures name native exports, not additional L words. `Arc`, `Future`,
`Pin`, `PathBuf` and `Formatter` below denote their standard Rust types.

| Native export | Inputs | Result |
| --- | --- | --- |
| `runtime_new` | None | `Result<Runtime>` |
| `runtime_run` | `&Runtime`, `impl Future<Output = Result<()>>` | `impl Future<Output = u8>` |
| `runtime_copy` | `&Runtime` | `Runtime` |
| `error_copy` | `&Error` | `Error` |
| `error_display` | `&Error`, `&mut Formatter<'_>` | `std::fmt::Result` |
| `error_native_cause` | `&Error` | `Option<&(dyn std::error::Error + 'static)>` |
| `error_from_io` | `std::io::Error` | `Error` |
| `snapshot_copy` | `&Snapshot` | `Snapshot` |
| `terminal_copy` | `&Terminal` | `Terminal` |
| `runtime_app` | `&Runtime`, `PathBuf` | `Arc<App>` |
| `app_set_source` | `&App`, `String` | `()` |
| `app_output` | `&App` | `Result<Arc<[u8]>>` |
| `app_error` | `&App` | `Result<Option<Error>>` |
| `error_new` | `Vec<u8>` | `Error` |
| `error_new_text` | `&str` | `Error` |
| `runtime_show` | `&Runtime`, `&[u8]` | `impl Future<Output = Result<()>> + Send` |
| `runtime_fail` | `&Runtime`, `Option<Error>` | `Error` |
| `format_number` | `f64` | `Vec<u8>` |
| `app_build` | `&App` | `Result<()>` |
| `app_run` | `&Arc<App>` | `Result<Arc<Invocation>>` |
| `app_stop` | `&App` | `Pin<Box<dyn Future<Output = Result<()>> + Send + '_>>` |
| `app_wait` | `&App` | `Pin<Box<dyn Future<Output = Result<Arc<Invocation>>> + Send + '_>>` |
| `invocation_claim` | `&Invocation` | `Result<bool>` |
| `invocation_finish` | `&Arc<Invocation>` | `()` |

Returned futures must support the caller's borrowed arguments for their actual
lifetimes. Error causes borrow the supplied Error. Generated Functions also use
`&(runtime::Runtime, LProgram)` contexts and futures returning
`runtime::Result<std::ops::ControlFlow<...>>`. Native compilation checks these
types and lifetimes at the actual call boundary.

A replacement preserves the [lifecycle contract](../../cli-build.md#invocation-and-results):
Run acknowledges launch, Stop requests termination without joining, and Await
joins finalized output and claims its continuation once. Preserve exact
application bytes, explicit Error bytes, ownership transfers, panic propagation,
process cleanup and the [shutdown limits](../../cli-build.md#shutdown-streams-and-limits).
Compiler-generated misuse diagnostics may use another presentation while
retaining their messages, source attribution, statuses and failure behavior.
Successful compilation establishes interface compatibility, not these runtime
guarantees. Behavioral conformance needs execution against the selected provider.

## Words

Words from the [Runtime dictionary](../../../stdlib/eng/runtime/dictionary.l) are
listed alphabetically below. Each link opens its definition, declared forms,
description and usage in the [Word Glossary](../../reference/word-glossary.md).

<!-- word-index:eng:runtime -->

## Previous reference sections

<a id="import-and-calling-forms"></a>

[Import and calling forms](../../reference/word-glossary.md#context-runtime--import-and-calling-forms).

<a id="public-and-private-names"></a>

[Public and private names](../../reference/word-glossary.md#context-runtime--public-and-private-names).

<a id="app"></a>

[App](../../reference/word-glossary.md#context-runtime--app).

<a id="capture"></a>

[Capture](../../reference/word-glossary.md#context-runtime--capture).

<a id="error"></a>

[Error](../../reference/word-glossary.md#context-runtime--error).

<a id="format"></a>

[Format](../../reference/word-glossary.md#context-runtime--format).

<a id="invocation"></a>

[Invocation](../../reference/word-glossary.md#context-runtime--invocation).

<a id="output-and-shutdown"></a>

[Output and shutdown](../../reference/word-glossary.md#context-runtime--output-and-shutdown).

<a id="startup"></a>

[Startup](../../reference/word-glossary.md#context-runtime--startup).

<a id="lifecycle-example"></a>

[Lifecycle example](../../reference/word-glossary.md#context-runtime--lifecycle-example).

<a id="sources-and-witnesses"></a>

[Sources and witnesses](../../reference/word-glossary.md#context-runtime--sources-and-witnesses).
