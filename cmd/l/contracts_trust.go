package main

import (
	"sort"

	"github.com/l-programming-language/l/contracts"
)

// appendNativeEvaluations reports the same prepared evaluations consumed by
// inference. Entry evidence comes only from executed typed analysis paths, never
// final facts or a second interpretation of the selected source/environment.
func appendNativeEvaluations(report *contracts.Report, evaluations []contracts.NativeEvaluation, trust contracts.TrustPolicyDocument) {
	context := contracts.EvidenceContext{Policy: trust.Policy, NativeWitnesses: make(map[string]contracts.NativeWitness), Current: make(map[string]contracts.EvidenceFingerprints), Checked: make(map[string]contracts.CheckedEvidence), Refutations: make(map[string]contracts.CheckedEvidence)}
	for _, witness := range trust.Witnesses {
		context.NativeWitnesses[witness.ID] = witness
	}
	checked := func(obligation contracts.Obligation) string {
		claim := obligation.Claim
		claim.ID = obligation.ID
		node := contracts.EvidenceNode{ID: claim.ID, Claim: claim.ID, Kind: "checked_derivation", Witness: contracts.StableID("native-applicability-check", obligation), Origin: obligation.Origin}
		node.Fingerprints = contracts.EvidenceFingerprints{Contract: contracts.ClaimFingerprint(claim), Implementation: contracts.AnalyzerVersion, Target: contracts.StableID("native-applicability-result", obligation)}
		context.Current[node.ID] = node.Fingerprints
		result := contracts.CheckedEvidence{ClaimFingerprint: contracts.ClaimFingerprint(claim), Checker: "typed-native-applicability", Fingerprints: node.Fingerprints}
		if obligation.Status == contracts.Established {
			context.Checked[node.Witness] = result
		} else if obligation.Status == contracts.Contradicted {
			node.Kind = "counterexample"
			context.Refutations[node.Witness] = result
		}
		report.Evidence.Claims = append(report.Evidence.Claims, claim)
		report.Evidence.Nodes = append(report.Evidence.Nodes, node)
		obligation.Evidence = []string{node.ID}
		report.Obligations = append(report.Obligations, obligation)
		return node.ID
	}
	for _, evaluation := range evaluations {
		entry := contracts.Obligation{ID: evaluation.Entry.ID, Claim: evaluation.Entry, Status: contracts.Unresolved, Reason: "no actual invocation-entry applicability result", Origin: evaluation.Use.Origin}
		found := false
		for _, observation := range report.Summary.Obligations {
			if observation.ID != entry.ID {
				continue
			}
			if !found || observation.Status == contracts.Contradicted || entry.Status == contracts.Established && observation.Status != contracts.Established {
				entry = observation
			}
			found = true
		}
		premises := []string{checked(evaluation.Applicability), checked(entry)}
		for _, component := range append(append([]contracts.NativeComponent(nil), evaluation.Components...), evaluation.Guarantees...) {
			node := contracts.EvidenceNode{ID: component.Claim.ID, Claim: component.Claim.ID, Kind: "native_assumption", Witness: component.Witness, Fingerprints: component.Current, Premises: premises, Origin: evaluation.Use.Origin}
			report.Evidence.Claims = append(report.Evidence.Claims, component.Claim)
			report.Evidence.Nodes = append(report.Evidence.Nodes, node)
			context.Current[node.ID] = component.Current
			result := contracts.EvidenceResult{Status: component.Status, Reason: component.Reason}
			if result.Status == contracts.Established {
				result = report.Evidence.Evaluate(node.ID, context)
			}
			component.Status, component.Reason = result.Status, result.Reason
			if component.Component != "guarantee" {
				report.NativeComponents = append(report.NativeComponents, component)
			}
			report.Obligations = append(report.Obligations, contracts.Obligation{ID: component.Claim.ID, Claim: component.Claim, Status: result.Status, Reason: result.Reason, Origin: evaluation.Use.Origin, Evidence: []string{node.ID}, Dependencies: result.Dependencies})
			if result.Status == contracts.Established {
				report.Assumptions = append(report.Assumptions, result.Assumptions...)
			}
		}
	}
	sort.Strings(report.Assumptions)
	unique := report.Assumptions[:0]
	for _, assumption := range report.Assumptions {
		if len(unique) == 0 || unique[len(unique)-1] != assumption {
			unique = append(unique, assumption)
		}
	}
	report.Assumptions = unique
}
