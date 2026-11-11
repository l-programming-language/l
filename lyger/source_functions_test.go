package Lyger

import (
	"fmt"

	"strings"
	"testing"
)

func compileFunctionPlan(t *testing.T, source string) (*orchestrationPlan, error) {
	t.Helper()
	c := compileMemberResolutionSource(t, memberResolutionSource{main: source})
	if c.Error != nil {
		return nil, c.Error
	}
	return planOrchestration(c.Semantics)
}

var identityFunction = functionSource("Echo|-es|-ed|-ing|-er", "\"Value\" as Text", "Text", "Return Value.")

func TestOrdinaryFunctionSourceCases(t *testing.T) {
	tests := []struct{ name, source, want string }{
		{"multiple results", functionSource("Pair|-s|-ed|-ing|-er", "\"First\" as Text and \"Second\" as Number", "Text and Number", "Return First and Second.") + "Pair `hello` and 7 as \"TextResult\" and \"NumberResult\".\nShow TextResult.\nShow NumberResult.\n", ""},
		{"nested call", identityFunction + functionSource("Wrap|-s|-ped|-ping|-per", "\"Message\" as Text", "Text", "Echo Message as \"Nested\".\nReturn Nested.") + "Wrap `hello` as \"Result\".\nShow Result.\n", ""},
		{"function argument", identityFunction + functionSource("Apply|--ies|--ied|-ing|--ier", "\"Replay|-s|-ed|-ing|-er\" as Echo and \"Message\" as Text", "Text", "Replay Message as \"Nested\".\nReturn Nested.") + "Apply Echo and `hello` as \"Result\".\nShow Result.\n", ""},
		{"Function signature alias", identityFunction + "Use Echo as \"Reflect|-s|-ed|-ing|-or\".\n\n" + functionSource("Apply|--ies|--ied|-ing|--ier", "\"Replay|-s|-ed|-ing|-er\" as Reflect and \"Message\" as Text", "Text", "Replay Message as \"Nested\".\nReturn Nested.") + "Apply Echo and `hello` as \"Result\".\nShow Result.\n", ""},
		{"function result", identityFunction + functionSource("Select|-s|-ed|-ing|-or", "", "Echo", "Return Echo.") + "Select as \"Choose|-s|--en|--ing|-r\".\nChoose `hello` as \"Result\".\nShow Result.\n", ""},
		{"recursive activation", functionSource("Repeat|-s|-ed|-ing|-er", "\"Ready\" as Truth and \"Message\" as Text", "Text", "If Ready, Return Message.\nRepeat true and Message as \"Nested\".\nReturn Nested.") + "Repeat false and `hello` as \"Result\".\nShow Result.\n", ""},
		{"conditional missing return", strings.Replace(identityFunction, "Return Value.", "If true, Return Value.", 1), "does not return"},
		{"conditional returns", strings.Replace(identityFunction, "Return Value.", "If true, Return Value, else Return `other`.", 1) + "Echo `hello` as \"Result\".\n", ""},
		{"captured value", "Define \"Prefix\" as `before`.\n\n" + functionSource("Read|-s|-|-ing|-er", "", "Text", "Return Prefix.") + "Read as \"Result\".\nShow Result.\n", ""},
		{"typed List argument", functionSource("Keep|-s|--pt|-ing|-er", "\"Items\" as List of Text", "List of Text", "Return Items."), ""},
		{"Function alias", identityFunction + "Use Echo as \"Replay|-s|-ed|-ing|-er\".\nReplay `hello` as \"Result\".\nShow Result.\n", ""},
		{"missing return", strings.Replace(identityFunction, "Return Value.", "Show Value.", 1), "does not return"},
		{"wrong return", strings.Replace(identityFunction, "Return Value.", "Return 7.", 1), "requires Text"},
		{"wrong argument", identityFunction + "Echo 7 as \"Result\".\n", "requires Text"},
		{"missing result", identityFunction + "Echo `hello`.\n", "result names"},
		{"return outside", "Return 7.\n", "unresolved"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, err := compileFunctionPlan(t, test.source)
			if test.want == "" {
				if err != nil {
					t.Fatal(err)
				}
			} else if err == nil || !strings.Contains(err.Error(), test.want) {
				t.Fatalf("wanted %q got %v", test.want, err)
			}
		})
	}
}

func TestOrdinaryFunctionAliasKeepsSelectedActivation(t *testing.T) {
	c := executableTestResult(t, memberResolutionSource{main: identityFunction + "Use Echo as \"Replay|-s|-ed|-ing|-er\".\nReplay `hello` as \"Result\".\n"}).Compilation
	var declaration, alias, call *instruction
	if err := c.Semantics.walkInstructions(func(current *instruction) error {
		switch current.Kind {
		case instructionFunction:
			declaration = current
		case instructionAssign:
			if current.Target.Record.runtimeFunctionAlias {
				alias = current
			}
		case instructionInvoke:
			call = current
		}
		return nil
	}); err != nil {
		t.Fatal(err)
	}
	if declaration == nil || alias == nil || call == nil {
		t.Fatal("source did not produce Function binding, alias and call")
	}
	c.Semantics.Instructions = []*instruction{declaration, alias, declaration, call}
	plan := requireOrchestrationPlan(t, c.Semantics)
	body := plan.Root.Body
	if body[0].Target.Storage != body[2].Target.Storage || body[1].Target.Storage == body[0].Target.Storage || body[3].Target.Storage != body[1].Target.Storage {
		t.Fatal("saved alias followed the declaration's replacement activation")
	}
	if body[0].Function.Body == body[2].Function.Body || body[0].Function.Inputs[0] == body[2].Function.Inputs[0] {
		t.Fatal("Function activations shared parameter storage")
	}
}

// Shared declaration framing preserves the authored signature and body bytes.
func functionSource(name, inputs, outputs, body string) string {
	if inputs == "" {
		inputs = "No Inputs"
	} else {
		inputs = "Inputs of " + inputs
	}
	if outputs == "" {
		outputs = "No Outputs"
	} else {
		outputs = "Outputs of " + outputs
	}
	return fmt.Sprintf("    Publicly create %q as a New Function with %s and %s.\n%s\n\n", name, inputs, outputs, body)
}

func requireFunctionPlan(t *testing.T, source string) *orchestrationPlan {
	t.Helper()
	plan, err := compileFunctionPlan(t, source)
	if err != nil {
		t.Fatal(err)
	}
	return plan
}
