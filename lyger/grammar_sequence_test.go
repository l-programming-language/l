package Lyger

import (
	"errors"
	"reflect"
	"testing"
)

func TestGrammarSequenceConsumptionBoundaries(t *testing.T) {
	first := &token{rawBytes: []byte("first")}
	second := &token{rawBytes: []byte("second")}
	separator := &token{rawBytes: []byte("and"), tType: "conjunctions.cumulative.root"}
	suffix := &token{rawBytes: []byte("suffix")}
	conjoined := &token{rawBytes: []byte("first,"), conjoined: true}
	parseFailure := errors.New("original parse failure")
	for _, tc := range []struct {
		name       string
		tokens     []*token
		multiple   bool
		invalid    *token
		consumed   int
		failure    *token
		want       []*token
		separators []*token
		errorText  string
	}{
		{name: "single leaves separator", tokens: []*token{first, separator, second}, want: []*token{first}},
		{name: "parent suffix", tokens: []*token{first, suffix}, multiple: true, want: []*token{first}},
		{name: "trailing separator", tokens: []*token{first, separator}, multiple: true, want: []*token{first}},
		{name: "failed lookahead retains separator", tokens: []*token{first, separator, second}, multiple: true, failure: second, want: []*token{first}},
		{name: "failed conjoined lookahead", tokens: []*token{conjoined, second}, multiple: true, failure: second, want: []*token{conjoined}},
		{name: "first parse failure", tokens: []*token{first}, failure: first, errorText: parseFailure.Error()},
		{name: "zero initial consumption", tokens: []*token{first}, invalid: first, errorText: "grammar sequence made no progress"},
		{name: "oversized initial consumption", tokens: []*token{first}, invalid: first, consumed: 2, errorText: "grammar sequence made no progress"},
		{name: "zero lookahead consumption", tokens: []*token{first, separator, second}, multiple: true, invalid: second, errorText: "grammar sequence made no progress"},
		{name: "oversized lookahead consumption", tokens: []*token{first, separator, second}, multiple: true, invalid: second, consumed: 2, errorText: "grammar sequence made no progress"},
		{name: "explicit separator and suffix", tokens: []*token{first, separator, second, suffix}, multiple: true, want: []*token{first, second}, separators: []*token{separator}},
		{name: "implicit comma separator", tokens: []*token{conjoined, second, suffix}, multiple: true, want: []*token{conjoined, second}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			original := append([]*token(nil), tc.tokens...)
			values, separators, err := grammarSequence(tc.tokens, tc.multiple, func(input []*token) (*token, []*token, error) {
				if input[0] == tc.failure {
					return nil, nil, parseFailure
				}
				if input[0] == tc.invalid {
					return input[0], make([]*token, tc.consumed), nil
				}
				return input[0], input[:1], nil
			})
			if tc.errorText != "" {
				if err == nil || err.Error() != tc.errorText || values != nil || separators != nil {
					t.Fatalf("failed parse = (%v, %v, %v), want nil results and %q", values, separators, err, tc.errorText)
				}
				if tc.failure == first && !errors.Is(err, parseFailure) {
					t.Fatal("first parse error identity changed")
				}
			} else if err != nil || !reflect.DeepEqual(values, tc.want) || !reflect.DeepEqual(separators, tc.separators) {
				t.Fatalf("sequence = (%v, %v, %v), want (%v, %v, nil)", values, separators, err, tc.want, tc.separators)
			}
			if !reflect.DeepEqual(tc.tokens, original) {
				t.Fatal("sequence changed the caller's token order")
			}
		})
	}
}

func TestGrammarSequenceRetainsLongestPrefixAndSortedTies(t *testing.T) {
	gram, err := loadGrammar(canonicalTestOptions())
	if err != nil {
		t.Fatal(err)
	}
	gram.grammarMap["object"] = grammeme{Structure: "CASE", Cases: map[string]string{
		"a_short": "ARTICLE", "b_full": "ARTICLE | #~#", "z_full": "ARTICLE | #~#",
	}}
	tokens := grammarTestTokens(t, "any Text and any Text with", gram)
	term, err := decodeGrammarTerm("[object]")
	if err != nil {
		t.Fatal(err)
	}
	objects, separators, err := buildGrammarObjects(term, tokens, gram)
	if err != nil {
		t.Fatal(err)
	}
	if len(objects) != 2 || !reflect.DeepEqual(separators, tokens[2:3]) {
		t.Fatalf("sequence = %d objects and %v separators", len(objects), separators)
	}
	for index, object := range objects {
		start := index * 3
		if object.cType != "b_full" || !reflect.DeepEqual(object.tokens, tokens[start:start+2]) {
			t.Fatalf("object %d lost longest-prefix selection or deterministic tie: %+v", index, object)
		}
	}
	if string(tokens[5].rawBytes) != "with" {
		t.Fatal("parent suffix changed")
	}
}

func TestGrammarSequenceParsesEachSuccessfulItemOnce(t *testing.T) {
	first, second, third := &token{rawBytes: []byte("first")}, &token{rawBytes: []byte("second")}, &token{rawBytes: []byte("third")}
	and := &token{rawBytes: []byte("and"), tType: "conjunctions.cumulative.root"}
	or := &token{rawBytes: []byte("or"), tType: "conjunctions.alternative.root"}
	calls := make(map[*token]int)
	parsed := make(map[*token]*struct{ token *token })
	values, separators, err := grammarSequence([]*token{first, and, second, or, third}, true, func(input []*token) (*struct{ token *token }, []*token, error) {
		calls[input[0]]++
		value := &struct{ token *token }{input[0]}
		if parsed[input[0]] == nil {
			parsed[input[0]] = value
		}
		return value, input[:1], nil
	})
	if err != nil || len(values) != 3 || !reflect.DeepEqual(separators, []*token{and, or}) {
		t.Fatalf("sequence = (%v, %v, %v)", values, separators, err)
	}
	for index, tok := range []*token{first, second, third} {
		if calls[tok] != 1 || values[index] != parsed[tok] {
			t.Errorf("item %q parsed %d times; emitted result is original = %t", tok.rawBytes, calls[tok], values[index] == parsed[tok])
		}
	}
}
