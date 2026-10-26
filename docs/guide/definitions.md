# Define vocabulary and behavior

L definitions introduce words, grammatical forms, typed arguments and ordered actions. Descriptors supply construction rules. Both participate in ordinary compilation rather than unrestricted native text substitution.

## Compose a source action

<!-- example:local-definitions -->

`Create ... for ...` defines Launch. `any App` declares its App argument; later uses of it refer to that same argument. Calling `Launch First` builds, runs and awaits First. Coordinated operands each receive the complete action body in source order.

`"Launch|-es|-ed|-ing"` declares Launch, Launches, Launched and Launching in the verb positions explained under [word forms](values.md#read-pipe-delimited-word-forms). Each paragraph owns its Launch action and Compact descriptor, so sibling definitions remain independent. The [paragraph test table](../../cmd/l/paragraph_integration_test.go) expects `First local\nSecond local\ndone`.

## Describe construction

`"Compact" describes a Generic App with a New Main Service using any Process` defines Compact as an adjective descriptor for Apps. The Generic App is the App shape described by the rule; `New Main Service` supplies its entry structure, and `using any Process` passes the supplied body of actions into that structure. `New Compact App` applies that pattern to the new App, much like Basic.

Descriptor targets, names, typed arguments, source ownership and order remain significant: a rule for an App cannot be substituted for a Service rule. Generic here does not declare type parameters.

## Know the implicit foundation

The [canonical core](../../builtins/eng/core.l) defines Start, Execute, Process, Function, Return, Objective, Main and Basic. Start and Execute are source actions for Apps, not primitive runtime opcodes or Service actions.

Process, Function, Objective and Main compose the entry used by Basic. [Typed Function declarations](../spec/applicational-types.md#functions-properties-and-operations) use the same builtin/logeme path and define ordinary invocation verbs. Function-local Return uses the selected builtin member; opaque native leaves retain their own type contracts.

## Diagnose a definition failure

Check unresolved arguments against their declared types rather than replacing them with String. Expansion cycles reject. Malformed native-body framing and unsupported Rust template compositions are different errors; retain their source span and original diagnostic.

Continue with [control flow](control-flow.md), [imports](imports.md), and [lifecycle](lifecycle.md) before the advanced [native-definition guide](native-definitions.md).
