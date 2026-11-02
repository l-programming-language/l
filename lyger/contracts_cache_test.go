package Lyger

import (
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
