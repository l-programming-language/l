package contracts

import "sort"

type Candidate struct {
	Meaning       MeaningKey `json:"meaning"`
	Origins       []Origin   `json:"origins"`
	Applicability Predicate  `json:"applicability"`
	LexicalRole   string     `json:"lexical_role"`
	Phrase        string     `json:"phrase,omitempty"`
}

type SelectionContext struct {
	Target      TargetRef             `json:"target"`
	Facts       []Predicate           `json:"facts,omitempty"`
	Environment BindingEnvironmentKey `json:"environment"`
	Policy      string                `json:"policy"`
	LexicalRole string                `json:"lexical_role"`
}

type CandidateJudgment struct {
	Candidate Candidate `json:"candidate"`
	Status    Status    `json:"status"`
	Reason    string    `json:"reason"`
}

type ReferenceSelection struct {
	Status                Status                `json:"status"`
	Meaning               MeaningKey            `json:"meaning,omitempty"`
	Reason                string                `json:"reason"`
	Candidates            []CandidateJudgment   `json:"candidates"`
	DependencyFingerprint string                `json:"dependency_fingerprint"`
	Policy                string                `json:"policy"`
	Target                TargetRef             `json:"target"`
	Environment           BindingEnvironmentKey `json:"environment"`
}

// SelectReference implements Track B only on already typed candidates. It has
// no profile argument and is deliberately not used by the source resolver.
func SelectReference(candidates []Candidate, context SelectionContext) ReferenceSelection {
	if context.Policy == "" {
		context.Policy = ReferencePolicy
	}
	result := ReferenceSelection{Status: Unresolved, Policy: context.Policy, Target: context.Target, Environment: context.Environment}
	ordered := append([]Candidate(nil), candidates...)
	sort.SliceStable(ordered, func(i, j int) bool { return StableID("candidate", ordered[i]) < StableID("candidate", ordered[j]) })
	result.DependencyFingerprint = StableID("binding", struct {
		Candidates []Candidate
		Context    SelectionContext
	}{ordered, context})
	unresolvedCandidates := func(reason string) ReferenceSelection {
		result.Reason = reason
		for _, candidate := range ordered {
			result.Candidates = append(result.Candidates, CandidateJudgment{Candidate: candidate, Status: Unresolved, Reason: reason})
		}
		return result
	}
	if context.Policy != ReferencePolicy {
		return unresolvedCandidates("unsupported reference-selection policy")
	}
	if context.Target.Unknown || context.Target.Root == "" {
		return unresolvedCandidates("attachment is unresolved")
	}
	facts, typeStatus, typeReason := authoritativeTargetFacts(context.Target, context.Facts)
	if typeStatus != Established {
		result = unresolvedCandidates(typeReason)
		result.Status = typeStatus
		return result
	}
	if len(context.Target.Path) > 0 {
		facts = append(facts, Predicate{Kind: "eq", Subject: "role", Value: Atom{Kind: "role", Text: context.Target.Path[len(context.Target.Path)-1]}})
	}
	states := make(map[MeaningKey]Status)
	for _, candidate := range ordered {
		status, reason := Entails(facts, candidate.Applicability)
		if candidate.Meaning == "" {
			status, reason = Unresolved, "candidate has no canonical meaning"
		}
		if candidate.LexicalRole != "" && context.LexicalRole != "" && candidate.LexicalRole != context.LexicalRole {
			status, reason = Contradicted, "different lexical role"
		}
		result.Candidates = append(result.Candidates, CandidateJudgment{candidate, status, reason})
		prior, present := states[candidate.Meaning]
		if !present || status == Established || prior == Contradicted && status == Unresolved {
			states[candidate.Meaning] = status
		}
	}
	var applicable []MeaningKey
	unknown := false
	for meaning, status := range states {
		if status == Established {
			applicable = append(applicable, meaning)
		}
		if status == Unresolved {
			unknown = true
		}
	}
	if len(applicable) > 1 {
		result.Status, result.Reason = Contradicted, "ambiguous: several distinct meanings are applicable"
		return result
	}
	if unknown {
		result.Reason = "candidate applicability is undetermined and may affect uniqueness"
		return result
	}
	if len(applicable) == 0 {
		result.Status, result.Reason = Contradicted, "no applicable meaning"
		return result
	}
	result.Status, result.Meaning, result.Reason = Established, applicable[0], "one canonical meaning applies; all competing meanings are inapplicable"
	return result
}
