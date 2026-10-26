package Logiceme

import (
	"bytes"
	"reflect"
	"testing"
)

func TestGenerationDefaultOptions(t *testing.T) {
	input := lucidRead(t, "grouping.aml.txt")
	original := bytes.Clone(input)
	service := NewLogicemeService(runtimeGenerationResources(t))
	want := lucidExpectedBundle(t, "grouping")
	for _, test := range []struct {
		name    string
		options []GenerationOptions
	}{
		{"omitted", nil},
		{"zero", []GenerationOptions{{}}},
	} {
		t.Run(test.name, func(t *testing.T) {
			got, err := service.Evaluate(input, test.options...)
			if err != nil {
				t.Fatalf("Rust source generation: %v", err)
			}
			if !reflect.DeepEqual(got, want) {
				t.Errorf("default generation differs from the complete Rust source oracle: paths=%v; want=%v", got.MapArrays, want.MapArrays)
			}
			if !bytes.Equal(input, original) {
				t.Fatal("generation mutated its AML input")
			}
		})
	}
}
