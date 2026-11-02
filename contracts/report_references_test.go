package contracts

import "testing"

func TestReportEvidenceReferences(t *testing.T) {
	report := NewReport("typed-fixture")
	report.Obligations = []Obligation{{ID: "guarantee", Status: Established, Evidence: []string{"missing-proof"}}}
	if _, err := EncodeDocument(report); err == nil {
		t.Fatal("obligation with dangling favorable evidence reference accepted")
	}
	claim := Claim{ID: "claim", Predicate: Predicate{Kind: "true"}}
	report.Evidence.Claims = []Claim{claim}
	report.Evidence.Nodes = []EvidenceNode{{ID: "missing-proof", Claim: claim.ID, Kind: "checked_derivation", Witness: "separate-checker"}}
	if _, err := EncodeDocument(report); err != nil {
		t.Fatal(err)
	}
	// Structural validation permits inspectable external evidence nodes; it
	// does not execute a checker or make that node an established proposition.
	if got := report.Evidence.Evaluate("missing-proof", EvidenceContext{}); got.Status != Unresolved {
		t.Fatalf("metadata reference self-certified: %+v", got)
	}
}
