package Lyger

import (
	"errors"
	"fmt"
	"strings"
	"sync"
	"testing"

	"github.com/l-programming-language/l/lerrors"
)

func TestGrammarSyntaxCacheKeepsTermsIndependent(t *testing.T) {
	for _, raw := range []string{"[OBJECT.GENERIC]*", `"ID"`, "VERB.reporting", "ID**"} {
		want, wantErr := parseGrammarTerm(raw)
		for attempt := 0; attempt < 2; attempt++ {
			got, err := decodeGrammarTerm(raw)
			if got != want || (err != nil) != (wantErr != nil) || err != nil && err.Error() != wantErr.Error() {
				t.Fatalf("decode %q = %+v, %v; want %+v, %v", raw, got, err, want, wantErr)
			}
			got.Name, got.Optional = "changed use", !got.Optional
		}
	}
	for _, rule := range []string{"ID | VERB", "VERB | ID", "[ID]* | `~`", "ID | VERB"} {
		if got := strings.Join(splitGramRule(rule), " | "); got != rule {
			t.Fatalf("changed rule text reused another sequence: %q, want %q", got, rule)
		}
	}
}

func TestGrammarSyntaxCacheKeepsSequentialFailures(t *testing.T) {
	witness := errors.New("first association failed")
	for _, failFirst := range []bool{true, false, true} {
		state := 0
		err := applyGrammarRule(&state, "ID | [ID*]", func(value int) int { return value }, nil,
			func(value *int, _ grammarTerm) error {
				if failFirst {
					return witness
				}
				*value++
				return nil
			}, nil)
		if failFirst {
			if state != 0 || !errors.Is(err, witness) {
				t.Fatalf("later malformed term replaced the first failure: state=%d err=%v", state, err)
			}
		} else if state != 1 || !errors.Is(err, lerrors.InvalidConfiguration) || err.Error() != `malformed grammar reference "[ID*]"` {
			t.Fatalf("decoding changed the earlier committed term: state=%d err=%v", state, err)
		}
	}
}

func TestGrammarSyntaxCacheBoundsConcurrentRetention(t *testing.T) {
	var cache grammarSyntaxCache[grammarTerm]
	var workers sync.WaitGroup
	for worker := 0; worker < 4; worker++ {
		workers.Add(1)
		go func() {
			defer workers.Done()
			for index := 0; index < grammarSyntaxEntries*2; index++ {
				raw := fmt.Sprintf("object.cache_%d", index)
				cache.store(raw, grammarTerm{Raw: raw, Name: raw, Kind: termGrammar, Family: familyObject})
				if term, found := cache.lookup(raw); found && term.Raw != raw {
					t.Errorf("lookup %q returned another term: %+v", raw, term)
				}
			}
		}()
	}
	workers.Wait()
	oversized := strings.Repeat("x", grammarSyntaxBytes+1)
	cache.store(oversized, grammarTerm{Raw: oversized})
	if _, exists := cache.lookup(oversized); exists || len(cache.entries) != grammarSyntaxEntries {
		t.Fatalf("syntax retention exceeded its limits: oversized=%t entries=%d", exists, len(cache.entries))
	}
}
