package Lyger

import (
	"strings"
	"testing"
)

func TestDeclarationRegistrationRejectsForeignToken(t *testing.T) {
	statement := extractionTestStatement(t, `"Owned" is 42`)
	original := statement.sGram.subject.noun
	foreign := *original
	statement.sGram.subject.noun = &foreign
	compilation := &Compilation{Memory: make(map[string]*record), Semantics: newSemanticProgram()}
	err := compilation.registerStatement(statement)
	if err == nil || !strings.Contains(err.Error(), "declaration token") {
		t.Fatalf("foreign declaration token accepted: %v", err)
	}
	if len(compilation.Semantics.Subjects) != 0 || len(compilation.Semantics.Declarations) != 0 || len(statement.records) != 0 {
		t.Fatal("rejected extraction published declaration relationships")
	}
}

func TestDeclarationRegistrationRetainsOriginalToken(t *testing.T) {
	for _, source := range []string{`"Owned" is 42`, `Create "Owned" as an App`} {
		t.Run(source, func(t *testing.T) {
			statement := extractionTestStatement(t, source)
			var declared *token
			for _, token := range statement.tokens {
				if token.isDeclare() {
					declared = token
				}
			}
			compilation := &Compilation{Memory: make(map[string]*record), Semantics: newSemanticProgram()}
			statement.source = &SourceStatement{Span: SourceSpan{Path: "registration.l", Paragraph: 1, Start: SourcePosition{Line: 1, Column: 1}}}
			var err error
			compilation.Scope, err = planSourceScopes([]*lStatement{statement}, compilation.Memory, nil)
			if err != nil {
				t.Fatal(err)
			}
			if err := compilation.registerStatement(statement); err != nil {
				t.Fatal(err)
			}
			if len(statement.records) != 1 {
				t.Fatalf("records = %d", len(statement.records))
			}
			record := statement.records[0]
			if record.ID != "Owned" || compilation.Semantics.Subjects[declared] != record || compilation.Semantics.Declarations[record] != statement {
				t.Fatal("declaration lost its original token, record, or statement")
			}
		})
	}
}
