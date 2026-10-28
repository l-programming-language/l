package Lyger

import (
	"strings"
	"testing"
)

var sourceActionArgumentCases = []struct{ name, source, output string }{
	{"List", "Define \"Items\" as a New List of `before`.\nCreate \"Report|-s|-ed|-ing\" for Showing any List.\nReport Items.\n", "[before]"},
	{"Map", "Define \"Items\" as a New Map of Text.\nPut `before` in Items as `key`.\nCreate \"Report|-s|-ed|-ing\" for Showing any Map.\nReport Items.\n", "{key: before}"},
	{"Null", "Define \"Nothing\" as Null.\nCreate \"Report|-s|-ed|-ing\" for Showing any Null.\nReport Nothing.\n", "null"},
	{"caller", "Define \"Items\" as a New List of `before`.\nCreate \"Report|-s|-ed|-ing\" for Showing Items and then Showing any Text.\nPlace `after` in Items at 0.\nReport a New Text from Items at 0.\n", "[before]after"},
	{"alias", "Define \"Items\" as a New List of `before`.\nCreate \"Report|-s|-ed|-ing\" for Showing Items and then Showing any Text.\nUse Report as \"Saved\".\nPlace `after` in Items at 0.\nSaved a New Text from Items at 0.\n", "[before]after"},
	{"forward", "Define \"Items\" as a New List of `before`.\nCreate \"Report|-s|-ed|-ing\" for Showing Items and then Showing any Text.\nCreate \"Forward|-s|-ed|-ing\" for Reporting any Text.\nPlace `after` in Items at 0.\nForward a New Text from Items at 0.\n", "[before]after"},
	{"once", "Define \"Items\" as a New List of `before`.\nCreate \"Report|-s|-ed|-ing\" for Showing any Text and then Showing it.\nReport a New Text from Items at 0.\n", "beforebefore"},
	{"caller-mutation", "Define \"Items\" as a New Pointer of `before`.\nCreate \"Report|-s|-ed|-ing\" for Putting `after` in Items and then Showing any Text.\nReport a New Text from Items.\n", "before"},
	{"List-snapshot", "Define \"Items\" as a New List of `before`.\nCreate \"Report|-s|-ed|-ing\" for Showing any List and then Placing `after` in it at 0.\nReport Items.\nShow Items.\n", "[before][before]"},
	{"Pointer-sharing", "Define \"Items\" as a New Pointer of `before`.\nCreate \"Report|-s|-ed|-ing\" for Showing any Pointer and then Putting `after` in it.\nReport Items.\nShow Items.\n", "beforeafter"},
}

func TestSourceActionArgumentBoundaries(t *testing.T) {
	for _, tc := range sourceActionArgumentCases {
		t.Run(tc.name, func(t *testing.T) {
			c := executableTestResult(t, memberResolutionSource{main: tc.source}).Compilation
			plan := requireOrchestrationPlan(t, c.Semantics)
			found := false
			visitCapturePlan(plan.Root.Body, func(call *orchestrationInstruction) {
				if call.Source.Kind != instructionCall || call.Source.Target.Record.ID != "Report" {
					return
				}
				found = true
				value := call.Actions[len(call.Actions)-1].Value
				if call.Actions[len(call.Actions)-1].Source.Kind == instructionUpdate {
					value = call.Actions[0].Value
				}
				if value.Storage == nil || value.Storage.Capture != nil {
					t.Fatal("actual argument became a constructor or callee capture")
				}
				if value.Type == valueList || value.Type == valueMap {
					if value.RuntimeType.Element == nil || value.RuntimeType.Element.Kind != valueText {
						t.Fatal("argument lost its complete element type")
					}
				}
			})
			if !found {
				t.Fatal("source action was not planned")
			}
		})
	}
}

func TestSourceActionArgumentsNative(t *testing.T) {
	var source, output strings.Builder
	for _, tc := range sourceActionArgumentCases {
		source.WriteString("\n    " + tc.source)
		output.WriteString(tc.output)
	}
	captureNativeResult(t, compileMemberResolutionSource(t, memberResolutionSource{main: source.String()}), output.String())
}
