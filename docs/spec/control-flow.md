# Operations and control flow

L sequences resolved source instructions and source-defined actions. The supported instruction set is intentionally smaller than the complete grammar/vocabulary inventory.

## Sequencing and evaluation order

Statements retain their original order after source expansion. Coordinated unary Build, Run, Await, Stop, and Show operations process operands in source order. A source-defined one-argument action applied to several operands instead expands its entire body for each operand before starting the next.

Run acknowledges launch before the next instruction, while the child process continues independently. This permits overlapping OS processes; it does not establish parallel evaluation of arbitrary expressions. Native fragments may have their own Rust evaluation rules. Logeme assembly order is not a general source expression-order promise.

**Supported coordination.** The [two-App program](../../demo/phase1/main.l) declares First and Second before these instructions:

<!-- example:control-ordered-apps -->

`Start First and Second` expands Build/Run for First, then Build/Run for Second. `Await Second and First` joins in the opposite order. The two Shows still display First before Second: the retained CLI test expects `First\nSecond\n`. Replacing Start with Execute would finish First's Start/Await composition before starting Second.

### Then

<!-- word:grammar-word:then -->

Separate successive actions within a composed source sequence.

**Usage and boundary.** Launch in the [local definitions program](../../demo/paragraphs/local-definitions.l) uses `and then Awaiting it` after Build/Run. Then preserves source sequencing; it is not an asynchronous continuation primitive.

### And

<!-- word:grammar-word:and -->

Coordinate source operands, declarations or actions in their owning grammatical sequence.

**Usage and boundary.** `Start First and Second` in the [two-App program](../../demo/phase1/main.l) expands the complete action for each operand.

**Partial support.** Operand, declaration and action coordination work, but an `and`-joined condition such as ``If true and false, Show `value`.`` is rejected during instruction lowering. The grammar recognizes the conditions; there is no Boolean join operator or short-circuit implementation for them.

### Or

<!-- word:grammar-word:or -->

Join alternative nominal targets in supported descriptor declarations, or occur in recognized condition coordination.

**Usage and boundary.** Main declares `Service, Object, or Function` in the [descriptor example](lexical-types.md#describe-with-an-adjective). These are alternative canonical targets.

**Partial support.** Descriptor target alternatives work, but an `or`-joined runtime condition is rejected during instruction lowering. The grammar recognizes the joined conditions; there is no Boolean join operator or short-circuit implementation for them.

## Assignment

Passive declarations such as `"count" is 7.` and supported Define forms introduce or assign typed values. Runtime assignment evaluates its right-hand side at the ordered instruction position. A stored Output value remains a snapshot across subsequent invocations.

Definite assignment intersects all paths that continue to a read. Exit branches cannot reach later reads and therefore do not require a value there. The compiler does not insert hidden defaults, and an unused Output assignment still performs the member read and its runtime availability check.

**Supported snapshot assignment — excerpt.** The complete [snapshot fixture](../../testdata/phase1/language/lang-11.l) first declares a Basic App named Worker that prints `saved`:

<!-- example:control-output-snapshot -->

The first Await makes Output available. Define reads it once; later invocations do not change `snapshot`. The last Show reads the latest Output separately. The [frontend witness](../../lyger/phase1_language_test.go) checks that the saved value retains its binding.

### Is

<!-- word:grammar-word:is -->

Associate a subject with its declared or assigned value; in a condition, compare supported like-typed values.

**Usage and boundary.** `"message" is` followed by Text in the [first program](../guide/first-program.md#read-the-program) supplies assignment. The comparison table below distinguishes conditional equality and the configured negative form `isnt`; neither implies arbitrary object equality.

## Comparisons

The current condition operators and their boundaries are:

| Condition | Supported operands | Meaning |
| --- | --- | --- |
| Bare value | Truth | Test the boolean value |
| `left is right` | Matching Text, Number, or Truth | Equality |
| `left equals right` | Matching Text, Number, or Truth | Equality |
| `left isnt right` | Matching Text, Number, or Truth | Inequality |
| `left less than right` | Number and Number | Strict order |
| `left greater than right` | Number and Number | Strict order |
| `left less than equal to right` | Number and Number | Inclusive order grammar |
| `left greater than equal to right` | Number and Number | Inclusive order grammar |
| `target has an Error` | Checked App and its Error | Test the selected process operation/invocation error |

The inclusive-order spelling follows L's grammar. Source phrases must satisfy both grammar and operand typing.

`not` or `no` negates only the condition forms whose grammar owns that modifier. Values do not have general truthiness, Text ordering is rejected, and mixed scalar comparisons do not coerce. The Error in a has condition must belong to the same target; it is not generic collection membership.

### Have

<!-- word:grammar-word:have -->

Test whether the selected App has the Error identified by its current operation or invocation.

**Usage and boundary.** `it has an Error` in the [conditional member fixture](../../lyger/test/member-resolution.l) follows Execute and refers to HelloWorldApp. Have is not general collection membership; the Error must belong to the same App.

### Equal

<!-- word:grammar-word:equal -->

Compare two supported scalar values of the same type for equality.

**Usage and boundary.** The [Truth condition](../../testdata/phase1/language/lang-13.l) uses `enabled equals true`. Text compares bytes and Number compares numeric values; mixed types do not coerce.

### Not

<!-- word:grammar-word:not -->

Negate a condition only where its grammatical form admits this modifier.

**Usage and boundary.** The grammar admits `not` in its bounded condition forms; see the [condition rules](../../lexicon/eng/grammar.json) and negative-condition [lowering witnesses](../../lyger/orchestrator_test.go). This is not an unrestricted prefix Boolean operator or a promise that joined conditions execute.

**Example.** Here `enabled` is a declared Truth value. Negating its false value selects the first branch and writes `disabled` without a newline.

<!-- example:word-not -->

### Less

<!-- word:grammar-word:less -->

Select a numeric less-than comparison, with an optional equality continuation.

**Usage and boundary.** The focused form `left less than right` requires two Number operands; `left less than equal to right` is the inclusive grammar. These [typed comparison witnesses](../../lyger/orchestrator_test.go) establish the native boundary; Text ordering is rejected.

### Greater

<!-- word:grammar-word:greater -->

Select a numeric greater-than comparison, with an optional equality continuation.

**Usage and boundary.** The focused form `left greater than right` requires Numbers; `left greater than equal to right` is the inclusive form. The [comparison witnesses](../../lyger/orchestrator_test.go) retain operand typing; no lexical Text ordering is implied.

## Selection

If selects a branch from a supported condition. An attached else belongs to the unmatched if. Otherwise can attach to an immediately preceding unmatched conditional in the same source paragraph; it does not search arbitrary earlier paragraphs or create an unconditional fallback after unrelated statements.

**Supported equality and selection.** This program makes the wait explicit on the branch that continues:

<!-- example:values-truth-condition -->

`enabled` is Truth. When it equals true, Await joins Worker; otherwise Exit terminates the program. Every path reaching the Output assignment has joined. Expected stdout is `condition\n`.

**Conditional ownership — excerpt from a frontend witness.** The [member-resolution source](../../lyger/test/member-resolution.l) declares an App, declares Text `x`, and executes the App before these clauses:

<!-- example:control-otherwise -->

The first `else` belongs to the Error test and assigns `y` from the completed invocation. Otherwise belongs to the immediately preceding `If x isnt y`. This [binding fixture](../../lyger/member_resolution_regression_test.go) requires the earlier declarations; its Text comparison differs by Print's final newline.

The grammar can parse multiple conditions, but outer lowering explicitly rejects joined condition lists. Consequently `and`/`or` in the vocabulary does not establish general boolean short-circuiting. Coordination of actions, declarations, and operands is a different supported use of conjunctions.

### If

<!-- word:grammar-word:if -->

Select a source branch from a supported typed condition.

**Usage and boundary.** The [Truth conditional program](../../testdata/phase1/language/lang-13.l) selects Await on true and Exit otherwise. Definite assignment is checked across all continuing paths.

**Partial support.** A single typed condition works. The configured list of conditions joined by And or Or is rejected during instruction lowering, and the configured `else if` continuation is rejected during grammar processing. Neither form is an executable extension of the supported If/Else pair.

### Else

<!-- word:grammar-word:else -->

Attach an alternative branch to its unmatched If in the supported conditional statement.

**Usage and boundary.** The [Truth conditional program](../../testdata/phase1/language/lang-13.l) puts `else exit` after its Await branch. Else is not an independent executable action or a fallback for arbitrary previous failures.

**Partial support.** A plain alternative branch works, but the configured `else if` continuation does not: a source chain such as `If false, Show 1, else if true, Show 2, else Show 3` is rejected during grammar processing. It does not produce an executable conditional chain.

### Otherwise

<!-- word:grammar-word:otherwise -->

Attach an alternative branch to the immediately preceding unmatched conditional in the same paragraph.

**Usage and boundary.** The [conditional member fixture](../../lyger/test/member-resolution.l) follows an If with Otherwise. Intervening statements or a paragraph boundary do not grant attachment to an earlier condition.

## Iteration and collection operations

While and three-part source loops remain unsupported. [Collections](collections-and-categories.md) support Put, Place and typed reads; configured Get/Take words do not introduce further operations.

**Unsupported loop — source fragment:**

<!-- example:control-while-unsupported -->

The [grammar configuration](../../lexicon/eng/grammar.json) declares a While rule, but this source fragment is rejected during grammar processing. The [instruction builder](../../lyger/instructions.go) also has no While operation; the configured rule is not an executable loop.

Within native Function bodies, the Storage implementation dictionary composes Iteration and typed Processes into Traverse, Enumerate and related [source words](../internals/storage.md#iteration-and-continuations). Their contracts specify ordering and error behavior. They do not add general source loop syntax or break/continue vocabulary. The [Rust target](../../builtins/README.md#iteration-and-definition-contract) supports their bounded iteration shapes and rejects the three-part X form.

### While

<!-- word:grammar-word:while -->

Name a configured conditional loop form with no executable source implementation.

**Unsupported operation.** The [unsupported loop illustration](../examples/control-while-unsupported.l) is rejected during grammar processing, and no outer While instruction is implemented. It is a labeled fragment, not a runnable loop or a promise of native iteration. Advanced [Storage traversal](../internals/storage.md#iteration-and-continuations) is a separate source composition facility.

## Arithmetic and native calculations

Number is binary64, but the configured English grammar has no general infix arithmetic expression system. No public precedence table for `+`, `-`, multiplication, division, or arithmetic grouping can be inferred from numeric literals or D/E templates.

D/E calculates native expressions; G/I selects branches. A source-defined native-backed word may use supported target fragments deliberately. Such a definition documents its own typed contract and target dependence; it is not a built-in arithmetic operator family.

## Show and output

Show is an outer observation instruction. It writes Text bytes unchanged and formats supported runtime values without adding a newline. Direct App and Error display rejects. It does not implicitly join a process, replay unrelated captured output, or turn child stderr into stdout. Use explicit Await before Output when completion is required.

IO.Print is an imported action that adds one newline. See [IO](../stdlib/eng/io.md#output-and-newlines) for raw-byte and numeric behavior.

In the [first program](../guide/first-program.md#read-the-program), Print adds LF inside Hello's process. Execute waits, then Show writes captured Output exactly, without adding a second LF.

## Exit and Return

Bare Exit terminates the whole L program after cleanup and succeeds if cleanup succeeds. Exit with an explicit Error/text failure follows the maintained error path and returns failure. Exit inside a continuation or expanded source action still exits the orchestrator; it is not merely a return from the nearest paragraph.

**Supported whole-program termination.**

<!-- example:control-bare-exit -->

Execute reaches First's continuation, which shows `First\n` and then exits. The following Show in the paragraph, Second's Execute operand, and the program's final Show are unreachable. The retained paragraph test expects that one output line and status 0 when cleanup succeeds. Compare the [explicit-error example](errors.md#runtime-failure-is-a-separate-stage), whose exit status is 1.

Function.Return completes the enclosing invocation through its selected definition. Root-authored returning actions use their caller's Return; definition-owned actions preserve their constructor contract. A conditional Return leaves other paths continuing, and constructing another Function does not return from its caller. See the [Return contract](../reference/builtins/definitions.md#functionreturn). This supplies neither general source exceptions nor return-value transport between App processes.

## Sources and witnesses

Primary sources: [instruction construction](../../lyger/instructions.go), [condition typing](../../lyger/orchestration_plan.go), [flow analysis](../../lyger/orchestration_flow.go), [emission](../../lyger/execution_composition.go), [condition kinds](../../lyger/orchestration_kinds_test.go), and [grammar](../../lexicon/eng/grammar.json). Existing witnesses: [instruction adjacency](../../lyger/instructions_test.go), [scalar conditions and snapshots](../../lyger/orchestrator_test.go), [paragraph grammar coordination](../../lyger/paragraph_grammar_architecture_test.go), [bare Exit](../../demo/paragraphs/exit-program.l), and [error Exit](../../demo/paragraphs/exit-error.l).
