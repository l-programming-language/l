package Lyger

import (
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

// These tests use configured grammar, independently of the deferred instruction
// implementation. A lexical success must not be reported as runtime evidence.
func phase1Grammar(t *testing.T) *grammar {
	t.Helper()
	g, err := loadGrammar(canonicalTestOptions())
	if err != nil {
		t.Fatal(err)
	}
	return g
}

func phase1Morphology(t *testing.T, root, family string, forms map[string]string) {
	t.Helper()
	g := phase1Grammar(t)
	for word, suffix := range forms {
		for _, spelling := range []string{word, strings.ToUpper(word), strings.ToUpper(word[:1]) + word[1:]} {
			tok := newToken([]byte(spelling), g)
			if err := tok.scan(); err != nil {
				t.Fatal(err)
			}
			if tok.mem == nil || tok.mem.Lexeme == nil {
				t.Fatalf("%s lacks lexical entry", spelling)
			}
			lex := tok.mem.Lexeme
			if tok.tType != "verbs."+family+"."+suffix || lex.Root != root || lex.POS != "verb" || lex.Type != "action."+family || string(tok.rawBytes) != spelling {
				t.Fatalf("%s: token=%s lexeme=%+v", spelling, tok.tType, lex)
			}
		}
	}
}

func phase1ParseGrammar(t *testing.T, source, role string) (*statementGram, error) {
	t.Helper()
	g := phase1Grammar(t)
	tokens := grammarTestTokens(t, source, g)
	// A grammar-only App reference stands in for binding; semantic tests below
	// compile actual declarations and check their canonical records separately.
	for _, tok := range tokens {
		if string(tok.rawBytes) == "Worker" || string(tok.rawBytes) == "Other" {
			tok.tType = "nouns.generic.root"
			tok.defined = true
		}
	}
	s, err := initStatementGram(role, tokens, g)
	if err == nil {
		err = s.build()
	}
	if err == nil && len(s.getUnassociatedTokens()) != 0 {
		t.Fatalf("accepted prefix with unconsumed tokens in %q", source)
	}
	return s, err
}

func phase1CanonicalAssets(t *testing.T) {
	t.Helper()
	g := phase1Grammar(t)
	if err := validateGrammarReferences(g.grammarMap); err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"grammar.json", "vocabulary.json"} {
		data, err := os.ReadFile(filepath.Join("..", "lexicon", "eng", name))
		if err != nil || len(data) == 0 {
			t.Fatalf("canonical resource %s unavailable: %v", name, err)
		}
	}
}

func TestPhase1Configuration(t *testing.T) {
	t.Run("checked-loading-and-reference-validation", func(t *testing.T) {
		g := phase1Grammar(t)
		if err := validateGrammarReferences(g.grammarMap); err != nil {
			t.Fatal(err)
		}
		want := lexeme{Root: "await", POS: "verb", Type: "action.suspensive", Postfixes: map[string]string{"-s": "active", "-ed": "passive", "-ing": "applicative"}, Value: "AWAIT(inputs[0])", Inputs: []string{"*.noun.app"}, Outputs: []string{"value"}}
		if !reflect.DeepEqual(g.lex.Verbs["suspensive"], want) {
			t.Fatalf("Await config differs: %+v", g.lex.Verbs["suspensive"])
		}
		for _, lex := range g.lex.Verbs {
			if lex.Root == "start" || lex.Root == "execute" {
				t.Fatalf("%s must remain a source-defined action", lex.Root)
			}
		}
		for _, form := range []string{"imperative", "active", "passive", "applicative"} {
			rule, ok := g.grammarMap["clause.suspensive."+form]
			if !ok {
				t.Fatalf("missing %s", form)
			}
			structure := "[ADVERB]* | VERB.SUSPENSIVE"
			if form != "imperative" {
				structure += "." + strings.ToUpper(form)
			}
			if form != "passive" {
				structure += " | OBJECT"
			}
			if rule.Structure != structure || rule.Category != "CLAUSE" || rule.Class != "SUSPENSIVE."+strings.ToUpper(form) || len(rule.Cases) != 0 {
				t.Fatalf("invalid %s clause: %+v", form, rule)
			}
			if got := g.grammarMap["predicate."+form].Cases["suspensive"]; got != "CLAUSE.SUSPENSIVE."+strings.ToUpper(form) {
				t.Fatalf("%s predicate dispatch=%s", form, got)
			}
		}
		for _, form := range []string{"objective.imperative", "subjective.active", "subjective.passive"} {
			suffix := strings.Split(form, ".")[1]
			if got := g.grammarMap["statement."+form].Cases["suspensive"]; got != "PREDICATE."+strings.ToUpper(suffix)+".SUSPENSIVE" {
				t.Fatalf("%s statement dispatch=%s", form, got)
			}
		}
		for key := range g.grammarMap {
			if strings.HasPrefix(key, "statement.") && strings.HasSuffix(key, ".applicative") {
				t.Fatal("standalone applicative statement must not be introduced")
			}
		}
	})
	t.Run("canonical-language-resources", phase1CanonicalAssets)
	t.Run("await-morphology", func(t *testing.T) {
		phase1Morphology(t, "await", "suspensive", map[string]string{"await": "root", "awaits": "active", "awaited": "passive", "awaiting": "applicative"})
	})
	t.Run("stop-morphology-unchanged", func(t *testing.T) {
		phase1Morphology(t, "stop", "interruptive", map[string]string{"stop": "root", "stops": "active", "stopped": "passive", "stopping": "applicative"})
	})
	t.Run("four-await-and-stop-forms", func(t *testing.T) {
		for _, tc := range []struct{ source, role string }{
			{"Await Worker", "verbs.suspensive.root"}, {"Worker awaits Other", "verbs.suspensive.active"}, {"Worker is awaited", "verbs.passive.root"},
			{"Create \"Join|-s|-ed|-ing|-er\" for Awaiting any App", "verbs.declarative.root"},
			{"Stop Worker", "verbs.interruptive.root"}, {"Worker stops Other", "verbs.interruptive.active"}, {"Worker is stopped", "verbs.passive.root"},
			{"Create \"Halt|-s|-ed|-ing|-er\" for Stopping any App", "verbs.declarative.root"},
		} {
			t.Run(tc.source, func(t *testing.T) {
				if _, err := phase1ParseGrammar(t, tc.source, tc.role); err != nil {
					t.Fatal(err)
				}
			})
		}
	})
	t.Run("invalid-forms-and-complete-consumption", func(t *testing.T) {
		for _, tc := range []struct{ source, role string }{
			{"Await", "verbs.suspensive.root"}, {"Await Worker extra", "verbs.suspensive.root"}, {"Worker is awaited Worker", "verbs.passive.root"},
			{"Awaiting Worker", "verbs.suspensive.applicative"}, {"Awaiter Worker", "verbs.suspensive.objective"},
			{"Create \"Join\" for Awaiting any App CONDITIONAL", "verbs.declarative.root"},
		} {
			t.Run(tc.source, func(t *testing.T) {
				if _, err := phase1ParseGrammar(t, tc.source, tc.role); err == nil {
					t.Fatalf("accepted malformed input %q", tc.source)
				}
			})
		}
	})
	t.Run("checked-loader-rejects-dangling-await-reference", func(t *testing.T) {
		g := phase1Grammar(t)
		delete(g.grammarMap, "clause.suspensive.applicative")
		data, err := json.Marshal(g.grammarMap)
		if err != nil {
			t.Fatal(err)
		}
		root := t.TempDir()
		dir := filepath.Join(root, "lexicon", "eng")
		if err := os.MkdirAll(dir, 0700); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(dir, "grammar.json"), data, 0600); err != nil {
			t.Fatal(err)
		}
		vocabulary, err := os.ReadFile("../lexicon/eng/vocabulary.json")
		if err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(dir, "vocabulary.json"), vocabulary, 0600); err != nil {
			t.Fatal(err)
		}
		got, err := loadGrammar(Options{Language: "eng", AssetRoot: root, Resources: os.DirFS(root)})
		if got != nil || err == nil || !strings.Contains(err.Error(), "suspensive.applicative") {
			t.Fatalf("checked loading accepted dangling suspensive reference: grammar=%v error=%v", got, err)
		}
	})
}
