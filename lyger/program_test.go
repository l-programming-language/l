package Lyger

import (
	"bytes"
	"os"
	"testing"
)

func TestDemoCompilationRetainsSemanticProgram(t *testing.T) {
	source, err := os.ReadFile("test/main.l")
	if err != nil {
		t.Fatal(err)
	}
	config := canonicalTestOptions()
	comp := NewLygerService(config).NewCompilation(source, "test/main.l")
	comp.Make()
	if comp.Error != nil {
		for id, record := range comp.Memory {
			value, err := record.Value.value()
			t.Logf("memory %s: %s / %v (error: %v)", id, record.Type, value, err)
		}
		t.Fatalf("compile: %v", comp.Error)
	}
	if comp.Program == nil || len(comp.Program.Books) != 1 || comp.Program.Books[0].Name != "HelloWorldApp" {
		t.Fatal("successful compilation must retain its source application graph")
	}
	mainStatements := 0
	for _, statement := range comp.SourceStatements {
		if statement.Span.Path == comp.MainPath {
			mainStatements++
		}
	}
	if mainStatements != 3 || len(comp.ProcStatements) != 0 {
		t.Fatalf("main/unresolved statements = %d/%d", mainStatements, len(comp.ProcStatements))
	}
	printDefinition, err := comp.builtinRecord("Print")
	if err != nil || printDefinition == nil || printDefinition.ID != "Print" {
		t.Fatalf("implicit builtin lost core Print: %v", err)
	}
	declaration := comp.Semantics.Declarations[printDefinition]
	if declaration == nil || declaration.source == nil || declaration.source.Unit == nil || declaration.source.Unit.Kind != SourceBuiltin || declaration.source.Unit.Origin.Name != "builtins/eng/core.l" {
		t.Fatal("Print lost its canonical builtin source provenance")
	}
	before := bytes.Clone(comp.ASMBytes)
	program := comp.Program
	comp.Make()
	if comp.Program != program || !bytes.Equal(comp.ASMBytes, before) {
		t.Fatal("repeated Make duplicated or changed a completed program")
	}
	second, err := NewLygerService(config).Compile(source, "test/main.l")
	if err != nil {
		t.Fatal(err)
	}
	if second.Compilation.Program == program || !bytes.Equal(second.AML, comp.ASMBytes) {
		t.Fatal("recompilation shared working state or changed emitted bytes")
	}
}
