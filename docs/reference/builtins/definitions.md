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

Inputs and Outputs preserve their typed cardinality. Main supplies explicit No Inputs and No Outputs. Nonempty source signatures need explicit property values; the construction pass rejects an unbound nonempty signature. The native definition facility supports concrete typed metadata, not automatic inference of Rust types from arbitrary source text.

Native named functions and callable fragments have separately bounded Rust target support. The native Function template does not infer receiver methods or native generic arguments; ordinary L Function captures and typed calls are checked by the frontend. See the [specification's definition boundary](../../spec/definitions.md#process-function-and-return).

## Function.Return

Return, Returns, Returned, Returning, and Returner are members of any Function in the builtin definition. The native template composes BOD/E with `return` and a typed Text Statement slot. The slot is encoded at the definition boundary; it is not unrestricted source injection.

The function composition determines its valid native return context and declared outputs. No generic cross-process return transport follows from this word. Use Exit for whole-program termination only when that behavior is intended; Exit is not an alias of Return.

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

A source-defined definition can refer to static values in its original environment without adopting a caller's shadowed value. A public runtime-dependent definition cannot escape private activation storage. These rules prevent superficially similar names from silently changing the generated program.

## Sources and witnesses

Sources: [core source](../../../builtins/eng/core.l), [descriptors](../../../lyger/semantics.go), [construction](../../../lyger/construction.go), and [native definition interface](../../../lyger/definition.go). Witnesses: [construction](../../../lyger/construction_test.go), [definitions](../../../lyger/definition_test.go), [descriptor identity](../../../lyger/paragraph_descriptor_identity_test.go), [capture](../../../lyger/paragraph_capture_architecture_test.go), and [Service boundary tests](../../../lyger/runnable_service_test.go).