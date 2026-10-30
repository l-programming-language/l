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
	if comp.Program == nil || len(comp.Program.Books) == 0 {
		t.Fatal("successful compilation must retain its application graph")
	}
	mainStatements := 0
	for _, statement := range comp.SourceStatements {
		if statement.Span.Path == comp.MainPath {
			mainStatements++
		}
	}
	if mainStatements != 8 || len(comp.ProcStatements) != 0 {
		t.Fatalf("main/unresolved statements = %d/%d", mainStatements, len(comp.ProcStatements))
	}
	printAlias := comp.Memory["Print"]
	printDefinition, err := canonicalRecord(printAlias, comp.Semantics.Aliases)
	if err != nil || printDefinition == nil || printDefinition.ID != "IO.Print" {
		t.Fatalf("selected import lost IO.Print: %v", err)
	}
	if comp.Memory["IO"] != nil || comp.Memory["Write"] != nil || comp.Memory["Discard"] != nil {
		t.Fatal("selected IO import published its unit bindings")
	}
	output := comp.Memory["y"].Value
	if output.Type.String() != "Primitive" || output.Primitive.Type.String() != "Text" || output.Runtime == nil || output.Runtime.OwnerID != "helloworldapp" || output.Runtime.Member.String() != "output" {
		t.Fatal("Output assignment must retain typed deferred member access")
	}
	if _, err := argumentFromToken(&token{mem: output}); err == nil {
		t.Fatal("runtime stdout was treated as a compile-time literal")
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
