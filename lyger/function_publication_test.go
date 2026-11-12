package Lyger

import (
	"strings"
	"testing"
)

var functionPublicationSource = functionSource("Publish|-es|-ed|-ing|-er", "\"Value\" as Text", "Text", "Publicly define \"Message\" as Value.\nPublicly create \"Report|-s|-ed|-ing|-er\" for Showing Message and then Showing any Text.\nReport `|`.\nReturn Message.") + "Publish `first` as \"First\".\nPublish `second` as \"Second\".\nShow First and Second.\n"

func TestFunctionBodyPublication(t *testing.T) {
	c := executableTestResult(t, memberResolutionSource{main: functionPublicationSource}).Compilation
	plan := requireOrchestrationPlan(t, c.Semantics)
	function := processTestFunction(plan.Root.Body)
	if function == nil || len(function.Captures) != 0 || len(function.CaptureHandles) != 0 {
		t.Fatal("Function-owned public declarations were captured before execution")
	}
	for _, name := range []string{"Message", "Report"} {
		rec := c.Memory[name]
		if rec == nil || rec.storageScope != c.Scope.Root {
			t.Fatalf("%s lost its declaring-unit publication", name)
		}
	}
	requireFunctionPlan(t, functionPublicationSource)
}

func TestFunctionPublicValuesRemainDefinitelyAssigned(t *testing.T) {
	for _, source := range []string{
		strings.Replace(functionPublicationSource, "Publicly define \"Message\" as Value.", "If false, Publicly define \"Message\" as Value.", 1),
		strings.Replace(functionPublicationSource, "Publicly define \"Message\" as Value.", "Show Message.\nPublicly define \"Message\" as Value.", 1),
		functionPublicationSource + "Show Message.\n",
		functionPublicationSource + "Report `outside`.\n",
	} {
		_, err := compileFunctionPlan(t, source)
		if err == nil || !strings.Contains(err.Error(), "not definitely") {
			t.Fatalf("unproven public initialization accepted: %v", err)
		}
	}
}

const functionMapSource = "Define \"callbacks\" as a New Map of Echo.\nPut Echo in callbacks as `key`.\nDefine \"replay|-s|-ed|-ing|-er\" as a New Echo from callbacks at `key`.\nreplay `mapped` as \"Result\".\nShow Result.\n"

func TestFunctionSignatureMap(t *testing.T) {
	requireFunctionPlan(t, identityFunction+functionMapSource)
}
