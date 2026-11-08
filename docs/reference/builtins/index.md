# Builtin reference

Builtins are available from the selected complete distribution without an explicit Use import. They comprise configured lexical facilities, compiler-typed operations, and definitions authored in `builtins/eng/core.l`. IO.Print belongs to the separately imported [standard library](../../stdlib/index.md).

## Source-defined facilities

<!-- word-index:builtins -->

The [construction and callable reference](definitions.md) explains their inputs and limits. Actor forms such as Invoker and Retainer construct uses of their verb. They are not separate primitives. Aliases preserve canonical identity and adjectives retain their declared targets. See [word forms](../../spec/definitions.md#inflection-declarations) for the declaration notation.

## Compiler-typed operations

| Operation | Forms from configured root | Operands and behavior |
| --- | --- | --- |
| Build | Build, Builds, Built, Building, Builder | Validate a checked App's prebuilt artifact at runtime |
| Run | Run, Runs, Ran, Running, Runner | Launch a checked target without joining |
| Await | Await, Awaits, Awaited, Awaiting | Join a checked target's finalized result and claim its continuation once |
| Stop | Stop, Stops, Stopped, Stopping, Stopper | Request termination of a checked target without joining |
| Show | Show, Shows, Showed, Showing, Shower | Observe supported runtime values without adding a newline |
| Exit | Exit, Exits, Exited, Exiting, Exiter | Terminate whole program, with optional supported failure operand |

These operations are verbs. Inflection recognition does not make every grammatical owner valid: native Await inside an App body rejects even though its word form is recognized. The [lifecycle reference](lifecycle.md) records supported contexts and failure modes.

## Declaration and vocabulary facilities

Create, Define, Describe, and Use establish declarations, typed/native definitions, descriptors, aliases, and supported imports. Public/Publicly controls visibility at one declaration site. New and Generic participate in construction. Noun/Verb/Adjective/Adverb are lexical categories, not native scalar types.

[Collections](../../spec/collections-and-categories.md) support Put, Place and typed reads. Get, Take and While remain unsupported operations. The [grammar chapter](../../spec/grammar.md) distinguishes executable actions from other configured vocabulary.

## Runtime members

Output and Error are noun members: Output contains finalized raw stdout as Text, and Error identifies the selected process operation/invocation failure. Possessive access keeps the canonical target. Input instead names signature parameters and typed Process bindings. It is not an implicit App member.

### Input

<!-- word:builtin:input -->

Name an ordered Function parameter or a typed binding supplied to a structured Process.

**Usage and example.** `with Inputs of "Value" as Text` declares Identity's parameter in the [complete Function example](../../examples/composed-functions.l). A selector's Process Inputs bind its selected payload inside that body. `No Inputs` explicitly declares no parameters. These forms do not provide stdin or argument transport to an App. The [execution limits](../../spec/execution.md#targets-and-process-ownership) remain unchanged.

Function calls check input count and types. A Process input belongs only to its selected body, with native borrow and move checks at compilation. These supported bindings introduce neither an untyped argument list nor an implicit clone.

## Availability and errors

English builtins and their backend resources are embedded and loaded automatically. Compiler developers can [override the resource bundle](../../development-harness.md#compiler-defaults).

Wrong action parameter counts/types, unsupported target owners, unknown descriptor targets, inaccessible source declarations, and cycles are real errors. Builtin definitions do not hide lifecycle misuse, insert Await, trim output, or infer initialization. The [specification](../../spec/index.md) separates their source contract from backend-native facilities.

## Sources and witnesses

Sources: [core.l](../../../builtins/eng/core.l), [vocabulary](../../../lexicon/eng/vocabulary.json), [typed process operations](../../../lyger/builtin_operations.go), and [native instruction roles](../../../lyger/ir_kinds.go). Witnesses: [builtin operation contract](../../../lyger/builtin_operations_contract_test.go), [operation validation](../../../lyger/builtin_operations_validation_test.go), [action definitions](../../../lyger/action_definition_test.go), and [Service boundary coverage](../../../testdata/service-execution/coverage.md).
