# Backend support and evidence

## Current corpus

The [Lucid suite guide](../../../logiceme/testdata/lucid/README.md#current-inventory) owns the case inventory and generation/native-compilation obligations. Three positive future cases additionally execute generated assertions. The [generated inventory](../../_generated/lucid.md) lists the registered cases.

## Rejection corpus

Malformed compositions must return classified errors without changing inputs, returning partial bundles, or panicking. The Lucid guide records the rejection families and focused checks. Its [frozen manifest](../../../logiceme/testdata/lucid/templates.sha256.json) protects the eighteen segment definitions; expected artifacts remain independent oracles.

## What the evidence does not say

Generation, compilation, and execution establish different facts. Compile-only selection and scheduling cases do not establish runtime fairness, cancellation, cleanup, or lifetime inference. Generated start/stop scripts remain unexecuted. See the [backend ledger](../../../logiceme/testdata/lucid/FAILURES.md) and separate [frontend/runtime ledger](../../../testdata/integration/IMPLEMENTATION.md) for results tied to specific revisions and hosts.

## Integration limits

Deferred and three-part loops are unsupported. CV/DV and CO/COB provide [bounded future composition](../../../builtins/README.md#future-construction-awaiting-and-value-bindings). Function receiver methods, async-FW modifiers, general enums/patterns, implementations, and method generic parameters remain outside the [maintained contracts](../../../builtins/README.md#current-integration-limits).

## Support assets

Config/Cargo templates supply configuration and manifests; runtime and process templates supply lifecycle support. Start/stop templates produce deployment artifacts. These assets are separate from the eighteen segments and L's public standard library.
