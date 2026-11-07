package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	distribution "github.com/l-programming-language/l"
	"github.com/l-programming-language/l/contracts"
)

func TestLayoutNativeEvidenceRejectsChangedAssets(t *testing.T) {
	assets := t.TempDir()
	if err := os.CopyFS(assets, distribution.Files()); err != nil {
		t.Fatal(err)
	}
	input := writeContractFile(t, filepath.Join(t.TempDir(), "main.l"), []byte("Show `selected provider`.\n"))
	inspect := func() contracts.Report {
		report, _ := contractJSON(t, input, 0, "inspect", "--assets", assets)
		if len(report.Executions) != 0 {
			t.Fatal("inspection invented native execution evidence")
		}
		return report
	}
	initial := inspect()
	// This record tests applicability only. It grants no implementation approval
	// and claims no observed build, native execution, or production authority.
	environment := contracts.SealNativeEnvironment(contracts.NativeEnvironment{
		SchemaVersion: contracts.SchemaVersion, Kind: "native_environment", Producer: "test-only-layout-evidence",
		Version: "1", Method: "test-only", Host: "fixture-host", Target: "fixture-target", Compiler: "fixture-compiler",
		PanicPolicy: "fixture-policy", Runtime: "fixture-runtime", Allocator: "fixture-allocator",
		Inputs: initial.Inputs, Dependencies: initial.Lowering.Fingerprints,
		Limitations: []string{"Synthetic applicability control; no production approval or executed build"},
	})
	assertApplicability := func(want contracts.Status) {
		t.Helper()
		current := inspect()
		if got, reason := contracts.EnvironmentApplicability(environment, current.Inputs, current.Lowering.Fingerprints); got != want {
			t.Fatalf("selected-provider evidence applicability = %s, want %s: %s", got, want, reason)
		}
	}
	assertApplicability(contracts.Established)
	host := filepath.Join(assets, "runtime", "rust", "host.rs")
	native := readContractFile(t, host)
	writeContractFile(t, host, append(append([]byte(nil), native...), []byte("\n// changed selected native input\n")...))
	assertApplicability(contracts.Unresolved)
	writeContractFile(t, host, native)
	assertApplicability(contracts.Established)

	// Move a real included source inside its owning dictionary root. Regenerate
	// from that provider, preserving its words while changing source provenance.
	storage := filepath.Join(assets, "builtins", "eng", "storage")
	entry := filepath.Join(storage, "dictionary.l")
	source := string(readContractFile(t, entry))
	if strings.Count(source, "Use `bindings.l`.") != 1 {
		t.Fatal("canonical Storage bindings include is absent or ambiguous")
	}
	if err := os.Rename(filepath.Join(storage, "bindings.l"), filepath.Join(storage, "relocated-bindings.l")); err != nil {
		t.Fatal(err)
	}
	writeContractFile(t, entry, []byte(strings.Replace(source, "Use `bindings.l`.", "Use `relocated-bindings.l`.", 1)))
	composeScopedTestProvider(t, assets)
	assertApplicability(contracts.Unresolved)
}
