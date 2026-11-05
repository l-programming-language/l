package contracts

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"
)

// Explanation contains only observed selection and checker records. It does
// not assign blame or introduce source qualification or attachment syntax.
type Explanation struct {
	ID           string              `json:"id"`
	Kind         string              `json:"kind"`
	Status       Status              `json:"status"`
	Target       TargetRef           `json:"target"`
	Meaning      MeaningKey          `json:"meaning,omitempty"`
	Reason       string              `json:"reason"`
	Requirement  Predicate           `json:"requirement"`
	Candidates   []CandidateJudgment `json:"candidates,omitempty"`
	Origins      []Origin            `json:"origins,omitempty"`
	Dependencies []string            `json:"dependencies,omitempty"`
}

func ExplainReport(report Report) []Explanation {
	var result []Explanation
	for _, binding := range report.Bindings {
		status := Established
		if binding.Target.Unknown {
			status = Unresolved
		}
		result = append(result, Explanation{ID: string(binding.Occurrence), Kind: "binding", Status: status, Target: binding.Target, Meaning: binding.Meaning, Reason: binding.SelectionReason, Origins: []Origin{binding.Origin, binding.DefinitionOrigin}, Dependencies: append([]string(nil), binding.Dependencies...)})
	}
	for _, selection := range report.ReferenceSelections {
		explanation := Explanation{ID: selection.DependencyFingerprint, Kind: "reference_selection", Status: selection.Status, Target: selection.Target, Meaning: selection.Meaning, Reason: selection.Reason, Candidates: append([]CandidateJudgment(nil), selection.Candidates...), Dependencies: []string{string(selection.Environment), selection.Policy}}
		for _, candidate := range selection.Candidates {
			explanation.Origins = append(explanation.Origins, candidate.Candidate.Origins...)
		}
		result = append(result, explanation)
	}
	obligations := append([]Obligation(nil), report.Obligations...)
	if report.Configuration != nil {
		obligations = append(obligations, report.Configuration.Obligations...)
	}
	for _, profile := range report.Profiles {
		obligations = append(obligations, profile.Obligations...)
	}
	writes := FootprintsFromEffects(report.Summary.Effects)
	for _, certificate := range report.Certificates {
		obligations = append(obligations, certificate.Obligations...)
		writes = append(writes, FootprintsFromEffects(certificate.Summary.Effects)...)
	}
	for _, certificate := range report.SourceCertificates {
		result = append(result, Explanation{ID: certificate.ProjectionFingerprint, Kind: "source_logical_certificate", Status: certificate.LogicalStatus, Target: certificate.Target, Meaning: certificate.Meaning, Reason: "source elaboration " + certificate.Mode + " theorem; native implementation assurance remains separate", Dependencies: append([]string{certificate.DeclarationFingerprint, string(certificate.Environment)}, certificate.Assumptions...)})
		obligations = append(obligations, certificate.LogicalObligations...)
		obligations = append(obligations, certificate.LoweringObligations...)
		writes = append(writes, FootprintsFromEffects(certificate.Summary.Effects)...)
	}
	for _, declaration := range report.SourceDeclarations {
		for _, requirement := range declaration.Requires {
			result = append(result, Explanation{ID: StableID("source-premise", requirement), Kind: "source_assumption", Status: Unresolved, Target: requirement.Target, Meaning: declaration.Meaning, Reason: "conditional source premise; must be checked against each concrete before-rule target", Requirement: requirement, Origins: []Origin{declaration.Origin}})
		}
	}
	for _, instance := range report.SourceInstantiations {
		result = append(result, Explanation{ID: instance.CertificateFingerprint, Kind: "source_instantiation", Status: instance.Status, Target: instance.Target, Meaning: instance.Meaning, Reason: "concrete before-rule requirements and structural writes checked; native assurance remains separate"})
		obligations = append(obligations, instance.Obligations...)
		obligations = append(obligations, instance.LoweringObligations...)
	}
	seen := make(map[string]bool)
	for _, obligation := range obligations {
		key := StableID("explanation-obligation", obligation)
		if seen[key] {
			continue
		}
		seen[key] = true
		explanation := Explanation{ID: obligation.ID, Kind: "obligation", Status: obligation.Status, Target: obligation.Claim.Target, Reason: obligation.Reason, Requirement: obligation.Claim.Predicate, Origins: []Origin{obligation.Origin}, Dependencies: append([]string(nil), obligation.Dependencies...)}
		if report.Configuration != nil {
			for _, selection := range report.Configuration.Selections {
				if TargetRelation(selection.Slot.Target, obligation.Claim.Target) == Established {
					explanation.Origins = append(explanation.Origins, selection.Origins...)
				}
			}
		}
		for _, write := range writes {
			if containsString(obligation.Dependencies, StableID("typed-write", write)) {
				explanation.Origins = append(explanation.Origins, write.Origin)
			}
		}
		result = append(result, explanation)
	}
	for i := range result {
		result[i].Origins = uniqueExplanationOrigins(result[i].Origins)
		sort.Strings(result[i].Dependencies)
	}
	sort.SliceStable(result, func(i, j int) bool {
		return StableID("explanation", result[i]) < StableID("explanation", result[j])
	})
	return result
}

func uniqueExplanationOrigins(origins []Origin) []Origin {
	result := distinctRecords("origin", origins)
	sort.Slice(result, func(i, j int) bool { return StableID("origin", result[i]) < StableID("origin", result[j]) })
	return result
}

func FormatExplanations(explanations []Explanation) string {
	var output strings.Builder
	for _, explanation := range explanations {
		fmt.Fprintf(&output, "%s %s %s target=%s meaning=%s: %s\n", explanation.Kind, explanation.ID, explanation.Status, explanation.Target.String(), explanation.Meaning, explanation.Reason)
		if explanation.Requirement.Kind != "" {
			fmt.Fprintf(&output, "  requirement=%s\n", explanationPredicate(explanation.Requirement))
		}
		for _, candidate := range explanation.Candidates {
			fmt.Fprintf(&output, "  candidate=%s applicability=%s %s: %s\n", candidate.Candidate.Meaning, explanationPredicate(candidate.Candidate.Applicability), candidate.Status, candidate.Reason)
		}
		for _, origin := range explanation.Origins {
			fmt.Fprintf(&output, "  origin=%s:%d-%d %s\n", origin.Source, origin.Start, origin.End, origin.Label)
		}
		for _, dependency := range explanation.Dependencies {
			fmt.Fprintf(&output, "  dependency=%s\n", dependency)
		}
	}
	return output.String()
}

func explanationPredicate(predicate Predicate) string {
	encoded, err := json.Marshal(predicate)
	if err != nil {
		return "unserializable predicate"
	}
	return string(encoded)
}
