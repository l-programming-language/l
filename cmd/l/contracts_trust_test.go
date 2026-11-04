package main

import (
	"bytes"
	"strings"
	"testing"

	"github.com/l-programming-language/l/contracts"
)

func contractTrustFixture(t *testing.T) (contracts.Report, contractMetadata) {
	t.Helper()
	target := contracts.TargetRef{Root: "test-app", Path: []string{"worker"}, Type: "Fixture", Phase: contracts.Invocation}
	report := contracts.NewReport("test-only typed fixture")
	report.Scope = contracts.ReportScope{Phase: contracts.Invocation, Boundary: "whole_region", Target: "test-only-target", Platform: "test-only-platform", Compiler: "test-only-compiler", PanicPolicy: "unwind"}
	report.Inputs = map[string]string{"fixture.l": "source-v1"}
	report.Lowering.Steps = []contracts.LoweringStep{{ID: "use-1", Meaning: "fixture:meaning", Name: "Fixture", Kind: "native", Target: target, Implementation: "body-v1"}}
	report.Lowering.Fingerprints = map[string]string{"fixture:meaning": "body-v1", "drop-helper": "drop-v1"}
	report.Summary.Coverage = contracts.Partial
	report.Summary.Unknowns = []contracts.UnknownReason{{Code: "other-native", Message: "surrounding native closure remains unresolved"}}
	claim := contracts.Claim{ID: "leaf-no-block", Predicate: contracts.Predicate{Kind: "effects_closed", Subject: "blocking", Value: contracts.Atom{Kind: "bool", Bool: false}}, Target: target, Scope: "whole_region", Phase: contracts.Invocation, Outcomes: []contracts.Outcome{contracts.Normal}}
	declaration := contracts.DeclaredContract{ID: "fixture-contract", Meaning: "fixture:meaning", Name: "Fixture", Guarantees: []contracts.Claim{claim}, Origin: contracts.Origin{Source: "test-only-fixture"}}
	current, err := contractNativeFingerprints(report, declaration)
	if err != nil {
		t.Fatal(err)
	}
	witness := contracts.NativeWitness{ID: "test-native-approval", ClaimFingerprint: contracts.ClaimFingerprint(claim), ContractFingerprint: current.Contract, ImplementationFingerprint: current.Implementation, TargetFingerprint: current.Target, Authority: "test-only-authority", Rationale: "Explicit synthetic test assumption, never production approval", Dependencies: current.Dependencies, TestOnly: true}
	metadata := contractMetadata{Native: []contracts.DeclaredContract{declaration}, Trust: contracts.TrustPolicyDocument{SchemaVersion: contracts.SchemaVersion, Kind: "trust_policy", Producer: "test-only-fixture", Policy: contracts.TrustPolicy{Version: contracts.SchemaVersion, ID: "test-only-policy", Authorities: []string{witness.Authority}, ApprovedWitnesses: []string{witness.ID}, AllowTestOnly: true}, Witnesses: []contracts.NativeWitness{witness}}}
	return report, metadata
}

func TestContractTrustIsPropositionSpecific(t *testing.T) {
	t.Run("approved-conditional-claim", func(t *testing.T) {
		report, metadata := contractTrustFixture(t)
		if err := applyContractTrust(&report, metadata); err != nil {
			t.Fatal(err)
		}
		if len(report.Obligations) != 1 || report.Obligations[0].Status != contracts.Established || len(report.Assumptions) != 1 || report.Assumptions[0] != "test-native-approval" {
			t.Fatalf("explicit conditional assurance missing: %+v", report.Obligations)
		}
		if report.Summary.Coverage != contracts.Partial || len(report.Summary.Unknowns) != 1 || len(report.Summary.Assumptions) != 0 {
			t.Fatal("native claim silently promoted surrounding summary")
		}
		if report.TrustPolicy == nil || len(report.NativeWitnesses) != 1 || len(report.Evidence.Nodes) != 1 {
			t.Fatal("approval provenance absent from report")
		}
		if _, err := contracts.EncodeDocument(report); err != nil {
			t.Fatal(err)
		}
		var text bytes.Buffer
		if err := writeContractText(&text, report); err != nil {
			t.Fatal(err)
		}
		if !strings.Contains(text.String(), "Assumption: test-native-approval") || !strings.Contains(text.String(), "no production approval") {
			t.Fatal("text report omitted conditional test-only approval scope")
		}
	})
	for _, mutation := range []string{"missing-policy", "body", "cleanup", "source", "target", "role", "declaration", "unknown-attachment"} {
		t.Run(mutation, func(t *testing.T) {
			report, metadata := contractTrustFixture(t)
			switch mutation {
			case "missing-policy":
				metadata.Trust = contracts.TrustPolicyDocument{}
			case "body":
				report.Lowering.Steps[0].Implementation = "body-v2"
			case "cleanup":
				report.Lowering.Fingerprints["drop-helper"] = "drop-v2"
			case "source":
				report.Inputs["fixture.l"] = "source-v2"
			case "target":
				report.Scope.PanicPolicy = "abort"
			case "role":
				metadata.Native[0].Guarantees[0].Target.Path = []string{"other-worker"}
			case "declaration":
				metadata.Native[0].Requirements = []contracts.Predicate{{Kind: "present", Subject: "extra-precondition"}}
			case "unknown-attachment":
				report.Lowering.Steps[0].Target = contracts.TargetRef{Unknown: true}
			}
			if err := applyContractTrust(&report, metadata); err != nil {
				t.Fatal(err)
			}
			if len(report.Obligations) == 0 || report.Obligations[len(report.Obligations)-1].Status != contracts.Unresolved || len(report.Assumptions) != 0 {
				t.Fatalf("stale/unapproved assertion accepted: %+v", report.Obligations)
			}
		})
	}
}

func TestContractTrustRequiresConcreteNativeConfiguration(t *testing.T) {
	report, metadata := contractTrustFixture(t)
	report.Scope.Target = "native-workspace"
	report.Scope.PanicPolicy = "selected-runtime"
	// Even a witness freshly keyed to the source-only placeholder environment
	// cannot certify an unknown Rust toolchain or panic configuration.
	fingerprint, err := contracts.Fingerprint(report.Scope)
	if err != nil {
		t.Fatal(err)
	}
	metadata.Trust.Witnesses[0].TargetFingerprint = fingerprint
	if err := applyContractTrust(&report, metadata); err != nil {
		t.Fatal(err)
	}
	if len(report.Obligations) != 1 || report.Obligations[0].Status != contracts.Unresolved || !strings.Contains(report.Obligations[0].Reason, "Rust toolchain") {
		t.Fatalf("unknown native configuration granted assurance: %+v", report.Obligations)
	}
}

func TestContractTrustRequiresCurrentApplicability(t *testing.T) {
	for _, test := range []struct {
		name string
		want contracts.Status
	}{
		{"false-requirement", contracts.Contradicted},
		{"unknown-entry-requirement", contracts.Unresolved},
		{"wrong-target-restriction", contracts.Contradicted},
		{"unsupported-target-restriction", contracts.Unresolved},
		{"unprojected-signature", contracts.Unresolved},
		{"valid-target-type", contracts.Established},
	} {
		t.Run(test.name, func(t *testing.T) {
			report, metadata := contractTrustFixture(t)
			declaration := &metadata.Native[0]
			switch test.name {
			case "false-requirement":
				declaration.Requirements = []contracts.Predicate{{Kind: "false"}}
			case "unknown-entry-requirement":
				requirement := contracts.Predicate{Kind: "present", Subject: "available-before-call"}
				declaration.Requirements = []contracts.Predicate{requirement}
				// An end-state observation is deliberately not an entry premise.
				report.Summary.Paths = []contracts.PathSummary{{Outcome: contracts.Normal, Reachable: true, Facts: []contracts.Predicate{requirement}}}
			case "wrong-target-restriction":
				declaration.TargetRestrictions = []contracts.Predicate{{Kind: "type", Subject: "target", Value: contracts.Atom{Kind: "type", Text: "OtherType"}}}
			case "unsupported-target-restriction":
				declaration.TargetRestrictions = []contracts.Predicate{{Kind: "native-platform-predicate", Subject: "target"}}
			case "unprojected-signature":
				declaration.Signature = &contracts.NativeSignature{Inputs: []string{"unverified-input"}, Outputs: []string{"unverified-output"}}
			case "valid-target-type":
				declaration.Requirements = []contracts.Predicate{{Kind: "type", Subject: "target", Value: contracts.Atom{Kind: "type", Text: "Fixture"}}}
			}
			current, err := contractNativeFingerprints(report, *declaration)
			if err != nil {
				t.Fatal(err)
			}
			metadata.Trust.Witnesses[0].ContractFingerprint = current.Contract
			if got := contracts.ApproveNativeClaim(metadata.Trust.Policy, metadata.Trust.Witnesses[0], declaration.Guarantees[0], current); got.Status != contracts.Established {
				t.Fatalf("fixture must have a current valid native witness: %+v", got)
			}
			if err := applyContractTrust(&report, metadata); err != nil {
				t.Fatal(err)
			}
			if len(report.Obligations) != 2 || report.Obligations[0].Status != test.want {
				t.Fatalf("applicability obligation omitted or mischecked: %+v", report.Obligations)
			}
			guarantee := report.Obligations[1]
			if test.want == contracts.Established {
				if guarantee.Status != contracts.Established || len(report.Assumptions) != 1 {
					t.Fatalf("supported typed applicability failed: %+v", report.Obligations)
				}
			} else if guarantee.Status != contracts.Unresolved || len(report.Assumptions) != 0 {
				t.Fatalf("current witness erased unsatisfied applicability: %+v", report.Obligations)
			}
			node := report.Evidence.Nodes[len(report.Evidence.Nodes)-1]
			if len(node.Premises) != 1 || node.Premises[0] != report.Evidence.Nodes[0].ID {
				t.Fatalf("guarantee graph omitted applicability dependency: %+v", node)
			}
			context := contracts.EvidenceContext{Policy: metadata.Trust.Policy, NativeWitnesses: map[string]contracts.NativeWitness{metadata.Trust.Witnesses[0].ID: metadata.Trust.Witnesses[0]}, Current: map[string]contracts.EvidenceFingerprints{node.ID: current}}
			if got := report.Evidence.Evaluate(node.ID, context); got.Status != contracts.Unresolved {
				t.Fatalf("native witness alone bypassed required checker evidence: %+v", got)
			}
			if _, err := contracts.EncodeDocument(report); err != nil {
				t.Fatal(err)
			}
		})
	}
}

func TestNativeRequirementsWithoutAdvertisedGuarantees(t *testing.T) {
	report, metadata := contractTrustFixture(t)
	metadata.Native[0].Guarantees = nil
	metadata.Native[0].Requirements = []contracts.Predicate{{Kind: "false"}}
	metadata.Trust = contracts.TrustPolicyDocument{}
	if err := applyContractTrust(&report, metadata); err != nil {
		t.Fatal(err)
	}
	if len(report.Obligations) != 1 || report.Obligations[0].Status != contracts.Contradicted {
		t.Fatalf("selected requirements ignored without an advertised guarantee: %+v", report.Obligations)
	}
}
