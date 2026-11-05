package contracts

import (
	"fmt"
	"reflect"
	"sort"
)

type NativeComponent struct {
	Use       string               `json:"use"`
	Component string               `json:"component"`
	Outcome   Outcome              `json:"outcome,omitempty"`
	Claim     Claim                `json:"claim"`
	Current   EvidenceFingerprints `json:"current"`
	Status    Status               `json:"status"`
	Reason    string               `json:"reason"`
	Witness   string               `json:"witness,omitempty"`
}

// EvaluateNativeUses evaluates the selected native closure once. Missing typed
// uses remain unresolved declarations; no binding or display name creates one.
func EvaluateNativeUses(uses []NativeUse, declarations []DeclaredContract, environment NativeEnvironment, policy TrustPolicy, witnesses []NativeWitness, applicability Status, selected []LoweringStep) []NativeEvaluation {
	ordered := append([]NativeWitness(nil), witnesses...)
	sort.Slice(ordered, func(i, j int) bool { return ordered[i].ID < ordered[j].ID })
	context := nativeEvaluationContext{policy: policy, witnesses: make(map[string][]NativeWitness)}
	context.fingerprint = StableID("native-assurance-context", struct {
		Environment NativeEnvironment
		Policy      TrustPolicy
		Witnesses   []NativeWitness
	}{environment, policy, ordered})
	for _, witness := range ordered {
		context.witnesses[witness.ClaimFingerprint] = append(context.witnesses[witness.ClaimFingerprint], witness)
	}
	validEnvironment := applicability == Established && environment.Validate() == nil && (environment.Method != "test-only" || policy.AllowTestOnly)
	inputs := make(map[string]string)
	for name, value := range environment.Inputs {
		inputs["source:"+name] = value
	}
	for name, value := range environment.Dependencies {
		inputs["lowering:"+name] = value
	}
	byMeaning := make(map[MeaningKey][]NativeUse)
	dependencies := make(map[string]bool)
	for _, use := range uses {
		byMeaning[use.Meaning] = append(byMeaning[use.Meaning], use)
		dependencies[string(use.Meaning)] = dependencies[string(use.Meaning)] || use.Implementation != ""
	}
	for _, step := range selected {
		dependencies[string(step.Meaning)] = dependencies[string(step.Meaning)] || step.Implementation != ""
	}
	var evaluations []NativeEvaluation
	for _, declaration := range declarations {
		valid := validEnvironment && (NativeContractManifest{SchemaVersion: SchemaVersion, Kind: "native_contract_manifest", Producer: AnalyzerVersion, Declarations: []DeclaredContract{declaration}}).Validate() == nil
		for _, dependency := range declaration.Dependencies {
			valid = valid && dependencies[dependency]
		}
		contract := StableID("native-contract", declaration)
		matching := byMeaning[declaration.Meaning]
		if len(matching) == 0 {
			matching = []NativeUse{{Meaning: declaration.Meaning, Target: TargetRef{Unknown: true}, Origin: declaration.Origin}}
		}
		for _, use := range matching {
			current := EvidenceFingerprints{Contract: contract, Implementation: use.Implementation, Target: environment.Fingerprint, Dependencies: make(map[string]string)}
			for name, value := range inputs {
				current.Dependencies[name] = value
			}
			for name, value := range use.Dependencies {
				current.Dependencies["use:"+name] = value
			}
			current.Dependencies["application"] = StableID("native-use", use)
			evaluations = append(evaluations, context.evaluate(use, declaration, current, valid))
		}
	}
	return evaluations
}

type nativeEvaluationContext struct {
	policy      TrustPolicy
	witnesses   map[string][]NativeWitness
	fingerprint string
}

func nativeComponentClaims(use NativeUse, declaration DeclaredContract, current EvidenceFingerprints) []NativeComponent {
	var result []NativeComponent
	add := func(component string, outcome Outcome, payload any) {
		outcomes := declaration.Outcomes
		if outcome != "" {
			outcomes = []Outcome{outcome}
		}
		claim := Claim{Predicate: Predicate{Kind: "native_component", Subject: component, Value: Atom{Kind: "fingerprint", Text: StableID("component", payload)}}, Target: use.Target, Scope: use.Boundary, Phase: use.Phase, Outcomes: outcomes}
		claim.ID = StableID("native-component", struct {
			Use     string
			Meaning MeaningKey
			Claim   Claim
		}{use.ID, declaration.Meaning, claim})
		result = append(result, NativeComponent{Use: use.ID, Component: component, Outcome: outcome, Claim: claim, Current: current, Status: Unresolved, Reason: "component has no applicable approved witness"})
	}
	if len(declaration.Outcomes) > 0 {
		add("outcomes", "", declaration.Outcomes)
	}
	for _, outcome := range declaration.Outcomes {
		contract := nativeOutcome(declaration, outcome)
		if contract.Effects != nil {
			add("effects", outcome, contract.Effects)
		}
		for _, kind := range contract.AbsentEffects {
			add("absent:"+string(kind), outcome, kind)
		}
		if contract.Costs != nil {
			add("costs", outcome, contract.Costs)
		}
		for i, state := range contract.States {
			add(fmt.Sprintf("state:%d", i), outcome, state)
		}
	}
	return result
}

func nativeOutcome(declaration DeclaredContract, outcome Outcome) NativeOutcomeContract {
	for _, item := range declaration.OutcomeContracts {
		if item.Outcome == outcome {
			return item
		}
	}
	return NativeOutcomeContract{Outcome: outcome, Effects: declaration.Effects, Costs: declaration.Costs}
}

// evaluate prepares conditional summary components and report-only guarantees.
// Actual entry premises are discharged later by inference, not by metadata.
func (context nativeEvaluationContext) evaluate(use NativeUse, declaration DeclaredContract, current EvidenceFingerprints, valid bool) NativeEvaluation {
	application := NativeEvaluation{Use: use, Requirements: append(append([]Predicate(nil), declaration.Requirements...), declaration.TargetRestrictions...), ContextFingerprint: context.fingerprint}
	application.Entry = Claim{Predicate: Predicate{Kind: "native_entry_requirements", Subject: use.ID, Value: Atom{Kind: "fingerprint", Text: StableID("requirements", application.Requirements)}}, Target: use.Target, Scope: "invocation_entry", Phase: use.Phase}
	application.Entry.ID = StableID("native-entry-applicability", struct {
		Meaning MeaningKey
		Claim   Claim
	}{declaration.Meaning, application.Entry})
	application.Applicability = Obligation{ID: StableID("native-use-applicability", struct {
		Use     string
		Meaning MeaningKey
	}{use.ID, declaration.Meaning}), Claim: Claim{Predicate: Predicate{Kind: "native_use_applicability", Subject: use.ID}, Target: use.Target, Scope: use.Boundary, Phase: use.Phase}, Status: Unresolved, Reason: "native use, signature, environment or exact attachment is not applicable", Origin: use.Origin}
	application.Applicability.Claim.ID = application.Applicability.ID
	components := nativeComponentClaims(use, declaration, current)
	valid = valid && use.ID != "" && use.Boundary != "" && use.Phase != "" && use.Meaning == declaration.Meaning && use.Implementation != "" && !use.Target.Unknown && use.Target.Root != "" && use.Signature != nil && declaration.Signature != nil && reflect.DeepEqual(use.Signature, declaration.Signature)
	if valid {
		application.Applicability.Status, application.Applicability.Reason = Established, "exact typed native use, signature and current environment are applicable"
	}
	for i, promised := range declaration.Guarantees {
		promised.ID = StableID("native-guarantee", struct {
			Use      string
			Contract string
			Index    int
			Claim    Claim
		}{use.ID, declaration.ID, i, promised})
		application.Guarantees = append(application.Guarantees, NativeComponent{Use: use.ID, Component: "guarantee", Claim: promised, Current: current, Status: Unresolved, Reason: "declared guarantee has no applicable approved witness"})
	}
	approved := make(map[string]NativeComponent)
	approve := func(component *NativeComponent) {
		if !valid {
			component.Reason = "native use, signature, environment or exact attachment is not applicable"
			return
		}
		if TargetRelation(component.Claim.Target, use.Target) != Established || component.Claim.Target.Type != "" && component.Claim.Target.Type != use.Target.Type {
			component.Reason = "native claim has no exact resolved attachment for its selected meaning"
			return
		}
		if component.Claim.Scope != use.Boundary || component.Claim.Phase != use.Phase {
			component.Reason = "native claim does not cover this exact execution boundary and phase"
			return
		}
		if component.Component == "guarantee" {
			if len(component.Claim.Outcomes) == 0 {
				component.Reason = "native guarantee requires explicit covered outcomes"
				return
			}
			for _, outcome := range component.Claim.Outcomes {
				if len(declaration.Outcomes) > 0 && !containsOutcome(declaration.Outcomes, outcome) {
					component.Reason = "native guarantee covers an outcome outside the selected contract"
					return
				}
			}
		}
		for _, witness := range context.witnesses[ClaimFingerprint(component.Claim)] {
			result := ApproveNativeClaim(context.policy, witness, component.Claim, component.Current)
			component.Status, component.Reason = result.Status, result.Reason
			if result.Status == Established {
				component.Witness = witness.ID
				break
			}
		}
	}
	for i := range components {
		approve(&components[i])
		if components[i].Status == Established {
			approved[components[i].Component+":"+string(components[i].Outcome)] = components[i]
			application.Evidence = append(application.Evidence, components[i].Claim.ID)
		}
	}
	for i := range application.Guarantees {
		approve(&application.Guarantees[i])
	}
	application.Components = components
	unknown := func(path *PathSummary, code string, components []string) {
		path.Unknowns = append(path.Unknowns, UnknownReason{Code: code, Boundary: use.ID, Message: code + " for native use " + use.ID, Phase: use.Phase, Components: components})
	}
	outcomes := declaration.Outcomes
	_, closedOutcomes := approved["outcomes:"]
	if !closedOutcomes {
		outcomes = []Outcome{Normal}
	}
	for _, outcome := range outcomes {
		path := PathSummary{Outcome: outcome, Reachable: false, Demand: make(map[string]Multiplicity), Owned: make(map[string]OwnedResource), Resources: UnknownResourceCost("native costs are not approved")}
		if len(outcomes) == 1 {
			path.Reachable = true
		}
		if !closedOutcomes {
			unknown(&path, "native_outcomes", []string{"outcomes"})
		}
		contract := nativeOutcome(declaration, outcome)
		if component, ok := approved["effects:"+string(outcome)]; ok && contract.Effects != nil {
			path.Effects = append([]Effect(nil), contract.Effects.Effects...)
			path.Assumptions = append(path.Assumptions, component.Witness)
			if contract.Effects.UnknownTail == nil || *contract.Effects.UnknownTail || len(contract.Effects.Parameters) > 0 {
				unknown(&path, "native_effect_tail", []string{"effects"})
			}
		} else {
			unknown(&path, "native_effects", []string{"effects"})
		}
		var excluded []EffectKind
		for _, kind := range contract.AbsentEffects {
			if component, ok := approved["absent:"+string(kind)+":"+string(outcome)]; ok {
				excluded = append(excluded, kind)
				path.Assumptions = append(path.Assumptions, component.Witness)
			}
		}
		for i := range path.Unknowns {
			if path.Unknowns[i].Code == "native_effects" || path.Unknowns[i].Code == "native_effect_tail" {
				path.Unknowns[i].ExcludedEffects = excluded
			}
		}
		if component, ok := approved["costs:"+string(outcome)]; ok && contract.Costs != nil {
			path.Resources = *contract.Costs
			path.Assumptions = append(path.Assumptions, component.Witness)
		}
		for i, state := range contract.States {
			if component, ok := approved[fmt.Sprintf("state:%d:%s", i, outcome)]; ok {
				state.Evidence = []string{component.Claim.ID}
				path.StateRelations = append(path.StateRelations, state)
				path.Assumptions = append(path.Assumptions, component.Witness)
			}
		}
		if component, ok := approved["outcomes:"]; ok {
			path.Assumptions = append(path.Assumptions, component.Witness)
		}
		// Absence of ownership/demand declarations never grants their guarantees.
		unknown(&path, "native_ownership", []string{"ownership"})
		unknown(&path, "native_demand", []string{"demand"})
		application.Summary.Paths = append(application.Summary.Paths, path)
	}
	application.Summary = summarize(application.Summary.Paths)
	return application
}
