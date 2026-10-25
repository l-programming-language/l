package Lyger

import (
	"reflect"
	"strings"
	"testing"
)

func grammarTestTokens(t *testing.T, source string, gram *grammar) []*token {
	t.Helper()
	stmt, err := newLStatement([]byte(source), gram)
	if err != nil {
		t.Fatal(err)
	}
	if err := stmt.tokenize(); err != nil {
		t.Fatal(err)
	}
	return stmt.tokens
}

func TestGrammarTokenSpans(t *testing.T) {
	a, b, c := &token{rawBytes: []byte("same")}, &token{rawBytes: []byte("same")}, &token{rawBytes: []byte("c")}
	for _, test := range []struct {
		name          string
		source, match []*token
		found         bool
		positions     []int
		remaining     []*token
	}{
		{"suffix", []*token{a, b, c}, []*token{b, c}, true, []int{1, 2}, []*token{a}},
		{"single terminal", []*token{a, b, c}, []*token{c}, true, []int{2, 2}, []*token{a, b}},
		{"repeated occurrence", []*token{a, b, c}, []*token{b}, true, []int{1, 1}, []*token{a, c}},
		{"equal length mismatch", []*token{a, b}, []*token{b, c}, false, []int{0, 0}, []*token{a, b}},
		{"empty", []*token{a, b}, nil, false, []int{0, 0}, []*token{a, b}},
	} {
		t.Run(test.name, func(t *testing.T) {
			original := append([]*token(nil), test.source...)
			match, positions := tokenMatch(test.source, test.match)
			if match != test.found || !reflect.DeepEqual(positions, test.positions) {
				t.Fatalf("match=%v positions=%v", match, positions)
			}
			if remaining := tokenReducer(test.source, test.match); !reflect.DeepEqual(remaining, test.remaining) {
				t.Fatalf("remaining=%v want=%v", remaining, test.remaining)
			}
			if !reflect.DeepEqual(test.source, original) {
				t.Fatal("reduction mutated shared source tokens")
			}
		})
	}
}

func TestGrammarPossessiveSubjectConsumesSourceOrder(t *testing.T) {
	gram, err := loadGrammar(canonicalTestOptions())
	if err != nil {
		t.Fatal(err)
	}
	tokens := grammarTestTokens(t, "any Function's \"Return|-s|-ed|-ing|-er\" to `definition`", gram)
	// A resolved dynamic noun retains its source morphology; binding is tested separately.
	tokens[1].tType = "nouns.generic.root.possessive"
	tokens[1].defined = true
	subject, err := initSubjectGram("subject.utilization", tokens, gram)
	if err != nil {
		t.Fatal(err)
	}
	if err = subject.build(); err != nil {
		t.Fatal(err)
	}
	if subject.noun != tokens[2] || !reflect.DeepEqual(subject.getLexTokens(), tokens[:3]) || len(subject.tokens) != len(tokens) {
		t.Fatalf("subject did not consume source order: %+v", subject)
	}
	if subject.phrase == nil || subject.phrase.obj == nil || subject.phrase.obj.noun != tokens[4] {
		t.Fatal("return definition was not associated")
	}
}

func TestGrammarBoundReferenceHasDeterministicCase(t *testing.T) {
	gram, err := loadGrammar(canonicalTestOptions())
	if err != nil {
		t.Fatal(err)
	}
	for attempt := 0; attempt < 100; attempt++ {
		tok := &token{rawBytes: []byte("x"), tType: "primitive.text", defined: true, gram: gram}
		object, err := initObjectGram("object", []*token{tok}, gram)
		if err != nil {
			t.Fatal(err)
		}
		if err = object.build(); err != nil {
			t.Fatal(err)
		}
		if object.cType != "definitive" || object.noun != tok {
			t.Fatalf("reference selected %s", object.cType)
		}
	}
}

func TestGrammarAlternativeRollsBackNestedPhrase(t *testing.T) {
	gram, err := loadGrammar(canonicalTestOptions())
	if err != nil {
		t.Fatal(err)
	}
	gram.grammarMap["subject"] = grammeme{Structure: "CASE", Cases: map[string]string{
		"a_failed": "\"ID\" | PHRASE.DECLARATIVE.DEFINITIONAL | VERB",
		"z_valid":  "\"ID\" | PHRASE.DECLARATIVE.DEFINITIONAL",
	}}
	tokens := grammarTestTokens(t, "\"label\" as `definition`", gram)
	subject, err := initSubjectGram("subject", tokens, gram)
	if err != nil {
		t.Fatal(err)
	}
	if err = subject.build(); err != nil {
		t.Fatal(err)
	}
	if subject.cType != "z_valid" || subject.phrase == nil || subject.phrase.obj.noun != tokens[2] {
		t.Fatalf("failed alternative contaminated result: %+v", subject)
	}
}

func TestGrammarOptionalFailureLeavesParentSuffix(t *testing.T) {
	gram, err := loadGrammar(canonicalTestOptions())
	if err != nil {
		t.Fatal(err)
	}
	gram.grammarMap["object"] = grammeme{Structure: "ID | PHRASE.DECLARATIVE.DEFINITIONAL*"}
	tokens := grammarTestTokens(t, "it as", gram)
	object, err := initObjectGram("object", tokens, gram)
	if err != nil {
		t.Fatal(err)
	}
	if err = object.build(); err != nil {
		t.Fatal(err)
	}
	if object.phrase != nil || len(object.tokens) != 1 || object.tokens[0] != tokens[0] {
		t.Fatalf("failed optional phrase consumed parent suffix: %+v", object)
	}
}

func TestGrammarStatementReturnsParseError(t *testing.T) {
	gram, err := loadGrammar(canonicalTestOptions())
	if err != nil {
		t.Fatal(err)
	}
	tokens := grammarTestTokens(t, "Define any", gram)
	statement, err := initStatementGram("verbs.definitive.root", tokens, gram)
	if err != nil {
		t.Fatal(err)
	}
	if err = statement.build(); err == nil {
		t.Fatal("incomplete definition must return a parse error")
	}
}

func TestGrammarRepeatedObjectsAndClauses(t *testing.T) {
	for _, test := range []struct{ name, source, statementType string }{
		{"sequential actions", "Create \"Launch|-s|-ed|-ing|-er\" for Building any App and then Running it", "verbs.declarative.root"},
		{"descriptor targets", "\"Primary\" is a New Adjective describing any Service, Object, or Function", "verbs.passive.root"},
		{"conditional declaration", "If it has an Error, exit with the Error, else define \"result\" as it's Output", "evaluations.conditional.root"},
		{"otherwise sequence", "Otherwise, show it and then exit", "evaluations.alternative.root"},
		{"named descriptor", "It describes a Generic `entry` Function of any Process with No Inputs and No Outputs", "verbs.descriptive.active"},
		{"named object", "It describes a Generic `entry` Object with a New Objective Primary Function using any Process", "verbs.descriptive.active"},
		{"named service", "It describes a Generic `entry` Service with a New Primary Object using any Process", "verbs.descriptive.active"},
		{"app descriptor", "\"Simple\" describes a Generic App with a New Primary Service using any Process", "verbs.descriptive.active"},
		{"possessive declaration", "Define any Function's \"Result|-s|-ed|-ing|-er\" to `definition`", "verbs.definitive.root"},
	} {
		t.Run(test.name, func(t *testing.T) {
			gram, err := loadGrammar(canonicalTestOptions())
			if err != nil {
				t.Fatal(err)
			}
			tokens := grammarTestTokens(t, test.source, gram)
			for _, tok := range tokens {
				if string(tok.rawBytes) == "Function" || string(tok.rawBytes) == "Process" {
					tok.tType = "nouns.generic.root"
					tok.defined = true
				}
				if string(tok.rawBytes) == "Primary" || string(tok.rawBytes) == "Objective" {
					tok.tType = "adjectives.generic.root"
					tok.defined = true
				}
				if string(tok.rawBytes) == "Function's" {
					tok.tType = "nouns.generic.root.possessive"
					tok.defined = true
				}
			}
			statement, err := initStatementGram(test.statementType, tokens, gram)
			if err != nil {
				t.Fatal(err)
			}
			if err = statement.build(); err != nil {
				t.Fatal(err)
			}
			if len(statement.getUnassociatedTokens()) != 0 {
				t.Fatal("whole statement left unassociated tokens")
			}
			if test.name == "sequential actions" {
				phrase := statement.predicate.clauses[0].subject.phrase
				if len(phrase.clauses) != 2 || len(phrase.conjunctions) != 1 {
					t.Fatalf("sequence missing clauses or connector: %+v", phrase)
				}
			}
			if test.name == "descriptor targets" {
				clause := statement.predicate.obj.clause
				if clause == nil || len(clause.objects) != 3 || len(clause.conjunctions) != 1 {
					t.Fatalf("descriptor target list missing: %+v", clause)
				}
			}
			if test.name == "otherwise sequence" {
				conditional := statement.predicate.conditionals[0]
				if len(conditional.clauses) != 2 {
					t.Fatalf("otherwise lost an imperative: %+v", conditional)
				}
			}
			if test.name == "named descriptor" {
				object := statement.predicate.clauses[0].obj
				if object.name == nil || string(object.name.noun.rawBytes) != "`entry`" || string(object.noun.rawBytes) != "Function" {
					t.Fatalf("name and nominal were conflated: %+v", object)
				}
			}
		})
	}
}

func TestGrammarRejectsTrailingSyntax(t *testing.T) {
	gram, err := loadGrammar(canonicalTestOptions())
	if err != nil {
		t.Fatal(err)
	}
	for _, source := range []string{"Otherwise, show it and then exit stray", "Otherwise, show it and", "Create \"Example\" as an App stray"} {
		tokens := grammarTestTokens(t, source, gram)
		statementType := "evaluations.alternative.root"
		if string(tokens[0].rawBytes) == "Create" {
			statementType = "verbs.declarative.root"
		}
		statement, err := initStatementGram(statementType, tokens, gram)
		if err != nil {
			t.Fatal(err)
		}
		if err = statement.build(); err == nil {
			t.Errorf("accepted trailing syntax: %s", source)
		}
	}
}

func TestGrammarBoundObjectExtractionDoesNotMutateDefinition(t *testing.T) {
	gram, err := loadGrammar(canonicalTestOptions())
	if err != nil {
		t.Fatal(err)
	}
	definition, err := initLexicalEntry([]byte("<DEF>FW^<@Process@><DEF>"))
	if err != nil {
		t.Fatal(err)
	}
	definition.Lexeme.POS = "noun"
	if err := definition.Lexeme.reLabel(`"Widget|-s|-al|-ing"`); err != nil {
		t.Fatal(err)
	}
	memory := map[string]*record{"Widget": {ID: "Widget", Type: definition.Lexeme.getValueType(), Value: definition}}
	tokens := grammarTestTokens(t, "any Widget", gram)
	scope := &lexicalScope{records: memory}
	for _, tok := range tokens {
		tok.scope = scope
	}
	if state, err := (semanticResolver{}).bindToken(tokens[1]); err != nil || state != lookupFound {
		t.Fatal("dynamic noun did not bind")
	}
	object, err := initObjectGram("object", tokens, gram)
	if err != nil {
		t.Fatal(err)
	}
	if err = object.build(); err != nil {
		t.Fatal(err)
	}
	value, kind, err := object.getValue(true)
	if err != nil {
		t.Fatal(err)
	}
	if value.Type.String() != "Lexeme" || kind != definition.Lexeme.getValueType() || value.Lexeme.Value != definition.Lexeme.Value {
		t.Fatalf("bound object lost definition: value=%+v kind=%s", value, kind)
	}
	value.Lexeme.POS = "adjective"
	if err := value.Lexeme.reLabel(`"Fancy"`); err != nil {
		t.Fatal(err)
	}
	if definition.Lexeme.POS != "noun" || definition.Lexeme.Root != "widget" || tokens[1].mem.Lexeme.Root != "widget" {
		t.Fatal("declaration mutation changed the stored or token-bound definition")
	}
}

func TestGrammarDescriptorExtractionUsesExplicitDefinition(t *testing.T) {
	gram, err := loadGrammar(canonicalTestOptions())
	if err != nil {
		t.Fatal(err)
	}
	definition, err := initLexicalEntry([]byte("<DEF>FW^<@Process@><DEF>"))
	if err != nil {
		t.Fatal(err)
	}
	definition.Lexeme.POS = "noun"
	if err := definition.Lexeme.reLabel(`"Widget|-s|-al|-ing"`); err != nil {
		t.Fatal(err)
	}
	tokens := grammarTestTokens(t, "\"Decorated\" describes any Widget as `<DEF>A^<@Widget@><DEF>`", gram)
	scope := &lexicalScope{records: map[string]*record{"Widget": {ID: "Widget", Type: definition.Lexeme.getValueType(), Value: definition}}}
	for _, tok := range tokens {
		tok.scope = scope
	}
	if state, err := (semanticResolver{}).bindToken(tokens[3]); err != nil || state != lookupFound {
		t.Fatal("dynamic noun did not bind")
	}
	statement, err := initStatementGram("verbs.descriptive.active", tokens, gram)
	if err != nil {
		t.Fatal(err)
	}
	if err = statement.build(); err != nil {
		t.Fatal(err)
	}
	records, err := statement.getDeclarations()
	if err != nil {
		t.Fatal(err)
	}
	if len(records) != 1 || records[0].Record.ID != "Decorated" || records[0].Record.Value.Lexeme.POS != "adjective" || strings.Trim(records[0].Record.Value.Lexeme.Value, "`") != "A^<@Widget@>" {
		t.Fatalf("explicit descriptor definition was not extracted: record=%+v lexeme=%+v", records[0], records[0].Record.Value.Lexeme)
	}
	if definition.Lexeme.POS != "noun" || definition.Lexeme.Root != "widget" || definition.Lexeme.Value != "FW^<@Process@>" {
		t.Fatal("descriptor declaration mutated its semantic target")
	}
}

func TestGrammarNewLexicalTypeRetainsDescriptorClause(t *testing.T) {
	gram, err := loadGrammar(canonicalTestOptions())
	if err != nil {
		t.Fatal(err)
	}
	tokens := grammarTestTokens(t, "\"Featured\" is a New Adjective describing any Service or Object", gram)
	statement, err := initStatementGram("verbs.passive.root", tokens, gram)
	if err != nil {
		t.Fatal(err)
	}
	if err = statement.build(); err != nil {
		t.Fatal(err)
	}
	records, err := statement.getDeclarations()
	if err != nil {
		t.Fatal(err)
	}
	if len(records) != 1 || records[0].Record.ID != "Featured" || records[0].Record.Value.Lexeme.POS != "adjective" {
		t.Fatalf("raw lexical type was not initialized: %+v", records)
	}
	if statement.predicate.obj.clause == nil || len(statement.predicate.obj.clause.objects) != 2 {
		t.Fatal("new lexical type lost its descriptor targets")
	}
}

func TestGrammarNativeSubjectReferenceAndDefinedVerbs(t *testing.T) {
	for _, example := range []struct {
		source string
		valid  bool
	}{
		{"The App Renders `payload`", true},
		{"The Service Renders `payload`", true},
		{"Render the App", true},
		{"Render `payload`", true},
		{"App Renders `payload`", false},
		{"A New App Renders `payload`", false},
		{"The Unknown Renders `payload`", false},
		{"The App Renders", false},
		{"The App Renders `payload` stray", false},
		{"Render unknown", false},
	} {
		t.Run(example.source, func(t *testing.T) {
			gram, err := loadGrammar(canonicalTestOptions())
			if err != nil {
				t.Fatal(err)
			}
			verb, err := initLexicalEntry([]byte("<DEF>BOD^<@Input|Thing@><DEF>"))
			if err != nil {
				t.Fatal(err)
			}
			verb.Lexeme.POS = "verb"
			if err := verb.Lexeme.reLabel(`"Render|-s|-ed|-ing|-er"`); err != nil {
				t.Fatal(err)
			}
			tokens := grammarTestTokens(t, example.source, gram)
			scope := &lexicalScope{records: map[string]*record{"Render": {ID: "Render", Type: verb.Lexeme.getValueType(), Value: verb}}}
			kind := ""
			for _, tok := range tokens {
				tok.scope = scope
				state, err := (semanticResolver{}).bindToken(tok)
				if err != nil {
					t.Fatal(err)
				}
				if state == lookupFound {
					kind = tok.tType
				} else if state != lookupAbsent {
					t.Fatalf("unexpected synthetic binding state: %s", state)
				}
			}
			statement, err := initStatementGram(kind, tokens, gram)
			if err == nil {
				err = statement.build()
			}
			if (err == nil) != example.valid {
				t.Fatalf("valid=%v error=%v", example.valid, err)
			}
			if err != nil {
				return
			}
			if strings.HasPrefix(example.source, "The ") {
				if statement.subject == nil || statement.subject.cType != "reference" || !strings.HasPrefix(statement.subject.noun.tType, "type.") {
					t.Fatal("native subject lost its reference role or lexical type")
				}
				if statement.predicate == nil || len(statement.predicate.clauses) != 1 || statement.predicate.clauses[0].obj == nil {
					t.Fatal("bound active verb lost its object")
				}
			}
		})
	}
}
