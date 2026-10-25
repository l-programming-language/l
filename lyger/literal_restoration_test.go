package Lyger

import (
	"bytes"
	"errors"
	"reflect"
	"testing"
)

func literalRestorationToken(t *testing.T, raw string) *token {
	t.Helper()
	tok := newToken([]byte(raw), nil)
	if err := tok.scan(); err != nil {
		t.Fatal(err)
	}
	return tok
}

func TestLiteralRestorationCommaBoundary(t *testing.T) {
	for _, tc := range []struct {
		name, literal, suffix string
		conjoined             bool
	}{
		{"outside", "`snow 雪`", ",", true},
		{"inside", "`snow 雪,`", "", false},
		{"both", "`snow 雪,`", ",", true},
		{"empty", "``", "", false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			prepared, err := PrepareSource("commas.l", []byte("Show "+tc.literal+tc.suffix+".\n"))
			if err != nil {
				t.Fatal(err)
			}
			words := newParser(statementBody(prepared.Statements[0].Prepared)).parseTokens()
			if len(words) != 2 {
				t.Fatalf("tokens = %q", words)
			}
			tok := literalRestorationToken(t, string(words[1]))
			id := prepared.Literals[0].ID
			if string(tok.rawBytes) != id || tok.conjoined != tc.conjoined {
				t.Fatalf("cached token = %q, conjoined = %t", tok.rawBytes, tok.conjoined)
			}
			statement := &lStatement{tokens: []*token{tok}, sGram: &statementGram{tokens: []*token{tok}}}
			if err = statement.uncacheLiterals(prepared.TextCache); err != nil {
				t.Fatal(err)
			}
			if string(tok.rawBytes) != tc.literal || tok.mem.Primitive.Text != tc.literal || tok.conjoined != tc.conjoined {
				t.Fatalf("restored token = %q, value = %q, conjoined = %t", tok.rawBytes, tok.mem.Primitive.Text, tok.conjoined)
			}
		})
	}
}

func TestLiteralRestorationPreservesTokenMetadata(t *testing.T) {
	prepared, statement := refactorPreparedStatement(t, "\"message\" is `first\r\nsecond \\` <DEF>`.\n")
	id := prepared.Literals[0].ID
	var tok *token
	for _, candidate := range statement.tokens {
		if string(candidate.rawBytes) == id {
			tok = candidate
		}
	}
	if tok == nil {
		t.Fatal("literal token missing")
	}
	tok.binding = &tokenBinding{Record: &record{ID: "metadata-owner"}, Root: "metadata-owner", POS: "noun", Form: "active", Possessive: true}
	tok.bindingError = errors.New("retained diagnostic")
	tok.mem.Runtime = &runtimeReference{OwnerID: "runtime-owner", Member: memberOutput, Source: statement}
	tok.category, tok.subCategory, tok.class = "retained category", "retained subcategory", "retained class"
	tok.abstract, tok.conjoined, tok.defined = true, true, true
	before, entryBefore := *tok, *tok.mem
	if err := statement.uncacheLiterals(prepared.TextCache); err != nil {
		t.Fatal(err)
	}
	literal := string(prepared.Literals[0].Raw)
	if string(tok.rawBytes) != literal || tok.mem.Primitive.Text != literal {
		t.Fatal("restoration changed source escaping or definition markers")
	}
	before.rawBytes, entryBefore.Primitive = tok.rawBytes, tok.mem.Primitive
	if !reflect.DeepEqual(*tok, before) || !reflect.DeepEqual(*tok.mem, entryBefore) || tok.mem != before.mem || tok.binding != before.binding {
		t.Fatal("literal restoration changed token, binding, or runtime metadata")
	}
}

func TestLiteralRestorationVisitsOriginalTokensOnce(t *testing.T) {
	const first = "`11111111-1111-1111-1111-111111111111`"
	const second = "`22222222-2222-2222-2222-222222222222`"
	tokens := []*token{literalRestorationToken(t, first), literalRestorationToken(t, second)}
	pool := map[string][]byte{first: []byte(second), second: []byte("`payload`")}
	statement := &lStatement{tokens: tokens, sGram: &statementGram{tokens: tokens, predicate: &predicateGram{tokens: tokens, obj: &objectGram{tokens: tokens, noun: tokens[0]}}}}
	if err := statement.uncacheLiterals(pool); err != nil {
		t.Fatal(err)
	}
	if string(tokens[0].rawBytes) != second || tokens[0].mem.Primitive.Text != second || string(tokens[1].rawBytes) != "`payload`" {
		t.Fatalf("literal bytes were restored recursively: %q / %q", tokens[0].rawBytes, tokens[1].rawBytes)
	}
	if string(pool[first]) != second || string(pool[second]) != "`payload`" {
		t.Fatal("literal pool changed")
	}
}

func TestLiteralRestorationFailureDoesNotMutateToken(t *testing.T) {
	const key = "`11111111-1111-1111-1111-111111111111`"
	for _, invalid := range []string{"not a primitive literal", "42", "true", `"identifier"`} {
		t.Run(invalid, func(t *testing.T) {
			tok := literalRestorationToken(t, key)
			before, primitiveBefore := *tok, *tok.mem.Primitive
			statement := &lStatement{tokens: []*token{tok}, sGram: &statementGram{tokens: []*token{tok}}}
			if err := statement.uncacheLiterals(map[string][]byte{key: []byte(invalid)}); err == nil {
				t.Fatal("invalid literal pool entry was accepted")
			}
			if !reflect.DeepEqual(*tok, before) || !reflect.DeepEqual(*tok.mem.Primitive, primitiveBefore) || !bytes.Equal(tok.rawBytes, []byte(key)) {
				t.Fatal("failed restoration changed the token")
			}
		})
	}
}

func TestLiteralRestorationRequiresCompleteKey(t *testing.T) {
	const key = "`11111111-1111-1111-1111-111111111111`"
	tok := literalRestorationToken(t, `"name`+key+`"`)
	before, primitiveBefore := *tok, *tok.mem.Primitive
	statement := &lStatement{tokens: []*token{tok}, sGram: &statementGram{tokens: []*token{tok}}}
	if err := statement.uncacheLiterals(map[string][]byte{key: []byte("`payload`")}); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(*tok, before) || !reflect.DeepEqual(*tok.mem.Primitive, primitiveBefore) {
		t.Fatal("a key-looking substring changed a declaration")
	}
}

func TestLiteralRestorationMissingKeyLeavesLiteralUnchanged(t *testing.T) {
	tok := literalRestorationToken(t, "`ordinary Text`")
	before, primitiveBefore := *tok, *tok.mem.Primitive
	statement := &lStatement{tokens: []*token{tok}, sGram: &statementGram{tokens: []*token{tok}}}
	for _, pool := range []map[string][]byte{nil, {}, {"`unrelated-key`": []byte("`other`")}} {
		if err := statement.uncacheLiterals(pool); err != nil {
			t.Fatal(err)
		}
		if !reflect.DeepEqual(*tok, before) || !reflect.DeepEqual(*tok.mem.Primitive, primitiveBefore) {
			t.Fatal("a cache miss changed the token")
		}
	}
}
