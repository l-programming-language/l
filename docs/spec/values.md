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

<!-- word:type:text -->

Represent a sequence of bytes without an implicit newline or Unicode normalization.

**Usage and example.** Backtick literals initialize Text; the [public-values program](../../demo/paragraphs/public-values.l) stores and shows Greeting. The escape illustration below preserves its precise byte rules.

Backticks delimit Text. Interior bytes are retained, except that a backslash immediately before an escaped backtick is removed when the literal is unquoted. This is not general escape decoding: backslash followed by `n` remains two characters. Text has no implicit newline. `Show` writes it exactly; imported `IO.Print` appends one LF. See the [IO reference](../stdlib/eng/io.md).

Text preserves UTF-8, invalid UTF-8 octets, NUL, whitespace, and markup punctuation. Native encoding preserves these bytes without replacement characters. Browser displays cannot faithfully represent every octet; use the original source bytes for byte-sensitive examples.

Text supports equality with Text. General ordering, numeric conversion, indexing, concatenation operators, code-point iteration, and Unicode normalization are not currently exposed as L operations. Native definitions can supply target-specific operations.

<!-- example:lexical-text-escapes -->

The first Text contains a backtick; the second retains the two bytes `\n`; the third retains Unicode, braces, and an interior `<DEF>` marker as data. Successive Show calls write adjacent text. This fragment does not define a Bytes constructor or an escape for arbitrary octets.

## Number

<!-- word:type:number -->

Represent a finite binary64 number for typed storage, comparison and supported rendering.

**Usage and example.** `Define "Count" as 7` in [public-values](../../demo/paragraphs/public-values.l) supplies a Number. The printing fragment below demonstrates the formatter; it requires its App context.

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

`IO.Print` uses shortest round-tripping binary64 formatting, corresponding to `strconv.FormatFloat(value, 'g', -1, 64)`. One million prints as `1e+06`, followed by LF.

<!-- example:values-number-printing -->

The three calls print `3.5`, `-12.5`, and `1e+06`, each followed by LF when the App runs. This fragment supplies only the App body. The [first program](../guide/first-program.md) shows the lifecycle and output handling needed to execute it.

## Truth

<!-- word:type:truth -->

Represent the Boolean values true and false without implicit truthiness conversion.

**Usage and example.** The complete conditional example below compares a declared Truth and joins Worker only on its selected branch.

The literals are lowercase `true` and `false`. A bare condition requires Truth; equality requires another Truth. Number and Text have no implicit truthiness. The grammar's `?~?` notation is a type matcher, not a source literal.

Configured `not`/negative forms express negation in conditions, not a general prefix operator. Joined Boolean conditions are currently rejected by outer instruction lowering; no short-circuit behavior is defined.

<!-- example:values-truth-condition -->

`enabled` receives a Truth. Comparing it with `true` selects `await Worker`; the following declaration snapshots the finalized Output. Replacing `true` with backtick Text changes the type and does not create truthiness.

## Thing

<!-- word:type:thing -->

Retain a value's concrete runtime type at an explicitly broad typed boundary.

**Usage and example.** ``Define "boxed" as a New Thing of `first`.`` followed by a checked Text read is the [runtime-value witness](../../lyger/runtime_values_test.go). The Print illustration below shows that a broad parameter still selects a supported concrete renderer.

Thing preserves a contained runtime type and checks typed reads. It does not coerce values into Text. IO.Print selects typed rendering definitions for Text, Number, Truth and Null:

<!-- example:values-thing-printing -->

The public Print action composes private rendering and writing Functions. This fragment supplies an App body; declaration alone does not run it.

Thing does not provide dynamically typed arithmetic, unrestricted reflection, or printing for every application or collection value. Its usable arguments depend on the definition and native consumer.

## Byte and Bytes

### Byte

<!-- word:type:byte -->

Name the byte-oriented lexical category whose supported owned storage form is Bytes.

**Usage and limitation.** Use the plural [Bytes](#bytes) constructor for an owned byte sequence, then read an element as Number. A standalone Byte scalar constructor or integer-width arithmetic API is not established. The [byte excerpt](../examples/word-values.l) reads the `a` octet as 97.

### Bytes

<!-- word:type:bytes -->

Own a sequence of unchanged bytes copied from Text.

**Usage and example.** `New Bytes of` constructs the buffer; a checked Number read selects a zero-based octet. The complete [value program](../examples/word-values.l) contains this excerpt:

<!-- example:word-values-bytes -->

Indices must be nonnegative integers within bounds. Bytes mutation is not exposed, and byte indexing is not Unicode character indexing.

## Pointer

<!-- word:type:pointer -->

Hold a managed shared reference to a typed runtime value.

**Usage and example.** The [complete value program](../examples/word-values.l) copies a Pointer handle, changes the shared target and reads it as Text. The focused excerpt below preserves those exact source lines.

<!-- example:word-values-pointer -->

The Pointer owns a copy of the initial value. Both handles share that target, so this prints `second`; `message` remains `first`. Target types are checked. Pointer and capture-environment writes reject ownership cycles. Reads release their checked borrow before a nested call or Await. The last owner releases the target; there is no pointer arithmetic or user-managed deallocation.

## Null and internal states

### Null

<!-- word:type:null -->

Represent an explicitly initialized absence value, distinct from an unassigned binding.

**Usage and example.** The [value program](../examples/word-values.l) defines nothing as Null and shows `null`. IO.Print renders its native Null operand as `<nil>` instead; these policies are intentionally distinct.

`Define "nothing" as Null.` initializes an absence value; Show writes `null`. A typed Null read still evaluates its source and checks its type, index or key. Null does not replace definite assignment. The separate native-definition Null atom displays `<nil>`.

Compiler labels are separate from public types: `ID`, `None`, and `Unknown` track resolution; `Lexeme` stores lexical definitions; invalid sentinels indicate compiler state. These implementation labels are not additional application words. [Error](#error) is a separate runtime carrier, not an internal sentinel.

## Error

<!-- word:type:error -->

Carry a runtime failure value that can be stored, passed and retained in typed containers.

**Usage and example.** The [complete value program](../examples/word-values.l) uses ``Define "problem" as an Error of `failure`.`` and puts it in a List of Error. Showing that list writes `[failure]`; direct Show of an Error is unsupported. This is a value constructor, distinct from inspecting an [App's Error](../reference/builtins/lifecycle.md#error).

The value retains its failure information when transferred through Functions. Extracting an absent App Error rejects instead of manufacturing this carrier. Error values do not provide general source exceptions or automatic rollback.

## Conversion and aliasing boundaries

Matching Text, Number, and Truth values can be compared; ordering is Number-only. No implicit cross-scalar coercion is promised. Definition slots separately check their required argument kinds.

The repeated-activation demo contrasts an App alias with a Text snapshot:

<!-- example:values-output-snapshot -->

`Worker` and `Repeat` refer to the same App. After the first join, `snapshot` holds that invocation's Output. Running the App again does not update `snapshot`: `Show snapshot` reads the stored scalar, while `Show Repeat's Output` reads the later invocation. Both invocations happen to print the same message. This excerpt depends on the demo's earlier declarations.

## Sources and witnesses

Primary sources: [value tags](../../lyger/value_kinds.go), [primitive storage and loading](../../lyger/types.go), [type spellings](../../lyger/utilities.go), [native encoding](../../lyger/literal_atoms.go), [definition encoding](../../lyger/definition_bindings.go), and [typed orchestration](../../lyger/orchestration_plan.go). Existing witnesses: [Number encoding](../../lyger/definition_number_test.go), [numeric resolution](../../lyger/numeric_resolution_test.go), [raw Text](../../lyger/raw_text_test.go), [Text encoding](../../lyger/definition_text_encoding_test.go), [IO numeric formatting](../../lyger/io_numeric_test.go), [stored-value invariants](../../lyger/stored_values_test.go), and [scalar conditions](../../lyger/orchestrator_test.go). Example labels distinguish complete programs from fragments; website validation does not execute L.
