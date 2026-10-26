package Lyger

import (
	"errors"
	nativetest "github.com/l-programming-language/l/internal/testsupport/native"
	Logiceme "github.com/l-programming-language/l/logiceme"
	"os/exec"
	"regexp"
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
	source, err := functionGenerated(t, identityFunction+"Identity `hello` as \"Result\".\nShow Result.\nIdentity `again` as \"Second\".\nShow Second.\n")
	if err != nil {
		t.Fatal(err)
	}
	for _, needle := range []string{"fn lFunction", "std::ops::ControlFlow<(), (Vec<u8>,)>", "async move", ".await"} {
		if !strings.Contains(source, needle) {
			t.Fatalf("missing canonical native Function structure %q", needle)
		}
	}
	if !regexp.MustCompile(`for<'[A-Za-z][A-Za-z0-9_]*> fn\(`).MatchString(source) {
		t.Fatal("native Function pointer lost its borrowed context lifetime")
	}
	if len(regexp.MustCompile(`\(lCalled[0-9]+\.1\)\(`).FindAllString(source, -1)) != 2 {
		t.Fatal("both source invocations must use the native typed function pointer")
	}
	for _, removed := range []string{"struct LCallable", "LFunctionOutcome", ".call(runtime, lProgram, vec![", "lArguments: Vec<LValue>"} {
		if strings.Contains(source, removed) {
			t.Fatalf("parallel Function implementation remains: %s", removed)
		}
	}
}

func TestOrdinaryFunctionSourceCases(t *testing.T) {
	tests := []struct{ name, source, want string }{
		{"multiple results", "    Publicly create \"Pair\" as a New Function with Inputs of \"First\" as Text and \"Second\" as Number and Outputs of Text and Number.\nReturn First and Second.\n\nPair `hello` and 7 as \"TextResult\" and \"NumberResult\".\nShow TextResult.\nShow NumberResult.\n", ""},
		{"nested call", identityFunction + "    Publicly create \"Wrapper\" as a New Function with Inputs of \"Message\" as Text and Outputs of Text.\nIdentity Message as \"Nested\".\nReturn Nested.\n\nWrapper `hello` as \"Result\".\nShow Result.\n", ""},
		{"function argument", identityFunction + "    Publicly create \"Apply\" as a New Function with Inputs of \"Callback\" as Identity and \"Message\" as Text and Outputs of Text.\nCallback Message as \"Nested\".\nReturn Nested.\n\nApply Identity and `hello` as \"Result\".\nShow Result.\n", ""},
		{"Function signature alias", identityFunction + "Use Identity as \"Signature\".\n\n    Publicly create \"Apply\" as a New Function with Inputs of \"Callback\" as Signature and \"Message\" as Text and Outputs of Text.\nCallback Message as \"Nested\".\nReturn Nested.\n\nApply Identity and `hello` as \"Result\".\nShow Result.\n", ""},
		{"function result", identityFunction + "    Publicly create \"Select\" as a New Function with No Inputs and Outputs of Identity.\nReturn Identity.\n\nSelect as \"Chosen\".\nChosen `hello` as \"Result\".\nShow Result.\n", ""},
		{"recursive activation", "    Publicly create \"Again\" as a New Function with Inputs of \"Ready\" as Truth and \"Message\" as Text and Outputs of Text.\nIf Ready, Return Message.\nAgain true and Message as \"Nested\".\nReturn Nested.\n\nAgain false and `hello` as \"Result\".\nShow Result.\n", ""},
		{"conditional missing return", strings.Replace(identityFunction, "Return Value.", "If true, Return Value.", 1), "does not return"},
		{"conditional returns", strings.Replace(identityFunction, "Return Value.", "If true, Return Value, else Return `other`.", 1) + "Identity `hello` as \"Result\".\n", ""},
		{"captured value", "Define \"Prefix\" as `before`.\n\n    Publicly create \"Read\" as a New Function with No Inputs and Outputs of Text.\nReturn Prefix.\n\nRead as \"Result\".\nShow Result.\n", ""},
		{"typed List argument", "    Publicly create \"Keep\" as a New Function with Inputs of \"Items\" as List of Text and Outputs of List of Text.\nReturn Items.\n\n", ""},
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
	c := compileMemberResolutionSource(t, memberResolutionSource{main: identityFunction + "Use Identity as \"Saved\".\nSaved `hello` as \"Result\".\n"})
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
		"    Publicly create \"Apply\" as a New Function with Inputs of \"Callback\" as Identity and \"Message\" as Text and Outputs of Text.\nCallback Message as \"Nested\".\nReturn Nested.\n\n" +
		"    Publicly create \"Select\" as a New Function with No Inputs and Outputs of Identity.\nReturn Identity.\n\n" +
		"    Publicly create \"Again\" as a New Function with Inputs of \"Ready\" as Truth and \"Message\" as Text and Outputs of Text.\nIf Ready, Return Message.\nAgain true and Message as \"Nested\".\nReturn Nested.\n\n" +
		"Define \"Prefix\" as `capture`.\n\n    Publicly create \"Read\" as a New Function with No Inputs and Outputs of Text.\nReturn Prefix.\n\n" +
		"Identity `one` as \"One\".\nIdentity `two` as \"Two\".\nPair `pair` and 7 as \"Label\" and \"Count\".\nApply Identity and `callback` as \"Applied\".\nSelect as \"Chosen\".\nChosen `selected` as \"Selected\".\nAgain false and `recursive` as \"Repeated\".\nRead as \"Captured\".\nUse Identity as \"Saved\".\nSaved `alias` as \"Aliased\".\nShow One and Two and Label and Count and Applied and Selected and Repeated and Captured and Aliased.\nDone as \"NeverRead\".\nShow `must not continue`.\n"
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
	source := "    Publicly create \"RejectCall\" as a New Function with No Inputs and No Outputs.\nExit with an Error of `failed`.\n\n    Publicly create \"Relay\" as a New Function with No Inputs and No Outputs.\nRejectCall.\nShow `unexpected-function`.\n\nRelay.\nShow `unexpected-root`.\n"
	bundle := runtimeValuesBundle(t, source)
	recordBundle(t, bundle)
	binary := buildActualBundle(t, bundle, "entry", true, "")
	stdout, stderr, err := nativetest.RuntimeExecution.Run(t, binary, "", nil)
	var exit *exec.ExitError
	if !errors.As(err, &exit) || exit.ExitCode() != 1 || len(stdout) != 0 || string(stderr) != "failed\n" {
		t.Fatalf("Function error stdout=%q stderr=%q error=%v", stdout, stderr, err)
	}
}
