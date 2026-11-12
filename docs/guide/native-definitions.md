# Write native-backed definitions

Compose library algorithms in L and keep each native leaf to one bounded call, operator, projection, constructor, pattern or return. Lyger lowers grammatical control flow and bindings through frozen AML interfaces. These compiler instructions do not require library implementations.

## Start from the existing foundation

[Core](../../builtins/eng/core.l) supplies Process, Function, Invoke, Return, Main, Basic and Print. Invoke is a root verb for applying a selected native-backed source Function, while Return belongs to the selected Function constructor. Print composes a Process using its private Log action, whose native body performs one `println!` call. Print accepts Text and appends one newline. That leaf supplies output, not a runtime.

Before adding a leaf, identify the missing target operation, the typed inputs and result, and the policy that remains in L. Inspect any native callee: a single call to a new helper can conceal an entire algorithm.

## Keep branch data inside its Process

A Process supplies a composition body. Typed inputs and intermediate labels belong to that body and must remain distinct through repeated and nested use. Preserve once-only operand evaluation and definition-site binding. A Process is not a callable output or a capture runtime.

## Compose a wrapper

A wrapper must retain its canonical base, original operands and result identity. State moves, borrows, evaluation order and failure behavior. Do not silently clone a consumed value or communicate through fixed generated names. The former IO and Storage wrappers are absent and are not current examples.

## Typed slots cross an encoding boundary

Native definitions use a complete backtick-delimited `<DEF>…<DEF>` frame. [Typed slots](../spec/definitions.md#typed-slots) carry checked values, names, types and selected compositions through existing AML fields. A Label, a native Type and a Process body are different roles.

Text is encoded safely for native use, and Number uses finite binary64 encoding. Keep source bytes, source provenance and explicit ownership. Native syntax does not become authored L syntax.

## Apply a selected native Function

[Invoke](../reference/word-glossary.md#word-builtin-invoke) uses a Callable slot followed by terminal Atoms. Callable selects a source Function binding that has its own complete verb family and a native-backed source definition. Atoms carries that target's arguments in order, with the count checked against its signature. The pair is determined by slot roles and order, not by the labels Target and Arguments.

This native-authoring fragment declares Read with one native `f64` input, then uses the actor Invoker to apply Read to 7 inside Exercise. Each Function owns its Return. The maintained `TestCoreInvokeArguments` witness in [family registration tests](../../lyger/authored_family_registration_test.go) supplies the source pattern. This fragment is static source evidence, not a complete standalone App or an execution result.

<!-- example:native-invoke -->

A complete verb alias or copy can select the same native-backed Function. A single ordinary value label or noun alias cannot supply the required invocation family. Text such as a native name or path cannot stand in for the Callable binding. Ordinary runtime Function values and callbacks remain unavailable.

## Read a complete backend witness

The [declaration/evaluation witness](../../logiceme/testdata/lucid/declaration-evaluation.aml.txt) illustrates the backend interface. [Alphabet and composition](../reference/logemes/alphabet.md) explains the glyphs. Generation, native compilation and execution are distinct observations; a retained witness does not establish a fresh test result.

## Stay inside the approved surface

The eighteen Rust segments and four logeme templates are frozen. Preserve their bytes and interfaces. Lifetime/generic inputs are glyph-local native inputs, not new Func/Prop/Op fields or inherited defaults. Deferred metadata rejects.

Do not hide executable algorithms or Function bodies in Statement, EA, opaque closures or native helpers. The [builtin](../../builtins/CONTRACTS.md), [Rust](../../builtins/README.md) and [AML](../../logiceme/AML.md) contracts define the retained surface. Runtime reconstruction is separate work in L.
