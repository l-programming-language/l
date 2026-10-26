package Lyger

import (
	"strings"
	"testing"
)

func TestFunctionInvocationWords(t *testing.T) {
	for _, source := range []string{
		identityFunction + "Identity `direct` as \"Result\".\nShow Result.\n",
		identityFunction + "Define \"Saved\" as Identity.\nSaved `copy` as \"Result\".\nShow Result.\n",
		functionSource("Echo|-es|-ed|-ing|-er", "\"Value\" as Text", "Text", "Return Value.") + "Echo `direct` as \"Result\".\nShow Result.\n",
		functionSource("Done", "", "", "Show `done`.") + "Done.\n",
	} {
		c := executableTestResult(t, memberResolutionSource{main: source}).Compilation
		requireOrchestrationPlan(t, c.Semantics)
	}
	c := compileMemberResolutionSource(t, memberResolutionSource{main: identityFunction + "Call Identity with `unsupported` as \"Result\".\n"})
	if c.Error == nil {
		t.Fatal("removed Call vocabulary remains accepted")
	}
}

func TestFunctionWordAndValueShareBinding(t *testing.T) {
	word := newLexeme("verb")
	if err := word.reLabel(`"Echo|-es|-ed|-ing|-er"`); err != nil {
		t.Fatal(err)
	}
	rec := &record{ID: "Echo", Value: initEntry("Function", storageApplicative)}
	rec.Value.Applicative.Function.Source = &sourceFunction{Word: word}
	for _, spelling := range []string{"Echo", "Echoes", "Echoed", "Echoing", "Echoer"} {
		binding, ok := rec.binding(newBindingQuery([]byte(spelling)))
		if !ok || binding.Record != rec {
			t.Fatalf("%s lost the Function value binding", spelling)
		}
	}
	copy := *rec
	copy.ID = "Saved"
	if binding, ok := copy.binding(newBindingQuery([]byte("Saved"))); !ok || binding.Record != &copy {
		t.Fatal("copy has no invocation word")
	}
	if _, ok := copy.binding(newBindingQuery([]byte("Echo"))); ok {
		t.Fatal("copy retained the source word's root")
	}
}

func TestFunctionMemberSentence(t *testing.T) {
	const source = "Create \"User\" as a Dictionary.\nDefine \"Store\" as `store`.\n\n" +
		"    Publicly create User's \"Walk|-s|-ed|-ing|-er\" as a New Function with Inputs of \"Location\" as Text and No Outputs.\nShow Location.\n\n" +
		"The User walks to the Store.\n"
	c := executableTestResult(t, memberResolutionSource{main: source}).Compilation
	requireOrchestrationPlan(t, c.Semantics)
	last := c.Semantics.Instructions[len(c.Semantics.Instructions)-1]
	if last.Kind != instructionInvoke || last.Target.Record.ID != "User.Walk" || len(last.Arguments) != 1 || last.Arguments[0].Record.ID != "Store" {
		t.Fatal("subject, inflected member, or prepositional argument changed binding")
	}
}

func TestFunctionReturnUsesBoundMember(t *testing.T) {
	c := executableTestResult(t, memberResolutionSource{main: identityFunction}).Compilation
	for _, function := range c.Semantics.Functions {
		if function.Template == nil || function.Return == nil || function.ReturnTemplate == nil {
			t.Fatal("Function lost its source definitions")
		}
		if len(function.Body) != 1 || function.Body[0].Kind != instructionReturn || function.Body[0].SourceToken.binding.Record != function.Return {
			t.Fatal("return effect did not use the bound Function member")
		}
	}
	if _, ok := c.Gram.lex.Verbs["returning"]; ok {
		t.Fatal("parallel Return root remains")
	}
	if _, ok := c.Gram.lex.Verbs["invocative"]; ok {
		t.Fatal("Call root remains")
	}
}

func TestFunctionWordsComposeWithActions(t *testing.T) {
	var walk = functionSource("Walk|-s|-ed|-ing|-er", "\"Location\" as Text", "", "Show Location.")
	for name, source := range map[string]string{
		"generic":     walk + "Create \"Travel|-s|-ed|-ing|-er\" for Walking any Text.\nTravel `store`.\n",
		"coordinated": walk + "Walk `store` and then Show `after`.\n",
		"conditional": walk + "If true, Walk `store`, else Walk `park`.\n",
	} {
		t.Run(name, func(t *testing.T) {
			c := executableTestResult(t, memberResolutionSource{main: source}).Compilation
			requireOrchestrationPlan(t, c.Semantics)
			last := c.Semantics.Instructions[len(c.Semantics.Instructions)-1]
			switch name {
			case "generic":
				if last.Kind != instructionCall || len(last.Actions) != 1 || last.Actions[0].Kind != instructionInvoke || last.Actions[0].Target.Record.ID != "Walk" {
					t.Fatal("composed word lost its Function invocation")
				}
			case "coordinated":
				if last.Kind != instructionShow || c.Semantics.Instructions[len(c.Semantics.Instructions)-2].Kind != instructionInvoke {
					t.Fatal("Function invocation discarded or reordered its coordinated effect")
				}
			case "conditional":
				if last.Kind != instructionIf || len(last.Then) != 1 || len(last.Else) != 1 || last.Then[0].Kind != instructionInvoke || last.Else[0].Kind != instructionInvoke {
					t.Fatal("Function invocation escaped its branch")
				}
			}
		})
	}
}

func TestFunctionReturnAliasesKeepSelectedDefinition(t *testing.T) {
	for name, body := range map[string]string{
		"direct alias":  "Use Return as \"Finish\".\nFinish Value.",
		"chained alias": "Use Return as \"Finish\".\nUse Finish as \"Complete\".\nComplete Value.",
		"member alias":  "Use Function's Return as \"Finish\".\nFinish Value.",
	} {
		t.Run(name, func(t *testing.T) {
			source := strings.Replace(identityFunction, "Return Value.", body, 1) + "Identity `done` as \"Result\".\nShow Result.\n"
			c := executableTestResult(t, memberResolutionSource{main: source}).Compilation
			for _, function := range c.Semantics.Functions {
				if len(function.Body) != 1 || function.Body[0].Kind != instructionReturn || function.Body[0].Function != function {
					t.Fatal("aliased return lost its Function binding")
				}
				bound, err := canonicalRecord(function.Body[0].SourceToken.binding.Record, c.Semantics.Aliases)
				if err != nil || bound != function.Return || function.ReturnTemplate != function.Members["Return"] {
					t.Fatalf("aliased return changed selected definition: %v", err)
				}
			}
			requireFunctionGenerated(t, source)
		})
	}
}

func TestFunctionImportedReturnAlias(t *testing.T) {
	main := "Use Finish from `local-words.l` as \"Complete\".\n\n" +
		functionSource("Identity", "\"Value\" as Text", "Text", "Complete Value.") + "Identity `done` as \"Result\".\nShow Result.\n"
	c := executableTestResult(t, memberResolutionSource{main: main, library: "Use Function's Return as \"Finish\".\n"}).Compilation
	result := &CompilationResult{Compilation: c}
	if _, err := result.OrchestratorUnit("entry", executableTestCompanions(result)); err != nil {
		t.Fatal(err)
	}
}

func TestFunctionConstructorAliasKeepsCallbackContract(t *testing.T) {
	const source = "Use Function as \"Callable\".\n\n" +
		"    Publicly create \"Identity\" as a New Callable with Inputs of \"Value\" as Text and Outputs of Text.\nReturn Value.\n\n" +
		"    Publicly create \"Apply\" as a New Callable with Inputs of \"Callback\" as Identity and Outputs of Text.\nCallback `done` as \"Result\".\nReturn Result.\n\n" +
		"Apply Identity as \"Selected\".\nShow Selected.\n"
	requireFunctionGenerated(t, source)
}

func TestFunctionEmptyCallbackKeepsBoundDefinition(t *testing.T) {
	var source = functionSource("Done", "", "", "Show `done`.") +
		functionSource("Apply", "\"Callback\" as Function", "", "Callback.") + "Apply Done.\n"
	requireFunctionGenerated(t, source)
}

func TestFunctionMemberAndComposedWordsNative(t *testing.T) {
	const source = "Create \"User\" as a Dictionary.\nDefine \"Store\" as `store`.\n\n" +
		"    Publicly create User's \"Walk|-s|-ed|-ing|-er\" as a New Function with Inputs of \"Location\" as Text and No Outputs.\nUse Return as \"Finish\".\nShow Location.\nFinish.\nShow `unreachable`.\n\n" +
		"The User walks to the Store.\nUse User's Walk as \"Walk\".\nCreate \"Travel|-s|-ed|-ing|-er\" for Walking any Text.\nTravel `park` and then Show `after`.\nIf false, Walk `wrong`, else Walk `home`.\n"
	runtimeBundleResult(t, runtimeValuesBundle(t, source), "storeparkafterhome", "")
}
