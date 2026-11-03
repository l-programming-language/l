package contracts

import (
	"fmt"
	"reflect"
	"sort"
)

type SymbolicActionContract struct {
	Requires      []Predicate   `json:"requires,omitempty"`
	Ensures       []Predicate   `json:"ensures,omitempty"`
	Effects       []Effect      `json:"effects,omitempty"`
	EffectsClosed bool          `json:"effects_closed"`
	Outcomes      []Outcome     `json:"outcomes,omitempty"`
	Costs         *ResourceCost `json:"costs,omitempty"`
}

type ContractTransformer struct {
	Meaning               MeaningKey                        `json:"meaning"`
	Version               string                            `json:"version"`
	Target                TargetRef                         `json:"target"`
	Requires              []Predicate                       `json:"requires,omitempty"`
	Provides              []Predicate                       `json:"provides,omitempty"`
	Callbacks             map[string]SymbolicActionContract `json:"callbacks,omitempty"`
	Demand                map[string]Multiplicity           `json:"demand,omitempty"`
	Body                  Plan                              `json:"body"`
	Profiles              []Profile                         `json:"profiles,omitempty"`
	Frame                 *FrameContract                    `json:"frame,omitempty"`
	Facets                []FacetSchema                     `json:"facets,omitempty"`
	Slots                 []FacetSlot                       `json:"slots,omitempty"`
	Contributions         []ConfigurationContribution       `json:"contributions,omitempty"`
	Rules                 []FacetRule                       `json:"rules,omitempty"`
	UnsupportedReflection bool                              `json:"unsupported_reflection,omitempty"`
	Origin                Origin                            `json:"origin"`
}

type AdjectiveCertificate struct {
	Meaning                MeaningKey          `json:"meaning"`
	TransformerFingerprint string              `json:"transformer_fingerprint"`
	Status                 Status              `json:"status"`
	Mode                   string              `json:"mode"`
	Assumptions            []string            `json:"assumptions,omitempty"`
	Obligations            []Obligation        `json:"obligations"`
	Summary                DerivedSummary      `json:"summary"`
	Configuration          ConfigurationResult `json:"configuration"`
}

func VerifyAdjective(transformer ContractTransformer, options AnalysisOptions) (AdjectiveCertificate, error) {
	certificate := AdjectiveCertificate{Meaning: transformer.Meaning, TransformerFingerprint: StableID("transformer", transformer), Status: Unresolved, Mode: "symbolic"}
	if transformer.Meaning == "" || transformer.Version == "" || transformer.Target.Root == "" {
		return certificate, fmt.Errorf("adjective needs canonical meaning, version and symbolic target")
	}
	plan := transformer.Body
	plan.Callbacks = make(map[string]Node, len(transformer.Callbacks))
	names := make([]string, 0, len(transformer.Callbacks))
	for name := range transformer.Callbacks {
		names = append(names, name)
	}
	sort.Strings(names)
	for _, name := range names {
		contract := transformer.Callbacks[name]
		certificate.Assumptions = append(certificate.Assumptions, "symbolic callback "+name+":"+StableID("contract", contract))
		if !contract.EffectsClosed {
			plan.Callbacks[name] = Node{Kind: Native, Native: "symbolic callback " + name, Origin: transformer.Origin}
			continue
		}
		outcomes := contract.Outcomes
		if len(outcomes) == 0 {
			plan.Callbacks[name] = Node{Kind: Native, Native: "symbolic callback " + name + " has no declared outcome coverage", Origin: transformer.Origin}
			continue
		}
		costs := contract.Costs
		if costs == nil {
			unknown := UnknownResourceCost("symbolic callback costs were not declared")
			costs = &unknown
		}
		var alternatives []Node
		for _, outcome := range outcomes {
			alternatives = append(alternatives, Node{Kind: Sequence, Children: []Node{{Kind: Require, Requires: contract.Requires, Origin: transformer.Origin}, {Kind: Event, Facts: contract.Ensures, Effects: contract.Effects, Costs: costs, Outcome: outcome, Origin: transformer.Origin}}})
		}
		callback := alternatives[0]
		for _, alternative := range alternatives[1:] {
			callback = Node{Kind: Branch, Then: []Node{callback}, Else: []Node{alternative}}
		}
		plan.Callbacks[name] = callback
	}
	plan.Entry = Node{Kind: Sequence, Children: []Node{{Kind: Event, Facts: transformer.Requires}, plan.Entry}}
	summary, err := Infer(plan, options)
	if err != nil {
		return certificate, err
	}
	certificate.Summary = summary
	certificate.Obligations = append(certificate.Obligations, summary.Obligations...)
	certificate.Obligations = append(certificate.Obligations, checkTransformerSummary(transformer, summary)...)
	configuration, err := CheckConfiguration(transformer.Facets, transformer.Slots, transformer.Contributions, transformer.Rules, transformer.Requires)
	if err != nil {
		return certificate, err
	}
	certificate.Configuration = configuration
	certificate.Obligations = append(certificate.Obligations, configuration.Obligations...)
	if transformer.UnsupportedReflection {
		certificate.Mode = "instantiation_only"
		certificate.Obligations = append(certificate.Obligations, facetObligation(transformer.Target, transformer.Origin, Unresolved, "body inspects target internals outside the supported symbolic fragment", Predicate{Kind: "symbolic_body"}))
	}
	certificate.Status = obligationStatus(certificate.Obligations)
	return certificate, nil
}

func obligationStatus(obligations []Obligation) Status {
	status := Established
	for _, obligation := range obligations {
		if obligation.Status == Contradicted {
			return Contradicted
		}
		if obligation.Status != Established {
			status = Unresolved
		}
	}
	return status
}

func checkTransformerSummary(transformer ContractTransformer, summary DerivedSummary) []Obligation {
	obligations := checkSummaryPromises(summary, RoleOperation{Ensures: transformer.Provides, Profiles: transformer.Profiles}, transformer.Target, transformer.Origin)
	for callback, expected := range transformer.Demand {
		for _, path := range summary.Paths {
			actual := demandAt(path.Demand, callback)
			status, reason := Entails([]Predicate{{Kind: "multiplicity", Subject: callback, Lower: &actual.Lower, Upper: actual.Upper}}, Predicate{Kind: "multiplicity", Subject: callback, Lower: &expected.Lower, Upper: expected.Upper})
			obligations = append(obligations, facetObligation(transformer.Target, transformer.Origin, status, "callback demand: "+reason, Predicate{Kind: "multiplicity", Subject: callback, Lower: &expected.Lower, Upper: expected.Upper}))
		}
	}
	if transformer.Frame != nil {
		writes := FootprintsFromEffects(summary.Effects)
		if len(summary.Unknowns) > 0 {
			writes = append(writes, FrameFootprint{Kind: "unknown", Origin: transformer.Origin})
		}
		obligations = append(obligations, CheckFrame(*transformer.Frame, writes, FrameContext{})...)
	}
	return obligations
}

type AdjectiveInstance struct {
	Target    TargetRef       `json:"target"`
	Facts     []Predicate     `json:"facts,omitempty"`
	Callbacks map[string]Node `json:"callbacks,omitempty"`
	Lowering  []Node          `json:"lowering,omitempty"`
}

type AdjectiveInstantiation struct {
	Certificate AdjectiveCertificate `json:"certificate"`
	Summary     DerivedSummary       `json:"summary"`
	Obligations []Obligation         `json:"obligations"`
	Status      Status               `json:"status"`
}

// InstantiateAdjective never trusts a serialized success bit. It revalidates
// the symbolic certificate and checks concrete callback/target/lowering inputs.
func InstantiateAdjective(transformer ContractTransformer, certificate AdjectiveCertificate, instance AdjectiveInstance, options AnalysisOptions) (AdjectiveInstantiation, error) {
	verified, err := VerifyAdjective(transformer, options)
	result := AdjectiveInstantiation{Certificate: verified, Status: Unresolved}
	if err != nil {
		return result, err
	}
	if certificate.TransformerFingerprint != verified.TransformerFingerprint || certificate.Meaning != transformer.Meaning {
		result.Obligations = append(result.Obligations, facetObligation(instance.Target, transformer.Origin, Unresolved, "certificate dependency fingerprint changed", Predicate{Kind: "certificate"}))
	}
	if verified.Mode != "instantiation_only" {
		result.Obligations = append(result.Obligations, verified.Obligations...)
	}
	if instance.Target.Unknown || instance.Target.Root == "" {
		result.Obligations = append(result.Obligations, facetObligation(instance.Target, transformer.Origin, Unresolved, "concrete adjective attachment is unresolved", Predicate{Kind: "attachment"}))
	}
	if transformer.Target.Type != "" {
		status, reason := Established, "concrete target type satisfies the symbolic attachment domain"
		if instance.Target.Type == "" {
			status, reason = Unresolved, "concrete target type is missing for symbolic attachment domain "+transformer.Target.Type
		} else if instance.Target.Type != transformer.Target.Type {
			status, reason = Contradicted, "concrete target type "+instance.Target.Type+" does not match symbolic attachment domain "+transformer.Target.Type
		}
		result.Obligations = append(result.Obligations, facetObligation(instance.Target, transformer.Origin, status, reason, Predicate{Kind: "type", Subject: "target", Value: Atom{Kind: "type", Text: transformer.Target.Type}}))
	}
	facts, typeStatus, typeReason := authoritativeTargetFacts(instance.Target, instance.Facts)
	instance.Facts = facts
	if typeStatus != Established {
		result.Obligations = append(result.Obligations, facetObligation(instance.Target, transformer.Origin, typeStatus, typeReason, Predicate{Kind: "type", Subject: "target"}))
	}
	concrete := rebaseTransformer(transformer, instance.Target)
	for _, requirement := range concrete.Requires {
		status, reason := Entails(instance.Facts, requirement)
		result.Obligations = append(result.Obligations, facetObligation(instance.Target, transformer.Origin, status, "instantiation requirement: "+reason, requirement))
	}
	for name, contract := range concrete.Callbacks {
		body, exists := instance.Callbacks[name]
		if !exists {
			result.Obligations = append(result.Obligations, facetObligation(instance.Target, transformer.Origin, Unresolved, "missing concrete callback "+name, Predicate{Kind: "callback", Subject: name}))
			continue
		}
		callbackPlan := Plan{Entry: Node{Kind: Sequence, Children: []Node{{Kind: Event, Facts: contract.Requires}, body}}, Definitions: concrete.Body.Definitions, Callbacks: instance.Callbacks}
		summary, analysisErr := Infer(callbackPlan, options)
		if analysisErr != nil {
			return result, analysisErr
		}
		result.Obligations = append(result.Obligations, summary.Obligations...)
		result.Obligations = append(result.Obligations, checkSummaryPromises(summary, RoleOperation{Ensures: contract.Ensures, Outcomes: contract.Outcomes}, instance.Target, transformer.Origin)...)
		if contract.EffectsClosed {
			status, reason := Established, "concrete callback effects are within the symbolic row"
			if len(summary.Unknowns) > 0 {
				status, reason = Unresolved, "concrete callback has an unknown effect tail"
			}
			for _, effect := range summary.Effects {
				allowed := false
				for _, bound := range contract.Effects {
					if effect.Kind == bound.Kind && effect.Resource == bound.Resource && (reflect.DeepEqual(effect.Target, bound.Target) || TargetRelation(effect.Target, bound.Target) == Established) && (bound.Phase == "" || effect.Phase == bound.Phase) && (bound.Condition == "" || effect.Condition == bound.Condition) {
						allowed = true
					}
				}
				if !allowed {
					if effect.Definite {
						status, reason = Contradicted, "concrete callback adds an effect outside the symbolic contract"
					} else if status != Contradicted {
						status, reason = Unresolved, "concrete callback may add an effect outside the symbolic contract"
					}
				}
			}
			result.Obligations = append(result.Obligations, facetObligation(instance.Target, transformer.Origin, status, reason, Predicate{Kind: "effects_closed", Subject: name}))
		}
		if contract.Costs != nil {
			result.Obligations = append(result.Obligations, checkCallbackCosts(name, summary.Resources, *contract.Costs, instance.Target, transformer.Origin)...)
		}
	}
	plan := concrete.Body
	plan.Callbacks = instance.Callbacks
	children := []Node{{Kind: Event, Facts: instance.Facts}, plan.Entry}
	children = append(children, instance.Lowering...)
	plan.Entry = Node{Kind: Sequence, Children: children}
	summary, err := Infer(plan, options)
	if err != nil {
		return result, err
	}
	result.Summary = summary
	result.Obligations = append(result.Obligations, summary.Obligations...)
	result.Obligations = append(result.Obligations, checkTransformerSummary(concrete, summary)...)
	configuration, err := CheckConfiguration(concrete.Facets, concrete.Slots, concrete.Contributions, concrete.Rules, instance.Facts)
	if err != nil {
		return result, err
	}
	result.Obligations = append(result.Obligations, configuration.Obligations...)
	result.Status = obligationStatus(result.Obligations)
	return result, nil
}

func checkCallbackCosts(name string, actual, bound ResourceCost, target TargetRef, origin Origin) []Obligation {
	var obligations []Obligation
	if bound.NetLive.Kind != "" && bound.NetLive.Kind != "unknown" {
		status, reason := Unresolved, "exact callback live-resource change is not established"
		if actual.Model == bound.Model && reflect.DeepEqual(actual.NetLive, bound.NetLive) {
			status, reason = Established, "exact callback net live change matches"
		} else if value, known := actual.NetLive.Constant(); known {
			if expected, boundKnown := bound.NetLive.Constant(); boundKnown && value != expected {
				status, reason = Contradicted, "callback live-resource change differs; release credit cannot be reused"
			}
		}
		obligations = append(obligations, facetObligation(target, origin, status, reason, Predicate{Kind: "eq", Subject: name + ".net_live"}))
	}
	metrics := []struct {
		name          string
		actual, bound CostExpr
	}{{"work", actual.Work, bound.Work}, {"allocated", actual.Allocated, bound.Allocated}, {"payload_copied", actual.PayloadCopied, bound.PayloadCopied}, {"metadata_copied", actual.MetadataCopied, bound.MetadataCopied}, {"io", actual.IO, bound.IO}, {"peak_live", actual.PeakLive, bound.PeakLive}, {"retained", actual.Retained, bound.Retained}}
	for _, metric := range metrics {
		if metric.bound.Kind == "" || metric.bound.Kind == "unknown" {
			continue
		}
		status, reason := Unresolved, "callback cost inequality is outside the supported constant fragment"
		if actual.Model != bound.Model {
			reason = "callback resource models differ"
		} else if upper, known := metric.bound.Constant(); known {
			if value, actualKnown := metric.actual.Constant(); actualKnown {
				status, reason = Established, "derived callback cost satisfies the symbolic envelope"
				if value > upper {
					status, reason = Contradicted, "derived callback cost exceeds the symbolic envelope"
				}
			}
		} else if reflect.DeepEqual(metric.actual, metric.bound) {
			status, reason = Established, "identical supported symbolic callback cost"
		}
		obligations = append(obligations, facetObligation(target, origin, status, reason, Predicate{Kind: "cost_leq", Subject: name + "." + metric.name}))
	}
	return obligations
}

func rebaseAdjectiveTarget(value, from, to TargetRef) TargetRef {
	if value.Root != from.Root {
		return value
	}
	value.Root = to.Root
	value.Path = append(append([]string(nil), to.Path...), value.Path...)
	value.Generation = to.Generation
	if value.Phase == "" {
		value.Phase = to.Phase
	}
	return value
}

func rebasePredicates(values []Predicate, from, to TargetRef) []Predicate {
	result := append([]Predicate(nil), values...)
	for i := range result {
		result[i].Target = rebaseAdjectiveTarget(result[i].Target, from, to)
		result[i].Args = rebasePredicates(result[i].Args, from, to)
	}
	return result
}

func rebaseNodes(values []Node, from, to TargetRef) []Node {
	result := append([]Node(nil), values...)
	for i := range result {
		node := &result[i]
		node.Target = rebaseAdjectiveTarget(node.Target, from, to)
		node.Children = rebaseNodes(node.Children, from, to)
		if node.Body != nil {
			body := rebaseNodes([]Node{*node.Body}, from, to)[0]
			node.Body = &body
		}
		node.Then = rebaseNodes(node.Then, from, to)
		node.Else = rebaseNodes(node.Else, from, to)
		node.Cleanup = rebaseNodes(node.Cleanup, from, to)
		node.Facts = rebasePredicates(node.Facts, from, to)
		node.Requires = rebasePredicates(node.Requires, from, to)
		node.Effects = append([]Effect(nil), node.Effects...)
		for j := range node.Effects {
			node.Effects[j].Target = rebaseAdjectiveTarget(node.Effects[j].Target, from, to)
		}
	}
	return result
}

func rebaseTransformer(value ContractTransformer, target TargetRef) ContractTransformer {
	from := value.Target
	value.Target = target
	value.Requires = rebasePredicates(value.Requires, from, target)
	value.Provides = rebasePredicates(value.Provides, from, target)
	value.Body.Entry = rebaseNodes([]Node{value.Body.Entry}, from, target)[0]
	definitions := make(map[MeaningKey]Node, len(value.Body.Definitions))
	for key, node := range value.Body.Definitions {
		definitions[key] = rebaseNodes([]Node{node}, from, target)[0]
	}
	value.Body.Definitions = definitions
	callbacks := make(map[string]SymbolicActionContract, len(value.Callbacks))
	for name, contract := range value.Callbacks {
		contract.Requires = rebasePredicates(contract.Requires, from, target)
		contract.Ensures = rebasePredicates(contract.Ensures, from, target)
		contract.Effects = append([]Effect(nil), contract.Effects...)
		for i := range contract.Effects {
			contract.Effects[i].Target = rebaseAdjectiveTarget(contract.Effects[i].Target, from, target)
		}
		callbacks[name] = contract
	}
	value.Callbacks = callbacks
	if value.Frame != nil {
		frame := *value.Frame
		frame.Target = rebaseAdjectiveTarget(frame.Target, from, target)
		frame.Allowed = append([]TargetRef(nil), frame.Allowed...)
		frame.Preserves = append([]TargetRef(nil), frame.Preserves...)
		for i := range frame.Allowed {
			frame.Allowed[i] = rebaseAdjectiveTarget(frame.Allowed[i], from, target)
		}
		for i := range frame.Preserves {
			frame.Preserves[i] = rebaseAdjectiveTarget(frame.Preserves[i], from, target)
		}
		value.Frame = &frame
	}
	value.Slots = append([]FacetSlot(nil), value.Slots...)
	for i := range value.Slots {
		value.Slots[i].Target = rebaseAdjectiveTarget(value.Slots[i].Target, from, target)
	}
	value.Contributions = append([]ConfigurationContribution(nil), value.Contributions...)
	for i := range value.Contributions {
		value.Contributions[i].Target = rebaseAdjectiveTarget(value.Contributions[i].Target, from, target)
		value.Contributions[i].Condition = rebasePredicates(value.Contributions[i].Condition, from, target)
	}
	value.Rules = append([]FacetRule(nil), value.Rules...)
	for i := range value.Rules {
		value.Rules[i].When = rebasePredicates(value.Rules[i].When, from, target)
		value.Rules[i].Requires = rebasePredicates(value.Rules[i].Requires, from, target)
		value.Rules[i].Consequences = rebasePredicates(value.Rules[i].Consequences, from, target)
	}
	return value
}
