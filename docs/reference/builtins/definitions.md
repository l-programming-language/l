# Construction and definition builtins

These facilities are defined in the implicit source unit `builtins/eng/core.l`. Their source definitions are the authority for composition; their names do not create conventional object-oriented or functional-language semantics beyond the current implementation.

## Process

Process has forms Process, Processes, Processive, and Processing. Its native definition is operation W integrated with the ordered Definitions/Syllables parameter:

```text
<DEF>W^<@Definitions|Syllables@><DEF>
```

A Process collects existing instantiated definition expressions. It is not an OS process handle; App invocations own those. Processing a source-defined rendering or writing operation in IO is an application-body composition, not a direct runtime launch instruction.

Structured selection and iteration bodies can declare typed Process Inputs. Lyger binds each input within its body and retains its declared native type and source dependency; an explicit local annotation lets native compilation check type and borrow compatibility. `_` requests native type inference. These bindings create no callable or closure and cannot escape to a sibling branch or later sentence.

## Function

Function has forms Function, Functions, Functional, and Functioning. It supplies an FW native definition with Func metadata. Name is Text, Inputs is Props, Outputs is Texts, and Process is an already instantiated definition reference.

Inputs and Outputs preserve their typed cardinality. Main supplies No Inputs and No Outputs; ordinary Function declarations bind their checked signatures and paragraph bodies to the same selected definition. The constructor is not a marker for a second emitter.

Invocation words use normal grammatical binding, as Print does. Captures, snapshots and Pointer sharing follow [scope rules](../../spec/names-and-scope.md#runtime-storage-and-definite-assignment). Native signature/lifetime metadata remains bounded by the [definition contract](../../spec/definitions.md#process-function-and-return).

## Function.Return

Return, Returns, Returned, Returning, and Returner are Function members. A Function body resolves Return to its selected member definition and checks the declared result types and counts. A returning source action defined outside a Function uses its enclosing caller's Return; a Return defined within a Function retains that constructor contract and rejects an incompatible caller.

Return completes that invocation; Exit terminates the enclosing program. Neither transports runtime captures into an App process.

## Retain

Retain takes an explicit borrowed Arc and returns another handle to the same allocation. It increments shared ownership without cloning the payload. Atomic reference counting does not make an arbitrary payload Send or Sync.

The implementation dictionary's `Storage.App.Retain` accepts a borrowed App Arc and delegates to this core word. It performs no launch, stop, wait or invocation transition. `Storage.Retain` remains the separate Rc operation. `Storage.Release` consumes a value through native drop; any destructor behavior belongs to that value's representation.

## Storage source compositions

The [Storage implementation dictionary](../../../stdlib/eng/storage/dictionary.l) keeps algorithms in source Functions and Processes. Native leaves perform their individual calls, operators, borrows or patterns.

| Family | Source contract |
| --- | --- |
| Option.Select / Result.Select | Select an explicit branch Process, with its payload or error scoped to that branch |
| Traverse / Enumerate | Visit each element or pair once in iterator order; Paired describes the pair-shaped Iteration and Applier |
| Examine / Review / Inspect / Audit | Bind callback results explicitly and propagate the first failure through Verify |
| Decide / Probe / Search | Propagate a callback error, continue after false, and activate the supplied continuation after true |
| Require / Accept | Return a successful value without constructing a failure; Accept releases a replaced error before constructing its replacement |
| Index.Check | Accept a finite nonnegative integral Number strictly below `usize::MAX as f64` before native conversion; reject other values with the existing index diagnostic |
| Reject | Construct Failure, wrap it in Rejection, and return from the enclosing Function through its selected Return |
| Function.Yield / Function.Exit | Wrap the selected control-flow value with shared Success and invoke the selected Function.Return using explicit generated result bindings |
| Bytes.Extend / Text.Enclose | Append a borrowed byte slice to an exclusively borrowed vector; Enclose consumes its input vector and builds opening + value + closing in one output buffer |

Append and enclosure can copy bytes and reallocate. These words establish no general allocation, blocking or destructor-cost bound. Result.Expect remains a separate fail-fast operation with an explicit message.

The former Propagate, Function.Success, Function.Completion, Rejection.Error and Rejection.Result helpers are removed. Paired replaces the former Storage.Indexed descriptor, and explicit Routine generic parameter lists replace Storage.Fallible. No aliases preserve those retired identities. The runtime's duplicate Retain alias and Dispose leaf are also removed; their consumers use core Retain and Storage.Release.

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
