package Lyger

import (
	"bytes"
	"os"
	"reflect"
	"strings"
	"testing"
)

func scopeTestStatements(t *testing.T, path, source string, gram *grammar) []*lStatement {
	t.Helper()
	prepared, err := PrepareSource(path, []byte(source))
	if err != nil {
		t.Fatal(err)
	}
	return scopePreparedStatements(t, prepared.Statements, gram)
}

func scopePreparedStatements(t *testing.T, prepared []SourceStatement, gram *grammar) []*lStatement {
	t.Helper()
	var statements []*lStatement
	for index := range prepared {
		parser := newParser(prepared[index].Prepared)
		raw := parser.parseStatements()
		if len(raw) != 1 {
			t.Fatalf("prepared statement split into %d", len(raw))
		}
		statement, err := newLStatement(raw[0], gram)
		if err != nil {
			t.Fatal(err)
		}
		if err = statement.tokenize(); err != nil {
			t.Fatal(err)
		}
		statement.source = &prepared[index]
		statements = append(statements, statement)
	}
	return statements
}

func scopeTestGrammar(t testing.TB) *grammar {
	t.Helper()
	gram, err := loadGrammar(canonicalTestOptions())
	if err != nil {
		t.Fatal(err)
	}
	return gram
}

func TestScopePlansIncomingRHSBeforePublishingDeclaration(t *testing.T) {
	statements := scopeTestStatements(t, "main.l", "Create \"Worker\" as a New Basic App.\nExecute it.\nIf it has an Error, exit with the Error, else define \"result\" as it's Output.\nShow it.\n", scopeTestGrammar(t))
	var before [][]byte
	var beforeTypes []string
	for _, statement := range statements {
		for _, token := range statement.tokens {
			before = append(before, bytes.Clone(token.rawBytes))
			beforeTypes = append(beforeTypes, token.tType)
		}
	}
	plan, err := planSourceScopes(statements, nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	if plan.Ordered[2].IncomingSubjectID != "worker" || plan.Ordered[2].DeclaredSubjectID != "result" || plan.Ordered[3].IncomingSubjectID != "result" {
		t.Fatalf("conditional declaration context = %+v", plan.Ordered)
	}
	for index, statement := range statements {
		for _, token := range statement.tokens {
			if binding, exists := plan.Pronouns[token]; exists {
				want := "worker"
				if index == 3 {
					want = "result"
				}
				if binding.SubjectID != want || binding.Kind != scopeParagraph {
					t.Fatalf("%q bound to %+v", token.rawBytes, binding)
				}
				if string(token.rawBytes) == "it's" && !binding.Possessive {
					t.Fatal("possessive pronoun lost member-access role")
				}
			}
		}
	}
	var after [][]byte
	var afterTypes []string
	for _, statement := range statements {
		for _, token := range statement.tokens {
			after, afterTypes = append(after, token.rawBytes), append(afterTypes, token.tType)
		}
	}
	if !reflect.DeepEqual(before, after) || !reflect.DeepEqual(beforeTypes, afterTypes) {
		t.Fatal("scope planning mutated lexical token roles or contents")
	}
	originalFirst := statements[0]
	statements[0], statements[3] = statements[3], statements[0]
	if plan.Ordered[0].Statement != originalFirst || plan.Ordered[0].Order != 0 || plan.Ordered[3].Order != 3 {
		t.Fatal("retry queue order changed the planned instruction order")
	}
}

func TestScopeKeepsDescriptorRefinementsOnOriginalSubject(t *testing.T) {
	statements := scopeTestStatements(t, "descriptors.l", "\"Primary\" is a New Adjective describing any Service, Object, or Function.\nIt describes a Generic `entry` Function of any Process with No Inputs and No Outputs.\nIt describes a Generic `entry` Object with a New Objective Primary Function using any Process.\n\"Simple\" describes a Generic App with a New Primary Service using any Process.\nIt describes a Generic App.\n", scopeTestGrammar(t))
	plan, err := planSourceScopes(statements, nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	for _, index := range []int{1, 2, 4} {
		want := "primary"
		if index == 4 {
			want = "simple"
		}
		if got := plan.Pronouns[statements[index].tokens[0]]; got.SubjectID != want || got.Kind != scopeParagraph {
			t.Fatalf("refinement %d binding = %+v", index, got)
		}
		if plan.Ordered[index].DeclaredSubjectID != "" {
			t.Fatal("nested generic description replaced its enclosing descriptor")
		}
	}
}

func TestScopeSeparatesGenericActionArgumentsFromParagraphSubjects(t *testing.T) {
	statements := scopeTestStatements(t, "actions.l", "\"prior\" is `data`.\nCreate \"Dispatch|-es|-ed|-ing\" for Building any Object and then Running it.\nShow it.\n", scopeTestGrammar(t))
	plan, err := planSourceScopes(statements, nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	action := statements[1]
	local := plan.Pronouns[action.tokens[len(action.tokens)-1]]
	if local.Kind != scopeActionArgument || local.SubjectID != "dispatch::argument:5" || local.TargetToken != action.tokens[5] || local.TargetTokenIndex != 5 || local.Declaration != action {
		t.Fatalf("directive pronoun binding = %+v", local)
	}
	if string(local.TargetToken.rawBytes) != "Object" || plan.Ordered[1].IncomingSubjectID != "prior" || plan.Ordered[1].DeclaredSubjectID != "dispatch" {
		t.Fatal("action argument was confused with its enclosing declaration")
	}
	paragraph := plan.Pronouns[statements[2].tokens[1]]
	if paragraph.Kind != scopeParagraph || paragraph.SubjectID != "dispatch" {
		t.Fatalf("local argument leaked into following source: %+v", paragraph)
	}
}

func TestScopeUsesFileAndParagraphIdentityAcrossImports(t *testing.T) {
	gram := scopeTestGrammar(t)
	main := scopeTestStatements(t, "main.l", "\"Outer\" is `outer`.\nShow it.\n", gram)
	library := scopeTestStatements(t, "library.l", "\"Inner\" is `inner`.\nShow it.\n", gram)
	plan, err := planSourceScopes([]*lStatement{main[0], library[0], library[1], main[1]}, nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	if plan.Pronouns[main[1].tokens[1]].SubjectID != "outer" || plan.Pronouns[library[1].tokens[1]].SubjectID != "inner" {
		t.Fatal("imported statements changed the caller's paragraph subject")
	}
	unbound := scopeTestStatements(t, "main.l", "\"Outer\" is `outer`.\n\nShow it.\n", gram)
	if _, err = planSourceScopes(unbound, nil, nil); err == nil || !strings.Contains(err.Error(), "main.l:3:1:") || !strings.Contains(err.Error(), "source paragraph 2") {
		t.Fatalf("cross-paragraph pronoun should be unbound: %v", err)
	}
}

func TestScopeQualifiesDeclaredMemberRoot(t *testing.T) {
	statements := scopeTestStatements(t, "library.l", "Create \"Library\" as a Dictionary.\nCreate Library's \"Render|-s|-ed|-ing\" for Showing any Text.\nShow it.\n", scopeTestGrammar(t))
	plan, err := planSourceScopes(statements, nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	if plan.Ordered[1].DeclaredSubjectID != "library.render" || plan.Pronouns[statements[2].tokens[1]].SubjectID != "library.render" {
		t.Fatalf("member declaration identity = %+v", plan.Ordered[1])
	}
}

func TestScopePlansActualExpandedFixture(t *testing.T) {
	raw, err := os.ReadFile("test/main.l")
	if err != nil {
		t.Fatal(err)
	}
	prepared, err := PrepareSource("test/main.l", raw)
	if err != nil {
		t.Fatal(err)
	}
	gram := scopeTestGrammar(t)
	built := newPreparedBuilder(prepared, gram)
	if err = built.process(); err != nil {
		t.Fatal(err)
	}
	statements := scopePreparedStatements(t, built.SourceStatements, gram)
	plan, err := planSourceScopes(statements, nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	counts := make(map[string]int)
	for _, binding := range plan.Pronouns {
		counts[binding.SubjectID]++
	}
	if len(plan.Ordered) != len(built.SourceStatements) || counts["execute::argument:5"] != 1 || counts["main"] != 3 || counts["helloworldapp"] != 3 {
		t.Fatalf("expanded fixture scope counts = %v across %d statements", counts, len(plan.Ordered))
	}
}

func TestScopeRejectsMissingProvenanceAndAmbiguousActionArguments(t *testing.T) {
	if _, err := planSourceScopes([]*lStatement{{}}, nil, nil); err == nil {
		t.Fatal("missing original source silently acquired a paragraph")
	}
	statements := scopeTestStatements(t, "actions.l", "Create \"Combine\" for Building any App and Running any Service and then Running it.\n", scopeTestGrammar(t))
	if _, err := planSourceScopes(statements, nil, nil); err == nil || !strings.Contains(err.Error(), "more than one preceding generic action argument") {
		t.Fatalf("ambiguous action argument = %v", err)
	}
}
