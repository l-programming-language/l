# Applicational Types

| Type | What it represents | Current use in L |
| --- | --- | --- |
| [App](#apps) | An application with source, identity and process state. | Construct, build, run, stop and await a separate process; read its Output or Error. |
| [Service](#services) | A unit of code composed into an App. | Main supplies the entry structure; additional App-constructed Services become libraries. A standalone Service is structural. |
| [Class](#classes-and-objects) | A group of source modules, Objects and nested Classes. | Organize generated code; no independent process or runtime scope. |
| [Object](#classes-and-objects) | A source-file unit containing definitions. | Own generated file content; no automatic runtime lifetime. |
| [Function](#functions-properties-and-operations) | A callable with typed inputs, results and retained captures. | Invoke its descriptive verb with fresh parameters and locals. |
| [Attribute](#attributes-and-definitions) | A composed part of an Object. | Internal construction representation; no general source field-construction API. |
| [Definition](#attributes-and-definitions) | The implementation attached to a constructed element. | Internal representation used by source and logeme-backed definitions; no general standalone constructor. |
| [Property](#functions-properties-and-operations) | Field or parameter information. | Internal/native representation; reading a member does not construct a Property. |
| [Operation](#functions-properties-and-operations) | A structured native operation. | Internal/native representation; source actions are defined as words. |

Only Apps own process lifecycle. Internal categories do not imply source constructors. [Lexical Types](lexical-types.md) explains the words that express these structures.

## Apps

<!-- word:type:app -->

Own an executable source composition and its separately supervised native process invocations.

**Usage and example.** `Create "Hello" as a New Basic App` in the [first program](../guide/first-program.md#read-the-program) declares the App; Execute then runs its body. An App alias shares its invocation identity.

### Create and execute an App

The [first program](../guide/first-program.md#read-the-program) constructs a Basic App, supplies its Print body and observes its captured output. Basic composes the Main Service, Object and entry Function; Execute builds, runs and awaits the App.

`l build` publishes the native program; running that program performs its lifecycle actions. Aliases preserve App identity. Separate declarations remain separate Apps even if they use the same name in different private scopes.

## Services

<!-- word:type:service -->

Group source code within an App, supplying its primary entry structure or an additional constructed library.

**Usage and example.** The [Service composition fixture](../../lyger/test/workspace/services.l) constructs Main, Left and Right Services within one App. A standalone Service remains structural and cannot be Run or Awaited.

Basic supplies the primary Main Service. Additional Services constructed inside an App become native libraries whose nonprivate composed items form their interface. The [Service composition example](../guide/apps-and-services.md#service-composition) builds Main, Left and Right Services inside an App and explains how the body is passed into them.

A standalone Service declaration creates no package or process. Standalone body effects, native lifecycle operations and implicit Output/Error reject; explicit static members and source-defined actions still use normal lexical lookup. See the [Service composition contract](../cli-build.md#service-composition).

## Classes and Objects

### Class

<!-- word:type:class -->

Group Objects and nested Classes into native source modules.

**Usage and example.** Outer, Middle and Leaf in the [recursive construction example](#compose-an-object-inside-nested-classes) retain their explicit parent order. Classes organize sources; they add no runtime scope or process lifetime.

### Object

<!-- word:type:object -->

Own a native source file containing composed definitions and Functions.

**Usage and example.** The Main Object in the [recursive construction example](#compose-an-object-inside-nested-classes) owns the entry Function. Object construction supplies source content, not an object initializer or process launch.

Native imports and visibility belong to their owning file/module; Class metadata does not silently become every child's default.

### Compose an Object inside nested Classes

<!-- example:recursive-classes -->

Read descriptors inside out: Leaf supplies the Main Object; Middle and Outer enclose it; Grouped adds the Service; Tree adds the App. Applying Tree produces `Recursive` → Service `main` → Class `outer` → Class `middle` → Class `leaf` → Object `main`, which owns the entry Function. Construction alone performs no lifecycle action.

## Attributes and Definitions

### Attribute

<!-- word:type:attribute -->

Name a composed part of an Object in its applicational representation.

**Context and illustration.** A Function inside the [Object fixture](../../lyger/test/classes/interleaved.l) becomes an Attribute containing its Definition. Use the supported Function/Object composition; this category does not establish a general `New Attribute` field-construction API.

### Definition

<!-- word:type:definition -->

Name the implementation attached to a constructed element or lexical word.

**Context and illustration.** The [native definition example](definitions.md#framing-and-escaping) associates framed source with a word. Definition describes that representation; it is not an interchangeable standalone runtime constructor. [Descriptors](lexical-types.md#describe-with-an-adjective) compose the actual selected definitions.

## Functions, Properties, and Operations

[Function](../reference/builtins/definitions.md#function) is the canonical word entry for the selected callable constructor. This section explains its value and capture behavior. Property/Operation and the configured Func/Prop/Op spellings describe native representation; their presence does not add public application constructors.

A Function declaration also defines its invocation verb. Its encapsulated paragraph supplies the body:

<!-- example:composed-function-definition -->

Four leading spaces begin the private body; a blank line ends it. Public exposes Identity, not Value. The [complete Function example](../guide/definitions.md#give-a-result-bearing-operation-a-function) calls it through Relay and names the caller's result explicitly. A label may declare related [word forms](lexical-structure.md#vocabulary-and-inflection).

Invocations check ordered types and have fresh locals. Function signatures include canonical constructor identity; incompatible constructors reject, including boxed values. Copies and aliases retain signatures and captures. Every continuing result-bearing path must Return through its bound member; Exit terminates the program.

The header retains surrounding conditions and coordinated effects. Captures must be assigned when it executes; ordinary values snapshot and explicit Pointers share targets. An invocation assigns its explicit results, without inferring assignment from side effects. A Function may Await an App and its continuation. Dictionary/Class units have no executable initialization phase.

Source declarations and Object construction use the same selected Function/logeme definitions. Ordinary signatures do not infer native lifetime or generic parameters. Native-backed constructors can accept explicit atoms within the [definition contract](definitions.md#composition-and-source-ownership); this introduces no overloads, variadic Function signatures or inference across opaque native bodies.

### Construct a Function in an Object

<!-- example:function-object-descriptor -->

`first` has No Inputs and No Outputs; `any Process` receives its body. `Objective First` supplies the Object, which the [complete fixture](../../lyger/test/classes/interleaved.l) composes into a Class and App.

`Worker's Output` selects an App member after Await; it does not create a Property record. Native `Func`, `Prop` and `Op` metadata have separate [logeme contracts](../../builtins/README.md). To name a sequence of source actions, [define a verb such as Launch](lexical-types.md#define-and-use-a-word); this does not construct an internal Operation.

## Sources and witnesses

See [application structures](../../lyger/application_types.go), [type recognition](../../lyger/utilities.go), [construction](../../lyger/construction.go), [builtins](../../builtins/eng/core.l), [workspace projection](../../lyger/workspace_units.go), [recursive-Class tests](../../lyger/recursive_classes_test.go), and [Service coverage](../../testdata/service-execution/coverage.md).
