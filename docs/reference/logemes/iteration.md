# Iteration and conditional binding

## X and P: four supported categories

X forwards iteration to P's Iterate primitive. Each supported form has one Operation body. Inputs select mutually exclusive categories:

| Category | Inputs | Rust shape |
| --- | --- | --- |
| Traversal | Iterable, existing Key/Index/RangeArity metadata, no condition child | `for pattern in iterable { body }` |
| Unconditional repetition | No Iterable or condition child | `loop { body }` |
| Boolean repetition | One Calculation or an I/Evaluation adapter holding ordered Calculations | `while condition { body }` |
| Definitional repetition | One OG or DOG Definition | `while let pattern = value { body }` |

The caller owns iterator expressions, patterns, borrowing and explicit native await. The generator inserts no iterator adapter, enumeration or channel-drain loop. I remains a Boolean condition adapter, not an initialization/step/body carrier.

## OG, DOG and PO

OG takes one Key/Keys native pattern and one nonempty Value/Values, without children. DOG takes the pattern and exactly one Calculation, without native Values. They only occur as X conditions and preserve `let pattern = value` as a fragment. PO attaches one native label to one Iteration, with no extra wrapper block, native Values or trailing semicolon.

This allows the label to target the loop itself. Condition bindings have their Rust loop-body scope; they do not become surrounding L paragraph bindings. Native pattern syntax is supplied data, not a general pattern language invented at the AML boundary.

## Unsupported forms and native failures

Three-part initialization/condition/step loops are unsupported and rejected before rendering. Multiple bodies, mixed iterator/condition inputs, wrong child roles, misplaced OG/DOG and invalid label compositions also reject.

[Conditions](../../../logiceme/testdata/lucid/iteration-conditions.aml.txt), [bindings](../../../logiceme/testdata/lucid/iteration-bindings.aml.txt), [traversal](../../../logiceme/testdata/lucid/iteration-traversal.aml.txt), [labels](../../../logiceme/testdata/lucid/iteration-labels.aml.txt), [nested scope](../../../logiceme/testdata/lucid/iteration-nesting-scope.aml.txt) and [explicit async](../../../logiceme/testdata/lucid/iteration-async.aml.txt) are bounded complete-AML witnesses. Native negative controls require E0308 for a nonboolean condition, E0425 for a binding used outside its scope and E0728 for await in synchronous context.

The corpus has fifty stored malformed iteration inputs. Generation and compilation do not run these loops or establish termination. Read [Iteration](../../../builtins/segments/vowels/iteration.lu), [Iterate](../../../builtins/segments/consonants/iterate.lu), [iteration tests](../../../logiceme/lucid_iteration_test.go), and the maintained [iteration contract](../../../builtins/README.md#iteration-and-definition-contract).
