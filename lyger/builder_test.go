package Lyger

import (
	"bytes"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func writeSourceFile(t *testing.T, root, name, source string) string {
	t.Helper()
	path := filepath.Join(root, name)
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(source), 0600); err != nil {
		t.Fatal(err)
	}
	return path
}

func buildTemporarySource(t *testing.T, path string) (*builder, error) {
	t.Helper()
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	prepared, err := PrepareSource(path, raw)
	if err != nil {
		t.Fatal(err)
	}
	gram := scopeTestGrammar(t)
	if gram == nil {
		t.Fatal("fixture grammar failed to load")
	}
	built := newPreparedBuilder(prepared, gram)
	return built, built.process()
}

func TestBuilderExpandsNestedRelativeImportsAndDefinitions(t *testing.T) {
	root := t.TempDir()
	bootstrap := writeSourceFile(t, root, "boot/definitions.l", "\"Bootstrap\" is `bootstrap`.\n")
	leaf := writeSourceFile(t, root, "shared/leaf.l", "\"Leaf\" is `leaf`.\n")
	library := writeSourceFile(t, root, "lib folder/local-words.l", "Use `../shared/leaf.l`.\nCreate \"Tools\" as a Dictionary.\n")
	mainText := "Use `boot/definitions.l`.\nUse Tools's Render from `lib folder/local-words.l` as \"Pretty Print\".\n\n\"x\" is `main`.\n"
	main := writeSourceFile(t, root, "main.l", mainText)
	built, err := buildTemporarySource(t, main)
	if err != nil {
		t.Fatal(err)
	}
	var paths, bodies []string
	statements := testPhysicalStatements(root, built.SourceStatements)
	for _, statement := range statements {
		paths = append(paths, statement.Span.Path)
		bodies = append(bodies, string(statementBody(statement.Prepared)))
	}
	if !reflect.DeepEqual(paths, []string{bootstrap, leaf, library, main, main}) {
		t.Fatalf("expanded source order = %v", paths)
	}
	var joined []byte
	for _, statement := range built.SourceStatements {
		joined = append(joined, statement.Prepared...)
	}
	if !bytes.Equal(joined, built.BuiltBytes) || len(statements) != 5 {
		t.Fatal("built bytes and source metadata disagree")
	}
	if bodies[3] != "Use Tools's Render as \"Pretty Print\"" {
		t.Fatalf("residual alias = %q", bodies[3])
	}
	alias := statements[3]
	if string(alias.Raw) != strings.SplitAfter(mainText, "\n")[1] || alias.Span.Start.Line != 2 || alias.Span.Paragraph != 1 || !bytes.HasSuffix(alias.Prepared, []byte(".\n")) {
		t.Fatalf("alias lost import origin or terminator: %+v", alias)
	}
	if statements[4].Span.Start.Line != 4 || statements[4].Span.Paragraph != 2 {
		t.Fatalf("main provenance after expansion = %+v", statements[4].Span)
	}
	for _, literal := range []string{"`bootstrap`", "`leaf`", "`main`"} {
		found := false
		for _, cached := range built.TextCache {
			found = found || string(cached) == literal
		}
		if !found {
			t.Errorf("expanded literal pool lost %s", literal)
		}
	}
}

func TestBuilderAllowsRepeatedImportsWithDistinctAliases(t *testing.T) {
	root := t.TempDir()

	writeSourceFile(t, root, "local-words.l", "Create \"Tools\" as a Dictionary.\n")
	main := writeSourceFile(t, root, "main.l", "Use Tools's Render from `local-words.l` as \"First\".\nUse Tools's Render from `local-words.l` as \"Second\".\n")
	built, err := buildTemporarySource(t, main)
	if err != nil {
		t.Fatal(err)
	}
	statements := testPhysicalStatements(root, built.SourceStatements)
	if len(statements) != 4 {
		t.Fatalf("repeated imports produced %d statements", len(statements))
	}
	for index, name := range []string{"First", "Second"} {
		statement := statements[index*2+1]
		if string(statementBody(statement.Prepared)) != "Use Tools's Render as \""+name+"\"" || statement.Span.Start.Line != index+1 {
			t.Fatalf("repeated alias = %s, %+v", statement.Prepared, statement.Span)
		}
	}
}

func TestBuilderReportsNestedImportFailures(t *testing.T) {
	for _, scenario := range []string{"cycle", "missing", "definitions-cycle"} {
		t.Run(scenario, func(t *testing.T) {
			root := t.TempDir()

			main := writeSourceFile(t, root, "main.l", "Use `nested/child.l`.\n")
			childText := "Use `../main.l`.\n"
			if scenario == "missing" {
				childText = "Use `missing.l`.\n"
			}
			child := writeSourceFile(t, root, "nested/child.l", childText)
			if scenario == "definitions-cycle" {
				writeSourceFile(t, root, "nested/child.l", "Use `../definitions.l`.\n")
				writeSourceFile(t, root, "definitions.l", "Use `main.l`.\n")
			}
			_, err := buildTemporarySource(t, main)
			if err == nil {
				t.Fatal("invalid import graph succeeded")
			}
			if scenario == "missing" {
				if !strings.Contains(err.Error(), child+":1:1:") || !strings.Contains(err.Error(), filepath.Join(root, "nested/missing.l")) {
					t.Fatalf("missing import lacks origin: %v", err)
				}
			} else if !strings.Contains(err.Error(), "import cycle:") || !strings.Contains(err.Error(), main) {
				t.Fatalf("cycle lacks source chain: %v", err)
			}
		})
	}
}

func TestBuilderRetainsNonEmittingSourceOccurrences(t *testing.T) {
	for _, definitionsSource := range []string{"", "Use `boot.l`.\n"} {
		name := "empty definitions"
		if definitionsSource != "" {
			name = "import-only definitions"
		}
		t.Run(name, func(t *testing.T) {
			root := t.TempDir()
			definitions := writeSourceFile(t, root, "definitions.l", definitionsSource)
			boot := writeSourceFile(t, root, "boot.l", "*** No declarations yet. ***\n")
			leaf := writeSourceFile(t, root, "leaf.l", "\"Leaf\" is `value`.\n")
			middle := writeSourceFile(t, root, "nested/middle.l", "Use `../leaf.l`.\n")
			alias := filepath.Join(root, "alias.l")
			if err := os.Symlink(leaf, alias); err != nil {
				t.Fatal(err)
			}
			main := writeSourceFile(t, root, "main.l", "Use `definitions.l`.\nUse `nested/middle.l`.\nUse `nested/middle.l`.\nUse `alias.l`.\n")
			built, err := buildTemporarySource(t, main)
			if err != nil {
				t.Fatal(err)
			}
			want := []string{main, definitions}
			if definitionsSource != "" {
				want = append(want, boot)
			}
			want = append(want, middle, leaf, middle, leaf, alias)
			if !reflect.DeepEqual(testProgramPaths(built.SourcePaths, root), want) {
				t.Fatalf("source occurrences = %v; want %v", built.SourcePaths, want)
			}
			statements := testPhysicalStatements(root, built.SourceStatements)
			if len(statements) != 3 {
				t.Fatalf("repeated source expansion changed: %d statements", len(statements))
			}
			for i, path := range []string{leaf, leaf, alias} {
				if statements[i].Span.Path != path || string(statements[i].Raw) != "\"Leaf\" is `value`.\n" {
					t.Fatal("input inventory changed emitted source provenance or bytes")
				}
			}
			otherRoot := t.TempDir()
			empty := writeSourceFile(t, otherRoot, "empty.l", "")
			otherMain := writeSourceFile(t, otherRoot, "main.l", "Use `empty.l`.\n")
			other, err := buildTemporarySource(t, otherMain)
			if err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(testProgramPaths(other.SourcePaths, otherRoot), []string{otherMain, empty}) || len(testPhysicalStatements(otherRoot, other.SourceStatements)) != 0 || !reflect.DeepEqual(testProgramPaths(built.SourcePaths, root), want) {
				t.Fatal("source inventory leaked across independent expansions")
			}
		})
	}
}
