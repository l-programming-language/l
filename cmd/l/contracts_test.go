package main

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"
	"testing"
	"testing/fstest"

	"github.com/l-programming-language/l/contracts"
	distributiontest "github.com/l-programming-language/l/internal/testsupport/distribution"
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

func readContractFile(t *testing.T, path string) []byte {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return data
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

func TestContractsProtectPhysicalInputs(t *testing.T) {
	root := distributiontest.CompleteOverride(t)
	input := writeContractFile(t, filepath.Join(root, "main.l"), []byte("Use `helper.l`.\nExit.\n"))
	helper := writeContractFile(t, filepath.Join(root, "helper.l"), []byte("Use `empty.l`.\n"))
	empty := writeContractFile(t, filepath.Join(root, "empty.l"), nil)
	hard, sym := filepath.Join(root, "alias"), filepath.Join(root, "link")
	if err := os.Link(empty, hard); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(helper, sym); err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir(filepath.Join(root, "contracts"), 0700); err != nil {
		t.Fatal(err)
	}
	index := filepath.Join(root, "contracts", "index.json")
	manifest := writeContractFile(t, filepath.Join(root, "outside.json"), []byte(`{"schema_version":"1.1","kind":"native_contract_manifest","producer":"test","declarations":[]}`))
	report := filepath.Join(root, "report.json")
	t.Setenv("PATH", t.TempDir()) // A capture reaching tool discovery would fail with status 1.
	reject := func(output string, args ...string) {
		t.Helper()
		before := assetTree(t, root)
		_, diagnostic := contractOutput(t, 2, append(args, "--assets", root, "--output", output, input)...)
		if !strings.Contains(diagnostic, "must not replace") || !reflect.DeepEqual(before, assetTree(t, root)) {
			t.Fatalf("rejection changed inputs or missed collision: %s", diagnostic)
		}
	}
	reject(index, "inspect") // A report must not become the optional provider index.
	writeContractFile(t, index, []byte(`{"schema_version":"1.1","manifests":["outside.json"]}`))
	for _, path := range []string{helper, empty, hard, sym, index, manifest} {
		reject(path, "inspect", "--report", report)
		reject(report, "inspect", "--report", path)
	}
	reject(empty, "evidence", "capture")
	writeContractFile(t, report, []byte("previous report"))
	alias := filepath.Join(root, "report-alias")
	if err := os.Link(report, alias); err != nil {
		t.Fatal(err)
	}
	reject(report, "inspect", "--report", report)
	reject(alias, "inspect", "--report", report)
	before := assetTree(t, root)
	contractOutput(t, 2, "inspect", "--report", report, "--output", root, input)
	if !reflect.DeepEqual(before, assetTree(t, root)) {
		t.Fatal("invalid second destination caused a partial write")
	}
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	relative, err := filepath.Rel(cwd, report)
	if err != nil {
		t.Fatal(err)
	}
	_, want := contractJSON(t, input, 0, "inspect", "--assets", root)
	contractOutput(t, 0, "inspect", "--assets", root, "--output", relative, "--format", "json", input)
	if got, err := os.ReadFile(report); err != nil || !bytes.Equal(got, want) {
		t.Fatalf("relative output differs from stdout: %v", err)
	}
}

func TestContractsInspectRepeatedStop(t *testing.T) {
	source := filepath.Join("..", "..", "testdata", "phase1", "f23-repeat-stop-after-completion", "main.l")
	contractOutput(t, 0, "inspect", source)
	data, _ := contractOutput(t, 0, "inspect", "--format", "json", source)
	if _, err := contracts.DecodeDocument(bytes.Replace(data, []byte(`"kind": "signal"`), []byte(`"kind": "signla"`), 1)); err == nil {
		t.Fatal("strict reader accepted an unknown effect kind")
	}
	document, err := contracts.DecodeDocument(data)
	if err != nil {
		t.Fatal(err)
	}
	report, ok := document.(contracts.Report)
	if !ok {
		t.Fatalf("expected a contract report, got %T", document)
	}
	var events []contracts.Effect
	for _, effect := range report.Summary.Effects {
		if effect.Origin.Source == "source:main.l" && effect.Resource.Kind == "process" {
			events = append(events, effect)
		}
	}
	if len(events) != 4 || events[0].Origin.Start >= events[1].Origin.Start || events[0].Target.Root == "" {
		t.Fatalf("missing distinct authored Stop occurrences and process target: %+v", events)
	}
	for i, kind := range []contracts.EffectKind{"signal", "signal", "wait", "suspend"} {
		e := events[i]
		if e.Kind != kind || e.Target.Root != events[0].Target.Root || e.Resource.ID != events[0].Resource.ID {
			t.Fatalf("expected %s on the selected process, got %+v", kind, e)
		}
	}
}
