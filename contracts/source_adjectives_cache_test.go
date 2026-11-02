package contracts

import "testing"

func TestSourceCertificateCacheBoundaries(t *testing.T) {
	projection, declaration := sourceAdjectiveFixture()
	session := NewSession()
	certificate, err := session.VerifySourceAdjective(projection, declaration, AnalysisOptions{})
	if err != nil || certificate.LogicalStatus != Established {
		t.Fatalf("initial: %+v %v", certificate, err)
	}
	instance := sourceInstanceFixture(projection)
	if _, err := session.VerifySourceAdjectiveInstance(projection, certificate, declaration, instance, AnalysisOptions{}); err != nil {
		t.Fatal(err)
	}
	if _, err := session.VerifySourceAdjectiveInstance(projection, certificate, declaration, instance, AnalysisOptions{}); err != nil {
		t.Fatal(err)
	}
	before := session.CacheStats()
	if before.Implementation.Hits != 1 {
		t.Fatalf("concrete cache not used: %+v", before)
	}
	projection.Dependencies = map[string]string{"source-adjective.l": "native-body-v2"}
	projection.Environment = "revalidated-import-environment"
	projection.LoweringObligations = []Obligation{{ID: "native-body-v2", Status: Unresolved, Reason: "changed implementation requires conformance"}}
	current, err := session.VerifySourceAdjective(projection, declaration, AnalysisOptions{NativeWitnesses: map[string]string{"leaf": "ignored-native-input"}, ApprovedAssumptions: map[string]bool{"ignored-native-input": true}})
	if err != nil || current.LogicalStatus != Established || current.Environment != projection.Environment || current.LoweringObligations[0].ID != "native-body-v2" || current.ProjectionFingerprint == certificate.ProjectionFingerprint {
		t.Fatalf("provenance/native obligations stale: %+v %v", current, err)
	}
	if session.CacheStats().Logical.Hits != before.Logical.Hits+1 {
		t.Fatal("native-only edit invalidated semantic theorem")
	}
	if _, err := session.VerifySourceAdjectiveInstance(projection, current, declaration, instance, AnalysisOptions{}); err != nil {
		t.Fatal(err)
	}
	if session.CacheStats().Implementation.Misses != before.Implementation.Misses+1 {
		t.Fatal("native edit reused concrete assurance")
	}
	old := session.CacheStats()
	projection.After.Children[0].Name = "changed-source-body"
	if _, err := session.VerifySourceAdjective(projection, declaration, AnalysisOptions{}); err != nil {
		t.Fatal(err)
	}
	if session.CacheStats().Logical.Misses != old.Logical.Misses+1 {
		t.Fatal("semantic source edit reused theorem")
	}
}
