# Backend support and evidence

## Current corpus

The [Lucid suite guide](../../../logiceme/testdata/lucid/README.md#current-inventory) and [registered cases](../../../logiceme/lucid_generation_test.go) define 59 applications and 295 exact artifact comparisons. Each Expected/Generated native phase requires 45 successful compilations and 14 specified Rust errors. Three generated future cases also execute assertions.

| Family | Positive or native-negative examples | Main boundary |
| --- | --- | --- |
| Baseline | grouping, statements, fields, calculations, communication, conditionals | Ordered composed source and native fragments. |
| Selection | select-communication/bindings/default-order/degenerate/nesting-hygiene/closed-receive; three native negatives | CI/BI/OC/DOC and explicit native async/ownership constraints. |
| Iteration | conditions, bindings, traversal, labels, nesting-scope, async; three native negatives | Four loop categories, labels and native scope. |
| Declaration compositions | operation, description, evaluation, instantiation, iteration | Singular child forwarding. |
| Invocation/names | invocation-named/void, function-void-flat/record | Receiver absence and native name provenance. |
| Callable/traits/empty | typed flat/record, outputs, closures, unit mismatch, Void output, trait outputs, empty instantiations | Bounded Func/Props, optional outputs and empty shells. |
| Keyed patterns | keyed-conditionals plus two native negatives | Ordered patterns and explicit fallback. |
| Generics/lifetimes | functions, records, traits, calls plus three native negatives | Local native interiors, no inference. |
| Scheduling | schedule-direct/composed | Supplied futures, generation and compilation only. |
| Futures | composed, move, infer; borrow-escape and move-reuse negatives | CV/DV/CO/COB, native ownership rejection and bounded execution. |

## Rejection corpus

There are 192 stored malformed compositions: 62 callable, 35 selection, 50 iteration, five empty-composition, four Deferred, two invocation, fourteen keyed-GI and twenty generic/lifetime cases. Existing tests assert classified causes, unchanged inputs, zero returned bundles and no panic. Fifty focused generic/lifetime primitive profiles supplement complete applications; they are not the entire low-level test inventory.

The [frozen manifest](../../../logiceme/testdata/lucid/templates.sha256.json) records all eighteen segment hashes. Documentation may inventory these hashes but cannot update them to accommodate source changes. Expected Rust/source-support files are independent oracles, not regenerated output.

## What the evidence does not say

An accepted grammar or template route is not operational support. Generated source is not a compiled program. Compile-only selection and scheduling cases do not establish runtime fairness, cancellation, cleanup or lifetime inference. Generated start/stop scripts remain unexecuted. Supervised frontend/runtime tests are a separate boundary recorded in the [integration ledger](../../../testdata/integration/IMPLEMENTATION.md).

The [latest recorded backend ledger](../../../logiceme/testdata/lucid/FAILURES.md) describes the successful October 15 macOS ARM64 run. Its results apply to that source identity and host. This website's checks establish documentation inventory/freshness and static-site behavior only.

## Integration limits

Deferred and three-part loops are unsupported. CV/DV and CO/COB provide [bounded future composition](../../../builtins/README.md#future-construction-awaiting-and-value-bindings). Function receiver methods, async-FW modifiers, general enums/patterns, implementations and method generic parameters remain outside the [maintained contracts](../../../builtins/README.md#current-integration-limits). Rust checks native ownership; unused compiler branches establish no further interface.

## Support assets

Config and Cargo templates assemble application configuration and manifests. Start/stop templates produce deployment artifacts; runtime/main, runtime/runtime and Unix/other process templates supply lifecycle support. They are separate from the eighteen segments and from L's public standard library.
