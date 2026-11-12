package Lyger

import (
	"strings"
	"testing"
)

func TestFunctionCopyRetainsSignature(t *testing.T) {
	for _, declaration := range []string{
		"Define \"Replay|-s|-ed|-ing|-er\" as Echo.",
		"\"Replay|-s|-ed|-ing|-er\" is Echo.",
		"Define \"First\" as Echo.\nDefine \"Replay|-s|-ed|-ing|-er\" as First.",
		"Use Echo as \"Relay|-s|-ed|-ing|-er\".\nDefine \"Replay|-s|-ed|-ing|-er\" as Relay.",
	} {
		t.Run(declaration, func(t *testing.T) {
			requireFunctionPlan(t, identityFunction+declaration+"\nReplay `copy` as \"Result\".\nShow Result.\n")
		})
	}
	var apply = functionSource("Apply|--ies|--ied|-ing|--ier", "\"Forward|-s|-ed|-ing|-er\" as Echo", "Text", "Define \"Replay|-s|-ed|-ing|-er\" as Forward.\nReplay `parameter` as \"Result\".\nReturn Result.") + "Apply Echo as \"Result\".\nShow Result.\n"
	requireFunctionPlan(t, identityFunction+apply)
	var selectCopy = functionSource("Select|-s|-ed|-ing|-or", "", "Echo", "Return Echo.") + "Select as \"Returned\".\nDefine \"Replay|-s|-ed|-ing|-er\" as Returned.\nReplay `result` as \"Result\".\nShow Result.\n"
	requireFunctionPlan(t, identityFunction+selectCopy)
	for _, use := range []string{"Replay 7 as \"Result\".", "Replay `copy`."} {
		if _, err := compileFunctionPlan(t, identityFunction+"Define \"Replay|-s|-ed|-ing|-er\" as Echo.\n"+use+"\n"); err == nil {
			t.Fatalf("copied Function accepted incompatible call: %s", use)
		}
	}
	if _, err := compileFunctionPlan(t, identityFunction+"If false, Define \"Replay|-s|-ed|-ing|-er\" as Echo.\nReplay `skipped` as \"Result\".\n"); err == nil || !strings.Contains(err.Error(), "not definitely assigned") {
		t.Fatalf("copied signature invented runtime initialization: %v", err)
	}
}

func TestFunctionModuleHeaderHasNoInitializationEffects(t *testing.T) {
	c := compileMemberResolutionSource(t, memberResolutionSource{main: "Use `local-words.l`.\nEcho `value` as \"Result\".\nShow Result.\n", library: identityFunction})
	if c.Error != nil {
		t.Fatalf("program-file Function include: %v", c.Error)
	}
	for _, suffix := range []string{"", " and then Show `effect`", "conditional"} {
		files := dictionaryIndexFilesystem(t)
		installTestDictionary(t, files, "catalog", "Catalog", "Create \"Catalog\" as a Dictionary.\n")
		header := "Publicly create"
		tail := suffix
		if suffix == "conditional" {
			header, tail = "If false, Publicly create", ""
		}
		files["stdlib/eng/catalog/dictionary.l"].Data = []byte("Create \"Catalog\" as a Dictionary.\n\n    " + header + " Catalog's \"Echo|-es|-ed|-ing|-er\" as a New Function with Inputs of \"Value\" as Text and Outputs of Text" + tail + ".\nReturn Value.\n")
		_, err := NewLygerService(Options{Resources: files}).Compile([]byte("Use Catalog's Echo from `catalog` as \"Echo\".\nEcho `value` as \"Result\".\nShow Result.\n"), "main.l")
		if err == nil || !strings.Contains(err.Error(), "initialization effects") {
			t.Fatalf("module Function header executed initialization (%q): %v", suffix, err)
		}
	}
}

const functionHeaderCaptureSource = "Define \"Enclose\" as a New Pointer of `first`.\n\n    Publicly Define \"Items\" as a New List of `snapshot` and then Define \"Handle\" as Enclose and then Publicly create \"HeaderRead|-s|-|-ing|-er\" as a New Function with Inputs of \"Suffix\" as Text and Outputs of Text and then Put `shared` in Enclose.\nDefine \"Local\" as a New Text from Items at 0.\nDefine \"Shared\" as a New Text from Handle.\nShow Local and Shared.\nReturn Suffix.\n\nPlace `changed` in Items at 0.\nHeaderRead `done` as \"Final\".\nShow Final.\n"

func TestFunctionHeaderCapturesOnlyOuterValues(t *testing.T) {
	c := executableTestResult(t, memberResolutionSource{main: functionHeaderCaptureSource}).Compilation
	for _, function := range c.Semantics.Functions {
		if len(function.Captures) != 2 || function.Captures[0].Record.ID != "Items" || function.Captures[1].Record.ID != "Handle" || len(function.CaptureCalls) != 0 {
			t.Fatalf("header captures included parameters/body locals or lost header siblings: %#v", function.Captures)
		}
	}
	requireFunctionPlan(t, functionHeaderCaptureSource)
	if _, err := compileFunctionPlan(t, functionSource("Read|-s|-|-ing|-er", "", "List of Text and then Define \"Items\" as a New List of `late`", "Return Items.")); err == nil || !strings.Contains(err.Error(), "not definitely assigned") {
		t.Fatalf("header capture read a value initialized after its binding: %v", err)
	}
}
