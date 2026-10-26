package Lyger

import (
	"errors"
	nativetest "github.com/l-programming-language/l/internal/testsupport/native"
	Logiceme "github.com/l-programming-language/l/logiceme"
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
	result := &CompilationResult{Compilation: c}
	unit, err := result.OrchestratorUnit("entry", executableTestCompanions(result))
	if err != nil {
		return "", err
	}
	bundle, err := Logiceme.NewLogicemeService(canonicalTestOptions().Resources).Evaluate(unit.BackendAML, Logiceme.GenerationOptions{IncludeRuntime: true})
	if err != nil {
		return "", err
	}
	return builtinProcessRust(bundle), nil
}

const identityFunction = "    Publicly create \"Identity\" as a New Function with Inputs of \"Value\" as Text and Outputs of Text.\nReturn Value.\n\n"

func TestOrdinaryFunctionSourceGeneration(t *testing.T) {
	source, err := functionGenerated(t, identityFunction+"Call Identity with `hello` as \"Result\".\nShow Result.\nCall Identity with `again` as \"Second\".\nShow Second.\n")
	if err != nil {
		t.Fatal(err)
	}
	for _, needle := range []string{"struct LCallable", "struct LProgram", "LProgram::new(", "LType::Function(vec![LType::Text],vec![LType::Text])", ".call(runtime, lProgram, vec![", "return Ok(LFunctionOutcome::Returned(vec![LValue::Text("} {
		if !strings.Contains(source, needle) {
			t.Fatalf("missing %q in generated source", needle)
		}
	}
}

func TestOrdinaryFunctionSourceCases(t *testing.T) {
	tests := []struct{ name, source, want string }{
		{"multiple results", "    Publicly create \"Pair\" as a New Function with Inputs of \"First\" as Text and \"Second\" as Number and Outputs of Text and Number.\nReturn First and Second.\n\nCall Pair with `hello` and 7 as \"TextResult\" and \"NumberResult\".\nShow TextResult.\nShow NumberResult.\n", ""},
		{"nested call", identityFunction + "    Publicly create \"Wrapper\" as a New Function with Inputs of \"Message\" as Text and Outputs of Text.\nCall Identity with Message as \"Nested\".\nReturn Nested.\n\nCall Wrapper with `hello` as \"Result\".\nShow Result.\n", ""},
		{"function argument", identityFunction + "    Publicly create \"Apply\" as a New Function with Inputs of \"Callback\" as Identity and \"Message\" as Text and Outputs of Text.\nCall Callback with Message as \"Nested\".\nReturn Nested.\n\nCall Apply with Identity and `hello` as \"Result\".\nShow Result.\n", ""},
		{"Function signature alias", identityFunction + "Use Identity as \"Signature\".\n\n    Publicly create \"Apply\" as a New Function with Inputs of \"Callback\" as Signature and \"Message\" as Text and Outputs of Text.\nCall Callback with Message as \"Nested\".\nReturn Nested.\n\nCall Apply with Identity and `hello` as \"Result\".\nShow Result.\n", ""},
		{"function result", identityFunction + "    Publicly create \"Select\" as a New Function with No Inputs and Outputs of Identity.\nReturn Identity.\n\nCall Select as \"Chosen\".\nCall Chosen with `hello` as \"Result\".\nShow Result.\n", ""},
		{"recursive activation", "    Publicly create \"Again\" as a New Function with Inputs of \"Ready\" as Truth and \"Message\" as Text and Outputs of Text.\nIf Ready, Return Message.\nCall Again with true and Message as \"Nested\".\nReturn Nested.\n\nCall Again with false and `hello` as \"Result\".\nShow Result.\n", ""},
		{"conditional missing return", strings.Replace(identityFunction, "Return Value.", "If true, Return Value.", 1), "does not return"},
		{"conditional returns", strings.Replace(identityFunction, "Return Value.", "If true, Return Value, else Return `other`.", 1) + "Call Identity with `hello` as \"Result\".\n", ""},
		{"captured value", "Define \"Prefix\" as `before`.\n\n    Publicly create \"Read\" as a New Function with No Inputs and Outputs of Text.\nReturn Prefix.\n\nCall Read as \"Result\".\nShow Result.\n", ""},
		{"typed List argument", "    Publicly create \"Keep\" as a New Function with Inputs of \"Items\" as List of Text and Outputs of List of Text.\nReturn Items.\n\n", ""},
		{"Function alias", identityFunction + "Use Identity as \"Saved\".\nCall Saved with `hello` as \"Result\".\nShow Result.\n", ""},
		{"missing return", strings.Replace(identityFunction, "Return Value.", "Show Value.", 1), "does not return"},
		{"wrong return", strings.Replace(identityFunction, "Return Value.", "Return 7.", 1), "requires Text"},
		{"wrong argument", identityFunction + "Call Identity with 7 as \"Result\".\n", "requires Text"},
		{"missing result", identityFunction + "Call Identity with `hello`.\n", "result names"},
		{"return outside", "Return 7.\n", "Function activation"},
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
	c := compileMemberResolutionSource(t, memberResolutionSource{main: identityFunction + "Use Identity as \"Saved\".\nCall Saved with `hello` as \"Result\".\n"})
	if c.Error != nil {
		t.Fatal(c.Error)
	}
	var declaration, alias, call *instruction
	var walk func([]*instruction)
	walk = func(body []*instruction) {
		for _, current := range body {
			if current.Kind == instructionFunction {
				declaration = current
			}
			if current.Kind == instructionAssign && current.Target.Record.runtimeFunctionAlias {
				alias = current
			}
			if current.Kind == instructionInvoke {
				call = current
			}
			if current.Block != nil {
				walk(current.Block.Body)
			}
		}
	}
	walk(c.Semantics.Instructions)
	if declaration == nil || alias == nil || call == nil {
		t.Fatal("source did not produce Function binding, alias and call")
	}
	c.Semantics.Instructions = []*instruction{declaration, alias, declaration, call}
	plan, err := planOrchestration(c.Semantics)
	if err != nil {
		t.Fatal(err)
	}
	body := plan.Root.Body
	if body[0].Target.Storage != body[2].Target.Storage || body[1].Target.Storage == body[0].Target.Storage || body[3].Target.Storage != body[1].Target.Storage {
		t.Fatal("saved alias followed the declaration's replacement activation")
	}
	if body[0].Function.Body == body[2].Function.Body || body[0].Function.Inputs[0] == body[2].Function.Inputs[0] {
		t.Fatal("Function activations shared parameter storage")
	}
}

func TestOrdinaryFunctionsNative(t *testing.T) {
	source := "    Publicly create \"Done\" as a New Function with No Inputs and Outputs of Text.\nIf false, Exit with an Error of `unused`, else Exit.\n\n" + identityFunction +
		"    Publicly create \"Pair\" as a New Function with Inputs of \"First\" as Text and \"Second\" as Number and Outputs of Text and Number.\nReturn First and Second.\n\n" +
		"    Publicly create \"Apply\" as a New Function with Inputs of \"Callback\" as Identity and \"Message\" as Text and Outputs of Text.\nCall Callback with Message as \"Nested\".\nReturn Nested.\n\n" +
		"    Publicly create \"Select\" as a New Function with No Inputs and Outputs of Identity.\nReturn Identity.\n\n" +
		"    Publicly create \"Again\" as a New Function with Inputs of \"Ready\" as Truth and \"Message\" as Text and Outputs of Text.\nIf Ready, Return Message.\nCall Again with true and Message as \"Nested\".\nReturn Nested.\n\n" +
		"Define \"Prefix\" as `capture`.\n\n    Publicly create \"Read\" as a New Function with No Inputs and Outputs of Text.\nReturn Prefix.\n\n" +
		"Call Identity with `one` as \"One\".\nCall Identity with `two` as \"Two\".\nCall Pair with `pair` and 7 as \"Label\" and \"Count\".\nCall Apply with Identity and `callback` as \"Applied\".\nCall Select as \"Chosen\".\nCall Chosen with `selected` as \"Selected\".\nCall Again with false and `recursive` as \"Repeated\".\nCall Read as \"Captured\".\nUse Identity as \"Saved\".\nCall Saved with `alias` as \"Aliased\".\nShow One and Two and Label and Count and Applied and Selected and Repeated and Captured and Aliased.\nCall Done as \"NeverRead\".\nShow `must not continue`.\n"
	c := compileMemberResolutionSource(t, memberResolutionSource{main: source})
	if c.Error != nil {
		t.Fatal(c.Error)
	}
	result := &CompilationResult{Compilation: c}
	unit, err := result.OrchestratorUnit("entry", executableTestCompanions(result))
	if err != nil {
		t.Fatal(err)
	}
	bundle, err := Logiceme.NewLogicemeService(canonicalTestOptions().Resources).Evaluate(unit.BackendAML, Logiceme.GenerationOptions{IncludeRuntime: true})
	if err != nil {
		t.Fatal(err)
	}
	recordBundle(t, bundle)
	binary := buildActualBundle(t, bundle, "entry", true, "")
	stdout, stderr, err := nativetest.RuntimeExecution.Run(t, binary, "", nil)
	if err != nil || len(stderr) != 0 || string(stdout) != "onetwopair7callbackselectedrecursivecapturealias" {
		t.Fatalf("functions stdout=%q stderr=%q error=%v", stdout, stderr, err)
	}
}

func TestOrdinaryFunctionErrorNative(t *testing.T) {
	source := "    Publicly create \"RejectCall\" as a New Function with No Inputs and No Outputs.\nExit with an Error of `failed`.\n\n    Publicly create \"Relay\" as a New Function with No Inputs and No Outputs.\nCall RejectCall.\nShow `unexpected-function`.\n\nCall Relay.\nShow `unexpected-root`.\n"
	bundle := runtimeValuesBundle(t, source)
	recordBundle(t, bundle)
	binary := buildActualBundle(t, bundle, "entry", true, "")
	stdout, stderr, err := nativetest.RuntimeExecution.Run(t, binary, "", nil)
	var exit *exec.ExitError
	if !errors.As(err, &exit) || exit.ExitCode() != 1 || len(stdout) != 0 || string(stderr) != "failed\n" {
		t.Fatalf("Function error stdout=%q stderr=%q error=%v", stdout, stderr, err)
	}
}
