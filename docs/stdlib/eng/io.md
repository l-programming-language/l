# IO

<a id="availability-and-limits"></a>

The embedded English `io` dictionary provides value printing and explicit byte output. Import it with Use; it has no filesystem, terminal-input, or networking API.

## Import and calling forms

The supported member import is:

```l
Use IO's Print from `io` as "Print".
```

Use Print in an App body, as in `The App Prints message.` The [first program](../../guide/first-program.md) shows a complete example. ``Use IO from `io`.`` instead keeps the dictionary namespace for qualified member access.

## Public and private names

| Public word and kind | Declared forms | Contract |
| --- | --- | --- |
| IO.Print — verb | Print, Prints, Printed, Printing, Printer | Render a supported value and write it to stdout followed by LF |
| IO.Output — noun | Output | Obtain the standard-output handle |
| IO.Diagnostic — noun | Diagnostic | Obtain the standard-error handle |
| IO.Send — verb | Send, Sends, Sent, Sending, Sender | Write borrowed bytes through a supplied mutable writer and return its native IO Result |
| IO.Lined — adjective | Lined | Describe a Sender that appends LF after its payload succeeds |
| IO.Flush — verb | Flush, Flushes, Flushed, Flushing, Flusher | Flush a supplied mutable writer and return its native IO Result |

Each member can be imported with the same member-import form as Print, including Lined. Sender and Flusher are actor forms of their verbs. Render, Write, locking and line-completion helpers remain private to the IO source unit; importing IO does not expose them.

## Explicit byte output

Send is available to source Functions with native signatures. Its writer is a borrowed `&mut W` where `W: std::io::Write + ?Sized`; its payload is a borrowed `&P` where `P: AsRef<[u8]> + ?Sized`. It obtains one byte view and calls `write_all`, returning `std::io::Result<()>`. It adds no newline, lock, flush or panic policy and does not convert the payload to owned Text. Send requires encoded bytes; Print supplies supported scalar rendering.

The native `write_all` operation continues after short writes and retries Interrupted. A zero-length write before completion fails with WriteZero; an empty payload requires no write. Other errors propagate, potentially after partial output. IO adds no further retry policy.

`New IO.Lined IO's Sender` retains the selected Sender invocation. If that invocation succeeds, an L-authored Result selection sends the existing LF bytes through the same writer. A payload failure skips LF; a newline failure follows the completed payload. The operation returns the first failure or successful unit. It does not reacquire a handle or lock.

Output and Diagnostic obtain their named handles without writing or locking. Flush is a separate request; Send, Lined and Print do not add an explicit flush. The writer determines allocation and blocking behavior, and flushing does not promise durable storage. A failed send can leave partial output; a line is not a transaction.

## Value handling

Print accepts supported Text, Number, Truth and Null operands and evaluates each argument once. Text bytes remain unchanged, including whitespace, NUL and invalid UTF-8. Truth renders as `true` or `false`; Null renders as `<nil>`. Numbers use the formatting below.

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

Print has no public return value. Its private Write consumes the rendered byte vector, obtains stdout and one guard, invokes a Lined Sender while that guard remains live, and uses `Storage.Result.Expect` with `failed printing to stdout`. An output failure panics in the native App context; it is not a compiler diagnostic. The guard is released on success or unwinding. The program can inspect the App's Error after Await.

Send, Lined and Flush instead return the native IO Result for the caller to handle. [Storage's result words](../../internals/storage.md#results-and-errors) distinguish selection, propagation and the explicit Expect panic policy.

## Sources and witnesses

Source: [IO dictionary](../../../stdlib/eng/io/dictionary.l). Tests cover [public byte IO and failure paths](../../../lyger/word_library_acceptance_test.go), [Write's guard and panic policy](../../../lyger/word_library_io_policy_test.go), [raw Text](../../../lyger/raw_text_test.go), [Number formatting](../../../lyger/io_numeric_test.go), [imports](../../../lyger/configured_import_test.go) and [private ownership](../../../lyger/dictionary_ownership_test.go).
