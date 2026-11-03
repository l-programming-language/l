package main

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	distribution "github.com/l-programming-language/l"
	"github.com/l-programming-language/l/contracts"
	Lyger "github.com/l-programming-language/l/lyger"
)

func scopedCLIReport(t *testing.T, input string, expected int, options ...string) contracts.Report {
	t.Helper()
	args := append([]string{"contracts"}, options...)
	args = append(args, "--format", "json", input)
	var stdout, stderr bytes.Buffer
	if status := run(context.Background(), args, &stdout, &stderr); status != expected {
		t.Fatalf("%v exit %d want %d: %s\n%s", args, status, expected, &stderr, &stdout)
	}
	var report contracts.Report
	if err := json.Unmarshal(stdout.Bytes(), &report); err != nil {
		t.Fatalf("decode %s: %v", &stdout, err)
	}
	return report
}

func scopedWrite(t *testing.T, directory, name string, value any) string {
	t.Helper()
	data, err := contracts.EncodeDocument(value)
	if err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(directory, name)
	if err := os.WriteFile(path, data, 0600); err != nil {
		t.Fatal(err)
	}
	return path
}

type scopedNativeFixture struct {
	assets                                                       string
	input, environmentPath, manifestPath, trustPath, requestPath string
	environment                                                  contracts.NativeEnvironment
	manifest                                                     contracts.NativeContractManifest
	trust                                                        contracts.TrustPolicyDocument
	requests                                                     contracts.CheckRequests
	report                                                       contracts.Report
}

func newScopedNativeFixture(t *testing.T, profile, relation string) scopedNativeFixture {
	return newScopedNativeFixtureAssets(t, profile, relation, "")
}

func newScopedNativeFixtureAssets(t *testing.T, profile, relation, assets string) scopedNativeFixture {
	t.Helper()
	directory := t.TempDir()
	f := scopedNativeFixture{input: filepath.Join(directory, "main.l"), assets: assets}
	if err := os.WriteFile(f.input, []byte("Create \"items\" as a New List of 1 and 2.\nPlace 4 in items at 9.\nPlace 5 in items at 10.\n"), 0600); err != nil {
		t.Fatal(err)
	}
	inspection := []string{"inspect"}
	if assets != "" {
		inspection = append(inspection, "--assets", assets)
	}
	f.report = scopedCLIReport(t, f.input, 0, inspection...)
	f.environment = contracts.SealNativeEnvironment(contracts.NativeEnvironment{SchemaVersion: contracts.SchemaVersion, Kind: "native_environment", Producer: "explicit-test-only-fixture", Version: "1", Method: "test-only", Host: "fixture-host", Target: "fixture-target", Compiler: "fixture-toolchain", PanicPolicy: "fixture-unwind", Runtime: "fixture-runtime", Allocator: "fixture-allocator", Inputs: f.report.Inputs, Dependencies: f.report.Lowering.Fingerprints, Limitations: []string{"Synthetic test data, no production authority or executed build"}})
	f.manifest = contracts.NativeContractManifest{SchemaVersion: contracts.SchemaVersion, Kind: "native_contract_manifest", Producer: "test-only-fixture"}
	f.trust = contracts.TrustPolicyDocument{SchemaVersion: contracts.SchemaVersion, Kind: "trust_policy", Producer: "test-only-fixture", Policy: contracts.TrustPolicy{Version: contracts.SchemaVersion, ID: "test-only-policy", Authorities: []string{"fixture-author"}, AllowTestOnly: true}}
	f.requests = contracts.CheckRequests{SchemaVersion: contracts.SchemaVersion, Kind: "contract_check_requests", Version: "1", Producer: "test-only-fixture"}
	seen := make(map[contracts.MeaningKey]bool)
	for _, use := range f.report.NativeUses {
		if use.Name != "Storage.List.Update" {
			continue
		}
		outcome := contracts.Normal
		if relation != "" {
			outcome = contracts.ReturnedFailure
		}
		closed := false
		component := contracts.NativeOutcomeContract{Outcome: outcome, Effects: &contracts.DeclaredEffectRow{UnknownTail: &closed}}
		if relation != "" {
			component.States = []contracts.StateRelation{{Location: contracts.FailureLocation{Resource: contracts.ResourceRef{ID: "List"}}, Observation: "state_at_completion", Relation: relation}}
		}
		declaration := contracts.DeclaredContract{ID: string(use.Meaning), Meaning: use.Meaning, Name: use.Name, Signature: use.Signature, Outcomes: []contracts.Outcome{outcome}, OutcomeContracts: []contracts.NativeOutcomeContract{component}}
		if !seen[use.Meaning] {
			f.manifest.Declarations = append(f.manifest.Declarations, declaration)
			seen[use.Meaning] = true
		}
		for _, claim := range evaluateNativeUse(use, declaration, f.environment, contracts.TrustPolicy{}, nil, contracts.Unresolved).Components {
			witness := contracts.NativeWitness{ID: "test-only:" + claim.Claim.ID, ClaimFingerprint: contracts.ClaimFingerprint(claim.Claim), ContractFingerprint: claim.Current.Contract, ImplementationFingerprint: claim.Current.Implementation, TargetFingerprint: claim.Current.Target, Dependencies: claim.Current.Dependencies, Authority: "fixture-author", Rationale: "Explicit test-only native assumption, never production approval", TestOnly: true}
			f.trust.Witnesses = append(f.trust.Witnesses, witness)
			f.trust.Policy.ApprovedWitnesses = append(f.trust.Policy.ApprovedWitnesses, witness.ID)
		}
		for _, scope := range f.report.Scopes {
			if scope.Boundary != "native_body" || scope.NodeID != use.ID {
				continue
			}
			request := contracts.CheckRequest{ID: use.ID, Scope: scope.ID, Profile: contracts.Profile{Name: profile}}
			if relation != "" {
				request.Footprint = &contracts.FailureFootprint{Locations: []contracts.FailureLocation{{Resource: use.Resources["List"]}}, Observation: "state_at_completion", Outcomes: []contracts.Outcome{contracts.ReturnedFailure}}
			}
			f.requests.Checks = append(f.requests.Checks, request)
		}
	}
	if len(f.requests.Checks) != 2 {
		t.Fatalf("expected two actual native invocation regions: %+v", f.report.Scopes)
	}
	f.environmentPath = scopedWrite(t, directory, "environment.json", f.environment)
	f.manifestPath = scopedWrite(t, directory, "manifest.json", f.manifest)
	f.trustPath = scopedWrite(t, directory, "trust.json", f.trust)
	f.requestPath = scopedWrite(t, directory, "requests.json", f.requests)
	return f
}

func (f scopedNativeFixture) check(t *testing.T, status int, extra ...string) contracts.Report {
	t.Helper()
	args := []string{"check", "--request-file", f.requestPath, "--native-environment", f.environmentPath, "--manifest", f.manifestPath, "--trust-policy", f.trustPath}
	if f.assets != "" {
		args = append(args, "--assets", f.assets)
	}
	return scopedCLIReport(t, f.input, status, append(args, extra...)...)
}

func TestScopedNativeCLICustomProviderFailureMutation(t *testing.T) {
	assets := t.TempDir()
	if err := os.CopyFS(assets, distribution.Files()); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(assets, "stdlib", "eng", "storage", "dictionary.l")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	text := string(data)
	start := strings.Index(text, "Publicly create Storage's List's \"Update")
	if start < 0 {
		t.Fatal("test provider update body not found")
	}
	end := strings.Index(text[start:], "Publicly create Storage's Map's \"Create") + start
	if end <= start {
		t.Fatal("test provider update boundary not found")
	}
	body := text[start:end]
	needle := "Create \"Found\" as a New Storage's Locator of Values and Position."
	mutation := "Create \"Replacement\" as a New Storage's Value's Copier of Borrowed.\nCreate \"First\" as a New Storage's Locator of Values and `0_usize`.\nCreate \"FirstTarget\" as a New Storage.Checked Binding of a New Storage's Requirer of First and `empty test list`.\nCreate \"FirstPrevious\" as a New Storage's Replacer of FirstTarget and Replacement.\n"
	if strings.Count(body, needle) != 1 {
		t.Fatal("custom fixture mutation point is not unique")
	}
	body = strings.Replace(body, needle, mutation+needle, 1)
	if err := os.WriteFile(path, []byte(text[:start]+body+text[end:]), 0600); err != nil {
		t.Fatal(err)
	}
	composeScopedTestProvider(t, assets)
	f := newScopedNativeFixtureAssets(t, "PreservesOnFailure", "different", assets)
	f.requests.Checks = f.requests.Checks[:1]
	scopedWrite(t, filepath.Dir(f.input), "requests.json", f.requests)
	report := f.check(t, 1)
	if report.Requests[0].Status != contracts.Contradicted {
		t.Fatal("approved custom provider mutation was not contradicted")
	}
	// Restoring the provider invalidates its native witness rather than reusing
	// a relation for a different implementation.
	if err := os.WriteFile(path, data, 0600); err != nil {
		t.Fatal(err)
	}
	composeScopedTestProvider(t, assets)
	report = f.check(t, 1)
	if report.Requests[0].Status != contracts.Unresolved {
		t.Fatal("provider change reused stale approval")
	}
}

// The temporary custom provider has its own genuinely compiled support AML.
// Repository resources, frozen expected bytes and integrity manifests are never
// modified. Merely replacing a source hash would leave stale native behavior.
func composeScopedTestProvider(t *testing.T, root string) {
	t.Helper()
	path := filepath.Join(root, "builtins", "runtime", "eng", "dictionary.l")
	source, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	c := Lyger.NewLygerService(Lyger.Options{Resources: os.DirFS(root), AssetRoot: root}).NewCompilation(source, path)
	c.Make()
	if c.Error != nil {
		t.Fatal(c.Error)
	}
	aml, selection, err := c.ComposeRuntimeLibrary()
	if err != nil {
		t.Fatal(err)
	}
	manifest := distribution.RuntimeComposition{Version: 1, SHA256: nativeBytesHash(aml), Sources: make(map[string]string), Selection: selection}
	for _, origin := range c.SourceOrigins {
		name := origin.Name
		if filepath.IsAbs(name) {
			name, err = filepath.Rel(root, name)
			if err != nil {
				t.Fatal(err)
			}
		}
		manifest.Sources[filepath.ToSlash(name)] = origin.SHA256
	}
	metadata, err := json.Marshal(manifest)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "builtins", "runtime", "support.aml"), aml, 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "builtins", "runtime", "support.json"), metadata, 0600); err != nil {
		t.Fatal(err)
	}
}

func TestScopedNativeCLIApprovalAndRemoval(t *testing.T) {
	f := newScopedNativeFixture(t, "NoBlocking", "")
	report := f.check(t, 0)
	if len(report.Requests) != 2 || len(report.Summary.Unknowns) == 0 || len(report.Assumptions) == 0 {
		t.Fatal("scoped success lost outside unknowns or displayed test-only assumptions")
	}
	for _, request := range report.Requests {
		if request.Scope.Boundary != "native_body" || len(request.Scope.Exclusions) == 0 || request.Status != contracts.Established {
			t.Fatal("narrow boundary was mislabeled")
		}
	}
	f.check(t, 1, "--profile", "NoBlocking")
	f.trust.Witnesses = f.trust.Witnesses[:0]
	f.trust.Policy.ApprovedWitnesses = nil
	scopedWrite(t, filepath.Dir(f.input), "trust.json", f.trust)
	f.check(t, 1)
}

func TestScopedNativeCLIFailureCompletionRelations(t *testing.T) {
	for _, test := range []struct {
		relation string
		status   int
		want     contracts.Status
	}{{"unchanged", 0, contracts.Established}, {"different", 1, contracts.Contradicted}, {"unknown", 1, contracts.Unresolved}} {
		t.Run(test.relation, func(t *testing.T) {
			f := newScopedNativeFixture(t, "PreservesOnFailure", test.relation)
			// A propagated failure ends the first invocation. Select that actual
			// reachable boundary; a later unreachable selector must be rejected.
			f.requests.Checks = f.requests.Checks[:1]
			scopedWrite(t, filepath.Dir(f.input), "requests.json", f.requests)
			report := f.check(t, test.status)
			if report.Requests[0].Status != test.want {
				t.Fatalf("completion relation: %+v", report.Requests[0])
			}
		})
	}
}

func TestScopedNativeCLIStaleEvidenceAndOfflineCommands(t *testing.T) {
	f := newScopedNativeFixture(t, "NoBlocking", "")
	trap := t.TempDir()
	marker := filepath.Join(trap, "executed")
	for _, name := range []string{"cargo", "rustc", "rustup", "git", "curl", "sh"} {
		if err := os.WriteFile(filepath.Join(trap, name), []byte("#!/bin/sh\nprintf executed > '"+marker+"'\nexit 99\n"), 0700); err != nil {
			t.Fatal(err)
		}
	}
	t.Setenv("PATH", trap)
	f.check(t, 0)
	f.environment.Dependencies = map[string]string{"changed": "dependency"}
	f.environment = contracts.SealNativeEnvironment(f.environment)
	scopedWrite(t, filepath.Dir(f.input), "environment.json", f.environment)
	f.check(t, 1)
	if _, err := os.Stat(marker); !os.IsNotExist(err) {
		t.Fatal("offline contract checking executed an external tool")
	}
	var stdout, stderr bytes.Buffer
	if status := run(context.Background(), []string{"contracts", "check", "--profile", "PreservesOnFailure", f.input}, &stdout, &stderr); status != 2 || !strings.Contains(stderr.String(), "footprint") {
		t.Fatalf("missing footprint: %d %s", status, &stderr)
	}
}

func TestScopedNativeCLIMultiOperationInvocation(t *testing.T) {
	directory := t.TempDir()
	input := filepath.Join(directory, "call.l")
	source := "    Publicly create \"Identity\" as a New Function with Inputs of \"Value\" as Number and Outputs of Number.\nReturn Value.\n\nIdentity 7 as \"Result\".\nShow Result.\n"
	if err := os.WriteFile(input, []byte(source), 0600); err != nil {
		t.Fatal(err)
	}
	report := scopedCLIReport(t, input, 0, "inspect")
	var selected contracts.AnalysisScope
	for _, scope := range report.Scopes {
		if scope.Boundary == "source_invocation" {
			if selected.ID != "" {
				t.Fatal("ambiguous fixture invocation")
			}
			selected = scope
		}
	}
	if selected.ID == "" || len(selected.NativeUses) < 2 {
		t.Fatalf("missing composed invocation: %+v", report.Scopes)
	}
	environment := contracts.SealNativeEnvironment(contracts.NativeEnvironment{SchemaVersion: contracts.SchemaVersion, Kind: "native_environment", Producer: "test-only", Version: "1", Method: "test-only", Host: "fixture", Target: "fixture", Compiler: "fixture", PanicPolicy: "fixture", Runtime: "fixture", Allocator: "fixture", Inputs: report.Inputs, Dependencies: report.Lowering.Fingerprints})
	manifest := contracts.NativeContractManifest{SchemaVersion: contracts.SchemaVersion, Kind: "native_contract_manifest", Producer: "test-only"}
	trust := contracts.TrustPolicyDocument{SchemaVersion: contracts.SchemaVersion, Kind: "trust_policy", Producer: "test-only", Policy: contracts.TrustPolicy{Version: contracts.SchemaVersion, ID: "test-only", Authorities: []string{"test-only"}, AllowTestOnly: true}}
	wanted := make(map[string]bool)
	for _, id := range selected.NativeUses {
		wanted[id] = true
	}
	declarations := make(map[contracts.MeaningKey]contracts.DeclaredContract)
	destruction := false
	for _, use := range report.NativeUses {
		if !wanted[use.ID] {
			continue
		}
		destruction = destruction || use.Phase == contracts.Destruction
		closed := false
		declaration := contracts.DeclaredContract{ID: string(use.Meaning), Meaning: use.Meaning, Name: use.Name, Signature: use.Signature, Outcomes: []contracts.Outcome{contracts.Normal}, Effects: &contracts.DeclaredEffectRow{UnknownTail: &closed}}
		if old, exists := declarations[use.Meaning]; exists {
			if contracts.StableID("declaration", old) != contracts.StableID("declaration", declaration) {
				t.Fatalf("same meaning incompatible use signatures %s", use.Name)
			}
		} else {
			declarations[use.Meaning] = declaration
			manifest.Declarations = append(manifest.Declarations, declaration)
		}
		for _, component := range evaluateNativeUse(use, declaration, environment, contracts.TrustPolicy{}, nil, contracts.Unresolved).Components {
			witness := contracts.NativeWitness{ID: "test-only:" + component.Claim.ID, ClaimFingerprint: contracts.ClaimFingerprint(component.Claim), ContractFingerprint: component.Current.Contract, ImplementationFingerprint: component.Current.Implementation, TargetFingerprint: component.Current.Target, Dependencies: component.Current.Dependencies, Authority: "test-only", Rationale: "Synthetic individually approved invocation component; no production approval", TestOnly: true}
			trust.Witnesses = append(trust.Witnesses, witness)
			trust.Policy.ApprovedWitnesses = append(trust.Policy.ApprovedWitnesses, witness.ID)
		}
	}
	if !destruction {
		t.Fatal("invocation omitted owned finalization")
	}
	request := contracts.CheckRequests{SchemaVersion: contracts.SchemaVersion, Kind: "contract_check_requests", Version: "1", Producer: "test-only", Checks: []contracts.CheckRequest{{ID: "composed-invocation", Scope: selected.ID, Profile: contracts.Profile{Name: "NoBlocking"}}}}
	args := []string{"check", "--native-environment", scopedWrite(t, directory, "environment.json", environment), "--manifest", scopedWrite(t, directory, "manifest.json", manifest), "--trust-policy", scopedWrite(t, directory, "trust.json", trust), "--request-file", scopedWrite(t, directory, "request.json", request)}
	checked := scopedCLIReport(t, input, 0, args...)
	if len(checked.Requests[0].Summary.Assumptions) < 2 || len(checked.Summary.Unknowns) == 0 {
		t.Fatal("composition dropped assumptions or outside unknowns")
	}
	// A helper inside the selected invocation loses only its own approval;
	// the composed profile must become unresolved without provider fallback.
	trust.Policy.ApprovedWitnesses = trust.Policy.ApprovedWitnesses[1:]
	trust.Witnesses = trust.Witnesses[1:]
	scopedWrite(t, directory, "trust.json", trust)
	scopedCLIReport(t, input, 1, args...)
}
