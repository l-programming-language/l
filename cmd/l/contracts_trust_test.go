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
	claim := contracts.Claim{ID: "leaf-no-block", Predicate: contracts.Predicate{Kind: "effects_closed", Subject: "blocking", Value: contracts.Atom{Kind: "bool", Bool: false}}, Target: target, Scope: "native_body", Phase: contracts.Invocation, Outcomes: []contracts.Outcome{contracts.Normal}}
	signature := &contracts.NativeSignature{Inputs: []string{"Fixture"}, Outputs: []string{"Fixture"}}
	use := contracts.NativeUse{ID: "use-1", Meaning: "fixture:meaning", Name: "Fixture", Target: target, Phase: contracts.Invocation, Boundary: "native_body", Signature: signature, Implementation: "body-v1", OutcomeDisposition: "value"}
	report.NativeUses = []contracts.NativeUse{use}
	environment := contracts.SealNativeEnvironment(contracts.NativeEnvironment{SchemaVersion: contracts.SchemaVersion, Kind: "native_environment", Producer: "test-only-fixture", Version: "1", Method: "test-only", Host: "test-only", Target: "test-only-target", Compiler: "test-only-compiler", PanicPolicy: "unwind", Runtime: "test-only", Allocator: "test-only", Inputs: report.Inputs, Dependencies: report.Lowering.Fingerprints})
	declaration := contracts.DeclaredContract{ID: "fixture-contract", Meaning: "fixture:meaning", Name: "Fixture", Signature: signature, Guarantees: []contracts.Claim{claim}, Origin: contracts.Origin{Source: "test-only-fixture"}}
	current := evaluateNativeUse(use, declaration, environment).Guarantees[0].Current
	witness := contracts.NativeWitness{ID: "test-native-approval", ClaimFingerprint: contracts.ClaimFingerprint(claim), ContractFingerprint: current.Contract, ImplementationFingerprint: current.Implementation, TargetFingerprint: current.Target, Authority: "test-only-authority", Rationale: "Explicit synthetic test assumption, never production approval", Dependencies: current.Dependencies, TestOnly: true}
	metadata := contractMetadata{Environment: &environment, Native: []contracts.DeclaredContract{declaration}, Trust: contracts.TrustPolicyDocument{SchemaVersion: contracts.SchemaVersion, Kind: "trust_policy", Producer: "test-only-fixture", Policy: contracts.TrustPolicy{Version: contracts.SchemaVersion, ID: "test-only-policy", Authorities: []string{witness.Authority}, ApprovedWitnesses: []string{witness.ID}, AllowTestOnly: true}, Witnesses: []contracts.NativeWitness{witness}}}
	return report, metadata
}

func evaluateContractTrust(t *testing.T, report *contracts.Report, metadata contractMetadata) {
	t.Helper()
	options, evaluations, err := prepareNativeApplications(report, metadata)
	if err != nil {
		t.Fatal(err)
	}
	var nodes []contracts.Node
	for _, use := range report.NativeUses {
		nodes = append(nodes, contracts.Node{Kind: contracts.Native, ID: use.ID, Native: use.ID, Meaning: use.Meaning, Target: use.Target, Phase: use.Phase})
	}
	report.Summary, err = contracts.Infer(contracts.Plan{Entry: contracts.Node{Kind: contracts.Sequence, Children: nodes}}, options)
	if err != nil {
		t.Fatal(err)
	}
	appendNativeEvaluations(report, evaluations, metadata.Trust)
}

func TestContractTrustIsPropositionSpecific(t *testing.T) {
	t.Run("approved-conditional-claim", func(t *testing.T) {
		report, metadata := contractTrustFixture(t)
		evaluateContractTrust(t, &report, metadata)
		if len(report.Obligations) == 0 || report.Obligations[len(report.Obligations)-1].Status != contracts.Established || len(report.Assumptions) != 1 || report.Assumptions[0] != "test-native-approval" {
			t.Fatalf("explicit conditional assurance missing: %+v", report.Obligations)
		}
		if report.Summary.Coverage != contracts.Partial || len(report.Summary.Unknowns) == 0 || len(report.Summary.Assumptions) != 0 {
			t.Fatal("native claim silently promoted surrounding summary")
		}
		if report.TrustPolicy == nil || len(report.NativeWitnesses) != 1 || len(report.Evidence.Nodes) < 3 {
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
	for _, mutation := range []string{"missing-policy", "missing-environment", "body", "cleanup", "source", "target", "role", "declaration", "unknown-attachment"} {
		t.Run(mutation, func(t *testing.T) {
			report, metadata := contractTrustFixture(t)
			switch mutation {
			case "missing-policy":
				metadata.Trust = contracts.TrustPolicyDocument{}
			case "missing-environment":
				// Concrete-looking report labels cannot replace sealed evidence.
				metadata.Environment = nil
			case "body":
				report.NativeUses[0].Implementation = "body-v2"
			case "cleanup":
				report.Lowering.Fingerprints["drop-helper"] = "drop-v2"
			case "source":
				report.Inputs["fixture.l"] = "source-v2"
			case "target":
				metadata.Environment.PanicPolicy = "abort"
				*metadata.Environment = contracts.SealNativeEnvironment(*metadata.Environment)
			case "role":
				metadata.Native[0].Guarantees[0].Target.Path = []string{"other-worker"}
			case "declaration":
				metadata.Native[0].Requirements = []contracts.Predicate{{Kind: "present", Subject: "extra-precondition"}}
			case "unknown-attachment":
				report.NativeUses[0].Target = contracts.TargetRef{Unknown: true}
			}
			evaluateContractTrust(t, &report, metadata)
			if len(report.Obligations) == 0 || report.Obligations[len(report.Obligations)-1].Status != contracts.Unresolved || len(report.Assumptions) != 0 {
				t.Fatalf("stale/unapproved assertion accepted: %+v", report.Obligations)
			}
		})
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
			current := evaluateNativeUse(report.NativeUses[0], *declaration, *metadata.Environment).Guarantees[0].Current
			metadata.Trust.Witnesses[0].ContractFingerprint = current.Contract
			if got := contracts.ApproveNativeClaim(metadata.Trust.Policy, metadata.Trust.Witnesses[0], declaration.Guarantees[0], current); got.Status != contracts.Established {
				t.Fatalf("fixture must have a current valid native witness: %+v", got)
			}
			evaluateContractTrust(t, &report, metadata)
			if len(report.Obligations) != 3 || report.Obligations[1].Status != test.want && !(test.name == "unprojected-signature" && report.Obligations[0].Status == test.want) {
				t.Fatalf("applicability obligation omitted or mischecked: %+v", report.Obligations)
			}
			guarantee := report.Obligations[len(report.Obligations)-1]
			if test.want == contracts.Established {
				if guarantee.Status != contracts.Established || len(report.Assumptions) != 1 {
					t.Fatalf("supported typed applicability failed: %+v", report.Obligations)
				}
			} else if guarantee.Status != contracts.Unresolved || len(report.Assumptions) != 0 {
				t.Fatalf("current witness erased unsatisfied applicability: %+v", report.Obligations)
			}
			node := report.Evidence.Nodes[len(report.Evidence.Nodes)-1]
			if len(node.Premises) != 2 || node.Premises[0] != report.Evidence.Nodes[0].ID || node.Premises[1] != report.Evidence.Nodes[1].ID {
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
	evaluateContractTrust(t, &report, metadata)
	if len(report.Obligations) != 2 || report.Obligations[1].Status != contracts.Contradicted {
		t.Fatalf("selected requirements ignored without an advertised guarantee: %+v", report.Obligations)
	}
}

func TestNativeEvaluationRetainsOrphanDeclarations(t *testing.T) {
	report, metadata := contractTrustFixture(t)
	report.NativeUses = nil
	// Multiple orphan declarations must retain distinct unresolved propositions.
	other := metadata.Native[0]
	other.ID, other.Meaning = "other-contract", "other-meaning"
	metadata.Native = append(metadata.Native, other)
	for i := range metadata.Native {
		metadata.Native[i].Outcomes = []contracts.Outcome{contracts.Normal}
	}
	evaluateContractTrust(t, &report, metadata)
	guarantees := 0
	for _, obligation := range report.Obligations {
		if obligation.Claim.Predicate.Kind == "effects_closed" {
			guarantees++
			if obligation.Status != contracts.Unresolved {
				t.Fatal("orphan guarantee became applicable")
			}
		}
	}
	if guarantees != 2 || len(report.NativeUses) != 0 || len(report.Assumptions) != 0 {
		t.Fatal("orphan declarations disappeared or fabricated native uses")
	}
	if _, err := contracts.EncodeDocument(report); err != nil {
		t.Fatal(err)
	}
}

func TestNativeEvaluationAggregatesEntryContexts(t *testing.T) {
	report, metadata := contractTrustFixture(t)
	requirement := contracts.Predicate{Kind: "present", Subject: "entry-ready"}
	metadata.Native[0].Requirements = []contracts.Predicate{requirement}
	current := evaluateNativeUse(report.NativeUses[0], metadata.Native[0], *metadata.Environment).Guarantees[0].Current
	metadata.Trust.Witnesses[0].ContractFingerprint = current.Contract
	options, evaluations, err := prepareNativeApplications(&report, metadata)
	if err != nil {
		t.Fatal(err)
	}
	use := report.NativeUses[0]
	native := contracts.Node{Kind: contracts.Native, ID: use.ID, Native: use.ID, Meaning: use.Meaning, Target: use.Target, Phase: use.Phase}
	plan := contracts.Plan{Entry: contracts.Node{Kind: contracts.Sequence, Children: []contracts.Node{{Kind: contracts.Branch, Then: []contracts.Node{{Kind: contracts.Event, Facts: []contracts.Predicate{requirement}}}}, native}}}
	report.Summary, err = contracts.Infer(plan, options)
	if err != nil {
		t.Fatal(err)
	}
	statuses := make(map[contracts.Status]bool)
	for _, obligation := range report.Summary.Obligations {
		if obligation.ID == evaluations[0].Entry.ID {
			statuses[obligation.Status] = true
		}
	}
	if !statuses[contracts.Established] || !statuses[contracts.Unresolved] {
		t.Fatal("fixture must exercise both favorable and unknown actual entry contexts")
	}
	appendNativeEvaluations(&report, evaluations, metadata.Trust)
	if report.Obligations[len(report.Obligations)-1].Status != contracts.Unresolved || len(report.Assumptions) != 0 {
		t.Fatal("one favorable entry context approved every invocation")
	}
}

func TestNativeEvaluationComponentsAndGuaranteesShareEvidence(t *testing.T) {
	report, metadata := contractTrustFixture(t)
	declaration := &metadata.Native[0]
	declaration.Outcomes = []contracts.Outcome{contracts.Normal}
	declaration.Guarantees = nil
	prepared := evaluateNativeUse(report.NativeUses[0], *declaration, *metadata.Environment)
	declaration.Guarantees = []contracts.Claim{prepared.Components[0].Claim}
	prepared = evaluateNativeUse(report.NativeUses[0], *declaration, *metadata.Environment)
	component := prepared.Components[0]
	witness := &metadata.Trust.Witnesses[0]
	witness.ClaimFingerprint, witness.ContractFingerprint = contracts.ClaimFingerprint(component.Claim), component.Current.Contract
	witness.ImplementationFingerprint, witness.TargetFingerprint, witness.Dependencies = component.Current.Implementation, component.Current.Target, component.Current.Dependencies
	evaluateContractTrust(t, &report, metadata)
	if len(report.NativeComponents) != 1 || report.NativeComponents[0].Status != contracts.Established || report.Obligations[len(report.Obligations)-1].Status != contracts.Established {
		t.Fatalf("same supported proposition disagrees across summary/report: %+v", report.Obligations)
	}
	if contracts.AllEstablished(contracts.CheckProfile(report.Summary, contracts.Profile{Name: "NoBlocking"})) {
		t.Fatal("outcome approval certified unrelated effects")
	}
}

func evaluateNativeUse(use contracts.NativeUse, declaration contracts.DeclaredContract, environment contracts.NativeEnvironment) contracts.NativeEvaluation {
	return contracts.EvaluateNativeUses([]contracts.NativeUse{use}, []contracts.DeclaredContract{declaration}, environment, contracts.TrustPolicy{}, nil, contracts.Unresolved, nil)[0]
}

func TestNativeEvaluationRejectsMismatchedGuaranteeScope(t *testing.T) {
	for _, mutation := range []string{"scope", "phase", "role", "outcomes", "missing-outcomes"} {
		t.Run(mutation, func(t *testing.T) {
			report, metadata := contractTrustFixture(t)
			claim := &metadata.Native[0].Guarantees[0]
			switch mutation {
			case "scope":
				claim.Scope = "whole_region"
			case "phase":
				claim.Phase = contracts.Construction
			case "role":
				claim.Target.Path = []string{"same-type-other-role"}
			case "outcomes":
				metadata.Native[0].Outcomes = []contracts.Outcome{contracts.Normal}
				claim.Outcomes = []contracts.Outcome{contracts.Panic}
			case "missing-outcomes":
				claim.Outcomes = nil
			}
			evaluation := evaluateNativeUse(report.NativeUses[0], metadata.Native[0], *metadata.Environment)
			current := evaluation.Guarantees[0].Current
			witness := &metadata.Trust.Witnesses[0]
			witness.ClaimFingerprint, witness.ContractFingerprint = contracts.ClaimFingerprint(*claim), current.Contract
			if result := contracts.ApproveNativeClaim(metadata.Trust.Policy, *witness, *claim, current); result.Status != contracts.Established {
				t.Fatalf("fixture witness must approve the exact authored proposition: %+v", result)
			}
			evaluateContractTrust(t, &report, metadata)
			if report.Obligations[len(report.Obligations)-1].Status != contracts.Unresolved || len(report.Assumptions) != 0 {
				t.Fatal("fresh witness erased the native use's exact scope, phase, target or outcomes")
			}
		})
	}
}
