package Logiceme

import (
	"bytes"
	"encoding/json"
	"io/fs"
	"path/filepath"
	"reflect"
	"testing"

	distribution "github.com/l-programming-language/l"
)

// Reusing MY's indexed Funcs must not leak records or loop bindings into another
// evaluation, mutate earlier bundles, or depend on the selected asset provider.
func TestLucidTraitSignatureIsolation(t *testing.T) {
	cases := []string{"trait-signature-outputs", "callable-output-void", "trait-signature-outputs"}
	assets := []string{
		"builtins/segments/consonants/instantiate.lu",
		"builtins/segments/vowels/instantiation.lu",
	}
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
			for _, asset := range assets {
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
				r, dir := lucidRecord(t, "trait-"+provider+"-"+string(rune('A'+i)), "isolation")
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
			for _, asset := range assets {
				if !resources.reads[asset] {
					t.Errorf("selected resource was not loaded: %s", asset)
				}
			}
		})
	}
}
