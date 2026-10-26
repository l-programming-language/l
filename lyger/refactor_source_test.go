package Lyger

import (
	"bytes"
	"os"
	"reflect"
	"testing"
)

func TestRefactorSourceUnicodeCRLFPositions(t *testing.T) {
	raw := []byte("\"雪\" is `α\r\nβ`.\r\n\r\n\t\"終\" is `tail`.")
	prepared, err := PrepareSource("雪/main.l", raw)
	if err != nil {
		t.Fatal(err)
	}
	wantStatements := []SourceSpan{
		{Path: "雪/main.l", Start: SourcePosition{Offset: 0, Line: 1, Column: 1}, End: SourcePosition{Offset: 20, Line: 3, Column: 1}, Paragraph: 1, Order: 0},
		{Path: "雪/main.l", Start: SourcePosition{Offset: 23, Line: 4, Column: 2}, End: SourcePosition{Offset: 39, Line: 4, Column: 18}, Paragraph: 2, Order: 1},
	}
	wantLiterals := []SourceSpan{
		{Path: "雪/main.l", Start: SourcePosition{Offset: 9, Line: 1, Column: 10}, End: SourcePosition{Offset: 17, Line: 2, Column: 4}, Paragraph: 1, Order: 0},
		{Path: "雪/main.l", Start: SourcePosition{Offset: 32, Line: 4, Column: 11}, End: SourcePosition{Offset: 38, Line: 4, Column: 17}, Paragraph: 2, Order: 1},
	}
	if len(prepared.Statements) != len(wantStatements) || len(prepared.Literals) != len(wantLiterals) {
		t.Fatalf("statements/literals = %d/%d", len(prepared.Statements), len(prepared.Literals))
	}
	var reconstructed []byte
	for i, statement := range prepared.Statements {
		if statement.Span != wantStatements[i] {
			t.Errorf("statement %d span = %+v; want %+v", i, statement.Span, wantStatements[i])
		}
		reconstructed = append(reconstructed, statement.Raw...)
		if literal := prepared.Literals[i]; literal.Span != wantLiterals[i] || !bytes.Equal(literal.Raw, raw[literal.Span.Start.Offset:literal.Span.End.Offset]) {
			t.Errorf("literal %d = %+v; want span %+v", i, literal, wantLiterals[i])
		}
	}
	if !bytes.Equal(reconstructed, raw) || string(prepared.Statements[0].Terminator) != ".\r\n" || string(prepared.Statements[1].Terminator) != "." {
		t.Fatal("CRLF/EOF source bytes changed")
	}
}

func TestRefactorSourceParagraphBoundaryLocations(t *testing.T) {
	for _, tc := range []struct {
		name       string
		source     string
		lines      []int
		paragraphs []int
	}{
		{"leading blanks then comment", "\r\n\t\r\n*** comment\r\n\r\n***\r\n\"雪\" is `one`.\r\n", []int{6}, []int{3}},
		{"blank after multiline comment", "\"a\" is `one`.\n*** comment\n\nstill comment ***\n\n\"b\" is `two`.", []int{1, 6}, []int{1, 2}},
		{"blank after multiline literal", "\"a\" is `first\n\nsecond`.\n\n\"b\" is `two`.", []int{1, 5}, []int{1, 2}},
		{"blank after multiline identifier", "\"first\n\nname\" is `one`.\n\n\"b\" is `two`.", []int{1, 5}, []int{1, 2}},
		{"adjacent comment frames", "*** first *** *** second ***\n\"a\" is `one`.", []int{2}, []int{1}},
		{"escaped quote in identifier", "\"a\\\"b\" is `one`.\n\"b\" is `two`.", []int{1, 2}, []int{1, 1}},
		{"comment only line is occupied", "\"a\" is `one`.\n \t*** comment *** \t\n\"b\" is `two`.", []int{1, 3}, []int{1, 1}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			prepared, err := PrepareSource("boundary.l", []byte(tc.source))
			if err != nil {
				t.Fatal(err)
			}
			if len(prepared.Statements) != len(tc.lines) {
				t.Fatalf("statements = %d; want %d", len(prepared.Statements), len(tc.lines))
			}
			for i, statement := range prepared.Statements {
				if statement.Span.Start.Line != tc.lines[i] || statement.Span.Start.Column != 1 || statement.Span.Paragraph != tc.paragraphs[i] {
					t.Errorf("statement %d location = %+v; want line %d column 1 paragraph %d", i, statement.Span, tc.lines[i], tc.paragraphs[i])
				}
				if statement.Span.End.Offset > len(tc.source) || statement.Span.Start.Offset >= statement.Span.End.Offset || tc.source[statement.Span.Start.Offset] != '"' {
					t.Errorf("statement %d offsets = %+v", i, statement.Span)
				}
			}
			for _, literal := range prepared.Literals {
				if !bytes.Equal(literal.Raw, []byte(tc.source)[literal.Span.Start.Offset:literal.Span.End.Offset]) || literal.Span.Paragraph != tc.paragraphs[literal.Span.Order] {
					t.Errorf("literal provenance = %+v", literal)
				}
			}
		})
	}
	for _, source := range []string{"", " \t\r\n\r\n", "*** comment\n\nonly ***\n"} {
		prepared, err := PrepareSource("empty.l", []byte(source))
		if err != nil || len(prepared.Statements) != 0 || !bytes.Equal(prepared.Original, []byte(source)) {
			t.Errorf("non-emitting source %q: %+v, %v", source, prepared, err)
		}
	}
}

func TestRefactorRepeatedImportsOwnLiteralOccurrences(t *testing.T) {
	root := t.TempDir()

	leaf := writeSourceFile(t, root, "leaf.l", "\"message\" is `snow 雪`.\n")
	main := writeSourceFile(t, root, "main.l", "Use `leaf.l`.\nUse `leaf.l`.\n")
	built, err := buildTemporarySource(t, main)
	if err != nil {
		t.Fatal(err)
	}
	if len(testPhysicalStatements(root, built.SourceStatements)) != 2 || !reflect.DeepEqual(testProgramPaths(built.SourcePaths, root), []string{main, leaf, leaf}) {
		t.Fatalf("repeated occurrences = %+v / %v", built.SourceStatements, built.SourcePaths)
	}
	var ids []string
	for id, raw := range built.TextCache {
		if string(raw) == "`snow 雪`" {
			ids = append(ids, id)
		}
	}
	if len(ids) != 2 || ids[0] == ids[1] {
		t.Fatalf("each imported occurrence needs a fresh literal ID: %v", ids)
	}
	for _, id := range ids {
		occurrences := 0
		for _, statement := range built.SourceStatements {
			if bytes.Contains(statement.Prepared, []byte(id)) {
				occurrences++
			}
		}
		if occurrences != 1 {
			t.Fatalf("literal %s belongs to %d statements", id, occurrences)
		}
	}
	secondRaw := bytes.Clone(built.SourceStatements[1].Raw)
	secondPrepared := bytes.Clone(built.SourceStatements[1].Prepared)
	built.TextCache[ids[0]][1] = 'X'
	built.SourceStatements[0].Raw[0] = 'X'
	built.SourceStatements[0].Prepared[0] = 'X'
	if string(built.TextCache[ids[1]]) != "`snow 雪`" || !bytes.Equal(built.SourceStatements[1].Raw, secondRaw) || !bytes.Equal(built.SourceStatements[1].Prepared, secondPrepared) {
		t.Fatal("repeated import occurrences share mutable buffers")
	}
	if disk, err := os.ReadFile(leaf); err != nil || string(disk) != "\"message\" is `snow 雪`.\n" {
		t.Fatalf("source changed: %q, %v", disk, err)
	}
}

func refactorPreparedStatement(tb testing.TB, source string) (*PreparedSource, *lStatement) {
	tb.Helper()
	prepared, err := PrepareSource("literal.l", []byte(source))
	if err != nil {
		tb.Fatal(err)
	}
	config := canonicalTestOptions()
	gram, err := loadGrammar(config)
	if err != nil {
		tb.Fatal(err)
	}
	parts := newParser(prepared.Statements[0].Prepared).parseStatements()
	if len(parts) != 1 {
		tb.Fatalf("prepared statement split into %d parts", len(parts))
	}
	statement, err := newLStatement(parts[0], gram)
	if err != nil {
		tb.Fatal(err)
	}
	statement.source = &prepared.Statements[0]
	if err = statement.tokenize(); err != nil {
		tb.Fatal(err)
	}
	statement.determineType()
	if err = statement.loadGram(); err != nil {
		tb.Fatal(err)
	}
	return prepared, statement
}

func TestRefactorLiteralRestorationSharesGrammarTokens(t *testing.T) {
	for _, literal := range []string{"`snow 雪, ordinary <DEF> text`", "`first\r\nsecond \\` quoted`", "`uuid-like 00000000-0000-0000-0000-000000000000`"} {
		t.Run(literal, func(t *testing.T) {
			prepared, statement := refactorPreparedStatement(t, "\"message\" is "+literal+".\n")
			id := prepared.Literals[0].ID
			var cached *token
			for _, tok := range statement.tokens {
				if string(tok.rawBytes) == id {
					cached = tok
				}
			}
			if cached == nil || statement.sGram == nil || statement.sGram.predicate == nil {
				t.Fatal("prepared literal is not a complete parsed token")
			}
			for name, tokens := range map[string][]*token{"statement": statement.sGram.tokens, "predicate": statement.sGram.predicate.tokens} {
				found := false
				for _, tok := range tokens {
					found = found || tok == cached
				}
				if !found {
					t.Fatalf("%s grammar lost original token identity", name)
				}
			}
			original := bytes.Clone(prepared.Original)
			parserView := bytes.Clone(prepared.PreparedBytes)
			lexicalType, category, binding := cached.tType, cached.category, cached.binding
			if err := statement.uncacheLiterals(prepared.TextCache); err != nil {
				t.Fatal(err)
			}
			if string(cached.rawBytes) != literal || cached.mem == nil || cached.mem.Primitive == nil || cached.mem.Primitive.Text != literal {
				t.Fatalf("literal restoration = %q / %+v", cached.rawBytes, cached.mem)
			}
			if cached.tType != lexicalType || cached.category != category || cached.binding != binding || string(prepared.TextCache[id]) != literal || !bytes.Equal(prepared.Original, original) || !bytes.Equal(prepared.PreparedBytes, parserView) {
				t.Fatal("restoration mutated lexical identity, source, or literal pool")
			}
			prepared.TextCache[id][1] = 'X'
			if string(cached.rawBytes) != literal || cached.mem.Primitive.Text != literal {
				t.Fatal("restored token borrows the mutable literal pool")
			}
		})
	}
}

func TestRefactorInstructionClauseOrderingBoundaries(t *testing.T) {
	a, b, missing := &token{}, &token{}, &token{}
	first := &clauseGram{verb: a}
	tie := &clauseGram{verb: a}
	second := &clauseGram{verb: b}
	unknown := &clauseGram{verb: missing}
	unknownTie := &clauseGram{verb: nil}
	input := []*clauseGram{unknown, second, tie, nil, first, second, unknownTie}
	want := []*clauseGram{tie, first, second, unknown, unknownTie}
	if got := orderedInstructionClauses(input, []*token{a, b, a}); !reflect.DeepEqual(got, want) {
		t.Fatalf("stable first-position ordering = %p; want %p", got, want)
	}
	if input[0] != unknown || input[1] != second || input[3] != nil {
		t.Fatal("ordering mutated its caller's slice")
	}
}
