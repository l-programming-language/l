package Logiceme

import (
	"bytes"
	"encoding/json"
	"errors"
	"testing"

	"github.com/l-programming-language/l/lerrors"
)

// The expected literal spans below are authored from the reviewed keyed branch:
// one match shell, one Count body per pair, and only explicit Void carriers
// selecting the wildcard arm. They are not captured from generator output.
func TestLucidKeyedEvaluatePrimitives(t *testing.T) {
	for _, c := range []struct {
		name, key, keyType, want string
		patterns, bodies         []string
	}{
		{
			name: "single-binding", key: "subject", keyType: "String",
			patterns: []string{"value"}, bodies: []string{"use_value(value);"},
			want: "\n\n\nmatch subject {\n" +
				"\n\n    value => {\n        use_value(value);\n    },\n\n\n" +
				"\n}\n\n",
		},
		{
			name: "ordered-final-binding", key: "subject", keyType: "String",
			patterns: []string{"0", "last"}, bodies: []string{"first();", "use_last(last);"},
			want: "\n\n\nmatch subject {\n" +
				"\n\n    0 => {\n        first();\n    },\n\n\n" +
				"\n\n    last => {\n        use_last(last);\n    },\n\n\n" +
				"\n}\n\n",
		},
		{
			name: "fallback-only", key: "subject", keyType: "String",
			patterns: []string{"Void"}, bodies: []string{"fallback();"},
			want: "\n\n\nmatch subject {\n" +
				"\n\n\n    _ => {\n        fallback();\n    },\n\n" +
				"\n}\n\n",
		},
		{
			name: "terminal-fallback", key: "subject", keyType: "String",
			patterns: []string{"Choice::Void", "Void"}, bodies: []string{"explicit_variant();", "fallback();"},
			want: "\n\n\nmatch subject {\n" +
				"\n\n    Choice::Void => {\n        explicit_variant();\n    },\n\n\n" +
				"\n\n\n    _ => {\n        fallback();\n    },\n\n" +
				"\n}\n\n",
		},
		{
			name: "native-void-subject", key: "Void", keyType: "SourceString",
			patterns: []string{"_"}, bodies: []string{"supplied_wildcard();"},
			want: "\n\n\nmatch Void {\n" +
				"\n\n    _ => {\n        supplied_wildcard();\n    },\n\n\n" +
				"\n}\n\n",
		},
	} {
		t.Run(c.name, func(t *testing.T) {
			defer executionRegistryPanicAsFailure(t)
			ss, err := loadSegmentalSystem(runtimeGenerationResources(t))
			if err != nil {
				t.Fatal(err)
			}
			g := Glyph{SegmentalSystem: &ss}
			g.Meta.SetAdd("Key", c.key, c.keyType)
			tags := g.FunctionalTags(ss.Evaluate.Content, "Evaluation")
			if len(tags) != 1 {
				t.Fatalf("Evaluate Evaluation selected %d executable cases; want one", len(tags))
			}
			calculations := CompiledConjoinment{Key: "Calculations", Type: "Calculation", Plural: true}
			operations := CompiledConjoinment{Key: "Operations", Type: "Operation", Plural: true}
			for _, pattern := range c.patterns {
				calculations.Contents = append(calculations.Contents, []byte(pattern))
			}
			for _, body := range c.bodies {
				operations.Contents = append(operations.Contents, []byte(body))
			}
			li := LucidInterpreter{
				RawContent: tags[0].GetTemplate(), SubTags: tags[0].GetSubTags(),
				Meta: g.Meta, SegmentalSystem: &ss,
				Conjoinments:     []CompiledConjoinment{calculations, operations},
				WorkingVariables: []*WorkingVariable{{VarName: "Saved", VarType: "String", VarValue: "untouched"}},
			}
			metadata := lucidKeyedSnapshot(t, li.Meta)
			bindings := lucidKeyedSnapshot(t, li.WorkingVariables)
			parsed := lucidKeyedSnapshot(t, tags)
			children := lucidKeyedSnapshot(t, li.Conjoinments)
			source := bytes.Clone(li.RawContent)
			for run := 0; run < 2; run++ {
				li.Interpret()
				if got := string(li.InterpretedContent); got != c.want {
					t.Errorf("run %d rendered %q; want %q", run, got, c.want)
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

// Effective file metadata and local overrides must make the same keyed/unkeyed
// decision as the template. The absent String/Void control intentionally keeps
// the existing unkeyed one-condition-plus-fallback preparation contract.
func TestLucidKeyedEvaluationPreparation(t *testing.T) {
	for _, c := range []struct {
		name, inherited, inheritedType, local, wantKey, wantType string
		localSupplied, extraBody, invalid                        bool
	}{
		{name: "absent-key-unkeyed", wantKey: "Void", wantType: "String", extraBody: true},
		{name: "inherited-absence-unkeyed", inherited: "Void", inheritedType: "String", wantKey: "Void", wantType: "String", extraBody: true},
		{name: "inherited-key", inherited: "inherited", inheritedType: "String", wantKey: "inherited", wantType: "String"},
		{name: "inherited-key-enforces-pairing", inherited: "inherited", inheritedType: "String", extraBody: true, invalid: true},
		{name: "local-key-overrides-inherited", inherited: "inherited", inheritedType: "String", local: "local", localSupplied: true, wantKey: "local", wantType: "String"},
		{name: "local-key-overrides-absence", inherited: "Void", inheritedType: "String", local: "local", localSupplied: true, wantKey: "local", wantType: "String"},
		{name: "inherited-native-void", inherited: "Void", inheritedType: "SourceString", wantKey: "Void", wantType: "SourceString"},
		{name: "inherited-native-void-enforces-pairing", inherited: "Void", inheritedType: "SourceString", extraBody: true, invalid: true},
		{name: "local-native-void", inherited: "inherited", inheritedType: "String", local: "Void", localSupplied: true, wantKey: "Void", wantType: "SourceString"},
		{name: "local-native-void-enforces-pairing", local: "Void", localSupplied: true, extraBody: true, invalid: true},
		{name: "local-key-overrides-native-void", inherited: "Void", inheritedType: "SourceString", local: "local", localSupplied: true, wantKey: "local", wantType: "String"},
		{name: "local-empty-overrides-inherited", inherited: "inherited", inheritedType: "String", localSupplied: true, invalid: true},
		{name: "local-whitespace-overrides-inherited", inherited: "inherited", inheritedType: "String", local: " \t ", localSupplied: true, invalid: true},
	} {
		t.Run(c.name, func(t *testing.T) {
			defer executionRegistryPanicAsFailure(t)
			ss, err := loadSegmentalSystem(runtimeGenerationResources(t))
			if err != nil {
				t.Fatal(err)
			}
			if c.inheritedType != "" {
				ss.fileMeta.SetAdd("Key", c.inherited, c.inheritedType)
			}
			inherited := lucidKeyedSnapshot(t, ss.fileMeta)
			source := "GI"
			if c.localSupplied {
				source += "{|Key: ~" + c.local + "~|}"
			}
			source += "^(E{|Statement: ~true~|}*W"
			if c.extraBody {
				source += "*W"
			}
			source += ")"
			g, err := ss.initializeGlyphChecked([]byte(source), nil, nil)
			if err != nil {
				t.Fatal(err)
			}
			input := g.InputString
			var previous []byte
			for run := 0; run < 2; run++ {
				err := prepareGlyph(&g, glyphContext{callable: true, parent: "BOG", role: "expression"}, &generatedNames{used: source, counts: map[string]int{}})
				if c.invalid {
					if !errors.Is(err, lerrors.InvalidInput) {
						t.Errorf("run %d invalid keyed composition produced %v; want InvalidInput", run, err)
					}
					if err != nil {
						cause := err
						for errors.Unwrap(cause) != nil {
							cause = errors.Unwrap(cause)
						}
						want := "Rust GI: keyed selection requires a nonempty Key"
						if c.extraBody {
							want = "Rust GI: keyed selection requires one Operation per Calculation"
						}
						if cause.Error() != want {
							t.Errorf("run %d leaf cause %q; want %q", run, cause.Error(), want)
						}
					}
				} else {
					if err != nil {
						t.Fatalf("run %d preparation failed: %v", run, err)
					}
					if key := g.Meta.Get("Key"); key.Value != c.wantKey || key.Type != c.wantType {
						t.Errorf("run %d effective Key=%#v; want %s/%s", run, key, c.wantKey, c.wantType)
					}
					// Local-input map order is incidental; compare each effective
					// metadata record, not dictionary insertion order.
					semanticMeta := map[string]MetaObj{}
					for _, entry := range g.Meta.GetMetaData() {
						if _, exists := semanticMeta[entry.Key]; exists {
							t.Fatalf("duplicate effective metadata key %q", entry.Key)
						}
						semanticMeta[entry.Key] = entry
					}
					current := lucidKeyedSnapshot(t, semanticMeta)
					if run > 0 && !bytes.Equal(previous, current) {
						t.Error("repeated preparation changed effective metadata")
					}
					previous = current
				}
				if !bytes.Equal(inherited, lucidKeyedSnapshot(t, ss.fileMeta)) || g.InputString != input || g.Name != source {
					t.Error("preparation mutated file-owned metadata or supplied input")
				}
			}
		})
	}
}

func lucidKeyedSnapshot(t *testing.T, value interface{}) []byte {
	t.Helper()
	data, err := json.Marshal(value)
	if err != nil {
		t.Fatal(err)
	}
	return data
}
