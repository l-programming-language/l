---
tableOfContents:
  maxHeadingLevel: 2
---

# Word Glossary

Find each word by its qualified meaning. Entries give its part of speech, every
declared form and case, exact source definition, human description and usage.
[Native Words](words.md) indexes implicit vocabulary; the
[library indexes](../stdlib/index.md) cover explicitly imported dictionaries.

The [support legend](words.md#how-to-read-the-labels) separates implementation
support from evidence. **Authoring** means the word requires its documented
native or compiler context. A qualified implementation name can be addressable
without being an application API. Source definitions and usage examples expand
independently. Their disclosures distinguish fragments, compilation and execution.

<nav aria-label="Glossary alphabet" data-pagefind-ignore>

[A](#glossary-a) · [B](#glossary-b) · [C](#glossary-c) · [D](#glossary-d) · [E](#glossary-e) · [F](#glossary-f) · [G](#glossary-g) · [H](#glossary-h) · [I](#glossary-i) · [J](#glossary-j) · [K](#glossary-k) · [L](#glossary-l) · [M](#glossary-m) · [N](#glossary-n) · [O](#glossary-o) · [P](#glossary-p) · [Q](#glossary-q) · [R](#glossary-r) · [S](#glossary-s) · [T](#glossary-t) · [U](#glossary-u) · [V](#glossary-v) · [W](#glossary-w) · [Y](#glossary-y)

</nav>

## A

<a id="glossary-a"></a>

### A

<a id="word-grammar-word-a"></a>

<!-- word:grammar-word:a -->

Introduce an indefinite declaration or construction using the declared alternative form An where appropriate.

<!-- word-definition:grammar-word:a -->

**Usage and boundary.** `a New Basic App` in the [first program](../guide/first-program.md#read-the-program) and `an Error` in the [value program](../examples/word-values.l) use the same article family. A/An does not select a type independently of its noun.

### Runtime.Abort

<a id="word-authoring-runtime-abort"></a>

<!-- word:authoring:Runtime.Abort -->

Request cancellation of the selected native task without joining it. The selected operation is `Task.abort()`.

<!-- word-definition:authoring:Runtime.Abort -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

This statement is used inside `Runtime.Quell`. Its enclosing native signature and preceding bindings remain required.

<!-- example:glossary-runtime-abort-caller -->

</details>

### Storage.Absent

<a id="word-authoring-storage-absent"></a>

<!-- word:authoring:Storage.Absent -->

Construct the native None variant of an Option without a payload.

<!-- word-definition:authoring:Storage.Absent -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

This statement is used inside `Runtime.Capture.Adapt`. Its enclosing native signature and preceding bindings remain required.

<!-- example:glossary-storage-absent-caller -->

</details>

### Storage.Accept

<a id="word-authoring-storage-accept"></a>

<!-- word:authoring:Storage.Accept -->

Consume a Result, preserve its successful payload, or release its old error before constructing a runtime failure from the supplied message. The message argument is evaluated normally even on success.

<!-- word-definition:authoring:Storage.Accept -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

This statement is used inside [Storage.Slot.Read](#word-authoring-storage-slot-read). Its enclosing native signature, preceding local bindings and selected native module are required.

<!-- example:glossary-builtins-eng-storage-program-28-28 -->

</details>

### Storage.Access

<a id="word-authoring-storage-access"></a>

<!-- word:authoring:Storage.Access -->

Borrow the payload of an Option and return that reference as a runtime success, or construct a runtime failure from the supplied message when the Option is empty.

<!-- word-definition:authoring:Storage.Access -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

This statement is used inside `Storage.Environment.Require`. Its enclosing native signature and preceding bindings remain required.

<!-- example:glossary-storage-access-caller -->

</details>

### Runtime.Error.Accumulate

<a id="word-authoring-runtime-error-accumulate"></a>

<!-- word:authoring:Runtime.Error.Accumulate -->

Unpack a diagnostic traversal context and append one Error using its shared output buffer, visited identities, and separator state.

<!-- word-definition:authoring:Runtime.Error.Accumulate -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

This statement is used inside `Runtime.Error.Collect`. Its enclosing native signature and preceding bindings remain required.

<!-- example:glossary-runtime-error-accumulate-caller -->

</details>

### Runtime.Acquire

<a id="word-authoring-runtime-acquire"></a>

<!-- word:authoring:Runtime.Acquire -->

Obtain the lock future for the selected asynchronous mutex; awaiting it is a separate composition step. The selected operation is `Mutex.lock()`.

<!-- word-definition:authoring:Runtime.Acquire -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

This statement is used inside `Runtime.Finish`. Its enclosing native signature and preceding bindings remain required.

<!-- example:glossary-runtime-acquire-caller -->

</details>

### Runtime.App.Acquire

<a id="word-authoring-runtime-app-acquire"></a>

<!-- word:authoring:Runtime.App.Acquire -->

Reject an outstanding App invocation, then allocate and select a new invocation, mark it outstanding, clear the previous operation error, and return its handle and request receiver.

<!-- word-definition:authoring:Runtime.App.Acquire -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

This statement is used inside `Runtime.App.Run`. Its enclosing native signature and preceding bindings remain required.

<!-- example:glossary-runtime-app-acquire-caller -->

</details>

### Storage.Program.Actions

<a id="word-authoring-storage-program-actions"></a>

<!-- word:authoring:Storage.Program.Actions -->

Borrow `Program.actions` without moving the referenced native field.

<!-- word-definition:authoring:Storage.Program.Actions -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

This statement is used inside [Storage.Program.Actions.Read](#word-authoring-storage-program-actions-read). Its enclosing native signature, preceding local bindings and selected native module are required.

<!-- example:glossary-builtins-eng-storage-program-57-57 -->

</details>

### Runtime.Capture.Adapt

<a id="word-authoring-runtime-capture-adapt"></a>

<!-- word:authoring:Runtime.Capture.Adapt -->

Convert a present concrete async reader into the declared boxed reader interface while preserving absence.

<!-- word-definition:authoring:Runtime.Capture.Adapt -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

This statement is used inside `Runtime.Capture.Read`. Its enclosing native signature and preceding bindings remain required.

<!-- example:glossary-runtime-capture-adapt-caller -->

</details>

### Runtime.Error.Address

<a id="word-authoring-runtime-error-address"></a>

<!-- word:authoring:Runtime.Error.Address -->

Read the allocation pointer of the selected Arc-backed ErrorData without consuming it. The selected operation is `std::sync::Arc::as_ptr(&Value.data)`.

<!-- word-definition:authoring:Runtime.Error.Address -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

This statement is used inside `Runtime.Error.Identify`. Its enclosing native signature and preceding bindings remain required.

<!-- example:glossary-runtime-error-address-caller -->

</details>

### Storage.Address

<a id="word-authoring-storage-address"></a>

<!-- word:authoring:Storage.Address -->

Read the allocation pointer of the supplied Rc handle without consuming its payload. The selected operation is `std::rc::Rc::as_ptr(Value)`.

<!-- word-definition:authoring:Storage.Address -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

This statement is used inside `Storage.Identify`. Its enclosing native signature and preceding bindings remain required.

<!-- example:glossary-storage-address-caller -->

</details>

### Adjective

<a id="word-type-adjective"></a>

<!-- word:type:adjective -->

Classify a word that qualifies a noun or describes construction for exact nominal targets.

<!-- word-definition:type:adjective -->

**Usage and example.** [Main](#word-builtin-main) declares Service, Object and Function targets. The [recursive fixture](../../lyger/test/classes/recursive.l) applies Main's Object rule through further descriptors. An unrelated same-spelled target does not satisfy that declaration.

### Runtime.Admission

<a id="word-authoring-runtime-admission"></a>

<!-- word:authoring:Runtime.Admission -->

Borrow `Value.state` without moving the referenced native field.

<!-- word-definition:authoring:Runtime.Admission -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

This statement is used inside `Runtime.Choose`. Its enclosing native signature and preceding bindings remain required.

<!-- example:glossary-runtime-admission-caller -->

</details>

### Runtime.App.Admission

<a id="word-authoring-runtime-app-admission"></a>

<!-- word:authoring:Runtime.App.Admission -->

Borrow `Value.state` without moving the referenced native field.

<!-- word-definition:authoring:Runtime.App.Admission -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

This statement is used inside `Runtime.App.Describe`. Its enclosing native signature and preceding bindings remain required.

<!-- example:glossary-runtime-app-admission-caller -->

</details>

### Runtime.Invocation.Admission

<a id="word-authoring-runtime-invocation-admission"></a>

<!-- word:authoring:Runtime.Invocation.Admission -->

Borrow `Value.state` without moving the referenced native field.

<!-- word-definition:authoring:Runtime.Invocation.Admission -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

This statement is used inside `Runtime.Invocation.Claim`. Its enclosing native signature and preceding bindings remain required.

<!-- example:glossary-runtime-invocation-admission-caller -->

</details>

### Runtime.Admit

<a id="word-authoring-runtime-admit"></a>

<!-- word:authoring:Runtime.Admit -->

Return true only when runtime state is neither stopping, terminal-selected, nor finalized.

<!-- word-definition:authoring:Runtime.Admit -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

This statement is used inside `Runtime.Check`. Its enclosing native signature and preceding bindings remain required.

<!-- example:glossary-runtime-admit-caller -->

</details>

### Runtime.Path.Admit

<a id="word-authoring-runtime-path-admit"></a>

<!-- word:authoring:Runtime.Path.Admit -->

Accept only normal or current-directory native path components; reject roots, prefixes and parent-directory components.

<!-- word-definition:authoring:Runtime.Path.Admit -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

This statement is used inside `Runtime.App.Locate`. Its enclosing native signature and preceding bindings remain required.

<!-- example:glossary-runtime-path-admit-caller -->

</details>

### Runtime.Advance

<a id="word-authoring-runtime-advance"></a>

<!-- word:authoring:Runtime.Advance -->

Mutably borrow `Value.next` so the caller can update that exact native field.

<!-- word-definition:authoring:Runtime.Advance -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

This statement is used inside `Runtime.Count`. Its enclosing native signature and preceding bindings remain required.

<!-- example:glossary-runtime-advance-caller -->

</details>

### Runtime.Capture.Advance

<a id="word-authoring-runtime-capture-advance"></a>

<!-- word:authoring:Runtime.Capture.Advance -->

Race a stream read against the current cleanup deadline and deadline changes. Record read, write, closed-notification, or expired-deadline failures and return whether capture should stop.

<!-- word-definition:authoring:Runtime.Capture.Advance -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

This statement is used inside `Runtime.Capture.Start`. Its enclosing native signature and preceding bindings remain required.

<!-- example:glossary-runtime-capture-advance-caller -->

</details>

### Adverb

<a id="word-type-adverb"></a>

<!-- word:type:adverb -->

Classify a word that modifies an action, declaration, comparison or source sequence.

<!-- word-definition:type:adverb -->

**Usage.** Publicly on the Identity declaration in the [Function example](../examples/composed-functions.l) publishes that Function, while Then in a [source action](../guide/definitions.md#compose-a-source-action) separates composed actions.

**Partial support.** The grammar recognizes adverbs, and configured roles such as visibility, condition negation, comparison and progression have specific handling. Declaring a new Adverb does not provide a general mechanism for applying its authored body to an action. Adverb bodies have no counterpart to the source-defined adjective constructor rules. Use the implemented configured modifiers in their documented contexts.

### Storage.Agreement

<a id="word-authoring-storage-agreement"></a>

<!-- word:authoring:Storage.Agreement -->

Compute the native Boolean conjunction of the supplied left and right operands.

<!-- word-definition:authoring:Storage.Agreement -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

This statement is used inside `Formatting.Balance`. Its enclosing native signature and preceding bindings remain required.

<!-- example:glossary-storage-agreement-caller -->

</details>

### Runtime.Invocation.Allocate

<a id="word-authoring-runtime-invocation-allocate"></a>

<!-- word:authoring:Runtime.Invocation.Allocate -->

Construct shared invocation state with weak App/runtime owners, sequence and host dependencies, completion and deadline channels, process-group gate, and a termination-request channel.

<!-- word-definition:authoring:Runtime.Invocation.Allocate -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

This statement is used inside `Runtime.App.Acquire`. Its enclosing native signature and preceding bindings remain required.

<!-- example:glossary-runtime-invocation-allocate-caller -->

</details>

### Storage.Slot.Allocate

<a id="word-authoring-storage-slot-allocate"></a>

<!-- word:authoring:Storage.Slot.Allocate -->

Allocate a RefCell containing the requested number of empty optional slots.

<!-- word-definition:authoring:Storage.Slot.Allocate -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

This statement is used inside [Storage.Program.Create](#word-authoring-storage-program-create). Its enclosing native signature, preceding local bindings and selected native module are required.

<!-- example:glossary-builtins-eng-storage-program-48-48 -->

</details>

### Storage.Alternative

<a id="word-authoring-storage-alternative"></a>

<!-- word:authoring:Storage.Alternative -->

Compute the native Boolean disjunction of the supplied left and right operands.

<!-- word-definition:authoring:Storage.Alternative -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

This statement is used inside `Storage.Signature.Match`. Its enclosing native signature and preceding bindings remain required.

<!-- example:glossary-storage-alternative-caller -->

</details>

### Runtime.Amend

<a id="word-authoring-runtime-amend"></a>

<!-- word:authoring:Runtime.Amend -->

Apply the supplied modifying callback to a watch channel's current value and notify its receivers. The selected operation is `Sender.send_modify(Modifier)`.

<!-- word-definition:authoring:Runtime.Amend -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

This statement is used inside `Runtime.Revise`. Its enclosing native signature and preceding bindings remain required.

<!-- example:glossary-runtime-amend-caller -->

</details>

### And

<a id="word-grammar-word-and"></a>

<!-- word:grammar-word:and -->

Coordinate source operands, declarations or actions in their owning grammatical sequence.

<!-- word-definition:grammar-word:and -->

**Usage and boundary.** `Start First and Second` in the [two-App program](../../demo/phase1/main.l) expands the complete action for each operand.

**Partial support.** Operand, declaration and action coordination work, but an `and`-joined condition such as ``If true and false, Show `value`.`` is rejected during instruction lowering. The grammar recognizes the conditions, but there is no Boolean join operator or short-circuit implementation for them.

### Storage.Annotation

<a id="word-authoring-storage-annotation"></a>

<!-- word:authoring:Storage.Annotation -->

Create a native binding with the supplied explicit type and initializer.

<!-- word-definition:authoring:Storage.Annotation -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

This statement is used inside `runtime/eng/process.l`. Its enclosing native signature and preceding bindings remain required.

<!-- example:glossary-storage-annotation-caller -->

</details>

### Any

<a id="word-grammar-word-any"></a>

<!-- word:grammar-word:any -->

Declare a variable argument or nominal target in a supported source signature or descriptor.

<!-- word-definition:grammar-word:any -->

**Usage and boundary.** Start declares `any App` in the [source-action example](../spec/definitions.md#source-actions), while Main declares exact target categories. Any does not erase the supplied value's type or create unrestricted native generics.

### App

<a id="word-type-app"></a>

<!-- word:type:app -->

Own an executable source composition and its separately supervised native process invocations.

<!-- word-definition:type:app -->

**Usage and example.** `Create "Hello" as a New Basic App` in the [first program](../guide/first-program.md#read-the-program) declares the App. Execute then runs its body. An App alias shares its invocation identity.

#### Create and execute an App

<a id="word-type-app--create-and-execute-an-app"></a>

The [first program](../guide/first-program.md#read-the-program) constructs a Basic App, supplies its Print body and observes its captured output. Basic composes the Main Service, Object and entry Function. Execute builds, runs and awaits the App.

`l build` publishes the native program. Running that program performs its lifecycle actions. Aliases preserve App identity. Separate declarations remain separate Apps even if they use the same name in different private scopes.

### Runtime.App

<a id="word-authoring-runtime-app"></a>

<!-- word:authoring:Runtime.App -->

Declare the native `App` record with fields `runtime: std::sync::Weak<RuntimeInner>`, `path: std::path::PathBuf`, `state: std::sync::Mutex<AppState>`.

<!-- word-definition:authoring:Runtime.App -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

This statement is used inside `Runtime.App.Launch`. Its enclosing native signature and preceding bindings remain required.

<!-- example:glossary-runtime-app-caller -->

</details>

### Runtime.Invocation.App

<a id="word-authoring-runtime-invocation-app"></a>

<!-- word:authoring:Runtime.Invocation.App -->

Borrow `Value.app` without moving the referenced native field.

<!-- word-definition:authoring:Runtime.Invocation.App -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

This statement is used inside `Runtime.Invocation.Finish`. Its enclosing native signature and preceding bindings remain required.

<!-- example:glossary-runtime-invocation-app-caller -->

</details>

### Storage.App

<a id="word-authoring-storage-app"></a>

<!-- word:authoring:Storage.App -->

Name the Storage namespace for native App handles, source attribution, process-member access and shared-handle retention.

<!-- word-definition:authoring:Storage.App -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

This statement is used inside `Storage.Program.Select`. Its enclosing native signature and preceding bindings remain required.

<!-- example:glossary-storage-app-caller -->

</details>

### Formatting.Append

<a id="word-authoring-formatting-append"></a>

<!-- word:authoring:Formatting.Append -->

Append enough zero digits to place the decimal point after the requested position, returning the extended digit string.

<!-- word-definition:authoring:Formatting.Append -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

This statement is used inside `Formatting.Expand`. Its enclosing native signature and preceding bindings remain required.

<!-- example:glossary-formatting-append-caller -->

</details>

### Runtime.Error.Append

<a id="word-authoring-runtime-error-append"></a>

<!-- word:authoring:Runtime.Error.Append -->

Append an Error's diagnostic bytes once per error identity, preserving child order and LF separation. Empty joined parents delegate to children instead of adding an empty parent line.

<!-- word-definition:authoring:Runtime.Error.Append -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

This statement is used inside `Runtime.Error.Render`. Its enclosing native signature and preceding bindings remain required.

<!-- example:glossary-runtime-error-append-caller -->

</details>

### Storage.Append

<a id="word-authoring-storage-append"></a>

<!-- word:authoring:Storage.Append -->

Append the supplied value to the selected native vector. The selected operation is `Collection.push(Value)`.

<!-- word-definition:authoring:Storage.Append -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

This statement is used inside [Storage.Slot.Populate](#word-authoring-storage-slot-populate). Its enclosing native signature, preceding local bindings and selected native module are required.

<!-- example:glossary-builtins-eng-storage-program-17-17 -->

</details>

### Storage.List.Append

<a id="word-authoring-storage-list-append"></a>

<!-- word:authoring:Storage.List.Append -->

Validate a value against a List's complete element signature, then append the value to the List's owned vector.

<!-- word-definition:authoring:Storage.List.Append -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

The wrapper below imports `Storage` and `Runtime` and is a member of the example `GlossaryExamples` Dictionary. Its declared native inputs, outputs, lifetime/generic descriptors, and the original selected module's native types/support are required. Lyger resolves the source and composes its library AML; this fragment has no native compilation or execution claim.

<!-- example:glossary-native-storage-list-append -->

</details>

### Storage.Apply

<a id="word-authoring-storage-apply"></a>

<!-- word:authoring:Storage.Apply -->

Invoke the supplied typed callback with Item and Context, preserving that argument order and binding its result.

<!-- word-definition:authoring:Storage.Apply -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

This statement is used inside `builtins/eng/storage/bindings.l`. Its enclosing native signature and preceding bindings remain required.

<!-- example:glossary-storage-apply-caller -->

</details>

### Runtime.Appoint

<a id="word-authoring-runtime-appoint"></a>

<!-- word:authoring:Runtime.Appoint -->

Create an Arc-backed App handle with its companion path, weak runtime owner, and initial mutex-protected App state.

<!-- word-definition:authoring:Runtime.Appoint -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

The wrapper below imports `Storage` and `Runtime` and is a member of the example `GlossaryExamples` Dictionary. Its declared native inputs, outputs, lifetime/generic descriptors, and the original selected module's native types/support are required. Lyger resolves the source and composes its library AML; this fragment has no native compilation or execution claim.

<!-- example:glossary-native-runtime-appoint -->

</details>

### Storage.Program.Apps

<a id="word-authoring-storage-program-apps"></a>

<!-- word:authoring:Storage.Program.Apps -->

Borrow `Program.apps` without moving the referenced native field.

<!-- word-definition:authoring:Storage.Program.Apps -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

This statement is used inside `Storage.Program.Select`. Its enclosing native signature and preceding bindings remain required.

<!-- example:glossary-storage-program-apps-caller -->

</details>

### As

<a id="word-grammar-word-as"></a>

<!-- word:grammar-word:as -->

Relate a declaration or alias to its value, category, construction or new local name.

<!-- word-definition:grammar-word:as -->

**Usage and boundary.** `Create "Hello" as a New Basic App` and the IO import alias in the [first program](../guide/first-program.md#read-the-program) use distinct owning actions. As alone does not convert runtime types.

### Runtime.Ascend

<a id="word-authoring-runtime-ascend"></a>

<!-- word:authoring:Runtime.Ascend -->

Attempt to upgrade the selected weak shared-ownership reference to a strong handle. The selected operation is `Value.upgrade()`.

<!-- word-definition:authoring:Runtime.Ascend -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

This statement is used inside `Runtime.Signals.Watch`. Its enclosing native signature and preceding bindings remain required.

<!-- example:glossary-runtime-ascend-caller -->

</details>

### Runtime.Assemble

<a id="word-authoring-runtime-assemble"></a>

<!-- word:authoring:Runtime.Assemble -->

Construct a shared runtime handle from host dependencies and output streams, creating its state, cancellation and completion channels, signal-task slot, stream locks, and finishing gate.

<!-- word-definition:authoring:Runtime.Assemble -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

This statement is used inside `Runtime.Install`. Its enclosing native signature and preceding bindings remain required.

<!-- example:glossary-runtime-assemble-caller -->

</details>

### Runtime.Assess

<a id="word-authoring-runtime-assess"></a>

<!-- word:authoring:Runtime.Assess -->

Collect recorded cleanup errors in runtime order, join them when present, and return either the combined failure or successful unit.

<!-- word-definition:authoring:Runtime.Assess -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

This statement is used inside `Runtime.Reap`. Its enclosing native signature and preceding bindings remain required.

<!-- example:glossary-runtime-assess-caller -->

</details>

### Storage.Assign

<a id="word-authoring-storage-assign"></a>

<!-- word:authoring:Storage.Assign -->

Assign the supplied value to the named native binding; declaration is separate.

<!-- word-definition:authoring:Storage.Assign -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

**Compiler use.** Select direct assignment for ordinary storage and publish into optional storage when required. This retained compiler excerpt selects the L-owned definition with typed native name/type/value or Process bindings. Those bindings come from the compiler’s resolved plan; this is an integration fragment, not an ordinary standalone L call.

<!-- example:glossary-lyger-runtime-value-words-177-181 -->

</details>

### Runtime.Asynchronous

<a id="word-authoring-runtime-asynchronous"></a>

<!-- word:authoring:Runtime.Asynchronous -->

Replace the canonical Runtime.Mutex construction with a Tokio asynchronous mutex around the original value.

<!-- word-definition:authoring:Runtime.Asynchronous -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

This statement is used inside `Runtime.Assemble`. Its enclosing native signature and preceding bindings remain required.

<!-- example:glossary-runtime-asynchronous-caller -->

</details>

### At

<a id="word-grammar-word-at"></a>

<!-- word:grammar-word:at -->

Select a Number index or Text key for a supported typed read or update.

<!-- word-definition:grammar-word:at -->

**Usage and boundary.** The [value program](../examples/word-values.l) uses `items at 0` and the Text key `first` in names. The containing operation checks the key type, bounds and result type. At supplies no unchecked access.

### Runtime.Attach

<a id="word-authoring-runtime-attach"></a>

<!-- word:authoring:Runtime.Attach -->

Store a spawned signal-watcher task in the runtime's protected optional task slot.

<!-- word-definition:authoring:Runtime.Attach -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

This statement is used inside `Runtime.Install`. Its enclosing native signature and preceding bindings remain required.

<!-- example:glossary-runtime-attach-caller -->

</details>

### Runtime.Error.Attach

<a id="word-authoring-runtime-error-attach"></a>

<!-- word:authoring:Runtime.Error.Attach -->

Preserve an Error when stderr is empty; otherwise construct a parent message from its rendered text, colon-space, and the exact stderr bytes, retaining the original Error as a cause.

<!-- word-definition:authoring:Runtime.Error.Attach -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

This statement is used inside `Runtime.Supervision.Decorate`. Its enclosing native signature and preceding bindings remain required.

<!-- example:glossary-runtime-error-attach-caller -->

</details>

### Attribute

<a id="word-type-attribute"></a>

<!-- word:type:attribute -->

Name a composed part of an Object in its applicational representation.

<!-- word-definition:type:attribute -->

**Context and illustration.** A Function inside the [Object fixture](../../lyger/test/classes/interleaved.l) becomes an Attribute containing its Definition.

**Partial support.** Attribute is recognized as an applicational category, and [Object projection](../../lyger/source_hierarchy.go) creates Attributes for composed Functions. It has no standalone data-field or runtime-value constructor: using Attribute as a source Function parameter or List element [rejects](../../lyger/runtime_types.go) with `Attribute is a source structure, not a runtime type`. Use Function/Object composition for the implemented form.

**Usage illustration.** In the linked Object fixture, `New Objective First Function` supplies a Function definition to the Object. The compiler represents that composed item as an Attribute. This illustrates the implemented structural role; it is not an `Attribute` runtime constructor.

### Storage.App.Attribute

<a id="word-authoring-storage-app-attribute"></a>

<!-- word:authoring:Storage.App.Attribute -->

Construct source-attribution text once and apply it to the supplied App handle.

<!-- word-definition:authoring:Storage.App.Attribute -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

**Compiler use.** Attach the resolved source attribution text to the selected App handle. This retained compiler excerpt selects the L-owned definition with typed native name/type/value or Process bindings. Those bindings come from the compiler’s resolved plan; this is an integration fragment, not an ordinary standalone L call.

<!-- example:glossary-lyger-execution-composition-429-436 -->

</details>

### Storage.Audit

<a id="word-authoring-storage-audit"></a>

<!-- word:authoring:Storage.Audit -->

Visit pair-shaped items in iterator order, invoking the supplied callback with both items and context. Discard successful callback payloads and propagate the first failure.

<!-- word-definition:authoring:Storage.Audit -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

This statement is used inside `Storage.Map.Render`. Its enclosing native signature and preceding bindings remain required.

<!-- example:glossary-storage-audit-caller -->

</details>

### Runtime.Authority

<a id="word-authoring-runtime-authority"></a>

<!-- word:authoring:Runtime.Authority -->

Borrow `Value.dependencies` without moving the referenced native field.

<!-- word-definition:authoring:Runtime.Authority -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

This statement is used inside `Runtime.App.Acquire`. Its enclosing native signature and preceding bindings remain required.

<!-- example:glossary-runtime-authority-caller -->

</details>

### Runtime.Invocation.Available

<a id="word-authoring-runtime-invocation-available"></a>

<!-- word:authoring:Runtime.Invocation.Available -->

Check whether the selected native Option contains a payload. The selected operation is `Value.snapshot.is_some()`.

<!-- word-definition:authoring:Runtime.Invocation.Available -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

This statement is used inside `Runtime.Invocation.Claim`. Its enclosing native signature and preceding bindings remain required.

<!-- example:glossary-runtime-invocation-available-caller -->

</details>

### Await

<a id="word-builtin-await"></a>

<!-- word:builtin:await -->

Join an App invocation's finalized result and activate its continuation once.

<!-- word-definition:builtin:await -->

**Usage and example.** The [continuation example](../spec/execution.md#continuations-and-activation) awaits Worker twice, but only the first join runs the paragraph. `Worker is awaited` is the supported outer passive form.

Await joins reaping, captured streams and owned-group cleanup. Failed launches and stopped Apps still have awaitable results. The [invocation contract](../cli-build.md#invocation-and-results) owns continuation claims, cached/self-Await, nested Await and slot release.

Imperative, outer passive and source-action applicative forms are supported. Native App-body Await rejects ownership, and standalone Service body effects reject. See [supported language forms](../cli-build.md#launch-and-language-forms).

### Runtime.Supervision.Await

<a id="word-authoring-runtime-supervision-await"></a>

<!-- word:authoring:Runtime.Supervision.Await -->

Poll the child wait alongside termination requests until the wait completes, close the request receiver, acknowledge queued requests, and return the saved wait result.

<!-- word-definition:authoring:Runtime.Supervision.Await -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

This statement is used inside `Runtime.Invocation.Supervise`. Its enclosing native signature and preceding bindings remain required.

<!-- example:glossary-runtime-supervision-await-caller -->

</details>

### Storage.Await

<a id="word-authoring-storage-await"></a>

<!-- word:authoring:Storage.Await -->

Await the supplied future and bind its outcome to the explicit result label. This operation neither constructs nor pins the future.

<!-- word-definition:authoring:Storage.Await -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

This statement is used inside `Delay`. Its enclosing native signature and preceding bindings remain required.

<!-- example:glossary-storage-await-caller -->

</details>

## B

<a id="glossary-b"></a>

### Formatting.Balance

<a id="word-authoring-formatting-balance"></a>

<!-- word:authoring:Formatting.Balance -->

Compare a decimal midpoint representation with the exact binary significand and exponent of a finite float to identify a formatting tie.

<!-- word-definition:authoring:Formatting.Balance -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

This statement is used inside `Formatting.Render`. Its enclosing native signature and preceding bindings remain required.

<!-- example:glossary-formatting-balance-caller -->

</details>

### Basic

<a id="word-builtin-basic"></a>

<!-- word:builtin:basic -->

Describe an App containing a Main Service that receives the App's source Process.

<!-- word-definition:builtin:basic -->

**Usage and example.** `Create "Hello" as a New Basic App` in the [first program](../guide/first-program.md#read-the-program) creates the body owner. Execute is a separate lifecycle action. Construction alone launches nothing.

Basic describes a Generic App with a New Main Service using a Process. It is the composition behind `Create "Hello" as a New Basic App.` The App's source actions supply its body, and the orchestrator separately controls launch and output.

### Runtime.App.Before

<a id="word-authoring-runtime-app-before"></a>

<!-- word:authoring:Runtime.App.Before -->

Bind the native constant `crate::runtime::host::Checkpoint::PreStart` for this composition.

<!-- word-definition:authoring:Runtime.App.Before -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

This statement is used inside `Runtime.App.Run`. Its enclosing native signature and preceding bindings remain required.

<!-- example:glossary-runtime-app-before-caller -->

</details>

### Runtime.Begin

<a id="word-authoring-runtime-begin"></a>

<!-- word:authoring:Runtime.Begin -->

Claim shutdown ownership once and spawn cleanup of the captured owned-invocation set. Repeated calls after shutdown begins do nothing.

<!-- word-definition:authoring:Runtime.Begin -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

This statement is used inside `Runtime.Clean`. Its enclosing native signature and preceding bindings remain required.

<!-- example:glossary-runtime-begin-caller -->

</details>

### Storage.Function.Bind

<a id="word-authoring-storage-function-bind"></a>

<!-- word:authoring:Storage.Function.Bind -->

Construct and check the capture environment, cast the selected function target and construct their tuple. Environment failure propagates before the Function value is published.

<!-- word-definition:authoring:Storage.Function.Bind -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

**Compiler use.** functionOperation selects Storage.Function.Bind and binds the typed Function pointer and captured environment. This retained compiler excerpt selects the L-owned definition with typed native name/type/value or Process bindings. Those bindings come from the compiler’s resolved plan; this is an integration fragment, not an ordinary standalone L call. The `functionOperation` helper in `lyger/function_composition.go` selects `Storage.Function.` plus the supplied member name and preserves the selected Return composition.

<!-- example:glossary-lyger-function-composition-421-429 -->

</details>

### Storage.Binding

<a id="word-authoring-storage-binding"></a>

<!-- word:authoring:Storage.Binding -->

Declare a native local binding initialized with the supplied value.

<!-- word-definition:authoring:Storage.Binding -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

This statement is used inside `Runtime.Error.Cancel`. Its enclosing native signature and preceding bindings remain required.

<!-- example:glossary-storage-binding-caller -->

</details>

### Storage.Borrow

<a id="word-authoring-storage-borrow"></a>

<!-- word:authoring:Storage.Borrow -->

Create an immutable native borrow of the supplied value.

<!-- word-definition:authoring:Storage.Borrow -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

This statement is used inside `Storage.Text.Enclose`. Its enclosing native signature and preceding bindings remain required.

<!-- example:glossary-storage-borrow-caller -->

</details>

### Runtime.Bound

<a id="word-authoring-runtime-bound"></a>

<!-- word:authoring:Runtime.Bound -->

Create the native final process-group cleanup operation from a process ID, signal result and optional deadline. The selected operation is `crate::process::final_group_result(Pid, Result, Deadline)`.

<!-- word-definition:authoring:Runtime.Bound -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

This statement is used inside `Runtime.Invocation.Signal`. Its enclosing native signature and preceding bindings remain required.

<!-- example:glossary-runtime-bound-caller -->

</details>

### Runtime.Invocation.Bound

<a id="word-authoring-runtime-invocation-bound"></a>

<!-- word:authoring:Runtime.Invocation.Bound -->

Publish a cleanup deadline through the invocation's deadline sender, preserving the earlier existing deadline.

<!-- word-definition:authoring:Runtime.Invocation.Bound -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

This statement is used inside `Runtime.Invocation.Supervise`. Its enclosing native signature and preceding bindings remain required.

<!-- example:glossary-runtime-invocation-bound-caller -->

</details>

### Runtime.Capture.Box

<a id="word-authoring-runtime-capture-box"></a>

<!-- word:authoring:Runtime.Capture.Box -->

Allocate a native Box containing the supplied value or mutable reference. The selected operation is `Box::new(Value)`.

<!-- word-definition:authoring:Runtime.Capture.Box -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

This statement is used inside `Runtime.Capture.Start`. Its enclosing native signature and preceding bindings remain required.

<!-- example:glossary-runtime-capture-box-caller -->

</details>

### Runtime.Capture.Buffer

<a id="word-authoring-runtime-capture-buffer"></a>

<!-- word:authoring:Runtime.Capture.Buffer -->

Create a mutable zero-filled 8192-byte array for stream capture reads.

<!-- word-definition:authoring:Runtime.Capture.Buffer -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

This statement is used inside `Runtime.Capture.Start`. Its enclosing native signature and preceding bindings remain required.

<!-- example:glossary-runtime-capture-buffer-caller -->

</details>

### Storage.Map.Buffer

<a id="word-authoring-storage-map-buffer"></a>

<!-- word:authoring:Storage.Map.Buffer -->

Construct an empty native ordered map with key and value types selected by its enclosing composition. The selected operation is `std::collections::BTreeMap::new()`.

<!-- word-definition:authoring:Storage.Map.Buffer -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

This statement is used inside `Storage.Map.Create`. Its enclosing native signature and preceding bindings remain required.

<!-- example:glossary-storage-map-buffer-caller -->

</details>

### Storage.Vector.Buffer

<a id="word-authoring-storage-vector-buffer"></a>

<!-- word:authoring:Storage.Vector.Buffer -->

Construct an empty native vector with the type required by its declared binding or enclosing composition. The selected operation is `Vec::new()`.

<!-- word-definition:authoring:Storage.Vector.Buffer -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

This statement is used inside `Runtime.Error.Render`. Its enclosing native signature and preceding bindings remain required.

<!-- example:glossary-storage-vector-buffer-caller -->

</details>

### Build

<a id="word-builtin-build"></a>

<!-- word:builtin:build -->

Check that an App's prebuilt executable is available, recording the operation result without launching it.

<!-- word-definition:builtin:build -->

**Usage and example.** `Build Worker` in the [explicit lifecycle example](../spec/execution.md#build-run-stop-and-await) precedes Run. It is not the CLI compilation command.

Use an explicit Error check when handling the result. Build creates no invocation, and building while one remains outstanding is misuse. Run revalidates the artifact, including after Build failure. See [invocation and result rules](../cli-build.md#invocation-and-results).

### Runtime.App.Build

<a id="word-authoring-runtime-app-build"></a>

<!-- word:authoring:Runtime.App.Build -->

Check runtime admission and reject an outstanding invocation, validate the prebuilt App companion path, record any path error, and clear the selected invocation. It does not invoke a compiler. Missing or invalid path failures are stored in the App operation error while the Function itself returns Ok(()); callers inspect App.Error for that recorded result.

<!-- word-definition:authoring:Runtime.App.Build -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

The wrapper below imports `Storage` and `Runtime` and is a member of the example `GlossaryExamples` Dictionary. Its declared native inputs, outputs, lifetime/generic descriptors, and the original selected module's native types/support are required. Lyger resolves the source and composes its library AML; this fragment has no native compilation or execution claim.

<!-- example:glossary-native-runtime-app-build -->

</details>

### Runtime.App.Busy

<a id="word-authoring-runtime-app-busy"></a>

<!-- word:authoring:Runtime.App.Busy -->

Select the native field `Value.outstanding` by value. Its declared type determines whether this copies or moves the field.

<!-- word-definition:authoring:Runtime.App.Busy -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

This statement is used inside `Runtime.App.Build`. Its enclosing native signature and preceding bindings remain required.

<!-- example:glossary-runtime-app-busy-caller -->

</details>

### By

<a id="word-grammar-word-by"></a>

<!-- word:grammar-word:by -->

Introduce a Function argument or name an actor in the configured indicative phrase.

<!-- word-definition:grammar-word:by -->

**Usage example (source fragment).** After declaring Identity in the [Function example](../examples/composed-functions.l), ``Identity by `value` as "Result".`` uses the generic argument phrase. This binds an ordinary argument. It does not select an actor or callback.

**Partial support.** The [indicative grammar](../../lexicon/eng/grammar.json) declares `by Actor`, but no enclosing clause uses that phrase to implement passive actor execution. Generic Function arguments work, but the configured actor-specific meaning does not.

### Byte

<a id="word-type-byte"></a>

<!-- word:type:byte -->

Name the byte-oriented lexical category whose supported owned storage form is Bytes.

<!-- word-definition:type:byte -->

**Usage.** Use the plural [Bytes](#word-type-bytes) constructor for an owned byte sequence, then read an element as Number. The [byte excerpt](../examples/word-values.l) reads the `a` octet as 97.

**Partial support.** The configured Byte/Bytes word family is recognized, but only the plural Bytes storage type is implemented. The source-type whitelist has no singular Byte scalar type: individual octets are read as Number, and `New Byte` does not supply a one-byte value constructor. See the [source-type boundary test](../../lyger/refactor_value_characterization_test.go).

### Bytes

<a id="word-type-bytes"></a>

<!-- word:type:bytes -->

Own a sequence of unchanged bytes copied from Text.

<!-- word-definition:type:bytes -->

**Usage and example.** `New Bytes of` constructs the buffer. A checked Number read selects a zero-based octet. The complete [value program](../examples/word-values.l) contains this excerpt:

<!-- example:word-values-bytes -->

Indices must be nonnegative integers within bounds. Bytes mutation is not exposed, and byte indexing is not Unicode character indexing.

### Runtime.Capture.Bytes

<a id="word-authoring-runtime-capture-bytes"></a>

<!-- word:authoring:Runtime.Capture.Bytes -->

Select the native field `Value.bytes` by value. Its declared type determines whether this copies or moves the field.

<!-- word-definition:authoring:Runtime.Capture.Bytes -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

This statement is used inside `Runtime.Capture.Share`. Its enclosing native signature and preceding bindings remain required.

<!-- example:glossary-runtime-capture-bytes-caller -->

</details>

### Runtime.Error.Bytes

<a id="word-authoring-runtime-error-bytes"></a>

<!-- word:authoring:Runtime.Error.Bytes -->

Borrow the selected vector as a native slice. The selected operation is `Value.data.message.as_slice()`.

<!-- word-definition:authoring:Runtime.Error.Bytes -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

This statement is used inside `Runtime.Error.Read`. Its enclosing native signature and preceding bindings remain required.

<!-- example:glossary-runtime-error-bytes-caller -->

</details>

### Storage.Bytes

<a id="word-authoring-storage-bytes"></a>

<!-- word:authoring:Storage.Bytes -->

Name the Storage namespace for byte-slice extension, checked byte reads and byte-vector display.

<!-- word-definition:authoring:Storage.Bytes -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

This statement is used inside `Runtime.Error.Attach`. Its enclosing native signature and preceding bindings remain required.

<!-- example:glossary-storage-bytes-caller -->

</details>

### Storage.Text.Bytes

<a id="word-authoring-storage-text-bytes"></a>

<!-- word:authoring:Storage.Text.Bytes -->

Borrow a native string as its exact UTF-8 byte slice. The selected operation is `Value.as_bytes()`.

<!-- word-definition:authoring:Storage.Text.Bytes -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

This statement is used inside `Storage.Text.Own`. Its enclosing native signature and preceding bindings remain required.

<!-- example:glossary-storage-text-bytes-caller -->

</details>

## C

<a id="glossary-c"></a>

### Runtime.Error.Cancel

<a id="word-authoring-runtime-error-cancel"></a>

<!-- word:authoring:Runtime.Error.Cancel -->

Construct the cancellation Error `source execution stopped` with its cancellation flag set.

<!-- word-definition:authoring:Runtime.Error.Cancel -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

This statement is used inside `Runtime.Leave`. Its enclosing native signature and preceding bindings remain required.

<!-- example:glossary-runtime-error-caller -->

</details>

### Runtime.Inner.Cancellation

<a id="word-authoring-runtime-inner-cancellation"></a>

<!-- word:authoring:Runtime.Inner.Cancellation -->

Borrow `Value.cancelled` without moving the referenced native field.

<!-- word-definition:authoring:Runtime.Inner.Cancellation -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

This statement is used inside `Runtime.Gather`. Its enclosing native signature and preceding bindings remain required.

<!-- example:glossary-runtime-inner-caller -->

</details>

### Runtime.App.Capture

<a id="word-authoring-runtime-app-capture"></a>

<!-- word:authoring:Runtime.App.Capture -->

Read an App's current invocation and return retained stdout bytes only when that invocation has a completed snapshot.

<!-- word-definition:authoring:Runtime.App.Capture -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

This statement is used inside `Runtime.App.Read`. Its enclosing native signature and preceding bindings remain required.

<!-- example:glossary-runtime-app-capture-caller -->

</details>

### Runtime.Capture

<a id="word-authoring-runtime-capture"></a>

<!-- word:authoring:Runtime.Capture -->

Declare the native `Capture` record with fields `bytes: Vec<u8>`, `error: Option<Error>`.

<!-- word-definition:authoring:Runtime.Capture -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

This statement is used inside `Runtime.Capture.Start`. Its enclosing native signature and preceding bindings remain required.

<!-- example:glossary-runtime-capture-buffer-caller -->

</details>

### Runtime.Invocation.Capture

<a id="word-authoring-runtime-invocation-capture"></a>

<!-- word:authoring:Runtime.Invocation.Capture -->

Mutably borrow `Value.snapshot` so the caller can update that exact native field.

<!-- word-definition:authoring:Runtime.Invocation.Capture -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

This statement is used inside `Runtime.Invocation.Publish`. Its enclosing native signature and preceding bindings remain required.

<!-- example:glossary-runtime-invocation-capture-caller -->

</details>

### Storage.Cast

<a id="word-authoring-storage-cast"></a>

<!-- word:authoring:Storage.Cast -->

Convert the supplied native value to the explicitly supplied native type using Rust cast semantics. This operation does not validate ranges or replace checked language conversions.

<!-- word-definition:authoring:Storage.Cast -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

This statement is used inside `Storage.Identify`. Its enclosing native signature and preceding bindings remain required.

<!-- example:glossary-storage-cast-caller -->

</details>

### Runtime.Catch

<a id="word-authoring-runtime-catch"></a>

<!-- word:authoring:Runtime.Catch -->

Wrap a source future in the host panic-catching adapter. The selected operation is `crate::runtime::host::catch(Future)`.

<!-- word-definition:authoring:Runtime.Catch -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

This statement is used inside `Runtime.Run`. Its enclosing native signature and preceding bindings remain required.

<!-- example:glossary-runtime-catch-caller -->

</details>

### Runtime.Error.Caused

<a id="word-authoring-runtime-error-caused"></a>

<!-- word:authoring:Runtime.Error.Caused -->

Describe a Storage.Routine with `E: std::error::Error + Send + Sync + 'static` so a native error can be retained as a shared cause.

<!-- word-definition:authoring:Runtime.Error.Caused -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

This statement is used inside `Runtime.Error.Explain`. Its enclosing native signature and preceding bindings remain required.

<!-- example:glossary-runtime-error-caused-caller -->

</details>

### Storage.Cell

<a id="word-authoring-storage-cell"></a>

<!-- word:authoring:Storage.Cell -->

Move the supplied value into a new RefCell with dynamically checked borrowing. The selected operation is `std::cell::RefCell::new(Value)`.

<!-- word-definition:authoring:Storage.Cell -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

This statement is used inside [Storage.Slot.Allocate](#word-authoring-storage-slot-allocate). Its enclosing native signature, preceding local bindings and selected native module are required.

<!-- example:glossary-builtins-eng-storage-program-24-24 -->

</details>

### Runtime.Change

<a id="word-authoring-runtime-change"></a>

<!-- word:authoring:Runtime.Change -->

Wait for the selected watch receiver's value to change, or report that its channel has closed. The selected operation is `Value.changed()`.

<!-- word-definition:authoring:Runtime.Change -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

This statement is used inside `Runtime.Capture.Advance`. Its enclosing native signature and preceding bindings remain required.

<!-- example:glossary-runtime-change-caller -->

</details>

### Runtime.Capture.Channel

<a id="word-authoring-runtime-capture-channel"></a>

<!-- word:authoring:Runtime.Capture.Channel -->

Bind a supplied boxed reader as the declared AsyncRead + Unpin + Send interface.

<!-- word-definition:authoring:Runtime.Capture.Channel -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

This statement is used inside `Runtime.Capture.Adapt`. Its enclosing native signature and preceding bindings remain required.

<!-- example:glossary-runtime-capture-channel-caller -->

</details>

### Runtime.Channel

<a id="word-authoring-runtime-channel"></a>

<!-- word:authoring:Runtime.Channel -->

Create an unbounded native message channel and return its sender/receiver pair. The selected operation is `tokio::sync::mpsc::unbounded_channel()`.

<!-- word-definition:authoring:Runtime.Channel -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

This statement is used inside `Runtime.Invocation.Allocate`. Its enclosing native signature and preceding bindings remain required.

<!-- example:glossary-runtime-channel-caller -->

</details>

### Runtime.Check

<a id="word-authoring-runtime-check"></a>

<!-- word:authoring:Runtime.Check -->

Return successful unit while runtime admission remains open, otherwise return the runtime cancellation Error.

<!-- word-definition:authoring:Runtime.Check -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

This statement is used inside `Runtime.App.Stop`. Its enclosing native signature and preceding bindings remain required.

<!-- example:glossary-runtime-check-caller -->

</details>

### Runtime.Error.Check

<a id="word-authoring-runtime-error-check"></a>

<!-- word:authoring:Runtime.Error.Check -->

Read a runtime Error's cancellation flag.

<!-- word-definition:authoring:Runtime.Error.Check -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

This statement is used inside `Runtime.Resolve`. Its enclosing native signature and preceding bindings remain required.

<!-- example:glossary-runtime-error-check-caller -->

</details>

### Runtime.Path.Check

<a id="word-authoring-runtime-path-check"></a>

<!-- word:authoring:Runtime.Path.Check -->

Apply the supplied predicate to successive iterator items, stopping with false at the first rejected item. The selected operation is `Values.all(Visit)`.

<!-- word-definition:authoring:Runtime.Path.Check -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

This statement is used inside `Runtime.App.Locate`. Its enclosing native signature and preceding bindings remain required.

<!-- example:glossary-runtime-path-admit-caller -->

</details>

### Storage.Environment.Check

<a id="word-authoring-storage-environment-check"></a>

<!-- word:authoring:Storage.Environment.Check -->

Check one captured value against an expected signature and return unit on success or a runtime failure with the supplied message.

<!-- word-definition:authoring:Storage.Environment.Check -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

This statement is used inside `Storage.Environment.Create`. Its enclosing native signature and preceding bindings remain required.

<!-- example:glossary-storage-environment-check-caller -->

</details>

### Storage.Index.Check

<a id="word-authoring-storage-index-check"></a>

<!-- word:authoring:Storage.Index.Check -->

Accept finite, nonnegative, integral numeric indices strictly below `usize::MAX as f64`, and return the converted native index. Reject other values with `index must be a nonnegative integer`.

<!-- word-definition:authoring:Storage.Index.Check -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

This statement is used inside `Storage.Bytes.Read`. Its enclosing native signature and preceding bindings remain required.

<!-- example:glossary-storage-index-check-caller -->

</details>

### Storage.List.Check

<a id="word-authoring-storage-list-check"></a>

<!-- word:authoring:Storage.List.Check -->

Check a value against a List element signature and use `List element type mismatch` when it does not match.

<!-- word-definition:authoring:Storage.List.Check -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

This statement is used inside `Storage.List.Create`. Its enclosing native signature and preceding bindings remain required.

<!-- example:glossary-storage-inspect-caller -->

</details>

### Storage.Value.Check

<a id="word-authoring-storage-value-check"></a>

<!-- word:authoring:Storage.Value.Check -->

Compare a stored value's full signature with the expected signature and return unit or a runtime failure with the supplied message.

<!-- word-definition:authoring:Storage.Value.Check -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

This statement is used inside `Storage.Environment.Check`. Its enclosing native signature and preceding bindings remain required.

<!-- example:glossary-storage-value-check-caller -->

</details>

### Storage.Checked

<a id="word-authoring-storage-checked"></a>

<!-- word:authoring:Storage.Checked -->

Describe a Binding so its initializer applies Rust's `?` operator, propagating a residual from the enclosing compatible Function before binding a successful value.

<!-- word-definition:authoring:Storage.Checked -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

This statement is used inside `Storage.Resolve`. Its enclosing native signature and preceding bindings remain required.

<!-- example:glossary-storage-checked-caller -->

</details>

### Runtime.Checkpoint

<a id="word-authoring-runtime-checkpoint"></a>

<!-- word:authoring:Runtime.Checkpoint -->

Bind the native constant `crate::runtime::host::Checkpoint::CapturedStopTarget` for this composition.

<!-- word-definition:authoring:Runtime.Checkpoint -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

This statement is used inside `Runtime.App.Stop`. Its enclosing native signature and preceding bindings remain required.

<!-- example:glossary-runtime-checkpoint-caller -->

</details>

### Runtime.Error.Children

<a id="word-authoring-runtime-error-children"></a>

<!-- word:authoring:Runtime.Error.Children -->

Borrow the selected vector as a native slice. The selected operation is `Value.data.causes.as_slice()`.

<!-- word-definition:authoring:Runtime.Error.Children -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

This statement is used inside `Runtime.Error.Inspect`. Its enclosing native signature and preceding bindings remain required.

<!-- example:glossary-runtime-error-children-caller -->

</details>

### Storage.Map.Children

<a id="word-authoring-storage-map-children"></a>

<!-- word:authoring:Storage.Map.Children -->

Obtain a borrowed iterator over the selected ordered map's values. The selected operation is `Value.values.values()`.

<!-- word-definition:authoring:Storage.Map.Children -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

This statement is used inside `Storage.Map.Reach`. Its enclosing native signature and preceding bindings remain required.

<!-- example:glossary-storage-map-children-caller -->

</details>

### Storage.Choice

<a id="word-authoring-storage-choice"></a>

<!-- word:authoring:Storage.Choice -->

Select the Then value when the native Boolean condition is true, otherwise select the Else value, and bind the selected result.

<!-- word-definition:authoring:Storage.Choice -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

This statement is used inside `Runtime.Report`. Its enclosing native signature and preceding bindings remain required.

<!-- example:glossary-storage-choice-caller -->

</details>

### Runtime.Choose

<a id="word-authoring-runtime-choose"></a>

<!-- word:authoring:Runtime.Choose -->

Select the first terminal status and optional Error while holding runtime state, then broadcast cancellation. Refuse replacement after a terminal choice or finalization.

<!-- word-definition:authoring:Runtime.Choose -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

This statement is used inside `Runtime.Request`. Its enclosing native signature and preceding bindings remain required.

<!-- example:glossary-runtime-choose-caller -->

</details>

### Runtime.Invocation.Claim

<a id="word-authoring-runtime-invocation-claim"></a>

<!-- word:authoring:Runtime.Invocation.Claim -->

Under runtime admission and invocation locks, claim a completed invocation's unclaimed continuation once. Return false when it is not ready or already claimed.

<!-- word-definition:authoring:Runtime.Invocation.Claim -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

The wrapper below imports `Storage` and `Runtime` and is a member of the example `GlossaryExamples` Dictionary. Its declared native inputs, outputs, lifetime/generic descriptors, and the original selected module's native types/support are required. Lyger resolves the source and composes its library AML; this fragment has no native compilation or execution claim.

<!-- example:glossary-native-runtime-invocation-claim -->

</details>

### Class

<a id="word-type-class"></a>

<!-- word:type:class -->

Group Objects and nested Classes into native source modules.

<!-- word-definition:type:class -->

**Usage and example.** Outer, Middle and Leaf in the [recursive construction example](../spec/applicational-types.md#compose-an-object-inside-nested-classes) retain their explicit parent order. Classes organize sources. They add no runtime scope or process lifetime.

### Runtime.Clean

<a id="word-authoring-runtime-clean"></a>

<!-- word:authoring:Runtime.Clean -->

Begin shutdown if necessary, wait for its published cleanup result, and return that result. A closed completion channel becomes an explicit runtime failure.

<!-- word-definition:authoring:Runtime.Clean -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

This statement is used inside `Runtime.Finish`. Its enclosing native signature and preceding bindings remain required.

<!-- example:glossary-runtime-clean-caller -->

</details>

### Runtime.Invocation.Cleaned

<a id="word-authoring-runtime-invocation-cleaned"></a>

<!-- word:authoring:Runtime.Invocation.Cleaned -->

Select the native field `Value.cleaned` by value. Its declared type determines whether this copies or moves the field.

<!-- word-definition:authoring:Runtime.Invocation.Cleaned -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

This statement is used inside `Runtime.Invocation.Signal`. Its enclosing native signature and preceding bindings remain required.

<!-- example:glossary-runtime-invocation-cleaned-caller -->

</details>

### Storage.Clone

<a id="word-authoring-storage-clone"></a>

<!-- word:authoring:Storage.Clone -->

Borrow a value implementing the declared Clone bound and return its native clone. The selected native type determines whether this copies data or retains shared ownership.

<!-- word-definition:authoring:Storage.Clone -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

The wrapper below imports `Storage` and `Runtime` and is a member of the example `GlossaryExamples` Dictionary. Its declared native inputs, outputs, lifetime/generic descriptors, and the original selected module's native types/support are required. Lyger resolves the source and composes its library AML; this fragment has no native compilation or execution claim.

<!-- example:glossary-native-storage-clone -->

</details>

### Storage.Cloneable

<a id="word-authoring-storage-cloneable"></a>

<!-- word:authoring:Storage.Cloneable -->

Describe a Storage.Routine with the native generic bound `T: Clone`.

<!-- word-definition:authoring:Storage.Cloneable -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

This statement is used inside `Storage.Clone`. Its enclosing native signature and preceding bindings remain required.

<!-- example:glossary-storage-cloneable-caller -->

</details>

### Runtime.Cloned

<a id="word-authoring-runtime-cloned"></a>

<!-- word:authoring:Runtime.Cloned -->

Adapt the selected borrowed iterator to yield cloned items. The selected operation is `Iterator.cloned()`.

<!-- word-definition:authoring:Runtime.Cloned -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

This statement is used inside `Runtime.Gather`. Its enclosing native signature and preceding bindings remain required.

<!-- example:glossary-runtime-cloned-caller -->

</details>

### Runtime.Supervision.Close

<a id="word-authoring-runtime-supervision-close"></a>

<!-- word:authoring:Runtime.Supervision.Close -->

Close the selected native request receiver to further sends while retaining queued requests for draining. The selected operation is `Value.close()`.

<!-- word-definition:authoring:Runtime.Supervision.Close -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

This statement is used inside `Runtime.Supervision.Await`. Its enclosing native signature and preceding bindings remain required.

<!-- example:glossary-runtime-supervision-close-caller -->

</details>

### Runtime.Code

<a id="word-authoring-runtime-code"></a>

<!-- word:authoring:Runtime.Code -->

Read a native IO error's optional operating-system error code. The selected operation is `Error.raw_os_error()`.

<!-- word-definition:authoring:Runtime.Code -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

This statement is used inside `Runtime.Normalize`. Its enclosing native signature and preceding bindings remain required.

<!-- example:glossary-runtime-code-caller -->

</details>

### Runtime.Supervision.Code

<a id="word-authoring-runtime-supervision-code"></a>

<!-- word:authoring:Runtime.Supervision.Code -->

Convert a native process ExitStatus to the platform-selected optional numeric exit code. The selected operation is `crate::process::status_code(Status)`.

<!-- word-definition:authoring:Runtime.Supervision.Code -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

This statement is used inside `Runtime.Supervision.Inspect`. Its enclosing native signature and preceding bindings remain required.

<!-- example:glossary-runtime-supervision-caller -->

</details>

### Runtime.Error.Collect

<a id="word-authoring-runtime-error-collect"></a>

<!-- word:authoring:Runtime.Error.Collect -->

Traverse child Errors in order, appending each through a shared diagnostic-output and identity-tracking context.

<!-- word-definition:authoring:Runtime.Error.Collect -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

This statement is used inside `Runtime.Error.Append`. Its enclosing native signature and preceding bindings remain required.

<!-- example:glossary-runtime-error-collect-caller -->

</details>

### Storage.List.Collect

<a id="word-authoring-storage-list-collect"></a>

<!-- word:authoring:Storage.List.Collect -->

Copy one stored value with Value.Copy and append the copied value to a supplied collection buffer.

<!-- word-definition:authoring:Storage.List.Collect -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

This statement is used inside `Storage.List.Copy`. Its enclosing native signature and preceding bindings remain required.

<!-- example:glossary-storage-list-collect-caller -->

</details>

### Storage.Map.Collect

<a id="word-authoring-storage-map-collect"></a>

<!-- word:authoring:Storage.Map.Collect -->

Copy a byte-vector key and a stored value, insert them into the supplied ordered map, and return any replaced value.

<!-- word-definition:authoring:Storage.Map.Collect -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

This statement is used inside `Storage.Map.Copy`. Its enclosing native signature and preceding bindings remain required.

<!-- example:glossary-storage-enumerate-caller -->

</details>

### Runtime.Collection

<a id="word-authoring-runtime-collection"></a>

<!-- word:authoring:Runtime.Collection -->

Collect the selected iterator into a native vector. The selected operation is `Iterator.collect()`.

<!-- word-definition:authoring:Runtime.Collection -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

This statement is used inside `Runtime.Gather`. Its enclosing native signature and preceding bindings remain required.

<!-- example:glossary-runtime-cloned-caller -->

</details>

### Runtime.Error.Collection

<a id="word-authoring-runtime-error-collection"></a>

<!-- word:authoring:Runtime.Error.Collection -->

Construct a mutable tuple holding diagnostic output, visited identities and separator state.

<!-- word-definition:authoring:Runtime.Error.Collection -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

This statement is used inside `Runtime.Error.Collect`. Its enclosing native signature and preceding bindings remain required.

<!-- example:glossary-runtime-error-collection-caller -->

</details>

### Runtime.App.Command

<a id="word-authoring-runtime-app-command"></a>

<!-- word:authoring:Runtime.App.Command -->

Construct a mutable native child-command builder for the supplied executable path. The selected operation is `tokio::process::Command::new(Path)`.

<!-- word-definition:authoring:Runtime.App.Command -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

This statement is used inside `Runtime.App.Launch`. Its enclosing native signature and preceding bindings remain required.

<!-- example:glossary-runtime-app-command-caller -->

</details>

### Storage.Compare

<a id="word-authoring-storage-compare"></a>

<!-- word:authoring:Storage.Compare -->

Expand the selected Storage.Comparison with its supplied operator, operands and output binding.

<!-- word-definition:authoring:Storage.Compare -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

**Compiler use.** Bind the already-resolved comparison operator and both operand atoms. This retained compiler excerpt selects the L-owned definition with typed native name/type/value or Process bindings. Those bindings come from the compiler’s resolved plan; this is an integration fragment, not an ordinary standalone L call.

<!-- example:glossary-lyger-execution-composition-398-401 -->

</details>

### Storage.Comparison

<a id="word-authoring-storage-comparison"></a>

<!-- word:authoring:Storage.Comparison -->

Apply the explicitly supplied native comparison operator to the left and right operands and bind its result.

<!-- word-definition:authoring:Storage.Comparison -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

This statement is used inside `builtins/eng/storage/bindings.l`. Its enclosing native signature and preceding bindings remain required.

<!-- example:glossary-storage-comparison-caller -->

</details>

### Runtime.Invocation.Complete

<a id="word-authoring-runtime-invocation-complete"></a>

<!-- word:authoring:Runtime.Invocation.Complete -->

Subscribe to an invocation's completion flag and wait until it is true or its notification channel closes.

<!-- word-definition:authoring:Runtime.Invocation.Complete -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

This statement is used inside `Runtime.Reap`. Its enclosing native signature and preceding bindings remain required.

<!-- example:glossary-runtime-invocation-complete-caller -->

</details>

### Runtime.Supervision.Complete

<a id="word-authoring-runtime-supervision-complete"></a>

<!-- word:authoring:Runtime.Supervision.Complete -->

Combine wait, capture, stream and cleanup failures, attach stderr to the joined Error, publish the finalized Snapshot, and remove the invocation from the runtime-owned set.

<!-- word-definition:authoring:Runtime.Supervision.Complete -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

This statement is used inside `Runtime.Invocation.Supervise`. Its enclosing native signature and preceding bindings remain required.

<!-- example:glossary-runtime-supervision-complete-caller -->

</details>

### Storage.Completion

<a id="word-authoring-storage-completion"></a>

<!-- word:authoring:Storage.Completion -->

Construct successful native unit, `Ok(())`.

<!-- word-definition:authoring:Storage.Completion -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

**Compiler use.** Construct and return successful unit completion for the generated entry. This retained compiler excerpt selects the L-owned definition with typed native name/type/value or Process bindings. Those bindings come from the compiler’s resolved plan; this is an integration fragment, not an ordinary standalone L call.

<!-- example:glossary-lyger-execution-composition-111-118 -->

</details>

### Runtime.Conclude

<a id="word-authoring-runtime-conclude"></a>

<!-- word:authoring:Runtime.Conclude -->

Mark runtime state finalized and return its selected terminal decision, or successful status zero with no Error when no decision was selected.

<!-- word-definition:authoring:Runtime.Conclude -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

This statement is used inside `Runtime.Finish`. Its enclosing native signature and preceding bindings remain required.

<!-- example:glossary-runtime-conclude-caller -->

</details>

### Runtime.App.Configure

<a id="word-authoring-runtime-app-configure"></a>

<!-- word:authoring:Runtime.App.Configure -->

Apply the native platform process configuration to the supplied child-command builder. The selected operation is `crate::process::configure(Command)`.

<!-- word-definition:authoring:Runtime.App.Configure -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

This statement is used inside `Runtime.App.Launch`. Its enclosing native signature and preceding bindings remain required.

<!-- example:glossary-runtime-app-configure-caller -->

</details>

### Runtime.App.Content

<a id="word-authoring-runtime-app-content"></a>

<!-- word:authoring:Runtime.App.Content -->

Construct a native `App` value with `runtime` set to `Runtime`, `path` set to `Path`, `state` set to `State`.

<!-- word-definition:authoring:Runtime.App.Content -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

This statement is used inside `Runtime.Appoint`. Its enclosing native signature and preceding bindings remain required.

<!-- example:glossary-runtime-app-content-caller -->

</details>

### Runtime.App.State.Content

<a id="word-authoring-runtime-app-state-content"></a>

<!-- word:authoring:Runtime.App.State.Content -->

Construct a native `AppState` value with `source` set to `Source`, `invocation` set to `None`, `outstanding` set to `false`, `selected` set to `None`, `operation_error` set to `None`.

<!-- word-definition:authoring:Runtime.App.State.Content -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

This statement is used inside `Runtime.Appoint`. Its enclosing native signature and preceding bindings remain required.

<!-- example:glossary-runtime-app-state-caller -->

</details>

### Runtime.Capture.Content

<a id="word-authoring-runtime-capture-content"></a>

<!-- word:authoring:Runtime.Capture.Content -->

Construct a native `Capture` value with `bytes` set to `Bytes`, `error` set to `Error`.

<!-- word-definition:authoring:Runtime.Capture.Content -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

This statement is used inside `Runtime.Capture.Start`. Its enclosing native signature and preceding bindings remain required.

<!-- example:glossary-runtime-capture-content-caller -->

</details>

### Runtime.Error.Data.Content

<a id="word-authoring-runtime-error-data-content"></a>

<!-- word:authoring:Runtime.Error.Data.Content -->

Construct a native `ErrorData` value with `message` set to `Message`, `causes` set to `Causes`, `cancelled` set to `Cancelled`, `native` set to `Native`.

<!-- word-definition:authoring:Runtime.Error.Data.Content -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

This statement is used inside `Runtime.Error.Wrap`. Its enclosing native signature and preceding bindings remain required.

<!-- example:glossary-runtime-error-data-caller -->

</details>

### Runtime.Handle.Content

<a id="word-authoring-runtime-handle-content"></a>

<!-- word:authoring:Runtime.Handle.Content -->

Construct a native `Runtime` value with `inner` set to `Inner`.

<!-- word-definition:authoring:Runtime.Handle.Content -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

This statement is used inside `Runtime.Copy`. Its enclosing native signature and preceding bindings remain required.

<!-- example:glossary-runtime-handle-caller -->

</details>

### Runtime.Inner.Content

<a id="word-authoring-runtime-inner-content"></a>

<!-- word:authoring:Runtime.Inner.Content -->

Construct a native `RuntimeInner` value with `companion_directory` set to `Directory`, `state` set to `State`, `cancelled` set to `Cancelled`, `shutdown_done` set to `Done`, `signals` set to `Signals`, `stdout` set to `Output`, `stderr` set to `Error`, `dependencies` set to `Dependencies`, `finish_gate` set to `Gate`.

<!-- word-definition:authoring:Runtime.Inner.Content -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

This statement is used inside `Runtime.Assemble`. Its enclosing native signature and preceding bindings remain required.

<!-- example:glossary-runtime-inner-content-caller -->

</details>

### Runtime.Invocation.Content

<a id="word-authoring-runtime-invocation-content"></a>

<!-- word:authoring:Runtime.Invocation.Content -->

Construct a native `Invocation` value with `app` set to `App`, `runtime` set to `Runtime`, `sequence` set to `Sequence`, `dependencies` set to `Dependencies`, `wait_delay` set to `Delay`, `state` set to `State`, `done` set to `Done`, `group` set to `Group`, `drain_deadline` set to `Deadline`, `leader_requests` set to `Requests`.

<!-- word-definition:authoring:Runtime.Invocation.Content -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

This statement is used inside `Runtime.Invocation.Allocate`. Its enclosing native signature and preceding bindings remain required.

<!-- example:glossary-runtime-invocation-content-caller -->

</details>

### Runtime.Invocation.Group.Content

<a id="word-authoring-runtime-invocation-group-content"></a>

<!-- word:authoring:Runtime.Invocation.Group.Content -->

Construct a native `GroupState` value with `pid` set to `Pid`, `cleaned` set to `false`, `termination_deadline` set to `None`.

<!-- word-definition:authoring:Runtime.Invocation.Group.Content -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

This statement is used inside `Runtime.Invocation.Allocate`. Its enclosing native signature and preceding bindings remain required.

<!-- example:glossary-runtime-invocation-group-caller -->

</details>

### Runtime.Invocation.State.Content

<a id="word-authoring-runtime-invocation-state-content"></a>

<!-- word:authoring:Runtime.Invocation.State.Content -->

Construct a native `InvocationState` value with `snapshot` set to `None`, `claim` set to `0`.

<!-- word-definition:authoring:Runtime.Invocation.State.Content -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

This statement is used inside `Runtime.Invocation.Allocate`. Its enclosing native signature and preceding bindings remain required.

<!-- example:glossary-runtime-invocation-state-caller -->

</details>

### Runtime.Snapshot.Content

<a id="word-authoring-runtime-snapshot-content"></a>

<!-- word:authoring:Runtime.Snapshot.Content -->

Construct a native `Snapshot` value with `stdout` set to `Output`, `stderr` set to `Error`, `error` set to `Fault`, `status` set to `Status`.

<!-- word-definition:authoring:Runtime.Snapshot.Content -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

This statement is used inside `Runtime.Snapshot.Reject`. Its enclosing native signature and preceding bindings remain required.

<!-- example:glossary-runtime-snapshot-content-caller -->

</details>

### Runtime.Terminal.Content

<a id="word-authoring-runtime-terminal-content"></a>

<!-- word:authoring:Runtime.Terminal.Content -->

Construct a native `Terminal` value with `status` set to `Status`, `error` set to `Error`.

<!-- word-definition:authoring:Runtime.Terminal.Content -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

This statement is used inside `Runtime.Conclude`. Its enclosing native signature and preceding bindings remain required.

<!-- example:glossary-runtime-terminal-content-caller -->

</details>

### Storage.Environment.Content

<a id="word-authoring-storage-environment-content"></a>

<!-- word:authoring:Storage.Environment.Content -->

Construct a native `LEnvironment` value with `types` set to `Types`, `values` set to `Values`.

<!-- word-definition:authoring:Storage.Environment.Content -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

This statement is used inside `Storage.Environment.Copy`. Its enclosing native signature and preceding bindings remain required.

<!-- example:glossary-storage-environment-content-caller -->

</details>

### Storage.List.Content

<a id="word-authoring-storage-list-content"></a>

<!-- word:authoring:Storage.List.Content -->

Construct a native `LList` value with `element` set to `Element`, `values` set to `Values`.

<!-- word-definition:authoring:Storage.List.Content -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

This statement is used inside `Storage.List.Copy`. Its enclosing native signature and preceding bindings remain required.

<!-- example:glossary-storage-list-content-caller -->

</details>

### Storage.Map.Content

<a id="word-authoring-storage-map-content"></a>

<!-- word:authoring:Storage.Map.Content -->

Construct a native `LMap` value with `element` set to `Element`, `values` set to `Values`.

<!-- word-definition:authoring:Storage.Map.Content -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

This statement is used inside `Storage.Map.Create`. Its enclosing native signature and preceding bindings remain required.

<!-- example:glossary-storage-map-content-caller -->

</details>

### Storage.Pointer.Content

<a id="word-authoring-storage-pointer-content"></a>

<!-- word:authoring:Storage.Pointer.Content -->

Construct a native `LPointer` value with `element` set to `Element`, `target` set to `Target`.

<!-- word-definition:authoring:Storage.Pointer.Content -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

This statement is used inside `Storage.Pointer.Copy`. Its enclosing native signature and preceding bindings remain required.

<!-- example:glossary-storage-pointer-content-caller -->

</details>

### Storage.Program.Content

<a id="word-authoring-storage-program-content"></a>

<!-- word:authoring:Storage.Program.Content -->

Construct a native `LProgram` value with `apps` set to `Apps`, `types` set to `Types`, `values` set to `Values`, `actions` set to `Actions`.

<!-- word-definition:authoring:Storage.Program.Content -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

This statement is used inside [Storage.Program.Create](#word-authoring-storage-program-create). Its enclosing native signature, preceding local bindings and selected native module are required.

<!-- example:glossary-builtins-eng-storage-program-50-50 -->

</details>

### Storage.Value.Content

<a id="word-authoring-storage-value-content"></a>

<!-- word:authoring:Storage.Value.Content -->

Construct a native `LValue` value with `ty` set to `Signature`, `data` set to `Data`, `show` set to `Renderer`, `reaches` set to `Visitor`.

<!-- word-definition:authoring:Storage.Value.Content -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

This statement is used inside `Storage.Value.Wrap`. Its enclosing native signature and preceding bindings remain required.

<!-- example:glossary-storage-value-content-caller -->

</details>

### Runtime.App.Context

<a id="word-authoring-runtime-app-context"></a>

<!-- word:authoring:Runtime.App.Context -->

Construct a native String using the exact format pattern and arguments specified by this definition. The selected operation is `format!("start application {}", Path)`.

<!-- word-definition:authoring:Runtime.App.Context -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

This statement is used inside `Runtime.App.Launch`. Its enclosing native signature and preceding bindings remain required.

<!-- example:glossary-runtime-app-context-caller -->

</details>

### Runtime.Context

<a id="word-authoring-runtime-context"></a>

<!-- word:authoring:Runtime.Context -->

Construct a native String using the exact format pattern and arguments specified by this definition. The selected operation is `format!("{}: {}", Source, Message)`.

<!-- word-definition:authoring:Runtime.Context -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

This statement is used inside `Runtime.Error.Explain`. Its enclosing native signature and preceding bindings remain required.

<!-- example:glossary-runtime-context-caller -->

</details>

### Runtime.Path.Context

<a id="word-authoring-runtime-path-context"></a>

<!-- word:authoring:Runtime.Path.Context -->

Construct a native String using the exact format pattern and arguments specified by this definition. The selected operation is `format!("application companion {}", Path)`.

<!-- word-definition:authoring:Runtime.Path.Context -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

This statement is used inside `Runtime.App.Locate`. Its enclosing native signature and preceding bindings remain required.

<!-- example:glossary-runtime-path-context-caller -->

</details>

### Storage.Context

<a id="word-authoring-storage-context"></a>

<!-- word:authoring:Storage.Context -->

Name the Storage namespace for pairing runtime and program storage and projecting their generated execution bindings.

<!-- word-definition:authoring:Storage.Context -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

This statement is used inside `builtins/eng/storage/execution.l`. Its enclosing native signature and preceding bindings remain required.

<!-- example:glossary-storage-context-caller -->

</details>

### Storage.Function.Continue

<a id="word-authoring-storage-function-continue"></a>

<!-- word:authoring:Storage.Function.Continue -->

Construct the Continue variant of native ControlFlow around the supplied values.

<!-- word-definition:authoring:Storage.Function.Continue -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

This statement is used inside `builtins/eng/storage/execution.l`. Its enclosing native signature and preceding bindings remain required.

<!-- example:glossary-storage-function-continue-caller -->

</details>

### Formatting.Contract

<a id="word-authoring-formatting-contract"></a>

<!-- word:authoring:Formatting.Contract -->

Insert a decimal point after the first significant digit when needed and append the adjusted signed scientific exponent.

<!-- word-definition:authoring:Formatting.Contract -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

This statement is used inside `Formatting.Render`. Its enclosing native signature and preceding bindings remain required.

<!-- example:glossary-formatting-contract-caller -->

</details>

### Runtime.Copy

<a id="word-authoring-runtime-copy"></a>

<!-- word:authoring:Runtime.Copy -->

Return another runtime handle retaining the same Arc-backed RuntimeInner allocation.

<!-- word-definition:authoring:Runtime.Copy -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

This statement is used inside `Runtime.Begin`. Its enclosing native signature and preceding bindings remain required.

<!-- example:glossary-runtime-copy-caller -->

</details>

Existing qualified aliases: `Runtime.Host.Copy`. They select this canonical meaning.

### Runtime.Error.Copy

<a id="word-authoring-runtime-error-copy"></a>

<!-- word:authoring:Runtime.Error.Copy -->

Retain the Arc-backed ErrorData allocation and return another Error handle.

<!-- word-definition:authoring:Runtime.Error.Copy -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

This statement is used inside `Runtime.Refuse`. Its enclosing native signature and preceding bindings remain required.

<!-- example:glossary-runtime-error-copy-caller -->

</details>

Existing qualified aliases: `Runtime.Host.Error.Copy`. They select this canonical meaning.

### Runtime.Snapshot.Copy

<a id="word-authoring-runtime-snapshot-copy"></a>

<!-- word:authoring:Runtime.Snapshot.Copy -->

Retain a Snapshot's shared stdout/stderr bytes, copy its optional Error, and preserve its exit status.

<!-- word-definition:authoring:Runtime.Snapshot.Copy -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>


The wrapper below imports `Storage` and `Runtime` and is a member of the example `GlossaryExamples` Dictionary. Its declared native inputs, outputs, lifetime/generic descriptors, and the original selected module's native types/support are required. Lyger resolves the source and composes its library AML; this fragment has no native compilation or execution claim.

<!-- example:glossary-native-runtime-snapshot-copy -->

</details>

Existing qualified aliases: `Runtime.Host.Snapshot.Copy`. They select this canonical meaning.

### Runtime.Terminal.Copy

<a id="word-authoring-runtime-terminal-copy"></a>

<!-- word:authoring:Runtime.Terminal.Copy -->

Copy a terminal decision's status and optional Error into another Terminal value.

<!-- word-definition:authoring:Runtime.Terminal.Copy -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>


The wrapper below imports `Storage` and `Runtime` and is a member of the example `GlossaryExamples` Dictionary. Its declared native inputs, outputs, lifetime/generic descriptors, and the original selected module's native types/support are required. Lyger resolves the source and composes its library AML; this fragment has no native compilation or execution claim.

<!-- example:glossary-native-runtime-terminal-copy -->

</details>

Existing qualified aliases: `Runtime.Host.Terminal.Copy`. They select this canonical meaning.

### Storage.Copy

<a id="word-authoring-storage-copy"></a>

<!-- word:authoring:Storage.Copy -->

Invoke the native Clone implementation on the supplied borrowed value. The selected operation is `std::clone::Clone::clone(Value)`.

<!-- word-definition:authoring:Storage.Copy -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

This statement is used inside `Storage.Clone`. Its enclosing native signature and preceding bindings remain required.

<!-- example:glossary-storage-copy-caller -->

</details>

### Storage.Environment.Copy

<a id="word-authoring-storage-environment-copy"></a>

<!-- word:authoring:Storage.Environment.Copy -->

Return another capture-environment handle sharing both the Rc-backed type signatures and the mutable captured-value storage.

<!-- word-definition:authoring:Storage.Environment.Copy -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

This statement is used inside `Storage.Environment.Require`. Its enclosing native signature and preceding bindings remain required.

<!-- example:glossary-storage-environment-copy-caller -->

</details>

### Storage.Function.Copy

<a id="word-authoring-storage-function-copy"></a>

<!-- word:authoring:Storage.Function.Copy -->

Copy a typed Function pair by retaining its capture environment and preserving its callable target. The Function signature remains concrete.

<!-- word-definition:authoring:Storage.Function.Copy -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

The wrapper below imports `Storage` and `Runtime` and is a member of the example `GlossaryExamples` Dictionary. Its declared native inputs, outputs, lifetime/generic descriptors, and the original selected module's native types/support are required. Lyger resolves the source and composes its library AML; this fragment has no native compilation or execution claim.

<!-- example:glossary-native-storage-function-copy -->

</details>

### Storage.List.Copy

<a id="word-authoring-storage-list-copy"></a>

<!-- word:authoring:Storage.List.Copy -->

Copy a List's signature and rebuild its vector by applying Value.Copy to each element in order.

<!-- word-definition:authoring:Storage.List.Copy -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

The wrapper below imports `Storage` and `Runtime` and is a member of the example `GlossaryExamples` Dictionary. Its declared native inputs, outputs, lifetime/generic descriptors, and the original selected module's native types/support are required. Lyger resolves the source and composes its library AML; this fragment has no native compilation or execution claim.

<!-- example:glossary-native-storage-list-copy -->

</details>

### Storage.Map.Copy

<a id="word-authoring-storage-map-copy"></a>

<!-- word:authoring:Storage.Map.Copy -->

Copy a Map's element signature and rebuild its ordered map by copying each key and applying Value.Copy to each value.

<!-- word-definition:authoring:Storage.Map.Copy -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

The wrapper below imports `Storage` and `Runtime` and is a member of the example `GlossaryExamples` Dictionary. Its declared native inputs, outputs, lifetime/generic descriptors, and the original selected module's native types/support are required. Lyger resolves the source and composes its library AML; this fragment has no native compilation or execution claim.

<!-- example:glossary-native-storage-map-copy -->

</details>

### Storage.Pointer.Copy

<a id="word-authoring-storage-pointer-copy"></a>

<!-- word:authoring:Storage.Pointer.Copy -->

Copy a Pointer's element signature while retaining the same Rc-backed target cell.

<!-- word-definition:authoring:Storage.Pointer.Copy -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

The wrapper below imports `Storage` and `Runtime` and is a member of the example `GlossaryExamples` Dictionary. Its declared native inputs, outputs, lifetime/generic descriptors, and the original selected module's native types/support are required. Lyger resolves the source and composes its library AML; this fragment has no native compilation or execution claim.

<!-- example:glossary-native-storage-pointer-copy -->

</details>

### Storage.Value.Copy

<a id="word-authoring-storage-value-copy"></a>

<!-- word:authoring:Storage.Value.Copy -->

Copy a stored value's signature, retain its Rc-backed payload, and preserve its renderer and ownership visitor Functions.

<!-- word-definition:authoring:Storage.Value.Copy -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

This statement is used inside `Storage.Pointer.Read`. Its enclosing native signature and preceding bindings remain required.

<!-- example:glossary-storage-value-copy-caller -->

</details>

### Storage.Copyable

<a id="word-authoring-storage-copyable"></a>

<!-- word:authoring:Storage.Copyable -->

Describe a Storage.Routine with the native generic bound `T: Copy`.

<!-- word-definition:authoring:Storage.Copyable -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

This statement is used inside `Storage.Function.Copy`. Its enclosing native signature and preceding bindings remain required.

<!-- example:glossary-storage-copyable-caller -->

</details>

### Runtime.Count

<a id="word-authoring-runtime-count"></a>

<!-- word:authoring:Runtime.Count -->

Increment the runtime's invocation-sequence counter under its state lock and return the new sequence number.

<!-- word-definition:authoring:Runtime.Count -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

This statement is used inside `Runtime.App.Acquire`. Its enclosing native signature and preceding bindings remain required.

<!-- example:glossary-runtime-count-caller -->

</details>

### Create

<a id="word-builtin-create"></a>

<!-- word:builtin:create -->

Introduce a named supported value, construction or source action in the current lexical scope.

<!-- word-definition:builtin:create -->

**Usage and example.** `Create "Hello" as a New Basic App` in the [first program](../guide/first-program.md#read-the-program) creates an App. `Create "Launch..." for ...` in the [action example](../guide/definitions.md#compose-a-source-action) creates a reusable composition. Its parameter has the declared App meaning.

### Storage.App.Create

<a id="word-authoring-storage-app-create"></a>

<!-- word:authoring:Storage.App.Create -->

Construct the PathBuf once and create the App handle from the explicit runtime and path operands.

<!-- word-definition:authoring:Storage.App.Create -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

**Compiler use.** Bind the compiler-selected App name and built companion executable path. This retained compiler excerpt selects the L-owned definition with typed native name/type/value or Process bindings. Those bindings come from the compiler’s resolved plan; this is an integration fragment, not an ordinary standalone L call.

<!-- example:glossary-lyger-execution-composition-85-90 -->

</details>

### Storage.Context.Create

<a id="word-authoring-storage-context-create"></a>

<!-- word:authoring:Storage.Context.Create -->

Clone the runtime handle, pair it with the supplied program storage, and expose the program through an explicitly named borrowed binding.

<!-- word-definition:authoring:Storage.Context.Create -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

**Compiler use.** Bind the Runtime reference and typed LProgram value into the callable context. This retained compiler excerpt selects the L-owned definition with typed native name/type/value or Process bindings. Those bindings come from the compiler’s resolved plan; this is an integration fragment, not an ordinary standalone L call.

<!-- example:glossary-lyger-storage-words-156-159 -->

</details>

### Storage.Environment.Create

<a id="word-authoring-storage-environment-create"></a>

<!-- word:authoring:Storage.Environment.Create -->

Validate equal numbers of capture signatures and values, check every value against its signature, then place the values in shared mutable storage and return the capture environment.

<!-- word-definition:authoring:Storage.Environment.Create -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

The wrapper below imports `Storage` and `Runtime` and is a member of the example `GlossaryExamples` Dictionary. Its declared native inputs, outputs, lifetime/generic descriptors, and the original selected module's native types/support are required. Lyger resolves the source and composes its library AML; this fragment has no native compilation or execution claim.

<!-- example:glossary-native-storage-environment-create -->

</details>

### Storage.Error.Create

<a id="word-authoring-storage-error-create"></a>

<!-- word:authoring:Storage.Error.Create -->

Construct a runtime Error from supplied owned message bytes through `runtime::error_new`.

<!-- word-definition:authoring:Storage.Error.Create -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

**Compiler use.** Construct a native runtime Error from the transferred source Error payload. This retained compiler excerpt selects the L-owned definition with typed native name/type/value or Process bindings. Those bindings come from the compiler’s resolved plan; this is an integration fragment, not an ordinary standalone L call.

<!-- example:glossary-lyger-execution-composition-337-343 -->

</details>

### Storage.List.Create

<a id="word-authoring-storage-list-create"></a>

<!-- word:authoring:Storage.List.Create -->

Check every initial value against the declared element signature, then construct an owned List from the supplied signature and values.

<!-- word-definition:authoring:Storage.List.Create -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

The wrapper below imports `Storage` and `Runtime` and is a member of the example `GlossaryExamples` Dictionary. Its declared native inputs, outputs, lifetime/generic descriptors, and the original selected module's native types/support are required. Lyger resolves the source and composes its library AML; this fragment has no native compilation or execution claim.

<!-- example:glossary-native-storage-list-create -->

</details>

### Storage.Map.Create

<a id="word-authoring-storage-map-create"></a>

<!-- word:authoring:Storage.Map.Create -->

Construct an empty ordered byte-keyed Map with the supplied complete element signature.

<!-- word-definition:authoring:Storage.Map.Create -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

The wrapper below imports `Storage` and `Runtime` and is a member of the example `GlossaryExamples` Dictionary. Its declared native inputs, outputs, lifetime/generic descriptors, and the original selected module's native types/support are required. Lyger resolves the source and composes its library AML; this fragment has no native compilation or execution claim.

<!-- example:glossary-native-storage-map-create -->

</details>

### Storage.Pointer.Create

<a id="word-authoring-storage-pointer-create"></a>

<!-- word:authoring:Storage.Pointer.Create -->

Check an initial value against the Pointer's element signature, then place it in a new Rc-backed mutable cell.

<!-- word-definition:authoring:Storage.Pointer.Create -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

The wrapper below imports `Storage` and `Runtime` and is a member of the example `GlossaryExamples` Dictionary. Its declared native inputs, outputs, lifetime/generic descriptors, and the original selected module's native types/support are required. Lyger resolves the source and composes its library AML; this fragment has no native compilation or execution claim.

<!-- example:glossary-native-storage-pointer-create -->

</details>

### Storage.Program.Create

<a id="word-authoring-storage-program-create"></a>

<!-- word:authoring:Storage.Program.Create -->

Construct program storage from App handles and value signatures, allocating initially empty value slots and the requested number of action slots.

<!-- word-definition:authoring:Storage.Program.Create -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

The wrapper below imports `Storage` and `Runtime` and is a member of the example `GlossaryExamples` Dictionary. Its declared native inputs, outputs, lifetime/generic descriptors, and the original selected module's native types/support are required. Lyger resolves the source and composes its library AML; this fragment has no native compilation or execution claim.

<!-- example:glossary-native-storage-program-create -->

</details>

### Runtime.App.Current

<a id="word-authoring-runtime-app-current"></a>

<!-- word:authoring:Runtime.App.Current -->

Obtain the selected value's native borrowed view, preserving any optional wrapper provided by that method. The selected operation is `Value.invocation.as_ref()`.

<!-- word-definition:authoring:Runtime.App.Current -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

This statement is used inside `Runtime.App.Resolve`. Its enclosing native signature and preceding bindings remain required.

<!-- example:glossary-runtime-app-current-caller -->

</details>

### Runtime.Current

<a id="word-authoring-runtime-current"></a>

<!-- word:authoring:Runtime.Current -->

Borrow the current value of the selected watch receiver. The selected operation is `Value.borrow()`.

<!-- word-definition:authoring:Runtime.Current -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

This statement is used inside `Runtime.Observe`. Its enclosing native signature and preceding bindings remain required.

<!-- example:glossary-runtime-current-caller -->

</details>

## D

<a id="glossary-d"></a>

### Runtime.Error.Data

<a id="word-authoring-runtime-error-data"></a>

<!-- word:authoring:Runtime.Error.Data -->

Declare the native `ErrorData` record with fields `message: Vec<u8>`, `causes: Vec<Error>`, `cancelled: bool`, `native: Option<std::sync::Arc<dyn std::error::Error + Send + Sync>>`.

<!-- word-definition:authoring:Runtime.Error.Data -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

This statement is used inside `Runtime.Error.Wrap`. Its enclosing native signature and preceding bindings remain required.

<!-- example:glossary-runtime-error-data-caller -->

</details>

### Storage.Value.Data

<a id="word-authoring-storage-value-data"></a>

<!-- word:authoring:Storage.Value.Data -->

Borrow `Value.data` without moving the referenced native field.

<!-- word-definition:authoring:Storage.Value.Data -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

This statement is used inside `Storage.Value.Restore`. Its enclosing native signature and preceding bindings remain required.

<!-- example:glossary-storage-value-data-caller -->

</details>

### Runtime.Error.Datum

<a id="word-authoring-runtime-error-datum"></a>

<!-- word:authoring:Runtime.Error.Datum -->

Borrow `Value.data` without moving the referenced native field.

<!-- word-definition:authoring:Runtime.Error.Datum -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

This statement is used inside `Runtime.Error.Copy`. Its enclosing native signature and preceding bindings remain required.

<!-- example:glossary-runtime-error-datum-caller -->

</details>

### Runtime.Invocation.Deadline

<a id="word-authoring-runtime-invocation-deadline"></a>

<!-- word:authoring:Runtime.Invocation.Deadline -->

Select the native field `Value.termination_deadline` by value. Its declared type determines whether this copies or moves the field.

<!-- word-definition:authoring:Runtime.Invocation.Deadline -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

This statement is used inside `Runtime.Invocation.Signal`. Its enclosing native signature and preceding bindings remain required.

<!-- example:glossary-runtime-invocation-deadline-caller -->

</details>

### Runtime.Decide

<a id="word-authoring-runtime-decide"></a>

<!-- word:authoring:Runtime.Decide -->

Read whether runtime state already contains a terminal selection.

<!-- word-definition:authoring:Runtime.Decide -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

This statement is used inside `Runtime.Run`. Its enclosing native signature and preceding bindings remain required.

<!-- example:glossary-runtime-decide-caller -->

</details>

### Storage.Decide

<a id="word-authoring-storage-decide"></a>

<!-- word:authoring:Storage.Decide -->

Propagate a failed Boolean Result, execute the supplied Process for a successful true value, and perform no continuation work for false.

<!-- word-definition:authoring:Storage.Decide -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

This statement is used inside `builtins/eng/storage/traversal.l`. Its enclosing native signature and preceding bindings remain required.

<!-- example:glossary-storage-decide-caller -->

</details>

### Storage.Declaration

<a id="word-authoring-storage-declaration"></a>

<!-- word:authoring:Storage.Declaration -->

Declare an uninitialized native local with the supplied name and type. Any later read must satisfy the caller’s definite-assignment rules.

<!-- word-definition:authoring:Storage.Declaration -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

**Compiler use.** Declare the selected native storage type without inventing an initialized L value. This retained compiler excerpt selects the L-owned definition with typed native name/type/value or Process bindings. Those bindings come from the compiler’s resolved plan; this is an integration fragment, not an ordinary standalone L call.

<!-- example:glossary-lyger-execution-composition-207-226 -->

</details>

### Runtime.Supervision.Decorate

<a id="word-authoring-runtime-supervision-decorate"></a>

<!-- word:authoring:Runtime.Supervision.Decorate -->

Preserve absence of an Error, or attach the supplied stderr bytes to a present invocation Error.

<!-- word-definition:authoring:Runtime.Supervision.Decorate -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

This statement is used inside `Runtime.Supervision.Complete`. Its enclosing native signature and preceding bindings remain required.

<!-- example:glossary-runtime-supervision-decorate-caller -->

</details>

### Define

<a id="word-builtin-define"></a>

<!-- word:builtin:define -->

Associate a declaration with a value or an explicitly supplied native-backed definition.

<!-- word-definition:builtin:define -->

**Usage and example.** The [public-values program](../../demo/paragraphs/public-values.l) defines Text, Number and Truth. [Framed definitions](../spec/definitions.md#framing-and-escaping) bind native source through typed slots. Define does not bypass ownership, type or visibility checks.

### Definition

<a id="word-type-definition"></a>

<!-- word:type:definition -->

Name the implementation attached to a constructed element or lexical word.

<!-- word-definition:type:definition -->

**Context and illustration.** The [native definition example](../spec/definitions.md#framing-and-escaping) associates framed source with a word. [Descriptors](../spec/lexical-types.md#describe-with-an-adjective) compose the actual selected definitions.

**Partial support.** Framed definition templates can be parsed, bound and composed, and constructed Functions carry their Definition in the [Object representation](../../lyger/source_hierarchy.go). The recognized Definition category has no standalone runtime-value constructor: using Definition as a source Function parameter or List element [rejects](../../lyger/runtime_types.go) with `Definition is a source structure, not a runtime type`. Attach a definition to its owning word through Define instead.

**Usage illustration.** The [Objective source composition](#word-builtin-objective) takes an existing Function definition and groups it into Object content. The native-definition example linked above shows how Define attaches such implementation content to its owning word. Neither example constructs a standalone runtime Definition value.

### Runtime.Invocation.Delay

<a id="word-authoring-runtime-invocation-delay"></a>

<!-- word:authoring:Runtime.Invocation.Delay -->

Select the native field `Value.wait_delay` by value. Its declared type determines whether this copies or moves the field.

<!-- word-definition:authoring:Runtime.Invocation.Delay -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

This statement is used inside `Runtime.Invocation.Signal`. Its enclosing native signature and preceding bindings remain required.

<!-- example:glossary-runtime-invocation-delay-caller -->

</details>

### Delegate

<a id="word-builtin-delegate"></a>

<!-- word:builtin:delegate -->

Call an explicitly selected native target without supplying an argument list.

<!-- word-definition:builtin:delegate -->

**Usage and example.** The [recursive Class fixture](../../lyger/test/classes/recursive.l) places Main below nested modules. Its generated entry uses Delegate with a compiler-resolved native path to that Main Function. The [source topology implementation](../../lyger/source_units_topology.go) supplies this reference as the entry's sole tail call. A quoted Text target is not this native reference, and no general Text-to-call API is promised.

Delegate and [Function.Invoke](#word-builtin-function-invoke) preserve the selected target and lexical ownership. Native type and symbol checking remain separate from grammatical recognition.

### Runtime.Dependencies

<a id="word-authoring-runtime-dependencies"></a>

<!-- word:authoring:Runtime.Dependencies -->

Move the supplied value into a new Arc allocation. The selected operation is `std::sync::Arc::new(Value)`.

<!-- word-definition:authoring:Runtime.Dependencies -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

This statement is used inside `Runtime.Open`. Its enclosing native signature and preceding bindings remain required.

<!-- example:glossary-runtime-dependencies-caller -->

</details>

### Runtime.Invocation.Dependencies

<a id="word-authoring-runtime-invocation-dependencies"></a>

<!-- word:authoring:Runtime.Invocation.Dependencies -->

Borrow `Value.dependencies` without moving the referenced native field.

<!-- word-definition:authoring:Runtime.Invocation.Dependencies -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

This statement is used inside `Runtime.Invocation.Observe`. Its enclosing native signature and preceding bindings remain required.

<!-- example:glossary-runtime-invocation-dependencies-caller -->

</details>

### Storage.Dereference

<a id="word-authoring-storage-dereference"></a>

<!-- word:authoring:Storage.Dereference -->

Select the value behind a native reference or dereferenceable value; its concrete type determines whether the result copies or moves.

<!-- word-definition:authoring:Storage.Dereference -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

This statement is used inside `Runtime.Limit`. Its enclosing native signature and preceding bindings remain required.

<!-- example:glossary-storage-dereference-caller -->

</details>

### Describe

<a id="word-builtin-describe"></a>

<!-- word:builtin:describe -->

Associate a descriptor with its canonical target and ordered construction refinements.

<!-- word-definition:builtin:describe -->

**Usage and example.** Main's `It describes` rules in the [descriptor example](../spec/lexical-types.md#describe-with-an-adjective) construct different Function/Object/Service targets. The [recursive caller](../../lyger/test/classes/recursive.l) applies the selected Object rule. Description does not make unrelated homonyms equivalent.

### Runtime.App.Describe

<a id="word-authoring-runtime-app-describe"></a>

<!-- word:authoring:Runtime.App.Describe -->

Read an App's diagnostic source label and prefix it to a message when nonempty, then construct a runtime Error.

<!-- word-definition:authoring:Runtime.App.Describe -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

This statement is used inside `Runtime.App.Explain`. Its enclosing native signature and preceding bindings remain required.

<!-- example:glossary-runtime-app-describe-caller -->

</details>

### Runtime.Error.Describe

<a id="word-authoring-runtime-error-describe"></a>

<!-- word:authoring:Runtime.Error.Describe -->

Copy a borrowed message string into bytes and construct a runtime Error with no causes, native cause, or cancellation flag.

<!-- word-definition:authoring:Runtime.Error.Describe -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

This statement is used inside `Runtime.App.Own`. Its enclosing native signature and preceding bindings remain required.

<!-- example:glossary-runtime-error-describe-caller -->

</details>

### Storage.List.Describe

<a id="word-authoring-storage-list-describe"></a>

<!-- word:authoring:Storage.List.Describe -->

Render one stored value, propagate a rendering failure, and append its byte vector to the supplied display-parts buffer.

<!-- word-definition:authoring:Storage.List.Describe -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

This statement is used inside `Storage.List.Render`. Its enclosing native signature and preceding bindings remain required.

<!-- example:glossary-storage-list-describe-caller -->

</details>

### Storage.Map.Describe

<a id="word-authoring-storage-map-describe"></a>

<!-- word:authoring:Storage.Map.Describe -->

Render one Map value, prefix it with its unchanged byte key and `: `, then append the result to the display-parts buffer.

<!-- word-definition:authoring:Storage.Map.Describe -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

This statement is used inside `Storage.Map.Render`. Its enclosing native signature and preceding bindings remain required.

<!-- example:glossary-storage-audit-caller -->

</details>

### Runtime.Capture.Description

<a id="word-authoring-runtime-capture-description"></a>

<!-- word:authoring:Runtime.Capture.Description -->

Construct a native String using the exact format pattern and arguments specified by this definition. The selected operation is `format!("capture application {} task: {}", Stream, Error)`.

<!-- word-definition:authoring:Runtime.Capture.Description -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

This statement is used inside `Runtime.Capture.Join`. Its enclosing native signature and preceding bindings remain required.

<!-- example:glossary-runtime-capture-description-caller -->

</details>

### Runtime.Path.Description

<a id="word-authoring-runtime-path-description"></a>

<!-- word:authoring:Runtime.Path.Description -->

Construct a native String using the exact format pattern and arguments specified by this definition. The selected operation is `format!("invalid application companion path {:?}", Path)`.

<!-- word-definition:authoring:Runtime.Path.Description -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

This statement is used inside `Runtime.App.Locate`. Its enclosing native signature and preceding bindings remain required.

<!-- example:glossary-runtime-path-description-caller -->

</details>

### Runtime.Supervision.Description

<a id="word-authoring-runtime-supervision-description"></a>

<!-- word:authoring:Runtime.Supervision.Description -->

Construct a native String using the exact format pattern and arguments specified by this definition. The selected operation is `format!("application {}", Status)`.

<!-- word-definition:authoring:Runtime.Supervision.Description -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

This statement is used inside `Runtime.Supervision.Inspect`. Its enclosing native signature and preceding bindings remain required.

<!-- example:glossary-runtime-supervision-description-caller -->

</details>

### Storage.Description

<a id="word-authoring-storage-description"></a>

<!-- word:authoring:Storage.Description -->

Format the supplied receiver into an owned native String through its Display implementation. The selected operation is `Value.to_string()`.

<!-- word-definition:authoring:Storage.Description -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

This statement is used inside `Formatting.Render`. Its enclosing native signature and preceding bindings remain required.

<!-- example:glossary-storage-description-caller -->

</details>

### Runtime.Capture.Destination

<a id="word-authoring-runtime-capture-destination"></a>

<!-- word:authoring:Runtime.Capture.Destination -->

Obtain the selected host capture destination for a stream and supplied byte-buffer writer. The selected operation is `Dependencies.capture(Stream, Bytes)`.

<!-- word-definition:authoring:Runtime.Capture.Destination -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

This statement is used inside `Runtime.Capture.Start`. Its enclosing native signature and preceding bindings remain required.

<!-- example:glossary-runtime-capture-box-caller -->

</details>

### Runtime.Error.Destination

<a id="word-authoring-runtime-error-destination"></a>

<!-- word:authoring:Runtime.Error.Destination -->

Mutably reborrow the diagnostic byte buffer in the first traversal-context component.

<!-- word-definition:authoring:Runtime.Error.Destination -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

This statement is used inside `Runtime.Error.Accumulate`. Its enclosing native signature and preceding bindings remain required.

<!-- example:glossary-runtime-error-destination-caller -->

</details>

### Runtime.Diagnose

<a id="word-authoring-runtime-diagnose"></a>

<!-- word:authoring:Runtime.Diagnose -->

Join optional primary and cleanup Errors, render their diagnostic bytes, append one LF, and return absence when neither Error exists.

<!-- word-definition:authoring:Runtime.Diagnose -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

This statement is used inside `Runtime.Finish`. Its enclosing native signature and preceding bindings remain required.

<!-- example:glossary-runtime-diagnose-caller -->

</details>

### IO.Diagnostic

<a id="word-stdlib-export-eng-io-io-diagnostic"></a>

<!-- word:stdlib-export:eng:io:IO.Diagnostic -->

Obtain a handle to the current process's standard-error stream without writing or locking it.

<!-- word-definition:stdlib-export:eng:io:IO.Diagnostic -->

**Usage:** In a native-signature Function, `a New IO's Diagnostic` produces `std::io::Stderr`. The caller decides whether to lock or borrow that handle and which operation to perform. Obtaining it does not create an App error or inspect captured stderr.

**Example:** The `Writing.Diagnostic` Function in [Native callers](#context-io--native-callers) returns the handle to its host. Compose a writer with [Send](#word-stdlib-export-eng-io-io-send) or [Flush](#word-stdlib-export-eng-io-io-flush). Use [Output](#word-stdlib-export-eng-io-io-output) for stdout instead.

### Dictionary

<a id="word-type-dictionary"></a>

<!-- word:type:dictionary -->

Own a namespace of lexical declarations and their canonical member identities.

<!-- word-definition:type:dictionary -->

**Usage and example.** The [public-values program](../../demo/paragraphs/public-values.l) declares Shared and its Display action. The [file-import example](../spec/imports.md#explicit-source-file-imports) imports a project-owned member without changing its owner.

A Dictionary owns named vocabulary members. `Create "Shared" as a New Public Dictionary` declares the owner, and `Shared's "Display|-s|-ed|-ing|-er"` declares a member word. Aliases preserve the owner and member identities. See the [public-values example](../guide/values.md#text-number-and-truth-together).

An imported dictionary has its own source scope and exports. [Indexed imports](#word-builtin-use) select those exports. Computed Map keys do not participate in lexical lookup.

### Runtime.Directory

<a id="word-authoring-runtime-directory"></a>

<!-- word:authoring:Runtime.Directory -->

Borrow `Value.companion_directory` without moving the referenced native field.

<!-- word-definition:authoring:Runtime.Directory -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

This statement is used inside `Runtime.App.Locate`. Its enclosing native signature and preceding bindings remain required.

<!-- example:glossary-runtime-directory-caller -->

</details>

### Runtime.Supervision.Discard

<a id="word-authoring-runtime-supervision-discard"></a>

<!-- word:authoring:Runtime.Supervision.Discard -->

Send the supplied value through the selected native channel. The selected operation is `Reply.send(Result)`.

<!-- word-definition:authoring:Runtime.Supervision.Discard -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

This statement is used inside `Runtime.Supervision.Release`. Its enclosing native signature and preceding bindings remain required.

<!-- example:glossary-runtime-supervision-discard-caller -->

</details>

### Storage.Discard

<a id="word-authoring-storage-discard"></a>

<!-- word:authoring:Storage.Discard -->

Evaluate the supplied value expression and bind it to the native discard pattern `_`.

<!-- word-definition:authoring:Storage.Discard -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

This statement is used inside `stdlib/eng/io/dictionary.l`. Its enclosing native signature and preceding bindings remain required.

<!-- example:glossary-storage-discard-caller -->

</details>

### Runtime.Error.Display

<a id="word-authoring-runtime-error-display"></a>

<!-- word:authoring:Runtime.Error.Display -->

Render an Error, convert its bytes to a lossy UTF-8 view, and write that text through the supplied native formatter.

<!-- word-definition:authoring:Runtime.Error.Display -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>


The wrapper below imports `Storage` and `Runtime` and is a member of the example `GlossaryExamples` Dictionary. Its declared native inputs, outputs, lifetime/generic descriptors, and the original selected module's native types/support are required. Lyger resolves the source and composes its library AML; this fragment has no native compilation or execution claim.

<!-- example:glossary-native-runtime-error-display -->

</details>

Existing qualified aliases: `Runtime.Host.Error.Display`. They select this canonical meaning.

### Runtime.Invocation.Done

<a id="word-authoring-runtime-invocation-done"></a>

<!-- word:authoring:Runtime.Invocation.Done -->

Borrow `Value.done` without moving the referenced native field.

<!-- word-definition:authoring:Runtime.Invocation.Done -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

This statement is used inside `Runtime.Invocation.Publish`. Its enclosing native signature and preceding bindings remain required.

<!-- example:glossary-runtime-invocation-done-caller -->

</details>

### Runtime.Duration

<a id="word-authoring-runtime-duration"></a>

<!-- word:authoring:Runtime.Duration -->

Construct a native duration from the supplied whole-second count. The selected operation is `tokio::time::Duration::from_secs(Seconds)`.

<!-- word-definition:authoring:Runtime.Duration -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

This statement is used inside `Runtime.App.Acquire`. Its enclosing native signature and preceding bindings remain required.

<!-- example:glossary-runtime-duration-caller -->

</details>

## E

<a id="glossary-e"></a>

### Storage.Editable

<a id="word-authoring-storage-editable"></a>

<!-- word:authoring:Storage.Editable -->

Select a mutable borrow for Storage.Borrow, a fallible mutable RefCell borrow for Storage.Read, or a mutable dereference/reborrow for Storage.Reference. Each refinement retains its exact canonical target.

<!-- word-definition:authoring:Storage.Editable -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

This statement is used inside `Runtime.Error.Render`. Its enclosing native signature and preceding bindings remain required.

<!-- example:glossary-storage-editable-caller -->

</details>

### Storage.List.Editable

<a id="word-authoring-storage-list-editable"></a>

<!-- word:authoring:Storage.List.Editable -->

Replace the canonical List.Values projection with a mutable borrow of the same value's `values` field. Map.Editable is an alias of this same modifier.

<!-- word-definition:authoring:Storage.List.Editable -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

This statement is used inside `Storage.Map.Update`. Its enclosing native signature and preceding bindings remain required.

<!-- example:glossary-storage-list-editable-caller -->

</details>

Existing qualified aliases: `Storage.Map.Editable`. They select this canonical meaning.

### Storage.Pointer.Element

<a id="word-authoring-storage-pointer-element"></a>

<!-- word:authoring:Storage.Pointer.Element -->

Borrow `Value.element` without moving the referenced native field.

<!-- word-definition:authoring:Storage.Pointer.Element -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

This statement is used inside `Storage.Map.Update`. Its enclosing native signature and preceding bindings remain required.

<!-- example:glossary-storage-pointer-element-caller -->

</details>

Existing qualified aliases: `Storage.List.Element`, `Storage.Map.Element`. They select this canonical meaning.

### Else

<a id="word-grammar-word-else"></a>

<!-- word:grammar-word:else -->

Attach an alternative branch to its unmatched If in the supported conditional statement.

<!-- word-definition:grammar-word:else -->

**Usage and boundary.** The [Truth conditional program](../../testdata/phase1/language/lang-13.l) puts `else exit` after its Await branch. Else is not an independent executable action or a fallback for arbitrary previous failures.

**Partial support.** A plain alternative branch works, but the configured `else if` continuation does not: a source chain such as `If false, Show 1, else if true, Show 2, else Show 3` is rejected during grammar processing. It does not produce an executable conditional chain.

### Runtime.Empty

<a id="word-authoring-runtime-empty"></a>

<!-- word:authoring:Runtime.Empty -->

Construct an empty native String. The selected operation is `String::new()`.

<!-- word-definition:authoring:Runtime.Empty -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

This statement is used inside `Runtime.Appoint`. Its enclosing native signature and preceding bindings remain required.

<!-- example:glossary-runtime-app-state-caller -->

</details>

### Runtime.Snapshot.Empty

<a id="word-authoring-runtime-snapshot-empty"></a>

<!-- word:authoring:Runtime.Snapshot.Empty -->

Construct an empty shared Arc-backed byte slice with `std::sync::Arc::from([])`.

<!-- word-definition:authoring:Runtime.Snapshot.Empty -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

This statement is used inside `Runtime.Snapshot.Reject`. Its enclosing native signature and preceding bindings remain required.

<!-- example:glossary-runtime-snapshot-caller -->

</details>

### Storage.Empty

<a id="word-authoring-storage-empty"></a>

<!-- word:authoring:Storage.Empty -->

Select an empty Process, an initially None mutable optional Binding, an empty typed Vector, or the unit Tuple for the corresponding canonical target.

<!-- word-definition:authoring:Storage.Empty -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

This statement is used inside `Runtime.Error.Join`. Its enclosing native signature and preceding bindings remain required.

<!-- example:glossary-storage-empty-caller -->

</details>

### Storage.Text.Enclose

<a id="word-authoring-storage-text-enclose"></a>

<!-- word:authoring:Storage.Text.Enclose -->

Consume a byte vector and return opening delimiter bytes, the unchanged vector bytes, and closing delimiter bytes in one result buffer. This preserves arbitrary bytes but can copy and reallocate.

<!-- word-definition:authoring:Storage.Text.Enclose -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

This statement is used inside `Storage.List.Render`. Its enclosing native signature and preceding bindings remain required.

<!-- example:glossary-storage-text-enclose-caller -->

</details>

### Storage.Encoding

<a id="word-authoring-storage-encoding"></a>

<!-- word:authoring:Storage.Encoding -->

Consume the selected String and return its owned UTF-8 byte vector. The selected operation is `Value.into_bytes()`.

<!-- word-definition:authoring:Storage.Encoding -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

This statement is used inside `Formatting.Finish`. Its enclosing native signature and preceding bindings remain required.

<!-- example:glossary-storage-encoding-caller -->

</details>

### Runtime.Checkpoint.Enter

<a id="word-authoring-runtime-checkpoint-enter"></a>

<!-- word:authoring:Runtime.Checkpoint.Enter -->

Call the selected host checkpoint observer with the supplied event and invocation. The selected operation is `Dependencies.checkpoint(Event, Invocation)`.

<!-- word-definition:authoring:Runtime.Checkpoint.Enter -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

This statement is used inside `Runtime.Invocation.Observe`. Its enclosing native signature and preceding bindings remain required.

<!-- example:glossary-runtime-checkpoint-enter-caller -->

</details>

### Runtime.Enter

<a id="word-authoring-runtime-enter"></a>

<!-- word:authoring:Runtime.Enter -->

Notify the selected host dependencies that the supplied terminal request has been entered. The selected operation is `Dependencies.terminal_entered(Request)`.

<!-- word-definition:authoring:Runtime.Enter -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

This statement is used inside `Runtime.Signals.Watch`. Its enclosing native signature and preceding bindings remain required.

<!-- example:glossary-runtime-enter-caller -->

</details>

### Storage.Map.Entries

<a id="word-authoring-storage-map-entries"></a>

<!-- word:authoring:Storage.Map.Entries -->

Borrow the selected collection and obtain its native iterator. The selected operation is `Value.values.iter()`.

<!-- word-definition:authoring:Storage.Map.Entries -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

This statement is used inside `Storage.Map.Render`. Its enclosing native signature and preceding bindings remain required.

<!-- example:glossary-storage-map-entries-caller -->

</details>

### Runtime.Entry

<a id="word-authoring-runtime-entry"></a>

<!-- word:authoring:Runtime.Entry -->

Look up the supplied key through a native map entry handle. The selected operation is `Values.entry(Key)`.

<!-- word-definition:authoring:Runtime.Entry -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

This statement is used inside `Runtime.Invocation.Record`. Its enclosing native signature and preceding bindings remain required.

<!-- example:glossary-runtime-entry-caller -->

</details>

### Storage.Entry

<a id="word-authoring-storage-entry"></a>

<!-- word:authoring:Storage.Entry -->

Name the Storage namespace that composes generated main execution and returns the orchestrator status.

<!-- word-definition:authoring:Storage.Entry -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

This statement is used inside `builtins/eng/storage/execution.l`. Its enclosing native signature and preceding bindings remain required.

<!-- example:glossary-storage-entry-caller -->

</details>

### Storage.Enumerate

<a id="word-authoring-storage-enumerate"></a>

<!-- word:authoring:Storage.Enumerate -->

Visit pair-shaped items in iterator order and invoke the supplied callback with both items and context, discarding each returned result, including an error result.

<!-- word-definition:authoring:Storage.Enumerate -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

This statement is used inside `Storage.Map.Copy`. Its enclosing native signature and preceding bindings remain required.

<!-- example:glossary-storage-enumerate-caller -->

</details>

### Storage.Environment

<a id="word-authoring-storage-environment"></a>

<!-- word:authoring:Storage.Environment -->

Declare the native `LEnvironment` record with fields `values: std::rc::Rc<std::cell::RefCell<Vec<Value>>>`, `types: std::rc::Rc<Vec<Vec<u64>>>`.

<!-- word-definition:authoring:Storage.Environment -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

This statement is used inside `Storage.Environment.Copy`. Its enclosing native signature and preceding bindings remain required.

<!-- example:glossary-storage-environment-caller -->

</details>

### Storage.Function.Environment

<a id="word-authoring-storage-function-environment"></a>

<!-- word:authoring:Storage.Function.Environment -->

Borrow `Value.0` without moving the referenced native field.

<!-- word-definition:authoring:Storage.Function.Environment -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

This statement is used inside `Storage.Function.Copy`. Its enclosing native signature and preceding bindings remain required.

<!-- example:glossary-storage-function-environment-caller -->

</details>

### Storage.Layout.Environment

<a id="word-authoring-storage-layout-environment"></a>

<!-- word:authoring:Storage.Layout.Environment -->

Declare the native `LEnvironment` record with fields `values: std::rc::Rc<std::cell::RefCell<Vec<Value>>>`, `types: std::rc::Rc<Vec<Vec<u64>>>`.

<!-- word-definition:authoring:Storage.Layout.Environment -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

This statement is used inside `builtins/eng/storage/values.l`. Its enclosing native signature and preceding bindings remain required.

<!-- example:glossary-storage-layout-environment-caller -->

</details>

### Equal

<a id="word-grammar-word-equal"></a>

<!-- word:grammar-word:equal -->

Compare two supported scalar values of the same type for equality.

<!-- word-definition:grammar-word:equal -->

**Usage and boundary.** The [Truth condition](../../testdata/phase1/language/lang-13.l) uses `enabled equals true`. Text compares bytes and Number compares numeric values. Mixed types do not coerce.

### Storage.Equality

<a id="word-authoring-storage-equality"></a>

<!-- word:authoring:Storage.Equality -->

Compare the supplied native operands `Left, Right` for equality and bind the Boolean result.

<!-- word-definition:authoring:Storage.Equality -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

This statement is used inside `Runtime.Admit`. Its enclosing native signature and preceding bindings remain required.

<!-- example:glossary-storage-equality-caller -->

</details>

### App.Error

<a id="word-builtin-error"></a>

<!-- word:builtin:error -->

Select an App's current operation or invocation failure for inspection.

<!-- word-definition:builtin:error -->

**Usage and example.** `it has an Error` in the [conditional member example](../../lyger/test/member-resolution.l) checks the declared HelloWorldApp after Execute. Extracting an absent Error rejects. This member is distinct from the [Error value type](#word-type-error).

Build/Stop can select operation errors, while Await reselects its immutable invocation result. `target has an Error` checks matching owner identity. See [error selection](../cli-build.md#invocation-and-results).

### Error

<a id="word-type-error"></a>

<!-- word:type:error -->

Carry a runtime failure value that can be stored, passed and retained in typed containers.

<!-- word-definition:type:error -->

**Usage and example.** The [complete value program](../examples/word-values.l) uses ``Define "problem" as an Error of `failure`.`` and puts it in a List of Error. Showing that list writes `[failure]`. Direct Show of an Error is unsupported. This is a value constructor, distinct from inspecting an [App's Error](#word-builtin-error).

The value retains its failure information when transferred through Functions. Extracting an absent App Error rejects instead of manufacturing this carrier. Error values do not provide general source exceptions or automatic rollback.

### Runtime.Capture.Error

<a id="word-authoring-runtime-capture-error"></a>

<!-- word:authoring:Runtime.Capture.Error -->

Select the native field `Value.error` by value. Its declared type determines whether this copies or moves the field.

<!-- word-definition:authoring:Runtime.Capture.Error -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

This statement is used inside `Runtime.Capture.Share`. Its enclosing native signature and preceding bindings remain required.

<!-- example:glossary-runtime-capture-error-caller -->

</details>

### Runtime.Error

<a id="word-authoring-runtime-error"></a>

<!-- word:authoring:Runtime.Error -->

Declare the native `Error` record with fields `data: std::sync::Arc<ErrorData>`.

<!-- word-definition:authoring:Runtime.Error -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

This statement is used inside `Runtime.Leave`. Its enclosing native signature and preceding bindings remain required.

<!-- example:glossary-runtime-error-caller -->

</details>

### Runtime.Host.Error

<a id="word-authoring-runtime-host-error"></a>

<!-- word:authoring:Runtime.Host.Error -->

Name the adapter namespace reusing canonical Runtime.Error copying, display, native-cause tracing and IO-error conversion.

<!-- word-definition:authoring:Runtime.Host.Error -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

After importing Runtime, bind the existing namespace for qualified member access.

<!-- example:glossary-namespace-runtime-host-error -->

</details>

### Runtime.Snapshot.Error

<a id="word-authoring-runtime-snapshot-error"></a>

<!-- word:authoring:Runtime.Snapshot.Error -->

Borrow `Value.stderr` without moving the referenced native field.

<!-- word-definition:authoring:Runtime.Snapshot.Error -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

This statement is used inside `Runtime.Snapshot.Copy`. Its enclosing native signature and preceding bindings remain required.

<!-- example:glossary-runtime-snapshot-error-caller -->

</details>

### Runtime.Supervision.Error

<a id="word-authoring-runtime-supervision-error"></a>

<!-- word:authoring:Runtime.Supervision.Error -->

Bind the native constant `crate::runtime::host::Stream::Stderr` for this composition.

<!-- word-definition:authoring:Runtime.Supervision.Error -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

This statement is used inside `Runtime.Invocation.Supervise`. Its enclosing native signature and preceding bindings remain required.

<!-- example:glossary-runtime-capture-follow-caller -->

</details>

### Runtime.Terminal.Error

<a id="word-authoring-runtime-terminal-error"></a>

<!-- word:authoring:Runtime.Terminal.Error -->

Borrow `Value.error` without moving the referenced native field.

<!-- word-definition:authoring:Runtime.Terminal.Error -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

This statement is used inside `Runtime.Finish`. Its enclosing native signature and preceding bindings remain required.

<!-- example:glossary-runtime-terminal-error-caller -->

</details>

### Storage.App.Error

<a id="word-authoring-storage-app-error"></a>

<!-- word:authoring:Storage.App.Error -->

Borrow an App handle and call the runtime operation that reads its selected invocation or operation Error.

<!-- word-definition:authoring:Storage.App.Error -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

**Compiler use.** Read the selected App member through its typed App reference; Error is optional. This retained compiler excerpt selects the L-owned definition with typed native name/type/value or Process bindings. Those bindings come from the compiler’s resolved plan; this is an integration fragment, not an ordinary standalone L call.

<!-- example:glossary-lyger-execution-composition-318-335 -->

</details>

### Storage.Error

<a id="word-authoring-storage-error"></a>

<!-- word:authoring:Storage.Error -->

Name the Storage namespace for runtime Error construction and erased-value display. It is distinct from the runtime Error value type.

<!-- word-definition:authoring:Storage.Error -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

This statement is used inside `Storage.Error.Render`. Its enclosing native signature and preceding bindings remain required.

<!-- example:glossary-storage-error-caller -->

</details>

### Runtime.State.Errors

<a id="word-authoring-runtime-state-errors"></a>

<!-- word:authoring:Runtime.State.Errors -->

Obtain a borrowed iterator over the selected ordered map's values. The selected operation is `Value.cleanup_errors.values()`.

<!-- word-definition:authoring:Runtime.State.Errors -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

This statement is used inside `Runtime.Assess`. Its enclosing native signature and preceding bindings remain required.

<!-- example:glossary-runtime-state-errors-caller -->

</details>

### Storage.Examine

<a id="word-authoring-storage-examine"></a>

<!-- word:authoring:Storage.Examine -->

Invoke a callback with one item and context, then discard a successful payload or propagate its failure from the enclosing Function.

<!-- word-definition:authoring:Storage.Examine -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

This statement is used inside `builtins/eng/storage/collections.l`. Its enclosing native signature and preceding bindings remain required.

<!-- example:glossary-storage-examine-caller -->

</details>

### Runtime.Executable

<a id="word-authoring-runtime-executable"></a>

<!-- word:authoring:Runtime.Executable -->

Read the native path of the running orchestrator executable. The selected operation is `std::env::current_exe()`.

<!-- word-definition:authoring:Runtime.Executable -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

This statement is used inside `Runtime.Locate`. Its enclosing native signature and preceding bindings remain required.

<!-- example:glossary-runtime-executable-caller -->

</details>

### Execute

<a id="word-builtin-execute"></a>

<!-- word:builtin:execute -->

Start an App and await its finalized result and first continuation before proceeding.

<!-- word-definition:builtin:execute -->

**Usage and example.** `Execute Hello` in the [first program](../guide/first-program.md#read-the-program) supplies the completion boundary before `Show Hello's Output`. Its source parameter remains `any App`, and no explicit Runtime import is needed.

A source-defined action with coordinated Apps performs its complete body for the first App, then the next. Thus coordinating Execute has a different overlap pattern from coordinating Run followed by coordinating Await. That ordering follows source expansion, not parallel implicit execution.

### Exit

<a id="word-builtin-exit"></a>

<!-- word:builtin:exit -->

Terminate the whole generated L program, cleaning up its owned processes.

<!-- word-definition:builtin:exit -->

**Usage and example.** The [bare Exit example](../spec/control-flow.md#exit-and-return) ends execution inside a continuation. Later actions do not run. [Function.Return](#word-builtin-function-return) instead completes one invocation.

Bare Exit succeeds if cleanup succeeds. Exit with a supported explicit failure/Error operand selects failure, including inside a continuation or source action. Cleanup does not activate unawaited continuations. The [shutdown contract](../cli-build.md#shutdown-streams-and-limits) specifies status codes, signal handling and cleanup failures.

### Runtime.Exit

<a id="word-authoring-runtime-exit"></a>

<!-- word:authoring:Runtime.Exit -->

Bind the native constant `crate::runtime::host::TerminalRequest::Exit` for this composition.

<!-- word-definition:authoring:Runtime.Exit -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

This statement is used inside `Runtime.Leave`. Its enclosing native signature and preceding bindings remain required.

<!-- example:glossary-runtime-exit-caller -->

</details>

### Storage.Function.Exit

<a id="word-authoring-storage-function-exit"></a>

<!-- word:authoring:Storage.Function.Exit -->

Wrap the supplied result in success and apply the selected enclosing Function return definition. This is separate from program Exit.

<!-- word-definition:authoring:Storage.Function.Exit -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

**Compiler use.** functionOperation selects Storage.Function.Exit with the constructed Break value and the caller-selected Return composition. This retained compiler excerpt selects the L-owned definition with typed native name/type/value or Process bindings. Those bindings come from the compiler’s resolved plan; this is an integration fragment, not an ordinary standalone L call. The `functionOperation` helper in `lyger/function_composition.go` selects `Storage.Function.` plus the supplied member name and preserves the selected Return composition.

<!-- example:glossary-lyger-execution-composition-611-618 -->

</details>

### Formatting.Expand

<a id="word-authoring-formatting-expand"></a>

<!-- word:authoring:Formatting.Expand -->

Place a decimal point in the digit string at the position determined by its exponent, prepending or appending zero digits when the point falls outside the string.

<!-- word-definition:authoring:Formatting.Expand -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

This statement is used inside `Formatting.Render`. Its enclosing native signature and preceding bindings remain required.

<!-- example:glossary-formatting-expand-caller -->

</details>

### Storage.Expect

<a id="word-authoring-storage-expect"></a>

<!-- word:authoring:Storage.Expect -->

Consume a native Option or Result and return its payload, or panic when it is absent or failed. The message argument is evaluated normally even on success.

<!-- word-definition:authoring:Storage.Expect -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

This statement is used inside `Runtime.Clean`. Its enclosing native signature and preceding bindings remain required.

<!-- example:glossary-storage-expect-caller -->

</details>

### Storage.Value.Expect

<a id="word-authoring-storage-value-expect"></a>

<!-- word:authoring:Storage.Value.Expect -->

Consume a stored value, verify its complete signature against an expected signature, and return the unchanged value on success.

<!-- word-definition:authoring:Storage.Value.Expect -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

The wrapper below imports `Storage` and `Runtime` and is a member of the example `GlossaryExamples` Dictionary. Its declared native inputs, outputs, lifetime/generic descriptors, and the original selected module's native types/support are required. Lyger resolves the source and composes its library AML; this fragment has no native compilation or execution claim.

<!-- example:glossary-native-storage-value-expect -->

</details>

### Runtime.App.Command.Explain

<a id="word-authoring-runtime-app-command-explain"></a>

<!-- word:authoring:Runtime.App.Command.Explain -->

Wrap a native IO error with the context `start application`, retaining its native cause.

<!-- word-definition:authoring:Runtime.App.Command.Explain -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

This statement is used inside `Runtime.App.Launch`. Its enclosing native signature and preceding bindings remain required.

<!-- example:glossary-runtime-app-command-explain-caller -->

</details>

### Runtime.App.Explain

<a id="word-authoring-runtime-app-explain"></a>

<!-- word:authoring:Runtime.App.Explain -->

Describe App misuse with its source label and request runtime failure using that Error.

<!-- word-definition:authoring:Runtime.App.Explain -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

This statement is used inside `Runtime.App.Read`. Its enclosing native signature and preceding bindings remain required.

<!-- example:glossary-runtime-app-explain-caller -->

</details>

### Runtime.Capture.Explain

<a id="word-authoring-runtime-capture-explain"></a>

<!-- word:authoring:Runtime.Capture.Explain -->

Attach `capture application stream` context to an IO error, store it as the capture failure, and report that capture should stop.

<!-- word-definition:authoring:Runtime.Capture.Explain -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

This statement is used inside `Runtime.Capture.Receive`. Its enclosing native signature and preceding bindings remain required.

<!-- example:glossary-runtime-capture-explain-caller -->

</details>

### Runtime.Error.Explain

<a id="word-authoring-runtime-error-explain"></a>

<!-- word:authoring:Runtime.Error.Explain -->

Construct a message from context and a native error's display text, retaining the native error as a shared cause.

<!-- word-definition:authoring:Runtime.Error.Explain -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

This statement is used inside `Runtime.App.Command.Explain`. Its enclosing native signature and preceding bindings remain required.

<!-- example:glossary-runtime-error-explain-caller -->

</details>

### Runtime.Explain

<a id="word-authoring-runtime-explain"></a>

<!-- word:authoring:Runtime.Explain -->

Transform a native Result's error with the supplied callback while preserving its successful payload. The selected operation is `Result.map_err(Visit)`.

<!-- word-definition:authoring:Runtime.Explain -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

This statement is used inside `Runtime.Invocation.Terminate`. Its enclosing native signature and preceding bindings remain required.

<!-- example:glossary-runtime-explain-caller -->

</details>

### Runtime.Invocation.Explain

<a id="word-authoring-runtime-invocation-explain"></a>

<!-- word:authoring:Runtime.Invocation.Explain -->

Wrap an IO error with `terminate application process group` context and retain its native cause.

<!-- word-definition:authoring:Runtime.Invocation.Explain -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

This statement is used inside `Runtime.Invocation.Terminate`. Its enclosing native signature and preceding bindings remain required.

<!-- example:glossary-runtime-explain-caller -->

</details>

### Runtime.Path.Explain

<a id="word-authoring-runtime-path-explain"></a>

<!-- word:authoring:Runtime.Path.Explain -->

Wrap an IO error with `locate orchestrator` context and retain its native cause.

<!-- word-definition:authoring:Runtime.Path.Explain -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

This statement is used inside `Runtime.Locate`. Its enclosing native signature and preceding bindings remain required.

<!-- example:glossary-runtime-path-explain-caller -->

</details>

### Runtime.Signals.Explain

<a id="word-authoring-runtime-signals-explain"></a>

<!-- word:authoring:Runtime.Signals.Explain -->

Wrap an IO error with `install application signal subscriptions` context and retain its native cause.

<!-- word-definition:authoring:Runtime.Signals.Explain -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

This statement is used inside `Runtime.Install`. Its enclosing native signature and preceding bindings remain required.

<!-- example:glossary-runtime-signals-explain-caller -->

</details>

### Runtime.Supervision.Explain

<a id="word-authoring-runtime-supervision-explain"></a>

<!-- word:authoring:Runtime.Supervision.Explain -->

Wrap an IO error with `terminate application leader` context and retain its native cause.

<!-- word-definition:authoring:Runtime.Supervision.Explain -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

This statement is used inside `Runtime.Supervision.Terminate`. Its enclosing native signature and preceding bindings remain required.

<!-- example:glossary-runtime-supervision-explain-caller -->

</details>

### Export

<a id="word-builtin-export"></a>

<!-- word:builtin:export -->

Recognize a configured export action whose general native execution is not implemented.

<!-- word-definition:builtin:export -->

**Unsupported operation.** The configured Export action has no executable instruction and is rejected during instruction lowering. Use [Public/Publicly](#word-builtin-public) for declaration visibility, as in the [public-values program](../../demo/paragraphs/public-values.l). That is not a runtime export operation. A separately authored action must supply its own definition.

**Rejected source illustration (not executed).** After defining a Text binding named Message, `Export message.` does not publish it or provide a native export operation. Use Public/Publicly on the declaration when visibility is intended.

### Storage.Bytes.Extend

<a id="word-authoring-storage-bytes-extend"></a>

<!-- word:authoring:Storage.Bytes.Extend -->

Append the supplied byte slice to a mutable byte buffer using `extend_from_slice`, preserving the exact bytes and returning the native unit result.

<!-- word-definition:authoring:Storage.Bytes.Extend -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

This statement is used inside `Runtime.Error.Attach`. Its enclosing native signature and preceding bindings remain required.

<!-- example:glossary-storage-bytes-caller -->

</details>

### Storage.Extract

<a id="word-authoring-storage-extract"></a>

<!-- word:authoring:Storage.Extract -->

Consume a Result and return its successful payload as Some, or discard its error and return None.

<!-- word-definition:authoring:Storage.Extract -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

This statement is used inside `Runtime.Capture.Receive`. Its enclosing native signature and preceding bindings remain required.

<!-- example:glossary-storage-extract-caller -->

</details>

## F

<a id="glossary-f"></a>

### Runtime.Capture.Fail

<a id="word-authoring-runtime-capture-fail"></a>

<!-- word:authoring:Runtime.Capture.Fail -->

Replace the optional capture failure with the supplied Error and return true to stop the capture loop.

<!-- word-definition:authoring:Runtime.Capture.Fail -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

This statement is used inside `Runtime.Capture.Explain`. Its enclosing native signature and preceding bindings remain required.

<!-- example:glossary-runtime-capture-fail-caller -->

</details>

### Runtime.Fail

<a id="word-authoring-runtime-fail"></a>

<!-- word:authoring:Runtime.Fail -->

Request runtime failure with a present Error; an absent Error becomes the explicit failure `cannot exit with an absent Error`.

<!-- word-definition:authoring:Runtime.Fail -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

This statement is used inside `Runtime.App.Explain`. Its enclosing native signature and preceding bindings remain required.

<!-- example:glossary-runtime-app-describe-caller -->

</details>

### Storage.Program.Fail

<a id="word-authoring-storage-program-fail"></a>

<!-- word:authoring:Storage.Program.Fail -->

Construct a runtime failure from the explicit failure input, bind its error and wrap that error in the canonical Rejection.

<!-- word-definition:authoring:Storage.Program.Fail -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

**Compiler use.** Wrap the transferred Error in the selected present value and return a failed program result. This retained compiler excerpt selects the L-owned definition with typed native name/type/value or Process bindings. Those bindings come from the compiler’s resolved plan; this is an integration fragment, not an ordinary standalone L call.

<!-- example:glossary-lyger-execution-composition-629-637 -->

</details>

### Runtime.Failure

<a id="word-authoring-runtime-failure"></a>

<!-- word:authoring:Runtime.Failure -->

Consume a native Result and retain its error as Some, or discard its successful payload as None. The selected operation is `Value.err()`.

<!-- word-definition:authoring:Runtime.Failure -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

This statement is used inside `Runtime.Reject`. Its enclosing native signature and preceding bindings remain required.

<!-- example:glossary-runtime-failure-caller -->

</details>

### Storage.Failure

<a id="word-authoring-storage-failure"></a>

<!-- word:authoring:Storage.Failure -->

Construct a runtime Error from a supplied borrowed message through `runtime::error_new_text`; it neither returns nor exits by itself.

<!-- word-definition:authoring:Storage.Failure -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

This statement is used inside `Storage.Validate`. Its enclosing native signature and preceding bindings remain required.

<!-- example:glossary-storage-failure-caller -->

</details>

### Runtime.Failures

<a id="word-authoring-runtime-failures"></a>

<!-- word:authoring:Runtime.Failures -->

Mutably borrow `Value.cleanup_errors` so the caller can update that exact native field.

<!-- word-definition:authoring:Runtime.Failures -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

This statement is used inside `Runtime.Invocation.Record`. Its enclosing native signature and preceding bindings remain required.

<!-- example:glossary-runtime-failures-caller -->

</details>

### Runtime.Fatal

<a id="word-authoring-runtime-fatal"></a>

<!-- word:authoring:Runtime.Fatal -->

Bind the native constant `crate::runtime::host::TerminalRequest::Failure` for this composition.

<!-- word-definition:authoring:Runtime.Fatal -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

This statement is used inside `Runtime.Refuse`. Its enclosing native signature and preceding bindings remain required.

<!-- example:glossary-runtime-fatal-caller -->

</details>

### Runtime.App.Fault

<a id="word-authoring-runtime-app-fault"></a>

<!-- word:authoring:Runtime.App.Fault -->

Borrow `Value.operation_error` without moving the referenced native field.

<!-- word-definition:authoring:Runtime.App.Fault -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

This statement is used inside `Runtime.App.Select`. Its enclosing native signature and preceding bindings remain required.

<!-- example:glossary-runtime-app-fault-caller -->

</details>

### Runtime.Error.Fault

<a id="word-authoring-runtime-error-fault"></a>

<!-- word:authoring:Runtime.Error.Fault -->

Construct a native `Error` value with `data` set to `Data`.

<!-- word-definition:authoring:Runtime.Error.Fault -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

This statement is used inside `Runtime.Error.Copy`. Its enclosing native signature and preceding bindings remain required.

<!-- example:glossary-runtime-error-datum-caller -->

</details>

### Runtime.Fault

<a id="word-authoring-runtime-fault"></a>

<!-- word:authoring:Runtime.Fault -->

Check whether the selected native Result is an error. The selected operation is `Value.is_err()`.

<!-- word-definition:authoring:Runtime.Fault -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

This statement is used inside `Runtime.Signals.Watch`. Its enclosing native signature and preceding bindings remain required.

<!-- example:glossary-runtime-fault-caller -->

</details>

### Runtime.Snapshot.Fault

<a id="word-authoring-runtime-snapshot-fault"></a>

<!-- word:authoring:Runtime.Snapshot.Fault -->

Borrow `Value.error` without moving the referenced native field.

<!-- word-definition:authoring:Runtime.Snapshot.Fault -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

This statement is used inside `Runtime.Snapshot.Copy`. Its enclosing native signature and preceding bindings remain required.

<!-- example:glossary-runtime-snapshot-fault-caller -->

</details>

### Runtime.App.Fetch

<a id="word-authoring-runtime-app-fetch"></a>

<!-- word:authoring:Runtime.App.Fetch -->

Lock App state and copy its optional current invocation handle.

<!-- word-definition:authoring:Runtime.App.Fetch -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

This statement is used inside `Runtime.App.Capture`. Its enclosing native signature and preceding bindings remain required.

<!-- example:glossary-runtime-app-fetch-caller -->

</details>

### Runtime.Capture.Fill

<a id="word-authoring-runtime-capture-fill"></a>

<!-- word:authoring:Runtime.Capture.Fill -->

Create an async read operation into the supplied capture buffer. The selected operation is `tokio::io::AsyncReadExt::read(Reader, Buffer)`.

<!-- word-definition:authoring:Runtime.Capture.Fill -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

This statement is used inside `Runtime.Capture.Advance`. Its enclosing native signature and preceding bindings remain required.

<!-- example:glossary-runtime-capture-fill-caller -->

</details>

### Runtime.State.Finalize

<a id="word-authoring-runtime-state-finalize"></a>

<!-- word:authoring:Runtime.State.Finalize -->

Mutably borrow `Value.finalized` so the caller can update that exact native field.

<!-- word-definition:authoring:Runtime.State.Finalize -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

This statement is used inside `Runtime.Conclude`. Its enclosing native signature and preceding bindings remain required.

<!-- example:glossary-runtime-state-finalize-caller -->

</details>

### Runtime.Finalized

<a id="word-authoring-runtime-finalized"></a>

<!-- word:authoring:Runtime.Finalized -->

Select the native field `Value.finalized` by value. Its declared type determines whether this copies or moves the field.

<!-- word-definition:authoring:Runtime.Finalized -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

This statement is used inside `Runtime.Admit`. Its enclosing native signature and preceding bindings remain required.

<!-- example:glossary-runtime-finalized-caller -->

</details>

### Formatting.Finish

<a id="word-authoring-formatting-finish"></a>

<!-- word:authoring:Formatting.Finish -->

Prefix a minus sign when the original float is sign-negative, then consume the completed digit string into a byte vector.

<!-- word-definition:authoring:Formatting.Finish -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

This statement is used inside `Formatting.Render`. Its enclosing native signature and preceding bindings remain required.

<!-- example:glossary-formatting-finish-caller -->

</details>

### Runtime.Finish

<a id="word-authoring-runtime-finish"></a>

<!-- word:authoring:Runtime.Finish -->

Serialize finalization, preserve a previously recorded status, handle the source result, await cleanup, report combined diagnostics, stop signal observation, and cache the final exit status.

<!-- word-definition:authoring:Runtime.Finish -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

This statement is used inside `Runtime.Run`. Its enclosing native signature and preceding bindings remain required.

<!-- example:glossary-runtime-finish-caller -->

</details>

### Runtime.Invocation.Finish

<a id="word-authoring-runtime-invocation-finish"></a>

<!-- word:authoring:Runtime.Invocation.Finish -->

Advance a claimed continuation to finished state and release its App's outstanding flag when that App's current invocation matches this invocation.

<!-- word-definition:authoring:Runtime.Invocation.Finish -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

The wrapper below imports `Storage` and `Runtime` and is a member of the example `GlossaryExamples` Dictionary. Its declared native inputs, outputs, lifetime/generic descriptors, and the original selected module's native types/support are required. Lyger resolves the source and composes its library AML; this fragment has no native compilation or execution claim.

<!-- example:glossary-native-runtime-invocation-finish -->

</details>

### Storage.Tuple.First

<a id="word-authoring-storage-tuple-first"></a>

<!-- word:authoring:Storage.Tuple.First -->

Select the first tuple component by value; the component's native type determines copy or move behavior.

<!-- word-definition:authoring:Storage.Tuple.First -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

This statement is used inside `Runtime.Invocation.Kill`. Its enclosing native signature and preceding bindings remain required.

<!-- example:glossary-storage-tuple-first-caller -->

</details>

### IO.Flush

<a id="word-stdlib-export-eng-io-io-flush"></a>

<!-- word:stdlib-export:eng:io:IO.Flush -->

Request a flush from a supplied mutable writer and return that writer's native IO result.

<!-- word-definition:stdlib-export:eng:io:IO.Flush -->

**Usage:** `a New IO's Flusher of Writer` calls `std::io::Write::flush` on the supplied `&mut W`. The caller handles `std::io::Result<()>`. Flush neither chooses stdout nor sends a payload, and successful flushing does not universally mean durable storage.

**Example:** `Writing.Flush` in [Native callers](#context-io--native-callers) uses the same mutable byte vector supplied by its host. [Send](#word-stdlib-export-eng-io-io-send), [Lined](#word-stdlib-export-eng-io-io-lined) and [Print](#word-stdlib-export-eng-io-io-print) add no explicit flush. The concrete writer determines what a flush does.

### Runtime.Flush

<a id="word-authoring-runtime-flush"></a>

<!-- word:authoring:Runtime.Flush -->

Create an async flush operation for the supplied output writer. The selected operation is `tokio::io::AsyncWriteExt::flush(Output)`.

<!-- word-definition:authoring:Runtime.Flush -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

This statement is used inside `Runtime.Show`. Its enclosing native signature and preceding bindings remain required.

<!-- example:glossary-runtime-flush-caller -->

</details>

### Runtime.Capture.Follow

<a id="word-authoring-runtime-capture-follow"></a>

<!-- word:authoring:Runtime.Capture.Follow -->

Subscribe to an invocation's drain deadline, retain its host dependencies, and spawn capture of the selected child stream.

<!-- word-definition:authoring:Runtime.Capture.Follow -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

This statement is used inside `Runtime.Invocation.Supervise`. Its enclosing native signature and preceding bindings remain required.

<!-- example:glossary-runtime-capture-follow-caller -->

</details>

### Storage.Value.Follow

<a id="word-authoring-storage-value-follow"></a>

<!-- word:authoring:Storage.Value.Follow -->

Read the target identity and visited-set references from a traversal context, then perform Value.Reach on the supplied value.

<!-- word-definition:authoring:Storage.Value.Follow -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

This statement is used inside `Storage.Value.Search`. Its enclosing native signature and preceding bindings remain required.

<!-- example:glossary-storage-search-caller -->

</details>

### For

<a id="word-grammar-word-for"></a>

<!-- word:grammar-word:for -->

Introduce the composed body of a source action or structured Process.

<!-- word-definition:grammar-word:for -->

**Usage and boundary.** Launch in the [local definitions program](../../demo/paragraphs/local-definitions.l) uses `for Building ...`, and the [selector caller](../examples/word-selection.l) uses `Process ... for Returning ...`. For is not a new loop primitive.

### Runtime.Error.Form

<a id="word-authoring-runtime-error-form"></a>

<!-- word:authoring:Runtime.Error.Form -->

Construct a runtime Error from owned message bytes with no child causes, native cause, or cancellation flag.

<!-- word-definition:authoring:Runtime.Error.Form -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

This statement is used inside `Runtime.Error.Describe`. Its enclosing native signature and preceding bindings remain required.

<!-- example:glossary-runtime-error-form-caller -->

</details>

### Formatting.Format

<a id="word-authoring-formatting-format"></a>

<!-- word:authoring:Formatting.Format -->

Convert a finite binary64 value into its shortest round-tripping decimal bytes, preserving signed zero. Use ordinary notation for nonzero magnitudes from 0.0001 inclusive to 1000000 exclusive, and signed scientific notation otherwise.

<!-- word-definition:authoring:Formatting.Format -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

This statement is used inside `Number.Render`. Its enclosing native signature and preceding bindings remain required.

<!-- example:glossary-formatting-format-caller -->

</details>

Existing qualified aliases: `Runtime.Format`. They select this canonical meaning.

### Storage.Number.Format

<a id="word-authoring-storage-number-format"></a>

<!-- word:authoring:Storage.Number.Format -->

Format the referenced finite Number through the canonical runtime number formatter and bind its byte vector.

<!-- word-definition:authoring:Storage.Number.Format -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

This statement is used inside `Storage.Number.Render`. Its enclosing native signature and preceding bindings remain required.

<!-- example:glossary-storage-number-caller -->

</details>

### Storage.Value.Format

<a id="word-authoring-storage-value-format"></a>

<!-- word:authoring:Storage.Value.Format -->

Borrow a stored value's erased payload and invoke its selected typed renderer Function, returning that renderer's byte result.

<!-- word-definition:authoring:Storage.Value.Format -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

This statement is used inside `Storage.Pointer.Render`. Its enclosing native signature and preceding bindings remain required.

<!-- example:glossary-storage-value-caller -->

</details>

### Formatting

<a id="word-authoring-formatting"></a>

<!-- word:authoring:Formatting -->

Name the explicit dictionary whose source Functions compose finite Number formatting into owned decimal bytes.

<!-- word-definition:authoring:Formatting -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

Import the dictionary to make its qualified members available.

<!-- example:glossary-namespace-formatting -->

</details>

### Storage.Number.Fraction

<a id="word-authoring-storage-number-fraction"></a>

<!-- word:authoring:Storage.Number.Fraction -->

Return the fractional part of the selected native floating-point value. The selected operation is `Value.fract()`.

<!-- word-definition:authoring:Storage.Number.Fraction -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

This statement is used inside `Storage.Index.Check`. Its enclosing native signature and preceding bindings remain required.

<!-- example:glossary-storage-number-fraction-caller -->

</details>

### From

<a id="word-grammar-word-from"></a>

<!-- word:grammar-word:from -->

Select the source of an import or a checked typed value read.

<!-- word-definition:grammar-word:from -->

**Usage and boundary.** The [first program](../guide/first-program.md#read-the-program) imports IO.Print from the `io` dictionary. The [value program](../examples/word-values.l) reads a Text from names at the key `first`. The owning import/read determines provider and type checks.

### Function

<a id="word-builtin-function"></a>

<!-- word:builtin:function -->

Define a callable with a selected constructor, ordered typed inputs and results, and a private source body.

<!-- word-definition:builtin:function -->

**Usage and example.** The [complete Identity/Relay program](../examples/composed-functions.l) declares two Functions, calls their invocation verbs and names each result. Ordinary source signatures and explicitly native signatures share this selected definition. Native generic/lifetime atoms require the [native definition context](../spec/definitions.md#process-function-and-return).

Function supplies an FW native definition with Func metadata. Name is Text, Inputs is Props, Outputs is Texts, and Process is an already instantiated definition reference.

Inputs and Outputs preserve their typed cardinality. Main supplies No Inputs and No Outputs. Ordinary Function declarations bind their checked signatures and paragraph bodies to the same selected definition. The constructor is not a marker for a second emitter.

**Supported boundary.** A Function declares one ordered signature and owns its body paragraph. A source Function with Outputs must return them on every continuing path. Recursive calls and Function-valued arguments/results use that same checked signature. Generic inference, overload resolution and variadic parameters are not part of this interface.

Invocation words use normal grammatical binding. Captures, snapshots and Pointer sharing follow [scope rules](../spec/names-and-scope.md#runtime-storage-and-definite-assignment). Native signature/lifetime metadata remains bounded by the [definition contract](../spec/definitions.md#process-function-and-return).

### Storage.Function

<a id="word-authoring-storage-function"></a>

<!-- word:authoring:Storage.Function -->

Name the Storage namespace for typed callable capture pairs, invocation results, ownership traversal and display.

<!-- word-definition:authoring:Storage.Function -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

This statement is used inside `Storage.Function.Copy`. Its enclosing native signature and preceding bindings remain required.

<!-- example:glossary-storage-function-caller -->

</details>

### Storage.Future

<a id="word-authoring-storage-future"></a>

<!-- word:authoring:Storage.Future -->

Construct and bind a move-capturing native future from the supplied body.

<!-- word-definition:authoring:Storage.Future -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

This statement is used inside `builtins/eng/storage/execution.l`. Its enclosing native signature and preceding bindings remain required.

<!-- example:glossary-storage-future-caller -->

</details>

## G

<a id="glossary-g"></a>

### Runtime.Inner.Gate

<a id="word-authoring-runtime-inner-gate"></a>

<!-- word:authoring:Runtime.Inner.Gate -->

Borrow `Value.finish_gate` without moving the referenced native field.

<!-- word-definition:authoring:Runtime.Inner.Gate -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

This statement is used inside `Runtime.Finish`. Its enclosing native signature and preceding bindings remain required.

<!-- example:glossary-runtime-acquire-caller -->

</details>

### Runtime.Invocation.Gate

<a id="word-authoring-runtime-invocation-gate"></a>

<!-- word:authoring:Runtime.Invocation.Gate -->

Borrow `Value.group` without moving the referenced native field.

<!-- word-definition:authoring:Runtime.Invocation.Gate -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

This statement is used inside `Runtime.Invocation.Identify`. Its enclosing native signature and preceding bindings remain required.

<!-- example:glossary-runtime-invocation-gate-caller -->

</details>

### Runtime.Gather

<a id="word-authoring-runtime-gather"></a>

<!-- word:authoring:Runtime.Gather -->

Claim shutdown once, set stopping state and cancellation, then snapshot the runtime-owned invocation handles for cleanup.

<!-- word-definition:authoring:Runtime.Gather -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

This statement is used inside `Runtime.Begin`. Its enclosing native signature and preceding bindings remain required.

<!-- example:glossary-runtime-gather-caller -->

</details>

### Generic

<a id="word-builtin-generic"></a>

<!-- word:builtin:generic -->

Describe a reusable source construction pattern whose explicit operands are supplied when applied.

<!-- word-definition:builtin:generic -->

**Usage and example.** Compact in the [source-action example](../../demo/paragraphs/local-definitions.l) describes a Generic App using any Process. `New Compact App` supplies its body. Generic alone does not infer native Rust type parameters, lifetimes or unconstrained constructor support.

**Supported boundary.** The selected source constructor declares the operands it accepts. Its callers and refinements retain that signature and canonical target. Generic describes this reusable construction mechanism, not a promise of inferred polymorphism.

### Get

<a id="word-builtin-get"></a>

<!-- word:builtin:get -->

Recognize a configured receptive action without a general supported collection instruction.

<!-- word-definition:builtin:get -->

**Unsupported operation.** The configured Get action has no executable instruction and is rejected during instruction lowering. Use the checked typed read form in the [value program](../examples/word-values.l), such as `Define "first" as a New Number from items at 0`. Get is not an alternate spelling for that operation. A separately authored action must supply its own definition.

**Rejected source illustration (not executed).** Even after `Define "items" as a New List of Number.` and `Put 7 in items.`, the configured request `Get 0 from items.` has no supported collection execution. Use the checked typed read described above.

### Storage.Get

<a id="word-authoring-storage-get"></a>

<!-- word:authoring:Storage.Get -->

Look up a native collection element by the supplied index or key and return its optional borrowed reference. The selected operation is `Collection.get(Index)`.

<!-- word-definition:authoring:Storage.Get -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

This statement is used inside `Storage.Program.Select`. Its enclosing native signature and preceding bindings remain required.

<!-- example:glossary-storage-get-caller -->

</details>

### Greater

<a id="word-grammar-word-greater"></a>

<!-- word:grammar-word:greater -->

Select a numeric greater-than comparison, with an optional equality continuation.

<!-- word-definition:grammar-word:greater -->

**Usage and boundary.** The focused form `left greater than right` requires Numbers. The phrase `left greater than equal to right` is the inclusive form. The [comparison witnesses](../../lyger/orchestrator_test.go) retain operand typing. No lexical Text ordering is implied.

**Usage example (source fragment, not executed here).** The declared Number on the left is greater than the Number on the right.

```l
Define "left" as 2.
Define "right" as 1.
If left greater than right, Show `larger`, else Show `not larger`.
```

### Runtime.Group

<a id="word-authoring-runtime-group"></a>

<!-- word:authoring:Runtime.Group -->

Construct the host signal target for the supplied process-group ID. The selected operation is `crate::runtime::host::SignalTarget::Group(Pid)`.

<!-- word-definition:authoring:Runtime.Group -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

This statement is used inside `Runtime.Invocation.Signal`. Its enclosing native signature and preceding bindings remain required.

<!-- example:glossary-runtime-group-caller -->

</details>

### Runtime.Invocation.Group

<a id="word-authoring-runtime-invocation-group"></a>

<!-- word:authoring:Runtime.Invocation.Group -->

Declare the native `GroupState` record with fields `pid: Option<u32>`, `cleaned: bool`, `termination_deadline: Option<tokio::time::Instant>`.

<!-- word-definition:authoring:Runtime.Invocation.Group -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

This statement is used inside `Runtime.Invocation.Allocate`. Its enclosing native signature and preceding bindings remain required.

<!-- example:glossary-runtime-invocation-group-caller -->

</details>

## H

<a id="glossary-h"></a>

### Runtime.Halt

<a id="word-authoring-runtime-halt"></a>

<!-- word:authoring:Runtime.Halt -->

Break from the enclosing native loop when the supplied Boolean condition is true; otherwise continue without an effect.

<!-- word-definition:authoring:Runtime.Halt -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

This statement is used inside `Runtime.Supervision.Await`. Its enclosing native signature and preceding bindings remain required.

<!-- example:glossary-runtime-halt-caller -->

</details>

### Runtime.Handle

<a id="word-authoring-runtime-handle"></a>

<!-- word:authoring:Runtime.Handle -->

Declare the native `Runtime` record with fields `inner: std::sync::Arc<RuntimeInner>`.

<!-- word-definition:authoring:Runtime.Handle -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

This statement is used inside `Runtime.Copy`. Its enclosing native signature and preceding bindings remain required.

<!-- example:glossary-runtime-handle-caller -->

</details>

### Have

<a id="word-grammar-word-have"></a>

<!-- word:grammar-word:have -->

Test whether the selected App has the Error identified by its current operation or invocation.

<!-- word-definition:grammar-word:have -->

**Usage and boundary.** `it has an Error` in the [conditional member fixture](../../lyger/test/member-resolution.l) follows Execute and refers to HelloWorldApp. Have is not general collection membership. The Error must belong to the same App.

### Storage.History

<a id="word-authoring-storage-history"></a>

<!-- word:authoring:Storage.History -->

Construct an empty ordered set used to track visited ownership identities. The selected operation is `std::collections::BTreeSet::new()`.

<!-- word-definition:authoring:Storage.History -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

This statement is used inside `Storage.Value.Protect`. Its enclosing native signature and preceding bindings remain required.

<!-- example:glossary-storage-history-caller -->

</details>

### Storage.Value.History

<a id="word-authoring-storage-value-history"></a>

<!-- word:authoring:Storage.Value.History -->

Mutably reborrow the visited-identity set stored in the second traversal-context component.

<!-- word-definition:authoring:Storage.Value.History -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

This statement is used inside `Storage.Value.Follow`. Its enclosing native signature and preceding bindings remain required.

<!-- example:glossary-storage-value-history-caller -->

</details>

### Runtime.Host

<a id="word-authoring-runtime-host"></a>

<!-- word:authoring:Runtime.Host -->

Name the runtime adapter namespace whose aliases select canonical startup and native-trait support definitions.

<!-- word-definition:authoring:Runtime.Host -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

This source-authoring fragment comes from `runtime/eng/dictionary.l` and retains that dictionary's surrounding imports and native binding context.

<!-- example:glossary-runtime-eng-dictionary-18-18 -->

</details>

## I

<a id="glossary-i"></a>

### Runtime.Identical

<a id="word-authoring-runtime-identical"></a>

<!-- word:authoring:Runtime.Identical -->

Compare the supplied Arc handles by allocation identity. The selected operation is `std::sync::Arc::ptr_eq(Left, Right)`.

<!-- word-definition:authoring:Runtime.Identical -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

This statement is used inside `Runtime.App.Resolve`. Its enclosing native signature and preceding bindings remain required.

<!-- example:glossary-runtime-identical-caller -->

</details>

### Storage.Identical

<a id="word-authoring-storage-identical"></a>

<!-- word:authoring:Storage.Identical -->

Replace a Storage.Comparison with Arc pointer-identity comparison of its original left and right operands.

<!-- word-definition:authoring:Storage.Identical -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

This descriptor applies to [Comparison](word-glossary.md#word-authoring-storage-comparison).

**Compiler use.** Apply the descriptor to canonical Storage.Comparison with two resolved App identity references. This retained compiler excerpt selects the L-owned definition with typed native name/type/value or Process bindings. Those bindings come from the compiler’s resolved plan; this is an integration fragment, not an ordinary standalone L call.

<!-- example:glossary-lyger-builtin-app-selection-51-56 -->

</details>

### Runtime.Error.Identify

<a id="word-authoring-runtime-error-identify"></a>

<!-- word:authoring:Runtime.Error.Identify -->

Return an ErrorData allocation address as `usize` for diagnostic deduplication.

<!-- word-definition:authoring:Runtime.Error.Identify -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

This statement is used inside `Runtime.Error.Append`. Its enclosing native signature and preceding bindings remain required.

<!-- example:glossary-runtime-error-identify-caller -->

</details>

### Runtime.Invocation.Identify

<a id="word-authoring-runtime-invocation-identify"></a>

<!-- word:authoring:Runtime.Invocation.Identify -->

Set an unstarted invocation's process ID while holding its process-group gate; failure to obtain that gate violates the explicit expectation.

<!-- word-definition:authoring:Runtime.Invocation.Identify -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

This statement is used inside `Runtime.App.Supervise`. Its enclosing native signature and preceding bindings remain required.

<!-- example:glossary-runtime-invocation-identify-caller -->

</details>

### Storage.Identify

<a id="word-authoring-storage-identify"></a>

<!-- word:authoring:Storage.Identify -->

Borrow an Rc-backed value and return the address of its allocation converted to `usize` for ownership-identity comparisons.

<!-- word-definition:authoring:Storage.Identify -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

This statement is used inside `Storage.Pointer.Reach`. Its enclosing native signature and preceding bindings remain required.

<!-- example:glossary-storage-identify-caller -->

</details>

### Runtime.Error.Identities

<a id="word-authoring-runtime-error-identities"></a>

<!-- word:authoring:Runtime.Error.Identities -->

Mutably reborrow the visited-error identity set in the second traversal-context component.

<!-- word-definition:authoring:Runtime.Error.Identities -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

This statement is used inside `Runtime.Error.Accumulate`. Its enclosing native signature and preceding bindings remain required.

<!-- example:glossary-runtime-error-identities-caller -->

</details>

### Runtime.Supervision.Identity

<a id="word-authoring-runtime-supervision-identity"></a>

<!-- word:authoring:Runtime.Supervision.Identity -->

Read the selected native child process's optional process ID. The selected operation is `Child.id()`.

<!-- word-definition:authoring:Runtime.Supervision.Identity -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

This statement is used inside `Runtime.Invocation.Supervise`. Its enclosing native signature and preceding bindings remain required.

<!-- example:glossary-runtime-supervision-identity-caller -->

</details>

### If

<a id="word-grammar-word-if"></a>

<!-- word:grammar-word:if -->

Select a source branch from a supported typed condition.

<!-- word-definition:grammar-word:if -->

**Usage and boundary.** The [Truth conditional program](../../testdata/phase1/language/lang-13.l) selects Await on true and Exit otherwise. Definite assignment is checked across all continuing paths.

**Partial support.** A single typed condition works. The configured list of conditions joined by And or Or is rejected during instruction lowering, and the configured `else if` continuation is rejected during grammar processing. Neither form is an executable extension of the supported If/Else pair.

### Storage.Ignore

<a id="word-authoring-storage-ignore"></a>

<!-- word:authoring:Storage.Ignore -->

Return successful false for a scalar ownership-reachability query without following any children.

<!-- word-definition:authoring:Storage.Ignore -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

The wrapper below imports `Storage` and `Runtime` and is a member of the example `GlossaryExamples` Dictionary. Its declared native inputs, outputs, lifetime/generic descriptors, and the original selected module's native types/support are required. Lyger resolves the source and composes its library AML; this fragment has no native compilation or execution claim.

<!-- example:glossary-native-storage-ignore -->

</details>

### Import

<a id="word-builtin-import"></a>

<!-- word:builtin:import -->

Recognize a configured import action whose general native execution is not implemented.

<!-- word-definition:builtin:import -->

**Unsupported operation.** The configured Import action has no executable instruction and is rejected during instruction lowering. Use [Use](#word-builtin-use) with an explicit source or registered dictionary. The [file-import program](../../demo/dictionaries/file-import/main.l) demonstrates that supported spelling. Import is not an alias for it. A separately authored action must supply its own definition.

**Rejected source illustration (not executed).** ``Import `io`.`` is not the supported indexed import. Write ``Use IO from `io`.`` to introduce that dictionary. The Import spelling identifies the unavailable operation.

### Runtime.Error.Import

<a id="word-authoring-runtime-error-import"></a>

<!-- word:authoring:Runtime.Error.Import -->

Construct a runtime Error from a native IO error's display bytes and retain that IO error as its native cause.

<!-- word-definition:authoring:Runtime.Error.Import -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

This statement is used inside `Runtime.App.Launch`. Its enclosing native signature and preceding bindings remain required.

<!-- example:glossary-runtime-error-import-caller -->

</details>

Existing qualified aliases: `Runtime.Host.Error.Import`. They select this canonical meaning.

### In

<a id="word-grammar-word-in"></a>

<!-- word:grammar-word:in -->

Identify the destination of a supported insertion or update.

<!-- word-definition:grammar-word:in -->

**Usage and boundary.** `Put 7 in items` in the [value program](../examples/word-values.l) appends a Number to its List. Pointer and Map updates have their own checked contracts. In does not select a lexical namespace.

### Storage.Index

<a id="word-authoring-storage-index"></a>

<!-- word:authoring:Storage.Index -->

Name the Storage namespace for checking numeric indices against the native index range.

<!-- word-definition:authoring:Storage.Index -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

This statement is used inside `Storage.Index.Check`. Its enclosing native signature and preceding bindings remain required.

<!-- example:glossary-storage-index-caller -->

</details>

### Runtime.Initial

<a id="word-authoring-runtime-initial"></a>

<!-- word:authoring:Runtime.Initial -->

Construct a native `RuntimeState` value with `owned` set to `Owned`, `cleanup_errors` set to `Errors`, `next` set to `0`, `stopping` set to `false`, `terminal` set to `None`, `finalized` set to `false`, `final_status` set to `None`.

<!-- word-definition:authoring:Runtime.Initial -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

This statement is used inside `Runtime.Initialize`. Its enclosing native signature and preceding bindings remain required.

<!-- example:glossary-runtime-initial-caller -->

</details>

### Runtime.Initialize

<a id="word-authoring-runtime-initialize"></a>

<!-- word:authoring:Runtime.Initialize -->

Construct initial runtime state with empty owned-invocation and cleanup-error maps.

<!-- word-definition:authoring:Runtime.Initialize -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

This statement is used inside `Runtime.Assemble`. Its enclosing native signature and preceding bindings remain required.

<!-- example:glossary-runtime-initialize-caller -->

</details>

### Runtime.Inner

<a id="word-authoring-runtime-inner"></a>

<!-- word:authoring:Runtime.Inner -->

Declare the native `RuntimeInner` record with fields `companion_directory: Result<std::path::PathBuf>`, `state: std::sync::Mutex<RuntimeState>`, `cancelled: tokio::sync::watch::Sender<bool>`, `shutdown_done: tokio::sync::watch::Sender<Option<Result<()>>>`, `signals: std::sync::Mutex<Option<tokio::task::JoinHandle<()>>>`, `stdout: tokio::sync::Mutex<crate::runtime::host::Output>`, `stderr: tokio::sync::Mutex<crate::runtime::host::Output>`, `dependencies: std::sync::Arc<dyn crate::runtime::host::Dependencies>`, `finish_gate: tokio::sync::Mutex<()>`.

<!-- word-definition:authoring:Runtime.Inner -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

This statement is used inside `Runtime.Gather`. Its enclosing native signature and preceding bindings remain required.

<!-- example:glossary-runtime-inner-caller -->

</details>

### Input

<a id="word-builtin-input"></a>

<!-- word:builtin:input -->

Name an ordered Function parameter or a typed binding supplied to a structured Process.

<!-- word-definition:builtin:input -->

**Usage and example.** `with Inputs of "Value" as Text` declares Identity's parameter in the [complete Function example](../examples/composed-functions.l). A selector's Process Inputs bind its selected payload inside that body. `No Inputs` explicitly declares no parameters. These forms do not provide stdin or argument transport to an App. The [execution limits](../spec/execution.md#targets-and-process-ownership) remain unchanged.

Function calls check input count and types. A Process input belongs only to its selected body, with native borrow and move checks at compilation. These supported bindings introduce neither an untyped argument list nor an implicit clone.

### Runtime.Insert

<a id="word-authoring-runtime-insert"></a>

<!-- word:authoring:Runtime.Insert -->

Insert the supplied value only when the selected native map entry is vacant. The selected operation is `Entry.or_insert(Value)`.

<!-- word-definition:authoring:Runtime.Insert -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

This statement is used inside `Runtime.Invocation.Record`. Its enclosing native signature and preceding bindings remain required.

<!-- example:glossary-runtime-insert-caller -->

</details>

### Storage.Insert

<a id="word-authoring-storage-insert"></a>

<!-- word:authoring:Storage.Insert -->

Insert the supplied value or key/value pair into the selected native collection and return that collection's insertion result. The selected operation is `Collection.insert(Key, Value)`.

<!-- word-definition:authoring:Storage.Insert -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

This statement is used inside `Storage.Map.Collect`. Its enclosing native signature and preceding bindings remain required.

<!-- example:glossary-storage-insert-caller -->

</details>

### Runtime.App.Inspect

<a id="word-authoring-runtime-app-inspect"></a>

<!-- word:authoring:Runtime.App.Inspect -->

Read the selected invocation's finalized Error when a snapshot exists; otherwise return the saved operation error when no invocation is selected, or absence while the selected invocation is incomplete.

<!-- word-definition:authoring:Runtime.App.Inspect -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

The wrapper below imports `Storage` and `Runtime` and is a member of the example `GlossaryExamples` Dictionary. Its declared native inputs, outputs, lifetime/generic descriptors, and the original selected module's native types/support are required. Lyger resolves the source and composes its library AML; this fragment has no native compilation or execution claim.

<!-- example:glossary-native-runtime-app-inspect -->

</details>

### Runtime.Error.Inspect

<a id="word-authoring-runtime-error-inspect"></a>

<!-- word:authoring:Runtime.Error.Inspect -->

Borrow the slice of child Errors retained by an Error.

<!-- word-definition:authoring:Runtime.Error.Inspect -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

This statement is used inside `Runtime.Error.Remember`. Its enclosing native signature and preceding bindings remain required.

<!-- example:glossary-runtime-error-inspect-caller -->

</details>

### Runtime.Path.Inspect

<a id="word-authoring-runtime-path-inspect"></a>

<!-- word:authoring:Runtime.Path.Inspect -->

Check the supplied companion path using the native executable-validation operation. The selected operation is `crate::process::validate_executable(Path)`.

<!-- word-definition:authoring:Runtime.Path.Inspect -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

This statement is used inside `Runtime.App.Locate`. Its enclosing native signature and preceding bindings remain required.

<!-- example:glossary-runtime-path-inspect-caller -->

</details>

### Runtime.Supervision.Inspect

<a id="word-authoring-runtime-supervision-inspect"></a>

<!-- word:authoring:Runtime.Supervision.Inspect -->

Convert a child wait result into an optional exit code, appending attributed wait or unsuccessful-exit Errors to the supplied error collection.

<!-- word-definition:authoring:Runtime.Supervision.Inspect -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

This statement is used inside `Runtime.Supervision.Complete`. Its enclosing native signature and preceding bindings remain required.

<!-- example:glossary-runtime-supervision-inspect-caller -->

</details>

### Storage.Inspect

<a id="word-authoring-storage-inspect"></a>

<!-- word:authoring:Storage.Inspect -->

Visit values in iterator order and apply Examine to each item, discarding successful callback payloads and propagating the first failure.

<!-- word-definition:authoring:Storage.Inspect -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

This statement is used inside `Storage.List.Create`. Its enclosing native signature and preceding bindings remain required.

<!-- example:glossary-storage-inspect-caller -->

</details>

### Runtime.Install

<a id="word-authoring-runtime-install"></a>

<!-- word:authoring:Runtime.Install -->

Install signal subscriptions, assemble a runtime from supplied host dependencies and output streams, spawn its signal watcher, and retain that task in the runtime.

<!-- word-definition:authoring:Runtime.Install -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

This statement is used inside `Runtime.Open`. Its enclosing native signature and preceding bindings remain required.

<!-- example:glossary-runtime-install-caller -->

</details>

### Runtime.Error.Interface

<a id="word-authoring-runtime-error-interface"></a>

<!-- word:authoring:Runtime.Error.Interface -->

Cast a shared native error to the Arc-backed Error + Send + Sync interface.

<!-- word-definition:authoring:Runtime.Error.Interface -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

This statement is used inside `Runtime.Error.Explain`. Its enclosing native signature and preceding bindings remain required.

<!-- example:glossary-runtime-error-interface-caller -->

</details>

### Runtime.Signals.Interpret

<a id="word-authoring-runtime-signals-interpret"></a>

<!-- word:authoring:Runtime.Signals.Interpret -->

Map signal status one to an `application output pipe closed` Error and other statuses to absence.

<!-- word-definition:authoring:Runtime.Signals.Interpret -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

This statement is used inside `Runtime.Signals.Watch`. Its enclosing native signature and preceding bindings remain required.

<!-- example:glossary-runtime-signals-interpret-caller -->

</details>

### Runtime.App.Invocation

<a id="word-authoring-runtime-app-invocation"></a>

<!-- word:authoring:Runtime.App.Invocation -->

Borrow `Value.invocation` without moving the referenced native field.

<!-- word-definition:authoring:Runtime.App.Invocation -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

This statement is used inside `Runtime.App.Fetch`. Its enclosing native signature and preceding bindings remain required.

<!-- example:glossary-runtime-app-invocation-caller -->

</details>

### Runtime.Invocation

<a id="word-authoring-runtime-invocation"></a>

<!-- word:authoring:Runtime.Invocation -->

Declare the native `Invocation` record with fields `app: std::sync::Weak<App>`, `runtime: std::sync::Weak<RuntimeInner>`, `sequence: u64`, `dependencies: std::sync::Arc<dyn crate::runtime::host::Dependencies>`, `wait_delay: tokio::time::Duration`, `state: std::sync::Mutex<InvocationState>`, `done: tokio::sync::watch::Sender<bool>`, `group: tokio::sync::Mutex<GroupState>`, `drain_deadline: tokio::sync::watch::Sender<Option<tokio::time::Instant>>`, `leader_requests: tokio::sync::mpsc::UnboundedSender<tokio::sync::oneshot::Sender<Result<()>>>`.

<!-- word-definition:authoring:Runtime.Invocation -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

This statement is used inside `Runtime.Invocation.Publish`. Its enclosing native signature and preceding bindings remain required.

<!-- example:glossary-runtime-invocation-caller -->

</details>

### Function.Invoke

<a id="word-builtin-function-invoke"></a>

<!-- word:builtin:function.invoke -->

Apply a selected Function target to its ordered, resolved argument atoms.

<!-- word-definition:builtin:function.invoke -->

**Usage and example.** The `Identity Message as "Nested"` call in the [complete Function example](../examples/composed-functions.l) uses the selected Function.Invoke definition. An author normally invokes the Function's own verb. Its Target and Arguments atoms are compiler-bound references, not an executable body hidden in text. [Selection tests](../../lyger/function_composition_test.go) establish that this implementation controls emitted calls.

Existing qualified aliases: `Storage.Routine.Invoke`. They select this canonical meaning.

### Storage.Invoke

<a id="word-authoring-storage-invoke"></a>

<!-- word:authoring:Storage.Invoke -->

Evaluate the selected Function value once, copy its environment, borrow the context, construct and await its invocation, then propagate failure and select the Continue payload or the enclosing caller’s Exit.

<!-- word-definition:authoring:Storage.Invoke -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

**Compiler use.** Compose a typed invocation definition, Function.Return-dependent exit alternative, environment copy, and context references. This retained compiler excerpt selects the L-owned definition with typed native name/type/value or Process bindings. Those bindings come from the compiler’s resolved plan; this is an integration fragment, not an ordinary standalone L call.

<!-- example:glossary-lyger-function-composition-282-300 -->

</details>

### IO

<a id="word-stdlib-eng-io"></a>

<!-- word:stdlib:eng:io -->

Name the explicit library dictionary for value printing and writer-directed byte output, with separate handle, send, newline and flush operations.

<!-- word-definition:stdlib:eng:io -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

Import the dictionary to make its qualified members available.

<!-- example:glossary-namespace-io -->

</details>

### Is

<a id="word-grammar-word-is"></a>

<!-- word:grammar-word:is -->

Associate a subject with its declared or assigned value. In a condition, compare supported like-typed values.

<!-- word-definition:grammar-word:is -->

**Usage and boundary.** `"message" is` followed by Text in the [first program](../guide/first-program.md#read-the-program) supplies assignment. The [comparison contract](../spec/control-flow.md#comparisons) distinguishes conditional equality and the configured negative form `isnt`. Neither implies arbitrary object equality.

### It

<a id="word-grammar-word-it"></a>

<!-- word:grammar-word:it -->

Refer to the last subject defined in the original source paragraph, or the applicable declared source-action argument.

<!-- word-definition:grammar-word:it -->

**Usage and boundary.** `Start it` in the [pronoun example](../../testdata/phase1/language/lang-06.l) uses the Job alias just defined. The [rejection fixture](../../testdata/phase1/n11-pronoun-crosses-paragraph/main.l) demonstrates the shared [paragraph boundary](../spec/names-and-scope.md#pronouns).

### Storage.Iterate

<a id="word-authoring-storage-iterate"></a>

<!-- word:authoring:Storage.Iterate -->

Borrow the selected collection and obtain its native iterator. The selected operation is `Value.iter()`.

<!-- word-definition:authoring:Storage.Iterate -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

This statement is used inside `Storage.Environment.Reach`. Its enclosing native signature and preceding bindings remain required.

<!-- example:glossary-storage-iterate-caller -->

</details>

### Storage.Iteration

<a id="word-authoring-storage-iteration"></a>

<!-- word:authoring:Storage.Iteration -->

Iterate the supplied native iterable, bind each item under the supplied label, and execute the supplied body in order.

<!-- word-definition:authoring:Storage.Iteration -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

This statement is used inside `builtins/eng/storage/bindings.l`. Its enclosing native signature and preceding bindings remain required.

<!-- example:glossary-storage-apply-caller -->

</details>

### Storage.Iterative

<a id="word-authoring-storage-iterative"></a>

<!-- word:authoring:Storage.Iterative -->

Describe a lending Storage.Routine with lifetime `'a` and `I: IntoIterator<Item = &'a LValue>`.

<!-- word-definition:authoring:Storage.Iterative -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

This statement is used inside `Storage.Value.Search`. Its enclosing native signature and preceding bindings remain required.

<!-- example:glossary-storage-iterative-caller -->

</details>

### Runtime.Error.Iterator

<a id="word-authoring-runtime-error-iterator"></a>

<!-- word:authoring:Runtime.Error.Iterator -->

Consume the selected collection into its owning iterator. The selected operation is `Value.into_iter()`.

<!-- word-definition:authoring:Runtime.Error.Iterator -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

This statement is used inside `Runtime.Error.Trace`. Its enclosing native signature and preceding bindings remain required.

<!-- example:glossary-runtime-error-iterator-caller -->

</details>

## J

<a id="glossary-j"></a>

### Runtime.Capture.Join

<a id="word-authoring-runtime-capture-join"></a>

<!-- word:authoring:Runtime.Capture.Join -->

Convert a completed capture task into shared bytes and an optional Error. A task-join failure produces empty captured bytes and a stream-attributed failure.

<!-- word-definition:authoring:Runtime.Capture.Join -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

This statement is used inside `Runtime.Invocation.Supervise`. Its enclosing native signature and preceding bindings remain required.

<!-- example:glossary-runtime-capture-join-caller -->

</details>

### Runtime.Error.Join

<a id="word-authoring-runtime-error-join"></a>

<!-- word:authoring:Runtime.Error.Join -->

Return None for an empty error list, preserve its sole Error for a singleton, or construct an empty-message parent retaining multiple causes.

<!-- word-definition:authoring:Runtime.Error.Join -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

This statement is used inside `Runtime.Diagnose`. Its enclosing native signature and preceding bindings remain required.

<!-- example:glossary-runtime-error-join-caller -->

</details>

### Runtime.Join

<a id="word-authoring-runtime-join"></a>

<!-- word:authoring:Runtime.Join -->

Visit the supplied invocations in order, constructing and awaiting each callback future before advancing. Reap uses a complete settle pass followed by a complete completion pass.

<!-- word-definition:authoring:Runtime.Join -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

This statement is used inside `Runtime.Reap`. Its enclosing native signature and preceding bindings remain required.

<!-- example:glossary-runtime-invocation-settle-caller -->

</details>

### Runtime.Path.Join

<a id="word-authoring-runtime-path-join"></a>

<!-- word:authoring:Runtime.Path.Join -->

Append Path to the native Parent path and return the resulting owned path through PathBuf::join.

<!-- word-definition:authoring:Runtime.Path.Join -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

This statement is used inside `Runtime.App.Locate`. Its enclosing native signature and preceding bindings remain required.

<!-- example:glossary-runtime-path-join-caller -->

</details>

### Storage.Join

<a id="word-authoring-storage-join"></a>

<!-- word:authoring:Storage.Join -->

Join the byte sequences in Values with the supplied Separator byte sequence using the native slice join operation.

<!-- word-definition:authoring:Storage.Join -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

This statement is used inside `Storage.Text.Join`. Its enclosing native signature and preceding bindings remain required.

<!-- example:glossary-storage-join-caller -->

</details>

### Storage.Text.Join

<a id="word-authoring-storage-text-join"></a>

<!-- word:authoring:Storage.Text.Join -->

Join byte-vector parts with the exact UTF-8 bytes of the supplied separator string, returning an owned byte vector.

<!-- word-definition:authoring:Storage.Text.Join -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

This statement is used inside `Storage.Map.Describe`. Its enclosing native signature and preceding bindings remain required.

<!-- example:glossary-storage-text-join-caller -->

</details>

## K

<a id="glossary-k"></a>

### Runtime.Invocation.Kill

<a id="word-authoring-runtime-invocation-kill"></a>

<!-- word:authoring:Runtime.Invocation.Kill -->

Send a leader-termination request to the invocation supervisor and await its reply. A closed request or reply channel is treated as already settled.

<!-- word-definition:authoring:Runtime.Invocation.Kill -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

This statement is used inside `Runtime.Invocation.Settle`. Its enclosing native signature and preceding bindings remain required.

<!-- example:glossary-runtime-invocation-kill-caller -->

</details>

### Runtime.Kill

<a id="word-authoring-runtime-kill"></a>

<!-- word:authoring:Runtime.Kill -->

Bind the native constant `nix::sys::signal::Signal::SIGKILL` for this composition.

<!-- word-definition:authoring:Runtime.Kill -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

This statement is used inside `Runtime.Invocation.Signal`. Its enclosing native signature and preceding bindings remain required.

<!-- example:glossary-runtime-kill-caller -->

</details>

## L

<a id="glossary-l"></a>

### Runtime.App.Label

<a id="word-authoring-runtime-app-label"></a>

<!-- word:authoring:Runtime.App.Label -->

Borrow `Value.source` without moving the referenced native field.

<!-- word-definition:authoring:Runtime.App.Label -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

This statement is used inside `Runtime.App.Describe`. Its enclosing native signature and preceding bindings remain required.

<!-- example:glossary-runtime-app-label-caller -->

</details>

### Runtime.Path.Label

<a id="word-authoring-runtime-path-label"></a>

<!-- word:authoring:Runtime.Path.Label -->

Obtain a native path-display adapter for diagnostic formatting. The selected operation is `Path.display()`.

<!-- word-definition:authoring:Runtime.Path.Label -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

This statement is used inside `Runtime.App.Locate`. Its enclosing native signature and preceding bindings remain required.

<!-- example:glossary-runtime-path-context-caller -->

</details>

### Runtime.App.Launch

<a id="word-authoring-runtime-app-launch"></a>

<!-- word:authoring:Runtime.App.Launch -->

Locate and validate an App companion, configure a native child with null stdin and piped stdout/stderr, and start it through the selected host dependencies. Wrap validation and launch failures with context.

<!-- word-definition:authoring:Runtime.App.Launch -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

This statement is used inside `Runtime.App.Run`. Its enclosing native signature and preceding bindings remain required.

<!-- example:glossary-runtime-app-launch-caller -->

</details>

### Storage.Layout

<a id="word-authoring-storage-layout"></a>

<!-- word:authoring:Storage.Layout -->

Name the Storage namespace containing native layout declarations for environments, Lists, Maps, Pointers and program storage.

<!-- word-definition:authoring:Storage.Layout -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

This statement is used inside `builtins/eng/storage/values.l`. Its enclosing native signature and preceding bindings remain required.

<!-- example:glossary-storage-layout-caller -->

</details>

### Runtime.Leave

<a id="word-authoring-runtime-leave"></a>

<!-- word:authoring:Runtime.Leave -->

Record a normal-exit request, request status zero with no Error, and return the cancellation Error used to stop source execution. A previously selected terminal decision can take precedence.

<!-- word-definition:authoring:Runtime.Leave -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

The wrapper below imports `Storage` and `Runtime` and is a member of the example `GlossaryExamples` Dictionary. Its declared native inputs, outputs, lifetime/generic descriptors, and the original selected module's native types/support are required. Lyger resolves the source and composes its library AML; this fragment has no native compilation or execution claim.

<!-- example:glossary-native-runtime-leave -->

</details>

### Storage.Lending

<a id="word-authoring-storage-lending"></a>

<!-- word:authoring:Storage.Lending -->

Add explicitly supplied native lifetime parameters to a Storage.Routine while retaining its callable signature, generic type parameters and Process body.

<!-- word-definition:authoring:Storage.Lending -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

This statement is used inside `builtins/eng/storage/execution.l`. Its enclosing native signature and preceding bindings remain required.

<!-- example:glossary-storage-lending-caller -->

</details>

### Storage.Length

<a id="word-authoring-storage-length"></a>

<!-- word:authoring:Storage.Length -->

Read the selected native value's length. The selected operation is `Value.len()`.

<!-- word-definition:authoring:Storage.Length -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

This statement is used inside `Runtime.Error.Attach`. Its enclosing native signature and preceding bindings remain required.

<!-- example:glossary-storage-length-caller -->

</details>

### Storage.Lent

<a id="word-authoring-storage-lent"></a>

<!-- word:authoring:Storage.Lent -->

Describe a lending Storage.Routine with a native type parameter `T` and lifetime `'a`.

<!-- word-definition:authoring:Storage.Lent -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

This statement is used inside `Storage.Access`. Its enclosing native signature and preceding bindings remain required.

<!-- example:glossary-storage-lent-caller -->

</details>

### Less

<a id="word-grammar-word-less"></a>

<!-- word:grammar-word:less -->

Select a numeric less-than comparison, with an optional equality continuation.

<!-- word-definition:grammar-word:less -->

**Usage and boundary.** The focused form `left less than right` requires two Number operands. The phrase `left less than equal to right` is the inclusive grammar. These [typed comparison witnesses](../../lyger/orchestrator_test.go) establish the native boundary. Text ordering is rejected.

**Usage example (source fragment, not executed here).** The operands are declared Numbers. The strict comparison chooses the first branch when Left is smaller.

```l
Define "left" as 1.
Define "right" as 2.
If left less than right, Show `smaller`, else Show `not smaller`.
```

### Runtime.Invocation.Limit

<a id="word-authoring-runtime-invocation-limit"></a>

<!-- word:authoring:Runtime.Invocation.Limit -->

Mutably borrow `Value.termination_deadline` so the caller can update that exact native field.

<!-- word-definition:authoring:Runtime.Invocation.Limit -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

This statement is used inside `Runtime.Invocation.Signal`. Its enclosing native signature and preceding bindings remain required.

<!-- example:glossary-runtime-invocation-delay-caller -->

</details>

### Runtime.Limit

<a id="word-authoring-runtime-limit"></a>

<!-- word:authoring:Runtime.Limit -->

Set an absent cleanup deadline or replace an existing deadline with the earlier of its current value and the supplied instant.

<!-- word-definition:authoring:Runtime.Limit -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

This statement is used inside `Runtime.Revise`. Its enclosing native signature and preceding bindings remain required.

<!-- example:glossary-runtime-amend-caller -->

</details>

### Storage.Index.Limit

<a id="word-authoring-storage-index-limit"></a>

<!-- word:authoring:Storage.Index.Limit -->

Convert the native maximum usize value to f64 for the exclusive numeric-index range check.

<!-- word-definition:authoring:Storage.Index.Limit -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

This statement is used inside `Storage.Index.Check`. Its enclosing native signature and preceding bindings remain required.

<!-- example:glossary-storage-index-caller -->

</details>

### IO.Lined

<a id="word-stdlib-export-eng-io-io-lined"></a>

<!-- word:stdlib-export:eng:io:IO.Lined -->

Describe IO.Send's Sender actor so that it sends one LF through the same writer after successful payload delivery.

<!-- word-definition:stdlib-export:eng:io:IO.Lined -->

**Usage:** `a New IO.Lined IO's Sender of Writer and Payload` targets **IO.Send's Sender actor (`IO.Sender`)**, retaining its selected invocation and original writer. Lined is not a modifier for any unrelated word spelled Sender.

**Example:** `Writing.Line` in [Native callers](#context-io--native-callers) returns the lined operation's `std::io::Result<()>`. A payload failure skips LF. A newline failure follows an already completed payload. Read [Explicit byte output](#context-io--explicit-byte-output) before choosing a retry policy: a line is not a transaction.

### List

<a id="word-type-list"></a>

<!-- word:type:list -->

Own an ordered sequence of values with a checked element type.

<!-- word-definition:type:list -->

**Usage and example.** The [value program](../examples/word-values.l) declares a List of Number, appends 7, replaces it with 9 and reads the checked element. The following excerpt contains that sequence.

#### Declare a List

<a id="word-type-list--declare-a-list"></a>

<!-- example:word-values-list -->

`of Number` declares the element type. The read states its expected result type. Incompatible reads or updates reject. Indices must be nonnegative integers within bounds. A List constructed from differently typed values infers Thing elements and preserves their individual kinds. Typed reads verify complete element types and Function signatures. A List owns its elements. An L List does not become native `Props` or `Texts` metadata.

### Storage.Layout.List

<a id="word-authoring-storage-layout-list"></a>

<!-- word:authoring:Storage.Layout.List -->

Declare the native `LList` record with fields `element: Vec<u64>`, `values: Vec<Value>`.

<!-- word-definition:authoring:Storage.Layout.List -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

This statement is used inside `builtins/eng/storage/values.l`. Its enclosing native signature and preceding bindings remain required.

<!-- example:glossary-storage-layout-list-caller -->

</details>

### Storage.List

<a id="word-authoring-storage-list"></a>

<!-- word:authoring:Storage.List -->

Declare the native `LList` record with fields `element: Vec<u64>`, `values: Vec<Value>`.

<!-- word-definition:authoring:Storage.List -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

This statement is used inside `Storage.List.Read`. Its enclosing native signature and preceding bindings remain required.

<!-- example:glossary-storage-list-caller -->

</details>

### Runtime.Listen

<a id="word-authoring-runtime-listen"></a>

<!-- word:authoring:Runtime.Listen -->

Observe a completed shutdown result or wait for a notification change; store an explicit failure if the notification channel closes before completion.

<!-- word-definition:authoring:Runtime.Listen -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

This statement is used inside `Runtime.Clean`. Its enclosing native signature and preceding bindings remain required.

<!-- example:glossary-runtime-listen-caller -->

</details>

### Storage.Literal

<a id="word-authoring-storage-literal"></a>

<!-- word:authoring:Storage.Literal -->

Bind the supplied compiler-encoded Thing literal as a native value.

<!-- word-definition:authoring:Storage.Literal -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

**Compiler use.** Bind the resolved Truth or finite Number atom to the selected literal definition. This retained compiler excerpt selects the L-owned definition with typed native name/type/value or Process bindings. Those bindings come from the compiler’s resolved plan; this is an integration fragment, not an ordinary standalone L call.

<!-- example:glossary-lyger-word-values-218-232 -->

</details>

### Runtime.App.Locate

<a id="word-authoring-runtime-app-locate"></a>

<!-- word:authoring:Runtime.App.Locate -->

Reject blank or nonrelative companion paths, join an allowed path to the orchestrator directory, validate the executable, and return its native path or an attributed runtime Error.

<!-- word-definition:authoring:Runtime.App.Locate -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

This statement is used inside `Runtime.App.Build`. Its enclosing native signature and preceding bindings remain required.

<!-- example:glossary-runtime-app-locate-caller -->

</details>

### Runtime.Locate

<a id="word-authoring-runtime-locate"></a>

<!-- word:authoring:Runtime.Locate -->

Locate and canonicalize the orchestrator executable, then return its parent directory or an attributed runtime failure.

<!-- word-definition:authoring:Runtime.Locate -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

This statement is used inside `Runtime.Assemble`. Its enclosing native signature and preceding bindings remain required.

<!-- example:glossary-runtime-locate-caller -->

</details>

### Storage.Locate

<a id="word-authoring-storage-locate"></a>

<!-- word:authoring:Storage.Locate -->

Look up a native collection element by the supplied index and return an optional mutable reference. The selected operation is `Collection.get_mut(Index)`.

<!-- word-definition:authoring:Storage.Locate -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

This statement is used inside `Storage.Environment.Update`. Its enclosing native signature and preceding bindings remain required.

<!-- example:glossary-storage-locate-caller -->

</details>

### Runtime.Invocation.Lock

<a id="word-authoring-runtime-invocation-lock"></a>

<!-- word:authoring:Runtime.Invocation.Lock -->

Attempt to acquire the selected asynchronous mutex immediately, returning its native guard or failure. The selected operation is `Gate.try_lock()`.

<!-- word-definition:authoring:Runtime.Invocation.Lock -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

This statement is used inside `Runtime.Invocation.Identify`. Its enclosing native signature and preceding bindings remain required.

<!-- example:glossary-runtime-invocation-gate-caller -->

</details>

### Storage.Loop

<a id="word-authoring-storage-loop"></a>

<!-- word:authoring:Storage.Loop -->

Expand the supplied body as a native unbounded loop; the body owns termination and return behavior.

<!-- word-definition:authoring:Storage.Loop -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

This statement is used inside `runtime/eng/process.l`. Its enclosing native signature and preceding bindings remain required.

<!-- example:glossary-storage-loop-caller -->

</details>

## M

<a id="glossary-m"></a>

### Main

<a id="word-builtin-main"></a>

<!-- word:builtin:main -->

Describe the canonical native entry structure for a Function, Object or Service.

<!-- word-definition:builtin:main -->

**Usage and example.** The [nested construction example](../spec/applicational-types.md#compose-an-object-inside-nested-classes) applies Main to an Object, while Basic composes its Service rule for an App. Choose the actual target below rather than treating all three refinements as interchangeable.

Main is a New Adjective describing Service, Object, or Function. Its ordered refinements compose:

#### Main Function

<a id="word-builtin-main--main-function"></a>

The Function rule constructs a Generic `main` Function from the supplied Process, with No Inputs and No Outputs. It chooses an entry signature. It does not accept arbitrary Function arguments.

#### Main Object

<a id="word-builtin-main--main-object"></a>

The Object rule constructs a Generic `main` Object containing a New Objective Main Function. The Process becomes that Function's body, not an Object initializer.

#### Main Service

<a id="word-builtin-main--main-service"></a>

The Service rule constructs a Generic `main` Service containing a New Main Object. Basic uses this structure inside an App.

These are ordinary source-defined descriptor rules. Aliases preserve their identity, and each construction applies every rule matching its target in authored order. Main currently has one rule for each target above. Repeated construction remains distinct.

Main does not grant standalone Service execution. A standalone declaration is structural. Standalone Service body effects and native process operations reject. Apps own executable identity through their construction.

### Map

<a id="word-type-map"></a>

<!-- word:type:map -->

Own values indexed by exact Text keys with a checked value type.

<!-- word-definition:type:map -->

**Usage and example.** The [value program](../examples/word-values.l) stores Ada under the Text key first and reads it back. A Map is runtime data, unlike a vocabulary Dictionary.

<!-- example:word-values-map -->

`as` supplies the Text key, and `at` selects it. Reading a missing key reports an error. No default value is inserted. Text keys preserve their bytes rather than applying Unicode normalization.

### Runtime.Map

<a id="word-authoring-runtime-map"></a>

<!-- word:authoring:Runtime.Map -->

Construct an empty native ordered map with key and value types selected by its enclosing composition. The selected operation is `std::collections::BTreeMap::new()`.

<!-- word-definition:authoring:Runtime.Map -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

This statement is used inside `Runtime.Initialize`. Its enclosing native signature and preceding bindings remain required.

<!-- example:glossary-runtime-map-caller -->

</details>

### Storage.Layout.Map

<a id="word-authoring-storage-layout-map"></a>

<!-- word:authoring:Storage.Layout.Map -->

Declare the native `LMap` record with fields `element: Vec<u64>`, `values: std::collections::BTreeMap<Vec<u8>, Value>`.

<!-- word-definition:authoring:Storage.Layout.Map -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

This statement is used inside `builtins/eng/storage/values.l`. Its enclosing native signature and preceding bindings remain required.

<!-- example:glossary-storage-layout-caller -->

</details>

### Storage.Map

<a id="word-authoring-storage-map"></a>

<!-- word:authoring:Storage.Map -->

Declare the native `LMap` record with fields `element: Vec<u64>`, `values: std::collections::BTreeMap<Vec<u8>, Value>`.

<!-- word-definition:authoring:Storage.Map -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

This statement is used inside `Storage.Map.Read`. Its enclosing native signature and preceding bindings remain required.

<!-- example:glossary-storage-environment-values-caller -->

</details>

### Runtime.Invocation.Mark

<a id="word-authoring-runtime-invocation-mark"></a>

<!-- word:authoring:Runtime.Invocation.Mark -->

Mutably borrow `Value.claim` so the caller can update that exact native field.

<!-- word-definition:authoring:Runtime.Invocation.Mark -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

This statement is used inside `Runtime.Invocation.Claim`. Its enclosing native signature and preceding bindings remain required.

<!-- example:glossary-runtime-invocation-mark-caller -->

</details>

### Storage.Signature.Match

<a id="word-authoring-storage-signature-match"></a>

<!-- word:authoring:Storage.Signature.Match -->

Accept an actual signature when it exactly equals the expected signature or the expected signature is the universal Thing signature.

<!-- word-definition:authoring:Storage.Signature.Match -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

This statement is used inside `Storage.Value.Check`. Its enclosing native signature and preceding bindings remain required.

<!-- example:glossary-storage-signature-match-caller -->

</details>

### Runtime.Minimum

<a id="word-authoring-runtime-minimum"></a>

<!-- word:authoring:Runtime.Minimum -->

Return the lesser of the supplied native values. The selected operation is `Left.min(Right)`.

<!-- word-definition:authoring:Runtime.Minimum -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

This statement is used inside `Runtime.Limit`. Its enclosing native signature and preceding bindings remain required.

<!-- example:glossary-runtime-minimum-caller -->

</details>

### Runtime.Missing

<a id="word-authoring-runtime-missing"></a>

<!-- word:authoring:Runtime.Missing -->

Bind the native constant `nix::libc::ESRCH` for this composition.

<!-- word-definition:authoring:Runtime.Missing -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

This statement is used inside `Runtime.Normalize`. Its enclosing native signature and preceding bindings remain required.

<!-- example:glossary-runtime-missing-caller -->

</details>

### Runtime.Modifier

<a id="word-authoring-runtime-modifier"></a>

<!-- word:authoring:Runtime.Modifier -->

Construct a native closure that calls the supplied deadline visitor with the current optional deadline and captured deadline value.

<!-- word-definition:authoring:Runtime.Modifier -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

This statement is used inside `Runtime.Revise`. Its enclosing native signature and preceding bindings remain required.

<!-- example:glossary-runtime-amend-caller -->

</details>

### Storage.Mutable

<a id="word-authoring-storage-mutable"></a>

<!-- word:authoring:Storage.Mutable -->

Describe a Storage.Binding as a mutable native binding while preserving its initializer.

<!-- word-definition:authoring:Storage.Mutable -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

This statement is used inside `Runtime.Error.Render`. Its enclosing native signature and preceding bindings remain required.

<!-- example:glossary-storage-mutable-caller -->

</details>

### Runtime.Mutex

<a id="word-authoring-runtime-mutex"></a>

<!-- word:authoring:Runtime.Mutex -->

Construct a synchronous native mutex around the supplied value. The selected operation is `std::sync::Mutex::new(Value)`.

<!-- word-definition:authoring:Runtime.Mutex -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

This statement is used inside `Runtime.Assemble`. Its enclosing native signature and preceding bindings remain required.

<!-- example:glossary-runtime-mutex-caller -->

</details>

## N

<a id="glossary-n"></a>

### Runtime.Path.Name

<a id="word-authoring-runtime-path-name"></a>

<!-- word:authoring:Runtime.Path.Name -->

Expose a Path as a borrowed native OsStr. The separate Vacancy predicate checks whether that representation is empty.

<!-- word-definition:authoring:Runtime.Path.Name -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

This statement is used inside `Runtime.App.Locate`. Its enclosing native signature and preceding bindings remain required.

<!-- example:glossary-runtime-path-name-caller -->

</details>

### Runtime.Error.Native

<a id="word-authoring-runtime-error-native"></a>

<!-- word:authoring:Runtime.Error.Native -->

Obtain the selected value's native borrowed view, preserving any optional wrapper provided by that method. The selected operation is `Value.data.native.as_ref()`.

<!-- word-definition:authoring:Runtime.Error.Native -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

This statement is used inside `Runtime.Error.Trace`. Its enclosing native signature and preceding bindings remain required.

<!-- example:glossary-runtime-error-native-caller -->

</details>

### Runtime.Native

<a id="word-authoring-runtime-native"></a>

<!-- word:authoring:Runtime.Native -->

Bind the native constant `crate::runtime::host::NativeDependencies` for this composition.

<!-- word-definition:authoring:Runtime.Native -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

This statement is used inside `Runtime.Open`. Its enclosing native signature and preceding bindings remain required.

<!-- example:glossary-runtime-dependencies-caller -->

</details>

### Storage.Negation

<a id="word-authoring-storage-negation"></a>

<!-- word:authoring:Storage.Negation -->

Compare a native Boolean value with false and bind the negated result.

<!-- word-definition:authoring:Storage.Negation -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

**Compiler use.** Negate the already-evaluated condition binding. This retained compiler excerpt selects the L-owned definition with typed native name/type/value or Process bindings. Those bindings come from the compiler’s resolved plan; this is an integration fragment, not an ordinary standalone L call.

<!-- example:glossary-lyger-execution-composition-406-408 -->

</details>

### New

<a id="word-builtin-new"></a>

<!-- word:builtin:new -->

Request construction through the selected noun and its matching descriptor rules.

<!-- word-definition:builtin:new -->

**Usage and example.** `New Basic App` in the [first program](../guide/first-program.md#read-the-program) applies Basic and Main to the App body. `New List of Number` in the [value program](../examples/word-values.l) selects checked List construction. New supplies no universal constructor, clone or initialization fallback.

### No

<a id="word-grammar-word-no"></a>

<!-- word:grammar-word:no -->

Express an absent signature component or negate a condition where that grammar permits it.

<!-- word-definition:grammar-word:no -->

**Usage and boundary.** Main's Function refinement uses `No Inputs and No Outputs` in the [descriptor example](../spec/lexical-types.md#describe-with-an-adjective). This explicit empty signature is not an unassigned value or implicit runtime default.

### Runtime.Normalize

<a id="word-authoring-runtime-normalize"></a>

<!-- word:authoring:Runtime.Normalize -->

Preserve successful signal results, treat the native missing-process error as success, and propagate other signal failures.

<!-- word-definition:authoring:Runtime.Normalize -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

This statement is used inside `Runtime.Invocation.Signal`. Its enclosing native signature and preceding bindings remain required.

<!-- example:glossary-runtime-normalize-caller -->

</details>

### Not

<a id="word-grammar-word-not"></a>

<!-- word:grammar-word:not -->

Negate a condition only where its grammatical form admits this modifier.

<!-- word-definition:grammar-word:not -->

**Usage and boundary.** The grammar admits `not` in its bounded condition forms. See the [condition rules](../../lexicon/eng/grammar.json) and negative-condition [lowering witnesses](../../lyger/orchestrator_test.go). This is not an unrestricted prefix Boolean operator or a promise that joined conditions execute.

**Example.** Here `enabled` is a declared Truth value. Negating its false value selects the first branch and writes `disabled` without a newline.

<!-- example:word-not -->

### Storage.Nothing

<a id="word-authoring-storage-nothing"></a>

<!-- word:authoring:Storage.Nothing -->

Construct the native unit value `()`.

<!-- word-definition:authoring:Storage.Nothing -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

This statement is used inside `Runtime.Invocation.Observe`. Its enclosing native signature and preceding bindings remain required.

<!-- example:glossary-storage-nothing-caller -->

</details>

### Runtime.Inner.Notification

<a id="word-authoring-runtime-inner-notification"></a>

<!-- word:authoring:Runtime.Inner.Notification -->

Borrow `Value.shutdown_done` without moving the referenced native field.

<!-- word-definition:authoring:Runtime.Inner.Notification -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

This statement is used inside `Runtime.Reap`. Its enclosing native signature and preceding bindings remain required.

<!-- example:glossary-runtime-inner-notification-caller -->

</details>

### Runtime.Invocation.Notification

<a id="word-authoring-runtime-invocation-notification"></a>

<!-- word:authoring:Runtime.Invocation.Notification -->

Borrow `Value.drain_deadline` without moving the referenced native field.

<!-- word-definition:authoring:Runtime.Invocation.Notification -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

This statement is used inside `Runtime.Invocation.Bound`. Its enclosing native signature and preceding bindings remain required.

<!-- example:glossary-runtime-invocation-notification-caller -->

</details>

### Runtime.Notify

<a id="word-authoring-runtime-notify"></a>

<!-- word:authoring:Runtime.Notify -->

Replace a watch channel's current value and notify receivers, discarding the previous value. The selected operation is `Sender.send_replace(Value)`.

<!-- word-definition:authoring:Runtime.Notify -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

This statement is used inside `Runtime.Gather`. Its enclosing native signature and preceding bindings remain required.

<!-- example:glossary-runtime-notify-caller -->

</details>

Existing qualified aliases: `Runtime.Invocation.Notify`. They select this canonical meaning.

### Noun

<a id="word-type-noun"></a>

<!-- word:type:noun -->

Classify a word that names a value, category, reference or constructed definition.

<!-- word-definition:type:noun -->

**Grammatical usage.** The [selector caller](../examples/word-selection.l) constructs `a New Process`, using Process as a noun. The [first program](../guide/first-program.md#read-the-program) likewise uses App and Text as nouns. Process’s Processing form is an applicative verb derived from its active slot. Noun classifies these words; it is not a scalar runtime type.

### Runtime.Now

<a id="word-authoring-runtime-now"></a>

<!-- word:authoring:Runtime.Now -->

Read the current native monotonic instant. The selected operation is `tokio::time::Instant::now()`.

<!-- word-definition:authoring:Runtime.Now -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

This statement is used inside `Runtime.Invocation.Supervise`. Its enclosing native signature and preceding bindings remain required.

<!-- example:glossary-runtime-now-caller -->

</details>

### Null

<a id="word-type-null"></a>

<!-- word:type:null -->

Represent an explicitly initialized absence value, distinct from an unassigned binding.

<!-- word-definition:type:null -->

**Usage and example.** The [value program](../examples/word-values.l) defines nothing as Null and shows `null`. IO.Print renders its native Null operand as `<nil>` instead. These policies are intentionally distinct.

`Define "nothing" as Null.` initializes an absence value, and Show writes `null`. A typed Null read still evaluates its source and checks its type, index or key. Null does not replace definite assignment. The separate native-definition Null atom displays `<nil>`.

Compiler labels are separate from public types. The labels `ID`, `None`, and `Unknown` track resolution, `Lexeme` stores lexical definitions, and invalid sentinels indicate compiler state. These implementation labels are not additional application words. [Error](#word-type-error) is a separate runtime carrier, not an internal sentinel.

### Runtime.App.Null

<a id="word-authoring-runtime-app-null"></a>

<!-- word:authoring:Runtime.App.Null -->

Construct the native null standard-stream configuration. The selected operation is `std::process::Stdio::null()`.

<!-- word-definition:authoring:Runtime.App.Null -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

This statement is used inside `Runtime.App.Launch`. Its enclosing native signature and preceding bindings remain required.

<!-- example:glossary-runtime-app-caller -->

</details>

### Storage.Null

<a id="word-authoring-storage-null"></a>

<!-- word:authoring:Storage.Null -->

Name the Storage namespace for displaying the explicit absence value.

<!-- word-definition:authoring:Storage.Null -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

This statement is used inside `Storage.Null.Render`. Its enclosing native signature and preceding bindings remain required.

<!-- example:glossary-storage-null-caller -->

</details>

### Number

<a id="word-type-number"></a>

<!-- word:type:number -->

Represent a finite binary64 number for typed storage, comparison and supported rendering.

<!-- word-definition:type:number -->

**Usage and example.** `Define "Count" as 7` in [public-values](../../demo/paragraphs/public-values.l) supplies a Number. The printing fragment below demonstrates the formatter and requires its App context.

Number uses finite IEEE-754 binary64 (`float64` in Lyger, `f64` in Rust). It is not an arbitrary-precision integer, and L has no source integer-width family.

The number parser accepts these forms. Source token boundaries also apply. Parser acceptance alone does not make a spelling a complete standalone token.

| Source spelling | Represented consequence |
| --- | --- |
| `3`, `+3`, `3.0` | The same binary64 value |
| `-12.5` | Negative finite fraction |
| `1e6` | One million |
| `0x1.8p+1` | Three |
| `1_000` | One thousand |
| `9007199254740993` | Rounded to the representable value 9007199254740992 |
| `5e-324` | Smallest positive subnormal |
| `1e-4000` | Underflows to positive zero |
| `-1e-4000` | Underflows to negative zero |

Positive and negative zero retain their sign through native encoding. Overflow, such as `1e999`, rejects with its range error preserved. Numeric-looking declared words can resolve as names. Although the lower parser recognizes nonfinite spellings, native scalar emission and typed definition binding reject NaN and Infinity.

Number supports equality and order comparisons with Number. Text/Number comparisons do not coerce. The English source grammar has no general arithmetic operator expressions. Native logeme calculations are a separate interface.

`IO.Print` uses shortest round-tripping binary64 formatting, corresponding to `strconv.FormatFloat(value, 'g', -1, 64)`. One million prints as `1e+06`, followed by LF.

<!-- example:values-number-printing -->

The three calls print `3.5`, `-12.5`, and `1e+06`, each followed by LF when the App runs. This fragment supplies only the App body. The [first program](../guide/first-program.md) shows the lifecycle and output handling needed to execute it.

### Storage.Number

<a id="word-authoring-storage-number"></a>

<!-- word:authoring:Storage.Number -->

Name the Storage namespace for native finite/fraction checks and canonical Number display.

<!-- word-definition:authoring:Storage.Number -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

This statement is used inside `Storage.Number.Render`. Its enclosing native signature and preceding bindings remain required.

<!-- example:glossary-storage-number-caller -->

</details>

## O

<a id="glossary-o"></a>

### Object

<a id="word-type-object"></a>

<!-- word:type:object -->

Own a native source file containing composed definitions and Functions.

<!-- word-definition:type:object -->

**Usage and example.** The Main Object in the [recursive construction example](../spec/applicational-types.md#compose-an-object-inside-nested-classes) owns the entry Function. Object construction supplies source content, not an object initializer or process launch.

Native imports and visibility belong to their owning file/module. Class metadata does not silently become every child's default.

### Objective

<a id="word-builtin-objective"></a>

<!-- word:builtin:objective -->

Group an existing Function definition as source content for its containing Object.

<!-- word-definition:builtin:objective -->

**Usage and example.** `New Objective First Function` in the [Object construction fixture](../../lyger/test/classes/interleaved.l) puts the selected Function into its Object. It does not invoke the Function or infer a signature.

Objective describes any Function by wrapping its existing definition in a grouped native source word containing A and the Function. It preserves the input definition and grouping. It does not clone its runtime semantics or infer additional function arguments.

### Runtime.Capture.Observe

<a id="word-authoring-runtime-capture-observe"></a>

<!-- word:authoring:Runtime.Capture.Observe -->

Continue after a successful deadline-change notice, or store `capture deadline notification closed` and stop after notification failure.

<!-- word-definition:authoring:Runtime.Capture.Observe -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

This statement is used inside `Runtime.Capture.Advance`. Its enclosing native signature and preceding bindings remain required.

<!-- example:glossary-runtime-capture-observe-caller -->

</details>

### Runtime.Invocation.Observe

<a id="word-authoring-runtime-invocation-observe"></a>

<!-- word:authoring:Runtime.Invocation.Observe -->

Invoke the selected host checkpoint hook with the supplied event and a retained invocation handle.

<!-- word-definition:authoring:Runtime.Invocation.Observe -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

This statement is used inside `Runtime.App.Stop`. Its enclosing native signature and preceding bindings remain required.

<!-- example:glossary-runtime-checkpoint-caller -->

</details>

### Runtime.Observe

<a id="word-authoring-runtime-observe"></a>

<!-- word:authoring:Runtime.Observe -->

Borrow the current watch-channel value and return its native clone.

<!-- word-definition:authoring:Runtime.Observe -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

This statement is used inside `Runtime.Listen`. Its enclosing native signature and preceding bindings remain required.

<!-- example:glossary-runtime-observe-caller -->

</details>

### Storage.Entry.Observe

<a id="word-authoring-storage-entry-observe"></a>

<!-- word:authoring:Storage.Entry.Observe -->

Call Main once, construct its runtime execution, await the resulting status and apply the supplied Return definition.

<!-- word-definition:authoring:Storage.Entry.Observe -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

**Compiler use.** Compose entry observation with the selected Function.Return definition and native u8 entry result. This retained compiler excerpt selects the L-owned definition with typed native name/type/value or Process bindings. Those bindings come from the compiler’s resolved plan; this is an integration fragment, not an ordinary standalone L call.

<!-- example:glossary-lyger-execution-composition-125-136 -->

</details>

### Of

<a id="word-grammar-word-of"></a>

<!-- word:grammar-word:of -->

Supply declared contents, type arguments or operands to a supported construction or signature.

<!-- word-definition:grammar-word:of -->

**Usage and boundary.** `New List of Number` in the [value program](../examples/word-values.l) gives the element type. The phrase `Inputs of "Value" as Text` in the [Function example](../examples/composed-functions.l) declares a parameter. The selected owner determines the contract.

### Runtime.Open

<a id="word-authoring-runtime-open"></a>

<!-- word:authoring:Runtime.Open -->

Create native host dependencies and pinned stdout/stderr outputs, then install the runtime and its signal watcher.

<!-- word-definition:authoring:Runtime.Open -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>


The wrapper below imports `Storage` and `Runtime` and is a member of the example `GlossaryExamples` Dictionary. Its declared native inputs, outputs, lifetime/generic descriptors, and the original selected module's native types/support are required. Lyger resolves the source and composes its library AML; this fragment has no native compilation or execution claim.

<!-- example:glossary-native-runtime-open -->

</details>

Existing qualified aliases: `Runtime.Host.Open`. They select this canonical meaning.

### Runtime.App.Operation

<a id="word-authoring-runtime-app-operation"></a>

<!-- word:authoring:Runtime.App.Operation -->

Mutably borrow `Value.operation_error` so the caller can update that exact native field.

<!-- word-definition:authoring:Runtime.App.Operation -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

This statement is used inside `Runtime.App.Build`. Its enclosing native signature and preceding bindings remain required.

<!-- example:glossary-runtime-app-operation-caller -->

</details>

### Storage.Option

<a id="word-authoring-storage-option"></a>

<!-- word:authoring:Storage.Option -->

Name the Storage dictionary namespace containing operations that compose native optional values.

<!-- word-definition:authoring:Storage.Option -->

**Usage:** After ``Use Storage from `storage`.`` qualify a member such as `Storage.Option.Select`. Option is a Dictionary namespace, not an L carrier type: the native Function signature uses `Option<T>`, with a real payload type in place of `T`.

**Example:** `Selection.Choose` in [Native callers](#context-storage--native-callers) receives `Option<u64>` and selects its Some or None case. The qualified member is [Option.Select](#word-authoring-storage-option-select). Result's two payload-bearing outcomes belong to [Result](#word-authoring-storage-result).

### Or

<a id="word-grammar-word-or"></a>

<!-- word:grammar-word:or -->

Join alternative nominal targets in supported descriptor declarations, or occur in recognized condition coordination.

<!-- word-definition:grammar-word:or -->

**Usage and boundary.** Main declares `Service, Object, or Function` in the [descriptor example](../spec/lexical-types.md#describe-with-an-adjective). These are alternative canonical targets.

**Partial support.** Descriptor target alternatives work, but an `or`-joined runtime condition is rejected during instruction lowering. The grammar recognizes the joined conditions, but there is no Boolean join operator or short-circuit implementation for them.

### Runtime.Ordinal

<a id="word-authoring-runtime-ordinal"></a>

<!-- word:authoring:Runtime.Ordinal -->

Select the native field `Value.next` by value. Its declared type determines whether this copies or moves the field.

<!-- word-definition:authoring:Runtime.Ordinal -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

This statement is used inside `Runtime.Count`. Its enclosing native signature and preceding bindings remain required.

<!-- example:glossary-runtime-ordinal-caller -->

</details>

### Otherwise

<a id="word-grammar-word-otherwise"></a>

<!-- word:grammar-word:otherwise -->

Attach an alternative branch to the immediately preceding unmatched conditional in the same paragraph.

<!-- word-definition:grammar-word:otherwise -->

**Usage and boundary.** The [conditional member fixture](../../lyger/test/member-resolution.l) follows an If with Otherwise. Intervening statements or a paragraph boundary do not grant attachment to an earlier condition.

### Out

<a id="word-grammar-word-out"></a>

<!-- word:grammar-word:out -->

Introduce a Function argument, with a separately configured deflective role that has no executable operation.

<!-- word-definition:grammar-word:out -->

**Usage example (source fragment).** After declaring Identity in the [Function example](../examples/composed-functions.l), ``Identity out `value` as "Result".`` uses the generic argument phrase. Out does not give that call output or export behavior.

**Partial support.** The [configured deflective role](../../lexicon/eng/vocabulary.json) has no corresponding operation or dedicated phrase in the current grammar. Only the generic Function argument use is implemented. Use the [Output member](#word-builtin-output) or IO operations for their distinct documented behavior.

### App.Output

<a id="word-builtin-output"></a>

<!-- word:builtin:output -->

Read the finalized raw standard-output bytes of an App invocation as Text.

<!-- word-definition:builtin:output -->

**Usage and example.** `Show Hello's Output` in the [first program](../guide/first-program.md#read-the-program) follows Execute. Assigning Output makes a snapshot, demonstrated by the [repeated-activation example](../../demo/paragraphs/repeated-activation.l). Function signature `Outputs` names result types instead. See [Function](#word-builtin-function).

Reading Output before finalization is fatal misuse. No hidden join occurs, and an already finalized read does not activate the continuation. The [result contract](../cli-build.md#invocation-and-results) covers availability, separate stderr capture and snapshots across reruns.

### IO.Output

<a id="word-stdlib-export-eng-io-io-output"></a>

<!-- word:stdlib-export:eng:io:IO.Output -->

Obtain a handle to the current process's standard-output stream without writing or locking it.

<!-- word-definition:stdlib-export:eng:io:IO.Output -->

**Usage:** In a native-signature Function, `a New IO's Output` produces `std::io::Stdout`. This is distinct from [builtin App Output](#word-builtin-output), which exposes finalized captured stdout bytes.

**Example:** `Writing.Output` in [Native callers](#context-io--native-callers) returns the handle. Choose [Diagnostic](#word-stdlib-export-eng-io-io-diagnostic) for stderr, or use [Print](#word-stdlib-export-eng-io-io-print) when its rendering, newline and failure policy are wanted together.

### Runtime.Capture.Output

<a id="word-authoring-runtime-capture-output"></a>

<!-- word:authoring:Runtime.Capture.Output -->

Convert the supplied bytes into a shared Arc-backed byte slice. The selected operation is `std::sync::Arc::from(Value)`.

<!-- word-definition:authoring:Runtime.Capture.Output -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

This statement is used inside `Runtime.Capture.Share`. Its enclosing native signature and preceding bindings remain required.

<!-- example:glossary-runtime-capture-bytes-caller -->

</details>

### Runtime.Output

<a id="word-authoring-runtime-output"></a>

<!-- word:authoring:Runtime.Output -->

Describe Storage.Pin's Pinner actor as a pinned box explicitly typed as the host Output interface.

<!-- word-definition:authoring:Runtime.Output -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

This statement is used inside `Runtime.Open`. Its enclosing native signature and preceding bindings remain required.

<!-- example:glossary-runtime-output-caller -->

</details>

### Runtime.Snapshot.Output

<a id="word-authoring-runtime-snapshot-output"></a>

<!-- word:authoring:Runtime.Snapshot.Output -->

Borrow `Value.stdout` without moving the referenced native field.

<!-- word-definition:authoring:Runtime.Snapshot.Output -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

This statement is used inside `Runtime.Snapshot.Copy`. Its enclosing native signature and preceding bindings remain required.

<!-- example:glossary-runtime-snapshot-output-caller -->

</details>

### Runtime.Supervision.Output

<a id="word-authoring-runtime-supervision-output"></a>

<!-- word:authoring:Runtime.Supervision.Output -->

Bind the native constant `crate::runtime::host::Stream::Stdout` for this composition.

<!-- word-definition:authoring:Runtime.Supervision.Output -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

This statement is used inside `Runtime.Invocation.Supervise`. Its enclosing native signature and preceding bindings remain required.

<!-- example:glossary-runtime-supervision-output-caller -->

</details>

### Storage.App.Output

<a id="word-authoring-storage-app-output"></a>

<!-- word:authoring:Storage.App.Output -->

Borrow an App handle and call the runtime operation that reads its finalized captured stdout bytes.

<!-- word-definition:authoring:Storage.App.Output -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

**Compiler use.** Read the selected App member through its typed App reference; Output retains the shared byte result. This retained compiler excerpt selects the L-owned definition with typed native name/type/value or Process bindings. Those bindings come from the compiler’s resolved plan; this is an integration fragment, not an ordinary standalone L call.

<!-- example:glossary-lyger-execution-composition-318-335 -->

</details>

### Runtime.App.Outstanding

<a id="word-authoring-runtime-app-outstanding"></a>

<!-- word:authoring:Runtime.App.Outstanding -->

Mutably borrow `Value.outstanding` so the caller can update that exact native field.

<!-- word-definition:authoring:Runtime.App.Outstanding -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

This statement is used inside `Runtime.App.Acquire`. Its enclosing native signature and preceding bindings remain required.

<!-- example:glossary-runtime-app-outstanding-caller -->

</details>

### Runtime.App.Own

<a id="word-authoring-runtime-app-own"></a>

<!-- word:authoring:Runtime.App.Own -->

Upgrade an App's weak runtime reference and return a runtime handle, or report `App runtime is unavailable`.

<!-- word-definition:authoring:Runtime.App.Own -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

This statement is used inside `Runtime.App.Read`. Its enclosing native signature and preceding bindings remain required.

<!-- example:glossary-runtime-app-own-caller -->

</details>

### Runtime.Invocation.Own

<a id="word-authoring-runtime-invocation-own"></a>

<!-- word:authoring:Runtime.Invocation.Own -->

Upgrade an invocation's weak runtime owner, or return `invocation runtime is unavailable`.

<!-- word-definition:authoring:Runtime.Invocation.Own -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

This statement is used inside `Runtime.Invocation.Claim`. Its enclosing native signature and preceding bindings remain required.

<!-- example:glossary-runtime-invocation-own-caller -->

</details>

### Storage.Own

<a id="word-authoring-storage-own"></a>

<!-- word:authoring:Storage.Own -->

Copy the selected native slice into an owned vector. The selected operation is `Value.to_vec()`.

<!-- word-definition:authoring:Storage.Own -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

This statement is used inside `Storage.Text.Own`. Its enclosing native signature and preceding bindings remain required.

<!-- example:glossary-storage-own-caller -->

</details>

### Storage.Text.Own

<a id="word-authoring-storage-text-own"></a>

<!-- word:authoring:Storage.Text.Own -->

Copy a borrowed native string's bytes into an owned byte vector.

<!-- word-definition:authoring:Storage.Text.Own -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

This statement is used inside `Storage.App.Render`. Its enclosing native signature and preceding bindings remain required.

<!-- example:glossary-storage-text-caller -->

</details>

### Runtime.State.Owned

<a id="word-authoring-runtime-state-owned"></a>

<!-- word:authoring:Runtime.State.Owned -->

Obtain a borrowed iterator over the selected ordered map's values. The selected operation is `Value.owned.values()`.

<!-- word-definition:authoring:Runtime.State.Owned -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

This statement is used inside `Runtime.Gather`. Its enclosing native signature and preceding bindings remain required.

<!-- example:glossary-runtime-cloned-caller -->

</details>

## P

<a id="glossary-p"></a>

### Storage.Pair

<a id="word-authoring-storage-pair"></a>

<!-- word:authoring:Storage.Pair -->

Pair the left iterator with the right iterator in order, stopping when either iterator ends. The selected operation is `Left.zip(Right)`.

<!-- word-definition:authoring:Storage.Pair -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

This statement is used inside `Storage.Environment.Create`. Its enclosing native signature and preceding bindings remain required.

<!-- example:glossary-storage-pair-caller -->

</details>

### Storage.Paired

<a id="word-authoring-storage-paired"></a>

<!-- word:authoring:Storage.Paired -->

Describe Storage.Iteration as pair-destructuring iteration, or Storage.Apply's actor as a callback invocation with first item, second item and context operands.

<!-- word-definition:authoring:Storage.Paired -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

This statement is used inside `builtins/eng/storage/traversal.l`. Its enclosing native signature and preceding bindings remain required.

<!-- example:glossary-storage-paired-caller -->

</details>

### Runtime.App.Parent

<a id="word-authoring-runtime-app-parent"></a>

<!-- word:authoring:Runtime.App.Parent -->

Borrow `Value.runtime` without moving the referenced native field.

<!-- word-definition:authoring:Runtime.App.Parent -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

This statement is used inside `Runtime.App.Own`. Its enclosing native signature and preceding bindings remain required.

<!-- example:glossary-runtime-app-parent-caller -->

</details>

### Runtime.Invocation.Parent

<a id="word-authoring-runtime-invocation-parent"></a>

<!-- word:authoring:Runtime.Invocation.Parent -->

Borrow `Value.runtime` without moving the referenced native field.

<!-- word-definition:authoring:Runtime.Invocation.Parent -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

This statement is used inside `Runtime.Invocation.Own`. Its enclosing native signature and preceding bindings remain required.

<!-- example:glossary-runtime-invocation-parent-caller -->

</details>

### Runtime.Parent

<a id="word-authoring-runtime-parent"></a>

<!-- word:authoring:Runtime.Parent -->

Read the supplied path's optional parent path. The selected operation is `Path.parent()`.

<!-- word-definition:authoring:Runtime.Parent -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

This statement is used inside `Runtime.Locate`. Its enclosing native signature and preceding bindings remain required.

<!-- example:glossary-runtime-parent-caller -->

</details>

### Runtime.Path.Parts

<a id="word-authoring-runtime-path-parts"></a>

<!-- word:authoring:Runtime.Path.Parts -->

Obtain a mutable native path-component iterator. The selected operation is `Value.components()`.

<!-- word-definition:authoring:Runtime.Path.Parts -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

This statement is used inside `Runtime.App.Locate`. Its enclosing native signature and preceding bindings remain required.

<!-- example:glossary-runtime-path-parts-caller -->

</details>

### Runtime.App.Path

<a id="word-authoring-runtime-app-path"></a>

<!-- word:authoring:Runtime.App.Path -->

Borrow `Value.path` without moving the referenced native field.

<!-- word-definition:authoring:Runtime.App.Path -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

This statement is used inside `Runtime.App.Locate`. Its enclosing native signature and preceding bindings remain required.

<!-- example:glossary-runtime-app-path-caller -->

</details>

### Runtime.Path

<a id="word-authoring-runtime-path"></a>

<!-- word:authoring:Runtime.Path -->

Copy the selected borrowed path into an owned PathBuf. The selected operation is `Path.to_path_buf()`.

<!-- word-definition:authoring:Runtime.Path -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

This statement is used inside `Runtime.Path.Admit`. Its enclosing native signature and preceding bindings remain required.

<!-- example:glossary-runtime-path-caller -->

</details>

### Storage.Permanent

<a id="word-authoring-storage-permanent"></a>

<!-- word:authoring:Storage.Permanent -->

Describe a Storage.Routine with the native generic bound `T: 'static`.

<!-- word-definition:authoring:Storage.Permanent -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

This statement is used inside `Storage.Value.Restore`. Its enclosing native signature and preceding bindings remain required.

<!-- example:glossary-storage-permanent-caller -->

</details>

### Runtime.Invocation.Phase

<a id="word-authoring-runtime-invocation-phase"></a>

<!-- word:authoring:Runtime.Invocation.Phase -->

Select the native field `Value.claim` by value. Its declared type determines whether this copies or moves the field.

<!-- word-definition:authoring:Runtime.Invocation.Phase -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

This statement is used inside `Runtime.Invocation.Claim`. Its enclosing native signature and preceding bindings remain required.

<!-- example:glossary-runtime-invocation-phase-caller -->

</details>

### Runtime.Invocation.Pid

<a id="word-authoring-runtime-invocation-pid"></a>

<!-- word:authoring:Runtime.Invocation.Pid -->

Select the native field `Value.pid` by value. Its declared type determines whether this copies or moves the field.

<!-- word-definition:authoring:Runtime.Invocation.Pid -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

This statement is used inside `Runtime.Invocation.Signal`. Its enclosing native signature and preceding bindings remain required.

<!-- example:glossary-runtime-invocation-pid-caller -->

</details>

### Storage.Pin

<a id="word-authoring-storage-pin"></a>

<!-- word:authoring:Storage.Pin -->

Move the supplied value into a pinned Box and bind it to the supplied result label. Future construction and awaiting remain separate operations.

<!-- word-definition:authoring:Storage.Pin -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

This statement is used inside `runtime/eng/process.l`. Its enclosing native signature and preceding bindings remain required.

<!-- example:glossary-storage-pin-caller -->

</details>

### Runtime.App.Pipe

<a id="word-authoring-runtime-app-pipe"></a>

<!-- word:authoring:Runtime.App.Pipe -->

Construct the native piped standard-stream configuration. The selected operation is `std::process::Stdio::piped()`.

<!-- word-definition:authoring:Runtime.App.Pipe -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

This statement is used inside `Runtime.App.Launch`. Its enclosing native signature and preceding bindings remain required.

<!-- example:glossary-runtime-app-pipe-caller -->

</details>

### Place

<a id="word-builtin-place"></a>

<!-- word:builtin:place -->

Replace an existing indexed List element or supported keyed runtime field.

<!-- word-definition:builtin:place -->

**Usage and example.** `Place 9 in items at 0` in the [value program](../examples/word-values.l) retains the List's Number element type. The index is checked. Place does not append past the end.

### Pointer

<a id="word-type-pointer"></a>

<!-- word:type:pointer -->

Hold a managed shared reference to a typed runtime value.

<!-- word-definition:type:pointer -->

**Usage and example.** The [complete value program](../examples/word-values.l) copies a Pointer handle, changes the shared target and reads it as Text. The focused excerpt below preserves those exact source lines.

<!-- example:word-values-pointer -->

The Pointer owns a copy of the initial value. Both handles share that target, so this prints `second`, while `message` remains `first`. Target types are checked. Pointer and capture-environment writes reject ownership cycles. Reads release their checked borrow before a nested call or Await. The last owner releases the target. There is no pointer arithmetic or user-managed deallocation.

### Storage.Layout.Pointer

<a id="word-authoring-storage-layout-pointer"></a>

<!-- word:authoring:Storage.Layout.Pointer -->

Declare the native `LPointer` record with fields `element: Vec<u64>`, `target: std::rc::Rc<std::cell::RefCell<Value>>`.

<!-- word-definition:authoring:Storage.Layout.Pointer -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

This statement is used inside `builtins/eng/storage/values.l`. Its enclosing native signature and preceding bindings remain required.

<!-- example:glossary-storage-layout-pointer-caller -->

</details>

### Storage.Pointer

<a id="word-authoring-storage-pointer"></a>

<!-- word:authoring:Storage.Pointer -->

Declare the native `LPointer` record with fields `element: Vec<u64>`, `target: std::rc::Rc<std::cell::RefCell<Value>>`.

<!-- word-definition:authoring:Storage.Pointer -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

This statement is used inside `Storage.Pointer.Reach`. Its enclosing native signature and preceding bindings remain required.

<!-- example:glossary-storage-pointer-caller -->

</details>

### Runtime.Supervision.Poll

<a id="word-authoring-runtime-supervision-poll"></a>

<!-- word:authoring:Runtime.Supervision.Poll -->

Race the child wait against a leader-termination request; store a completed wait result or service one request and continue.

<!-- word-definition:authoring:Runtime.Supervision.Poll -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

This statement is used inside `Runtime.Supervision.Await`. Its enclosing native signature and preceding bindings remain required.

<!-- example:glossary-runtime-supervision-poll-caller -->

</details>

### Storage.Polymorphic

<a id="word-authoring-storage-polymorphic"></a>

<!-- word:authoring:Storage.Polymorphic -->

Describe a Storage.Routine with the explicitly named native type parameter `T`.

<!-- word-definition:authoring:Storage.Polymorphic -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

This statement is used inside `Storage.Identify`. Its enclosing native signature and preceding bindings remain required.

<!-- example:glossary-storage-polymorphic-caller -->

</details>

### Storage.Slot.Populate

<a id="word-authoring-storage-slot-populate"></a>

<!-- word:authoring:Storage.Slot.Populate -->

Append one empty optional value to a supplied slot vector; the traversal index is intentionally unused.

<!-- word-definition:authoring:Storage.Slot.Populate -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

This statement is used inside [Storage.Slot.Allocate](#word-authoring-storage-slot-allocate). Its enclosing native signature, preceding local bindings and selected native module are required.

<!-- example:glossary-builtins-eng-storage-program-23-23 -->

</details>

### Runtime.App.Preference

<a id="word-authoring-runtime-app-preference"></a>

<!-- word:authoring:Runtime.App.Preference -->

Borrow `Value.selected` without moving the referenced native field.

<!-- word-definition:authoring:Runtime.App.Preference -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

This statement is used inside `Runtime.App.Select`. Its enclosing native signature and preceding bindings remain required.

<!-- example:glossary-runtime-app-preference-caller -->

</details>

### Formatting.Prepend

<a id="word-authoring-formatting-prepend"></a>

<!-- word:authoring:Formatting.Prepend -->

Build an ordinary decimal string with `0.`, the required leading zero digits, and the supplied significant digits.

<!-- word-definition:authoring:Formatting.Prepend -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

This statement is used inside `Formatting.Expand`. Its enclosing native signature and preceding bindings remain required.

<!-- example:glossary-formatting-prepend-caller -->

</details>

### Runtime.Error.Presence

<a id="word-authoring-runtime-error-presence"></a>

<!-- word:authoring:Runtime.Error.Presence -->

Mutably reborrow the diagnostic separator flag in the third traversal-context component.

<!-- word-definition:authoring:Runtime.Error.Presence -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

This statement is used inside `Runtime.Error.Accumulate`. Its enclosing native signature and preceding bindings remain required.

<!-- example:glossary-runtime-error-presence-caller -->

</details>

### Storage.Presence

<a id="word-authoring-storage-presence"></a>

<!-- word:authoring:Storage.Presence -->

Check whether the selected native Option contains a payload. The selected operation is `Value.is_some()`.

<!-- word-definition:authoring:Storage.Presence -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

This statement is used inside `Runtime.Supervision.Terminate`. Its enclosing native signature and preceding bindings remain required.

<!-- example:glossary-storage-presence-caller -->

</details>

### Storage.Present

<a id="word-authoring-storage-present"></a>

<!-- word:authoring:Storage.Present -->

Wrap the supplied value in the native Some variant. The selected operation is `Some(Value)`.

<!-- word-definition:authoring:Storage.Present -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

This statement is used inside `Runtime.App.Capture`. Its enclosing native signature and preceding bindings remain required.

<!-- example:glossary-storage-present-caller -->

</details>

### IO.Print

<a id="word-stdlib-export-eng-io-io-print"></a>

<!-- word:stdlib-export:eng:io:IO.Print -->

Render a supported value and write its bytes to standard output followed by one LF newline.

<!-- word-definition:stdlib-export:eng:io:IO.Print -->

**Usage:** Import Print and invoke it on an App, as in the complete program below. Each argument is evaluated once. See [Value handling](#context-io--value-handling) for accepted operands.

**Example:** The App prints the greeting, Execute waits for it, and Show displays the captured bytes including LF.

<!-- example:basic -->

**Behavior:** Print composes rendering and a lined stdout send with its [fail-fast error policy](#context-io--errors-and-return-behavior). It makes no zero-allocation or nonblocking promise. Use [Send](#word-stdlib-export-eng-io-io-send) with [Lined](#word-stdlib-export-eng-io-io-lined) to choose the writer and handle returned failure.

### Storage.Probe

<a id="word-authoring-storage-probe"></a>

<!-- word:authoring:Storage.Probe -->

Invoke a callback with one item and context, then pass its Boolean Result to Decide with the supplied Process continuation.

<!-- word-definition:authoring:Storage.Probe -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

This statement is used inside `builtins/eng/storage/traversal.l`. Its enclosing native signature and preceding bindings remain required.

<!-- example:glossary-storage-probe-caller -->

</details>

### Process

<a id="word-builtin-process"></a>

<!-- word:builtin:process -->

Compose an ordered source body from existing instantiated definition expressions.

<!-- word-definition:builtin:process -->

**Usage and example.** A `New Process with Inputs ... for ...` supplies a branch body to [Storage.Option.Select](#word-authoring-storage-option-select). Its typed input belongs to that branch. The [selector caller](../examples/word-selection.l) demonstrates enclosing Functions and both branches.

Process composes ordered Definitions/Syllables through operation W.

A Process collects existing instantiated definition expressions. It is not an OS process handle. App invocations own those. Processing a source-defined rendering or writing operation in IO is an application-body composition, not a direct runtime launch instruction.

Structured selection and iteration bodies can declare typed Process Inputs. Lyger binds each input within its body and retains its declared native type and source dependency. An explicit local annotation lets native compilation check type and borrow compatibility. `_` requests native type inference. These bindings create no callable or closure and cannot escape to a sibling branch or later sentence.

### Runtime.Invocation.Process

<a id="word-authoring-runtime-invocation-process"></a>

<!-- word:authoring:Runtime.Invocation.Process -->

Mutably borrow `Value.pid` so the caller can update that exact native field.

<!-- word-definition:authoring:Runtime.Invocation.Process -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

This statement is used inside `Runtime.Invocation.Identify`. Its enclosing native signature and preceding bindings remain required.

<!-- example:glossary-runtime-invocation-process-caller -->

</details>

### Runtime.State.Processes

<a id="word-authoring-runtime-state-processes"></a>

<!-- word:authoring:Runtime.State.Processes -->

Borrow the RuntimeState owned-invocation map mutably, preserving the original map identity.

<!-- word-definition:authoring:Runtime.State.Processes -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

This statement is used inside `Runtime.App.Supervise`. Its enclosing native signature and preceding bindings remain required.

<!-- example:glossary-runtime-state-processes-caller -->

</details>

### Storage.Produced

<a id="word-authoring-storage-produced"></a>

<!-- word:authoring:Storage.Produced -->

Describe a Storage.Binding whose initializer is a composed native value expression, retaining that expression's selected definition and result binding.

<!-- word-definition:authoring:Storage.Produced -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

This statement is used inside `Storage.Value.Format`. Its enclosing native signature and preceding bindings remain required.

<!-- example:glossary-storage-produced-caller -->

</details>

### Storage.Layout.Program

<a id="word-authoring-storage-layout-program"></a>

<!-- word:authoring:Storage.Layout.Program -->

Declare the native `LProgram` record with fields `apps: Vec<std::sync::Arc<runtime::App>>`, `types: Vec<Vec<u64>>`, `values: std::cell::RefCell<Vec<Option<Value>>>`, `actions: std::cell::RefCell<Vec<Option<Environment>>>`.

<!-- word-definition:authoring:Storage.Layout.Program -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

This source-authoring fragment comes from `builtins/eng/storage/program.l` and retains that dictionary's surrounding imports and native binding context.

<!-- example:glossary-builtins-eng-storage-program-4-4 -->

</details>

### Storage.Program

<a id="word-authoring-storage-program"></a>

<!-- word:authoring:Storage.Program -->

Declare the native `LProgram` record with fields `apps: Vec<std::sync::Arc<runtime::App>>`, `types: Vec<Vec<u64>>`, `values: std::cell::RefCell<Vec<Option<Value>>>`, `actions: std::cell::RefCell<Vec<Option<Environment>>>`.

<!-- word-definition:authoring:Storage.Program -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

This statement is used inside `Storage.Program.Select`. Its enclosing native signature and preceding bindings remain required.

<!-- example:glossary-storage-program-apps-caller -->

</details>

### Runtime.App.Progress

<a id="word-authoring-runtime-app-progress"></a>

<!-- word:authoring:Runtime.App.Progress -->

Bind the native constant `crate::runtime::host::Checkpoint::SuccessfulStart` for this composition.

<!-- word-definition:authoring:Runtime.App.Progress -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

This statement is used inside `Runtime.App.Run`. Its enclosing native signature and preceding bindings remain required.

<!-- example:glossary-runtime-app-progress-caller -->

</details>

### Storage.Project

<a id="word-authoring-storage-project"></a>

<!-- word:authoring:Storage.Project -->

Downcast a borrowed erased value to the declared native type and return its borrowed reference, or produce a runtime failure with the supplied message.

<!-- word-definition:authoring:Storage.Project -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

This statement is used inside `Storage.Value.Restore`. Its enclosing native signature and preceding bindings remain required.

<!-- example:glossary-storage-project-caller -->

</details>

### Storage.Projection

<a id="word-authoring-storage-projection"></a>

<!-- word:authoring:Storage.Projection -->

Attempt a borrowed downcast from erased Any storage to the native type selected by the enclosing composition. The selected operation is `Value.downcast_ref()`.

<!-- word-definition:authoring:Storage.Projection -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

This statement is used inside `Storage.Project`. Its enclosing native signature and preceding bindings remain required.

<!-- example:glossary-storage-projection-caller -->

</details>

### Storage.Value.Protect

<a id="word-authoring-storage-value-protect"></a>

<!-- word:authoring:Storage.Value.Protect -->

Search a value with a fresh visited set and reject it with the supplied message if it reaches the target ownership identity.

<!-- word-definition:authoring:Storage.Value.Protect -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

This statement is used inside `Storage.Pointer.Update`. Its enclosing native signature and preceding bindings remain required.

<!-- example:glossary-storage-value-protect-caller -->

</details>

### Public

<a id="word-builtin-public"></a>

<!-- word:builtin:public -->

Publish exactly the declaration qualified by this visibility adjective from its private paragraph into the source-unit root.

<!-- word-definition:builtin:public -->

**Usage and example.** `Create "Shared" as a New Public Dictionary` in the [public-values program](../../demo/paragraphs/public-values.l) exposes Shared. Publication retains identity and does not export unrelated declarations or private members. Use [Publicly](#word-builtin-publicly) when the modifier attaches to the declaration action.

### Runtime.Invocation.Publication

<a id="word-authoring-runtime-invocation-publication"></a>

<!-- word:authoring:Runtime.Invocation.Publication -->

Bind the native constant `crate::runtime::host::Checkpoint::PrePublication` for this composition.

<!-- word-definition:authoring:Runtime.Invocation.Publication -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

This statement is used inside `Runtime.Invocation.Publish`. Its enclosing native signature and preceding bindings remain required.

<!-- example:glossary-runtime-invocation-caller -->

</details>

### Publicly

<a id="word-builtin-publicly"></a>

<!-- word:builtin:publicly -->

Publish exactly the declaration selected by the action modified by this adverb.

<!-- word-definition:builtin:publicly -->

**Usage and example.** `Publicly define "Count" as 7` in the [public-values program](../../demo/paragraphs/public-values.l) makes Count available after its paragraph. `Publicly create Identity` in the [Function example](../examples/composed-functions.l) exposes the Function, not its input or body locals. This adverb is distinct from the [Public adjective](#word-builtin-public).

### Runtime.Invocation.Publish

<a id="word-authoring-runtime-invocation-publish"></a>

<!-- word:authoring:Runtime.Invocation.Publish -->

Run the result-publication checkpoint, store the finalized Snapshot under the invocation lock, then broadcast its completion flag.

<!-- word-definition:authoring:Runtime.Invocation.Publish -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

This statement is used inside `Runtime.Supervision.Complete`. Its enclosing native signature and preceding bindings remain required.

<!-- example:glossary-runtime-invocation-publish-caller -->

</details>

### Storage.Publish

<a id="word-authoring-storage-publish"></a>

<!-- word:authoring:Storage.Publish -->

Assign Some of the supplied value to the selected existing native binding.

<!-- word-definition:authoring:Storage.Publish -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

**Compiler use.** Assign the initialized captured callable value into its optional native storage binding. This retained compiler excerpt selects the L-owned definition with typed native name/type/value or Process bindings. Those bindings come from the compiler’s resolved plan; this is an integration fragment, not an ordinary standalone L call.

<!-- example:glossary-lyger-storage-words-56-58 -->

</details>

### Runtime.Error.Punctuation

<a id="word-authoring-runtime-error-punctuation"></a>

<!-- word:authoring:Runtime.Error.Punctuation -->

Bind the exact colon-space byte slice used when attaching stderr.

<!-- word-definition:authoring:Runtime.Error.Punctuation -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

This statement is used inside `Runtime.Error.Attach`. Its enclosing native signature and preceding bindings remain required.

<!-- example:glossary-runtime-error-punctuation-caller -->

</details>

### Runtime.Push

<a id="word-authoring-runtime-push"></a>

<!-- word:authoring:Runtime.Push -->

Append the supplied value to the selected native vector. The selected operation is `Values.push(Value)`.

<!-- word-definition:authoring:Runtime.Push -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

This statement is used inside `Runtime.Invocation.Settle`. Its enclosing native signature and preceding bindings remain required.

<!-- example:glossary-runtime-push-caller -->

</details>

### Put

<a id="word-builtin-put"></a>

<!-- word:builtin:put -->

Insert a value into a supported runtime destination with its required element or target type.

<!-- word-definition:builtin:put -->

**Usage and example.** `Put 7 in items` appends to a List. The [value program](../examples/word-values.l) also supplies a Map key and updates a Pointer target. `7 is Put in items` uses the same checked List update. Passive Put also accepts the existing Map-key phrase. Pointer updates remain imperative. Put does not mutate vocabulary Dictionaries.

## Q

<a id="glossary-q"></a>

### Runtime.Quell

<a id="word-authoring-runtime-quell"></a>

<!-- word:authoring:Runtime.Quell -->

Take the runtime's signal-watcher task if present, abort it, and await its termination.

<!-- word-definition:authoring:Runtime.Quell -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

This statement is used inside `Runtime.Finish`. Its enclosing native signature and preceding bindings remain required.

<!-- example:glossary-runtime-quell-caller -->

</details>

## R

<a id="glossary-r"></a>

### Runtime.Race

<a id="word-authoring-runtime-race"></a>

<!-- word:authoring:Runtime.Race -->

Create the host future that completes with whichever of two supplied futures becomes ready first. The selected operation is `crate::runtime::host::race(First, Second)`.

<!-- word-definition:authoring:Runtime.Race -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

This statement is used inside `Runtime.Supervision.Poll`. Its enclosing native signature and preceding bindings remain required.

<!-- example:glossary-runtime-race-caller -->

</details>

### Storage.Range

<a id="word-authoring-storage-range"></a>

<!-- word:authoring:Storage.Range -->

Construct the half-open native usize range from zero to the supplied limit.

<!-- word-definition:authoring:Storage.Range -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

This statement is used inside [Storage.Slot.Allocate](#word-authoring-storage-slot-allocate). Its enclosing native signature, preceding local bindings and selected native module are required.

<!-- example:glossary-builtins-eng-storage-program-21-21 -->

</details>

### Storage.Environment.Reach

<a id="word-authoring-storage-environment-reach"></a>

<!-- word:authoring:Storage.Environment.Reach -->

Check whether an environment or one of its captured values reaches a target ownership identity. Track visited storage identities to avoid revisiting cycles and propagate borrow failures.

<!-- word-definition:authoring:Storage.Environment.Reach -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

This statement is used inside `Storage.Environment.Visit`. Its enclosing native signature and preceding bindings remain required.

<!-- example:glossary-storage-environment-reach-caller -->

</details>

### Storage.Function.Reach

<a id="word-authoring-storage-function-reach"></a>

<!-- word:authoring:Storage.Function.Reach -->

Project an erased payload to the declared typed Function pair and search its capture environment for the target ownership identity.

<!-- word-definition:authoring:Storage.Function.Reach -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

The wrapper below imports `Storage` and `Runtime` and is a member of the example `GlossaryExamples` Dictionary. Its declared native inputs, outputs, lifetime/generic descriptors, and the original selected module's native types/support are required. Lyger resolves the source and composes its library AML; this fragment has no native compilation or execution claim.

<!-- example:glossary-native-storage-function-reach -->

</details>

### Storage.List.Reach

<a id="word-authoring-storage-list-reach"></a>

<!-- word:authoring:Storage.List.Reach -->

Project an erased payload to a List and search its elements for a target ownership identity, preserving the shared traversal history.

<!-- word-definition:authoring:Storage.List.Reach -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

The wrapper below imports `Storage` and `Runtime` and is a member of the example `GlossaryExamples` Dictionary. Its declared native inputs, outputs, lifetime/generic descriptors, and the original selected module's native types/support are required. Lyger resolves the source and composes its library AML; this fragment has no native compilation or execution claim.

<!-- example:glossary-native-storage-list-reach -->

</details>

### Storage.Map.Reach

<a id="word-authoring-storage-map-reach"></a>

<!-- word:authoring:Storage.Map.Reach -->

Project an erased payload to a Map and search its values for a target ownership identity.

<!-- word-definition:authoring:Storage.Map.Reach -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

The wrapper below imports `Storage` and `Runtime` and is a member of the example `GlossaryExamples` Dictionary. Its declared native inputs, outputs, lifetime/generic descriptors, and the original selected module's native types/support are required. Lyger resolves the source and composes its library AML; this fragment has no native compilation or execution claim.

<!-- example:glossary-native-storage-map-reach -->

</details>

### Storage.Pointer.Reach

<a id="word-authoring-storage-pointer-reach"></a>

<!-- word:authoring:Storage.Pointer.Reach -->

Check a Pointer cell's identity and recursively follow its stored value once per visited cell, propagating conflicting borrows.

<!-- word-definition:authoring:Storage.Pointer.Reach -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

This statement is used inside `Storage.Pointer.Visit`. Its enclosing native signature and preceding bindings remain required.

<!-- example:glossary-storage-pointer-reach-caller -->

</details>

### Storage.Value.Reach

<a id="word-authoring-storage-value-reach"></a>

<!-- word:authoring:Storage.Value.Reach -->

Invoke a stored value's typed ownership visitor with its borrowed erased payload, target identity and shared visited set.

<!-- word-definition:authoring:Storage.Value.Reach -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

This statement is used inside `Storage.Pointer.Reach`. Its enclosing native signature and preceding bindings remain required.

<!-- example:glossary-storage-value-reach-caller -->

</details>

### Runtime.App.Read

<a id="word-authoring-runtime-app-read"></a>

<!-- word:authoring:Runtime.App.Read -->

Return retained captured stdout bytes from an App's completed invocation, or report misuse when capture is not yet available.

<!-- word-definition:authoring:Runtime.App.Read -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

The wrapper below imports `Storage` and `Runtime` and is a member of the example `GlossaryExamples` Dictionary. Its declared native inputs, outputs, lifetime/generic descriptors, and the original selected module's native types/support are required. Lyger resolves the source and composes its library AML; this fragment has no native compilation or execution claim.

<!-- example:glossary-native-runtime-app-read -->

</details>

### Runtime.Capture.Read

<a id="word-authoring-runtime-capture-read"></a>

<!-- word:authoring:Runtime.Capture.Read -->

Adapt an optional concrete reader and return the capture future using the supplied deadline, host dependencies, and stream identity.

<!-- word-definition:authoring:Runtime.Capture.Read -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

This statement is used inside `Runtime.Capture.Follow`. Its enclosing native signature and preceding bindings remain required.

<!-- example:glossary-runtime-capture-read-caller -->

</details>

### Runtime.Error.Read

<a id="word-authoring-runtime-error-read"></a>

<!-- word:authoring:Runtime.Error.Read -->

Borrow the raw message bytes of an Error without rendering child causes.

<!-- word-definition:authoring:Runtime.Error.Read -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

This statement is used inside `Runtime.Error.Append`. Its enclosing native signature and preceding bindings remain required.

<!-- example:glossary-runtime-error-read-caller -->

</details>

### Runtime.Invocation.Read

<a id="word-authoring-runtime-invocation-read"></a>

<!-- word:authoring:Runtime.Invocation.Read -->

Lock invocation state and copy its optional finalized Snapshot.

<!-- word-definition:authoring:Runtime.Invocation.Read -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

This statement is used inside `Runtime.App.Capture`. Its enclosing native signature and preceding bindings remain required.

<!-- example:glossary-runtime-invocation-read-caller -->

</details>

### Runtime.Read

<a id="word-authoring-runtime-read"></a>

<!-- word:authoring:Runtime.Read -->

Read the runtime's optional cached final exit status under its state lock.

<!-- word-definition:authoring:Runtime.Read -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

This statement is used inside `Runtime.Finish`. Its enclosing native signature and preceding bindings remain required.

<!-- example:glossary-runtime-read-caller -->

</details>

### Runtime.Supervision.Read

<a id="word-authoring-runtime-supervision-read"></a>

<!-- word:authoring:Runtime.Supervision.Read -->

Try to receive a queued termination request without waiting. The selected operation is `Value.try_recv()`.

<!-- word-definition:authoring:Runtime.Supervision.Read -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

This statement is used inside `Runtime.Supervision.Release`. Its enclosing native signature and preceding bindings remain required.

<!-- example:glossary-runtime-supervision-read-caller -->

</details>

### Storage.Bytes.Read

<a id="word-authoring-storage-bytes-read"></a>

<!-- word:authoring:Storage.Bytes.Read -->

Validate a numeric byte index, read that byte with a bounds check, and return its value as a Number. Invalid indices and out-of-bounds reads return runtime failures.

<!-- word-definition:authoring:Storage.Bytes.Read -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

The wrapper below imports `Storage` and `Runtime` and is a member of the example `GlossaryExamples` Dictionary. Its declared native inputs, outputs, lifetime/generic descriptors, and the original selected module's native types/support are required. Lyger resolves the source and composes its library AML; this fragment has no native compilation or execution claim.

<!-- example:glossary-native-storage-bytes-read -->

</details>

### Storage.Context.Read

<a id="word-authoring-storage-context-read"></a>

<!-- word:authoring:Storage.Context.Read -->

Borrow the runtime and program components of the supplied context into independently supplied result labels.

<!-- word-definition:authoring:Storage.Context.Read -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

**Compiler use.** Project the typed context supplied to a generated callable Function. This retained compiler excerpt selects the L-owned definition with typed native name/type/value or Process bindings. Those bindings come from the compiler’s resolved plan; this is an integration fragment, not an ordinary standalone L call.

<!-- example:glossary-lyger-function-composition-333-337 -->

</details>

### Storage.Environment.Read

<a id="word-authoring-storage-environment-read"></a>

<!-- word:authoring:Storage.Environment.Read -->

Borrow a capture environment, check its numeric slot, and return a value snapshot through Value.Copy. Borrow conflicts and missing slots return runtime failures.

<!-- word-definition:authoring:Storage.Environment.Read -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

The wrapper below imports `Storage` and `Runtime` and is a member of the example `GlossaryExamples` Dictionary. Its declared native inputs, outputs, lifetime/generic descriptors, and the original selected module's native types/support are required. Lyger resolves the source and composes its library AML; this fragment has no native compilation or execution claim.

<!-- example:glossary-native-storage-environment-read -->

</details>

### Storage.List.Read

<a id="word-authoring-storage-list-read"></a>

<!-- word:authoring:Storage.List.Read -->

Validate a numeric List index, check bounds, and return Value.Copy of the selected element.

<!-- word-definition:authoring:Storage.List.Read -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

The wrapper below imports `Storage` and `Runtime` and is a member of the example `GlossaryExamples` Dictionary. Its declared native inputs, outputs, lifetime/generic descriptors, and the original selected module's native types/support are required. Lyger resolves the source and composes its library AML; this fragment has no native compilation or execution claim.

<!-- example:glossary-native-storage-list-read -->

</details>

### Storage.Map.Read

<a id="word-authoring-storage-map-read"></a>

<!-- word:authoring:Storage.Map.Read -->

Find a Map value by its exact byte key and return Value.Copy of that value, or return a missing-key failure.

<!-- word-definition:authoring:Storage.Map.Read -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

The wrapper below imports `Storage` and `Runtime` and is a member of the example `GlossaryExamples` Dictionary. Its declared native inputs, outputs, lifetime/generic descriptors, and the original selected module's native types/support are required. Lyger resolves the source and composes its library AML; this fragment has no native compilation or execution claim.

<!-- example:glossary-native-storage-map-read -->

</details>

### Storage.Pointer.Read

<a id="word-authoring-storage-pointer-read"></a>

<!-- word:authoring:Storage.Pointer.Read -->

Borrow a Pointer's target cell and return Value.Copy of the stored value. A conflicting borrow returns a runtime failure.

<!-- word-definition:authoring:Storage.Pointer.Read -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

This statement is used inside `Storage.Pointer.Render`. Its enclosing native signature and preceding bindings remain required.

<!-- example:glossary-storage-pointer-read-caller -->

</details>

### Storage.Program.Actions.Read

<a id="word-authoring-storage-program-actions-read"></a>

<!-- word:authoring:Storage.Program.Actions.Read -->

Read an initialized action-capture slot from a program and retain the selected environment, with distinct failures for borrow conflicts and uninitialized slots.

<!-- word-definition:authoring:Storage.Program.Actions.Read -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

The wrapper below imports `Storage` and `Runtime` and is a member of the example `GlossaryExamples` Dictionary. Its declared native inputs, outputs, lifetime/generic descriptors, and the original selected module's native types/support are required. Lyger resolves the source and composes its library AML; this fragment has no native compilation or execution claim.

<!-- example:glossary-native-storage-program-actions-read -->

</details>

### Storage.Program.Values.Read

<a id="word-authoring-storage-program-values-read"></a>

<!-- word:authoring:Storage.Program.Values.Read -->

Read an initialized program-value slot and return Value.Copy of its contents, preserving borrow and missing-initialization failures.

<!-- word-definition:authoring:Storage.Program.Values.Read -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

The wrapper below imports `Storage` and `Runtime` and is a member of the example `GlossaryExamples` Dictionary. Its declared native inputs, outputs, lifetime/generic descriptors, and the original selected module's native types/support are required. Lyger resolves the source and composes its library AML; this fragment has no native compilation or execution claim.

<!-- example:glossary-native-storage-program-values-read -->

</details>

### Storage.Read

<a id="word-authoring-storage-read"></a>

<!-- word:authoring:Storage.Read -->

Attempt an immutable RefCell borrow and return its guard or native borrow error. The selected operation is `Value.try_borrow()`.

<!-- word-definition:authoring:Storage.Read -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

This statement is used inside `Storage.Pointer.Reach`. Its enclosing native signature and preceding bindings remain required.

<!-- example:glossary-storage-read-caller -->

</details>

### Storage.Slot.Read

<a id="word-authoring-storage-slot-read"></a>

<!-- word:authoring:Storage.Slot.Read -->

Borrow a slot vector, require the selected slot and its initialized payload, then invoke the supplied typed copying Function. Use the supplied messages for borrow and absence failures.

<!-- word-definition:authoring:Storage.Slot.Read -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

This statement is used inside [Storage.Program.Values.Read](#word-authoring-storage-program-values-read). Its enclosing native signature, preceding local bindings and selected native module are required.

<!-- example:glossary-builtins-eng-storage-program-54-54 -->

</details>

### Runtime.Capture.Readable

<a id="word-authoring-runtime-capture-readable"></a>

<!-- word:authoring:Runtime.Capture.Readable -->

Describe a Storage.Routine with `R: tokio::io::AsyncRead + Unpin + Send + 'static` for owned capture-reader tasks.

<!-- word-definition:authoring:Runtime.Capture.Readable -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

This statement is used inside `Runtime.Capture.Adapt`. Its enclosing native signature and preceding bindings remain required.

<!-- example:glossary-runtime-capture-readable-caller -->

</details>

### Runtime.Reap

<a id="word-authoring-runtime-reap"></a>

<!-- word:authoring:Runtime.Reap -->

Settle every captured owned invocation, await their completion, collect cleanup errors, stop signal observation, and publish the shutdown result.

<!-- word-definition:authoring:Runtime.Reap -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

This statement is used inside `Runtime.Begin`. Its enclosing native signature and preceding bindings remain required.

<!-- example:glossary-runtime-reap-caller -->

</details>

### Runtime.Capture.Receive

<a id="word-authoring-runtime-capture-receive"></a>

<!-- word:authoring:Runtime.Capture.Receive -->

Handle one read result: stop at EOF, store a read/write failure, or write exactly the received byte prefix and continue.

<!-- word-definition:authoring:Runtime.Capture.Receive -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

This statement is used inside `Runtime.Capture.Advance`. Its enclosing native signature and preceding bindings remain required.

<!-- example:glossary-runtime-capture-receive-caller -->

</details>

### Runtime.Receive

<a id="word-authoring-runtime-receive"></a>

<!-- word:authoring:Runtime.Receive -->

Create a receive operation on the selected native event or request receiver. The selected operation is `Value.recv()`.

<!-- word-definition:authoring:Runtime.Receive -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

This statement is used inside `Runtime.Supervision.Poll`. Its enclosing native signature and preceding bindings remain required.

<!-- example:glossary-runtime-receive-caller -->

</details>

### Runtime.Invocation.Record

<a id="word-authoring-runtime-invocation-record"></a>

<!-- word:authoring:Runtime.Invocation.Record -->

If its runtime owner remains available, record an invocation cleanup Error under that invocation's sequence number.

<!-- word-definition:authoring:Runtime.Invocation.Record -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

This statement is used inside `Runtime.Invocation.Settle`. Its enclosing native signature and preceding bindings remain required.

<!-- example:glossary-runtime-invocation-record-caller -->

</details>

### Runtime.Terminal.Record

<a id="word-authoring-runtime-terminal-record"></a>

<!-- word:authoring:Runtime.Terminal.Record -->

Notify the selected host dependencies that the supplied terminal request has been entered. The selected operation is `Value.inner.dependencies.terminal_entered(Request)`.

<!-- word-definition:authoring:Runtime.Terminal.Record -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

This statement is used inside `Runtime.Refuse`. Its enclosing native signature and preceding bindings remain required.

<!-- example:glossary-runtime-terminal-record-caller -->

</details>

### Formatting.Reduce

<a id="word-authoring-formatting-reduce"></a>

<!-- word:authoring:Formatting.Reduce -->

Recursively divide a decimal coefficient by ten while it is divisible by ten, incrementing the exponent to preserve the represented value.

<!-- word-definition:authoring:Formatting.Reduce -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

This statement is used inside `Formatting.Render`. Its enclosing native signature and preceding bindings remain required.

<!-- example:glossary-formatting-reduce-caller -->

</details>

### Storage.Refer

<a id="word-authoring-storage-refer"></a>

<!-- word:authoring:Storage.Refer -->

Obtain the selected value's native borrowed view, preserving any optional wrapper provided by that method. The selected operation is `Value.as_ref()`.

<!-- word-definition:authoring:Storage.Refer -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

This statement is used inside `Storage.Value.Restore`. Its enclosing native signature and preceding bindings remain required.

<!-- example:glossary-storage-refer-caller -->

</details>

### Runtime.Reference

<a id="word-authoring-runtime-reference"></a>

<!-- word:authoring:Runtime.Reference -->

Borrow `Value.inner` without moving the referenced native field.

<!-- word-definition:authoring:Runtime.Reference -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

This statement is used inside `Runtime.App.Wait`. Its enclosing native signature and preceding bindings remain required.

<!-- example:glossary-runtime-reference-caller -->

</details>

### Storage.Reference

<a id="word-authoring-storage-reference"></a>

<!-- word:authoring:Storage.Reference -->

Create an immutable native reborrow through dereferencing the supplied value.

<!-- word-definition:authoring:Storage.Reference -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

This statement is used inside `Storage.Pointer.Read`. Its enclosing native signature and preceding bindings remain required.

<!-- example:glossary-storage-reference-caller -->

</details>

### Storage.Referenced

<a id="word-authoring-storage-referenced"></a>

<!-- word:authoring:Storage.Referenced -->

Describe a lending Storage.Routine with lifetime `'a` and native type bound `T: 'static`.

<!-- word-definition:authoring:Storage.Referenced -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

This statement is used inside `Storage.Project`. Its enclosing native signature and preceding bindings remain required.

<!-- example:glossary-storage-referenced-caller -->

</details>

### Runtime.Refuse

<a id="word-authoring-runtime-refuse"></a>

<!-- word:authoring:Runtime.Refuse -->

Record a fatal terminal request, request exit status one with a retained copy of the supplied Error, and return the original Error.

<!-- word-definition:authoring:Runtime.Refuse -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

This statement is used inside `Runtime.Fail`. Its enclosing native signature and preceding bindings remain required.

<!-- example:glossary-runtime-refuse-caller -->

</details>

### Runtime.Capture.Reject

<a id="word-authoring-runtime-capture-reject"></a>

<!-- word:authoring:Runtime.Capture.Reject -->

Construct a Capture containing an empty byte vector and the supplied Error.

<!-- word-definition:authoring:Runtime.Capture.Reject -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

This statement is used inside `Runtime.Capture.Start`. Its enclosing native signature and preceding bindings remain required.

<!-- example:glossary-runtime-capture-reject-caller -->

</details>

### Runtime.Invocation.Reject

<a id="word-authoring-runtime-invocation-reject"></a>

<!-- word:authoring:Runtime.Invocation.Reject -->

Construct a failure Snapshot with empty streams and publish it as the invocation result.

<!-- word-definition:authoring:Runtime.Invocation.Reject -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

This statement is used inside `Runtime.App.Run`. Its enclosing native signature and preceding bindings remain required.

<!-- example:glossary-runtime-invocation-reject-caller -->

</details>

### Runtime.Reject

<a id="word-authoring-runtime-reject"></a>

<!-- word:authoring:Runtime.Reject -->

Extract the error from a Result already known to have failed, panicking through the explicit expectation if supplied a success.

<!-- word-definition:authoring:Runtime.Reject -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

This statement is used inside `Runtime.Normalize`. Its enclosing native signature and preceding bindings remain required.

<!-- example:glossary-runtime-reject-caller -->

</details>

### Runtime.Snapshot.Reject

<a id="word-authoring-runtime-snapshot-reject"></a>

<!-- word:authoring:Runtime.Snapshot.Reject -->

Construct a Snapshot with empty stdout and stderr, the supplied Error, and no process exit status.

<!-- word-definition:authoring:Runtime.Snapshot.Reject -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

This statement is used inside `Runtime.Invocation.Reject`. Its enclosing native signature and preceding bindings remain required.

<!-- example:glossary-runtime-snapshot-reject-caller -->

</details>

### Storage.Reject

<a id="word-authoring-storage-reject"></a>

<!-- word:authoring:Storage.Reject -->

Construct a runtime Error from a Text message, wrap it in rejection, and immediately return it from the enclosing Function.

<!-- word-definition:authoring:Storage.Reject -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

This statement is used inside `Storage.Index.Check`. Its enclosing native signature and preceding bindings remain required.

<!-- example:glossary-storage-reject-caller -->

</details>

### Storage.Rejection

<a id="word-authoring-storage-rejection"></a>

<!-- word:authoring:Storage.Rejection -->

Construct native Err around the supplied error payload without returning from the enclosing Function.

<!-- word-definition:authoring:Storage.Rejection -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

This statement is used inside `Runtime.Normalize`. Its enclosing native signature and preceding bindings remain required.

<!-- example:glossary-storage-rejection-caller -->

</details>

### Runtime.Path.Relative

<a id="word-authoring-runtime-path-relative"></a>

<!-- word:authoring:Runtime.Path.Relative -->

Match a native path component and return true for Normal or CurDir, false for every other component.

<!-- word-definition:authoring:Runtime.Path.Relative -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

This statement is used inside `Runtime.Path.Admit`. Its enclosing native signature and preceding bindings remain required.

<!-- example:glossary-runtime-path-caller -->

</details>

### Runtime.App.Release

<a id="word-authoring-runtime-app-release"></a>

<!-- word:authoring:Runtime.App.Release -->

Clear an App's outstanding flag only when its current invocation is the supplied invocation.

<!-- word-definition:authoring:Runtime.App.Release -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

This statement is used inside `Runtime.Invocation.Finish`. Its enclosing native signature and preceding bindings remain required.

<!-- example:glossary-runtime-app-release-caller -->

</details>

### Runtime.Supervision.Release

<a id="word-authoring-runtime-supervision-release"></a>

<!-- word:authoring:Runtime.Supervision.Release -->

Try to receive one queued termination request, acknowledge it with success, and report whether the queue should stop draining.

<!-- word-definition:authoring:Runtime.Supervision.Release -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

This statement is used inside `Runtime.Supervision.Await`. Its enclosing native signature and preceding bindings remain required.

<!-- example:glossary-runtime-halt-caller -->

</details>

### Storage.Release

<a id="word-authoring-storage-release"></a>

<!-- word:authoring:Storage.Release -->

Consume the supplied value through native drop. Releasing a shared handle need not destroy its payload and does not imply flush, durability, cancellation or App completion.

<!-- word-definition:authoring:Storage.Release -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

This statement is used inside `Runtime.App.Describe`. Its enclosing native signature and preceding bindings remain required.

<!-- example:glossary-storage-release-caller -->

</details>

### Runtime.Error.Remember

<a id="word-authoring-runtime-error-remember"></a>

<!-- word:authoring:Runtime.Error.Remember -->

Insert an Error's identity and recursively visit its children in the supplied identity set.

<!-- word-definition:authoring:Runtime.Error.Remember -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

This statement is used inside `Runtime.Error.Append`. Its enclosing native signature and preceding bindings remain required.

<!-- example:glossary-runtime-error-remember-caller -->

</details>

### Runtime.Remember

<a id="word-authoring-runtime-remember"></a>

<!-- word:authoring:Runtime.Remember -->

Store a final exit status under the runtime state lock and return that status.

<!-- word-definition:authoring:Runtime.Remember -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

This statement is used inside `Runtime.Finish`. Its enclosing native signature and preceding bindings remain required.

<!-- example:glossary-runtime-remember-caller -->

</details>

### Runtime.State.Remember

<a id="word-authoring-runtime-state-remember"></a>

<!-- word:authoring:Runtime.State.Remember -->

Mutably borrow `Value.final_status` so the caller can update that exact native field.

<!-- word-definition:authoring:Runtime.State.Remember -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

This statement is used inside `Runtime.Remember`. Its enclosing native signature and preceding bindings remain required.

<!-- example:glossary-runtime-state-remember-caller -->

</details>

### Storage.Remember

<a id="word-authoring-storage-remember"></a>

<!-- word:authoring:Storage.Remember -->

Insert the supplied value or key/value pair into the selected native collection and return that collection's insertion result. The selected operation is `Collection.insert(Value)`.

<!-- word-definition:authoring:Storage.Remember -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

This statement is used inside `Storage.Pointer.Reach`. Its enclosing native signature and preceding bindings remain required.

<!-- example:glossary-storage-remember-caller -->

</details>

### Runtime.Supervision.Remove

<a id="word-authoring-runtime-supervision-remove"></a>

<!-- word:authoring:Runtime.Supervision.Remove -->

Remove the supplied invocation sequence from the selected owned-invocation map and discard the removed value. The selected operation is `Values.remove(Sequence)`.

<!-- word-definition:authoring:Runtime.Supervision.Remove -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

This statement is used inside `Runtime.Supervision.Complete`. Its enclosing native signature and preceding bindings remain required.

<!-- example:glossary-runtime-supervision-remove-caller -->

</details>

### Formatting.Render

<a id="word-authoring-formatting-render"></a>

<!-- word:authoring:Formatting.Render -->

Parse a native scientific spelling, adjust an exact midpoint when required, remove redundant coefficient zeros, select ordinary or scientific notation, and restore the original sign before encoding bytes.

<!-- word-definition:authoring:Formatting.Render -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

This statement is used inside `Formatting.Format`. Its enclosing native signature and preceding bindings remain required.

<!-- example:glossary-formatting-render-caller -->

</details>

### Runtime.Error.Render

<a id="word-authoring-runtime-error-render"></a>

<!-- word:authoring:Runtime.Error.Render -->

Render an Error graph into a fresh byte vector, tracking identities to avoid duplicate diagnostic lines.

<!-- word-definition:authoring:Runtime.Error.Render -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

This statement is used inside `Runtime.Error.Attach`. Its enclosing native signature and preceding bindings remain required.

<!-- example:glossary-runtime-error-render-caller -->

</details>

### Storage.App.Render

<a id="word-authoring-storage-app-render"></a>

<!-- word:authoring:Storage.App.Render -->

Return the literal bytes `App` as a successful display result. This renderer does not inspect the supplied erased value or execute an App.

<!-- word-definition:authoring:Storage.App.Render -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

The wrapper below imports `Storage` and `Runtime` and is a member of the example `GlossaryExamples` Dictionary. Its declared native inputs, outputs, lifetime/generic descriptors, and the original selected module's native types/support are required. Lyger resolves the source and composes its library AML; this fragment has no native compilation or execution claim.

<!-- example:glossary-native-storage-app-render -->

</details>

### Storage.Bytes.Render

<a id="word-authoring-storage-bytes-render"></a>

<!-- word:authoring:Storage.Bytes.Render -->

Require the erased payload to contain a byte vector and return a copied vector, using `expected Bytes` for a type mismatch.

<!-- word-definition:authoring:Storage.Bytes.Render -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

The wrapper below imports `Storage` and `Runtime` and is a member of the example `GlossaryExamples` Dictionary. Its declared native inputs, outputs, lifetime/generic descriptors, and the original selected module's native types/support are required. Lyger resolves the source and composes its library AML; this fragment has no native compilation or execution claim.

<!-- example:glossary-native-storage-bytes-render -->

</details>

### Storage.Environment.Render

<a id="word-authoring-storage-environment-render"></a>

<!-- word:authoring:Storage.Environment.Render -->

Return the runtime failure `capture environment is not an L value`; capture environments have no ordinary display representation.

<!-- word-definition:authoring:Storage.Environment.Render -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

The wrapper below imports `Storage` and `Runtime` and is a member of the example `GlossaryExamples` Dictionary. Its declared native inputs, outputs, lifetime/generic descriptors, and the original selected module's native types/support are required. Lyger resolves the source and composes its library AML; this fragment has no native compilation or execution claim.

<!-- example:glossary-native-storage-environment-render -->

</details>

### Storage.Error.Render

<a id="word-authoring-storage-error-render"></a>

<!-- word:authoring:Storage.Error.Render -->

Project an erased payload to the runtime Error type and return the bytes of its native display string, propagating a type mismatch as a runtime failure.

<!-- word-definition:authoring:Storage.Error.Render -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

The wrapper below imports `Storage` and `Runtime` and is a member of the example `GlossaryExamples` Dictionary. Its declared native inputs, outputs, lifetime/generic descriptors, and the original selected module's native types/support are required. Lyger resolves the source and composes its library AML; this fragment has no native compilation or execution claim.

<!-- example:glossary-native-storage-error-render -->

</details>

### Storage.Function.Render

<a id="word-authoring-storage-function-render"></a>

<!-- word:authoring:Storage.Function.Render -->

Return the fixed display text for a Function without inspecting or invoking the erased input.

<!-- word-definition:authoring:Storage.Function.Render -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

The wrapper below imports `Storage` and `Runtime` and is a member of the example `GlossaryExamples` Dictionary. Its declared native inputs, outputs, lifetime/generic descriptors, and the original selected module's native types/support are required. Lyger resolves the source and composes its library AML; this fragment has no native compilation or execution claim.

<!-- example:glossary-native-storage-function-render -->

</details>

### Storage.List.Render

<a id="word-authoring-storage-list-render"></a>

<!-- word:authoring:Storage.List.Render -->

Project an erased List, render its elements in order, join them with comma-space bytes, and enclose the result in square brackets. Stop on the first rendering failure.

<!-- word-definition:authoring:Storage.List.Render -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

The wrapper below imports `Storage` and `Runtime` and is a member of the example `GlossaryExamples` Dictionary. Its declared native inputs, outputs, lifetime/generic descriptors, and the original selected module's native types/support are required. Lyger resolves the source and composes its library AML; this fragment has no native compilation or execution claim.

<!-- example:glossary-native-storage-list-render -->

</details>

### Storage.Map.Render

<a id="word-authoring-storage-map-render"></a>

<!-- word:authoring:Storage.Map.Render -->

Project an erased Map, render its entries in map order, join them with comma-space bytes, and enclose them in braces. Rendering failures propagate.

<!-- word-definition:authoring:Storage.Map.Render -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

The wrapper below imports `Storage` and `Runtime` and is a member of the example `GlossaryExamples` Dictionary. Its declared native inputs, outputs, lifetime/generic descriptors, and the original selected module's native types/support are required. Lyger resolves the source and composes its library AML; this fragment has no native compilation or execution claim.

<!-- example:glossary-native-storage-map-render -->

</details>

### Storage.Null.Render

<a id="word-authoring-storage-null-render"></a>

<!-- word:authoring:Storage.Null.Render -->

Return the literal bytes `null` as a successful display result. This differs from IO.Print's `<nil>` spelling.

<!-- word-definition:authoring:Storage.Null.Render -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

The wrapper below imports `Storage` and `Runtime` and is a member of the example `GlossaryExamples` Dictionary. Its declared native inputs, outputs, lifetime/generic descriptors, and the original selected module's native types/support are required. Lyger resolves the source and composes its library AML; this fragment has no native compilation or execution claim.

<!-- example:glossary-native-storage-null-render -->

</details>

### Storage.Number.Render

<a id="word-authoring-storage-number-render"></a>

<!-- word:authoring:Storage.Number.Render -->

Project an erased payload to `f64` and return the canonical runtime Number-formatting bytes, or return `expected Number` for a type mismatch.

<!-- word-definition:authoring:Storage.Number.Render -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

The wrapper below imports `Storage` and `Runtime` and is a member of the example `GlossaryExamples` Dictionary. Its declared native inputs, outputs, lifetime/generic descriptors, and the original selected module's native types/support are required. Lyger resolves the source and composes its library AML; this fragment has no native compilation or execution claim.

<!-- example:glossary-native-storage-number-render -->

</details>

### Storage.Pointer.Render

<a id="word-authoring-storage-pointer-render"></a>

<!-- word:authoring:Storage.Pointer.Render -->

Project an erased Pointer, read its target value, and render that value. Projection, borrow and rendering failures propagate.

<!-- word-definition:authoring:Storage.Pointer.Render -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

The wrapper below imports `Storage` and `Runtime` and is a member of the example `GlossaryExamples` Dictionary. Its declared native inputs, outputs, lifetime/generic descriptors, and the original selected module's native types/support are required. Lyger resolves the source and composes its library AML; this fragment has no native compilation or execution claim.

<!-- example:glossary-native-storage-pointer-render -->

</details>

### Storage.Text.Render

<a id="word-authoring-storage-text-render"></a>

<!-- word:authoring:Storage.Text.Render -->

Require an erased byte-vector Text payload and return a copied vector, using `expected Text` for a type mismatch.

<!-- word-definition:authoring:Storage.Text.Render -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

The wrapper below imports `Storage` and `Runtime` and is a member of the example `GlossaryExamples` Dictionary. Its declared native inputs, outputs, lifetime/generic descriptors, and the original selected module's native types/support are required. Lyger resolves the source and composes its library AML; this fragment has no native compilation or execution claim.

<!-- example:glossary-native-storage-text-render -->

</details>

### Storage.Truth.Render

<a id="word-authoring-storage-truth-render"></a>

<!-- word:authoring:Storage.Truth.Render -->

Project an erased payload to `bool` and return the bytes of its native `true` or `false` spelling.

<!-- word-definition:authoring:Storage.Truth.Render -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

The wrapper below imports `Storage` and `Runtime` and is a member of the example `GlossaryExamples` Dictionary. Its declared native inputs, outputs, lifetime/generic descriptors, and the original selected module's native types/support are required. Lyger resolves the source and composes its library AML; this fragment has no native compilation or execution claim.

<!-- example:glossary-native-storage-truth-render -->

</details>

### Storage.Value.Renderer

<a id="word-authoring-storage-value-renderer"></a>

<!-- word:authoring:Storage.Value.Renderer -->

Select the native field `Value.show` by value. Its declared type determines whether this copies or moves the field.

<!-- word-definition:authoring:Storage.Value.Renderer -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

This statement is used inside `Storage.Value.Copy`. Its enclosing native signature and preceding bindings remain required.

<!-- example:glossary-storage-value-renderer-caller -->

</details>

### Runtime.Repeat

<a id="word-authoring-runtime-repeat"></a>

<!-- word:authoring:Runtime.Repeat -->

Expand the supplied Process as the body of a native loop; exits belong to the Process body, such as Runtime.Halt.

<!-- word-definition:authoring:Runtime.Repeat -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

This statement is used inside `Runtime.Supervision.Await`. Its enclosing native signature and preceding bindings remain required.

<!-- example:glossary-runtime-halt-caller -->

</details>

### Runtime.Replace

<a id="word-authoring-runtime-replace"></a>

<!-- word:authoring:Runtime.Replace -->

Assign the supplied value through the selected mutable reference.

<!-- word-definition:authoring:Runtime.Replace -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

This statement is used inside `Runtime.Invocation.Identify`. Its enclosing native signature and preceding bindings remain required.

<!-- example:glossary-runtime-replace-caller -->

</details>

### Storage.Replace

<a id="word-authoring-storage-replace"></a>

<!-- word:authoring:Storage.Replace -->

Replace the value at a supplied mutable location and return its previous value through `std::mem::replace`.

<!-- word-definition:authoring:Storage.Replace -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

This statement is used inside `Storage.List.Update`. Its enclosing native signature and preceding bindings remain required.

<!-- example:glossary-storage-replace-caller -->

</details>

### Runtime.Reply

<a id="word-authoring-runtime-reply"></a>

<!-- word:authoring:Runtime.Reply -->

Create a native one-shot reply channel and return its sender/receiver pair. The selected operation is `tokio::sync::oneshot::channel()`.

<!-- word-definition:authoring:Runtime.Reply -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

This statement is used inside `Runtime.Invocation.Kill`. Its enclosing native signature and preceding bindings remain required.

<!-- example:glossary-runtime-reply-caller -->

</details>

### Runtime.Report

<a id="word-authoring-runtime-report"></a>

<!-- word:authoring:Runtime.Report -->

When diagnostic bytes exist, write and flush them to runtime stderr. Return status one if either output step fails; otherwise preserve the supplied status.

<!-- word-definition:authoring:Runtime.Report -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

This statement is used inside `Runtime.Finish`. Its enclosing native signature and preceding bindings remain required.

<!-- example:glossary-runtime-remember-caller -->

</details>

### Runtime.Request

<a id="word-authoring-runtime-request"></a>

<!-- word:authoring:Runtime.Request -->

Attempt the first terminal selection and begin shutdown only if this request wins that selection.

<!-- word-definition:authoring:Runtime.Request -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

This statement is used inside `Runtime.Refuse`. Its enclosing native signature and preceding bindings remain required.

<!-- example:glossary-runtime-request-caller -->

</details>

### Runtime.Invocation.Requests

<a id="word-authoring-runtime-invocation-requests"></a>

<!-- word:authoring:Runtime.Invocation.Requests -->

Borrow `Value.leader_requests` without moving the referenced native field.

<!-- word-definition:authoring:Runtime.Invocation.Requests -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

This statement is used inside `Runtime.Invocation.Kill`. Its enclosing native signature and preceding bindings remain required.

<!-- example:glossary-runtime-invocation-requests-caller -->

</details>

### Storage.Environment.Require

<a id="word-authoring-storage-environment-require"></a>

<!-- word:authoring:Storage.Environment.Require -->

Require a borrowed optional capture environment, retain its shared storage on success, and return the supplied runtime failure when absent.

<!-- word-definition:authoring:Storage.Environment.Require -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

The wrapper below imports `Storage` and `Runtime` and is a member of the example `GlossaryExamples` Dictionary. Its declared native inputs, outputs, lifetime/generic descriptors, and the original selected module's native types/support are required. Lyger resolves the source and composes its library AML; this fragment has no native compilation or execution claim.

<!-- example:glossary-native-storage-environment-require -->

</details>

### Storage.Require

<a id="word-authoring-storage-require"></a>

<!-- word:authoring:Storage.Require -->

Consume an Option and return its payload as a runtime success, or construct a runtime failure when it is empty.

<!-- word-definition:authoring:Storage.Require -->

**Usage:** `a New Storage's Requirer of Value and Message` consumes `Option<T>` and an `&str` message and returns `runtime::Result<T>`. The enclosing native module must include the compiler-selected Runtime support. Require does not return directly from its caller. That caller chooses how to handle the returned Result.

**Example:** `Selection.Require` in [Native callers](#context-storage--native-callers) keeps a successful Box's identity and returns the supplied message as a runtime error for None. Failure construction happens only on None. Message arguments still evaluate normally, including on Some. Passing an effectful argument is not made lazy by selection.

### Runtime.App.Resolve

<a id="word-authoring-runtime-app-resolve"></a>

<!-- word:authoring:Runtime.App.Resolve -->

Update an App's selection only if the supplied invocation is still current: store an operation failure and clear selection on failure, or select the retained invocation on success.

<!-- word-definition:authoring:Runtime.App.Resolve -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

This statement is used inside `Runtime.App.Wait`. Its enclosing native signature and preceding bindings remain required.

<!-- example:glossary-runtime-app-resolve-caller -->

</details>

### Runtime.Resolve

<a id="word-authoring-runtime-resolve"></a>

<!-- word:authoring:Runtime.Resolve -->

Ignore a successful or cancellation source result, and request status one with any other returned Error.

<!-- word-definition:authoring:Runtime.Resolve -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

This statement is used inside `Runtime.Finish`. Its enclosing native signature and preceding bindings remain required.

<!-- example:glossary-runtime-resolve-caller -->

</details>

### Storage.Resolve

<a id="word-authoring-storage-resolve"></a>

<!-- word:authoring:Storage.Resolve -->

Propagate a failed borrowed-result value; on success invoke the supplied typed copying Function once and return its owned result wrapped in success.

<!-- word-definition:authoring:Storage.Resolve -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

This statement is used inside [Storage.Slot.Read](#word-authoring-storage-slot-read). Its enclosing native signature, preceding local bindings and selected native module are required.

<!-- example:glossary-builtins-eng-storage-program-33-33 -->

</details>

### Storage.Text.Restore

<a id="word-authoring-storage-text-restore"></a>

<!-- word:authoring:Storage.Text.Restore -->

Downcast an erased payload to a byte vector and return its native clone, using the supplied message if the type does not match.

<!-- word-definition:authoring:Storage.Text.Restore -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

This statement is used inside `Storage.Text.Render`. Its enclosing native signature and preceding bindings remain required.

<!-- example:glossary-storage-text-restore-caller -->

</details>

### Storage.Value.Restore

<a id="word-authoring-storage-value-restore"></a>

<!-- word:authoring:Storage.Value.Restore -->

Consume a stored value, project its erased payload to the declared native type, and invoke the supplied typed snapshot Function to produce an owned result.

<!-- word-definition:authoring:Storage.Value.Restore -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

The wrapper below imports `Storage` and `Runtime` and is a member of the example `GlossaryExamples` Dictionary. Its declared native inputs, outputs, lifetime/generic descriptors, and the original selected module's native types/support are required. Lyger resolves the source and composes its library AML; this fragment has no native compilation or execution claim.

<!-- example:glossary-native-storage-value-restore -->

</details>

### Runtime.Result

<a id="word-authoring-runtime-result"></a>

<!-- word:authoring:Runtime.Result -->

Declare the generated runtime's native `Result<T>` alias as `std::result::Result<T, Error>`.

<!-- word-definition:authoring:Runtime.Result -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

**Native type use.** This L Function header uses the native Result<()> alias defined by Runtime.Result. The type name is inside the native type atom, so it is not a separate lexical L word reference.

<!-- example:glossary-runtime-eng-app-58-58 -->

</details>

### Storage.Result

<a id="word-authoring-storage-result"></a>

<!-- word:authoring:Storage.Result -->

Name the Storage dictionary namespace containing operations that compose native success-or-error results.

<!-- word-definition:authoring:Storage.Result -->

**Usage:** Qualify [Result.Select](#word-authoring-storage-result-select) for branch selection or [Expect](#word-authoring-storage-expect) for an explicit panic policy. Result is not an L runtime carrier declaration. Write the actual native type, such as `Result<T, E>` or `runtime::Result<T>`, in the enclosing Function signature.

**Example:** `Selection.Preserve` in [Native callers](#context-storage--native-callers) handles `Result<u64, u8>`. Its error is a `u8`, while `Selection.Require` returns the runtime's Error type. Sharing the namespace does not unify those types or failure policies.

### Runtime.Resume

<a id="word-authoring-runtime-resume"></a>

<!-- word:authoring:Runtime.Resume -->

Resume unwinding with the supplied native panic payload. The selected operation is `std::panic::resume_unwind(Panic)`.

<!-- word-definition:authoring:Runtime.Resume -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

This statement is used inside `Runtime.Run`. Its enclosing native signature and preceding bindings remain required.

<!-- example:glossary-runtime-resume-caller -->

</details>

### Retain

<a id="word-builtin-retain"></a>

<!-- word:builtin:retain -->

Borrow an Arc handle and produce another handle to the same allocation.

<!-- word-definition:builtin:retain -->

**Usage and example.** The [native caller module](../examples/word-core.l) takes `&std::sync::Arc<u8>` and returns the retained Arc, preserving allocation identity.

Retain increments shared ownership without cloning the payload. Atomic reference counting does not make an arbitrary payload Send or Sync.

The [Storage ownership words](#context-storage--ownership-and-bytes) distinguish this Arc operation from Rc retention, copying and release.

### Storage.App.Retain

<a id="word-authoring-storage-app-retain"></a>

<!-- word:authoring:Storage.App.Retain -->

Borrow an Arc-backed App handle and return another handle to the same App allocation, using core Retain without changing process lifecycle state.

<!-- word-definition:authoring:Storage.App.Retain -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

This statement is used inside `Storage.Program.Select`. Its enclosing native signature and preceding bindings remain required.

<!-- example:glossary-storage-app-caller -->

</details>

### Storage.Function.Retain

<a id="word-authoring-storage-function-retain"></a>

<!-- word:authoring:Storage.Function.Retain -->

Copy the selected Function environment once and pair it with the cast function target, preserving shared capture identity.

<!-- word-definition:authoring:Storage.Function.Retain -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

**Compiler use.** Retain the Function self binding with its environment, native pointer type, and selected environment copier. This retained compiler excerpt selects the L-owned definition with typed native name/type/value or Process bindings. Those bindings come from the compiler’s resolved plan; this is an integration fragment, not an ordinary standalone L call.

<!-- example:glossary-lyger-function-composition-338-345 -->

</details>

### Storage.Retain

<a id="word-authoring-storage-retain"></a>

<!-- word:authoring:Storage.Retain -->

Borrow an Rc-backed value and return another handle to the same allocation, without copying its payload.

<!-- word-definition:authoring:Storage.Retain -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

This statement is used inside `Storage.Value.Copy`. Its enclosing native signature and preceding bindings remain required.

<!-- example:glossary-storage-retain-caller -->

</details>

### Function.Return

<a id="word-builtin-function-return"></a>

<!-- word:builtin:function.return -->

Complete the current Function invocation through its selected Return definition with the declared results.

<!-- word-definition:builtin:function.return -->

**Usage and example.** `Return Value` in the [Identity/Relay program](../examples/composed-functions.l) returns Text, while Relay separately binds and returns Nested. A Return from a selected Process still belongs to its enclosing Function.

Root-authored returning actions select their caller's Return. That selected implementation preserves its own lexical delegates. Function-owned actions retain their constructor contract and reject incompatible callers. Each Function establishes its own return context and checks declared result types and counts. Known unconditional exits through source actors and transparent Processes enter the typed plan. Opaque or conditional bodies do not establish unconditional exits.

Return completes that invocation. Exit terminates the enclosing program. Neither transports runtime captures into an App process.

Existing qualified aliases: `Storage.Routine.Return`. They select this canonical meaning.

### Storage.Function.Return

<a id="word-authoring-storage-function-return"></a>

<!-- word:authoring:Storage.Function.Return -->

Expand the selected Return definition supplied by the enclosing typed Function composition.

<!-- word-definition:authoring:Storage.Function.Return -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

This statement is used inside `builtins/eng/storage/execution.l`. Its enclosing native signature and preceding bindings remain required.

<!-- example:glossary-storage-function-return-caller -->

</details>

### Storage.Review

<a id="word-authoring-storage-review"></a>

<!-- word:authoring:Storage.Review -->

Invoke a callback with two items and context, then discard a successful payload or propagate its failure from the enclosing Function.

<!-- word-definition:authoring:Storage.Review -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

This statement is used inside `builtins/eng/storage/traversal.l`. Its enclosing native signature and preceding bindings remain required.

<!-- example:glossary-storage-review-caller -->

</details>

### Runtime.Revise

<a id="word-authoring-runtime-revise"></a>

<!-- word:authoring:Runtime.Revise -->

Apply Limit to a watch sender's optional deadline through its native modifying operation.

<!-- word-definition:authoring:Runtime.Revise -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

This statement is used inside `Runtime.Invocation.Bound`. Its enclosing native signature and preceding bindings remain required.

<!-- example:glossary-runtime-revise-caller -->

</details>

### Storage.Routine

<a id="word-authoring-storage-routine"></a>

<!-- word:authoring:Storage.Routine -->

Define a native callable from the supplied name, typed input/output declarations, explicit generic parameters and Process body.

<!-- word-definition:authoring:Storage.Routine -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

This statement is used inside `builtins/eng/storage/control.l`. Its enclosing native signature and preceding bindings remain required.

<!-- example:glossary-storage-routine-caller -->

</details>

### Run

<a id="word-builtin-run"></a>

<!-- word:builtin:run -->

Launch an App invocation in a separate operating-system process without joining it.

<!-- word-definition:builtin:run -->

**Usage and example.** `Run Worker` in the [continuation program](../../testdata/phase1/language/lang-01.l) launches the declared Basic App. The following Show runs before Await joins it. Use [Await](#word-builtin-await) for finalized output.

Run returns after process creation acknowledgment or recorded startup failure. Starting several Apps before Await permits overlap. A startup failure is still awaitable; running an outstanding target again rejects. The [launch contract](../cli-build.md#launch-and-language-forms) and [invocation rules](../cli-build.md#invocation-and-results) define acknowledgment, aliases and reruns.

### Runtime.App.Run

<a id="word-authoring-runtime-app-run"></a>

<!-- word:authoring:Runtime.App.Run -->

Allocate an App invocation, check runtime admission, launch its child or publish its startup failure, and arrange supervision for a launched child. Return the invocation without joining it.

<!-- word-definition:authoring:Runtime.App.Run -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

The wrapper below imports `Storage` and `Runtime` and is a member of the example `GlossaryExamples` Dictionary. Its declared native inputs, outputs, lifetime/generic descriptors, and the original selected module's native types/support are required. Lyger resolves the source and composes its library AML; this fragment has no native compilation or execution claim.

<!-- example:glossary-native-runtime-app-run -->

</details>

### Runtime.Run

<a id="word-authoring-runtime-run"></a>

<!-- word:authoring:Runtime.Run -->

Race the source future against cancellation, settle runtime cleanup and status in either outcome, and resume a source panic only when cleanup and terminal selection do not determine another result.

<!-- word-definition:authoring:Runtime.Run -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>


The wrapper below imports `Storage` and `Runtime` and is a member of the example `GlossaryExamples` Dictionary. Its declared native inputs, outputs, lifetime/generic descriptors, and the original selected module's native types/support are required. Lyger resolves the source and composes its library AML; this fragment has no native compilation or execution claim.

<!-- example:glossary-native-runtime-run -->

</details>

Existing qualified aliases: `Runtime.Host.Run`. They select this canonical meaning.

### Runtime

<a id="word-authoring-runtime"></a>

<!-- word:authoring:Runtime -->

Name the implementation dictionary composing App processes, capture, Errors, shutdown and host integration. Its members require the generated native runtime context.

<!-- word-definition:authoring:Runtime -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

Import the dictionary to make its qualified members available.

<!-- example:glossary-namespace-runtime -->

</details>

## S

<a id="glossary-s"></a>

### Formatting.Scale

<a id="word-authoring-formatting-scale"></a>

<!-- word:authoring:Formatting.Scale -->

Multiply an integer coefficient by powers of five for positive exponents, or divide exactly by powers of five for negative exponents. Return None for multiplication overflow or an inexact division.

<!-- word-definition:authoring:Formatting.Scale -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

This statement is used inside `Formatting.Balance`. Its enclosing native signature and preceding bindings remain required.

<!-- example:glossary-formatting-scale-caller -->

</details>

### Storage.Scope

<a id="word-authoring-storage-scope"></a>

<!-- word:authoring:Storage.Scope -->

Place the supplied Body inside a native loop with a trailing break. Body can transfer control before the trailing break.

<!-- word-definition:authoring:Storage.Scope -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

**Compiler use.** Supply a composed Process body, selecting Scope when source control can continue and Loop when it cannot. This retained compiler excerpt selects the L-owned definition with typed native name/type/value or Process bindings. Those bindings come from the compiler’s resolved plan; this is an integration fragment, not an ordinary standalone L call.

<!-- example:glossary-lyger-execution-composition-245-255 -->

</details>

### Runtime.Invocation.Seal

<a id="word-authoring-runtime-invocation-seal"></a>

<!-- word:authoring:Runtime.Invocation.Seal -->

Mutably borrow `Value.cleaned` so the caller can update that exact native field.

<!-- word-definition:authoring:Runtime.Invocation.Seal -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

This statement is used inside `Runtime.Invocation.Signal`. Its enclosing native signature and preceding bindings remain required.

<!-- example:glossary-runtime-bound-caller -->

</details>

### Runtime.Error.Search

<a id="word-authoring-runtime-error-search"></a>

<!-- word:authoring:Runtime.Error.Search -->

Visit the selected iterator in order with the supplied callback and return its first Some result. The selected operation is `Iterator.find_map(Action)`.

<!-- word-definition:authoring:Runtime.Error.Search -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

This statement is used inside `Runtime.Error.Trace`. Its enclosing native signature and preceding bindings remain required.

<!-- example:glossary-runtime-error-search-caller -->

</details>

### Storage.Search

<a id="word-authoring-storage-search"></a>

<!-- word:authoring:Storage.Search -->

Probe values in iterator order with the supplied callback and continuation. A falling-through continuation can run for every true result; early exit depends on the continuation itself.

<!-- word-definition:authoring:Storage.Search -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

This statement is used inside `Storage.Value.Search`. Its enclosing native signature and preceding bindings remain required.

<!-- example:glossary-storage-search-caller -->

</details>

### Storage.Value.Search

<a id="word-authoring-storage-value-search"></a>

<!-- word:authoring:Storage.Value.Search -->

Search values in iterator order, returning successful true at the first ownership-reachable value, successful false after exhaustion, or the first traversal failure.

<!-- word-definition:authoring:Storage.Value.Search -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

This statement is used inside `Storage.List.Reach`. Its enclosing native signature and preceding bindings remain required.

<!-- example:glossary-storage-value-search-caller -->

</details>

### Storage.Tuple.Second

<a id="word-authoring-storage-tuple-second"></a>

<!-- word:authoring:Storage.Tuple.Second -->

Select the second tuple component by value; the component's native type determines copy or move behavior.

<!-- word-definition:authoring:Storage.Tuple.Second -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

This statement is used inside `Runtime.Invocation.Kill`. Its enclosing native signature and preceding bindings remain required.

<!-- example:glossary-storage-tuple-second-caller -->

</details>

Existing qualified aliases: `Storage.Function.Target`. They select this canonical meaning.

### Runtime.App.Select

<a id="word-authoring-runtime-app-select"></a>

<!-- word:authoring:Runtime.App.Select -->

Lock App state and copy the selected invocation together with the saved operation error.

<!-- word-definition:authoring:Runtime.App.Select -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

This statement is used inside `Runtime.App.Inspect`. Its enclosing native signature and preceding bindings remain required.

<!-- example:glossary-runtime-app-select-caller -->

</details>

### Storage.Option.Select

<a id="word-authoring-storage-option-select"></a>

<!-- word:authoring:Storage.Option.Select -->

Select exactly one Process according to whether a supplied native Option contains a payload or is empty.

<!-- word-definition:authoring:Storage.Option.Select -->

**Usage:** Supply the Option value, a Some Process with one explicit typed Input, then a None Process with no payload Input. The payload type must match the enclosing Function's native carrier. Selector is Select's actor form in the Option namespace. It is distinct from [Result.Select](#word-authoring-storage-result-select).

**Example:** `Selection.Choose` in [Native callers](#context-storage--native-callers) returns the bound `u64` for Some and zero for None. The selected value is evaluated once. Only the chosen body runs, and its Input and declarations remain local to that Process. Moving or borrowing follows the supplied type. Selection does not clone the payload. Nested selections retain their own bindings.

### Storage.Program.Select

<a id="word-authoring-storage-program-select"></a>

<!-- word:authoring:Storage.Program.Select -->

Bounds-check a program's App-handle index and retain the selected Arc. An invalid index reports that the App handle is not owned by this program.

<!-- word-definition:authoring:Storage.Program.Select -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

The wrapper below imports `Storage` and `Runtime` and is a member of the example `GlossaryExamples` Dictionary. Its declared native inputs, outputs, lifetime/generic descriptors, and the original selected module's native types/support are required. Lyger resolves the source and composes its library AML; this fragment has no native compilation or execution claim.

<!-- example:glossary-native-storage-program-select -->

</details>

### Storage.Result.Select

<a id="word-authoring-storage-result-select"></a>

<!-- word:authoring:Storage.Result.Select -->

Select exactly one Process according to whether a supplied native Result is successful or contains an error.

<!-- word-definition:authoring:Storage.Result.Select -->

**Usage:** Supply the Result value, an Ok Process with one typed payload Input, then an Err Process with one typed error Input. Each Input matches its corresponding native type. Both bodies belong to the enclosing Function context, so a Return in the selected body returns from that Function.

**Example:** `Selection.Preserve` in [Native callers](#context-storage--native-callers) reconstructs Ok and Err using their original values. The Result is evaluated once, only one branch runs, and branch-local bindings do not leak into siblings. Selection adds no panic, error replacement or clone. [IO.Lined](#word-stdlib-export-eng-io-io-lined) composes this selector to send LF only after payload success.

### Storage.Select

<a id="word-authoring-storage-select"></a>

<!-- word:authoring:Storage.Select -->

Execute the supplied Then Process for a true native condition and the Else Process otherwise.

<!-- word-definition:authoring:Storage.Select -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

This statement is used inside `builtins/eng/storage/traversal.l`. Its enclosing native signature and preceding bindings remain required.

<!-- example:glossary-storage-select-caller -->

</details>

### Runtime.Selected

<a id="word-authoring-runtime-selected"></a>

<!-- word:authoring:Runtime.Selected -->

Check whether the selected native Option contains a payload. The selected operation is `Value.terminal.is_some()`.

<!-- word-definition:authoring:Runtime.Selected -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

This statement is used inside `Runtime.Decide`. Its enclosing native signature and preceding bindings remain required.

<!-- example:glossary-runtime-selected-caller -->

</details>

### Runtime.App.Selection

<a id="word-authoring-runtime-app-selection"></a>

<!-- word:authoring:Runtime.App.Selection -->

Mutably borrow `Value.selected` so the caller can update that exact native field.

<!-- word-definition:authoring:Runtime.App.Selection -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

This statement is used inside `Runtime.App.Build`. Its enclosing native signature and preceding bindings remain required.

<!-- example:glossary-runtime-app-selection-caller -->

</details>

### IO.Send

<a id="word-stdlib-export-eng-io-io-send"></a>

<!-- word:stdlib-export:eng:io:IO.Send -->

Write a borrowed byte view through a supplied mutable writer and return the native IO result.

<!-- word-definition:stdlib-export:eng:io:IO.Send -->

**Usage:** `a New IO's Sender of Writer and Payload` takes `&mut W` with `W: std::io::Write + ?Sized`, then `&P` with `P: AsRef<[u8]> + ?Sized`, and returns `std::io::Result<()>`. Sender is the actor form of Send, and Sent is its declared passive form, not a separate implementation.

**Example:** `Writing.Send` in [Native callers](#context-io--native-callers) sends a borrowed byte slice into a supplied vector. Send obtains one byte view and calls `write_all`. It adds no newline, lock, flush, owned Text conversion or panic policy. Use [Lined](#word-stdlib-export-eng-io-io-lined) to append LF and [Flush](#word-stdlib-export-eng-io-io-flush) for an explicit flush request.

### Runtime.Send

<a id="word-authoring-runtime-send"></a>

<!-- word:authoring:Runtime.Send -->

Ask the selected host dependencies to send the supplied signal to the supplied target. The selected operation is `Dependencies.signal(Target, Signal)`.

<!-- word-definition:authoring:Runtime.Send -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

This statement is used inside `Runtime.Invocation.Signal`. Its enclosing native signature and preceding bindings remain required.

<!-- example:glossary-runtime-normalize-caller -->

</details>

### Runtime.Invocation.Sequence

<a id="word-authoring-runtime-invocation-sequence"></a>

<!-- word:authoring:Runtime.Invocation.Sequence -->

Select the native field `Value.sequence` by value. Its declared type determines whether this copies or moves the field.

<!-- word-definition:authoring:Runtime.Invocation.Sequence -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

This statement is used inside `Runtime.App.Supervise`. Its enclosing native signature and preceding bindings remain required.

<!-- example:glossary-runtime-invocation-sequence-caller -->

</details>

### Service

<a id="word-type-service"></a>

<!-- word:type:service -->

Group source code within an App, supplying its primary entry structure or an additional constructed library.

<!-- word-definition:type:service -->

**Usage and example.** The [Service composition fixture](../../lyger/test/workspace/services.l) constructs Main, Left and Right Services within one App. A standalone Service remains structural and cannot be Run or Awaited.

Basic supplies the primary Main Service. Additional Services constructed inside an App become native libraries whose nonprivate composed items form their interface. The [Service composition example](../guide/apps-and-services.md#service-composition) builds Main, Left and Right Services inside an App and explains how the body is passed into them.

A standalone Service declaration creates no package or process. Standalone body effects, native lifecycle operations and implicit Output/Error reject. Explicit static members and source-defined actions still use normal lexical lookup. See the [Service composition contract](../cli-build.md#service-composition).

### Runtime.App.Set

<a id="word-authoring-runtime-app-set"></a>

<!-- word:authoring:Runtime.App.Set -->

Replace an App's diagnostic source-attribution string while holding its state lock.

<!-- word-definition:authoring:Runtime.App.Set -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

The wrapper below imports `Storage` and `Runtime` and is a member of the example `GlossaryExamples` Dictionary. Its declared native inputs, outputs, lifetime/generic descriptors, and the original selected module's native types/support are required. Lyger resolves the source and composes its library AML; this fragment has no native compilation or execution claim.

<!-- example:glossary-native-runtime-app-set -->

</details>

### Runtime.Error.Set

<a id="word-authoring-runtime-error-set"></a>

<!-- word:authoring:Runtime.Error.Set -->

Construct an empty native hash set used to track rendered Error identities. The selected operation is `std::collections::HashSet::new()`.

<!-- word-definition:authoring:Runtime.Error.Set -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

This statement is used inside `Runtime.Error.Render`. Its enclosing native signature and preceding bindings remain required.

<!-- example:glossary-runtime-error-set-caller -->

</details>

### Runtime.Invocation.Settle

<a id="word-authoring-runtime-invocation-settle"></a>

<!-- word:authoring:Runtime.Invocation.Settle -->

Request process-group termination; if it fails, request leader termination and record the joined cleanup errors.

<!-- word-definition:authoring:Runtime.Invocation.Settle -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

This statement is used inside `Runtime.Reap`. Its enclosing native signature and preceding bindings remain required.

<!-- example:glossary-runtime-invocation-settle-caller -->

</details>

### Runtime.Capture.Share

<a id="word-authoring-runtime-capture-share"></a>

<!-- word:authoring:Runtime.Capture.Share -->

Consume a Capture and return its bytes in an Arc-backed slice together with its optional Error.

<!-- word-definition:authoring:Runtime.Capture.Share -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

This statement is used inside `Runtime.Capture.Join`. Its enclosing native signature and preceding bindings remain required.

<!-- example:glossary-runtime-capture-share-caller -->

</details>

### Storage.Share

<a id="word-authoring-storage-share"></a>

<!-- word:authoring:Storage.Share -->

Allocate an Rc containing the supplied value. This creates shared ownership without cloning the value and is distinct from Arc allocation or retention.

<!-- word-definition:authoring:Storage.Share -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

This statement is used inside `Storage.Value.Wrap`. Its enclosing native signature and preceding bindings remain required.

<!-- example:glossary-storage-share-caller -->

</details>

### Runtime.Show

<a id="word-authoring-runtime-show"></a>

<!-- word:authoring:Runtime.Show -->

Check runtime admission, lock stdout, recheck admission, write the exact supplied bytes and flush. An output failure requests fatal runtime termination; no LF is added.

<!-- word-definition:authoring:Runtime.Show -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

The wrapper below imports `Storage` and `Runtime` and is a member of the example `GlossaryExamples` Dictionary. Its declared native inputs, outputs, lifetime/generic descriptors, and the original selected module's native types/support are required. Lyger resolves the source and composes its library AML; this fragment has no native compilation or execution claim.

<!-- example:glossary-native-runtime-show -->

</details>

### Show

<a id="word-builtin-show"></a>

<!-- word:builtin:show -->

Write a supported runtime value to the outer program's output without adding a newline.

<!-- word-definition:builtin:show -->

**Usage and example.** The [public-values program](../../demo/paragraphs/public-values.l) shows Text, Number and Truth as adjacent bytes. The [first program](../guide/first-program.md#read-the-program) shows finalized App Output.

Show preserves Text bytes and formats supported runtime values. Direct App and Error display rejects, and other targets' streams are not replayed. Imported [IO.Print](#word-stdlib-export-eng-io-io-print) supplies the LF in the first program's `Hello World!` output.

### Storage.Show

<a id="word-authoring-storage-show"></a>

<!-- word:authoring:Storage.Show -->

Call and await runtime_show using the enclosing runtime and a borrow of the supplied byte value.

<!-- word-definition:authoring:Storage.Show -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

**Compiler use.** Format the dynamic value, check the bytes, display them through runtime output, and check the returned result. This retained compiler excerpt selects the L-owned definition with typed native name/type/value or Process bindings. Those bindings come from the compiler’s resolved plan; this is an integration fragment, not an ordinary standalone L call.

<!-- example:glossary-lyger-execution-composition-595-609 -->

</details>

### Runtime.Invocation.Signal

<a id="word-authoring-runtime-invocation-signal"></a>

<!-- word:authoring:Runtime.Invocation.Signal -->

Serialize process-group termination, skip completed or already-cleaned invocations, signal the selected group, establish its drain deadline, and perform final bounded cleanup when requested.

<!-- word-definition:authoring:Runtime.Invocation.Signal -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

This statement is used inside `Runtime.Invocation.Terminate`. Its enclosing native signature and preceding bindings remain required.

<!-- example:glossary-runtime-invocation-signal-caller -->

</details>

### Runtime.Signal

<a id="word-authoring-runtime-signal"></a>

<!-- word:authoring:Runtime.Signal -->

Construct a host terminal request carrying the supplied signal status. The selected operation is `crate::runtime::host::TerminalRequest::Signal(Value)`.

<!-- word-definition:authoring:Runtime.Signal -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

This statement is used inside `Runtime.Signals.Watch`. Its enclosing native signature and preceding bindings remain required.

<!-- example:glossary-runtime-signal-caller -->

</details>

### Runtime.Supervision.Signal

<a id="word-authoring-runtime-supervision-signal"></a>

<!-- word:authoring:Runtime.Supervision.Signal -->

Request native termination of the supplied child-leader process ID. The selected operation is `crate::process::terminate_leader(Pid)`.

<!-- word-definition:authoring:Runtime.Supervision.Signal -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

This statement is used inside `Runtime.Supervision.Terminate`. Its enclosing native signature and preceding bindings remain required.

<!-- example:glossary-runtime-supervision-explain-caller -->

</details>

### Runtime.Inner.Signals

<a id="word-authoring-runtime-inner-signals"></a>

<!-- word:authoring:Runtime.Inner.Signals -->

Borrow `Value.signals` without moving the referenced native field.

<!-- word-definition:authoring:Runtime.Inner.Signals -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

This statement is used inside `Runtime.Unsubscribe`. Its enclosing native signature and preceding bindings remain required.

<!-- example:glossary-runtime-inner-signals-caller -->

</details>

### Runtime.Signals

<a id="word-authoring-runtime-signals"></a>

<!-- word:authoring:Runtime.Signals -->

Install the native process-signal subscriptions and return their receiver or IO failure. The selected operation is `crate::process::Signals::new()`.

<!-- word-definition:authoring:Runtime.Signals -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

This statement is used inside `Runtime.Install`. Its enclosing native signature and preceding bindings remain required.

<!-- example:glossary-runtime-signals-caller -->

</details>

### Storage.Signature

<a id="word-authoring-storage-signature"></a>

<!-- word:authoring:Storage.Signature -->

Name the Storage namespace for universal and exact runtime type-signature matching.

<!-- word-definition:authoring:Storage.Signature -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

This statement is used inside `Storage.Signature.Match`. Its enclosing native signature and preceding bindings remain required.

<!-- example:glossary-storage-signature-caller -->

</details>

### Storage.Value.Signature

<a id="word-authoring-storage-value-signature"></a>

<!-- word:authoring:Storage.Value.Signature -->

Borrow `Value.ty` without moving the referenced native field.

<!-- word-definition:authoring:Storage.Value.Signature -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

This statement is used inside `Storage.Value.Check`. Its enclosing native signature and preceding bindings remain required.

<!-- example:glossary-storage-value-signature-caller -->

</details>

### Storage.Slot

<a id="word-authoring-storage-slot"></a>

<!-- word:authoring:Storage.Slot -->

Name the Storage namespace for allocating, reading and writing optional typed storage slots.

<!-- word-definition:authoring:Storage.Slot -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

This statement is used inside [Storage.Program.Create](#word-authoring-storage-program-create). Its enclosing native signature, preceding local bindings and selected native module are required.

<!-- example:glossary-builtins-eng-storage-program-48-48 -->

</details>

### Runtime.Host.Snapshot

<a id="word-authoring-runtime-host-snapshot"></a>

<!-- word:authoring:Runtime.Host.Snapshot -->

Name the adapter namespace reusing canonical Snapshot copying for native host integration.

<!-- word-definition:authoring:Runtime.Host.Snapshot -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

After importing Runtime, bind the existing namespace for qualified member access.

<!-- example:glossary-namespace-runtime-host-snapshot -->

</details>

### Runtime.Invocation.Snapshot

<a id="word-authoring-runtime-invocation-snapshot"></a>

<!-- word:authoring:Runtime.Invocation.Snapshot -->

Borrow `Value.snapshot` without moving the referenced native field.

<!-- word-definition:authoring:Runtime.Invocation.Snapshot -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

This statement is used inside `Runtime.Invocation.Read`. Its enclosing native signature and preceding bindings remain required.

<!-- example:glossary-runtime-invocation-snapshot-caller -->

</details>

### Runtime.Snapshot

<a id="word-authoring-runtime-snapshot"></a>

<!-- word:authoring:Runtime.Snapshot -->

Declare the native `Snapshot` record with fields `pub stdout: std::sync::Arc<[u8]>`, `pub stderr: std::sync::Arc<[u8]>`, `pub error: Option<Error>`, `pub status: Option<i32>`.

<!-- word-definition:authoring:Runtime.Snapshot -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

This statement is used inside `Runtime.Snapshot.Reject`. Its enclosing native signature and preceding bindings remain required.

<!-- example:glossary-runtime-snapshot-caller -->

</details>

### Runtime.App.Source

<a id="word-authoring-runtime-app-source"></a>

<!-- word:authoring:Runtime.App.Source -->

Mutably borrow `Value.source` so the caller can update that exact native field.

<!-- word-definition:authoring:Runtime.App.Source -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

This statement is used inside `Runtime.App.Set`. Its enclosing native signature and preceding bindings remain required.

<!-- example:glossary-runtime-app-source-caller -->

</details>

### Runtime.Error.Source

<a id="word-authoring-runtime-error-source"></a>

<!-- word:authoring:Runtime.Error.Source -->

Cast a native cause reference to the borrowed Error interface with the declared static native-cause bound.

<!-- word-definition:authoring:Runtime.Error.Source -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

This statement is used inside `Runtime.Error.Trace`. Its enclosing native signature and preceding bindings remain required.

<!-- example:glossary-runtime-error-source-caller -->

</details>

### Runtime.Spawn

<a id="word-authoring-runtime-spawn"></a>

<!-- word:authoring:Runtime.Spawn -->

Spawn the supplied future on the active native executor and return its task handle. The selected operation is `tokio::spawn(Value)`.

<!-- word-definition:authoring:Runtime.Spawn -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

This statement is used inside `Runtime.Begin`. Its enclosing native signature and preceding bindings remain required.

<!-- example:glossary-runtime-reap-caller -->

</details>

### Storage.Specialize

<a id="word-authoring-storage-specialize"></a>

<!-- word:authoring:Storage.Specialize -->

Bind the supplied native Function with an explicitly selected native type argument.

<!-- word-definition:authoring:Storage.Specialize -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

**Compiler use.** Specialize the selected visitor Function with the resolved native type atom. This retained compiler excerpt selects the L-owned definition with typed native name/type/value or Process bindings. Those bindings come from the compiler’s resolved plan; this is an integration fragment, not an ordinary standalone L call.

<!-- example:glossary-lyger-word-values-310-321 -->

</details>

### Runtime.App.Start

<a id="word-authoring-runtime-app-start"></a>

<!-- word:authoring:Runtime.App.Start -->

Ask the selected host dependencies to start the supplied child command. The selected operation is `Dependencies.start(Command)`.

<!-- word-definition:authoring:Runtime.App.Start -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

This statement is used inside `Runtime.App.Launch`. Its enclosing native signature and preceding bindings remain required.

<!-- example:glossary-runtime-app-start-caller -->

</details>

### Runtime.Capture.Start

<a id="word-authoring-runtime-capture-start"></a>

<!-- word:authoring:Runtime.Capture.Start -->

Require a capture reader, obtain the host capture destination, and repeatedly advance capture until EOF, failure, or cleanup deadline. Release the reader and destination, then return the bytes collected and optional failure.

<!-- word-definition:authoring:Runtime.Capture.Start -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

This statement is used inside `Runtime.Capture.Read`. Its enclosing native signature and preceding bindings remain required.

<!-- example:glossary-runtime-capture-start-caller -->

</details>

### Start

<a id="word-builtin-start"></a>

<!-- word:builtin:start -->

Build an App and then launch its native process without waiting for completion.

<!-- word-definition:builtin:start -->

**Usage and example.** `Start First and Second` applies the whole Build/Run composition to each App in order. The [two-App program](../../demo/phase1/main.l) starts both before joining. Start does not promise application readiness.

### Runtime.App.State

<a id="word-authoring-runtime-app-state"></a>

<!-- word:authoring:Runtime.App.State -->

Declare the native `AppState` record with fields `source: String`, `invocation: Option<std::sync::Arc<Invocation>>`, `outstanding: bool`, `selected: Option<std::sync::Arc<Invocation>>`, `operation_error: Option<Error>`.

<!-- word-definition:authoring:Runtime.App.State -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

This statement is used inside `Runtime.Appoint`. Its enclosing native signature and preceding bindings remain required.

<!-- example:glossary-runtime-app-state-caller -->

</details>

### Runtime.Invocation.State

<a id="word-authoring-runtime-invocation-state"></a>

<!-- word:authoring:Runtime.Invocation.State -->

Declare the native `InvocationState` record with fields `snapshot: Option<Snapshot>`, `claim: u8`.

<!-- word-definition:authoring:Runtime.Invocation.State -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

This statement is used inside `Runtime.Invocation.Allocate`. Its enclosing native signature and preceding bindings remain required.

<!-- example:glossary-runtime-invocation-state-caller -->

</details>

### Runtime.State

<a id="word-authoring-runtime-state"></a>

<!-- word:authoring:Runtime.State -->

Declare the native `RuntimeState` record with fields `owned: std::collections::BTreeMap<u64, std::sync::Arc<Invocation>>`, `cleanup_errors: std::collections::BTreeMap<u64, Error>`, `next: u64`, `stopping: bool`, `terminal: Option<Terminal>`, `finalized: bool`, `final_status: Option<u8>`.

<!-- word-definition:authoring:Runtime.State -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

This statement is used inside `Runtime.Read`. Its enclosing native signature and preceding bindings remain required.

<!-- example:glossary-runtime-state-caller -->

</details>

### Runtime.Error.Status

<a id="word-authoring-runtime-error-status"></a>

<!-- word:authoring:Runtime.Error.Status -->

Select the native field `Value.data.cancelled` by value. Its declared type determines whether this copies or moves the field.

<!-- word-definition:authoring:Runtime.Error.Status -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

This statement is used inside `Runtime.Error.Check`. Its enclosing native signature and preceding bindings remain required.

<!-- example:glossary-runtime-error-status-caller -->

</details>

### Runtime.Snapshot.Status

<a id="word-authoring-runtime-snapshot-status"></a>

<!-- word:authoring:Runtime.Snapshot.Status -->

Select the native field `Value.status` by value. Its declared type determines whether this copies or moves the field.

<!-- word-definition:authoring:Runtime.Snapshot.Status -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

This statement is used inside `Runtime.Snapshot.Copy`. Its enclosing native signature and preceding bindings remain required.

<!-- example:glossary-runtime-snapshot-status-caller -->

</details>

### Runtime.State.Status

<a id="word-authoring-runtime-state-status"></a>

<!-- word:authoring:Runtime.State.Status -->

Select the native field `Value.final_status` by value. Its declared type determines whether this copies or moves the field.

<!-- word-definition:authoring:Runtime.State.Status -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

This statement is used inside `Runtime.Read`. Its enclosing native signature and preceding bindings remain required.

<!-- example:glossary-runtime-state-caller -->

</details>

### Runtime.Terminal.Status

<a id="word-authoring-runtime-terminal-status"></a>

<!-- word:authoring:Runtime.Terminal.Status -->

Select the native field `Value.status` by value. Its declared type determines whether this copies or moves the field.

<!-- word-definition:authoring:Runtime.Terminal.Status -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

This statement is used inside `Runtime.Terminal.Copy`. Its enclosing native signature and preceding bindings remain required.

<!-- example:glossary-runtime-terminal-caller -->

</details>

### Runtime.Inner.Stderr

<a id="word-authoring-runtime-inner-stderr"></a>

<!-- word:authoring:Runtime.Inner.Stderr -->

Borrow `Value.stderr` without moving the referenced native field.

<!-- word-definition:authoring:Runtime.Inner.Stderr -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

This statement is used inside `Runtime.Report`. Its enclosing native signature and preceding bindings remain required.

<!-- example:glossary-runtime-inner-stderr-caller -->

</details>

### Runtime.Stderr

<a id="word-authoring-runtime-stderr"></a>

<!-- word:authoring:Runtime.Stderr -->

Obtain the native asynchronous standard-error handle. The selected operation is `tokio::io::stderr()`.

<!-- word-definition:authoring:Runtime.Stderr -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

This statement is used inside `Runtime.Open`. Its enclosing native signature and preceding bindings remain required.

<!-- example:glossary-runtime-output-caller -->

</details>

### Runtime.Supervision.Stderr

<a id="word-authoring-runtime-supervision-stderr"></a>

<!-- word:authoring:Runtime.Supervision.Stderr -->

Borrow the child’s stderr Option mutably. The caller applies Take separately to move out the stream once.

<!-- word-definition:authoring:Runtime.Supervision.Stderr -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

This statement is used inside `Runtime.Invocation.Supervise`. Its enclosing native signature and preceding bindings remain required.

<!-- example:glossary-runtime-supervision-stderr-caller -->

</details>

### Runtime.Inner.Stdout

<a id="word-authoring-runtime-inner-stdout"></a>

<!-- word:authoring:Runtime.Inner.Stdout -->

Borrow `Value.stdout` without moving the referenced native field.

<!-- word-definition:authoring:Runtime.Inner.Stdout -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

This statement is used inside `Runtime.Show`. Its enclosing native signature and preceding bindings remain required.

<!-- example:glossary-runtime-inner-stdout-caller -->

</details>

### Runtime.Stdout

<a id="word-authoring-runtime-stdout"></a>

<!-- word:authoring:Runtime.Stdout -->

Obtain the native asynchronous standard-output handle. The selected operation is `tokio::io::stdout()`.

<!-- word-definition:authoring:Runtime.Stdout -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

This statement is used inside `Runtime.Open`. Its enclosing native signature and preceding bindings remain required.

<!-- example:glossary-runtime-stdout-caller -->

</details>

### Runtime.Supervision.Stdout

<a id="word-authoring-runtime-supervision-stdout"></a>

<!-- word:authoring:Runtime.Supervision.Stdout -->

Borrow the child’s stdout Option mutably. The caller applies Take separately to move out the stream once.

<!-- word-definition:authoring:Runtime.Supervision.Stdout -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

This statement is used inside `Runtime.Invocation.Supervise`. Its enclosing native signature and preceding bindings remain required.

<!-- example:glossary-runtime-supervision-stdout-caller -->

</details>

### Runtime.App.Stop

<a id="word-authoring-runtime-app-stop"></a>

<!-- word:authoring:Runtime.App.Stop -->

Require a current invocation and runtime admission, request termination, and record a termination failure in the App's operation error while clearing its selected invocation. It does not join the child or activate its continuation.

<!-- word-definition:authoring:Runtime.App.Stop -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

The wrapper below imports `Storage` and `Runtime` and is a member of the example `GlossaryExamples` Dictionary. Its declared native inputs, outputs, lifetime/generic descriptors, and the original selected module's native types/support are required. Lyger resolves the source and composes its library AML; this fragment has no native compilation or execution claim.

<!-- example:glossary-native-runtime-app-stop -->

</details>

### Runtime.State.Stop

<a id="word-authoring-runtime-state-stop"></a>

<!-- word:authoring:Runtime.State.Stop -->

Mutably borrow `Value.stopping` so the caller can update that exact native field.

<!-- word-definition:authoring:Runtime.State.Stop -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

This statement is used inside `Runtime.Gather`. Its enclosing native signature and preceding bindings remain required.

<!-- example:glossary-runtime-state-stop-caller -->

</details>

### Stop

<a id="word-builtin-stop"></a>

<!-- word:builtin:stop -->

Request termination of an App's owned process group without waiting for finalized completion.

<!-- word-definition:builtin:stop -->

**Usage and example.** The [Stop/Await witness](../spec/execution.md#build-run-stop-and-await) shows `stop-requested|` before joining. A natural exit may win the race, so always Await when completion is needed.

Stop before Run is misuse; after finalized completion it is an idempotent no-op. Signal-delivery failure selects a recoverable operation Error and preserves ownership. Use Await for finalized output and cleanup state. The [shutdown contract](../cli-build.md#shutdown-streams-and-limits) defines group signaling and its race with natural exit.

### Storage.Function.Stop

<a id="word-authoring-storage-function-stop"></a>

<!-- word:authoring:Storage.Function.Stop -->

Construct the Break variant of native ControlFlow with a unit payload.

<!-- word-definition:authoring:Storage.Function.Stop -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

**Compiler use.** Construct a Break value when an unvalued exit occurs inside a Function. This retained compiler excerpt selects the L-owned definition with typed native name/type/value or Process bindings. Those bindings come from the compiler’s resolved plan; this is an integration fragment, not an ordinary standalone L call.

<!-- example:glossary-lyger-execution-composition-611-618 -->

</details>

### Runtime.Stopping

<a id="word-authoring-runtime-stopping"></a>

<!-- word:authoring:Runtime.Stopping -->

Select the native field `Value.stopping` by value. Its declared type determines whether this copies or moves the field.

<!-- word-definition:authoring:Runtime.Stopping -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

This statement is used inside `Runtime.Admit`. Its enclosing native signature and preceding bindings remain required.

<!-- example:glossary-runtime-stopping-caller -->

</details>

### Storage

<a id="word-authoring-storage"></a>

<!-- word:authoring:Storage -->

Name the foundational dictionary of typed storage, ownership, traversal and execution compositions. Source callers explicitly import it even though prelude loads it privately.

<!-- word-definition:authoring:Storage -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

Import the dictionary to make its qualified members available.

<!-- example:glossary-namespace-storage -->

</details>

### Runtime.App.Store

<a id="word-authoring-runtime-app-store"></a>

<!-- word:authoring:Runtime.App.Store -->

Mutably borrow `Value.invocation` so the caller can update that exact native field.

<!-- word-definition:authoring:Runtime.App.Store -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

This statement is used inside `Runtime.App.Acquire`. Its enclosing native signature and preceding bindings remain required.

<!-- example:glossary-runtime-app-store-caller -->

</details>

### Runtime.Submit

<a id="word-authoring-runtime-submit"></a>

<!-- word:authoring:Runtime.Submit -->

Send the supplied value through the selected native channel. The selected operation is `Sender.send(Value)`.

<!-- word-definition:authoring:Runtime.Submit -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

This statement is used inside `Runtime.Invocation.Kill`. Its enclosing native signature and preceding bindings remain required.

<!-- example:glossary-runtime-submit-caller -->

</details>

### Runtime.Subscribe

<a id="word-authoring-runtime-subscribe"></a>

<!-- word:authoring:Runtime.Subscribe -->

Subscribe to the selected watch sender and return a receiver for its current and later values. The selected operation is `Value.subscribe()`.

<!-- word-definition:authoring:Runtime.Subscribe -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

This statement is used inside `Runtime.Capture.Follow`. Its enclosing native signature and preceding bindings remain required.

<!-- example:glossary-runtime-subscribe-caller -->

</details>

### Runtime.Supervision.Success

<a id="word-authoring-runtime-supervision-success"></a>

<!-- word:authoring:Runtime.Supervision.Success -->

Check whether the supplied native process status denotes success. The selected operation is `Status.success()`.

<!-- word-definition:authoring:Runtime.Supervision.Success -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

This statement is used inside `Runtime.Supervision.Inspect`. Its enclosing native signature and preceding bindings remain required.

<!-- example:glossary-runtime-supervision-success-caller -->

</details>

### Storage.Success

<a id="word-authoring-storage-success"></a>

<!-- word:authoring:Storage.Success -->

Construct native Ok around the supplied success payload without returning from the enclosing Function.

<!-- word-definition:authoring:Storage.Success -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

This statement is used inside `Runtime.Invocation.Claim`. Its enclosing native signature and preceding bindings remain required.

<!-- example:glossary-storage-success-caller -->

</details>

### Storage.Sum

<a id="word-authoring-storage-sum"></a>

<!-- word:authoring:Storage.Sum -->

Add the supplied native operands and bind the result; their concrete types determine the applicable addition operation.

<!-- word-definition:authoring:Storage.Sum -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

This statement is used inside `Formatting.Scale`. Its enclosing native signature and preceding bindings remain required.

<!-- example:glossary-storage-sum-caller -->

</details>

### Runtime.App.Supervise

<a id="word-authoring-runtime-app-supervise"></a>

<!-- word:authoring:Runtime.App.Supervise -->

Record a launched child's process ID, register its invocation as runtime-owned, and spawn its supervisor with capture and termination-request ownership.

<!-- word-definition:authoring:Runtime.App.Supervise -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

This statement is used inside `Runtime.App.Run`. Its enclosing native signature and preceding bindings remain required.

<!-- example:glossary-runtime-app-progress-caller -->

</details>

### Runtime.Invocation.Supervise

<a id="word-authoring-runtime-invocation-supervise"></a>

<!-- word:authoring:Runtime.Invocation.Supervise -->

Own the child wait, capture stdout and stderr, service termination requests, bound post-exit draining, perform final group cleanup, and publish one combined invocation Snapshot.

<!-- word-definition:authoring:Runtime.Invocation.Supervise -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

This statement is used inside `Runtime.App.Supervise`. Its enclosing native signature and preceding bindings remain required.

<!-- example:glossary-runtime-invocation-supervise-caller -->

</details>

### Runtime.Supervision

<a id="word-authoring-runtime-supervision"></a>

<!-- word:authoring:Runtime.Supervision -->

Name the Runtime namespace owning child waits, leader-termination requests, stream finalization and invocation publication.

<!-- word-definition:authoring:Runtime.Supervision -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

This statement is used inside `Runtime.Supervision.Inspect`. Its enclosing native signature and preceding bindings remain required.

<!-- example:glossary-runtime-supervision-caller -->

</details>

### Storage.Function.Suspend

<a id="word-authoring-storage-function-suspend"></a>

<!-- word:authoring:Storage.Function.Suspend -->

Construct a move future, pin it and yield the pinned value as the outer construction result. The enclosed Function body executes only when its future is polled.

<!-- word-definition:authoring:Storage.Function.Suspend -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

**Compiler use.** functionOperation selects Storage.Function.Suspend and supplies the composed body as a typed Process definition. This retained compiler excerpt selects the L-owned definition with typed native name/type/value or Process bindings. Those bindings come from the compiler’s resolved plan; this is an integration fragment, not an ordinary standalone L call. The `functionOperation` helper in `lyger/function_composition.go` selects `Storage.Function.` plus the supplied member name and preserves the selected Return composition.

<!-- example:glossary-lyger-function-composition-394-402 -->

</details>

## T

<a id="glossary-t"></a>

### Runtime.Take

<a id="word-authoring-runtime-take"></a>

<!-- word:authoring:Runtime.Take -->

Take the selected optional value, leaving None in its original location. The selected operation is `Value.take()`.

<!-- word-definition:authoring:Runtime.Take -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

This statement is used inside `Runtime.Unsubscribe`. Its enclosing native signature and preceding bindings remain required.

<!-- example:glossary-runtime-take-caller -->

</details>

### Take

<a id="word-builtin-take"></a>

<!-- word:builtin:take -->

Recognize a configured reductive action without a general supported removal instruction.

<!-- word-definition:builtin:take -->

**Unsupported operation.** The configured Take action has no executable instruction and is rejected during instruction lowering. The [collection contract](../spec/collections-and-categories.md) provides insertion, replacement and typed reads, not removal. A separately authored action must supply its own definition. The configured word is not a runnable removal example.

**Rejected source illustration (not executed).** With a List named Items, `Take 0 from items.` illustrates the requested removal operation. The configured word supplies no supported removal behavior. This is an invalid illustration, not a verified parser diagnostic or a runnable alternative to Put/Place.

### Storage.Pointer.Target

<a id="word-authoring-storage-pointer-target"></a>

<!-- word:authoring:Storage.Pointer.Target -->

Borrow `Value.target` without moving the referenced native field.

<!-- word-definition:authoring:Storage.Pointer.Target -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

This statement is used inside `Storage.Pointer.Reach`. Its enclosing native signature and preceding bindings remain required.

<!-- example:glossary-storage-pointer-caller -->

</details>

### Storage.Value.Target

<a id="word-authoring-storage-value-target"></a>

<!-- word:authoring:Storage.Value.Target -->

Read the target ownership identity from the first traversal-context component.

<!-- word-definition:authoring:Storage.Value.Target -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

This statement is used inside `Storage.Value.Follow`. Its enclosing native signature and preceding bindings remain required.

<!-- example:glossary-storage-value-target-caller -->

</details>

### Runtime.Host.Terminal

<a id="word-authoring-runtime-host-terminal"></a>

<!-- word:authoring:Runtime.Host.Terminal -->

Name the adapter namespace reusing canonical Terminal copying for native host integration.

<!-- word-definition:authoring:Runtime.Host.Terminal -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

After importing Runtime, bind the existing namespace for qualified member access.

<!-- example:glossary-namespace-runtime-host-terminal -->

</details>

### Runtime.State.Terminal

<a id="word-authoring-runtime-state-terminal"></a>

<!-- word:authoring:Runtime.State.Terminal -->

Mutably borrow `Value.terminal` so the caller can update that exact native field.

<!-- word-definition:authoring:Runtime.State.Terminal -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

This statement is used inside `Runtime.Choose`. Its enclosing native signature and preceding bindings remain required.

<!-- example:glossary-runtime-state-terminal-caller -->

</details>

### Runtime.Terminal

<a id="word-authoring-runtime-terminal"></a>

<!-- word:authoring:Runtime.Terminal -->

Declare the native `Terminal` record with fields `status: u8`, `error: Option<Error>`.

<!-- word-definition:authoring:Runtime.Terminal -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

This statement is used inside `Runtime.Terminal.Copy`. Its enclosing native signature and preceding bindings remain required.

<!-- example:glossary-runtime-terminal-caller -->

</details>

### Runtime.Invocation.Terminate

<a id="word-authoring-runtime-invocation-terminate"></a>

<!-- word:authoring:Runtime.Invocation.Terminate -->

Await Invocation.Signal and convert any native termination failure into a runtime Error with process-group context.

<!-- word-definition:authoring:Runtime.Invocation.Terminate -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

This statement is used inside `Runtime.Invocation.Settle`. Its enclosing native signature and preceding bindings remain required.

<!-- example:glossary-runtime-invocation-terminate-caller -->

</details>

### Runtime.Supervision.Terminate

<a id="word-authoring-runtime-supervision-terminate"></a>

<!-- word:authoring:Runtime.Supervision.Terminate -->

Signal the supplied leader process ID and convert a native failure to a runtime Error; an absent ID succeeds without signaling.

<!-- word-definition:authoring:Runtime.Supervision.Terminate -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

This statement is used inside `Runtime.Supervision.Poll`. Its enclosing native signature and preceding bindings remain required.

<!-- example:glossary-runtime-supervision-terminate-caller -->

</details>

### Storage.Test

<a id="word-authoring-storage-test"></a>

<!-- word:authoring:Storage.Test -->

Prefix a supplied native Function composition with `#[cfg(test)] pub(crate)` so it is emitted only for native test builds with crate visibility.

<!-- word-definition:authoring:Storage.Test -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

**Compiler use.** Wrap the generated observer Function definition in the native test-only declaration. This retained compiler excerpt selects the L-owned definition with typed native name/type/value or Process bindings. Those bindings come from the compiler’s resolved plan; this is an integration fragment, not an ordinary standalone L call.

<!-- example:glossary-lyger-execution-composition-133-139 -->

</details>

### Runtime.Text

<a id="word-authoring-runtime-text"></a>

<!-- word:authoring:Runtime.Text -->

Borrow the selected native String as a string slice. The selected operation is `Value.as_str()`.

<!-- word-definition:authoring:Runtime.Text -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

This statement is used inside `Runtime.App.Describe`. Its enclosing native signature and preceding bindings remain required.

<!-- example:glossary-runtime-text-caller -->

</details>

### Storage.Text

<a id="word-authoring-storage-text"></a>

<!-- word:authoring:Storage.Text -->

Name the Storage namespace for exact byte ownership, joining, delimiters and erased Text display.

<!-- word-definition:authoring:Storage.Text -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

This statement is used inside `Storage.App.Render`. Its enclosing native signature and preceding bindings remain required.

<!-- example:glossary-storage-text-caller -->

</details>

### Text

<a id="word-type-text"></a>

<!-- word:type:text -->

Represent a sequence of bytes without an implicit newline or Unicode normalization.

<!-- word-definition:type:text -->

**Usage and example.** Backtick literals initialize Text. The [public-values program](../../demo/paragraphs/public-values.l) stores and shows Greeting. The escape illustration below preserves its precise byte rules.

Backticks delimit Text. Interior bytes are retained, except that a backslash immediately before an escaped backtick is removed when the literal is unquoted. This is not general escape decoding: backslash followed by `n` remains two characters. Text has no implicit newline. `Show` writes it exactly, while imported `IO.Print` appends one LF. See the [IO reference](../stdlib/eng/io.md).

Text preserves UTF-8, invalid UTF-8 octets, NUL, whitespace, and markup punctuation. Native encoding preserves these bytes without replacement characters. Browser displays cannot faithfully represent every octet. Use the original source bytes for byte-sensitive examples.

Text supports equality with Text. General ordering, numeric conversion, indexing, concatenation operators, code-point iteration, and Unicode normalization are not currently exposed as L operations. Native definitions can supply target-specific operations.

<!-- example:lexical-text-escapes -->

The first Text contains a backtick, the second retains the two bytes `\n`, and the third retains Unicode, braces, and an interior `<DEF>` marker as data. Successive Show calls write adjacent text. This fragment does not define a Bytes constructor or an escape for arbitrary octets.

### Than

<a id="word-grammar-word-than"></a>

<!-- word:grammar-word:than -->

Connect the operands of an admitted ordered Number comparison.

<!-- word-definition:grammar-word:than -->

**Usage and boundary.** The focused `left less than right` and `left greater than right` forms follow the [comparison contract](../spec/control-flow.md#comparisons). Both operands must be Number. Than does not enable Text ordering.

**Usage example.** The [Less example](#word-grammar-word-less) supplies two declared Numbers and uses `than` between them. The [Greater example](#word-grammar-word-greater) shows its corresponding ordered comparison. Both are labeled source fragments.

### That

<a id="word-grammar-word-that"></a>

<!-- word:grammar-word:that -->

Refer to the last subject defined in the original source paragraph, or the applicable declared source-action argument.

<!-- word-definition:grammar-word:that -->

**Usage and boundary.** Within a paragraph that has just defined Job, `that` refers to Job under the shared [pronoun rules](../spec/names-and-scope.md#pronouns). It can replace `it` in the [Job illustration](#word-grammar-word-it).

**Usage example (source fragment, not executed here).** These sentences share one paragraph, so That selects the Text subject declared immediately before it.

```l
Define "message" as `hello`.
Show that.
```

### The

<a id="word-grammar-word-the"></a>

<!-- word:grammar-word:the -->

Refer definitively to the selected subject or value in its owning source context.

<!-- word-definition:grammar-word:the -->

**Usage and boundary.** `The App Prints message` in the [first program](../guide/first-program.md#read-the-program) supplies an App body effect. The does not make a private or unresolved name visible.

### Then

<a id="word-grammar-word-then"></a>

<!-- word:grammar-word:then -->

Separate successive actions within a composed source sequence.

<!-- word-definition:grammar-word:then -->

**Usage and boundary.** Launch in the [local definitions program](../../demo/paragraphs/local-definitions.l) uses `and then Awaiting it` after Build/Run. Then preserves source sequencing. It is not an asynchronous continuation primitive.

### Thing

<a id="word-type-thing"></a>

<!-- word:type:thing -->

Retain a value's concrete runtime type at an explicitly broad typed boundary.

<!-- word-definition:type:thing -->

**Usage and example.** ``Define "boxed" as a New Thing of `first`.`` followed by a checked Text read is the [runtime-value witness](../../lyger/runtime_values_test.go). The Print illustration below shows that a broad parameter still selects a supported concrete renderer.

Thing preserves a contained runtime type and checks typed reads. It does not coerce values into Text. IO.Print selects typed rendering definitions for Text, Number, Truth and Null:

<!-- example:values-thing-printing -->

The public Print action composes private rendering and writing Functions. This fragment supplies an App body. Declaration alone does not run it.

Thing does not provide dynamically typed arithmetic, unrestricted reflection, or printing for every application or collection value. Its usable arguments depend on the definition and native consumer.

### This

<a id="word-grammar-word-this"></a>

<!-- word:grammar-word:this -->

Refer to the last subject defined in the original source paragraph, or the applicable declared source-action argument.

<!-- word-definition:grammar-word:this -->

**Usage and boundary.** Within a paragraph that has just defined Job, `this` refers to Job under the shared [pronoun rules](../spec/names-and-scope.md#pronouns). It can replace `it` in the [Job illustration](#word-grammar-word-it).

**Usage example (source fragment, not executed here).** These sentences share one paragraph, so This selects the Text subject declared immediately before it.

```l
Define "message" as `hello`.
Show this.
```

### To

<a id="word-grammar-word-to"></a>

<!-- word:grammar-word:to -->

Introduce a native-backed action definition or a target-dependent utilization/comparison phrase.

<!-- word-definition:grammar-word:to -->

**Usage and boundary.** The [IO source](../../stdlib/eng/io/dictionary.l) defines Flush with `to` and one framed native flush call. In `less than equal to`, To belongs to the numeric comparison. It is not an independent conversion or transfer.

### Runtime.Error.Trace

<a id="word-authoring-runtime-error-trace"></a>

<!-- word:authoring:Runtime.Error.Trace -->

Return an Error's retained native cause when present, otherwise search its child Errors in order for a native cause.

<!-- word-definition:authoring:Runtime.Error.Trace -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>


The wrapper below imports `Storage` and `Runtime` and is a member of the example `GlossaryExamples` Dictionary. Its declared native inputs, outputs, lifetime/generic descriptors, and the original selected module's native types/support are required. Lyger resolves the source and composes its library AML; this fragment has no native compilation or execution claim.

<!-- example:glossary-native-runtime-error-trace -->

</details>

Existing qualified aliases: `Runtime.Host.Error.Trace`. They select this canonical meaning.

### Storage.Traverse

<a id="word-authoring-storage-traverse"></a>

<!-- word:authoring:Storage.Traverse -->

Visit values in iterator order and invoke the supplied callback with item and context, discarding every callback result, including returned errors.

<!-- word-definition:authoring:Storage.Traverse -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

This statement is used inside [Storage.Slot.Allocate](#word-authoring-storage-slot-allocate). Its enclosing native signature, preceding local bindings and selected native module are required.

<!-- example:glossary-builtins-eng-storage-program-23-23 -->

</details>

### Storage.Truth

<a id="word-authoring-storage-truth"></a>

<!-- word:authoring:Storage.Truth -->

Name the Storage namespace for native Boolean display.

<!-- word-definition:authoring:Storage.Truth -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

This statement is used inside `Storage.Truth.Render`. Its enclosing native signature and preceding bindings remain required.

<!-- example:glossary-storage-truth-caller -->

</details>

### Truth

<a id="word-type-truth"></a>

<!-- word:type:truth -->

Represent the Boolean values true and false without implicit truthiness conversion.

<!-- word-definition:type:truth -->

**Usage and example.** The complete conditional example below compares a declared Truth and joins Worker only on its selected branch.

The literals are lowercase `true` and `false`. A bare condition requires Truth, and equality requires another Truth. Number and Text have no implicit truthiness. The grammar's `?~?` notation is a type matcher, not a source literal.

Configured `not`/negative forms express negation in conditions, not a general prefix operator. Joined Boolean conditions are currently rejected by outer instruction lowering. No short-circuit behavior is defined.

<!-- example:values-truth-condition -->

`enabled` receives a Truth. Comparing it with `true` selects `await Worker`. The following declaration snapshots the finalized Output. Replacing `true` with backtick Text changes the type and does not create truthiness.

### Storage.Tuple

<a id="word-authoring-storage-tuple"></a>

<!-- word:authoring:Storage.Tuple -->

Construct a native tuple from the supplied values in their supplied order.

<!-- word-definition:authoring:Storage.Tuple -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

This statement is used inside `Storage.Function.Copy`. Its enclosing native signature and preceding bindings remain required.

<!-- example:glossary-storage-tuple-caller -->

</details>

### Storage.Environment.Types

<a id="word-authoring-storage-environment-types"></a>

<!-- word:authoring:Storage.Environment.Types -->

Borrow `Value.types` without moving the referenced native field.

<!-- word-definition:authoring:Storage.Environment.Types -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

This statement is used inside `Storage.Environment.Copy`. Its enclosing native signature and preceding bindings remain required.

<!-- example:glossary-storage-environment-caller -->

</details>

### Storage.Program.Types

<a id="word-authoring-storage-program-types"></a>

<!-- word:authoring:Storage.Program.Types -->

Borrow `Program.types` without moving the referenced native field.

<!-- word-definition:authoring:Storage.Program.Types -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

This statement is used inside [Storage.Program.Values.Write](#word-authoring-storage-program-values-write). Its enclosing native signature, preceding local bindings and selected native module are required.

<!-- example:glossary-builtins-eng-storage-program-61-61 -->

</details>

## U

<a id="glossary-u"></a>

### Storage.Signature.Universal

<a id="word-authoring-storage-signature-universal"></a>

<!-- word:authoring:Storage.Signature.Universal -->

Bind the borrowed signature slice `[1]` used to represent the universal Thing signature.

<!-- word-definition:authoring:Storage.Signature.Universal -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

This statement is used inside `Storage.Signature.Match`. Its enclosing native signature and preceding bindings remain required.

<!-- example:glossary-storage-signature-caller -->

</details>

### Runtime.Unsubscribe

<a id="word-authoring-runtime-unsubscribe"></a>

<!-- word:authoring:Runtime.Unsubscribe -->

Take and clear the runtime's optional signal-watcher task while holding its lock.

<!-- word-definition:authoring:Runtime.Unsubscribe -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

This statement is used inside `Runtime.Quell`. Its enclosing native signature and preceding bindings remain required.

<!-- example:glossary-runtime-unsubscribe-caller -->

</details>

### Storage.Unwrap

<a id="word-authoring-storage-unwrap"></a>

<!-- word:authoring:Storage.Unwrap -->

Consume an Option or Result and return its payload, or the supplied fallback. The fallback argument is evaluated normally, including when the payload is present.

<!-- word-definition:authoring:Storage.Unwrap -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

This statement is used inside `Runtime.Invocation.Kill`. Its enclosing native signature and preceding bindings remain required.

<!-- example:glossary-storage-unwrap-caller -->

</details>

### Storage.Environment.Update

<a id="word-authoring-storage-environment-update"></a>

<!-- word:authoring:Storage.Environment.Update -->

Reject a capture write that would create an ownership cycle, acquire mutable access, validate the slot and its declared type, then replace the captured value.

<!-- word-definition:authoring:Storage.Environment.Update -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

The wrapper below imports `Storage` and `Runtime` and is a member of the example `GlossaryExamples` Dictionary. Its declared native inputs, outputs, lifetime/generic descriptors, and the original selected module's native types/support are required. Lyger resolves the source and composes its library AML; this fragment has no native compilation or execution claim.

<!-- example:glossary-native-storage-environment-update -->

</details>

### Storage.List.Update

<a id="word-authoring-storage-list-update"></a>

<!-- word:authoring:Storage.List.Update -->

Validate the replacement value's element signature and numeric index, check bounds, then replace the selected List element.

<!-- word-definition:authoring:Storage.List.Update -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

The wrapper below imports `Storage` and `Runtime` and is a member of the example `GlossaryExamples` Dictionary. Its declared native inputs, outputs, lifetime/generic descriptors, and the original selected module's native types/support are required. Lyger resolves the source and composes its library AML; this fragment has no native compilation or execution claim.

<!-- example:glossary-native-storage-list-update -->

</details>

### Storage.Map.Update

<a id="word-authoring-storage-map-update"></a>

<!-- word:authoring:Storage.Map.Update -->

Check a replacement value against the Map's element signature and insert it under the supplied owned byte key.

<!-- word-definition:authoring:Storage.Map.Update -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

The wrapper below imports `Storage` and `Runtime` and is a member of the example `GlossaryExamples` Dictionary. Its declared native inputs, outputs, lifetime/generic descriptors, and the original selected module's native types/support are required. Lyger resolves the source and composes its library AML; this fragment has no native compilation or execution claim.

<!-- example:glossary-native-storage-map-update -->

</details>

### Storage.Pointer.Update

<a id="word-authoring-storage-pointer-update"></a>

<!-- word:authoring:Storage.Pointer.Update -->

Check the target type and reject ownership cycles before acquiring mutable access and replacing a Pointer's stored value.

<!-- word-definition:authoring:Storage.Pointer.Update -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

The wrapper below imports `Storage` and `Runtime` and is a member of the example `GlossaryExamples` Dictionary. Its declared native inputs, outputs, lifetime/generic descriptors, and the original selected module's native types/support are required. Lyger resolves the source and composes its library AML; this fragment has no native compilation or execution claim.

<!-- example:glossary-native-storage-pointer-update -->

</details>

### Use

<a id="word-builtin-use"></a>

<!-- word:builtin:use -->

Associate an existing canonical declaration with an alias, or load a supported explicit source or indexed dictionary dependency.

<!-- word-definition:builtin:use -->

**Usage and example.** The member import below selects IO.Print and names its local alias. `Use Worker as "Job"` in the [pronoun example](../../testdata/phase1/language/lang-06.l) instead aliases an existing App. Neither form copies its meaning.

Bare dictionary names resolve through the selected `dictionaries.json`. The current format is version 1, with language keys containing dictionary keys. Each entry names a repository/resource-relative `.l` entry and the registered Dictionary it exports.

For example:

<!-- example:imports-indexed-member -->

This imports the public member from the registered IO dictionary and binds a local alias. It does not search for a neighboring `io` directory or `io.l`, nor does it expose private rendering and writing helpers. Importing the dictionary itself retains a namespace instead of flattening all members into the program.

The registry validates duplicate JSON keys, supported source roots, safe resource paths, format, actual entry files, and consistent dictionary identity. It selects foundational dictionaries under `builtins/<language>/`, runtime source under `runtime/<language>/`, and standard libraries under `stdlib/<language>/`. There is one selected registry and no legacy registry lookup. An unknown key is an unresolved dictionary. An unsupported provider-style name reports its bounded unsupported-provider diagnostic. There is no package download or unrestricted remote dependency resolver hidden in Use.

Each indexed dictionary owns a private lexical root inheriting builtins. Its public export must resolve to the registered Dictionary. Imported member aliases resolve in that dictionary's environment. Helper definitions do not capture caller-added members. Repeated dictionary dependencies reuse canonical dictionary identity while preserving import-site attribution.

## V

<a id="glossary-v"></a>

### Runtime.Vacancy

<a id="word-authoring-runtime-vacancy"></a>

<!-- word:authoring:Runtime.Vacancy -->

Check whether the selected native string or path-string value is empty. The selected operation is `Value.is_empty()`.

<!-- word-definition:authoring:Runtime.Vacancy -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

This statement is used inside `Runtime.App.Describe`. Its enclosing native signature and preceding bindings remain required.

<!-- example:glossary-runtime-vacancy-caller -->

</details>

### Runtime.Valid

<a id="word-authoring-runtime-valid"></a>

<!-- word:authoring:Runtime.Valid -->

Check whether the selected native Result contains a success. The selected operation is `Value.is_ok()`.

<!-- word-definition:authoring:Runtime.Valid -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

This statement is used inside `Runtime.Normalize`. Its enclosing native signature and preceding bindings remain required.

<!-- example:glossary-runtime-valid-caller -->

</details>

### Storage.Validate

<a id="word-authoring-storage-validate"></a>

<!-- word:authoring:Storage.Validate -->

Return successful unit when a condition is true; otherwise construct and return a runtime error from the supplied message.

<!-- word-definition:authoring:Storage.Validate -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

This statement is used inside `Storage.Value.Check`. Its enclosing native signature and preceding bindings remain required.

<!-- example:glossary-storage-validate-caller -->

</details>

### Storage.Value

<a id="word-authoring-storage-value"></a>

<!-- word:authoring:Storage.Value -->

Declare the native `LValue` record with fields `ty: Vec<u64>`, `data: std::rc::Rc<dyn std::any::Any>`, `show: fn(&dyn std::any::Any) -> runtime::Result<Vec<u8>>`, `reaches: fn(&dyn std::any::Any, usize, &mut std::collections::BTreeSet<usize>) -> runtime::Result<bool>`.

<!-- word-definition:authoring:Storage.Value -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

This statement is used inside `Storage.Pointer.Render`. Its enclosing native signature and preceding bindings remain required.

<!-- example:glossary-storage-value-caller -->

</details>

### Storage.Environment.Values

<a id="word-authoring-storage-environment-values"></a>

<!-- word:authoring:Storage.Environment.Values -->

Borrow `Value.values` without moving the referenced native field.

<!-- word-definition:authoring:Storage.Environment.Values -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

This statement is used inside `Storage.Map.Read`. Its enclosing native signature and preceding bindings remain required.

<!-- example:glossary-storage-environment-values-caller -->

</details>

Existing qualified aliases: `Storage.List.Values`, `Storage.Map.Values`. They select this canonical meaning.

### Storage.Program.Values

<a id="word-authoring-storage-program-values"></a>

<!-- word:authoring:Storage.Program.Values -->

Borrow `Program.values` without moving the referenced native field.

<!-- word-definition:authoring:Storage.Program.Values -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

This statement is used inside [Storage.Program.Values.Read](#word-authoring-storage-program-values-read). Its enclosing native signature, preceding local bindings and selected native module are required.

<!-- example:glossary-builtins-eng-storage-program-53-53 -->

</details>

### Storage.Vector

<a id="word-authoring-storage-vector"></a>

<!-- word:authoring:Storage.Vector -->

Construct an owned vector from the supplied native array values. The selected operation is `Vec::from([Values])`.

<!-- word-definition:authoring:Storage.Vector -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

This statement is used inside `Runtime.Error.Attach`. Its enclosing native signature and preceding bindings remain required.

<!-- example:glossary-storage-vector-caller -->

</details>

### Verb

<a id="word-type-verb"></a>

<!-- word:type:verb -->

Classify a word that expresses an action or relationship with explicit operands.

<!-- word-definition:type:verb -->

**Usage and example.** The [IO source](../../stdlib/eng/io/dictionary.l) declares its private Render family with `as a New Verb`, then supplies typed rendering implementations. For an application-level source verb, Launch in the [source-action example](../guide/definitions.md#compose-a-source-action) composes Build, Run and Await. A verb's actor noun constructs that same selected meaning.

### Storage.Verify

<a id="word-authoring-storage-verify"></a>

<!-- word:authoring:Storage.Verify -->

Apply native `?` to the supplied result and discard its successful payload, propagating failure from the enclosing Function.

<!-- word-definition:authoring:Storage.Verify -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

This statement is used inside `Storage.Environment.Create`. Its enclosing native signature and preceding bindings remain required.

<!-- example:glossary-storage-verify-caller -->

</details>

### Storage.Environment.Visit

<a id="word-authoring-storage-environment-visit"></a>

<!-- word:authoring:Storage.Environment.Visit -->

Project an erased payload to a capture environment, then perform Environment.Reach with the supplied target and visited identities.

<!-- word-definition:authoring:Storage.Environment.Visit -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

The wrapper below imports `Storage` and `Runtime` and is a member of the example `GlossaryExamples` Dictionary. Its declared native inputs, outputs, lifetime/generic descriptors, and the original selected module's native types/support are required. Lyger resolves the source and composes its library AML; this fragment has no native compilation or execution claim.

<!-- example:glossary-native-storage-environment-visit -->

</details>

### Storage.Pointer.Visit

<a id="word-authoring-storage-pointer-visit"></a>

<!-- word:authoring:Storage.Pointer.Visit -->

Project an erased payload to a Pointer, then perform Pointer.Reach using the supplied target and traversal history.

<!-- word-definition:authoring:Storage.Pointer.Visit -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

The wrapper below imports `Storage` and `Runtime` and is a member of the example `GlossaryExamples` Dictionary. Its declared native inputs, outputs, lifetime/generic descriptors, and the original selected module's native types/support are required. Lyger resolves the source and composes its library AML; this fragment has no native compilation or execution claim.

<!-- example:glossary-native-storage-pointer-visit -->

</details>

### Storage.Value.Visitor

<a id="word-authoring-storage-value-visitor"></a>

<!-- word:authoring:Storage.Value.Visitor -->

Select the native field `Value.reaches` by value. Its declared type determines whether this copies or moves the field.

<!-- word-definition:authoring:Storage.Value.Visitor -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

This statement is used inside `Storage.Value.Reach`. Its enclosing native signature and preceding bindings remain required.

<!-- example:glossary-storage-value-visitor-caller -->

</details>

## W

<a id="glossary-w"></a>

### Runtime.App.Wait

<a id="word-authoring-runtime-app-wait"></a>

<!-- word:authoring:Runtime.App.Wait -->

Require a current invocation, wait for either runtime cancellation or invocation completion, and select the completed invocation. Continuation claiming is a separate operation.

<!-- word-definition:authoring:Runtime.App.Wait -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

The wrapper below imports `Storage` and `Runtime` and is a member of the example `GlossaryExamples` Dictionary. Its declared native inputs, outputs, lifetime/generic descriptors, and the original selected module's native types/support are required. Lyger resolves the source and composes its library AML; this fragment has no native compilation or execution claim.

<!-- example:glossary-native-runtime-app-wait -->

</details>

### Runtime.Supervision.Wait

<a id="word-authoring-runtime-supervision-wait"></a>

<!-- word:authoring:Runtime.Supervision.Wait -->

Construct the single host-owned child-wait future used by invocation supervision. Driving it, deadline policy and cleanup belong to the supervising source Functions.

<!-- word-definition:authoring:Runtime.Supervision.Wait -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

This statement is used inside `Runtime.Invocation.Supervise`. Its enclosing native signature and preceding bindings remain required.

<!-- example:glossary-runtime-supervision-wait-caller -->

</details>

### Runtime.Wait

<a id="word-authoring-runtime-wait"></a>

<!-- word:authoring:Runtime.Wait -->

Wait for a Boolean watch value to become true or for its notification channel to close.

<!-- word-definition:authoring:Runtime.Wait -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

This statement is used inside `Runtime.App.Wait`. Its enclosing native signature and preceding bindings remain required.

<!-- example:glossary-runtime-wait-caller -->

</details>

### Runtime.Signals.Watch

<a id="word-authoring-runtime-signals-watch"></a>

<!-- word:authoring:Runtime.Signals.Watch -->

Await a native signal event, upgrade the runtime owner if still available, notify host observation, and request the corresponding terminal status or an attributed reception failure.

<!-- word-definition:authoring:Runtime.Signals.Watch -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

This statement is used inside `Runtime.Install`. Its enclosing native signature and preceding bindings remain required.

<!-- example:glossary-runtime-signals-watch-caller -->

</details>

### Runtime.Watch

<a id="word-authoring-runtime-watch"></a>

<!-- word:authoring:Runtime.Watch -->

Create a native watch channel initialized with the supplied value. The selected operation is `tokio::sync::watch::channel(Value)`.

<!-- word-definition:authoring:Runtime.Watch -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

This statement is used inside `Runtime.Invocation.Allocate`. Its enclosing native signature and preceding bindings remain required.

<!-- example:glossary-runtime-watch-caller -->

</details>

### Runtime.Weak

<a id="word-authoring-runtime-weak"></a>

<!-- word:authoring:Runtime.Weak -->

Create a weak reference to the supplied Arc allocation. The selected operation is `std::sync::Arc::downgrade(Value)`.

<!-- word-definition:authoring:Runtime.Weak -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

This statement is used inside `Runtime.App.Acquire`. Its enclosing native signature and preceding bindings remain required.

<!-- example:glossary-runtime-weak-caller -->

</details>

### While

<a id="word-grammar-word-while"></a>

<!-- word:grammar-word:while -->

Name a configured conditional loop form with no executable source implementation.

<!-- word-definition:grammar-word:while -->

**Unsupported operation.** The [unsupported loop illustration](../examples/control-while-unsupported.l) is rejected during grammar processing, and no outer While instruction is implemented. It is a labeled fragment, not a runnable loop or a promise of native iteration. Advanced [Storage traversal](#context-storage--iteration-and-continuations) is a separate source composition facility.

### With

<a id="word-grammar-word-with"></a>

<!-- word:grammar-word:with -->

Introduce associated structures or signature components in the selected construction.

<!-- word-definition:grammar-word:with -->

**Usage and boundary.** The [Function example](../examples/composed-functions.l) uses `with Inputs`, and the [Main construction](../spec/lexical-types.md#describe-with-an-adjective) uses `with a New Main Object`. With preserves the owning construction's argument contract.

### Runtime.Error.Wrap

<a id="word-authoring-runtime-error-wrap"></a>

<!-- word:authoring:Runtime.Error.Wrap -->

Construct ErrorData from message bytes, child Errors, cancellation state and an optional native cause, then place it in an Arc-backed Error handle.

<!-- word-definition:authoring:Runtime.Error.Wrap -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

This statement is used inside `Runtime.Error.Explain`. Its enclosing native signature and preceding bindings remain required.

<!-- example:glossary-runtime-error-wrap-caller -->

</details>

### Storage.Value.Wrap

<a id="word-authoring-storage-value-wrap"></a>

<!-- word:authoring:Storage.Value.Wrap -->

Construct a shared dynamic value containing the supplied T value, signature, renderer, and visitor. The source stores those supplied components without checking their mutual consistency.

<!-- word-definition:authoring:Storage.Value.Wrap -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

The wrapper below imports `Storage` and `Runtime` and is a member of the example `GlossaryExamples` Dictionary. Its declared native inputs, outputs, lifetime/generic descriptors, and the original selected module's native types/support are required. Lyger resolves the source and composes its library AML; this fragment has no native compilation or execution claim.

<!-- example:glossary-native-storage-value-wrap -->

</details>

### Runtime.Capture.Write

<a id="word-authoring-runtime-capture-write"></a>

<!-- word:authoring:Runtime.Capture.Write -->

Write the selected capture-buffer prefix through the supplied native writer, returning its IO result. The selected operation is `std::io::Write::write_all(Destination, &Buffer[..Count])`.

<!-- word-definition:authoring:Runtime.Capture.Write -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

This statement is used inside `Runtime.Capture.Receive`. Its enclosing native signature and preceding bindings remain required.

<!-- example:glossary-runtime-capture-write-caller -->

</details>

### Runtime.Error.Write

<a id="word-authoring-runtime-error-write"></a>

<!-- word:authoring:Runtime.Error.Write -->

Write the supplied native text through the selected formatter and return its formatting result. The selected operation is `Formatter.write_str(&Text)`.

<!-- word-definition:authoring:Runtime.Error.Write -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

This statement is used inside `Runtime.Error.Display`. Its enclosing native signature and preceding bindings remain required.

<!-- example:glossary-runtime-error-write-caller -->

</details>

### Runtime.Write

<a id="word-authoring-runtime-write"></a>

<!-- word:authoring:Runtime.Write -->

Create an async operation that writes the supplied bytes through the selected output writer. The selected operation is `tokio::io::AsyncWriteExt::write_all(Output, Bytes)`.

<!-- word-definition:authoring:Runtime.Write -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

This statement is used inside `Runtime.Show`. Its enclosing native signature and preceding bindings remain required.

<!-- example:glossary-runtime-write-caller -->

</details>

### Storage.Program.Actions.Write

<a id="word-authoring-storage-program-actions-write"></a>

<!-- word:authoring:Storage.Program.Actions.Write -->

Acquire mutable access to the program's action slots and replace the selected slot with a present capture environment.

<!-- word-definition:authoring:Storage.Program.Actions.Write -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

The wrapper below imports `Storage` and `Runtime` and is a member of the example `GlossaryExamples` Dictionary. Its declared native inputs, outputs, lifetime/generic descriptors, and the original selected module's native types/support are required. Lyger resolves the source and composes its library AML; this fragment has no native compilation or execution claim.

<!-- example:glossary-native-storage-program-actions-write -->

</details>

### Storage.Program.Values.Write

<a id="word-authoring-storage-program-values-write"></a>

<!-- word:authoring:Storage.Program.Values.Write -->

Check the program slot and expected value signature, then store the supplied value in that slot through Slot.Write.

<!-- word-definition:authoring:Storage.Program.Values.Write -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

The wrapper below imports `Storage` and `Runtime` and is a member of the example `GlossaryExamples` Dictionary. Its declared native inputs, outputs, lifetime/generic descriptors, and the original selected module's native types/support are required. Lyger resolves the source and composes its library AML; this fragment has no native compilation or execution claim.

<!-- example:glossary-native-storage-program-values-write -->

</details>

### Storage.Slot.Write

<a id="word-authoring-storage-slot-write"></a>

<!-- word:authoring:Storage.Slot.Write -->

Acquire mutable slot-vector access, require the selected slot, and replace it with Some of the supplied value. Return unit on success.

<!-- word-definition:authoring:Storage.Slot.Write -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

This statement is used inside [Storage.Program.Values.Write](#word-authoring-storage-program-values-write). Its enclosing native signature, preceding local bindings and selected native module are required.

<!-- example:glossary-builtins-eng-storage-program-68-68 -->

</details>

## Y

<a id="glossary-y"></a>

### Storage.Function.Yield

<a id="word-authoring-storage-function-yield"></a>

<!-- word:authoring:Storage.Function.Yield -->

Wrap values in ControlFlow.Continue and success, then apply the selected enclosing Function return definition.

<!-- word-definition:authoring:Storage.Function.Yield -->

<details class="l-word-usage" data-pagefind-ignore>
<summary>Usage example</summary>

**Compiler use.** functionOperation selects Storage.Function.Yield and supplies the evaluated return tuple. This retained compiler excerpt selects the L-owned definition with typed native name/type/value or Process bindings. Those bindings come from the compiler’s resolved plan; this is an integration fragment, not an ordinary standalone L call. The `functionOperation` helper in `lyger/function_composition.go` selects `Storage.Function.` plus the supplied member name and preserves the selected Return composition.

<!-- example:glossary-lyger-function-composition-241-247 -->

</details>

## Builtin vocabulary usage and context

<a id="context-builtin-vocabulary--builtin-reference"></a>

Builtins are available from the selected complete distribution without an explicit Use import. They comprise configured lexical facilities, compiler-typed operations, and definitions authored in `builtins/eng/core.l`. IO.Print belongs to the separately imported [standard library](../stdlib/index.md).

### Builtin vocabulary

<a id="context-builtin-vocabulary--builtin-vocabulary"></a>

<a id="context-builtin-vocabulary--source-defined-facilities"></a>

The [Native Words index](words.md) groups the implicit vocabulary. Entries such as [Function](#word-builtin-function), [Main](#word-builtin-main) and [Retain](#word-builtin-retain) give their declared forms and supported contexts. Actor nouns such as Invoker and Retainer construct uses of their verb. See [word forms](../spec/definitions.md#inflection-declarations) for the declaration notation.

### Compiler-typed operations

<a id="context-builtin-vocabulary--compiler-typed-operations"></a>

| Operation | Operands and behavior |
| --- | --- |
| [Build](#word-builtin-build) | Validate a checked App's prebuilt artifact at runtime |
| [Run](#word-builtin-run) | Launch a checked target without joining |
| [Await](#word-builtin-await) | Join a checked target's finalized result and claim its continuation once |
| [Stop](#word-builtin-stop) | Request termination of a checked target without joining |
| [Show](#word-builtin-show) | Observe supported runtime values without adding a newline |
| [Exit](#word-builtin-exit) | Terminate whole program, with optional supported failure operand |

These operations are verbs. Inflection recognition does not make every grammatical owner valid: native Await inside an App body rejects even though its word form is recognized. The [App execution contract](../cli-build.md#app-execution-semantics) records operation ordering and failure modes.

### Declaration and vocabulary facilities

<a id="context-builtin-vocabulary--declaration-and-vocabulary-facilities"></a>

Create, Define, Describe, and Use establish declarations, typed/native definitions, descriptors, aliases, and supported imports. Public/Publicly controls visibility at one declaration site. New and Generic participate in construction. Noun/Verb/Adjective/Adverb are lexical categories, not native scalar types.

[Collections](../spec/collections-and-categories.md) support Put, Place and typed reads. Get, Take and While remain unsupported operations. The [grammar chapter](../spec/grammar.md) distinguishes executable actions from other configured vocabulary.

### Runtime members

<a id="context-builtin-vocabulary--runtime-members"></a>

Output and Error are noun members: Output contains finalized raw stdout as Text, and Error identifies the selected process operation/invocation failure. Possessive access keeps the canonical target. Input instead names signature parameters and typed Process bindings. It is not an implicit App member.

### Availability and errors

<a id="context-builtin-vocabulary--availability-and-errors"></a>

English builtins and their backend resources are embedded and loaded automatically. Compiler developers can [override the resource bundle](../development-harness.md#compiler-defaults).

Wrong action parameter counts/types, unsupported target owners, unknown descriptor targets, inaccessible source declarations, and cycles are real errors. Builtin definitions do not hide lifecycle misuse, insert Await, trim output, or infer initialization. The [specification](../spec/index.md) separates their source contract from backend-native facilities.

### Sources and witnesses

<a id="context-builtin-vocabulary--sources-and-witnesses"></a>

Sources: [core.l](../../builtins/eng/core.l), [vocabulary](../../lexicon/eng/vocabulary.json), [typed process operations](../../lyger/builtin_operations.go), and [native instruction roles](../../lyger/ir_kinds.go). Witnesses: [builtin operation contract](../../lyger/builtin_operations_contract_test.go), [operation validation](../../lyger/builtin_operations_validation_test.go), [action definitions](../../lyger/action_definition_test.go), and [Service boundary coverage](../../testdata/service-execution/coverage.md).

## Builtin construction usage and context

<a id="context-builtin-construction--construction-and-definition-builtins"></a>

The implicit [core source](../../builtins/eng/core.l) defines [Process](#word-builtin-process), [Function](#word-builtin-function), [Main](#word-builtin-main) and the other construction facilities. The [definition guide](../guide/definitions.md) shows how to compose source words.

### Function.Invoke and Delegate

<a id="context-builtin-construction--functioninvoke-and-delegate"></a>

[Function.Invoke](#word-builtin-function-invoke) invokes a selected target with arguments. [Delegate](#word-builtin-delegate) supplies the native entry call without arguments. Both use the selected definitions; neither introduces a Call primitive or a second dispatch runtime.

### Storage source compositions

<a id="context-builtin-construction--storage-source-compositions"></a>

Prelude includes core and imports Storage privately. The [Storage usage sections](#context-storage--storage) explain reusable result selection, traversal, return, ownership and byte operations. Programs that name Storage members import that dictionary explicitly.

### Failures and visibility

<a id="context-builtin-construction--failures-and-visibility"></a>

Descriptor target mismatch, a missing constructor rule, recursive descriptor expansion, malformed slots, unsupported signatures, and inaccessible captures reject at their respective semantic boundaries. Aliases preserve canonical descriptor identity. Public/Publicly controls publication independently of New/Generic construction.

Static references retain definition-site bindings. Runtime captures retain their declaration activation: ordinary values snapshot and explicit Pointer targets share. Caller shadowing does not select another environment.

### Sources and witnesses

<a id="context-builtin-construction--sources-and-witnesses"></a>

Sources: [core source](../../builtins/eng/core.l), [descriptors](../../lyger/semantics.go), [construction](../../lyger/construction.go), and [native definition interface](../../lyger/definition_bindings.go). Witnesses: [word contracts](../../lyger/word_library_acceptance_test.go), [ownership](../../lyger/word_library_ownership_test.go), [construction](../../lyger/construction_test.go), [definitions](../../lyger/definition_test.go), [descriptor identity](../../lyger/paragraph_descriptor_identity_test.go), [capture](../../lyger/paragraph_capture_architecture_test.go), and [Service boundary tests](../../lyger/runnable_service_test.go).

## Builtin lifecycle usage and context

<a id="context-builtin-lifecycle--lifecycle-and-observation-builtins"></a>

Build, Run, Stop, and Await act on checked process targets. Show and Exit act in the outer generated program. The full maintained contract is [Building and running L programs](../cli-build.md#app-execution-semantics).

### Start and Execute

<a id="context-builtin-lifecycle--start-and-execute"></a>

[Start](#word-builtin-start) composes Build and Run. [Execute](#word-builtin-execute) composes Start and Await. Both share the App-only lifecycle contract.

### Output and Error

<a id="context-builtin-lifecycle--output-and-error"></a>

[App.Output](#word-builtin-output) reads finalized stdout, and [App.Error](#word-builtin-error) selects a failure. The [IO.Output handle](#word-stdlib-export-eng-io-io-output) instead refers to the current process’s stdout writer.

### Sources and witnesses

<a id="context-builtin-lifecycle--sources-and-witnesses"></a>

Sources: [typed operations](../../lyger/builtin_operations.go), [instructions](../../lyger/instructions.go), [runtime plan](../../lyger/orchestration_plan.go), [core source composition](../../builtins/eng/core.l), and [runtime contract](../../testdata/phase1/runtime/contract.md). Witnesses: [phase-one language](../../lyger/phase1_language_test.go), [nested Await](../../lyger/phase1_nested_await_test.go), [paragraph examples](../../demo/paragraphs/README.md), and [Stop without joining](../../cmd/l/runnable_service_stop_test.go).

## IO usage and context

<a id="context-io--io"></a>

<a id="context-io--availability-and-limits"></a>

The English `io` dictionary provides value printing and explicit byte output. Import it with Use. It has no filesystem, terminal-input, or networking API. IO is an explicit library, not implicit builtin vocabulary.

### Import and calling forms

<a id="context-io--import-and-calling-forms"></a>

The supported member import is:

```l
Use IO's Print from `io` as "Print".
```

Use Print in an App body, as in `The App Prints message.` The [first program](../guide/first-program.md) shows a complete example. ``Use IO from `io`.`` instead keeps the dictionary namespace for qualified member access. The same member-import form works for the other public words, including Lined.

| Task | Word |
| --- | --- |
| Print a supported value to stdout with LF | [Print](#word-stdlib-export-eng-io-io-print) |
| Obtain stdout or stderr without writing | [Output](#word-stdlib-export-eng-io-io-output), [Diagnostic](#word-stdlib-export-eng-io-io-diagnostic) |
| Send encoded bytes through a supplied writer | [Send](#word-stdlib-export-eng-io-io-send) |
| Add LF after successful delivery | [Lined](#word-stdlib-export-eng-io-io-lined) |
| Request a writer flush | [Flush](#word-stdlib-export-eng-io-io-flush) |

### Public and private names

<a id="context-io--public-and-private-names"></a>

IO's six word entries are [Diagnostic](#word-stdlib-export-eng-io-io-diagnostic), [Flush](#word-stdlib-export-eng-io-io-flush), [Lined](#word-stdlib-export-eng-io-io-lined), [Output](#word-stdlib-export-eng-io-io-output), [Print](#word-stdlib-export-eng-io-io-print) and [Send](#word-stdlib-export-eng-io-io-send). Sender and Flusher are actor forms of their verbs. Render, Write, locking and line-completion helpers remain private to the IO source unit. Importing IO does not expose them. A public library entry does not certify allocation, blocking, durability or rollback guarantees.

### Words

<a id="context-io--words"></a>

Browse the [IO word index](../stdlib/eng/io.md) for the alphabetical list and links to each complete entry.

### Native callers

<a id="context-io--native-callers"></a>

These are complete Function declarations within a dictionary, not a standalone App. A native host supplies the standard Rust writer and payload types and invokes the generated Functions. The source uses IO's selected words rather than reimplementing delivery in a native helper. `Vec<u8>` makes the destination visible, so a host can inspect its exact bytes after Send or Line.

<!-- example:word-io -->

The [documentation caller witness](../../lyger/word_documentation_test.go) compiles these canonical bytes and exercises the generated calls. It checks handle types, embedded NUL/invalid UTF-8/LF bytes, line completion, and the unchanged buffer after vector flush. The separate [IO behavior witness](../../lyger/word_library_acceptance_test.go) covers failing and short-write writers. A vector alone would not establish those failure cases.

### Explicit byte output

<a id="context-io--explicit-byte-output"></a>

The native `write_all` operation continues after short writes and retries Interrupted. A zero-length write before completion fails with WriteZero. An empty payload requires no write. Other errors propagate, potentially after partial output. IO adds no further retry policy.

Lined uses L-authored Result selection: only successful payload delivery invokes the LF send through the same writer. It neither reacquires a handle nor locks it. A failed send can leave a prefix, and newline failure can follow a complete payload. Neither Send nor Lined promises rollback or safe retry. The writer determines allocation and blocking behavior.

### Value handling

<a id="context-io--value-handling"></a>

Print accepts supported Text, Number, Truth and Null operands. Text bytes remain unchanged, including whitespace, NUL and invalid UTF-8. Truth renders as `true` or `false`. Null renders as `<nil>`, distinct from the `null` display used by other value contexts. Numbers use the formatting below.

Lists, Dictionaries, Pointers and internal compiler states have no general Print contract.

### Output and newlines

<a id="context-io--output-and-newlines"></a>

Print does not trim its operand before appending LF. A value already ending with LF therefore produces another LF. Empty Text produces one LF. [Show](#word-builtin-show) preserves captured Output without adding a newline. Captured stderr remains separate.

### Number formatting

<a id="context-io--number-formatting"></a>

Number uses binary64. Print selects [Formatting.Format](#context-formatting--format) for the shortest round-tripping representation, including scientific notation and signed zero, then appends LF. For example, `1000000` prints `1e+06\n` and negative zero prints `-0\n`. The Formatting reference owns the numeric thresholds and exact byte examples.

Large source integers may already have rounded. Print reports that binary64 value. Nonfinite L Numbers are rejected before formatting.

### Errors and return behavior

<a id="context-io--errors-and-return-behavior"></a>

Print has no public return value. Its private Write consumes the rendered byte vector, obtains stdout and one guard, invokes a Lined Sender while that guard remains live, and uses [Storage.Expect](#word-authoring-storage-expect) with `failed printing to stdout`. An output failure panics in the native App context. It is not a compiler diagnostic. The guard is released on success or unwinding. The program can inspect the App's Error after Await.

Send, Lined and Flush instead return the native IO Result for the caller to handle. [Storage's result words](#context-storage--results-and-errors) distinguish selection, propagation and the explicit Expect panic policy.

### Sources and witnesses

<a id="context-io--sources-and-witnesses"></a>

Source: [IO dictionary](../../stdlib/eng/io/dictionary.l). Tests cover [public byte IO and failure paths](../../lyger/word_library_acceptance_test.go), [Write's guard and panic policy](../../lyger/word_library_io_policy_test.go), [raw Text](../../lyger/raw_text_test.go), [Number formatting](../../lyger/io_numeric_test.go), [imports](../../lyger/configured_import_test.go) and [private ownership](../../lyger/dictionary_ownership_test.go).

## Formatting usage and context

<a id="context-formatting--formatting"></a>

The English `format` dictionary supplies the shared Number formatter used by [IO.Print](#word-stdlib-export-eng-io-io-print) and [Runtime](#context-runtime--format). It is an explicitly importable implementation dictionary: its native-signature Function supports word authors and the compiler, rather than adding implicit application vocabulary.

### Import and calling forms

<a id="context-formatting--import-and-calling-forms"></a>

Import the dictionary for qualified access:

```l
Use Formatting from `format`.
```

In a Function with a native `f64` input named Value and a `Vec<u8>` result, `a New Formatting's Formatter of Value` constructs the formatted bytes. The member form ``Use Formatting's Format from `format` as "Format".`` binds the same definition locally. Formatter is Format's actor form, not a second word.

| Task | Word or composition |
| --- | --- |
| Obtain Number bytes without writing | [Formatting.Format](#context-formatting--format) |
| Print a scalar to stdout with LF | [IO.Print](#word-stdlib-export-eng-io-io-print) |
| Deliver already formatted bytes to a supplied writer | [IO.Send](#word-stdlib-export-eng-io-io-send) |
| Display a supported value from the orchestrator | [Show](#word-builtin-show) |

### Public and private names

<a id="context-formatting--public-and-private-names"></a>

Formatting.Format is the scalar-formatting entry point. Its declared family is Format, Formats, Formatted, Formatting and Formatter. The qualified Functions Scale, Balance, Reduce, Expand, Contract, Finish, Render, Append and Prepend are also addressable through an explicit Formatting import. Their native signatures support normalization, midpoint handling and byte construction. The [Formatting index](../stdlib/eng/format.md) links each meaning.

Formatting's unqualified [foundations](../../stdlib/eng/format/foundations.l), including fixed-signature numeric casts and imported Storage aliases, stay local to the dictionary. Formatting does not source-load Runtime or process supervision. Runtime.Format and IO's private formatting binding select the same canonical Format definition.

### Format

<a id="context-formatting--format"></a>

Convert a finite binary64 Number into its shortest round-tripping decimal representation and return the encoded bytes.

**Usage:** The selected native function, `format_number`, accepts `f64` and returns `Vec<u8>`. The returned vector contains neither an added newline nor a terminator. Format leaves output, locking and flushing to its caller.

**Example:** [Number output](#context-formatting--number-output-example) shows IO.Print using this formatter. Direct native-signature callers supply an `f64` and consume or borrow the resulting byte vector. This is not a general conversion for Text, collections or Errors.

**Behavior:** Preserve signed zero. Nonzero magnitudes from `0.0001` inclusive to `1000000` exclusive use ordinary decimal notation. Smaller or larger magnitudes use scientific notation. Scientific exponents include a sign and at least two digits.

| Number | Returned bytes shown as text |
| --- | --- |
| `3.5` | `3.5` |
| `-12.5` | `-12.5` |
| `999999` | `999999` |
| `1000000` | `1e+06` |
| `0.0001` | `0.0001` |
| `0.00001` | `1e-05` |
| `-0` | `-0` |

Large source integers may already have rounded to binary64. Formatting reports that stored value. The L [Number](#word-type-number) rules reject nonfinite values. Format's native `f64` signature does not itself enforce finiteness, so this reference does not extend the finite-number contract to arbitrary native NaN or infinity inputs.

### Number output example

<a id="context-formatting--number-output-example"></a>

This registered App-body fragment imports IO.Print and illustrates decimal and scientific output. It omits the surrounding lifecycle instructions and is not a standalone runnable program. Print appends one LF to each formatted value, while Format alone does not.

<!-- example:values-number-printing -->

For a complete App with execution and captured output, see [IO.Print](#word-stdlib-export-eng-io-io-print).

### Errors and ownership

<a id="context-formatting--errors-and-ownership"></a>

Format returns an owned byte vector, not a Result. Its internal expectations describe invariants of native finite-number spellings. They are not a recoverable parsing API. The composition makes no zero-allocation, allocation-count or nonblocking guarantee. It does not change IO's separate [output-failure policy](#context-io--errors-and-return-behavior).

### Sources and witnesses

<a id="context-formatting--sources-and-witnesses"></a>

Sources: [Formatting dictionary](../../stdlib/eng/format/dictionary.l) and its [private foundations](../../stdlib/eng/format/foundations.l). Shared casts and other foundational operations compose through canonical Storage definitions.

The [dictionary-selection tests](../../lyger/format_dictionary_test.go) check canonical Format identity, Runtime's alias and source-import isolation. [Source Number tests](../../lyger/io_numeric_test.go) exercise Print with exact output. [Native scalar tests](../../logiceme/scalar_runtime_test.go) cover the selected formatter's signed zero, extrema, subnormal values and midpoint cases. [Foundation tests](../../lyger/library_foundations_test.go) check shared ownership and cast composition.

## Runtime usage and context

<a id="context-runtime--runtime"></a>

The English `runtime` dictionary supplies L-authored App supervision, error, capture and shutdown policy. It is a compiler implementation dictionary, distinct from ordinary application vocabulary. Its generated native support works with the host and platform mechanisms. Importing it does not create a new runtime or launch a process.

### Import and calling forms

<a id="context-runtime--import-and-calling-forms"></a>

Word authors can name the dictionary explicitly:

```l
Use Runtime from `runtime`.
```

Its Functions require the native types, async context and host facilities of the generated runtime module. Ordinary applications use the [builtin lifecycle words](builtins/lifecycle.md) without importing Runtime. A source name such as Runtime.App.Run identifies an implementation Function. It is not an alternate spelling for a top-level Run instruction.

| Task | Application interface | Implementation group |
| --- | --- | --- |
| Check a prebuilt App | Build | [App](#context-runtime--app) |
| Launch an App | Run, or Start to compose Build then Run | [App](#context-runtime--app) |
| Request termination without joining | Stop | [App](#context-runtime--app) |
| Join an invocation and activate its continuation | Await | [App](#context-runtime--app), [Invocation](#context-runtime--invocation) |
| Build, launch and join | Execute | [App](#context-runtime--app), [Invocation](#context-runtime--invocation) |
| Read finalized output or failure | App Output and Error | [App](#context-runtime--app), [Error](#context-runtime--error) |
| Display a supported value | Show | [Output and shutdown](#context-runtime--output-and-shutdown), [Format](#context-runtime--format) |
| Settle process ownership on exit | Exit or normal completion | [Output and shutdown](#context-runtime--output-and-shutdown) |

### Public and private names

<a id="context-runtime--public-and-private-names"></a>

The groups below describe native support interfaces. Explicitly importing Runtime makes its qualified dictionary members addressable, including Runtime.Host. That namespace supplies aliases for startup and native traits; each alias retains its target’s native signature and requirements. Compiler selection of native export roots is separate from lexical visibility. Unqualified imported bindings and paragraph locals retain their private scopes.

Runtime privately imports required Storage words and the canonical Formatting implementation. Authored L policy lives under `runtime/eng/`, native entry, host and OS inputs live under `runtime/rust/`, and derived AML and provenance live under `runtime/generated/`. See [support assembly](../cli-build.md#source-generation-and-native-builds) for selection and freshness rules.

### App

<a id="context-runtime--app"></a>

Coordinate a prebuilt App executable and its current invocation.

| Member | Native inputs and result | Behavior |
| --- | --- | --- |
| Set | `&App`, `String`<br />Returns unit | Set source attribution for diagnostics |
| Build | `&App`<br />Returns `Result<()>` | Validate the prebuilt companion without invoking a compiler |
| Run | `&Arc<App>`<br />Returns `Result<Arc<Invocation>>` | Launch or record a terminal startup failure |
| Stop | `&App`<br />Returns a future of `Result<()>` | Request termination without joining |
| Wait | `&App`<br />Returns a future of `Result<Arc<Invocation>>` | Join the finalized invocation |
| Read | `&App`<br />Returns `Result<Arc<[u8]>>` | Read finalized stdout bytes |
| Inspect | `&App`<br />Returns `Result<Option<Error>>` | Read the selected failure |

These native types belong to the generated module. A launched process, a finalized invocation and a completed continuation are distinct states. Source Await composes joining with continuation ownership. Calling an arbitrary native member does not reproduce that full source operation. [Invocation and result rules](../cli-build.md#invocation-and-results) define misuse, reruns, cached results and failure selection.

### Capture

<a id="context-runtime--capture"></a>

Collect a child's output stream and preserve its bytes and capture failure.

Capture.Read accepts an optional async reader, a deadline receiver, host dependencies and a stream identity. It returns a future producing Capture. Capture.Join converts a capture-task result and stream label into `(Arc<[u8]>, Option<Error>)`. Snapshot holds stdout and stderr separately, with an optional Error and exit status.

There is no capture-size cap. The post-exit/cancellation drain bound is two seconds. It is not a normal process-duration timeout. Finalized output includes whitespace, NUL and non-UTF-8 bytes. See [stream and shutdown limits](../cli-build.md#shutdown-streams-and-limits).

### Error

<a id="context-runtime--error"></a>

Represent a shared runtime failure with message bytes and retained causes. Runtime.Result names the generated `std::result::Result<T, Error>` carrier.

| Members | Policy |
| --- | --- |
| Describe, Form | Construct from borrowed text or owned message bytes |
| Import, Explain | Retain an IO error or attach context to a native cause |
| Join | Return None for no causes, preserve a singleton, or construct a joined parent |
| Read, Render, Display | Borrow root message bytes, render diagnostic bytes, or write native formatter output |
| Copy, Identify | Retain the shared error and observe its identity |
| Inspect, Trace | Borrow child errors or the native cause |
| Cancel, Check | Construct cancellation or inspect its cancellation flag |
| Attach | Attach captured stderr |

Display's native text conversion may replace invalid UTF-8. Raw message and captured-output interfaces remain byte-oriented. An Error value is distinct from the App Error member that selects a current operation or invocation failure. Error construction does not itself exit, join or roll back effects.

### Format

<a id="context-runtime--format"></a>

Runtime.Format reexports [Formatting.Format](#context-formatting--format), retaining that definition's identity and signature, which accepts `f64` and returns `Vec<u8>`. It adds no lifecycle policy or second number-formatting implementation.

### Invocation

<a id="context-runtime--invocation"></a>

Own one App launch, its terminal result and its continuation claim.

Claim returns `Result<bool>` and admits the continuation once. Finish releases its ownership, Read returns an optional Snapshot, and Complete waits for completion. Allocate, Terminate, Kill, Record, Publish and Supervise support launch, termination and finalization. They require the real generated state, channels, process and host context, not values invented by an ordinary App program.

Snapshot.Copy preserves shared result bytes and error data. Stop can race with natural completion, but finalization publishes one coherent result and each launched direct child has one Wait owner. The [canonical lifecycle contract](../cli-build.md#app-execution-semantics) owns the detailed ordering rules.

### Output and shutdown

<a id="context-runtime--output-and-shutdown"></a>

Show accepts a Runtime and borrowed bytes, returning a future of `Result<()>`. It checks terminal admission, writes those exact bytes and flushes without adding LF. Fail, Leave and Request select failure, normal exit or an explicit terminal request. Begin starts shutdown, and Clean performs idempotent cleanup. Finish resolves a source Result to an exit status, while Run drives the source future through cancellation and finalization. Terminal.Copy preserves the selected terminal state.

Cleanup is explicit, and later requests cannot replace the first accepted terminal decision. Fatal output errors, signals, pending children and cleanup failures follow the [shutdown contract](../cli-build.md#shutdown-streams-and-limits). This is not a promise to clean up after an uncatchable orchestrator kill or host failure, or to own descendants that escaped their process group.

### Startup

<a id="context-runtime--startup"></a>

Open takes no arguments and returns `Result<Runtime>` using native dependencies and stdout/stderr. Install instead accepts `Arc<dyn host::Dependencies>` and two host Output values. Check tests runtime admission, Copy retains a runtime handle, and Appoint associates a companion path with a new `Arc<App>`.

The native entrypoint owns executor startup and invokes selected L policy. Runtime is not a second L interpreter or a Function-dispatch runtime. Current native execution targets are macOS and Linux. The other-platform adapter retains explicit unsupported behavior.

### Lifecycle example

<a id="context-runtime--lifecycle-example"></a>

This complete application uses the runtime through builtin words. Run launches the child. The first Await finalizes its output and activates its continuation, and the second Await reuses the result without repeating that continuation. No Runtime import is required.

<!-- example:runtime-continuation-once -->

For termination, see the [Stop/Await example](../guide/lifecycle.md). The host witnesses distinguish source compilation, actual process execution, stream capture and cleanup. Generating the page does not launch the example.

### Sources and witnesses

<a id="context-runtime--sources-and-witnesses"></a>

The [Runtime dictionary](../../runtime/eng/dictionary.l) assembles [errors](../../runtime/eng/errors.l), [process state](../../runtime/eng/process.l), [App operations](../../runtime/eng/app.l), [startup](../../runtime/eng/start.l), [shutdown](../../runtime/eng/shutdown.l), [supervision](../../runtime/eng/supervision.l) and [capture](../../runtime/eng/capture.l).

[Selection tests](../../lyger/runtime_library_test.go) and [host support selection tests](../../lyger/runtime_selection_resources_test.go) check compiler-owned dependencies. [Composition tests](../../logiceme/runtime_composition_test.go) reject missing/stale resources, and [native layout tests](../../logiceme/native_runtime_layout_test.go) check selected input paths. [Runtime parity](../../logiceme/runtime_parity_test.go) and [error-cause tests](../../logiceme/runtime_error_causes_test.go) observe native behavior. The maintained [integration suite](../../testdata/integration/README.md) covers lifecycle, capture, faults and shutdown on the executing host. It supplies no automatic cross-platform proof.

## Storage usage and context

<a id="context-storage--storage"></a>

<a id="context-storage--storage-source-compositions"></a>

The [Storage dictionary](../../builtins/eng/storage/dictionary.l) supplies reusable source words for the compiler's value, ownership and control-flow implementations. Prelude imports it privately. Source code that names its members uses ``Use Storage from `storage`.`` Its cohesive source includes share one dictionary unit and preserve declaration order, paragraph scope and canonical targets. These are native-backed composition facilities, with native types checked at their Function boundaries.

The groups below explain the shared policies rather than cataloging every representation helper. Native leaves perform individual calls, operators, borrows or patterns. Source Functions and Processes own their sequencing, branches and loops. Member paths below are relative to Storage unless fully qualified.

### Import and calling forms

<a id="context-storage--import-and-calling-forms"></a>

```l
Use Storage from `storage`.
```

Qualify the selected member in a Function with the required native signature, such as `a New Storage's Expecter of Value and Message`. Importing Storage reuses the canonical dictionary already loaded privately by prelude. It does not publish its helpers globally. The [native callers](#context-storage--native-callers) below show complete enclosing declarations and actual calls.

| Task | Composition |
| --- | --- |
| Choose an optional or fallible outcome | [Option.Select](#word-authoring-storage-option-select), [Result.Select](#word-authoring-storage-result-select) |
| Turn absence into a runtime failure | [Require](#word-authoring-storage-require) |
| Choose an explicit panic policy | [Expect](#word-authoring-storage-expect) |
| Visit values or pairs with explicit callback policy | [Iteration and continuations](#context-storage--iteration-and-continuations) |
| Construct a result or return from the caller | [Returns and control flow](#context-storage--returns-and-control-flow) |
| Retain shared identity or append exact bytes | [Ownership and bytes](#context-storage--ownership-and-bytes) |

### Results and errors

<a id="context-storage--results-and-errors"></a>

These are **advanced word-authoring meanings**. Explicitly importing Storage names the dictionary. Private bootstrap loading does not expose its members to ordinary application source. The [Storage word index](../internals/storage.md) distinguishes the native authoring facilities by their qualified names. Their entries state the required context; indexing a native helper does not make it ordinary application vocabulary.

#### Native callers

<a id="context-storage--native-callers"></a>

These dictionary Functions show the enclosing native signatures, explicit imports, typed Process Inputs and actual calls. A generated native library and host provide their execution context. They are not standalone App programs. The Require caller additionally needs the compiler's actual Runtime module.

<!-- example:word-selection -->

The [documentation caller witness](../../lyger/word_documentation_test.go) reads these exact bytes, composes their selected definitions, compiles them with real runtime support and executes both selector outcomes. It also observes Require's success identity/failure message and Expect's success identity/panic. The [selector and lazy-error witnesses](../../lyger/word_library_acceptance_test.go) additionally observe nested scopes, once-only values, inactive bodies, destruction order and conditional error construction.

#### Related error policies

<a id="context-storage--related-error-policies"></a>

Accept consumes `Result<T, E>` and a message, preserves Ok's payload, or releases the old error before constructing a replacement runtime error. If that destructor panics, no replacement is constructed. Its message argument still evaluates on success. Verify instead discards a successful payload or propagates an error from the enclosing Function. These policies are not interchangeable with selection or Expect.

### Iteration and continuations

<a id="context-storage--iteration-and-continuations"></a>

Iteration is a noun, and Paired is an adjective describing Iteration or the actor form Applier. The other words below are verbs with their declared [inflection families](../spec/definitions.md#inflection-declarations). An actor such as Examiner selects the same word's constructed form.

| Word | Ordered operands | Policy |
| --- | --- | --- |
| Apply | Action, Item, Context | Call the action with Item and Context |
| Paired Apply | Action, First, Second, Context | Call with both pair members and Context |
| Traverse / Enumerate | Values, Action, Context | Visit items or pairs in iterator order and discard each callback result, including errors |
| Examine / Review | Action, Item, Context or Action, First, Second, Context | Perform one callback and Verify its result |
| Inspect / Audit | Values, Action, Context | Repeat Examine or Review, discarding successful payloads and propagating the first error |
| Decide | Result, Continuation | Propagate an error, do nothing for false, or execute the supplied Process for true |
| Probe | Action, Item, Context, Continuation | Call the action and pass its result to Decide |
| Search | Values, Action, Context, Continuation | Repeat Probe in iterator order |

Enumerate consumes pair-shaped items. It does not generate numeric indexes. Empty iterators invoke no callback. Callback panics are not caught. Successful callback payloads discarded by Inspect/Audit are released before the next iteration.

Search does not inherently stop at the first true result. A continuation that falls through runs for every true result. A returning continuation can end the enclosing Function. For example, the [Value.Search composition](../../builtins/eng/storage/traversal.l) supplies a Returner to implement first-match behavior. Iteration bodies use scoped Process Inputs, so nested bodies can reuse labels such as Item and Outcome without sharing their bindings.

### Returns and control flow

<a id="context-storage--returns-and-control-flow"></a>

Failure constructs a runtime error from a message, Rejection wraps a value in Err, and Success wraps it in Ok. These nouns construct values without returning from their caller. Reject is a source action that composes Failure, Rejection and the caller's selected [Function.Return](#word-builtin-function-return). Its imperative and actor forms end that caller, subject to the same result-type and constructor checks as Return.

Index.Check accepts finite, nonnegative integral `f64` values strictly below `usize::MAX as f64`, including negative zero, then returns the converted `usize` in Success. Other inputs return `index must be a nonnegative integer`. Collection bounds are checked separately by their readers or writers.

`Storage.Function.Yield` wraps values in `ControlFlow::Continue` and Success before invoking the supplied selected Return. `Storage.Function.Exit` wraps the supplied control-flow value in Success and invokes that Return. These compositions use explicit result bindings. Neither is the global [Exit operation](#word-builtin-exit), which terminates the program.

### Ownership and bytes

<a id="context-storage--ownership-and-bytes"></a>

| Word | Ownership contract |
| --- | --- |
| Core Retain / Storage.App.Retain | Borrow an Arc and produce another handle to the same allocation. App.Retain delegates to core without lifecycle effects |
| Storage.Retain | Borrow an Rc and produce another handle to its allocation |
| Pointer.Copy / Value.Copy | Copy the signature vector while retaining the shared Rc target or data |
| Environment.Copy | Retain the Rc-backed type and value storage, sharing subsequent value mutations |
| Release | Consume a value through native drop. Releasing one shared handle need not destroy its payload |
| Bytes.Extend | Append a borrowed byte slice to a mutably borrowed byte vector and return unit |
| Text.Enclose | Consume a byte vector and return opening + its unchanged bytes + closing, using borrowed string delimiters |

Enclose builds one output buffer from the opening delimiter, then extends it with the payload and closing bytes. It preserves NUL and invalid UTF-8 payload bytes. It may copy and reallocate. It promises neither input-allocation reuse nor a single allocation. Retention does not clone the shared payload or confer Send/Sync on it, and none of these words supplies a general blocking or destructor-cost bound.

### Sources and witnesses

<a id="context-storage--sources-and-witnesses"></a>

The [source dictionary](../../builtins/eng/storage/dictionary.l) includes bindings, value representations, control, scalar display, collections, program storage, traversal and execution compositions in source order. These files add no namespaces or export wrappers. Shared [Runtime](../stdlib/eng/runtime.md) and [Formatting](../stdlib/eng/format.md) bindings and operators use Storage's canonical definitions. The existing Function.Target binding resolves to Tuple.Second's canonical by-value projection. Runtime casts use Storage.Cast. Formatting's private fixed-signature cast descriptors compose through that same operation while retaining their nominal targets. There is no separate Numeric dictionary or new implicit arithmetic vocabulary.

Existing tests cover [namespaces, lazy failures, selectors, bytes and retention](../../lyger/word_library_acceptance_test.go), [traversal and continuation policies](../../lyger/word_library_traversal_test.go), [caller returns](../../lyger/word_library_reject_test.go), and [selected App retention](../../lyger/word_library_ownership_test.go). These implementation contracts complement the language's [copying and aliasing rules](../spec/values.md#conversion-and-aliasing-boundaries).
