package Lyger

import (
	"strings"
	"testing"
)

func instructionRecord(id, kind, text string) *record {
	if kind == "App" {
		return &record{ID: id, Type: "applicative.app", Value: &entry{Type: storageApplicative, Applicative: &applicative{Type: valueApp, App: &App{ID: id}}}}
	}
	valueKind, ok := parseValueType(kind)
	if !ok {
		panic("invalid test value type: " + kind)
	}
	return &record{ID: id, Type: "primitive." + strings.ToLower(kind), Value: &entry{Type: storagePrimitive, Primitive: &primitive{Type: valueKind, Text: text}}}
}

func instructionAction(t *testing.T, label string) *record {
	t.Helper()
	lex := newLexeme("verb")
	if err := lex.reLabel(label); err != nil {
		t.Fatal(err)
	}
	return &record{ID: lex.Root, Type: "lexeme.verb.generic", Value: &entry{Type: storageLexeme, Lexeme: lex}}
}

func instructionTestGrammar(t *testing.T, source string, memory map[string]*record) ([]*lStatement, *scopePlan, *semanticProgram) {
	t.Helper()
	prepared, err := PrepareSource("renamed.l", []byte(source))
	if err != nil {
		t.Fatal(err)
	}
	statements := scopePreparedStatements(t, prepared.Statements, scopeTestGrammar(t))
	scopes, err := planSourceScopes(statements, memory, nil)
	if err != nil {
		t.Fatal(err)
	}
	semantics := newSemanticProgram()
	for _, ordered := range scopes.Ordered {
		for _, declaration := range ordered.Declarations {
			binding, err := resolveTokenBinding([]byte(declaration.SubjectID), memory)
			if err != nil {
				t.Fatal(err)
			}
			if binding != nil {
				semantics.Subjects[declaration.Token] = binding.Record
			}
		}
	}
	for _, statement := range statements {
		for _, tok := range statement.tokens {
			if tok.getPOS(false) == "unknown" {
				state, err := (semanticResolver{scopes: scopes, semantics: semantics}).bindToken(tok)
				if err != nil {
					t.Fatal(err)
				}
				if state != lookupFound && state != lookupAbsent {
					t.Fatalf("unexpected synthetic binding state: %s", state)
				}
			}
		}
		statement.determineType()
		if err = statement.loadGram(); err != nil {
			t.Fatalf("grammar %s: %v", statement.stmtBytes, err)
		}
		if err = statement.uncacheLiterals(prepared.TextCache); err != nil {
			t.Fatal(err)
		}
	}
	return statements, scopes, semantics
}

func TestInstructionsPreserveRuntimeBranchesAndCustomActionOrder(t *testing.T) {
	action := instructionAction(t, `"Launch|-es|-ed|-ing"`)
	job := instructionRecord("Worker", "App", "")
	message := instructionRecord("message", "Text", "`unchanged\ntext`")
	result := instructionRecord("response", "Text", "")
	memory := map[string]*record{action.ID: action, job.ID: job, message.ID: message, result.ID: result}
	source := "Create \"Launch|-es|-ed|-ing\" for Building any App and then Running it.\nCreate \"Worker\" as a New App.\nLaunch it.\nIf it has an Error, exit with the Error, else define \"response\" as it's Output.\nIf message isnt response, exit with an error of `unaltered failure\nmessage`.\nOtherwise, show response and then exit.\n"
	statements, scopes, semantics := instructionTestGrammar(t, source, memory)
	semantics.Declarations = map[*record]*lStatement{action: statements[0], job: statements[1], result: statements[3]}
	instructions, err := buildInstructions(statements, scopes, semantics)
	if err != nil {
		t.Fatal(err)
	}
	if len(instructions) != 3 || instructions[0].Kind.String() != "call" || instructions[1].Kind.String() != "if" || instructions[2].Kind.String() != "if" {
		t.Fatalf("outer instruction sequence = %+v", instructions)
	}
	call := instructions[0]
	if call.Definition != statements[0] || call.Statement != statements[2] || !call.Synchronous || len(call.Actions) != 2 || call.Actions[0].Kind.String() != "build" || call.Actions[1].Kind.String() != "run" {
		t.Fatalf("custom action expansion = %+v", call)
	}
	for _, expanded := range call.Actions {
		if expanded.Target.Record != job || !expanded.Synchronous || expanded.Source.Start.Line != 1 || expanded.CallSite != statements[2] || expanded.Order != call.Order {
			t.Fatalf("expanded action provenance/argument = %+v", expanded)
		}
	}
	failure := instructions[1]
	if failure.Condition.Operator != "has" || failure.Condition.Left.Record != job || failure.Condition.Right.Kind.String() != "member" || failure.Condition.Right.Member.String() != "error" {
		t.Fatalf("error condition = %+v", failure.Condition)
	}
	if len(failure.Then) != 1 || failure.Then[0].Kind.String() != "exit" || failure.Then[0].Value.Owner.Record != job || failure.Then[0].Value.Member.String() != "error" {
		t.Fatalf("error branch = %+v", failure.Then)
	}
	if len(failure.Else) != 1 || failure.Else[0].Kind.String() != "assign" || failure.Else[0].Target.Record != result || failure.Else[0].Value.Kind.String() != "member" || failure.Else[0].Value.Member.String() != "output" || failure.Else[0].Value.Type.String() != "Text" || failure.Else[0].Value.Owner.Record != job {
		t.Fatalf("conditional output assignment = %+v", failure.Else)
	}
	comparison := instructions[2]
	if comparison.Condition.Operator != "isnt" || comparison.Condition.Left.Kind.String() != "symbol" || comparison.Condition.Left.Record != message || comparison.Condition.Right.Record != result || !comparison.Condition.Right.Runtime {
		t.Fatalf("comparison lost symbol/runtime identity: %+v", comparison.Condition)
	}
	if len(comparison.Then) != 1 || comparison.Then[0].Value.Kind.String() != "error" || comparison.Then[0].Value.Payload.Literal.Text != "`unaltered failure\nmessage`" {
		t.Fatalf("error payload changed: %+v", comparison.Then)
	}
	if len(comparison.Else) != 2 || comparison.Else[0].Kind.String() != "show" || comparison.Else[0].Value.Record != result || comparison.Else[1].Kind.String() != "exit" || comparison.Else[1].Value != nil || comparison.ElseStatement != statements[5] {
		t.Fatalf("Otherwise association/sequence = %+v", comparison.Else)
	}
	if result.Value.Primitive.Text != "" || message.Value.Primitive.Text != "`unchanged\ntext`" {
		t.Fatal("instruction planning executed or rewrote a value")
	}
}

func TestInstructionsUseSourceActionAliasesAndIgnoreRetryOrder(t *testing.T) {
	action := instructionAction(t, `"Prepare|-s|-d|-ing"`)
	alias := instructionAction(t, `"Start|-s|-ed|-ing"`)
	job := instructionRecord("Other", "App", "")
	memory := map[string]*record{action.ID: action, alias.ID: alias, job.ID: job}
	source := "Create \"Prepare|-s|-d|-ing\" for Building any App and then Running it.\nCreate \"Other\" as a New App.\nStart it.\n"
	statements, scopes, semantics := instructionTestGrammar(t, source, memory)
	semantics.Declarations = map[*record]*lStatement{action: statements[0], job: statements[1]}
	semantics.Aliases = map[*record]*record{alias: action}
	reversed := []*lStatement{statements[2], statements[1], statements[0]}
	instructions, err := buildInstructions(reversed, scopes, semantics)
	if err != nil {
		t.Fatal(err)
	}
	if len(instructions) != 1 || instructions[0].Target.Record != action || instructions[0].Definition != statements[0] || instructions[0].Actions[0].Target.Record != job {
		t.Fatalf("aliased action was not expanded from source: %+v", instructions)
	}
	semantics.Aliases = map[*record]*record{alias: action, action: alias}
	if _, err = buildInstructions(statements, scopes, semantics); err == nil || !strings.Contains(err.Error(), "alias cycle") {
		t.Fatalf("action alias cycle = %v", err)
	}
}

func TestInstructionsRejectOtherwiseWithoutAnAdjacentConditional(t *testing.T) {
	memory := map[string]*record{"value": instructionRecord("value", "Text", "`x`")}
	for _, source := range []string{
		"Otherwise, show value and then exit.\n",
		"If value isnt value, exit.\nShow value.\nOtherwise, exit.\n",
		"If value isnt value, exit.\n\nOtherwise, exit.\n",
	} {
		statements, scopes, semantics := instructionTestGrammar(t, source, memory)
		if _, err := buildInstructions(statements, scopes, semantics); err == nil || !strings.Contains(err.Error(), "otherwise has no preceding unmatched conditional") {
			t.Errorf("invalid Otherwise attachment returned %v", err)
		}
	}
}

func TestInstructionsKeepAppBodyCallsSeparateFromImperativeCalls(t *testing.T) {
	action := instructionAction(t, `"Launch|-es|-ed|-ing"`)
	job := instructionRecord("Worker", "App", "")
	memory := map[string]*record{action.ID: action, job.ID: job}
	source := "Create \"Launch|-es|-ed|-ing\" for Building any App and then Running it.\nCreate \"Worker\" as a New App.\nThe App Launches it.\nLaunch it.\n"
	statements, scopes, semantics := instructionTestGrammar(t, source, memory)
	semantics.Declarations = map[*record]*lStatement{action: statements[0], job: statements[1]}
	instructions, err := buildInstructions(statements, scopes, semantics)
	if err != nil {
		t.Fatal(err)
	}
	if len(instructions) != 1 || instructions[0].Statement != statements[3] {
		t.Fatalf("application-body call became an outer instruction: %+v", instructions)
	}
}

func TestInstructionsRejectIncompleteInputsAndWrongActionArgument(t *testing.T) {
	action := instructionAction(t, `"Launch|-es|-ed|-ing"`)
	value := instructionRecord("message", "Text", "`message`")
	memory := map[string]*record{action.ID: action, value.ID: value}
	source := "Create \"Launch|-es|-ed|-ing\" for Building any App and then Running it.\nLaunch message.\n"
	statements, scopes, semantics := instructionTestGrammar(t, source, memory)
	declarations := map[*record]*lStatement{action: statements[0]}
	semantics.Declarations = declarations
	if _, err := buildInstructions(statements, scopes, semantics); err == nil || !strings.Contains(err.Error(), "requires app, got Text") {
		t.Fatalf("wrong action argument = %v", err)
	}
	semantics.Declarations = nil
	if _, err := buildInstructions(statements, scopes, semantics); err == nil || !strings.Contains(err.Error(), "no resolved source declaration") {
		t.Fatalf("missing action source = %v", err)
	}
	semantics.Declarations = declarations
	if _, err := buildInstructions(statements[:1], scopes, semantics); err == nil {
		t.Fatal("partial source silently omitted a scoped instruction")
	}
	statements[1].sGram = nil
	if _, err := buildInstructions(statements, scopes, semantics); err == nil || !strings.Contains(err.Error(), "requires successful statement grammar") {
		t.Fatalf("unbuilt grammar = %v", err)
	}
}

func TestInstructionsRejectRecursiveSourceAction(t *testing.T) {
	action := instructionAction(t, `"Repeat|-s|-ed|-ing"`)
	job := instructionRecord("Worker", "App", "")
	memory := map[string]*record{action.ID: action, job.ID: job}
	source := "Create \"Repeat|-s|-ed|-ing\" for Building any App and then Repeating it.\nCreate \"Worker\" as a New App.\nRepeat it.\n"
	statements, scopes, semantics := instructionTestGrammar(t, source, memory)
	semantics.Declarations = map[*record]*lStatement{action: statements[0]}
	if _, err := buildInstructions(statements, scopes, semantics); err == nil || !strings.Contains(err.Error(), "recursive source action") {
		t.Fatalf("recursive source action = %v", err)
	}
}
