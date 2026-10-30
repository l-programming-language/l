package Logiceme

import (
	"reflect"
	"testing"
)

func TestScannerRegexPreservesMultilineFrames(t *testing.T) {
	for _, frame := range []struct {
		kind, subtype, open, close string
	}{
		{"Comment", "", "/*|", "|*\\"},
		{"Array", "", "|[", "]|"},
		{"Hashmap", "", "|{", "}|"},
		{"Tag", "BOOK", "<BOOK", "</BOOK>"},
		{"Variable", "", `"`, `">`},
		{"Word", "", "<|W|>", "</W|>"},
		{"Syllable", "", "<|S|>", "</S|>"},
		{"MetaDefs", "", "{|", "|}"},
		{"GlyphExps", "", "(", ")"},
		{"Structs", "Op", "Op{", "}Op"},
		{"StructArray", "", "&[", "]&"},
		{"StructHashmap", "", "#[", "]#"},
	} {
		t.Run(frame.kind, func(t *testing.T) {
			parser, err := LoadRegex(frame.kind, frame.subtype)
			if err != nil {
				t.Fatal(err)
			}
			// Keep exact bytes, the first closing delimiter, adjacent frames,
			// empty interiors, and an unmatched trailing opener distinct.
			payload := "alpha\r\nbeta\rgamma\n雪\xff"
			first := frame.open + payload + frame.close
			empty := frame.open + frame.close
			input := []byte("prefix " + first + empty + " suffix " + frame.open + "unfinished")
			want := [][][]byte{{[]byte(first), []byte(payload)}, {[]byte(empty), []byte{}}}
			if got := parser.Reg.FindAllSubmatch(input, -1); !reflect.DeepEqual(got, want) {
				t.Fatalf("frames = %q; want %q", got, want)
			}
			if got := parser.Parse([]byte(frame.open + "unfinished")); len(got) != 0 {
				t.Fatalf("unclosed frame was accepted: %q", got)
			}
		})
	}
}
