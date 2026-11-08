# Construction and definition builtins

The implicit [core source](../../../builtins/eng/core.l) defines these composition facilities. See the [builtin index](index.md#source-defined-facilities) for their word forms and the [definition guide](../../guide/definitions.md) for source authoring.

## Process

<!-- word:builtin:process -->

Compose an ordered source body from existing instantiated definition expressions.

**Usage and example.** A `New Process with Inputs ... for ...` supplies a branch body to [Storage.Option.Select](../../internals/storage.md#optionselect). Its typed input belongs to that branch. The [selector caller](../../examples/word-selection.l) demonstrates enclosing Functions and both branches.

Process composes ordered Definitions/Syllables through operation W.

A Process collects existing instantiated definition expressions. It is not an OS process handle. App invocations own those. Processing a source-defined rendering or writing operation in IO is an application-body composition, not a direct runtime launch instruction.

Structured selection and iteration bodies can declare typed Process Inputs. Lyger binds each input within its body and retains its declared native type and source dependency. An explicit local annotation lets native compilation check type and borrow compatibility. `_` requests native type inference. These bindings create no callable or closure and cannot escape to a sibling branch or later sentence.

## Function

<!-- word:builtin:function -->

Define a callable with a selected constructor, ordered typed inputs and results, and a private source body.

**Usage and example.** The [complete Identity/Relay program](../../examples/composed-functions.l) declares two Functions, calls their invocation verbs and names each result. Ordinary source signatures and explicitly native signatures share this selected definition. Native generic/lifetime atoms require the advanced context below.

Function supplies an FW native definition with Func metadata. Name is Text, Inputs is Props, Outputs is Texts, and Process is an already instantiated definition reference.

Inputs and Outputs preserve their typed cardinality. Main supplies No Inputs and No Outputs. Ordinary Function declarations bind their checked signatures and paragraph bodies to the same selected definition. The constructor is not a marker for a second emitter.

**Supported boundary.** A Function declares one ordered signature and owns its body paragraph. A source Function with Outputs must return them on every continuing path. Recursive calls and Function-valued arguments/results use that same checked signature. Generic inference, overload resolution and variadic parameters are not part of this interface.

Invocation words use normal grammatical binding. Captures, snapshots and Pointer sharing follow [scope rules](../../spec/names-and-scope.md#runtime-storage-and-definite-assignment). Native signature/lifetime metadata remains bounded by the [definition contract](../../spec/definitions.md#process-function-and-return).

## Function.Invoke and Delegate

These advanced facilities compose invocation. They do not introduce a Call primitive or a second dispatch runtime.

### Function.Invoke

<!-- word:builtin:function.invoke -->

Apply a selected Function target to its ordered, resolved argument atoms.

**Usage and example.** The `Identity Message as "Nested"` call in the [complete Function example](../../../docs/examples/composed-functions.l) uses the selected Function.Invoke definition. An author normally invokes the Function's own verb. Its Target and Arguments atoms are compiler-bound references, not an executable body hidden in text. [Selection tests](../../../lyger/function_composition_test.go) establish that this implementation controls emitted calls.

### Delegate

<!-- word:builtin:delegate -->

Call an explicitly selected native target without supplying an argument list.

**Usage and example.** The [recursive Class fixture](../../../lyger/test/classes/recursive.l) places Main below nested modules. Its generated entry uses Delegate with a compiler-resolved native path to that Main Function. The [source topology implementation](../../../lyger/source_units_topology.go) supplies this reference as the entry's sole tail call. A quoted Text target is not this native reference, and no general Text-to-call API is promised.

Both facilities preserve the selected target and lexical ownership. Native type and symbol checking remain separate from grammatical recognition.

## Function.Return

<!-- word:builtin:function.return -->

Complete the current Function invocation through its selected Return definition with the declared results.

**Usage and example.** `Return Value` in the [Identity/Relay program](../../examples/composed-functions.l) returns Text, while Relay separately binds and returns Nested. A Return from a selected Process still belongs to its enclosing Function.

Root-authored returning actions select their caller's Return. That selected implementation preserves its own lexical delegates. Function-owned actions retain their constructor contract and reject incompatible callers. Each Function establishes its own return context and checks declared result types and counts. Known unconditional exits through source actors and transparent Processes enter the typed plan. Opaque or conditional bodies do not establish unconditional exits.

Return completes that invocation. Exit terminates the enclosing program. Neither transports runtime captures into an App process.

## Retain

<!-- word:builtin:retain -->

Borrow an Arc handle and produce another handle to the same allocation.

**Usage and example.** The [native caller module](../../examples/word-core.l) takes `&std::sync::Arc<u8>` and returns the retained Arc. It preserves allocation identity. [Storage retention](../../internals/storage.md#ownership-and-bytes) distinguishes Rc from this Arc operation.

Retain takes an explicit borrowed Arc and returns another handle to the same allocation. It increments shared ownership without cloning the payload. Atomic reference counting does not make an arbitrary payload Send or Sync.

The [Storage ownership words](../../internals/storage.md#ownership-and-bytes) distinguish this Arc operation from Rc retention, copying and release.

## Storage source compositions

Prelude includes core and imports Storage privately. Its [implementation reference](../../internals/storage.md) explains reusable result selection, traversal, return, ownership and byte operations. Programs that name Storage members import that dictionary explicitly.

## Objective

<!-- word:builtin:objective -->

Group an existing Function definition as source content for its containing Object.

**Usage and example.** `New Objective First Function` in the [Object construction fixture](../../../lyger/test/classes/interleaved.l) puts the selected Function into its Object. It does not invoke the Function or infer a signature.

Objective describes any Function by wrapping its existing definition in a grouped native source word containing A and the Function. It preserves the input definition and grouping. It does not clone its runtime semantics or infer additional function arguments.

## Main

<!-- word:builtin:main -->

Describe the canonical native entry structure for a Function, Object or Service.

**Usage and example.** The [nested construction example](../../spec/applicational-types.md#compose-an-object-inside-nested-classes) applies Main to an Object, while Basic composes its Service rule for an App. Choose the actual target below rather than treating all three refinements as interchangeable.

Main is a New Adjective describing Service, Object, or Function. Its ordered refinements compose:

### Main Function

The Function rule constructs a Generic `main` Function from the supplied Process, with No Inputs and No Outputs. It chooses an entry signature. It does not accept arbitrary Function arguments.

### Main Object

The Object rule constructs a Generic `main` Object containing a New Objective Main Function. The Process becomes that Function's body, not an Object initializer.

### Main Service

The Service rule constructs a Generic `main` Service containing a New Main Object. Basic uses this structure inside an App. It does not make a standalone Service executable.

These are ordinary source-defined descriptor rules. Aliases preserve their identity, and each construction applies every rule matching its target in authored order. Main currently has one rule for each target above. Repeated construction remains distinct.

Main does not grant standalone Service execution. A standalone declaration is structural. Standalone Service body effects and native process operations reject. Apps own executable identity through their construction.

## Basic

<!-- word:builtin:basic -->

Describe an App containing a Main Service that receives the App's source Process.

**Usage and example.** `Create "Hello" as a New Basic App` in the [first program](../../guide/first-program.md#read-the-program) creates the body owner. Execute is a separate lifecycle action. Construction alone launches nothing.

Basic describes a Generic App with a New Main Service using a Process. It is the composition behind `Create "Hello" as a New Basic App.` The App's source actions supply its body, and the orchestrator separately controls launch and output.

## Failures and visibility

Descriptor target mismatch, a missing constructor rule, recursive descriptor expansion, malformed slots, unsupported signatures, and inaccessible captures reject at their respective semantic boundaries. Aliases preserve canonical descriptor identity. Public/Publicly controls publication independently of New/Generic construction.

Static references retain definition-site bindings. Runtime captures retain their declaration activation: ordinary values snapshot and explicit Pointer targets share. Caller shadowing does not select another environment.

## Sources and witnesses

Sources: [core source](../../../builtins/eng/core.l), [descriptors](../../../lyger/semantics.go), [construction](../../../lyger/construction.go), and [native definition interface](../../../lyger/definition_bindings.go). Witnesses: [word contracts](../../../lyger/word_library_acceptance_test.go), [ownership](../../../lyger/word_library_ownership_test.go), [construction](../../../lyger/construction_test.go), [definitions](../../../lyger/definition_test.go), [descriptor identity](../../../lyger/paragraph_descriptor_identity_test.go), [capture](../../../lyger/paragraph_capture_architecture_test.go), and [Service boundary tests](../../../lyger/runnable_service_test.go).
