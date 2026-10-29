package Lyger

import "testing"

func TestGrammarAssociationsPreserveSourceOccurrences(t *testing.T) {
	first, lexical, child, last := &token{}, &token{}, &token{}, &token{}
	source := []*token{first, lexical, child, nil, first, last}
	want := []*token{first, nil, first, last}
	for name, node := range map[string]interface{ getUnassociatedTokens() []*token }{
		"condition":   &conditionGram{tokens: source, verb: lexical, values: []*valueGram{{tokens: []*token{child}}}},
		"conditional": &conditionalGram{tokens: source, adverb: lexical, conditions: []*conditionGram{{tokens: []*token{child}}}},
		"value":       &valueGram{tokens: source, noun: lexical, adjectives: []*token{child}},
		"subject":     &subjectGram{tokens: source, noun: lexical, phrase: &phraseGram{tokens: []*token{child}}},
		"object":      &objectGram{tokens: source, noun: lexical, name: &valueGram{tokens: []*token{child}}},
		"phrase":      &phraseGram{tokens: source, prepositions: []*token{lexical}, values: []*valueGram{{tokens: []*token{child}}}},
		"clause":      &clauseGram{tokens: source, verb: lexical, obj: &objectGram{tokens: []*token{child}}},
		"predicate":   &predicateGram{tokens: source, verb: lexical, clauses: []*clauseGram{{tokens: []*token{child}}}},
	} {
		t.Run(name, func(t *testing.T) {
			got := node.getUnassociatedTokens()
			if len(got) != len(want) {
				t.Fatalf("unassociated count = %d; want %d", len(got), len(want))
			}
			for index := range want {
				if got[index] != want[index] {
					t.Fatalf("occurrence %d = %p; want %p", index, got[index], want[index])
				}
			}
		})
	}
}
