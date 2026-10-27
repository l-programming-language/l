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

A recognized While clause can reach an Unsupported instruction boundary. A missing name reaches a reference boundary. A recognized structural category such as Attribute does not by itself establish a runtime constructor; documentation must distinguish compiler representation from executable values.

See the [type reference](values.md) and [control-flow limits](control-flow.md) for current supported operations.

### Rejected operands: a type error and a missing name

These complete fixtures demonstrate expected compiler rejections.

<!-- example:invalid-await-text -->

The backtick literal is Text. Await requires a process target, so this is a TypeMismatch at compilation, before any application can be launched. The [fixture contract](../../testdata/phase1/n01-await-text/case.json) expects a diagnostic containing `await` and `app` and no new published executable.

<!-- example:invalid-await-unresolved -->

Here the operand is a name with no declaration. The [separate negative contract](../../testdata/phase1/n02-await-unresolved/case.json) expects unresolved-reference evidence containing `missing` and `unresolved`. Supplying a declared Text instead would change the failure to the type boundary above; the two diagnoses are not interchangeable.

For a grammar-recognized but unsupported form, see the [While illustration](control-flow.md#iteration-and-collection-operations). Its limitation is an instruction-lowering boundary, not a missing vocabulary word.

### Runtime failure is a separate stage

**Intentional runtime misuse — complete source shown as a negative illustration.** This retained fixture builds an App but never launches it:

<!-- example:runtime-build-without-run -->

Build does not create an invocation. Await therefore fails at generated-runtime execution with an invocation-related diagnostic. The [N07 contract](../../testdata/phase1/n07-build-is-not-run/case.json) expects empty stdout, status 1, and diagnostic text containing `Await` and `invocation`; it is not a compiler-rejection test. This is a display-only negative example, not a successful tutorial. No language stage was executed here.

**Explicit program failure.**

<!-- example:paragraph-error-exit -->

First's child succeeds and its continuation shows `First\n`. The equality condition then requests an error Exit with `paragraph stopped`. The [paragraph witness](../../cmd/l/paragraph_integration_test.go) expects stdout `First\n`, stderr `paragraph stopped\n`, and status 1. The failure comes from the program's explicit Exit, not from a compiler diagnostic or an IO.Print exception. Neither the later paragraph tail nor Second's Execute operand runs.

## Compatibility and releases

L uses `vMAJOR.MINOR.PATCH` release tags. Backward compatibility is not guaranteed before `v1.0.0`; consult the [versioning policy and release notes](../releases/index.md#versioning-and-compatibility) before upgrading. Historical test results apply to their recorded revision and host; local previews report their own source identity.

Native fragments remain target-specific. Generated Go, Deferred and three-part loop forms are unsupported. New dictionaries or supported forms are documented with their release.

## Sources and witnesses

The maintained [error model](../error-model.md) owns category definitions and Go inspection behavior. Other primary sources are [diagnostics](../../lyger/diagnostics.go), [error kinds](../../lyger/error_kinds_test.go), [native causes](../../lyger/error_causes_test.go), [numeric resolution](../../lyger/numeric_resolution_test.go), [CLI/runtime contract](../cli-build.md), and [validation policy](../development-harness.md). The [integration ledger](../../testdata/integration/IMPLEMENTATION.md) records historical outcomes; it is not a result for this checkout.
