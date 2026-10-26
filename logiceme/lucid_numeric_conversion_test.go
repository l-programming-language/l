package Logiceme

import "testing"

func TestLucidNumericComparisonRetainsConversionFallbacks(t *testing.T) {
	for _, tc := range []struct {
		name, value, condition, operator, want string
	}{
		{"invalid integer value is zero", "invalid", "0", "GreaterOrEqual", "true"},
		{"invalid integer condition is zero", "1", "invalid", "Greater", "true"},
		{"positive integer overflow saturates", "999999999999999999999999999999", "0", "Greater", "true"},
		{"negative integer overflow saturates", "-999999999999999999999999999999", "0", "Lesser", "true"},
		{"integer condition overflow saturates", "0", "999999999999999999999999999999", "Lesser", "true"},
		{"invalid float value is zero", "invalid.number", "0", "GreaterOrEqual", "true"},
		{"invalid float condition is zero", "1.5", "invalid", "Greater", "true"},
		{"positive float overflow is infinity", "1.0e999", "1.0e300", "Greater", "true"},
		{"negative float overflow is infinity", "-1.0e999", "-1.0e300", "Lesser", "true"},
		{"float condition overflow is infinity", "1.0", "1.0e999", "Lesser", "true"},
		{"invalid integer stays at zero", "invalid", "0", "Greater", "false"},
		{"invalid float stays at zero", "invalid.number", "0", "Greater", "false"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			interpreter := LucidInterpreter{}
			if got := interpreter.conditionComparison(tc.value, "Number", tc.condition, tc.operator); got != tc.want {
				t.Fatalf("numeric comparison = %q, want %q", got, tc.want)
			}
		})
	}
}

func TestLucidSubtagCollectionRetainsOrderAndOwnership(t *testing.T) {
	for _, expressions := range [][]*TagExpression{
		nil,
		{},
		{{ExpressionSubTags: nil}},
		{{ExpressionSubTags: []*SubTag{}}},
	} {
		tag := LucidTag{Expressions: expressions}
		if got := tag.GetSubTags(); got != nil {
			t.Fatalf("empty subtag collection = %#v, want nil", got)
		}
	}
	first, second := &SubTag{}, &SubTag{}
	left, right := []*SubTag{first}, []*SubTag{second}
	tag := LucidTag{Expressions: []*TagExpression{{ExpressionSubTags: left}, {ExpressionSubTags: right}}}
	got := tag.GetSubTags()
	if len(got) != 2 || got[0] != first || got[1] != second {
		t.Fatalf("subtag collection changed order or element identity: %#v", got)
	}
	got[0], got[1] = second, first
	if left[0] != first || right[0] != second {
		t.Fatal("subtag collection reused an input slice")
	}
}

func TestCombinedVariableManagersRetainSliceSemantics(t *testing.T) {
	if got := combineVarMgmrs(nil, nil); got == nil || len(got) != 0 {
		t.Fatalf("empty variable manager = %#v, want a non-nil empty slice", got)
	}
	left, right := [][]string{{"left", "one"}}, [][]string{{"right", "two"}}
	got := combineVarMgmrs(left, right)
	if len(got) != 2 || len(got[0]) != 2 || len(got[1]) != 2 || got[0][0] != "left" || got[1][0] != "right" {
		t.Fatalf("variable-manager combination changed ordering: %#v", got)
	}
	got[0][1], got[1][1] = "updated left", "updated right"
	if left[0][1] != "updated left" || right[0][1] != "updated right" {
		t.Fatal("variable-manager combination deep-copied shared entries")
	}
	got[0], got[1] = nil, nil
	if left[0] == nil || right[0] == nil {
		t.Fatal("variable-manager combination reused an input outer slice")
	}
}
