package Lyger

import (
	"strings"
	"testing"
)

func TestFunctionHeaderPreservesControlFlow(t *testing.T) {
	conditional := strings.Replace(identityFunction, "Publicly create", "If false, Publicly create", 1)
	generated, err := functionGenerated(t, conditional)
	if err != nil || !strings.Contains(generated, "if false") {
		t.Fatalf("Function header lost its condition: %v", err)
	}
	if _, err := functionGenerated(t, conditional+"Identity `skipped` as \"Result\".\n"); err == nil || !strings.Contains(err.Error(), "not definitely assigned") {
		t.Fatalf("conditional Function binding was treated as initialized: %v", err)
	}
	effects := strings.Replace(identityFunction, "and Outputs of Text.", "and Outputs of Text and then Show `header-effect`.", 1)
	generated, err = functionGenerated(t, effects)
	if err != nil || !strings.Contains(generated, "header-effect") {
		t.Fatalf("Function header lost its sequenced effect: %v", err)
	}
	if _, err := functionGenerated(t, strings.Replace(conditional, "Return Value.", "Otherwise, Show `not a body branch`.\nReturn Value.", 1)); err == nil || !strings.Contains(err.Error(), "otherwise") {
		t.Fatalf("Function body Otherwise attached to its declaration's conditional: %v", err)
	}
	if _, err := functionGenerated(t, "    If true, Publicly create \"First\" as a New Function with No Inputs and No Outputs, else Publicly create \"Second\" as a New Function with No Inputs and No Outputs.\nReturn.\n"); err == nil || !strings.Contains(err.Error(), "one Function body") {
		t.Fatalf("two Functions claimed one paragraph body: %v", err)
	}
}

func TestFunctionDeclarationBoundariesNative(t *testing.T) {
	source := "    If false, Publicly create \"Unused\" as a New Function with No Inputs and No Outputs and then Show `skipped|`.\nReturn.\n\n" +
		strings.Replace(identityFunction, "and Outputs of Text.", "and Outputs of Text and then Show `header|`.", 1) +
		"Define \"FirstCopy\" as Identity.\n\"SecondCopy\" is FirstCopy.\nSecondCopy `copy|` as \"Copied\".\nShow Copied.\n\n" +
		"    Publicly create \"Apply\" as a New Function with Inputs of \"Callback\" as Identity and Outputs of Text.\nDefine \"Saved\" as Callback.\nSaved `parameter|` as \"Result\".\nReturn Result.\n\nApply Identity as \"Applied\".\nShow Applied.\n\n" +
		"    Publicly create \"Select\" as a New Function with No Inputs and Outputs of Identity.\nReturn Identity.\n\nSelect as \"Returned\".\nDefine \"Saved\" as Returned.\nSaved `result` as \"Selected\".\nShow Selected.\n"
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
			_, err := functionGenerated(t, identityFunction+declaration+"\nSaved `copy` as \"Result\".\nShow Result.\n")
			if err != nil {
				t.Fatal(err)
			}
		})
	}
	const apply = "    Publicly create \"Apply\" as a New Function with Inputs of \"Callback\" as Identity and Outputs of Text.\nDefine \"Saved\" as Callback.\nSaved `parameter` as \"Result\".\nReturn Result.\n\nApply Identity as \"Result\".\nShow Result.\n"
	if _, err := functionGenerated(t, identityFunction+apply); err != nil {
		t.Fatal(err)
	}
	const selectCopy = "    Publicly create \"Select\" as a New Function with No Inputs and Outputs of Identity.\nReturn Identity.\n\nSelect as \"Returned\".\nDefine \"Saved\" as Returned.\nSaved `result` as \"Result\".\nShow Result.\n"
	if _, err := functionGenerated(t, identityFunction+selectCopy); err != nil {
		t.Fatal(err)
	}
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
	c := compileMemberResolutionSource(t, memberResolutionSource{main: functionHeaderCaptureSource})
	if c.Error != nil {
		t.Fatal(c.Error)
	}
	for _, function := range c.Semantics.Functions {
		if len(function.Captures) != 2 || function.Captures[0].Record.ID != "Items" || function.Captures[1].Record.ID != "Handle" || len(function.CaptureCalls) != 0 {
			t.Fatalf("header captures included parameters/body locals or lost header siblings: %#v", function.Captures)
		}
	}
	if _, err := functionGenerated(t, functionHeaderCaptureSource); err != nil {
		t.Fatal(err)
	}
	if _, err := functionGenerated(t, "    Publicly create \"Read\" as a New Function with No Inputs and Outputs of List of Text and then Define \"Items\" as a New List of `late`.\nReturn Items.\n\n"); err == nil || !strings.Contains(err.Error(), "not definitely assigned") {
		t.Fatalf("header capture read a value initialized after its binding: %v", err)
	}
}
