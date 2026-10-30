package Lyger

import (
	"reflect"
	"testing"
)

func indexedTestWord(t *testing.T, name, label string, callable bool) *record {
	t.Helper()
	word := newLexeme("verb")
	if err := word.reLabel(label); err != nil {
		t.Fatal(err)
	}
	value := &entry{Type: storageLexeme, Lexeme: word}
	if callable {
		value = initEntry("Function", storageApplicative)
		value.Applicative.Function.Source = &sourceFunction{Word: word}
	}
	return &record{ID: name, Value: value}
}

func TestBindingIndexPreservesResolution(t *testing.T) {
	write := indexedTestWord(t, "Library.Write", `"Write|-s|--ote|--iting|-r"`, false)
	write.Value.Lexeme.Prefixes["un-"] = "negative"
	call := indexedTestWord(t, "Apply", `"Apply|--ies|--ied|--ying|-er"`, true)
	value := &record{ID: "Mixed", Value: initEntry("Text", storagePrimitive)}
	records := map[string]*record{"write": write, "call": call, "value": value, "duplicate": write, "nil": nil}
	scope := &lexicalScope{records: records, bindings: newRecordBindingIndex(records)}
	for _, spelling := range []string{"Library.Write", "LIBRARY.WRITES", "Library.Wrote's", "Library.Writing", "Library.Unwrite", "Write", "Apply", "applies", "APPLIED’S", "Applying", "Mixed's", "missing"} {
		want, wantErr := resolveTokenBinding([]byte(spelling), records)
		got, _, gotErr := scope.lookupLocal([]byte(spelling))
		if !reflect.DeepEqual(got, want) || (gotErr != nil) != (wantErr != nil) {
			t.Fatalf("%q indexed lookup = %+v, %v; direct = %+v, %v", spelling, got, gotErr, want, wantErr)
		}
	}
	other := indexedTestWord(t, "Library.Other", `"Write|-s|--ote|--iting|-r"`, false)
	scope.storeBinding("another", other)
	_, _, gotErr := scope.lookupLocal([]byte("Library.Writes"))
	_, wantErr := resolveTokenBinding([]byte("Library.Writes"), records)
	if gotErr == nil || wantErr == nil || gotErr.Error() != wantErr.Error() {
		t.Fatalf("indexed ambiguity = %v; direct = %v", gotErr, wantErr)
	}
}

func TestBindingIndexReplacementAndLifetime(t *testing.T) {
	before := indexedTestWord(t, "Before", `"Before|-s|-d|-ing|-r"`, false)
	after := indexedTestWord(t, "After", `"After|-s|-ed|-ing|-er"`, false)
	scope := &lexicalScope{records: map[string]*record{"same": before}}
	plan := &scopePlan{Root: scope}
	release := plan.indexBindings()
	scope.storeBinding("same", after)
	if got, state, err := scope.lookupLocal([]byte("Befores")); got != nil || state != lookupAbsent || err != nil {
		t.Fatalf("replaced spelling remained visible: %+v %v", got, err)
	}
	if got, state, err := scope.lookupLocal([]byte("Afters")); got == nil || got.Record != after || state != lookupFound || err != nil {
		t.Fatalf("replacement was not indexed: %+v %v", got, err)
	}
	release()
	if scope.bindings != nil {
		t.Fatal("semantic lookup index survived its lifetime")
	}
	// Memory remains an ordinary mutable map after semantic analysis returns.
	scope.records["same"] = before
	if got, state, err := scope.lookupLocal([]byte("Befores")); got == nil || got.Record != before || state != lookupFound || err != nil {
		t.Fatalf("post-analysis mutation used a stale index: %+v %v", got, err)
	}
}
