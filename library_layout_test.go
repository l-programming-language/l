package distribution_test

import (
	"encoding/json"
	"errors"
	"io/fs"
	"path/filepath"
	"strings"
	"testing"
	"testing/fstest"

	distribution "github.com/l-programming-language/l"
	distributiontest "github.com/l-programming-language/l/internal/testsupport/distribution"
	Logiceme "github.com/l-programming-language/l/logiceme"
	Lyger "github.com/l-programming-language/l/lyger"
)

func TestCoreResourceLayout(t *testing.T) {
	for _, name := range []string{"dictionaries.json", "builtins/eng/core.l"} {
		if _, err := fs.ReadFile(distribution.Files(), name); err != nil {
			t.Errorf("required resource %s: %v", name, err)
		}
	}
	for _, name := range []string{"builtins/eng/prelude.l", "builtins/eng/storage", "builtins/operations", "stdlib", "builtins/runtime", "builtins/templates/runtime", "runtime"} {
		if _, err := fs.Stat(distribution.Files(), name); !errors.Is(err, fs.ErrNotExist) {
			t.Errorf("retired resource %s remains available: %v", name, err)
		}
	}
	data, err := fs.ReadFile(distribution.Files(), "dictionaries.json")
	if err != nil {
		t.Fatal(err)
	}
	var registry struct {
		Languages map[string]map[string]json.RawMessage `json:"languages"`
	}
	if err := json.Unmarshal(data, &registry); err != nil {
		t.Fatal(err)
	}
	for language, entries := range registry.Languages {
		if len(entries) != 0 {
			t.Errorf("bundled dictionaries remain for %s: %v", language, entries)
		}
	}
}

func TestCoreCompilerWithoutProcessOperationAssets(t *testing.T) {
	files := distributiontest.Files(t)
	for name := range files {
		if strings.HasPrefix(name, "builtins/operations/") {
			delete(files, name)
		}
	}
	result, err := Lyger.NewLygerService(Lyger.Options{Resources: files}).Compile([]byte("Create \"Worker\" as a New Basic App.\nThe App Prints `independent`.\n"), filepath.Join(t.TempDir(), "main.l"))
	if err != nil {
		t.Fatal(err)
	}
	units, err := result.ApplicationUnits()
	if err != nil || len(units) != 1 {
		t.Fatalf("standalone units = %v, %v", units, err)
	}
	bundle, err := Logiceme.NewLogicemeService(files).Evaluate(units[0].BackendAML)
	if err != nil || len(bundle.ByteArrays) != 1 || len(bundle.ByteArrays[0]) == 0 {
		t.Fatalf("standalone source generation = %v", err)
	}
}

func TestSelectedRegistryHasNoFallback(t *testing.T) {
	files := distributiontest.Files(t)
	files["dictionaries.json"] = &fstest.MapFile{Data: []byte(`{"format":1,"languages":{"eng":{"probe":{"entry":"stdlib/eng/probe.l","dictionary":"Probe"}}}}`)}
	files["stdlib/eng/probe.l"] = &fstest.MapFile{Data: []byte("Create \"Probe\" as a Dictionary.\n")}
	compile := func() error {
		_, err := Lyger.NewLygerService(Lyger.Options{Resources: files}).Compile([]byte("Use Probe from `probe`.\n"), filepath.Join(t.TempDir(), "main.l"))
		return err
	}
	if err := compile(); err != nil {
		t.Fatalf("complete selected provider: %v", err)
	}
	files["stdlib/index.json"] = &fstest.MapFile{Data: append([]byte(nil), files["dictionaries.json"].Data...)}
	delete(files, "dictionaries.json")
	if err := compile(); !errors.Is(err, fs.ErrNotExist) || !strings.Contains(err.Error(), "dictionaries.json") {
		t.Fatalf("missing selected registry must reject despite old registry: %v", err)
	}
}

func TestCoreSourceUnitOwnership(t *testing.T) {
	result, err := Lyger.NewLygerService(Lyger.Options{}).Compile([]byte("Create \"Worker\" as a New Basic App.\nThe App Prints `owned`.\n"), filepath.Join(t.TempDir(), "main.l"))
	if err != nil {
		t.Fatal(err)
	}
	builtins := 0
	for _, unit := range result.Compilation.SourceUnits {
		if unit.Kind == Lyger.SourceBuiltin {
			builtins++
			if unit.Origin.Name != "builtins/eng/core.l" {
				t.Fatalf("builtin origin: %s", unit.Origin.Name)
			}
		}
		if unit.Kind == Lyger.SourceDictionary {
			t.Fatalf("core loaded an implicit dictionary: %s", unit.Dictionary)
		}
	}
	if builtins != 1 {
		t.Fatalf("builtin units=%d, want one", builtins)
	}
}

func TestSelectedCoreSourceHasNoFallback(t *testing.T) {
	files := distributiontest.Files(t)
	compile := func() error {
		_, err := Lyger.NewLygerService(Lyger.Options{Resources: files}).Compile([]byte("Create \"Worker\" as a New Basic App.\n"), filepath.Join(t.TempDir(), "main.l"))
		return err
	}
	if err := compile(); err != nil {
		t.Fatalf("complete selected provider: %v", err)
	}
	files["builtins/eng/prelude.l"] = &fstest.MapFile{Data: append([]byte(nil), files["builtins/eng/core.l"].Data...)}
	delete(files, "builtins/eng/core.l")
	if err := compile(); !errors.Is(err, fs.ErrNotExist) || !strings.Contains(err.Error(), "builtins/eng/core.l") {
		t.Fatalf("missing selected core used a fallback: %v", err)
	}
}
