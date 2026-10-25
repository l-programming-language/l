package Logiceme

import (
	"reflect"
	"testing"
)

func TestConjoinmentOptionalIndexRegression(t *testing.T) {
	for _, language := range []string{"Rust", ""} {
		for _, tc := range []struct {
			name, query, wantType string
			want                  []string
			fragment              string
		}{
			{"present", "Calculations[0]", "Calculation", []string{"present()"}, "present()"},
			{"missing", "Calculations[1]", "Void", []string{"Void"}, ""},
		} {
			t.Run(language+"/"+tc.name, func(t *testing.T) {
				defer executionRegistryPanicAsFailure(t)
				interpreter := LucidInterpreter{
					SegmentalSystem: &SegmentalSystem{Language: language},
					Conjoinments: []CompiledConjoinment{{
						Type: "Calculations", Plural: true, Contents: [][]byte{[]byte("present()")},
					}},
				}
				before := conjoinmentRegressionSnapshot(t, interpreter.Conjoinments)
				defer func() {
					if conjoinmentRegressionSnapshot(t, interpreter.Conjoinments) != before {
						t.Error("indexed interpolation mutated conjoinment source")
					}
				}()
				values, valueType := interpreter.conjoinmentsSwitch(tc.query)
				if valueType != tc.wantType || !reflect.DeepEqual(values, tc.want) {
					t.Errorf("lookup %s = (%q, %q); want (%q, %q)", tc.query, values, valueType, tc.want, tc.wantType)
				}
				for _, marker := range []string{"", "%", "&"} {
					placeholder := "[|" + marker + "Conjoinments." + tc.query + marker + "|]"
					want := "left(" + tc.fragment + ")right"
					if got := interpreter.interpretExpressionVars([]string{placeholder}, "left("+placeholder+")right"); got != want {
						t.Errorf("%s interpolation = %q; want %q", placeholder, got, want)
					}
				}
			})
		}
	}
}

func TestConjoinmentLiteralVoidRegression(t *testing.T) {
	for _, language := range []string{"Rust", ""} {
		for _, tc := range []struct {
			category, query string
			plural          bool
		}{
			{"Calculation", "Calculation", false},
			{"Calculations", "Calculations[0]", true},
		} {
			t.Run(language+"/"+tc.query, func(t *testing.T) {
				defer executionRegistryPanicAsFailure(t)
				interpreter := LucidInterpreter{
					SegmentalSystem: &SegmentalSystem{Language: language},
					Conjoinments: []CompiledConjoinment{{
						Type: tc.category, Plural: tc.plural, Contents: [][]byte{[]byte("Void")},
					}},
				}
				before := conjoinmentRegressionSnapshot(t, interpreter.Conjoinments)
				defer func() {
					if conjoinmentRegressionSnapshot(t, interpreter.Conjoinments) != before {
						t.Error("literal interpolation mutated conjoinment source")
					}
				}()
				values, valueType := interpreter.conjoinmentsSwitch(tc.query)
				if valueType != "Calculation" || !reflect.DeepEqual(values, []string{"Void"}) {
					t.Errorf("present Void lookup = (%q, %q); want ([Void], Calculation)", values, valueType)
				}
				for _, marker := range []string{"", "%", "&"} {
					placeholder := "[|" + marker + "Conjoinments." + tc.query + marker + "|]"
					if got := interpreter.interpretExpressionVars([]string{placeholder}, "left("+placeholder+")right"); got != "left(Void)right" {
						t.Errorf("%s interpolation = %q; want literal Void preserved", placeholder, got)
					}
				}
				// Existing conditions deliberately recognize this explicit sentinel.
				if got := interpreter.interpretCondition("if Conjoinments." + tc.query + " == Void"); got != "true" {
					t.Errorf("explicit Void condition = %q; want true", got)
				}
			})
		}
	}
}

func TestConjoinmentEmptyPresenceRegression(t *testing.T) {
	for _, language := range []string{"Rust", ""} {
		for _, tc := range []struct {
			name       string
			present    bool
			contents   [][]byte
			want       []string
			wantType   string
			lookupOnly bool
		}{
			{"absent", false, nil, []string{"Void"}, "Void", false},
			{"nil_contents", true, nil, nil, "Calculation", true},
			{"empty_contents", true, [][]byte{}, nil, "Calculation", true},
			{"nil_fragment", true, [][]byte{nil}, []string{""}, "Calculation", false},
			{"empty_fragment", true, [][]byte{[]byte{}}, []string{""}, "Calculation", false},
		} {
			t.Run(language+"/"+tc.name, func(t *testing.T) {
				defer executionRegistryPanicAsFailure(t)
				interpreter := LucidInterpreter{SegmentalSystem: &SegmentalSystem{Language: language}}
				if tc.present {
					interpreter.Conjoinments = []CompiledConjoinment{
						{Type: "Calculation", Contents: tc.contents},
						{Type: "Calculation", Contents: [][]byte{[]byte("later()")}},
					}
				}
				before := conjoinmentRegressionSnapshot(t, interpreter.Conjoinments)
				defer func() {
					if conjoinmentRegressionSnapshot(t, interpreter.Conjoinments) != before {
						t.Error("empty-fragment handling mutated conjoinment source")
					}
				}()
				values, valueType := interpreter.conjoinmentsSwitch("Calculation")
				if valueType != tc.wantType || !reflect.DeepEqual(values, tc.want) {
					t.Errorf("lookup = (%q, %q); want (%q, %q)", values, valueType, tc.want, tc.wantType)
				}
				if tc.lookupOnly {
					return // Preserve the existing rendering contract for zero-item records.
				}
				wantPresence := "false"
				if tc.present {
					wantPresence = "true"
				}
				if got := interpreter.interpretCondition("if Conjoinments.Calculation != Void"); got != wantPresence {
					t.Errorf("presence condition = %q; want %q", got, wantPresence)
				}
				for _, marker := range []string{"", "%", "&"} {
					placeholder := "[|" + marker + "Conjoinments.Calculation" + marker + "|]"
					if got := interpreter.interpretExpressionVars([]string{placeholder}, "left("+placeholder+")right"); got != "left()right" {
						t.Errorf("%s interpolation = %q; want left()right", placeholder, got)
					}
				}
			})
		}
	}
}

func TestConjoinmentMetadataFallbackRegression(t *testing.T) {
	child := Glyph{
		Type: "Calculation", Vowels: []Segment{{Symbol: "E"}},
		Content: []byte("[|Meta.Statement|] + [|Meta.Fallback|]"),
	}
	child.Meta.SetAdd("Statement", "child()", "String")
	childBefore := conjoinmentRegressionSnapshot(t, child)
	for _, value := range []string{"first()", "second()"} {
		t.Run(value, func(t *testing.T) {
			defer executionRegistryPanicAsFailure(t)
			parent := Glyph{Conjoinments: []Glyph{child}}
			parent.Meta.SetAdd("Statement", value, "String")
			parent.Meta.SetAdd("Fallback", "suffix()", "String")
			before := conjoinmentRegressionSnapshot(t, parent)
			defer func() {
				if conjoinmentRegressionSnapshot(t, parent) != before || conjoinmentRegressionSnapshot(t, child) != childBefore {
					t.Error("repeated composition mutated parent/child metadata or source")
				}
			}()
			compiled := parent.CompileConjoinments([]Parameter{{Key: "Calculation", Type: "Conjoinments", SubType: "Calculation"}})
			if parent.err != nil {
				t.Fatal(parent.err)
			}
			if len(compiled) != 1 || len(compiled[0].Contents) != 1 {
				t.Fatalf("expected one compiled child: %#v", compiled)
			}
			if got := string(compiled[0].Contents[0]); got != "child() + [|Meta.Fallback|]" {
				t.Errorf("child binding = %q; want local Statement and unresolved Fallback", got)
			}
			if got := string(InterpretConjoinments([]byte("[|Conjoinments.Calculation|]"), compiled, parent.Meta)); got != "child() + suffix()" {
				t.Errorf("parent composition = %q; want child() + suffix()", got)
			}
		})
	}
}

func TestConjoinmentMetadataArrayRegression(t *testing.T) {
	defer executionRegistryPanicAsFailure(t)
	child := Glyph{
		Type: "Calculation", Vowels: []Segment{{Symbol: "E"}},
		Content: []byte("[|Meta.Keys~|] / [|Meta.Keys|]"),
	}
	child.Meta.SetAdd("Keys", "&[alpha, beta]&", "Array-String")
	parent := Glyph{Conjoinments: []Glyph{child}}
	parent.Meta.SetAdd("Keys", "&[wrong, values]&", "Array-String")
	before := conjoinmentRegressionSnapshot(t, parent)
	defer func() {
		if conjoinmentRegressionSnapshot(t, parent) != before {
			t.Error("array composition mutated parent/child metadata or source")
		}
	}()
	compiled := parent.CompileConjoinments([]Parameter{{Key: "Calculation", Type: "Conjoinments", SubType: "Calculation"}})
	if parent.err != nil {
		t.Fatal(parent.err)
	}
	if len(compiled) != 1 || len(compiled[0].Contents) != 1 {
		t.Fatalf("expected one compiled child: %#v", compiled)
	}
	const want = "alpha, beta / &[alpha, beta]&"
	if got := string(compiled[0].Contents[0]); got != want {
		t.Errorf("child array binding = %q; want %q", got, want)
	}
	if got := string(InterpretConjoinments([]byte("[|Conjoinments.Calculation|]"), compiled, parent.Meta)); got != want {
		t.Errorf("parent array composition = %q; want %q", got, want)
	}
}
