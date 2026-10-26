# Declarations, definitions and file boundaries

## A and B: declaration

A opens an ordinarily composed source file. Its package metadata becomes a comment and its Headers loop emits collected `use` imports. The source assembler adds graph-owned module/safety items at that boundary; A does not discover a module tree, initialize a runtime or resolve dependencies.

BA declares the supplied Keys and Type without an explicit value. The primitive does not invent `Default::default()`, zero initialization, `mut` or cloning. DA, KA, HA, GA, MA and PA forward one Calculation, Operation, Description, Evaluation, Instantiation or Iteration respectively through Declare. Their child retains its native expression/item/statement role.

The [Declaration vowel](../../../builtins/segments/vowels/declaration.lu) and [Declare primitive](../../../builtins/segments/consonants/declare.lu) own punctuation and selectors. [Declaration composition witnesses](../../../logiceme/testdata/lucid/declaration-operation.aml.txt) exercise composed source, not a textual replacement of a completed Rust file.

## O and F: definition

| Forms | Value/body source | Output role |
| --- | --- | --- |
| O, OB | Native Values | Assignment; OB also declares. |
| FO, FOB | Exactly one Declaration | Define using the supplied declaration composition. |
| KO, KOB | Exactly one Operation | Assign/bind the supplied callable operation. |
| DO, DOB | Exactly one Calculation | Assign/bind the calculated expression. |
| GO, GOB | Exactly one Evaluation | Assign/bind a selection result. |
| MO, MOB | Exactly one Instantiation | Assign/bind a constructed/invoked value. |
| CO, COB | Exactly one CV or DV | Assign / declare one binding from a future or awaited result. |
| BO | Keys and Type | Declaration without an initializer. |
| BOD/BOK/BOH/BOG/BOM/BOP/BOC | One Calculation/Operation/Description/Evaluation/Instantiation/Iteration/Execution | Expose the child's logical role through Declare. |

Definitions and declarations generally supply their own statement punctuation. W adds ordered template newlines, not implicit semicolons to arbitrary native leaves. Multiple Keys and native syntax remain caller-owned; the table does not promise automatic tuple-binding repair.

```text
DOB{|Keys: ~&[result]&~|}^DE{|Ops: ~&[Op{Type: Calculate, Operator: +, Operands: &[1, 2]&}Op]&~|}
```

This fragment binds a calculated result using the existing native Op representation. It needs complete AML context. The [grouping](../../../logiceme/testdata/lucid/grouping.aml.txt) and [calculations](../../../logiceme/testdata/lucid/calculations.aml.txt) witnesses preserve arithmetic grouping and exact output independently.

## Restricted definition fragments

OC and DOC supply complete `name = async { expression }` guards only inside BI selection alternatives. OC takes exactly one explicit nonempty native value; DOC takes exactly one Calculation and excludes native Values. They require one caller-supplied name or `_`; no name, readiness future or unwrap is synthesized.

OG and DOG provide `let pattern = expression` only as X loop conditions. OG consumes one explicit native value; DOG consumes one Calculation. Their Key/Keys pattern remains native text. PO places one caller-supplied label directly on one Iteration, without an added wrapper block or semicolon. See [selection](selection.md) and [iteration](iteration.md).

## EA: complete literal file

EA is an explicit escape for one complete literal file. Its Sentence/Object supplies the output path; Contents is required and copied verbatim, including explicitly empty content. EA stands alone without A or composed siblings. It is not an inline fragment or a shortcut for implementing a missing composition. Inherited generic/lifetime metadata is still rejected at this file boundary.

The normal corpus uses composed FW/W forms. A literal-file result does not prove the behavior of the composition it could replace. See [source-file assembly](../../../logiceme/source_file.go), [Definition](../../../builtins/segments/vowels/definition.lu), [Define](../../../builtins/segments/consonants/define.lu), and the maintained [literal-file contract](../../../builtins/CONTRACTS.md#literal-file-exception).
