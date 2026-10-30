package Logiceme

import (
	"math/rand"
	"reflect"
	"testing"
)

func TestMetaInsertFramesMatchOriginalRegex(t *testing.T) {
	original, err := LoadRegex("MetaDefs", "")
	if err != nil {
		t.Fatal(err)
	}
	inputs := []string{
		"", "{|", "|}", "{||}", "{|{|x|}|}", "{|one|}{|two|}",
		"x {|α\r\nβ\rx\ny\xff|} after", "{|same|} {|same|}", "{|first|} {|unclosed",
	}
	random := rand.New(rand.NewSource(1)) //nolint:gosec // Deterministic parser differential corpus, not security randomness.
	alphabet := []byte("{|}x\r\n\xff")
	for sample := 0; sample < 500; sample++ {
		input := make([]byte, random.Intn(96))
		for index := range input {
			input[index] = alphabet[random.Intn(len(alphabet))]
		}
		inputs = append(inputs, string(input))
	}
	for _, input := range inputs {
		bytes := []byte(input)
		if got, want := metaInsertFrames(bytes), original.Parse(bytes); !reflect.DeepEqual(got, want) {
			t.Fatalf("%q: frames %q, want %q", input, got, want)
		}
	}
}
