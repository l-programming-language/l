package Lyger

import (
	"bytes"
	"os"
	"reflect"
	"strconv"
	"strings"
	"testing"
)

func TestLiteralFramingPreservesSourceText(t *testing.T) {
	for _, tc := range []struct {
		name, literal, displayed string
	}{
		{"plain", "`ordinary text`", "`ordinary text`"},
		{"embedded", "`literal <DEF> tag`", "`literal <DEF> tag`"},
		{"single marker", "`<DEF>`", "`<DEF>`"},
		{"opening marker", "`<DEF> opening`", "`<DEF> opening`"},
		{"closing marker", "`closing <DEF>`", "`closing <DEF>`"},
		{"interior pair", "`before <DEF>A<DEF> after`", "`before <DEF>A<DEF> after`"},
		{"inner whitespace", "` <DEF>A<DEF> `", "` <DEF>A<DEF> `"},
		{"escaped unicode", "`" + `雪 <DEF> \"quoted\" \\path` + "`", "`" + `雪 <DEF> \"quoted\" \\path` + "`"},
		{"paired frame", "`<DEF>W^<@Parts|Syllables@><DEF>`", "`W^<@Parts|Syllables@>`"},
		{"definition escaping", "`" + `<DEF>E{|Value: ~\"snow 雪\" \\path~|}<DEF>` + "`", "`" + `E{|Value: ~"snow 雪" \path~|}` + "`"},
		{"empty frame display", "`<DEF><DEF>`", "``"},
		{"nested frame display", "`<DEF>A<DEF><DEF>`", "`A<DEF>`"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			raw := []byte("\"value\" is " + tc.literal + ".\n")
			prepared, err := PrepareSource("literal.l", raw)
			if err != nil {
				t.Fatal(err)
			}
			originalPrepared := bytes.Clone(prepared.PreparedBytes)
			originalLiteral := prepared.Literals[0]
			originalLiteral.Raw = bytes.Clone(originalLiteral.Raw)
			want := "\"value\" is " + tc.displayed + "."
			for attempt := 0; attempt < 2; attempt++ {
				if got := string(renderExpandedSource(prepared.Statements, prepared.TextCache)); got != want {
					t.Errorf("expanded source = %q, want %q", got, want)
				}
			}
			if !bytes.Equal(prepared.Original, raw) || !bytes.Equal(prepared.PreparedBytes, originalPrepared) || !reflect.DeepEqual(prepared.Literals[0], originalLiteral) || string(prepared.TextCache[originalLiteral.ID]) != tc.literal {
				t.Fatal("display rendering changed original source, cache or literal provenance")
			}
		})
	}
}

func TestLiteralFramingClassification(t *testing.T) {
	for _, tc := range []struct {
		name, text string
		definition bool
	}{
		{"plain", "`ordinary text`", false},
		{"embedded", "`literal <DEF> tag`", false},
		{"single marker", "`<DEF>`", false},
		{"opening marker", "`<DEF> opening`", false},
		{"closing marker", "`closing <DEF>`", false},
		{"inner whitespace", "` <DEF>A<DEF> `", false},
		{"bare frame is not a source literal", "<DEF>A<DEF>", false},
		{"paired", "`<DEF>A<DEF>`", true},
		{"empty frame is validated by parser", "`<DEF><DEF>`", true},
		{"nested frame is validated by parser", "`<DEF>A<DEF><DEF>`", true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			literal := &token{rawBytes: []byte(tc.text), mem: &entry{Type: storagePrimitive, Primitive: &primitive{Type: valueText, Text: tc.text}}}
			value, err := literal.mem.value()
			if err != nil {
				t.Fatal(err)
			}
			if got := value.isDef(); got != tc.definition {
				t.Errorf("value classification = %v, want %v", got, tc.definition)
			}
			selected, err := definitionLiteral([]*token{literal})
			if err != nil {
				t.Fatal(err)
			}
			if (selected == literal) != tc.definition {
				t.Errorf("semantic template selection = %v, want %v", selected == literal, tc.definition)
			}
			if literal.mem.Primitive.Text != tc.text || string(literal.rawBytes) != tc.text {
				t.Fatal("classification changed literal bytes")
			}
		})
	}
}

func TestLiteralFramingParserForms(t *testing.T) {
	span := SourceSpan{Path: "literal.l", Start: SourcePosition{Line: 7, Column: 3}}
	for _, source := range []string{"A", "`A`", "<DEF>A<DEF>", "`<DEF>A<DEF>`", " \t`<DEF>A<DEF>`\n "} {
		t.Run("valid "+source, func(t *testing.T) {
			template, err := ParseDefinitionTemplate([]byte(source), span)
			if err != nil {
				t.Fatal(err)
			}
			instance, err := template.Instantiate(nil)
			if err != nil || instance.Expression != "A" {
				t.Fatalf("definition = %q, error = %v", instance.Expression, err)
			}
			if string(template.Original) != source || !reflect.DeepEqual(template.Span, span) {
				t.Fatal("definition parser changed original bytes or provenance")
			}
		})
	}
	for _, tc := range []struct{ source, message string }{
		{"`", "unterminated definition literal"},
		{"`<DEF>A<DEF>", "unterminated definition literal"},
		{"<DEF>", "definition requires paired <DEF> markers"},
		{"`<DEF>`", "definition requires paired <DEF> markers"},
		{"<DEF>A", "definition requires paired <DEF> markers"},
		{"A<DEF>", "definition requires paired <DEF> markers"},
		{"before <DEF>A<DEF> after", "definition requires paired <DEF> markers"},
		{"` <DEF>A<DEF> `", "definition requires paired <DEF> markers"},
		{"<DEF>A<DEF><DEF>", "nested <DEF> markers"},
		{"`<DEF>A<DEF><DEF>`", "nested <DEF> markers"},
		{"<DEF><DEF>", "empty definition"},
		{"`<DEF><DEF>`", "empty definition"},
	} {
		t.Run("invalid "+tc.source, func(t *testing.T) {
			_, err := ParseDefinitionTemplate([]byte(tc.source), span)
			want := "literal.l:7:3: definition: " + tc.message
			if err == nil || err.Error() != want {
				t.Fatalf("error = %v, want %q", err, want)
			}
		})
	}
}

func TestLiteralFramingLexemeStorage(t *testing.T) {
	for _, tc := range []struct{ name, source, value, kind string }{
		{"quoted sentence", "`<DEF>[[\"A\"]]<DEF>`", "`[[\"A\"]]`", "generic.sentence"},
		{"quoted word", "`<DEF>[\"A\"]<DEF>`", "`[\"A\"]`", "generic.word"},
		{"quoted syllable", "`<DEF>W^<@Body@><DEF>`", "`W^<@Body@>`", "generic.syllable"},
		{"bare syllable", "<DEF>W^<@Body@><DEF>", "W^<@Body@>", "generic.syllable"},
		{"encoded body", "`" + `<DEF>E{|Value: ~\"quote\" \\path~|}<DEF>` + "`", "`" + `E{|Value: ~\"quote\" \\path~|}` + "`", "generic.syllable"},
		{"interior marker awaits parser validation", "`<DEF>A<DEF><DEF>`", "`A<DEF>`", "generic.syllable"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			entry, err := initLexicalEntry([]byte(tc.source))
			if err != nil {
				t.Fatal(err)
			}
			if entry.Lexeme.Value != tc.value || entry.Lexeme.Type != tc.kind {
				t.Fatalf("lexeme value/type = %q/%q, want %q/%q", entry.Lexeme.Value, entry.Lexeme.Type, tc.value, tc.kind)
			}
		})
	}
	for _, source := range []string{"ordinary text", "`literal <DEF> tag`", "`<DEF>`", "<DEF>opening"} {
		_, err := initLexicalEntry([]byte(source))
		want := "invalid lexeme definition: " + source
		if err == nil || err.Error() != want {
			t.Errorf("lexeme definition %q error = %v, want %q", source, err, want)
		}
	}
}

func TestLiteralFramingCompilePreservesOrdinaryText(t *testing.T) {
	original, err := os.ReadFile("test/main.l")
	if err != nil {
		t.Fatal(err)
	}
	config := canonicalTestOptions()
	for _, message := range []string{"ordinary tag", "literal <DEF> tag", "<DEF>", "雪 <DEF> \"quoted\" text"} {
		t.Run(message, func(t *testing.T) {
			literal := "`" + message + "`"
			source := bytes.Replace(original, []byte("`Hello World!`"), []byte(literal), 1)
			result, err := NewLygerService(config).Compile(source, "test/main.l")
			if err != nil {
				t.Fatal(err)
			}
			if !bytes.Contains(result.ExpandedSource, []byte("\"x\" is "+literal+".")) {
				t.Errorf("expanded source lost ordinary literal %q", literal)
			}
			x := result.Compilation.Memory["x"].Value
			if x.Type.String() != "Primitive" || x.Primitive == nil || x.Primitive.Type.String() != "Text" || x.Primitive.Text != literal {
				t.Errorf("ordinary text was reinterpreted: %#v", x)
			}
			if !strings.Contains(string(result.AML), "Inputs: ~&["+strconv.Quote(message)+"]&~") {
				t.Errorf("generated statement lost ordinary literal %q", message)
			}
		})
	}
}
