package Lyger

import (
	"github.com/l-programming-language/l/lerrors"
)

// An App's runtime handle and its source continuation are different identities.
// Dynamic App values select among the complete canonical program inventory;
// successful paths retain the same per-invocation claim as a direct Await.
func (p *orchestrationPlanner) processInstruction(current *instruction, flow orchestrationFlow) (*orchestrationInstruction, orchestrationFlow, error) {
	target, err := p.value(current.Target, flow)
	if err != nil {
		return nil, flow, err
	}
	if target.Type != valueApp {
		return nil, flow, lerrors.TypeMismatch.Errorf("Build/Run or member access requires an application, got %s", target.Type)
	}
	if target.App != nil {
		return p.canonicalProcess(current, target, target.App, flow)
	}
	result := &orchestrationInstruction{Source: current, Target: target}
	joined := newOrchestrationFlow()
	joined.live = false
	for _, app := range p.semantics.Apps {
		if !p.apps[app] {
			p.apps[app] = true
			p.plan.Apps = append(p.plan.Apps, app)
		}
		branch, next, err := p.canonicalProcess(current, &orchestrationValue{Type: valueApp, App: app}, app, flow.clone())
		if err != nil {
			return nil, flow, err
		}
		result.AppCases = append(result.AppCases, branch)
		result.Effects |= branch.Effects
		joined = joinOrchestrationFlow(joined, next)
	}
	result.ReachesNext = joined.live
	return result, joined, nil
}
func (p *orchestrationPlanner) canonicalProcess(current *instruction, target *orchestrationValue, app *App, flow orchestrationFlow) (*orchestrationInstruction, orchestrationFlow, error) {
	result := &orchestrationInstruction{Source: current, Target: target, App: app}
	switch current.Kind {
	case instructionBuild:
		result.Effects = orchestrationEffectBuild
	case instructionRun:
		result.Effects = orchestrationEffectRun
	case instructionStop:
		result.Effects = orchestrationEffectStop
	case instructionAwait:
		result.Effects = orchestrationEffectAwait
		if continuation := p.semantics.Continuations[app]; continuation != nil && !p.active[app] {
			p.active[app] = true
			if p.functionDepth > 0 {
				p.plan.FunctionProgramState = true
			}
			// Continuations retain definition-site storage, not caller snapshots.
			savedCaptures := p.captureContexts
			p.captureContexts = nil
			var err error
			result.Continuation, flow, err = p.scoped(continuation, flow)
			p.captureContexts = savedCaptures
			delete(p.active, app)
			if err != nil {
				return nil, flow, err
			}
			result.Effects |= orchestrationInstructionsEffects(result.Continuation.Body)
		}
	}
	result.ReachesNext = flow.live
	return result, flow, nil
}
func samePlannedApp(left, right *orchestrationValue) bool {
	if left == nil || right == nil {
		return false
	}
	if left.App != nil || right.App != nil {
		return left.App != nil && left.App == right.App
	}
	return left.Storage != nil && left.Storage == right.Storage
}
