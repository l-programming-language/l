# Formatting

The English `format` dictionary supplies the shared Number formatter used by [IO.Print](io.md#print) and [Runtime](runtime.md#format). It is an explicitly importable implementation dictionary: its native-signature Function supports word authors and the compiler, rather than adding implicit application vocabulary.

## Import and calling forms

Import the dictionary for qualified access:

```l
Use Formatting from `format`.
```

In a Function with a native `f64` input named Value and a `Vec<u8>` result, `a New Formatting's Formatter of Value` constructs the formatted bytes. The member form ``Use Formatting's Format from `format` as "Format".`` binds the same definition locally. Formatter is Format's actor form, not a second word.

| Task | Word or composition |
| --- | --- |
| Obtain Number bytes without writing | [Formatting.Format](#format) |
| Print a scalar to stdout with LF | [IO.Print](io.md#print) |
| Deliver already formatted bytes to a supplied writer | [IO.Send](io.md#send) |
| Display a supported value from the orchestrator | [Show](../../reference/builtins/lifecycle.md#show) |

## Public and private names

Format is the dictionary's one public root. Its declared family is Format, Formats, Formatted, Formatting and Formatter. Private Functions such as Scale, Balance, Reduce, Expand, Contract, Finish and Render compose normalization, midpoint handling and byte construction. They are not additional application APIs.

Formatting loads its required [Storage](../../internals/storage.md) words and private foundations. It does not source-load Runtime or process supervision. Runtime.Format and IO's private formatting binding select this same canonical Format definition.

## Format

Convert a finite binary64 Number into its shortest round-tripping decimal representation and return the encoded bytes.

**Usage:** The selected native function, `format_number`, accepts `f64` and returns `Vec<u8>`. The returned vector contains neither an added newline nor a terminator. Format leaves output, locking and flushing to its caller.

**Example:** [Number output](#number-output-example) shows IO.Print using this formatter. Direct native-signature callers supply an `f64` and consume or borrow the resulting byte vector. This is not a general conversion for Text, collections or Errors.

**Behavior:** Preserve signed zero. Nonzero magnitudes from `0.0001` inclusive to `1000000` exclusive use ordinary decimal notation. Smaller or larger magnitudes use scientific notation. Scientific exponents include a sign and at least two digits.

| Number | Returned bytes shown as text |
| --- | --- |
| `3.5` | `3.5` |
| `-12.5` | `-12.5` |
| `999999` | `999999` |
| `1000000` | `1e+06` |
| `0.0001` | `0.0001` |
| `0.00001` | `1e-05` |
| `-0` | `-0` |

Large source integers may already have rounded to binary64. Formatting reports that stored value. The L [Number](../../spec/values.md#number) rules reject nonfinite values. Format's native `f64` signature does not itself enforce finiteness, so this reference does not extend the finite-number contract to arbitrary native NaN or infinity inputs.

## Number output example

This registered App-body fragment imports IO.Print and illustrates decimal and scientific output. It omits the surrounding lifecycle instructions and is not a standalone runnable program. Print appends one LF to each formatted value, while Format alone does not.

<!-- example:values-number-printing -->

For a complete App with execution and captured output, see [IO.Print](io.md#print). The source-number witness below exercises the formatting path through real L programs. Static page preparation does not execute the fragment.

## Errors and ownership

Format returns an owned byte vector, not a Result. Its internal expectations describe invariants of native finite-number spellings. They are not a recoverable parsing API. The composition makes no zero-allocation, allocation-count or nonblocking guarantee. It does not change IO's separate [output-failure policy](io.md#errors-and-return-behavior).

## Sources and witnesses

Sources: [Formatting dictionary](../../../stdlib/eng/format/dictionary.l) and its [private foundations](../../../stdlib/eng/format/foundations.l). Shared casts and other foundational operations compose through canonical Storage definitions.

The [dictionary-selection tests](../../../lyger/format_dictionary_test.go) check canonical Format identity, Runtime's alias and source-import isolation. [Source Number tests](../../../lyger/io_numeric_test.go) exercise Print with exact output. [Native scalar tests](../../../logiceme/scalar_runtime_test.go) cover the selected formatter's signed zero, extrema, subnormal values and midpoint cases. [Foundation tests](../../../lyger/library_foundations_test.go) check shared ownership and cast composition. These witnesses establish distinct properties. A documentation build alone establishes none of their runtime outcomes.
