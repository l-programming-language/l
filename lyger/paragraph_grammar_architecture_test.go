package Lyger

import (
	"reflect"
	"strings"
	"testing"
)

func paragraphGrammarStatement(t *testing.T, source string, memory map[string]*record) (*lStatement, error) {
	t.Helper()
	config := canonicalTestOptions()
	gram, err := loadGrammar(config)
	if err != nil {
		return nil, err
	}
	stmt, err := newLStatement([]byte(source), gram)
	if err != nil {
		return nil, err
	}
	if err = stmt.tokenize(); err != nil {
		return nil, err
	}
	scope := &lexicalScope{records: memory}
	for _, tok := range stmt.tokens {
		tok.scope = scope
		if tok.getPOS(false) == "unknown" {
			state, err := (semanticResolver{}).bindToken(tok)
			if err != nil {
				return nil, err
			}
			if state != lookupFound && state != lookupAbsent {
				t.Fatalf("unexpected synthetic binding state: %s", state)
			}
		}
	}
	stmt.determineType()
	return stmt, stmt.loadGram()
}

func TestParagraphGrammarVisibilityTargets(t *testing.T) {
	for _, tc := range []struct {
		name, source string
		want         []declarationVisibility
	}{
		{"first imperative", "Publicly define \"first\" as `left` and define \"second\" as `right`", []declarationVisibility{visibilityPublic, visibilityDefault}},
		{"second imperative", "Define \"first\" as `left` and publicly define \"second\" as `right`", []declarationVisibility{visibilityDefault, visibilityPublic}},
		{"first passive", "Public \"first\" is `left` and \"second\" is `right`", []declarationVisibility{visibilityPublic, visibilityDefault}},
		{"second passive value", "\"first\" is `left` and \"second\" is a Public App", []declarationVisibility{visibilityDefault, visibilityPublic}},
		{"conditional", "If true is true, publicly define \"first\" as `left` and define \"second\" as `right`, else publicly define \"third\" as `other`", []declarationVisibility{visibilityPublic, visibilityDefault, visibilityPublic}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			stmt, err := paragraphGrammarStatement(t, tc.source, nil)
			if err != nil {
				t.Fatal(err)
			}
			records, err := stmt.sGram.getDeclarations()
			if err != nil {
				t.Fatal(err)
			}
			var got []declarationVisibility
			for _, record := range records {
				got = append(got, record.Visibility)
				visibility, err := stmt.sGram.declarationVisibility(record.Token)
				if err != nil || visibility != record.Visibility {
					t.Fatalf("early visibility = %v, %v; extraction = %v", visibility, err, record.Visibility)
				}
				if strings.Contains(strings.ToLower(record.Record.ID), "public") {
					t.Fatalf("visibility changed ID %q", record.Record.ID)
				}
			}
			if !reflect.DeepEqual(got, tc.want) {
				t.Fatalf("visibility = %v, want %v", got, tc.want)
			}
		})
	}
}

func TestParagraphGrammarVisibilityRejectsNonDeclarations(t *testing.T) {
	if _, err := paragraphGrammarStatement(t, `Create "Worker" as a Public App`, nil); err != nil {
		t.Fatal(err)
	}
	for _, source := range []string{"Publicly Run any App", "Publicly Show `visible`", "Show Public `visible`", "Define \"first\" as `left` and publicly Show `visible`", `Create "Launch" for Publicly Building any App`} {
		t.Run(source, func(t *testing.T) {
			_, err := paragraphGrammarStatement(t, source, nil)
			if err == nil || !strings.Contains(err.Error(), "visibility modifier") {
				t.Fatalf("expected declaration modifier diagnostic, got %v", err)
			}
		})
	}
}

func TestParagraphGrammarEarlyVisibilityAndValueDescriptors(t *testing.T) {
	memory := map[string]*record{"Original": instructionRecord("Original", "App", "")}
	for _, source := range []string{`Publicly Use Original as "Copy"`, `Public "Shape" is a New Adjective describing any Service or Object`} {
		stmt, err := paragraphGrammarStatement(t, source, memory)
		if err != nil {
			t.Fatal(err)
		}
		var target *token
		for _, tok := range stmt.tokens {
			if tok.isDeclare() {
				target = tok
			}
		}
		visibility, err := stmt.sGram.declarationVisibility(target)
		if err != nil || visibility != visibilityPublic {
			t.Fatalf("early descriptor/alias visibility = %s, %v", visibility, err)
		}
	}
	config := canonicalTestOptions()
	gram, err := loadGrammar(config)
	if err != nil {
		t.Fatal(err)
	}
	tokens := grammarTestTokens(t, "Public Generic", gram)
	lex := newLexeme("noun")
	lex.Root = "item"
	noun := &token{rawBytes: []byte("Item"), mem: &entry{Type: storageLexeme, Lexeme: lex}}
	object := &objectGram{cType: "definitive", noun: noun, adjectives: tokens}
	value, _, err := object.getValue(false)
	if err != nil {
		t.Fatal(err)
	}
	if object.referenceID() != "Generic.Item" || len(value.Lexeme.Adjectives) != 1 || value.Lexeme.Adjectives[0] != tokens[1] || len(lex.Adjectives) != 0 {
		t.Fatal("visibility leaked into a reference/value descriptor or changed the original lexeme")
	}
}

func TestParagraphGrammarVisibilityFollowsConfiguredRole(t *testing.T) {
	config := canonicalTestOptions()
	gram, err := loadGrammar(config)
	if err != nil {
		t.Fatal(err)
	}
	adjective := gram.lex.Adjectives[publicVisibilityRole]
	adverb := gram.lex.Adverbs[publicVisibilityRole]
	adjective.Root, adverb.Root = "shared", "globally"
	gram.lex.Adjectives[publicVisibilityRole], gram.lex.Adverbs[publicVisibilityRole] = adjective, adverb
	for _, source := range []string{`Create "Worker" as a Shared App`, "Globally define \"Greeting\" as `hello`"} {
		stmt, err := newLStatement([]byte(source), gram)
		if err != nil {
			t.Fatal(err)
		}
		if err := stmt.tokenize(); err != nil {
			t.Fatal(err)
		}
		stmt.determineType()
		if err := stmt.loadGram(); err != nil {
			t.Fatal(err)
		}
		records, err := stmt.sGram.getDeclarations()
		if err != nil || len(records) != 1 || records[0].Visibility != visibilityPublic {
			t.Fatalf("renamed role lost visibility: %+v, %v", records, err)
		}
	}
	for _, pos := range []string{"noun", "verb", "unknown"} {
		if _, ok := (&lexeme{POS: pos, Type: publicVisibilityRole}).declarationVisibility(); ok {
			t.Fatalf("visibility accepted %s", pos)
		}
	}
	if visibilityDefault.String() != "default" || visibilityPublic.String() != "public" || declarationVisibility(255).String() != "declarationVisibility(255)" {
		t.Fatal("unchecked declaration visibility enum names")
	}
}

func TestParagraphGrammarDeclarations(t *testing.T) {
	for _, tc := range []struct {
		name, source string
		ids          []string
	}{
		{"passive coordination", "\"first\" is `left` and \"second\" is `right` and \"third\" is 42", []string{"first", "second", "third"}},
		{"imperative coordination", "Define \"first\" as `left` and define \"second\" as `right`", []string{"first", "second"}},
		{"passive action continuation", "\"first\" is `left` and define \"second\" as `right`", []string{"first", "second"}},
		{"conditional branches", "If true is true, define \"first\" as `left` and define \"second\" as `right`, else define \"third\" as `other` and define \"fourth\" as `last`", []string{"first", "second", "third", "fourth"}},
		{"public App", `Create "Worker" as a New Public App`, []string{"Worker"}},
		{"public Dictionary", `Create "Library" as a Public Dictionary`, []string{"Library"}},
		{"public Text", "Publicly define \"Greeting\" as `hello`", []string{"Greeting"}},
		{"public Number", `Publicly define "Count" as 42`, []string{"Count"}},
		{"public Truth", `Publicly define "Enabled" as true`, []string{"Enabled"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			stmt, err := paragraphGrammarStatement(t, tc.source, nil)
			if err != nil {
				t.Fatal(err)
			}
			records, err := stmt.sGram.getDeclarations()
			if err != nil {
				t.Fatal(err)
			}
			var ids []string
			var declared []*token
			for _, tok := range stmt.tokens {
				if tok.isDeclare() {
					declared = append(declared, tok)
				}
			}
			for i, extracted := range records {
				ids = append(ids, extracted.Record.ID)
				if i >= len(declared) || extracted.Token != declared[i] {
					t.Fatal("extraction changed source declaration order or token identity")
				}
				if _, err := extracted.Record.Value.value(); err != nil {
					t.Fatal(err)
				}
			}
			if !reflect.DeepEqual(ids, tc.ids) {
				t.Fatalf("declarations = %v, want %v", ids, tc.ids)
			}
		})
	}
}

func TestParagraphGrammarConjunctionOwnership(t *testing.T) {
	for _, tc := range []struct {
		source string
		valid  bool
	}{
		{"\"first\" is `left` and Show `visible`", true},
		{"\"first\" is `left` and \"second\" is `right`", true},
		{"\"first\" is `left` and", false},
		{"\"first\" is `left` and \"second\" is `right` and", false},
		{"\"first\" is `left` and unrecognized", false},
		{"Show `visible` and", false},
		{"and Show `visible`", false},
		{"If true is true, and Show `visible`", false},
	} {
		t.Run(tc.source, func(t *testing.T) {
			_, err := paragraphGrammarStatement(t, tc.source, nil)
			if (err == nil) != tc.valid {
				t.Fatalf("grammar error = %v, valid = %v", err, tc.valid)
			}
		})
	}
}

func TestParagraphGrammarCoordinatedActionObjects(t *testing.T) {
	memory := map[string]*record{"Launch": instructionAction(t, `"Launch|-es|-ed|-ing"`), "First": instructionRecord("First", "App", ""), "Second": instructionRecord("Second", "App", "")}
	stmt, err := paragraphGrammarStatement(t, "Launch First and Second and Show `done`", memory)
	if err != nil {
		t.Fatal(err)
	}
	clauses := stmt.sGram.predicate.clauses
	if len(clauses) != 2 || len(clauses[0].objects) != 2 || clauses[0].objects[0].noun.binding.Record != memory["First"] || clauses[0].objects[1].noun.binding.Record != memory["Second"] {
		t.Fatalf("coordinated action grammar lost its two operands or following Show: %+v", clauses)
	}
	if _, err := paragraphGrammarStatement(t, "Launch First and Second and", memory); err == nil {
		t.Fatal("accepted a trailing conjunction")
	}
}

func TestParagraphGrammarJoinedClauseRollback(t *testing.T) {
	gram, err := loadGrammar(canonicalTestOptions())
	if err != nil {
		t.Fatal(err)
	}
	term, err := decodeGrammarTerm("[CLAUSE.ASSIGN.PASSIVE]*")
	if err != nil {
		t.Fatal(err)
	}
	tokens := grammarTestTokens(t, "and \"second\" is `right` and \"third\"", gram)
	original := append([]*token(nil), tokens...)
	clauses, separators, err := buildGrammarClauses(term, tokens, gram)
	if err != nil {
		t.Fatal(err)
	}
	if len(clauses) != 1 || len(separators) != 1 || separators[0] != tokens[0] || !reflect.DeepEqual(clauses[0].tokens, tokens[1:4]) || clauses[0].predicate == nil {
		t.Fatal("joined passive clause lost prefix, predicate or leading separator identity")
	}
	if !reflect.DeepEqual(tokens, original) {
		t.Fatal("joined matching changed the source token slice")
	}
	clauses, separators, err = buildGrammarClauses(term, tokens[4:], gram)
	if err == nil || clauses != nil || separators != nil {
		t.Fatal("failed joined clause consumed its parent's separator")
	}
}
