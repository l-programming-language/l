# Work with values

Start with three source forms already used by the maintained examples: backtick-delimited Text, numeric literals, and `true`/`false` Truth values. Quoted declaration names are names rather than Text. The spelling of a source type and the compiler's storage tag are different interfaces.

## Text, Number and Truth together

<!-- example:public-values -->

The three public declarations survive the paragraph boundary. `Show Greeting`, `Show Count`, and `Show Enabled` produce the expected bytes `public7true`: each writes its value without inserting spaces or line breaks. The Dictionary declaration establishes a vocabulary namespace; it is not a demonstration of a generic key/value collection API.

## Read pipe-delimited word forms

In the example above, `"Display|-s|-ed|-ing|-er"` declares the spellings of one action. Each pipe (`|`) separates a grammatical form. This is called **inflection** or **lexical casing**: it tells L how a word can fit into a sentence. Capitalization is a separate matter; L looks up these names without regard to letter case.

The first field is the **root**, `Display`. Each later field gives an ending to apply to that root. A single leading hyphen means “append these letters,” so `-ed` produces `Displayed`. Apply every ending to `Display`, independently of the other endings.

For a **verb**, the positions are **root, active, passive, applicative, objective**:

| Field | Spelling | Grammatical role |
| --- | --- | --- |
| `Display` | Display | Root verb, used for a direct command. |
| `-s` | Displays | Active verb, used when a subject performs the action. |
| `-ed` | Displayed | Passive verb form, used in a passive clause. |
| `-ing` | Displaying | Applicative verb: the action form used in compositions, like `Showing` in the declaration above. |
| `-er` | Displayer | Objective form: the same definition expressed as a noun. |

These forms let the grammar recognize different uses of the same definition. For example, the `for Showing any Text` body defines what Display does, while the pipe fields define how its name is written in different grammatical positions. The passive spelling does not schedule an action in the past, and the noun spelling does not create a new object or App.

**Nouns have a different field order:** **root, plural, descriptive, active**. The builtin label `"Function|-s|-al|-ing"` declares `Function` and `Functions` as nouns, `Functional` as an adjective, and `Functioning` as an applicative verb. A word can therefore take on an adjective or verb role while still referring to its original definition.

L also recognizes other parts of speech, such as adjectives and adverbs. For example, `Public` modifies a noun in `a New Public Dictionary`, while `Publicly` modifies a declaration in `Publicly define`. Pipe-delimited casing is supported for noun and verb labels; it is not a general form list for every part of speech. These grammatical roles are separate from value types such as Text and Number.

Write the endings explicitly: L does not guess English spelling rules. Each additional hyphen removes one trailing byte from the root before adding the ending. In the builtin `"Execute|-s|-d|--ing|--or"`, `--ing` removes the final `e` and adds `ing`, producing `Executing`; `--or` produces `Executor`. This is a byte-based spelling rule, so do not assume it handles every Unicode word like an English word. You can leave off unused trailing forms, or use a plain quoted name when only a root is needed.

See [vocabulary and inflection](../spec/lexical-structure.md#vocabulary-and-inflection) and [clauses and grammatical forms](../spec/grammar.md#clauses-and-grammatical-forms) for the detailed grammar.

## Text is data, including whitespace

<!-- example:guide-text-value -->

The quoted name `"message"` receives the backtick-delimited Text `Hello World!`, including the space between the words. This is the value declaration from the [first program](first-program.md).

Literal handling preserves bytes and their source attribution. Periods and whitespace inside a quoted literal are not sentence boundaries. The quote scanner recognizes escaped delimiters so they do not close the literal, but that scanner is not a general promise of C-style escape decoding. Do not infer that writing a backslash followed by `n` automatically means an LF byte.

Captured App output is also Text and preserves raw stdout, including whitespace, NUL and non-ASCII bytes. A browser's Unicode display does not expose every byte.

`IO.Print` writes supported byte-like Text with an added newline; `Show` writes Text exactly. A read of Output is meaningful only after capture finalization. A stored copy remains a snapshot if the target is run again.

## Numbers have a concrete representation

<!-- example:guide-number-value -->

`Count` receives the Number `7`. The numeric literal has no quotes or backticks; the double quotes belong to the name being declared. `Publicly` makes that name available after the paragraph, as in the complete example above.

The current scalar Number representation is finite binary64. Its encoder uses a round-trip decimal representation for Rust and preserves negative zero and the parsed rounding. There is no public guarantee of arbitrary-precision arithmetic or a family of integer widths.

Numeric parsing and source-token boundaries both matter. In particular, an entry accepted by the number parser is not necessarily a complete standalone source token. Overflow remains an error with its native parse cause, and stored nonfinite values are unsupported. The [number encoder](../../lyger/native_encoding.go) and [numeric resolution tests](../../lyger/numeric_resolution_test.go) are the precise evidence for numeric boundaries.

## Other categories require their own contracts

Use [Primitive Types](../spec/values.md), [Iterative Types](../spec/collections-and-categories.md) and [Applicational Types](../spec/applicational-types.md) for construction, copying and sharing rules. Structural source categories and unresolved compiler states remain distinct from runtime data.

The [specification](../spec/index.md) distinguishes source spellings such as Byte from Bytes storage, constructible values from internal `ID`/`None`/`Unknown`, and vocabulary dictionaries from collection behavior. Use that reference before inferring copying, pointer ownership, indexing or iteration from a type's name.

Continue with [names and paragraphs](names-and-paragraphs.md).
