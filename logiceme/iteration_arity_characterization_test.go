package Logiceme

import (
	"bytes"
	"errors"
	"reflect"
	"strings"
	"testing"

	"github.com/l-programming-language/l/lerrors"
)

func TestIterationArityPublicBoundary(t *testing.T) {
	// Declaration.Default contributes the header with an empty Headers loop.
	// Operate.Definition supplies the function/Operation boundary; BOP supplies
	// one leading newline, and its trailing newline joins the W definition loop.
	const sourcePrefix = "\n\n// package main\n\n\n\n\n" + "fn main( ) -> (()) {\n    \n        " + "\n"
	const sourceSuffix = "\n\n" + "\n    \n}\n"
	// Literal separators around the selected Iterate conditions contribute
	// leading/trailing newline counts 3/4, 4/3, and 6/4 respectively. The loop
	// body is empty W; these fragments include the loop's own final newline.
	const wildcardSource = sourcePrefix + "\n\n\n" + "for _ in items {\n    \n}\n" + "\n\n\n\n" + sourceSuffix
	const singleSource = sourcePrefix + "\n\n\n\n" + "for value in items {\n    \n}\n" + "\n\n\n" + sourceSuffix
	const pairSource = sourcePrefix + "\n\n\n\n\n\n" + "for (_, value) in items {\n    \n}\n" + "\n\n\n\n" + sourceSuffix
	const arityCause = "Rust X: RangeArity requires 0, 1 or 2; CX requires 0 or 1"
	service := NewLogicemeService(runtimeGenerationResources(t))
	for _, test := range []struct {
		name, arity, children, source, cause string
		key                                  bool
	}{
		{name: "zero", arity: "0", children: "^W", source: wildcardSource},
		{name: "one", arity: "1", key: true, children: "^W", source: singleSource},
		{name: "two", arity: "2", key: true, children: "^W", source: pairSource},
		{name: "default", key: true, children: "^W", source: pairSource},
		// Atoi accepts these spellings, but the template compares their original
		// text with "0" and "1". Retain the existing pair branch and its bytes.
		{name: "explicit plus", arity: "+1", key: true, children: "^W", source: pairSource},
		{name: "leading zero", arity: "01", key: true, children: "^W", source: pairSource},
		{name: "negative", arity: "-1", children: "^W", cause: arityCause},
		{name: "above maximum", arity: "3", children: "^W", cause: arityCause},
		{name: "nonnumeric", arity: "invalid", children: "^W", cause: arityCause},
		{name: "overflow", arity: strings.Repeat("9", 100), children: "^W", cause: arityCause},
		{name: "body precedes arity", arity: "invalid", cause: "Rust X: Body requires 1..1 Operation children"},
		{name: "children precede arity", arity: "invalid", children: "^(E{|Statement: ~true~|}*W)", cause: "Rust X: iterator does not accept condition children"},
	} {
		t.Run(test.name, func(t *testing.T) {
			metadata := "Iterable: ~items~"
			if test.arity != "" {
				metadata += ", RangeArity: ~" + test.arity + "~"
			}
			if test.key {
				metadata += ", Key: ~value~"
			}
			input := []byte(`<PROGRAM><STRUCTURE><BOOK "arity_probe"><OUTLINE>
<JSON>{"AppMeta":{"AppGenDir":"/src"}}</JSON>
<JSON>{"Services":[{"Name":"main","Location":"/src/","GenericObjects":["main"]}]}</JSON>
<JSON>{"Objects":[{"Name":"main","Type":"root"}]}</JSON>
</OUTLINE><SETTINGS></SETTINGS><CONTENT><CHAPTER "main"><SENTENCE "main"><|W|>
<|S|>` + "`A`" + `</S|>
<|S|>` + "`FW{|Key: ~main~, Outputs: ~&[()]&~|}^W^BOP^X{|" + metadata + "|}" + test.children + "`" + `</S|>
</W|></SENTENCE></CHAPTER></CONTENT></BOOK></STRUCTURE></PROGRAM>`)
			original := bytes.Clone(input)
			bundle, err := service.Evaluate(input)
			if !bytes.Equal(input, original) {
				t.Fatal("iteration generation mutated the complete AML input")
			}
			if test.cause != "" {
				if !errors.Is(err, lerrors.InvalidInput) {
					t.Fatalf("iteration error = %v; want InvalidInput", err)
				}
				if !reflect.DeepEqual(bundle, Executable{}) {
					t.Fatalf("invalid iteration returned a partial bundle: %#v", bundle)
				}
				if !strings.Contains(err.Error(), "glyph bytes [") {
					t.Errorf("iteration diagnostic lost its glyph context: %v", err)
				}
				cause := err
				for errors.Unwrap(cause) != nil {
					cause = errors.Unwrap(cause)
				}
				if cause.Error() != test.cause {
					t.Errorf("iteration cause = %q; want %q", cause.Error(), test.cause)
				}
				return
			}
			if err != nil {
				t.Fatalf("accepted iteration spelling %q rejected: %v", test.arity, err)
			}
			wantPaths := [][]string{{"/src/main.rs", "/Cargo.toml", "/src/config.rs", "/start.sh", "/stop.sh"}}
			if !reflect.DeepEqual(bundle.MapArrays, wantPaths) || len(bundle.ByteArrays) != 1 || len(bundle.ByteArrays[0]) != len(wantPaths[0]) {
				t.Fatalf("iteration bundle inventory = %#v; want %#v", bundle.MapArrays, wantPaths)
			}
			if got := string(bundle.ByteArrays[0][0]); got != test.source {
				t.Errorf("iteration source = %q; want exact bytes %q", got, test.source)
			}
		})
	}
}
