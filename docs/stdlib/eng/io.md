# IO

<a id="availability-and-limits"></a>

The English `io` dictionary provides value printing and explicit byte output. Import it with Use. It has no filesystem, terminal-input, or networking API. IO is an explicit library, not implicit builtin vocabulary.

## Import and calling forms

The supported member import is:

```l
Use IO's Print from `io` as "Print".
```

Use Print in an App body, as in `The App Prints message.` The [first program](../../guide/first-program.md) shows a complete example. ``Use IO from `io`.`` instead keeps the dictionary namespace for qualified member access. The same member-import form works for the other public words, including Lined.

| Task | Word |
| --- | --- |
| Print a supported value to stdout with LF | [Print](#print) |
| Obtain stdout or stderr without writing | [Output](#output), [Diagnostic](#diagnostic) |
| Send encoded bytes through a supplied writer | [Send](#send) |
| Add LF after successful delivery | [Lined](#lined) |
| Request a writer flush | [Flush](#flush) |

## Public and private names

The six entries below are IO's public roots. Sender and Flusher are actor forms of their verbs. Render, Write, locking and line-completion helpers remain private to the IO source unit. Importing IO does not expose them. A public library entry does not certify allocation, blocking, durability or rollback guarantees.

## Words

<!-- word-index:eng:io -->

### Diagnostic

<!-- word:stdlib-export:eng:io:IO.Diagnostic -->

Obtain a handle to the current process's standard-error stream without writing or locking it.

**Usage:** In a native-signature Function, `a New IO's Diagnostic` produces `std::io::Stderr`. The caller decides whether to lock or borrow that handle and which operation to perform. Obtaining it does not create an App error or inspect captured stderr.

**Example:** The `Writing.Diagnostic` Function in [Native callers](#native-callers) returns the handle to its host. Compose a writer with [Send](#send) or [Flush](#flush). Use [Output](#output) for stdout instead.

### Flush

<!-- word:stdlib-export:eng:io:IO.Flush -->

Request a flush from a supplied mutable writer and return that writer's native IO result.

**Usage:** `a New IO's Flusher of Writer` calls `std::io::Write::flush` on the supplied `&mut W`. The caller handles `std::io::Result<()>`. Flush neither chooses stdout nor sends a payload, and successful flushing does not universally mean durable storage.

**Example:** `Writing.Flush` in [Native callers](#native-callers) uses the same mutable byte vector supplied by its host. [Send](#send), [Lined](#lined) and [Print](#print) add no explicit flush. The concrete writer determines what a flush does.

### Lined

<!-- word:stdlib-export:eng:io:IO.Lined -->

Describe IO.Send's Sender actor so that it sends one LF through the same writer after successful payload delivery.

**Usage:** `a New IO.Lined IO's Sender of Writer and Payload` targets **IO.Send's Sender actor (`IO.Sender`)**, retaining its selected invocation and original writer. Lined is not a modifier for any unrelated word spelled Sender.

**Example:** `Writing.Line` in [Native callers](#native-callers) returns the lined operation's `std::io::Result<()>`. A payload failure skips LF. A newline failure follows an already completed payload. Read [Explicit byte output](#explicit-byte-output) before choosing a retry policy: a line is not a transaction.

### Output

<!-- word:stdlib-export:eng:io:IO.Output -->

Obtain a handle to the current process's standard-output stream without writing or locking it.

**Usage:** In a native-signature Function, `a New IO's Output` produces `std::io::Stdout`. This is distinct from [builtin App Output](../../reference/builtins/lifecycle.md#output), which exposes finalized captured stdout bytes.

**Example:** `Writing.Output` in [Native callers](#native-callers) returns the handle. Choose [Diagnostic](#diagnostic) for stderr, or use [Print](#print) when its rendering, newline and failure policy are wanted together.

### Print

<!-- word:stdlib-export:eng:io:IO.Print -->

Render a supported value and write its bytes to standard output followed by one LF newline.

**Usage:** Import Print and invoke it on an App, as in the complete program below. Each argument is evaluated once. See [Value handling](#value-handling) for accepted operands.

**Example:** The App prints the greeting, Execute waits for it, and Show displays the captured bytes including LF.

<!-- example:basic -->

**Behavior:** Print composes rendering and a lined stdout send with its [fail-fast error policy](#errors-and-return-behavior). It makes no zero-allocation or nonblocking promise. Use [Send](#send) with [Lined](#lined) to choose the writer and handle returned failure.

### Send

<!-- word:stdlib-export:eng:io:IO.Send -->

Write a borrowed byte view through a supplied mutable writer and return the native IO result.

**Usage:** `a New IO's Sender of Writer and Payload` takes `&mut W` with `W: std::io::Write + ?Sized`, then `&P` with `P: AsRef<[u8]> + ?Sized`, and returns `std::io::Result<()>`. Sender is the actor form of Send, and Sent is its declared passive form, not a separate implementation.

**Example:** `Writing.Send` in [Native callers](#native-callers) sends a borrowed byte slice into a supplied vector. Send obtains one byte view and calls `write_all`. It adds no newline, lock, flush, owned Text conversion or panic policy. Use [Lined](#lined) to append LF and [Flush](#flush) for an explicit flush request.

## Native callers

These are complete Function declarations within a dictionary, not a standalone App. A native host supplies the standard Rust writer and payload types and invokes the generated Functions. The source uses IO's selected words rather than reimplementing delivery in a native helper. `Vec<u8>` makes the destination visible, so a host can inspect its exact bytes after Send or Line.

<!-- example:word-io -->

The [documentation caller witness](../../../lyger/word_documentation_test.go) compiles these canonical bytes and exercises the generated calls. It checks handle types, embedded NUL/invalid UTF-8/LF bytes, line completion, and the unchanged buffer after vector flush. The separate [IO behavior witness](../../../lyger/word_library_acceptance_test.go) covers failing and short-write writers. A vector alone would not establish those failure cases.

## Explicit byte output

The native `write_all` operation continues after short writes and retries Interrupted. A zero-length write before completion fails with WriteZero. An empty payload requires no write. Other errors propagate, potentially after partial output. IO adds no further retry policy.

Lined uses L-authored Result selection: only successful payload delivery invokes the LF send through the same writer. It neither reacquires a handle nor locks it. A failed send can leave a prefix, and newline failure can follow a complete payload. Neither Send nor Lined promises rollback or safe retry. The writer determines allocation and blocking behavior.

## Value handling

Print accepts supported Text, Number, Truth and Null operands. Text bytes remain unchanged, including whitespace, NUL and invalid UTF-8. Truth renders as `true` or `false`. Null renders as `<nil>`, distinct from the `null` display used by other value contexts. Numbers use the formatting below.

Lists, Dictionaries, Pointers and internal compiler states have no general Print contract.

## Output and newlines

Print does not trim its operand before appending LF. A value already ending with LF therefore produces another LF. Empty Text produces one LF. [Show](../../reference/builtins/lifecycle.md#show) preserves captured Output without adding a newline. Captured stderr remains separate.

## Number formatting

Number uses binary64. Print selects [Formatting.Format](format.md#format) for the shortest round-tripping representation, including scientific notation and signed zero, then appends LF. For example, `1000000` prints `1e+06\n` and negative zero prints `-0\n`. The Formatting reference owns the numeric thresholds and exact byte examples.

Large source integers may already have rounded. Print reports that binary64 value. Nonfinite L Numbers are rejected before formatting.

## Errors and return behavior

Print has no public return value. Its private Write consumes the rendered byte vector, obtains stdout and one guard, invokes a Lined Sender while that guard remains live, and uses [Storage.Result.Expect](../../internals/storage.md#resultexpect) with `failed printing to stdout`. An output failure panics in the native App context. It is not a compiler diagnostic. The guard is released on success or unwinding. The program can inspect the App's Error after Await.

Send, Lined and Flush instead return the native IO Result for the caller to handle. [Storage's result words](../../internals/storage.md#results-and-errors) distinguish selection, propagation and the explicit Expect panic policy.

## Sources and witnesses

Source: [IO dictionary](../../../stdlib/eng/io/dictionary.l). Tests cover [public byte IO and failure paths](../../../lyger/word_library_acceptance_test.go), [Write's guard and panic policy](../../../lyger/word_library_io_policy_test.go), [raw Text](../../../lyger/raw_text_test.go), [Number formatting](../../../lyger/io_numeric_test.go), [imports](../../../lyger/configured_import_test.go) and [private ownership](../../../lyger/dictionary_ownership_test.go).
