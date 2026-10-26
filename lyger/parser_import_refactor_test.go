package Lyger

import (
	"bytes"
	"reflect"
	"strings"
	"testing"
)

func TestFixedParserOperationsPreserveBytes(t *testing.T) {
	for _, tc := range []struct {
		name                        string
		mode                        parseKind
		source, replacement, result string
		matches                     []string
	}{
		{"stamps", parseLineStamp, "before $$HEAD<雪.l:7>HEAD$$ between $$HEAD<local-words.l:9>HEAD$$ after", "FILE:$1", "before FILE:雪.l:7 between FILE:local-words.l:9 after", []string{"$$HEAD<雪.l:7>HEAD$$", "$$HEAD<local-words.l:9>HEAD$$"}},
		{"inputs", parseInputs, "<@x|Text@> + <@y|Number@>", "($1)", "(x|Text) + (y|Number)", []string{"<@x|Text@>", "<@y|Number@>"}},
		{"multiline input", parseInputs, "before <@first\r\nsecond@> after", "$1", "before first\r\nsecond after", []string{"<@first\r\nsecond@>"}},
		{"absent stamp", parseLineStamp, "plain 雪", "", "plain 雪", nil},
		{"absent input", parseInputs, "plain 雪", "", "plain 雪", nil},
	} {
		t.Run(tc.name, func(t *testing.T) {
			source := []byte(tc.source)
			parser := newParser(source)
			matches, err := parser.parse(tc.mode)
			if err != nil {
				t.Fatal(err)
			}
			var got []string
			for _, match := range matches {
				got = append(got, string(match))
			}
			if !reflect.DeepEqual(got, tc.matches) || (tc.matches == nil && matches != nil) {
				t.Fatalf("matches = %q, want %q", matches, tc.matches)
			}
			result, err := parser.replaceAll(tc.mode, []byte(tc.replacement))
			if err != nil || string(result) != tc.result {
				t.Fatalf("replacement = %q, %v; want %q", result, err, tc.result)
			}
			if string(source) != tc.source || !bytes.Equal(parser.contentBytes, source) {
				t.Fatal("fixed parser operation mutated its source")
			}
		})
	}
}

func BenchmarkFixedParserOperations(b *testing.B) {
	for _, tc := range []struct {
		name   string
		mode   parseKind
		source string
	}{
		{"stamps", parseLineStamp, strings.Repeat("$$HEAD<file.l:7>HEAD$$ body\n", 32)},
		{"inputs", parseInputs, strings.Repeat("<@value|Text@> body\n", 32)},
	} {
		b.Run(tc.name, func(b *testing.B) {
			source := []byte(tc.source)
			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				parser := newParser(source)
				matches, err := parser.parse(tc.mode)
				if err != nil || len(matches) != 32 {
					b.Fatalf("matches = %d, %v", len(matches), err)
				}
				result, err := parser.replaceAll(tc.mode, nil)
				if err != nil || len(result) == 0 {
					b.Fatalf("replacement = %q, %v", result, err)
				}
			}
		})
	}
}

func TestParserModesAreChecked(t *testing.T) {
	for _, tc := range []struct {
		kind parseKind
		name string
	}{
		{parseInvalid, ""}, {parseLineStamp, "lineStamp"}, {parseInputs, "inputs"},
		{parseKindCount, "parseKind(3)"}, {parseKind(255), "parseKind(255)"},
	} {
		if got := tc.kind.String(); got != tc.name {
			t.Errorf("mode %d = %q, want %q", tc.kind, got, tc.name)
		}
	}
	parser := newParser([]byte("unchanged"))
	for _, invalid := range []parseKind{parseInvalid, parseKindCount, parseKind(255)} {
		if _, err := parser.parse(invalid); err == nil {
			t.Errorf("parse accepted invalid mode %d", invalid)
		}
		if _, err := parser.replaceAll(invalid, []byte("replacement")); err == nil {
			t.Errorf("replacement accepted invalid mode %d", invalid)
		}
		if string(parser.contentBytes) != "unchanged" {
			t.Fatal("invalid mode mutated parser contents")
		}
	}
}

func TestImportRecognitionRetainsFormsAndOwnership(t *testing.T) {
	gram := scopeTestGrammar(t)
	if gram == nil {
		t.Fatal("fixture grammar failed to load")
	}
	for _, tc := range []struct {
		name, source, module, member, path, alias string
	}{
		{"file", " uSe `dir with spaces/local-words.l`. *** note ***\r\n", "", "", "dir with spaces/local-words.l", ""},
		{"escaped path", "Use `dir/escaped\\`name.l`.\n", "", "", "dir/escaped`name.l", ""},
		{"member", "Use Render from `local-words.l`.\n", "", "Render", "local-words.l", ""},
		{"qualified", "Use Tools's Render from `local-words.l`.\n", "Tools", "Render", "local-words.l", ""},
		{"plural owner", "Use Tools’ Render from `local-words.l`.\n", "Tools", "Render", "local-words.l", ""},
		{"alias", "Use Tools’s Render from `local-words.l` as \"Pretty \\\"Print\\\"\".\n", "Tools", "Render", "local-words.l", "Pretty \"Print\""},
	} {
		t.Run(tc.name, func(t *testing.T) {
			raw := []byte(tc.source)
			matched, got, err := gram.isImportStatement(raw)
			if err != nil {
				t.Fatal(err)
			}
			want := &importStatement{statementBytes: []byte(tc.source), module: tc.module, keyWord: tc.member, filePath: tc.path, alias: tc.alias}
			if !matched || !reflect.DeepEqual(got, want) {
				t.Fatalf("import = %+v / %t, want %+v", got, matched, want)
			}
			raw[0] = 'X'
			if string(got.statementBytes) != tc.source {
				t.Fatal("import statement borrows the source buffer")
			}
		})
	}
	for _, source := range []string{
		"Use Render as \"Alias\".\n", "Use Render from.\n", "Use Tools's Other's Render from `local-words.l`.\n",
		"Use Tools's `local-words.l`.\n", "Use `local-words.l` trailing.\n", "Use `unterminated", "Show `ordinary Text`.\n", "",
	} {
		if matched, _, _ := gram.isImportStatement([]byte(source)); matched { //nolint:errcheck // Malformed examples intentionally exercise nonrecognition, independently of their parse errors.
			t.Errorf("non-import recognized: %q", source)
		}
	}
}
