package Lyger

import "testing"

func TestConditionOperatorNamesAndMappings(t *testing.T) {
	wants := []struct{ name, operator string }{
		{"", ""}, {"truth", ""}, {"has", ""}, {"is", "=="}, {"equals", "=="}, {"isnt", "!="},
		{"less", "<"}, {"greater", ">"}, {"less_equal", "<="}, {"greater_equal", ">="},
	}
	if int(conditionOperatorCount) != len(wants) {
		t.Fatal("condition variants changed without independent expectations")
	}
	for index, want := range wants {
		op := conditionOperator(index)
		if op.String() != want.name {
			t.Errorf("variant %d = %q, want %q", index, op.String(), want.name)
		}
		parsed, found := parseConditionOperator(want.name)
		if parsed != op || found != (index > 0) {
			t.Errorf("parse %q = %v/%v", want.name, parsed, found)
		}
		actual, binary := op.NativeOperator()
		if actual != want.operator || binary != (want.operator != "") {
			t.Errorf("operator %q = %q/%v", want.name, actual, binary)
		}
	}
	for _, raw := range []string{"IS", " is", "is ", "configured_extension"} {
		if op, ok := parseConditionOperator(raw); ok || op != conditionInvalid {
			t.Errorf("unknown %q accepted", raw)
		}
	}
	for _, invalid := range []conditionOperator{conditionInvalid, conditionOperatorCount, 255} {
		if _, ok := invalid.NativeOperator(); ok || invalid.accepts(valueNumber, valueNumber) {
			t.Errorf("invalid %d accepted operands", invalid)
		}
	}
	if conditionOperator(255).String() != "conditionOperator(255)" {
		t.Fatal("invalid operator diagnostic changed")
	}
}

func TestConditionOperatorScalarTypeBoundary(t *testing.T) {
	for _, tc := range []struct {
		op          string
		left, right valueType
		valid       bool
	}{
		{"is", valueText, valueText, true}, {"equals", valueTruth, valueTruth, true}, {"isnt", valueNumber, valueNumber, true},
		{"is", valueText, valueNumber, false}, {"equals", valueApp, valueApp, false}, {"isnt", valueError, valueError, false},
		{"is", valueInvalid, valueInvalid, false}, {"is", valueID, valueID, false},
		{"less", valueNumber, valueNumber, true}, {"greater", valueNumber, valueNumber, true},
		{"less_equal", valueNumber, valueNumber, true}, {"greater_equal", valueNumber, valueNumber, true},
		{"less", valueText, valueText, false}, {"greater", valueTruth, valueNumber, false},
		{"truth", valueTruth, valueInvalid, false}, {"has", valueApp, valueError, false},
	} {
		op, _ := parseConditionOperator(tc.op)
		if actual := op.accepts(tc.left, tc.right); actual != tc.valid {
			t.Errorf("%s(%s,%s) = %v", tc.op, tc.left, tc.right, actual)
		}
	}
}
