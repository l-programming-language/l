# Instantiation, records, traits and scheduling

## Y and M: records and values

Y constructs a named value from Props; HY emits a struct-shaped record declaration from Props. Empty or omitted Props preserve braces: `Key {}` and `struct Key {}` respectively. They do not silently become unit structs. Nonempty fields retain their supplied order, names and native types/values.

Empty HY accepts absent Shape or Record; Sum rejects. Empty HY/MY collections cannot accompany composed member/header children the wrappers would discard. These guards do not establish composed field/method implementations, enums or sum types. KY invocation is described with [callables](callables.md#invocation-through-ky).

The [empty-instantiations witness](../../../logiceme/testdata/lucid/empty-instantiations.aml.txt) covers omitted/explicit-empty records, construction, populated controls and traits. Five stored malformed inputs exercise unsupported combinations. [Instantiation](../../../builtins/segments/vowels/instantiation.lu), [Instantiate](../../../builtins/segments/consonants/instantiate.lu) and [empty validation](../../../logiceme/lucid_empty_validation_test.go) own this implementation boundary.

## Required trait signatures

MY emits receiverless required trait signatures from ordered Funcs. Each signature has a native name and named/typed inputs. Missing or empty Outputs omit the return annotation; supplied outputs retain their punctuation and terminating semicolon. Nested native type commas are decoded by the existing record machinery.

Any supplied Receiver field rejects, including empty Receiver. Trait-level Lifetimes/GenericTypes qualify the trait once, not each method. Default/composed methods, separate method parameter lists and implementation blocks remain outside the supported contract. Empty Funcs emit `trait Key {}`.

The [trait-signature witness](../../../logiceme/testdata/lucid/trait-signature-outputs.aml.txt) includes native names/types spelled Void and eight required signatures. Native compilation checks declarations and type syntax; it executes no trait behavior.

## V and C: scheduling

V schedules the invocation text in Key through `tokio::spawn`. MV schedules exactly one Instantiation child. The supplied invocation must already return a suitable future. Templates add no async wrapper, runtime, scope-exit callback or automatic join.

The [direct scheduling](../../../logiceme/testdata/lucid/schedule-direct.aml.txt) and [composed scheduling](../../../logiceme/testdata/lucid/schedule-composed.aml.txt) examples supply their own native scaffolding. They are compile-only witnesses. L Apps use the separately specified OS-process lifecycle; Tokio tasks do not replace that contract.

## Deferred is rejected

Rust Execute/Execution exclude Deferred from inputs, forwarding and rendering. Supplying Deferred is unsupported even when its value is False. There is no deferred-call or destructor/RAII cleanup contract inferred from V/MV. Four malformed cases and [scheduling tests](../../../logiceme/lucid_scheduling_test.go) preserve this rejection.

Read [Execution](../../../builtins/segments/vowels/execution.lu), [Execute](../../../builtins/segments/consonants/execute.lu), and [execution without Deferred](../../../builtins/README.md#execution-without-deferred). Runtime launch, Stop, Await and cleanup belong to the [CLI/runtime contract](../../cli-build.md), not these scheduling templates.
