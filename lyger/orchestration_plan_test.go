package Lyger

import (
	"reflect"
	"strings"
	"testing"
)

func planTestScalar(name string, kind valueType, scope *lexicalScope) *record {
	return &record{ID: name, storageScope: scope, Value: &entry{Type: storagePrimitive, Primitive: &primitive{Type: kind, Text: "`value`"}}}
}

func planTestRef(rec *record) *instructionValue {
	kind, _ := instructionType(rec) //nolint:errcheck // Every caller constructs a valid typed record; this helper only projects its known type.
	return &instructionValue{Kind: operandSymbol, Type: kind, Record: rec}
}

func planTestApp(name string) (*App, *instructionValue) {
	app := &App{ID: name}
	rec := &record{ID: name, Value: &entry{Type: storageApplicative, Applicative: &applicative{Type: valueApp, App: app}}}
	return app, planTestRef(rec)
}

func TestPlanOwnershipAtValueBoundaries(t *testing.T) {
	source := planTestScalar("source", valueText, nil)
	target := planTestScalar("target", valueText, nil)
	assign := &instruction{Kind: instructionAssign, Target: planTestRef(target), Value: planTestRef(source)}
	semantics := newSemanticProgram()
	semantics.Instructions = []*instruction{assign,
		{Kind: instructionShow, Value: planTestRef(source)},
		{Kind: instructionShow, Value: planTestRef(target)},
		assign}
	plan := requireOrchestrationPlan(t, semantics)
	if plan.Root.Body[0].Value.Access != orchestrationAccessSnapshot || plan.Root.Body[1].Value.Access != orchestrationAccessBorrow || plan.Root.Body[2].Value.Access != orchestrationAccessBorrow {
		t.Fatal("Text assignment must own an independent snapshot; reads must borrow retained storage")
	}
	if plan.Root.Body[0].Value.Record != source || plan.Root.Body[0].Target.Record != target {
		t.Fatal("planning rebound a source symbol")
	}
	if !plan.Root.Body[0].Target.Storage.Mutable || plan.Root.Body[1].Value.Storage.Mutable {
		t.Fatal("mutability must follow writes, not the non-Copy Text type")
	}
	if plan.Root.Body[0].Target.Storage.Initial != nil || plan.Root.Body[1].Value.Storage.Initial == nil {
		t.Fatal("dynamic storage acquired an implicit initial value or lost an explicit source value")
	}
}

func TestPlanDefiniteAssignmentUsesContinuingPaths(t *testing.T) {
	for _, scenario := range []string{"both", "missing", "exit"} {
		t.Run(scenario, func(t *testing.T) {
			rec := planTestScalar("result", valueNumber, nil)
			assign := &instruction{Kind: instructionAssign, Target: planTestRef(rec), Value: &instructionValue{Kind: operandLiteral, Type: valueNumber, Literal: &primitive{Type: valueNumber, Number: 3}}}
			branch := &instruction{Kind: instructionIf, Condition: &instructionCondition{Operator: "truth", Left: &instructionValue{Kind: operandLiteral, Type: valueTruth, Literal: &primitive{Type: valueTruth, Truth: true}}}, Then: []*instruction{assign}}
			switch scenario {
			case "both":
				branch.Else = []*instruction{assign}
			case "exit":
				branch.Else = []*instruction{{Kind: instructionExit}}
			}
			semantics := newSemanticProgram()
			semantics.Instructions = []*instruction{branch, {Kind: instructionShow, Value: planTestRef(rec), Source: SourceSpan{Path: "read.l", Start: SourcePosition{Line: 7, Column: 2}}}}
			plan, err := planOrchestration(semantics)
			if scenario == "missing" {
				if err == nil || !strings.Contains(err.Error(), "read.l:7:2") || !strings.Contains(err.Error(), "symbol result is not definitely assigned") || plan != nil {
					t.Fatalf("incomplete branch assignment = %#v/%v", plan, err)
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			if plan.Root.Body[1].Value.Storage.Mutable {
				t.Fatal("mutually exclusive initializations do not require mutable storage")
			}
		})
	}
}

func TestPlanActivationIdentityAndCapture(t *testing.T) {
	scope := &lexicalScope{kind: lexicalParagraph}
	first, firstRef := planTestApp("First")
	second, secondRef := planTestApp("Second")
	private := planTestScalar("saved", valueText, scope)
	assign := &instruction{Kind: instructionAssign, Target: planTestRef(private), Value: &instructionValue{Kind: operandMember, Type: valueText, Member: memberOutput, Owner: firstRef}}
	read := &instruction{Kind: instructionShow, Value: planTestRef(private)}
	semantics := newSemanticProgram()
	semantics.Apps = []*App{first, second}
	semantics.Continuations = map[*App]*scopedInstructions{
		first:  {Scope: scope, Body: []*instruction{assign, {Kind: instructionAwait, Target: secondRef}, read}},
		second: {Scope: scope, Body: []*instruction{assign, read}},
	}
	semantics.Instructions = []*instruction{{Kind: instructionAwait, Target: firstRef}}
	plan := requireOrchestrationPlan(t, semantics)
	outer := plan.Root.Body[0].Continuation
	inner := outer.Body[1].Continuation
	if outer.Scope != inner.Scope || outer.ID == inner.ID || outer.Storage[0] == inner.Storage[0] || outer.Body[2].Value.Storage != outer.Storage[0] {
		t.Fatal("nested activations sharing lexical scope must retain distinct storage and resume the caller snapshot")
	}
	if outer.Body[0].Source != assign || outer.Body[2].Source != read || outer.Body[0].Value.Access != orchestrationAccessOwned {
		t.Fatal("planning lost source identity or added a clone to an owned output snapshot")
	}
	secondBody := semantics.Continuations[second].Body
	semantics.Continuations[second].Body = append([]*instruction{read}, secondBody...)
	if _, err := planOrchestration(semantics); err == nil || !strings.Contains(err.Error(), "symbol saved is not definitely assigned") {
		t.Fatalf("callee inherited caller initialization: %v", err)
	}
	semantics.Continuations[second].Body = secondBody
	again, err := planOrchestration(semantics)
	if err != nil || !reflect.DeepEqual(plan, again) {
		t.Fatalf("repeated planning changed identity, storage or semantic inputs: %v", err)
	}
}

func TestPlanStaticCaptureAndBoundedContinuation(t *testing.T) {
	private := planTestScalar("captured", valueText, &lexicalScope{kind: lexicalParagraph})
	app, appRef := planTestApp("Worker")
	semantics := newSemanticProgram()
	semantics.Apps = []*App{app}
	semantics.Continuations = map[*App]*scopedInstructions{app: {Scope: &lexicalScope{kind: lexicalParagraph}, Body: []*instruction{{Kind: instructionAwait, Target: appRef}, {Kind: instructionShow, Value: planTestRef(private)}}}}
	semantics.Instructions = []*instruction{{Kind: instructionAwait, Target: appRef}}
	plan := requireOrchestrationPlan(t, semantics)
	body := plan.Root.Body[0].Continuation.Body
	if body[0].Continuation != nil || body[0].App != app || body[0].Effects&orchestrationEffectAwait == 0 || body[1].Value.Storage != nil || body[1].Value.Literal == nil {
		t.Fatal("self-Await must remain a real join; static captures retain their defining value without runtime storage")
	}
	if body[1].Value.Literal == private.Value.Primitive {
		t.Fatal("planned literal aliases mutable semantic storage")
	}
	private.Value.Runtime = &runtimeReference{OwnerID: "Worker", Member: memberOutput}
	if plan, err := planOrchestration(semantics); err == nil || !strings.Contains(err.Error(), "outside its paragraph activation") || plan != nil {
		t.Fatalf("escaping private runtime capture = %#v/%v", plan, err)
	}
}

func requireOrchestrationPlan(t *testing.T, semantics *semanticProgram) *orchestrationPlan {
	t.Helper()
	plan, err := planOrchestration(semantics)
	if err != nil {
		t.Fatal(err)
	}
	return plan
}
