# Work with values

Start with backtick-delimited Text, numeric literals, and `true`/`false` Truth values. Double-quoted declaration names identify values; they are not Text literals.

## Text, Number and Truth together

<!-- example:public-values -->

The three public declarations survive the paragraph boundary. `Show Greeting`, `Show Count`, and `Show Enabled` write `public7true` without spaces or line breaks. The Dictionary provides a vocabulary namespace.

## Read pipe-delimited word forms

`"Display|-s|-ed|-ing|-er"` gives one action several grammatical forms: Display, Displays, Displayed, Displaying and Displayer. Each ending applies to the root. L uses these explicit forms rather than guessing English spelling. See [vocabulary and inflection](../spec/lexical-structure.md#vocabulary-and-inflection) for noun/verb forms and spelling rules.

## Text is data, including whitespace

<!-- example:guide-text-value -->

`message` receives the Text `Hello World!`, including its space, as in the [first program](first-program.md).

Text preserves its bytes. Periods and whitespace inside it are not sentence boundaries. A backslash followed by `n` remains two characters; it is not a newline escape.

`IO.Print` adds a newline; `Show` writes Text exactly. Captured App Output is also Text. Read it after completion; a stored copy keeps its bytes even if the App runs again.

## Numbers have a concrete representation

<!-- example:guide-number-value -->

`Count` receives the Number `7`, written without quotes or backticks. `Publicly` makes the name available after the paragraph.

Number uses finite binary64 values, so large integers and fractions can round. See the [Number reference](../spec/values.md#number) for accepted spellings, precision and limits.

## Other categories require their own contracts

Use [Primitive Types](../spec/values.md), [Iterative Types](../spec/collections-and-categories.md) and [Applicational Types](../spec/applicational-types.md) for construction, copying and sharing rules.

Continue with [names and paragraphs](names-and-paragraphs.md).
