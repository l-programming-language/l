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

// These are complete AML documents with a neutral application/file context.
// Each row isolates a composition failure, leaving native pattern/type legality
// to the separately compiled complete-project cases.
var lucidIterationInvalid = []struct {
	name, form, reason string
	classification     error
}{
	{"calculation-evaluation-conflict", "X", "iteration accepts at most one condition", lerrors.InvalidInput},
	{"calculation-definition-conflict", "X", "iteration accepts at most one condition", lerrors.InvalidInput},
	{"evaluation-definition-conflict", "X", "iteration accepts at most one condition", lerrors.InvalidInput},
	{"repeated-calculation", "X", "iteration accepts at most one condition", lerrors.InvalidInput},
	{"repeated-evaluation", "X", "iteration accepts at most one condition", lerrors.InvalidInput},
	{"repeated-definition", "X", "iteration accepts at most one condition", lerrors.InvalidInput},
	{"iterator-calculation-conflict", "X", "iterator does not accept condition children", lerrors.InvalidInput},
	{"iterator-evaluation-conflict", "X", "iterator does not accept condition children", lerrors.InvalidInput},
	{"iterator-definition-conflict", "X", "iterator does not accept condition children", lerrors.InvalidInput},
	{"three-part", "X", "three-part iteration is unsupported; move initialization and steps into explicit compositions", lerrors.InvalidInput},
	{"missing-body", "X", "Body requires 1..1 Operation children", lerrors.InvalidInput},
	{"multiple-bodies", "X", "Body requires 1..1 Operation children", lerrors.InvalidInput},
	{"wrong-category-body", "X", "Body requires 1..1 Operation children", lerrors.InvalidInput},
	{"ordinary-o-condition", "X", "iteration Definition condition requires OG or DOG", lerrors.InvalidInput},
	{"ordinary-ob-condition", "X", "iteration Definition condition requires OG or DOG", lerrors.InvalidInput},
	{"ordinary-do-condition", "X", "iteration Definition condition requires OG or DOG", lerrors.InvalidInput},
	{"ordinary-dob-condition", "X", "iteration Definition condition requires OG or DOG", lerrors.InvalidInput},
	{"unsupported-evaluation-adapter", "X", "iteration Evaluation condition requires I", lerrors.InvalidInput},
	{"og-missing-pattern", "OG", "conditional definition requires exactly one nonempty pattern", lerrors.InvalidInput},
	{"og-empty-pattern", "OG", "conditional definition requires exactly one nonempty pattern", lerrors.InvalidInput},
	{"og-multiple-patterns", "OG", "conditional definition requires exactly one nonempty pattern", lerrors.InvalidInput},
	{"og-missing-value", "OG", "OG requires exactly one explicit nonempty value", lerrors.InvalidInput},
	{"og-empty-value", "OG", "OG requires exactly one explicit nonempty value", lerrors.InvalidInput},
	{"og-multiple-values", "OG", "OG requires exactly one explicit nonempty value", lerrors.InvalidInput},
	{"og-child", "OG", "OG cannot have children", lerrors.InvalidInput},
	{"og-outside-x", "OG", "conditional definition requires direct X condition context", lerrors.InvalidInput},
	{"dog-missing-pattern", "DOG", "conditional definition requires exactly one nonempty pattern", lerrors.InvalidInput},
	{"dog-empty-pattern", "DOG", "conditional definition requires exactly one nonempty pattern", lerrors.InvalidInput},
	{"dog-multiple-patterns", "DOG", "conditional definition requires exactly one nonempty pattern", lerrors.InvalidInput},
	{"dog-missing-child", "DOG", "DOG requires exactly one Calculation child", lerrors.InvalidInput},
	{"dog-multiple-children", "DOG", "DOG requires exactly one Calculation child", lerrors.InvalidInput},
	{"dog-wrong-child", "DOG", "DOG requires exactly one Calculation child", lerrors.InvalidInput},
	{"dog-native-value", "DOG", "unsupported metadata Value for this form", lerrors.InvalidInput},
	{"dog-outside-x", "DOG", "conditional definition requires direct X condition context", lerrors.InvalidInput},
	{"po-missing-child", "PO", "PO requires exactly one Iteration child", lerrors.InvalidInput},
	{"po-multiple-children", "PO", "PO requires exactly one Iteration child", lerrors.InvalidInput},
	{"po-wrong-child", "PO", "PO requires exactly one Iteration child", lerrors.InvalidInput},
	{"po-native-values", "PO", "unsupported metadata Values for this form", lerrors.InvalidInput},
	{"po-missing-label", "PO", "label definition requires exactly one nonempty label", lerrors.InvalidInput},
	{"po-empty-label", "PO", "label definition requires exactly one nonempty label", lerrors.InvalidInput},
	{"po-multiple-labels", "PO", "label definition requires exactly one nonempty label", lerrors.InvalidInput},
	{"po-unsupported-coda", "", "unsupported Rust syllable/coda POC", lerrors.Unsupported},
	{"po-file-scope", "PO", "label definition requires a callable W statement context", lerrors.InvalidInput},
	{"og-key-keys-conflict", "OG", "Key and Keys metadata are exclusive", lerrors.InvalidInput},
	{"og-value-values-conflict", "OG", "Value and Values metadata are exclusive", lerrors.InvalidInput},
	{"dog-key-keys-conflict", "DOG", "Key and Keys metadata are exclusive", lerrors.InvalidInput},
	{"po-key-keys-conflict", "PO", "Key and Keys metadata are exclusive", lerrors.InvalidInput},
	{"og-unsupported-type", "OG", "unsupported metadata Type for this form", lerrors.InvalidInput},
	{"dog-unsupported-mutable", "DOG", "Mutable only applies to newly declared simple bindings; it cannot qualify assignments or a complete pattern", lerrors.InvalidInput},
	{"po-unsupported-parameter", "PO", "unsupported metadata Parameter for this form", lerrors.InvalidInput},
}

var lucidIterationAssets = []string{
	"builtins/segments/consonants/iterate.lu",
	"builtins/segments/vowels/iteration.lu",
	"builtins/segments/consonants/define.lu",
	"builtins/segments/vowels/definition.lu",
}

func TestLucidIterationValidation(t *testing.T) {
	files, err := filepath.Glob(filepath.Join("testdata", "lucid", "iteration-invalid", "*.aml.txt"))
	if err != nil || len(files) != len(lucidIterationInvalid) || len(files) != 50 {
		t.Fatalf("invalid iteration fixture inventory: files=%d cases=%d error=%v", len(files), len(lucidIterationInvalid), err)
	}
	seen := map[string]bool{}
	for _, c := range lucidIterationInvalid {
		if seen[c.name] {
			t.Fatalf("duplicate invalid fixture %s", c.name)
		}
		seen[c.name] = true
		t.Run(c.name, func(t *testing.T) {
			defer executionRegistryPanicAsFailure(t)
			r, dir := lucidRecord(t, "iteration-"+c.name, "validation")
			input := lucidRead(t, "iteration-invalid/"+c.name+".aml.txt")
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
			for _, asset := range lucidIterationAssets {
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
			want := c.reason
			if c.form != "" {
				want = "Rust " + c.form + ": " + c.reason
			}
			if cause.Error() != want {
				lucidIssue(t, r, "production validation: cause = %q; want %q", cause.Error(), want)
			}
		})
	}
}

// This primitive check supplements complete AML coverage. Each row selects one
// legal branch and preserves the asset's full literal layout without formatting.
func TestLucidIterateBranches(t *testing.T) {
	for _, c := range []struct {
		name, conditionType, condition, iterable, key, index, arity string
	}{
		{name: "definition", conditionType: "Definition", condition: "let Some(value) = source"},
		{name: "calculation", conditionType: "Calculation", condition: "ready"},
		{name: "evaluation", conditionType: "Evaluation", condition: "ready"},
		{name: "unconditional"},
		{name: "wildcard", iterable: "items", arity: "0"},
		{name: "single", iterable: "items", key: "value", arity: "1"},
		{name: "discard-pair", iterable: "items", key: "value"},
		{name: "pair", iterable: "items", key: "value", index: "index", arity: "2"},
	} {
		t.Run(c.name, func(t *testing.T) {
			defer executionRegistryPanicAsFailure(t)
			r, dir := lucidRecord(t, "iterate-"+c.name, "primitive")
			ss, err := loadSegmentalSystem(runtimeGenerationResources(t))
			if err != nil {
				t.Fatal(err)
			}
			g := Glyph{SegmentalSystem: &ss}
			for key, value := range map[string]string{"Iterable": c.iterable, "Key": c.key, "Index": c.index, "RangeArity": c.arity} {
				if value == "" {
					value = "Void"
				}
				g.Meta.SetAdd(key, value, "String")
			}
			tags := g.FunctionalTags(ss.Iterate.Content, "Default")
			if len(tags) != 1 {
				t.Fatalf("Iterate Default selected %d executable cases; want one", len(tags))
			}
			conjoinments := []CompiledConjoinment{{Key: "Operation", Type: "Operation", Contents: [][]byte{[]byte("body();")}}}
			if c.conditionType != "" {
				conjoinments = append(conjoinments, CompiledConjoinment{Key: c.conditionType, Type: c.conditionType, Contents: [][]byte{[]byte(c.condition)}})
			}
			got := tags[0].Interpret(DataCompilation{Meta: g.Meta, SegmentalSystem: &ss, Conjoinments: conjoinments})
			file := "primitive/iterate-" + c.name + ".rs.txt"
			want := lucidRead(t, file)
			r.Expected["Iterate.Default"] = "logiceme/testdata/lucid/" + file
			r.ExpectedHashes["Iterate.Default"] = lucidHash(want)
			r.ActualHashes["Iterate.Default"] = lucidHash(got)
			lucidWrite(t, filepath.Join(dir, "actual.rs.txt"), got)
			if !bytes.Equal(got, want) {
				lucidIssue(t, r, "Iterate branch bytes = %q; want %q", got, want)
			}
		})
	}
}

func TestLucidIterationIsolation(t *testing.T) {
	providerFirst := map[string]Executable{}
	for _, provider := range []string{"filesystem", "embedded"} {
		t.Run(provider, func(t *testing.T) {
			defer executionRegistryPanicAsFailure(t)
			config := runtimeGenerationResources(t)
			if provider == "embedded" {
				config = distribution.Files()
			}
			resources := &lucidRecordingFS{FS: config, reads: map[string]bool{}}
			config = resources
			service := NewLogicemeService(config)
			var first Executable
			var firstBytes []byte
			for i, name := range []string{"iteration-bindings", "iteration-nesting-scope", "iteration-bindings"} {
				r, dir := lucidRecord(t, "iteration-"+provider+"-"+string(rune('A'+i)), "isolation")
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
					continue
				}
				lucidSaveBundle(t, r, dir, got)
				lucidCompare(t, r, dir, got, lucidExpectedBundle(t, name))
				if i == 0 {
					first = got
					providerFirst[provider] = got
					firstBytes, err = json.Marshal(first)
					if err != nil {
						t.Fatal(err)
					}
				} else {
					current, err := json.Marshal(first)
					if err != nil {
						t.Fatal(err)
					}
					if !bytes.Equal(current, firstBytes) {
						lucidIssue(t, r, "production isolation: later evaluation mutated earlier bundle")
					}
					if i == 2 && !reflect.DeepEqual(first, got) {
						lucidIssue(t, r, "production isolation: A/B/A changed the returned bundle")
					}
				}
			}
			for _, asset := range lucidIterationAssets {
				if !resources.reads[asset] {
					t.Errorf("missing Rust resource provenance for %s", asset)
				}
			}
		})
	}
	if len(providerFirst) == 2 && !reflect.DeepEqual(providerFirst["filesystem"], providerFirst["embedded"]) {
		t.Error("explicit and embedded providers returned different iteration bundles")
	}
}

func TestLucidIterationAssetProviders(t *testing.T) {
	config := runtimeGenerationResources(t)
	filesystem := config
	config = distribution.Files()
	embedded := config
	for _, asset := range lucidIterationAssets {
		left, e1 := fs.ReadFile(filesystem, asset)
		right, e2 := fs.ReadFile(embedded, asset)
		if e1 != nil || e2 != nil || !bytes.Equal(left, right) {
			t.Errorf("resource mismatch %s: filesystem=%v embedded=%v", asset, e1, e2)
		}
	}
	// The composed original Rust control rules out accidental target fallback
	// or an all-rejections implementation satisfying the malformed cases.
	input := lucidRead(t, "communication.aml.txt")
	before := bytes.Clone(input)
	got, err := NewLogicemeService(config).Evaluate(input, GenerationOptions{})
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(input, before) {
		t.Fatal("valid Rust control mutated input")
	}
	r, dir := lucidRecord(t, "iteration-resource-control", "isolation")
	lucidSaveBundle(t, r, dir, got)
	lucidCompare(t, r, dir, got, lucidExpectedBundle(t, "communication"))
}
