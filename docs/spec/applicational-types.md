# Applicational Types

| Type | What it represents | Current use in L |
| --- | --- | --- |
| [App](#apps) | An application with source, identity and process state. | Construct, build, run, stop and await a separate process; read its Output or Error. |
| [Service](#services) | A unit of code composed into an App. | Main supplies the entry structure; additional App-constructed Services become libraries. A standalone Service is structural. |
| [Class](#classes-and-objects) | A group of source modules, Objects and nested Classes. | Organize generated code; no independent process or runtime scope. |
| [Object](#classes-and-objects) | A source-file unit containing definitions. | Own generated file content; no automatic instance lifetime. |
| [Function](#functions-properties-and-operations) | A callable with typed inputs, results and retained captures. | Invoke its descriptive verb with fresh parameters and locals. |
| [Instance](#instances) | Runtime fields belonging to a canonical user-defined shape. | Own, snapshot, read and update typed fields. |
| [Attribute](#attributes-and-definitions) | A composed part of an Object. | Internal construction representation; no general source field-construction API. |
| [Definition](#attributes-and-definitions) | The implementation attached to a constructed element. | Internal representation used by source and logeme-backed definitions; no general standalone constructor. |
| [Property](#functions-properties-and-operations) | Field or parameter information. | Internal/native representation; reading a member does not construct a Property. |
| [Operation](#functions-properties-and-operations) | A structured native operation. | Internal/native representation; source actions are defined as words. |

Only Apps own process lifecycle. Internal categories do not imply source constructors. [Lexical Types](lexical-types.md) explains the words that express these structures.

## Apps

### Create and execute an App

<!-- example:basic -->

`New Basic App` constructs a Main Service, Object and entry Function. `The App Prints message` supplies its child-process body; `Execute Hello` builds, runs and awaits it. Show displays finalized stdout: `Hello World!` and one newline.

`l build` publishes the native program; running that program performs its lifecycle actions. Aliases preserve App identity. Separate declarations remain separate Apps even if they use the same name in different private scopes.

## Services

Basic supplies the primary Main Service. Additional Services constructed inside an App become native libraries whose nonprivate composed items form their interface. The [Service composition example](../guide/apps-and-services.md#service-composition) builds Main, Left and Right Services inside an App and explains how the body is passed into them.

A standalone Service declaration creates no package or process. Standalone body effects, native lifecycle operations and implicit Output/Error reject; explicit static members and source-defined actions still use normal lexical lookup. See the [Service composition contract](../cli-build.md#service-composition).

## Classes and Objects

Classes group modules and source files; Objects own files. Neither introduces a paragraph scope or process lifetime. Nested Classes retain explicit parents and source order. Native imports and visibility belong to their owning file/module; Class metadata does not silently become every child's default.

### Compose an Object inside nested Classes

<!-- example:recursive-classes -->

Read descriptors inside out: Leaf supplies the Main Object; Middle and Outer enclose it; Grouped adds the Service; Tree adds the App. Applying Tree produces `Recursive` → Service `main` → Class `outer` → Class `middle` → Class `leaf` → Object `main`, which owns the entry Function. Construction alone performs no lifecycle action.

## Attributes and Definitions

A Function composed inside an Object is represented as an Attribute containing a Definition. [Adjective descriptors](lexical-types.md#describe-with-an-adjective) and named definitions express this structure; these categories add no field syntax or interchangeable constructors.

## Functions, Properties, and Operations

A Function declaration also defines its invocation verb. Its encapsulated paragraph supplies the body:

```l
    Publicly create "Echo|-es|-ed|-ing|-er" as a New Function with Inputs of "Value" as Text and Outputs of Text.
Return Value.

Echo `hello` as "Result".
Show Result.
```

Four leading spaces begin the private body; a blank line ends it. Public exposes Echo, not Value. The label supplies ordinary verb inflections. A member word can form a sentence such as `The User walks to the Store`: User selects its Walk definition and Store supplies the Location argument. See [word forms](../guide/values.md#read-pipe-delimited-word-forms).

Each invocation checks ordered argument/result types and has fresh locals. Every continuing result-bearing path must Return. Return binds the Function's builtin member; Exit retains whole-program termination. Copies and aliases preserve signatures and captures, including callable parameters such as `"Callback" as Echo`.

The header retains surrounding conditions and coordinated effects. Captures must be assigned when it executes; ordinary values snapshot and explicit Pointers share targets. An invocation assigns its explicit results, without inferring assignment from side effects. A Function may Await an App and its continuation. Dictionary/Class units have no executable initialization phase.

Source declarations and Object construction use the same selected Function/logeme definitions. Rust checks the generated signatures and lifetimes; no user-written Rust lifetime parameters, overloads, variadic signatures or inference across opaque native bodies are implied.

### Construct a Function in an Object

<!-- example:function-object-descriptor -->

`first` has No Inputs and No Outputs; `any Process` receives its body. `Objective First` supplies the Object, which the [complete fixture](../../lyger/test/classes/interleaved.l) composes into a Class and App.

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
