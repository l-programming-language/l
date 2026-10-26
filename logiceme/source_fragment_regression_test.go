package Logiceme

import (
	"bytes"
	"reflect"
	"testing"
)

func TestExecutionOperandFramingRegression(t *testing.T) {
	for _, tc := range []struct {
		name, input string
		want        []string
	}{
		{"indexed_receive", "&[channels[0]]&", []string{"channels[0]"}},
		{"indexed_send", "&[ch, values[0]]&", []string{"ch", "values[0]"}},
		{"nested_indices", "&[matrix[0][1]]&", []string{"matrix[0][1]"}},
		{"slice_literal_index", "&[[]int{1}[0]]&", []string{"[]int{1}[0]"}},
		{"leading_ampersand", "&[&items[0]]&", []string{"&items[0]"}},
		{"marked_bracket_payload", "&[[channels[0]]]&", []string{"[channels[0]]"}},
		{"plain", "[ch]", []string{"ch"}},
		{"index", "[channels[0]]", []string{"channels[0]"}},
		{"unframed", "ch", []string{"ch"}},
		{"marked_empty", "&[]&", []string{""}},
		{"empty", "[]", []string{""}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			defer executionRegistryPanicAsFailure(t)
			meta := MetaObj{Type: "Struct", Value: "Op{Prefix: &[prefix]&, Type: &[Calculate]&, Operator: &[<-]&, Operands: " + tc.input + ", Postfix: &[postfix]&}Op"}
			before := meta
			got := LoadOp(&meta)
			want := Op{Prefix: "prefix", Type: "Calculate", Operator: "<-", Operands: tc.want, Postfix: "postfix"}
			if !reflect.DeepEqual(got, want) {
				t.Errorf("LoadOp operands %q = %#v; want %#v", tc.input, got, want)
			}
			if meta != before {
				t.Fatal("LoadOp mutated its metadata input")
			}
		})
	}
}

func TestExecutionParseStructFramingRegression(t *testing.T) {
	for _, tc := range []struct{ name, input, want string }{
		{"index", "&[channels[0]]&", "channels[0]"},
		{"slice_literal", "&[[]int{1}[0]]&", "[]int{1}[0]"},
		{"ampersand", "&[&items[0]]&", "&items[0]"},
		{"one_wrapper", "&[[channels[0]]]&", "[channels[0]]"},
		{"bare_brackets", "[channels[0]]", "[channels[0]]"},
		{"embedded_wrapper", "prefix&[items[0]]&suffix", "prefix&[items[0]]&suffix"},
		{"empty_wrapper", "&[]&", ""},
		{"plain", "Calculate", "Calculate"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			defer executionRegistryPanicAsFailure(t)
			got := ParseStruct("Prop{Value: " + tc.input + "}Prop")
			want := [][]string{{"Value", tc.want}}
			if !reflect.DeepEqual(got, want) {
				t.Fatalf("ParseStruct(%q) = %q; want %q", tc.input, got, want)
			}
		})
	}
}

// Exercise each existing interpolation entry point, rather than a cleanup helper,
// so losing the target language while forwarding a conjoinment is observable.
func TestSourceFragmentInterpolationRegression(t *testing.T) {
	for _, tc := range []struct {
		name, marker, source, want string
		system                     *SegmentalSystem
	}{
		{"rust", "%", "\nfirst\n\nlast\n", "firstlast", &SegmentalSystem{Language: "Rust"}},
		{"unspecified_language", "%", "\nfirst\n\nlast\n", "firstlast", &SegmentalSystem{}},
		{"nil_system", "%", "\nfirst\n\nlast\n", "firstlast", nil},
	} {
		for _, route := range []string{"singular", "indexed"} {
			t.Run(tc.name+"/"+route, func(t *testing.T) {
				defer executionRegistryPanicAsFailure(t)
				source := []byte(tc.source)
				before := append([]byte(nil), source...)
				var got string
				if route == "singular" {
					placeholder := "[|" + tc.marker + "Conjoinments.Calculation" + tc.marker + "|]"
					tag := LucidTag{Expressions: []*TagExpression{{}, {}, {ExpressionString: placeholder}}}
					compilation := DataCompilation{SegmentalSystem: tc.system, Conjoinments: []CompiledConjoinment{
						{Key: "Calculation", Type: "Calculation", Contents: [][]byte{source}},
					}}
					got = string(tag.Interpret(compilation))
					if tag.Expressions[2].ExpressionString != placeholder {
						t.Error("singular interpolation mutated its template")
					}
				} else {
					placeholder := "[|" + tc.marker + "Conjoinments.Calculations[1]" + tc.marker + "|]"
					interpreter := LucidInterpreter{SegmentalSystem: tc.system, Conjoinments: []CompiledConjoinment{
						{Type: "Calculations", Plural: true, Contents: [][]byte{[]byte("unselected()"), source}},
					}}
					got = interpreter.interpretExpressionVars([]string{placeholder}, placeholder)
				}
				if got != tc.want {
					t.Errorf("interpolated source = %q; want %q", got, tc.want)
				}
				if !bytes.Equal(source, before) {
					t.Errorf("interpolation mutated conjoinment bytes: got %q; want %q", source, before)
				}
			})
		}
	}
}

func TestSourceFragmentPublicInterpolation(t *testing.T) {
	// The public API has no target language; its percent behavior stays
	// generic regardless of the opaque source bytes.
	for _, tc := range []struct{ name, marker, want string }{
		{"percent", "%", "`firstlast`"},
		{"ordinary", "", "\n`first\n\nlast`\n"},
		{"leading_only", "&", "`first\n\nlast`\n"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			defer executionRegistryPanicAsFailure(t)
			source := []byte("\n`first\n\nlast`\n")
			before := append([]byte(nil), source...)
			template := []byte("[|" + tc.marker + "Conjoinments.Calculation" + tc.marker + "|]")
			templateBefore := append([]byte(nil), template...)
			got := InterpretConjoinments(template, []CompiledConjoinment{
				{Key: "Calculation", Type: "Calculation", Contents: [][]byte{source}},
			}, MetaDict{})
			if string(got) != tc.want {
				t.Errorf("InterpretConjoinments = %q; want %q", got, tc.want)
			}
			if !bytes.Equal(source, before) || !bytes.Equal(template, templateBefore) {
				t.Fatal("public interpolation mutated input bytes")
			}
		})
	}
}
