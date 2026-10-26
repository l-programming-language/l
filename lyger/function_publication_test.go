package Lyger

import (
	"strings"
	"testing"
)

const functionPublicationSource = "    Publicly create \"Publish\" as a New Function with Inputs of \"Value\" as Text and Outputs of Text.\nPublicly define \"Message\" as Value.\nPublicly create \"Report|-s|-ed|-ing\" for Showing Message and then Showing any Text.\nReport `|`.\nReturn Message.\n\nCall Publish with `first` as \"First\".\nCall Publish with `second` as \"Second\".\nShow First and Second.\n"

func TestFunctionBodyPublication(t *testing.T) {
	c := compileMemberResolutionSource(t, memberResolutionSource{main: functionPublicationSource})
	if c.Error != nil {
		t.Fatal(c.Error)
	}
	plan, err := planOrchestration(c.Semantics)
	if err != nil {
		t.Fatal(err)
	}
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
	if _, err := functionGenerated(t, functionPublicationSource); err != nil {
		t.Fatal(err)
	}
}

func TestFunctionPublicValuesRemainDefinitelyAssigned(t *testing.T) {
	for _, source := range []string{
		strings.Replace(functionPublicationSource, "Publicly define \"Message\" as Value.", "If false, Publicly define \"Message\" as Value.", 1),
		strings.Replace(functionPublicationSource, "Publicly define \"Message\" as Value.", "Show Message.\nPublicly define \"Message\" as Value.", 1),
		functionPublicationSource + "Show Message.\n",
		functionPublicationSource + "Report `outside`.\n",
	} {
		_, err := functionGenerated(t, source)
		if err == nil || !strings.Contains(err.Error(), "not definitely") {
			t.Fatalf("unproven public initialization accepted: %v", err)
		}
	}
}

func TestFunctionBodyPublicationNative(t *testing.T) {
	c := compileMemberResolutionSource(t, memberResolutionSource{main: functionPublicationSource})
	captureNativeResult(t, c, "first|second|firstsecond")
}

const functionMapSource = "Define \"callbacks\" as a New Map of Identity.\nPut Identity in callbacks as `key`.\nDefine \"callback\" as a New Identity from callbacks at `key`.\nCall callback with `mapped` as \"Result\".\nShow Result.\n"

func TestFunctionSignatureMap(t *testing.T) {
	_, err := functionGenerated(t, identityFunction+functionMapSource)
	if err != nil {
		t.Fatal(err)
	}
}

func functionMapExtractionSource() string {
	source := identityFunction + strings.Replace(functionMapSource, "Define \"callback\"", "Define \"boxed\" as a New Thing of callbacks.\nDefine \"copied\" as a New Map of Identity from boxed.\nDefine \"callback\"", 1)
	return strings.Replace(source, "from callbacks at", "from copied at", 1)
}

func TestFunctionSignatureMapExtraction(t *testing.T) {
	generated, err := functionGenerated(t, functionMapExtractionSource())
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(generated, ".expect_type(LType::Map(Box::new(LType::Function(") {
		t.Fatal("bound Function Map type lost its typed read")
	}
}

func TestFunctionSignatureMapNative(t *testing.T) {
	c := compileMemberResolutionSource(t, memberResolutionSource{main: functionMapExtractionSource()})
	captureNativeResult(t, c, "mapped")
}
