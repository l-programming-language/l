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

// Each stored document isolates one malformed callable composition. Application
// and file names are neutral: the leaf diagnostic must identify the actual rule.
var lucidCallableInvalid = []struct {
	name, form, reason string
	classification     error
}{
	{"fw-missing-name", "FW", "FW requires a named Func; use BW for a closure", lerrors.InvalidInput},
	{"fw-empty-name", "FW", "FW requires a named Func; use BW for a closure", lerrors.InvalidInput},
	{"fw-whitespace-name", "FW", "FW requires a named Func; use BW for a closure", lerrors.InvalidInput},
	{"fw-record-missing-name", "FW", "FW requires a named Func; use BW for a closure", lerrors.InvalidInput},
	{"fw-record-empty-name", "FW", "FW requires a named Func; use BW for a closure", lerrors.InvalidInput},
	{"bw-flat-name", "BW", "unsupported metadata Key for this form", lerrors.InvalidInput},
	{"bw-flat-receiver", "BW", "unsupported metadata Receiver for this form", lerrors.InvalidInput},
	{"bw-record-name", "BW", "BW excludes a function name and receiver", lerrors.InvalidInput},
	{"bw-record-receiver", "BW", "BW excludes a function name and receiver", lerrors.InvalidInput},
	{"bw-file-scope", "BW", "BW requires KY callee or KO/KOB value within callable context", lerrors.InvalidInput},
	{"bw-operation-body", "BW", "BW requires KY callee or KO/KOB value within callable context", lerrors.InvalidInput},
	{"bw-file-kob", "BW", "BW requires KY callee or KO/KOB value within callable context", lerrors.InvalidInput},
	{"bw-file-ky", "BW", "BW requires KY callee or KO/KOB value within callable context", lerrors.InvalidInput},
	{"ky-missing-callee", "KY", "anonymous KY requires exactly one BW Operation", lerrors.InvalidInput},
	{"ky-multiple-callees", "KY", "anonymous KY requires exactly one BW Operation", lerrors.InvalidInput},
	{"ky-fw-child", "KY", "anonymous KY requires exactly one BW Operation", lerrors.InvalidInput},
	{"ky-e-child", "KY", "anonymous KY requires exactly one BW Operation", lerrors.InvalidInput},
	{"ky-w-child", "KY", "anonymous KY requires exactly one BW Operation", lerrors.InvalidInput},
	{"ky-named-child", "KY", "named KY excludes a callable child", lerrors.InvalidInput},
	{"ky-anonymous-receiver", "KY", "anonymous KY excludes Receiver", lerrors.InvalidInput},
	{"ky-empty-key", "KY", "Key must be nonempty when supplied", lerrors.InvalidInput},
	{"ky-whitespace-key", "KY", "Key must be nonempty when supplied", lerrors.InvalidInput},
	{"fw-func-inputs-conflict", "FW", "Func record and flat signature metadata are exclusive: conflicting Inputs", lerrors.InvalidInput},
	{"bw-func-outputs-conflict", "BW", "Func record and flat signature metadata are exclusive: conflicting Outputs", lerrors.InvalidInput},
	{"fw-capture", "FW", "Capture is unsupported by the FW/BW template interface", lerrors.InvalidInput},
	{"bw-capture", "BW", "Capture is unsupported by the FW/BW template interface", lerrors.InvalidInput},
	{"fw-record-capture", "FW", "unsupported Func field Capture", lerrors.InvalidInput},
	{"bw-record-capture", "BW", "unsupported Func field Capture", lerrors.InvalidInput},
	{"fw-parameter-header", "FW", "unexpected child BA: wrong category or role", lerrors.InvalidInput},
	{"fw-requirements-header", "FW", "unexpected child U: wrong category or role", lerrors.InvalidInput},
	{"bw-parameter-header", "BW", "unexpected child BA: wrong category or role", lerrors.InvalidInput},
	{"bw-requirements-header", "BW", "unexpected child U: wrong category or role", lerrors.InvalidInput},
	{"fw-unknown-func-field", "FW", "unsupported Func field Alpha", lerrors.InvalidInput},
	{"bw-unknown-func-field", "BW", "unsupported Func field Inputs Outputs", lerrors.InvalidInput},
	{"bw-unknown-receiver-field", "BW", "unsupported Receiver Prop field Unexpected", lerrors.InvalidInput},
	{"input-flat-value", "FW", "unsupported input Prop field Value", lerrors.InvalidInput},
	{"input-record-value", "BW", "unsupported input Prop field Value", lerrors.InvalidInput},
	{"input-flat-struct-tags", "BW", "unsupported input Prop field StructTags", lerrors.InvalidInput},
	{"input-record-struct-tags", "FW", "unsupported input Prop field StructTags", lerrors.InvalidInput},
	{"input-flat-unknown", "FW", "unsupported input Prop field Unknown", lerrors.InvalidInput},
	{"input-record-unknown", "BW", "unsupported input Prop field Unknown", lerrors.InvalidInput},
	{"input-missing-name", "FW", "function parameters require name and type", lerrors.InvalidInput},
	{"input-empty-name", "BW", `invalid Prop field "Key:"`, lerrors.InvalidInput},
	{"input-missing-type", "BW", "function parameters require name and type", lerrors.InvalidInput},
	{"input-empty-type", "FW", `invalid Prop field "Type:"`, lerrors.InvalidInput},
	{"fw-empty-func", "FW", "malformed or unclosed expected Func record", lerrors.Syntax},
	{"bw-malformed-func", "BW", "malformed or unclosed expected Func record", lerrors.Syntax},
	{"fw-duplicate-func-key", "FW", "duplicate Func field Key", lerrors.InvalidInput},
	{"fw-malformed-input-prop", "FW", "malformed or unclosed expected Prop record", lerrors.Syntax},
	{"bw-duplicate-input-prop-key", "BW", "duplicate Prop field Key", lerrors.InvalidInput},
	{"fw-malformed-receiver-prop", "FW", "malformed or unclosed expected Prop record", lerrors.Syntax},
	{"my-receiver", "MY", "Receiver is unsupported by the MY Funcs template interface", lerrors.InvalidInput},
	{"my-empty-receiver", "MY", "Receiver is unsupported by the MY Funcs template interface", lerrors.InvalidInput},
	{"my-missing-name", "MY", "trait signatures require named methods", lerrors.InvalidInput},
	{"my-unknown-func-field", "MY", "unsupported Func field Unexpected", lerrors.InvalidInput},
	{"my-duplicate-func-key", "MY", "duplicate Func field Key", lerrors.InvalidInput},
	{"my-malformed-func", "MY", "malformed or unclosed expected Func record", lerrors.Syntax},
	{"my-input-missing-name", "MY", "function parameters require name and type", lerrors.InvalidInput},
	{"my-input-missing-type", "MY", "function parameters require name and type", lerrors.InvalidInput},
	{"my-input-unknown-field", "MY", "unsupported input Prop field Unexpected", lerrors.InvalidInput},
	{"my-duplicate-input-prop-key", "MY", "duplicate Prop field Key", lerrors.InvalidInput},
	{"my-malformed-input-prop", "MY", "malformed or unclosed expected Prop record", lerrors.Syntax},
}

var lucidCallableAssets = []string{
	"builtins/segments/consonants/operate.lu",
	"builtins/segments/vowels/operation.lu",
	"builtins/segments/consonants/instantiate.lu",
	"builtins/segments/vowels/instantiation.lu",
}

func TestLucidCallableValidation(t *testing.T) {
	files, err := filepath.Glob(filepath.Join("testdata", "lucid", "callable-invalid", "*.aml.txt"))
	if err != nil || len(files) != len(lucidCallableInvalid) || len(files) != 62 {
		t.Fatalf("invalid callable fixture inventory: files=%d cases=%d error=%v", len(files), len(lucidCallableInvalid), err)
	}
	seen := map[string]bool{}
	for _, c := range lucidCallableInvalid {
		if seen[c.name] {
			t.Fatalf("duplicate invalid fixture %s", c.name)
		}
		seen[c.name] = true
		t.Run(c.name, func(t *testing.T) {
			defer executionRegistryPanicAsFailure(t)
			r, dir := lucidRecord(t, "callable-"+c.name, "validation")
			input := lucidRead(t, "callable-invalid/"+c.name+".aml.txt")
			before := bytes.Clone(input)
			lucidWrite(t, filepath.Join(dir, "input.aml.txt"), input)
			config := runtimeGenerationResources(t)
			resources := &lucidRecordingFS{FS: config, reads: map[string]bool{}}
			config = resources
			bundle, err := NewLogicemeService(config).Evaluate(input, GenerationOptions{IncludeRuntime: false})
			if !bytes.Equal(input, before) {
				lucidIssue(t, r, "production validation: mutated input")
			}
			if !reflect.DeepEqual(bundle, Executable{}) {
				lucidIssue(t, r, "production validation: returned nonzero bundle on malformed input")
			}
			for _, asset := range lucidCallableAssets {
				if !resources.reads[asset] {
					lucidIssue(t, r, "production selection: did not load %s", asset)
				}
			}
			lucidSaveBundle(t, r, dir, bundle)
			if err == nil {
				lucidIssue(t, r, "production validation: accepted malformed composition; want %s", c.reason)
				return
			}
			r.GenerationError = err.Error()
			if !errors.Is(err, c.classification) {
				lucidIssue(t, r, "production validation: error %v is not %v", err, c.classification)
			}
			cause := err
			for errors.Unwrap(cause) != nil {
				cause = errors.Unwrap(cause)
			}
			want := "Rust " + c.form + ": " + c.reason
			if cause.Error() != want {
				lucidIssue(t, r, "production validation: cause=%q want=%q", cause.Error(), want)
			}
		})
	}
}

// Func/Prop records retain their existing formatting. In
// particular, changing callable inputs must not globally change other Props.
func TestLucidCallableParameterInterpolation(t *testing.T) {
	cases := []struct{ name, inputs, nativeWant string }{
		{"empty", "&[]&", ""},
		{"single", "&[Prop{Key: value, Type: i32}Prop]&", "value: i32"},
		{"multiple", "&[Prop{Key: first, Type: i32}Prop, Prop{Key: second, Type: bool}Prop]&", "first: i32, second: bool"},
		{"raw-identifier", "&[Prop{Key: r#type, Type: i32}Prop]&", "r#type: i32"},
		{"native-punctuation", "&[Prop{Key: data, Type: &'a [u8; 4]}Prop]&", "data: &'a [u8; 4]"},
	}
	for _, language := range []string{"Rust"} {
		for _, c := range cases {
			t.Run(language+"/"+c.name, func(t *testing.T) {
				defer executionRegistryPanicAsFailure(t)
				want := c.nativeWant
				lucidFuncRepeatedInterpretation(t, language,
					"#[Func{Key: helper, Inputs: "+c.inputs+"}Func]#", "unrelated",
					"before<[|Meta.Func.Inputs~|]>after", "before<"+want+">after")
			})
		}
		t.Run(language+"/receiver-prop-unchanged", func(t *testing.T) {
			defer executionRegistryPanicAsFailure(t)
			lucidFuncRepeatedInterpretation(t, language,
				"#[Func{Key: helper, Receiver: #[Prop{Key: self, Type: Widget}Prop]#}Func]#", "",
				"before<[|Meta.Func.Receiver~|]>after", "before<self Widget>after")
		})
		t.Run(language+"/ordinary-prop-unchanged", func(t *testing.T) {
			defer executionRegistryPanicAsFailure(t)
			meta := MetaDict{}
			meta.SetAdd("Prop", "#[Prop{Key: field, Type: Widget}Prop]#", "Hashmap")
			source := "before<[|Meta.Prop.Key|] [|Meta.Prop.Type|]>after"
			li := LucidInterpreter{
				RawContent: []byte(source), SubTags: RecursiveSubTagParser(source, &TagMeta{}),
				Meta: meta, SegmentalSystem: &SegmentalSystem{Language: language},
				WorkingVariables: []*WorkingVariable{{VarName: "Saved", VarType: "String", VarValue: "unchanged"}},
			}
			metadata := marshalTestJSON(t, li.Meta)
			bindings := marshalTestJSON(t, li.WorkingVariables)
			parsed := marshalTestJSON(t, li.SubTags)
			for run := 0; run < 2; run++ {
				li.Interpret()
				if got := string(li.InterpretedContent); got != "before<field Widget>after" {
					t.Errorf("run %d: rendered %q; want before<field Widget>after", run, got)
				}
				afterMetadata := marshalTestJSON(t, li.Meta)
				afterBindings := marshalTestJSON(t, li.WorkingVariables)
				afterParsed := marshalTestJSON(t, li.SubTags)
				if !bytes.Equal(metadata, afterMetadata) || !bytes.Equal(bindings, afterBindings) || !bytes.Equal(parsed, afterParsed) || string(li.RawContent) != source {
					t.Error("interpretation mutated metadata, bindings or parsed source")
				}
			}
		})
	}
}

func TestLucidCallableIsolation(t *testing.T) {
	cases := []string{"callable-typed-record", "closure-stored-capture", "callable-typed-record"}
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
			for _, asset := range lucidCallableAssets {
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
				r, dir := lucidRecord(t, "callable-"+provider+"-"+string(rune('A'+i)), "isolation")
				input := lucidRead(t, name+".aml.txt")
				before := bytes.Clone(input)
				lucidWrite(t, filepath.Join(dir, "input.aml.txt"), input)
				got, err := service.Evaluate(input, GenerationOptions{IncludeRuntime: false})
				if !bytes.Equal(input, before) {
					lucidIssue(t, r, "production isolation: mutated complete AML input")
				}
				if err != nil {
					r.GenerationError = err.Error()
					lucidIssue(t, r, "production isolation: generation failed: %v", err)
					return
				}
				lucidSaveBundle(t, r, dir, got)
				// A carries a nested import; B has none. Exact bundles check both
				// import ownership and callable signature/capture state isolation.
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
			for _, asset := range lucidCallableAssets {
				if !resources.reads[asset] {
					t.Errorf("selected resource was not loaded: %s", asset)
				}
			}
		})
	}
}
