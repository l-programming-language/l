# Metadata and native input contracts

## Records and conjoinments

Meta supplies scoped data to a glyph. Conjoinments supply composed child values under explicitly forwarded logical roles. Singular/plural forwarding and child order are part of the consuming form's interface. A preparation field or an inherited value is not automatically an authoring contract.

| Record | Fields | Typical consumers |
| --- | --- | --- |
| Func / Funcs | Key, Inputs (Props), Outputs (native strings), Receiver (Prop) | FW/BW signatures and MY required signatures. |
| Prop / Props | Key, Type, Value, StructTags | Typed parameters, fields, construction and updates. Allowed fields vary by form. |
| Op / Ops | Prefix, Type, Operands, Operator, Postfix | Calculation and communication. |

The [record implementation](../../../logiceme/segmentals.go) decodes these structures. For a function parameter only Key and Type are allowed. Record-field existence does not grant permission to supply Value or StructTags there. Metadata interpretation preserves literal spans and template-owned separators. It does not reparse inserted source as new Lucid control instructions.

## Absence and source provenance

The backend has an absence sentinel spelled Void, but supplied native identifiers/types named `Void` must remain data. Rust preparation tracks SourceString provenance for the specifically supported native fields. Exact `Meta.Func.Key`, `Meta.Func.Outputs~` and indexed MY Key/Outputs have bounded controls. That does not prove arbitrary indexed-field provenance.

Canonical absence, omitted collections and explicit empty collections depend on the consuming form. Unknown or malformed records must not silently become empty values. An inherited metadata default and glyph-local override are distinct origins. Service reuse must not leak either into prior or later bundles.

## Lifetimes and GenericTypes

The [Rust input contract](../../../builtins/README.md#lifetime-and-generic-inputs)
defines glyph-local native String list interiors:

```text
Lifetimes: ~'a, 'b~
GenericTypes: ~T, U: Clone + 'a~
```

FW/HY/MY declare parameters; Y and named KY apply arguments. These values sit
beside Func/Prop/Op records and are never inherited file defaults. The canonical
contract specifies absence, punctuation, provenance and the native-validation
boundary. The [generic witnesses](../../../logiceme/testdata/lucid/generics-functions.aml.txt)
and [validation tests](../../../logiceme/lucid_generics_validation_test.go) retain
native E0794/E0392/E0107 failures where appropriate.

## Imports and support

Imports remain file-owned Headers, separate from module declarations and Cargo dependencies. Preparation gathers nested imports before the file opener renders. Retained Evaluate options do not restore absent runtime assets, and arbitrary glyph metadata cannot select a runtime. See [metadata validation](../../../logiceme/glyph_metadata_validation.go), [preparation](../../../logiceme/glyph_prepare.go) and the maintained [local-input contract](../../../builtins/README.md#lifetime-and-generic-inputs).
