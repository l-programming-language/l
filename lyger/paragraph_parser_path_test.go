package Lyger

import (
	"bytes"
	"reflect"
	"testing"
)

func TestParagraphParserPreservesPathStamps(t *testing.T) {
	for _, path := range []string{"folder copy/main.l", "folder. copy/main.l", "folder/main. copy.l", "雪. copy/main.l"} {
		t.Run(path, func(t *testing.T) {
			prepared, err := PrepareSource(path, []byte("Show `first`. Show `second`.\n"))
			if err != nil {
				t.Fatal(err)
			}
			before := bytes.Clone(prepared.PreparedBytes)
			parsed := newParser(prepared.PreparedBytes).parseStatements()
			if !bytes.Equal(prepared.PreparedBytes, before) {
				t.Fatal("parsing mutated its caller's stamped source")
			}
			if len(prepared.Statements) != 2 || len(parsed) != 2 {
				t.Fatalf("prepared/parser statements = %d/%d; want 2/2: %q", len(prepared.Statements), len(parsed), parsed)
			}
			for i, statement := range prepared.Statements {
				want := bytes.TrimSuffix(statement.Prepared, []byte(".\n"))
				if !bytes.Equal(parsed[i], want) {
					t.Errorf("statement %d changed its stamp or cached body: %q; want %q", i, parsed[i], want)
				}
			}
		})
	}
}

func TestParagraphParserPreservesRawQuotedMarkerFragments(t *testing.T) {
	for _, tc := range []struct {
		name, source string
		want         []string
	}{
		{
			name:   "whole marker inside Text",
			source: "Show `$$HEAD<folder. copy/main.l:9>HEAD$$`. Show `after`.\n",
			want:   []string{"Show `$$HEAD<folder. copy/main.l:9>HEAD$$`", "Show `after`"},
		},
		{
			name:   "marker fragments in separate Text values",
			source: "Show `$$HEAD<literal`. Show `>HEAD$$`. Show `after`.\n",
			want:   []string{"Show `$$HEAD<literal`", "Show `>HEAD$$`", "Show `after`"},
		},
		{
			name:   "marker fragments in quoted identifiers",
			source: "Show \"$$HEAD<identifier\". Show \">HEAD$$\".\n",
			want:   []string{"Show \"$$HEAD<identifier\"", "Show \">HEAD$$\""},
		},
		{
			name:   "escaped quote before marker text",
			source: "Show `escaped \\` $$HEAD<folder. copy/main.l:9>HEAD$$`. Show `after`.\n",
			want:   []string{"Show `escaped \\` $$HEAD<folder. copy/main.l:9>HEAD$$`", "Show `after`"},
		},
		{
			name:   "marker fragments in separate comments",
			source: "*** $$HEAD<comment *** Show `one`. *** >HEAD$$ *** Show `two`.\n",
			want:   []string{"*** $$HEAD<comment *** Show `one`", "Show `two`"},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			input := []byte(tc.source)
			parsed := newParser(input).parseStatements()
			var got []string
			for _, statement := range parsed {
				got = append(got, string(statement))
			}
			if !reflect.DeepEqual(got, tc.want) || string(input) != tc.source {
				t.Fatalf("raw marker/literal source changed: %q; want %q", got, tc.want)
			}
		})
	}
}

func TestParagraphParserScanRetainsStatementProvenance(t *testing.T) {
	const path = "project. copy/main.l"
	prepared, err := PrepareSource(path, []byte("    \"message\" is `one`. Show message.\n"))
	if err != nil {
		t.Fatal(err)
	}
	config := canonicalTestOptions()
	gram, err := loadGrammar(config)
	if err != nil {
		t.Fatal(err)
	}
	compilation := &Compilation{Gram: gram, CompiledBytes: prepared.PreparedBytes, SourceStatements: prepared.Statements}
	compilation.scan()
	if compilation.Error != nil {
		t.Fatal(compilation.Error)
	}
	if len(compilation.Statements) != 2 {
		t.Fatalf("scan produced %d statements; want 2", len(compilation.Statements))
	}
	for i, statement := range compilation.Statements {
		if statement.source != &compilation.SourceStatements[i] || statement.source.Span != prepared.Statements[i].Span || !bytes.Equal(statement.headBytes, newLineStamp(1, path)) {
			t.Errorf("statement %d lost its original source association: %+v / %q", i, statement.source, statement.headBytes)
		}
	}
}
