package Lyger

import (
	"fmt"
	"testing"
)

func TestDecodeGrammarTermStrictBoundary(t *testing.T) {
	for _, tc := range []struct{ raw, kind, name, qualifier string }{
		{"ID.FOO", "identifier", "id.foo", "foo"},
		{"PRONOUN", "unknown", "pronoun", ""},
		{"CASE.FOO", "unknown", "case.foo", "foo"},
		{"VERB.REPORTING.ROOT", "lexical", "verb.reporting.root", "reporting.root"},
		{"OBJECT.自定_2", "grammar", "object.自定_2", "自定_2"},
	} {
		term, err := decodeGrammarTerm(tc.raw)
		if err != nil || term.Raw != tc.raw || term.Kind.String() != tc.kind || term.Name != tc.name || term.Qualifier != tc.qualifier {
			t.Errorf("decode %q = %+v, %v; want %+v", tc.raw, term, err, tc)
		}
	}
	for _, raw := range []string{`"ID.FOO"`, "[ID*]", "ID**", "[[ID]]", "[ID", "ID]", " ID", "ID ", "VERB..ROOT"} {
		term, err := decodeGrammarTerm(raw)
		want := fmt.Sprintf("malformed grammar reference %q", raw)
		if raw == `"ID.FOO"` {
			want = fmt.Sprintf("reference %q: only ID can be declared", raw)
		}
		if err == nil || err.Error() != want || term.Name != "" {
			t.Errorf("decode %q = %+v, %v; want %s", raw, term, err, want)
		}
	}
}

func TestGrammarEnumRelationships(t *testing.T) {
	for _, tc := range []struct {
		name  string
		arity int
	}{
		{"subject", 1}, {"object", 1}, {"value", 1}, {"condition", 1},
		{"phrase", 2}, {"conditional", 2}, {"predicate", 2}, {"clause", 3}, {"statement", 3},
	} {
		family, ok := parseGrammarFamily(tc.name)
		arity, valid := family.NameArity()
		if !ok || !valid || family.String() != tc.name || arity != tc.arity {
			t.Errorf("family %s = %s/%t, arity %d/%t", tc.name, family, ok, arity, valid)
		}
	}
	for _, tc := range []struct {
		name                string
		grammar, possessive bool
	}{
		{"noun", true, false}, {"verb", true, false}, {"adjective", true, true},
		{"adverb", true, true}, {"preposition", true, false}, {"article", true, false},
		{"conjunction", true, false}, {"evaluation", true, false}, {"pronoun", false, false},
	} {
		pos, ok := parsePartOfSpeech(tc.name)
		kind, supported := pos.GrammarKind()
		if !ok || supported != tc.grammar || pos.AcceptsPossessive() != tc.possessive || (supported && kind.String() != "lexical") {
			t.Errorf("POS %s = %s/%t, grammar %s/%t, possessive %t", tc.name, pos, ok, kind, supported, pos.AcceptsPossessive())
		}
	}
	for _, raw := range []string{"id", "case", "future_name", "NOUN"} {
		if _, ok := parsePartOfSpeech(raw); ok {
			t.Errorf("accepted non-POS name %q", raw)
		}
	}
	for _, raw := range []string{"case", "future_name", "CLAUSE"} {
		if _, ok := parseGrammarFamily(raw); ok {
			t.Errorf("accepted non-family name %q", raw)
		}
	}
	for _, n := range []uint8{0, 255} {
		if arity, ok := grammarFamily(n).NameArity(); ok || arity != 0 {
			t.Errorf("invalid family %d has arity %d/%t", n, arity, ok)
		}
		if _, ok := partOfSpeech(n).GrammarKind(); ok || partOfSpeech(n).AcceptsPossessive() {
			t.Errorf("invalid POS %d has a grammar mapping", n)
		}
	}
}

func TestGrammarEnumsCheckedNames(t *testing.T) {
	for _, tc := range []struct {
		name  string
		count int
		want  []string
		get   func(uint8) fmt.Stringer
	}{
		{"termKind", int(termKindCount), []string{"", "unknown", "grammar", "lexical", "type", "identifier", "case"}, func(n uint8) fmt.Stringer { return termKind(n) }},
		{"grammarFamily", int(grammarFamilyCount), []string{"", "condition", "conditional", "value", "subject", "object", "phrase", "clause", "predicate", "statement"}, func(n uint8) fmt.Stringer { return grammarFamily(n) }},
		{"partOfSpeech", int(partOfSpeechCount), []string{"", "noun", "verb", "adjective", "adverb", "preposition", "article", "conjunction", "evaluation", "pronoun"}, func(n uint8) fmt.Stringer { return partOfSpeech(n) }},
		{"typePattern", int(typePatternCount), []string{"", "`~`", "%~%", "?~?", "!~!", "#~#", "&~&", "{~}"}, func(n uint8) fmt.Stringer { return typePattern(n) }},
	} {
		if tc.count != len(tc.want) {
			t.Fatalf("%s count=%d, want %d", tc.name, tc.count, len(tc.want))
		}
		for n, want := range tc.want {
			if got := tc.get(uint8(n)).String(); got != want {
				t.Errorf("%s(%d)=%q, want %q", tc.name, n, got, want)
			}
		}
		for _, n := range []uint8{uint8(tc.count), 255} {
			if got, want := tc.get(n).String(), fmt.Sprintf("%s(%d)", tc.name, n); got != want {
				t.Errorf("%s, want %s", got, want)
			}
		}
	}
}

func TestGrammarMatcherRejectsMalformedOptionalBeforeDispatch(t *testing.T) {
	for _, tc := range []struct {
		raw   string
		build func(string) error
	}{
		{"[ID*]", (&valueGram{}).buildProcess},
		{"VERB*.ROOT", (&conditionGram{}).buildProcess},
	} {
		want := fmt.Sprintf("malformed grammar reference %q", tc.raw)
		if err := tc.build(tc.raw); err == nil || err.Error() != want {
			t.Errorf("matcher %q returned %v, want %s", tc.raw, err, want)
		}
	}
}

func TestGrammarTermValidationRetainsReferenceBoundaries(t *testing.T) {
	rules := grammarValidationRules()
	for _, raw := range []string{"ID.FOO", "[ID]*", "VERB.REPORTING.ROOT", "CLAUSE.GENERIC.ROOT"} {
		if _, err := validateGrammarSequence(rules, raw, false); err != nil {
			t.Errorf("valid reference %q: %v", raw, err)
		}
	}
	for _, tc := range []struct {
		raw, want string
		allowCase bool
	}{
		{"PRONOUN", `reference "PRONOUN": unknown grammatical element`, false},
		{"CASE.FOO", `reference "CASE.FOO": unknown grammatical element`, true},
		{"STATEMENT.OBJECTIVE.IMPERATIVE*", `reference "STATEMENT.OBJECTIVE.IMPERATIVE*": unknown grammatical element`, false},
		{"[CASE]*", "CASE cannot repeat or occur inside an alternative", true},
		{"CASE*", "CASE cannot repeat or occur inside an alternative", false},
	} {
		if _, err := validateGrammarSequence(rules, tc.raw, tc.allowCase); err == nil || err.Error() != tc.want {
			t.Errorf("reference %q returned %v, want %s", tc.raw, err, tc.want)
		}
	}
	for _, name := range []string{"object*", "[object]", "Object"} {
		configured := grammarValidationRules()
		configured[name] = grammeme{Structure: "ID"}
		want := fmt.Sprintf("rule %q has an invalid grammar family or category", name)
		if err := validateGrammarReferences(configured); err == nil || err.Error() != want {
			t.Errorf("rule %q returned %v, want %s", name, err, want)
		}
	}
	for _, raw := range []string{"PRONOUN*", "STATEMENT.OBJECTIVE.IMPERATIVE*", "FUTURE_ELEMENT*"} {
		want := fmt.Sprintf("%s: unknown grammar rule element %s", raw, raw)
		if err := (&valueGram{}).buildProcess(raw); err == nil || err.Error() != want {
			t.Errorf("optional reference %q returned %v, want %s", raw, err, want)
		}
	}
}
