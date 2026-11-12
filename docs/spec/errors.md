# Errors, diagnostics, and compatibility

Compiler diagnostics, native build failures, and standalone application output are distinct channels and stages.

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
| Standalone application | Process status, output or native failure observed by its caller |

Errors retain original causes where the maintained interface supports them. A failed prerequisite does not prove later stages passed or failed.

## Compiler categories

The Go `lerrors` package classifies InvalidInput, InvalidConfiguration, Syntax, UnresolvedReference, AmbiguousReference, DuplicateDeclaration, TypeMismatch, Cycle, Unsupported, IO, Tool, Internal, and Unknown. Numeric category values are not a serialization contract. Unknown is an error category, not success or a wildcard.

These are compiler/API inspection categories for Go callers. L does not thereby gain `try`, `catch`, thrown exceptions, or an imported standard-library error module. Runtime Error inspection is unavailable.

Grammar alternatives can aggregate candidate causes while preserving sorted diagnostics. A source numeric overflow may remain provisional until lexical resolution is exhausted. Direct definition Number binding has its own Syntax boundary. Ordinary numeric input overflow and nonfinite native lowering have different classified boundaries. Do not collapse these into one invented source exception.

## Source attribution

Diagnostics retain source file, byte-based line/column position, and original paragraph/source ownership. Native definition failures include definition/parameter context. Expanded actions retain call-site attribution. Imported source belongs to its provider even when the calling program is elsewhere.

Incidental UUIDs may be normalized for a diagnostic view. They must not be normalized in application output, canonical input, expected bytes, or generated artifacts merely to make a comparison pass. Original literal bytes and instruction order remain authoritative.

## CLI statuses and streams

The compiler CLI returns 0 for help and successful builds, 2 for invalid invocation or unsupported host, and 1 for initialization, compilation, generation, native build, publication, or output-write failure. Compiler diagnostics go to stderr. Successful build stdout contains the published executable path and newline.

The standalone App has its own process status and streams. Building it does not establish execution success. Runtime lifecycle, capture and cleanup policies are absent; see [execution](execution.md).

## Unsupported versus unresolved

While has configured vocabulary and grammar rules but no executable loop implementation. A missing name reaches a reference boundary. A recognized structural category such as Attribute does not by itself establish a runtime constructor. Documentation must distinguish compiler representation from executable values.

See the [type reference](values.md) and [control-flow limits](control-flow.md) for current supported operations.

### Rejected operands: a type error and a missing name

A resolved operand with the wrong type and an unresolved name are different semantic failures. Preserve the selected word, typed input and original source attribution in either diagnostic. Text-only Print does not gain implicit formatting to accept another type.

### Runtime failure is a separate stage

The current executable has no L process runtime. Runtime-dependent source effects reject during compilation instead of building a program that later performs lifecycle checks. Observe actual standalone execution separately from compilation; never relabel a failed or unrun stage as passing.

## Compatibility and releases

L uses `vMAJOR.MINOR.PATCH` release tags. Backward compatibility is not guaranteed before `v1.0.0`. Consult the [versioning policy and release notes](../releases/index.md#versioning-and-compatibility) before upgrading. Historical test results apply to their recorded revision and host. Local previews report their own source identity.

Native fragments remain target-specific. Generated Go, Deferred and three-part loop forms are unsupported. New dictionaries or supported forms are documented with their release.

## Sources and witnesses

The maintained [error model](../error-model.md) owns category definitions and Go inspection behavior. Other primary sources are [diagnostics](../../lyger/diagnostics.go), [error kinds](../../lyger/error_kinds_test.go), [native causes](../../lyger/error_causes_test.go), [numeric resolution](../../lyger/numeric_resolution_test.go), [build contract](../cli-build.md), and [validation policy](../development-harness.md). Historical outcomes are not results for this checkout.
