package Logiceme

import (
	"errors"
	"reflect"
	"testing"

	"github.com/l-programming-language/l/lerrors"
)

func TestCompilationBuildsOwnResults(t *testing.T) {
	syllables := []Syllable{{Name: "E", Tensor: Tensor{Name: "original"}}}
	words := []Word{{Name: "word", Syllables: syllables, Tensor: Tensor{Name: "original"}}}
	sentence := Sentence{Name: "sentence", Words: words}
	sentence.Compile()
	if syllables[0].Tensor.Name != "original" || words[0].Tensor.Name != "original" {
		t.Fatal("building mutated the caller's word or syllable storage")
	}
	previousWords, previousTensor := sentence.Words, sentence.Tensor
	sentence.Compile()
	if !reflect.DeepEqual(sentence.Words, previousWords) || !reflect.DeepEqual(sentence.Tensor, previousTensor) {
		t.Fatal("repeated building changed the compiled structure")
	}
	sentence.Words[0].Name = "changed"
	sentence.Words[0].Syllables[0].Name = "changed"
	sentence.Words[0].Tensor.Dimensions[0].Name = "changed"
	sentence.Tensor.Dimensions[0].Name = "changed"
	if previousWords[0].Name != "word" || previousWords[0].Syllables[0].Name != "E" ||
		previousWords[0].Tensor.Dimensions[0].Name != "E" || previousTensor.Dimensions[0].Name != "word" {
		t.Fatal("repeated building reused a previous result's mutable storage")
	}
}

func TestCompilationEmptyResultsRemainNonNil(t *testing.T) {
	word := Word{}
	word.Build()
	if word.Syllables == nil || word.Tensor.Dimensions == nil || len(word.Syllables) != 0 || len(word.Tensor.Dimensions) != 0 {
		t.Fatal("empty word lost its non-nil empty results")
	}
	sentence := Sentence{}
	sentence.Compile()
	if sentence.Words == nil || sentence.Tensor.Dimensions == nil || len(sentence.Words) != 0 || len(sentence.Tensor.Dimensions) != 0 {
		t.Fatal("empty sentence lost its non-nil empty results")
	}
	tensor := Tensor{Content: []byte("previous")}
	if err := tensor.Compile(); err != nil {
		t.Fatal(err)
	}
	if tensor.Dimensions == nil || len(tensor.Dimensions) != 0 || tensor.Content != nil {
		t.Fatal("empty tensor changed dimensions or retained old content")
	}
}

func TestTensorCompilationPreservesOwnedResultsAndFailureState(t *testing.T) {
	system, err := loadSegmentalSystem(runtimeGenerationResources(t))
	if err != nil {
		t.Fatal(err)
	}
	first := Tensor{Order: 1, Type: "Scalar", Subject: []byte("E{|Statement: ~first()~|}"), Content: []byte("previous")}
	for _, failing := range []bool{false, true} {
		t.Run(map[bool]string{false: "success", true: "failure"}[failing], func(t *testing.T) {
			second := first
			if failing {
				second.Subject = []byte("@")
			}
			original := []Tensor{first, second}
			tensor := Tensor{SegmentalSystem: system, Dimensions: original, Content: []byte("previous parent")}
			for run := 0; run < 2; run++ {
				before := tensor.Dimensions
				err := tensor.Compile()
				want := "\nfirst()\n\nfirst()\n"
				if failing {
					want = "\nfirst()\n"
					if !errors.Is(err, lerrors.Unsupported) || !reflect.DeepEqual(tensor.Dimensions, original) || &tensor.Dimensions[0] != &original[0] {
						t.Fatalf("failure replaced dimensions or lost its cause: %v", err)
					}
				} else if err != nil || &tensor.Dimensions[0] == &before[0] {
					t.Fatalf("successful compilation reused dimensions or failed: %v", err)
				}
				if string(tensor.Content) != want || string(original[0].Content) != "previous" || string(original[1].Content) != "previous" {
					t.Fatalf("content or caller storage changed: %q", tensor.Content)
				}
			}
		})
	}
	first.SegmentalSystem = system
	first.Dimensions = []Tensor{{Name: "retained scalar metadata"}}
	before := first.Dimensions
	if err := first.Compile(); err != nil || &first.Dimensions[0] != &before[0] || first.Dimensions[0].Name != "retained scalar metadata" {
		t.Fatalf("scalar compilation replaced unrelated dimensions: %v", err)
	}
}
