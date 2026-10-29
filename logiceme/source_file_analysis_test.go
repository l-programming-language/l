package Logiceme

import (
	"errors"
	"testing"

	"github.com/l-programming-language/l/lerrors"
)

func TestObjectAnalysisDoesNotRetainPublicSourceMutation(t *testing.T) {
	system, err := loadSegmentalSystem(runtimeGenerationResources(t))
	if err != nil {
		t.Fatal(err)
	}
	object := Object{Name: "literal"}
	for _, content := range []string{"first", "second"} {
		object.Sentence = Sentence{Words: []Word{{Syllables: []Syllable{{Name: "EA{|Contents: ~" + content + "~|}", Count: 1, Exponent: 1}}}}}
		object.Compile()
		compiled, err := object.Assemble(MetaDict{}, system)
		if err != nil || string(compiled.Content) != content {
			t.Fatalf("changed public Object = %q, %v; want %q", compiled.Content, err, content)
		}
	}
	object.prepared = &sourceFileAnalysis{contents: []byte("previously checked"), extension: "rs", literal: true}
	var inherited MetaDict
	inherited.Add("GenericTypes", "T", "String")
	if _, err := object.Assemble(inherited, system); !errors.Is(err, lerrors.InvalidInput) {
		t.Fatalf("prepared source hid inherited generic input: %v", err)
	}
}
