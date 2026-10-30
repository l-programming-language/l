package Logiceme

import "testing"

func TestGlyphInitializationRetainsFailedChildPrefix(t *testing.T) {
	system, err := loadSegmentalSystem(runtimeGenerationResources(t))
	if err != nil {
		t.Fatal(err)
	}
	for count, source := range []string{"E^@", "E^(E*@)"} {
		glyph := system.InitializeGlyph([]byte(source), nil, nil)
		if glyph.initializationErr == nil || glyph.initializationErr.Error() != `unknown glyph symbol "@" in @` {
			t.Fatalf("%s: initialization error = %v", source, glyph.initializationErr)
		}
		if len(glyph.Conjoinments) != count || (glyph.Conjoinments == nil) != (count == 0) {
			t.Fatalf("%s: successful child prefix = %#v", source, glyph.Conjoinments)
		}
		if count > 0 && glyph.Conjoinments[0].Name != "E" {
			t.Fatalf("%s: successful child source changed", source)
		}
	}
}

func TestGlyphInputSourcePreservesEmptyEdges(t *testing.T) {
	glyph := Glyph{Conjoinments: []Glyph{
		{InputString: "雪", Conjoinments: []Glyph{{}, {InputString: "\tβ "}}},
		{},
	}}
	if got, want := glyphInputSource(&glyph), " 雪  \tβ  "; got != want {
		t.Fatalf("input bytes = %q; want %q", got, want)
	}
}
