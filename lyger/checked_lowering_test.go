package Lyger

import (
	"reflect"
	"strings"
	"testing"
)

func TestCheckedDefinitionArgumentReads(t *testing.T) {
	for _, tc := range []struct {
		value *primitive
		kind  string
		text  string
	}{
		{&primitive{Type: valueText, Text: "`雪\n<DEF>`"}, "Text", "雪\n<DEF>"},
		{&primitive{Type: valueNumber, Number: -2.5}, "Number", "-2.5"},
		{&primitive{Type: valueTruth, Truth: true}, "Truth", "true"},
	} {
		actual, err := argumentFromToken(&token{mem: &entry{Type: storagePrimitive, Primitive: tc.value}})
		if err != nil || actual.Kind != tc.kind || actual.Text != tc.text {
			t.Fatalf("definition argument = %#v/%v, want %s/%q", actual, err, tc.kind, tc.text)
		}
	}
	malformed := &entry{Type: storagePrimitive, Primitive: &primitive{Type: valueText, Text: "`value`"}, Lexeme: &lexeme{}}
	if actual, err := argumentFromToken(&token{mem: malformed}); err == nil || !strings.Contains(err.Error(), "requires one storage payload") || !reflect.DeepEqual(actual, DefinitionArgument{}) {
		t.Fatalf("malformed argument storage produced %#v/%v", actual, err)
	}
	malformed.Runtime = &runtimeReference{OwnerID: "worker", Member: memberOutput}
	if _, err := argumentFromToken(&token{mem: malformed}); err == nil || err.Error() != "runtime worker.output is unavailable during source generation" {
		t.Fatalf("runtime unavailability lost precedence over storage errors: %v", err)
	}
	for _, value := range []*primitive{{Type: valueThing}, {Type: valueBytes}, {Type: valuePointer, Pointer: "reference"}} {
		if _, err := argumentFromToken(&token{mem: &entry{Type: storagePrimitive, Primitive: value}}); err == nil {
			t.Fatalf("unsupported stored %s became a definition argument", value.Type)
		}
	}
}

func TestCheckedLoweringRejectsMalformedResolvedStorage(t *testing.T) {
	for _, tc := range []struct {
		name, want string
		value      *entry
	}{
		{"missing App", "applicative App has no readable value", &entry{Type: storageApplicative, Applicative: &applicative{Type: valueApp}}},
		{"multiple wrappers", "requires one storage payload", &entry{Type: storageApplicative, Applicative: &applicative{Type: valueApp, App: &App{ID: "Worker"}}, Primitive: &primitive{Type: valueText}}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			record := &record{ID: "Worker", Value: tc.value}
			operand := &instructionValue{Kind: operandSymbol, Record: record}
			planner := orchestrationPlanner{semantics: newSemanticProgram()}
			if actual, kind, err := planner.record(operand); err == nil || !strings.Contains(err.Error(), tc.want) || actual != nil || kind != valueInvalid {
				t.Fatalf("omitted type annotation hid malformed symbol storage: %v/%v", actual, err)
			}
			noun := &token{rawBytes: []byte("Worker"), mem: tc.value, binding: &tokenBinding{Record: record, Root: "Worker"}}
			compilation := &Compilation{Semantics: newSemanticProgram()}
			statement := &lStatement{sGram: &statementGram{subject: &subjectGram{noun: noun}}}
			if actual, err := compilation.statementApp(statement); err == nil || !strings.Contains(err.Error(), tc.want) || actual != nil {
				t.Fatalf("application body accepted malformed owner storage: %v/%v", actual, err)
			}
			unboundNominal := &token{rawBytes: []byte("App"), mem: tc.value}
			if actual, err := compilation.constructObject(&objectGram{noun: unboundNominal}, make(map[*descriptorRule]bool)); err == nil || !strings.Contains(err.Error(), tc.want) || actual != nil {
				t.Fatalf("construction accepted malformed resolved nominal: %v/%v", actual, err)
			}
		})
	}
}

func TestCheckedDefinitionDiscoveryPropagatesStorageError(t *testing.T) {
	record := &record{ID: "ForeignDefinition"}
	invalid := &token{mem: &entry{Type: storagePrimitive, Primitive: &primitive{Type: valueText, Text: "`ordinary text`"}, Lexeme: &lexeme{}}}
	compilation := &Compilation{Semantics: newSemanticProgram()}
	target := &token{rawBytes: []byte(`"ForeignDefinition"`), tType: "primitive.text", mem: &entry{Type: storagePrimitive, Primitive: &primitive{Type: valueText, Text: `"ForeignDefinition"`}}}
	compilation.Semantics.Subjects[target] = record
	compilation.Semantics.Declarations[record] = &lStatement{sGram: &statementGram{subject: &subjectGram{noun: target}, predicate: &predicateGram{cType: "assign", obj: &objectGram{tokens: []*token{invalid}}}}}
	if actual, err := compilation.recordTemplate(record); err == nil || !strings.Contains(err.Error(), "requires one storage payload") || actual != nil {
		t.Fatalf("definition lookup converted malformed storage into an absent template: %v/%v", actual, err)
	}
}

func TestCheckedApplicationBodyUsesCanonicalAliasOwner(t *testing.T) {
	app := &App{ID: "Worker"}
	canonical := &record{ID: "Worker", Value: &entry{Type: storageApplicative, Applicative: &applicative{Type: valueApp, App: app}}}
	alias := &record{ID: "Alias", Value: canonical.Value}
	semantics := newSemanticProgram()
	semantics.Aliases[alias] = canonical
	compilation := &Compilation{Semantics: semantics}
	noun := &token{rawBytes: []byte("Alias"), mem: alias.Value, binding: &tokenBinding{Record: alias, Root: "Alias"}}
	actual, err := compilation.statementApp(&lStatement{sGram: &statementGram{subject: &subjectGram{noun: noun}}})
	if err != nil || actual != app {
		t.Fatalf("application body lost canonical alias owner: %p/%v, want %p", actual, err, app)
	}
}

func TestCheckedApplicationBodyIgnoresOtherApplicativePayloads(t *testing.T) {
	record := &record{ID: "App", Value: &entry{Type: storageApplicative, Applicative: &applicative{Type: valueService, Service: &service{}}}}
	noun := &token{rawBytes: []byte("App"), mem: record.Value, binding: &tokenBinding{Record: record, Root: "App"}}
	compilation := &Compilation{Semantics: newSemanticProgram()}
	actual, err := compilation.statementApp(&lStatement{sGram: &statementGram{subject: &subjectGram{noun: noun}}})
	if err != nil || actual != nil {
		t.Fatalf("non-App applicative was treated as an application subject: %v/%v", actual, err)
	}
}
