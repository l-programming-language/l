package Lyger

import (
	"reflect"
	"strings"
	"testing"
)

func TestRefactorSourceTypeWhitelist(t *testing.T) {
	for _, tc := range []struct{ source, value, storage string }{
		{"Null", "Null", "Primitive"},
		{"Text", "Text", "Primitive"},
		{"Number", "Number", "Primitive"},
		{"Truth", "Truth", "Primitive"},
		{"Bytes", "Bytes", "Primitive"},
		{"Thing", "Thing", "Primitive"},
		{"App", "App", "Applicative"},
		{"Service", "Service", "Applicative"},
		{"Class", "Class", "Applicative"},
		{"Object", "Object", "Applicative"},
		{"Attribute", "Attribute", "Applicative"},
		{"Definition", "Definition", "Applicative"},
		{"List", "List", "Iterative"},
		{"Dictionary", "Dictionary", "Iterative"},
		{"Noun", "Noun", "Lexeme"},
		{"Verb", "Verb", "Lexeme"},
		{"Adverb", "Adverb", "Lexeme"},
		{"Adjective", "Adjective", "Lexeme"},
		{"tExT", "Text", "Primitive"},
		{"dIcTiOnArY", "Dictionary", "Iterative"},
		{"Pointer", "Pointer", "Primitive"},
		{"Map", "Map", "Iterative"},
		{"Instance", "Instance", "Applicative"},
		{"Function", "Unknown", "Unknown"},
		{"Property", "Unknown", "Unknown"},
		{"Operation", "Unknown", "Unknown"},
		{"ID", "Unknown", "Unknown"},
		{"Byte", "Unknown", "Unknown"},
		{"Error", "Unknown", "Unknown"},
		{"Unknown", "Unknown", "Unknown"},
		{"None", "Unknown", "Unknown"},
		{" Text", "Unknown", "Unknown"},
		{"Text ", "Unknown", "Unknown"},
		{"", "Unknown", "Unknown"},
	} {
		t.Run(tc.source, func(t *testing.T) {
			value, storage := checkType([]byte(tc.source))
			if value != tc.value || storage.String() != tc.storage {
				t.Fatalf("checkType(%q) = (%q, %q), want (%q, %q)", tc.source, value, storage, tc.value, tc.storage)
			}
		})
	}
}

func TestRefactorTransientStorageTags(t *testing.T) {
	var zero entry
	if zero.Type.String() != "" || zero.Primitive != nil || zero.Lexeme != nil || zero.Applicative != nil || zero.Iterative != nil || zero.Runtime != nil {
		t.Fatal("zero entry must remain an uninitialized state")
	}
	for _, name := range []string{"", "Unknown", "None", "ID"} {
		kind, ok := parseValueType(name)
		if !ok {
			t.Fatalf("unknown characterization type %q", name)
		}
		if got := newPrimitive(kind); got.Type.String() != "Unknown" {
			t.Errorf("newPrimitive(%q).Type = %q, want Unknown", name, got.Type)
		}
	}
	null := newPrimitive(valueNull)
	value, err := null.value()
	if err != nil || null.Type != valueNull || value.Type != valueNull || value.Data != nil {
		t.Fatalf("explicit Null must be a readable initialized value: %+v, %v", value, err)
	}
	declared, err := loadPrimitive([]byte(`"Subject"`))
	if err != nil || declared.Type.String() != "ID" || declared.ID != `"subject"` {
		t.Fatalf("declaration token = %+v, %v", declared, err)
	}
	unknown, err := loadPrimitive([]byte("unresolved_reference"))
	if err == nil || unknown.Type.String() != "None" {
		t.Fatalf("unrecognized literal = %+v, %v; want None and an error", unknown, err)
	}
	bytesValue := newPrimitive(valueBytes)
	if bytesValue.Bytes == nil || len(bytesValue.Bytes) != 0 {
		t.Fatal("new Bytes storage must retain its nonnil empty payload")
	}
}

func TestRefactorValidStoredValueProjections(t *testing.T) {
	thing := &struct{ name string }{"opaque identity"}
	for _, tc := range []struct {
		name  string
		value *primitive
		want  interface{}
	}{
		{"Thing", &primitive{Type: valueThing, Thing: thing}, thing},
		{"nil Thing", &primitive{Type: valueThing}, nil},
		{"Text", &primitive{Type: valueText, Text: "`雪\nraw`"}, "`雪\nraw`"},
		{"Number", &primitive{Type: valueNumber, Number: 2.5}, float64(2.5)},
		{"Truth", &primitive{Type: valueTruth, Truth: true}, true},
		{"Bytes", &primitive{Type: valueBytes, Bytes: []byte{0, 255}}, []byte{0, 255}},
		{"nil Bytes", &primitive{Type: valueBytes}, []byte(nil)},
		{"Pointer", &primitive{Type: valuePointer, Pointer: "source-reference"}, "source-reference"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			stored := &entry{Type: storagePrimitive, Primitive: tc.value}
			got, err := stored.value()
			if err != nil {
				t.Fatal(err)
			}
			if got.Type.String() != strings.TrimPrefix(tc.name, "nil ") || !reflect.DeepEqual(got.Data, tc.want) || stored.Primitive != tc.value {
				t.Fatalf("stored projection = %+v; want %s / %#v without replacing storage", got, tc.name, tc.want)
			}
		})
	}
	app, serviceValue, classValue := &App{}, &service{}, &class{}
	objectValue, attributeValue, definitionValue := &object{}, &attribute{}, &def{}
	functionValue, propertyValue, operationValue := &funct{}, &prop{}, &op{}
	listValue, dictionaryValue, lexemeValue := newList(), newDict(), newLexeme("noun")
	for _, tc := range []struct {
		name  string
		entry *entry
		want  interface{}
	}{
		{"App", &entry{Type: storageApplicative, Applicative: &applicative{Type: valueApp, App: app}}, app},
		{"Service", &entry{Type: storageApplicative, Applicative: &applicative{Type: valueService, Service: serviceValue}}, serviceValue},
		{"Class", &entry{Type: storageApplicative, Applicative: &applicative{Type: valueClass, Class: classValue}}, classValue},
		{"Object", &entry{Type: storageApplicative, Applicative: &applicative{Type: valueObject, Object: objectValue}}, objectValue},
		{"Attribute", &entry{Type: storageApplicative, Applicative: &applicative{Type: valueAttribute, Attribute: attributeValue}}, attributeValue},
		{"Definition", &entry{Type: storageApplicative, Applicative: &applicative{Type: valueDefinition, Definition: definitionValue}}, definitionValue},
		{"Function", &entry{Type: storageApplicative, Applicative: &applicative{Type: valueFunction, Function: functionValue}}, functionValue},
		{"Property", &entry{Type: storageApplicative, Applicative: &applicative{Type: valueProperty, Property: propertyValue}}, propertyValue},
		{"Operation", &entry{Type: storageApplicative, Applicative: &applicative{Type: valueOperation, Operation: operationValue}}, operationValue},
		{"List", &entry{Type: storageIterative, Iterative: &iterative{Type: valueList, List: listValue}}, listValue},
		{"Dictionary", &entry{Type: storageIterative, Iterative: &iterative{Type: valueDictionary, Dictionary: dictionaryValue}}, dictionaryValue},
		{"Lexeme", &entry{Type: storageLexeme, Lexeme: lexemeValue}, lexemeValue},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got, err := tc.entry.value()
			if err != nil {
				t.Fatal(err)
			}
			if got.Type.String() != tc.name || got.Data != tc.want {
				t.Fatalf("stored projection = %+v; want %s and the original payload pointer", got, tc.name)
			}
		})
	}
}

func TestRefactorLexicalInitializationAndLabelOrder(t *testing.T) {
	lex := newLexeme("verb.generic")
	if lex.POS != "verb.generic" || lex.Type != "verb.generic" || lex.Arguments != nil || lex.Inputs == nil || lex.Outputs == nil {
		t.Fatal("lexical initialization changed raw metadata or unbound/formal slice states")
	}
	lex.POS = "noun"
	if err := lex.reLabel(`"Dispatch|-es|-ive|-ing"`); err != nil {
		t.Fatal(err)
	}
	if lex.POS != "noun" || lex.Type != "verb.generic" || lex.Root != "dispatch" || lex.Arguments != nil {
		t.Fatal("relabeling changed unrelated lexical metadata or bound arguments")
	}
	for _, tc := range []struct{ word, form string }{{"dispatch", "root"}, {"dispatches", "plural"}, {"dispatchive", "descriptive"}, {"dispatching", "active"}} {
		matched, form := lex.match([]byte(tc.word))
		if !matched || form != tc.form {
			t.Errorf("%s = (%t, %q), want (true, %q); POS must be assigned before morphological label parsing", tc.word, matched, form, tc.form)
		}
	}
}
