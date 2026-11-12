# Operations and control flow

The grammar retains sequencing, assignment and condition forms. The current executable boundary is a standalone App with core Print; runtime-dependent top-level control flow rejects. Recognition alone does not establish execution.

## Sequencing and evaluation order

Preserve original source order, once-only argument evaluation and definition-site binding. A composed source action must retain its selected operands and body placement. Native template assembly is not a universal evaluation-order guarantee.

### Then

[Then in the Word Glossary](../reference/word-glossary.md#word-grammar-word-then).

### And

[And in the Word Glossary](../reference/word-glossary.md#word-grammar-word-and).

### Or

[Or in the Word Glossary](../reference/word-glossary.md#word-grammar-word-or).

## Assignment

The basic program binds static Text for Print. Runtime assignment, capture and Output snapshots are not current executable facilities. Never manufacture default values to hide missing assignment.

### Is

[Is in the Word Glossary](../reference/word-glossary.md#word-grammar-word-is).

## Comparisons

Configured comparison words distinguish equality, order and negation. Their recognizers do not supply runtime conditions. Type checking must not silently coerce mixed values or infer general truthiness.

### Have

[Have in the Word Glossary](../reference/word-glossary.md#word-grammar-word-have).

### Equal

[Equal in the Word Glossary](../reference/word-glossary.md#word-grammar-word-equal).

### Not

[Not in the Word Glossary](../reference/word-glossary.md#word-grammar-word-not).

### Less

[Less in the Word Glossary](../reference/word-glossary.md#word-grammar-word-less).

### Greater

[Greater in the Word Glossary](../reference/word-glossary.md#word-grammar-word-greater).

## Selection

If, Else and Otherwise retain grammatical ownership rules. Paragraph boundaries and clause adjacency must survive parsing and expansion. Their presence does not restore the removed runtime executor.

### If

[If in the Word Glossary](../reference/word-glossary.md#word-grammar-word-if).

### Else

[Else in the Word Glossary](../reference/word-glossary.md#word-grammar-word-else).

### Otherwise

[Otherwise in the Word Glossary](../reference/word-glossary.md#word-grammar-word-otherwise).

## Iteration and collection operations

No bundled Storage traversal or iteration words remain. While recognition and native iteration templates do not establish general source loops, break/continue or runtime collection updates.

### While

[While in the Word Glossary](../reference/word-glossary.md#word-grammar-word-while).

## Arithmetic and native calculations

Number literals do not introduce general infix arithmetic syntax. D/E native calculations and G/I selection retain their separate [logeme contracts](../../builtins/README.md). Source words must compose policy explicitly.

## Show and output

Show and captured Output inspection are unavailable. Core Print accepts Text and writes it followed by one newline from the standalone App. See the [first program](../guide/first-program.md).

## Exit and Return

Runtime Exit is unavailable. Core Function's Return is a callable boundary, not an orchestrator exit or a process result transport. Keep the selected callable's identity and typed results.

## Sources and witnesses

See [grammar](../../lexicon/eng/grammar.json), [instruction construction](../../lyger/instructions.go), [definitions](definitions.md) and [execution](execution.md). Historical runtime tests are outside the temporary gate.
