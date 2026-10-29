package Lyger

import (
	"reflect"
	"testing"
)

func TestGrammarCompletePrefixEndsAlternativeSearch(t *testing.T) {
	type state struct {
		consumed int
		selected string
	}
	var target state
	var tried []string
	tokens := []*token{{rawBytes: []byte("first")}, {rawBytes: []byte("second")}}
	err := selectGrammarCase(&target, map[string]string{
		"a_short": "short", "b_complete": "complete", "c_tie": "tie", "d_later": "later",
	}, func(value state) state { return value }, func(value *state, rule string) error {
		tried = append(tried, rule)
		if rule == "short" {
			value.consumed = 1
		} else {
			value.consumed = len(tokens)
		}
		return nil
	}, func(value *state) []*token {
		return tokens[value.consumed:]
	}, func(value *state, name string) {
		value.selected = name
	}, tokens)
	if err != nil || target.consumed != len(tokens) || target.selected != "b_complete" {
		t.Fatalf("complete match lost longest-prefix or sorted tie precedence: %+v, %v", target, err)
	}
	if want := []string{"short", "complete"}; !reflect.DeepEqual(tried, want) {
		t.Fatalf("searched alternatives after the result was determined: %v; want %v", tried, want)
	}
}
