package Lyger

import (
	"bytes"
	"fmt"
	"os"
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
			source := paragraphSource(t)
			source.main = paragraphWitness + "\n    " + tc.declaration + "\n" + tc.body + "\"UnpublishedSibling\" is `private`.\n" +
				"\nUse " + tc.name + " as \"PublishedAlias\".\nUse PublishedAlias as \"PublishedAgain\".\n"
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
			if compilation.Semantics.Declarations[original] != statement || statement.source.Span.Paragraph != 2 {
				t.Fatal("publication replaced defining-paragraph provenance")
			}
			for _, name := range []string{"PublishedAlias", "PublishedAgain"} {
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
	source := paragraphSource(t)
	source.main = paragraphWitness + paragraphAnnounce + "\n    Create \"Shared\" as a Public Dictionary.\n" +
		"Create Shared's \"Speak|-s|-ed|-ing|-er\" for Processing a New Announcer.\n" +
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
	for _, tc := range []struct{ name, declaration, body string }{
		{"SecretText", "\"SecretText\" is `private`.", ""},
		{"SecretDictionary", "Create \"SecretDictionary\" as a Dictionary.", ""},
		{"SecretApp", "Create \"SecretApp\" as a New Basic App.", "The App Prints x2.\n"},
		{"SecretAction", "Create \"SecretAction\" for Showing any Text.", ""},
	} {
		for _, sibling := range []bool{false, true} {
			t.Run(fmt.Sprintf("%s/sibling=%t", tc.name, sibling), func(t *testing.T) {
				source := paragraphSource(t)
				source.main = paragraphWitness + "\n    " + tc.declaration + "\n" + tc.body + "Use " + tc.name + " as \"LocalAlias\".\n"
				control := compileMemberResolutionSource(t, source)
				if control.Error != nil {
					t.Fatalf("private declaration and in-scope alias control: %v", control.Error)
				}
				tail := "\n"
				if sibling {
					tail += "    \"Neighbour\" is `next`.\n"
				}
				source.main += tail + "Use " + tc.name + " as \"Escaped\".\n"
				compilation := compileMemberResolutionSource(t, source)
				line := strings.Count(strings.TrimSuffix(source.main, "\n"), "\n") + 1
				if compilation.Error == nil || !strings.Contains(compilation.Error.Error(), fmt.Sprintf("main.l:%d:", line)) || !strings.Contains(compilation.Error.Error(), tc.name) {
					t.Fatalf("private %s escaped or visibility error lost its source: %v", tc.name, compilation.Error)
				}
			})
		}
	}
}

// Read the maintained examples themselves, preserving their source-relative
// bootstrap/imports. The CLI suite owns execution expectations for these files.
func TestParagraphDemosCompleteFrontend(t *testing.T) {

	for _, name := range []string{
		"main.l", "fallthrough.l", "exit-program.l", "exit-error.l",
		"multiple-subjects.l", "repeated-activation.l", "run-within-paragraph.l",
		"coordinate-existing.l", "public-values.l", "local-shadowing.l",
		"local-definitions.l", "local-imports.l",
	} {
		path := filepath.Join("..", "demo", "paragraphs", name)
		t.Run(name, func(t *testing.T) {
			raw, err := os.ReadFile(path)
			if err != nil {
				t.Fatal(err)
			}
			config := canonicalTestOptions()
			result, err := NewLygerService(config).Compile(raw, path)
			if err != nil {
				t.Fatal(err)
			}
			if result == nil || result.Compilation == nil || result.Compilation.Program == nil || len(result.Compilation.ProcStatements) != 0 || !bytes.Equal(result.Compilation.MainBytes, raw) {
				t.Fatal("demo lost its complete graph, original source, or unresolved work")
			}
			assertParagraphDemoVisibility(t, name, path, result.Compilation)
		})
	}
}

func assertParagraphDemoVisibility(t *testing.T, name, path string, compilation *Compilation) {
	t.Helper()
	private := map[string][]string{
		"main.l": {"y"}, "fallthrough.l": {"result"},
		"multiple-subjects.l":   {"firstResult", "secondResult"},
		"repeated-activation.l": {"result"}, "run-within-paragraph.l": {"result"},
		"coordinate-existing.l": {"Selected"},
		"local-definitions.l":   {"Launch", "Compact", "First", "Second", "result"},
		"local-imports.l":       {"Display", "LocalLibrary", "Paint", "Worker"},
	}[name]
	declared := make(map[string][]*record)
	statements := make(map[*record]*lStatement)
	for _, statement := range compilation.Statements {
		if filepath.Clean(statement.source.Span.Path) != filepath.Clean(path) {
			continue
		}
		for _, tok := range statement.tokens {
			if root := tok.declarationRoot(); root != "" {
				rec := compilation.Semantics.Subjects[tok]
				if rec == nil {
					t.Fatalf("%s:%d declaration %s lost its canonical record", name, statement.source.Span.Start.Line, root)
				}
				declared[root] = append(declared[root], rec)
				statements[rec] = statement
			}
		}
	}
	for _, name := range private {
		if len(declared[strings.ToLower(name)]) == 0 {
			t.Errorf("demo no longer declares expected private %s", name)
		}
		if compilation.Memory[name] != nil {
			t.Errorf("demo's private %s leaked into root memory", name)
		}
	}
	if name == "coordinate-existing.l" || name == "local-imports.l" {
		global := compilation.Memory["result"]
		if global == nil || global.Value.Primitive == nil || global.Value.Primitive.Text != "`done`" || len(declared["result"]) != 2 {
			t.Error("scoped Output result replaced the outer result or lost a declaration")
		} else {
			for _, statement := range compilation.Statements {
				if filepath.Clean(statement.source.Span.Path) != filepath.Clean(path) || strings.TrimSpace(string(statement.source.Raw)) != "Show result." {
					continue
				}
				want := global
				for _, rec := range declared["result"] {
					if statements[rec].source.Span.Paragraph == statement.source.Span.Paragraph {
						want = rec
					}
				}
				for _, tok := range statement.tokens {
					if string(tok.rawBytes) == "result" && (tok.binding == nil || tok.binding.Record != want) {
						t.Errorf("%s:%d Show result did not retain its lexical declaration", name, statement.source.Span.Start.Line)
					}
				}
			}
		}
	}
	if name == "coordinate-existing.l" && len(declared["selected"]) == 1 {
		owner, err := canonicalRecord(declared["selected"][0], compilation.Semantics.Aliases)
		if err != nil || owner != compilation.Memory["Second"] {
			t.Fatalf("local App alias changed the canonical outer App: %v", err)
		}
	}
	if name == "local-imports.l" {
		for _, imported := range []string{"LocalWords", "LocalWords.Paint"} {
			if compilation.Memory[imported] != nil {
				t.Errorf("local import published its underlying namespace %s globally", imported)
			}
		}
		foundImport := false
		for _, imported := range compilation.SourcePaths {
			if filepath.Clean(imported) == filepath.Clean(filepath.Join(filepath.Dir(path), "local-words.l")) {
				foundImport = true
			}
		}
		if !foundImport {
			t.Error("local import disappeared from the source inventory")
		}
		var definition *record
		for _, alias := range []string{"Display", "Paint"} {
			if len(declared[strings.ToLower(alias)]) != 1 {
				t.Fatalf("local import lost alias %s", alias)
			}
			target, err := canonicalRecord(declared[strings.ToLower(alias)][0], compilation.Semantics.Aliases)
			if err != nil || target == nil {
				t.Fatalf("local import alias %s lost its target: %v", alias, err)
			}
			if definition != nil && target != definition {
				t.Error("aliases of the same imported member acquired separate definitions")
			}
			definition = target
			origin := compilation.Semantics.Declarations[target]
			if origin == nil || target.ID != "IO.Print" || !strings.Contains(string(origin.source.Raw), "Create IO's") {
				t.Errorf("local import alias %s lost canonical IO.Print definition provenance", alias)
			}
		}
		uses := 0
		for _, statement := range compilation.Statements {
			if filepath.Clean(statement.source.Span.Path) != filepath.Clean(path) {
				continue
			}
			for _, tok := range statement.tokens {
				if string(tok.rawBytes) != "Paints" {
					continue
				}
				uses++
				if tok.binding == nil {
					t.Fatal("renamed imported verb lost its grammatical binding")
				}
				target, err := canonicalRecord(tok.binding.Record, compilation.Semantics.Aliases)
				if err != nil || target != definition {
					t.Errorf("renamed imported verb did not bind its canonical imported definition: %v", err)
				}
			}
		}
		if uses != 1 {
			t.Fatalf("expected one inflected Paints use, got %d", uses)
		}
	}
	if name == "local-definitions.l" {
		global := compilation.Memory["message"]
		if global == nil || global.Value.Primitive == nil || global.Value.Primitive.Text != "`done`" || len(declared["message"]) != 3 {
			t.Error("sibling local messages replaced the outer message or lost a declaration")
		} else {
			seen := make(map[*record]bool)
			for _, rec := range declared["message"] {
				if seen[rec] {
					t.Error("outer and sibling message declarations share a record")
				}
				seen[rec] = true
			}
			for _, statement := range compilation.Statements {
				if filepath.Clean(statement.source.Span.Path) == filepath.Clean(path) && strings.TrimSpace(string(statement.source.Raw)) == "Show message." {
					for _, tok := range statement.tokens {
						if string(tok.rawBytes) == "message" && (tok.binding == nil || tok.binding.Record != global) {
							t.Error("final Show message did not resolve the enclosing global declaration")
						}
					}
				}
			}
		}
		for _, name := range []string{"Launch", "Compact", "result"} {
			if records := declared[strings.ToLower(name)]; len(records) != 2 || records[0] == records[1] {
				t.Errorf("same-spelled local %s declarations did not retain two identities", name)
			}
		}
		for _, word := range []string{"Launch", "Compact"} {
			uses := 0
			for _, statement := range compilation.Statements {
				if filepath.Clean(statement.source.Span.Path) != filepath.Clean(path) {
					continue
				}
				for _, tok := range statement.tokens {
					if string(tok.rawBytes) != word {
						continue
					}
					uses++
					var want *record
					for _, rec := range declared[strings.ToLower(word)] {
						if statements[rec].source.Span.Paragraph == statement.source.Span.Paragraph {
							want = rec
						}
					}
					if want == nil || tok.binding == nil || tok.binding.Record != want {
						t.Errorf("%s:%d %s did not retain its local source definition", name, statement.source.Span.Start.Line, word)
					}
				}
			}
			if uses != 2 {
				t.Errorf("expected two scoped uses of %s, got %d", word, uses)
			}
		}
		for _, value := range []string{"First local", "Second local"} {
			if !bytes.Contains(compilation.ASMBytes, []byte(fmt.Sprintf("(&&LPrintValue(&(%q))).l_print();", value))) {
				t.Errorf("local definitions lost their own captured Text %q", value)
			}
		}
	}
}
