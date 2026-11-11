package Lyger

import (
	"testing"
)

var captureFunctionInteropSource = "Define \"Items\" as a New List of `kept` and `other`.\nDefine \"Payload\" as a New Text from Items at 0.\nCreate \"Report|-s|-ed|-ing\" for Showing Payload and then Showing any Text.\nUse Report as \"Saved\".\n\n" + functionSource("Invoke", "\"Suffix\" as Text", "", "Saved Suffix.") + "Invoke `one`.\nInvoke `two`.\n"

func TestFunctionCapturesSourceActionEnvironment(t *testing.T) {
	c := executableTestResult(t, memberResolutionSource{main: captureFunctionInteropSource}).Compilation
	plan := requireOrchestrationPlan(t, c.Semantics)
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
	requireFunctionGenerated(t, captureFunctionInteropSource)
}
