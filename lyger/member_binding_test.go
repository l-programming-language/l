package Lyger

import (
	"bytes"
	"testing"
)

func memberBindingStatement(t *testing.T) *lStatement {
	t.Helper()
	config := canonicalTestOptions()
	gram, err := loadGrammar(config)
	if err != nil {
		t.Fatal(err)
	}
	statement, err := newLStatement([]byte(`Use Library's Render as "Alias"`), gram)
	if err != nil {
		t.Fatal(err)
	}
	if err := statement.tokenize(); err != nil {
		t.Fatal(err)
	}
	return statement
}

func memberBindingRecord(t *testing.T, id string) *record {
	t.Helper()
	definition := newLexeme("verb")
	if err := definition.reLabel(`"Render|-s|-ed|-ing|-er"`); err != nil {
		t.Fatal(err)
	}
	definition.Inputs = []string{"original"}
	return &record{ID: id, Type: "lexeme.verb.generic", Value: &entry{Type: storageLexeme, Lexeme: definition}}
}

func TestMemberBindingMissingQualifiedNameDoesNotUseGlobal(t *testing.T) {
	for _, previouslyBound := range []bool{false, true} {
		name := "unbound member"
		if previouslyBound {
			name = "stale global binding"
		}
		t.Run(name, func(t *testing.T) {
			statement := memberBindingStatement(t)
			owner, member := statement.tokens[1], statement.tokens[2]
			library := &record{ID: "Library", Value: initEntry("Dictionary", storageIterative)}
			global := memberBindingRecord(t, "Render")
			memory := map[string]*record{"Library": library, "Render": global}
			scope := &lexicalScope{records: memory}
			for _, tok := range statement.tokens {
				tok.scope = scope
			}
			if previouslyBound {
				if state, err := (semanticResolver{}).bindToken(member); err != nil || state != lookupFound {
					t.Fatal("could not arrange the stale global reference")
				}
			}
			if _, err := (semanticResolver{semantics: newSemanticProgram()}).bindStatementMembers(statement); err != nil {
				t.Fatal(err)
			}
			if owner.binding == nil || owner.binding.Record != library {
				t.Fatal("member owner did not resolve")
			}
			if member.binding != nil || member.defined || member.bindingError != nil || member.getPOS(false) != "unknown" {
				t.Fatalf("missing qualified member retained a global reference: %+v", member)
			}
			if member.mem.Lexeme != nil {
				t.Fatal("missing qualified member retained the global definition")
			}
			statement.determineType()
			if statement.sType != "unknown" {
				t.Fatal("unresolved member did not remain eligible for dependency retry")
			}

			qualified := memberBindingRecord(t, "Library.Render")
			memory[qualified.ID] = qualified
			memory["Library"].members = map[string]*record{qualified.ID: qualified}
			if _, err := (semanticResolver{semantics: newSemanticProgram()}).bindStatementMembers(statement); err != nil {
				t.Fatal(err)
			}
			if member.binding == nil || member.binding.Record != qualified || member.bindingError != nil {
				t.Fatal("newly available qualified definition did not resolve on retry")
			}
			if !bytes.Equal(member.rawBytes, []byte("Render")) || !bytes.Equal(owner.rawBytes, []byte("Library's")) {
				t.Fatal("namespace resolution altered source token bytes")
			}
		})
	}
}

func TestMemberBindingQualifiedNameClearsUnqualifiedAmbiguity(t *testing.T) {
	statement := memberBindingStatement(t)
	member := statement.tokens[2]
	global := memberBindingRecord(t, "Render")
	memory := map[string]*record{
		"Library": {ID: "Library", Value: initEntry("Dictionary", storageIterative)},
		"Render":  global,
	}
	scope := &lexicalScope{records: memory}
	for _, tok := range statement.tokens {
		tok.scope = scope
	}
	if state, err := (semanticResolver{}).bindToken(member); err != nil || state != lookupFound {
		t.Fatal("could not arrange the stale global reference")
	}
	memory["render"] = memberBindingRecord(t, "render")
	if state, err := (semanticResolver{}).bindToken(member); state != lookupAbsent || err == nil || member.bindingError == nil {
		t.Fatal("unqualified lookup did not expose the expected ambiguity")
	}
	qualified := memberBindingRecord(t, "Library.Render")
	memory[qualified.ID] = qualified
	memory["Library"].members = map[string]*record{qualified.ID: qualified}
	if _, err := (semanticResolver{semantics: newSemanticProgram()}).bindStatementMembers(statement); err != nil {
		t.Fatal(err)
	}
	if member.binding == nil || member.binding.Record != qualified || member.bindingError != nil || !member.defined {
		t.Fatalf("qualified resolution retained stale state: %+v", member)
	}
	if member.tType != "verbs.generic.root" || member.mem.Lexeme == qualified.Value.Lexeme {
		t.Fatal("qualified resolution lost lexical identity or reused the stored definition")
	}
	member.mem.Lexeme.Inputs[0] = "changed"
	if qualified.Value.Lexeme.Inputs[0] != "original" || global.Value.Lexeme.Inputs[0] != "original" {
		t.Fatal("qualified invocation mutated a stored definition")
	}
}
