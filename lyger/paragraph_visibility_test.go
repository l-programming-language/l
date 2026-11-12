package Lyger

import (
	"bytes"
	"fmt"
	"path/filepath"
	"strings"
	"testing"
)

func TestParagraphVisibilityVocabulary(t *testing.T) {
	lex, err := loadLexicon(canonicalTestOptions().Resources, "eng")
	if err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct{ spelling, category string }{
		{"Public", "adjectives."}, {"Publicly", "adverbs."},
	} {
		t.Run(tc.spelling, func(t *testing.T) {
			word, category, err := lex.lookup([]byte(tc.spelling))
			if err != nil || word == nil || !strings.HasPrefix(category, tc.category) {
				t.Fatalf("configured %s lookup = (%+v, %q, %v); want %s", tc.spelling, word, category, err, tc.category)
			}
		})
	}
}

const paragraphWitness = paragraphGlobals + "Create \"Witness\" as a New Basic App.\nThe App Prints x1.\n"

func paragraphDeclaration(t *testing.T, compilation *Compilation, name string) (*record, *lStatement) {
	t.Helper()
	var found *record
	var source *lStatement
	for _, statement := range compilation.Statements {
		if filepath.Base(statement.source.Span.Path) != "main.l" {
			continue
		}
		for _, tok := range statement.tokens {
			if !strings.EqualFold(tok.declarationRoot(), name) {
				continue
			}
			if found != nil {
				t.Fatalf("test expected one declaration of %s", name)
			}
			found, source = compilation.Semantics.Subjects[tok], statement
		}
	}
	if found == nil || source == nil {
		t.Fatalf("declaration %s lost its source-to-record identity", name)
	}
	return found, source
}

func TestParagraphPublicDeclarationsRetainCanonicalIdentity(t *testing.T) {
	for _, tc := range []struct {
		name, declaration, body string
		kind                    valueType
	}{
		{"Greeting", "Publicly define \"Greeting\" as `hello`.", "", valueText},
		{"Count", "Publicly define \"Count\" as 42.", "", valueNumber},
		{"Enabled", "Publicly define \"Enabled\" as true.", "", valueTruth},
		{"Shared", "Create \"Shared\" as a Public Dictionary.", "", valueDictionary},
		{"Worker", "Create \"Worker\" as a New Public Basic App.", "The App Prints x2.\n", valueApp},
		{"Chirp", "Publicly define \"Chirp|-s|-ed|-ing|-er\" to `<DEF>BOD^E{|Statement: ~::std::println!(\"{}\", <@Output|Thing@>);~|}<DEF>`.", "", valueLexeme},
	} {
		t.Run(tc.name, func(t *testing.T) {
			source := memberResolutionSource{}
			first, second := "PublishedAlias", "PublishedAgain"
			firstLabel, secondLabel := first, second
			if tc.kind == valueLexeme {
				first, second = "Relay", "Forward"
				firstLabel, secondLabel = "Relay|-s|-ed|-ing|-er", "Forward|-s|-ed|-ing|-er"
			}
			source.main = paragraphWitness + "\n    " + tc.declaration + "\n" + tc.body + "\"UnpublishedSibling\" is `private`.\n" +
				"\nUse " + tc.name + " as \"" + firstLabel + "\".\nUse " + first + " as \"" + secondLabel + "\".\n"
			compilation := executableTestResult(t, source).Compilation
			if compilation.Memory["UnpublishedSibling"] != nil {
				t.Fatal("Public applied to the whole paragraph instead of its declaration")
			}
			original, statement := paragraphDeclaration(t, compilation, tc.name)
			if compilation.Memory[tc.name] != original {
				t.Fatal("Public copied the declaration or did not publish its canonical record globally")
			}
			value, err := original.Value.value()
			if err != nil || value.Type != tc.kind {
				t.Fatalf("published %s lost its value type: %+v, %v", tc.name, value, err)
			}
			if compilation.Semantics.Declarations[original] != statement || statement.source.Span.Start.Line != strings.Count(paragraphWitness, "\n")+2 {
				t.Fatal("publication replaced defining-paragraph provenance")
			}
			for _, name := range []string{first, second} {
				alias := compilation.Memory[name]
				canonical, err := canonicalRecord(alias, compilation.Semantics.Aliases)
				if err != nil || alias == nil || canonical != original {
					t.Fatalf("public alias %s lost canonical identity: %v", name, err)
				}
			}
			if tc.kind == valueApp {
				app := original.Value.Applicative.App
				count := 0
				for _, candidate := range compilation.Semantics.Apps {
					if candidate == app {
						count++
					}
				}
				if count != 1 || len(compilation.Program.Books) != 2 {
					t.Fatal("public App aliases created duplicate applications")
				}
			}
		})
	}
}

func TestParagraphPublicDictionaryKeepsMemberNamespace(t *testing.T) {
	source := memberResolutionSource{}
	source.main = paragraphWitness + paragraphAnnounce + "\n    Create \"Shared\" as a Public Dictionary.\n" +
		"Create Shared's \"Speak|-s|----oke|-ing|-er\" for Processing a New Announcer.\n" +
		"\nUse Shared as \"SharedAlias\".\nUse SharedAlias's Speak as \"Speak\".\n" +
		"Create \"Reader\" as a New Basic App.\nThe App Speaks x2.\n"
	compilation := executableTestResult(t, source).Compilation
	owner, _ := paragraphDeclaration(t, compilation, "Shared")
	member := compilation.Memory["Shared.Speak"]
	alias := compilation.Memory["Speak"]
	canonical, err := canonicalRecord(alias, compilation.Semantics.Aliases)
	if err != nil || member == nil || canonical != member || owner.Value.Iterative.Dictionary.Values["Speak"] != member.Value.Lexeme {
		t.Fatalf("Public dictionary member alias lost canonical namespace: %v", err)
	}
	if compilation.Memory["SharedAlias.Speak"] != nil || !bytes.Contains(compilation.ASMBytes, []byte(`::std::println!("{}", "Hello Again World!");`)) {
		t.Fatal("public dictionary alias cloned its namespace or lost local behavior")
	}
}

func TestParagraphPrivateValuesCannotEscape(t *testing.T) {
	for _, tc := range []struct{ name, declaration, body, localAlias, escapedAlias string }{
		{"SecretText", "\"SecretText\" is `private`.", "", "LocalAlias", "Escaped"},
		{"SecretDictionary", "Create \"SecretDictionary\" as a Dictionary.", "", "LocalAlias", "Escaped"},
		{"SecretApp", "Create \"SecretApp\" as a New Basic App.", "The App Prints x2.\n", "LocalAlias", "Escaped"},
		{"SecretAction", "Create \"SecretAction|-s|-ed|-ing|-er\" for Showing any Text.", "", "Relay|-s|-ed|-ing|-er", "Forward|-s|-ed|-ing|-er"},
	} {
		for _, sibling := range []bool{false, true} {
			t.Run(fmt.Sprintf("%s/sibling=%t", tc.name, sibling), func(t *testing.T) {
				source := memberResolutionSource{}
				source.main = paragraphWitness + "\n    " + tc.declaration + "\n" + tc.body + "Use " + tc.name + " as \"" + tc.localAlias + "\".\n"
				control := compileMemberResolutionSource(t, source)
				if control.Error != nil {
					t.Fatalf("private declaration and in-scope alias control: %v", control.Error)
				}
				tail := "\n"
				if sibling {
					tail += "    \"Neighbour\" is `next`.\n"
				}
				source.main += tail + "Use " + tc.name + " as \"" + tc.escapedAlias + "\".\n"
				compilation := compileMemberResolutionSource(t, source)
				line := strings.Count(strings.TrimSuffix(source.main, "\n"), "\n") + 1
				if compilation.Error == nil || !strings.Contains(compilation.Error.Error(), fmt.Sprintf("main.l:%d:", line)) || !strings.Contains(compilation.Error.Error(), tc.name) {
					t.Fatalf("private %s escaped or visibility error lost its source: %v", tc.name, compilation.Error)
				}
			})
		}
	}
}
