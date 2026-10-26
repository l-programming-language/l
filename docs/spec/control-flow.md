# Operations and control flow

L sequences resolved source instructions and source-defined actions. The supported instruction set is intentionally smaller than the complete grammar/vocabulary inventory.

## Sequencing and evaluation order

Statements retain their original order after source expansion. Coordinated unary Build, Run, Await, Stop, and Show operations process operands in source order. A source-defined one-argument action applied to several operands instead expands its entire body for each operand before starting the next.

Run acknowledges launch before the next instruction, while the child process continues independently. This permits overlapping OS processes; it does not establish parallel evaluation of arbitrary expressions. Native fragments may have their own Rust evaluation rules. Logeme assembly order is not a general source expression-order promise.

**Supported coordination.** The [two-App program](../../demo/phase1/main.l) declares First and Second before these instructions:

<!-- example:control-ordered-apps -->

`Start First and Second` expands Build/Run for First, then Build/Run for Second. `Await Second and First` joins in the opposite order. The two Shows still display First before Second: the retained CLI test expects `First\nSecond\n`. Replacing Start with Execute would finish First's Start/Await composition before starting Second.

## Assignment

Passive declarations such as `"count" is 7.` and supported Define forms introduce or assign typed values. Runtime assignment evaluates its right-hand side at the ordered instruction position. A stored Output value remains a snapshot across subsequent invocations.

Definite assignment intersects all paths that continue to a read. Exit branches cannot reach later reads and therefore do not require a value there. The compiler does not insert hidden defaults, and an unused Output assignment still performs the member read and its runtime availability check.

**Supported snapshot assignment — excerpt.** The complete [snapshot fixture](../../testdata/phase1/language/lang-11.l) first declares a Basic App named Worker that prints `saved`:

<!-- example:control-output-snapshot -->

The first Await makes Output available. Define reads that finalized value once; the second invocation does not turn `snapshot` into a live member reference. The last Show performs a separate read of the latest Output. Both invocations happen to print the same text in this witness, so its `saved\nsaved\n` output alone is not proof of distinct storage; the [frontend witness](../../lyger/phase1_language_test.go) also checks that the saved value retains its binding and is not reevaluated. Neither witness was run here.

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

The unusual inclusive-order spelling follows the actual grammar sequence; it is not corrected into a conventional operator. Native typed-IR witnesses cover the operators, while a particular source phrase additionally depends on grammar and binding. A grammar match must not be presented as a newly executed program.

`not` or `no` negates only the condition forms whose grammar owns that modifier. Values do not have general truthiness, Text ordering is rejected, and mixed scalar comparisons do not coerce. The Error in a has condition must belong to the same target; it is not generic collection membership.

## Selection

If selects a branch from a supported condition. An attached else belongs to the unmatched if. Otherwise can attach to an immediately preceding unmatched conditional in the same source paragraph; it does not search arbitrary earlier paragraphs or create an unconditional fallback after unrelated statements.

**Supported equality and selection.** This program makes the wait explicit on the branch that continues:

<!-- example:values-truth-condition -->

`enabled` is Truth. When it equals true, Await joins Worker; otherwise Exit terminates the program. Every path that reaches the Output assignment has therefore joined. The existing native witness expects `condition\n`. This establishes the displayed equality form, not every operator combination in the table.

**Conditional ownership — excerpt from a frontend witness.** The [member-resolution source](../../lyger/test/member-resolution.l) declares an App, declares Text `x`, and executes the App before these clauses:

<!-- example:control-otherwise -->

The first `else` belongs to the Error test and assigns `y` from the completed invocation. Otherwise belongs to the immediately preceding `If x isnt y`, not the earlier Error test. The source spelling `it's Output` is retained exactly. This excerpt requires those preceding declarations and is not a standalone program. Its complete fixture compares `Hello World!` against the newline-bearing output from Print, so it must not be presented as a successful equality/output tutorial. The [member-resolution tests](../../lyger/member_resolution_regression_test.go) inspect binding and emitted instructions; no runtime success is claimed here.

The grammar can parse multiple conditions, but outer lowering explicitly rejects joined condition lists. Consequently `and`/`or` in the vocabulary does not establish general boolean short-circuiting. Coordination of actions, declarations, and operands is a different supported use of conjunctions.

## Iteration and collection operations

While and three-part source loops remain unsupported. [Collections](collections-and-categories.md) support Put, Place and typed reads; configured Get/Take words do not introduce further operations.

**Unsupported outer loop — explanatory source fragment, static inspection only.** This illustrates the configured While shape; it is deliberately not a runnable loop recipe:

<!-- example:control-while-unsupported -->

The [grammar's `conditional.while` rule](../../lexicon/eng/grammar.json) recognizes a condition followed by an imperative clause. The [instruction builder](../../lyger/instructions.go) accepts If/else/Otherwise conditionals and rejects the While kind. It supplies no supported way to repeat Show; use it to recognize the unsupported boundary.

The Rust backend separately supports bounded iteration compositions described in the [target reference](../../builtins/README.md#iteration-and-definition-contract). It rejects the three-part X form. Native loop composition does not confer source-level iteration semantics, collection ordering, bounds behavior, or break/continue vocabulary.

## Arithmetic and native calculations

Number is binary64, but the configured English grammar has no general infix arithmetic expression system. No public precedence table for `+`, `-`, multiplication, division, or arithmetic grouping can be inferred from numeric literals or D/E templates.

D/E calculates native expressions; G/I selects branches. A source-defined native-backed word may use supported target fragments deliberately. Such a definition documents its own typed contract and target dependence; it is not a built-in arithmetic operator family.

## Show and output

Show is an outer observation instruction. It writes Text bytes unchanged and formats supported runtime values without adding a newline. Direct App and Error display rejects. It does not implicitly join a process, replay unrelated captured output, or turn child stderr into stdout. Use explicit Await before Output when completion is required.

IO.Print is an imported application/native action that adds one newline. See [IO](../stdlib/eng/io.md#output-and-newlines) for raw-byte and numeric behavior. Mixing Print and Show without accounting for that newline is a real source-program error, not a reason to trim runtime output.

**Supported application print and outer observation — complete canonical example.**

<!-- example:basic -->

`The App Prints message` belongs to Hello's generated program and appends LF through imported IO.Print. Execute waits for that invocation; `Show Hello's Output` then writes the captured `Hello World!\n` exactly. Show does not add a second LF. The canonical CLI witness records that expected output.

## Exit and Return

Bare Exit terminates the whole L program after cleanup and succeeds if cleanup succeeds. Exit with an explicit Error/text failure follows the maintained error path and returns failure. Exit inside a continuation or expanded source action still exits the orchestrator; it is not merely a return from the nearest paragraph.

**Supported whole-program termination.**

<!-- example:control-bare-exit -->

Execute reaches First's continuation, which shows `First\n` and then exits. The following Show in the paragraph, Second's Execute operand, and the program's final Show are unreachable. The retained paragraph test expects that one output line and status 0 when cleanup succeeds. Compare the [explicit-error example](errors.md#runtime-failure-is-a-separate-stage), whose exit status is 1.

Function.Return belongs to the builtin native-backed Function facility. It emits target return syntax within its composed function. The presence of Return does not make Exit a function return, nor establish general source exceptions or arbitrary return-value transport between processes.

## Sources and witnesses

Primary sources: [instruction construction](../../lyger/instructions.go), [condition typing](../../lyger/orchestration_plan.go), [flow analysis](../../lyger/orchestration_flow.go), [emission](../../lyger/orchestration_emit.go), [condition kinds](../../lyger/orchestration_kinds_test.go), and [grammar](../../lexicon/eng/grammar.json). Existing witnesses: [instruction adjacency](../../lyger/instructions_test.go), [scalar conditions and snapshots](../../lyger/orchestrator_test.go), [paragraph grammar coordination](../../lyger/paragraph_grammar_architecture_test.go), [bare Exit](../../demo/paragraphs/exit-program.l), and [error Exit](../../demo/paragraphs/exit-error.l).
