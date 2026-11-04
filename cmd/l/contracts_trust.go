package main

import (
	"fmt"
	"sort"

	"github.com/l-programming-language/l/contracts"
)

// applyContractTrust checks individual native propositions. It deliberately
// leaves independent native-closure gaps and whole-program profiles untouched.
func applyContractTrust(report *contracts.Report, metadata contractMetadata) error {
	if report == nil {
		return fmt.Errorf("contract trust requires a report")
	}
	policy := metadata.Trust.Policy
	if metadata.Trust.Kind != "" {
		if err := metadata.Trust.Validate(); err != nil {
			return err
		}
		report.TrustPolicy = &policy
		report.NativeWitnesses = append([]contracts.NativeWitness(nil), metadata.Trust.Witnesses...)
	}
	witnesses := append([]contracts.NativeWitness(nil), metadata.Trust.Witnesses...)
	sort.Slice(witnesses, func(i, j int) bool { return witnesses[i].ID < witnesses[j].ID })
	context := contracts.EvidenceContext{Policy: policy, NativeWitnesses: make(map[string]contracts.NativeWitness), Current: make(map[string]contracts.EvidenceFingerprints), Checked: make(map[string]contracts.CheckedEvidence), Refutations: make(map[string]contracts.CheckedEvidence)}
	for _, witness := range witnesses {
		context.NativeWitnesses[witness.ID] = witness
	}
	for _, declaration := range metadata.Native {
		current, fingerprintErr := contractNativeFingerprints(*report, declaration)
		// Applicability is required at every known application, including a
		// declaration that advertises no favorable guarantee.
		for _, target := range contractDeclarationTargets(*report, declaration.Meaning) {
			contractNativeApplicability(report, declaration, target, &context)
		}
		for index, promised := range declaration.Guarantees {
			claim := promised
			claim.ID = contracts.StableID("declared-native-claim", struct {
				Contract    string
				Index       int
				Proposition string
			}{declaration.ID, index, contracts.ClaimFingerprint(promised)})
			node := contracts.EvidenceNode{ID: contracts.StableID("native-claim-evidence", claim.ID), Claim: claim.ID, Kind: "native_assumption", Fingerprints: current, Origin: declaration.Origin}
			node.Premises = contractNativeApplicability(report, declaration, claim.Target, &context)
			result := contracts.EvidenceResult{Status: contracts.Unresolved, Reason: "declared guarantee has no explicitly approved current native witness"}
			if fingerprintErr != nil {
				result.Reason = fingerprintErr.Error()
			} else if !contractClaimTargetSelected(*report, declaration.Meaning, claim.Target) {
				result.Reason = "native claim has no exact resolved attachment for its selected meaning"
			} else {
				for _, witness := range witnesses {
					if witness.ClaimFingerprint != contracts.ClaimFingerprint(claim) {
						continue
					}
					checked := contracts.ApproveNativeClaim(policy, witness, claim, current)
					result, node.Witness = checked, witness.ID
					if checked.Status == contracts.Established {
						break
					}
				}
			}
			report.Evidence.Claims = append(report.Evidence.Claims, claim)
			report.Evidence.Nodes = append(report.Evidence.Nodes, node)
			context.Current[node.ID] = current
			if result.Status == contracts.Established {
				// The graph carries the same applicability dependencies as the
				// displayed obligation; a native witness cannot bypass them.
				result = report.Evidence.Evaluate(node.ID, context)
			}
			report.Obligations = append(report.Obligations, contracts.Obligation{ID: contracts.StableID("native-claim-obligation", claim.ID), Claim: claim, Status: result.Status, Reason: result.Reason, Origin: declaration.Origin, Evidence: []string{node.ID}, Dependencies: result.Dependencies})
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
	return nil
}

func contractDeclarationTargets(report contracts.Report, meaning contracts.MeaningKey) []contracts.TargetRef {
	var targets []contracts.TargetRef
	seen := make(map[string]bool)
	add := func(target contracts.TargetRef) {
		if target.Unknown || target.Root == "" {
			return
		}
		key := contracts.StableID("target", target)
		if !seen[key] {
			seen[key] = true
			targets = append(targets, target)
		}
	}
	for _, binding := range report.Bindings {
		if binding.Meaning == meaning {
			add(binding.Target)
		}
	}
	for _, step := range report.Lowering.Steps {
		if step.Meaning == meaning {
			add(step.Target)
		}
	}
	if len(targets) == 0 {
		targets = append(targets, contracts.TargetRef{Unknown: true})
	}
	return targets
}

func contractNativeApplicability(report *contracts.Report, declaration contracts.DeclaredContract, target contracts.TargetRef, context *contracts.EvidenceContext) []string {
	facts, selected := contractSelectedTargetFacts(*report, declaration.Meaning, target)
	var premises []string
	add := func(kind string, index int, predicate contracts.Predicate) {
		claim := contracts.Claim{Predicate: predicate, Target: target, Scope: "selected_attachment", Phase: target.Phase}
		claim.ID = contracts.StableID("native-applicability", struct {
			Contract string
			Kind     string
			Index    int
			Claim    contracts.Claim
		}{declaration.ID, kind, index, claim})
		node := contracts.EvidenceNode{ID: contracts.StableID("applicability-evidence", claim.ID), Claim: claim.ID, Kind: "checked_derivation", Witness: contracts.StableID("typed-target-check", claim.ID), Origin: declaration.Origin}
		status, reason := contracts.Entails(facts, predicate)
		if !selected {
			status, reason = contracts.Unresolved, "exact selected target facts are unavailable"
		}
		if kind == "signature" {
			status, reason = contracts.Unresolved, "selected source signature has no typed conformance projection; authored signature metadata is not applicability evidence"
		}
		if status == contracts.Contradicted {
			node.Kind = "counterexample"
		}
		node.Fingerprints = contracts.EvidenceFingerprints{Contract: contracts.ClaimFingerprint(claim), Implementation: contracts.AnalyzerVersion, Target: contracts.StableID("selected-target-facts", facts)}
		context.Current[node.ID] = node.Fingerprints
		checked := contracts.CheckedEvidence{ClaimFingerprint: contracts.ClaimFingerprint(claim), Checker: "typed-target-applicability/1", Fingerprints: node.Fingerprints}
		if status == contracts.Established {
			context.Checked[node.Witness] = checked
		} else if status == contracts.Contradicted {
			context.Refutations[node.Witness] = checked
		}
		premises = append(premises, node.ID)
		for _, existing := range report.Evidence.Nodes {
			if existing.ID == node.ID {
				return
			}
		}
		report.Evidence.Claims = append(report.Evidence.Claims, claim)
		report.Evidence.Nodes = append(report.Evidence.Nodes, node)
		report.Obligations = append(report.Obligations, contracts.Obligation{ID: claim.ID, Claim: claim, Status: status, Reason: "native " + kind + ": " + reason, Origin: declaration.Origin, Evidence: []string{node.ID}, Dependencies: []string{node.Fingerprints.Target}})
	}
	for i, requirement := range declaration.Requirements {
		add("requirement", i, requirement)
	}
	for i, restriction := range declaration.TargetRestrictions {
		add("target restriction", i, restriction)
	}
	if declaration.Signature != nil {
		add("signature", 0, contracts.Predicate{Kind: "signature_conformance", Subject: string(declaration.Meaning), Value: contracts.Atom{Kind: "signature", Text: contracts.StableID("signature", declaration.Signature)}})
	}
	return premises
}

// Only facts inherent in the selected canonical target are available here.
// Post-call path facts, advertised requirements and metadata never enter this set.
func contractSelectedTargetFacts(report contracts.Report, meaning contracts.MeaningKey, target contracts.TargetRef) ([]contracts.Predicate, bool) {
	var common []contracts.Predicate
	selected := false
	for _, actual := range contractDeclarationTargets(report, meaning) {
		if actual.Unknown || actual.Root == "" || contracts.TargetRelation(actual, target) != contracts.Established || target.Type != "" && target.Type != actual.Type {
			continue
		}
		facts := []contracts.Predicate{{Kind: "meaning", Subject: "target", Value: contracts.Atom{Kind: "meaning", Text: string(meaning)}}}
		if actual.Type != "" {
			facts = append(facts, contracts.Predicate{Kind: "type", Subject: "target", Value: contracts.Atom{Kind: "type", Text: actual.Type}})
		}
		if len(actual.Path) > 0 {
			facts = append(facts, contracts.Predicate{Kind: "eq", Subject: "role", Value: contracts.Atom{Kind: "role", Text: actual.Path[len(actual.Path)-1]}})
		}
		for _, fact := range append([]contracts.Predicate(nil), facts...) {
			fact.Target = actual
			facts = append(facts, fact)
		}
		if !selected {
			common = facts
			selected = true
			continue
		}
		var intersection []contracts.Predicate
		for _, fact := range common {
			for _, candidate := range facts {
				if contracts.StableID("fact", fact) == contracts.StableID("fact", candidate) {
					intersection = append(intersection, fact)
					break
				}
			}
		}
		common = intersection
	}
	return common, selected
}

func contractClaimTargetSelected(report contracts.Report, meaning contracts.MeaningKey, target contracts.TargetRef) bool {
	if target.Root == "" || target.Unknown {
		return false
	}
	matches := func(actual contracts.TargetRef) bool {
		return contracts.TargetRelation(actual, target) == contracts.Established && (target.Type == "" || actual.Type == target.Type)
	}
	for _, binding := range report.Bindings {
		if binding.Meaning == meaning && matches(binding.Target) {
			return true
		}
	}
	for _, step := range report.Lowering.Steps {
		if step.Meaning == meaning && matches(step.Target) {
			return true
		}
	}
	return false
}

func contractNativeFingerprints(report contracts.Report, declaration contracts.DeclaredContract) (contracts.EvidenceFingerprints, error) {
	var current contracts.EvidenceFingerprints
	contract, err := contracts.Fingerprint(declaration)
	if err != nil {
		return current, err
	}
	current.Contract = contract
	for _, step := range report.Lowering.Steps {
		if step.Meaning != declaration.Meaning {
			continue
		}
		if step.Implementation == "" || current.Implementation != "" && current.Implementation != step.Implementation {
			return current, fmt.Errorf("native meaning has missing or ambiguous current implementation fingerprints")
		}
		current.Implementation = step.Implementation
	}
	if current.Implementation == "" {
		return current, fmt.Errorf("native meaning has no current selected implementation fingerprint")
	}
	if report.Scope.Target == "" || report.Scope.Target == "unspecified" || report.Scope.Compiler == "" || report.Scope.PanicPolicy == "" {
		return current, fmt.Errorf("native assurance requires a concrete compiler, target and panic configuration")
	}
	if report.Scope.Target == "native-workspace" || report.Scope.PanicPolicy == "selected-runtime" {
		return current, fmt.Errorf("source-only inspection has not identified the concrete Rust toolchain, target and panic configuration; native implementation assurance remains unresolved")
	}
	current.Target, err = contracts.Fingerprint(report.Scope)
	if err != nil {
		return current, err
	}
	// Until a proven precise native dependency query exists, conservatively bind
	// all selected inputs. This includes cleanup-only helpers and runtime placement.
	current.Dependencies = make(map[string]string)
	for name, fingerprint := range report.Inputs {
		if fingerprint == "" {
			return current, fmt.Errorf("source dependency %s has no current fingerprint", name)
		}
		current.Dependencies["source:"+name] = fingerprint
	}
	for name, fingerprint := range report.Lowering.Fingerprints {
		if fingerprint == "" {
			return current, fmt.Errorf("lowering dependency %s has no current fingerprint", name)
		}
		current.Dependencies["lowering:"+name] = fingerprint
	}
	for _, dependency := range declaration.Dependencies {
		found := false
		for _, step := range report.Lowering.Steps {
			if string(step.Meaning) == dependency && step.Implementation != "" {
				found = true
			}
		}
		if !found {
			return current, fmt.Errorf("native contract dependency %s lacks a current selected implementation", dependency)
		}
	}
	return current, nil
}
