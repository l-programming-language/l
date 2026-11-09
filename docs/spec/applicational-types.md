# Applicational Types

| Type | What it represents | Current use in L |
| --- | --- | --- |
| [App](../reference/word-glossary.md#word-type-app) | An application with source, identity and process state. | Construct, build, run, stop and await a separate process. Read its Output or Error. |
| [Service](../reference/word-glossary.md#word-type-service) | A unit of code composed into an App. | Main supplies the entry structure. Additional App-constructed Services become libraries. A standalone Service is structural. |
| [Class](#classes-and-objects) | A group of source modules, Objects and nested Classes. | Organize generated code without an independent process or runtime scope. |
| [Object](#classes-and-objects) | A source-file unit containing definitions. | Own generated file content without an automatic runtime lifetime. |
| [Function](#functions-properties-and-operations) | A callable with typed inputs, results and retained captures. | Invoke its descriptive verb with fresh parameters and locals. |
| [Attribute](#attributes-and-definitions) | A composed part of an Object. | Internal construction representation with no general source field-construction API. |
| [Definition](#attributes-and-definitions) | The implementation attached to a constructed element. | Internal representation used by source and logeme-backed definitions, with no general standalone constructor. |
| [Property](#functions-properties-and-operations) | Field or parameter information. | Internal/native representation. Reading a member does not construct a Property. |
| [Operation](#functions-properties-and-operations) | A structured native operation. | Internal/native representation. Source actions are defined as words. |

Only Apps own process lifecycle. Internal categories do not imply source constructors. [Lexical Types](lexical-types.md) explains the words that express these structures.

## Apps

[App in the Word Glossary](../reference/word-glossary.md#word-type-app).

### Create and execute an App

[Create and execute an App in the Word Glossary](../reference/word-glossary.md#word-type-app--create-and-execute-an-app).

## Services

[Service in the Word Glossary](../reference/word-glossary.md#word-type-service).

## Classes and Objects

### Class

[Class in the Word Glossary](../reference/word-glossary.md#word-type-class).

### Object

[Object in the Word Glossary](../reference/word-glossary.md#word-type-object).

### Compose an Object inside nested Classes

<!-- example:recursive-classes -->

Read descriptors inside out: Leaf supplies the Main Object, Middle and Outer enclose it, Grouped adds the Service, and Tree adds the App. Applying Tree produces App `Recursive` containing Service `main` with nested Classes `outer`, `middle` and `leaf`, in that order. The innermost Class contains Object `main`, which owns the entry Function. Construction alone performs no lifecycle action.

## Attributes and Definitions

### Attribute

[Attribute in the Word Glossary](../reference/word-glossary.md#word-type-attribute).

### Definition

[Definition in the Word Glossary](../reference/word-glossary.md#word-type-definition).

## Functions, Properties, and Operations

[Function](../reference/word-glossary.md#word-builtin-function) is the canonical word entry for the selected callable constructor. This section explains its value and capture behavior. Property/Operation and the configured Func/Prop/Op spellings describe native representation. Their presence does not add public application constructors.

A Function declaration also defines its invocation verb. Its encapsulated paragraph supplies the body:

<!-- example:composed-function-definition -->

Four leading spaces begin the private body, and a blank line ends it. Public exposes Identity, not Value. The [complete Function example](../guide/definitions.md#give-a-result-bearing-operation-a-function) calls it through Relay and names the caller's result explicitly. A label may declare related [word forms](lexical-structure.md#vocabulary-and-inflection).

Invocations check ordered types and have fresh locals. Function signatures include canonical constructor identity. Incompatible constructors reject, including boxed values. Copies and aliases retain signatures and captures. Every continuing result-bearing path must Return through its bound member. Exit terminates the program.

The header retains surrounding conditions and coordinated effects. Captures must be assigned when it executes. Ordinary values snapshot and explicit Pointers share targets. An invocation assigns its explicit results, without inferring assignment from side effects. A Function may Await an App and its continuation. Dictionary/Class units have no executable initialization phase.

Source declarations and Object construction use the same selected Function/logeme definitions. Ordinary signatures do not infer native lifetime or generic parameters. Native-backed constructors can accept explicit atoms within the [definition contract](definitions.md#composition-and-source-ownership). This introduces no overloads, variadic Function signatures or inference across opaque native bodies.

### Construct a Function in an Object

<!-- example:function-object-descriptor -->

`first` has No Inputs and No Outputs, and `any Process` receives its body. `Objective First` supplies the Object, which the [complete fixture](../../lyger/test/classes/interleaved.l) composes into a Class and App.

`Worker's Output` selects an App member after Await. It does not create a Property record. Native `Func`, `Prop` and `Op` metadata have separate [logeme contracts](../../builtins/README.md). To name a sequence of source actions, [define a verb such as Launch](lexical-types.md#define-and-use-a-word). This does not construct an internal Operation.

## Sources and witnesses

See [application structures](../../lyger/application_types.go), [type recognition](../../lyger/utilities.go), [construction](../../lyger/construction.go), [builtins](../../builtins/eng/core.l), [workspace projection](../../lyger/workspace_units.go), [recursive-Class tests](../../lyger/recursive_classes_test.go), and [Service coverage](../../testdata/service-execution/coverage.md).
