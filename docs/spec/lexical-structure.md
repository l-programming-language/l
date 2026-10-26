# Lexical structure and source organization

This chapter describes the byte-preserving source preparation and English lexical lookup implemented by Lyger. It distinguishes source punctuation from the notation used in grammar configuration.

## Source bytes and positions

Source preparation retains the original byte sequence and records source file, byte offset, line, column, paragraph, and occurrence ownership before removing comments or caching literals. Offsets are zero-based; diagnostic lines and columns are one-based and count bytes. A multibyte Unicode character therefore occupies more than one diagnostic column. CRLF and LF input both retain meaningful original attribution.

Text can contain invalid UTF-8 octets. This is a supported preservation boundary, not a promise that every identifier operation accepts arbitrary bytes: case conversion and identifier validation have their own Unicode behavior. Do not normalize a source file, replace invalid octets, or trim application output when reproducing a failure.

## Whitespace, sentences, and paragraphs

Space, tab, carriage return, and newline separate unquoted source words. A period ends a sentence when followed by one of those whitespace bytes, a comment opener, or end of file. Decimal points, dots in qualified names, and other uninterrupted-token dots remain part of the token. Newlines alone are not implicit sentence terminators.

A sentence may span lines; several sentences may share a line. A physical whitespace-only line outside a quoted literal or comment separates source paragraphs. Blank lines inside a multiline comment or quoted Text do not create a new paragraph. Paragraph identity is captured before imports and expansions, so importing text does not silently join the caller's pronoun context.

A paragraph whose first code line begins with four spaces opens an encapsulated lexical scope. Its following lines need not remain indented. A blank line closes it. Indentation is consequently not a general nested-block syntax. See [paragraph scope](names-and-scope.md#paragraph-scope) and [activation](execution.md#continuations-and-activation) for the different lexical and runtime roles.

The following fragment combines sentence-boundary shapes from the existing source-preparation witnesses:

<!-- example:lexical-sentence-boundaries -->

The first line contains two sentences: the decimal point in `3.14` stays inside the Number token, while the period after it ends the declaration. The next `Show` belongs to the same paragraph even though it is unindented. The empty line begins a second paragraph; the period inside `tail.` belongs to Text. This documentation fragment illustrates **source preparation only**. The related tests inspect sentence spans and paragraph identity; this combined fragment has not been compiled or run.

## Comments

Comments use paired triple asterisks:

<!-- example:lexical-comments -->

This is an explanatory fragment. The scanner ignores sentence punctuation inside a comment and rejects an unterminated comment. It does not establish nested-comment semantics. A comment next to a period can delimit that sentence without an intervening space.

## Text and declared names

Backticks delimit Text. Double quotes identify a declaration label. A bare use of a declared name refers to its binding:

<!-- example:lexical-declared-name -->

This is an excerpt from [the complete basic example](../guide/first-program.md), whose earlier import binds `Print`. `"message"` declares a name; backtick-delimited `Hello World!` supplies its Text value. The later bare `message` reads that binding as the App's print operand. The excerpt omits import and execution steps; use the linked program to run it.

The scanner skips a backslash and following byte while locating a quote terminator. Unquoting a backtick literal specifically replaces a backslash-escaped backtick with a literal backtick. General Text is not a C or JSON string with automatic `\n` interpretation. Definition templates have an explicitly narrower escape-decoding step described in [native definitions](definitions.md#framing-and-escaping). Literal contents, including braces, angle brackets, periods, and markup-like strings, must remain data through preprocessing.

<!-- example:lexical-text-escapes -->

In this documentation-only fragment, the first operand contains a literal backtick, the second contains backslash followed by `n`, and the third contains Unicode and markup punctuation. The interior `<DEF>` is ordinary Text because it is not a complete definition frame. `Show` adds no separator or newline between operands. These examples are based on inspected quote-scanning, literal-framing, and Text-emission contracts; the fragment itself has not been compiled or run. See [Text](values.md#text) for the byte-preservation boundary.

Quoted source names cannot contain ordinary spaces in the primitive identifier classification. Case-insensitive lookup uses lowercase identity. Source-defined words may declare inflections with vertical bars inside the quoted label; those bars belong to the lexical declaration, not statement conjunction.

## Names, possessives, and qualification

Names are matched through lexical bindings, not a universal identifier regex. Builtin native type names, configured vocabulary, source-defined lexical roots, aliases, and dotted members are distinct lookup inputs. Double-quoted declared labels are retained for diagnostics; lookup normally folds them to lowercase.

Possessive forms select a member of a particular owner, as in `Hello's Output` and `IO's Print`. A dotted form such as `IO.Print` names the same namespace role where qualified resolution is supported. Once lookup selects an owner, a missing member cannot fall through to a same-named member on an unrelated owner. Pronoun possessives, including the repository's `it's Output` form, use paragraph-bound owner identity; they are not English punctuation correction opportunities.

For example, the basic program's final two sentences use an explicit owner:

<!-- example:lexical-possessive-output -->

`Hello` must already name the App declared earlier in that program. `Execute` completes its lifecycle before `Hello's Output` is shown; `Output` is the member of that App, not a global variable. This excerpt shares the basic program's existing execution witness and has not been run independently.

Native definition slots have a stricter identifier contract: each dot-separated segment begins with a Unicode letter or underscore, followed by letters, digits, or underscores. Empty segments, hyphens, leading digits, and combining marks rejected by that validator are not permitted. Rust or Go reserved words are not automatically reserved as L slot names. Native emitted names have separate target validation.

## Vocabulary and inflection

Configured English vocabulary has noun, verb, adjective, adverb, preposition, article, conjunction, and evaluation sections. Words such as `it`, `this`, and `that` are configured noun entries with pronoun roles. Lexical role, inflection, storage category, and value type remain distinct.

A source noun declaration can supply root, plural, descriptive, and active forms. A verb declaration can supply root, active, passive, applicative, and objective forms. For example, the builtin declaration `"Execute|-s|-d|--ing|--or"` yields Execute, Executes, Executed, Executing, and Executor. A postfix with one hyphen appends its letters; each additional hyphen removes one trailing byte from the root before appending. This byte-oriented rule must not be generalized into linguistic Unicode inflection.

The implicit core defines the two lifecycle words this way:

<!-- example:lexical-action-inflections -->

`Starting` in the second definition uses an inflection of the preceding `Start`; `it` refers to the action's bound App operand. The inflection bars declare forms of one lexical identity, while `and then` sequences the action's steps. This is an excerpt of already-loaded builtin definitions, not code to redeclare in a user program. Existing builtin/frontend witnesses cover their expansion; no language execution was performed here.

An active noun form can serve as an applicative verb; a descriptive noun form can serve as an adjective; an objective verb form can serve as a noun. These conversions preserve the canonical definition. Prefix storage exists internally, but source label declaration currently assigns its root first and does not establish a general public prefix-declaration language.

## Literal and token precedence

Token loading first recognizes Text, exact lowercase `true` or `false`, quoted identifiers, and numeric candidates; then native types and configured vocabulary participate. A numeric-range candidate can remain provisional while source-defined names resolve. If a configured or source binding supplies that spelling, the compiler does not prematurely replace it with a numeric error. Unresolved explicit members remain reference failures.

This canonical observer demonstrates why name resolution matters:

<!-- example:lexical-numeric-name -->

The quoted label declares a Text binding named `1e999`. `Show 1e999` therefore reads `payload`; it is not an attempt to construct an overflowing Number. Without a resolving binding, the numeric spelling has a range error. The numeric-resolution test checks this binding and retained literal.

Grammar patterns such as `%~%` and `?~?` name typed token categories. They are not instructions to surround ordinary numbers and Truth literals with those characters. The actual source examples use `7`, `-12.5`, `true`, and `false`.

## Sources and witnesses

Implementations: [source preparation](../../lyger/source.go), [parser](../../lyger/parser.go), [literal unquoting](../../lyger/grammar.go), [token loading](../../lyger/lexicon.go), [lexical forms](../../lyger/lexical_types.go), [type and inflection utilities](../../lyger/utilities.go), and [definition slots](../../lyger/definition.go). Existing witnesses: [source framing](../../lyger/source_test.go), [paragraph positions](../../lyger/paragraph_source_test.go), [literal framing](../../lyger/literal_framing_regression_test.go), [numeric resolution](../../lyger/numeric_resolution_test.go), and [slot names](../../lyger/definition_slot_names_test.go).
