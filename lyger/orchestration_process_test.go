package Lyger

import (
	distribution "github.com/l-programming-language/l"
	"strings"
	"testing"
)

func dynamicAppTestRecord(name string) *record {
	_, ref := planTestApp(name)
	rec := ref.Record
	rec.Value.Runtime = &runtimeReference{}
	return rec
}
func TestDynamicAppAwaitJoinsOnlyContinuingFacts(t *testing.T) {
	for _, scenario := range []string{"both", "missing", "exit"} {
		t.Run(scenario, func(t *testing.T) {
			first, firstRef := planTestApp("First")
			second, _ := planTestApp("Second")
			selected := dynamicAppTestRecord("Selected")
			result := planTestScalar("result", valueNumber, nil)
			assign := &instruction{Kind: instructionAssign, Target: planTestRef(result), Value: &instructionValue{Kind: operandLiteral, Type: valueNumber, Literal: &primitive{Type: valueNumber, Number: 7}}}
			semantics := newSemanticProgram()
			semantics.Apps = []*App{first, second}
			semantics.Continuations = map[*App]*scopedInstructions{first: {Scope: &lexicalScope{kind: lexicalParagraph}, Body: []*instruction{assign}}, second: {Scope: &lexicalScope{kind: lexicalParagraph}}}
			if scenario == "both" {
				semantics.Continuations[second].Body = []*instruction{assign}
			} else if scenario == "exit" {
				semantics.Continuations[second].Body = []*instruction{{Kind: instructionExit}}
			}
			semantics.Instructions = []*instruction{{Kind: instructionAssign, Target: planTestRef(selected), Value: firstRef}, {Kind: instructionAwait, Target: planTestRef(selected)}, {Kind: instructionShow, Value: planTestRef(result)}}
			plan, err := planOrchestration(semantics)
			if scenario == "missing" {
				if err == nil || !strings.Contains(err.Error(), "not definitely assigned") {
					t.Fatalf("incomplete dynamic continuation facts: %v", err)
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			await := plan.Root.Body[1]
			if await.Target.App != nil || await.Target.Storage == nil || len(await.AppCases) != 2 || await.AppCases[0].App != first || await.AppCases[1].App != second {
				t.Fatal("dynamic handle lost canonical dispatch alternatives")
			}
			if await.AppCases[0].Continuation == await.AppCases[1].Continuation {
				t.Fatal("distinct Apps shared one continuation activation")
			}
			catalog, err := loadBuiltinProcesses(distribution.Files())
			if err != nil {
				t.Fatal(err)
			}
			emitter := &orchestrationEmitter{plan: plan, apps: map[*App]string{first: "lApp0", second: "lApp1"}, processes: catalog}
			generated, err := emitter.processInstruction(await)
			if err != nil {
				t.Fatal(err)
			}
			source := strings.Join(generated, "\n")
			for _, part := range []string{"Arc::ptr_eq(&lSelectedApp0, &lApp0)", "Arc::ptr_eq(&lSelectedApp0, &lApp1)", "App handle is not owned by this program"} {
				if !strings.Contains(source, part) {
					t.Fatalf("missing %s in %s", part, source)
				}
			}
			if strings.Count(source, "claim_continuation") != 2 || strings.Count(source, "finish_continuation") != 2 {
				t.Fatal("dynamic Await bypassed selected builtin continuation claim/finish")
			}
		})
	}
}

func TestDynamicAppFunctionParameterUsesRuntimeStorage(t *testing.T) {
	source := memberResolutionSource{main: "Create \"First\" as a New Basic App.\nCreate \"Second\" as a New Basic App.\n\n    Publicly create \"Join\" as a New Function with Inputs of \"Target\" as App and No Outputs.\nRun Target.\nAwait Target.\nIf Target has an Error, exit with Target's Error.\n\nJoin First.\n"}
	compilation := compileMemberResolutionSource(t, source)
	if compilation.Error != nil {
		t.Fatal(compilation.Error)
	}
	plan, err := planOrchestration(compilation.Semantics)
	if err != nil {
		t.Fatal(err)
	}
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
	if _, err := planOrchestration(semantics); err != nil {
		t.Fatal(err)
	}
}

func TestFunctionAwaitUsesProgramRootsAndPreservesCallerSnapshot(t *testing.T) {
	source := memberResolutionSource{main: functionAwaitRootsSource}
	compilation := compileMemberResolutionSource(t, source)
	if compilation.Error != nil {
		t.Fatal(compilation.Error)
	}
	plan, err := planOrchestration(compilation.Semantics)
	if err != nil {
		t.Fatal(err)
	}
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
	source := memberResolutionSource{main: "    Create \"Worker\" as a New Public Basic App.\nPublicly define \"Late\" as Worker's Output.\n\n    Publicly create \"Join\" as a New Function with Inputs of \"Target\" as App and No Outputs.\nAwait Target.\n\nRun Worker.\nJoin Worker.\nShow Late.\n"}
	compilation := compileMemberResolutionSource(t, source)
	if compilation.Error != nil {
		t.Fatal(compilation.Error)
	}
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

const functionAwaitRootsSource = "Define \"Message\" as a New List of `before` and `before`.\nDefine \"Witness\" as a New Text from Message at 0.\nCreate \"Report|-s|-ed|-ing\" for Showing Witness and then Showing any Text.\n\n    Create \"Worker\" as a New Public Basic App.\nPlace `after` in Message at 0.\nDefine \"Observed\" as a New Text from Message at 0.\nShow Observed.\nReport `|`.\n\n    Publicly create \"Join\" as a New Function with Inputs of \"Target\" as App and Outputs of List of Text.\nAwait Target.\nReturn Message.\n\nRun Worker.\nJoin Worker as \"Snapshot\".\nDefine \"Copied\" as a New Text from Snapshot at 0.\nShow Copied.\nDefine \"Current\" as a New Text from Message at 0.\nShow Current.\n"
