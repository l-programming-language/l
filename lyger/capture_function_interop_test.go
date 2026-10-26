package Lyger

import (
	nativetest "github.com/l-programming-language/l/internal/testsupport/native"
	Logiceme "github.com/l-programming-language/l/logiceme"
	"os"
	"path/filepath"
	"testing"
)

const captureFunctionInteropSource = "Define \"Items\" as a New List of `kept` and `other`.\nDefine \"Payload\" as a New Text from Items at 0.\nCreate \"Report|-s|-ed|-ing\" for Showing Payload and then Showing any Text.\nUse Report as \"Saved\".\n\n    Publicly create \"Invoke\" as a New Function with Inputs of \"Suffix\" as Text and No Outputs.\nSaved Suffix.\n\nCall Invoke with `one`.\nCall Invoke with `two`.\n"

func TestFunctionCapturesSourceActionEnvironment(t *testing.T) {
	c := compileMemberResolutionSource(t, memberResolutionSource{main: captureFunctionInteropSource})
	if c.Error != nil {
		t.Fatal(c.Error)
	}
	plan, err := planOrchestration(c.Semantics)
	if err != nil {
		t.Fatal(err)
	}
	function := processTestFunction(plan.Root.Body)
	if function == nil || len(function.CaptureHandles) != 1 || len(function.Captures) != 0 {
		t.Fatal("Function did not retain exactly the selected action environment")
	}
	call := function.Body.Body[0]
	if call.CaptureSource == nil || call.CaptureSource.Capture == nil || call.CaptureSource.Slot != nil {
		t.Fatal("Function action call rereads a source binding instead of its saved handle")
	}
	if call.Actions[0].Value.Storage.Capture == nil || call.Actions[1].Value.Storage != function.Inputs[0] {
		t.Fatal("action environment and Function parameter ownership were conflated")
	}
	if _, err := functionGenerated(t, captureFunctionInteropSource); err != nil {
		t.Fatal(err)
	}
}

func TestFunctionCapturesSourceActionNative(t *testing.T) {
	c := compileMemberResolutionSource(t, memberResolutionSource{main: captureFunctionInteropSource})
	captureNativeResult(t, c, "keptonekepttwo")
}

func TestFunctionAwaitProgramRootsNative(t *testing.T) {
	c := compileMemberResolutionSource(t, memberResolutionSource{main: functionAwaitRootsSource})
	captureNativeResult(t, c, "afterbefore|beforeafter")
}

func TestCaptureActivationAliasesNative(t *testing.T) {
	c := compileMemberResolutionSource(t, memberResolutionSource{main: "Show `unused`.\n"})
	scope := &lexicalScope{kind: lexicalParagraph}
	value := planTestScalar("private", valueText, scope)
	action, alias := captureTestAction("Report"), captureTestAction("Saved")
	semantics := newSemanticProgram()
	semantics.Aliases[alias] = action
	bind := captureTestBind(action, value)
	first := &instruction{Kind: instructionBlock, Block: &scopedInstructions{Scope: scope, Body: []*instruction{captureTestSet(value, "first"), bind, captureTestCall(action, captureTestSet(value, "mutated"), captureTestShow(value)), captureTestShow(value)}}}
	second := &instruction{Kind: instructionBlock, Block: &scopedInstructions{Scope: scope, Body: []*instruction{captureTestSet(value, "second"), bind}}}
	semantics.Instructions = []*instruction{first, {Kind: instructionAliasAction, Callable: alias, Target: &instructionValue{Kind: operandSymbol, Record: action}}, second, captureTestCall(alias, captureTestShow(value)), captureTestCall(action, captureTestShow(value))}
	c.Semantics = semantics
	captureNativeResult(t, c, "mutatedfirstmutatedsecond")
}

func captureNativeResult(t *testing.T, c *Compilation, expected string) {
	t.Helper()
	if c.Error != nil {
		t.Fatal(c.Error)
	}
	result := &CompilationResult{Compilation: c}
	units, err := result.ApplicationUnits()
	if err != nil {
		t.Fatal(err)
	}
	service := Logiceme.NewLogicemeService(canonicalTestOptions().Resources)
	children := make(map[string]string)
	for _, unit := range units {
		bundle, err := service.Evaluate(unit.BackendAML)
		if err != nil {
			t.Fatal(err)
		}
		children[unit.ID] = buildActualBundle(t, bundle, unit.ID, false, "")
	}
	unit, err := result.OrchestratorUnit("entry", executableTestCompanions(result))
	if err != nil {
		t.Fatal(err)
	}
	bundle, err := service.Evaluate(unit.BackendAML, Logiceme.GenerationOptions{IncludeRuntime: true})
	if err != nil {
		t.Fatal(err)
	}
	recordBundle(t, bundle)
	binary := buildActualBundle(t, bundle, "entry", true, "")
	for name, child := range children {
		target := filepath.Join(filepath.Dir(binary), "main.apps", name)
		data, err := os.ReadFile(child)
		if err != nil {
			t.Fatal(err)
		}
		if err := os.MkdirAll(filepath.Dir(target), 0700); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(target, data, 0700); err != nil {
			t.Fatal(err)
		}
	}
	stdout, stderr, err := nativetest.RuntimeExecution.Run(t, binary, "", nil)
	if err != nil || len(stderr) != 0 || string(stdout) != expected {
		t.Fatalf("captures stdout=%q stderr=%q error=%v; expected %q", stdout, stderr, err, expected)
	}
}
