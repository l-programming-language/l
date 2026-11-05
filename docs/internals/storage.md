# Storage source compositions

The [Storage dictionary](../../stdlib/eng/storage/dictionary.l) supplies reusable source words for the compiler's value, ownership and control-flow implementations. Core imports it privately; source code that names its members uses ``Use Storage from `storage`.`` These are native-backed composition facilities, with native types checked at their Function boundaries.

The groups below explain the shared policies rather than cataloging every representation helper. Native leaves perform individual calls, operators, borrows or patterns; source Functions and Processes own their sequencing, branches and loops. Member paths below are relative to Storage unless fully qualified.

## Results and errors

Option and Result are Dictionary namespaces. They do not declare L runtime carrier types; native signatures spell types such as `Option<T>` and `Result<T, E>` explicitly.

| Word | Inputs and behavior |
| --- | --- |
| Option.Select | Select one of two Processes for Some or None; the Some body binds its payload through one typed Input, and the None body takes none |
| Result.Select | Select one of two Processes for Ok or Err; each body binds its payload or error through one typed Input |
| Require | Consume `Option<T>` and a message; return `runtime::Result<T>`, preserving Some's payload or constructing a failure for None |
| Accept | Consume `Result<T, E>` and a message; preserve Ok's payload, or release the old error and construct a replacement runtime error |
| Result.Expect | Consume a Result and explicit message; return the successful payload or panic on the error |
| Verify | Discard a successful payload or propagate an error from the enclosing Function |

Select evaluates its selected value once. Only the chosen branch runs, and its Inputs and local declarations stay inside that body. Moving or borrowing the payload follows the supplied native type; selection adds no clone. Branches can compose further selections or iterations.

Require and Accept construct failures only on their failure branches. Their message arguments still evaluate normally, including on success. Accept releases the replaced error before constructing the new one; if its destructor panics, no replacement is constructed. Expect is a separate panic policy, not error propagation or conversion to a runtime error.

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

Search does not inherently stop at the first true result. A continuation that falls through runs for every true result; a returning continuation can end the enclosing Function. For example, the [Value.Search composition](../../stdlib/eng/storage/dictionary.l) supplies a Returner to implement first-match behavior. Iteration bodies use scoped Process Inputs, so nested bodies can reuse labels such as Item and Outcome without sharing their bindings.

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

The [source dictionary](../../stdlib/eng/storage/dictionary.l) owns these compositions. Existing tests cover [namespaces, lazy failures, selectors, bytes and retention](../../lyger/word_library_acceptance_test.go), [traversal and continuation policies](../../lyger/word_library_traversal_test.go), [caller returns](../../lyger/word_library_reject_test.go), and [selected App retention](../../lyger/word_library_ownership_test.go). These implementation contracts complement the language's [copying and aliasing rules](../spec/values.md#conversion-and-aliasing-boundaries).
