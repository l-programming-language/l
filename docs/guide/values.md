# Work with values

Start with three source forms already used by the maintained examples: backtick-delimited Text, numeric literals, and `true`/`false` Truth values. Quoted declaration names are names rather than Text. The spelling of a source type and the compiler's storage tag are different interfaces.

## Text, Number and Truth together

<!-- example:public-values -->

The three public declarations survive the paragraph boundary. `Show Greeting`, `Show Count`, and `Show Enabled` produce the expected bytes `public7true`: each writes its value without inserting spaces or line breaks. The Dictionary declaration establishes a vocabulary namespace; it is not a demonstration of a generic key/value collection API.

The public-values case is covered by [paragraph integration assertions](../../cmd/l/paragraph_integration_test.go). The example's expected byte stream is recorded in the manifest; it was not run by the website build.

## Text is data, including whitespace

Literal handling preserves bytes and their source attribution. Periods and whitespace inside a quoted literal are not sentence boundaries. The quote scanner recognizes escaped delimiters so they do not close the literal, but that scanner is not a general promise of C-style escape decoding. Do not infer that writing a backslash followed by `n` automatically means an LF byte.

Captured App output is also Text and preserves raw stdout, including whitespace, NUL and non-ASCII bytes. The browser's Unicode text display is not a byte inspector. The documentation pipeline refuses unsafe decoding rather than silently claiming a changed display is the canonical source.

`IO.Print` writes supported byte-like Text with an added newline; `Show` writes Text exactly. A read of Output is meaningful only after capture finalization. A stored copy remains a snapshot if the target is run again.

## Numbers have a concrete representation

The current scalar Number representation is finite binary64. Its encoder uses a round-trip decimal representation for Rust and preserves negative zero and the parsed rounding. There is no public guarantee of arbitrary-precision arithmetic or a family of integer widths.

Numeric parsing and source-token boundaries both matter. In particular, an entry accepted by the number parser is not necessarily a complete standalone source token. Overflow remains an error with its native parse cause, and stored nonfinite values are unsupported. The [number encoder](../../lyger/native_encoding.go) and [numeric resolution tests](../../lyger/numeric_resolution_test.go) are the precise evidence for numeric boundaries.

Number formatting in the native print definition is deliberate, including decimal/scientific display and signed zero. Do not substitute JavaScript number formatting for the language's expected output in documentation checks.

## Other categories require their own contracts

The frontend has representations for Thing, Bytes, Pointer, collections and application structures, plus internal unresolved states. Those internal tags do not prove that every representation has a public literal, implicit conversion, mutable collection operation or printable form.

The [specification](../spec/index.md) distinguishes source spellings such as Byte from Bytes storage, constructible values from internal `ID`/`None`/`Unknown`, and vocabulary dictionaries from collection behavior. Use that reference before inferring copying, pointer ownership, indexing or iteration from a type's name.

Continue with [names and paragraphs](names-and-paragraphs.md).
