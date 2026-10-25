package Lyger

import (
	"encoding/json"
	"reflect"
	"strings"
	"testing"
)

func actionDefinitionSnapshot(t *testing.T, value *lexeme) string {
	t.Helper()
	data, err := json.Marshal(value)
	if err != nil {
		t.Fatal(err)
	}
	return string(data)
}

func TestActionArgumentsPreservePositionalSignatureAndTemplate(t *testing.T) {
	for _, formals := range [][]string{{"Text", "Text"}, {"Text", "Text", "Text"}} {
		definition := &lexeme{Root: "collect", Inputs: append([]string(nil), formals...), Value: "source-defined prototype"}
		arguments := make([]string, len(formals))
		for index := range arguments {
			arguments[index] = "same-reference"
		}
		use := definition.cloneUse()
		for attempt := 0; attempt < 2; attempt++ {
			if err := use.bindArguments(arguments); err != nil {
				t.Fatal(err)
			}
			if err := use.construct(); err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(use.Inputs, formals) || !reflect.DeepEqual(use.Arguments, arguments) || use.Value != definition.Value {
				t.Fatalf("positional declaration/use changed: %+v", use)
			}
		}
		arguments[0] = "caller-mutated"
		clone := use.cloneUse()
		clone.Arguments[0] = "clone-mutated"
		if use.Arguments[0] != "same-reference" || definition.Arguments != nil || !reflect.DeepEqual(definition.Inputs, formals) {
			t.Fatal("argument storage aliases its caller or definition")
		}
	}
}

func TestActionArgumentValidationRetainsInvalidUseForDiagnosis(t *testing.T) {
	for _, arguments := range [][]string{{}, {"first"}, {"first", "second", "third"}} {
		use := &lexeme{Root: "pair", Inputs: []string{"Text", "Text"}, Value: "PAIR(inputs[0], inputs[1])"}
		if err := use.bindArguments(arguments); err != nil {
			t.Fatal(err)
		}
		before := actionDefinitionSnapshot(t, use)
		if err := use.construct(); err == nil || !strings.Contains(err.Error(), "requires 2 arguments") {
			t.Fatalf("wrong arity returned %v", err)
		}
		if actionDefinitionSnapshot(t, use) != before {
			t.Fatal("failed validation discarded arguments or changed the template")
		}
	}
	prototype := &lexeme{Root: "pair", Inputs: []string{"Text", "Text"}}
	if err := prototype.construct(); err != nil {
		t.Fatalf("unbound prototype rejected: %v", err)
	}
	zero := &lexeme{Root: "empty", Value: "EMPTY()"}
	if err := zero.bindArguments(nil); err != nil {
		t.Fatal(err)
	}
	if zero.cloneUse().Arguments == nil {
		t.Fatal("clone lost explicit zero-argument invocation")
	}
	if err := zero.construct(); err != nil {
		t.Fatal(err)
	}
	for _, inputs := range [][]string{{""}, {"valid", "\t"}} {
		prior := actionDefinitionSnapshot(t, zero)
		if err := zero.bindArguments(inputs); err == nil {
			t.Fatal("blank reference accepted")
		}
		if actionDefinitionSnapshot(t, zero) != prior {
			t.Fatal("failed argument load partially changed the use")
		}
	}
	if err := (&lexeme{Inputs: []string{""}}).validateInputs(); err == nil {
		t.Fatal("empty formal accepted")
	}
}

func TestActionSequencesStayFlatOrderedAndIsolated(t *testing.T) {
	for _, synchronous := range []bool{false, true} {
		first := &lexeme{Root: "prepare", Inputs: []string{"Resource"}, Arguments: []string{"target"}, Value: "PREPARE(inputs[0])"}
		second := &lexeme{Root: "submit", Inputs: []string{"Resource"}, Arguments: []string{"it"}, Value: "SUBMIT(inputs[0])"}
		if synchronous {
			second.Adverbs = []*token{{tType: "adverbs.progression.root"}}
		}
		sequence := []*lexeme{first, second}
		definition := &lexeme{Root: "dispatch"}
		for attempt := 0; attempt < 2; attempt++ {
			if err := definition.loadActions(sequence); err != nil {
				t.Fatal(err)
			}
			if len(definition.Actions) != 2 || definition.Actions[0].Root != first.Root || definition.Actions[1].Root != second.Root || definition.Synchronous != synchronous {
				t.Fatalf("sequence order/policy changed: %+v", definition)
			}
			if len(definition.Actions[0].Actions) != 0 || len(first.Actions) != 0 || definition.Value != "" {
				t.Fatal("following action became a dependency or a fabricated template")
			}
		}
		definition.Actions[0].Inputs[0] = "changed-signature"
		definition.Actions[1].Arguments[0] = "changed-reference"
		if first.Inputs[0] != "Resource" || second.Arguments[0] != "it" {
			t.Fatal("stored sequence aliases source uses")
		}
	}
}

func TestActionGraphErrorsDoNotPartiallyLoad(t *testing.T) {
	cycle := &lexeme{Root: "recursive"}
	cycle.Actions = []*lexeme{cycle}
	progression := &token{tType: "adverbs.progression.root"}
	for _, actions := range [][]*lexeme{
		nil,
		{nil},
		{{Root: "before"}, cycle},
		{{Root: "before", Adverbs: []*token{progression}}, {Root: "after"}},
		{{Root: "broken", Adverbs: []*token{nil}}},
	} {
		prior := &lexeme{Root: "retained"}
		definition := &lexeme{Root: "container", Actions: []*lexeme{prior}, Synchronous: true}
		if err := definition.loadActions(actions); err == nil {
			t.Fatal("invalid action graph accepted")
		}
		if len(definition.Actions) != 1 || definition.Actions[0] != prior || !definition.Synchronous {
			t.Fatal("failed action load changed stored sequence")
		}
	}
	for _, invalid := range []*lexeme{nil, cycle, {Root: "missing", Actions: []*lexeme{nil}}} {
		if err := invalid.construct(); err == nil {
			t.Fatal("invalid legacy graph accepted")
		}
	}
	container := &lexeme{Root: "container"}
	if err := container.loadSyncActions([]*lexeme{container}); err == nil {
		t.Fatal("self-containing action sequence accepted")
	}
}

func TestActionConstructionDoesNotResolveScopedReferences(t *testing.T) {
	child := &lexeme{Root: "apply", Inputs: []string{"it"}, Arguments: []string{"it"}, Value: "original child template"}
	definition := &lexeme{Root: "compose", Inputs: []string{"Resource"}, Arguments: []string{"different"}, Actions: []*lexeme{child}}
	if err := definition.construct(); err != nil {
		t.Fatal(err)
	}
	if child.Inputs[0] != "it" || child.Arguments[0] != "it" || child.Value != "original child template" || definition.Value != "" {
		t.Fatal("grammar-only construction guessed a pronoun target or rewrote source")
	}
}

func TestActionDeclarationExtractionPreservesSourceUses(t *testing.T) {
	statements, _, _ := instructionTestGrammar(t, "Create \"Stage|-s|-d|-ing\" for Building any App and then Running it.\n", nil)
	clauses := directiveClauses(statements[0])
	if len(clauses) != 2 {
		t.Fatalf("source action clauses = %d", len(clauses))
	}
	prototypes := []*lexeme{clauses[0].verb.mem.Lexeme.cloneUse(), clauses[1].verb.mem.Lexeme.cloneUse()}
	for attempt := 0; attempt < 2; attempt++ {
		records, err := statements[0].sGram.getDeclarations()
		if err != nil || len(records) != 1 {
			t.Fatalf("extract declaration: %v (%d records)", err, len(records))
		}
		definition := records[0].Record.Value.Lexeme
		if len(definition.Actions) != 2 || !definition.Synchronous || definition.Value != "" {
			t.Fatalf("source action definition = %+v", definition)
		}
		for index, action := range definition.Actions {
			if action.Root != prototypes[index].Root || !reflect.DeepEqual(action.Inputs, prototypes[index].Inputs) || action.Value != prototypes[index].Value || len(action.Actions) != 0 {
				t.Fatalf("action %d lost prototype or order: %+v", index, action)
			}
			if !reflect.DeepEqual(action.Arguments, []string{clauses[index].obj.referenceID()}) {
				t.Fatalf("action %d references = %v", index, action.Arguments)
			}
			if actionDefinitionSnapshot(t, clauses[index].verb.mem.Lexeme) != actionDefinitionSnapshot(t, prototypes[index]) {
				t.Fatal("extracting declaration mutated its source token")
			}
		}
		if definition.Actions[1].Arguments[0] != "it" {
			t.Fatal("declaration extraction resolved a scoped pronoun early")
		}
	}
}

func TestActionExtractionReportsUnboundAndEmptyReferences(t *testing.T) {
	for _, clause := range []*clauseGram{nil, {}, {verb: &token{mem: &entry{Type: storagePrimitive}}}} {
		if _, err := clause.getAction(); err == nil {
			t.Fatal("unbound directive action accepted")
		}
	}
	prototype := &lexeme{Root: "save", Inputs: []string{"Resource"}}
	clause := &clauseGram{verb: &token{mem: &entry{Type: storageLexeme, Lexeme: prototype}}, obj: &objectGram{}}
	if _, err := clause.getAction(); err == nil || !strings.Contains(err.Error(), "empty argument") {
		t.Fatalf("missing action reference returned %v", err)
	}
	if prototype.Arguments != nil {
		t.Fatal("failed action extraction changed source lexeme")
	}
}
