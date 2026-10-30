package Lyger

import "testing"

func TestFunctionHeaderRegistersBeforeConstantLocals(t *testing.T) {
	c := compileDefinitionSource(t, "Use Storage from `storage`.\n", "header.l")
	allocator, err := c.exportedRecord("Storage.Slot.Allocate")
	if err != nil {
		t.Fatal(err)
	}
	local, _, err := allocator.function.Scope.lookupLocal([]byte("Values"))
	if err != nil || local == nil || local.Record.definitionArgument == nil || local.Record.staticConstructor == nil {
		t.Fatalf("constant local committed before its composed Function header: %v", err)
	}
}
