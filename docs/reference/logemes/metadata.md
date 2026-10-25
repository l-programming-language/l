# Metadata and native input contracts

## Records and conjoinments

Meta supplies scoped data to a glyph. Conjoinments supply composed child values under explicitly forwarded logical roles. Singular/plural forwarding and child order are part of the consuming form's interface. A preparation field or an inherited value is not automatically an authoring contract.

| Record | Fields | Typical consumers |
| --- | --- | --- |
| Func / Funcs | Key, Inputs (Props), Outputs (native strings), Receiver (Prop) | FW/BW signatures and MY required signatures. |
| Prop / Props | Key, Type, Value, StructTags | Typed parameters, fields, construction and updates; allowed fields vary by form. |
| Op / Ops | Prefix, Type, Operands, Operator, Postfix | Calculation and communication. |

The [record implementation](../../../logiceme/segmentals.go) decodes these structures. For a function parameter only Key and Type are allowed; record-field existence does not grant permission to supply Value or StructTags there. Metadata interpretation preserves literal spans and template-owned separators. It does not reparse inserted source as new Lucid control instructions.

## Absence and source provenance

The backend has an absence sentinel spelled Void, but supplied native identifiers/types named `Void` must remain data. Rust preparation tracks SourceString provenance for the specifically supported native fields. Exact `Meta.Func.Key`, `Meta.Func.Outputs~` and indexed MY Key/Outputs have bounded controls; that does not prove arbitrary indexed-field provenance.

Canonical absence, omitted collections and explicit empty collections depend on the consuming form. Unknown or malformed records must not silently become empty values. An inherited metadata default and glyph-local override are distinct origins; service reuse must not leak either into prior or later bundles.

## Lifetimes and GenericTypes

The only approved generic/lifetime interfaces are glyph-local native String list interiors:

```text
Lifetimes: ~'a, 'b~
GenericTypes: ~T, U: Clone + 'a~
```

FW/HY/MY declare parameters; Y and named KY apply arguments. Both absent emits nothing. Both supplied places Lifetimes first and adds one template-owned separator. Callers own commas inside either value. Empty/whitespace-only supplied inputs reject; native text Void remains a supplied type.

These values are not Func/Prop/Op fields and cannot be inherited Class/Object/file metadata—even with a local override or an EA literal file. BW, anonymous KY and other forms reject them. Parent lists do not propagate into child calls. Named KY attaches arguments to its final named component: Key `Vec::new` plus GenericTypes `u8` yields `Vec::new::<u8>()`.

Rust validates syntax, bounds, arity and lifetime legality. There is no parsing of a new generic parameter language, ownership/lifetime inference, implicit borrowing/cloning, added PhantomData or const-generic/where-clause interface. The [generic witnesses](../../../logiceme/testdata/lucid/generics-functions.aml.txt) and [validation tests](../../../logiceme/lucid_generics_validation_test.go) retain native E0794/E0392/E0107 failures where appropriate.

## Imports and support

Imports remain file-owned Headers, separate from module declarations and Cargo dependencies. Preparation gathers nested imports before the file opener renders. Runtime support selection belongs to Evaluate options/source assembly, not arbitrary glyph metadata. See [metadata validation](../../../logiceme/glyph_metadata_validation.go), [preparation](../../../logiceme/glyph_prepare.go) and the maintained [local-input contract](../../../builtins/README.md#lifetime-and-generic-inputs).
