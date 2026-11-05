# Define vocabulary and behavior

Give each word a coherent purpose and explicit inputs, results and effects. Decompose behavior where a step has a useful independent meaning, not to meet a line limit or add forwarding layers.

## Compose a source action

<!-- example:local-definitions -->

`Create ... for ...` defines Launch. `any App` declares its argument; `it` refers to that argument. `Launch First` builds, runs and awaits First. Coordinated operands each receive the complete body in order. Source actions retain their [definition-site bindings](../spec/names-and-scope.md#runtime-storage-and-definite-assignment).

`"Launch|-es|-ed|-ing"` declares Launch, Launches, Launched and Launching; see [word forms](values.md#read-pipe-delimited-word-forms). Each paragraph owns its Launch action and Compact descriptor, so sibling definitions remain independent.

## Give a result-bearing operation a Function

<!-- example:composed-functions -->

Identity returns its Text input. Relay names that call's result Nested and returns it. The outer call names Result before Show reads it. The [first-program commands](first-program.md#build-and-run) build and run this complete example; it writes exactly `hello`, without a newline. Relay makes the dataflow visible here; a forwarding word is useful in real code when it adds a meaningful boundary.

Each Function's four-space paragraph owns its inputs and body; the blank line ends it. Publicly exposes the Function, not its local names. A Function invocation has fresh parameters and locals. A source action such as Launch expands its ordered actions; a Process supplies a body to another composition; an App owns a separate OS process. Choose the form that matches the work.

## Organize words by meaning and grammatical role

Use single-word lexical roots and Dictionary/member ownership to group related meanings. IO owns Send and Lined; imported aliases retain that ownership. Keep helpers private and expose the operations callers need. [Project dictionary imports](imports.md#import-a-project-owned-member) show this organization.

A verb describes an action; its actor noun constructs that action's definition. IO's `"Send|-s|--t|-ing|-er"` declares Send, Sends, Sent, Sending and Sender as forms of one identity. The adjective Lined describes how to construct a Sender. Follow the [inflection rules](../spec/lexical-structure.md#vocabulary-and-inflection) rather than guessing spellings.

## Keep branch data inside its Process

The implementation dictionary's Require illustrates composition in a Function with native types:

<!-- example:source-require -->

Option.Select receives Value and two Process bodies. The present branch binds Payload as `T`; only the absent branch constructs Failure from Message. Returning completes Require. This fragment needs its dictionary declarations and native caller; it is not a standalone program.

Name intermediate results the next operation needs. Declare each Process input's type and ownership; `_` requests native inference. Inputs and locals stay in that body, with distinct bindings on repeated or nested use. See [Process](../reference/builtins/definitions.md#process) for scope/cardinality and [Return](../reference/builtins/definitions.md#functionreturn) for caller selection and lexical delegation.

## Describe construction

In the Launch example, Compact describes a Generic App with a Main Service. `using any Process` passes the supplied body into that entry structure. `New Compact App` applies the pattern, much like Basic.

Descriptor targets are canonical identities: an App rule cannot describe a Service, and a same-named private declaration is a different target. Generic here selects a construction pattern; it does not declare type parameters.

A wrapper composes around its retained base; a replacement supplies a new construction. Checked value operands evaluate once in source order, while Process bodies stay within their branch or iteration. Reusing a value does not clone it or erase its borrow contract. The [native guide's Lined example](native-definitions.md#compose-a-wrapper) shows a source wrapper that uses the base result and original writer.

## Know the implicit foundation

The [implicit foundation](../reference/builtins/index.md#source-defined-facilities) includes Start/Execute and the Process/Function/Objective/Main composition behind Basic. Extend existing words before adding a native leaf; the [native guide](native-definitions.md) explains that boundary.

## Diagnose a definition failure

Check the selected word, its argument count and types, and the original source location. Return belongs to a Function; Exit terminates the program. Every continuing result-bearing Function path must return. Expansion cycles, malformed native framing and unsupported target compositions are distinct failures.

Continue with [control flow](control-flow.md), [imports](imports.md), and [lifecycle](lifecycle.md) before the advanced [native-definition guide](native-definitions.md).
