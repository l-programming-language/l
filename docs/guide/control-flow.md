# Decisions, sequencing and termination

L preserves ordered source actions. Grammar recognition, semantic binding and lowering all contribute to whether a form executes. A familiar English word or a backend loop template is not enough to establish a source-language control-flow feature.

## Conditions and Exit

The maintained error-exit program demonstrates an ordinary condition and the effect of terminating from a continuation:

<!-- example:paragraph-error-exit -->

After First finishes, its continuation shows finalized output. The equality condition is true, so `exit with an error of` stops the enclosing program. The later paragraph tail and later Execute operand do not run. The retained assertion is stdout `First\n`, stderr `paragraph stopped\n`, and status 1.

Bare Exit also terminates the enclosing program, but selects successful status when cleanup succeeds. It is not simply a return from an App's local paragraph. The [bare-Exit example](../../demo/paragraphs/exit-program.l) and [paragraph tests](../../cmd/l/paragraph_integration_test.go) preserve that distinction.

Condition spelling comes from the selected vocabulary and grammar. Existing sources use `equals`, `isnt`, error-member checks and conditional clauses. The [grammar reference](../spec/index.md) records the recognized alternatives and their support rather than translating them into invented C-style expressions.

## Sequence before assuming parallelism

<!-- example:two-apps -->

Start First and Second expands Build/Run for First, then Build/Run for Second. Because Run launches without joining, the two child processes may overlap. Await Second and First joins in that written order. The subsequent Shows explicitly choose the final display order, yielding the retained expected `First\nSecond\n`.

By comparison, `Execute First and Second` expands Start/Await for First and then Start/Await for Second. It waits for First's continuation before proceeding to Second. Source action composition and coordinated primitive operations must not be conflated when predicting scheduling.

## Branches and value availability

The frontend checks that a value is assigned on every path that continues to a use. It does not silently initialize missing values or clone everything to satisfy Rust. The current typed planner tracks required storage and mutability; member readiness such as Output still has its own runtime rule.

Do not infer undocumented short-circuit or operator-precedence guarantees from English phrasing. Read the supported condition/operator reference and use a maintained source witness for an exact combination.

## Iteration has a bounded implementation surface

The backend supports documented iteration compositions in its frozen Rust template surface and Lucid corpus. That does not establish a conventional source `for`, `while`, collection iterator or three-part loop syntax. This guide therefore does not provide invented runnable L loop forms.

The [logeme iteration reference](../reference/logemes/index.md) and [Lucid suite guide](../../logiceme/testdata/lucid/README.md) identify supported backend cases and removed forms. Those witnesses are compile-only at that boundary; they do not prove a general source-language iteration API or runtime scheduling model.

Continue with [imports and dictionaries](imports.md).
