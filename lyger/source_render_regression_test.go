package Lyger

import (
	"bytes"
	"reflect"
	"testing"
)

func TestExpandedSourceReplacementDoesNotRevisitInsertedLiterals(t *testing.T) {
	const first = "`11111111-1111-1111-1111-111111111111`"
	const second = "`22222222-2222-2222-2222-222222222222`"
	// These fixed prepared IDs make the renderer boundary deterministic. Each
	// valid escaped source frame produces the other ID as ordinary display text.
	// Both appear in one statement, so either map iteration order revisits one
	// inserted value in the old repeated-replacement implementation.
	firstRaw := []byte("`<DEF>A \\`22222222-2222-2222-2222-222222222222\\`<DEF>`")
	secondRaw := []byte("`<DEF>B \\`11111111-1111-1111-1111-111111111111\\`<DEF>`")
	cache := map[string][]byte{first: bytes.Clone(firstRaw), second: bytes.Clone(secondRaw)}
	prepared := []byte("Show " + first + " and then show " + second + ".\n")
	raw := []byte("Show " + string(firstRaw) + " and then show " + string(secondRaw) + ".\n")
	statements := []SourceStatement{{Prepared: bytes.Clone(prepared), Raw: bytes.Clone(raw)}}
	want := []byte("Show `A " + second + "` and then show `B " + first + "`.")
	for attempt := 0; attempt < 3; attempt++ {
		got := renderExpandedSource(statements, cache)
		if !bytes.Equal(got, want) {
			t.Fatalf("inserted literal was interpreted as a prepared key:\ngot  %q\nwant %q", got, want)
		}
		got[0] = 'X'
	}
	if !bytes.Equal(statements[0].Prepared, prepared) || !bytes.Equal(statements[0].Raw, raw) || !bytes.Equal(cache[first], firstRaw) || !bytes.Equal(cache[second], secondRaw) {
		t.Fatal("display rendering mutated the prepared statement, raw source, or literal pool")
	}
}

func TestExpandedSourceDisplayDecodesEachFrameOnce(t *testing.T) {
	ordinary := "`" + `ordinary 雪 \"quoted\" \\n <DEF> marker` + "`"
	framed := "`<DEF>" + `E{|Statement: ~::std::concat!(\"雪\", "\\\\n")~|}` + "<DEF>`"
	displayed := "`" + `E{|Statement: ~::std::concat!("雪", "\\n")~|}` + "`"
	nested := "`<DEF><DEF>keep this body<DEF><DEF>`"
	source := []byte("\"ordinary\" is " + ordinary + ".\r\nDefine \"Quoted\" as " + framed + ".\nDefine \"Nested\" as " + nested + ".")
	prepared, err := PrepareSource("render 雪.l", source)
	if err != nil {
		t.Fatal(err)
	}
	parserView := bytes.Clone(prepared.PreparedBytes)
	cache := make(map[string][]byte, len(prepared.TextCache))
	for key, value := range prepared.TextCache {
		cache[key] = bytes.Clone(value)
	}
	want := []byte("\"ordinary\" is " + ordinary + ".\nDefine \"Quoted\" as " + displayed + ".\nDefine \"Nested\" as `<DEF>keep this body<DEF>`.")
	for attempt := 0; attempt < 3; attempt++ {
		if got := renderExpandedSource(prepared.Statements, prepared.TextCache); !bytes.Equal(got, want) {
			t.Fatalf("frame display changed escaping/order/EOF:\ngot  %q\nwant %q", got, want)
		}
	}
	if !bytes.Equal(prepared.Original, source) || !bytes.Equal(prepared.PreparedBytes, parserView) || !reflect.DeepEqual(prepared.TextCache, cache) {
		t.Fatal("display decoding changed source or semantic literal storage")
	}
}
