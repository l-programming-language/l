package contracts

import "fmt"

func SupportedProfile(name string) bool {
	switch name {
	case "NoHeapAllocation", "NoBlocking", "NoExternalIO", "NoPanic":
		return true
	}
	return false
}

// CheckProfile is enforcement over the closed, outcome-sensitive summary.
// Possible effects block assurance but do not constitute proven counterexamples.
func CheckProfile(summary DerivedSummary, profile Profile) []Obligation {
	claim := Claim{ID: StableID("profile", profile), Predicate: Predicate{Kind: "profile", Subject: profile.Name}, Target: profile.Target, Scope: profile.Scope, Outcomes: profile.Outcomes, Assumptions: summary.Assumptions}
	if claim.Scope == "" {
		claim.Scope = "whole_region"
	}
	base := Obligation{ID: claim.ID, Claim: claim, Status: Established, Reason: "closed modeled behavior satisfies the requested profile"}
	if !SupportedProfile(profile.Name) {
		base.Status = Unresolved
		base.Reason = "unsupported profile " + profile.Name
		return []Obligation{base}
	}
	if len(summary.Paths) == 0 {
		base.Status = Unresolved
		base.Reason = "summary has no established outcome coverage"
		return []Obligation{base}
	}
	matched := false
	for _, path := range summary.Paths {
		for _, unknown := range path.Unknowns {
			if profile.Scope == "body_only" && (unknown.Phase == Construction || unknown.Phase == Destruction || unknown.Phase == Cleanup) {
				continue
			}
			for _, component := range unknown.Components {
				if component == "outcomes" {
					selectFinding(&base, Unresolved, unknown.Message, unknown.Origin)
				}
			}
		}
		if len(profile.Outcomes) > 0 && !containsOutcome(profile.Outcomes, path.Outcome) {
			continue
		}
		matched = true
		for _, obligation := range path.Obligations {
			if obligation.Status != Established {
				selectFinding(&base, Unresolved, "body requirement is not established: "+obligation.Reason, obligation.Origin)
			}
		}
		if profile.Name == "NoPanic" && path.Outcome == Panic {
			status := Unresolved
			if path.Reachable {
				status = Contradicted
			}
			selectFinding(&base, status, "covered path may panic or unwind", Origin{})
		}
		for _, effect := range path.Effects {
			if profile.Scope == "body_only" && (effect.Phase == Construction || effect.Phase == Destruction || effect.Phase == Cleanup) {
				continue
			}
			forbidden := profile.Name == "NoHeapAllocation" && effect.Kind == "allocate" || profile.Name == "NoBlocking" && effect.Kind == "block" || profile.Name == "NoExternalIO" && effect.Kind == "io" || profile.Name == "NoPanic" && effect.Kind == "panic"
			if !forbidden {
				continue
			}
			status := Unresolved
			if effect.Definite && path.Reachable {
				status = Contradicted
			}
			reason := fmt.Sprintf("%s is introduced during %s", effect.Kind, effect.Phase)
			if effect.Condition != "" {
				reason += " when " + effect.Condition
			}
			selectFinding(&base, status, reason, effect.Origin)
		}
		for _, unknown := range path.Unknowns {
			if profile.Scope == "body_only" && (unknown.Phase == Construction || unknown.Phase == Destruction || unknown.Phase == Cleanup) {
				continue
			}
			relevant := len(unknown.Components) == 0
			for _, component := range unknown.Components {
				relevant = relevant || component == "effects" || component == "outcomes"
			}
			if relevant {
				selectFinding(&base, Unresolved, unknown.Message, unknown.Origin)
			}
		}
	}
	if !matched {
		selectFinding(&base, Unresolved, "requested outcome scope has no established applicable paths", Origin{})
	}
	if len(base.Claim.Assumptions) > 0 && base.Status == Established {
		base.Reason = "established conditional on the displayed approved native assumptions"
	}
	return []Obligation{base}
}

func selectFinding(obligation *Obligation, status Status, reason string, origin Origin) {
	if obligation.Status == Contradicted || obligation.Status == Unresolved && status != Contradicted {
		return
	}
	obligation.Status, obligation.Reason, obligation.Origin = status, reason, origin
}
