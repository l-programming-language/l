package main

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strconv"
	"testing"
	"testing/fstest"

	"github.com/l-programming-language/l/contracts"
)

func contractOutput(t *testing.T, expected int, args ...string) ([]byte, string) {
	t.Helper()
	var stdout, stderr bytes.Buffer
	if status := run(context.Background(), append([]string{"contracts"}, args...), &stdout, &stderr); status != expected {
		t.Fatalf("%v exit %d want %d: %s\n%s", args, status, expected, &stderr, &stdout)
	}
	return stdout.Bytes(), stderr.String()
}

func contractJSON(t *testing.T, input string, expected int, options ...string) (contracts.Report, []byte) {
	t.Helper()
	data, _ := contractOutput(t, expected, append(options, "--format", "json", input)...)
	var report contracts.Report
	if expected != 2 {
		if err := json.Unmarshal(data, &report); err != nil {
			t.Fatalf("decode %s: %v", data, err)
		}
	}
	return report, data
}

func writeContractFile(t *testing.T, path string, data []byte) string {
	t.Helper()
	if err := os.WriteFile(path, data, 0600); err != nil {
		t.Fatal(err)
	}
	return path
}

func contractCLIInput(t *testing.T) string {
	t.Helper()
	source := "Use IO's Print from `io` as \"Print\".\nCreate \"Worker\" as a New Basic App.\nThe App Prints `hello`.\nRun Worker.\nAwait Worker.\nShow Worker's Output.\n"
	return writeContractFile(t, filepath.Join(t.TempDir(), "main.l"), []byte(source))
}

func TestContractsCLI(t *testing.T) {
	input := contractCLIInput(t)
	report, first := contractJSON(t, input, 0, "inspect")
	if len(report.Bindings) == 0 || len(report.Lowering.Native) == 0 || report.ResolutionPolicy != contracts.TrackAPolicy {
		t.Fatal("inspection omitted actual bindings or native closure")
	}
	if len(report.Executions) != 0 {
		t.Fatal("inspection fabricated execution evidence")
	}
	_, repeated := contractJSON(t, input, 0, "inspect")
	if !bytes.Equal(first, repeated) {
		t.Fatal("same-source report is not deterministic")
	}
	entries, err := os.ReadDir(filepath.Dir(input))
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 1 {
		t.Fatal("source inspection created build/publication artifacts")
	}
	t.Run("LEX-039", func(t *testing.T) {
		selected := report.Bindings[len(report.Bindings)-1]
		explained, _ := contractJSON(t, input, 0, "explain", "--binding", string(selected.Occurrence))
		if len(explained.Bindings) != 1 || explained.Bindings[0].Meaning != selected.Meaning || explained.Bindings[0].Target.String() != selected.Target.String() {
			t.Fatal("explanation changed selected meaning or exact target")
		}
	})
}

func TestContractsCLIRejectsInvalidRequests(t *testing.T) {
	input := contractCLIInput(t)
	for index, args := range [][]string{{"check", input}, {"check", "--profile", "Fast", input}, {"inspect", "--format", "yaml", input}, {"explain", input}, {"inspect", input, "--format", "json"}} {
		t.Run(strconv.Itoa(index), func(t *testing.T) {
			contractOutput(t, 2, args...)
		})
	}
}

func TestContractsCLIMetadataCannotSelfCertify(t *testing.T) {
	input := contractCLIInput(t)
	bad := `{"schema_version":"1.1","kind":"native_contract_manifest","producer":"test","declarations":[],"established":true}`
	manifest := writeContractFile(t, filepath.Join(filepath.Dir(input), "native.json"), []byte(bad))
	contractOutput(t, 2, "inspect", "--manifest", manifest, input)
}

func TestContractsDictionaryVerificationIsNotLoading(t *testing.T) {
	report, _ := contractJSON(t, "io", 1, "verify-dictionary")
	if len(report.Definitions) == 0 || len(report.Obligations) == 0 || contracts.AllEstablished(report.Obligations) {
		t.Fatal("dictionary loading was treated as verification")
	}
}

func TestContractsMetadataProviderAndConflicts(t *testing.T) {
	// An explicit empty provider must not inherit embedded declarations.
	metadata, err := loadContractMetadata(fstest.MapFS{}, contractsCommand{})
	if err != nil || len(metadata.Native) != 0 || len(metadata.Trust.Witnesses) != 0 {
		t.Fatalf("optional absent metadata: %#v %v", metadata, err)
	}
	for _, index := range []string{
		`{"schema_version":"1.1","manifests":[],"manifests":[]}`,
		`{"schema_version":"1.1","manifests":[]} {}`,
		`{"schema_version":"1.1","manifests":null}`,
		`{"schema_version":"1.1","Manifests":[]}`,
		`{"schema_version":"1.0","manifests":[]}`,
	} {
		if _, err := loadContractMetadata(fstest.MapFS{"contracts/index.json": &fstest.MapFile{Data: []byte(index)}}, contractsCommand{}); err == nil {
			t.Fatalf("accepted malformed metadata index %s", index)
		}
	}
	declaration := contracts.DeclaredContract{ID: "contract:x", Meaning: "meaning:x", Name: "X"}
	manifest := contracts.NativeContractManifest{SchemaVersion: contracts.SchemaVersion, Kind: "native_contract_manifest", Producer: "test-only", Declarations: []contracts.DeclaredContract{declaration}}
	first, err := contracts.EncodeDocument(manifest)
	if err != nil {
		t.Fatal(err)
	}
	manifest.Declarations[0].Name = "conflict"
	second, err := contracts.EncodeDocument(manifest)
	if err != nil {
		t.Fatal(err)
	}
	for _, names := range []string{`["first.json","second.json"]`, `["second.json","first.json"]`} {
		resources := fstest.MapFS{
			"contracts/index.json": &fstest.MapFile{Data: []byte(`{"schema_version":"1.1","manifests":` + names + `}`)},
			"first.json":           &fstest.MapFile{Data: first},
			"second.json":          &fstest.MapFile{Data: second},
		}
		if _, err := loadContractMetadata(resources, contractsCommand{}); err == nil {
			t.Fatal("manifest order overrode a conflicting declaration")
		}
	}
}
