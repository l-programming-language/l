package Logiceme

import (
	"reflect"
	"strings"
	"testing"
)

func TestPreparedGlyphDefaultsRemainIndependent(t *testing.T) {
	system, err := loadSegmentalSystem(runtimeGenerationResources(t))
	if err != nil {
		t.Fatal(err)
	}
	prepare := func() Glyph {
		t.Helper()
		glyph := system.InitializeGlyph([]byte("E{|Statement: ~observe()~|}"), nil, nil)
		if err := prepareGlyph(&glyph, glyphContext{}, &generatedNames{counts: map[string]int{}}); err != nil {
			t.Fatal(err)
		}
		return glyph
	}
	keys := strings.Fields("Key Type Parameter Relation Shape AssignTarget Iterable Index RangeArity Default Lifetimes GenericTypes")
	check := func(glyph Glyph) {
		t.Helper()
		if len(glyph.Meta.Metas) < len(keys) || len(glyph.Meta.Index.Keys) < len(keys) || len(glyph.Meta.Index.Indexes) < len(keys) {
			t.Fatal("preparation lost default metadata")
		}
		for i, key := range keys {
			if got, want := *glyph.Meta.Metas[i], (MetaObj{Key: key, Value: "Void", Type: "String"}); got != want || glyph.Meta.Index.Keys[i] != key || glyph.Meta.Index.Indexes[i] != i {
				t.Fatalf("default %d = %#v, index %q/%d; want %#v", i, got, glyph.Meta.Index.Keys[i], glyph.Meta.Index.Indexes[i], want)
			}
		}
	}
	first, second := prepare(), prepare()
	check(first)
	check(second)
	before := second.Meta.GetMetaData()
	// Exercise both mutable records and index storage, rather than an update
	// method that could conceal sharing by replacing a pointer.
	first.Meta.Metas[0].Value = "changed"
	first.Meta.Index.Keys[0] = "changed"
	first.Meta.Index.Indexes[0] = 1
	if !reflect.DeepEqual(second.Meta.GetMetaData(), before) {
		t.Fatal("mutating one prepared glyph changed another glyph")
	}
	check(second)
	check(prepare())
}

func TestNativeIdentifierPattern(t *testing.T) {
	for _, input := range []string{"a", "Z9", "_name", "__", "r#type", "r#_", "type"} {
		if !isNativeIdentifier(input) {
			t.Errorf("rejected native identifier %q", input)
		}
	}
	for _, input := range []string{"", "_", "9name", "r#", "r##name", "r#9name", "a-b", "a::b", "a b", "a\n", "雪", "café"} {
		if isNativeIdentifier(input) {
			t.Errorf("accepted invalid native identifier %q", input)
		}
	}
}

func TestGlyphInputAccumulationPreservesParsing(t *testing.T) {
	system, err := loadSegmentalSystem(runtimeGenerationResources(t))
	if err != nil {
		t.Fatal(err)
	}
	for _, test := range []struct {
		source, want string
		invalid      bool
	}{
		{"E{outer{inner}tail}", "outer{inner}tail", false},
		{"E{first}{second}", "second", false},
		{"E{saved}{unfinished", "saved", false},
		{"E{saved}@", "saved", true},
		// Unframed inputs historically accumulate each byte as a rune. Framed
		// metadata is restored separately and must retain its original UTF-8.
		{"E{雪}", "\u00e9\u009b\u00aa", false},
		{"E{\xff}", "\u00ff", false},
		{"E{|Statement: ~雪~|}", "Statement: ~雪~", false},
	} {
		glyph := system.InitializeGlyph([]byte(test.source), nil, nil)
		if glyph.InputString != test.want || (glyph.initializationErr != nil) != test.invalid {
			t.Errorf("%q: input %q, error %v; want %q, invalid=%v", test.source, glyph.InputString, glyph.initializationErr, test.want, test.invalid)
		}
	}
	glyph := system.InitializeGlyph([]byte("W^(E{first{nested}}*W^E{second})"), nil, nil)
	if glyph.initializationErr != nil || len(glyph.Conjoinments) != 2 || len(glyph.Conjoinments[1].Conjoinments) != 1 {
		t.Fatalf("nested input composition failed: %#v, %v", glyph.Conjoinments, glyph.initializationErr)
	}
	if glyph.Conjoinments[0].InputString != "first{nested}" || glyph.Conjoinments[1].Conjoinments[0].InputString != "second" {
		t.Fatal("nested glyphs shared input accumulation")
	}
}
