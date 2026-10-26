# IO

The English `io` dictionary exports the IO namespace and its Print member. It supplies output through a source-defined native action. It does not introduce general filesystem, terminal-input, formatting, or networking APIs.

## Import and calling forms

The supported member import is:

```l
Use IO's Print from `io` as "Print".
```

The public declaration is `Create IO's "Print|-s|-ed|-ing|-er" for Processing a New Logger.` Its forms are Print, Prints, Printed, Printing, and Printer. The alias keeps the imported definition's canonical source identity and dependencies; it does not copy private helpers into the caller's lexical root.

The maintained basic example uses the active application-body form `The App Prints message.` Its complete source is included below:

<!-- example:basic -->

Use Print in an App body, as shown here; its grammatical forms do not make every execution context valid.

## Public and private names

IO.Print is the public member. Log and its objective noun Logger implement the native operation privately inside the dictionary source unit. They are dependencies of Print, not additional callable public exports.

The imported dictionary owns a private lexical root inheriting implicit builtins. Its helpers cannot capture caller-added members, and the caller cannot resolve them unqualified through a Print import. IO remains a namespace; a same-named local collection or neighboring `io.l` file does not override the index entry.

## Value handling

Print's private definition has an Output/Thing slot. The typed definition boundary can encode supported Text, Number, and Truth operands without converting every argument to Text first. The native dispatch uses a borrowed wrapper around the operand, preserving the one-evaluation contract even when the expression has an effect.

Raw Text can contain valid Unicode, invalid UTF-8 octets, NUL, or markup delimiters. The native byte-capable implementation writes those bytes unchanged. Invalid UTF-8 is not replaced with a Unicode replacement character and is not rejected merely to make a UTF-8 display convenient.

The native helper also has bounded Rust Display dispatch for deliberately supplied native values. That implementation is not a promise that every L internal value tag can be passed through a general generic Print source API. Lists, Dictionaries, Pointer, and internal resolution states do not acquire supported printing semantics by this fallback's existence.

## Output and newlines

Print appends exactly one LF after the supplied byte/text or formatted scalar output. It does not trim the operand first. A value already ending with LF therefore produces another LF. Empty Text produces one LF.

Show is different: it is the outer observation builtin and adds no newline. In the basic example, Print writes `Hello World!` plus LF inside the App; Execute joins it; Show preserves that captured Output, including the LF. The deliberately negative paragraph example must remain negative when it compares captured `Hello World!\n` against `Hello World!`.

Captured child stderr is not automatically forwarded as stdout. IO.Print writes to the child/native context where the composed action runs. An orchestrator Show of Output is the explicit step that publishes captured bytes in the common tutorial.

## Number formatting

L Number is binary64. IO's native formatter preserves the established shortest round-tripping formatting policy corresponding to `strconv.FormatFloat(value, 'g', -1, 64)`, including scientific notation thresholds and signed zero. Representative existing expectations include:

| Number | Printed bytes shown with escaped LF |
| --- | --- |
| `3.5` | `3.5\n` |
| `-12.5` | `-12.5\n` |
| `999999` | `999999\n` |
| `1000000` | `1e+06\n` |
| `0.0001` | `0.0001\n` |
| `0.00001` | `1e-05\n` |
| `-0` | `-0\n` |

Large source integers may already have rounded during Number parsing; Print reports that binary64 value. Nonfinite source/native Number inputs are rejected at earlier typed boundaries. Native integer Thing arguments intentionally supplied to a native definition retain their own Display formatting and do not define an L integer-width family.

## Errors and return behavior

Print is an output effect. No public return-value API is established. Byte/numeric branches lock stdout and use write-all operations followed by the newline; an IO failure calls the native `expect("failed printing to stdout")` path. The Display fallback uses the native printing macro. These failures belong to the generated application/native context, not the Go compiler's lerrors API.

The enclosing App result can therefore be unsuccessful. L program code still decides how to handle its process Error after Await. Print does not automatically trim, retry output, swallow a broken writer, or Exit the orchestrator directly through a source exception system.

## Availability and limits

The English IO dictionary at `stdlib/eng/io/dictionary.l` is embedded in the compiler. Import it with `Use`; no separate library installation is needed.

Additional standard-library dictionaries belong in the index-driven [catalog](../index.md), with their own reviewed public exports and evidence. IO's single-file implementation is not a restriction on future dictionary organization.

## Sources and witnesses

Definition: [IO dictionary](../../../stdlib/eng/io/dictionary.l). Resolution: [index](../../../stdlib/index.json), [export/member checks](../../../lyger/source_units.go), and [binding](../../../lyger/semantic_resolution.go). Existing witnesses: [raw text and typed dispatch](../../../lyger/raw_text_test.go), [numeric formatting](../../../lyger/io_numeric_test.go), [dictionary imports](../../../lyger/configured_import_test.go), [private ownership](../../../lyger/dictionary_ownership_test.go), and [basic source](../../../demo/basic/main.l). These tests were inspected, not executed here.
