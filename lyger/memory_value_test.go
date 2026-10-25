package Lyger

import (
	"strings"
	"testing"
)

func TestMemoryPublicationChecksStoredValues(t *testing.T) {
	for _, qualified := range []bool{false, true} {
		compilation := &Compilation{Memory: make(map[string]*record), Semantics: newSemanticProgram()}
		id := "Broken"
		if qualified {
			id = "Library.Broken"
			compilation.Memory["Library"] = &record{ID: "Library", Value: initEntry("Dictionary", storageIterative)}
		}
		rec := &record{ID: id, Value: &entry{Type: storageApplicative, Applicative: &applicative{Type: valueApp}}}
		if err := publishTestRecord(t, compilation, rec); err == nil {
			t.Error("malformed record accepted")
		}
		if compilation.Memory[id] != nil {
			t.Error("failed read published a record")
		}
	}
}

func TestMemoryPublicationChecksDictionaryOwner(t *testing.T) {
	dictionary := newDict()
	owner := &record{ID: "Library", Value: &entry{
		Type: storageIterative, Iterative: &iterative{Type: valueDictionary, Dictionary: dictionary},
		Primitive: &primitive{Type: valueText},
	}}
	compilation := &Compilation{Memory: map[string]*record{"Library": owner}, Semantics: newSemanticProgram()}
	rec := &record{ID: "Library.Member", Value: &entry{Type: storagePrimitive, Primitive: &primitive{Type: valueText, Text: "`retained`"}}}
	if err := publishTestRecord(t, compilation, rec); err == nil {
		t.Fatal("malformed namespace accepted")
	}
	if len(dictionary.Values) != 0 || len(dictionary.Types) != 0 || compilation.Memory[rec.ID] != nil {
		t.Fatal("failed owner read published member data")
	}
}

// Prepare source bindings for the same malformed publication values.
func publishTestRecord(t *testing.T, compilation *Compilation, rec *record) error {
	t.Helper()
	source := `Create "` + rec.ID + `" as a Text.`
	if dot := strings.LastIndexByte(rec.ID, '.'); dot >= 0 {
		source = `Create ` + rec.ID[:dot] + `'s "` + rec.ID[dot+1:] + `" as a Text.`
	}
	statements := scopeTestStatements(t, "memory.l", source, scopeTestGrammar(t))
	var err error
	compilation.Scope, err = planSourceScopes(statements, compilation.Memory, nil)
	if err != nil {
		t.Fatal(err)
	}
	return compilation.publishRegistration(statements[0], statementRegistration{declarations: []extractedDeclaration{{Token: compilation.Scope.Ordered[0].Declarations[0].Token, Record: rec}}})
}
