package Lyger

import (
	"errors"
	"reflect"
	"testing"
)

func TestGrammarRuleNestedOptionalRollback(t *testing.T) {
	state := []string{"original"}
	clone := func(value []string) []string { return append([]string(nil), value...) }
	rejected := errors.New("rejected association")
	err := applyGrammarRule(&state, "ID | VERB* | ARTICLE", clone, nil,
		func(value *[]string, term grammarTerm) error {
			switch term.POS {
			case posInvalid: // ID invokes a nested rule on the current candidate.
				return applyGrammarRule(value, "ARTICLE | VERB*", clone, nil,
					func(nested *[]string, nestedTerm grammarTerm) error {
						if nestedTerm.Optional {
							(*nested)[0] = "rejected nested change"
							return rejected
						}
						*nested = append(*nested, "nested")
						return nil
					}, nil)
			case posVerb:
				(*value)[0] = "rejected outer change"
				*value = append(*value, "rejected")
				return rejected
			case posArticle:
				*value = append(*value, "outer")
			default:
				t.Fatalf("unexpected term: %+v", term)
			}
			return nil
		}, nil)
	if want := []string{"original", "nested", "outer"}; err != nil || !reflect.DeepEqual(state, want) {
		t.Fatalf("nested rule or rejected optional association changed committed state: %v, %v", state, err)
	}
}

func TestGrammarCasesKeepSelectedAssociations(t *testing.T) {
	type state struct {
		associations []string
		caseName     string
	}
	selected := state{associations: []string{"original"}}
	tokens := []*token{{rawBytes: []byte("first")}, {rawBytes: []byte("second")}}
	err := selectGrammarCase(&selected, map[string]string{
		"a_full": "full", "b_short": "short", "c_failed": "failed", "d_tie": "full",
	}, func(value state) state {
		value.associations = append([]string(nil), value.associations...)
		return value
	}, func(value *state, rule string) error {
		if !reflect.DeepEqual(value.associations, []string{"original"}) || value.caseName != "" {
			t.Fatalf("alternative started with another attempt's state: %+v", value)
		}
		value.associations[0] = rule
		value.associations = append(value.associations, "first")
		if rule == "failed" {
			return errors.New("rejected alternative")
		}
		if rule == "full" {
			value.associations = append(value.associations, "second")
		}
		return nil
	}, func(value *state) []*token {
		return tokens[len(value.associations)-1:]
	}, func(value *state, name string) {
		value.caseName = name
	}, tokens)
	want := state{associations: []string{"full", "first", "second"}, caseName: "a_full"}
	if err != nil || !reflect.DeepEqual(selected, want) {
		t.Fatalf("later alternatives changed selected case: %+v, %v", selected, err)
	}
}
