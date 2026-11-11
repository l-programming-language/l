package Lyger

import (
	"bytes"
	Logiceme "github.com/l-programming-language/l/logiceme"
	"strings"
	"testing"
	"testing/fstest"
)

func TestFunctionInvocationUsesEnclosingReturnComposition(t *testing.T) {
	source := "Define \"Callerreturn\" to `<DEF>BOD^E{|Statement: ~return /* caller */ <@Statement@>;~|}<DEF>`.\n" +
		"Define \"Calleereturn\" to `<DEF>BOD^E{|Statement: ~return /* callee */ <@Statement@>;~|}<DEF>`.\n" +
		"Define \"Invocation\" to `<DEF>KY{|Key: ~(/* selected invocation */ <@Target@>)~, Inputs: ~&[<@Arguments|Atoms@>]&~|}<DEF>`.\n\n" +
		"    Publicly create \"Done\" as a New Function with No Inputs and No Outputs.\nReturn.\n"
	c := executableTestResult(t, memberResolutionSource{main: source}).Compilation
	var definition *sourceFunction
	for _, function := range c.Semantics.Functions {
		definition = function
	}
	if definition == nil {
		t.Fatal("missing Function")
	}
	caller, callee := *definition, *definition
	selected := func(name string) *record {
		rec, err := c.exportedRecord(name)
		if err != nil {
			t.Fatal(err)
		}
		return rec
	}
	caller.Return = selected("Callerreturn")
	callee.Return = selected("Calleereturn")
	callee.Invoke = selected("Invocation")

	emitter := orchestrationEmitter{compilation: c,
		plan: &orchestrationPlan{Identities: map[*record]uint64{callee.Definition: 1}}, functionDepth: 1, functionWords: &caller,
		functionContracts: map[*record]*sourceFunction{callee.Definition: &callee},
	}
	statements, err := emitter.functionInstruction(&orchestrationInstruction{
		Source: &instruction{Kind: instructionInvoke},
		Target: &orchestrationValue{Type: valueFunction, RuntimeType: callee.signature(), Storage: &orchestrationStorage{}},
	})
	if err != nil {
		t.Fatal(err)
	}
	aml := strings.Join(statements, "*")
	if !strings.Contains(aml, "/* selected invocation */") || !strings.Contains(aml, "/* caller */") || strings.Contains(aml, "/* callee */") {
		t.Fatalf("invocation or enclosing return lost its selected definition: %s", aml)
	}
}

// Main construction must consume the selected Function definition even when
// the standalone App needs no Runtime composition.
func TestCanonicalFunctionDefinitionsOwnEmission(t *testing.T) {
	const core = "builtins/eng/core.l"
	const marker = "/* selected_Function */"
	files := dictionaryIndexFilesystem(t)
	original := bytes.Clone(files[core].Data)
	from := []byte("^(<@Process@>)")
	if bytes.Count(original, from) != 1 {
		t.Fatal("mutation must select exactly the core Function composition")
	}
	selected := bytes.Replace(original, from, []byte("^(W^(BOD^E{|Statement: ~"+marker+"~|})*<@Process@>)"), 1)
	files[core] = &fstest.MapFile{Data: selected}
	result, err := NewLygerService(Options{Resources: files}).Compile([]byte("Create \"Child\" as a New Basic App.\nThe App Prints `selected`.\n"), "selected-function.l")
	if err != nil {
		t.Fatal(err)
	}
	units, err := result.ApplicationUnits()
	if err != nil || len(units) != 1 {
		t.Fatalf("standalone App units: %v, %v", units, err)
	}
	bundle, err := Logiceme.NewLogicemeService(files).Evaluate(units[0].BackendAML)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Contains(units[0].BackendAML, []byte(marker)) || !strings.Contains(builtinProcessRust(bundle), marker) {
		t.Fatal("selected Function definition did not control App AML and Rust")
	}
	if !bytes.Equal(files[core].Data, selected) {
		t.Fatal("generation mutated selected source")
	}
}
