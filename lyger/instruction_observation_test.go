package Lyger

import (
	"reflect"
	"testing"
)

func flattenedInstructions(in []*instruction) []*instruction {
	var out []*instruction
	for _, i := range in {
		if i.Kind.String() == "call" {
			out = append(out, flattenedInstructions(i.Actions)...)
		} else if i.Block != nil {
			out = append(out, flattenedInstructions(i.Block.Body)...)
		} else {
			out = append(out, i)
			out = append(out, flattenedInstructions(i.Then)...)
			out = append(out, flattenedInstructions(i.Else)...)
		}
	}
	return out
}
func flattenedInstructionKinds(in []*instruction) []string {
	var out []string
	for _, i := range flattenedInstructions(in) {
		out = append(out, i.Kind.String())
	}
	return out
}
func requireLifecycleInstructions(t *testing.T, r *CompilationResult, want []string) []*instruction {
	t.Helper()
	var got []string
	var ops []*instruction
	for _, i := range flattenedInstructions(r.Compilation.Semantics.Instructions) {
		switch i.Kind.String() {
		case "build", "run", "await", "stop":
			got = append(got, i.Kind.String())
			ops = append(ops, i)
			if i.Target == nil || i.Target.Type.String() != "App" || i.Target.Record == nil {
				t.Fatalf("%s lost typed App operand: %+v", i.Kind, i.Target)
			}
		}
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("lifecycle IR=%v want %v", got, want)
	}
	return ops
}
func canonicalInstructionTarget(t *testing.T, r *CompilationResult, i *instruction) *record {
	t.Helper()
	rec, err := canonicalRecord(i.Target.Record, r.Compilation.Semantics.Aliases)
	if err != nil {
		t.Fatal(err)
	}
	return rec
}

const instructionWorkerSource = "\n\"message\" is `child`.\nCreate \"Worker\" as a New Basic App.\nThe App Prints message.\n"
