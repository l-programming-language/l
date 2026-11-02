package contracts

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func nativeManifestFixture() NativeContractManifest {
	return NativeContractManifest{SchemaVersion: SchemaVersion, Kind: "native_contract_manifest", Producer: "test-fixture", Declarations: []DeclaredContract{{ID: "fixture-native", Meaning: "fixture:meaning", Name: "Fixture", Origin: Origin{Source: "typed-fixture", Start: 1, End: 2}}}}
}

func TestMetadataStrictness(t *testing.T) {
	manifest := nativeManifestFixture()
	data, err := EncodeDocument(manifest)
	if err != nil {
		t.Fatal(err)
	}
	t.Run("BASE-004", func(t *testing.T) {
		for name, invalid := range map[string][]byte{
			"version":                bytes.Replace(data, []byte(`"schema_version": "1.1"`), []byte(`"schema_version": "99"`), 1),
			"misspelled-guarantee":   bytes.Replace(data, []byte(`"name": "Fixture",`), []byte(`"name": "Fixture", "guarantes": [],`), 1),
			"case-insensitive-field": bytes.Replace(data, []byte(`"name": "Fixture"`), []byte(`"NaMe": "Fixture"`), 1),
			"self-established":       bytes.Replace(data, []byte(`"name": "Fixture",`), []byte(`"name": "Fixture", "status": "established",`), 1),
			"duplicate-key":          bytes.Replace(data, []byte(`"name": "Fixture"`), []byte(`"name": "Fixture", "name": "Other"`), 1),
			"null-identity":          bytes.Replace(data, []byte(`"meaning": "fixture:meaning"`), []byte(`"meaning": null`), 1),
			"missing-identity":       bytes.Replace(data, []byte(`"meaning": "fixture:meaning",`), nil, 1),
			"trailing-document":      append(append([]byte(nil), data...), []byte(`{}`)...),
		} {
			t.Run(name, func(t *testing.T) {
				if _, err := DecodeDocument(invalid); err == nil {
					t.Fatalf("accepted malformed metadata: %s", invalid)
				}
			})
		}
		ambiguous := manifest
		ambiguous.Declarations = append(append([]DeclaredContract(nil), manifest.Declarations...), manifest.Declarations[0])
		ambiguous.Declarations[1].ID = "another-contract"
		if _, err := EncodeDocument(ambiguous); err == nil {
			t.Fatal("ambiguous canonical meaning accepted")
		}
	})
	t.Run("FALS-006", func(t *testing.T) {
		invalid := bytes.Replace(data, []byte(`"name": "Fixture",`), []byte(`"name": "Fixture", "noHeapAllocaton": true,`), 1)
		if _, err := DecodeDocument(invalid); err == nil {
			t.Fatal("favorable misspelling ignored")
		}
		loaded, err := ReadNativeManifest(data)
		if err != nil {
			t.Fatal(err)
		}
		if got := NormalizeNativeContract(loaded.Declarations[0]); got.Assurance != Unresolved || got.Effects.UnknownTail == nil || !*got.Effects.UnknownTail {
			t.Fatalf("omission granted no-allocation: %+v", got)
		}
	})
}

func TestNativeNormalizationAndEvidenceSeparation(t *testing.T) {
	declaration := nativeManifestFixture().Declarations[0]
	t.Run("BASE-005", func(t *testing.T) {
		normalized := NormalizeNativeContract(declaration)
		for family, coverage := range normalized.Coverage {
			if coverage != Unknown {
				t.Fatalf("omitted family %s became %s", family, coverage)
			}
		}
		if normalized.Effects.UnknownTail == nil || !*normalized.Effects.UnknownTail || len(normalized.Unknowns) < 10 {
			t.Fatalf("unknowns not normalized: %+v", normalized)
		}
	})
	t.Run("BASE-006", func(t *testing.T) {
		claim, witness, _, current := evidenceFixture()
		favorable := declaration
		favorable.Guarantees = []Claim{claim}
		closed := false
		favorable.Effects = &DeclaredEffectRow{UnknownTail: &closed}
		if got := NormalizeNativeContract(favorable); got.Assurance != Unresolved || len(got.Declaration.Guarantees) != 1 {
			t.Fatalf("declaration became proof: %+v", got)
		}
		if got := ApproveNativeClaim(TrustPolicy{}, witness, claim, current); got.Status != Unresolved {
			t.Fatalf("favorable declaration auto-approved: %+v", got)
		}
	})
	t.Run("BASE-007", func(t *testing.T) {
		irrelevant := declaration
		irrelevant.Coverage = map[string]Coverage{"C3": NotApplicable, "C4": NotApplicable}
		irrelevant.Effects = &DeclaredEffectRow{Effects: []Effect{{Kind: "block", Phase: Destruction}}}
		got := NormalizeNativeContract(irrelevant)
		if got.Coverage["C3"] != Unknown || got.Coverage["C4"] != Unknown || !*got.Effects.UnknownTail || len(got.Effects.Effects) != 1 {
			t.Fatalf("not-applicable assertion erased behavior: %+v", got)
		}
	})
}

func TestReportSerializationDeterministic(t *testing.T) {
	t.Run("BASE-008", func(t *testing.T) {
		report := NewReport("typed-fixture")
		report.Inputs["z-source"], report.Inputs["a-source"] = "z-hash", "a-hash"
		report.Scope.Exclusions = []string{"external forced termination"}
		report.Summary.Unknowns = []UnknownReason{{Code: "opaque", Message: "native callback effects", Components: []string{"effects"}}}
		report.Targets = []TargetRef{{Root: "A", Path: []string{"first"}, Type: "Theme"}, {Root: "A", Path: []string{"second"}, Type: "Theme"}}
		first, err := EncodeDocument(report)
		if err != nil {
			t.Fatal(err)
		}
		report.Inputs = map[string]string{"a-source": "a-hash", "z-source": "z-hash"}
		second, err := EncodeDocument(report)
		if err != nil || !bytes.Equal(first, second) {
			t.Fatalf("map insertion order changed serialization: %v", err)
		}
		decoded, err := DecodeDocument(first)
		if err != nil {
			t.Fatal(err)
		}
		third, err := EncodeDocument(decoded)
		if err != nil || !bytes.Equal(first, third) || !reflect.DeepEqual(decoded, report) {
			t.Fatalf("round trip lost semantic data: %v", err)
		}
	})
}

func executionFixture() ExecutionRecord {
	exit := 0
	return SealExecutionRecord(ExecutionRecord{ID: "fixture-run", Kind: "test", Command: []string{"fixture-test"}, WorkingDirectory: "/fixture", Status: "passed", Started: "2026-11-01T00:00:00Z", Finished: "2026-11-01T00:00:01Z", ExitCode: &exit, StdoutFingerprint: "empty-stdout", StderrFingerprint: "empty-stderr"})
}

func TestExecutedEvidenceRequiresRecordedOperation(t *testing.T) {
	t.Run("FALS-010", func(t *testing.T) {
		report := NewReport("fixture")
		claim := Claim{ID: "observed", Predicate: Predicate{Kind: "observed", Subject: "one-input"}, Phase: Invocation, Outcomes: []Outcome{Normal}}
		report.Evidence.Claims = []Claim{claim}
		report.Evidence.Nodes = []EvidenceNode{{ID: "empirical", Claim: claim.ID, Kind: "empirical", Witness: "fixture-run"}}
		if _, err := EncodeDocument(report); err == nil {
			t.Fatal("fabricated executed evidence accepted without record")
		}
		report.Executions = []ExecutionRecord{{ID: "fixture-run", Kind: "test", Command: []string{"fixture-test"}, WorkingDirectory: "/fixture", Status: "passed"}}
		if _, err := EncodeDocument(report); err == nil {
			t.Fatal("fabricated success with no timestamps/status/provenance accepted")
		}
		report.Executions = []ExecutionRecord{executionFixture()}
		if _, err := EncodeDocument(report); err != nil {
			t.Fatal(err)
		}
		report.Executions[0].Kind = "integrity"
		report.Executions[0] = SealExecutionRecord(report.Executions[0])
		if _, err := EncodeDocument(report); err == nil || !strings.Contains(err.Error(), "scope") {
			t.Fatalf("integrity observation became compiler test: %v", err)
		}
	})
}

func TestDocumentSchemasFollowSharedModel(t *testing.T) {
	for _, kind := range []string{"native_contract_manifest", "facet_and_role_manifest", "trust_policy", "contract_report", "adjective_certificate"} {
		t.Run(kind, func(t *testing.T) {
			first, err := JSONSchema(kind)
			if err != nil {
				t.Fatal(err)
			}
			second, err := JSONSchema(kind)
			if err != nil || !bytes.Equal(first, second) {
				t.Fatalf("schema generation nondeterministic: %v", err)
			}
			retained, err := os.ReadFile(filepath.Join("schemas", kind+".schema.json"))
			if err != nil || !bytes.Equal(first, retained) {
				t.Fatalf("retained schema differs from shared model; regenerate with go run ./tools/contracts-schema %s: %v", kind, err)
			}
			var schema map[string]json.RawMessage
			if err := json.Unmarshal(first, &schema); err != nil {
				t.Fatal(err)
			}
			if len(schema["$defs"]) == 0 || len(schema["$ref"]) == 0 || !bytes.Contains(first, []byte(`"additionalProperties": false`)) {
				t.Fatalf("schema lacks model/reference validation: %s", first)
			}
		})
	}
}
