# Decisions, sequencing and termination

L executes resolved actions in source order. The examples below distinguish supported source control flow from grammar or backend facilities.

## Conditions and Exit

<!-- example:values-truth-condition -->

`If enabled equals true, await Worker, else exit.` has a condition, a comma before the selected action, and an optional `else` action. Here the true branch joins Worker; the false branch terminates the program. Only the joined path reaches Output. Expected stdout is `condition\n`.

`Otherwise` can supply the alternative to an immediately preceding unmatched If in the same paragraph:

<!-- example:control-otherwise -->

This excerpt requires the fixture's earlier App and Text declarations. The inline `else` belongs to the Error test; Otherwise belongs to `If x isnt y`. It cannot attach across a blank paragraph boundary or intervening statement. This is a binding witness, not a successful output recipe: the fixture compares Text with newline-bearing Print output.

Bare Exit ends the whole program, including from a continuation, and succeeds if cleanup succeeds. The [error-Exit example](../../demo/paragraphs/exit-error.l) instead writes `First\n` to stdout, `paragraph stopped\n` to stderr and exits 1; later actions do not run.

## Sequence before assuming parallelism

<!-- example:two-apps -->

Start expands Build/Run for First, then Second. Their processes may overlap. Await joins Second before First; the Shows choose display order, producing `First\nSecond\n`.

`Execute First and Second` instead completes Start/Await and the continuation for First before starting Second. Coordinating action operands with `and` is supported; joining boolean conditions with `and` or `or` is rejected. See [comparisons and selection](../spec/control-flow.md#comparisons) for exact forms, including `isnt` and Error checks.

## Branches and value availability

A value must be assigned on every path that continues to its use; Exit paths cannot reach later reads. The compiler supplies neither hidden defaults nor blanket cloning. Output readiness is checked separately at runtime.

## Iteration has a bounded implementation surface

`While` is rejected in the enclosing program; general source collection iteration is unsupported. Backend loop compositions do not supply L `for`/`while` syntax, arithmetic precedence or short-circuit guarantees. The [iteration reference](../reference/logemes/iteration.md) and [Lucid suite](../../logiceme/testdata/lucid/README.md) document compile-only backend witnesses and removed forms.

Continue with [imports and dictionaries](imports.md).
