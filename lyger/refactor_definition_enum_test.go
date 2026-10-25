package Lyger

import (
	"strings"
	"testing"
)

func TestRefactorDefinitionEnumNamesAndBounds(t *testing.T) {
	cases := []struct {
		kind definitionArgKind
		name string
	}{
		{definitionArgInvalid, ""},
		{definitionArgText, "Text"},
		{definitionArgNumber, "Number"},
		{definitionArgTruth, "Truth"},
		{definitionArgNull, "Null"},
		{definitionArgSyllable, "Syllable"},
		{definitionArgSyllables, "Syllables"},
		{definitionArgProps, "Props"},
		{definitionArgTexts, "Texts"},
		{definitionArgDefinition, "Definition"},
	}
	if len(cases) != int(definitionArgKindCount) {
		t.Fatal("enum name expectations do not cover every declared variant")
	}
	for _, tc := range cases {
		if got := tc.kind.String(); got != tc.name {
			t.Errorf("kind %d String = %q, want %q", tc.kind, got, tc.name)
		}
		if tc.kind.Valid() != (tc.kind != definitionArgInvalid) {
			t.Errorf("kind %d has incorrect validity", tc.kind)
		}
	}
	for _, tc := range []struct {
		kind definitionArgKind
		name string
	}{
		{definitionArgKindCount, "definitionArgKind(10)"},
		{definitionArgKind(255), "definitionArgKind(255)"},
	} {
		if tc.kind.Valid() || tc.kind.String() != tc.name {
			t.Errorf("invalid kind %d: Valid=%v String=%q", tc.kind, tc.kind.Valid(), tc.kind.String())
		}
	}
}

func TestRefactorDefinitionEnumExactAndFoldedBoundaries(t *testing.T) {
	for _, tc := range []struct {
		name string
		kind definitionArgKind
	}{
		{"Text", definitionArgText}, {"Number", definitionArgNumber},
		{"Truth", definitionArgTruth}, {"Null", definitionArgNull},
		{"Syllable", definitionArgSyllable}, {"Syllables", definitionArgSyllables},
		{"Props", definitionArgProps}, {"Texts", definitionArgTexts},
		{"Definition", definitionArgDefinition},
	} {
		if got, ok := parseDefinitionArgKindExact(tc.name); !ok || got != tc.kind {
			t.Errorf("exact %q = %v/%v", tc.name, got, ok)
		}
		for _, raw := range []string{tc.name, strings.ToLower(tc.name), strings.ToUpper(tc.name)} {
			if got, ok := parseDefinitionArgKindFolded(raw); !ok || got != tc.kind {
				t.Errorf("folded %q = %v/%v", raw, got, ok)
			}
		}
		if got, ok := parseDefinitionArgKindExact(strings.ToLower(tc.name)); ok || got != definitionArgInvalid {
			t.Errorf("exact boundary accepted lowercase %q", tc.name)
		}
	}
	for _, raw := range []string{"", "Thing", "SourceSpecific", " Text", "Text ", "Text\n", "ſyllable"} {
		for _, parse := range []func(string) (definitionArgKind, bool){parseDefinitionArgKindExact, parseDefinitionArgKindFolded} {
			if got, ok := parse(raw); ok || got != definitionArgInvalid {
				t.Errorf("unknown boundary %q = %v/%v", raw, got, ok)
			}
		}
	}
}

func TestRefactorDefinitionValueTypeMapping(t *testing.T) {
	for _, tc := range []struct {
		value valueType
		want  definitionArgKind
	}{
		{valueText, definitionArgText},
		{valueNumber, definitionArgNumber},
		{valueTruth, definitionArgTruth},
	} {
		if actual, ok := tc.value.DefinitionArgKind(); !ok || actual != tc.want {
			t.Errorf("%s definition kind = %s/%v, want %s", tc.value, actual, ok, tc.want)
		}
	}
	for value := valueInvalid; value < valueTypeCount; value++ {
		if value == valueText || value == valueNumber || value == valueTruth {
			continue
		}
		if actual, ok := value.DefinitionArgKind(); ok || actual != definitionArgInvalid {
			t.Errorf("unsupported stored value %s maps to %s/%v", value, actual, ok)
		}
	}
	for _, value := range []valueType{valueTypeCount, valueType(255)} {
		if actual, ok := value.DefinitionArgKind(); ok || actual != definitionArgInvalid {
			t.Errorf("out-of-range value %d maps to %s/%v", value, actual, ok)
		}
	}
}

func TestRefactorDefinitionArgumentReceiver(t *testing.T) {
	argument := DefinitionArgument{Kind: "tExT", Text: "value"}
	actual, err := argument.renderFor(DefinitionSlot{Name: "Value", Type: "Thing"})
	if err != nil || actual != `"value"` || argument.Kind != "tExT" || argument.Text != "value" {
		t.Fatalf("argument receiver = %q/%v; source argument = %#v", actual, err, argument)
	}
	if _, err := (DefinitionArgument{Kind: "ſyllable", Text: "A"}).renderFor(DefinitionSlot{Name: "Value"}); err == nil {
		t.Fatal("rendering changed lowercase comparison into broader Unicode folding")
	}
}
