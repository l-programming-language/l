package Logiceme

import (
	"encoding/json"
	"reflect"
	"testing"
)

func TestLucidIndexedOperationFields(t *testing.T) {
	ops := []Op{{Type: "Calculate", Prefix: "(", Operator: "+", Operands: []string{"left", "right"}, Postfix: ")"}, {Type: "Calculate", Prefix: "[", Operands: []string{"other"}, Postfix: "]"}}
	encoded, err := json.Marshal(ops)
	if err != nil {
		t.Fatal(err)
	}
	for _, language := range []string{"Rust"} {
		for _, tc := range []struct{ name, source, want string }{
			{"literal_index", "[|Meta.Ops[0].Operands[0]|]/[|Meta.Ops[0].Operands[1]|]", "left/right"},
			{"computed_index", "[|Meta.Ops[Add(Count, -1)].Prefix|][|Meta.Ops[Add(Count, -1)].Operands[0]|][|Meta.Ops[Add(Count, -1)].Postfix|]", "(left)"},
			{"computed_field_index", "[|Meta.Ops[0].Operands[Add(Count, 0)]|]", "right"},
			{"parenthesized", "[|(Meta.Ops[0].Operands[0])|]", "(left)"},
			{"inline", "[|%Meta.Ops[0].Operands[0]%|]", "left"},
			{"trim_prefix", "[|&Meta.Ops[0].Operands[0]&|]", "left"},
			{"loop_binding", nestedRenderLoop("Current", "Meta.Ops", "[|Current.Prefix|][|Current.Operands[0]|][|Current.Postfix|];"), "(left);[other];"},
		} {
			t.Run(language+"/"+tc.name, func(t *testing.T) {
				meta := MetaDict{}
				meta.SetAdd("Ops", string(encoded), "PreparedOps")
				li := LucidInterpreter{RawContent: []byte(tc.source), SubTags: RecursiveSubTagParser(tc.source, &TagMeta{}), Meta: meta,
					SegmentalSystem: &SegmentalSystem{Language: language}, WorkingVariables: []*WorkingVariable{{VarName: "Count", VarType: "Number", VarValue: reflect.ValueOf(1)}}}
				before := marshalTestJSON(t, li.WorkingVariables)
				for run := 0; run < 2; run++ {
					li.Interpret()
					if got := string(li.InterpretedContent); got != tc.want {
						t.Errorf("run %d: got %q; want %q", run, got, tc.want)
					}
					after := marshalTestJSON(t, li.WorkingVariables)
					if string(before) != string(after) || string(li.RawContent) != tc.source {
						t.Error("indexed lookup changed source or bindings")
					}
				}
			})
		}
	}
}

func nestedRenderCondition(body string, enabled bool) string {
	op := "!="
	if !enabled {
		op = "=="
	}
	return "(/={tags: Meta.Tags}=={if Meta.Key " + op + " Void}=={" + body + "}=/)"
}

func nestedRenderLoop(name, collection, body string) string {
	return "(/={tags: Meta.Tags}=={for " + name + " in " + collection + "}=={" + body + "}=/)"
}

func TestLucidNestedRendering(t *testing.T) {
	child := nestedRenderCondition("inside", true)
	for _, language := range []string{"Rust"} {
		for _, tc := range []struct{ name, source, want string }{
			{"mixed_literals", nestedRenderCondition("before"+child+"between"+child+"after", true), "beforeinsidebetweeninsideafter"},
			{"false_parent", nestedRenderCondition("before"+child+"after", false), ""},
			{"empty_children", nestedRenderCondition("before"+nestedRenderCondition("hidden", false)+"after", true), "beforeafter"},
			{"empty_body", nestedRenderCondition("", true), ""},
			{"three_levels", nestedRenderCondition("a"+nestedRenderCondition("b"+child+"c", true)+"d", true), "abinsidecd"},
			{"crlf_unicode", nestedRenderCondition("雪\r\n"+child+"\r\n光", true), "雪\r\ninside\r\n光"},
			{"braces", nestedRenderCondition("{"+child+"}", true), "{inside}"},
			{"empty_loop", nestedRenderLoop("Item", "Empty", "[|Item|];"), ""},
			{"one_iteration", nestedRenderLoop("Item", "One", "[|Item|];"), "z;"},
			{"explicit_separator", nestedRenderLoop("Item", "Items", "[|Item|];"), "a;b;"},
			{"explicit_newline", nestedRenderLoop("Item", "Items", "[|Item|]\r\n"), "a\r\nb\r\n"},
			{"nested_pairs", nestedRenderLoop("Outer", "Items", nestedRenderLoop("Inner", "Other", "[|Outer|][|Inner|];")), "ax;ay;bx;by;"},
			{"loop_mixed_literals", nestedRenderLoop("Outer", "Items", "[|Outer|]("+nestedRenderLoop("Inner", "Other", "[|Inner|]")+")[|Outer|];"), "a(xy)a;b(xy)b;"},
			{"condition_in_loop", nestedRenderLoop("Item", "Items", nestedRenderCondition("[|Item|];", true)), "a;b;"},
			{"loop_in_condition", nestedRenderCondition("("+nestedRenderLoop("Item", "Items", "[|Item|]")+")", true), "(ab)"},
			{"same_name_shadowing", nestedRenderLoop("Item", "Items", "[|Item|]("+nestedRenderLoop("Item", "Other", "[|Item|]")+")[|Item|];"), "a(xy)a;b(xy)b;"},
			{"sibling_loops", nestedRenderLoop("Item", "Items", "[|Item|]") + "/" + nestedRenderLoop("Item", "Other", "[|Item|]") + "/[|Item|]", "ab/xy/saved"},
			{"restored_existing", nestedRenderLoop("Item", "Items", "[|Item|]") + "[|Item|]", "absaved"},
			{"restored_empty", nestedRenderLoop("Item", "Empty", "[|Item|]") + "[|Item|]", "saved"},
		} {
			t.Run(language+"/"+tc.name, func(t *testing.T) {
				meta := MetaDict{}
				meta.SetAdd("Key", "present", "String")
				li := LucidInterpreter{
					RawContent: []byte(tc.source), SubTags: RecursiveSubTagParser(tc.source, &TagMeta{}),
					SegmentalSystem: &SegmentalSystem{Language: language}, Meta: meta,
					WorkingVariables: []*WorkingVariable{
						{VarName: "Items", VarType: "Array-String", VarValue: []string{"a", "b"}},
						{VarName: "Other", VarType: "Array-String", VarValue: []string{"x", "y"}},
						{VarName: "Empty", VarType: "Array-String", VarValue: []string{}},
						{VarName: "One", VarType: "Array-String", VarValue: []string{"z"}},
						{VarName: "Item", VarType: "String", VarValue: "saved"},
					},
				}
				before := marshalTestJSON(t, li.SubTags)
				bindings := make([]WorkingVariable, len(li.WorkingVariables))
				for i, v := range li.WorkingVariables {
					bindings[i] = *v
				}
				for run := 0; run < 2; run++ {
					li.Interpret()
					if got := string(li.InterpretedContent); got != tc.want {
						t.Errorf("run %d: rendered %q; want %q", run, got, tc.want)
					}
					if len(li.WorkingVariables) != len(bindings) {
						t.Errorf("run %d: loop bindings leaked", run)
					} else {
						for i, v := range li.WorkingVariables {
							if !reflect.DeepEqual(*v, bindings[i]) {
								t.Errorf("run %d: binding %s changed", run, v.VarName)
							}
						}
					}
				}
				after := marshalTestJSON(t, li.SubTags)
				if string(before) != string(after) || string(li.RawContent) != tc.source {
					t.Error("interpretation mutated source or parsed children")
				}
			})
		}
	}
}

func TestLucidNestedControlOccurrences(t *testing.T) {
	for _, language := range []string{"Rust"} {
		t.Run(language, func(t *testing.T) {
			child := nestedRenderCondition("inside", true)
			source := nestedRenderCondition("[|Meta.Payload|]", true) + "/" + child
			meta := MetaDict{}
			meta.SetAdd("Key", "present", "String")
			meta.SetAdd("Payload", child, "String")
			li := LucidInterpreter{RawContent: []byte(source), SubTags: RecursiveSubTagParser(source, &TagMeta{}), SegmentalSystem: &SegmentalSystem{Language: language}, Meta: meta}
			li.Interpret()
			if got, want := string(li.InterpretedContent), child+"/inside"; got != want {
				t.Errorf("rendered %q; want generated control-like text unchanged: %q", got, want)
			}
		})
	}
}

func TestLucidNestedBoundConjoinmentModifiers(t *testing.T) {
	for _, language := range []string{"Rust"} {
		for _, tc := range []struct{ token, want string }{
			{"[|Calculation|]", "\n\nvalue()\n"},
			{"[|%Calculation%|]", "value()"},
			{"[|&Calculation&|]", "value()\n"},
		} {
			t.Run(language+"/"+tc.token, func(t *testing.T) {
				li := LucidInterpreter{SegmentalSystem: &SegmentalSystem{Language: language}, WorkingVariables: []*WorkingVariable{
					{VarName: "Calculation", VarType: "Conjoinment.Calculation", VarValue: reflect.ValueOf("\n\nvalue()\n")},
				}}
				if got := li.interpretExpressionVars([]string{tc.token}, tc.token); got != tc.want {
					t.Errorf("rendered %q; want %q", got, tc.want)
				}
			})
		}
	}
}

func TestLucidRootMetadataPassCompatibility(t *testing.T) {
	for _, language := range []string{"Rust"} {
		t.Run(language, func(t *testing.T) {
			source := "[|Meta.Statement|]" + nestedRenderCondition("tail", true)
			meta := MetaDict{}
			meta.SetAdd("Key", "present", "String")
			meta.SetAdd("Statement", "[|Meta.Other|]", "String")
			meta.SetAdd("Other", "second-pass", "String")
			li := LucidInterpreter{RawContent: []byte(source), SubTags: RecursiveSubTagParser(source, &TagMeta{}), SegmentalSystem: &SegmentalSystem{Language: language}, Meta: meta}
			li.Interpret()
			if got, want := string(li.InterpretedContent), "[|Meta.Other|]tail"; got != want {
				t.Errorf("root metadata rendering = %q; want existing single global pass %q", got, want)
			}
		})
	}
}

func TestLucidResolvedNumericConditions(t *testing.T) {
	for _, language := range []string{"Rust"} {
		for _, tc := range []struct{ name, expression, want string }{
			{"nonempty_greater", "Length(Conjoinments.Calculations) > 0", "yes"},
			{"nonempty_greater_boundary", "Length(Conjoinments.Calculations) > 2", ""},
			{"nonempty_greater_equal", "Length(Conjoinments.Calculations) >= 2", "yes"},
			{"nonempty_lesser", "Length(Conjoinments.Calculations) < 3", "yes"},
			{"nonempty_lesser_equal", "Length(Conjoinments.Calculations) <= 2", "yes"},
			{"nonempty_lesser_equal_false", "Length(Conjoinments.Calculations) <= 1", ""},
			{"empty_greater", "Length(Conjoinments.Definitions) > 0", ""},
			{"empty_greater_equal", "Length(Conjoinments.Definitions) >= 0", "yes"},
			{"empty_lesser", "Length(Conjoinments.Definitions) < 1", "yes"},
			{"empty_lesser_equal", "Length(Conjoinments.Definitions) <= 0", "yes"},
			{"metadata_greater", "Meta.Count > 1", "yes"},
			{"metadata_lesser_equal", "Meta.Count <= 2", "yes"},
			{"computed_metadata_greater", "Length(Meta.Items) > 0", "yes"},
			{"computed_metadata_lesser_equal", "Length(Meta.Items) <= 2", "yes"},
			{"metadata_rhs", "Meta.Count >= Length(Conjoinments.Calculations)", "yes"},
			{"float_metadata", "Meta.Ratio > 1.25", "yes"},
			{"working_count", "Count >= 2", "yes"},
			{"working_float", "Ratio < 2", "yes"},
			{"working_with_computed_rhs", "Count <= Length(Meta.Items)", "yes"},
			{"numeric_string_not_coerced", "Meta.TextCount >= 1", ""},
		} {
			t.Run(language+"/"+tc.name, func(t *testing.T) {
				source := "(/={tags: Meta.Tags}=={if " + tc.expression + "}=={yes}=/)"
				meta := MetaDict{}
				meta.SetAdd("Count", "2", "Number")
				meta.SetAdd("Ratio", "1.5", "Number")
				meta.SetAdd("TextCount", "2", "String")
				meta.SetAdd("Items", `["first","second"]`, "PreparedStrings")
				li := LucidInterpreter{RawContent: []byte(source), SubTags: RecursiveSubTagParser(source, &TagMeta{}), SegmentalSystem: &SegmentalSystem{Language: language}, Meta: meta,
					Conjoinments:     []CompiledConjoinment{{Key: "Calculations", Type: "Calculations", Plural: true, Contents: [][]byte{[]byte("first"), []byte("second")}}, {Key: "Definitions", Type: "Definitions", Plural: true, Contents: [][]byte{}}},
					WorkingVariables: []*WorkingVariable{{VarName: "Count", VarType: "Number", VarValue: 2}, {VarName: "Ratio", VarType: "Number", VarValue: 1.5}},
				}
				before := marshalTestJSON(t, li.WorkingVariables)
				parsed := marshalTestJSON(t, li.SubTags)
				for run := 0; run < 2; run++ {
					li.Interpret()
					if got := string(li.InterpretedContent); got != tc.want {
						t.Errorf("run %d: rendered %q; want %q", run, got, tc.want)
					}
					after := marshalTestJSON(t, li.WorkingVariables)
					if string(before) != string(after) {
						t.Error("numeric comparison created or mutated a working binding")
					}
				}
				after := marshalTestJSON(t, li.SubTags)
				if string(parsed) != string(after) || string(li.RawContent) != source {
					t.Error("numeric comparison mutated parsed source")
				}
			})
		}
	}
}

func TestLucidResolvedProjectionLengthIgnoresNameCollision(t *testing.T) {
	for _, count := range []int{0, 1, 2} {
		t.Run(string(rune('0'+count)), func(t *testing.T) {
			source := "(/={tags: Meta.Tags}=={if Length(Conjoinments.Arguments) > 0}=={<}=/)" +
				nestedRenderLoop("Argument", "Conjoinments.Arguments", "[|Argument|];") +
				"(/={tags: Meta.Tags}=={if Length(Conjoinments.Arguments) >= 1}=={>}=/)" +
				"(/={tags: Meta.Tags}=={if Length(Conjoinments.Arguments) <= 0}=={empty}=/)"
			values := [][]byte{[]byte("T"), []byte("U")}[:count]
			li := LucidInterpreter{RawContent: []byte(source), SubTags: RecursiveSubTagParser(source, &TagMeta{}), SegmentalSystem: &SegmentalSystem{Language: "Rust"},
				Conjoinments:     []CompiledConjoinment{{Key: "Arguments", Type: "Arguments", Plural: true, Contents: values}},
				WorkingVariables: []*WorkingVariable{{VarName: "Length(Conjoinments.Arguments)", VarType: "Number", VarValue: -99}, {VarName: "Argument", VarType: "String", VarValue: "saved"}},
			}
			want := []string{"empty", "<T;>", "<T;U;>"}[count]
			before := marshalTestJSON(t, li.WorkingVariables)
			for run := 0; run < 2; run++ {
				li.Interpret()
				if got := string(li.InterpretedContent); got != want {
					t.Errorf("run %d: rendered %q; want %q", run, got, want)
				}
				after := marshalTestJSON(t, li.WorkingVariables)
				if string(before) != string(after) {
					t.Error("comparison rebound its expression-shaped working variable")
				}
			}
		})
	}
}

func TestLucidPreparedNativeVoidProvenance(t *testing.T) {
	ops, err := json.Marshal([]Op{{Prefix: "Void", Operands: []string{"Void"}, Postfix: "Void"}})
	if err != nil {
		t.Fatal(err)
	}
	for _, language := range []string{"Rust"} {
		t.Run(language, func(t *testing.T) {
			meta := MetaDict{}
			meta.SetAdd("Name", "Void", "SourceString")
			meta.SetAdd("Absent", "Void", "String")
			meta.SetAdd("Ops", string(ops), "PreparedOps")
			source := "[|Meta.Name|]/[|Meta.Ops[0].Prefix|]/[|Meta.Ops[0].Operands[0]|]/[|Meta.Ops[0].Postfix|]/[|Meta.Absent|]"
			li := LucidInterpreter{RawContent: []byte(source), Meta: meta, SegmentalSystem: &SegmentalSystem{Language: language}}
			li.Interpret()
			want := "Void/Void/Void/Void/"
			if string(li.InterpretedContent) != want {
				t.Fatalf("got %q; want %q", li.InterpretedContent, want)
			}
		})
	}
}
