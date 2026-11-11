package contracts

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestContractsSchemaVersionBoundary(t *testing.T) {
	if SchemaVersion != "2.0" || AnalyzerVersion != "l-contracts-2.0/1" {
		t.Errorf("unexpected current model identity: %s / %s", SchemaVersion, AnalyzerVersion)
	}
	_, _, environment := nativeAssuranceFixture()
	_, witness, policy, _ := evidenceFixture()
	documents := map[string]any{
		"native_environment":       environment,
		"contract_check_requests":  CheckRequests{SchemaVersion: SchemaVersion, Kind: "contract_check_requests", Producer: "test-only", Version: "1", Checks: []CheckRequest{{ID: "no-block", Scope: "whole_entry", Profile: Profile{Name: "NoBlocking"}}}},
		"native_contract_manifest": nativeManifestFixture(),
		"facet_and_role_manifest":  FacetRoleManifest{SchemaVersion: SchemaVersion, Kind: "facet_and_role_manifest", Producer: "test-only"},
		"trust_policy":             TrustPolicyDocument{SchemaVersion: SchemaVersion, Kind: "trust_policy", Producer: "test-only", Policy: policy, Witnesses: []NativeWitness{witness}},
		"contract_report":          NewReport("test-only"),
		"adjective_certificate":    AdjectiveCertificateDocument{SchemaVersion: SchemaVersion, Kind: "adjective_certificate", Producer: "test-only"},
	}
	for kind, document := range documents {
		t.Run(kind, func(t *testing.T) {
			data, err := EncodeDocument(document)
			if err != nil {
				t.Fatal(err)
			}
			if _, err := DecodeDocument(data); err != nil {
				t.Fatalf("current document rejected: %v", err)
			}
			for _, version := range []string{"1.1", "99"} {
				t.Run(version, func(t *testing.T) {
					invalid := replaceDocumentField(t, data, "schema_version", version)
					if _, err := DecodeDocument(invalid); err == nil || !strings.Contains(err.Error(), "unsupported contracts schema version") {
						t.Fatalf("unsupported version %s was not rejected: %v", version, err)
					}
				})
			}
		})
	}
}

func TestRetiredAssuranceFieldsReject(t *testing.T) {
	_, _, environment := nativeAssuranceFixture()
	for _, test := range []struct {
		name     string
		document any
		field    string
		value    any
	}{
		{"lifecycle", NewReport("test-only"), "lifecycle", []any{}},
		{"runtime", environment, "runtime", "test-only-runtime"},
	} {
		t.Run(test.name, func(t *testing.T) {
			data, err := EncodeDocument(test.document)
			if err != nil {
				t.Fatal(err)
			}
			data = replaceDocumentField(t, data, test.field, test.value)
			if _, err := DecodeDocument(data); err == nil || !strings.Contains(err.Error(), "unknown contract field document."+test.field) {
				t.Fatalf("retired field was not strictly rejected in schema 2.0: %v", err)
			}
		})
	}
}

func TestRetiredLifecycleProfileRejects(t *testing.T) {
	if _, supported := LookupProfile("OwnedLifecycle"); supported {
		t.Error("retired lifecycle checker remains selectable")
	}
	requests := CheckRequests{SchemaVersion: SchemaVersion, Kind: "contract_check_requests", Producer: "test-only", Version: "1", Checks: []CheckRequest{{ID: "retired", Scope: "whole_entry", Profile: Profile{Name: "OwnedLifecycle"}}}}
	if _, err := EncodeDocument(requests); err == nil || !strings.Contains(err.Error(), "unsupported requested profile") {
		t.Fatalf("retired lifecycle profile was not rejected: %v", err)
	}
}

func TestRetiredTrustPolicyVersionCannotApprove(t *testing.T) {
	claim, witness, policy, current := evidenceFixture()
	policy.Version = "1.1"
	if got := ApproveNativeClaim(policy, witness, claim, current); got.Status != Unresolved {
		t.Fatalf("retired policy approved native evidence: %+v", got)
	}
	document := TrustPolicyDocument{SchemaVersion: SchemaVersion, Kind: "trust_policy", Producer: "test-only", Policy: policy, Witnesses: []NativeWitness{witness}}
	if _, err := EncodeDocument(document); err == nil {
		t.Fatal("current document accepted a retired embedded policy version")
	}
}

func replaceDocumentField(t *testing.T, data []byte, field string, value any) []byte {
	t.Helper()
	var document map[string]json.RawMessage
	if err := json.Unmarshal(data, &document); err != nil {
		t.Fatal(err)
	}
	encoded, err := json.Marshal(value)
	if err != nil {
		t.Fatal(err)
	}
	document[field] = encoded
	result, err := json.Marshal(document)
	if err != nil {
		t.Fatal(err)
	}
	return result
}
