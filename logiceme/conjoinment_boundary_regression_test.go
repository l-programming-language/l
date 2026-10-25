package Logiceme

import (
	"encoding/json"
	"fmt"
	"reflect"
	"testing"
)

// Snapshot values, including nested source buffers and metadata, rather than
// retaining aliases that would hide an in-place mutation during composition.
func conjoinmentRegressionSnapshot(t *testing.T, value interface{}) string {
	t.Helper()
	data, err := json.Marshal(value)
	if err != nil {
		t.Fatal(err)
	}
	return string(data)
}

func TestConjoinmentPresenceRegression(t *testing.T) {
	newInterpreter := func(category string, present bool) LucidInterpreter {
		unrelated := "Operation"
		if category == unrelated {
			unrelated = "Calculation"
		}
		interpreter := LucidInterpreter{
			SegmentalSystem: &SegmentalSystem{},
			Conjoinments: []CompiledConjoinment{
				{Type: unrelated, Contents: [][]byte{[]byte("unrelated()")}},
				{Type: category + "s", Plural: true, Contents: [][]byte{[]byte("plural()")}},
			},
		}
		if present {
			interpreter.Conjoinments = append(interpreter.Conjoinments, CompiledConjoinment{
				Type: category, Contents: [][]byte{[]byte("present()")},
			})
		}
		return interpreter
	}

	for _, category := range []string{
		"Declaration", "Evaluation", "Calculation", "Definition", "Description",
		"Operation", "Iteration", "Instantiation", "Execution",
	} {
		for _, present := range []bool{false, true} {
			name := "missing"
			if present {
				name = "present"
			}
			t.Run(category+"/lookup/"+name, func(t *testing.T) {
				defer executionRegistryPanicAsFailure(t)
				interpreter := newInterpreter(category, present)
				before := conjoinmentRegressionSnapshot(t, interpreter.Conjoinments)
				defer func() {
					if conjoinmentRegressionSnapshot(t, interpreter.Conjoinments) != before {
						t.Error("lookup mutated conjoinment inputs")
					}
				}()
				wantType, wantValues := "Void", []string{"Void"}
				if present {
					wantType, wantValues = category, []string{"present()"}
				}
				values, valueType := interpreter.conjoinmentsSwitch(category)
				if valueType != wantType || !reflect.DeepEqual(values, wantValues) {
					t.Errorf("lookup %s = (%q, %q); want (%q, %q)", category, values, valueType, wantValues, wantType)
				}
			})
		}
	}

	for _, category := range []string{"Definition", "Evaluation", "Calculation", "Operation"} {
		for _, tc := range []struct{ name, operator, want string }{
			{"equals_void", "==", "true"},
			{"not_equals_void", "!=", "false"},
		} {
			t.Run(category+"/condition/"+tc.name, func(t *testing.T) {
				defer executionRegistryPanicAsFailure(t)
				interpreter := newInterpreter(category, false)
				before := conjoinmentRegressionSnapshot(t, interpreter.Conjoinments)
				defer func() {
					if conjoinmentRegressionSnapshot(t, interpreter.Conjoinments) != before {
						t.Error("condition evaluation mutated conjoinment inputs")
					}
				}()
				condition := "if Conjoinments." + category + " " + tc.operator + " Void"
				if got := interpreter.interpretCondition(condition); got != tc.want {
					t.Errorf("%s = %q; want %q", condition, got, tc.want)
				}
			})
		}
		for _, tc := range []struct{ name, marker string }{
			{"ordinary", ""}, {"percent", "%"}, {"leading_only", "&"},
		} {
			t.Run(category+"/interpolation/"+tc.name, func(t *testing.T) {
				defer executionRegistryPanicAsFailure(t)
				interpreter := newInterpreter(category, false)
				before := conjoinmentRegressionSnapshot(t, interpreter.Conjoinments)
				defer func() {
					if conjoinmentRegressionSnapshot(t, interpreter.Conjoinments) != before {
						t.Error("interpolation mutated conjoinment inputs")
					}
				}()
				placeholder := "[|" + tc.marker + "Conjoinments." + category + tc.marker + "|]"
				if got := interpreter.interpretExpressionVars([]string{placeholder}, "left("+placeholder+")right"); got != "left()right" {
					t.Errorf("absent %s interpolation = %q; want %q", category, got, "left()right")
				}
			})
		}
	}
}

func TestConjoinmentMetadataScopeRegression(t *testing.T) {
	newChild := func(category, symbol, statement, content string) Glyph {
		meta := MetaDict{}
		meta.SetAdd("Statement", statement, "String")
		return Glyph{Type: category, Vowels: []Segment{{Symbol: symbol}}, Content: []byte(content), Meta: meta}
	}
	for _, tc := range []struct {
		name              string
		plural, reordered bool
		content           string
		want              []string
	}{
		{"singular", false, false, "[|Meta.Statement|]", []string{"println(1)"}},
		{"singular_reordered_siblings", false, true, "[|Meta.Statement|]", []string{"println(1)"}},
		{"plural", true, false, "[|Meta.Statement|]", []string{"println(1)", "println(2)"}},
		{"plural_reordered_siblings", true, true, "[|Meta.Statement|]", []string{"println(2)", "println(1)"}},
		{"void_literal", false, false, `println("Void")`, []string{`println("Void")`}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			defer executionRegistryPanicAsFailure(t)
			parent := Glyph{Conjoinments: []Glyph{
				newChild("Calculation", "E", "println(1)", tc.content),
				newChild("Definition", "O", "println(9)", "[|Meta.Statement|]"),
			}}
			if tc.plural {
				parent.Conjoinments = append(parent.Conjoinments, newChild("Calculation", "E", "println(2)", tc.content))
			}
			if tc.reordered {
				for i, j := 0, len(parent.Conjoinments)-1; i < j; i, j = i+1, j-1 {
					parent.Conjoinments[i], parent.Conjoinments[j] = parent.Conjoinments[j], parent.Conjoinments[i]
				}
			}
			parent.Meta.SetAdd("Statement", "println(99)", "String")
			before := conjoinmentRegressionSnapshot(t, parent)
			defer func() {
				if conjoinmentRegressionSnapshot(t, parent) != before {
					t.Error("composition mutated child source buffers or parent/child metadata")
				}
			}()
			category := "Calculation"
			if tc.plural {
				category += "s"
			}
			compiled := parent.CompileConjoinments([]Parameter{{Key: category, Type: "Conjoinments", SubType: category, Plural: tc.plural}})
			if parent.err != nil {
				t.Fatal(parent.err)
			}
			if len(compiled) != 1 || len(compiled[0].Contents) != len(tc.want) {
				t.Fatalf("compiled %s collection does not contain the expected children: %#v", category, compiled)
			}
			for i, want := range tc.want {
				var got string
				if tc.plural {
					interpreter := LucidInterpreter{SegmentalSystem: &SegmentalSystem{}, Meta: parent.Meta, Conjoinments: compiled}
					placeholder := fmt.Sprintf("[|Conjoinments.Calculations[%d]|]", i)
					got = interpreter.interpretExpressionVars([]string{placeholder}, placeholder)
				} else {
					got = string(InterpretConjoinments([]byte("[|Conjoinments.Calculation|]"), compiled, parent.Meta))
				}
				if got != want {
					t.Errorf("child %d source = %q; want its own statement %q", i, got, want)
				}
			}
		})
	}
}
