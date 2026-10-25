package Lyger

import (
	"errors"
	"strings"
	"testing"

	"github.com/l-programming-language/l/lerrors"
)

func TestPlanLifecycleEffects(t *testing.T) {
	for _, tc := range []struct {
		name                 string
		kind                 instructionKind
		effects              orchestrationEffects
		registerContinuation bool
		wantContinuation     bool
	}{
		{"Build", instructionBuild, orchestrationEffectBuild, true, false},
		{"Run", instructionRun, orchestrationEffectRun, true, false},
		{"Stop", instructionStop, orchestrationEffectStop, true, false},
		{"Await", instructionAwait, orchestrationEffectAwait, false, false},
		{"Await continuation", instructionAwait, orchestrationEffectAwait | orchestrationEffectShow, true, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			app, appRef := planTestApp("Worker")
			current := &instruction{Kind: tc.kind, Target: appRef, Source: SourceSpan{Path: "lifecycle.l", Start: SourcePosition{Line: 7, Column: 3}}}
			show := &instruction{Kind: instructionShow, Value: &instructionValue{Kind: operandLiteral, Literal: &primitive{Type: valueNumber, Number: 3}}}
			continuation := &scopedInstructions{Scope: &lexicalScope{kind: lexicalParagraph}, Body: []*instruction{show}}
			semantics := newSemanticProgram()
			semantics.Instructions = []*instruction{current}
			if tc.registerContinuation {
				semantics.Continuations = map[*App]*scopedInstructions{app: continuation}
			}
			plan, err := planOrchestration(semantics)
			if err != nil {
				t.Fatal(err)
			}
			if len(plan.Root.Body) != 1 || len(plan.Apps) != 1 || plan.Apps[0] != app {
				t.Fatalf("lifecycle plan lost its instruction or application identity: %+v", plan)
			}
			planned := plan.Root.Body[0]
			if planned.Source != current || planned.App != app || planned.Target.Source != appRef || planned.Target.App != app || planned.Effects != tc.effects || !planned.ReachesNext {
				t.Fatalf("lifecycle identity, effects or reachability changed: %+v; want effects %d", planned, tc.effects)
			}
			if !tc.wantContinuation {
				if planned.Continuation != nil {
					t.Fatal("lifecycle instruction acquired an unexpected continuation")
				}
				return
			}
			body := planned.Continuation
			if body == nil || body.Scope != continuation.Scope || len(body.Body) != 1 || body.Body[0].Source != show || body.Body[0].Effects != orchestrationEffectShow {
				t.Fatalf("Await lost the continuation identity or effects: %+v", body)
			}
		})
	}
}

func TestPlanRejectsInvalidInstructionKinds(t *testing.T) {
	for _, tc := range []struct {
		name string
		kind instructionKind
	}{
		{"zero", instructionInvalid},
		{"count", instructionKindCount},
		{"out of range", instructionKind(255)},
	} {
		t.Run(tc.name, func(t *testing.T) {
			semantics := newSemanticProgram()
			semantics.Instructions = []*instruction{{Kind: tc.kind, Source: SourceSpan{Path: "lifecycle.l", Start: SourcePosition{Line: 7, Column: 3}}}}
			plan, err := planOrchestration(semantics)
			if plan != nil || !errors.Is(err, lerrors.Internal) || !strings.Contains(err.Error(), "lifecycle.l:7:3") || !strings.Contains(err.Error(), "unsupported instruction") {
				t.Fatalf("invalid instruction lost its classification or source context: plan=%+v, error=%v", plan, err)
			}
		})
	}
}
