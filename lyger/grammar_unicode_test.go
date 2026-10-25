package Lyger

import (
	"errors"
	"testing"

	"github.com/l-programming-language/l/lerrors"
)

func TestClauseTypingRetainsUnicodeLowercaseIdentity(t *testing.T) {
	for _, tc := range []struct {
		name, expected, actual string
		accept                 bool
	}{
		{"ASCII case", "imperative", "Imperative", true},
		{"sigma case", "σ", "Σ", true},
		{"final sigma stays distinct", "ς", "Σ", false},
		{"long s stays distinct", "ſ", "S", false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			clause := clauseGram{tokens: []*token{{tType: "verbs." + tc.actual + ".active"}}}
			err := clause.loadTyping("clause." + tc.expected)
			if tc.accept {
				if err != nil {
					t.Fatalf("case-equivalent grammar name rejected: %v", err)
				}
			} else if !errors.Is(err, lerrors.Syntax) {
				t.Fatalf("distinct lowercase grammar name returned %v, want syntax mismatch", err)
			}
		})
	}
}
