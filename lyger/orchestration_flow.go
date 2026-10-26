package Lyger

import "github.com/l-programming-language/l/lerrors"

// Definite facts intersect; possible writes unite. A terminating predecessor
// contributes to neither continuing set. No facts encode Rust source strings.
type orchestrationFlow struct {
	assigned map[*orchestrationStorage]bool
	written  map[*orchestrationStorage]bool
	live     bool
	bound    map[*orchestrationCallable]bool
}

func newOrchestrationFlow() orchestrationFlow {
	return orchestrationFlow{assigned: make(map[*orchestrationStorage]bool), written: make(map[*orchestrationStorage]bool), live: true, bound: make(map[*orchestrationCallable]bool)}
}

func (flow orchestrationFlow) clone() orchestrationFlow {
	cloned := newOrchestrationFlow()
	cloned.live = flow.live
	for slot, bound := range flow.bound {
		cloned.bound[slot] = bound
	}
	for slot, assigned := range flow.assigned {
		cloned.assigned[slot] = assigned
	}
	for slot, written := range flow.written {
		cloned.written[slot] = written
	}
	return cloned
}

func joinOrchestrationFlow(left, right orchestrationFlow) orchestrationFlow {
	if !left.live {
		return right
	}
	if !right.live {
		return left
	}
	joined := left.clone()
	for slot := range joined.bound {
		joined.bound[slot] = right.bound[slot]
	}
	for slot := range joined.assigned {
		joined.assigned[slot] = right.assigned[slot]
	}
	for slot, written := range right.written {
		joined.written[slot] = joined.written[slot] || written
	}
	return joined
}

func (p *orchestrationPlanner) instructions(body []*instruction, flow orchestrationFlow) ([]*orchestrationInstruction, orchestrationFlow, error) {
	var result []*orchestrationInstruction
	for _, current := range body {
		if !flow.live {
			break
		}
		planned, next, err := p.instruction(current, flow)
		if err != nil {
			return nil, flow, current.wrapError(err)
		}
		result = append(result, planned)
		flow = next
	}
	return result, flow, nil
}

func orchestrationInstructionsEffects(body []*orchestrationInstruction) orchestrationEffects {
	var effects orchestrationEffects
	for _, current := range body {
		effects |= current.Effects
	}
	return effects
}

func (p *orchestrationPlanner) instruction(current *instruction, flow orchestrationFlow) (*orchestrationInstruction, orchestrationFlow, error) {
	if current == nil {
		return nil, flow, lerrors.Internal.Errorf("nil instruction")
	}
	result := &orchestrationInstruction{Source: current}
	var err error
	switch current.Kind {
	case instructionFunction, instructionInvoke, instructionReturn:
		return p.functionInstruction(current, flow)
	case instructionBlock:
		result.Block, flow, err = p.scoped(current.Block, flow)
		if err == nil {
			result.Effects = orchestrationInstructionsEffects(result.Block.Body)
		}
	case instructionBindAction, instructionAliasAction:
		return p.bindCapture(current, flow)
	case instructionCall:
		if !current.Synchronous {
			return nil, flow, lerrors.Unsupported.Errorf("source action requires synchronous expanded instructions")
		}
		if len(current.Actions) == 0 {
			return nil, flow, lerrors.Internal.Errorf("source action requires synchronous expanded instructions")
		}
		restore, captureErr := p.enterCaptureCall(current, result, flow)
		if captureErr != nil {
			return nil, flow, captureErr
		}
		result.Actions, flow, err = p.instructions(current.Actions, flow)
		restore()
		result.Effects = orchestrationInstructionsEffects(result.Actions)
	case instructionBuild, instructionRun, instructionStop, instructionAwait:
		return p.processInstruction(current, flow)
	case instructionShow, instructionExit:
		if current.Kind == instructionExit {
			result.Effects = orchestrationEffectExit
			if current.Value == nil {
				flow.live = false
				break
			}
		} else {
			result.Effects = orchestrationEffectShow
		}
		result.Value, err = p.value(current.Value, flow)
		if err != nil {
			break
		}
		if current.Kind == instructionShow {
			kind := result.Value.Type
			if _, ok := runtimeTypeName(kind); !ok || kind == valueApp || kind == valueError {
				return nil, flow, lerrors.TypeMismatch.Errorf("Show requires a scalar value, got %s", kind)
			}
		} else {
			if result.Value.Type != valueError {
				return nil, flow, lerrors.TypeMismatch.Errorf("Exit requires Error, got %s", result.Value.Type)
			}
			flow.live = false
		}
		result.Effects |= orchestrationValueEffects(result.Value)
	case instructionUpdate:
		result, err = p.runtimeUpdate(current, flow)
	case instructionAssign:
		rec, kind, resolveErr := p.record(current.Target)
		if resolveErr != nil {
			return nil, flow, resolveErr
		}
		result.Value, err = p.value(current.Value, flow)
		if err != nil {
			break
		}
		slot, _, storageErr := p.storage(rec, kind)
		if storageErr != nil {
			return nil, flow, storageErr
		}
		if result.Value.Type != kind {
			return nil, flow, lerrors.TypeMismatch.Errorf("assignment to %s requires %s, got %s", rec.ID, kind, result.Value.Type)
		}
		if !slot.RuntimeType.accepts(result.Value.RuntimeType) {
			return nil, flow, lerrors.TypeMismatch.Errorf("assignment to %s has incompatible runtime type", rec.ID)
		}
		result.Target = &orchestrationValue{Source: current.Target, Type: kind, RuntimeType: slot.RuntimeType, Record: rec, Storage: slot}
		if result.Value.Access == orchestrationAccessBorrow {
			result.Value.Access = orchestrationAccessSnapshot
		}
		slot.Mutable = slot.Mutable || slot.Initial != nil || flow.written[slot]
		flow.assigned[slot], flow.written[slot] = true, true
		result.Effects = orchestrationEffectAssign | orchestrationValueEffects(result.Value)
	case instructionIf:
		result.Condition, err = p.condition(current.Condition, flow)
		if err != nil {
			break
		}
		var thenFlow, elseFlow orchestrationFlow
		result.Then, thenFlow, err = p.instructions(current.Then, flow.clone())
		if err != nil {
			break
		}
		result.Else, elseFlow, err = p.instructions(current.Else, flow.clone())
		if err != nil {
			break
		}
		flow = joinOrchestrationFlow(thenFlow, elseFlow)
		result.Effects = orchestrationValueEffects(result.Condition.Left) | orchestrationValueEffects(result.Condition.Right) | orchestrationInstructionsEffects(result.Then) | orchestrationInstructionsEffects(result.Else)
	default:
		return nil, flow, lerrors.Internal.Errorf("unsupported instruction %s", current.Kind)
	}
	if err != nil {
		return nil, flow, err
	}
	result.ReachesNext = flow.live
	return result, flow, nil
}
