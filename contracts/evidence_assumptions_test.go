package contracts

import (
	"reflect"
	"strings"
	"testing"
)

func TestEvidenceRetainsClaimAssumptions(t *testing.T) {
	claim, witness, policy, current := evidenceFixture()
	claim.Assumptions = []string{"required-native-X", "required-native-X"}
	witness.ClaimFingerprint = ClaimFingerprint(claim)
	want := []string{"required-native-X", witness.ID}
	if got := ApproveNativeClaim(policy, witness, claim, current); got.Status != Established || !reflect.DeepEqual(got.Assumptions, want) || !strings.Contains(got.Reason, "test-only") || !strings.Contains(got.Reason, "no production approval") || !strings.Contains(got.Reason, "required-native-X") {
		t.Fatalf("native approval dropped conditional premise: %+v", got)
	}
	graph := EvidenceGraph{Claims: []Claim{claim}, Nodes: []EvidenceNode{{ID: "checked", Claim: claim.ID, Kind: "checked_derivation", Witness: "checker"}}}
	context := EvidenceContext{Checked: map[string]CheckedEvidence{"checker": {ClaimFingerprint: ClaimFingerprint(claim), Checker: "typed-checker", Fingerprints: current}}, Current: map[string]EvidenceFingerprints{"checked": current}}
	if got := graph.Evaluate("checked", context); got.Status != Established || !reflect.DeepEqual(got.Assumptions, []string{"required-native-X"}) {
		t.Fatalf("logical checker dropped conditional premise: %+v", got)
	}
}
