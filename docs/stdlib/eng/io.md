# IO

The English `io` dictionary exports IO.Print for writing output. It has no filesystem, terminal-input, or networking API.

## Import and calling forms

The supported member import is:

```l
Use IO's Print from `io` as "Print".
```

The action's forms are Print, Prints, Printed, Printing, and Printer. Use it in an App body, as in `The App Prints message.` The [first program](../../guide/first-program.md) shows a complete example.

## Public and private names

IO.Print is public. Its typed rendering and writing helpers remain private dependencies of that import.

## Value handling

Print accepts supported Text, Number, Truth and Null operands and evaluates each argument once. It does not first convert every argument to Text.

Text bytes are written unchanged, including whitespace, NUL and invalid UTF-8.

Lists, Dictionaries, Pointers and internal compiler states have no general Print contract.

## Output and newlines

Print appends exactly one LF after the supplied byte/text or formatted scalar output. It does not trim the operand first. A value already ending with LF therefore produces another LF. Empty Text produces one LF.

Show adds no newline. In the first program, Print writes `Hello World!` plus LF inside the App; Execute waits for it; Show displays its captured Output, including that LF. Captured stderr remains separate.

## Number formatting

Number uses binary64. Print uses the shortest round-tripping representation, including scientific notation and signed zero:

| Number | Printed bytes shown with escaped LF |
| --- | --- |
| `3.5` | `3.5\n` |
| `-12.5` | `-12.5\n` |
| `999999` | `999999\n` |
| `1000000` | `1e+06\n` |
| `0.0001` | `0.0001\n` |
| `0.00001` | `1e-05\n` |
| `-0` | `-0\n` |

Large source integers may already have rounded; Print reports that binary64 value. Nonfinite Numbers are rejected.

## Errors and return behavior

Print has no public return value. An output failure fails in the native App context; it is not a compiler diagnostic. Print does not retry or suppress a broken writer. The program can inspect the App's Error after Await.

## Availability and limits

IO is embedded in the compiler. Import it with `Use`; no separate library installation is needed. See the [standard-library catalog](../index.md) for available dictionaries.

## Sources and witnesses

Source: [IO dictionary](../../../stdlib/eng/io/dictionary.l). Tests cover [raw Text](../../../lyger/raw_text_test.go), [Number formatting](../../../lyger/io_numeric_test.go), [imports](../../../lyger/configured_import_test.go) and [private ownership](../../../lyger/dictionary_ownership_test.go).
