package Lyger

import (
	"bytes"
	"math/rand"
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

func TestParserFramesKeepMultilineRepetition(t *testing.T) {
	for _, kind := range []parseKind{parseLineStamp, parseInputs} {
		pattern, err := kind.pattern()
		if err != nil {
			t.Fatal(err)
		}
		oracle := regexp.MustCompile(regexp.QuoteMeta(string(pattern.open)) + `((?:.*?\r?\n?)*)` + regexp.QuoteMeta(string(pattern.close)))
		pieces := [][]byte{pattern.open, pattern.close, []byte("x"), []byte("\r"), []byte("\n"), []byte("\xff")}
		random := rand.New(rand.NewSource(1)) //nolint:gosec // Fixed parser differential corpus, not security randomness.
		for sample := 0; sample < 2000; sample++ {
			var source []byte
			for count := random.Intn(32); count > 0; count-- {
				source = append(source, pieces[random.Intn(len(pieces))]...)
			}
			parser := newParser(source)
			got, err := parser.parse(kind)
			if want := oracle.FindAll(source, -1); err != nil || !reflect.DeepEqual(got, want) {
				t.Fatalf("%s parse %q: %q, %v; want %q", kind, source, got, err, want)
			}
			gotReplacement, err := parser.replaceAll(kind, []byte("<$1>"))
			if want := oracle.ReplaceAll(source, []byte("<$1>")); err != nil || !reflect.DeepEqual(gotReplacement, want) {
				t.Fatalf("%s replace %q: %q, %v; want %q", kind, source, gotReplacement, err, want)
			}
		}
	}
}
