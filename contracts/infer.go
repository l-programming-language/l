package contracts

import (
	"fmt"
	"sort"
)

type NodeKind string

const (
	Sequence NodeKind = "sequence"
	Branch   NodeKind = "branch"
	Event    NodeKind = "event"
	Call     NodeKind = "call"
	Callback NodeKind = "callback"
	Native   NodeKind = "native"
	Guard    NodeKind = "guard"
	Require  NodeKind = "require"
	Acquire  NodeKind = "acquire"
	Move     NodeKind = "move"
	Release  NodeKind = "release"
	Scope    NodeKind = "scope"
	Repeat   NodeKind = "repeat"
	Mutation NodeKind = "mutation"
)

type Substitutions struct {
	Resources map[string]ResourceRef `json:"resources,omitempty"`
	Targets   map[string]TargetRef   `json:"targets,omitempty"`
	Values    map[string]Atom        `json:"values,omitempty"`
}

// Node is compiler-produced typed behavior, not authored evidence. Native
// behavior must pass through Native and explicit approved conformance inputs.
type Node struct {
	Kind            NodeKind       `json:"kind"`
	ID              string         `json:"id,omitempty"`
	Origin          Origin         `json:"origin"`
	Meaning         MeaningKey     `json:"meaning,omitempty"`
	Callback        string         `json:"callback,omitempty"`
	Native          string         `json:"native,omitempty"`
	Children        []Node         `json:"children,omitempty"`
	Then            []Node         `json:"then,omitempty"`
	Else            []Node         `json:"else,omitempty"`
	Effects         []Effect       `json:"effects,omitempty"`
	Facts           []Predicate    `json:"facts,omitempty"`
	Requires        []Predicate    `json:"requires,omitempty"`
	Outcome         Outcome        `json:"outcome,omitempty"`
	Resource        ResourceRef    `json:"resource,omitempty"`
	Target          TargetRef      `json:"target,omitempty"`
	Phase           Phase          `json:"phase,omitempty"`
	Ownership       string         `json:"ownership,omitempty"`
	SuspensionSafe  bool           `json:"suspension_safe,omitempty"`
	Count           *int64         `json:"count,omitempty"`
	CountSymbol     string         `json:"count_symbol,omitempty"`
	StateInvariant  bool           `json:"state_invariant,omitempty"`
	Costs           *ResourceCost  `json:"costs,omitempty"`
	Cleanup         []Node         `json:"cleanup,omitempty"`
	CleanupOutcomes []Outcome      `json:"cleanup_outcomes,omitempty"`
	Condition       string         `json:"condition,omitempty"`
	FinalOwner      Status         `json:"final_owner,omitempty"`
	Substitutions   Substitutions  `json:"substitutions,omitempty"`
	MutationFacts   *MutationFacts `json:"mutation_facts,omitempty"`
}

// MutationFacts describe the already selected representation; they never select
// a new representation. Sharing and capacity are independent proof obligations.
type MutationFacts struct {
	CopyOnWrite    bool     `json:"copy_on_write"`
	Sharing        Status   `json:"sharing"`
	CapacityEnough Status   `json:"capacity_enough"`
	NestedClosed   bool     `json:"nested_closed"`
	NestedEffects  []Effect `json:"nested_effects,omitempty"`
}

type Plan struct {
	Entry       Node                `json:"entry"`
	Definitions map[MeaningKey]Node `json:"definitions,omitempty"`
	Callbacks   map[string]Node     `json:"callbacks,omitempty"`
	// Residuals and NativeClosure describe incomplete analysis coverage, not
	// operations scheduled after source execution. They cover every outcome.
	Residuals     []UnknownReason  `json:"residuals,omitempty"`
	NativeClosure []NativeBoundary `json:"native_closure,omitempty"`
}

type UnknownReason struct {
	Code       string   `json:"code"`
	Message    string   `json:"message"`
	Boundary   string   `json:"boundary,omitempty"`
	Origin     Origin   `json:"origin"`
	Components []string `json:"components,omitempty"`
	Phase      Phase    `json:"phase,omitempty"`
}

type NativeBoundary struct {
	ID       string     `json:"id"`
	Meaning  MeaningKey `json:"meaning,omitempty"`
	Target   TargetRef  `json:"target"`
	Origin   Origin     `json:"origin"`
	Status   Status     `json:"status"`
	Reason   string     `json:"reason"`
	Witness  string     `json:"witness,omitempty"`
	Implicit bool       `json:"implicit,omitempty"`
}

type Multiplicity struct {
	Lower  int64  `json:"lower"`
	Upper  *int64 `json:"upper,omitempty"`
	Symbol string `json:"symbol,omitempty"`
}

type OwnedResource struct {
	Resource       ResourceRef `json:"resource"`
	Mode           string      `json:"mode"`
	State          string      `json:"state"`
	SuspensionSafe bool        `json:"suspension_safe"`
	FinalOwner     Status      `json:"final_owner"`
	Destructor     []Node      `json:"destructor,omitempty"`
	Origin         Origin      `json:"origin"`
}

type PathSummary struct {
	Outcome          Outcome                  `json:"outcome"`
	Reachable        bool                     `json:"reachable"`
	Facts            []Predicate              `json:"facts,omitempty"`
	Effects          []Effect                 `json:"effects,omitempty"`
	Demand           map[string]Multiplicity  `json:"demand,omitempty"`
	Resources        ResourceCost             `json:"resources"`
	Owned            map[string]OwnedResource `json:"owned,omitempty"`
	OwnershipOrder   []string                 `json:"ownership_order,omitempty"`
	Unknowns         []UnknownReason          `json:"unknowns,omitempty"`
	Obligations      []Obligation             `json:"obligations,omitempty"`
	Assumptions      []string                 `json:"assumptions,omitempty"`
	NativeBoundaries []NativeBoundary         `json:"native_boundaries,omitempty"`
}

type DerivedSummary struct {
	Paths            []PathSummary    `json:"paths"`
	Effects          []Effect         `json:"effects,omitempty"`
	Unknowns         []UnknownReason  `json:"unknowns,omitempty"`
	Coverage         Coverage         `json:"coverage"`
	Resources        ResourceCost     `json:"resources"`
	NativeBoundaries []NativeBoundary `json:"native_boundaries,omitempty"`
	Assumptions      []string         `json:"assumptions,omitempty"`
	Obligations      []Obligation     `json:"obligations,omitempty"`
	Fingerprint      string           `json:"fingerprint"`
}

// ApprovedAssumptions is supplied by the evidence/trust checker, never loaded
// from a native declaration. A witness must be explicitly approved to be used.
type AnalysisOptions struct {
	MaxAlternatives     int
	WidenAfter          int
	MaxSCCRounds        int
	Aliases             []ResourcePair
	Disjoint            []ResourcePair
	NativeSummaries     map[string]DerivedSummary
	NativeWitnesses     map[string]string
	ApprovedAssumptions map[string]bool
}

func (o AnalysisOptions) normalized() AnalysisOptions {
	if o.MaxAlternatives <= 0 {
		o.MaxAlternatives = 32
	}
	if o.WidenAfter <= 0 {
		o.WidenAfter = 8
	}
	if o.MaxSCCRounds <= 0 {
		o.MaxSCCRounds = 64
	}
	return o
}

type inference struct {
	plan          Plan
	options       AnalysisOptions
	summaries     map[MeaningKey]DerivedSummary
	callbackStack map[string]bool
	callStack     map[MeaningKey]bool
	summarizing   bool
}

func initialPath() PathSummary {
	return PathSummary{Outcome: Normal, Reachable: true, Demand: make(map[string]Multiplicity), Owned: make(map[string]OwnedResource), Resources: ZeroResourceCost()}
}

func clonePath(p PathSummary) PathSummary {
	p.Facts = append([]Predicate(nil), p.Facts...)
	p.Effects = append([]Effect(nil), p.Effects...)
	p.Unknowns = append([]UnknownReason(nil), p.Unknowns...)
	p.Obligations = append([]Obligation(nil), p.Obligations...)
	p.Assumptions = append([]string(nil), p.Assumptions...)
	p.NativeBoundaries = append([]NativeBoundary(nil), p.NativeBoundaries...)
	p.OwnershipOrder = append([]string(nil), p.OwnershipOrder...)
	demand := make(map[string]Multiplicity, len(p.Demand))
	for key, value := range p.Demand {
		demand[key] = value
	}
	p.Demand = demand
	owned := make(map[string]OwnedResource, len(p.Owned))
	for key, value := range p.Owned {
		owned[key] = value
	}
	p.Owned = owned
	return p
}

func Infer(plan Plan, options AnalysisOptions) (DerivedSummary, error) {
	options = options.normalized()
	for _, pair := range options.Disjoint {
		if AliasRelation(pair.Left, pair.Right, options.Aliases, nil) == Established {
			return DerivedSummary{}, fmt.Errorf("resource %s is both aliased and declared disjoint from %s", pair.Left.ID, pair.Right.ID)
		}
	}
	a := &inference{plan: plan, options: options, summaries: make(map[MeaningKey]DerivedSummary), callbackStack: make(map[string]bool), callStack: make(map[MeaningKey]bool)}
	if err := a.inferDefinitions(); err != nil {
		return DerivedSummary{}, err
	}
	paths, err := a.run(plan.Entry, []PathSummary{initialPath()})
	if err != nil {
		return DerivedSummary{}, err
	}
	for i := range paths {
		paths[i].Unknowns = append(paths[i].Unknowns, plan.Residuals...)
		for _, boundary := range plan.NativeClosure {
			if boundary.Status == Established {
				if boundary.Witness == "" || !options.ApprovedAssumptions[boundary.Witness] {
					boundary.Status = Unresolved
					boundary.Reason = "native closure assertion has no explicitly approved conformance witness"
				} else {
					paths[i].Assumptions = append(paths[i].Assumptions, boundary.Witness)
				}
			}
			paths[i].NativeBoundaries = append(paths[i].NativeBoundaries, boundary)
			if boundary.Status != Established {
				paths[i].Unknowns = append(paths[i].Unknowns, UnknownReason{Code: "native_closure", Message: boundary.Reason, Boundary: boundary.ID, Origin: boundary.Origin, Components: []string{"effects", "outcomes", "costs", "ownership"}})
			}
		}
		if len(plan.Residuals) > 0 || len(plan.NativeClosure) > 0 {
			for _, unknown := range paths[i].Unknowns {
				for _, component := range unknown.Components {
					if component == "costs" || component == "resources" {
						paths[i].Resources = UnknownResourceCost(unknown.Message)
					}
				}
			}
		}
	}
	result := summarize(paths)
	result.Fingerprint = StableID("inference", struct {
		Plan                        Plan
		Policy                      string
		Alternatives, Widen, Rounds int
		Aliases, Disjoint           []ResourcePair
		Native                      map[string]DerivedSummary
		Witnesses                   map[string]string
		Approved                    map[string]bool
	}{plan, AnalyzerVersion, options.MaxAlternatives, options.WidenAfter, options.MaxSCCRounds, options.Aliases, options.Disjoint, options.NativeSummaries, options.NativeWitnesses, options.ApprovedAssumptions})
	return result, nil
}

func (a *inference) sequence(nodes []Node, paths []PathSummary) ([]PathSummary, error) {
	var err error
	for _, node := range nodes {
		paths, err = a.run(node, paths)
		if err != nil {
			return nil, err
		}
		paths = a.bound(paths)
	}
	return paths, nil
}

func (a *inference) run(node Node, paths []PathSummary) ([]PathSummary, error) {
	var result []PathSummary
	for _, path := range paths {
		if path.Outcome != Normal {
			result = append(result, path)
			continue
		}
		out, err := a.step(node, clonePath(path))
		if err != nil {
			return nil, err
		}
		result = append(result, out...)
	}
	return a.bound(result), nil
}

func (a *inference) step(node Node, p PathSummary) ([]PathSummary, error) {
	if node.Kind != Require && len(node.Requires) > 0 {
		checked, err := a.step(Node{Kind: Require, Requires: node.Requires, Target: node.Target, Phase: node.Phase, Origin: node.Origin}, p)
		if err != nil {
			return nil, err
		}
		node.Requires = nil
		return a.run(node, checked)
	}
	switch node.Kind {
	case Sequence:
		return a.sequence(node.Children, []PathSummary{p})
	case Branch:
		left, err := a.sequence(node.Then, []PathSummary{clonePath(p)})
		if err != nil {
			return nil, err
		}
		right, err := a.sequence(node.Else, []PathSummary{p})
		return append(left, right...), err
	case Event:
		a.applyEvent(&p, node)
		if node.Outcome != "" {
			p.Outcome = node.Outcome
			switch node.Outcome {
			case Normal, ReturnedFailure, GuardRejected, Panic, Abort, Cancelled, ProcessFailure:
			default:
				a.unknown(&p, node, "unsupported_outcome", "typed event has an unsupported outcome classification")
			}
		}
		p.Facts = postFacts(p.Facts, node.Facts)
		return []PathSummary{p}, nil
	case Mutation:
		if node.MutationFacts == nil {
			a.unknown(&p, node, "representation", "mutation has no selected representation contract")
			return []PathSummary{p}, nil
		}
		facts := node.MutationFacts
		effects := append([]Effect(nil), facts.NestedEffects...)
		effects = append(effects, Effect{Kind: "write", Resource: node.Resource, Target: node.Target, Phase: node.Phase, Origin: node.Origin, Definite: true})
		if facts.CopyOnWrite && facts.Sharing != Contradicted {
			effects = append(effects, Effect{Kind: "allocate", Resource: node.Resource, Target: node.Target, Phase: node.Phase, Origin: node.Origin, Definite: facts.Sharing == Established, Condition: "shared backing materialization"})
		}
		if facts.CapacityEnough != Established {
			effects = append(effects, Effect{Kind: "allocate", Resource: node.Resource, Target: node.Target, Phase: node.Phase, Origin: node.Origin, Definite: facts.CapacityEnough == Contradicted, Condition: "capacity growth"})
		}
		node.Effects = append(effects, node.Effects...)
		a.applyEvent(&p, node)
		if !facts.NestedClosed {
			a.unknown(&p, node, "nested_mutation", "nested mutation effects and costs are unknown")
		}
		return []PathSummary{p}, nil
	case Require:
		for _, want := range node.Requires {
			status, reason := Entails(p.Facts, want)
			p.Obligations = append(p.Obligations, Obligation{ID: StableID("require", struct {
				Origin    Origin
				Predicate Predicate
			}{node.Origin, want}), Claim: Claim{Predicate: want, Target: node.Target, Scope: "use", Phase: node.Phase, Outcomes: []Outcome{Normal}}, Status: status, Reason: reason, Origin: node.Origin})
		}
		return []PathSummary{p}, nil
	case Guard:
		a.applyEvent(&p, node)
		all, impossible := true, false
		for _, fact := range node.Facts {
			status, _ := Entails(p.Facts, fact)
			all = all && status == Established
			impossible = impossible || status == Contradicted
		}
		var result []PathSummary
		if !impossible {
			success := clonePath(p)
			success.Facts = append(success.Facts, node.Facts...)
			result = append(result, success)
		}
		if !all {
			p.Outcome = GuardRejected
			result = append(result, p)
		}
		return result, nil
	case Native:
		witness := a.options.NativeWitnesses[node.Native]
		if summary, ok := a.options.NativeSummaries[node.Native]; ok && witness != "" && a.options.ApprovedAssumptions[witness] {
			out := a.applySummary(p, summary, node.Substitutions)
			for i := range out {
				out[i].Assumptions = append(out[i].Assumptions, witness)
				out[i].NativeBoundaries = append(out[i].NativeBoundaries, NativeBoundary{ID: node.Native, Meaning: node.Meaning, Target: node.Target, Origin: node.Origin, Status: Established, Reason: "conditional on explicitly approved native conformance witness", Witness: witness, Implicit: node.Phase == Destruction || node.Phase == Cleanup})
			}
			return out, nil
		}
		a.unknown(&p, node, "native_contract", "native behavior has no approved implementation-conformance witness")
		p.NativeBoundaries = append(p.NativeBoundaries, NativeBoundary{ID: node.Native, Meaning: node.Meaning, Target: node.Target, Origin: node.Origin, Status: Unresolved, Reason: "missing native contract or approved conformance witness", Implicit: node.Phase == Destruction || node.Phase == Cleanup})
		return []PathSummary{p}, nil
	case Call:
		actuals, err := a.sequence(node.Children, []PathSummary{p})
		if err != nil {
			return nil, err
		}
		var out []PathSummary
		for _, path := range actuals {
			if path.Outcome != Normal {
				out = append(out, path)
				continue
			}
			if body, ok := a.plan.Definitions[node.Meaning]; ok && !a.summarizing && !a.callStack[node.Meaning] {
				a.callStack[node.Meaning] = true
				called, err := a.run(rebaseNode(body, node.Substitutions), []PathSummary{path})
				delete(a.callStack, node.Meaning)
				if err != nil {
					return nil, err
				}
				out = append(out, called...)
				continue
			}
			if summary, ok := a.summaries[node.Meaning]; ok {
				out = append(out, a.applySummary(path, summary, node.Substitutions)...)
			} else {
				a.unknown(&path, node, "missing_definition", "selected callee has no typed body")
				out = append(out, path)
			}
		}
		return out, nil
	case Callback:
		one := int64(1)
		d := demandAt(p.Demand, node.Callback)
		d = addDemand(d, Multiplicity{Lower: 1, Upper: &one})
		p.Demand[node.Callback] = d
		if body, ok := a.plan.Callbacks[node.Callback]; ok && !a.callbackStack[node.Callback] {
			a.callbackStack[node.Callback] = true
			out, err := a.run(body, []PathSummary{p})
			delete(a.callbackStack, node.Callback)
			return out, err
		}
		a.unknown(&p, node, "callback_effects", "callback behavior, escape, and cost are not established")
		return []PathSummary{p}, nil
	case Acquire:
		if node.Resource.ID == "" {
			return nil, fmt.Errorf("acquire requires a resource identity")
		}
		if old, ok := p.Owned[node.Resource.ID]; ok && old.State == "owned" {
			return nil, fmt.Errorf("acquire replaces initialized resource %s without an explicit release", node.Resource.ID)
		}
		a.applyEvent(&p, node)
		p.Owned[node.Resource.ID] = OwnedResource{Resource: node.Resource, Mode: node.Ownership, State: "owned", SuspensionSafe: node.SuspensionSafe, FinalOwner: node.FinalOwner, Destructor: node.Cleanup, Origin: node.Origin}
		p.OwnershipOrder = append(p.OwnershipOrder, node.Resource.ID)
		return a.sequence(node.Children, []PathSummary{p})
	case Move:
		owned, ok := p.Owned[node.Resource.ID]
		if !ok || owned.State != "owned" {
			return nil, fmt.Errorf("move requires an initialized owned resource %s", node.Resource.ID)
		}
		owned.State = "moved"
		p.Owned[node.Resource.ID] = owned
		a.applyEvent(&p, node)
		return []PathSummary{p}, nil
	case Release:
		return a.release(p, node)
	case Scope:
		before := make(map[string]bool)
		for id, owned := range p.Owned {
			before[id] = owned.State == "owned"
		}
		paths, err := a.sequence(node.Children, []PathSummary{p})
		if err != nil {
			return nil, err
		}
		var out []PathSummary
		for _, path := range paths {
			covered := path.Outcome == Normal
			if len(node.CleanupOutcomes) > 0 {
				covered = containsOutcome(node.CleanupOutcomes, path.Outcome)
			}
			if path.Outcome == Abort || !covered {
				out = append(out, path)
				continue
			}
			outcome := path.Outcome
			path.Outcome = Normal
			closed := []PathSummary{path}
			for i := len(path.OwnershipOrder) - 1; i >= 0; i-- {
				id := path.OwnershipOrder[i]
				if before[id] {
					continue
				}
				closed, err = a.run(Node{Kind: Release, Resource: ResourceRef{ID: id}, Origin: node.Origin, Phase: Cleanup}, closed)
				if err != nil {
					return nil, err
				}
			}
			closed, err = a.sequence(node.Cleanup, closed)
			if err != nil {
				return nil, err
			}
			for i := range closed {
				if closed[i].Outcome == Normal {
					closed[i].Outcome = outcome
				}
			}
			out = append(out, closed...)
		}
		return out, nil
	case Repeat:
		return a.repeat(node, p)
	default:
		return nil, fmt.Errorf("unsupported typed contract node kind %q", node.Kind)
	}
}

func (a *inference) applyEvent(p *PathSummary, node Node) {
	if node.Costs != nil {
		if allocated, ok := node.Costs.Allocated.Constant(); ok && allocated > 0 {
			present := false
			for _, effect := range node.Effects {
				present = present || effect.Kind == "allocate"
			}
			if !present {
				node.Effects = append(append([]Effect(nil), node.Effects...), Effect{Kind: "allocate", Resource: node.Resource, Target: node.Target, Phase: node.Phase, Definite: true, Origin: node.Origin})
			}
		}
	}
	for _, effect := range node.Effects {
		switch effect.Kind {
		case "read", "write", "replace", "release", "allocate", "block", "io", "panic", "spawn", "signal", "wait", "suspend", "retain", "escape", "close":
		default:
			a.unknown(p, node, "unsupported_effect", "typed event has an unsupported effect classification: "+string(effect.Kind))
		}
		if effect.Phase == "" {
			effect.Phase = node.Phase
		}
		if effect.Origin.Source == "" {
			effect.Origin = node.Origin
		}
		if !p.Reachable {
			effect.Definite = false
		}
		p.Effects = append(p.Effects, effect)
		if effect.Kind == "write" || effect.Kind == "replace" || effect.Kind == "release" {
			a.invalidate(p, effect.Resource)
			a.invalidateStructural(p, effect.Target, string(effect.Kind))
		}
		if effect.Kind == "suspend" {
			for _, owned := range p.Owned {
				if owned.State == "owned" && !owned.SuspensionSafe && owned.Mode == "guard" {
					p.Obligations = append(p.Obligations, Obligation{ID: StableID("suspension", owned.Resource), Claim: Claim{Predicate: Predicate{Kind: "suspension_safe", Resource: owned.Resource}, Scope: "whole_region"}, Status: Contradicted, Reason: "non-suspension-safe guard is live across suspension", Origin: effect.Origin})
				}
			}
		}
	}
	if node.Costs != nil {
		p.Resources = SequenceResources(p.Resources, *node.Costs)
	} else if len(node.Effects) > 0 {
		cost := ZeroResourceCost()
		cost.Work = UnknownCost("event work bound was not supplied")
		for _, effect := range node.Effects {
			switch effect.Kind {
			case "allocate":
				cost.Allocated = UnknownCost("allocation quantity was not supplied")
				cost.NetLive = UnknownCost("allocation live change was not supplied")
				cost.PeakLive = UnknownCost("allocation peak was not supplied")
			case "release":
				cost.NetLive = UnknownCost("release quantity was not supplied")
			case "io":
				cost.IO = UnknownCost("IO volume was not supplied")
			}
		}
		p.Resources = SequenceResources(p.Resources, cost)
	}
}

func (a *inference) invalidate(p *PathSummary, resource ResourceRef) {
	var facts []Predicate
	for _, fact := range p.Facts {
		if fact.Resource.ID == "" && !fact.Resource.Unknown || AliasRelation(fact.Resource, resource, a.options.Aliases, a.options.Disjoint) == Contradicted {
			facts = append(facts, fact)
		}
	}
	p.Facts = facts
}

func (a *inference) invalidateStructural(p *PathSummary, target TargetRef, kind string) {
	var facts []Predicate
	for _, fact := range p.Facts {
		if !structuralFactAffected(fact, target, kind) {
			facts = append(facts, fact)
		}
	}
	p.Facts = facts
}

func (a *inference) unknown(p *PathSummary, node Node, code, message string) {
	p.Unknowns = append(p.Unknowns, UnknownReason{Code: code, Message: message, Boundary: node.Native, Origin: node.Origin, Phase: node.Phase, Components: []string{"effects", "outcomes", "ownership", "costs"}})
	p.Resources = SequenceResources(p.Resources, UnknownResourceCost(message))
	a.invalidate(p, ResourceRef{Unknown: true})
	a.invalidateStructural(p, TargetRef{Unknown: true}, "unknown")
	for _, owned := range p.Owned {
		if owned.State == "owned" && (owned.Mode == "exclusive" || owned.Mode == "guard") {
			p.Obligations = append(p.Obligations, Obligation{ID: StableID("unknown-interference", struct {
				Resource ResourceRef
				Origin   Origin
			}{owned.Resource, node.Origin}), Claim: Claim{Predicate: Predicate{Kind: "ownership", Resource: owned.Resource}, Scope: "whole_region"}, Status: Unresolved, Reason: "unknown behavior may interfere with a live exclusive resource or guard", Origin: node.Origin})
		}
	}
}

func (a *inference) release(p PathSummary, node Node) ([]PathSummary, error) {
	owned, ok := p.Owned[node.Resource.ID]
	if !ok || owned.State != "owned" && owned.State != "conditional" {
		return []PathSummary{p}, nil
	}
	if owned.State == "conditional" {
		missing := clonePath(p)
		owned.State = "owned"
		p.Owned[node.Resource.ID] = owned
		possible, err := a.release(p, node)
		for i := range possible {
			possible[i].Reachable = false
			for j := range possible[i].Effects {
				possible[i].Effects[j].Definite = false
			}
		}
		return append([]PathSummary{missing}, possible...), err
	}
	owned.State = "released"
	p.Owned[node.Resource.ID] = owned
	effect := Effect{Kind: "release", Resource: owned.Resource, Phase: node.Phase, Definite: p.Reachable, Origin: node.Origin}
	if effect.Phase == "" {
		effect.Phase = Destruction
	}
	p.Effects = append(p.Effects, effect)
	a.invalidate(&p, owned.Resource)
	if node.Costs != nil {
		p.Resources = SequenceResources(p.Resources, *node.Costs)
	}
	if owned.FinalOwner == Contradicted {
		return []PathSummary{p}, nil
	}
	if len(owned.Destructor) == 0 {
		owned.Destructor = []Node{{Kind: Native, Native: "destructor:" + owned.Resource.ID, Resource: owned.Resource, Phase: Destruction, Origin: owned.Origin}}
	}
	if owned.FinalOwner == Established || owned.Mode != "shared" {
		return a.sequence(owned.Destructor, []PathSummary{p})
	}
	with, err := a.sequence(owned.Destructor, []PathSummary{clonePath(p)})
	for i := range with {
		for j := len(p.Effects); j < len(with[i].Effects); j++ {
			with[i].Effects[j].Definite = false
			with[i].Effects[j].Condition = "final owner of " + owned.Resource.ID
		}
		with[i].Reachable = false
	}
	return append([]PathSummary{p}, with...), err
}

func (a *inference) repeat(node Node, p PathSummary) ([]PathSummary, error) {
	if node.Count != nil && *node.Count < 0 {
		return nil, fmt.Errorf("negative repetition count")
	}
	if node.Count != nil && *node.Count <= 1024 {
		paths := []PathSummary{p}
		var err error
		for i := int64(0); i < *node.Count; i++ {
			paths, err = a.sequence(node.Children, paths)
			if err != nil {
				return nil, err
			}
		}
		return paths, nil
	}
	body, err := a.sequence(node.Children, []PathSummary{initialPath()})
	if err != nil {
		return nil, err
	}
	summary := summarize(body)
	if node.Count != nil && node.StateInvariant && len(body) == 1 && body[0].Outcome == Normal && len(body[0].Owned) == 0 {
		summary.Paths[0].Resources = RepeatResources(body[0].Resources, *node.Count)
		for key, demand := range summary.Paths[0].Demand {
			demand.Lower = 0
			if demand.Upper != nil {
				n, ok := MultiplyCost(ConstantCost(*demand.Upper), ConstantCost(*node.Count)).Constant()
				if ok {
					demand.Upper = &n
				} else {
					demand.Upper = nil
				}
			}
			summary.Paths[0].Demand[key] = demand
		}
		return a.applySummary(p, summary, node.Substitutions), nil
	}
	a.unknown(&p, node, "iteration_bound", "iteration count or evolving callback state has no supported finite bound")
	p.Effects = append(p.Effects, summary.Effects...)
	for i := range p.Effects {
		p.Effects[i].Definite = false
	}
	p.Resources.Work = EvolvingCallbackWork(node.CountSymbol, SymbolicCost(node.CountSymbol))
	for _, path := range body {
		for key := range path.Demand {
			p.Demand[key] = Multiplicity{Symbol: node.CountSymbol}
		}
	}
	return []PathSummary{p}, nil
}

func addDemand(a, b Multiplicity) Multiplicity {
	lower, ok := addInteger(a.Lower, b.Lower)
	if !ok {
		lower = 0
	}
	result := Multiplicity{Lower: lower, Symbol: a.Symbol + b.Symbol}
	if a.Upper != nil && b.Upper != nil {
		value, ok := addInteger(*a.Upper, *b.Upper)
		if ok {
			result.Upper = &value
		}
	}
	return result
}

func demandAt(demands map[string]Multiplicity, name string) Multiplicity {
	if demand, ok := demands[name]; ok {
		return demand
	}
	zero := int64(0)
	return Multiplicity{Upper: &zero}
}

func (a *inference) applySummary(p PathSummary, summary DerivedSummary, sub Substitutions) []PathSummary {
	if len(summary.Paths) == 0 {
		a.unknown(&p, Node{}, "missing_summary", "callee has no outcome summary")
		return []PathSummary{p}
	}
	var out []PathSummary
	for _, tail := range summary.Paths {
		path := clonePath(p)
		path.Outcome = tail.Outcome
		path.Reachable = path.Reachable && tail.Reachable
		for _, unknown := range tail.Unknowns {
			for _, component := range unknown.Components {
				if component == "effects" || component == "ownership" {
					a.invalidate(&path, ResourceRef{Unknown: true})
					a.invalidateStructural(&path, TargetRef{Unknown: true}, "unknown")
				}
			}
		}
		for _, effect := range tail.Effects {
			effect.Resource = rebaseResource(effect.Resource, sub)
			effect.Target = rebaseTarget(effect.Target, sub)
			if !path.Reachable {
				effect.Definite = false
			}
			path.Effects = append(path.Effects, effect)
			if effect.Kind == "write" || effect.Kind == "replace" || effect.Kind == "release" {
				a.invalidate(&path, effect.Resource)
				a.invalidateStructural(&path, effect.Target, string(effect.Kind))
			}
		}
		for _, fact := range tail.Facts {
			path.Facts = postFacts(path.Facts, []Predicate{rebasePredicate(fact, sub)})
		}
		for key, demand := range tail.Demand {
			path.Demand[key] = addDemand(demandAt(path.Demand, key), demand)
		}
		path.Resources = SequenceResources(path.Resources, tail.Resources)
		path.Unknowns = append(path.Unknowns, tail.Unknowns...)
		path.Obligations = append(path.Obligations, tail.Obligations...)
		path.Assumptions = append(path.Assumptions, tail.Assumptions...)
		path.NativeBoundaries = append(path.NativeBoundaries, tail.NativeBoundaries...)
		for _, key := range tail.OwnershipOrder {
			owned := tail.Owned[key]
			owned.Resource = rebaseResource(owned.Resource, sub)
			path.Owned[owned.Resource.ID] = owned
			path.OwnershipOrder = append(path.OwnershipOrder, owned.Resource.ID)
		}
		out = append(out, path)
	}
	return out
}

func rebaseResource(resource ResourceRef, sub Substitutions) ResourceRef {
	if mapped, ok := sub.Resources[resource.ID]; ok {
		return mapped
	}
	return resource
}
func rebaseTarget(target TargetRef, sub Substitutions) TargetRef {
	if mapped, ok := sub.Targets[target.Root]; ok {
		mapped.Path = append(append([]string(nil), mapped.Path...), target.Path...)
		// Substituting the enclosing attachment does not turn a child role into
		// that enclosing type. The child's canonical nominal remains its own.
		if len(target.Path) != 0 {
			mapped.Type, mapped.Nominal = target.Type, target.Nominal
		}
		mapped.Unknown = mapped.Unknown || target.Unknown
		return mapped
	}
	return target
}
func rebasePredicate(fact Predicate, sub Substitutions) Predicate {
	fact.Resource = rebaseResource(fact.Resource, sub)
	fact.Target = rebaseTarget(fact.Target, sub)
	if value, ok := sub.Values[fact.Subject]; ok {
		fact.Value = value
	}
	fact.Args = append([]Predicate(nil), fact.Args...)
	for i := range fact.Args {
		fact.Args[i] = rebasePredicate(fact.Args[i], sub)
	}
	return fact
}

func rebaseNode(node Node, sub Substitutions) Node {
	node.Resource = rebaseResource(node.Resource, sub)
	node.Target = rebaseTarget(node.Target, sub)
	node.Facts = append([]Predicate(nil), node.Facts...)
	for i := range node.Facts {
		node.Facts[i] = rebasePredicate(node.Facts[i], sub)
	}
	node.Requires = append([]Predicate(nil), node.Requires...)
	for i := range node.Requires {
		node.Requires[i] = rebasePredicate(node.Requires[i], sub)
	}
	node.Effects = append([]Effect(nil), node.Effects...)
	for i := range node.Effects {
		node.Effects[i].Resource = rebaseResource(node.Effects[i].Resource, sub)
		node.Effects[i].Target = rebaseTarget(node.Effects[i].Target, sub)
	}
	children := func(nodes []Node) []Node {
		out := make([]Node, len(nodes))
		for i := range nodes {
			out[i] = rebaseNode(nodes[i], sub)
		}
		return out
	}
	node.Children = children(node.Children)
	node.Then = children(node.Then)
	node.Else = children(node.Else)
	node.Cleanup = children(node.Cleanup)
	return node
}

func containsOutcome(outcomes []Outcome, want Outcome) bool {
	for _, outcome := range outcomes {
		if outcome == want {
			return true
		}
	}
	return false
}

// Postconditions describe the new state. Replacing an exact scalar/state fact
// must not leave its previous value available as a simultaneous proof premise.
func postFacts(before, after []Predicate) []Predicate {
	result := append([]Predicate(nil), before...)
	for _, fact := range after {
		var retained []Predicate
		for _, old := range result {
			status, _ := Entails([]Predicate{old}, fact)
			if status == Contradicted && old.Subject == fact.Subject {
				continue
			}
			retained = append(retained, old)
		}
		result = append(retained, fact)
	}
	return result
}

func summarize(paths []PathSummary) DerivedSummary {
	result := DerivedSummary{Paths: paths, Coverage: Known, Resources: ZeroResourceCost()}
	for i, path := range paths {
		result.Effects = append(result.Effects, path.Effects...)
		result.Unknowns = append(result.Unknowns, path.Unknowns...)
		result.Assumptions = append(result.Assumptions, path.Assumptions...)
		result.Obligations = append(result.Obligations, path.Obligations...)
		result.NativeBoundaries = append(result.NativeBoundaries, path.NativeBoundaries...)
		if i == 0 {
			result.Resources = path.Resources
		} else {
			result.Resources = BranchResources(result.Resources, path.Resources)
		}
	}
	result.Effects = uniqueEffects(result.Effects)
	result.Unknowns = uniqueUnknowns(result.Unknowns)
	result.Assumptions = sortedStrings(result.Assumptions)
	if len(result.Unknowns) > 0 {
		result.Coverage = Partial
	}
	return result
}

func uniqueEffects(effects []Effect) []Effect {
	seen := make(map[string]bool)
	var out []Effect
	for _, effect := range effects {
		key := StableID("effect", effect)
		if !seen[key] {
			seen[key] = true
			out = append(out, effect)
		}
	}
	return out
}
func uniqueUnknowns(unknowns []UnknownReason) []UnknownReason {
	seen := make(map[string]bool)
	var out []UnknownReason
	for _, unknown := range unknowns {
		key := StableID("unknown", unknown)
		if !seen[key] {
			seen[key] = true
			out = append(out, unknown)
		}
	}
	return out
}

func (a *inference) bound(paths []PathSummary) []PathSummary {
	if len(paths) <= a.options.MaxAlternatives {
		return paths
	}
	groups := make(map[Outcome][]PathSummary)
	for _, path := range paths {
		groups[path.Outcome] = append(groups[path.Outcome], path)
	}
	keys := make([]string, 0, len(groups))
	for outcome := range groups {
		keys = append(keys, string(outcome))
	}
	sort.Strings(keys)
	var out []PathSummary
	for _, key := range keys {
		group := groups[Outcome(key)]
		merged := clonePath(group[0])
		merged.Reachable = false
		for _, path := range group[1:] {
			var common []Predicate
			for _, fact := range merged.Facts {
				if status, _ := Entails(path.Facts, fact); status == Established {
					common = append(common, fact)
				}
			}
			merged.Facts = common
			merged.Effects = append(merged.Effects, path.Effects...)
			merged.Unknowns = append(merged.Unknowns, path.Unknowns...)
			merged.Resources = BranchResources(merged.Resources, path.Resources)
			merged.Obligations = append(merged.Obligations, path.Obligations...)
			merged.Assumptions = append(merged.Assumptions, path.Assumptions...)
			merged.NativeBoundaries = append(merged.NativeBoundaries, path.NativeBoundaries...)
			for id, owned := range merged.Owned {
				other, ok := path.Owned[id]
				if !ok || other.State != owned.State {
					owned.State = "conditional"
					merged.Owned[id] = owned
				}
			}
			for id, owned := range path.Owned {
				if _, ok := merged.Owned[id]; !ok {
					owned.State = "conditional"
					merged.Owned[id] = owned
					merged.OwnershipOrder = append(merged.OwnershipOrder, id)
				}
			}
			for callback, demand := range merged.Demand {
				other := path.Demand[callback]
				if other.Lower < demand.Lower {
					demand.Lower = other.Lower
				}
				if other.Upper == nil || demand.Upper == nil {
					demand.Upper = nil
				} else if *other.Upper > *demand.Upper {
					demand.Upper = other.Upper
				}
				merged.Demand[callback] = demand
			}
			for callback, demand := range path.Demand {
				if _, ok := merged.Demand[callback]; !ok {
					demand.Lower = 0
					merged.Demand[callback] = demand
				}
			}
		}
		for i := range merged.Effects {
			merged.Effects[i].Definite = false
		}
		merged.Effects = uniqueEffects(merged.Effects)
		merged.Unknowns = append(uniqueUnknowns(merged.Unknowns), UnknownReason{Code: "alternative_budget", Message: "bounded disjunction widened; only common facts and conservative behavior retained", Components: []string{"relations", "ownership"}})
		out = append(out, merged)
	}
	return out
}
