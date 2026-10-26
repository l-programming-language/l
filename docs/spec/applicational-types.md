# Applicational Types

| Type | What it represents | Current use in L |
| --- | --- | --- |
| [App](#apps) | An application with source, identity and process state. | Construct, build, run, stop and await a separate process; read its Output or Error. |
| [Service](#services) | A unit of code composed into an App. | Main supplies the entry structure; additional App-constructed Services become libraries. A standalone Service is structural. |
| [Class](#classes-and-objects) | A group of source modules, Objects and nested Classes. | Organize generated code; no independent process or runtime scope. |
| [Object](#classes-and-objects) | A source-file unit containing definitions. | Own generated file content; no automatic instance lifetime. |
| [Function](#functions-properties-and-operations) | A callable with typed inputs, results and retained captures. | Call an L body with fresh parameters and locals. |
| [Instance](#instances) | Runtime fields belonging to a canonical user-defined shape. | Own, snapshot, read and update typed fields. |
| [Attribute](#attributes-and-definitions) | A composed part of an Object. | Internal construction representation; no general source field-construction API. |
| [Definition](#attributes-and-definitions) | The implementation attached to a constructed element. | Internal representation used by source and logeme-backed definitions; no general standalone constructor. |
| [Property](#functions-properties-and-operations) | Field or parameter information. | Internal/native representation; reading a member does not construct a Property. |
| [Operation](#functions-properties-and-operations) | A structured native operation. | Internal/native representation; source actions are defined as words. |

These types describe program structure and behavior. Only Apps own process lifecycle. The table includes internal categories so that recognizing a name does not imply a source constructor for it. [Lexical Types](lexical-types.md) explains how nouns, verbs and descriptors express these structures in sentences.

## Apps

### Create and execute an App

<!-- example:basic -->

`New Basic App` applies the builtin Basic descriptor. It constructs a Main Service containing an Object and entry Function. `The App Prints message` supplies the child process's body; `Execute Hello` builds, runs and awaits it. `Show Hello's Output` displays the finalized stdout: `Hello World!` followed by one newline.

`l build` publishes the native program; running that program performs its lifecycle actions. Aliases preserve App identity. Separate declarations remain separate Apps even if they use the same name in different private scopes.

## Services

Basic supplies the primary Main Service. Additional Services constructed inside an App become native libraries whose nonprivate composed items form their interface. The [Service composition example](../guide/apps-and-services.md#service-composition) builds Main, Left and Right Services inside an App and explains how the body is passed into them.

A standalone Service declaration creates no package or process. Standalone body effects, native lifecycle operations and implicit Output/Error reject; explicit static members and source-defined actions still use normal lexical lookup. See the [Service composition contract](../cli-build.md#service-composition).

## Classes and Objects

Classes group modules and source files; Objects own files. Neither introduces a paragraph scope or process lifetime. Nested Classes retain explicit parents and source order. Native imports and visibility belong to their owning file/module; Class metadata does not silently become every child's default.

### Compose an Object inside nested Classes

<!-- example:recursive-classes -->

Read the descriptors from the inside out: Leaf puts a Main Object in a Class; Middle and Outer add enclosing Classes; Grouped wraps them in a Service; Tree wraps the Service in an App. The final Create applies Tree to Recursive. Its ownership chain is `Recursive` → Service `main` → Class `outer` → Class `middle` → Class `leaf` → Object `main`. The Object owns the entry Function. This construction example contains no lifecycle action.

## Attributes and Definitions

When a descriptor constructs a Function inside an Object, the compiler represents it as an Attribute containing a Definition. Users normally express this through [adjective descriptors](lexical-types.md#describe-with-an-adjective) and named definitions. These categories do not add record-field syntax or interchangeable value constructors.

## Functions, Properties, and Operations

An ordinary Function owns the statements in its encapsulated paragraph:

```l
    Publicly create "Identity" as a New Function with Inputs of "Value" as Text and Outputs of Text.
Return Value.

Call Identity with `hello` as "Result".
Show Result.
```

The four leading spaces begin the Function's private body scope; the blank line ends it. Public exposes Identity, not Value. Call checks argument and result counts/types. Each invocation has fresh locals, and every continuing path in a result-bearing Function must Return. Exit retains its program-termination meaning. Existing Function names can specify a callable parameter/result signature, such as `"Callback" as Identity`.

Calls establish assignment only for their explicit result bindings. Public declarations initialize when their body runs. A Function may Await an App and run its continuation; free values read there must already be initialized at the Function declaration. The caller does not infer assignment from side effects.

Ordinary captures snapshot; explicit Pointers share targets. Callable copies retain their environment. There are no user-written Rust lifetime parameters, inferred contracts for opaque Rust, overloads or variadic signatures. The native Function/Function.Return composition remains separate:

### Construct a Function in an Object

<!-- example:function-object-descriptor -->

The Function is named `first`, with No Inputs and No Outputs. `any Process` receives its body when the descriptor is applied; `Objective First` wraps the Function in an Object. The [complete fixture](../../lyger/test/classes/interleaved.l) composes that Object into a Class and App.

`Worker's Output` selects an App member after Await; it does not create a Property record. Native `Func`, `Prop` and `Op` metadata have separate [logeme contracts](../../builtins/README.md). To name a sequence of source actions, [define a verb such as Launch](lexical-types.md#define-and-use-a-word); this does not construct an internal Operation.

## Instances

An adjective descriptor defines a runtime field shape, distinct from Class and Object:

```l
"Person" is a New Adjective describing any Instance.
It describes a Generic Instance with a `name` Text and a `age` Number.
Define "ada" as a New Person Instance.
Put `Ada` in ada as `name`.
Put 37 in ada as `age`.
Define "name" as a New Text from ada at `name`.
Show name.
```

The canonical Person declaration identifies the shape; unrelated descriptors with the same spelling remain distinct. Fields accept only their declared types. Reading an unassigned field or an unknown field rejects. Assignment snapshots the Instance's ordinary fields and retains nested Pointer targets.

## Sources and witnesses

See [application structures](../../lyger/application_types.go), [type recognition](../../lyger/utilities.go), [construction](../../lyger/construction.go), [builtins](../../builtins/eng/core.l), [workspace projection](../../lyger/workspace_units.go), [recursive-Class tests](../../lyger/recursive_classes_test.go), and [Service coverage](../../testdata/service-execution/coverage.md).
