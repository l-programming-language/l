# Lexical structure and source organization

This chapter describes the byte-preserving source preparation and English lexical lookup implemented by Lyger. It distinguishes source punctuation from the notation used in grammar configuration.

## Source bytes and positions

Source preparation retains the original byte sequence and records source file, byte offset, line, column, paragraph, and occurrence ownership before removing comments or caching literals. Offsets are zero-based, while diagnostic lines and columns are one-based and count bytes. A multibyte Unicode character therefore occupies more than one diagnostic column. CRLF and LF input both retain meaningful original attribution.

Text can contain invalid UTF-8 octets. This is a supported preservation boundary, not a promise that every identifier operation accepts arbitrary bytes: case conversion and identifier validation have their own Unicode behavior. Do not normalize a source file, replace invalid octets, or trim application output when reproducing a failure.

## Whitespace, sentences, and paragraphs

Space, tab, carriage return, and newline separate unquoted source words. A period ends a sentence when followed by one of those whitespace bytes, a comment opener, or end of file. Decimal points and other uninterrupted-token dots remain part of the token. That scanner behavior does not permit dot-qualified authored names. Newlines alone are not implicit sentence terminators.

A sentence may span lines, and several sentences may share a line. A physical whitespace-only line outside a quoted literal or comment separates source paragraphs. Blank lines inside a multiline comment or quoted Text do not create a new paragraph. Paragraph identity is captured before imports and expansions, so importing text does not silently join the caller's pronoun context.

A paragraph whose first code line begins with four spaces opens an encapsulated lexical scope. Its following lines need not remain indented. A blank line closes it. Indentation is consequently not a general nested-block syntax. See [paragraph scope](names-and-scope.md#paragraph-scope) for source ownership.

The following fragment combines sentence-boundary shapes from the existing source-preparation witnesses:

<!-- example:lexical-sentence-boundaries -->

The first line contains two sentences: the decimal point in `3.14` stays inside the Number token, while the period after it ends the declaration. The following declaration belongs to the same paragraph even though it is unindented. The empty line begins a second paragraph. The period inside a Text literal belongs to that literal. This fragment illustrates **source preparation only**.

## Comments

Comments use paired triple asterisks:

<!-- example:lexical-comments -->

This is an explanatory fragment. The scanner ignores sentence punctuation inside a comment and rejects an unterminated comment. It does not establish nested-comment semantics. A comment next to a period can delimit that sentence without an intervening space.

## Text and declared names

Backticks delimit Text. Double quotes identify a declaration label. A bare use of a declared name refers to its binding:

The [basic program](../guide/first-program.md) declares `message` and supplies its Text to core Print. No import or runtime orchestration is needed.

The scanner skips a backslash and following byte while locating a quote terminator. Unquoting a backtick literal specifically replaces a backslash-escaped backtick with a literal backtick. General Text is not a C or JSON string with automatic `\n` interpretation. Definition templates have an explicitly narrower escape-decoding step described in [native definitions](definitions.md#framing-and-escaping). Literal contents, including braces, angle brackets, periods, and markup-like strings, must remain data through preprocessing.

<!-- example:lexical-text-escapes -->

The first declaration contains a literal backtick, the second declaration contains backslash followed by `n`, and the third declaration contains Unicode and markup punctuation. The interior `<DEF>` is ordinary Text because it is not a complete definition frame. These declarations demonstrate source bytes, not output behavior. See [Text](values.md#text) for the byte-preservation boundary.

Quoted source names cannot contain ordinary spaces in the primitive identifier classification. Case-insensitive lookup uses lowercase identity. Source-defined words may declare inflections with vertical bars inside the quoted label. Those bars belong to the lexical declaration, not statement conjunction.

## Names, possessives, and qualification

Names are matched through lexical bindings. Builtin types, configured vocabulary, source-defined roots and aliases retain their distinct roles. Lookup normally folds declared labels to lowercase while preserving the source spelling for diagnostics.

Use grammatical possessives such as `Function's Return`, or explicit member aliases. Dot-qualified names are not valid authored L. Internal compiler identity strings are separate. Once lookup selects an owner, a missing member cannot fall through to a same-spelled member of another owner. Pronoun possessives use paragraph-bound owner identity.

Native definition slots have a stricter identifier contract: each dot-separated segment begins with a Unicode letter or underscore, followed by letters, digits, or underscores. Empty segments, hyphens, leading digits, and combining marks rejected by that validator are not permitted. Rust or Go reserved words are not automatically reserved as L slot names. Native emitted names have separate target validation.

## Vocabulary and inflection

Configured English vocabulary has noun, verb, adjective, adverb, preposition, article, conjunction, and evaluation sections. Words such as `it`, `this`, and `that` are configured noun entries with pronoun roles. Lexical role, inflection, storage category, and value type remain distinct.

A source noun declaration can supply root, plural, descriptive, and active forms. A verb declaration can supply root, active, passive, applicative, and objective forms. For example, the builtin declaration `"Execute|-s|-d|--ing|--or"` yields Execute, Executes, Executed, Executing, and Executor. A postfix with one hyphen appends its letters. Each additional hyphen removes one trailing byte from the root before appending. This byte-oriented rule must not be generalized into linguistic Unicode inflection.

Core retains Start and Execute declarations, but their Build/Run/Await effects reject during the reset. Their morphology is source evidence, not current lifecycle support. Print's `"Print|-s|-ed|-ing|-er"` declares Print, Prints, Printed, Printing and Printer.

When declared forms share a spelling, context-free lookup keeps the root form. The enclosing grammar can select another declared role of that same canonical word. Failed alternatives do not change the token binding. A complete Function-value assignment, such as `"Saved" is Read`, retains assignment precedence. This does not introduce general passive Function calls.

An active noun form can serve as an applicative verb, a descriptive noun form as an adjective, and an objective verb form as a noun. These conversions preserve the canonical definition. Prefix storage exists internally, but source label declaration currently assigns its root first and does not establish a general public prefix-declaration language.

## Literal and token precedence

Token loading first recognizes Text, exact lowercase `true` or `false`, quoted identifiers, and numeric candidates. Then native types and configured vocabulary participate. A numeric-range candidate can remain provisional while source-defined names resolve. If a configured or source binding supplies that spelling, the compiler does not prematurely replace it with a numeric error. Unresolved explicit members remain reference failures.

A declared name with numeric-looking spelling remains a name if binding resolves it. Otherwise a numeric range error retains its original cause. Recognition does not convert an overflowing Number into a successful value.

Grammar patterns such as `%~%` and `?~?` name typed token categories. They are not instructions to surround ordinary numbers and Truth literals with those characters. The actual source examples use `7`, `-12.5`, `true`, and `false`.

## Sources and witnesses

Implementations: [source preparation](../../lyger/source.go), [parser](../../lyger/parser.go), [literal unquoting](../../lyger/grammar.go), [token loading](../../lyger/lexicon.go), [lexical forms](../../lyger/lexical_types.go), [type and inflection utilities](../../lyger/utilities.go), and [definition slots](../../lyger/definition_bindings.go). Existing witnesses: [source framing](../../lyger/source_test.go), [paragraph positions](../../lyger/paragraph_source_test.go), [literal framing](../../lyger/literal_framing_regression_test.go), [numeric resolution](../../lyger/numeric_resolution_test.go), and [slot names](../../lyger/definition_slot_names_test.go).
