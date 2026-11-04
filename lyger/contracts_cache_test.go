package Lyger

import (
	"bytes"
	"reflect"
	"testing"

	"github.com/l-programming-language/l/contracts"
)

func TestContractsAnalysisSessionWarmCold(t *testing.T) {
	session := contracts.NewSession()
	service := NewLygerService(Options{})
	options := ContractAnalysisOptions{Session: session}
	source := []byte("Create \"Worker\" as a New Basic App.\nRun Worker.\nAwait Worker.\n")
	first, err := service.AnalyzeContracts(source, "/project/session.l", options)
	if err != nil {
		t.Fatal(err)
	}
	warm, err := service.AnalyzeContracts(source, "/project/session.l", options)
	if err != nil {
		t.Fatal(err)
	}
	cold, err := service.AnalyzeContracts(source, "/project/session.l", ContractAnalysisOptions{})
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(first.Bindings, warm.Bindings) || !reflect.DeepEqual(warm.Bindings, cold.Bindings) || contracts.StableID("summary", first.Summary) != contracts.StableID("summary", warm.Summary) || contracts.StableID("summary", warm.Summary) != contracts.StableID("summary", cold.Summary) {
		t.Fatalf("cached source analysis differs from cold Track A resolution/inference: bindings first/warm=%v warm/cold=%v; summaries=%s %s %s", reflect.DeepEqual(first.Bindings, warm.Bindings), reflect.DeepEqual(warm.Bindings, cold.Bindings), contracts.StableID("summary", first.Summary), contracts.StableID("summary", warm.Summary), contracts.StableID("summary", cold.Summary))
	}
	stats := session.CacheStats()
	if stats.Binding.Hits == 0 || stats.Logical.Hits != 1 || stats.Implementation.Hits != 1 {
		t.Fatalf("independent source caches not exercised: %+v", stats)
	}
	// A native body change invalidates current assurance while the unchanged
	// typed logical summary (with opaque obligations) remains reusable.
	warm.Lowering.Fingerprints["native-body"] = "changed"
	if err := warm.ReinferContracts(options); err != nil {
		t.Fatal(err)
	}
	updated := session.CacheStats()
	if updated.Logical.Hits != stats.Logical.Hits+1 || updated.Implementation.Misses != stats.Implementation.Misses+1 {
		t.Fatalf("native edit conflated logical and implementation caches: before=%+v after=%+v", stats, updated)
	}
	// New native closure changes logical unknowns and must agree with cold.
	warm.Plan.Residuals = append(warm.Plan.Residuals, contracts.UnknownReason{Code: "new-cleanup", Message: "new cleanup dependency", Components: []string{"effects"}})
	if err := warm.ReinferContracts(options); err != nil {
		t.Fatal(err)
	}
	expected, err := contracts.Infer(warm.Plan, options.Inference)
	if err != nil || contracts.StableID("summary", warm.Summary) != contracts.StableID("summary", expected) {
		t.Fatalf("closure changed cached result: %v", err)
	}
}

func TestContractsResourceLifecycleReportRoundTrip(t *testing.T) {
	analysis, err := NewLygerService(Options{}).AnalyzeContracts([]byte(contractNativeUpdateSource), "/project/lifecycle-report.l", ContractAnalysisOptions{})
	if err != nil {
		t.Fatal(err)
	}
	owned := 0
	for _, path := range analysis.Summary.Paths {
		for _, resource := range path.Owned {
			owned++
			if resource.FinalOwner == "" {
				t.Fatal("owned resource omitted normalized final-owner uncertainty")
			}
		}
	}
	if owned == 0 {
		t.Fatal("source fixture did not exercise owned runtime resources")
	}
	report := contracts.NewReport("/project/lifecycle-report.l")
	report.Summary = analysis.Summary
	encoded, err := contracts.EncodeDocument(report)
	if err != nil {
		t.Fatal(err)
	}
	decoded, err := contracts.DecodeDocument(encoded)
	if err != nil {
		t.Fatal(err)
	}
	again, err := contracts.EncodeDocument(decoded)
	if err != nil || !bytes.Equal(encoded, again) {
		t.Fatalf("resource lifecycle report lost normalized state: %v", err)
	}
}

func TestContractsAnalysisSessionSourceChange(t *testing.T) {
	session := contracts.NewSession()
	service := NewLygerService(Options{})
	initial := []byte("Create \"Worker\" as a New Basic App.\nRun Worker.\n")
	if _, err := service.AnalyzeContracts(initial, "/project/session.l", ContractAnalysisOptions{Session: session}); err != nil {
		t.Fatal(err)
	}
	changed := []byte("Use IO's Print from `io` as \"Print\".\nCreate \"Other\" as a New Basic App.\nRun Other.\nAwait Other.\n")
	warm, err := service.AnalyzeContracts(changed, "/project/session.l", ContractAnalysisOptions{Session: session})
	if err != nil {
		t.Fatal(err)
	}
	cold, err := service.AnalyzeContracts(changed, "/project/session.l", ContractAnalysisOptions{})
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(warm.Bindings, cold.Bindings) || contracts.StableID("summary", warm.Summary) != contracts.StableID("summary", cold.Summary) {
		t.Fatal("new import/target reused old bindings or assurance")
	}
}

// Exercise real source scopes while keeping approval conspicuously test-only.
// The reusable logical result remains independent of native implementation trust.
func TestContractsNativeApplicationsSessionWarmCold(t *testing.T) {
	session := contracts.NewSession()
	service := NewLygerService(Options{})
	source := []byte(contractNativeUpdateSource)
	analysis, err := service.AnalyzeContracts(source, "/project/native-cache.l", ContractAnalysisOptions{Session: session})
	if err != nil {
		t.Fatal(err)
	}
	var uses []contracts.NativeUse
	for _, use := range analysis.NativeUses {
		if use.Name == "Storage.List.Update" {
			uses = append(uses, use)
		}
	}
	if len(uses) != 2 {
		t.Fatalf("selected update uses: %+v", uses)
	}
	use := uses[0]
	declaration := contracts.DeclaredContract{ID: "test-only:cache-native", Meaning: use.Meaning, Name: use.Name, Signature: use.Signature, Outcomes: []contracts.Outcome{contracts.Normal}, OutcomeContracts: []contracts.NativeOutcomeContract{{Outcome: contracts.Normal, AbsentEffects: []contracts.EffectKind{"block"}}}}
	environment := contracts.SealNativeEnvironment(contracts.NativeEnvironment{SchemaVersion: contracts.SchemaVersion, Kind: "native_environment", Producer: contracts.AnalyzerVersion, Version: "1", Method: "test-only", Host: "test-only", Target: "test-only", Compiler: "test-only", PanicPolicy: "test-only", Runtime: "test-only", Allocator: "test-only", Inputs: map[string]string{"source": "v1"}})
	policy := contracts.TrustPolicy{Version: contracts.SchemaVersion, ID: "test-only:cache-policy", Authorities: []string{"test-only:authority"}, AllowTestOnly: true}
	var witnesses []contracts.NativeWitness
	for _, component := range contracts.NativeComponentClaims(use, declaration, environment) {
		witness := contracts.NativeWitness{ID: "test-only:" + component.Claim.ID, ClaimFingerprint: contracts.ClaimFingerprint(component.Claim), ContractFingerprint: component.Current.Contract, ImplementationFingerprint: component.Current.Implementation, TargetFingerprint: component.Current.Target, Dependencies: component.Current.Dependencies, Authority: policy.Authorities[0], Rationale: "Synthetic cache regression assumption; no production approval", TestOnly: true}
		witnesses = append(witnesses, witness)
		policy.ApprovedWitnesses = append(policy.ApprovedWitnesses, witness.ID)
	}
	for _, change := range []string{"baseline", "environment", "witness", "trust", "budget", "restored"} {
		t.Run(change, func(t *testing.T) {
			currentEnvironment, currentPolicy := environment, policy
			currentWitnesses := append([]contracts.NativeWitness(nil), witnesses...)
			inference := contracts.AnalysisOptions{}
			switch change {
			case "environment":
				currentEnvironment.PanicPolicy = "changed"
				currentEnvironment = contracts.SealNativeEnvironment(currentEnvironment)
			case "witness":
				currentWitnesses[0].ImplementationFingerprint = "stale"
			case "trust":
				currentPolicy.ApprovedWitnesses = nil
			case "budget":
				inference.MaxAlternatives = 1
			}
			application, _ := contracts.ApproveNativeApplication(use, declaration, currentEnvironment, currentPolicy, currentWitnesses, contracts.Established)
			inference.NativeApplications = map[string]contracts.ApprovedNativeApplication{use.ID: application}
			options := ContractAnalysisOptions{Session: session, Inference: inference}
			warm, err := service.AnalyzeContracts(source, "/project/native-cache.l", options)
			if err != nil {
				t.Fatal(err)
			}
			cold, err := service.AnalyzeContracts(source, "/project/native-cache.l", ContractAnalysisOptions{Inference: inference})
			if err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(warm.Bindings, cold.Bindings) || contracts.StableID("summary", warm.Summary) != contracts.StableID("summary", cold.Summary) {
				t.Fatal("cached native application differs from cold analysis")
			}
			if err := warm.ReinferContracts(options); err != nil {
				t.Fatal(err)
			}
			if contracts.StableID("summary", warm.Summary) != contracts.StableID("summary", cold.Summary) {
				t.Fatal("warm assurance reload changed its result")
			}
			for i, selected := range uses {
				warmScope, err := contracts.InferScope(warm.Plan, selected.ID, inference)
				if err != nil {
					t.Fatal(err)
				}
				coldScope, err := contracts.InferScope(cold.Plan, selected.ID, inference)
				if err != nil {
					t.Fatal(err)
				}
				if contracts.StableID("scope", warmScope) != contracts.StableID("scope", coldScope) {
					t.Fatal("warm/cold selected scope differs")
				}
				want := i == 0 && (change == "baseline" || change == "budget" || change == "restored")
				if contracts.AllEstablished(contracts.CheckProfile(warmScope, contracts.Profile{Name: "NoBlocking"})) != want {
					t.Fatalf("change=%s scope=%d retained stale or cross-scope native approval: %+v", change, i, warmScope)
				}
			}
			if contracts.AllEstablished(contracts.CheckProfile(warm.Summary, contracts.Profile{Name: "NoBlocking"})) {
				t.Fatal("selected leaf approval closed the whole source/native closure")
			}
		})
	}
	stats := session.CacheStats()
	if stats.Logical.Hits == 0 || stats.Implementation.Hits == 0 || stats.Implementation.Misses < 5 {
		t.Fatalf("cache boundaries were not exercised: %+v", stats)
	}
}
