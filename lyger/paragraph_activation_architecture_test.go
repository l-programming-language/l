package Lyger

import (
	nativetest "github.com/l-programming-language/l/internal/testsupport/native"

	"path/filepath"
	"reflect"

	"strings"
	"testing"

	Logiceme "github.com/l-programming-language/l/logiceme"
)

// Inspect actual generated Rust, without materializing or executing the returned
// bundle. Generated variable names and the frontend's scope representation are
// deliberately not part of these assertions.
func paragraphActivationTrace(t *testing.T, result *CompilationResult) []string {
	t.Helper()
	unit, err := result.OrchestratorUnit("entry", executableTestCompanions(result))
	if err != nil {
		t.Fatal(err)
	}
	bundle, err := Logiceme.NewLogicemeService(phase1Options(t).Resources).Evaluate(unit.BackendAML, Logiceme.GenerationOptions{IncludeRuntime: true})
	recordBundle(t, bundle)
	if err != nil {
		t.Fatal(err)
	}
	file := observeBundleMain(t, bundle)
	body := file.function(t, "l_main").Body
	apps, stored := map[string]string{}, map[string]string{}
	var trace []string
	// Index assignments in source order. The source observer does not consult IR.
	calls := nativetest.RustCalls(body)
	at := map[int]nativetest.RustCall{}
	for _, call := range calls {
		at[call.At] = call
	}
	for i, tok := range body {
		if tok.Text == "=" && i > 0 {
			name := body[i-1].Text
			// A typed let owns the identifier before its type, not the final type token.
			for left := i - 1; left >= 0; left-- {
				if body[left].Text == ";" || body[left].Text == "{" || body[left].Text == "}" {
					break
				}
				if body[left].Text == "let" {
					binding := left + 1
					if body[binding].Text == "mut" {
						binding++
					}
					name = body[binding].Text
					break
				}
			}
			j := i + 1
			for j < len(body) && body[j].Text != ";" {
				j++
			}
			rhs := body[i+1 : j]
			for _, call := range nativetest.RustCalls(rhs) {
				switch call.Name {
				case "app":
					if len(call.Args) != 1 || len(call.Args[0]) != 1 {
						t.Fatal("companion path lost literal")
					}
					path, e := nativetest.DecodeRustLiteral(call.Args[0][0].Text)
					if e != nil {
						t.Fatal(e)
					}
					apps[name] = filepath.Base(path)
				case "output":
					stored[name] = "saved:" + apps[call.Receiver]
				}
			}
			if len(rhs) > 0 {
				if value, e := nativetest.DecodeRustLiteral(rhs[0].Text); e == nil {
					stored[name] = value
				}
			}
		}
		if tok.Text == "if" {
			j := i + 1
			for j < len(body) && body[j].Text != "{" {
				j++
			}
			for _, c := range nativetest.RustCalls(body[i+1 : j]) {
				if c.Name == "error" {
					trace = append(trace, "guard:"+apps[c.Receiver])
				}
			}
		}
		if tok.Text == "return" {
			trace = append(trace, "exit")
		}
		call, ok := at[i]
		if !ok {
			continue
		}
		switch call.Name {
		case "build", "run", "wait", "output":
			name := call.Name
			if name == "wait" {
				name = "await"
			}
			trace = append(trace, name+":"+apps[call.Receiver])
		case "fail":
			trace = append(trace, "fail")
		case "show":
			if len(call.Args) != 1 {
				t.Fatal("show arity")
			}
			arg := call.Args[0]
			for len(arg) > 0 && arg[0].Text == "&" {
				arg = arg[1:]
			}
			if len(arg) > 0 && (len(arg) == 1 || nativetest.RustTokensText(arg[1:]) == ".as_bytes().to_vec()" || nativetest.RustTokensText(arg[1:]) == ".to_vec()") {
				if value, e := nativetest.DecodeRustLiteral(arg[0].Text); e == nil {
					trace = append(trace, "show:"+value)
				} else {
					trace = append(trace, "show:"+stored[arg[0].Text])
				}
			} else {
				trace = append(trace, "show:expression")
			}
		}
	}
	return trace
}

func TestParagraphActivationConditionalKeepsGuardAndRHSOwners(t *testing.T) {
	source := paragraphSource(t)
	source.main = paragraphGlobals +
		"\n    Create \"First\" as a New Public Basic App.\nThe App Prints x1.\n" +
		"Create \"Second\" as a New Public Basic App.\nThe App Prints x2.\n" +
		"If Second has an Error, exit with Second's Error, else define \"firstResult\" as First's Output.\n" +
		"Show firstResult.\n\nRun First.\nAwait First.\nRun Second.\nAwait Second.\n"
	result := executableTestResult(t, source)
	actual := paragraphActivationTrace(t, result)

	want := []string{"run:First", "await:First", "run:Second", "await:Second", "guard:Second", "fail", "exit", "output:First", "show:saved:First"}
	if !reflect.DeepEqual(actual, want) {
		t.Fatalf("guard-selected continuation = %v; want %v", actual, want)
	}
}

func TestParagraphActivationRunCalleeDoesNotOwnCallerContinuation(t *testing.T) {
	source := paragraphSource(t)
	source.main = paragraphGlobals +
		"\n    Create \"Other\" as a New Public Basic App.\nThe App Prints x2.\nShow `other continuation`.\n" +
		"\n    Create \"Worker\" as a New Public Basic App.\nThe App Prints x1.\nRun Other.\nAwait Other.\nShow `worker continuation`.\n" +
		"\nRun Worker.\nAwait Worker.\n"
	result := executableTestResult(t, source)
	actual := paragraphActivationTrace(t, result)
	want := []string{"run:Worker", "await:Worker", "run:Other", "await:Other", "show:other continuation", "show:worker continuation"}
	if !reflect.DeepEqual(actual, want) {
		t.Fatalf("nested Run changed continuation affiliation = %v; want %v", actual, want)
	}
}

func TestParagraphActivationCoordinatedDeclarationEffectOwner(t *testing.T) {
	for _, tc := range []struct {
		name, tail, following string
		want                  []string
	}{
		{
			name:      "Show uses the newly declared App",
			tail:      "Show `second continuation`",
			following: "Show `second tail`.\n",
			want:      []string{"run:First", "await:First", "run:Second", "await:Second", "show:second continuation", "show:second tail"},
		},
		{
			name: "bare Exit uses the newly declared App",
			tail: "Exit",
			want: []string{"run:First", "await:First", "run:Second", "await:Second", "exit"},
		},
		{
			name:      "explicit RHS owner remains authoritative",
			tail:      `Define "firstResult" as First's Output`,
			following: "Show firstResult.\n",
			want:      []string{"run:First", "await:First", "output:First", "show:saved:First", "run:Second", "await:Second"},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			source := paragraphSource(t)
			source.main = paragraphGlobals +
				"\n    Create \"First\" as a New Public Basic App.\nThe App Prints x1.\n" +
				"Create \"Second\" as a New Public Basic App and " + tc.tail + ".\nSecond Prints x2.\n" + tc.following +
				"\nRun First.\nAwait First.\nRun Second.\nAwait Second.\n"
			result := executableTestResult(t, source)
			actual := paragraphActivationTrace(t, result)
			if !reflect.DeepEqual(actual, tc.want) {
				t.Fatalf("coordinated declaration effect = %v; want %v", actual, tc.want)
			}
		})
	}
}

func TestParagraphActivationImportedChildRetainsParentContext(t *testing.T) {
	for _, tc := range []struct {
		name                  string
		appLed, childIsScoped bool
	}{
		{"ordinary parent and scoped child", false, true},
		{"App-led parent and inherited child", true, false},
		{"App-led parent and scoped child", true, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			source := paragraphSource(t)
			source.main = "Use IO's Print from `io` as \"Print\".\n\"message\" is `parent`.\n"
			if tc.appLed {
				source.main += "\n    Create \"Worker\" as a New Public Basic App.\nThe App Prints message.\n"
			} else {
				source.main += "Create \"Worker\" as a New Basic App.\nThe App Prints message.\n\n    Run Worker.\nAwait Worker.\n"
			}
			source.main += "Define \"snapshot\" as Worker's Output.\nUse `local-words.l`.\nShow snapshot.\n"
			if tc.appLed {
				source.main += "\nRun Worker.\nAwait Worker.\n"
			}
			source.library = "\"localMarker\" is `child marker`.\nShow snapshot.\nShow localMarker.\n"
			if tc.childIsScoped {
				source.library = "    " + source.library
			}
			result := executableTestResult(t, source)
			actual := paragraphActivationTrace(t, result)
			want := []string{"run:Worker", "await:Worker", "output:Worker", "show:saved:Worker", "show:child marker", "show:saved:Worker"}
			if !reflect.DeepEqual(actual, want) {
				t.Fatalf("import changed parent activation lifetime or saved value = %v; want %v", actual, want)
			}
		})
	}
}

func TestParagraphActivationFirstRunExitRetainsNestedInvocation(t *testing.T) {
	source := paragraphSource(t)
	source.main = paragraphGlobals +
		"Create \"Launch|-es|-ed|-ing\" for Building any App and Running it and then Awaiting it.\n" +
		"Create \"Dispatch|-es|-ed|-ing\" for Launching any App and then Showing `pending action`.\n" +
		"Create \"Second\" as a New Basic App.\nThe App Prints x2.\n" +
		"\n    Create \"First\" as a New Public Basic App.\nThe App Prints x1.\nExit.\n" +
		"\nDispatch First.\nRun Second.\nShow `pending outer`.\n"
	result := executableTestResult(t, source)
	var invocation *lStatement
	for _, statement := range result.Compilation.Statements {
		if strings.TrimSpace(string(statement.source.Raw)) == "Dispatch First." {
			invocation = statement
		}
	}
	if invocation == nil {
		t.Fatal("outer source invocation disappeared")
	}
	var operations int
	var inspect func([]*instruction)
	inspect = func(instructions []*instruction) {
		for _, current := range instructions {
			if current.Kind == instructionRun && current.CallSite != nil {
				operations++
				if current.CallSite != invocation {
					t.Errorf("nested Run diagnostics point to %s:%d, want outer invocation at line %d", current.CallSite.source.Span.Path, current.CallSite.source.Span.Start.Line, invocation.source.Span.Start.Line)
				}
				if current.Statement == nil || !strings.HasPrefix(strings.TrimSpace(string(current.Statement.source.Raw)), "Create \"Launch") {
					t.Error("nested Run lost its own definition-site attribution")
				}
			}
			inspect(current.Actions)
			inspect(current.Then)
			inspect(current.Else)
			if current.Block != nil {
				inspect(current.Block.Body)
			}
		}
	}
	inspect(result.Compilation.Semantics.Instructions)
	if operations != 1 {
		t.Fatalf("nested Run operations = %d; want one source-expanded Run", operations)
	}
	actual := paragraphActivationTrace(t, result)

	for index, operation := range actual {
		if operation == "exit" {
			actual = actual[:index+1]
			break
		}
	}
	want := []string{"build:First", "run:First", "await:First", "exit"}
	if !reflect.DeepEqual(actual, want) {
		t.Fatalf("first Await did not terminate pending action and outer effects = %v; want %v", actual, want)
	}
}

func TestParagraphActivationRetainsEscapingPublicRuntimeCapture(t *testing.T) {
	source := paragraphSource(t)
	const declaration = "Publicly create \"Report|-s|-ed|-ing\" for Showing privateText and then Showing any Text."
	const invocation = "Report `caller`."
	makeSource := func(value string) string {
		return paragraphGlobals + "\n    Create \"Witness\" as a New Public Basic App.\nThe App Prints x1.\n" +
			"Define \"privateText\" as " + value + ".\n" + declaration + "\n\nRun Witness.\nAwait Witness.\n" + invocation + "\n"
	}
	source.main = makeSource("`captured at definition`")
	control := executableTestResult(t, source)
	if actual := paragraphActivationTrace(t, control); !reflect.DeepEqual(actual, []string{"run:Witness", "await:Witness", "show:captured at definition", "show:caller"}) {
		t.Fatalf("valid static definition-site capture = %v", actual)
	}
	source.main = makeSource("Witness's Output")
	runtimeResult := executableTestResult(t, source)
	plan := requireOrchestrationPlan(t, runtimeResult.Compilation.Semantics)
	var boundInContinuation bool
	var invocationPlan *orchestrationInstruction
	visitCapturePlan(plan.Root.Body, func(current *orchestrationInstruction) {
		if current.Continuation != nil {
			for _, effect := range current.Continuation.Body {
				boundInContinuation = boundInContinuation || effect.Source.Kind == instructionBindAction
			}
		}
		if current.Source.Kind == instructionCall && current.Source.Callable == runtimeResult.Compilation.Memory["Report"] {
			invocationPlan = current
		}
	})
	if !boundInContinuation || invocationPlan == nil || invocationPlan.Actions[0].Value.Storage.Capture == nil {
		t.Fatal("Await did not create and preserve its public action environment")
	}
}
