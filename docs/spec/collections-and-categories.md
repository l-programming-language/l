# Collections and application categories

The source vocabulary names several kinds of value and source organization. This chapter separates their actual uses from internal storage structures and backend-native record types.

The examples below show the supported source forms and their boundaries. Existing test witnesses are identified beside the examples; the website build does not compile or execute them.

## Lists

List is a recognized source category backed internally by an ordered `Values` slice and a `Types` slice. The grammar contains list construction and collection-action forms. The internal list initializer creates empty storage.

### Declare a List

The declaration shape names a List:

<!-- example:list-declaration -->

`"items"` supplies the name and `as a List` supplies its category. This documentation fragment illustrates the grammar and recognized storage type; it has no compilation or runtime witness. The repository does not establish ordinary source list-element literals or a complete mutable-list interface, so there is no supported example of filling this List and iterating over its contents yet. An unused initialized compiler container and a container operated on at runtime are different claims.

The current typed outer instruction set does not implement a general List insertion, indexing, update, membership, traversal, or removal API. The configured Put, Place, Get, and Take families do not by themselves establish native execution. Consequently no element homogeneity rule, bounds exception, mutation aliasing promise, or loop semantics can be specified as an operational public contract from those structures alone.

List-related grammar entries remain in the complete grammar inventory with partial/unresolved classification. Do not translate them into conventional bracket literals or a for-loop syntax. Typed metadata arrays such as Props and Texts in native definitions are separately defined interfaces, not proof that source Lists lower into them automatically.

## Dictionaries

A collection Dictionary has internal key/value and key/type maps. A Dictionary also provides a demonstrated source namespace for named members. For example, the maintained public-values demo declares `Shared` as a Dictionary and defines `Shared's Display` as a source action. A possessive or qualified member reference retains the canonical owner.

A concrete namespace construction is `Create "Tools" as a Dictionary.` The retained example below includes a public Dictionary member alongside public scalar values. Its declared Display action is a source vocabulary member, not insertion of an arbitrary runtime map key.

<!-- example:public-values -->

This namespace use is supported and is the basis of indexed standard-library exports. General runtime map insertion, retrieval by arbitrary computed key, iteration order, membership, missing-key defaults, and removal are not established source contracts. Go map representation does not promise iteration order to L programs.

An imported vocabulary dictionary is a source unit selected by `stdlib/index.json`. It owns a private lexical root and exports its registered Dictionary and selected members. This meaning of dictionary is different from a runtime collection, although it uses a Dictionary declaration to define the public namespace. See [imports](imports.md#indexed-dictionaries).

## Apps

An App is an application value with a canonical identity and generated source. A `New Basic App` expands through builtin descriptors into a Main Service, Object, and Function structure. Actions such as `The App Prints message` become application-body source; outer lifecycle instructions control the resulting process.

### Create and execute an App

The basic demo creates an App, gives it a body, executes it, and displays its captured output:

<!-- example:basic -->

`Use` imports the public Print action. `Create` names the App `Hello`, and `The App Prints message` supplies the work it will perform. `Execute Hello` expands to Build, Run, then Await; `Show Hello's Output` displays the finalized stdout. The existing CLI witness expects `Hello World!` followed by one newline. Building this program with `l build` creates native artifacts; the lifecycle actions happen when you run the resulting executable.

Aliases preserve App identity rather than copying applications. Distinct declarations remain distinct even when their surface names match in separate private scopes. The compiler allocates safe native artifact names independently of the source IDs. An App is not an in-process closure or merely a Rust class instance.

## Services

A Service organizes source within App construction. Basic supplies the App's primary Main Service; additional constructed Services become library packages. Their nonprivate composed items form their native library interface.

Standalone Service declarations remain valid structural values but produce no packages. Standalone body effects such as `The Service Prints ...`, native lifecycle operations and implicit process Output/Error reject. Explicit static members and ordinary source-defined actions retain lexical lookup.

### Run an App and read its output

<!-- example:runnable-app -->

`The App Prints` supplies Worker's body. Run launches the App, Await joins its finalized output, and Show displays it. Expected stdout is `work completed` followed by one newline. Main/Basic are unchanged source-defined construction descriptors; a Service declaration does not gain a process entry role.

See the [Service composition contract](../cli-build.md#service-composition) for the distinction between L declarations and raw AML library/bin targets.

## Classes and Objects

Classes group modules and source files; Objects represent files. They are neither source paragraphs nor runtime activation scopes. Nested populated Classes use the maintained version-2 source topology with explicit immediate parents and direct dimensions. A Class does not acquire a process lifetime, and its metadata cannot be used as an implicit default for every child Object.

Application construction can compose an Object around a Function and a Service around that Object. Source ownership follows the bound construction graph and original source units. Native imports and module visibility belong to the owning file/module; moving them to a surrounding class would change ownership.

### Compose an Object inside nested Classes

The recursive-Class fixture is a complete construction example:

<!-- example:recursive-classes -->

Read the descriptors from the inside out. `Leaf` constructs a Class named `leaf` containing a Main Object. `Middle` puts that Class inside `middle`, and `Outer` puts it inside `outer`. `Grouped` puts the outer Class in a Service; `Tree` puts that Service in an App. The final `Create` applies the composition to `Recursive`.

The resulting ownership chain is `Recursive` → Service `main` → Class `outer` → Class `middle` → Class `leaf` → Object `main`. The Main Object owns the generated `main` function. The existing `TestFrontendRecursiveClassSource` witness checks that hierarchy, generated Rust module paths, and native compilation. It does not run this example, and the source contains no lifecycle action.

## Attributes, Definitions, and lexical categories

Attribute and Definition are recognized category names with internal applicative storage. Source-defined definitions also use lexical records and typed templates. The demonstrated source mechanisms are adjective descriptors and `Define`; recognition of the internal categories does not establish a general user-facing Attribute constructor or record-field API.

The source categories Noun, Verb, Adjective, and Adverb describe lexical roles. Adjective descriptors can constrain targets and supply constructors or native wrappers. They are not ordinary struct fields. Public is a visibility modifier; New selects construction; Generic participates in generic source composition. These words do not imply general type parameters or prototype inheritance.

### Describe with an adjective

The implicit Main descriptor is itself written in L:

<!-- example:main-descriptor -->

The first sentence declares Main as an Adjective for Service, Object, or Function. Each `It describes` refers to that adjective and adds a construction rule. The Function rule supplies its name and empty signature; the Object rule wraps that Function; the Service rule wraps the Object. Those rules explain the `New Main Service` used above. They are already loaded as builtins, so this excerpt is for reading, not a definition to paste again into your program.

When lowering the Object, the compiler stores its constructed Function as an internal Attribute containing a Definition. That representation does not introduce a separate L attribute-field declaration. See [source hierarchy construction](../../lyger/source_hierarchy.go) for this distinction.

### Define and use a word

This retained source fixture defines a native-backed operation used by App bodies:

<!-- example:native-dictionary-definition -->

`Log|-s|-ged|-ging|-ger` declares the verb's forms, including Logger. The `<@Output|Thing@>` slot binds the supplied value and the `<DEF>` framing contains the Rust implementation. ProbeIO.Print composes that definition as ordinary L source. The fixture supplies vocabulary for an App body; it does not independently execute.

For a definition made entirely from L actions, see [source operations](#source-operations) below. `Launch` is a verb, `Compact` an adjective descriptor, and `First` the App's noun/name in that example. The [Dictionary example](#dictionaries) also demonstrates the adverb `Publicly` on declarations. These lexical roles determine how words participate in sentences; they do not imply four interchangeable runtime value constructors.

## Functions, Properties, and Operations

Function is an implicit source-defined facility that composes FW from a Name, Inputs, Outputs, and Process. Function.Return is a native-backed member. This is distinct from the internal Function, Property, and Operation value tags and from the backend's Func/Prop/Op metadata records.

Nonempty source Function signatures need explicitly bound property values; a bare nonempty Inputs or Outputs relation is rejected. The documented Main composition uses No Inputs and No Outputs. General closures, dynamic dispatch, object methods, async functions, and automatic generic argument forwarding are not introduced by the Function spelling.

### Construct a Function in an Object

This excerpt from the interleaved-Class fixture defines a named Function and an Object that contains it:

<!-- example:function-object-descriptor -->

The Function is named `first`, takes No Inputs, and has No Outputs. `any Process` supplies its body when the descriptor is used. The Object rule applies `Objective First` to that Function. The [complete fixture](../../lyger/test/classes/interleaved.l) composes a `New First Object` into a Class and App; `TestFrontendInterleavedClassSource` checks its emitted source and native compilation. The three lines shown here are descriptor definitions, not a standalone function-call program.

### Read a member value

L's possessive member syntax is demonstrated by the finalized App output:

<!-- example:service-output-member -->

This excerpt continues the `Worker` declaration in the [App example](#run-an-app-and-read-its-output). `Worker's Output` reads the runtime member after Await; it does not create a user-defined Property record. General source construction of the internal Property category is not established. Native `Prop`/`Props` records describe fields or parameters at the logeme boundary and must not be confused with this member access.

### Source operations

The local-definitions demo composes a new action from existing lifecycle actions and calls it:

<!-- example:local-launch -->

`Create "Launch…" for` gives the action a typed `any App` argument. Its body builds that App, runs it, then awaits it; `it` refers to the argument. `Launch First` applies those steps to First. The following `Define "result"` binds the App's Output for `Show` to display. This is the first local paragraph of a [complete runnable demo](definitions.md#descriptors-and-constructors); the full demo's existing witness covers both local paragraphs and the final outer Show. The leading indentation is retained because it establishes the paragraph context.

This source action is the supported way illustrated here to name a sequence of operations. It does not construct the internal Operation value tag or the backend's `Op` metadata record. The [source-type whitelist witness](../../lyger/refactor_value_characterization_test.go) distinguishes those internal tags from supported source categories; a familiar name alone is not a public construction API.

Func, Prop, and Op nouns in the configured vocabulary refer to native metadata concepts. Detailed callable, property, trait, and record capabilities belong to the [Rust logeme reference](../../builtins/README.md). A supported backend template composition does not establish a matching top-level L construction form.

## Sources and witnesses

Primary sources: [storage types](../../lyger/types.go), [type recognition](../../lyger/utilities.go), [construction](../../lyger/construction.go), [builtin definitions](../../builtins/eng/core.l), [App workspace projection](../../lyger/workspace_units.go), and [source topology](../../lyger/source_topology.go). Witnesses: [public dictionary example](../../demo/paragraphs/public-values.l), [dictionary member aliases](../../lyger/member_resolution_regression_test.go), [recursive Classes](../../lyger/recursive_classes_test.go), [Service coverage](../../testdata/service-execution/coverage.md), and [stored value tests](../../lyger/stored_values_test.go). Collection mutation semantics remain an explicit support boundary.
