package contracts

// Application requirements are discharged against the actual incoming path.
// No end-of-program fact or authored guarantee is a call-entry premise.
func (a *inference) applyNativeApplication(p PathSummary, node Node, application ApprovedNativeApplication) []PathSummary {
	use := application.Use
	valid := node.Native == use.ID && node.Meaning == use.Meaning && TargetRelation(node.Target, use.Target) == Established && node.Phase == use.Phase
	for _, requirement := range application.Requirements {
		want := rebasePredicate(requirement, node.Substitutions)
		status, reason := Entails(p.Facts, want)
		p.Obligations = append(p.Obligations, Obligation{ID: StableID("native-entry", struct {
			Use  string
			Want Predicate
		}{use.ID, want}), Claim: Claim{Predicate: want, Target: node.Target, Scope: "invocation_entry", Phase: node.Phase}, Status: status, Reason: reason, Origin: node.Origin})
		valid = valid && status == Established
	}
	if !valid {
		a.unknown(&p, node, "native_applicability", "approved components do not satisfy this exact use or its invocation-entry requirements")
		return []PathSummary{p}
	}
	summary := application.Summary
	summary.Paths = append([]PathSummary(nil), summary.Paths...)
	for i := range summary.Paths {
		path := &summary.Paths[i]
		path.ObservedOutcome = path.Outcome
		if use.OutcomeDisposition == "value" && path.Outcome == ReturnedFailure {
			path.Outcome = Normal
		} else if use.OutcomeDisposition != "value" && use.OutcomeDisposition != "propagate_returned_failure" {
			path.Unknowns = append(path.Unknowns, UnknownReason{Code: "native_control", Boundary: use.ID, Message: "native result propagation is not modeled", Components: []string{"outcomes"}, Phase: use.Phase})
		}
	}
	out := a.applySummary(p, summary, node.Substitutions)
	for i := range out {
		out[i].NativeBoundaries = append(out[i].NativeBoundaries, NativeBoundary{ID: use.ID, Meaning: use.Meaning, Target: node.Target, Origin: node.Origin, Status: Unresolved, Reason: "only individually approved components are available; native ownership, demand and transitive closure remain separate"})
	}
	return out
}
