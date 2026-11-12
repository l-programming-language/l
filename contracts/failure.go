package contracts

type FailureLocation struct {
	Resource ResourceRef `json:"resource"`
	Path     []string    `json:"path,omitempty"`
}

type FailureChange struct {
	Location FailureLocation `json:"location"`
	Definite bool            `json:"definite"`
	Partial  bool            `json:"partial"`
	Origin   Origin          `json:"origin"`
}

type FailureSummary struct {
	Outcome        Outcome         `json:"outcome"`
	Observation    string          `json:"observation"`
	Changes        []FailureChange `json:"changes,omitempty"`
	Coverage       Coverage        `json:"coverage"`
	Invariants     []Predicate     `json:"invariants,omitempty"`
	FailureChannel string          `json:"failure_channel"`
}

type FailureFootprint struct {
	Locations   []FailureLocation `json:"locations"`
	Observation string            `json:"observation"`
	Outcomes    []Outcome         `json:"outcomes"`
}

type FailureResourcePair struct {
	Left  string `json:"left"`
	Right string `json:"right"`
}

func CheckFailurePreservation(summary FailureSummary, footprint FailureFootprint, disjoint []FailureResourcePair) Obligation {
	result := Obligation{ID: StableID("failure-preservation", footprint), Status: Established, Reason: "covered failure leaves the specified observable footprint unchanged"}
	if summary.Coverage != Known || footprint.Observation == "" || summary.Observation != footprint.Observation || !containsOutcome(footprint.Outcomes, summary.Outcome) || len(footprint.Locations) == 0 {
		result.Status, result.Reason = Unresolved, "failure outcome, observation boundary or complete footprint coverage is missing"
		return result
	}
	for _, location := range footprint.Locations {
		if location.Resource.Unknown || location.Resource.ID == "" {
			result.Status, result.Reason = Unresolved, "protected resource identity is unknown"
			continue
		}
		for _, change := range summary.Changes {
			relation := failureLocationRelation(location, change.Location, disjoint)
			if relation == "disjoint" {
				continue
			}
			if relation == "overlap" && change.Definite {
				result.Status, result.Reason = Contradicted, "failure changes the protected observable footprint"
				if change.Partial {
					result.Reason = "failure follows partial externally visible progress in the protected footprint"
				}
				result.Origin = change.Origin
				return result
			}
			result.Status, result.Reason = Unresolved, "possible write or resource aliasing prevents a preservation proof"
		}
	}
	return result
}

func failureLocationRelation(left, right FailureLocation, disjoint []FailureResourcePair) string {
	if left.Resource.Unknown || right.Resource.Unknown || left.Resource.ID == "" || right.Resource.ID == "" {
		return "unknown"
	}
	if left.Resource.ID != right.Resource.ID {
		for _, pair := range disjoint {
			if pair.Left == left.Resource.ID && pair.Right == right.Resource.ID || pair.Right == left.Resource.ID && pair.Left == right.Resource.ID {
				return "disjoint"
			}
		}
		return "unknown"
	}
	if pathPrefix(left.Path, right.Path) || pathPrefix(right.Path, left.Path) {
		return "overlap"
	}
	return "disjoint"
}

// CheckRetryFailureBeforeEffect supports the bounded failure-before-effect
// fragment. Idempotency, deduplication and restart protocols require separate
// proposition-specific evidence; an error-shaped return is never sufficient.
func CheckRetryFailureBeforeEffect(summary FailureSummary, footprint FailureFootprint, disjoint []FailureResourcePair) Obligation {
	result := CheckFailurePreservation(summary, footprint, disjoint)
	result.ID = StableID("retry-failure-before-effect", footprint)
	if summary.Outcome == Normal || summary.FailureChannel == "" {
		result.Status, result.Reason = Unresolved, "retry requires a covered failure channel and applicable failure protocol"
	} else if result.Status == Established {
		result.Reason = "failure-before-effect established only for the specified operation footprint and outcomes"
	}
	return result
}

// CheckFailureSummaryProfile projects actual outcome effects onto an explicit
// runtime resource footprint. Structural target paths are not resource paths.
func CheckFailureSummaryProfile(summary DerivedSummary, profile Profile, footprint FailureFootprint, disjoint []FailureResourcePair) []Obligation {
	claim := Claim{ID: StableID("failure-profile", footprint), Predicate: Predicate{Kind: "preserves_failure", Subject: StableID("footprint", footprint)}, Target: profile.Target, Scope: profile.Scope, Outcomes: footprint.Outcomes, Assumptions: summary.Assumptions}
	descriptor, supported := LookupProfile(profile.Name)
	if !supported || !descriptor.RequiresFootprint || !descriptor.SupportsScope(profile.Scope) || len(footprint.Locations) == 0 || len(footprint.Outcomes) == 0 || footprint.Observation == "" {
		return []Obligation{{ID: claim.ID, Claim: claim, Status: Unresolved, Reason: "PreservesOnFailure requires an explicit runtime resource footprint, observation boundary and failure outcomes"}}
	}
	var results []Obligation
	for index, path := range summary.Paths {
		outcome := path.Outcome
		if profile.Scope == "native_body" && path.Outcome == Normal && path.ObservedOutcome != "" {
			outcome = path.ObservedOutcome
		}
		if !containsOutcome(footprint.Outcomes, outcome) {
			continue
		}
		obligation := checkPathFinalState(path, footprint, disjoint)
		obligation.ID = StableID("failure-path", struct {
			Claim string
			Index int
		}{claim.ID, index})
		obligation.Claim = claim
		results = append(results, obligation)
	}
	incomplete := hasWorkCutoff(summary.Unknowns) || hasOutcomeUnknown(summary.Paths)
	if len(results) == 0 {
		status, reason := Established, "closed outcome summary contains no covered failure path"
		if len(summary.Paths) == 0 || incomplete {
			status, reason = Unresolved, "failure outcome coverage is incomplete"
		}
		results = append(results, Obligation{ID: claim.ID, Claim: claim, Status: status, Reason: reason})
	} else if incomplete {
		results = append(results, Obligation{ID: claim.ID + ":outcome-coverage", Claim: claim, Status: Unresolved, Reason: "additional failure outcomes remain unknown"})
	}
	return results
}

func hasOutcomeUnknown(paths []PathSummary) bool {
	for _, path := range paths {
		for _, unknown := range path.Unknowns {
			if len(unknown.Components) == 0 || containsString(unknown.Components, "outcomes") {
				return true
			}
		}
	}
	return false
}

// Write history alone cannot establish a different final value. Approved
// completion relations and exact typed state operations are separate evidence.
func checkPathFinalState(path PathSummary, footprint FailureFootprint, disjoint []FailureResourcePair) Obligation {
	result := Obligation{Status: Established, Reason: "selected failure preserves the modeled entry-to-completion state after cleanup"}
	for _, requirement := range path.Obligations {
		if requirement.Status != Established {
			selectFinding(&result, Unresolved, requirement.Reason, requirement.Origin)
		}
	}
	for _, location := range footprint.Locations {
		if location.Resource.ID == "" || location.Resource.Unknown {
			selectFinding(&result, Unresolved, "protected runtime resource identity is unresolved", Origin{})
			continue
		}
		if footprint.Observation == "state_at_completion" {
			status, reason := TypedFinalValueRelation(path, location)
			if status == Established {
				continue
			}
			if status == Contradicted {
				selectFinding(&result, status, reason, Origin{})
				continue
			}
		}
		proven := false
		for _, relation := range path.StateRelations {
			if relation.Observation != footprint.Observation || len(relation.Evidence) == 0 || relation.Location.Resource.Unknown || relation.Location.Resource.ID != location.Resource.ID || !pathPrefix(relation.Location.Path, location.Path) {
				continue
			}
			if relation.Relation == "different" && len(relation.Location.Path) == len(location.Path) {
				status := Unresolved
				if path.Reachable {
					status = Contradicted
				}
				selectFinding(&result, status, "approved completion relation changes the selected state", relation.Origin)
			} else if relation.Relation == "unchanged" {
				proven = true
			}
		}
		if proven {
			continue
		}
		for _, unknown := range path.Unknowns {
			if len(unknown.Components) == 0 || containsString(unknown.Components, "effects") || containsString(unknown.Components, "ownership") || containsString(unknown.Components, "outcomes") {
				selectFinding(&result, Unresolved, "completion state depends on unknown native, alias, outcome or cleanup behavior: "+unknown.Message, unknown.Origin)
			}
		}
		for _, effect := range path.Effects {
			switch effect.Kind {
			case "write", "io", "replace", "release", "deallocate", "spawn", "transfer":
				if failureLocationRelation(location, FailureLocation{Resource: effect.Resource}, disjoint) != "disjoint" {
					selectFinding(&result, Unresolved, "a possible mutation lacks a supported final-state relation", effect.Origin)
				}
			}
		}
	}
	return result
}
