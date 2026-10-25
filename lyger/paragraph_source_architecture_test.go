package Lyger

import (
	"bytes"
	"reflect"
	"strings"
	"testing"
)

func TestParagraphSourceInlineCommentSeparators(t *testing.T) {
	for _, separator := range []string{". *** trailing. comment *** ", ".*** trailing. comment ***"} {
		t.Run(separator, func(t *testing.T) {
			first := "Show `first`" + separator
			source := first + "Show `second`.\r\n"
			prepared, err := PrepareSource("comments.l", []byte(source))
			if err != nil {
				t.Fatal(err)
			}
			if len(prepared.Statements) != 2 {
				t.Fatalf("inline comment-separated statements = %d; want 2", len(prepared.Statements))
			}
			if string(prepared.Statements[0].Raw) != first || string(prepared.Statements[0].Terminator) != separator || string(prepared.Statements[1].Terminator) != ".\r\n" {
				t.Fatalf("source separators changed: %+v", prepared.Statements)
			}
			if got := prepared.Statements[1].Span; got.Start != (SourcePosition{Offset: len(first), Line: 1, Column: len(first) + 1}) || got.Paragraph != 1 || got.Order != 1 {
				t.Fatalf("inline source position = %+v", got)
			}
			var reconstructed []byte
			for i, statement := range prepared.Statements {
				reconstructed = append(reconstructed, statement.Raw...)
				want := []string{"Show `first`.", "Show `second`."}[i]
				if got := string(renderExpandedSource([]SourceStatement{statement}, prepared.TextCache)); got != want {
					t.Errorf("individual rendered sentence = %q; want %q", got, want)
				}
			}
			if !bytes.Equal(reconstructed, []byte(source)) || !bytes.Equal(prepared.Original, []byte(source)) {
				t.Fatal("inline comment splitting changed original bytes")
			}
			parsed := newParser([]byte(source)).parseStatements()
			if !reflect.DeepEqual(parsed, [][]byte{[]byte("Show `first`"), []byte("Show `second`")}) {
				t.Fatalf("shared parser boundary = %q", parsed)
			}
		})
	}
}

func TestParagraphSourceNoImplicitLineTerminator(t *testing.T) {
	const source = "    Show `first`\n\nShow `second`.\n"
	prepared, err := PrepareSource("explicit-period.l", []byte(source))
	if err != nil {
		t.Fatal(err)
	}
	if len(prepared.Statements) != 1 || len(newParser([]byte(source)).parseStatements()) != 1 {
		t.Fatal("a blank line became an implicit sentence terminator")
	}
	if got := prepared.Statements[0]; string(got.Raw) != source || got.Span.Start.Column != 5 || got.Span.Paragraph != 1 {
		t.Fatalf("unterminated continuation changed provenance: %+v", got)
	}
	if got := string(renderExpandedSource(prepared.Statements, prepared.TextCache)); got != "Show `first`\n\nShow `second`." {
		t.Fatalf("explicit-period sentence = %q", got)
	}
}

func TestParagraphSourceInlineImportsPreserveCallsite(t *testing.T) {
	root := t.TempDir()
	empty := writeSourceFile(t, root, "empty.l", "*** intentionally empty ***\n")
	library := writeSourceFile(t, root, "lib/local-words.l", "Create \"Tools\" as a Dictionary.\n")
	const imported = "Use Tools's Render from `lib/local-words.l` as \"Alias\". "
	const source = "    Use `empty.l`. " + imported + "Show Alias.\n"
	main := writeSourceFile(t, root, "main.l", source)
	built, err := buildTemporarySource(t, main)
	if err != nil {
		t.Fatal(err)
	}
	statements := testPhysicalStatements(root, built.SourceStatements)
	if !reflect.DeepEqual(testProgramPaths(built.SourcePaths, root), []string{main, empty, library}) {
		t.Fatalf("inline import inventory = %v", built.SourcePaths)
	}
	if len(statements) != 3 {
		t.Fatalf("inline import expansion = %d statements; want library, alias, Show", len(statements))
	}
	if got := statements[0]; got.Span.Path != library || string(got.Raw) != "Create \"Tools\" as a Dictionary.\n" {
		t.Fatalf("library source changed: %+v", got)
	}
	alias := statements[1]
	start := strings.Index(source, imported)
	if alias.Span.Path != main || alias.Span.Start != (SourcePosition{Offset: start, Line: 1, Column: start + 1}) || alias.Span.Paragraph != 1 || string(alias.Raw) != imported || string(statementBody(alias.Prepared)) != "Use Tools's Render as \"Alias\"" {
		t.Fatalf("inline alias callsite changed: %+v", alias)
	}
	last := statements[2]
	if last.Span.Path != main || string(statementBody(last.Prepared)) != "Show Alias" || last.Span.Start.Offset != start+len(imported) {
		t.Fatalf("statement after inline imports changed: %+v", last)
	}
}

func TestParagraphSourceLayoutUsesFirstCodeLine(t *testing.T) {
	for _, tc := range []struct {
		name, source string
		first        []string
		lines        []int
		columns      []int
		paragraphs   []int
		encapsulated []bool
	}{
		{
			name: "inline and continuation statements inherit the opener",
			source: "    Show `local`. Show `inline`.\nShow `continuation`.\n\n" +
				"Show `root`.\n    Show `still root`.\n",
			first: []string{"Show `local`", "Show `root`"}, lines: []int{1, 4}, columns: []int{5, 1}, paragraphs: []int{1, 2}, encapsulated: []bool{true, false},
		},
		{
			name: "comment-only lines do not become the opener",
			source: "    *** heading\n\ncontinued ***\nShow `root`.\n\n" +
				"*** another heading ***\n    Show `local`.\n",
			first: []string{"Show `root`", "Show `local`"}, lines: []int{4, 7}, columns: []int{1, 5}, paragraphs: []int{1, 2}, encapsulated: []bool{false, true},
		},
		{
			name:   "comment before first code on an indented line",
			source: "    *** lead *** Show `local`.\n",
			first:  []string{"Show `local`"}, lines: []int{1}, columns: []int{18}, paragraphs: []int{1}, encapsulated: []bool{true},
		},
		{
			name:   "four ASCII spaces are distinct from tabs",
			source: "\tShow `tabbed`.\r\n\r\n   Show `three`.\r\n\r\n        Show `single level`.\r\n",
			first:  []string{"Show `tabbed`", "Show `three`", "Show `single level`"}, lines: []int{1, 3, 5}, columns: []int{2, 4, 9}, paragraphs: []int{1, 2, 3}, encapsulated: []bool{false, false, true},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var want []paragraphLayout
			for i, first := range tc.first {
				offset := strings.Index(tc.source, first)
				if offset < 0 {
					t.Fatalf("test opener %q is missing", first)
				}
				want = append(want, paragraphLayout{Number: tc.paragraphs[i], FirstCode: SourcePosition{Offset: offset, Line: tc.lines[i], Column: tc.columns[i]}, Encapsulated: tc.encapsulated[i]})
			}
			prepared, err := PrepareSource("layout.l", []byte(tc.source))
			if err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(prepared.paragraphs, want) {
				t.Fatalf("first-code layout = %+v; want %+v", prepared.paragraphs, want)
			}
			if !bytes.Equal(prepared.Original, []byte(tc.source)) {
				t.Fatal("layout discovery changed original source")
			}
		})
	}
	for _, source := range []string{"", " \t\r\n", "    *** comment only ***\n"} {
		prepared, err := PrepareSource("empty.l", []byte(source))
		if err != nil || len(prepared.paragraphs) != 0 {
			t.Fatalf("non-emitting source acquired code layout: %+v, %v", prepared, err)
		}
	}
}

func TestParagraphSourceImportOccurrenceAncestry(t *testing.T) {
	root := t.TempDir()
	empty := writeSourceFile(t, root, "empty.l", "    *** no code ***\n")
	leaf := writeSourceFile(t, root, "leaf.l", "Create \"Tools\" as a Dictionary.\n")
	middle := writeSourceFile(t, root, "nested/middle.l", "Use `../leaf.l`.\n")
	const source = "*** opener ***\n    Use `empty.l`. Use Tools's Render from `nested/middle.l` as \"First\".\n\n" +
		"    Use Tools's Render from `nested/middle.l` as \"Second\".\n"
	main := writeSourceFile(t, root, "main.l", source)
	built, err := buildTemporarySource(t, main)
	if err != nil {
		t.Fatal(err)
	}
	occurrences := testPhysicalOccurrences(root, built.occurrences)
	statements := testPhysicalStatements(root, built.SourceStatements)
	wantPaths := []string{main, empty, middle, leaf, middle, leaf}
	if len(occurrences) != len(wantPaths) || !reflect.DeepEqual(testProgramPaths(built.SourcePaths, root), wantPaths) {
		t.Fatalf("physical occurrence inventory = %+v / %v", occurrences, built.SourcePaths)
	}
	byID := make(map[sourceOccurrenceID]sourceOccurrence)
	for _, occurrence := range built.occurrences {
		byID[occurrence.ID] = occurrence
	}
	for i, occurrence := range occurrences {
		if occurrence.Path != wantPaths[i] || occurrence.ID == 0 {
			t.Errorf("occurrence %d = %+v", i, occurrence)
		}
		if i == 0 {
			if occurrence.Parent != 0 || occurrence.ImportSite != (SourceSpan{}) {
				t.Error("main acquired a physical import parent")
			}
		} else {
			wantParent := occurrences[0].ID
			if i == 3 {
				wantParent = occurrences[2].ID
			}
			if i == 5 {
				wantParent = occurrences[4].ID
			}
			if occurrence.Parent != wantParent {
				t.Errorf("occurrence %d parent = %d; want %d", i, occurrence.Parent, wantParent)
			}
		}
	}
	if len(built.main.Statements) != 3 || len(statements) != 4 {
		t.Fatalf("main/expanded statements = %d/%d", len(built.main.Statements), len(statements))
	}
	for i, index := range []int{1, 2, 4} {
		if occurrences[index].ImportSite != built.main.Statements[i].Span {
			t.Errorf("import %d lost callsite", i)
		}
	}
	for _, index := range []int{3, 5} {
		occurrence := occurrences[index]
		parent := byID[occurrence.Parent]
		site := occurrence.ImportSite
		if site.Path != parent.Path || site.Start != (SourcePosition{Offset: 0, Line: 1, Column: 1}) || site.Paragraph != 1 || site.Order != 0 {
			t.Errorf("nested import lost parent: %+v", occurrence)
		}
	}
	if len(occurrences[1].Paragraphs) != 0 {
		t.Fatal("comment-only import acquired code layout")
	}
	layout := occurrences[0].Paragraphs
	if len(layout) != 2 || !layout[0].Encapsulated || !layout[1].Encapsulated || layout[0].FirstCode != (SourcePosition{Offset: strings.Index(source, "Use `empty.l`"), Line: 2, Column: 5}) {
		t.Fatalf("main layout = %+v", layout)
	}
	for i, want := range []sourceOccurrenceID{occurrences[3].ID, occurrences[0].ID, occurrences[5].ID, occurrences[0].ID} {
		if statements[i].occurrence != want {
			t.Errorf("statement %d occurrence = %d; want %d", i, statements[i].occurrence, want)
		}
	}
	if statements[0].Span != statements[2].Span || !bytes.Equal(statements[0].Raw, statements[2].Raw) {
		t.Fatal("repeated source identity changed")
	}
	for _, statement := range built.main.Statements {
		if statement.occurrence != 0 {
			t.Fatal("prepared source was mutated")
		}
	}
	built.main.paragraphs[0].Encapsulated = false
	if !occurrences[0].Paragraphs[0].Encapsulated {
		t.Fatal("occurrence aliases prepared layout")
	}
	occurrences[2].Paragraphs[0].FirstCode.Line = 99
	if occurrences[4].Paragraphs[0].FirstCode.Line != 1 {
		t.Fatal("repeated occurrences share mutable layout")
	}
	otherRoot := t.TempDir()
	otherMain := writeSourceFile(t, otherRoot, "main.l", "")
	other, err := buildTemporarySource(t, otherMain)
	if err != nil {
		t.Fatal(err)
	}
	otherOccurrences := testPhysicalOccurrences(otherRoot, other.occurrences)
	if len(otherOccurrences) != 1 || otherOccurrences[0].Path != otherMain || len(testPhysicalStatements(otherRoot, other.SourceStatements)) != 0 || !reflect.DeepEqual(testProgramPaths(other.SourcePaths, otherRoot), []string{otherMain}) {
		t.Fatal("independent source inventory changed")
	}
}
