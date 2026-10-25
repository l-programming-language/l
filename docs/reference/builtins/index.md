# Builtin reference

Builtins are available from the selected complete distribution without an explicit Use import. They comprise configured lexical facilities, compiler-typed operations, and definitions authored in `builtins/eng/core.l`. IO.Print belongs to the separately imported [standard library](../../stdlib/index.md).

## Source-defined facilities

| Facility | Declared forms | Contract |
| --- | --- | --- |
| Start | Start, Starts, Started, Starting, Starter | Build any App, then Run it |
| Execute | Execute, Executes, Executed, Executing, Executor | Start any App, then Await it |
| Process | Process, Processes, Processive, Processing | Compose ordered definition syllables into operation W |
| Function | Function, Functions, Functional, Functioning | Compose FW with typed native signature/body inputs |
| Function.Return | Return, Returns, Returned, Returning, Returner | Native-backed return statement inside the supplied Function composition |
| Objective | Objective | Wrap a Function's source grouping |
| Main | Main | Describe the canonical bounded Service/Object/Function entry structure |
| Basic | Basic | Compose an App with a Main Service using a Process |

The exact source-defined [construction and callable reference](definitions.md) explains their inputs and limits. These names are source-defined facilities: aliases preserve canonical identity and descriptor targets govern composition.

## Compiler-typed operations

| Operation | Forms from configured root | Operands and behavior |
| --- | --- | --- |
| Build | Build, Builds, Built, Building, Builder | Checked App; validate prebuilt artifact at runtime |
| Run | Run, Runs, Ran, Running, Runner | Checked target; launch without join |
| Await | Await, Awaits, Awaited, Awaiting | Checked target; join finalized result and claim continuation once |
| Stop | Stop, Stops, Stopped, Stopping, Stopper | Checked target; request termination without join |
| Show | Show, Shows, Showed, Showing, Shower | Supported scalar/member observation; no added newline |
| Exit | Exit, Exits, Exited, Exiting, Exiter | Terminate whole program, with optional supported failure operand |

Inflection recognition does not make every grammatical owner valid. For example, native Await inside an App body rejects even though its word form is recognized. The [lifecycle reference](lifecycle.md) records supported contexts and failure modes.

## Declaration and vocabulary facilities

Create, Define, Describe, and Use establish declarations, typed/native definitions, descriptors, aliases, and supported imports. Public/Publicly controls visibility at one declaration site; New and Generic participate in construction. Noun/Verb/Adjective/Adverb are lexical categories, not native scalar types.

The configured vocabulary also contains Put, Place, Get, Take, Import, Export, Have, Equal, articles, prepositions, conjunctions, condition markers, and type/category nouns. Some are grammar-only or bounded roles: Have/Equal participate in conditions; collection actions and While do not have general executable outer lowering. Read the [grammar chapter](../../spec/grammar.md) rather than assuming every configured word is an implemented procedure.

## Runtime members

Output is Text containing finalized raw stdout. Error is the selected process operation/invocation failure. These are compiler-typed process members, so a possessive keeps its canonical target. Input exists in vocabulary/signature metadata but runtime process argument/input transport is not provided by the current App contract.

## Availability and errors

English builtins and their backend resources are embedded and loaded automatically. Compiler developers can [override the resource bundle](../../development-harness.md#compiler-defaults).

Wrong action parameter counts/types, unsupported target owners, unknown descriptor targets, inaccessible source declarations, and cycles are real errors. Builtin definitions do not hide lifecycle misuse, insert Await, trim output, or infer initialization. The [specification](../../spec/index.md) separates their source contract from backend-native facilities.

## Sources and witnesses

Sources: [core.l](../../../builtins/eng/core.l), [vocabulary](../../../lexicon/eng/vocabulary.json), [typed process operations](../../../lyger/builtin_operations.go), and [native instruction roles](../../../lyger/ir_kinds.go). Witnesses: [builtin operation contract](../../../lyger/builtin_operations_contract_test.go), [operation validation](../../../lyger/builtin_operations_validation_test.go), [action definitions](../../../lyger/action_definition_test.go), and [Service boundary coverage](../../../testdata/service-execution/coverage.md).
