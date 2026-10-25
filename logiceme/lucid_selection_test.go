package Logiceme

import (
	"bytes"
	"encoding/json"
	"errors"
	"io/fs"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/l-programming-language/l/lerrors"

	distribution "github.com/l-programming-language/l"
)

// These documents remain structurally complete. Every violation is isolated
// under a neutral application name so context cannot satisfy a rule assertion.
var lucidSelectionInvalid = []struct {
	name, form, reason string
	classification     error
}{
	{"ci-file-scope", "CI", "select requires callable context", lerrors.InvalidInput},
	{"ci-non-bi-child", "CI", "select accepts BI child branches only", lerrors.InvalidInput},
	{"bi-orphan", "BI", "BI requires CI parent", lerrors.InvalidInput},
	{"missing-guard", "BI", "select branch requires exactly one OC/DOC Definition", lerrors.InvalidInput},
	{"mixed-guards", "BI", "select branch requires exactly one OC/DOC Definition", lerrors.InvalidInput},
	{"bare-calculation", "BI", "select branch requires an OC/DOC Definition; wrap the Calculation in DOC with an explicit key", lerrors.InvalidInput},
	{"doc-wrong-child", "DOC", "DOC requires exactly one Calculation child", lerrors.InvalidInput},
	{"unsupported-definition", "BI", "select branch requires OC/DOC; use an explicit O in the body for outer assignments", lerrors.InvalidInput},
	{"legacy-do-guard", "BI", "select branch requires OC/DOC; use an explicit O in the body for outer assignments", lerrors.InvalidInput},
	{"missing-name", "DOC", "asynchronous definition requires exactly one binding name", lerrors.InvalidInput},
	{"multiple-names", "DOC", "asynchronous definition requires exactly one binding name", lerrors.InvalidInput},
	{"pattern-binding", "DOC", "asynchronous definition requires a simple binding name or _", lerrors.InvalidInput},
	{"assignment-target", "DOC", "asynchronous definition requires a simple binding name or _", lerrors.InvalidInput},
	{"duplicate-default", "CI", "select has multiple defaults", lerrors.InvalidInput},
	{"default-missing-definition", "BI", "select branch requires exactly one OC/DOC Definition", lerrors.InvalidInput},
	{"invalid-default", "BI", "Default must be True or False", lerrors.InvalidInput},
	{"multiple-bodies", "BI", "select branch accepts at most one W body", lerrors.InvalidInput},
	{"non-w-body", "BI", "select branch body requires W", lerrors.InvalidInput},
	{"guard-type", "DOC", "unsupported metadata Type for this form", lerrors.InvalidInput},
	{"guard-mutable", "DOC", "Mutable only applies to newly declared simple bindings; it cannot qualify assignments or a complete pattern", lerrors.InvalidInput},
	{"guard-parameter", "DOC", "unsupported metadata Parameter for this form", lerrors.InvalidInput},
	{"caller-key", "BI", "unsupported metadata Key for this form", lerrors.InvalidInput},
	{"bare-ce", "BI", "select branch requires an OC/DOC Definition; wrap the Calculation in DOC with an explicit key", lerrors.InvalidInput},
	{"legacy-dob-guard", "BI", "select branch requires OC/DOC; use an explicit O in the body for outer assignments", lerrors.InvalidInput},
	{"multiple-definitions", "BI", "select branch requires exactly one OC/DOC Definition", lerrors.InvalidInput},
	{"oc-missing-value", "OC", "OC requires exactly one explicit nonempty value", lerrors.InvalidInput},
	{"oc-empty-value", "OC", "OC requires exactly one explicit nonempty value", lerrors.InvalidInput},
	{"oc-multiple-values", "OC", "OC requires exactly one explicit nonempty value", lerrors.InvalidInput},
	{"oc-child", "OC", "OC cannot have children", lerrors.InvalidInput},
	{"doc-missing-child", "DOC", "DOC requires exactly one Calculation child", lerrors.InvalidInput},
	{"doc-multiple-children", "DOC", "DOC requires exactly one Calculation child", lerrors.InvalidInput},
	{"doc-native-value", "DOC", "unsupported metadata Value for this form", lerrors.InvalidInput},
	{"oc-outside-bi", "OC", "asynchronous definition requires direct BI selection context", lerrors.InvalidInput},
	{"doc-outside-bi", "DOC", "asynchronous definition requires direct BI selection context", lerrors.InvalidInput},
	{"unsupported-c-qualified", "", "unsupported Rust syllable/coda FOC", lerrors.Unsupported},
}

type lucidRecordingFS struct {
	fs.FS
	reads map[string]bool
}

func (r *lucidRecordingFS) Open(name string) (fs.File, error) {
	r.reads[name] = true
	return r.FS.Open(name)
}

func TestLucidSelectionValidation(t *testing.T) {
	files, err := filepath.Glob(filepath.Join("testdata", "lucid", "selection-invalid", "*.aml.txt"))
	if err != nil || len(files) != len(lucidSelectionInvalid) {
		t.Fatalf("invalid fixture inventory: files=%d cases=%d error=%v", len(files), len(lucidSelectionInvalid), err)
	}
	seen := map[string]bool{}
	for _, c := range lucidSelectionInvalid {
		if seen[c.name] {
			t.Fatalf("duplicate invalid fixture %s", c.name)
		}
		seen[c.name] = true
		t.Run(c.name, func(t *testing.T) {
			defer executionRegistryPanicAsFailure(t)
			r, dir := lucidRecord(t, c.name, "validation")
			input := lucidRead(t, "selection-invalid/"+c.name+".aml.txt")
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
			if !resources.reads["builtins/segments/consonants/define.lu"] || !resources.reads["builtins/segments/consonants/evaluate.lu"] || !resources.reads["builtins/segments/vowels/evaluation.lu"] {
				lucidIssue(t, r, "production selection: did not load Rust Define/Evaluate/Evaluation assets")
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
			// Compare the responsible leaf diagnostic, not the app/file/child
			// wrappers, which can contain an invalid case's descriptive name.
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

func TestLucidSelectionIsolation(t *testing.T) {
	for _, provider := range []string{"filesystem", "embedded"} {
		t.Run(provider, func(t *testing.T) {
			defer executionRegistryPanicAsFailure(t)
			config := runtimeGenerationResources(t)
			if provider == "embedded" {
				config = distribution.Files()
			}
			service := NewLogicemeService(config)
			var first Executable
			var firstBytes []byte
			for i, name := range []string{"select-nesting-hygiene", "select-bindings", "select-nesting-hygiene"} {
				r, dir := lucidRecord(t, provider+"-"+string(rune('A'+i)), "isolation")
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
		})
	}
}

// An explicit resource provider must account for the same frozen files as the
// embedded provider; successful compilation is not needed to prove selection.
func TestLucidSelectionAssetProviders(t *testing.T) {
	config := runtimeGenerationResources(t)
	filesystem := config
	config = distribution.Files()
	embedded := config
	for _, name := range []string{"define", "evaluate", "iterate"} {
		p := "builtins/segments/consonants/" + name + ".lu"
		left, e1 := fs.ReadFile(filesystem, p)
		right, e2 := fs.ReadFile(embedded, p)
		if e1 != nil || e2 != nil || !bytes.Equal(left, right) {
			t.Errorf("resource mismatch %s: filesystem=%v embedded=%v", p, e1, e2)
		}
	}
	p := "builtins/segments/vowels/evaluation.lu"
	left, e1 := fs.ReadFile(filesystem, p)
	right, e2 := fs.ReadFile(embedded, p)
	if e1 != nil || e2 != nil || !bytes.Equal(left, right) {
		t.Errorf("resource mismatch %s: filesystem=%v embedded=%v", p, e1, e2)
	}
	// The known valid original control also prevents an all-rejections path
	// from masquerading as validation coverage for the selected language.
	input := lucidRead(t, "communication.aml.txt")
	got, err := NewLogicemeService(config).Evaluate(input, GenerationOptions{})
	if err != nil {
		t.Fatal(err)
	}
	if len(got.ByteArrays) != 1 || len(got.ByteArrays[0]) != 5 || !strings.Contains(string(got.ByteArrays[0][0]), ".recv().await") {
		t.Fatal("valid Rust control did not produce the expected Rust application")
	}
}
