package Lyger

import (
	"sort"

	"github.com/l-programming-language/l/contracts"
)

// This projection consumes typed instructions and the same canonical App
// records as lowering. It describes covered source control flow; opaque native
// activities remain a separate closure obligation.
type contractLifecyclePath struct {
	apps     map[string][]contracts.AppProtocolState
	terminal bool
	returned bool
}

type contractLifecycleAnalysis struct {
	recorder    *contractRecorder
	functions   map[*record]*orchestrationFunction
	active      map[*record]bool
	obligations []contracts.Obligation
	ordinal     int
}

func (r *contractRecorder) contractLifecycle(plan *orchestrationPlan) ([]contracts.AppProtocolState, []contracts.Obligation) {
	a := contractLifecycleAnalysis{recorder: r, functions: make(map[*record]*orchestrationFunction), active: make(map[*record]bool)}
	a.collectFunctions(plan.Root.Body)
	path := contractLifecyclePath{apps: make(map[string][]contracts.AppProtocolState)}
	for _, app := range r.compilation.Semantics.Apps {
		target := r.contractAppTarget(app)
		path.apps[target.Root] = []contracts.AppProtocolState{{App: target, Stage: contracts.AppIdle, Claim: contracts.ClaimUnclaimed, Owner: "orchestrator-supervisor", Authority: contracts.TransitionOwner}}
	}
	paths := a.sequence(plan.Root.Body, []contractLifecyclePath{path})
	var states []contracts.AppProtocolState
	for _, path := range paths {
		var names []string
		for name := range path.apps {
			names = append(names, name)
		}
		sort.Strings(names)
		for _, name := range names {
			states = contracts.JoinAppProtocol(states, contracts.CloseAppEnvironment(path.apps[name]))
		}
	}
	return states, a.obligations
}

// CheckLifecycleProfile combines the current concrete closure with the typed
// lifecycle projection. A conditional source projection alone never certifies
// unknown native activities or cleanup. Coordinator reinference updates the
// unknowns used here without changing source transition choices.
func (a *ContractAnalysis) CheckLifecycleProfile(profile contracts.Profile) []contracts.Obligation {
	if a == nil {
		return []contracts.Obligation{{ID: "lifecycle-analysis", Status: contracts.Unresolved, Reason: "lifecycle analysis is absent"}}
	}
	result := contracts.CheckLifecycleProfile(a.Lifecycle, profile)
	result = append(result, a.LifecycleObligations...)
	seen := make(map[string]bool)
	for _, unknown := range a.Summary.Unknowns {
		relevant := len(unknown.Components) == 0
		for _, component := range unknown.Components {
			relevant = relevant || component == "ownership" || component == "effects" || component == "outcomes" || component == "cleanup"
		}
		if !relevant {
			continue
		}
		id := contracts.StableID("lifecycle-closure", unknown)
		if seen[id] {
			continue
		}
		seen[id] = true
		result = append(result, contracts.Obligation{ID: id, Claim: contracts.Claim{Predicate: contracts.Predicate{Kind: "lifecycle_closed", Subject: unknown.Boundary}, Target: profile.Target, Scope: profile.Scope, Outcomes: profile.Outcomes}, Status: contracts.Unresolved, Reason: unknown.Message, Origin: unknown.Origin})
	}
	return result
}

func cloneContractLifecyclePath(path contractLifecyclePath) contractLifecyclePath {
	result := contractLifecyclePath{apps: make(map[string][]contracts.AppProtocolState), terminal: path.terminal, returned: path.returned}
	for key, states := range path.apps {
		result.apps[key] = append([]contracts.AppProtocolState(nil), states...)
	}
	return result
}

func (a *contractLifecycleAnalysis) collectFunctions(body []*orchestrationInstruction) {
	for _, current := range body {
		if current == nil {
			continue
		}
		if current.Function != nil && current.Function.Body != nil {
			a.functions[current.Function.Source.Record] = current.Function
			a.collectFunctions(current.Function.Body.Body)
		}
		for _, block := range []*orchestrationBlock{current.Block, current.Continuation} {
			if block != nil {
				a.collectFunctions(block.Body)
			}
		}
		for _, nested := range [][]*orchestrationInstruction{current.Then, current.Else, current.Actions, current.AppCases} {
			a.collectFunctions(nested)
		}
	}
}

func (a *contractLifecycleAnalysis) sequence(body []*orchestrationInstruction, paths []contractLifecyclePath) []contractLifecyclePath {
	for _, current := range body {
		if current == nil || current.Source == nil {
			continue
		}
		var next []contractLifecyclePath
		for _, path := range paths {
			if path.terminal || path.returned {
				next = append(next, path)
				continue
			}
			next = append(next, a.step(current, path)...)
		}
		paths = a.bound(next)
	}
	return paths
}

func (a *contractLifecycleAnalysis) step(current *orchestrationInstruction, path contractLifecyclePath) []contractLifecyclePath {
	if len(current.AppCases) > 0 {
		var result []contractLifecyclePath
		for _, branch := range current.AppCases {
			result = append(result, a.step(branch, cloneContractLifecyclePath(path))...)
		}
		return result
	}
	source := current.Source
	origin := a.recorder.origin(source.SourceToken)
	for _, value := range append(append([]*orchestrationValue{current.Value, current.Target, current.Key}, current.Arguments...), current.Results...) {
		a.observeValue(value, path, origin)
	}
	if current.Condition != nil {
		a.observeValue(current.Condition.Left, path, origin)
		a.observeValue(current.Condition.Right, path, origin)
	}
	switch source.Kind {
	case instructionIf:
		return append(a.sequence(current.Then, []contractLifecyclePath{cloneContractLifecyclePath(path)}), a.sequence(current.Else, []contractLifecyclePath{path})...)
	case instructionBlock:
		if current.Block != nil {
			return a.sequence(current.Block.Body, []contractLifecyclePath{path})
		}
	case instructionCall:
		return a.sequence(current.Actions, []contractLifecyclePath{path})
	case instructionInvoke:
		var rec *record
		if current.Target != nil {
			var err error
			rec, err = canonicalRecord(current.Target.Record, a.recorder.compilation.Semantics.Aliases)
			if err != nil {
				rec = nil
			}
		}
		if function := a.functions[rec]; function != nil && !a.active[rec] {
			a.active[rec] = true
			paths := a.sequence(function.Body.Body, []contractLifecyclePath{path})
			delete(a.active, rec)
			for i := range paths {
				paths[i].returned = false
			}
			return paths
		}
		a.obligations = append(a.obligations, contracts.Obligation{ID: contracts.StableID("lifecycle-call", origin), Status: contracts.Unresolved, Origin: origin, Reason: "dynamic, native or recursive call has unresolved activity ownership and transition effects"})
		for name, states := range path.apps {
			for i := range states {
				states[i].Owner = ""
				states[i].Authority = contracts.UnknownAuthority
			}
			path.apps[name] = states
		}
	case instructionReturn:
		path.returned = true
	case instructionExit:
		path.terminal = true
	case instructionBuild, instructionRun, instructionStop, instructionAwait:
		if current.App == nil {
			a.obligations = append(a.obligations, contracts.Obligation{ID: contracts.StableID("lifecycle-target", origin), Status: contracts.Unresolved, Origin: origin, Reason: "lifecycle target is not a resolved canonical App"})
			return []contractLifecyclePath{path}
		}
		target := a.recorder.contractAppTarget(current.App)
		kind := map[instructionKind]string{instructionBuild: "Build", instructionRun: "Run", instructionStop: "Stop", instructionAwait: "Await"}[source.Kind]
		operation := contracts.ProtocolOperation{Kind: kind, Owner: "orchestrator-supervisor"}
		if source.Kind == instructionRun {
			a.ordinal++
			operation.Invocation = contracts.StableID("source-invocation", struct {
				Target  contracts.TargetRef
				Origin  contracts.Origin
				Ordinal int
			}{target, origin, a.ordinal})
		}
		transition := contracts.ApplyAppProtocol(path.apps[target.Root], operation)
		a.obligations = append(a.obligations, contracts.Obligation{ID: contracts.StableID("source-transition", struct {
			Operation contracts.ProtocolOperation
			Target    contracts.TargetRef
			Origin    contracts.Origin
		}{operation, target, origin}), Claim: contracts.Claim{Predicate: contracts.Predicate{Kind: "protocol", Subject: kind, Target: target}, Target: target, Scope: "source_transition"}, Status: transition.Status, Reason: transition.Reason, Origin: origin})
		var cancellation []contractLifecyclePath
		if transition.CancellationPossible {
			cancelled := cloneContractLifecyclePath(path)
			cancelled.terminal = true
			cancellation = append(cancellation, cancelled)
		}
		path.apps[target.Root] = transition.States
		if source.Kind == instructionRun {
			// Startup failure is a finalized invocation retained by the supervisor.
			// Build its alternative from the admitted transition, never retry selection.
			for _, state := range transition.States {
				if state.Invocation == operation.Invocation {
					state.Stage = contracts.AppFinalized
					state.StartupFailure = true
					path.apps[target.Root] = contracts.JoinAppProtocol(path.apps[target.Root], []contracts.AppProtocolState{state})
				}
			}
		}
		if source.Kind == instructionAwait {
			return append(cancellation, a.awaitContinuation(current, path, target, origin)...)
		}
		return append(cancellation, path)
	}
	return []contractLifecyclePath{path}
}

func (a *contractLifecycleAnalysis) observeValue(value *orchestrationValue, path contractLifecyclePath, origin contracts.Origin) {
	if value == nil {
		return
	}
	for _, nested := range append([]*orchestrationValue{value.Owner, value.Payload, value.Key}, value.Elements...) {
		a.observeValue(nested, path, origin)
	}
	if value.Source == nil || value.Source.Member != memberOutput {
		return
	}
	var app *App
	if value.Owner != nil {
		app = value.Owner.App
	}
	target := a.recorder.contractAppTarget(app)
	result := contracts.ApplyAppProtocol(path.apps[target.Root], contracts.ProtocolOperation{Kind: "ReadOutput"})
	a.obligations = append(a.obligations, contracts.Obligation{ID: contracts.StableID("source-output", struct {
		Origin contracts.Origin
		Target contracts.TargetRef
	}{origin, target}), Claim: contracts.Claim{Predicate: contracts.Predicate{Kind: "protocol", Subject: "ReadOutput", Target: target}, Target: target, Scope: "source_transition"}, Status: result.Status, Reason: result.Reason, Origin: origin})
}

// Await's cancellation alternative retains the existing owner; Stop never
// joins, and no lexical scope is interpreted as implicit process termination.
func (a *contractLifecycleAnalysis) awaitContinuation(current *orchestrationInstruction, path contractLifecyclePath, target contracts.TargetRef, origin contracts.Origin) []contractLifecyclePath {
	var result []contractLifecyclePath
	for _, state := range path.apps[target.Root] {
		branch := cloneContractLifecyclePath(path)
		branch.apps[target.Root] = []contracts.AppProtocolState{state}
		claimed := contracts.ApplyAppProtocol([]contracts.AppProtocolState{state}, contracts.ProtocolOperation{Kind: "ClaimContinuation", Invocation: state.Invocation})
		branch.apps[target.Root] = claimed.States
		if claimed.Status != contracts.Established {
			a.obligations = append(a.obligations, contracts.Obligation{ID: contracts.StableID("source-continuation", struct {
				Origin contracts.Origin
				State  contracts.AppProtocolState
			}{origin, state}), Status: claimed.Status, Reason: claimed.Reason, Origin: origin})
		}
		if !claimed.ContinuationActivated {
			result = append(result, branch)
			continue
		}
		paths := []contractLifecyclePath{branch}
		if current.Continuation != nil {
			paths = a.sequence(current.Continuation.Body, paths)
		}
		for _, out := range paths {
			if !out.terminal {
				finished := contracts.ApplyAppProtocol(out.apps[target.Root], contracts.ProtocolOperation{Kind: "FinishContinuation", Invocation: state.Invocation})
				out.apps[target.Root] = finished.States
			}
			result = append(result, out)
		}
	}
	return result
}

func (a *contractLifecycleAnalysis) bound(paths []contractLifecyclePath) []contractLifecyclePath {
	if len(paths) <= 32 {
		return paths
	}
	joined := contractLifecyclePath{apps: make(map[string][]contracts.AppProtocolState)}
	for _, path := range paths {
		for key, states := range path.apps {
			joined.apps[key] = contracts.JoinAppProtocol(joined.apps[key], states)
		}
	}
	a.obligations = append(a.obligations, contracts.Obligation{ID: "source-lifecycle-alternatives", Status: contracts.Unresolved, Reason: "more than 32 lifecycle alternatives; correlations and terminal control widened conservatively"})
	return []contractLifecyclePath{joined}
}
