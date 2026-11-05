# Builtin reference

Builtins are available from the selected complete distribution without an explicit Use import. They comprise configured lexical facilities, compiler-typed operations, and definitions authored in `builtins/eng/core.l`. IO.Print belongs to the separately imported [standard library](../../stdlib/index.md).

## Source-defined facilities

| Word and kind | Declared forms | Contract |
| --- | --- | --- |
| Start — verb | Start, Starts, Started, Starting, Starter | Build any App, then Run it |
| Execute — verb | Execute, Executes, Executed, Executing, Executor | Start any App, then Await it |
| Process — noun | Process, Processes, Processive, Processing | Compose ordered definition expressions |
| Function — noun | Function, Functions, Functional, Functioning | Define a callable with typed inputs, outputs and body |
| Function.Invoke — verb | Invoke, Invokes, Invoked, Invoking, Invoker | Apply a selected callable to ordered argument references |
| Function.Return — verb | Return, Returns, Returned, Returning, Returner | Complete the current Function with its checked results |
| Delegate — verb | Delegate, Delegates, Delegated, Delegating, Delegator | Call the selected native target |
| Retain — verb | Retain, Retains, Retained, Retaining, Retainer | Borrow an Arc and return another handle to the same allocation |
| Objective — adjective | Objective | Group a Function's source definition |
| Main — adjective | Main | Construct the Service/Object/Function entry structure |
| Basic — adjective | Basic | Construct an App with a Main Service using a Process |

The [construction and callable reference](definitions.md) explains their inputs and limits. Actor forms such as Invoker and Retainer construct uses of their verb; they are not separate primitives. Aliases preserve canonical identity and adjectives retain their declared targets. See [word forms](../../spec/definitions.md#inflection-declarations) for the declaration notation.

## Compiler-typed operations

| Operation | Forms from configured root | Operands and behavior |
| --- | --- | --- |
| Build | Build, Builds, Built, Building, Builder | Checked App; validate prebuilt artifact at runtime |
| Run | Run, Runs, Ran, Running, Runner | Checked target; launch without join |
| Await | Await, Awaits, Awaited, Awaiting | Checked target; join finalized result and claim continuation once |
| Stop | Stop, Stops, Stopped, Stopping, Stopper | Checked target; request termination without join |
| Show | Show, Shows, Showed, Showing, Shower | Observe supported runtime values; no added newline |
| Exit | Exit, Exits, Exited, Exiting, Exiter | Terminate whole program, with optional supported failure operand |

These operations are verbs. Inflection recognition does not make every grammatical owner valid: native Await inside an App body rejects even though its word form is recognized. The [lifecycle reference](lifecycle.md) records supported contexts and failure modes.

## Declaration and vocabulary facilities

Create, Define, Describe, and Use establish declarations, typed/native definitions, descriptors, aliases, and supported imports. Public/Publicly controls visibility at one declaration site; New and Generic participate in construction. Noun/Verb/Adjective/Adverb are lexical categories, not native scalar types.

[Collections](../../spec/collections-and-categories.md) support Put, Place and typed reads. Get, Take and While remain unsupported operations. The [grammar chapter](../../spec/grammar.md) distinguishes executable actions from other configured vocabulary.

## Runtime members

Output and Error are noun members: Output contains finalized raw stdout as Text; Error identifies the selected process operation/invocation failure. Possessive access keeps the canonical target. Input exists in vocabulary/signature metadata but runtime process argument/input transport is not provided by the current App contract.

## Availability and errors

English builtins and their backend resources are embedded and loaded automatically. Compiler developers can [override the resource bundle](../../development-harness.md#compiler-defaults).

Wrong action parameter counts/types, unsupported target owners, unknown descriptor targets, inaccessible source declarations, and cycles are real errors. Builtin definitions do not hide lifecycle misuse, insert Await, trim output, or infer initialization. The [specification](../../spec/index.md) separates their source contract from backend-native facilities.

## Sources and witnesses

Sources: [core.l](../../../builtins/eng/core.l), [vocabulary](../../../lexicon/eng/vocabulary.json), [typed process operations](../../../lyger/builtin_operations.go), and [native instruction roles](../../../lyger/ir_kinds.go). Witnesses: [builtin operation contract](../../../lyger/builtin_operations_contract_test.go), [operation validation](../../../lyger/builtin_operations_validation_test.go), [action definitions](../../../lyger/action_definition_test.go), and [Service boundary coverage](../../../testdata/service-execution/coverage.md).
