package Lyger

import "testing"

func TestLexicalReservationsFollowPublicationAndNewDeclarations(t *testing.T) {
	parent := &lexicalScope{records: map[string]*record{}}
	scope := &lexicalScope{parent: parent, records: map[string]*record{}}
	declaration := &scopeDeclaration{Token: &token{rawBytes: []byte(`"Write|-s|--ten|--ing"`)}}
	scope.declarations = append(scope.declarations, declaration)
	for _, spelling := range []string{"Write", "Writes", "Written", "Writing", "WRITES’"} {
		binding, state, err := scope.lookupLocal([]byte(spelling))
		if err != nil || binding != nil || state != lookupPending {
			t.Fatalf("%q reservation = %v, %s, %v", spelling, binding, state, err)
		}
	}
	word := newLexeme("verb")
	if err := word.reLabel(`"Write|-s|--ten|--ing"`); err != nil {
		t.Fatal(err)
	}
	rec := &record{ID: "Write", Value: &entry{Type: storageLexeme, Lexeme: word}}
	declaration.Record, scope.records[rec.ID] = rec, rec
	if binding, state, err := scope.lookupLocal([]byte("Written")); err != nil || state != lookupFound || binding.Record != rec {
		t.Fatalf("published reservation = %v, %s, %v", binding, state, err)
	}
	scope.declarations = append(scope.declarations, &scopeDeclaration{Token: &token{rawBytes: []byte(`"Later"`)}})
	if binding, state, err := scope.lookupLocal([]byte("Later")); err != nil || binding != nil || state != lookupPending {
		t.Fatalf("appended reservation = %v, %s, %v", binding, state, err)
	}
}

func TestLexicalReservationsRetainOwnerNamespaces(t *testing.T) {
	scope := &lexicalScope{parent: &lexicalScope{}, records: map[string]*record{}, declarations: []*scopeDeclaration{
		{Token: &token{rawBytes: []byte(`"Retain|-s"`)}, Owner: &token{rawBytes: []byte("App's")}},
		{Token: &token{rawBytes: []byte(`"Private"`)}, Owner: &token{rawBytes: []byte("Library's")}},
	}}
	for _, spelling := range []string{"APP.Retain", "App.Retains’"} {
		if binding, state, err := scope.lookupLocal([]byte(spelling)); err != nil || binding != nil || state != lookupPending {
			t.Fatalf("qualified native reservation %q = %v, %s, %v", spelling, binding, state, err)
		}
	}
	for _, spelling := range []string{"Retain", "Private", "Library.Private", "Other.Retain"} {
		if binding, state, err := scope.lookupLocal([]byte(spelling)); err != nil || binding != nil || state != lookupAbsent {
			t.Fatalf("unrelated namespace %q = %v, %s, %v", spelling, binding, state, err)
		}
	}
}
