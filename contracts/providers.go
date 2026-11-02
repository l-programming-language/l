package contracts

import "fmt"

type RoleOperation struct {
	Name        string      `json:"name"`
	Requires    []Predicate `json:"requires,omitempty"`
	Ensures     []Predicate `json:"ensures,omitempty"`
	Profiles    []Profile   `json:"profiles,omitempty"`
	Outcomes    []Outcome   `json:"outcomes,omitempty"`
	MaximumWork *int64      `json:"maximum_work,omitempty"`
}

type RoleInterface struct {
	ID           string          `json:"id"`
	Version      string          `json:"version"`
	Requires     []Predicate     `json:"requires,omitempty"`
	Capabilities []string        `json:"capabilities,omitempty"`
	Operations   []RoleOperation `json:"operations"`
	Frame        *FrameContract  `json:"frame,omitempty"`
	Origin       Origin          `json:"origin"`
}

type ProviderImplementation struct {
	ID           MeaningKey      `json:"id"`
	Target       TargetRef       `json:"target"`
	Requires     []Predicate     `json:"requires,omitempty"`
	Capabilities []string        `json:"capabilities,omitempty"`
	Operations   map[string]Plan `json:"operations"`
	Origin       Origin          `json:"origin"`
}

type ProviderConformance struct {
	Provider     MeaningKey                `json:"provider"`
	Role         string                    `json:"role"`
	Target       TargetRef                 `json:"target"`
	Obligations  []Obligation              `json:"obligations"`
	Summaries    map[string]DerivedSummary `json:"summaries"`
	Dependencies []string                  `json:"dependencies"`
}

func (role RoleInterface) Validate() error {
	if role.ID == "" || role.Version == "" {
		return fmt.Errorf("role interface needs identity and version")
	}
	seen := make(map[string]bool)
	for _, operation := range role.Operations {
		if operation.Name == "" || seen[operation.Name] {
			return fmt.Errorf("role %s has empty or duplicate operation identity", role.ID)
		}
		seen[operation.Name] = true
	}
	return nil
}

// CheckProvider analyzes the actual supplied typed operations under the role's
// caller premises. Matching operation names and provider labels are insufficient.
func CheckProvider(role RoleInterface, provider ProviderImplementation, facts []Predicate, options AnalysisOptions) (ProviderConformance, error) {
	result := ProviderConformance{Provider: provider.ID, Role: role.ID, Target: provider.Target, Summaries: make(map[string]DerivedSummary), Dependencies: []string{StableID("role", role), StableID("provider", provider)}}
	if err := role.Validate(); err != nil {
		return result, err
	}
	premises := append(append([]Predicate(nil), facts...), role.Requires...)
	premises, typeStatus, typeReason := authoritativeTargetFacts(provider.Target, premises)
	if typeStatus != Established {
		result.Obligations = append(result.Obligations, facetObligation(provider.Target, provider.Origin, typeStatus, typeReason, Predicate{Kind: "type", Subject: "target"}))
	}
	for _, requirement := range provider.Requires {
		status, reason := Entails(premises, requirement)
		result.Obligations = append(result.Obligations, facetObligation(provider.Target, provider.Origin, status, "provider caller requirement: "+reason, requirement))
	}
	for _, capability := range provider.Capabilities {
		status := Established
		if !containsString(role.Capabilities, capability) {
			status = Contradicted
		}
		result.Obligations = append(result.Obligations, facetObligation(provider.Target, provider.Origin, status, "provider may not add caller capability requirements", Predicate{Kind: "capability", Subject: capability}))
	}
	for _, operation := range role.Operations {
		plan, found := provider.Operations[operation.Name]
		if !found {
			result.Obligations = append(result.Obligations, facetObligation(provider.Target, provider.Origin, Contradicted, "provider lacks required operation "+operation.Name, Predicate{Kind: "member", Subject: operation.Name}))
			continue
		}
		operationPremises := append(append([]Predicate(nil), premises...), operation.Requires...)
		operationPremises, typeStatus, typeReason := authoritativeTargetFacts(provider.Target, operationPremises)
		if typeStatus != Established {
			result.Obligations = append(result.Obligations, facetObligation(provider.Target, provider.Origin, typeStatus, typeReason, Predicate{Kind: "type", Subject: "target"}))
		}
		plan.Entry = Node{Kind: "sequence", Children: []Node{{Kind: "event", Facts: operationPremises}, plan.Entry}}
		summary, err := Infer(plan, options)
		if err != nil {
			return result, err
		}
		result.Summaries[operation.Name] = summary
		result.Obligations = append(result.Obligations, summary.Obligations...)
		result.Obligations = append(result.Obligations, checkSummaryPromises(summary, operation, provider.Target, provider.Origin)...)
		if role.Frame != nil {
			writes := FootprintsFromEffects(summary.Effects)
			if len(summary.Unknowns) > 0 {
				writes = append(writes, FrameFootprint{Kind: "unknown", Origin: provider.Origin})
			}
			result.Obligations = append(result.Obligations, CheckFrame(*role.Frame, writes, FrameContext{})...)
		}
	}
	return result, nil
}

func checkSummaryPromises(summary DerivedSummary, operation RoleOperation, target TargetRef, origin Origin) []Obligation {
	var obligations []Obligation
	for _, guarantee := range operation.Ensures {
		status, reason := Established, "postcondition holds on every covered return path"
		covered := false
		for _, path := range summary.Paths {
			if path.Outcome != Normal {
				continue
			}
			covered = true
			s, r := Entails(path.Facts, guarantee)
			if s == Contradicted || s == Unresolved && status == Established {
				status, reason = s, r
			}
		}
		if !covered {
			status, reason = Unresolved, "no normal path establishes the requested postcondition"
		}
		obligations = append(obligations, facetObligation(target, origin, status, reason, guarantee))
	}
	if len(operation.Outcomes) > 0 {
		for _, path := range summary.Paths {
			allowed := false
			for _, outcome := range operation.Outcomes {
				if path.Outcome == outcome {
					allowed = true
				}
			}
			if !allowed {
				obligations = append(obligations, facetObligation(target, origin, Contradicted, "provider admits an outcome excluded by its role", Predicate{Kind: "outcome", Subject: string(path.Outcome)}))
			}
		}
		if len(summary.Unknowns) > 0 {
			obligations = append(obligations, facetObligation(target, origin, Unresolved, "unknown behavior prevents a closed outcome guarantee", Predicate{Kind: "outcome"}))
		}
	}
	for _, profile := range operation.Profiles {
		profile.Target = target
		obligations = append(obligations, CheckProfile(summary, profile)...)
	}
	if operation.MaximumWork != nil {
		status, reason := Unresolved, "work bound is not a known constant"
		if work, known := summary.Resources.Work.Constant(); known {
			status, reason = Established, "derived work is within the role envelope"
			if work > *operation.MaximumWork {
				status, reason = Contradicted, "derived work exceeds the role envelope"
			}
		}
		obligations = append(obligations, facetObligation(target, origin, status, reason, Predicate{Kind: "cost_leq", Subject: "work", Upper: operation.MaximumWork}))
	}
	return obligations
}
