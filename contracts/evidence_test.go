package contracts

import (
	"strings"
	"testing"
)

func evidenceFixture() (Claim, NativeWitness, TrustPolicy, EvidenceFingerprints) {
	claim := Claim{ID: "no-block", Predicate: Predicate{Kind: "effects_closed", Subject: "blocking", Value: Atom{Kind: "bool", Bool: false}}, Target: TargetRef{Root: "fixture", Path: []string{"worker"}}, Scope: "whole_region", Phase: Invocation, Outcomes: []Outcome{Normal}}
	current := EvidenceFingerprints{Contract: "contract-v1", Implementation: "body-v1", Target: "fixture-target-v1", Dependencies: map[string]string{"drop-helper": "drop-v1"}}
	witness := NativeWitness{ID: "test-only-native", ClaimFingerprint: ClaimFingerprint(claim), ContractFingerprint: current.Contract, ImplementationFingerprint: current.Implementation, TargetFingerprint: current.Target, Authority: "test-authority", Rationale: "Explicit synthetic fixture assumption; grants no production approval", Dependencies: map[string]string{"drop-helper": "drop-v1"}, TestOnly: true}
	policy := TrustPolicy{Version: SchemaVersion, ID: "test-only-policy", Authorities: []string{"test-authority"}, ApprovedWitnesses: []string{witness.ID}, AllowTestOnly: true}
	return claim, witness, policy, current
}

func TestEvidenceDependencies(t *testing.T) {
	t.Run("BASE-043", func(t *testing.T) {
		claim, witness, policy, current := evidenceFixture()
		copyClaim := claim
		copyClaim.ID = "caller-no-block"
		graph := EvidenceGraph{Claims: []Claim{claim, copyClaim}, Nodes: []EvidenceNode{{ID: "native", Claim: claim.ID, Kind: "native_assumption", Witness: witness.ID}, {ID: "caller", Claim: copyClaim.ID, Kind: "derived", Rule: "identity", Premises: []string{"native"}}}}
		context := EvidenceContext{Policy: policy, NativeWitnesses: map[string]NativeWitness{witness.ID: witness}, Current: map[string]EvidenceFingerprints{"native": current}}
		result := graph.Evaluate("caller", context)
		if result.Status != Established || len(result.Assumptions) != 1 || result.Assumptions[0] != witness.ID || !strings.Contains(result.Reason, "derived") || !containsString(result.Evidence, "native") {
			t.Fatalf("derived trust closure was lost: %+v", result)
		}
	})
	t.Run("BASE-044", func(t *testing.T) {
		claim, _, _, current := evidenceFixture()
		graph := EvidenceGraph{Claims: []Claim{claim}, Nodes: []EvidenceNode{{ID: "unit-suite", Claim: claim.ID, Kind: "empirical", Witness: "tests"}, {ID: "composite", Claim: claim.ID, Kind: "derived", Rule: "identity", Premises: []string{"unit-suite"}}}}
		context := EvidenceContext{Observations: map[string]CheckedEvidence{"tests": {ClaimFingerprint: ClaimFingerprint(claim), Checker: "separate-unit-suites", Fingerprints: current}}, Current: map[string]EvidenceFingerprints{"unit-suite": current}}
		if got := graph.Evaluate("composite", context); got.Status != Unresolved || !strings.Contains(got.Reason, "empirical") {
			t.Fatalf("empirical unit tests became integration theorem: %+v", got)
		}
	})
	t.Run("BASE-045", func(t *testing.T) {
		claim, _, _, _ := evidenceFixture()
		graph := EvidenceGraph{Claims: []Claim{claim}, Nodes: []EvidenceNode{{ID: "a", Claim: claim.ID, Kind: "derived", Rule: "identity", Premises: []string{"b"}}, {ID: "b", Claim: claim.ID, Kind: "derived", Rule: "identity", Premises: []string{"a"}}}}
		if got := graph.Evaluate("a", EvidenceContext{}); got.Status != Unresolved || !strings.Contains(got.Reason, "circular") {
			t.Fatalf("circular proof accepted: %+v", got)
		}
	})
	t.Run("FALS-004", func(t *testing.T) {
		claim, witness, policy, current := evidenceFixture()
		if got := ApproveNativeClaim(policy, witness, claim, current); got.Status != Established {
			t.Fatal(got)
		}
		policy.ApprovedWitnesses = nil
		if got := ApproveNativeClaim(policy, witness, claim, current); got.Status != Unresolved {
			t.Fatalf("removed approval still establishes: %+v", got)
		}
	})
}

func TestEvidenceExactScopeAndCurrentImplementation(t *testing.T) {
	claim, witness, policy, current := evidenceFixture()
	for name, change := range map[string]func(*Claim, *NativeWitness, *TrustPolicy, *EvidenceFingerprints){
		"target-role": func(c *Claim, _ *NativeWitness, _ *TrustPolicy, _ *EvidenceFingerprints) {
			c.Target.Path = []string{"other-worker"}
		},
		"body": func(_ *Claim, _ *NativeWitness, _ *TrustPolicy, f *EvidenceFingerprints) {
			f.Implementation = "body-v2"
		},
		"target": func(_ *Claim, _ *NativeWitness, _ *TrustPolicy, f *EvidenceFingerprints) { f.Target = "other-target" },
		"cleanup": func(_ *Claim, _ *NativeWitness, _ *TrustPolicy, f *EvidenceFingerprints) {
			f.Dependencies = map[string]string{"drop-helper": "drop-v2"}
		},
		"omitted-dependency": func(_ *Claim, w *NativeWitness, _ *TrustPolicy, _ *EvidenceFingerprints) { w.Dependencies = nil },
		"whole-region":       func(c *Claim, _ *NativeWitness, _ *TrustPolicy, _ *EvidenceFingerprints) { c.Scope = "body_only" },
		"panic-outcome": func(c *Claim, _ *NativeWitness, _ *TrustPolicy, _ *EvidenceFingerprints) {
			c.Outcomes = []Outcome{Normal, Panic}
		},
		"production": func(_ *Claim, _ *NativeWitness, p *TrustPolicy, _ *EvidenceFingerprints) { p.AllowTestOnly = false },
		"no-policy":  func(_ *Claim, _ *NativeWitness, p *TrustPolicy, _ *EvidenceFingerprints) { *p = TrustPolicy{} },
	} {
		t.Run(name, func(t *testing.T) {
			c, w, p, f := claim, witness, policy, current
			change(&c, &w, &p, &f)
			if got := ApproveNativeClaim(p, w, c, f); got.Status != Unresolved {
				t.Fatalf("mismatched assurance accepted: %+v", got)
			}
		})
	}
}

func TestEvidenceGuardDominanceAndStability(t *testing.T) {
	claim, _, _, _ := evidenceFixture()
	guard := GuardWitness{ID: "checked-state", ClaimFingerprint: ClaimFingerprint(claim), Dominates: true, ResourceVersions: map[string]string{"process": "generation-1"}, Outcomes: []Outcome{Normal}}
	graph := EvidenceGraph{Claims: []Claim{claim}, Nodes: []EvidenceNode{{ID: "guard", Claim: claim.ID, Kind: "runtime_guard", Guard: guard.ID}}}
	context := EvidenceContext{Guards: map[string]GuardWitness{guard.ID: guard}, ResourceVersions: map[string]string{"process": "generation-1"}}
	if got := graph.Evaluate("guard", context); got.Status != Established {
		t.Fatal(got)
	}
	context.ResourceVersions["process"] = "generation-2"
	if got := graph.Evaluate("guard", context); got.Status != Unresolved {
		t.Fatalf("stale guard accepted: %+v", got)
	}
	context.ResourceVersions["process"] = "generation-1"
	guard.Dominates = false
	context.Guards[guard.ID] = guard
	if got := graph.Evaluate("guard", context); got.Status != Unresolved {
		t.Fatalf("non-dominating guard accepted: %+v", got)
	}
}

func TestEvidenceConjunctionRequiresEachProposition(t *testing.T) {
	claim, _, _, current := evidenceFixture()
	other := claim
	other.ID, other.Predicate = "finite", Predicate{Kind: "finite", Subject: "number"}
	combined := claim
	combined.ID, combined.Predicate = "both", Predicate{Kind: "and", Args: []Predicate{claim.Predicate, other.Predicate}}
	graph := EvidenceGraph{Claims: []Claim{claim, other, combined}, Nodes: []EvidenceNode{{ID: "a", Claim: claim.ID, Kind: "checked_derivation", Witness: "check-a"}, {ID: "b", Claim: other.ID, Kind: "checked_derivation", Witness: "check-b"}, {ID: "both", Claim: combined.ID, Kind: "derived", Rule: "conjunction", Premises: []string{"a", "b"}}}}
	context := EvidenceContext{Current: map[string]EvidenceFingerprints{"a": current, "b": current}, Checked: map[string]CheckedEvidence{"check-a": {ClaimFingerprint: ClaimFingerprint(claim), Checker: "typed-effects", Fingerprints: current}, "check-b": {ClaimFingerprint: ClaimFingerprint(other), Checker: "finite-number", Fingerprints: current}}}
	if got := graph.Evaluate("both", context); got.Status != Established {
		t.Fatal(got)
	}
	graph.Nodes[2].Premises = []string{"a", "a"}
	if got := graph.Evaluate("both", context); got.Status != Unresolved {
		t.Fatalf("same evidence reused for unrelated proposition: %+v", got)
	}
}
