package Lyger

import (
	"errors"
	"testing"

	"github.com/l-programming-language/l/lerrors"
)

func TestLexiconIndex(t *testing.T) {
	noun := lexeme{Root: "signal", POS: "noun", Postfixes: map[string]string{"-s": "plural"}, Prefixes: map[string]string{"re-": "repeated"}}
	verb := lexeme{Root: "signal", POS: "verb", Postfixes: map[string]string{"-ed": "passive"}}
	lex := &lexicon{Nouns: map[string]lexeme{"message": noun}, Verbs: map[string]lexeme{"send": verb}}
	check := func(spelling, category, root string) {
		t.Helper()
		word, got, err := lex.lookup([]byte(spelling))
		if category == "" {
			if !errors.Is(err, lerrors.UnresolvedReference) {
				t.Fatalf("%s: missing word returned %v", spelling, err)
			}
			return
		}
		if err != nil || got != category || word.Root != root {
			t.Fatalf("%s: got %s/%s, %v; want %s/%s", spelling, got, word.Root, err, category, root)
		}
		word.Root = "changed use"
	}
	for _, indexed := range []bool{false, true} {
		release := func() {}
		if indexed {
			release = lex.indexLookups()
		}
		check("signal", "nouns.message.root", "signal")
		check("SIGNAL'S", "nouns.message.root.possessive", "signal")
		check("Signals", "nouns.message.plural", "signal")
		check("Resignal", "nouns.message.repeated", "signal")
		check("Signaled", "verbs.send.passive", "signal")
		check("missing", "", "")
		release()
	}
	if lex.lookupIndex != nil {
		t.Fatal("stage retained its configured-word index")
	}
	// Replacing a configured word between stages must alter lexical priority,
	// including same-size replacement and a newly indexed next stage.
	noun.Root = "notice"
	lex.Nouns["message"] = noun
	check("signal", "verbs.send.root", "signal")
	check("notice", "nouns.message.root", "notice")
	release := lex.indexLookups()
	check("signal", "verbs.send.root", "signal")
	check("notices", "nouns.message.plural", "notice")
	release()
}
