package Lyger

import (
	"bytes"
	"errors"
	"testing"

	"github.com/l-programming-language/l/lerrors"
)

func TestAppProcessArtifactsIgnoreStructuralServices(t *testing.T) {
	result := runnableServiceRequire(t, "Create \"Host\" as a New Basic App.\nCreate \"Unused\" as a New Main Service.\nCreate \"First\" as a New Basic App.\nCreate \"Library\" as a New Service.\nRun First and Host.\nAwait Host and First.\n", nil)
	apps := []AppArtifact{{ID: "Host", RelativeExecutable: "generation/app-000"}, {ID: "First", RelativeExecutable: "generation/app-001"}}
	before, backend := bytes.Clone(result.AML), bytes.Clone(result.BackendAML)
	unit, err := result.OrchestratorUnit("entry", apps)
	if err != nil {
		t.Fatal(err)
	}
	want := bytes.Clone(unit.BackendAML)
	unit.BackendAML[0] = '!'
	again, err := result.OrchestratorUnit("entry", apps)
	if err != nil || !bytes.Equal(want, again.BackendAML) || !bytes.Equal(result.AML, before) || !bytes.Equal(result.BackendAML, backend) {
		t.Fatalf("repeated App binding mutated source views or shared generated buffers: %v", err)
	}
	if len(result.Compilation.Program.Apps) != 2 {
		t.Fatal("structural Services changed the App graph")
	}
	for _, tc := range []struct {
		name string
		apps []AppArtifact
	}{
		{"missing", apps[:1]},
		{"structural Service", []AppArtifact{{"Unused", "unused"}, apps[1]}},
		{"unknown", []AppArtifact{{"Unknown", "unknown"}, apps[1]}},
		{"duplicate identity", []AppArtifact{apps[0], apps[0]}},
		{"empty path", []AppArtifact{{"Host", ""}, apps[1]}},
		{"absolute path", []AppArtifact{{"Host", "/app"}, apps[1]}},
		{"escape path", []AppArtifact{{"Host", "../app"}, apps[1]}},
		{"dot path", []AppArtifact{{"Host", "."}, apps[1]}},
		{"nul path", []AppArtifact{{"Host", "app\x00"}, apps[1]}},
		{"shared executable", []AppArtifact{apps[0], {"First", apps[0].RelativeExecutable}}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			invalid, err := result.OrchestratorUnit("entry", tc.apps)
			if !errors.Is(err, lerrors.InvalidInput) || invalid.BackendAML != nil {
				t.Fatalf("invalid artifact binding returned source or unclassified error: %v", err)
			}
		})
	}
}
