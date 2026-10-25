package Logiceme

import (
	"bytes"
	"errors"
	"reflect"
	"testing"

	"github.com/l-programming-language/l/lerrors"
)

// Every prefix/tail below is independently authored from the retained template
// literals. The suffixes are explicit expected bytes, not a second template
// renderer, output capture, or replacement within generated source.
func TestLucidGenericPrimitives(t *testing.T) {
	profiles := []struct {
		name, selector, props, funcs, receiver, prefix, tail string
		operate, application                                 bool
	}{
		{"FW", "Definition", "&[]&", "&[]&", "Void", "fn sample", "( ) -> (u8) {\n    \n        body();\n    \n}\n", true, false},
		{"Y-empty", "Default", "&[]&", "&[]&", "Void", "Item", " {}\n", false, true},
		{"Y-single", "Default", "&[Prop{Key: x, Value: 1}Prop]&", "&[]&", "Void", "Item", "{x: 1} \n", false, true},
		{"Y-multiple", "Default", "&[Prop{Key: x, Value: 1}Prop, Prop{Key: y, Value: 2}Prop]&", "&[]&", "Void", "Item", "{x: 1, y: 2} \n", false, true},
		{"HY-empty", "Description", "&[]&", "&[]&", "Void", "\nstruct Item", " {}\n", false, false},
		{"HY-nonempty", "Description", "&[Prop{Key: x, Type: u8}Prop]&", "&[]&", "Void", "\n\n\nstruct Item", " {\n\n\n\n    x: u8, // \n\n\n}\n\n\n", false, false},
		{"MY-empty", "Instantiation", "&[]&", "&[]&", "Void", "\ntrait Item", " {}\n", false, false},
		{"MY-nonempty", "Instantiation", "&[]&", "&[Func{Key: method, Inputs: &[]&}Func]&", "Void", "\n\n\ntrait Item", " {\n\n\nfn method();\n\n\n}\n\n\n", false, false},
		{"KY-named", "Operation", "&[]&", "&[]&", "Void", "\n\ncall", "()\n", false, true},
		{"KY-receiver", "Operation", "&[]&", "&[]&", "receiver", "\nreceiver.call", "()\n\n", false, true},
	}
	modes := []struct {
		name, lifetimes, genericTypes, kind, declaration, application string
	}{
		{"absent", "", "", "String", "", ""},
		{"lifetimes", "'a", "", "String", "<'a>", "::<'a>"},
		{"types", "", "T", "String", "<T>", "::<T>"},
		{"both", "'a, 'b", "T, Vec<(u8, u16)>", "String", "<'a, 'b, T, Vec<(u8, u16)>>", "::<'a, 'b, T, Vec<(u8, u16)>>"},
		{"native-Void", "", "Void", "SourceString", "<Void>", "::<Void>"},
	}
	if len(profiles)*len(modes) != 50 {
		t.Fatal("generic primitive coverage must contain exactly 50 profiles")
	}
	for _, c := range profiles {
		for _, mode := range modes {
			t.Run(c.name+"/"+mode.name, func(t *testing.T) {
				defer executionRegistryPanicAsFailure(t)
				ss, err := loadSegmentalSystem(runtimeGenerationResources(t))
				if err != nil {
					t.Fatal(err)
				}
				g := Glyph{SegmentalSystem: &ss}
				key := "Item"
				if c.selector == "Operation" {
					key = "call"
				}
				g.Meta.SetAdd("Key", key, "String")
				g.Meta.SetAdd("Receiver", c.receiver, "String")
				g.Meta.SetAdd("Inputs", "&[]&", "Array")
				g.Meta.SetAdd("Props", c.props, "Array")
				g.Meta.SetAdd("Funcs", c.funcs, "Array")
				g.Meta.SetAdd("Func", "Func{Key: sample, Inputs: &[]&, Outputs: &[u8]&}Func", "Func")
				if mode.lifetimes != "" {
					g.Meta.SetAdd("Lifetimes", mode.lifetimes, "String")
				}
				if mode.genericTypes != "" {
					g.Meta.SetAdd("GenericTypes", mode.genericTypes, mode.kind)
				}
				asset := ss.Instantiate.Content
				if c.operate {
					asset = ss.Operate.Content
				}
				tags := g.FunctionalTags(asset, c.selector)
				if len(tags) != 1 {
					t.Fatalf("%s selected %d executable cases; want one", c.name, len(tags))
				}
				suffix := mode.declaration
				if c.application {
					suffix = mode.application
				}
				want := c.prefix + suffix + c.tail
				li := LucidInterpreter{
					RawContent: tags[0].GetTemplate(), SubTags: tags[0].GetSubTags(),
					Meta: g.Meta, SegmentalSystem: &ss,
					Conjoinments:     []CompiledConjoinment{{Key: "Operations", Type: "Operation", Plural: true, Contents: [][]byte{[]byte("body();")}}},
					WorkingVariables: []*WorkingVariable{{VarName: "Saved", VarType: "String", VarValue: "untouched"}},
				}
				metadata := lucidKeyedSnapshot(t, li.Meta)
				bindings := lucidKeyedSnapshot(t, li.WorkingVariables)
				parsed := lucidKeyedSnapshot(t, tags)
				children := lucidKeyedSnapshot(t, li.Conjoinments)
				source := bytes.Clone(li.RawContent)
				for run := 0; run < 2; run++ {
					li.Interpret()
					if got := string(li.InterpretedContent); got != want {
						t.Errorf("run %d rendered %q; want %q", run, got, want)
					}
					if !bytes.Equal(metadata, lucidKeyedSnapshot(t, li.Meta)) || !bytes.Equal(bindings, lucidKeyedSnapshot(t, li.WorkingVariables)) || !bytes.Equal(parsed, lucidKeyedSnapshot(t, tags)) || !bytes.Equal(children, lucidKeyedSnapshot(t, li.Conjoinments)) || !bytes.Equal(source, li.RawContent) {
						t.Error("interpretation mutated metadata, bindings, parsed tags, children or source")
					}
					if li.getVariable("Count").VarType != "NotFound" {
						t.Error("Count binding leaked from the template loop")
					}
				}
			})
		}
	}
}

// Missing generic inputs acquire absence semantics in Rust. Explicit
// SourceString/Void remains native text.
func TestLucidGenericLookup(t *testing.T) {
	for _, language := range []string{"Rust"} {
		for _, field := range []string{"Lifetimes", "GenericTypes"} {
			for _, c := range []struct{ name, value, kind string }{
				{"missing", "", ""},
				{"canonical-absence", "Void", "String"},
				{"native-Void", "Void", "SourceString"},
				{"ordinary", "T", "String"},
			} {
				t.Run(language+"/"+field+"/"+c.name, func(t *testing.T) {
					defer executionRegistryPanicAsFailure(t)
					meta := MetaDict{}
					if c.kind != "" {
						meta.SetAdd(field, c.value, c.kind)
					}
					li := LucidInterpreter{Meta: meta, SegmentalSystem: &SegmentalSystem{Language: language}}
					before := lucidKeyedSnapshot(t, meta)
					wantValue, wantKind := c.value, "String"
					if c.kind == "" {
						wantValue, wantKind = "Void", "String"
					}
					for run := 0; run < 2; run++ {
						value, kind := li.metaSwitch(field)
						if value != wantValue || kind != wantKind {
							t.Errorf("lookup = %#v/%s; want %q/%s", value, kind, wantValue, wantKind)
						}
						wantNative := c.kind == "SourceString"
						if got := li.explicitSourceString("Meta." + field); got != wantNative {
							t.Errorf("native provenance = %v; want %v", got, wantNative)
						}
						if language == "Rust" {
							wantAbsent := c.kind == "" || c.kind == "String" && c.value == "Void"
							if got := li.interpretCondition("if Meta." + field + " == Void"); (got == "true") != wantAbsent {
								t.Errorf("bare Void comparison = %s; want absence %v", got, wantAbsent)
							}
						}
						if !bytes.Equal(before, lucidKeyedSnapshot(t, li.Meta)) {
							t.Error("lookup mutated metadata")
						}
					}
				})
			}
		}
	}
	t.Run("other-missing-fields-unchanged", func(t *testing.T) {
		li := LucidInterpreter{SegmentalSystem: &SegmentalSystem{Language: "Rust"}}
		if value, kind := li.metaSwitch("Unrelated"); value != "NotFound" || kind != "NotFound" {
			t.Fatalf("unrelated missing lookup = %#v/%s", value, kind)
		}
	})
}

func lucidGenericSource(form, metadata string) string {
	base := map[string]string{
		"FW": "Key: ~sample~, Inputs: ~&[]&~, Outputs: ~&[()]&~",
		"HY": "Key: ~Item~, Props: ~&[]&~",
		"MY": "Key: ~Item~, Funcs: ~&[]&~",
		"Y":  "Key: ~Item~, Props: ~&[]&~",
		"KY": "Key: ~sample~, Inputs: ~&[]&~",
	}[form]
	if metadata != "" {
		base += ", " + metadata
	}
	source := form + "{|" + base + "|}"
	if form == "FW" {
		source += "^W"
	}
	return source
}

func lucidGenericLeaf(err error) error {
	for err != nil && errors.Unwrap(err) != nil {
		err = errors.Unwrap(err)
	}
	return err
}

func TestLucidGenericPreparation(t *testing.T) {
	for _, form := range []string{"FW", "HY", "MY", "Y", "KY"} {
		for _, c := range []struct{ name, metadata, lifetime, genericTypes, lifetimeKind, genericKind string }{
			{"absent", "", "Void", "Void", "String", "String"},
			{"supplied", "Lifetimes: ~'a, 'b~, GenericTypes: ~T, Vec<(u8, u16)>~", "'a, 'b", "T, Vec<(u8, u16)>", "String", "String"},
			{"native-Void", "GenericTypes: ~Void~", "Void", "Void", "String", "SourceString"},
			{"native-bytes", "Lifetimes: ~ 'a,\n'b ~, GenericTypes: ~ Vec <(u8, u16)>, fn(u8, u16) -> (u8, u16) ~", " 'a,\n'b ", " Vec <(u8, u16)>, fn(u8, u16) -> (u8, u16) ", "String", "String"},
		} {
			t.Run(form+"/"+c.name, func(t *testing.T) {
				defer executionRegistryPanicAsFailure(t)
				ss, err := loadSegmentalSystem(runtimeGenerationResources(t))
				if err != nil {
					t.Fatal(err)
				}
				source := lucidGenericSource(form, c.metadata)
				g, err := ss.initializeGlyphChecked([]byte(source), nil, nil)
				if err != nil {
					t.Fatal(err)
				}
				input := g.InputString
				fileMeta := lucidKeyedSnapshot(t, ss.fileMeta)
				var previous map[string]MetaObj
				for run := 0; run < 2; run++ {
					if err := prepareGlyph(&g, glyphContext{callable: true, role: "expression"}, &generatedNames{used: source, counts: map[string]int{}}); err != nil {
						t.Fatalf("run %d preparation: %v", run, err)
					}
					for _, want := range []MetaObj{{Key: "Lifetimes", Value: c.lifetime, Type: c.lifetimeKind}, {Key: "GenericTypes", Value: c.genericTypes, Type: c.genericKind}} {
						if got := g.Meta.Get(want.Key); *got != want {
							t.Errorf("run %d %s = %#v; want %#v", run, want.Key, *got, want)
						}
					}
					current := map[string]MetaObj{}
					for _, entry := range g.Meta.GetMetaData() {
						if _, exists := current[entry.Key]; exists {
							t.Fatalf("duplicate effective metadata %q", entry.Key)
						}
						current[entry.Key] = entry
					}
					if run > 0 && !reflect.DeepEqual(current, previous) {
						t.Error("repeated preparation changed effective metadata")
					}
					previous = current
					if g.InputString != input || g.Name != source || !bytes.Equal(fileMeta, lucidKeyedSnapshot(t, ss.fileMeta)) {
						t.Error("preparation mutated native input, source or file metadata")
					}
				}
			})
		}
	}
}

func TestLucidGenericBlankPreparation(t *testing.T) {
	for _, form := range []string{"FW", "HY", "MY", "Y", "KY"} {
		for _, field := range []string{"Lifetimes", "GenericTypes"} {
			for _, blank := range []struct{ name, value string }{{"empty", ""}, {"whitespace", " \t "}} {
				t.Run(form+"/"+field+"/"+blank.name, func(t *testing.T) {
					defer executionRegistryPanicAsFailure(t)
					ss, err := loadSegmentalSystem(runtimeGenerationResources(t))
					if err != nil {
						t.Fatal(err)
					}
					source := lucidGenericSource(form, field+": ~"+blank.value+"~")
					g, err := ss.initializeGlyphChecked([]byte(source), nil, nil)
					if err != nil {
						t.Fatal(err)
					}
					input := g.InputString
					for run := 0; run < 2; run++ {
						err := prepareGlyph(&g, glyphContext{callable: true, role: "expression"}, &generatedNames{used: source, counts: map[string]int{}})
						want := "Rust " + form + ": " + field + " must be nonempty when supplied"
						if !errors.Is(err, lerrors.InvalidInput) || lucidGenericLeaf(err) == nil || lucidGenericLeaf(err).Error() != want {
							t.Errorf("run %d error=%v; want InvalidInput with leaf %q", run, err, want)
						}
						if g.InputString != input || g.Name != source {
							t.Error("invalid preparation mutated input or source")
						}
					}
				})
			}
		}
	}
}

func TestLucidGenericInheritedPreparation(t *testing.T) {
	for _, field := range []string{"Lifetimes", "GenericTypes"} {
		for _, inherited := range []struct{ name, value, kind string }{
			{"value", "T", "String"}, {"empty", "", "String"}, {"whitespace", " \t ", "String"},
			{"absence-sentinel", "Void", "String"}, {"native-Void", "Void", "SourceString"},
		} {
			for _, override := range []bool{false, true} {
				name := field + "/" + inherited.name
				if override {
					name += "/local-override"
				}
				t.Run(name, func(t *testing.T) {
					defer executionRegistryPanicAsFailure(t)
					ss, err := loadSegmentalSystem(runtimeGenerationResources(t))
					if err != nil {
						t.Fatal(err)
					}
					ss.fileMeta.SetAdd(field, inherited.value, inherited.kind)
					before := lucidKeyedSnapshot(t, ss.fileMeta)
					local := ""
					if override {
						local = field + ": ~local~"
					}
					source := lucidGenericSource("FW", local)
					g, err := ss.initializeGlyphChecked([]byte(source), nil, nil)
					if err != nil {
						t.Fatal(err)
					}
					input := g.InputString
					for run := 0; run < 2; run++ {
						err := prepareGlyph(&g, glyphContext{callable: true, role: "expression"}, &generatedNames{used: source, counts: map[string]int{}})
						want := "Rust " + field + " is glyph-local; inherited file metadata is unsupported"
						if !errors.Is(err, lerrors.InvalidInput) || lucidGenericLeaf(err) == nil || lucidGenericLeaf(err).Error() != want {
							t.Errorf("run %d error=%v; want InvalidInput with leaf %q", run, err, want)
						}
						if !bytes.Equal(before, lucidKeyedSnapshot(t, ss.fileMeta)) || g.InputString != input || g.Name != source {
							t.Error("inherited rejection mutated file metadata, input or source")
						}
					}
				})
			}
		}
	}
}

func TestLucidGenericChildIsolation(t *testing.T) {
	ss, err := loadSegmentalSystem(runtimeGenerationResources(t))
	if err != nil {
		t.Fatal(err)
	}
	source := "FW{|Key: ~sample~, Lifetimes: ~'a~, GenericTypes: ~T~|}^W^BOM^KY{|Key: ~consume~, Inputs: ~&[]&~|}"
	g, err := ss.initializeGlyphChecked([]byte(source), nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	input := g.InputString
	for run := 0; run < 2; run++ {
		if err := prepareGlyph(&g, glyphContext{}, &generatedNames{used: source, counts: map[string]int{}}); err != nil {
			t.Fatal(err)
		}
		child := &g
		for _, form := range []string{"W", "BOM", "KY"} {
			if len(child.Conjoinments) != 1 {
				t.Fatalf("expected one %s descendant", form)
			}
			child = &child.Conjoinments[0]
			if glyphForm(child) != form {
				t.Fatalf("descendant form %q; want %q", glyphForm(child), form)
			}
			for _, field := range []string{"Lifetimes", "GenericTypes"} {
				if got := child.Meta.Get(field); got.Type != "String" || got.Value != "Void" {
					t.Errorf("parent parameter leaked to %s: %s=%#v", form, field, got)
				}
			}
		}
		if g.InputString != input || g.Name != source {
			t.Error("recursive preparation mutated input or source")
		}
	}
}
