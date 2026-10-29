package Lyger

import (
	"bytes"
	"reflect"
	"regexp"
	"testing"
)

// The retained expression implementation is an independent oracle for fixed
// framing and replacement captures, including malformed and arbitrary bytes.
func TestParserFramesPreserveExpressionSemantics(t *testing.T) {
	for _, tc := range []struct {
		kind              parseKind
		open, close, expr string
	}{
		{parseLineStamp, "$$HEAD<", ">HEAD$$", `\$\$HEAD\<((?:.*?\r?\n?)*)\>HEAD\$\$`},
		{parseInputs, "<@", "@>", `\<\@((?:.*?\r?\n?)*)\@\>`},
	} {
		t.Run(tc.kind.String(), func(t *testing.T) {
			oracle := regexp.MustCompile(tc.expr)
			parts := []string{"", "plain", "雪\r\n\x00\xff", tc.open, tc.close, tc.open + tc.close,
				tc.open + "one\r\ntwo" + tc.close, tc.open + tc.open + "nested" + tc.close,
				tc.open + "one" + tc.close + tc.open + "two" + tc.close}
			for _, left := range parts {
				for _, right := range parts {
					source := []byte(left + right)
					before := bytes.Clone(source)
					parser := newParser(source)
					got, err := parser.parse(tc.kind)
					if err != nil || !reflect.DeepEqual(got, oracle.FindAll(source, -1)) {
						t.Fatalf("parse %q: %q, %v; want %q", source, got, err, oracle.FindAll(source, -1))
					}
					for _, replacement := range []string{"", "literal", "$1", "${1}x", "$0/$1/$$", "$2/$name", "$1suffix", "$", "\xff$1"} {
						got, err := parser.replaceAll(tc.kind, []byte(replacement))
						want := oracle.ReplaceAll(source, []byte(replacement))
						if err != nil || !reflect.DeepEqual(got, want) {
							t.Fatalf("replace %q with %q: %q, %v; want %q", source, replacement, got, err, want)
						}
						if len(got) > 0 {
							got[0] ^= 1
						}
						if !bytes.Equal(source, before) {
							t.Fatal("replacement changed borrowed source bytes")
						}
					}
				}
			}
		})
	}
}
