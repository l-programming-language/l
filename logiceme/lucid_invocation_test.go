package Logiceme

import (
	"bytes"
	"encoding/json"
	"errors"
	"io/fs"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/l-programming-language/l/lerrors"

	distribution "github.com/l-programming-language/l"
)

var lucidInvocationAssets = []string{
	"builtins/segments/consonants/instantiate.lu",
	"builtins/segments/vowels/instantiation.lu",
}

func TestLucidInvocationInvalid(t *testing.T) {
	cases := []string{"empty-receiver", "whitespace-receiver"}
	files, err := filepath.Glob(filepath.Join("testdata", "lucid", "invocation-invalid", "*.aml.txt"))
	if err != nil || len(files) != len(cases) {
		t.Fatalf("invalid fixture inventory: files=%d cases=%d error=%v", len(files), len(cases), err)
	}
	for _, name := range cases {
		t.Run(name, func(t *testing.T) {
			defer executionRegistryPanicAsFailure(t)
			r, dir := lucidRecord(t, name, "validation")
			input := lucidRead(t, "invocation-invalid/"+name+".aml.txt")
			before := bytes.Clone(input)
			lucidWrite(t, filepath.Join(dir, "input.aml.txt"), input)
			config := runtimeGenerationResources(t)
			resources := &lucidRecordingFS{FS: config, reads: map[string]bool{}}
			config = resources
			got, err := NewLogicemeService(config).Evaluate(input, GenerationOptions{IncludeRuntime: false})
			lucidSaveBundle(t, r, dir, got)
			if !bytes.Equal(input, before) {
				lucidIssue(t, r, "production validation: mutated input")
			}
			if !reflect.DeepEqual(got, Executable{}) {
				lucidIssue(t, r, "production validation: returned partial bundle")
			}
			for _, asset := range lucidInvocationAssets {
				if !resources.reads[asset] {
					lucidIssue(t, r, "production selection: did not load %s", asset)
				}
			}
			if err == nil {
				lucidIssue(t, r, "production validation: accepted empty Receiver")
				return
			}
			r.GenerationError = err.Error()
			if !errors.Is(err, lerrors.InvalidInput) {
				lucidIssue(t, r, "production validation: expected InvalidInput, got %v", err)
			}
			cause := err
			for errors.Unwrap(cause) != nil {
				cause = errors.Unwrap(cause)
			}
			const want = "Rust KY: Receiver must be nonempty when supplied"
			if cause.Error() != want {
				lucidIssue(t, r, "production validation: cause=%q want=%q", cause.Error(), want)
			}
		})
	}
}

func TestLucidInvocationIsolation(t *testing.T) {
	cases := []string{"invocation-named", "schedule-composed", "invocation-named"}
	var explicit []Executable
	for _, provider := range []string{"filesystem", "embedded"} {
		t.Run(provider, func(t *testing.T) {
			defer executionRegistryPanicAsFailure(t)
			config := runtimeGenerationResources(t)
			filesystem := config
			if provider == "embedded" {
				config = distribution.Files()
			}
			resources := &lucidRecordingFS{FS: config, reads: map[string]bool{}}
			for _, asset := range lucidInvocationAssets {
				left, e1 := fs.ReadFile(filesystem, asset)
				right, e2 := fs.ReadFile(resources.FS, asset)
				if e1 != nil || e2 != nil || !bytes.Equal(left, right) {
					t.Fatalf("resource mismatch %s: filesystem=%v selected=%v", asset, e1, e2)
				}
			}
			config = resources
			service := NewLogicemeService(config)
			var first Executable
			var firstBytes []byte
			for i, name := range cases {
				r, dir := lucidRecord(t, "invocation-"+provider+"-"+string(rune('A'+i)), "isolation")
				input := lucidRead(t, name+".aml.txt")
				before := bytes.Clone(input)
				lucidWrite(t, filepath.Join(dir, "input.aml.txt"), input)
				got, err := service.Evaluate(input, GenerationOptions{IncludeRuntime: false})
				if !bytes.Equal(input, before) {
					lucidIssue(t, r, "production isolation: mutated input")
				}
				if err != nil {
					r.GenerationError = err.Error()
					lucidIssue(t, r, "production isolation: generation failed: %v", err)
					return
				}
				lucidSaveBundle(t, r, dir, got)
				// B's method receiver and import must not leak into A's free calls.
				lucidCompare(t, r, dir, got, lucidExpectedBundle(t, name))
				if provider == "filesystem" {
					explicit = append(explicit, got)
				} else if len(explicit) != len(cases) || !reflect.DeepEqual(explicit[i], got) {
					lucidIssue(t, r, "production isolation: resource providers produced different bundles")
				}
				if i == 0 {
					first = got
					firstBytes, err = json.Marshal(first)
					if err != nil {
						t.Fatal(err)
					}
				} else {
					current, err := json.Marshal(first)
					if err != nil {
						t.Fatal(err)
					}
					if !bytes.Equal(firstBytes, current) {
						lucidIssue(t, r, "production isolation: mutated earlier bundle")
					}
					if i == len(cases)-1 && !reflect.DeepEqual(first, got) {
						lucidIssue(t, r, "production isolation: A/B/A changed output")
					}
				}
			}
			for _, asset := range lucidInvocationAssets {
				if !resources.reads[asset] {
					t.Errorf("selected resource was not loaded: %s", asset)
				}
			}
		})
	}
}

// Bare Void tests absence; quoted Void continues to compare its literal bytes.
// The Rust distinction must preserve prepared metadata.
func TestLucidInvocationVoidComparisons(t *testing.T) {
	cases := []struct {
		name, expression string
		nativeWant       bool
	}{
		{"native-bare-equal", "Meta.Native == Void", false},
		{"native-bare-unequal", "Meta.Native != Void", true},
		{"native-quoted-equal", `Meta.Native == "Void"`, true},
		{"native-quoted-unequal", `Meta.Native != "Void"`, false},
		{"absent-bare-equal", "Meta.Absent == Void", true},
		{"absent-bare-unequal", "Meta.Absent != Void", false},
		{"absent-quoted-equal", `Meta.Absent == "Void"`, true},
		{"absent-quoted-unequal", `Meta.Absent != "Void"`, false},
		{"ordinary-bare-equal", "Meta.Ordinary == Void", false},
		{"ordinary-bare-unequal", "Meta.Ordinary != Void", true},
		{"ordinary-quoted-equal", `Meta.Ordinary == "ready"`, true},
		{"ordinary-quoted-unequal", `Meta.Ordinary != "ready"`, false},
	}
	for _, language := range []string{"Rust"} {
		for _, c := range cases {
			t.Run(language+"/"+c.name, func(t *testing.T) {
				meta := MetaDict{}
				meta.SetAdd("Native", "Void", "SourceString")
				meta.SetAdd("Absent", "Void", "String")
				meta.SetAdd("Ordinary", "ready", "SourceString")
				source := "(/={tags: Meta.Tags}=={if " + c.expression + "}=={yes}=/)"
				li := LucidInterpreter{
					RawContent: []byte(source), SubTags: RecursiveSubTagParser(source, &TagMeta{}),
					Meta: meta, SegmentalSystem: &SegmentalSystem{Language: language},
					WorkingVariables: []*WorkingVariable{{VarName: "Saved", VarType: "String", VarValue: "unchanged"}},
				}
				metadata := marshalTestJSON(t, li.Meta)
				bindings := marshalTestJSON(t, li.WorkingVariables)
				parsed := marshalTestJSON(t, li.SubTags)
				wantCondition := c.nativeWant
				want := ""
				if wantCondition {
					want = "yes"
				}
				for run := 0; run < 2; run++ {
					li.Interpret()
					if got := string(li.InterpretedContent); got != want {
						t.Errorf("run %d: rendered %q; want %q", run, got, want)
					}
					afterMetadata := marshalTestJSON(t, li.Meta)
					afterBindings := marshalTestJSON(t, li.WorkingVariables)
					afterParsed := marshalTestJSON(t, li.SubTags)
					if !bytes.Equal(metadata, afterMetadata) || !bytes.Equal(bindings, afterBindings) || !bytes.Equal(parsed, afterParsed) || string(li.RawContent) != source {
						t.Error("comparison mutated metadata, bindings or parsed source")
					}
				}
			})
		}
	}
}

func TestLucidInvocationReceiverPreparation(t *testing.T) {
	for _, c := range []struct{ name, input, kind, source string }{
		{"omitted", "KY{|Key: ~callee~|}", "String", "\n\ncallee()\n"},
		{"native-void", "KY{|Key: ~callee~, Receiver: ~Void~|}", "SourceString", "\nVoid.callee()\n\n"},
	} {
		t.Run(c.name, func(t *testing.T) {
			defer executionRegistryPanicAsFailure(t)
			ss, err := loadSegmentalSystem(runtimeGenerationResources(t))
			if err != nil {
				t.Fatal(err)
			}
			// Inherited native text has the same bytes as the absence sentinel.
			// Resetting local provenance must not alter the owning file metadata.
			ss.fileMeta.SetAdd("Receiver", "Void", "SourceString")
			inherited := marshalTestJSON(t, ss.fileMeta)
			g, err := ss.initializeGlyphChecked([]byte(c.input), nil, nil)
			if err != nil {
				t.Fatal(err)
			}
			input := g.InputString
			for run := 0; run < 2; run++ {
				if err := prepareGlyph(&g, glyphContext{callable: true, role: "expression"}, &generatedNames{used: c.input, counts: map[string]int{}}); err != nil {
					t.Fatal(err)
				}
				receiver := g.Meta.Get("Receiver")
				if receiver.Value != "Void" || receiver.Type != c.kind {
					t.Errorf("run %d: Receiver=%#v; want Void/%s", run, receiver, c.kind)
				}
				if err := g.compilePreparedGlyph(); err != nil {
					t.Fatal(err)
				}
				if got := string(g.Content); got != c.source {
					t.Errorf("run %d: rendered %q; want %q", run, got, c.source)
				}
				after := marshalTestJSON(t, ss.fileMeta)
				if !bytes.Equal(inherited, after) || g.InputString != input {
					t.Error("preparation mutated inherited metadata or supplied input")
				}
			}
		})
	}
}
