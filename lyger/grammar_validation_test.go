package Lyger

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func grammarValidationRules() map[string]grammeme {
	return map[string]grammeme{
		"object":                         {Structure: "CASE", Cases: map[string]string{"text": "`~`", "definitive": "ID"}},
		"subject":                        {Structure: "CASE", Cases: map[string]string{"declaration": `"ID" | PHRASE.GENERIC*`, "definitive": "ID"}},
		"phrase.generic":                 {Structure: "[PREPOSITION] | OBJECT"},
		"clause.generic.imperative":      {Structure: "[ADVERB]* | VERB | OBJECT"},
		"predicate.imperative":           {Structure: "CLAUSE.GENERIC.ROOT | [CLAUSE.IMPERATIVE]*"},
		"statement.objective.imperative": {Structure: "PREDICATE.IMPERATIVE"},
	}
}

func TestGrammarReferenceValidationAcceptsConfiguredGrammar(t *testing.T) {
	gram, err := loadGrammar(canonicalTestOptions())
	if err != nil {
		t.Fatal(err)
	}
	if err := validateGrammarReferences(gram.grammarMap); err != nil {
		t.Fatalf("current configured grammar does not validate: %v", err)
	}
	rules := grammarValidationRules()
	if err := validateGrammarReferences(rules); err != nil {
		t.Fatalf("dynamic clause selection or root-to-imperative reference failed: %v", err)
	}
	// Categories remain configuration-owned; validation resolves references
	// without restricting their names to the bundled language's vocabulary.
	rules["clause.reporting.active"] = grammeme{Structure: "VERB.ACTIVE | OBJECT"}
	rules["predicate.active"] = grammeme{Structure: "CLAUSE.REPORTING.ACTIVE"}
	rules["statement.subjective.active"] = grammeme{Structure: "SUBJECT | PREDICATE.ACTIVE"}
	if err := validateGrammarReferences(rules); err != nil {
		t.Fatalf("new configured category and its references were rejected: %v", err)
	}
}

func TestGrammarReferenceValidationRejectsInvalidConfiguration(t *testing.T) {
	for _, example := range []struct {
		name, rule, sequence, want string
		inCase                     bool
	}{
		{"missing optional rule", "subject", `"ID" | PHRASE.MISSING*`, `rule "subject" case "declaration": reference "PHRASE.MISSING*": rule "phrase.missing" is not configured`, true},
		{"missing case", "subject", "OBJECT.MISSING", `case "missing" is not configured on rule "object"`, true},
		{"unknown family", "subject", "OBJCET", "unknown grammatical element", true},
		{"missing family dispatch", "predicate.imperative", "CLAUSE.APPLICATIVE", "dynamic grammar family has no configured target", false},
		{"unclosed repetition", "subject", "[OBJECT", "malformed grammar reference", true},
		{"single quote", "subject", `"`, "malformed grammar reference", true},
		{"misplaced optional marker", "subject", "OBJECT*.TEXT", "malformed grammar reference", true},
		{"repeated type modifier", "subject", "#~#**", "malformed grammar reference", true},
		{"missing separator spaces", "subject", "ID|OBJECT", "malformed grammar reference", true},
		{"empty alternative", "subject", "", "empty grammar sequence", true},
		{"recursive case placeholder", "subject", "CASE", "CASE cannot repeat or occur inside an alternative", true},
		{"missing selector", "subject", "ID", "inconsistent CASE structure", false},
		{"multiple selectors", "subject", "CASE | CASE", "inconsistent CASE structure", false},
	} {
		t.Run(example.name, func(t *testing.T) {
			rules := grammarValidationRules()
			rule := rules[example.rule]
			if example.inCase {
				rule.Cases["declaration"] = example.sequence
			} else {
				rule.Structure = example.sequence
			}
			rules[example.rule] = rule
			first := ""
			for attempt := 0; attempt < 10; attempt++ {
				err := validateGrammarReferences(rules)
				if err == nil || !strings.Contains(err.Error(), example.want) {
					t.Fatalf("expected %q, got %v", example.want, err)
				}
				if attempt > 0 && err.Error() != first {
					t.Fatal("map iteration changed the configuration diagnostic")
				}
				first = err.Error()
			}
		})
	}
	for _, rules := range []map[string]grammeme{nil, {}, {"Unknown": {Structure: "ID"}}, {"subject": {Structure: "CASE"}}} {
		if err := validateGrammarReferences(rules); err == nil {
			t.Fatalf("incomplete configuration was accepted: %#v", rules)
		}
	}
}

func TestGrammarReferenceValidationRunsInCheckedLoader(t *testing.T) {
	root := t.TempDir()
	directory := filepath.Join(root, "lexicon", "eng")
	if err := os.MkdirAll(directory, 0700); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(directory, "grammar.json")
	rules := grammarValidationRules()
	rule := rules["subject"]
	rule.Cases["declaration"] = `"ID" | PHRASE.MISSING*`
	rules["subject"] = rule
	content, err := json.Marshal(rules)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, content, 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(directory, "vocabulary.json"), []byte(`{"nouns":{}}`), 0600); err != nil {
		t.Fatal(err)
	}
	gram, err := loadGrammar(Options{Language: "eng", AssetRoot: root, Resources: os.DirFS(root)})
	if gram != nil || err == nil || !strings.Contains(err.Error(), "lexicon/eng/grammar.json") || !strings.Contains(err.Error(), `rule "subject" case "declaration"`) {
		t.Fatalf("checked loader did not retain the configuration path and reference: grammar=%v error=%v", gram, err)
	}
}
