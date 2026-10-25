package Lyger

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCheckedLanguageLoading(t *testing.T) {
	config := canonicalTestOptions()
	gram, err := loadGrammar(config)
	if err != nil {
		t.Fatal(err)
	}
	if gram == nil || gram.lex == nil || len(gram.lex.Nouns) == 0 || len(gram.grammarMap) == 0 {
		t.Fatalf("checked loader returned an incomplete language: %#v", gram)
	}
	for _, config := range []Options{{Language: " \t", AssetRoot: ".."}, {Language: "../eng"}, {Language: "eng/sub"}} {
		if gram, err := loadGrammar(config); gram != nil || err == nil {
			t.Errorf("invalid configuration produced grammar=%#v, error=%v", gram, err)
		}
	}
	if lex, err := loadLexicon(config.Resources, ""); lex != nil || err == nil {
		t.Fatalf("empty language produced lexicon=%#v, error=%v", lex, err)
	}
	missingLanguage := "not-installed-" + filepath.Base(t.TempDir())
	path := filepath.Join("lexicon", missingLanguage, "vocabulary.json")
	if gram, err := loadGrammar(Options{AssetRoot: "..", Language: missingLanguage}); gram != nil || !errors.Is(err, os.ErrNotExist) || !strings.Contains(err.Error(), filepath.Base(path)) {
		t.Fatalf("missing language should retain its path and filesystem cause: grammar=%#v, error=%v", gram, err)
	}
}

func TestReadLanguageObjects(t *testing.T) {
	directory := t.TempDir()
	vocabulary := filepath.Join(directory, "vocabulary.json")
	if err := os.WriteFile(vocabulary, []byte(`{"nouns":{"sample":{"root":"sample","pos":"noun","type":"generic"}}}`), 0600); err != nil {
		t.Fatal(err)
	}
	var lex lexicon
	if err := readLanguageJSON(os.DirFS(filepath.Dir(vocabulary)), filepath.Base(vocabulary), &lex); err != nil {
		t.Fatal(err)
	}
	if word := lex.Nouns["sample"]; word.Root != "sample" || word.POS != "noun" || word.Type != "generic" {
		t.Fatalf("vocabulary content was lost: %#v", lex)
	}
	grammarFile := filepath.Join(directory, "grammar.json")
	if err := os.WriteFile(grammarFile, []byte(`{"statement.example":{"id":"example","structure":"SUBJECT | PREDICATE","cases":{"one":"CASE"}}}`), 0600); err != nil {
		t.Fatal(err)
	}
	var rules map[string]grammeme
	if err := readLanguageJSON(os.DirFS(filepath.Dir(grammarFile)), filepath.Base(grammarFile), &rules); err != nil {
		t.Fatal(err)
	}
	if rule := rules["statement.example"]; rule.ID != "example" || rule.Structure != "SUBJECT | PREDICATE" || rule.Cases["one"] != "CASE" {
		t.Fatalf("grammar content was lost: %#v", rules)
	}
}

func TestReadLanguageJSONErrors(t *testing.T) {
	for _, tc := range []struct{ name, content string }{
		{"empty", ""}, {"null", "null"}, {"array", "[]"}, {"truncated", `{"nouns":`},
		{"wrong field shape", `{"nouns":42}`}, {"trailing text", `{} trailing`}, {"second object", `{} {}`},
	} {
		t.Run(tc.name, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "invalid.json")
			if err := os.WriteFile(path, []byte(tc.content), 0600); err != nil {
				t.Fatal(err)
			}
			var lex lexicon
			if err := readLanguageJSON(os.DirFS(filepath.Dir(path)), filepath.Base(path), &lex); err == nil || !strings.Contains(err.Error(), filepath.Base(path)) || !strings.Contains(err.Error(), "decode language definition") {
				t.Fatalf("malformed language JSON should return a path/phase error, got %v", err)
			}
		})
	}
	path := filepath.Join(t.TempDir(), "missing.json")
	var rules map[string]grammeme
	if err := readLanguageJSON(os.DirFS(filepath.Dir(path)), filepath.Base(path), &rules); !errors.Is(err, os.ErrNotExist) || !strings.Contains(err.Error(), filepath.Base(path)) {
		t.Fatalf("missing language definition lost underlying error: %v", err)
	}
}

func TestCompilationLoaderValidation(t *testing.T) {
	for _, tc := range []struct {
		name    string
		config  Options
		missing bool
	}{
		{"invalid language", Options{Language: "../eng"}, false},
		{"blank language", Options{Language: " \t", AssetRoot: ".."}, false},
		{"unavailable language", Options{AssetRoot: "..", Language: "not-installed-loader-test"}, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			service := NewLygerService(tc.config)
			source := []byte("\"value\" is `text`.")
			compilation := service.NewCompilation(source, "loader-source.l")
			if compilation == nil || compilation.Error == nil {
				t.Fatalf("invalid language setup did not return a compilation error: %#v", compilation)
			}
			if compilation.Gram != nil || compilation.Program != nil || compilation.CompiledBytes != nil {
				t.Fatal("invalid setup continued into preprocessing or semantic construction")
			}
			result, err := service.Compile(source, "loader-source.l")
			if result != nil || err == nil {
				t.Fatalf("Compile should return no result after loader failure: %#v, %v", result, err)
			}
			if tc.missing && !errors.Is(err, os.ErrNotExist) {
				t.Fatalf("Compile lost the underlying language file error: %v", err)
			}
		})
	}
}
