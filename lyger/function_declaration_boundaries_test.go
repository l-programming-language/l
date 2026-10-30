package Lyger

import (
	"strings"
	"testing"
)

func TestFunctionDeclarationBoundariesNative(t *testing.T) {
	source := "    If false, Publicly create \"Unused\" as a New Function with No Inputs and No Outputs and then Show `skipped|`.\nReturn.\n\n" +
		strings.Replace(identityFunction, "and Outputs of Text.", "and Outputs of Text and then Show `header|`.", 1) +
		"Define \"FirstCopy\" as Identity.\n\"SecondCopy\" is FirstCopy.\nSecondCopy `copy|` as \"Copied\".\nShow Copied.\n\n" +
		functionSource("Apply", "\"Callback\" as Identity", "Text", "Define \"Saved\" as Callback.\nSaved `parameter|` as \"Result\".\nReturn Result.") + "Apply Identity as \"Applied\".\nShow Applied.\n\n" +
		functionSource("Select", "", "Identity", "Return Identity.") + "Select as \"Returned\".\nDefine \"Saved\" as Returned.\nSaved `result` as \"Selected\".\nShow Selected.\n"
	c := compileMemberResolutionSource(t, memberResolutionSource{main: source + "\n" + functionHeaderCaptureSource})
	captureNativeResult(t, c, "header|copy|parameter|resultsnapshotshareddone")
}

func TestFunctionCopyRetainsSignature(t *testing.T) {
	for _, declaration := range []string{
		"Define \"Saved\" as Identity.",
		"\"Saved\" is Identity.",
		"Define \"First\" as Identity.\nDefine \"Saved\" as First.",
		"Use Identity as \"Alias\".\nDefine \"Saved\" as Alias.",
	} {
		t.Run(declaration, func(t *testing.T) {
			requireFunctionGenerated(t, identityFunction+declaration+"\nSaved `copy` as \"Result\".\nShow Result.\n")
		})
	}
	var apply = functionSource("Apply", "\"Callback\" as Identity", "Text", "Define \"Saved\" as Callback.\nSaved `parameter` as \"Result\".\nReturn Result.") + "Apply Identity as \"Result\".\nShow Result.\n"
	requireFunctionGenerated(t, identityFunction+apply)
	var selectCopy = functionSource("Select", "", "Identity", "Return Identity.") + "Select as \"Returned\".\nDefine \"Saved\" as Returned.\nSaved `result` as \"Result\".\nShow Result.\n"
	requireFunctionGenerated(t, identityFunction+selectCopy)
	for _, use := range []string{"Saved 7 as \"Result\".", "Saved `copy`."} {
		if _, err := functionGenerated(t, identityFunction+"Define \"Saved\" as Identity.\n"+use+"\n"); err == nil {
			t.Fatalf("copied Function accepted incompatible call: %s", use)
		}
	}
	if _, err := functionGenerated(t, identityFunction+"If false, Define \"Saved\" as Identity.\nSaved `skipped` as \"Result\".\n"); err == nil || !strings.Contains(err.Error(), "not definitely assigned") {
		t.Fatalf("copied signature invented runtime initialization: %v", err)
	}
}

func TestFunctionModuleHeaderHasNoInitializationEffects(t *testing.T) {
	c := compileMemberResolutionSource(t, memberResolutionSource{main: "Use `local-words.l`.\nIdentity `value` as \"Result\".\nShow Result.\n", library: identityFunction})
	if c.Error != nil {
		t.Fatalf("program-file Function include: %v", c.Error)
	}
	for _, suffix := range []string{"", " and then Show `effect`", "conditional"} {
		files := dictionaryIndexFilesystem(t)
		header := "Publicly create"
		tail := suffix
		if suffix == "conditional" {
			header, tail = "If false, Publicly create", ""
		}
		files["stdlib/eng/io/dictionary.l"].Data = []byte("Create \"IO\" as a Dictionary.\n\n    " + header + " IO's \"Identity\" as a New Function with Inputs of \"Value\" as Text and Outputs of Text" + tail + ".\nReturn Value.\n")
		_, err := NewLygerService(Options{Resources: files}).Compile([]byte("Use IO's Identity from `io` as \"Identity\".\nIdentity `value` as \"Result\".\nShow Result.\n"), "main.l")
		if err == nil || !strings.Contains(err.Error(), "initialization effects") {
			t.Fatalf("module Function header executed initialization (%q): %v", suffix, err)
		}
	}
}

const functionHeaderCaptureSource = "Define \"Outer\" as a New Pointer of `first`.\n\n    Publicly Define \"Items\" as a New List of `snapshot` and then Define \"Handle\" as Outer and then Publicly create \"HeaderRead\" as a New Function with Inputs of \"Suffix\" as Text and Outputs of Text and then Put `shared` in Outer.\nDefine \"Local\" as a New Text from Items at 0.\nDefine \"Shared\" as a New Text from Handle.\nShow Local and Shared.\nReturn Suffix.\n\nPlace `changed` in Items at 0.\nHeaderRead `done` as \"Final\".\nShow Final.\n"

func TestFunctionHeaderCapturesOnlyOuterValues(t *testing.T) {
	c := executableTestResult(t, memberResolutionSource{main: functionHeaderCaptureSource}).Compilation
	for _, function := range c.Semantics.Functions {
		if len(function.Captures) != 2 || function.Captures[0].Record.ID != "Items" || function.Captures[1].Record.ID != "Handle" || len(function.CaptureCalls) != 0 {
			t.Fatalf("header captures included parameters/body locals or lost header siblings: %#v", function.Captures)
		}
	}
	requireFunctionGenerated(t, functionHeaderCaptureSource)
	if _, err := functionGenerated(t, functionSource("Read", "", "List of Text and then Define \"Items\" as a New List of `late`", "Return Items.")); err == nil || !strings.Contains(err.Error(), "not definitely assigned") {
		t.Fatalf("header capture read a value initialized after its binding: %v", err)
	}
}
