package Logiceme

import (
	"bytes"
	"encoding/json"
	"testing"
)

// Func remains authoritative when a root Key happens to contain different
// native source. Only its nonempty Key gains Rust name provenance; empty
// signature fields retain their existing, target-specific absence behavior.
func TestLucidFuncKeyProvenance(t *testing.T) {
	cases := []struct {
		name, record, rootKey, field string
		nativeText                   string
		nativeBare                   bool
		nativeQuoted                 bool
	}{
		{"native-void", "#[Func{Key: Void}Func]#", "", "Key", "Void", false, true},
		{"ordinary-name", "#[Func{Key: helper}Func]#", "", "Key", "helper", false, false},
		{"void-record-over-root", "#[Func{Key: Void}Func]#", "different", "Key", "Void", false, true},
		{"ordinary-record-over-void-root", "#[Func{Key: helper}Func]#", "Void", "Key", "helper", false, false},
		// An unset dictionary entry currently falls back to an empty Func;
		// the explicit prepared absence sentinel remains distinct.
		{"missing-func-over-root", "", "Void", "Key", "", false, false},
		{"absent-func-sentinel", "Void", "Void", "Key", "", true, true},
		{"empty-key-over-root", "#[Func{}Func]#", "Void", "Key", "", false, false},
		{"absent-receiver", "#[Func{Key: Void}Func]#", "Void", "Receiver~", "", true, true},
		{"empty-inputs", "#[Func{Key: Void}Func]#", "Void", "Inputs~", "", true, true},
		{"empty-outputs", "#[Func{Key: Void}Func]#", "Void", "Outputs~", "", true, true},
		{"explicit-empty-inputs", "#[Func{Key: Void, Inputs: &[]&, Outputs: &[]&}Func]#", "Void", "Inputs~", "", true, true},
		{"explicit-empty-outputs", "#[Func{Key: Void, Inputs: &[]&, Outputs: &[]&}Func]#", "Void", "Outputs~", "", true, true},
	}
	for _, language := range []string{"Rust"} {
		for _, c := range cases {
			t.Run(language+"/"+c.name, func(t *testing.T) {
				text, bare, quoted := c.nativeText, c.nativeBare, c.nativeQuoted
				variable := "Meta.Func." + c.field
				t.Run("interpolation", func(t *testing.T) {
					lucidFuncRepeatedInterpretation(t, language, c.record, c.rootKey,
						"before<[|"+variable+"|]>after", "before<"+text+">after")
				})
				for _, comparison := range []struct {
					name, operator, value string
					want                  bool
				}{
					{"bare-equal", "==", "Void", bare},
					{"bare-unequal", "!=", "Void", !bare},
					{"quoted-equal", "==", `"Void"`, quoted},
					{"quoted-unequal", "!=", `"Void"`, !quoted},
				} {
					t.Run(comparison.name, func(t *testing.T) {
						source := "before(/={tags: Meta.Tags}=={if " + variable + " " + comparison.operator + " " + comparison.value + "}=={yes}=/)after"
						want := "beforeafter"
						if comparison.want {
							want = "beforeyesafter"
						}
						lucidFuncRepeatedInterpretation(t, language, c.record, c.rootKey, source, want)
					})
				}
			})
		}
	}
}

// Joined return types are native Rust source only when the Func actually has
// outputs. Keep the empty signature controls explicit.
func TestLucidFuncOutputProvenance(t *testing.T) {
	type expected struct {
		joined, parenthesized string
		bareVoid, quotedVoid  bool
		outputsPresent        bool
	}
	cases := []struct {
		name, record string
		nativeResult expected
	}{
		{
			"native-void", "#[Func{Key: helper, Outputs: &[Void]&}Func]#",
			expected{"Void", "(Void)", false, true, true},
		},
		{
			"ordinary-output", "#[Func{Key: helper, Outputs: &[u8]&}Func]#",
			expected{"u8", "(u8)", false, false, true},
		},
		{
			"explicit-unit", "#[Func{Key: helper, Outputs: &[()]&}Func]#",
			expected{"()", "(())", false, false, true},
		},
		{
			"multiple-outputs", "#[Func{Key: helper, Outputs: &[Void, u8]&}Func]#",
			expected{"Void, u8", "(Void, u8)", false, false, true},
		},
		{
			"missing-outputs", "#[Func{Key: helper}Func]#",
			expected{"", "", true, true, false},
		},
		{
			"explicit-empty-outputs", "#[Func{Key: helper, Outputs: &[]&}Func]#",
			expected{"", "", true, true, false},
		},
		{
			"missing-func", "",
			expected{"", "", true, true, false},
		},
		{
			"absent-func-sentinel", "Void",
			expected{"", "", true, true, false},
		},
		{
			"empty-func-record", "#[Func{}Func]#",
			expected{"", "", true, true, false},
		},
	}
	for _, language := range []string{"Rust"} {
		for _, c := range cases {
			t.Run(language+"/"+c.name, func(t *testing.T) {
				want := c.nativeResult
				for _, interpolation := range []struct{ name, source, text string }{
					{"joined", "[|Meta.Func.Outputs~|]", want.joined},
					{"parenthesized", "[|(Meta.Func.Outputs~)|]", want.parenthesized},
				} {
					t.Run(interpolation.name, func(t *testing.T) {
						lucidFuncRepeatedInterpretation(t, language, c.record, "",
							"before<"+interpolation.source+">after", "before<"+interpolation.text+">after")
					})
				}
				for _, comparison := range []struct {
					name, field, operator, value string
					result                       bool
				}{
					{"bare-equal", "Outputs~", "==", "Void", want.bareVoid},
					{"bare-unequal", "Outputs~", "!=", "Void", !want.bareVoid},
					{"quoted-equal", "Outputs~", "==", `"Void"`, want.quotedVoid},
					{"quoted-unequal", "Outputs~", "!=", `"Void"`, !want.quotedVoid},
					{"output-list-absent", "Outputs", "==", "Void", !want.outputsPresent},
					{"output-list-present", "Outputs", "!=", "Void", want.outputsPresent},
				} {
					t.Run(comparison.name, func(t *testing.T) {
						source := "before(/={tags: Meta.Tags}=={if Meta.Func." + comparison.field + " " + comparison.operator + " " + comparison.value + "}=={yes}=/)after"
						text := "beforeafter"
						if comparison.result {
							text = "beforeyesafter"
						}
						lucidFuncRepeatedInterpretation(t, language, c.record, "", source, text)
					})
				}
			})
		}
	}
}

func lucidFuncRepeatedInterpretation(t *testing.T, language, record, rootKey, source, want string) {
	t.Helper()
	meta := MetaDict{}
	if record != "" {
		meta.SetAdd("Func", record, "Hashmap")
	}
	if rootKey != "" {
		meta.SetAdd("Key", rootKey, "SourceString")
	}
	li := LucidInterpreter{
		RawContent: []byte(source), SubTags: RecursiveSubTagParser(source, &TagMeta{}),
		Meta: meta, SegmentalSystem: &SegmentalSystem{Language: language},
		WorkingVariables: []*WorkingVariable{{VarName: "Saved", VarType: "String", VarValue: "unchanged"}},
	}
	snapshot := func(value interface{}) []byte {
		t.Helper()
		encoded, err := json.Marshal(value)
		if err != nil {
			t.Fatal(err)
		}
		return encoded
	}
	metadata := snapshot(li.Meta)
	bindings := snapshot(li.WorkingVariables)
	parsed := snapshot(li.SubTags)
	for run := 0; run < 2; run++ {
		li.Interpret()
		if got := string(li.InterpretedContent); got != want {
			t.Errorf("run %d: rendered %q; want %q", run, got, want)
		}
		if !bytes.Equal(metadata, snapshot(li.Meta)) || !bytes.Equal(bindings, snapshot(li.WorkingVariables)) || !bytes.Equal(parsed, snapshot(li.SubTags)) || string(li.RawContent) != source {
			t.Error("interpretation mutated metadata, working bindings, raw source or parsed tags")
		}
	}
}
