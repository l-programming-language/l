package Lyger

import (
	"strings"
	"testing"
)

func TestBindingPreservesStorageAndInstanceIdentity(t *testing.T) {
	text := &entry{Type: storagePrimitive, Primitive: &primitive{Type: valueText, Text: "`snow 雪`"}}
	app := &App{ID: "sample"}
	dict := newDict()
	for _, original := range []*entry{text, {Type: storageApplicative, Applicative: &applicative{Type: valueApp, App: app}}, {Type: storageIterative, Iterative: &iterative{Type: valueDictionary, Dictionary: dict}}} {
		rec := &record{ID: "named", Value: original}
		token := &token{rawBytes: []byte("named"), scope: &lexicalScope{records: map[string]*record{"named": rec}}}
		if state, err := (semanticResolver{}).bindToken(token); err != nil || state != lookupFound {
			t.Fatal("reference was not bound")
		}
		got, err := token.mem.value()
		if err != nil {
			t.Fatal(err)
		}
		want, err := original.value()
		if err != nil {
			t.Fatal(err)
		}
		if token.mem.Type != original.Type || got.Type != want.Type {
			t.Fatalf("storage/leaf type changed: %#v", token.mem)
		}
		if got.Data != want.Data {
			t.Fatal("binding changed scalar or instance identity")
		}
	}
}

func TestBindingPreservesMorphologyAndDefinition(t *testing.T) {
	for _, tc := range []struct{ pos, label, use, want, form string }{
		{"noun", `"Process|-es|-ive|-ing"`, "Processing", "verbs.generic.applicative", "active"},
		{"verb", `"Log|-s|-ged|-ging|-ger"`, "Logger", "nouns.generic.root", "objective"},
		{"verb", `"Print|-s|-ed|-ing|-er"`, "Prints", "verbs.generic.active", "active"},
		{"noun", `"Plan|-s"`, "Plans’", "nouns.generic.plural.possessive", "plural"},
	} {
		t.Run(tc.use, func(t *testing.T) {
			definition := newLexeme(tc.pos)
			if err := definition.reLabel(tc.label); err != nil {
				t.Fatal(err)
			}
			definition.Inputs = []string{"original"}
			definition.Actions = []*lexeme{{Root: "inner", Inputs: []string{"parameter"}}}
			rec := &record{ID: definition.Root, Value: &entry{Type: storageLexeme, Lexeme: definition}}
			token := &token{rawBytes: []byte(tc.use), scope: &lexicalScope{records: map[string]*record{rec.ID: rec}}}
			if state, err := (semanticResolver{}).bindToken(token); err != nil || state != lookupFound {
				t.Fatal("unbound")
			}
			if token.tType != tc.want || token.binding.Form != tc.form || token.mem.Lexeme.POS != tc.pos {
				t.Fatalf("lost lexical identity: %#v / %s", token.binding, token.tType)
			}
			token.mem.Lexeme.Inputs[0] = "changed"
			token.mem.Lexeme.Actions[0].Inputs[0] = "changed"
			if definition.Inputs[0] != "original" || definition.Actions[0].Inputs[0] != "parameter" {
				t.Fatal("invocation mutated definition")
			}
		})
	}
}

func TestBindingQualifiedMembersAndAmbiguity(t *testing.T) {
	l := newLexeme("verb")
	if err := l.reLabel(`"Write|-s"`); err != nil {
		t.Fatal(err)
	}
	member := &record{ID: "Library.Write", Value: &entry{Type: storageLexeme, Lexeme: l}}
	if b, err := resolveTokenBinding([]byte("Writes"), map[string]*record{"member": member}); b != nil || err != nil {
		t.Fatal("qualified member leaked into global scope")
	}
	if b, err := resolveTokenBinding([]byte("Library.Writes"), map[string]*record{"member": member}); b == nil || err != nil {
		t.Fatal("qualified member missing", err)
	}
	one := &record{ID: "One", Value: &entry{Type: storageLexeme, Lexeme: l}}
	two := &record{ID: "Two", Value: &entry{Type: storageLexeme, Lexeme: l}}
	for i := 0; i < 20; i++ {
		_, err := resolveTokenBinding([]byte("Writes"), map[string]*record{"z": two, "a": one})
		if err == nil || !strings.Contains(err.Error(), "One and Two") {
			t.Fatalf("ambiguous binding must report stable alternatives: %v", err)
		}
	}
}
