# Storage

<a id="storage-source-compositions"></a>

The [Storage dictionary](../../builtins/eng/storage/dictionary.l) supplies reusable source words for the compiler's value, ownership and control-flow implementations. Prelude imports it privately; source code that names its members uses ``Use Storage from `storage`.`` Its cohesive source includes share one dictionary unit and preserve declaration order, paragraph scope and canonical targets. These are native-backed composition facilities, with native types checked at their Function boundaries.

The groups below explain the shared policies rather than cataloging every representation helper. Native leaves perform individual calls, operators, borrows or patterns; source Functions and Processes own their sequencing, branches and loops. Member paths below are relative to Storage unless fully qualified.

## Import and calling forms

```l
Use Storage from `storage`.
```

Qualify the selected member in a Function with the required native signature, such as `a New Storage's Result's Expecter of Value and Message`. Importing Storage reuses the canonical dictionary already loaded privately by prelude; it does not publish its helpers globally. The [native callers](#native-callers) below show complete enclosing declarations and actual calls.

| Task | Composition |
| --- | --- |
| Choose an optional or fallible outcome | [Option.Select](#optionselect), [Result.Select](#resultselect) |
| Turn absence into a runtime failure | [Require](#require) |
| Choose an explicit panic policy | [Result.Expect](#resultexpect) |
| Visit values or pairs with explicit callback policy | [Iteration and continuations](#iteration-and-continuations) |
| Construct a result or return from the caller | [Returns and control flow](#returns-and-control-flow) |
| Retain shared identity or append exact bytes | [Ownership and bytes](#ownership-and-bytes) |

## Results and errors

These are **advanced word-authoring meanings**. Explicitly importing Storage names the dictionary; private bootstrap loading does not expose its members to ordinary application source. The entries below describe selected composition facilities, not a promise that every Storage helper is application API.

### Option

<!-- word:authoring:Storage.Option -->

Name the Storage dictionary namespace containing operations that compose native optional values.

**Usage:** After ``Use Storage from `storage`.`` qualify a member such as `Storage.Option.Select`. Option is a Dictionary namespace, not an L carrier type: the native Function signature uses `Option<T>`, with a real payload type in place of `T`.

**Example:** `Selection.Choose` in [Native callers](#native-callers) receives `Option<u64>` and selects its Some or None case. The qualified member is [Option.Select](#optionselect); Result's two payload-bearing outcomes belong to [Result](#result).

### Option.Select

<!-- word:authoring:Storage.Option.Select -->

Select exactly one Process according to whether a supplied native Option contains a payload or is empty.

**Usage:** Supply the Option value, a Some Process with one explicit typed Input, then a None Process with no payload Input. The payload type must match the enclosing Function's native carrier. Selector is Select's actor form in the Option namespace; it is distinct from [Result.Select](#resultselect).

**Example:** `Selection.Choose` in [Native callers](#native-callers) returns the bound `u64` for Some and zero for None. The selected value is evaluated once. Only the chosen body runs, and its Input and declarations remain local to that Process. Moving or borrowing follows the supplied type; selection does not clone the payload. Nested selections retain their own bindings.

### Result

<!-- word:authoring:Storage.Result -->

Name the Storage dictionary namespace containing operations that compose native success-or-error results.

**Usage:** Qualify [Result.Select](#resultselect) for branch selection or [Result.Expect](#resultexpect) for an explicit panic policy. Result is not an L runtime carrier declaration. Write the actual native type, such as `Result<T, E>` or `runtime::Result<T>`, in the enclosing Function signature.

**Example:** `Selection.Preserve` in [Native callers](#native-callers) handles `Result<u64, u8>`. Its error is a `u8`, while `Selection.Require` returns the runtime's Error type; sharing the namespace does not unify those types or failure policies.

### Result.Expect

<!-- word:authoring:Storage.Result.Expect -->

Consume a native Result and an explicit message, returning its successful payload or panicking on its error.

**Usage:** `a New Storage's Result's Expecter of Value and Message` calls the Result's native `expect`. The error type must support its native debug formatting, and the message is `&str`. Expecter belongs to this Expect verb family, not to [Result.Select](#resultselect).

**Example:** `Selection.Expect` in [Native callers](#native-callers) preserves the successful Box's allocation. An error triggers the supplied message together with the error's debug representation. This is a panic policy, not propagation, rollback, or conversion into `runtime::Error`. [IO.Print](../stdlib/eng/io.md#print) deliberately uses it for output failure.

### Result.Select

<!-- word:authoring:Storage.Result.Select -->

Select exactly one Process according to whether a supplied native Result is successful or contains an error.

**Usage:** Supply the Result value, an Ok Process with one typed payload Input, then an Err Process with one typed error Input. Each Input matches its corresponding native type. Both bodies belong to the enclosing Function context; a Return in the selected body returns from that Function.

**Example:** `Selection.Preserve` in [Native callers](#native-callers) reconstructs Ok and Err using their original values. The Result is evaluated once; only one branch runs, and branch-local bindings do not leak into siblings. Selection adds no panic, error replacement or clone. [IO.Lined](../stdlib/eng/io.md#lined) composes this selector to send LF only after payload success.

### Require

<!-- word:authoring:Storage.Require -->

Consume an Option and return its payload as a runtime success, or construct a runtime failure when it is empty.

**Usage:** `a New Storage's Requirer of Value and Message` consumes `Option<T>` and an `&str` message and returns `runtime::Result<T>`. The enclosing native module must include the compiler-selected Runtime support. Require does not return directly from its caller; that caller chooses how to handle the returned Result.

**Example:** `Selection.Require` in [Native callers](#native-callers) keeps a successful Box's identity and returns the supplied message as a runtime error for None. Failure construction happens only on None. Message arguments still evaluate normally, including on Some; passing an effectful argument is not made lazy by selection.

### Native callers

These dictionary Functions show the enclosing native signatures, explicit imports, typed Process Inputs and actual calls. A generated native library and host provide their execution context; they are not standalone App programs. The Require caller additionally needs the compiler's actual Runtime module.

<!-- example:word-selection -->

The [documentation caller witness](../../lyger/word_documentation_test.go) reads these exact bytes, composes their selected definitions, compiles them with real runtime support and executes both selector outcomes. It also observes Require's success identity/failure message and Expect's success identity/panic. The [selector and lazy-error witnesses](../../lyger/word_library_acceptance_test.go) additionally observe nested scopes, once-only values, inactive bodies, destruction order and conditional error construction.

### Related error policies

Accept consumes `Result<T, E>` and a message, preserves Ok's payload, or releases the old error before constructing a replacement runtime error. If that destructor panics, no replacement is constructed. Its message argument still evaluates on success. Verify instead discards a successful payload or propagates an error from the enclosing Function. These policies are not interchangeable with selection or Expect.

## Iteration and continuations

Iteration is a noun; Paired is an adjective describing Iteration or the actor form Applier. The other words below are verbs with their declared [inflection families](../spec/definitions.md#inflection-declarations). An actor such as Examiner selects the same word's constructed form.

| Word | Ordered operands and policy |
| --- | --- |
| Apply | Action, Item, Context; call the action with Item and Context |
| Paired Apply | Action, First, Second, Context; call with both pair members and Context |
| Traverse / Enumerate | Values, Action, Context; visit items or pairs in iterator order and discard each callback result, including errors |
| Examine / Review | Action, Item, Context or Action, First, Second, Context; perform one callback and Verify its result |
| Inspect / Audit | Values, Action, Context; repeat Examine or Review, discarding successful payloads and propagating the first error |
| Decide | Result, Continuation; propagate an error, do nothing for false, or execute the supplied Process for true |
| Probe | Action, Item, Context, Continuation; call the action and pass its result to Decide |
| Search | Values, Action, Context, Continuation; repeat Probe in iterator order |

Enumerate consumes pair-shaped items; it does not generate numeric indexes. Empty iterators invoke no callback. Callback panics are not caught. Successful callback payloads discarded by Inspect/Audit are released before the next iteration.

Search does not inherently stop at the first true result. A continuation that falls through runs for every true result; a returning continuation can end the enclosing Function. For example, the [Value.Search composition](../../builtins/eng/storage/traversal.l) supplies a Returner to implement first-match behavior. Iteration bodies use scoped Process Inputs, so nested bodies can reuse labels such as Item and Outcome without sharing their bindings.

## Returns and control flow

Failure constructs a runtime error from a message; Rejection wraps a value in Err, and Success wraps it in Ok. These nouns construct values without returning from their caller. Reject is a source action that composes Failure, Rejection and the caller's selected [Function.Return](../reference/builtins/definitions.md#functionreturn). Its imperative and actor forms end that caller, subject to the same result-type and constructor checks as Return.

Index.Check accepts finite, nonnegative integral `f64` values strictly below `usize::MAX as f64`, including negative zero, then returns the converted `usize` in Success. Other inputs return `index must be a nonnegative integer`. Collection bounds are checked separately by their readers or writers.

The internal `Storage.Function.Yield` wraps values in `ControlFlow::Continue` and Success before invoking the supplied selected Return. `Storage.Function.Exit` wraps the supplied control-flow value in Success and invokes that Return. These compositions use explicit result bindings; neither is the global [Exit operation](../reference/builtins/lifecycle.md#exit), which terminates the program.

## Ownership and bytes

| Word | Ownership contract |
| --- | --- |
| Core Retain / Storage.App.Retain | Borrow an Arc and produce another handle to the same allocation; App.Retain delegates to core without lifecycle effects |
| Storage.Retain | Borrow an Rc and produce another handle to its allocation |
| Pointer.Copy / Value.Copy | Copy the signature vector while retaining the shared Rc target or data |
| Environment.Copy | Retain the Rc-backed type and value storage, sharing subsequent value mutations |
| Release | Consume a value through native drop; releasing one shared handle need not destroy its payload |
| Bytes.Extend | Append a borrowed byte slice to a mutably borrowed byte vector; return unit |
| Text.Enclose | Consume a byte vector and return opening + its unchanged bytes + closing, using borrowed string delimiters |

Enclose builds one output buffer from the opening delimiter, then extends it with the payload and closing bytes. It preserves NUL and invalid UTF-8 payload bytes. It may copy and reallocate; it promises neither input-allocation reuse nor a single allocation. Retention does not clone the shared payload or confer Send/Sync on it, and none of these words supplies a general blocking or destructor-cost bound.

## Sources and witnesses

The [source dictionary](../../builtins/eng/storage/dictionary.l) includes bindings, value representations, control, scalar display, collections, program storage, traversal and execution compositions in source order. These files add no namespaces or export wrappers. Shared [Runtime](../stdlib/eng/runtime.md) and [Formatting](../stdlib/eng/format.md) bindings and operators use Storage's canonical definitions. The existing Function.Target binding resolves to Tuple.Second's canonical by-value projection. Runtime casts use Storage.Integral; Formatting's private fixed-signature cast descriptors compose through that same operation while retaining their nominal targets. There is no separate Numeric dictionary or new implicit arithmetic vocabulary.

Existing tests cover [namespaces, lazy failures, selectors, bytes and retention](../../lyger/word_library_acceptance_test.go), [traversal and continuation policies](../../lyger/word_library_traversal_test.go), [caller returns](../../lyger/word_library_reject_test.go), and [selected App retention](../../lyger/word_library_ownership_test.go). These implementation contracts complement the language's [copying and aliasing rules](../spec/values.md#conversion-and-aliasing-boundaries).
