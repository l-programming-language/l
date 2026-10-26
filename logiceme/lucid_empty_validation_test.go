package Logiceme

import (
	"bytes"
	"errors"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/l-programming-language/l/lerrors"
)

// Empty native collections must not make supplied composition disappear. Each
// neutral complete document reaches the intended preparation rule; neither its
// application nor its file name can satisfy the leaf diagnostic assertion.
var lucidEmptyInvalid = []struct{ name, form, reason string }{
	{"hy-missing-props-parameters", "HY", "empty Props cannot discard composed child BA"},
	{"hy-empty-props-parameters", "HY", "empty Props cannot discard composed child BA"},
	{"my-empty-funcs-parameters", "MY", "empty Funcs cannot discard composed child BA"},
	{"my-missing-funcs-method", "MY", "empty Funcs cannot discard composed child FW"},
	{"hy-empty-sum", "HY", "empty Props requires absent Shape or Shape: Record; Shape: Sum is unsupported"},
}

func TestLucidEmptyInstantiationValidation(t *testing.T) {
	files, err := filepath.Glob(filepath.Join("testdata", "lucid", "empty-instantiation-invalid", "*.aml.txt"))
	if err != nil || len(files) != len(lucidEmptyInvalid) || len(files) != 5 {
		t.Fatalf("invalid empty-instantiation fixture inventory: files=%d cases=%d error=%v", len(files), len(lucidEmptyInvalid), err)
	}
	seen := map[string]bool{}
	for _, c := range lucidEmptyInvalid {
		if seen[c.name] {
			t.Fatalf("duplicate invalid fixture %s", c.name)
		}
		seen[c.name] = true
		t.Run(c.name, func(t *testing.T) {
			defer executionRegistryPanicAsFailure(t)
			r, dir := lucidRecord(t, "empty-"+c.name, "validation")
			input := lucidRead(t, "empty-instantiation-invalid/"+c.name+".aml.txt")
			before := bytes.Clone(input)
			lucidWrite(t, filepath.Join(dir, "input.aml.txt"), input)
			bundle, err := NewLogicemeService(runtimeGenerationResources(t)).Evaluate(input, GenerationOptions{IncludeRuntime: false})
			if !bytes.Equal(input, before) {
				lucidIssue(t, r, "production validation: mutated input")
			}
			if !reflect.DeepEqual(bundle, Executable{}) {
				lucidIssue(t, r, "production validation: returned nonzero bundle on malformed input")
			}
			lucidSaveBundle(t, r, dir, bundle)
			if err == nil {
				lucidIssue(t, r, "production validation: accepted malformed composition; want %s", c.reason)
				return
			}
			r.GenerationError = err.Error()
			if !errors.Is(err, lerrors.InvalidInput) {
				lucidIssue(t, r, "production validation: error %v is not InvalidInput", err)
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
