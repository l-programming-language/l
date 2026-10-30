package Logiceme

import "testing"

func TestGlyphBindingIndexRetainsPublicSubstitutions(t *testing.T) {
	system, err := loadSegmentalSystem(runtimeGenerationResources(t))
	if err != nil {
		t.Fatal(err)
	}
	for _, test := range []struct {
		source, want string
		values       [][]string
	}{
		{"E{outer}", "Statement: ~supplied~", [][]string{{"outer", "inner"}, {"inner", "Statement: ~supplied~"}}},
		{"E{prefix{|Statement: ~supplied~|}suffix}", "prefix{Statement: ~supplied~}suffix", nil},
	} {
		glyph := system.InitializeGlyph([]byte(test.source), test.values, nil)
		if glyph.initializationErr != nil || glyph.InputString != test.want {
			t.Fatalf("public binding %q = %q, %v; want %q", test.source, glyph.InputString, glyph.initializationErr, test.want)
		}
	}
	glyph := system.InitializeGlyph([]byte("W^|supplied|"), nil, [][]string{{"supplied", "E{|Statement: ~nested~|}"}})
	if glyph.initializationErr != nil || len(glyph.Conjoinments) != 1 || glyph.Conjoinments[0].InputString != "Statement: ~nested~" {
		t.Fatalf("public group binding lost nested metadata: %#v, %v", glyph.Conjoinments, glyph.initializationErr)
	}
}
