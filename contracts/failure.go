package contracts

import "strings"

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
	if summary.Coverage != Known || footprint.Observation == "" || summary.Observation != footprint.Observation || !evidenceOutcomeContains(footprint.Outcomes, summary.Outcome) || len(footprint.Locations) == 0 {
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
	if failurePathPrefix(left.Path, right.Path) || failurePathPrefix(right.Path, left.Path) {
		return "overlap"
	}
	return "disjoint"
}

func failurePathPrefix(left, right []string) bool {
	if len(left) > len(right) {
		return false
	}
	for i, part := range left {
		if part != right[i] {
			return false
		}
	}
	return true
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

type CleanupFailure struct {
	ID          string `json:"id"`
	Message     string `json:"message"`
	LaunchOrder int    `json:"launch_order"`
}

type CleanupReport struct {
	Primary    string           `json:"primary,omitempty"`
	Secondary  []CleanupFailure `json:"secondary,omitempty"`
	ExitStatus int              `json:"exit_status"`
	Policy     string           `json:"policy"`
	Coverage   Coverage         `json:"coverage"`
}

// RuntimeCleanupReport models the existing documented orchestrator policy:
// preserve the accepted primary failure, append cleanup failures once in launch
// order, and turn an otherwise successful exit into failure. Native conformance
// remains an independent obligation; this function does not grant it.
func RuntimeCleanupReport(primary string, acceptedStatus int, cleanup []CleanupFailure) CleanupReport {
	return RuntimeCleanupReportWithCauses(primary, nil, acceptedStatus, cleanup)
}

// RuntimeCleanupReportWithCauses retains error identity across an accepted
// primary cause chain and shutdown cleanup. Equal text does not imply the same
// failure. The caller supplies identities from the runtime error cause graph.
func RuntimeCleanupReportWithCauses(primary string, primaryCauses []string, acceptedStatus int, cleanup []CleanupFailure) CleanupReport {
	result := CleanupReport{Primary: primary, ExitStatus: acceptedStatus, Policy: "l-runtime-primary-plus-launch-ordered-cleanup/1", Coverage: Known}
	seen := make(map[string]bool)
	for _, id := range primaryCauses {
		if id != "" {
			seen[id] = true
		}
	}
	for _, failure := range cleanup {
		if failure.ID == "" {
			result.Coverage = Partial
			result.Secondary = append(result.Secondary, failure)
			continue
		}
		if !seen[failure.ID] {
			seen[failure.ID] = true
			result.Secondary = append(result.Secondary, failure)
		}
	}
	// Stable insertion preserves actual order for equal launch ordinals.
	for i := 1; i < len(result.Secondary); i++ {
		for j := i; j > 0 && result.Secondary[j].LaunchOrder < result.Secondary[j-1].LaunchOrder; j-- {
			result.Secondary[j], result.Secondary[j-1] = result.Secondary[j-1], result.Secondary[j]
		}
	}
	if acceptedStatus == 0 && (primary != "" || len(result.Secondary) != 0) {
		result.ExitStatus = 1
	}
	return result
}

func (report CleanupReport) Explanation() string {
	parts := []string{}
	if report.Primary != "" {
		parts = append(parts, report.Primary)
	}
	for _, failure := range report.Secondary {
		parts = append(parts, failure.Message)
	}
	return strings.Join(parts, "; ")
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
		if !evidenceOutcomeContains(footprint.Outcomes, outcome) {
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
	if len(results) == 0 {
		status, reason := Established, "closed outcome summary contains no covered failure path"
		if len(summary.Paths) == 0 || hasOutcomeUnknown(summary.Paths) {
			status, reason = Unresolved, "failure outcome coverage is incomplete"
		}
		results = append(results, Obligation{ID: claim.ID, Claim: claim, Status: status, Reason: reason})
	} else if hasOutcomeUnknown(summary.Paths) {
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
			if relation.Observation != footprint.Observation || len(relation.Evidence) == 0 || relation.Location.Resource.Unknown || relation.Location.Resource.ID != location.Resource.ID || !failurePathPrefix(relation.Location.Path, location.Path) {
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
