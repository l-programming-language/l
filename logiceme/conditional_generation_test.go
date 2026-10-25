package Logiceme

import (
	"reflect"
	"testing"
)

func TestConditionalConjoinmentLookup(t *testing.T) {
	system := SegmentalSystem{Language: "Rust"}
	interpreter := LucidInterpreter{SegmentalSystem: &system, Conjoinments: []CompiledConjoinment{
		{Type: "Operations", Plural: true, Contents: [][]byte{[]byte("first body"), []byte("second body")}},
		{Type: "Calculations", Plural: true, Contents: [][]byte{[]byte("true"), []byte("Void")}},
	}}
	interpreter.createVariable("Count", "Number", reflect.ValueOf(2))
	for _, tc := range []struct{ input, want string }{
		{"[|Conjoinments.Operations[0]|]", "first body"},
		{"[|Conjoinments.Operations[Add(Count, -1)]|]", "second body"},
		{"[|%Conjoinments.Calculations[Add(Count, -1)]%|]", "Void"},
	} {
		if got := interpreter.interpretExpressionVars([]string{tc.input}, tc.input); got != tc.want {
			t.Errorf("%s = %q; want %q", tc.input, got, tc.want)
		}
	}
	if got := interpreter.interpretCondition("if Conjoinments.Calculations[Add(Count, -1)] == Void"); got != "true" {
		t.Errorf("indexed Void condition = %s; want true", got)
	}
}
