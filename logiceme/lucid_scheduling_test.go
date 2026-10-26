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

var lucidSchedulingAssets = []string{
	"builtins/segments/consonants/execute.lu",
	"builtins/segments/vowels/execution.lu",
}

func TestLucidDeferredRejected(t *testing.T) {
	cases := []struct{ name, form string }{
		{"v-deferred-true", "V"}, {"v-deferred-false", "V"},
		{"mv-deferred-true", "MV"}, {"mv-deferred-false", "MV"},
	}
	files, err := filepath.Glob(filepath.Join("testdata", "lucid", "scheduling-invalid", "*.aml.txt"))
	if err != nil || len(files) != len(cases) {
		t.Fatalf("invalid fixture inventory: files=%d cases=%d error=%v", len(files), len(cases), err)
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			defer executionRegistryPanicAsFailure(t)
			r, dir := lucidRecord(t, c.name, "validation")
			input := lucidRead(t, "scheduling-invalid/"+c.name+".aml.txt")
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
			for _, asset := range lucidSchedulingAssets {
				if !resources.reads[asset] {
					lucidIssue(t, r, "production selection: did not load %s", asset)
				}
			}
			if err == nil {
				lucidIssue(t, r, "production validation: accepted unsupported Deferred input")
				return
			}
			r.GenerationError = err.Error()
			if !errors.Is(err, lerrors.Unsupported) {
				lucidIssue(t, r, "production validation: expected Unsupported, got %v", err)
			}
			cause := err
			for errors.Unwrap(cause) != nil {
				cause = errors.Unwrap(cause)
			}
			want := "Rust " + c.form + ": Deferred is unsupported; Rust has no deferred execution operation"
			if cause.Error() != want {
				lucidIssue(t, r, "production validation: cause=%q want=%q", cause.Error(), want)
			}
		})
	}
}

func TestLucidSchedulingIsolation(t *testing.T) {
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
			for _, asset := range lucidSchedulingAssets {
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
			for i, name := range []string{"schedule-direct", "schedule-composed", "schedule-direct"} {
				r, dir := lucidRecord(t, "scheduling-"+provider+"-"+string(rune('A'+i)), "isolation")
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
				// Exact comparison also checks that B's import cannot leak into A.
				lucidCompare(t, r, dir, got, lucidExpectedBundle(t, name))
				if provider == "filesystem" {
					explicit = append(explicit, got)
				} else if len(explicit) != 3 || !reflect.DeepEqual(explicit[i], got) {
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
					if i == 2 && !reflect.DeepEqual(first, got) {
						lucidIssue(t, r, "production isolation: A/B/A changed output")
					}
				}
			}
			for _, asset := range lucidSchedulingAssets {
				if !resources.reads[asset] {
					t.Errorf("selected resource was not loaded: %s", asset)
				}
			}
		})
	}
}
