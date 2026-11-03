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

// NativeUseFingerprints binds a specific application to selected source and
// native configuration, without observing or executing a historical tool.
func NativeUseFingerprints(use NativeUse, declaration DeclaredContract, environment NativeEnvironment) EvidenceFingerprints {
	dependencies := make(map[string]string)
	for name, value := range use.Dependencies {
		dependencies["use:"+name] = value
	}
	for name, value := range environment.Inputs {
		dependencies["source:"+name] = value
	}
	for name, value := range environment.Dependencies {
		dependencies["lowering:"+name] = value
	}
	dependencies["application"] = StableID("native-use", use)
	return EvidenceFingerprints{Contract: StableID("native-contract", declaration), Implementation: use.Implementation, Target: environment.Fingerprint, Dependencies: dependencies}
}

func NativeComponentClaims(use NativeUse, declaration DeclaredContract, environment NativeEnvironment) []NativeComponent {
	var result []NativeComponent
	add := func(component string, outcome Outcome, payload any) {
		outcomes := declaration.Outcomes
		if outcome != "" {
			outcomes = []Outcome{outcome}
		}
		claim := Claim{Predicate: Predicate{Kind: "native_component", Subject: component, Value: Atom{Kind: "fingerprint", Text: StableID("component", payload)}}, Target: use.Target, Scope: "invocation", Phase: use.Phase, Outcomes: outcomes}
		if use.Boundary != "" {
			claim.Scope = use.Boundary
		}
		claim.ID = StableID("native-component", struct {
			Use   string
			Claim Claim
		}{use.ID, claim})
		result = append(result, NativeComponent{Use: use.ID, Component: component, Outcome: outcome, Claim: claim, Current: NativeUseFingerprints(use, declaration, environment), Status: Unresolved, Reason: "component has no applicable approved witness"})
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

// ApproveNativeApplication translates only individually approved component
// propositions. Requirements are deliberately retained for call-entry checking.
func ApproveNativeApplication(use NativeUse, declaration DeclaredContract, environment NativeEnvironment, policy TrustPolicy, witnesses []NativeWitness, applicability Status) (ApprovedNativeApplication, []NativeComponent) {
	application := ApprovedNativeApplication{Use: use, Requirements: append(append([]Predicate(nil), declaration.Requirements...), declaration.TargetRestrictions...)}
	application.ContextFingerprint = StableID("native-assurance-context", struct {
		Environment NativeEnvironment
		Policy      TrustPolicy
		Witnesses   []NativeWitness
	}{environment, policy, witnesses})
	components := NativeComponentClaims(use, declaration, environment)
	valid := applicability == Established && use.ID != "" && use.Meaning == declaration.Meaning && use.Implementation != "" && !use.Target.Unknown && use.Target.Root != "" && use.Signature != nil && declaration.Signature != nil && reflect.DeepEqual(use.Signature, declaration.Signature)
	valid = valid && environment.Validate() == nil
	valid = valid && (NativeContractManifest{SchemaVersion: SchemaVersion, Kind: "native_contract_manifest", Producer: AnalyzerVersion, Declarations: []DeclaredContract{declaration}}).Validate() == nil
	if environment.Method == "test-only" && !policy.AllowTestOnly {
		valid = false
	}
	ordered := append([]NativeWitness(nil), witnesses...)
	sort.Slice(ordered, func(i, j int) bool { return ordered[i].ID < ordered[j].ID })
	approved := make(map[string]NativeComponent)
	for i := range components {
		component := &components[i]
		if !valid {
			component.Reason = "native use, signature, environment or exact attachment is not applicable"
			continue
		}
		for _, witness := range ordered {
			if witness.ClaimFingerprint != ClaimFingerprint(component.Claim) {
				continue
			}
			result := ApproveNativeClaim(policy, witness, component.Claim, component.Current)
			component.Status, component.Reason = result.Status, result.Reason
			if result.Status == Established {
				component.Witness = witness.ID
				approved[component.Component+":"+string(component.Outcome)] = *component
				application.Evidence = append(application.Evidence, component.Claim.ID)
				break
			}
		}
	}
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
	return application, components
}
