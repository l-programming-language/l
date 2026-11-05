package Lyger

import "testing"

func TestWordLibraryQualifiedDescriptorReference(t *testing.T) {
	c := compileMemberResolutionSource(t, memberResolutionSource{
		main:    "Use Catalog's Decorated from `local-words.l` as \"Selected\".\n",
		library: "Create \"Catalog\" as a Dictionary.\nPublicly define Catalog's \"Decorated\" as a New Adjective describing any Process.\n",
	})
	if c.Error != nil {
		t.Fatal(c.Error)
	}
	record := c.Memory["Selected"]
	canonical, err := canonicalRecord(record, c.Semantics.Aliases)
	if err != nil || canonical == nil || c.Semantics.Descriptors[canonical] == nil {
		t.Fatalf("qualified descriptor alias lost its identity: %v", err)
	}
}

func TestWordLibraryModifierHomonym(t *testing.T) {
	verb := indexedTestWord(t, "Fold", `"Fold|-s|-ed|-ing|-er"`, false)
	adjective := indexedTestWord(t, "Folded", `"Folded"`, false)
	adjective.Value.Lexeme.POS = "adjective"
	nominal := indexedTestWord(t, "Parcel", `"Parcel"`, false)
	nominal.Value.Lexeme.POS = "noun"
	scope := &lexicalScope{records: map[string]*record{"Fold": verb, "Folded": adjective, "Parcel": nominal}}
	newToken := &token{rawBytes: []byte("New"), tType: "adjectives.initialization.root", mem: &entry{Type: storageLexeme, Lexeme: &lexeme{POS: "adjective", Type: "initialization"}}, scope: scope}
	modifier := &token{rawBytes: []byte("Folded"), tType: "primitive.unknown", mem: &entry{Type: storagePrimitive, Primitive: newPrimitive(valueUnknown)}, scope: scope}
	parcel := &token{rawBytes: []byte("Parcel"), tType: "nouns.generic.root", mem: nominal.Value, scope: scope}
	statement := &lStatement{tokens: []*token{newToken, modifier, parcel}}
	resolver := semanticResolver{scopes: &scopePlan{Root: scope}, semantics: newSemanticProgram()}
	if _, err := resolver.bindStatementMembers(statement); err != nil || modifier.binding == nil || modifier.binding.Record != adjective {
		t.Fatalf("constructor modifier selected the wrong homonym: %+v, %v", modifier.binding, err)
	}
	if binding, err := resolveTokenBinding([]byte("Folded"), map[string]*record{"Fold": verb}); err != nil || binding == nil || binding.Record != verb || binding.Form != "passive" {
		t.Fatalf("passive family was changed: %+v, %v", binding, err)
	}
	auxiliary := &token{rawBytes: []byte("is"), tType: "verbs.action.passive.root", mem: &entry{Type: storageLexeme, Lexeme: &lexeme{POS: "verb", Type: "action.passive"}}, scope: scope}
	passive := &token{rawBytes: []byte("Folded"), tType: "primitive.unknown", mem: &entry{Type: storagePrimitive, Primitive: newPrimitive(valueUnknown)}, scope: scope}
	if _, err := resolver.bindStatementMembers(&lStatement{tokens: []*token{parcel, auxiliary, passive}}); err != nil || passive.binding == nil || passive.binding.Record != verb {
		t.Fatalf("passive position selected the adjective: %+v, %v", passive.binding, err)
	}
	other := indexedTestWord(t, "Other", `"Folded"`, false)
	other.Value.Lexeme.POS = "adjective"
	scope.records["Other"] = other
	modifier.binding, modifier.tType = nil, "primitive.unknown"
	if _, err := resolver.bindStatementMembers(statement); err == nil {
		t.Fatal("same-POS ambiguity was silently resolved")
	}
}
