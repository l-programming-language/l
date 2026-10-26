# Compiler and build errors

Compiler and build diagnostics use the standard-library-only `lerrors` package.
The category identifies why an operation failed; existing diagnostic text and
native causes remain independently inspectable. Categories do not select CLI
statuses, parser alternatives, dependency retries, or runtime behavior.

## Public inspection

```go
result, err := service.Compile(source, path)
if err != nil {
    if errors.Is(err, lerrors.UnresolvedReference) {
        // A binding could not be resolved by the existing compilation process.
    }
    var classified *lerrors.Error
    if errors.As(err, &classified) {
        category := classified.Kind() // nearest classified error in the chain
        _ = category
    }
    var fileError *os.PathError
    if errors.As(err, &fileError) {
        // Inspect the original filesystem operation and path.
    }
}
_ = result
```

`errors.Is` visits the complete cause tree, so a categorized wrapper can match
both its category and a category deeper in its cause. `errors.As` finds the
first matching type in the standard traversal order. Native errors remain
available, including JSON errors, `*strconv.NumError`, `strconv.ErrRange`,
`*os.PathError`, `*exec.ExitError`, cancellation and `exec.ErrWaitDelay`.

## Categories

`Kind` is a `uint8` enum using `iota`. `String` and `Error` use indexed names;
out-of-range values format safely as `Kind(n)`. Numeric values are not a
serialization contract.

| Kind | Meaning and originating boundary |
|---|---|
| Unknown | Defined category for an unspecified failure; not a wildcard or success |
| InvalidInput | Invalid caller arguments or source-file selection, numeric overflow |
| InvalidConfiguration | Invalid compiler setup, language selection, grammar/vocabulary assets or templates |
| Syntax | Malformed L, definition or AML syntax |
| UnresolvedReference | A binding remains unresolved at an existing error boundary |
| AmbiguousReference | More than one valid binding |
| DuplicateDeclaration | Publication rejects a duplicate declaration |
| TypeMismatch | Resolved values or arguments have incompatible types |
| Cycle | Import, alias, action or definition expansion cycle |
| Unsupported | Recognized construct, dependency or host that is not supported |
| IO | Failed filesystem operation, retaining its native cause |
| Tool | Cargo/toolchain startup, execution, interruption or invalid output |
| Internal | Inconsistent compiler storage, provenance or generated structures |

A nil error interface is the only successful result. `Unknown` itself is an
error. A nil `*lerrors.Error` is safe to inspect (`Error` is `"<nil>"`, `Kind` is
`Unknown`, `Unwrap` is nil, `Is` is false), but storing that pointer in an error
interface does not make the interface nil.

## Origin and context responsibilities

Classify once where the failure's meaning becomes known:

```go
return lerrors.TypeMismatch.Errorf("requires %s, got %s", expected, actual)
return lerrors.IO.Errorf("read source %s: %w", path, err)
```

`Kind.Errorf` retains the complete `fmt.Errorf` result as its cause, including
multiple `%w` operands. Pass dynamic diagnostics as `Errorf("%s", message)` so
literal percent characters cannot become formatting directives.

`Kind.Wrap(nil, message)` returns nil. With an empty message, `Wrap` keeps the
cause's exact display text. With a nonempty message, that message is the complete
diagnostic; the cause is not appended again:

```go
return lerrors.IO.Wrap(err, existingOwnershipDiagnostic)
```

Ordinary context uses `fmt.Errorf("phase: %w", err)` without reclassification.
Source context belongs to `SourceSpan.wrapError`; definition context belongs to
`DefinitionTemplate.failure`; instruction and expanded-action call-site context
belong to `instruction.wrapError`. `statementDiagnostic` still restores literal
cache values only in display text and keeps the original cause. The compilation
retains its first error.

Grammar alternatives preserve their existing sorted aggregate messages and keep
candidate causes with `errors.Join`. An empty alternative set is still a Syntax
error. Optional parsing, lookahead, longest-match selection, pending bindings,
transactional parsing and dependency scheduling are unchanged.

## Boundary behavior

Numeric candidates are parsed once in `loadPrimitive` after Text, Truth and
identifier checks. Range failures are InvalidInput with the original numeric
cause. Direct primitive loading returns that failure immediately. Source tokens
retain it provisionally after native-type and configured-vocabulary lookup miss,
allowing source-defined names to resolve through the existing dependency worklist.
Only when resolution makes no progress does the compiler report a remaining
numeric candidate, preserving its original `tokenize` diagnostic and cause.
Declared but unresolved names and explicit member names remain reference errors.
Ordinary nonnumeric primitive misses retain their lexical fallback.

Source-file validation returns non-NotExist filesystem failures as errors. Missing sources keep their original diagnostic and native cause.
Ownership checks distinguish failed inspection (IO) from an inspected, unrelated
layout (InvalidInput). Interrupted Cargo builds keep cancellation and child/cleanup
causes together, while their diagnostic and CLI status remain unchanged.

## Scope and limits

Classified returned errors cover Lyger, checked logeme source
generation, source-file utilities and the CLI/build coordinator. Help retains
`flag.ErrHelp`; the unsupported-host sentinel retains its identity. CLI statuses
remain 0 for help/success, 2 for invalid invocation/unsupported host and 1 for
initialization, compilation, generation or build failure.

The independent harness and generated process runtime do not depend on
`lerrors`. The [execution contract](cli-build.md) independently governs App launch, Await, raw streams and process cleanup.

Unchecked regex/metadata/math indexing and UUID paths, suppressed recursive-parser errors,
and non-error-returning convenience loaders retain their existing behavior.
There is no blanket recovery, severity system, serialization API, recursive
classifier, status mapping, retry policy or centralized message catalog.

Follow the [development guide](development-harness.md) for validation commands
and the current local-validation pause.
