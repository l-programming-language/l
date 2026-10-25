package Lyger

import "testing"

func TestRefactorGrammarTermMetadata(t *testing.T) {
	for _, tc := range []struct {
		raw, name, kind, category, qualifier string
		multiple, optional, declared         bool
	}{
		{raw: "ID", name: "id", kind: "identifier", category: "id"},
		{raw: `"ID"`, name: "id", kind: "identifier", category: "id", declared: true},
		{raw: `["ID"]*`, name: "id", kind: "identifier", category: "id", multiple: true, optional: true, declared: true},
		{raw: "[ADVERB]*", name: "adverb", kind: "lexical", category: "adverb", multiple: true, optional: true},
		{raw: "VERB.REPORTING.ROOT", name: "verb.reporting.root", kind: "lexical", category: "verb", qualifier: "reporting.root"},
		{raw: "[CLAUSE.REPORTING.ACTIVE]*", name: "clause.reporting.active", kind: "grammar", category: "clause", multiple: true, optional: true},
		{raw: "OBJECT.CUSTOM_VALUE*", name: "object.custom_value", kind: "grammar", category: "object", optional: true},
		{raw: "CASE*", name: "case", kind: "case", category: "case", optional: true},
		{raw: "`~`", name: "`~`", kind: "type", category: "primitive.text"},
		{raw: "%~%", name: "%~%", kind: "type", category: "primitive.number"},
		{raw: "?~?", name: "?~?", kind: "type", category: "primitive.truth"},
		{raw: "!~!", name: "!~!", kind: "type", category: "primitive.null"},
		{raw: "#~#", name: "#~#", kind: "type", category: "type"},
		{raw: "[&~&]*", name: "&~&", kind: "type", category: "type.iterative.list", multiple: true, optional: true},
		{raw: "{~}", name: "{~}", kind: "type", category: "type.iterative.dictionary"},
	} {
		t.Run(tc.raw, func(t *testing.T) {
			term, err := decodeGrammarTerm(tc.raw)
			if err != nil || term.Name != tc.name {
				t.Fatalf("reference = %q, %v; want %q", term.Name, err, tc.name)
			}
			kind, category := term.Kind.String(), term.className()
			qualifier := ""
			if term.Kind == termLexical || term.Kind == termIdentifier {
				qualifier = term.Qualifier
			}
			if kind != tc.kind || category != tc.category || qualifier != tc.qualifier || term.Multiple != tc.multiple || term.Optional != tc.optional || term.Declared != tc.declared {
				t.Fatalf("metadata = kind:%q category:%q qualifier:%q repeated:%t optional:%t declared:%t; want %+v", kind, category, qualifier, term.Multiple, term.Optional, term.Declared, tc)
			}
		})
	}
}

func TestRefactorGrammarReferenceExactErrors(t *testing.T) {
	for _, tc := range []struct{ raw, want string }{
		{`"TEXT"`, `reference "\"TEXT\"": only ID can be declared`},
		{" [ID]*", `malformed grammar reference " [ID]*"`},
		{"ID *", `malformed grammar reference "ID *"`},
		{`"ID".NOUN`, `malformed grammar reference "\"ID\".NOUN"`},
	} {
		term, err := decodeGrammarTerm(tc.raw)
		if err == nil || err.Error() != tc.want || term.Name != "" {
			t.Errorf("reference %q = %q, %v; want empty name and %q", tc.raw, term.Name, err, tc.want)
		}
	}
	term, err := decodeGrammarTerm("FUTURE_ELEMENT")
	if err != nil || term.Name != "future_element" {
		t.Fatalf("syntactically valid unknown name = %q, %v", term.Name, err)
	}
	_, err = validateGrammarSequence(grammarValidationRules(), "FUTURE_ELEMENT", false)
	if err == nil || err.Error() != `reference "FUTURE_ELEMENT": unknown grammatical element` {
		t.Fatalf("unknown name lost contextual validation error: %v", err)
	}
}

func TestRefactorConfiguredGrammarNamesRemainOpen(t *testing.T) {
	rules := grammarValidationRules()
	object := rules["object"]
	object.Cases["custom_value"] = "ID"
	rules["object"] = object
	rules["clause.dispatch_v2.active"] = grammeme{Structure: "VERB.REPORTING.ROOT | OBJECT.CUSTOM_VALUE"}
	rules["predicate.active"] = grammeme{Structure: "CLAUSE.DISPATCH_V2.ACTIVE"}
	rules["statement.subjective.active"] = grammeme{Structure: "SUBJECT | PREDICATE.ACTIVE"}
	if err := validateGrammarReferences(rules); err != nil {
		t.Fatalf("configured category, lexical qualifier and case names were restricted: %v", err)
	}
	gram := &grammar{grammarMap: rules}
	got, err := gram.getGrammeme("clause.dispatch_v2.active")
	if err != nil || got.Structure != "VERB.REPORTING.ROOT | OBJECT.CUSTOM_VALUE" {
		t.Fatalf("configured rule lookup changed: %+v, %v", got, err)
	}
}

func TestRefactorPredicateCaseInitialization(t *testing.T) {
	for _, tc := range []struct{ name, rule, wantError string }{
		{name: "no case", rule: "predicate.active"},
		{name: "configured case", rule: "predicate.active.named"},
		{name: "missing case", rule: "predicate.active.missing", wantError: "case missing for: missing"},
		{name: "missing rule", rule: "predicate.absent.named", wantError: "grammeme type not found: predicate.absent"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			gram := &grammar{grammarMap: map[string]grammeme{"predicate.active": {Structure: "CASE", Cases: map[string]string{"named": "CLAUSE.REPORTING.ACTIVE"}}}}
			source := []*token{{rawBytes: []byte("show")}}
			got, err := initPredicateGram(tc.rule, source, gram)
			if tc.wantError != "" {
				if err == nil || err.Error() != tc.wantError {
					t.Fatalf("error = %v; want %q", err, tc.wantError)
				}
				if got.tokens != nil {
					t.Fatal("failed initialization published tokens")
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			if len(got.tokens) != 1 || got.tokens[0] != source[0] {
				t.Fatal("initialization lost source token identity")
			}
			if got.rule.Structure != "CASE" {
				t.Fatal("initialization changed configured rule")
			}
		})
	}
}
