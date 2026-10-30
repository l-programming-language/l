package Lyger

import (
	"bytes"
	"strings"
	"testing"

	Logiceme "github.com/l-programming-language/l/logiceme"
)

func TestDictionaryLibraryOwnsDeclarationDependencies(t *testing.T) {
	source := "Create \"Words\" as a Dictionary.\n\n" +
		"    Publicly create \"Echo|-es|-ed|-ing|-er\" as a Generic Function with Inputs of \"value\" as `f64` and Outputs of `f64`.\nReturn Value.\n\n" +
		"    Publicly create Words's \"Supply|-ies|-ied|-ying|-ier\" as a Generic Function with No Inputs and Outputs of `f64`.\nCreate \"Result\" as a New Echoer of 2.\nReturn Result.\n\n" +
		"    Publicly create \"Unused\" as a Generic Function with No Inputs and Outputs of `f64`.\nReturn 3.\n"
	c := compileDefinitionSource(t, source, "library.l")
	aml, err := c.ComposeLibrary("Words", "runtime")
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Contains(aml, []byte(`"Kind":"lib"`)) || strings.Count(string(aml), "Key: ~echo~") != 1 || bytes.Contains(aml, []byte("unused")) {
		t.Fatalf("library lost selected declaration closure: %s", aml)
	}
	bundle, err := Logiceme.NewLogicemeService(canonicalTestOptions().Resources).Evaluate(aml)
	if err != nil {
		t.Fatal(err)
	}
	var generated []byte
	for i, paths := range bundle.MapArrays {
		for j, path := range paths {
			if path == "/src/runtime.rs" {
				generated = bundle.ByteArrays[i][j]
			}
		}
	}
	if len(generated) == 0 || !bytes.Contains(generated, []byte("fn echo(")) || bytes.Contains(generated, []byte("fn unused(")) {
		t.Fatalf("normal AML library did not render selected declarations: %s", generated)
	}
}

func TestDeclarationClosureRejectsConflictingExpansions(t *testing.T) {
	rec := &record{ID: "selected"}
	first := DefinitionInstance{Symbol: rec.ID, definition: rec, Expression: "W"}
	second := first
	second.Expression = "W^E"
	if _, err := definitionClosure(first, second); err == nil {
		t.Fatal("conflicting expansions of one canonical declaration were accepted")
	}
	result, err := definitionClosure(first, first)
	if err != nil || len(result) != 1 {
		t.Fatalf("identical repeated dependency: %v, %v", result, err)
	}
}
