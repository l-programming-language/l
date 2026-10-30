package Lyger

import "testing"

func TestRefactorTokenInspectionPreservesLexicalRoles(t *testing.T) {
	var absent *token
	if absent.lexeme() != nil || absent.hasRole("noun", "pronoun.personal") || absent.isPronoun() || absent.declarationRoot() != "" {
		t.Fatal("absent token acquired lexical information")
	}
	lex := &lexeme{Root: "renamed", POS: "noun", Type: "pronoun.personal"}
	tok := &token{mem: &entry{Type: storageLexeme, Lexeme: lex}}
	if tok.lexeme() != lex || !tok.hasRole("noun", "pronoun.personal") || !tok.isPronoun() || tok.hasRole("verb", "pronoun.personal") {
		t.Fatal("lexical inspection lost identity or conflated POS with role")
	}
	lex.Type = "source_defined_category"
	if tok.isPronoun() || !tok.hasRole("noun", "source_defined_category") {
		t.Fatal("open lexical category changed")
	}
}

func TestRefactorInstructionContextAttribution(t *testing.T) {
	span := SourceSpan{Path: "library.l", Start: SourcePosition{Line: 7}}
	stmt := &lStatement{source: &SourceStatement{Span: span}}
	callSite := &lStatement{}
	ctx := instructionContext{statement: stmt, order: 12, callSite: callSite}
	got := ctx.newInstruction(instructionShow)
	if got.Kind.String() != "show" || got.Source != span || got.Statement != stmt || got.Order != 12 || got.CallSite != callSite {
		t.Fatalf("instruction lost definition/call-site attribution: %+v", got)
	}
}

func TestRefactorDescriptorMergeRetainsOrderedIndependentRelations(t *testing.T) {
	first, second := &constructionNode{Name: "first"}, &constructionNode{Name: "second"}
	left, right := &descriptorRule{}, &descriptorRule{}
	parameter := &objectGram{}
	definition := &record{ID: "sourceDefinition"}
	dst := &constructionNode{Nominal: "sourceNominal", Name: "original", Children: []*constructionNode{first}, Wrappers: []*descriptorRule{left}, InputsDefined: true}
	src := &constructionNode{Name: "replacement", Definition: definition, Children: []*constructionNode{second}, Wrappers: []*descriptorRule{right}, Parameters: []*objectGram{parameter}, OutputsDefined: true}
	dst.mergeDescriptor(src)
	if dst.Nominal != "sourceNominal" || dst.Name != "replacement" || dst.Definition != definition || !dst.InputsDefined || !dst.OutputsDefined {
		t.Fatal("descriptor merge changed nominal identity or lost independent flags")
	}
	if len(dst.Children) != 2 || dst.Children[0] != first || dst.Children[1] != second || len(dst.Wrappers) != 2 || dst.Wrappers[0] != left || dst.Wrappers[1] != right || len(dst.Parameters) != 1 || dst.Parameters[0] != parameter {
		t.Fatal("descriptor merge reordered additive relations")
	}
	if src.InputsDefined || len(src.Children) != 1 || src.Children[0] != second || len(src.Wrappers) != 1 {
		t.Fatal("merge mutated the descriptor source")
	}
}
