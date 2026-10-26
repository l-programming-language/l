# Construction and definition builtins

These facilities are defined in the implicit source unit `builtins/eng/core.l`. Their source definitions are the authority for composition; their names do not create conventional object-oriented or functional-language semantics beyond the current implementation.

## Process

Process has forms Process, Processes, Processive, and Processing. Its native definition is operation W integrated with the ordered Definitions/Syllables parameter:

```text
<DEF>W^<@Definitions|Syllables@><DEF>
```

A Process collects existing instantiated definition expressions. It is not an OS process handle; App invocations own those. Processing a source-defined Logger in IO is an application-body composition, not a direct runtime launch instruction.

## Function

Function has forms Function, Functions, Functional, and Functioning. It supplies an FW native definition with Func metadata. Name is Text, Inputs is Props, Outputs is Texts, and Process is an already instantiated definition reference.

Inputs and Outputs preserve their typed cardinality. Main supplies No Inputs and No Outputs; ordinary Function declarations bind their checked signatures and paragraph bodies to the same selected definition. The constructor is not a marker for a second emitter.

Invocation words use normal grammatical binding, as Print does. Captures, snapshots and Pointer sharing follow [scope rules](../../spec/names-and-scope.md#runtime-storage-and-definite-assignment). Native signature/lifetime metadata remains bounded by the [definition contract](../../spec/definitions.md#process-function-and-return).

## Function.Return

Return, Returns, Returned, Returning, and Returner are Function members. A Function body resolves Return to its selected member definition and checks the declared result types and counts. Native composition determines the return context.

Return completes that invocation; Exit terminates the enclosing program. Neither transports runtime captures into an App process.

## Objective

Objective describes any Function by wrapping its existing definition in a grouped native source word containing A and the Function. It preserves the input definition and grouping; it does not clone its runtime semantics or infer additional function arguments.

The source-native grouping is explicitly represented as JSON arrays within the framed definition. Treat it as escaped code in documentation, not HTML/AML to be reparsed by the page renderer.

## Main

Main is a New Adjective describing Service, Object, or Function. Its ordered refinements compose:

| Target | Builtin construction |
| --- | --- |
| Function | Generic `main` Function of a Process, with No Inputs and No Outputs |
| Object | Generic `main` Object with a New Objective Main Function using a Process |
| Service | Generic `main` Service with a New Main Object using a Process |

These are ordinary source-defined descriptor rules. Aliases preserve their identity, and each construction selects the rule matching its target type.

Main does not grant standalone Service execution. A standalone declaration is structural; standalone Service body effects and native process operations reject. Apps own executable identity through their construction.

## Basic

Basic describes a Generic App with a New Main Service using a Process. It is the composition behind `Create "Hello" as a New Basic App.` The App's source actions supply its body; the orchestrator separately controls launch and output.

Basic is implicit; copying its declaration into a project bootstrap is unnecessary and can alter canonical identity or collide with existing definitions. Custom descriptors can compose equivalent supported structure while retaining their own declaration identity; they must obey constructor target and parameter rules.

## Failures and visibility

Descriptor target mismatch, a missing constructor rule, recursive descriptor expansion, malformed slots, unsupported signatures, and inaccessible captures reject at their respective semantic boundaries. Aliases preserve canonical descriptor identity. Public/Publicly controls publication independently of New/Generic construction.

Static references retain definition-site bindings. Runtime captures retain their declaration activation: ordinary values snapshot and explicit Pointer targets share. Caller shadowing does not select another environment.

## Sources and witnesses

Sources: [core source](../../../builtins/eng/core.l), [descriptors](../../../lyger/semantics.go), [construction](../../../lyger/construction.go), and [native definition interface](../../../lyger/definition.go). Witnesses: [construction](../../../lyger/construction_test.go), [definitions](../../../lyger/definition_test.go), [descriptor identity](../../../lyger/paragraph_descriptor_identity_test.go), [capture](../../../lyger/paragraph_capture_architecture_test.go), and [Service boundary tests](../../../lyger/runnable_service_test.go).