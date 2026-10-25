package Logiceme

import (
	"bytes"
	"errors"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/l-programming-language/l/lerrors"
)

// Every complete document uses neutral application/file identities. Only the
// classified leaf cause may satisfy its keyed-composition diagnostic contract.
var lucidKeyedInvalid = []struct{ name, reason string }{
	{"no-entries", "keyed selection requires at least one Calculation/Operation pair"},
	{"pattern-without-body", "keyed selection requires one Operation per Calculation"},
	{"body-without-pattern", "keyed selection requires one Operation per Calculation"},
	{"excess-patterns", "keyed selection requires one Operation per Calculation"},
	{"excess-bodies", "keyed selection requires one Operation per Calculation"},
	{"nonterminal-fallback", "keyed selection fallback must be terminal"},
	{"duplicate-fallback", "keyed selection has multiple fallback patterns"},
	{"nonterminal-composed-fallback", "keyed selection fallback must be terminal"},
	{"duplicate-composed-fallback", "keyed selection has multiple fallback patterns"},
	{"unexpected-child", "unexpected child O: wrong category or role"},
	{"empty-key", "keyed selection requires a nonempty Key"},
	{"whitespace-key", "keyed selection requires a nonempty Key"},
	{"empty-pattern", "keyed selection requires a nonempty pattern at entry 1"},
	{"whitespace-pattern", "keyed selection requires a nonempty pattern at entry 1"},
}

func TestLucidKeyedValidation(t *testing.T) {
	files, err := filepath.Glob(filepath.Join("testdata", "lucid", "keyed-invalid", "*.aml.txt"))
	if err != nil || len(files) != len(lucidKeyedInvalid) || len(files) != 14 {
		t.Fatalf("invalid keyed fixture inventory: files=%d cases=%d error=%v", len(files), len(lucidKeyedInvalid), err)
	}
	seen := map[string]bool{}
	for _, c := range lucidKeyedInvalid {
		if seen[c.name] {
			t.Fatalf("duplicate invalid fixture %s", c.name)
		}
		seen[c.name] = true
		t.Run(c.name, func(t *testing.T) {
			defer executionRegistryPanicAsFailure(t)
			r, dir := lucidRecord(t, "keyed-"+c.name, "validation")
			input := lucidRead(t, "keyed-invalid/"+c.name+".aml.txt")
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
			want := "Rust GI: " + c.reason
			if cause.Error() != want {
				lucidIssue(t, r, "production validation: cause=%q want=%q", cause.Error(), want)
			}
		})
	}
}
