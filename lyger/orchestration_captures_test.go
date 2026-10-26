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

func TestCapturePlanSnapshotsAndAliasesRetainInstance(t *testing.T) {
	scope := &lexicalScope{kind: lexicalParagraph}
	value := planTestScalar("private", valueText, scope)
	action := captureTestAction("Report")
	alias := captureTestAction("Saved")
	semantics := newSemanticProgram()
	semantics.Aliases[alias] = action
	bind := captureTestBind(action, value)
	first := &instruction{Kind: instructionBlock, Block: &scopedInstructions{Scope: scope, Body: []*instruction{captureTestSet(value, "first"), bind, captureTestSet(value, "changed"), captureTestCall(action, captureTestShow(value))}}}
	second := &instruction{Kind: instructionBlock, Block: &scopedInstructions{Scope: scope, Body: []*instruction{captureTestSet(value, "second"), bind}}}
	semantics.Instructions = []*instruction{first, {Kind: instructionAliasAction, Callable: alias, Target: &instructionValue{Kind: operandSymbol, Record: action}}, second, captureTestCall(alias, captureTestShow(value)), captureTestCall(action, captureTestShow(value))}
	plan, err := planOrchestration(semantics)
	if err != nil {
		t.Fatal(err)
	}
	a, b := plan.Root.Body[0].Block, plan.Root.Body[2].Block
	if a.Body[1].CaptureValues[0].Storage == b.Body[1].CaptureValues[0].Storage {
		t.Fatal("two activations share private source storage")
	}
	if a.Body[3].Actions[0].Value.Storage.Capture == nil {
		t.Fatal("active call reads current source storage instead of declaration snapshot")
	}
	if plan.Root.Body[1].Callable == plan.Root.Body[1].CaptureSource.Slot {
		t.Fatal("alias shares a mutable name binding slot")
	}
	if plan.Root.Body[3].CaptureSource.Slot != plan.Root.Body[1].Callable || plan.Root.Body[4].CaptureSource.Slot != a.Body[1].Callable {
		t.Fatal("call selected the canonical name instead of its own alias binding")
	}
	if a.Body[1].Callable != b.Body[1].Callable {
		t.Fatal("public name cannot be rebound by a later activation")
	}
	emitter := &orchestrationEmitter{plan: plan}
	emitted, err := emitter.block(plan.Root, false)
	if err != nil {
		t.Fatal(err)
	}
	source := strings.Join(emitted, "\n")
	if strings.Count(source, "LEnvironment::typed(") != 2 || !strings.Contains(source, ".get(0)?") || !strings.Contains(source, ".clone()") {
		t.Fatalf("missing independent environment allocation/selection: %s", source)
	}
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

func TestCapturePlanWritesStayInEnvironmentAndNestedCallsKeepHandle(t *testing.T) {
	value := planTestScalar("saved", valueText, nil)
	inner, outer := captureTestAction("Inner"), captureTestAction("Outer")
	semantics := newSemanticProgram()
	bindOuter := captureTestBind(outer)
	bindOuter.CaptureCalls = []*record{inner}
	semantics.Instructions = []*instruction{captureTestSet(value, "original"), captureTestBind(inner, value), bindOuter,
		captureTestCall(outer, captureTestCall(inner, captureTestSet(value, "closure"), captureTestShow(value))), captureTestShow(value)}
	plan, err := planOrchestration(semantics)
	if err != nil {
		t.Fatal(err)
	}
	outerCall := plan.Root.Body[3]
	innerCall := outerCall.Actions[0]
	if innerCall.CaptureSource.Capture == nil || innerCall.CaptureSource.Slot != nil {
		t.Fatal("nested action reselected global name instead of saved handle")
	}
	if innerCall.Actions[0].Target.Storage.Capture == nil || plan.Root.Body[4].Value.Storage.Capture != nil {
		t.Fatal("captured write escaped into original storage")
	}
	emitted, err := (&orchestrationEmitter{plan: plan}).block(plan.Root, false)
	if err != nil {
		t.Fatal(err)
	}
	source := strings.Join(emitted, "\n")
	if !strings.Contains(source, "LValue::Environment(") || !strings.Contains(source, ".into_environment()?") || !strings.Contains(source, ".set(0,") {
		t.Fatalf("missing managed captured handle/write: %s", source)
	}
	if strings.Contains(source, "borrow_mut") || strings.Contains(source, ".borrow()") {
		t.Fatal("emitter exposed environment guards across calls")
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
