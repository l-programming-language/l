# Primitive Types

The compiler distinguishes primitive value categories from lexical roles and executable support. The current standalone App uses Text with core Print. Other retained type tags do not imply a bundled runtime or formatting library.

## Declarations, assignment, and reading values

Double-quoted labels declare names; backticks delimit Text. The [basic program](../guide/first-program.md) binds `message` to Text and supplies that binding to Print. Preserve source identity and typed operands rather than treating every token as a string.

## Text

[Text](../reference/word-glossary.md#word-type-text) preserves its bytes, including whitespace. A literal backslash followed by `n` is not a newline escape. Core Print accepts Text and adds one newline. There is no captured Output interface.

## Number

[Number](../reference/word-glossary.md#word-type-number) uses finite binary64. Large integers and fractions can round. Native definition operands use checked encoding with an explicit `f64` representation; malformed and nonfinite inputs retain their error boundary. Number is not accepted by Text-only Print.

## Truth

[Truth](../reference/word-glossary.md#word-type-truth) represents boolean values. It is distinct from Number and Text. Its recognition does not provide a runtime conditional executor.

## Thing

[Thing](../reference/word-glossary.md#word-type-thing) is a retained general value category. It does not permit Print to accept arbitrary values or supply the removed formatting policy.

## Byte and Bytes

### Byte

[Byte](../reference/word-glossary.md#word-type-byte) is a configured type meaning; it is distinct from a textual character.

### Bytes

[Bytes](../reference/word-glossary.md#word-type-bytes) is a retained internal representation. There is no bundled Storage/IO byte API.

## Pointer

[Pointer](../reference/word-glossary.md#word-type-pointer) remains a distinct compiler category. Its presence does not establish an executable sharing or lifetime-management API in the reset.

## Null and internal states

### Null

[Null](../reference/word-glossary.md#word-type-null) is distinct from unresolved compiler state and from a native null pointer. Unknown/None states must not become successful ordinary values.

## Error

Runtime Error inspection is unavailable. Compiler failures use the separate [error model](../error-model.md).

## Conversion and aliasing boundaries

Keep canonical identity, value representation, borrow/move behavior and conversion policy explicit. Do not add implicit formatting, cloning or type coercion to compensate for a removed library.

## Sources and witnesses

See [value kinds](../../lyger/value_kinds.go), [stored values](../../lyger/types.go), [native encoding](../../lyger/literal_atoms.go) and [definition bindings](../../lyger/definition_bindings.go). Retained tests supply assertions, not evidence of a fresh run.
