package Lyger

import (
	"bytes"
	"reflect"
	"strings"
	"testing"
)

func TestPrepareSourcePreservesOriginalProvenance(t *testing.T) {
	literal := "`first.\n\nsecond *** literal *** \\` quoted`"
	raw := []byte("*** lead ***\n\"x\" is " + literal + ".\n*** comment-only ***\n\"y\" is `two`.\n\n  \"z\" is `three`.\n")
	prepared, err := PrepareSource("source/main.l", raw)
	if err != nil {
		t.Fatal(err)
	}
	if len(prepared.Statements) != 3 || len(prepared.Literals) != 3 {
		t.Fatalf("got %d statements and %d literals", len(prepared.Statements), len(prepared.Literals))
	}
	var restored []byte
	for index, statement := range prepared.Statements {
		restored = append(restored, statement.Raw...)
		wantLine := []int{2, 6, 8}[index]
		wantColumn := []int{1, 1, 3}[index]
		wantParagraph := []int{1, 1, 2}[index]
		if statement.Span.Path != prepared.Path || statement.Span.Start.Line != wantLine || statement.Span.Start.Column != wantColumn || statement.Span.Paragraph != wantParagraph || statement.Span.Order != index {
			t.Errorf("statement %d source span = %+v", index, statement.Span)
		}
		if !bytes.HasPrefix(statement.Prepared, newLineStamp(wantLine, prepared.Path)) {
			t.Errorf("statement %d lost original source line: %s", index, statement.Prepared)
		}
		if !bytes.Equal(statement.Terminator, []byte(".\n")) {
			t.Errorf("statement %d terminator = %q", index, statement.Terminator)
		}
	}
	if !bytes.Equal(restored, raw) || !bytes.Equal(prepared.Original, raw) {
		t.Fatal("original bytes were not preserved")
	}
	first := prepared.Literals[0]
	if string(first.Raw) != literal || string(prepared.TextCache[first.ID]) != literal {
		t.Fatalf("literal changed before caching: %q", first.Raw)
	}
	if string(raw[first.Span.Start.Offset:first.Span.End.Offset]) != literal || first.Span.Start.Line != 2 || first.Span.End.Line != 4 {
		t.Fatalf("literal span = %+v", first.Span)
	}
	if bytes.Contains(prepared.PreparedBytes, []byte("literal ***")) || bytes.Contains(first.Raw, []byte("$$HEAD")) {
		t.Fatal("cache and parser views were mixed")
	}
	raw[0] = 'x'
	if prepared.Original[0] != '*' || prepared.Statements[0].Raw[0] != '*' {
		t.Fatal("prepared source aliases caller's input buffer")
	}
}

func TestPrepareSourceParagraphsIgnoreBlankLinesInsideComments(t *testing.T) {
	prepared, err := PrepareSource("main.l", []byte("\"x\" is `x`.\n*** multi\n\nline comment ***\n\"y\" is `y`.\n \t\n\"z\" is `z`."))
	if err != nil {
		t.Fatal(err)
	}
	var paragraphs []int
	for _, statement := range prepared.Statements {
		paragraphs = append(paragraphs, statement.Span.Paragraph)
	}
	if !reflect.DeepEqual(paragraphs, []int{1, 1, 2}) {
		t.Fatalf("source paragraphs = %v", paragraphs)
	}
	if string(prepared.Statements[2].Terminator) != "." {
		t.Fatalf("EOF terminator = %q", prepared.Statements[2].Terminator)
	}
}

func TestPrepareSourceRejectsUnterminatedSource(t *testing.T) {
	for _, source := range []string{"\"x\" is `open.", "\"open", "*** open"} {
		if _, err := PrepareSource("bad.l", []byte(source)); err == nil || !strings.Contains(err.Error(), "bad.l:") {
			t.Errorf("PrepareSource(%q) error = %v", source, err)
		}
	}
}

func TestImportScanPreservesRawAndCachedIdentity(t *testing.T) {
	raw := []byte("  Use  Tools’s Render from `lib folder/local-words.l` as \"Pretty Print\". *** note ***\r\n")
	gram := scopeTestGrammar(t)
	isImport, imported, err := gram.isImportStatement(raw)
	if err != nil || !isImport || imported == nil {
		t.Fatalf("raw import scan = %v, %v", imported, err)
	}
	want := &importStatement{statementBytes: raw, module: "Tools", keyWord: "Render", filePath: "lib folder/local-words.l", alias: "Pretty Print"}
	if !reflect.DeepEqual(imported, want) {
		t.Fatalf("raw import = %+v, want %+v", imported, want)
	}
	prepared, err := PrepareSource("main.l", raw)
	if err != nil {
		t.Fatal(err)
	}
	if len(prepared.Statements) != 1 {
		t.Fatalf("prepared import split into %d statements", len(prepared.Statements))
	}
	isImport, got, err := gram.isImportStatement(prepared.Statements[0].Prepared)
	if err != nil || !isImport || got == nil {
		t.Fatalf("prepared import scan = %v, %v", got, err)
	}
	if !bytes.Equal(got.statementBytes, prepared.Statements[0].Prepared) || got.module != want.module || got.keyWord != want.keyWord || got.alias != want.alias {
		t.Fatalf("cached import lost raw statement or normalized fields: %+v", got)
	}
	if string(prepared.TextCache["`"+got.filePath+"`"]) != "`lib folder/local-words.l`" {
		t.Fatalf("cached path %q no longer resolves to its original literal", got.filePath)
	}
	if isImport, _, err := gram.isImportStatement([]byte("Use Tools's Render as \"Pretty Print\".\n")); err != nil || isImport {
		t.Fatal("member alias without a file was treated as a dependency")
	}
}

func TestParserRetainsQuotedWordsAndIgnoresLiteralTerminators(t *testing.T) {
	source := []byte("Use Tools's Render from `nested folder/file.l` as \"Pretty Print\".\n\"x\" is `a.\nb`.\nExit.")
	statements := newParser(source).parseStatements()
	if len(statements) != 3 || string(statements[2]) != "Exit" {
		t.Fatalf("statements = %q", statements)
	}
	words := newParser(statements[0]).parseTokens()
	if len(words) != 7 || string(words[4]) != "`nested folder/file.l`" || string(words[6]) != "\"Pretty Print\"" {
		t.Fatalf("quoted words were split: %q", words)
	}
}
