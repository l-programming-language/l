package Logiceme

import (
	"bytes"
	"strings"
	"testing"
)

func TestRepeatedCompositionGroupKeepsLaterParentIntact(t *testing.T) {
	system, err := loadSegmentalSystem(runtimeGenerationResources(t))
	if err != nil {
		t.Fatal(err)
	}
	input := []byte("W^(BOG^GI^(E{|Statement: ~true~|}*W^(BOD^E{|Statement: ~first();~|}))*BOG^GI^(E{|Statement: ~true~|}*W^(BOG^GI^(E{|Statement: ~true~|}*W^(BOD^E{|Statement: ~first();~|}))*BOD^E{|Statement: ~second();~|})))")
	original := bytes.Clone(input)
	glyph := system.InitializeGlyph(input, nil, nil)
	if err := glyph.Compile(); err != nil {
		t.Fatal(err)
	}
	if got := strings.Fields(string(glyph.Content)); strings.Join(got, " ") != "if true { first(); } if true { if true { first(); } second(); }" {
		t.Fatalf("nested group order/occurrences changed: %q", glyph.Content)
	}
	if !bytes.Equal(input, original) {
		t.Fatal("group parsing mutated caller input")
	}
}
