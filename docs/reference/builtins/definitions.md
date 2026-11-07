# Construction and definition builtins

The implicit [core source](../../../builtins/eng/core.l) defines these composition facilities. See the [builtin index](index.md#source-defined-facilities) for their word forms and the [definition guide](../../guide/definitions.md) for source authoring.

## Process

Process composes ordered Definitions/Syllables through operation W.

A Process collects existing instantiated definition expressions. It is not an OS process handle; App invocations own those. Processing a source-defined rendering or writing operation in IO is an application-body composition, not a direct runtime launch instruction.

Structured selection and iteration bodies can declare typed Process Inputs. Lyger binds each input within its body and retains its declared native type and source dependency; an explicit local annotation lets native compilation check type and borrow compatibility. `_` requests native type inference. These bindings create no callable or closure and cannot escape to a sibling branch or later sentence.

## Function

Function supplies an FW native definition with Func metadata. Name is Text, Inputs is Props, Outputs is Texts, and Process is an already instantiated definition reference.

Inputs and Outputs preserve their typed cardinality. Main supplies No Inputs and No Outputs; ordinary Function declarations bind their checked signatures and paragraph bodies to the same selected definition. The constructor is not a marker for a second emitter.

Invocation words use normal grammatical binding. Captures, snapshots and Pointer sharing follow [scope rules](../../spec/names-and-scope.md#runtime-storage-and-definite-assignment). Native signature/lifetime metadata remains bounded by the [definition contract](../../spec/definitions.md#process-function-and-return).

## Function.Invoke and Delegate

Function.Invoke applies its Target with ordered Arguments atoms through KY. Delegate supplies a direct native Target call without an argument-list slot. These are composition facilities; each source Function declaration binds its own ordinary invocation verb, with checked inputs and outputs.

## Function.Return

Root-authored returning actions select their caller's Return; that selected implementation preserves its own lexical delegates. Function-owned actions retain their constructor contract and reject incompatible callers. Each Function establishes its own return context and checks declared result types and counts. Known unconditional exits through source actors and transparent Processes enter the typed plan; opaque or conditional bodies do not establish unconditional exits.

Return completes that invocation; Exit terminates the enclosing program. Neither transports runtime captures into an App process.

## Retain

Retain takes an explicit borrowed Arc and returns another handle to the same allocation. It increments shared ownership without cloning the payload. Atomic reference counting does not make an arbitrary payload Send or Sync.

The [Storage ownership words](../../internals/storage.md#ownership-and-bytes) distinguish this Arc operation from Rc retention, copying and release.

## Storage source compositions

Prelude includes core and imports Storage privately. Its [implementation reference](../../internals/storage.md) explains reusable result selection, traversal, return, ownership and byte operations. Programs that name Storage members import that dictionary explicitly.

## Objective

Objective describes any Function by wrapping its existing definition in a grouped native source word containing A and the Function. It preserves the input definition and grouping; it does not clone its runtime semantics or infer additional function arguments.

## Main

Main is a New Adjective describing Service, Object, or Function. Its ordered refinements compose:

| Target | Builtin construction |
| --- | --- |
| Function | Generic `main` Function of a Process, with No Inputs and No Outputs |
| Object | Generic `main` Object with a New Objective Main Function using a Process |
| Service | Generic `main` Service with a New Main Object using a Process |

These are ordinary source-defined descriptor rules. Aliases preserve their identity, and each construction applies every rule matching its target in authored order. Main currently has one rule for each target above; repeated construction remains distinct.

Main does not grant standalone Service execution. A standalone declaration is structural; standalone Service body effects and native process operations reject. Apps own executable identity through their construction.

## Basic

Basic describes a Generic App with a New Main Service using a Process. It is the composition behind `Create "Hello" as a New Basic App.` The App's source actions supply its body; the orchestrator separately controls launch and output.

## Failures and visibility

Descriptor target mismatch, a missing constructor rule, recursive descriptor expansion, malformed slots, unsupported signatures, and inaccessible captures reject at their respective semantic boundaries. Aliases preserve canonical descriptor identity. Public/Publicly controls publication independently of New/Generic construction.

Static references retain definition-site bindings. Runtime captures retain their declaration activation: ordinary values snapshot and explicit Pointer targets share. Caller shadowing does not select another environment.

## Sources and witnesses

Sources: [core source](../../../builtins/eng/core.l), [descriptors](../../../lyger/semantics.go), [construction](../../../lyger/construction.go), and [native definition interface](../../../lyger/definition_bindings.go). Witnesses: [word contracts](../../../lyger/word_library_acceptance_test.go), [ownership](../../../lyger/word_library_ownership_test.go), [construction](../../../lyger/construction_test.go), [definitions](../../../lyger/definition_test.go), [descriptor identity](../../../lyger/paragraph_descriptor_identity_test.go), [capture](../../../lyger/paragraph_capture_architecture_test.go), and [Service boundary tests](../../../lyger/runnable_service_test.go).
