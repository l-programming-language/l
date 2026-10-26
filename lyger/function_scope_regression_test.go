package Lyger

import (
	"strings"
	"testing"
)

var functionScopeCases = []struct {
	name, main, library, output string
}{
	{"imported effects", functionSource("Act", "\"Value\" as Text", "", "Use `local-words.l`.\nReturn.") + "Show `before`.\nAct `value`.\nShow `after`.\n", "    Show `inner`.\n", "beforeinnerafter"},
	{"imported locals and return", functionSource("Echo", "\"Value\" as Text", "Text", "Use `local-words.l`.\nShow `unreachable`.") + "Show `before`.\nEcho `one` as \"First\".\nShow First.\nEcho `two` as \"Second\".\nShow Second.\nShow `after`.\n", "    Define \"Local\" as Value.\nShow Local.\nReturn Local.\n", "beforeoneonetwotwoafter"},
	{"imported Function header", functionSource("Outer", "\"Value\" as Text", "Text", "Use `local-words.l`.\nInner as \"Result\".\nReturn Result.") + "Outer `one` as \"First\".\nOuter `two` as \"Second\".\nShow First and Second.\n", "    Publicly create \"Inner\" as a New Function with No Inputs and Outputs of Text.\nReturn Value.\n", "onetwo"},
	{"Function continuation producers", distributionIOImport + "\n" + functionSource("Owner", "", "", "Use `local-words.l`.\nExecute First.\nReturn.") + "    Create \"Worker\" as a New Public Basic App.\nShow Message.\n\nOwner.\nAwait First.\nExecute Worker.\n", "    Create \"First\" as a New Public Basic App.\nThe App Prints `kept`.\nPublicly define \"Message\" as First's Output.\n", "kept\n"},
	{"composed return", strings.Replace(identityFunction, "Return Value.", "Create \"Finish|-es|-ed|-ing\" for Returning any Text.\nCreate \"Complete|-s|-d|--ing\" for Finishing any Text.\nIf true, Complete Value, else Complete `other`.\nShow `unreachable`.", 1) + "Identity `kept` as \"Result\".\nShow Result.\n", "", "kept"},
}

func TestFunctionNestedScopeOwnership(t *testing.T) {
	for _, test := range functionScopeCases {
		t.Run(test.name, func(t *testing.T) {
			c := compileMemberResolutionSource(t, memberResolutionSource{main: test.main, library: test.library})
			if c.Error != nil {
				t.Fatal(c.Error)
			}
			if _, err := planOrchestration(c.Semantics); err != nil {
				t.Fatal(err)
			}
			if test.library != "" {
				var child bool
				for _, function := range c.Semantics.Functions {
					for _, continuation := range c.Semantics.Continuations {
						child = child || continuation.Scope.parent == function.Scope
					}
					for _, current := range function.Body {
						child = child || current.Block != nil && current.Block.Scope.parent == function.Scope
					}
				}
				if !child {
					t.Fatal("imported paragraph lost its enclosing Function activation")
				}
			}
		})
	}
}

func TestFunctionImportedScopeRejectsEscapes(t *testing.T) {
	for _, test := range []struct{ body, library, diagnostic string }{
		{"Use `local-words.l`.\nReturn Local.", "    Define \"Local\" as Value.\n", "unresolved"},
		{"Use `local-words.l`.", "    If true, Return Value.\n", "does not return"},
	} {
		c := compileMemberResolutionSource(t, memberResolutionSource{main: strings.Replace(identityFunction, "Return Value.", test.body, 1), library: test.library})
		err := c.Error
		if err == nil {
			_, err = planOrchestration(c.Semantics)
		}
		if err == nil || !strings.Contains(err.Error(), test.diagnostic) {
			t.Fatalf("wanted %q, got %v", test.diagnostic, err)
		}
	}
}

func TestFunctionNestedScopesNative(t *testing.T) {
	for _, test := range functionScopeCases {
		t.Run(test.name, func(t *testing.T) {
			captureNativeResult(t, compileMemberResolutionSource(t, memberResolutionSource{main: test.main, library: test.library}), test.output)
		})
	}
}
