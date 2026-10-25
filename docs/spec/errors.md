# Errors, diagnostics, and compatibility

Compiler diagnostics, native build failures, runtime lifecycle misuse, and application output are distinct channels and stages. A successful website build does not establish successful language compilation or native execution.

## Failure stages

| Stage | Typical failures |
| --- | --- |
| Invocation | Unknown command/option, misplaced option, extra source argument, unsupported host |
| Resource initialization | Missing or malformed selected grammar, vocabulary, builtin, template, or dictionary index |
| Source preparation and grammar | Unterminated literal/comment, malformed definition framing, unassociated tokens |
| Binding and semantics | Unresolved/ambiguous references, duplicate publication, alias/import/expansion cycle, incompatible types |
| Typed lowering | Unsupported recognized feature, unavailable capture, definite-assignment failure |
| Source generation | Invalid AML framing, unsupported glyph composition, wrong child roles or metadata |
| Native build/publication | Missing pinned toolchain/dependency, Cargo failure, unsafe source bundle, invalid artifact identity, filesystem failure |
| Generated runtime | Lifecycle misuse, launch/signal/capture/cleanup failure, failed output writer |
| Child application | Nonzero status, native panic, stderr output, application-specific failure |

Errors retain original causes where the maintained interface supports them. A failed prerequisite does not prove later stages passed or failed.

## Compiler categories

The Go `lerrors` package classifies InvalidInput, InvalidConfiguration, Syntax, UnresolvedReference, AmbiguousReference, DuplicateDeclaration, TypeMismatch, Cycle, Unsupported, IO, Tool, Internal, and Unknown. Numeric category values are not a serialization contract. Unknown is an error category, not success or a wildcard.

These are compiler/API inspection categories for Go callers. L does not thereby gain `try`, `catch`, thrown exceptions, or an imported standard-library error module. Native process Error is a separate runtime member.

Grammar alternatives can aggregate candidate causes while preserving sorted diagnostics. A source numeric overflow may remain provisional until lexical resolution is exhausted. Direct definition Number binding has its own Syntax boundary; ordinary numeric input overflow and nonfinite native lowering have different classified boundaries. Do not collapse these into one invented source exception.

## Source attribution

Diagnostics retain source file, byte-based line/column position, and original paragraph/source ownership. Native definition failures include definition/parameter context; expanded actions retain call-site attribution. Imported source belongs to its provider even when the calling program is elsewhere.

Incidental UUIDs may be normalized for a diagnostic view. They must not be normalized in application output, canonical input, expected bytes, or generated artifacts merely to make a comparison pass. Original literal bytes and instruction order remain authoritative.

## CLI statuses and streams

The compiler CLI returns 0 for help and successful builds, 2 for invalid invocation or unsupported host, and 1 for initialization, compilation, generation, native build, publication, or output-write failure. Compiler diagnostics go to stderr. Successful build stdout contains the published executable path and newline.

Generated programs have their own runtime statuses: normal fallthrough/bare Exit succeeds if cleanup succeeds; explicit failure or fatal misuse returns 1; supported SIGINT/SIGTERM termination uses 130/143. A child application's nonzero result is not automatically the orchestrator's exit status. See [execution](execution.md).

## Unsupported versus unresolved

A recognized While clause can reach an Unsupported instruction boundary. A missing name reaches a reference boundary. A configured Pointer noun or backend callable template does not establish an operational source construct; its documentation must state the narrower known facts rather than promise a future behavior.

The specification audit records Byte/Bytes/Pointer/Null construction gaps, collection action support, joined conditions, source versus native callable forms, and other unresolved distinctions. Those notes are reasons to investigate a specific feature, not permission to change compiler behavior or protected fixtures during documentation work.

### Rejected operands: a type error and a missing name

These are complete negative compiler fixtures, not runnable examples. The website includes their source; the compiler rejection tests were not run in this pass.

<!-- example:invalid-await-text -->

The backtick literal is Text. Await requires a process target, so this is a TypeMismatch at compilation, before any application can be launched. The [fixture contract](../../testdata/phase1/n01-await-text/case.json) expects a diagnostic containing `await` and `app` and no new published executable.

<!-- example:invalid-await-unresolved -->

Here the operand is a name with no declaration. The [separate negative contract](../../testdata/phase1/n02-await-unresolved/case.json) expects unresolved-reference evidence containing `missing` and `unresolved`. Supplying a declared Text instead would change the failure to the type boundary above; the two diagnoses are not interchangeable.

For a grammar-recognized but unsupported form, see the [While illustration](control-flow.md#iteration-and-collection-operations). Its limitation is an instruction-lowering boundary, not a missing vocabulary word.

### Runtime failure is a separate stage

**Intentional runtime misuse — complete source shown as a negative illustration.** This retained fixture builds an App but never launches it:

<!-- example:runtime-build-without-run -->

Build does not create an invocation. Await therefore fails at generated-runtime execution with an invocation-related diagnostic. The [N07 contract](../../testdata/phase1/n07-build-is-not-run/case.json) expects empty stdout, status 1, and diagnostic text containing `Await` and `invocation`; it is not a compiler-rejection test. This is a display-only negative example, not a successful tutorial. No language stage was executed here.

**Explicit program failure — complete maintained runtime witness, not run here.**

<!-- example:paragraph-error-exit -->

First's child succeeds and its continuation shows `First\n`. The equality condition then requests an error Exit with `paragraph stopped`. The [paragraph witness](../../cmd/l/paragraph_integration_test.go) expects stdout `First\n`, stderr `paragraph stopped\n`, and status 1. The failure comes from the program's explicit Exit, not from a compiler diagnostic or an IO.Print exception. Neither the later paragraph tail nor Second's Execute operand runs.

## Compatibility and releases

No official released L version is manufactured by this documentation. Local/unreleased identity, source fingerprints, and Git base revision describe the checkout actually used. Historical test results apply only to their recorded revision and host. A dirty checkout is not fully identified by HEAD.

The maintained current contracts and protected resources define compatibility obligations. There is no inferred promise of Semantic Versioning, complete backward compatibility across unreleased changes, portable native fragments, or cross-target behavior. The retired generated-Go target and rejected Deferred/legacy loop forms must not reappear as supported options through old examples.

Native fragments and byte-preserving source APIs have intentional implementation boundaries. Source lexicon selection is distinct from target selection. Documentation for a later indexed dictionary or release requires its own reviewed source evidence; current absence is not evidence of a permanently closed standard library.

## Sources and witnesses

The maintained [error model](../error-model.md) owns category definitions and Go inspection behavior. Other primary sources are [diagnostics](../../lyger/diagnostics.go), [error kinds](../../lyger/error_kinds_test.go), [native causes](../../lyger/error_causes_test.go), [numeric resolution](../../lyger/numeric_resolution_test.go), [CLI/runtime contract](../cli-build.md), and [validation policy](../development-harness.md). The [integration ledger](../../testdata/integration/IMPLEMENTATION.md) records historical outcomes; it is not a result for this checkout.
