# Primitive values

L's current executable scalar path supports Text, Number, and Truth. The compiler contains additional storage tags and source type names; their existence must not be mistaken for equally complete source construction, operations, or native execution.

## Declarations, assignment, and reading values

A declaration names a value, for example `"message" is ` followed by backtick Text, or `Define "count" as 7.` A runtime assignment is emitted at its ordered instruction position. A runtime declaration does not receive a hidden default merely because a native Rust variable needs initialization. Internal type-token constructors allocate empty/zeroed storage wrappers, but that implementation detail is not evidence that every recognized nominal category has a complete executable default-value contract.

Static values can be captured by source-defined definitions. Runtime values need storage in the owning activation and must be definitely assigned along every continuing path before a read. Reading an application's Output is an operation on an invocation; copying it into a scalar creates a snapshot. That snapshot does not become a fresh member read when the application runs again.

The maintained public-values program demonstrates explicit Text, Number, and Truth initial values and their Show behavior:

<!-- example:public-values -->

Aliases retain canonical binding identity. Scalar assignment and aliases therefore describe different relationships: assignment copies the value required at that instruction, while `Use value as "Alias"` preserves a reference to the source binding. The compiler copies scalar wrappers for safe internal use; that implementation detail is not a general user-visible reference/pointer API.

## Text

Text is introduced by backticks. Ordinary interior bytes are retained; when the source literal is unquoted, a backslash immediately before a backtick is removed so that the backtick becomes data. This is not general escape decoding: the two characters backslash and `n` remain those characters. Text has no implicit newline. `Show` writes supplied Text exactly; imported `IO.Print` writes it and appends one LF. The [IO reference](../stdlib/eng/io.md) explains the deliberate distinction.

Valid UTF-8, embedded NUL, whitespace, markup punctuation, and invalid UTF-8 octets have explicit preservation witnesses. Native orchestration storage uses `Vec<u8>`. For a typed native definition, valid UTF-8 can be emitted as a Rust string literal; invalid UTF-8 uses a byte string. Both paths preserve the represented bytes rather than applying replacement characters.

Text equality compares matching Text operands. General Text ordering, automatic numeric conversion, indexing, concatenation operators, code-point iteration, and a public Unicode normalization API are not established by the current source surface. Native fragments may perform target-specific operations, but those operations remain native code.

An escaped quote can affect source framing. Do not infer that the characters backslash and `n` in ordinary Text automatically produce LF. Use the canonical fixture bytes for examples that depend on control characters. A browser code display cannot faithfully show every invalid octet; byte-sensitive witnesses must retain their original downloadable/source representation.

The following explanatory fragment makes the distinction visible:

<!-- example:lexical-text-escapes -->

The first Text contains a backtick; the second retains the two bytes `\n`; the third retains Unicode, braces, and an interior `<DEF>` marker as data. `Show` does not add line endings, so successive writes would be adjacent. The quote and Text-emission paths are statically inspected and have existing focused witnesses; this documentation-only combination has not been compiled or executed. It does not demonstrate a Bytes constructor or a source-level escape for arbitrary octets.

## Number

Number is binary64: Lyger stores a Go `float64`, and Rust scalar emission uses `f64`. It is not an arbitrary-precision integer type, and no source integer-width family is implemented by the Number spelling.

The numeric parser uses `strconv.ParseFloat` with 64-bit precision. Existing encoding witnesses include decimal integers and fractions, signs, exponent notation, hexadecimal floating-point notation, and permitted underscores:

| Source spelling | Represented consequence |
| --- | --- |
| `3`, `+3`, `3.0` | The same binary64 value |
| `-12.5` | Negative finite fraction |
| `1e6` | One million |
| `0x1.8p+1` | Three |
| `1_000` | One thousand |
| `9007199254740993` | Rounded to the representable value 9007199254740992 |
| `5e-324` | Smallest positive subnormal in the retained witness |
| `1e-4000` | Underflows to positive zero |
| `-1e-4000` | Underflows to negative zero |

Positive and negative zero retain their IEEE-754 sign through native encoding. Overflow such as `1e999` is an invalid numeric input with its underlying range error preserved. A numeric-looking declared word can still resolve as a name before a provisional numeric error is reported.

The lower primitive parser can recognize nonfinite spellings, but the native scalar emitter rejects nonfinite Number. Typed definition binding rejects nonfinite input at its own Syntax boundary. This is a staged limitation, not a supported NaN/Infinity language feature.

Equality and order comparison are supported for Number operands. Mixed Text/Number comparisons do not coerce. The configured English surface does not provide a general arithmetic operator expression language; D/E native calculations are a separate logeme capability. No source guarantee of overflow arithmetic, integer division, numeric casts, or operator precedence should be inferred from the representation.

`IO.Print` preserves the established shortest round-tripping binary64 formatting policy corresponding to `strconv.FormatFloat(value, 'g', -1, 64)`. For example, one million prints as `1e+06`, and a newline follows. Rust native integer operands supplied deliberately through native definitions retain their own Display contract; they are not new L integer types.

To put Number values in an App body, use numeric operands without Text delimiters:

<!-- example:values-number-printing -->

These three print calls are a small documentation excerpt of the source shapes exercised by `TestSourceLogNumberSourceParity`. Its existing native witness expects `3.5`, `-12.5`, and `1e+06`, each followed by LF. The fragment only describes the App body; it has no `Execute`/`Show` steps and produces no output merely by being declared. The test builds and runs the generated App separately; neither the test nor this fragment was run here. The [first program](../guide/first-program.md) shows complete lifecycle and output handling.

## Truth

The primitive source literals are lowercase `true` and `false`. They map to Go `bool` and Rust `bool`. Internal helpers also recognize wrapped diagnostic spellings in narrow contexts, but `?~?` in the grammar is a type matcher, not the required source spelling.

A bare conditional requires Truth. Equality requires another Truth. Number and Text do not acquire truthiness. Negation is represented by configured `not`/negative forms in the appropriate condition grammar; it is not a general prefix operator accepted anywhere. Multiple joined conditions are currently rejected by outer instruction lowering, so no short-circuit language contract is established.

The maintained conditional-Await program stores a Truth and compares it with `true`:

<!-- example:values-truth-condition -->

`enabled` receives a Truth, not the Text spelling of that word. Its equality comparison selects the `await Worker` branch; the following declaration snapshots the finalized Output. The existing LANG-13 native witness expects `condition` followed by one LF. Replacing `true` with backtick Text would change the operand type and would not create implicit truthiness.

## Thing

Thing is a broad primitive representation and an accepted typed native parameter category. In a definition slot, it admits supported scalar argument kinds and preserves their native encoding. This permits IO's native implementation to receive Text, Number, and Truth without coercing them all to Text.

Thing does not establish dynamically typed arithmetic, unrestricted object reflection, or a universal printable source value. A `Thing` storage tag can carry host data internally; executable source support still depends on the actual value, context, and native consumer. User code should pass a supported scalar to a documented operation rather than relying on Go's `interface{}` representation.

For example, the public `IO.Print` action ultimately supplies the selected scalar to its native implementation's `Thing` slot:

<!-- example:values-thing-printing -->

The three arguments remain Text, Number, and Truth respectively. There is no `Thing(...)` cast: `Print` binds each supplied value and the native implementation handles its type. This documentation-only App-body fragment is based on the inspected IO definition, the source-number witness, and the native typed-dispatch witness. That last witness binds arguments directly at the definition boundary, so it is not an end-to-end run of this exact L source. No execution is claimed for the fragment, and no lifecycle instruction is included. Only the public `Print` action is used; the dictionary's `Log` implementation remains private.

## Byte and Bytes

The English vocabulary contains the noun `Byte` and plural inflection `Bytes`; native type recognition explicitly recognizes `Bytes`. The internal primitive representation is `[]byte`, and byte wrappers are copied when bindings are copied. Those facts do not establish a source byte-literal syntax, a numeric 8-bit scalar, indexing, arithmetic, or conversion APIs.

Ordinary Text already preserves raw bytes through its supported path. The Byte/Bytes spelling and construction discrepancy is recorded in the audit as partial/unresolved support. Do not replace Text with a guessed `Byte` declaration to handle invalid UTF-8.

**Example boundary:** use the supported [Text example](#text) for source text. No byte-literal or byte-mutation example is supplied here because the checkout does not establish that public API.

## Pointer

Pointer is a configured noun and an internal primitive tag with string payload, but the native source type recognizer and executable scalar emitter do not expose a complete Pointer path. There is no documented public address-of/dereference syntax, pointer arithmetic, allocation API, or ownership inference. Treat it as a recognized, unresolved category rather than an operational memory feature.

**Example boundary:** there is no source-backed public pointer operation to demonstrate. Native Rust pointer syntax belongs to an explicit native definition, not to an L code example.

## Null and internal states

Null is recognized as a native type name and by the grammar's null type pattern. It is not one of the ordinary literal branches in `loadPrimitive`, nor one of the executable scalar types. Typed native definition arguments can explicitly supply Null; the emitter constructs a local display atom that prints `<nil>`. That bounded native binding contract does not imply Rust null pointers or an optional-value source type system.

`ID`, `None`, and `Unknown` represent compiler resolution states. `Error` represents the process error member in runtime instructions, not a general thrown exception object. `Lexeme` is lexical storage, while invalid sentinels indicate compiler state. None of these should be offered as user constructors simply because an enum contains them.

**Example boundary:** the existing Null witness supplies a typed argument directly to the native-definition binder. It does not establish an L `null` literal or an optional-value declaration. There is therefore no ordinary source construction example for Null or the internal state tags.

## Conversion and aliasing boundaries

No implicit cross-scalar coercion is promised. The typed instruction planner checks operands before native emission; definition slots check their requested kind separately. Matching Text/Number/Truth values can be compared; ordering is Number-only. Assignments produce owned Text snapshots where needed, while observation and comparison can borrow storage internally. That is compiler-managed Rust encoding, not a user-visible borrow checker or cloning directive.

The end of the repeated-activation demo contrasts an App alias with a Text snapshot:

<!-- example:values-output-snapshot -->

`Worker` aliases the already-declared `Repeat` App, so both names address the same application identity. After the first join, `snapshot` receives that invocation's Output. Running the App again does not turn `snapshot` into a live member read: `Show snapshot` reads the retained scalar, while `Show Repeat's Output` reads the later invocation. The demo prints the same message on both invocations, so identical displayed text alone cannot prove the distinction; the [stored runtime-value witnesses](../../lyger/orchestrator_test.go) also check the snapshot representation. This excerpt needs the demo's earlier declarations. Its full program has existing native witnesses that were not rerun here.

## Sources and witnesses

Primary sources: [value tags](../../lyger/value_kinds.go), [primitive storage and loading](../../lyger/types.go), [type spellings](../../lyger/utilities.go), [native encoding](../../lyger/native_encoding.go), [definition encoding](../../lyger/definition.go), and [typed orchestration](../../lyger/orchestration_plan.go). Existing witnesses: [Number encoding](../../lyger/definition_number_test.go), [numeric resolution](../../lyger/numeric_resolution_test.go), [raw Text](../../lyger/raw_text_test.go), [Text encoding](../../lyger/definition_text_encoding_test.go), [IO numeric formatting](../../lyger/io_numeric_test.go), [stored-value invariants](../../lyger/stored_values_test.go), and [scalar conditions](../../lyger/orchestrator_test.go). No new native execution is claimed.
