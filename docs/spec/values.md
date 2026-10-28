# Primitive Types

| Type | Meaning | Current use |
| --- | --- | --- |
| [Text](#text) | A sequence of bytes, including whitespace. | Write a backtick-delimited literal, such as `` `hello` ``; store, compare, or print it. |
| [Number](#number) | A finite binary64 number, such as `7` or `-12.5`. | Store, compare, or print it. Integer spellings use the same number type. |
| [Truth](#truth) | A Boolean value: `true` or `false`. | Store, compare, print, or use it as a condition. |
| [Thing](#thing) | A value retaining its concrete runtime type. | Hold different value categories at an explicitly broad boundary. |
| [Byte / Bytes](#byte-and-bytes) | An owned sequence of bytes. | Construct from Text and read numbered bytes. |
| [Pointer](#pointer) | A managed reference to a typed value. | Copy handles to share a target; checked reads and writes. |
| [Null](#null-and-internal-states) | An explicit absence value. | Initialize a value with `Null`; distinct from an unassigned slot. |

Primitive types describe data; [lexical types](lexical-types.md) describe words. [Collections](collections-and-categories.md) and [Functions](applicational-types.md#functions-properties-and-operations) use the same ownership rules.

## Declarations, assignment, and reading values

A declaration names a value: `Define "count" as 7.` Runtime assignments take effect in source order, and every continuing path must assign a value before reading it. Declaring a type does not promise an implicit initial value.

The public-values program gives Text, Number, and Truth explicit initial values, then displays them with Show:

<!-- example:public-values -->

Assignment, Function arguments/results and captures snapshot ordinary values, recursively through collections; embedded Pointers retain their shared target. App handles retain process identity. Stored Output keeps its invocation's bytes after reruns. See [capture environments](names-and-scope.md#runtime-storage-and-definite-assignment).

## Text

Backticks delimit Text. Interior bytes are retained, except that a backslash immediately before an escaped backtick is removed when the literal is unquoted. This is not general escape decoding: backslash followed by `n` remains two characters. Text has no implicit newline. `Show` writes it exactly; imported `IO.Print` appends one LF. See the [IO reference](../stdlib/eng/io.md).

Text preserves UTF-8, invalid UTF-8 octets, NUL, whitespace, and markup punctuation. Native encoding preserves these bytes without replacement characters. Browser displays cannot faithfully represent every octet; use the original source bytes for byte-sensitive examples.

Text supports equality with Text. General ordering, numeric conversion, indexing, concatenation operators, code-point iteration, and Unicode normalization are not currently exposed as L operations. Native definitions can supply target-specific operations.

<!-- example:lexical-text-escapes -->

The first Text contains a backtick; the second retains the two bytes `\n`; the third retains Unicode, braces, and an interior `<DEF>` marker as data. Successive Show calls write adjacent text. This fragment does not define a Bytes constructor or an escape for arbitrary octets.

## Number

Number uses finite IEEE-754 binary64 (`float64` in Lyger, `f64` in Rust). It is not an arbitrary-precision integer, and L has no source integer-width family.

The number parser accepts these forms. Source token boundaries also apply; parser acceptance alone does not make a spelling a complete standalone token.

| Source spelling | Represented consequence |
| --- | --- |
| `3`, `+3`, `3.0` | The same binary64 value |
| `-12.5` | Negative finite fraction |
| `1e6` | One million |
| `0x1.8p+1` | Three |
| `1_000` | One thousand |
| `9007199254740993` | Rounded to the representable value 9007199254740992 |
| `5e-324` | Smallest positive subnormal |
| `1e-4000` | Underflows to positive zero |
| `-1e-4000` | Underflows to negative zero |

Positive and negative zero retain their sign through native encoding. Overflow, such as `1e999`, rejects with its range error preserved. Numeric-looking declared words can resolve as names. Although the lower parser recognizes nonfinite spellings, native scalar emission and typed definition binding reject NaN and Infinity.

Number supports equality and order comparisons with Number. Text/Number comparisons do not coerce. The English source grammar has no general arithmetic operator expressions; native logeme calculations are a separate interface.

`IO.Print` uses shortest round-tripping binary64 formatting, corresponding to `strconv.FormatFloat(value, 'g', -1, 64)`. One million prints as `1e+06`, followed by LF. Native integer operands in explicit native definitions retain their own formatting; they do not introduce L integer types.

<!-- example:values-number-printing -->

The three calls print `3.5`, `-12.5`, and `1e+06`, each followed by LF when the App runs. This fragment supplies only the App body. The [first program](../guide/first-program.md) shows the lifecycle and output handling needed to execute it.

## Truth

The literals are lowercase `true` and `false`. A bare condition requires Truth; equality requires another Truth. Number and Text have no implicit truthiness. The grammar's `?~?` notation is a type matcher, not a source literal.

Configured `not`/negative forms express negation in conditions, not a general prefix operator. Joined Boolean conditions are currently rejected by outer instruction lowering; no short-circuit behavior is defined.

<!-- example:values-truth-condition -->

`enabled` receives a Truth. Comparing it with `true` selects `await Worker`; the following declaration snapshots the finalized Output. Replacing `true` with backtick Text changes the type and does not create truthiness.

## Thing

Thing preserves a contained runtime type and checks typed reads. It does not coerce values into Text. Native-definition slots remain a separate interface: IO accepts Text, Number and Truth through its Thing slot:

<!-- example:values-thing-printing -->

The arguments retain their original types. There is no `Thing(...)` cast: Print binds each supplied value and its native implementation handles the type. This fragment supplies an App body; declaration alone does not run it. The public Print action uses a private Log implementation.

Thing does not provide dynamically typed arithmetic, unrestricted reflection, or printing for every application or collection value. Its usable arguments depend on the definition and native consumer.

## Byte and Bytes

``Define "buffer" as a New Bytes of `raw`.`` copies Text bytes. `Define "octet" as a New Number from buffer at 1.` reads the zero-based byte as Number. Indices must be nonnegative integers within bounds; Bytes mutation is not exposed. Byte is vocabulary for byte-oriented storage, not an additional numeric-width family.

## Pointer

```l
Define "message" as `first`.
Define "handle" as a New Pointer of message.
Define "shared" as handle.
Put `second` in shared.
Define "selected" as a New Text from handle.
Show selected.
```

The Pointer owns a copy of the initial value. Both handles share that target, so this prints `second`; `message` remains `first`. Target types are checked. Pointer and capture-environment writes reject ownership cycles. Reads release their checked borrow before a nested call or Await. The last owner releases the target; there is no pointer arithmetic or user-managed deallocation.

## Null and internal states

`Define "nothing" as Null.` initializes an absence value; Show writes `null`. A typed Null read still evaluates its source and checks its type, index or key. Null does not replace definite assignment. The separate native-definition Null atom displays `<nil>`.

Compiler labels are separate from public types: `ID`, `None`, and `Unknown` track resolution; `Lexeme` stores lexical definitions; invalid sentinels indicate compiler state. An Error value can be created from Text or read from a failed App, stored and passed like other values. Reading an absent App Error rejects; it does not fabricate an Error.

## Conversion and aliasing boundaries

Matching Text, Number, and Truth values can be compared; ordering is Number-only. No implicit cross-scalar coercion is promised. Definition slots separately check their required argument kinds.

The repeated-activation demo contrasts an App alias with a Text snapshot:

<!-- example:values-output-snapshot -->

`Worker` and `Repeat` refer to the same App. After the first join, `snapshot` holds that invocation's Output. Running the App again does not update `snapshot`: `Show snapshot` reads the stored scalar, while `Show Repeat's Output` reads the later invocation. Both invocations happen to print the same message. This excerpt depends on the demo's earlier declarations.

## Sources and witnesses

Primary sources: [value tags](../../lyger/value_kinds.go), [primitive storage and loading](../../lyger/types.go), [type spellings](../../lyger/utilities.go), [native encoding](../../lyger/literal_atoms.go), [definition encoding](../../lyger/definition_bindings.go), and [typed orchestration](../../lyger/orchestration_plan.go). Existing witnesses: [Number encoding](../../lyger/definition_number_test.go), [numeric resolution](../../lyger/numeric_resolution_test.go), [raw Text](../../lyger/raw_text_test.go), [Text encoding](../../lyger/definition_text_encoding_test.go), [IO numeric formatting](../../lyger/io_numeric_test.go), [stored-value invariants](../../lyger/stored_values_test.go), and [scalar conditions](../../lyger/orchestrator_test.go). Example labels distinguish complete programs from fragments; website validation does not execute L.
