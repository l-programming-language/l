package contracts

import (
	"fmt"
	"slices"
)

// CheckProfile is enforcement over the closed, outcome-sensitive summary.
// Possible effects block assurance but do not constitute proven counterexamples.
func CheckProfile(summary DerivedSummary, profile Profile) []Obligation {
	claim := Claim{ID: StableID("profile", profile), Predicate: Predicate{Kind: "profile", Subject: profile.Name}, Target: profile.Target, Scope: profile.Scope, Outcomes: profile.Outcomes, Assumptions: summary.Assumptions}
	if claim.Scope == "" {
		claim.Scope = "whole_region"
	}
	base := Obligation{ID: claim.ID, Claim: claim, Status: Established, Reason: "closed modeled behavior satisfies the requested profile"}
	descriptor, supported := LookupProfile(profile.Name)
	if !supported || descriptor.ForbiddenEffect == "" || !descriptor.SupportsScope(profile.Scope) {
		base.Status = Unresolved
		base.Reason = "unsupported effect profile or scope: " + profile.Name + " / " + profile.Scope
		return []Obligation{base}
	}
	paths := summary.Paths
	if hasWorkCutoff(summary.Unknowns) {
		base.Status, base.Reason = Unresolved, errWorkLimit.Error()
		if len(profile.Outcomes) == 0 {
			// Witnessed prefix effects can refute an unrestricted check.
			paths = append(append([]PathSummary(nil), paths...), PathSummary{Reachable: true, Effects: summary.Effects})
		}
	}
	if len(paths) == 0 {
		base.Status, base.Reason = Unresolved, "summary has no established outcome coverage"
		return []Obligation{base}
	}
	includesPhase := func(phase Phase) bool {
		return profile.Scope != "body_only" || phase != Construction && phase != Destruction && phase != Cleanup
	}
	matched := false
	for _, path := range paths {
		for _, unknown := range path.Unknowns {
			if !includesPhase(unknown.Phase) {
				continue
			}
			for _, component := range unknown.Components {
				if component == "outcomes" {
					selectFinding(&base, Unresolved, unknown.Message, unknown.Origin)
				}
			}
		}
		outcome := path.Outcome
		if profile.Scope == "native_body" && path.Outcome == Normal && path.ObservedOutcome != "" {
			outcome = path.ObservedOutcome
		}
		if len(profile.Outcomes) > 0 && !containsOutcome(profile.Outcomes, outcome) {
			continue
		}
		matched = true
		for _, obligation := range path.Obligations {
			if obligation.Status != Established {
				selectFinding(&base, Unresolved, "body requirement is not established: "+obligation.Reason, obligation.Origin)
			}
		}
		if descriptor.ForbiddenEffect == "panic" && path.Outcome == Panic {
			status := Unresolved
			if path.Reachable {
				status = Contradicted
			}
			selectFinding(&base, status, "covered path may panic or unwind", Origin{})
		}
		for _, effect := range path.Effects {
			if !includesPhase(effect.Phase) {
				continue
			}
			if effect.Kind != descriptor.ForbiddenEffect {
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
			if !includesPhase(unknown.Phase) {
				continue
			}
			relevant := len(unknown.Components) == 0
			for _, component := range unknown.Components {
				relevant = relevant || component == "outcomes" || component == "effects" && !slices.Contains(unknown.ExcludedEffects, descriptor.ForbiddenEffect)
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
