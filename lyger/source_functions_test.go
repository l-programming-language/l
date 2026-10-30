package Lyger

import (
	"errors"
	"fmt"
	nativetest "github.com/l-programming-language/l/internal/testsupport/native"
	Logiceme "github.com/l-programming-language/l/logiceme"
	"io/fs"
	"os/exec"
	"strings"
	"testing"
)

func functionGenerated(t *testing.T, source string) (string, error) {
	t.Helper()
	c := compileMemberResolutionSource(t, memberResolutionSource{main: source})
	if c.Error != nil {
		return "", c.Error
	}
	_, bundle, err := orchestratorTestBundle(canonicalTestOptions().Resources, &CompilationResult{Compilation: c})
	return builtinProcessRust(bundle), err
}

var identityFunction = functionSource("Identity", "\"Value\" as Text", "Text", "Return Value.")

func TestOrdinaryFunctionSourceCases(t *testing.T) {
	tests := []struct{ name, source, want string }{
		{"multiple results", functionSource("Pair", "\"First\" as Text and \"Second\" as Number", "Text and Number", "Return First and Second.") + "Pair `hello` and 7 as \"TextResult\" and \"NumberResult\".\nShow TextResult.\nShow NumberResult.\n", ""},
		{"nested call", identityFunction + functionSource("Wrapper", "\"Message\" as Text", "Text", "Identity Message as \"Nested\".\nReturn Nested.") + "Wrapper `hello` as \"Result\".\nShow Result.\n", ""},
		{"function argument", identityFunction + functionSource("Apply", "\"Callback\" as Identity and \"Message\" as Text", "Text", "Callback Message as \"Nested\".\nReturn Nested.") + "Apply Identity and `hello` as \"Result\".\nShow Result.\n", ""},
		{"Function signature alias", identityFunction + "Use Identity as \"Signature\".\n\n" + functionSource("Apply", "\"Callback\" as Signature and \"Message\" as Text", "Text", "Callback Message as \"Nested\".\nReturn Nested.") + "Apply Identity and `hello` as \"Result\".\nShow Result.\n", ""},
		{"function result", identityFunction + functionSource("Select", "", "Identity", "Return Identity.") + "Select as \"Chosen\".\nChosen `hello` as \"Result\".\nShow Result.\n", ""},
		{"recursive activation", functionSource("Again", "\"Ready\" as Truth and \"Message\" as Text", "Text", "If Ready, Return Message.\nAgain true and Message as \"Nested\".\nReturn Nested.") + "Again false and `hello` as \"Result\".\nShow Result.\n", ""},
		{"conditional missing return", strings.Replace(identityFunction, "Return Value.", "If true, Return Value.", 1), "does not return"},
		{"conditional returns", strings.Replace(identityFunction, "Return Value.", "If true, Return Value, else Return `other`.", 1) + "Identity `hello` as \"Result\".\n", ""},
		{"captured value", "Define \"Prefix\" as `before`.\n\n" + functionSource("Read", "", "Text", "Return Prefix.") + "Read as \"Result\".\nShow Result.\n", ""},
		{"typed List argument", functionSource("Keep", "\"Items\" as List of Text", "List of Text", "Return Items."), ""},
		{"Function alias", identityFunction + "Use Identity as \"Saved\".\nSaved `hello` as \"Result\".\nShow Result.\n", ""},
		{"missing return", strings.Replace(identityFunction, "Return Value.", "Show Value.", 1), "does not return"},
		{"wrong return", strings.Replace(identityFunction, "Return Value.", "Return 7.", 1), "requires Text"},
		{"wrong argument", identityFunction + "Identity 7 as \"Result\".\n", "requires Text"},
		{"missing result", identityFunction + "Identity `hello`.\n", "result names"},
		{"return outside", "Return 7.\n", "unresolved"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, err := functionGenerated(t, test.source)
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
	c := executableTestResult(t, memberResolutionSource{main: identityFunction + "Use Identity as \"Saved\".\nSaved `hello` as \"Result\".\n"}).Compilation
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

func TestOrdinaryFunctionsNative(t *testing.T) {
	source := functionSource("Done", "", "Text", "If false, Exit with an Error of `unused`, else Exit.") + identityFunction +
		functionSource("Pair", "\"First\" as Text and \"Second\" as Number", "Text and Number", "Return First and Second.") +
		functionSource("Apply", "\"Callback\" as Identity and \"Message\" as Text", "Text", "Callback Message as \"Nested\".\nReturn Nested.") +
		functionSource("Select", "", "Identity", "Return Identity.") +
		functionSource("Again", "\"Ready\" as Truth and \"Message\" as Text", "Text", "If Ready, Return Message.\nAgain true and Message as \"Nested\".\nReturn Nested.") +
		"Define \"Prefix\" as `capture`.\n\n" + functionSource("Read", "", "Text", "Return Prefix.") +
		"Identity `one` as \"One\".\nIdentity `two` as \"Two\".\nPair `pair` and 7 as \"Label\" and \"Count\".\nApply Identity and `callback` as \"Applied\".\nSelect as \"Chosen\".\nChosen `selected` as \"Selected\".\nAgain false and `recursive` as \"Repeated\".\nRead as \"Captured\".\nUse Identity as \"Saved\".\nSaved `alias` as \"Aliased\".\nShow One and Two and Label and Count and Applied and Selected and Repeated and Captured and Aliased.\nDone as \"NeverRead\".\nShow `must not continue`.\n"
	captureNativeResult(t, executableTestResult(t, memberResolutionSource{main: source}).Compilation, "onetwopair7callbackselectedrecursivecapturealias")
}

func TestOrdinaryFunctionErrorNative(t *testing.T) {
	source := functionSource("RejectCall", "", "", "Exit with an Error of `failed`.") + functionSource("Relay", "", "", "RejectCall.\nShow `unexpected-function`.") + "Relay.\nShow `unexpected-root`.\n"
	bundle := runtimeValuesBundle(t, source)
	recordBundle(t, bundle)
	binary := buildActualBundle(t, bundle, "entry", true, "")
	stdout, stderr, err := nativetest.RuntimeExecution.Run(t, binary, "", nil)
	var exit *exec.ExitError
	if !errors.As(err, &exit) || exit.ExitCode() != 1 || len(stdout) != 0 || string(stderr) != "failed\n" {
		t.Fatalf("Function error stdout=%q stderr=%q error=%v", stdout, stderr, err)
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

func requireFunctionGenerated(t *testing.T, source string) string {
	t.Helper()
	generated, err := functionGenerated(t, source)
	if err != nil {
		t.Fatal(err)
	}
	return generated
}

func orchestratorTestBundle(resources fs.FS, result *CompilationResult) (BuildUnit, Logiceme.Executable, error) {
	unit, err := result.OrchestratorUnit("entry", executableTestCompanions(result))
	if err != nil {
		return unit, Logiceme.Executable{}, err
	}
	bundle, err := Logiceme.NewLogicemeService(resources).Evaluate(unit.BackendAML, Logiceme.GenerationOptions{IncludeRuntime: true})
	return unit, bundle, err
}
