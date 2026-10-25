package Lyger

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"reflect"
	"strings"
	"testing"

	"github.com/l-programming-language/l/lerrors"
)

func workspaceDependencyResult(t *testing.T) *CompilationResult {
	t.Helper()
	source, err := os.ReadFile("test/workspace/services.l")
	if err != nil {
		t.Fatal(err)
	}
	cfg := canonicalTestOptions()
	result, err := NewLygerService(cfg).Compile(source, "test/workspace/services.l")
	if err != nil {
		t.Fatal(err)
	}
	return result
}

func workspaceDependencyLocations(plan *WorkspacePlan) map[string]string {
	locations := make(map[string]string)
	for index, item := range plan.Packages() {
		locations[item.ID] = fmt.Sprintf("packages/package-%d", index)
	}
	return locations
}

func workspaceDependencyTestRows(t *testing.T, unit WorkspaceUnit) []string {
	t.Helper()
	_, after, ok := bytes.Cut(unit.BackendAML, []byte("<JSON>"))
	if !ok {
		t.Fatal("workspace unit omitted application metadata")
	}
	raw, _, ok := bytes.Cut(after, []byte("</JSON>"))
	if !ok {
		t.Fatal("workspace unit has unterminated application metadata")
	}
	var metadata struct{ AppMeta amlAppMetadata }
	if err := json.Unmarshal(raw, &metadata); err != nil {
		t.Fatal(err)
	}
	var rows []string
	if metadata.AppMeta.Deps != "" {
		if err := json.Unmarshal([]byte(metadata.AppMeta.Deps), &rows); err != nil {
			t.Fatal(err)
		}
	}
	return rows
}

func TestWorkspacePreservesSuppliedDependencies(t *testing.T) {
	result := workspaceDependencyResult(t)
	app := result.Compilation.Program.Apps[0]
	rows := []string{
		`tokio = { version = "=1.53.1", features = ["rt", "sync"] }`,
		`nix = { version = "=0.31.3", features = ["signal", "process"] }`,
	}
	encoded, err := json.Marshal(rows)
	if err != nil {
		t.Fatal(err)
	}
	app.Meta["Deps"] = string(encoded)
	beforeMeta := cloneWorkspaceMetadata(app.Meta)
	beforeAML, beforeBackend := bytes.Clone(result.AML), bytes.Clone(result.BackendAML)
	plan, err := result.WorkspacePlan()
	if err != nil {
		t.Fatal(err)
	}
	locations := workspaceDependencyLocations(plan)
	units, err := plan.Units(locations)
	if err != nil {
		t.Fatal(err)
	}
	want := append(append([]string(nil), rows...),
		`left = { package = "l_service_4669727374_1", path = "../package-1" }`,
		`right = { package = "l_service_4669727374_2", path = "../package-2" }`)
	if got := workspaceDependencyTestRows(t, units[0]); !reflect.DeepEqual(got, want) {
		t.Fatalf("workspace dependency rows changed order or bytes:\n got %#v\nwant %#v", got, want)
	}
	for _, unit := range units[1:3] {
		if got := workspaceDependencyTestRows(t, unit); !reflect.DeepEqual(got, rows) {
			t.Fatalf("library %s changed supplied dependencies: %#v", unit.ID, got)
		}
	}
	again, err := plan.Units(locations)
	if err != nil || !reflect.DeepEqual(units, again) || !reflect.DeepEqual(app.Meta, beforeMeta) || !bytes.Equal(result.AML, beforeAML) || !bytes.Equal(result.BackendAML, beforeBackend) {
		t.Fatalf("workspace dependency projection mutated caller state or repeated output: %v", err)
	}
	// A plan owns its metadata snapshot. Neither a later caller mutation nor an
	// earlier returned byte slice changes that snapshot.
	app.Meta["Deps"] = `[]`
	units[0].BackendAML[0] = '!'
	afterMutation, err := plan.Units(locations)
	if err != nil || !reflect.DeepEqual(afterMutation, again) {
		t.Fatalf("workspace plan retained caller-owned dependency state: %v", err)
	}
}

func TestWorkspaceRejectsMalformedDependencies(t *testing.T) {
	result := workspaceDependencyResult(t)
	app := result.Compilation.Program.Apps[0]
	for _, tc := range []struct{ name, raw, cause string }{
		{"empty", "", "JSON-encoded array"},
		{"framing", "[", "JSON-encoded array"},
		{"null", "null", "not null"},
		{"scalar", `"tokio = 1"`, "JSON-encoded array"},
		{"non-string row", `[3]`, "JSON-encoded array"},
		{"missing equals", `["tokio"]`, "normalized TOML dependency assignment"},
		{"missing key", `["= 1"]`, "normalized TOML dependency assignment"},
		{"missing value", `["tokio = "]`, "normalized TOML dependency assignment"},
		{"outer whitespace", `[" tokio = 1"]`, "normalized TOML dependency assignment"},
		{"newline", `["tokio = 1\nother = 2"]`, "normalized TOML dependency assignment"},
		{"non-bare key", `["a.b = 1"]`, "bare dependency key"},
		{"duplicate", `["tokio = 1", "tokio = 2"]`, "duplicate dependency"},
		{"Service alias conflict", `["left = { version = \"1\" }"]`, "collides with constructed Service"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			app.Meta["Deps"] = tc.raw
			before := cloneWorkspaceMetadata(app.Meta)
			plan, err := result.WorkspacePlan()
			if err != nil {
				t.Fatal(err)
			}
			units, err := plan.Units(workspaceDependencyLocations(plan))
			if !errors.Is(err, lerrors.InvalidInput) || !strings.Contains(err.Error(), tc.cause) || units != nil {
				t.Fatalf("invalid dependency metadata returned %d units, %v; require %s", len(units), err, tc.cause)
			}
			if !reflect.DeepEqual(app.Meta, before) {
				t.Fatal("rejected dependency metadata changed caller state")
			}
		})
	}
}
