package Lyger

import (
	"strings"
	"testing"
)

func captureTestAction(name string) *record {
	return &record{ID: name, Value: &entry{Type: storageLexeme, Lexeme: newLexeme(name)}}
}
func captureTestBind(action *record, values ...*record) *instruction {
	result := &instruction{Kind: instructionBindAction, Callable: action, Target: &instructionValue{Kind: operandSymbol, Type: valueLexeme, Record: action}}
	for _, value := range values {
		result.Captures = append(result.Captures, planTestRef(value))
	}
	return result
}
func captureTestCall(action *record, body ...*instruction) *instruction {
	return &instruction{Kind: instructionCall, Callable: action, Synchronous: true, Actions: body}
}
func captureTestLiteral(value string) *instructionValue {
	return &instructionValue{Kind: operandLiteral, Type: valueText, Literal: &primitive{Type: valueText, Text: "`" + value + "`"}}
}
func captureTestSet(rec *record, value string) *instruction {
	return &instruction{Kind: instructionAssign, Target: planTestRef(rec), Value: captureTestLiteral(value)}
}
func captureTestShow(rec *record) *instruction {
	return &instruction{Kind: instructionShow, Value: planTestRef(rec)}
}

func TestCapturePlanDefiniteBindingAndInitialization(t *testing.T) {
	for _, scenario := range []string{"before", "missing", "both", "exit", "unassigned", "alias-before"} {
		t.Run(scenario, func(t *testing.T) {
			value := planTestScalar("saved", valueText, nil)
			action := captureTestAction("Report")
			alias := captureTestAction("Relay")
			semantics := newSemanticProgram()
			semantics.Aliases[alias] = action
			set, bind, call := captureTestSet(value, "value"), captureTestBind(action, value), captureTestCall(action, captureTestShow(value))
			branch := &instruction{Kind: instructionIf, Condition: &instructionCondition{Operator: "truth", Left: &instructionValue{Kind: operandLiteral, Type: valueTruth, Literal: &primitive{Type: valueTruth, Truth: true}}}, Then: []*instruction{bind}}
			switch scenario {
			case "before":
				semantics.Instructions = []*instruction{set, call, bind}
			case "missing":
				semantics.Instructions = []*instruction{set, branch, call}
			case "both":
				branch.Else = []*instruction{bind}
				semantics.Instructions = []*instruction{set, branch, call}
			case "exit":
				branch.Else = []*instruction{{Kind: instructionExit}}
				semantics.Instructions = []*instruction{set, branch, call}
			case "unassigned":
				semantics.Instructions = []*instruction{bind, set, call}
			case "alias-before":
				semantics.Instructions = []*instruction{{Kind: instructionAliasAction, Callable: alias, Target: &instructionValue{Kind: operandSymbol, Record: action}}, set, bind}
			}
			_, err := planOrchestration(semantics)
			if scenario == "both" || scenario == "exit" {
				if err != nil {
					t.Fatal(err)
				}
				return
			}
			want := "not definitely bound"
			if scenario == "unassigned" {
				want = "not definitely assigned"
			}
			if err == nil || !strings.Contains(err.Error(), want) {
				t.Fatalf("%s: %v", scenario, err)
			}
		})
	}
}

func visitCapturePlan(body []*orchestrationInstruction, visit func(*orchestrationInstruction)) {
	for _, current := range body {
		visit(current)
		for _, nested := range [][]*orchestrationInstruction{current.Actions, current.Then, current.Else} {
			visitCapturePlan(nested, visit)
		}
		if current.Block != nil {
			visitCapturePlan(current.Block.Body, visit)
		}
		if current.Continuation != nil {
			visitCapturePlan(current.Continuation.Body, visit)
		}
	}
}
