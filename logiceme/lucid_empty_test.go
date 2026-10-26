package Logiceme

import (
	"bytes"
	"encoding/json"
	"errors"
	"testing"

	"github.com/l-programming-language/l/lerrors"
)

// These expectations are derived from the approved empty branches and the
// unchanged Count-loop literals. In particular, an empty record retains braces;
// it never becomes a unit struct or a bare constructor name.
func TestLucidEmptyInstantiatePrimitives(t *testing.T) {
	for _, c := range []struct {
		name, selector, collection, records, want string
	}{
		{"Y-empty", "Default", "Props", "&[]&", "Empty {}\n"},
		{"HY-empty", "Description", "Props", "&[]&", "\nstruct Empty {}\n"},
		{"MY-empty", "Instantiation", "Funcs", "&[]&", "\ntrait Empty {}\n"},
		{"Y-nonempty", "Default", "Props", "&[Prop{Key: field, Value: 7}Prop]&", "Empty{field: 7} \n"},
		{"HY-nonempty", "Description", "Props", "&[Prop{Key: field, Type: u8}Prop]&", "\n\n\nstruct Empty {\n\n\n\n    field: u8, // \n\n\n}\n\n\n"},
		{"MY-nonempty", "Instantiation", "Funcs", "&[Func{Key: method, Inputs: &[]&}Func]&", "\n\n\ntrait Empty {\n\n\nfn method();\n\n\n}\n\n\n"},
	} {
		t.Run(c.name, func(t *testing.T) {
			defer executionRegistryPanicAsFailure(t)
			ss, err := loadSegmentalSystem(runtimeGenerationResources(t))
			if err != nil {
				t.Fatal(err)
			}
			g := Glyph{SegmentalSystem: &ss}
			g.Meta.SetAdd("Key", "Empty", "String")
			g.Meta.SetAdd(c.collection, c.records, "Array")
			tags := g.FunctionalTags(ss.Instantiate.Content, c.selector)
			if len(tags) != 1 {
				t.Fatalf("Instantiate %s selected %d executable cases; want one", c.selector, len(tags))
			}
			// Interpret the actual selected asset body, retaining visible bindings
			// so repeated interpretation detects mutation and Count leakage.
			li := LucidInterpreter{
				RawContent: tags[0].GetTemplate(), SubTags: tags[0].GetSubTags(),
				Meta: g.Meta, SegmentalSystem: &ss,
				WorkingVariables: []*WorkingVariable{{VarName: "Saved", VarType: "String", VarValue: "untouched"}},
			}
			metadata := lucidEmptySnapshot(t, li.Meta)
			bindings := lucidEmptySnapshot(t, li.WorkingVariables)
			parsed := lucidEmptySnapshot(t, tags)
			source := bytes.Clone(li.RawContent)
			for run := 0; run < 2; run++ {
				li.Interpret()
				if got := string(li.InterpretedContent); got != c.want {
					t.Errorf("run %d rendered %q; want %q", run, got, c.want)
				}
				if !bytes.Equal(metadata, lucidEmptySnapshot(t, li.Meta)) || !bytes.Equal(bindings, lucidEmptySnapshot(t, li.WorkingVariables)) || !bytes.Equal(parsed, lucidEmptySnapshot(t, tags)) || !bytes.Equal(source, li.RawContent) {
					t.Error("interpretation mutated metadata, bindings, parsed tags or source")
				}
				if li.getVariable("Count").VarType != "NotFound" {
					t.Error("Count binding leaked from the template loop")
				}
			}
		})
	}
}

// Exercise the real merge boundary: file-owned metadata is inherited, explicit
// local metadata wins, and only an absent collection becomes canonical empty.
func TestLucidEmptyInstantiationPreparation(t *testing.T) {
	for _, form := range []struct{ name, collection, record, malformed string }{
		{"Y", "Props", "&[Prop{Key: field, Value: 7}Prop]&", "&[Prop{Key: field}Func]&"},
		{"HY", "Props", "&[Prop{Key: field, Type: u8}Prop]&", "&[Prop{Key: field}Func]&"},
		{"MY", "Funcs", "&[Func{Key: method, Inputs: &[]&}Func]&", "&[Func{Key: method}Prop]&"},
	} {
		for _, c := range []struct {
			name, inherited, local, want string
			invalid, canonical           bool
		}{
			{name: "missing", want: "&[]&", canonical: true},
			{name: "explicit-empty", local: "&[]&", want: "&[]&"},
			{name: "inherited-nonempty", inherited: form.record, want: form.record},
			{name: "local-empty-overrides-inherited", inherited: form.record, local: "&[]&", want: "&[]&"},
			{name: "local-nonempty-overrides-empty", inherited: "&[]&", local: form.record, want: form.record},
			{name: "malformed-local", local: form.malformed, invalid: true},
			{name: "malformed-inherited", inherited: form.malformed, invalid: true},
		} {
			t.Run(form.name+"/"+c.name, func(t *testing.T) {
				defer executionRegistryPanicAsFailure(t)
				ss, err := loadSegmentalSystem(runtimeGenerationResources(t))
				if err != nil {
					t.Fatal(err)
				}
				if c.inherited != "" {
					ss.fileMeta.SetAdd(form.collection, c.inherited, "Array")
				}
				inherited := lucidEmptySnapshot(t, ss.fileMeta)
				source := form.name + "{|Key: ~Empty~"
				if c.local != "" {
					source += ", " + form.collection + ": ~" + c.local + "~"
				}
				source += "|}"
				g, err := ss.initializeGlyphChecked([]byte(source), nil, nil)
				if err != nil {
					t.Fatal(err)
				}
				input := g.InputString
				var previous []byte
				for run := 0; run < 2; run++ {
					err := prepareGlyph(&g, glyphContext{callable: true, role: "expression"}, &generatedNames{used: source, counts: map[string]int{}})
					if c.invalid {
						if !errors.Is(err, lerrors.Syntax) {
							t.Errorf("run %d malformed %s metadata produced %v; want Syntax", run, form.collection, err)
						}
						if err != nil {
							cause := err
							for errors.Unwrap(cause) != nil {
								cause = errors.Unwrap(cause)
							}
							kind := "Prop"
							if form.collection == "Funcs" {
								kind = "Func"
							}
							want := "Rust " + form.name + ": malformed or unclosed expected " + kind + " record"
							if cause.Error() != want {
								t.Errorf("leaf cause %q; want %q", cause.Error(), want)
							}
						}
					} else {
						if err != nil {
							t.Fatal(err)
						}
						got := g.Meta.Get(form.collection)
						if got.Value != c.want {
							t.Errorf("run %d %s = %#v; want value %q", run, form.collection, got, c.want)
						}
						if c.canonical && got.Type != "Array" {
							t.Errorf("missing collection has type %q; want Array", got.Type)
						}
						// Preparation iterates local input maps; dictionary storage
						// order is not part of the effective metadata contract.
						// Compare keys, values and types, never generated bytes.
						semanticMeta := map[string]MetaObj{}
						for _, entry := range g.Meta.GetMetaData() {
							if _, exists := semanticMeta[entry.Key]; exists {
								t.Fatalf("duplicate effective metadata key %q", entry.Key)
							}
							semanticMeta[entry.Key] = entry
						}
						current := lucidEmptySnapshot(t, semanticMeta)
						if run > 0 && !bytes.Equal(previous, current) {
							t.Error("repeated preparation changed effective metadata")
						}
						previous = current
					}
					if !bytes.Equal(inherited, lucidEmptySnapshot(t, ss.fileMeta)) || g.InputString != input {
						t.Error("preparation changed file-owned metadata or supplied input")
					}
				}
			})
		}
	}
}

func lucidEmptySnapshot(t *testing.T, value interface{}) []byte {
	t.Helper()
	data, err := json.Marshal(value)
	if err != nil {
		t.Fatal(err)
	}
	return data
}
