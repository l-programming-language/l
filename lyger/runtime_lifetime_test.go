package Lyger

import (
	"os"
	"testing"

	Logiceme "github.com/l-programming-language/l/logiceme"
)

// The fixture observes the selected production words' actual Rc targets. Its
// assertions are composed native consumers, never a replacement storage model.
func runtimeLifetimeBundle(t *testing.T) Logiceme.Executable {
	t.Helper()
	fixture, err := os.ReadFile("test/native/runtime-lifetimes.l")
	if err != nil {
		t.Fatal(err)
	}
	result := requireDistributionFrontend(t, string(fixture)+"\nCreate \"Witness\" as a New Basic App.\n", nil)
	compilation := result.Compilation
	// The native observer deliberately consumes these private storage operations.
	// Its imports are explicit; production emission has no support catalogue.
	var declarations []DefinitionInstance
	for _, name := range []string{
		"Require", "Ignore", "Value.Wrap", "Value.Restore", "Value.Expect", "Value.Format",
		"Environment.Create", "Environment.Read", "Environment.Update", "Environment.Copy", "Environment.Visit", "Environment.Render",
		"Map.Create", "Map.Read", "Map.Update", "Map.Copy", "Map.Render", "Map.Reach",
		"Pointer.Create", "Pointer.Read", "Pointer.Update", "Pointer.Copy", "Pointer.Render", "Pointer.Visit",
		"Function.Copy", "Function.Render", "Function.Reach", "Program.Create", "Program.Values.Write", "Program.Actions.Write", "Null.Render", "Text.Render",
	} {
		record, err := compilation.builtinRecord("Storage." + name)
		if err != nil {
			t.Fatal(err)
		}
		declaration, err := compilation.composeExport(record)
		if err != nil {
			t.Fatal(err)
		}
		declarations = append(declarations, declaration)
	}
	emitter := orchestrationEmitter{compilation: compilation}
	pointerCopy, err := emitter.functionReference("Storage.Pointer.Copy")
	if err != nil {
		t.Fatal(err)
	}
	declarations, err = definitionClosure(append(declarations, emitter.dependencies...)...)
	if err != nil {
		t.Fatal(err)
	}
	instantiate := func(name string) DefinitionInstance {
		t.Helper()
		binding, state, err := compilation.Scope.Root.lookupLocal([]byte(name))
		if err != nil || state != lookupFound || binding == nil {
			t.Fatalf("native observer definition %s: %v", name, err)
		}
		template, err := compilation.recordTemplate(binding.Record)
		if err != nil || template == nil {
			t.Fatalf("native observer template %s: %v", name, err)
		}
		bound, err := template.Instantiate(map[string]DefinitionArgument{"PointerCopy": pointerCopy})
		if err != nil {
			t.Fatal(err)
		}
		return bound
	}
	callback := instantiate("Lifetimecallback")
	declarations = append(declarations, callback)
	observation := instantiate("Lifetimeobservation")
	var body []string
	for _, word := range observation.Words {
		body = append(body, word...)
	}
	if observation.Expression != "" {
		body = append(body, observation.Expression)
	}
	entry, err := compilation.futureEntry("l_main", "runtime::Result<()>", body)
	if err != nil {
		t.Fatal(err)
	}
	declarations = append(declarations, entry)
	main := newRootObject("main")
	primary := newRootService("main")
	primary.Objects = []*object{main}
	app := &App{ID: "entry", Meta: appMetadata("entry"), Services: []*service{primary}}
	word := &w{Syllables: []*s{{CompiledBytes: []byte("A")}}}
	for _, declaration := range declarations {
		word.Syllables = append(word.Syllables, &s{CompiledBytes: []byte(declaration.Expression)})
	}
	program := &Program{Books: []*book{{Name: "entry", Chapters: []*chapter{{Name: "main", sentences: []*sentence{{Name: "main", Words: []*w{word}}}}}}}}
	_, aml, err := renderProgram(program, []*App{app})
	if err != nil {
		t.Fatal(err)
	}
	bundle, err := Logiceme.NewLogicemeService(phase1Options(t).Resources).Evaluate(aml, Logiceme.GenerationOptions{IncludeRuntime: true})
	if err != nil {
		t.Fatal(err)
	}
	return bundle
}

func TestRuntimeLifetimeSourceGeneration(t *testing.T) { runtimeLifetimeBundle(t) }

func TestRuntimeLifetimesNative(t *testing.T) {
	runtimeBundleResult(t, runtimeLifetimeBundle(t), "", "")
}
