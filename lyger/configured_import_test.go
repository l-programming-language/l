package Lyger

import (
	"bytes"
	"encoding/json"
	"fmt"
	distribution "github.com/l-programming-language/l"
	"io/fs"
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"strings"
	"testing"
	"testing/fstest"
)

func configuredImportAssets(t *testing.T, root string, words [3]string) Options {
	t.Helper()
	resources := fstest.MapFS{}
	if err := fs.WalkDir(distribution.Files(), ".", func(name string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() {
			return nil
		}
		raw, err := fs.ReadFile(distribution.Files(), name)
		if err == nil {
			resources[name] = &fstest.MapFile{Data: raw}
		}
		return err
	}); err != nil {
		t.Fatal(err)
	}

	for name, file := range resources {
		if strings.HasSuffix(name, ".l") {
			file.Data = configuredSourceWords(t, file.Data, words)
		}
	}
	directory := filepath.Join(root, "assets", "lexicon", "eng")
	if err := os.MkdirAll(directory, 0700); err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"grammar.json", "vocabulary.json"} {
		raw, err := os.ReadFile(filepath.Join("..", "lexicon", "eng", name))
		if err != nil {
			t.Fatal(err)
		}
		if name == "vocabulary.json" {
			var lex lexicon
			if err = json.Unmarshal(raw, &lex); err != nil {
				t.Fatal(err)
			}
			for i, item := range []struct {
				entries map[string]lexeme
				key     string
			}{{lex.Verbs, "associative"}, {lex.Prepositions, "conjoinment.objective"}, {lex.Prepositions, "definement.declarative"}} {
				entry, exists := item.entries[item.key]
				if !exists {
					t.Fatalf("fixture vocabulary has no %s", item.key)
				}
				entry.Root = strings.ToLower(words[i])
				item.entries[item.key] = entry
			}
			raw, err = json.Marshal(lex)
			if err != nil {
				t.Fatal(err)
			}
		}
		resources["lexicon/eng/"+name] = &fstest.MapFile{Data: raw}
		if err = os.WriteFile(filepath.Join(directory, name), raw, 0600); err != nil {
			t.Fatal(err)
		}
	}
	return Options{Language: "eng", Resources: resources}
}

var configuredGrammarWords = regexp.MustCompile(`(?i)\b(use|from|as|using)\b`)

// A configured distribution translates its source grammar, including private
// dictionary imports. Literal labels, native leaves and data retain their bytes.
func configuredSourceWords(t *testing.T, source []byte, words [3]string) []byte {
	t.Helper()
	spellings := map[string]string{"use": strings.ToLower(words[0]), "from": words[1], "as": words[2], "using": processLexFix(strings.ToLower(words[0]), "--ing", "postfix")}
	replace := func(source string) string {
		return configuredGrammarWords.ReplaceAllStringFunc(source, func(word string) string {
			if word == "Use" {
				return words[0]
			}
			return spellings[strings.ToLower(word)]
		})
	}
	var translated strings.Builder
	start := 0
	for at := 0; at < len(source); at++ {
		if source[at] != '`' && source[at] != '"' {
			continue
		}
		end, err := quotedEnd(source, at)
		if err != nil {
			t.Fatal(err)
		}
		translated.WriteString(replace(string(source[start:at])))
		translated.Write(source[at:end])
		start, at = end, end-1
	}
	translated.WriteString(replace(string(source[start:])))
	return []byte(translated.String())
}

func TestConfiguredImportVocabulary(t *testing.T) {
	for _, words := range [][3]string{
		{"Use", "from", "as"}, {"Adopt", "from", "as"}, {"Use", "via", "as"},
		{"Use", "from", "named"}, {"AdOpT", "vIa", "NaMeD"},
	} {
		t.Run(strings.Join(words[:], "_"), func(t *testing.T) {
			for _, form := range []string{"ordinary alias control", "whole file", "member", "member alias", "qualified alias"} {
				t.Run(form, func(t *testing.T) {
					root := t.TempDir()
					config := configuredImportAssets(t, root, words)

					library := "\"value\" is `leaf 雪`.\nCreate \"Tools\" " + words[2] + " a Dictionary.\n" +
						"Define Tools's \"Value\" " + words[2] + " `member 雪`.\n"
					libraryPath := writeSourceFile(t, root, "lib folder/values.l", library)
					var main, target, alias string
					switch form {
					case "ordinary alias control":
						main = "\"value\" is `leaf 雪`.\n" + words[0] + " value " + words[2] + " \"Copy\".\nShow Copy.\n"
						target, alias = "value", "Copy"
					case "whole file":
						main = words[0] + " `lib folder/values.l`.\nShow value.\n"
						target = "value"
					case "member":
						main = words[0] + " value " + words[1] + " `lib folder/values.l`.\nShow value.\n"
						target = "value"
					case "member alias":
						main = words[0] + " value " + words[1] + " `lib folder/values.l` " + words[2] + " \"Copy\".\nShow Copy.\n"
						target, alias = "value", "Copy"
					case "qualified alias":
						main = "  " + words[0] + " Tools’s Value " + words[1] + " `lib folder/values.l` " + words[2] + " \"Copy\". *** note ***\r\nShow Copy.\r\n"
						target, alias = "Tools.Value", "Copy"
					}
					mainPath := writeSourceFile(t, root, "main.l", main)
					c := NewLygerService(config).NewCompilation([]byte(main), mainPath)
					c.Make()
					if c.Error != nil {
						t.Fatal(c.Error)
					}
					if !bytes.Equal(c.MainBytes, []byte(main)) || !bytes.Equal(c.Prepared.Original, []byte(main)) || c.Memory[target] == nil {
						t.Fatal("configured import lost original source or its target")
					}
					if alias != "" {
						canonical, err := canonicalRecord(c.Memory[alias], c.Semantics.Aliases)
						if err != nil || canonical != c.Memory[target] {
							t.Fatalf("configured alias lost canonical target: %v", err)
						}
					}
					if form == "ordinary alias control" {
						return
					}
					if got := c.Semantics.Declarations[c.Memory[target]].source.Span.Path; got != libraryPath {
						t.Fatalf("imported target provenance = %s; want %s", got, libraryPath)
					}
					literal := c.Prepared.Literals[0]
					if string(literal.Raw) != "`lib folder/values.l`" || !bytes.Equal(c.TextCache[literal.ID], literal.Raw) {
						t.Fatal("import recognition changed cached path identity or bytes")
					}
					if form != "qualified alias" {
						return
					}
					original := c.Prepared.Statements[0]
					var synthesized *SourceStatement
					for index := range c.SourceStatements {
						statement := &c.SourceStatements[index]
						if statement.Span.Path == mainPath && statement.Span.Order == 0 {
							synthesized = statement
						}
						if statement.Span.Path == libraryPath && synthesized != nil {
							t.Fatal("alias preceded its imported declarations")
						}
					}
					if synthesized == nil || synthesized.Span != original.Span || !bytes.Equal(synthesized.Raw, original.Raw) || synthesized.Span.Start.Column != 3 {
						t.Fatal("generated alias lost raw source or original byte span")
					}
					wantBody := words[0] + " Tools’s Value " + words[2] + " \"Copy\""
					if got := string(statementBody(synthesized.Prepared)); got != wantBody {
						t.Fatalf("generated alias = %q; want source operator spellings %q", got, wantBody)
					}
					var mainOccurrence sourceOccurrenceID
					for _, occurrence := range c.SourceOccurrences {
						if occurrence.Path == mainPath {
							mainOccurrence = occurrence.ID
						}
					}
					if mainOccurrence == 0 || synthesized.occurrence != mainOccurrence || c.Semantics.Declarations[c.Memory[alias]].source != synthesized {
						t.Fatal("generated alias changed its expansion occurrence or declaration provenance")
					}
				})
			}
		})
	}
}

func TestConfiguredImportNestedRepeatedOccurrences(t *testing.T) {
	root := t.TempDir()
	words := [3]string{"Adopt", "via", "named"}
	config := configuredImportAssets(t, root, words)
	middle := writeSourceFile(t, root, "lib/middle.l", "Adopt `../leaf.l`.\n")
	leaf := writeSourceFile(t, root, "leaf.l", "\"value\" is `imported 雪`.\n")
	paragraph := "    Adopt value via `lib/middle.l` named \"Copy\".\nShow Copy.\n"
	main := paragraph + "\n" + paragraph
	mainPath := writeSourceFile(t, root, "main.l", main)
	c := NewLygerService(config).NewCompilation([]byte(main), mainPath)
	c.Make()
	if c.Error != nil {
		t.Fatal(c.Error)
	}
	if !reflect.DeepEqual(c.SourcePaths, []string{mainPath, middle, leaf, middle, leaf}) {
		t.Fatalf("nested import order = %v", c.SourcePaths)
	}
	aliases := paragraphBindingDeclarations(t, c, "Copy", 2)
	var targets []*record
	for _, alias := range aliases {
		target, err := canonicalRecord(alias.record, c.Semantics.Aliases)
		if err != nil || target == nil || c.Semantics.Declarations[target].source.Span.Path != leaf || target.storageScope != alias.record.storageScope {
			t.Fatalf("repeated import alias lost its local originating target: %v", err)
		}
		targets = append(targets, target)
	}
	if targets[0] == targets[1] || targets[0].storageScope == targets[1].storageScope || c.Memory["value"] != nil || c.Memory["Copy"] != nil {
		t.Fatal("repeated import merged private declarations or published them globally")
	}
	first, second := c.Semantics.Declarations[targets[0]].source, c.Semantics.Declarations[targets[1]].source
	if first.occurrence == second.occurrence || first.Span != second.Span || !bytes.Equal(first.Raw, second.Raw) || !bytes.Equal(c.MainBytes, []byte(main)) {
		t.Fatal("repeated imports lost independent occurrence identity or original source")
	}
	for _, occurrence := range c.SourceOccurrences {
		if occurrence.Path != leaf {
			continue
		}
		var parent *sourceOccurrence
		for index := range c.SourceOccurrences {
			if c.SourceOccurrences[index].ID == occurrence.Parent {
				parent = &c.SourceOccurrences[index]
			}
		}
		if parent == nil || parent.Path != middle || occurrence.ImportSite.Path != middle || occurrence.ImportSite.Start.Line != 1 {
			t.Fatal("nested import occurrence lost its real importing parent")
		}
	}
}

func TestConfiguredImportLexicalBoundaries(t *testing.T) {
	config := configuredImportAssets(t, t.TempDir(), [3]string{"Adopt", "via", "named"})
	gram, err := loadGrammar(config)
	if err != nil {
		t.Fatal(err)
	}
	for _, source := range []string{
		"Use `old spelling.l`.\n", "Adopts `active form.l`.\n", "Adopt value named \"Alias\".\n",
		"Show `Adopt via named`.\n", "\"Adopt\" is `ordinary Text`.\n",
	} {
		t.Run(source, func(t *testing.T) {
			isImport, imported, err := gram.isImportStatement([]byte(source))
			if err != nil || isImport {
				t.Fatalf("non-import = %v, %v", imported, err)
			}
		})
	}

	gram.lex.Nouns["import_shadow"] = lexeme{Root: "adopt", POS: "noun", Type: "generic"}
	isImport, imported, err := gram.isImportStatement([]byte("Adopt `noun.l`.\n"))
	if err != nil || isImport {
		t.Fatalf("import classifier overrode existing lexical precedence: %v, %v", imported, err)
	}
}

func TestConfiguredImportRequiresLexicon(t *testing.T) {
	for _, tc := range []struct {
		name string
		gram *grammar
	}{{"nil grammar", nil}, {"unloaded grammar", &grammar{}}} {
		t.Run(tc.name, func(t *testing.T) {
			isImport, imported, err := tc.gram.isImportStatement([]byte("Use `local-words.l`.\n"))
			if err == nil || isImport {
				t.Fatalf("missing configured lexicon used an implicit vocabulary: %v, %v", imported, err)
			}
		})
	}
}

func TestConfiguredDictionaryImportBindings(t *testing.T) {
	for _, namespace := range []bool{false, true} {
		t.Run(fmt.Sprint(namespace), func(t *testing.T) {
			root := t.TempDir()
			cfg := configuredImportAssets(t, root, [3]string{"Adopt", "via", "named"})
			source := "Adopt IO's Print via `io`.\n"
			if namespace {
				source = "Adopt IO via `io`.\nAdopt IO's Print named \"Print\".\n"
			}
			source += "Create \"Hello\" named a New Basic App.\nThe App Prints `dialect`.\nExecute Hello.\n"
			main := writeSourceFile(t, root, "main.l", source)
			result, err := NewLygerService(cfg).Compile([]byte(source), main)
			if err != nil {
				t.Fatal(err)
			}
			if !bytes.Contains(result.BackendAML, []byte(`Inputs: ~&["dialect"]&~`)) {
				t.Fatalf("selected dictionary failed under configured import words: %s", result.BackendAML)
			}
		})
	}
}
