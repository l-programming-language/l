# IO

The English `io` dictionary provides value printing and explicit byte output. It has no filesystem, terminal-input, or networking API.

## Import and calling forms

The supported member import is:

```l
Use IO's Print from `io` as "Print".
```

The action's forms are Print, Prints, Printed, Printing, and Printer. Use it in an App body, as in `The App Prints message.` The [first program](../../guide/first-program.md) shows a complete example.

## Public and private names

| Public member | Contract |
| --- | --- |
| IO.Print | Render a supported value and write it to stdout followed by LF |
| IO.Output | Obtain the standard-output handle |
| IO.Diagnostic | Obtain the standard-error handle |
| IO.Send / Sender | Write borrowed bytes through a supplied mutable writer and return its native IO Result |
| IO.Lined | Describe a Sender that appends LF after its payload succeeds |
| IO.Flush / Flusher | Flush a supplied mutable writer and return its native IO Result |

Each member can be imported with the same member-import form as Print, including the Lined adjective. Render, Write, locking and line-completion helpers remain private implementation dependencies. The former Stream, Check and Slice helpers have been removed; no compatibility aliases are provided.

## Explicit byte output

Send is available to source Functions with native signatures. Its writer is a borrowed `&mut W` where `W: std::io::Write + ?Sized`; its payload is a borrowed `&P` where `P: AsRef<[u8]> + ?Sized`. It obtains a byte view and calls `write_all`, returning `std::io::Result<()>`. It adds no newline, lock, flush or panic policy and does not convert the payload to owned Text. Render Numbers and other formatted values before sending them.

`New IO.Lined IO's Sender` retains the selected Sender invocation. If that invocation succeeds, an L-authored Result selection sends the existing LF bytes through the same writer. A payload failure skips LF; a newline failure follows the completed payload. The operation returns the first failure or successful unit. It does not reacquire a handle or lock.

Output and Diagnostic obtain their named handles without writing or locking. Flush is a separate request; Send, Lined and Print do not add an explicit flush. The writer determines allocation and blocking behavior, and flushing does not promise durable storage. A failed send can leave partial output; a line is not a transaction.

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

Print has no public return value. Its private Write consumes the rendered byte vector, obtains stdout and one guard, invokes a Lined Sender while that guard remains live, and uses `Storage.Result.Expect` with `failed printing to stdout`. An output failure therefore fails in the native App context; it is not a compiler diagnostic. Print adds no retry policy and does not suppress a failed writer. The program can inspect the App's Error after Await.

Send, Lined and Flush instead return the native IO Result for the caller to handle. `Storage.Result.Expect` is an explicit fail-fast choice, separate from result selection or propagation.

## Availability and limits

IO is embedded in the compiler. Import it with `Use`; no separate library installation is needed. See the [standard-library catalog](../index.md) for available dictionaries.

## Sources and witnesses

Source: [IO dictionary](../../../stdlib/eng/io/dictionary.l). Tests cover [public byte IO and failure paths](../../../lyger/word_library_acceptance_test.go), [Write's guard and panic policy](../../../lyger/word_library_io_policy_test.go), [raw Text](../../../lyger/raw_text_test.go), [Number formatting](../../../lyger/io_numeric_test.go), [imports](../../../lyger/configured_import_test.go) and [private ownership](../../../lyger/dictionary_ownership_test.go).
