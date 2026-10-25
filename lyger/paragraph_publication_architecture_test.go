package Lyger

import (
	"bytes"
	"strings"
	"testing"
)

func TestParagraphPublicationRejectsBatchWithoutPartialState(t *testing.T) {
	for _, failure := range []string{"foreign second token", "invalid second value", "duplicate second name"} {
		t.Run(failure, func(t *testing.T) {
			statement, err := paragraphGrammarStatement(t, `Publicly define "first" as 1 and define "second" as 2`, nil)
			if err != nil {
				t.Fatal(err)
			}
			span := SourceSpan{Path: "publication.l", Paragraph: 1, Start: SourcePosition{Offset: 4, Line: 1, Column: 5}}
			statement.source = &SourceStatement{Span: span, occurrence: 1}
			compilation := &Compilation{Memory: make(map[string]*record), Semantics: newSemanticProgram()}
			occurrences := []sourceOccurrence{{ID: 1, Path: span.Path, Paragraphs: []paragraphLayout{{Number: 1, FirstCode: span.Start, Encapsulated: true}}}}
			compilation.Scope, err = planSourceScopes([]*lStatement{statement}, compilation.Memory, occurrences)
			if err != nil {
				t.Fatal(err)
			}
			declarations, err := statement.sGram.getDeclarations()
			if err != nil || len(declarations) != 2 || declarations[0].Visibility != visibilityPublic {
				t.Fatalf("batch preparation = %+v, %v", declarations, err)
			}
			firstToken, secondToken := declarations[0].Token, declarations[1].Token
			local := compilation.Scope.declarations[firstToken].Scope
			var prior *record
			var previousStatement *lStatement
			wantError := ""
			switch failure {
			case "foreign second token":
				foreign := *secondToken
				declarations[1].Token = &foreign
				wantError = "declaration token"
			case "invalid second value":
				declarations[1].Record.Value = &entry{Type: storagePrimitive}
			case "duplicate second name":
				prior = &record{ID: "second", Type: "primitive.number", Value: &entry{Type: storagePrimitive, Primitive: newPrimitive(valueNumber)}, storageScope: local}
				previousStatement = &lStatement{source: &SourceStatement{Span: span}}
				local.records["second"] = prior
				compilation.Semantics.Declarations[prior] = previousStatement
				wantError = "duplicate declaration"
			}
			err = compilation.publishRegistration(statement, statementRegistration{declarations: declarations})
			if err == nil || (wantError != "" && !strings.Contains(err.Error(), wantError)) {
				t.Fatalf("invalid publication batch error = %v; want %q", err, wantError)
			}
			if len(compilation.Memory) != 0 || local.records["first"] != nil || local.records["second"] != prior {
				t.Errorf("rejected batch published root/local records: root=%v local=%v", compilation.Memory, local.records)
			}
			wantPrior := 0
			if prior != nil {
				wantPrior = 1
				if compilation.Semantics.Declarations[prior] != previousStatement {
					t.Error("rejected batch changed the existing declaration's provenance")
				}
			}
			if len(local.records) != wantPrior || len(compilation.Semantics.Subjects) != 0 || len(compilation.Semantics.Declarations) != wantPrior || len(statement.records) != 0 {
				t.Error("rejected batch published declaration relationships")
			}
			for _, token := range []*token{firstToken, secondToken} {
				if compilation.Scope.declarations[token].Record != nil {
					t.Error("rejected batch resolved a reserved declaration")
				}
			}
			if declarations[0].Record.storageScope != nil {
				t.Error("rejected batch mutated the unpublished first record's storage scope")
			}
		})
	}
}

// Stop at semantic publication so these checks are independent of activation
// lowering and generated executables. All source still uses the normal pipeline.
func paragraphPublicationSource(t *testing.T, source string) *Compilation {
	t.Helper()
	root := t.TempDir()

	path := writeSourceFile(t, root, "main.l", source)
	compilation := newCompilation([]byte(source), path)
	compilation.init(canonicalTestOptions())
	for _, phase := range []func(){compilation.preProcess, compilation.build, compilation.scan, compilation.analyze} {
		if compilation.Error != nil {
			break
		}
		phase()
	}
	if !bytes.Equal(compilation.MainBytes, []byte(source)) {
		t.Fatal("publication changed original source")
	}
	return compilation
}

func TestParagraphPublicationPublicMemberKeepsPrivateOwner(t *testing.T) {
	for _, memberDeclaration := range []string{
		`Create Hidden's "Exported" as a Public Dictionary.`,
		`Publicly create Hidden's "Exported" as a Dictionary.`,
	} {
		t.Run(memberDeclaration, func(t *testing.T) {
			source := "    Create \"Hidden\" as a Dictionary.\n" + memberDeclaration + "\n" +
				"\"PrivateSibling\" is `private`.\n\nUse Hidden.Exported as \"Alias\".\n"
			compilation := paragraphPublicationSource(t, source)
			if compilation.Error != nil {
				t.Fatal(compilation.Error)
			}
			member, alias := compilation.Memory["Hidden.Exported"], compilation.Memory["Alias"]
			if member == nil || alias == nil {
				t.Fatal("Public member did not publish its qualified binding for the outer alias")
			}
			canonical, err := canonicalRecord(alias, compilation.Semantics.Aliases)
			if err != nil || canonical != member {
				t.Fatalf("outer member alias lost canonical identity: %v", err)
			}
			var owner *record
			for _, statement := range compilation.Statements {
				for _, tok := range statement.tokens {
					if tok.declarationRoot() == "hidden" {
						owner = compilation.Semantics.Subjects[tok]
					}
				}
			}
			if owner == nil || owner.members["Hidden.Exported"] != member || owner.Value.Iterative.Dictionary.Values["Exported"] != member.Value.Iterative.Dictionary {
				t.Fatal("Public member stopped addressing the private owner's canonical namespace")
			}
			if owner.storageScope.records["Hidden.Exported"] != member || member.storageScope != compilation.Scope.Root {
				t.Fatal("Public member lost its local binding or global storage scope")
			}
			if alias.Value.Iterative.Dictionary != member.Value.Iterative.Dictionary || member.Value.Iterative.Dictionary.ID != "Hidden.Exported" {
				t.Fatal("outer alias cloned or renamed the published dictionary member")
			}
			if compilation.Memory["Hidden"] != nil || compilation.Memory["PrivateSibling"] != nil {
				t.Fatal("Public member published its private owner or sibling")
			}
		})
	}
}

func TestParagraphPublicationRejectedAliasKeepsExistingGraph(t *testing.T) {
	const source = "\"origin\" is 1.\nPublicly use origin as \"Alias\".\nPublicly use origin as \"Alias\".\n"
	compilation := paragraphPublicationSource(t, source)
	if compilation.Error == nil || !strings.Contains(compilation.Error.Error(), "duplicate declaration Alias") {
		t.Fatalf("duplicate public alias error = %v", compilation.Error)
	}
	origin, alias := compilation.Memory["origin"], compilation.Memory["Alias"]
	if origin == nil || alias == nil || len(compilation.Semantics.Aliases) != 1 || compilation.Semantics.Aliases[alias] != origin {
		t.Fatalf("rejected alias changed the existing alias graph: %v", compilation.Semantics.Aliases)
	}
	subjects, declarations, _ := testProgramSemanticCounts(compilation)
	if subjects != 2 || declarations != 2 || compilation.Semantics.Declarations[alias].source.Span.Start.Line != 2 {
		t.Fatal("rejected alias published or replaced declaration provenance")
	}
	for _, statement := range compilation.Statements {
		if statement.source.Unit.Kind != SourceProgram || statement.source.Span.Start.Line != 3 {
			continue
		}
		for _, tok := range statement.tokens {
			if tok.isDeclare() && (compilation.Semantics.Subjects[tok] != nil || compilation.Scope.declarations[tok].Record != nil) {
				t.Fatal("rejected alias resolved its reserved declaration")
			}
		}
		if len(statement.records) != 0 {
			t.Fatal("rejected alias appended statement records")
		}
	}
}
