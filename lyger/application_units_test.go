package Lyger

import (
	"bytes"
	"strings"
	"testing"
)

func executableTestResult(t *testing.T, source memberResolutionSource) *CompilationResult {
	t.Helper()
	comp := compileMemberResolutionSource(t, source)
	if comp.Error != nil {
		t.Fatal(comp.Error)
	}
	return &CompilationResult{Compilation: comp, ExpandedSource: bytes.Clone(comp.ExpandedBytes), AML: bytes.Clone(comp.ASMBytes), BackendAML: bytes.Clone(comp.Program.BackendBytes)}
}

func executableTestCompanions(result *CompilationResult) []AppArtifact {
	var resultApps []AppArtifact
	for _, app := range result.Compilation.Program.Apps {
		resultApps = append(resultApps, AppArtifact{ID: app.buildName(), RelativeExecutable: "main.apps/" + app.buildName()})
	}
	return resultApps
}

func TestApplicationUnitsPreserveSnapshotsAndBookOrder(t *testing.T) {
	source := memberResolutionFixture(t)
	source.main += "\n\nCreate \"SecondWorker\" as a New Basic App.\nThe App Prints x.\n"
	result := executableTestResult(t, source)
	beforeAML, beforeBackend := bytes.Clone(result.AML), bytes.Clone(result.BackendAML)
	program := result.Compilation.Program
	units, err := result.ApplicationUnits()
	if err != nil {
		t.Fatal(err)
	}
	if len(units) != 2 || units[0].ID != "HelloWorldApp" || units[1].ID != "SecondWorker" {
		t.Fatalf("application unit order = %+v", units)
	}
	for _, unit := range units {
		if bytes.Count(unit.BackendAML, []byte("<BOOK ")) != 1 || !bytes.Contains(unit.BackendAML, []byte(`<BOOK "`+unit.ID+`">`)) || !bytes.Contains(unit.BackendAML, []byte("<|S|>`")) {
			t.Fatalf("unit %s is not a complete canonical single-book document", unit.ID)
		}
	}
	units[0].BackendAML[0] = '!'
	again, err := result.ApplicationUnits()
	if err != nil || again[0].BackendAML[0] != '<' || units[1].BackendAML[0] != '<' {
		t.Fatalf("application units share output buffers: %v", err)
	}
	if result.Compilation.Program != program || !bytes.Equal(result.AML, beforeAML) || !bytes.Equal(result.BackendAML, beforeBackend) || !bytes.Equal(program.BackendBytes, beforeBackend) {
		t.Fatal("build-unit generation changed the original compilation views or graph")
	}
}

func TestApplicationUnitsSingleAppMatchesExistingBackendDocument(t *testing.T) {
	result := executableTestResult(t, memberResolutionFixture(t))
	units, err := result.ApplicationUnits()
	if err != nil || len(units) != 1 || !bytes.Equal(units[0].BackendAML, result.BackendAML) {
		t.Fatalf("single-App source generation changed: %v", err)
	}
}

func TestBuildUnitsRejectIncompleteGraphsAndArtifactMappings(t *testing.T) {
	var absent *CompilationResult
	if _, err := absent.ApplicationUnits(); err == nil {
		t.Fatal("nil compilation accepted")
	}
	if _, err := (&CompilationResult{}).ApplicationUnits(); err == nil {
		t.Fatal("empty compilation accepted")
	}
	result := executableTestResult(t, memberResolutionFixture(t))
	for _, tc := range []struct {
		name      string
		id        string
		artifacts []AppArtifact
	}{
		{name: "missing", id: "main"},
		{name: "unknown", id: "main", artifacts: []AppArtifact{{ID: "Other", RelativeExecutable: "main.apps/Other"}}},
		{name: "duplicate", id: "main", artifacts: []AppArtifact{{ID: "HelloWorldApp", RelativeExecutable: "main.apps/one"}, {ID: "HelloWorldApp", RelativeExecutable: "main.apps/two"}}},
		{name: "absolute", id: "main", artifacts: []AppArtifact{{ID: "HelloWorldApp", RelativeExecutable: "/tmp/Worker"}}},
		{name: "escape", id: "main", artifacts: []AppArtifact{{ID: "HelloWorldApp", RelativeExecutable: "../Worker"}}},
		{name: "empty path", id: "main", artifacts: []AppArtifact{{ID: "HelloWorldApp"}}},
		{name: "empty id", artifacts: []AppArtifact{{ID: "HelloWorldApp", RelativeExecutable: "main.apps/Worker"}}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if _, err := result.OrchestratorUnit(tc.id, tc.artifacts); err == nil {
				t.Fatal("invalid build artifact mapping accepted")
			}
		})
	}
	result.Compilation.Program.Books[0].Name = "Missing"
	if _, err := result.ApplicationUnits(); err == nil || !strings.Contains(err.Error(), "application") {
		t.Fatalf("mismatched graph accepted: %v", err)
	}
}
