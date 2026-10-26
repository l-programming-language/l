package Logiceme

import (
	"bytes"
	"errors"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/l-programming-language/l/lerrors"
)

// Neutral document identities ensure that context cannot accidentally satisfy
// the exact semantic leaf cause. Each stored input isolates one invalid use.
var lucidGenericsInvalid = []struct{ name, cause string }{
	{"fw-empty-lifetimes", "Rust FW: Lifetimes must be nonempty when supplied"},
	{"fw-whitespace-lifetimes", "Rust FW: Lifetimes must be nonempty when supplied"},
	{"hy-empty-generic-types", "Rust HY: GenericTypes must be nonempty when supplied"},
	{"hy-whitespace-generic-types", "Rust HY: GenericTypes must be nonempty when supplied"},
	{"bw-lifetimes", "Rust BW: unsupported metadata Lifetimes for this form"},
	{"bw-generic-types", "Rust BW: unsupported metadata GenericTypes for this form"},
	{"anonymous-ky-lifetimes", "Rust KY: Lifetimes requires a named KY with a local Key"},
	{"anonymous-ky-generic-types", "Rust KY: GenericTypes requires a named KY with a local Key"},
	{"e-lifetimes", "Rust E: unsupported metadata Lifetimes for this form"},
	{"o-generic-types", "Rust O: unsupported metadata GenericTypes for this form"},
	{"func-lifetimes", "Rust FW: unsupported Func field Lifetimes"},
	{"func-generic-types", "Rust FW: unsupported Func field GenericTypes"},
	{"funcs-lifetimes", "Rust MY: unsupported Func field Lifetimes"},
	{"funcs-generic-types", "Rust MY: unsupported Func field GenericTypes"},
	{"duplicate-lifetimes", "Rust FW: duplicate metadata field Lifetimes"},
	{"duplicate-generic-types", "Rust FW: duplicate metadata field GenericTypes"},
	{"class-lifetimes", "Rust Lifetimes is glyph-local; inherited file metadata is unsupported"},
	{"class-generic-types", "Rust GenericTypes is glyph-local; inherited file metadata is unsupported"},
	{"object-lifetimes", "Rust Lifetimes is glyph-local; inherited file metadata is unsupported"},
	{"object-generic-types-ea", "Rust GenericTypes is glyph-local; inherited file metadata is unsupported"},
}

func TestLucidGenericsValidation(t *testing.T) {
	files, err := filepath.Glob(filepath.Join("testdata", "lucid", "generics-invalid", "*.aml.txt"))
	if err != nil || len(files) != len(lucidGenericsInvalid) || len(files) != 20 {
		t.Fatalf("invalid generics fixture inventory: files=%d cases=%d error=%v", len(files), len(lucidGenericsInvalid), err)
	}
	seen := map[string]bool{}
	for _, c := range lucidGenericsInvalid {
		if seen[c.name] {
			t.Fatalf("duplicate invalid fixture %s", c.name)
		}
		seen[c.name] = true
		t.Run(c.name, func(t *testing.T) {
			defer executionRegistryPanicAsFailure(t)
			r, dir := lucidRecord(t, "generics-"+c.name, "validation")
			input := lucidRead(t, "generics-invalid/"+c.name+".aml.txt")
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
				lucidIssue(t, r, "production validation: accepted malformed composition; want %s", c.cause)
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
			if cause.Error() != c.cause {
				lucidIssue(t, r, "production validation: cause=%q want=%q", cause.Error(), c.cause)
			}
		})
	}
}
