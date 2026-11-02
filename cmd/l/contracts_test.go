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

func contractCLIInput(t *testing.T) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "main.l")
	source := "Use IO's Print from `io` as \"Print\".\nCreate \"Worker\" as a New Basic App.\nThe App Prints `hello`.\nRun Worker.\nAwait Worker.\nShow Worker's Output.\n"
	if err := os.WriteFile(path, []byte(source), 0600); err != nil {
		t.Fatal(err)
	}
	return path
}

func TestContractsCLI(t *testing.T) {
	input := contractCLIInput(t)
	var stdout, stderr bytes.Buffer
	status := run(context.Background(), []string{"contracts", "inspect", "--format", "json", input}, &stdout, &stderr)
	if status != 0 {
		t.Fatalf("inspect exit %d: %s", status, stderr.String())
	}
	var report contracts.Report
	if err := json.Unmarshal(stdout.Bytes(), &report); err != nil {
		t.Fatal(err)
	}
	if len(report.Bindings) == 0 || len(report.Lowering.Native) == 0 || report.ResolutionPolicy != contracts.TrackAPolicy {
		t.Fatal("inspection omitted actual bindings or native closure")
	}
	if len(report.Executions) != 0 {
		t.Fatal("inspection fabricated execution evidence")
	}
	var repeated, repeatedError bytes.Buffer
	if status := run(context.Background(), []string{"contracts", "inspect", "--format", "json", input}, &repeated, &repeatedError); status != 0 {
		t.Fatalf("repeated inspection failed: %s", repeatedError.String())
	}
	if !bytes.Equal(stdout.Bytes(), repeated.Bytes()) {
		t.Fatal("same-source report is not deterministic")
	}
	entries, err := os.ReadDir(filepath.Dir(input))
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 1 {
		t.Fatal("source inspection created build/publication artifacts")
	}
	t.Run("BASE-028", func(t *testing.T) {
		stdout.Reset()
		stderr.Reset()
		status := run(context.Background(), []string{"contracts", "check", "--profile", "NoBlocking", "--format", "json", input}, &stdout, &stderr)
		if status != 1 {
			t.Fatalf("unresolved enforcement exit %d: %s", status, stderr.String())
		}
		var checked contracts.Report
		if err := json.Unmarshal(stdout.Bytes(), &checked); err != nil {
			t.Fatal(err)
		}
		if len(checked.Profiles) != 1 || checked.Profiles[0].Status != contracts.Unresolved {
			t.Fatal("unknown native behavior did not stay unresolved")
		}
	})
	t.Run("LEX-039", func(t *testing.T) {
		selected := report.Bindings[len(report.Bindings)-1]
		stdout.Reset()
		stderr.Reset()
		status := run(context.Background(), []string{"contracts", "explain", "--binding", string(selected.Occurrence), "--format", "json", input}, &stdout, &stderr)
		if status != 0 {
			t.Fatalf("explain exit %d: %s", status, stderr.String())
		}
		var explained contracts.Report
		if err := json.Unmarshal(stdout.Bytes(), &explained); err != nil {
			t.Fatal(err)
		}
		if len(explained.Bindings) != 1 || explained.Bindings[0].Meaning != selected.Meaning || explained.Bindings[0].Target.String() != selected.Target.String() {
			t.Fatal("explanation changed selected meaning or exact target")
		}
	})
	t.Run("LEX-040", func(t *testing.T) {
		if report.ResolutionPolicy != contracts.TrackAPolicy {
			t.Fatal("contracts changed default resolver")
		}
	})
}

func TestContractsCLIRejectsInvalidRequests(t *testing.T) {
	input := contractCLIInput(t)
	for index, args := range [][]string{{"check", input}, {"check", "--profile", "Fast", input}, {"inspect", "--format", "yaml", input}, {"explain", input}, {"inspect", input, "--format", "json"}} {
		t.Run(strconv.Itoa(index), func(t *testing.T) {
			var stdout, stderr bytes.Buffer
			if status := run(context.Background(), append([]string{"contracts"}, args...), &stdout, &stderr); status != 2 {
				t.Fatalf("invalid command exit %d: %s", status, stderr.String())
			}
		})
	}
}

func TestContractsCLIMetadataCannotSelfCertify(t *testing.T) {
	input := contractCLIInput(t)
	manifest := filepath.Join(filepath.Dir(input), "native.json")
	bad := `{"schema_version":"1.1","kind":"native_contract_manifest","producer":"test","declarations":[],"established":true}`
	if err := os.WriteFile(manifest, []byte(bad), 0600); err != nil {
		t.Fatal(err)
	}
	var stdout, stderr bytes.Buffer
	if status := run(context.Background(), []string{"contracts", "inspect", "--manifest", manifest, input}, &stdout, &stderr); status != 2 {
		t.Fatalf("authored evidence accepted: %d %s", status, stderr.String())
	}
}

func TestContractsDictionaryVerificationIsNotLoading(t *testing.T) {
	var stdout, stderr bytes.Buffer
	status := run(context.Background(), []string{"contracts", "verify-dictionary", "--format", "json", "io"}, &stdout, &stderr)
	if status != 1 {
		t.Fatalf("unverified dictionary exit %d: %s", status, stderr.String())
	}
	var report contracts.Report
	if err := json.Unmarshal(stdout.Bytes(), &report); err != nil {
		t.Fatal(err)
	}
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
