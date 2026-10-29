package Lyger

import "testing"

func TestTokenCategoriesFollowCurrentClassification(t *testing.T) {
	var tok token
	for _, tc := range []struct {
		typeName, pos, singular, caseName string
	}{
		{"", "", "", ""},
		{"nouns", "nouns", "noun", ""},
		{"verbs.", "verbs", "verb", ""},
		{".leading", "", "", "leading"},
		{"primitive.unknown", "unknown", "unknown", "unknown"},
		{"nouns.NotFound.possessive", "unknown", "unknown", "NotFound.possessive"},
		{"nouns.common.root.possessive", "nouns", "noun", "common.possessive"},
		{"verbs.active.root.root", "verbs", "verb", "active.root"},
		{"verbs.root", "verbs", "verb", "root"},
		{"nouns..root.root", "nouns", "noun", ".root"},
		{"雪.kind", "雪", "雪", "kind"},
	} {
		tok.tType = tc.typeName
		if got := tok.getPOS(false); got != tc.pos {
			t.Errorf("%q part of speech = %q, want %q", tc.typeName, got, tc.pos)
		}
		if got := tok.getPOS(true); got != tc.singular {
			t.Errorf("%q singular part of speech = %q, want %q", tc.typeName, got, tc.singular)
		}
		if got := tok.getCase(); got != tc.caseName {
			t.Errorf("%q case = %q, want %q", tc.typeName, got, tc.caseName)
		}
	}
}
