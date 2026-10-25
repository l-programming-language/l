# Calculations and descriptions

## E and D: calculations

E emits `Meta.Statement` as a supplied native fragment. DE renders ordered `Meta.Ops`; an Op has Prefix, Type, Operands, Operator and Postfix. The existing `Type == Calculate` selector participates in grouping. Native operands/operators remain explicit data: this is not an alternate parser for L expressions and does not imply coercion or arithmetic rules beyond the frontend's typed contract.

GE forwards Calculations and Key to Evaluate; I has a related adapter and grouping selector. Parentheses and ordered Op records matter. The [grouping](../../../logiceme/testdata/lucid/grouping.aml.txt) witness retains `2 * (1 + 2)`; [calculations](../../../logiceme/testdata/lucid/calculations.aml.txt) retains multiple Ops, prefixes and postfixes. These witnesses compare exact Rust bytes before separate native compilation.

## Communication calculations

CE interprets the existing operand model: one operand receives using `.recv().await`; two send using `.send(...).await`. The caller supplies the channel expression/payload and an appropriate async context. CE neither unwraps nor otherwise consumes the returned Option/Result and does not synthesize a binding.

The [communication witness](../../../logiceme/testdata/lucid/communication.aml.txt) explicitly creates its Tokio runtime, async block and channel in native fragments. Its success does not establish L source-level concurrency, V/MV scheduling, readiness selection, runtime fairness or cancellation behavior. CI/BI has separate [selection contracts](selection.md).

## U and H: descriptions

U describes ordered Operations. MU describes updates through the existing Props loop, preserving field order and supplied values. Description is structural composition, not an assignment-scoping or ownership-inference mechanism. A Prop carries Key, Type, Value and StructTags; its role depends on the consuming form.

Record field declarations and constructors use separate [HY and Y forms](instantiation.md#y-and-m-records-and-values). Rust StructTags are retained as comments rather than interpreted as attributes. Raw multiline tags are not normalized into safe multiline comments; native authors must respect that limitation.

The [fields witness](../../../logiceme/testdata/lucid/fields.aml.txt) combines a native record declaration with composed construction. It establishes that particular construction's fields and ordering, not every structural-description route. The [declaration-description witness](../../../logiceme/testdata/lucid/declaration-description.aml.txt) exercises the declaration adapter.

## Source and evidence

Read [Calculation](../../../builtins/segments/vowels/calculation.lu), [Calculate](../../../builtins/segments/consonants/calculate.lu), [Description](../../../builtins/segments/vowels/description.lu) and [Describe](../../../builtins/segments/consonants/describe.lu). [Op/Prop decoding](../../../logiceme/segmentals.go) and [glyph role validation](../../../logiceme/glyph_roles.go) preserve the separation between literal native metadata and structured child roles.
