package Lyger

import (
	"reflect"
	"testing"
)

func extractionTestStatement(t *testing.T, source string) *lStatement {
	t.Helper()
	config := canonicalTestOptions()
	gram, err := loadGrammar(config)
	if err != nil {
		t.Fatal(err)
	}
	stmt, err := newLStatement([]byte(source), gram)
	if err != nil {
		t.Fatal(err)
	}
	if err = stmt.tokenize(); err != nil {
		t.Fatal(err)
	}
	stmt.determineType()
	if err = stmt.loadGram(); err != nil {
		t.Fatal(err)
	}
	return stmt
}

func TestRefactorSourceDeclarationExtraction(t *testing.T) {
	for _, tc := range []struct {
		source, id, path, kind string
		want                   any
	}{
		{"\"Textual\" is `雪\nraw <DEF> and \\\\ text`", "Textual", "primitive.text", "Text", "`雪\nraw <DEF> and \\\\ text`"},
		{`"Numeric" is 42.5`, "Numeric", "primitive.number", "Number", float64(42.5)},
		{`"Logical" is true`, "Logical", "primitive.truth", "Truth", true},
		{`Create "Library" as a Dictionary`, "Library", "type.iterative.dictionary", "Dictionary", nil},
		{`Create "Worker" as an App`, "Worker", "type.applicative.app", "App", nil},
	} {
		t.Run(tc.id, func(t *testing.T) {
			stmt := extractionTestStatement(t, tc.source)
			records, err := stmt.sGram.getDeclarations()
			if err != nil {
				t.Fatal(err)
			}
			if len(records) != 1 || records[0].Record.ID != tc.id || records[0].Record.Type != tc.path {
				t.Fatalf("declarations = %+v, want %s/%s", records, tc.id, tc.path)
			}
			value, err := records[0].Record.Value.value()
			if err != nil {
				t.Fatal(err)
			}
			if value.Type.String() != tc.kind || (tc.want != nil && !reflect.DeepEqual(value.Data, tc.want)) {
				t.Fatalf("value = %+v; want %s / %#v", value, tc.kind, tc.want)
			}
			if tc.kind == "App" && value.Data != records[0].Record.Value.Applicative.App {
				t.Fatal("App extraction changed instance identity")
			}
			if tc.kind == "Dictionary" && value.Data != records[0].Record.Value.Iterative.Dictionary {
				t.Fatal("Dictionary extraction changed instance identity")
			}
		})
	}
}

func TestRefactorBoundExtractionRetainsStorageIdentity(t *testing.T) {
	for _, original := range []*entry{
		{Type: storageApplicative, Applicative: &applicative{Type: valueApp, App: &App{}}},
		{Type: storageIterative, Iterative: &iterative{Type: valueDictionary, Dictionary: newDict()}},
	} {
		noun := &token{rawBytes: []byte("existing"), mem: original, defined: true}
		object := &objectGram{cType: "definitive", noun: noun}
		value, path, err := object.getValue(false)
		if err != nil {
			t.Fatal(err)
		}
		stored, err := value.value()
		if err != nil {
			t.Fatal(err)
		}
		prior, err := original.value()
		if err != nil {
			t.Fatal(err)
		}
		if value == nil || value == original || stored.Data != prior.Data {
			t.Fatal("bound extraction must copy the wrapper and preserve the stored instance")
		}
		want := "applicative.app"
		if original.Type == storageIterative {
			want = "iterative.dictionary"
		}
		if path != want {
			t.Fatalf("record path = %q, want %q", path, want)
		}
	}
}

func TestRefactorDeclarationExtractionRejectsMalformedStorage(t *testing.T) {
	for _, tc := range []struct {
		name  string
		value *entry
	}{
		{"missing App payload", &entry{Type: storageApplicative, Applicative: &applicative{Type: valueApp}}},
		{"wrong App payload", &entry{Type: storageApplicative, Applicative: &applicative{Type: valueApp, Service: &service{}}}},
		{"missing Dictionary payload", &entry{Type: storageIterative, Iterative: &iterative{Type: valueDictionary}}},
		{"conflicting outer wrappers", &entry{Type: storagePrimitive, Primitive: &primitive{Type: valueText, Text: "`valid`"}, Applicative: &applicative{Type: valueApp, App: &App{}}}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			stmt := extractionTestStatement(t, `"Target" is `+"`value`")
			object := stmt.sGram.predicate.obj
			object.cType = "definitive"
			object.noun.mem = tc.value
			if records, err := stmt.sGram.getDeclarations(); err == nil {
				t.Fatalf("malformed resolved storage was published: %+v", records)
			}
		})
	}
}

func TestRefactorDeclarationExtractionPropagatesActionFailure(t *testing.T) {
	subject := &subjectGram{
		cType: "directive", noun: &token{rawBytes: []byte(`"Action"`)},
		phrase: &phraseGram{cType: "definitive", sType: "phrase.directive", clauses: []*clauseGram{{}}},
	}
	if _, err := subject.getDeclaration(); err == nil || err.Error() != "directive action requires a bound lexeme" {
		t.Fatalf("action extraction lost its original failure: %v", err)
	}
}
