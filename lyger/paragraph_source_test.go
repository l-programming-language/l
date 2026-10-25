package Lyger

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
)

type paragraphSourceSentence struct {
	text                    string
	line, column, paragraph int
}

func assertParagraphSource(t *testing.T, path, source string, want []paragraphSourceSentence, literals []string) {
	t.Helper()
	// Check the independently specified byte columns against the input before
	// the production statement count can fail and hide a bad test expectation.
	lines := bytes.Split([]byte(source), []byte("\n"))
	starts := make([]int, len(want))
	for i, expected := range want {
		if expected.line < 1 || expected.line > len(lines) || expected.column < 1 || expected.column > len(lines[expected.line-1]) {
			t.Fatalf("test expectation %d has an invalid line/byte column", i)
		}
		start := expected.column - 1
		for _, line := range lines[:expected.line-1] {
			start += len(line) + 1
		}
		end := start + len(expected.text)
		if end > len(source) || source[start:end] != expected.text {
			t.Fatalf("test expectation %d does not match line %d byte column %d in the original source", i, expected.line, expected.column)
		}
		starts[i] = start
	}
	prepared, err := PrepareSource(path, []byte(source))
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(prepared.Original, []byte(source)) {
		t.Fatal("preparation changed the original source")
	}
	if len(prepared.Statements) != len(want) {
		t.Fatalf("explicit period-delimited statements = %d; want %d", len(prepared.Statements), len(want))
	}
	var reconstructed []byte
	for i, expected := range want {
		statement := prepared.Statements[i]
		span := statement.Span
		start := starts[i]
		if span.Path != path || span.Order != i || span.Paragraph != expected.paragraph || span.Start != (SourcePosition{Offset: start, Line: expected.line, Column: expected.column}) {
			t.Errorf("statement %d span = %+v; want line %d column %d paragraph %d offset %d", i, span, expected.line, expected.column, expected.paragraph, start)
		}
		end := start + len(expected.text)
		if span.End.Offset < end || span.End.Offset > len(source) {
			t.Errorf("statement %d end = %+v; must contain exactly its sentence and following separator", i, span.End)
		} else if len(statement.Terminator) == 0 || statement.Terminator[0] != '.' || span.End.Offset-len(statement.Terminator) != end-1 || !bytes.Equal(statement.Terminator, []byte(source[end-1:span.End.Offset])) {
			t.Errorf("statement %d terminator = %q; lost the original period/separator bytes", i, statement.Terminator)
		}
		if !bytes.HasPrefix(statement.Prepared, newLineStamp(expected.line, path)) {
			t.Errorf("statement %d lost its original line stamp", i)
		}
		parsed := newParser(statement.Prepared).parseStatements()
		if len(parsed) != 1 {
			t.Errorf("statement %d prepared view parsed as %d items; want exactly one", i, len(parsed))
		} else if expectedBody := bytes.TrimSuffix(statement.Prepared, []byte(".\n")); !bytes.Equal(parsed[0], expectedBody) {
			t.Errorf("statement %d parser changed its stamped/cached body: %q; want %q", i, parsed[0], expectedBody)
		}
		if rendered := renderExpandedSource([]SourceStatement{statement}, prepared.TextCache); string(rendered) != expected.text {
			t.Errorf("statement %d rendered as %q; want the exact individual sentence %q", i, rendered, expected.text)
		}
		reconstructed = append(reconstructed, statement.Raw...)
	}
	if !bytes.Equal(reconstructed, []byte(source)) {
		t.Fatal("statement raw spans did not reconstruct the exact original source")
	}
	if len(prepared.Literals) != len(literals) || len(prepared.TextCache) != len(literals) {
		t.Fatalf("literal occurrences/cache entries = %d/%d; want %d", len(prepared.Literals), len(prepared.TextCache), len(literals))
	}
	for i, expected := range literals {
		literal := prepared.Literals[i]
		span := literal.Span
		if string(literal.Raw) != expected || string(prepared.TextCache[literal.ID]) != expected {
			t.Errorf("literal %d changed in the source or cache view: %q / %q", i, literal.Raw, prepared.TextCache[literal.ID])
		}
		if span.Start.Offset < 0 || span.End.Offset > len(source) || span.End.Offset < span.Start.Offset || source[span.Start.Offset:span.End.Offset] != expected {
			t.Errorf("literal %d original span = %+v", i, span)
		}
		if span.Path != path || span.Order < 0 || span.Order >= len(want) || span.Paragraph != want[span.Order].paragraph {
			t.Errorf("literal %d lost its statement/paragraph ownership: %+v", i, span)
		}
	}
}

func TestParagraphSourceDemoSentences(t *testing.T) {
	path := filepath.Join("..", "demo", "paragraphs", "main.l")
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	// The author corrected the declaration's missing period before these tests
	// were added. Use the demo unchanged: no implicit newline terminator or
	// test-only punctuation insertion is part of this contract.
	assertParagraphSource(t, path, string(raw), []paragraphSourceSentence{
		{"Use IO's Print from `io` as \"Print\".", 1, 1, 1},
		{"\"x1\" is `Hello World!` and \"x2\" is `Hello Again World!`.", 2, 1, 1},
		{"Create \"HelloWorldApp1\" as a New Public Basic App.", 5, 5, 2},
		{"The App Prints x1.", 5, 56, 2},
		{"If it has an Error, exit with the Error, else define \"y\" as it's Output.", 5, 75, 2},
		{"If x1 isnt y, exit with an error of `text values of both x1 & y should be Hello World`.", 6, 1, 2},
		{"Otherwise, show y and then exit with error.", 6, 89, 2},
		{"Create \"HelloWorldApp2\" as a New Public Basic App.", 8, 5, 3},
		{"The App Prints x2.", 8, 56, 3},
		{"If it has an Error, exit with the Error, else define \"y\" as it's Output.", 8, 75, 3},
		{"If x2 isnt y, exit with an error of `text values of both x2 & y should be Hello World`.", 9, 1, 3},
		{"Otherwise, show y and then exit with error.", 9, 89, 3},
		{"Execute HelloWorldApp1 and HelloWorldApp2.", 12, 1, 4},
	}, []string{
		"`io`", "`Hello World!`", "`Hello Again World!`",
		"`text values of both x1 & y should be Hello World`",
		"`text values of both x2 & y should be Hello World`",
	})
}

func TestParagraphSourceInlineAndContinuationLines(t *testing.T) {
	source := "    \"first\" is `one.`. Show first.\n" +
		"Exit.\n\n" +
		"    \"second\" is `two.`. Show second.\n" +
		"Exit.\n"
	assertParagraphSource(t, "renamed-paragraphs.l", source, []paragraphSourceSentence{
		{"\"first\" is `one.`.", 1, 5, 1},
		{"Show first.", 1, 24, 1},
		{"Exit.", 2, 1, 1},
		{"\"second\" is `two.`.", 4, 5, 2},
		{"Show second.", 4, 25, 2},
		{"Exit.", 5, 1, 2},
	}, []string{"`one.`", "`two.`"})
}

func TestParagraphSourceInlineUnicodeCRLFByteColumns(t *testing.T) {
	source := "    \"雪\" is `α. β`. Show \"雪\".\r\n" +
		"Show `再見.`.\r\n\r\n" +
		"    \"終\" is `終わり`. Show \"終\".\r\n"
	// UTF-8 makes these inline starts byte columns 24 and 27, rather than
	// character columns. The physical CRLF-only line separates the two blocks.
	assertParagraphSource(t, "雪/paragraphs.l", source, []paragraphSourceSentence{
		{"\"雪\" is `α. β`.", 1, 5, 1},
		{"Show \"雪\".", 1, 24, 1},
		{"Show `再見.`.", 2, 1, 1},
		{"\"終\" is `終わり`.", 4, 5, 2},
		{"Show \"終\".", 4, 27, 2},
	}, []string{"`α. β`", "`再見.`", "`終わり`"})
}

func TestParagraphSourceLiteralAndCommentBoundaries(t *testing.T) {
	literal := "`first. sentence\n\nsecond *** literal *** \\` quoted`"
	source := "    \"payload.name\" is " + literal + ".\n" +
		"*** comment. sentence\n\nstill comment ***\n" +
		"Show \"payload.name\". Show `inline. text`.\n\n" +
		"    \"other\" is `tail.`.\nShow other.\n"
	assertParagraphSource(t, "quoted-paragraphs.l", source, []paragraphSourceSentence{
		{"\"payload.name\" is " + literal + ".", 1, 5, 1},
		{"Show \"payload.name\".", 7, 1, 1},
		{"Show `inline. text`.", 7, 22, 1},
		{"\"other\" is `tail.`.", 9, 5, 2},
		{"Show other.", 10, 1, 2},
	}, []string{literal, "`inline. text`", "`tail.`"})
}

func TestParagraphSourcePeriodsInsideUnquotedValues(t *testing.T) {
	// A decimal's internal period is not a sentence boundary. These controls
	// exercise preprocessing only, without assigning a numeric language type.
	source := "    \"fraction\" is 3.14. Show fraction.\nShow fraction.\n"
	assertParagraphSource(t, "decimal-paragraph.l", source, []paragraphSourceSentence{
		{"\"fraction\" is 3.14.", 1, 5, 1},
		{"Show fraction.", 1, 25, 1},
		{"Show fraction.", 2, 1, 1},
	}, nil)
}
