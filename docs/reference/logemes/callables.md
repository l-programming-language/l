# Operations and callable forms

## W and K: ordered operations

W emits ordered Definitions. FW defines a named function using Func metadata and ordered Operations; BW forms an anonymous synchronous closure. The templates own Rust function/closure punctuation, while preparation validates signature records and forwards explicit children. They add no async wrapper, implicit move, cloning or implementation block.

FW requires a nonempty name and named, typed input Props. BW requires typed input Props too; it does not expose general inferred-parameter syntax. Inputs permit Key/Type, not supplied Value, StructTags or unknown fields. Flat signature metadata and an explicit Func record are exclusive. Func has Key, Inputs, Outputs and Receiver, but a retained field does not mean every form consumes it.

## Results and absence

Omitted or empty Outputs omit the entire return annotation. A named function then returns unit; a closure uses Rust's output inference. Explicit outputs preserve their existing parenthesized punctuation, including native types named `Void`. The absence sentinel differs from supplied native String `Void`.

The [typed flat](../../../logiceme/testdata/lucid/callable-typed-flat.aml.txt), [typed record](../../../logiceme/testdata/lucid/callable-typed-record.aml.txt), [output](../../../logiceme/testdata/lucid/callable-outputs.aml.txt) and [Void output](../../../logiceme/testdata/lucid/callable-output-void.aml.txt) cases witness those boundaries. [Unit mismatch](../../../logiceme/testdata/lucid/callable-unit-mismatch.aml.txt) intentionally produces E0308, proving that omitted output metadata does not cause a hidden return-type repair.

## Invocation through KY

Named KY consumes its Key and Inputs, with no callable child. Absent Receiver emits `Key(Inputs)`; a supplied nonblank Receiver emits `Receiver.Key(Inputs)`. Empty/whitespace Receiver rejects. Native Key/Receiver named `Void` remain actual native data.

Unnamed KY requires exactly one BW child and no Receiver. BW is allowed only directly under KY or KO/KOB, within enclosing callable scope. KO/KOB assign or bind the resulting closure. Native Rust infers captures; a Capture interface, receiver methods, FWC/KYC semantics and general ownership analysis are not established by these witnesses.

The [inline closure](../../../logiceme/testdata/lucid/closure-inline.aml.txt), [stored borrowed capture](../../../logiceme/testdata/lucid/closure-stored-capture.aml.txt), [named invocation](../../../logiceme/testdata/lucid/invocation-named.aml.txt) and [Void invocation](../../../logiceme/testdata/lucid/invocation-void.aml.txt) are complete compile-only examples. Their canonical AML is not L source syntax.

## Local native parameter lists

FW may declare glyph-local Lifetimes/GenericTypes; named KY may apply them to the final path component. BW and anonymous KY cannot. These native String interiors belong to [metadata](metadata.md#lifetimes-and-generictypes), not an L closure/generics API. Trait signatures have a separate [MY contract](instantiation.md#required-trait-signatures).

See [Operation](../../../builtins/segments/vowels/operation.lu), [Operate](../../../builtins/segments/consonants/operate.lu), [callable validation](../../../logiceme/lucid_callable_test.go), and the [maintained callable contract](../../../builtins/README.md#typed-functions-and-anonymous-callables).
