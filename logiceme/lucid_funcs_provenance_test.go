package Logiceme

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
)

// The indexed record must remain authoritative, including native source named
// Void.
func TestLucidIndexedFuncsProvenance(t *testing.T) {
	type expected struct {
		joined, parenthesized         string
		bareVoid, quotedVoid, present bool
	}
	cases := []struct {
		name, record string
		nativeResult expected
	}{
		{"missing", "Func{Key: missing}Func", expected{"", "", true, true, false}},
		{"empty", "Func{Key: empty, Outputs: &[]&}Func", expected{"", "", true, true, false}},
		{"unit", "Func{Key: unit, Outputs: &[()]&}Func", expected{"()", "(())", false, false, true}},
		{"scalar", "Func{Key: scalar, Outputs: &[u8]&}Func", expected{"u8", "(u8)", false, false, true}},
		{"multiple", "Func{Key: multiple, Outputs: &[Void, u8]&}Func", expected{"Void, u8", "(Void, u8)", false, false, true}},
		{"native-void", "Func{Key: Void, Outputs: &[Void]&}Func", expected{"Void", "(Void)", false, true, true}},
	}
	for _, language := range []string{"Rust"} {
		for _, c := range cases {
			t.Run(language+"/"+c.name, func(t *testing.T) {
				want := c.nativeResult
				// Index 1 avoids accidentally treating the first record as the
				// selected record and ensures root/singular fields cannot win.
				records := "&[Func{Key: decoy, Outputs: &[bool]&}Func, " + c.record + "]&"
				for _, interpolation := range []struct{ name, source, text string }{
					{"joined", "[|Meta.Funcs[1].Outputs~|]", want.joined},
					{"parenthesized", "[|(Meta.Funcs[1].Outputs~)|]", want.parenthesized},
				} {
					t.Run(interpolation.name, func(t *testing.T) {
						lucidFuncsRepeatedInterpretation(t, language, records, "before<"+interpolation.source+">after", "before<"+interpolation.text+">after")
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
					{"list-present", "Outputs", "!=", "Void", want.present},
					{"list-absent", "Outputs", "==", "Void", !want.present},
				} {
					t.Run(comparison.name, func(t *testing.T) {
						source := "before(/={tags: Meta.Tags}=={if Meta.Funcs[1]." + comparison.field + " " + comparison.operator + " " + comparison.value + "}=={yes}=/)after"
						text := "beforeafter"
						if comparison.result {
							text = "beforeyesafter"
						}
						lucidFuncsRepeatedInterpretation(t, language, records, source, text)
					})
				}
			})
		}
		for _, name := range []string{"helper", "Void"} {
			t.Run(language+"/name-"+name, func(t *testing.T) {
				records := "&[Func{Key: decoy}Func, Func{Key: " + name + "}Func]&"
				text := name
				lucidFuncsRepeatedInterpretation(t, language, records, "before<[|Meta.Funcs[1].Key|]>after", "before<"+text+">after")
				for _, comparison := range []struct {
					name, operator, value string
					result                bool
				}{
					{"bare-equal", "==", "Void", false},
					{"bare-unequal", "!=", "Void", true},
					{"quoted-equal", "==", `"Void"`, name == "Void"},
					{"quoted-unequal", "!=", `"Void"`, name != "Void"},
				} {
					t.Run(comparison.name, func(t *testing.T) {
						source := "before(/={tags: Meta.Tags}=={if Meta.Funcs[1].Key " + comparison.operator + " " + comparison.value + "}=={yes}=/)after"
						want := "beforeafter"
						if comparison.result {
							want = "beforeyesafter"
						}
						lucidFuncsRepeatedInterpretation(t, language, records, source, want)
					})
				}
			})
		}
	}
}

func TestLucidIndexedFuncsLoop(t *testing.T) {
	records := "&[Func{Key: missing}Func, Func{Key: empty, Outputs: &[]&}Func, Func{Key: unit, Outputs: &[()]&}Func, Func{Key: scalar, Outputs: &[u8]&}Func, Func{Key: multiple, Outputs: &[Void, u8]&}Func, Func{Key: Void, Outputs: &[Void]&}Func]&"
	// This is MY's actual Count/Add indexing and optional output presence
	// condition, with visible separators independently owned by the template.
	source := "prefix(/={tags: Meta.Tags}=={for Count in Count(Meta.Funcs)}=={[|Meta.Funcs[Add(Count, -1)].Key|](/={tags: Meta.Tags}=={if Meta.Funcs[Add(Count, -1)].Outputs != Void}=={ -> [|(Meta.Funcs[Add(Count, -1)].Outputs~)|]}=/); }=/)suffix"
	for _, c := range []struct{ language, want string }{
		{"Rust", "prefixmissing; empty; unit -> (()); scalar -> (u8); multiple -> (Void, u8); Void -> (Void); suffix"},
	} {
		t.Run(c.language, func(t *testing.T) { lucidFuncsRepeatedInterpretation(t, c.language, records, source, c.want) })
		for _, empty := range []struct{ name, records string }{{"missing-collection", ""}, {"empty-collection", "&[]&"}} {
			t.Run(c.language+"/"+empty.name, func(t *testing.T) {
				lucidFuncsRepeatedInterpretation(t, c.language, empty.records, source, "prefixsuffix")
			})
		}
	}
}

func TestLucidIndexedFuncsParameters(t *testing.T) {
	for _, c := range []struct{ language, want string }{
		{"Rust", "a: u8, b: bool"},
	} {
		t.Run(c.language+"/simple-controls", func(t *testing.T) {
			records := "&[Func{Key: method, Inputs: &[Prop{Key: a, Type: u8}Prop, Prop{Key: b, Type: bool}Prop]&}Func]&"
			lucidFuncsRepeatedInterpretation(t, c.language, records, "before<[|Meta.Funcs[0].Inputs~|]>after", "before<"+c.want+">after")
		})
	}
	t.Run("Rust/nested-commas", func(t *testing.T) {
		records := "&[Func{Key: nested, Inputs: &[Prop{Key: pair, Type: (u8, u16)}Prop, Prop{Key: map, Type: std::collections::HashMap<u8, Vec<(u16, u32)>>}Prop, Prop{Key: callback, Type: fn(u8, u16) -> (u32, bool)}Prop]&, Outputs: &[Result<(u8, u16), ()>, (u32, bool)]&}Func]&"
		source := "[|Meta.Funcs[0].Key|]([|Meta.Funcs[0].Inputs~|]) -> [|(Meta.Funcs[0].Outputs~)|]"
		want := "nested(pair: (u8, u16), map: std::collections::HashMap<u8, Vec<(u16, u32)>>, callback: fn(u8, u16) -> (u32, bool)) -> (Result<(u8, u16), ()>, (u32, bool))"
		lucidFuncsRepeatedInterpretation(t, "Rust", records, source, want)
	})
}

func TestLucidIndexedFuncsProvenanceBounds(t *testing.T) {
	for _, language := range []string{"Rust"} {
		for _, c := range []struct {
			field   string
			present bool
		}{
			{"Meta.Funcs[0].Key", true}, {"Meta.Funcs[0].Outputs~", true},
			{"Meta.Funcs[1].Key", false}, {"Meta.Funcs[1].Outputs~", false},
			{"Meta.Funcs[2].Key", true}, {"Meta.Funcs[2].Outputs~", false},
			{"Meta.Funcs[-1].Key", false}, {"Meta.Funcs[3].Key", false},
			{"Meta.Funcs[3].Outputs~", false}, {"Meta.Funcs[0].Outputs", false},
			{"Meta.Funcs[0].Inputs~", false}, {"Meta.Funcs[0].Receiver~", false},
		} {
			t.Run(language+"/"+c.field, func(t *testing.T) {
				defer executionRegistryPanicAsFailure(t)
				meta := MetaDict{}
				meta.SetAdd("Funcs", "&[Func{Key: Void, Outputs: &[Void]&}Func, Func{}Func, Func{Key: empty, Outputs: &[]&}Func]&", "Array")
				li := LucidInterpreter{Meta: meta, SegmentalSystem: &SegmentalSystem{Language: language}}
				want := c.present
				if got := li.explicitSourceString(c.field); got != want {
					t.Fatalf("provenance = %t, want %t", got, want)
				}
			})
		}
	}
}

func TestLucidFuncsMalformedDecode(t *testing.T) {
	for _, c := range []struct{ name, record string }{
		{"wrong-record-kind", "&[Prop{Key: wrong}Prop]&"},
		{"malformed-record", "&[Func{Key: broken}Prop]&"},
		{"malformed-input", "&[Func{Key: broken, Inputs: &[Prop{Key: value, Type: u8}Func]&}Func]&"},
		{"no-partial-records", "&[Func{Key: valid}Func, Prop{Key: wrong}Prop]&"},
	} {
		t.Run(c.name, func(t *testing.T) {
			defer executionRegistryPanicAsFailure(t)
			meta := MetaDict{}
			meta.SetAdd("Funcs", c.record, "Array")
			li := LucidInterpreter{Meta: meta, SegmentalSystem: &SegmentalSystem{Language: "Rust"}}
			for run := 0; run < 2; run++ {
				value, kind := li.metaSwitch("Funcs")
				functions, ok := value.([]Func)
				if kind != "Funcs" || !ok || len(functions) != 0 {
					t.Errorf("run %d decoded %T %#v (%s), want no Func records", run, value, value, kind)
				}
				if li.Meta.Get("Funcs").Value != c.record {
					t.Fatal("decode mutated Funcs metadata")
				}
			}
		})
	}
}

func lucidFuncsRepeatedInterpretation(t *testing.T, language, records, source, want string) {
	t.Helper()
	defer executionRegistryPanicAsFailure(t)
	meta := MetaDict{}
	if records != "" {
		meta.SetAdd("Funcs", records, "Array")
	}
	meta.SetAdd("Func", "#[Func{Key: singular-decoy, Outputs: &[bool]&}Func]#", "Hashmap")
	meta.SetAdd("Key", "root-decoy", "SourceString")
	li := LucidInterpreter{
		RawContent: []byte(source), SubTags: RecursiveSubTagParser(source, &TagMeta{}),
		Meta: meta, SegmentalSystem: &SegmentalSystem{Language: language},
		WorkingVariables: []*WorkingVariable{{VarName: "Saved", VarType: "String", VarValue: "unchanged"}},
	}
	snapshot := func(value interface{}) []byte {
		t.Helper()
		data, err := json.Marshal(value)
		if err != nil {
			t.Fatal(err)
		}
		return data
	}
	metadata, bindings, parsed := snapshot(li.Meta), snapshot(li.WorkingVariables), snapshot(li.SubTags)
	for run := 0; run < 2; run++ {
		li.Interpret()
		if got := string(li.InterpretedContent); got != want {
			t.Errorf("run %d rendered %q; want %q", run, got, want)
		}
		if !bytes.Equal(metadata, snapshot(li.Meta)) || !bytes.Equal(bindings, snapshot(li.WorkingVariables)) || !bytes.Equal(parsed, snapshot(li.SubTags)) || string(li.RawContent) != source {
			t.Error("interpretation mutated metadata, bindings, parsed tags or raw source")
		}
		if strings.Contains(source, "Count(Meta.Funcs)") && li.getVariable("Count").VarType != "NotFound" {
			t.Error("loop binding leaked")
		}
	}
}
