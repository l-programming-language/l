package Lyger

import (
	"strings"
	"testing"
)

func dynamicAppTestRecord(name string) *record {
	_, ref := planTestApp(name)
	rec := ref.Record
	rec.Value.Runtime = &runtimeReference{}
	return rec
}

func TestDynamicAppFunctionParameterUsesRuntimeStorage(t *testing.T) {
	source := memberResolutionSource{main: "Create \"First\" as a New Basic App.\nCreate \"Second\" as a New Basic App.\n\n" + functionSource("Join", "\"Target\" as App", "", "Run Target.\nAwait Target.\nIf Target has an Error, exit with Target's Error.") + "Join First.\n"}
	compilation := executableTestResult(t, source).Compilation
	plan := requireOrchestrationPlan(t, compilation.Semantics)
	function := processTestFunction(plan.Root.Body)
	if function == nil {
		t.Fatal("Function binding missing")
	}
	for _, instruction := range function.Body.Body {
		if instruction.Source.Kind == instructionRun || instruction.Source.Kind == instructionAwait {
			if instruction.Target.App != nil || instruction.Target.Storage == nil || !instruction.Target.Storage.Parameter || len(instruction.AppCases) != 2 {
				t.Fatal("App parameter was resolved to a placeholder canonical App")
			}
		}
	}
}

func TestDynamicAppMemberConditionRetainsOwnerIdentity(t *testing.T) {
	first, firstRef := planTestApp("First")
	a, b := dynamicAppTestRecord("A"), dynamicAppTestRecord("B")
	semantics := newSemanticProgram()
	semantics.Apps = []*App{first}
	condition := &instructionCondition{Operator: "has", Left: planTestRef(a), Right: &instructionValue{Kind: operandMember, Type: valueError, Owner: planTestRef(b), Member: memberError}}
	semantics.Instructions = []*instruction{{Kind: instructionAssign, Target: planTestRef(a), Value: firstRef}, {Kind: instructionAssign, Target: planTestRef(b), Value: firstRef}, {Kind: instructionIf, Condition: condition}}
	if _, err := planOrchestration(semantics); err == nil || !strings.Contains(err.Error(), "different application") {
		t.Fatalf("independent dynamic owners were conflated: %v", err)
	}
	condition.Right.Owner = condition.Left
	requireOrchestrationPlan(t, semantics)
}

func TestFunctionAwaitUsesProgramRootsAndPreservesCallerSnapshot(t *testing.T) {
	source := memberResolutionSource{main: functionAwaitRootsSource}
	compilation := executableTestResult(t, source).Compilation
	plan := requireOrchestrationPlan(t, compilation.Semantics)
	if !plan.FunctionProgramState {
		t.Fatal("Function continuation did not retain program state")
	}
	function := processTestFunction(plan.Root.Body)
	if function == nil {
		t.Fatal("Function missing")
	}
	await := function.Body.Body[0]
	if len(await.AppCases) != 1 || await.AppCases[0].Continuation == nil {
		t.Fatal("dynamic Await lost Worker continuation")
	}
	continuation := await.AppCases[0].Continuation
	write, read := continuation.Body[0].Target.Storage, continuation.Body[1].Value.Owner.Storage
	returned := function.Body.Body[1].Arguments[0].Storage
	if write.Capture != nil || write.Owner != plan.Root || read != write {
		t.Fatal("continuation was rebound to caller capture instead of program root")
	}
	if returned.Capture == nil || returned == write {
		t.Fatal("caller did not resume its independent captured snapshot")
	}
	result := &CompilationResult{Compilation: compilation}
	if _, err := result.OrchestratorUnit("entry", executableTestCompanions(result)); err != nil {
		t.Fatal(err)
	}
}

func TestFunctionInvocationDoesNotInventGlobalAssignmentFacts(t *testing.T) {
	source := memberResolutionSource{main: "    Create \"Worker\" as a New Public Basic App.\nPublicly define \"Late\" as Worker's Output.\n\n" + functionSource("Join", "\"Target\" as App", "", "Await Target.") + "Run Worker.\nJoin Worker.\nShow Late.\n"}
	compilation := executableTestResult(t, source).Compilation
	if _, err := planOrchestration(compilation.Semantics); err == nil || !strings.Contains(err.Error(), "Late is not definitely assigned") {
		t.Fatalf("Function invocation invented a caller initialization fact: %v", err)
	}
}

func processTestFunction(body []*orchestrationInstruction) *orchestrationFunction {
	for _, current := range body {
		if current.Function != nil {
			return current.Function
		}
		if current.Block != nil {
			if result := processTestFunction(current.Block.Body); result != nil {
				return result
			}
		}
	}
	return nil
}

var functionAwaitRootsSource = "Define \"Message\" as a New List of `before` and `before`.\nDefine \"Witness\" as a New Text from Message at 0.\nCreate \"Report|-s|-ed|-ing\" for Showing Witness and then Showing any Text.\n\n    Create \"Worker\" as a New Public Basic App.\nPlace `after` in Message at 0.\nDefine \"Observed\" as a New Text from Message at 0.\nShow Observed.\nReport `|`.\n\n" + functionSource("Join", "\"Target\" as App", "List of Text", "Await Target.\nReturn Message.") + "Run Worker.\nJoin Worker as \"Snapshot\".\nDefine \"Copied\" as a New Text from Snapshot at 0.\nShow Copied.\nDefine \"Current\" as a New Text from Message at 0.\nShow Current.\n"
