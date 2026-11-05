package contracts

import "testing"

func TestGuardRequiresPropositionResourceDependencies(t *testing.T) {
	claim, _, _, _ := evidenceFixture()
	claim.Predicate = Predicate{Kind: "protocol", Subject: "Running", Resource: ResourceRef{ID: "process"}}
	guard := GuardWitness{ID: "running", ClaimFingerprint: ClaimFingerprint(claim), Dominates: true, Outcomes: []Outcome{Normal}}
	graph := EvidenceGraph{Claims: []Claim{claim}, Nodes: []EvidenceNode{{ID: "guard", Claim: claim.ID, Kind: "runtime_guard", Guard: guard.ID}}}
	context := EvidenceContext{Guards: map[string]GuardWitness{guard.ID: guard}, ResourceVersions: map[string]string{"process": "before-completion"}}
	if got := graph.Evaluate("guard", context); got.Status != Unresolved {
		t.Fatalf("missing dependency established mutable protocol fact: %+v", got)
	}
	guard.ResourceVersions = map[string]string{"process": "before-completion"}
	context.Guards[guard.ID] = guard
	if got := graph.Evaluate("guard", context); got.Status != Established {
		t.Fatalf("stable checked proposition not supported: %+v", got)
	}
	context.ResourceVersions["process"] = "external-completion"
	if got := graph.Evaluate("guard", context); got.Status != Unresolved {
		t.Fatalf("external completion retained stale Running fact: %+v", got)
	}
}
