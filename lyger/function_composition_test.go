package Lyger

import (
	"bytes"
	Logiceme "github.com/l-programming-language/l/logiceme"
	"strings"
	"testing"
	"testing/fstest"
)

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
	if !bytes.Contains(units[0].BackendAML, []byte(marker)) || !strings.Contains(generatedRustSource(bundle), marker) {
		t.Fatal("selected Function definition did not control App AML and Rust")
	}
	if !bytes.Equal(files[core].Data, selected) {
		t.Fatal("generation mutated selected source")
	}
}
