# Conditional and readiness selection

## I and G: evaluation

I and GE can adapt Calculations with their existing Key/grouping selector. GI with no Key selects an ordered Boolean branch chain. GI with Key emits native pattern matching: ordered supplied pattern/body pairs and an optional explicit terminal Void fallback. It does not insert selector bindings, comparison guards or ownership decisions.

Native Rust checks pattern legality and exhaustiveness. [Keyed conditionals](../../../logiceme/testdata/lucid/keyed-conditionals.aml.txt) exercise positive composition. [Nonexhaustive matching](../../../logiceme/testdata/lucid/keyed-nonexhaustive.aml.txt) intentionally produces E0004; [call-pattern syntax](../../../logiceme/testdata/lucid/keyed-call-pattern.aml.txt) intentionally produces E0164. Those are successful generation with expected native rejection, not supported executable programs.

## CI and BI: bounded readiness selection

CI forwards ordered BI Evaluation children to biased Tokio selection. Each BI requires one OC/DOC Definition, zero or one W Operation body, and optional Default metadata. Ordinary branch order is preserved; the single explicit default is placed last. Default changes ordering only. It invents no binding, expression, readiness or timeout.

| Guard | Meaning |
| --- | --- |
| DOC with CE receive | Await the supplied receive; expose its explicit Option result under the caller's binding. |
| DOC with CE send | Await the supplied send; preserve its Result or explicitly discard with `_`. |
| DOC with another Calculation | Keep that expression as the async block's tail. |
| OC with explicit `()` | A caller-supplied ready alternative. |
| Empty CI | Remain pending using the existing future expression. |

OC is childless and takes one nonempty Value/Values. DOC takes one Calculation and excludes native Values. Both need one caller-supplied Key/Keys name or `_`. BI owns the selected body; guard bindings belong to that alternative. Bare calculations, DO/DOB guards, multiple guards/bodies/defaults and invalid placement reject before rendering.

## Limits and witnesses

The corpus covers [communication](../../../logiceme/testdata/lucid/select-communication.aml.txt), [bindings](../../../logiceme/testdata/lucid/select-bindings.aml.txt), [default placement](../../../logiceme/testdata/lucid/select-default-order.aml.txt), [empty/default-only forms](../../../logiceme/testdata/lucid/select-degenerate.aml.txt), [nesting](../../../logiceme/testdata/lucid/select-nesting-hygiene.aml.txt) and [closed receives](../../../logiceme/testdata/lucid/select-closed-receive.aml.txt). Thirty-five malformed inputs have classified failure assertions.

Native negative controls deliberately retain E0728 outside an async context, E0499 for conflicting borrows, and E0382 for a moved send value. Templates do not repair those Rust programs. Compilation does not verify runtime fairness, cancellation, destructor timing or scheduling behavior. No L source async API follows merely from these target forms.

See [Evaluation](../../../builtins/segments/vowels/evaluation.lu), [Evaluate](../../../builtins/segments/consonants/evaluate.lu), [selection validation](../../../logiceme/lucid_selection_test.go), and the maintained [CI/BI contract](../../../builtins/README.md#cibi-selection-contract).
