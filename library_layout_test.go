package distribution_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"io/fs"
	"path/filepath"
	"strings"
	"testing"
	"testing/fstest"

	distribution "github.com/l-programming-language/l"
	Lyger "github.com/l-programming-language/l/lyger"
)

func layoutResources(t *testing.T) fstest.MapFS {
	t.Helper()
	files := fstest.MapFS{}
	if err := fs.WalkDir(distribution.Files(), ".", func(name string, entry fs.DirEntry, err error) error {
		if err != nil || entry.IsDir() {
			return err
		}
		data, err := fs.ReadFile(distribution.Files(), name)
		if err == nil {
			files[name] = &fstest.MapFile{Data: data}
		}
		return err
	}); err != nil {
		t.Fatal(err)
	}
	return files
}

// LAY-003, 005, 009, 016, 024, 044, 046: one physical distribution layout.
func TestLibraryLayoutResources(t *testing.T) {
	for _, name := range []string{
		"dictionaries.json", "builtins/eng/prelude.l", "builtins/eng/core.l",
		"builtins/eng/storage/dictionary.l", "runtime/eng/dictionary.l",
		"runtime/generated/support.aml", "runtime/generated/support.json",
		"runtime/rust/main.rs", "runtime/rust/host.rs",
		"runtime/rust/process_unix.rs", "runtime/rust/process_other.rs",
		"stdlib/eng/format/dictionary.l", "stdlib/eng/io/dictionary.l",
	} {
		if _, err := fs.ReadFile(distribution.Files(), name); err != nil {
			t.Errorf("required canonical resource %s: %v", name, err)
		}
	}
	for _, name := range []string{"stdlib/index.json", "stdlib/eng/storage", "builtins/runtime", "builtins/templates/runtime"} {
		if _, err := fs.Stat(distribution.Files(), name); !errors.Is(err, fs.ErrNotExist) {
			t.Errorf("retired resource %s remains available: %v", name, err)
		}
	}
	core, err := fs.ReadFile(distribution.Files(), "builtins/eng/core.l")
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(core), "from `storage`") {
		t.Error("core still owns foundational support loading")
	}
}

// LAY-006, 014, 043: a complete selected provider is the positive control;
// a misleading former registry must not repair a missing canonical registry.
func TestLibraryLayoutRegistryNoFallback(t *testing.T) {
	files := layoutResources(t)
	source := []byte("Create \"Worker\" as a New Basic App.\n")
	compile := func() error {
		_, err := Lyger.NewLygerService(Lyger.Options{Resources: files}).Compile(source, filepath.Join(t.TempDir(), "main.l"))
		return err
	}
	if err := compile(); err != nil {
		t.Fatalf("complete selected provider: %v", err)
	}
	if current := files["dictionaries.json"]; current != nil {
		files["stdlib/index.json"] = &fstest.MapFile{Data: append([]byte(nil), current.Data...)}
	}
	delete(files, "dictionaries.json")
	err := compile()
	if !errors.Is(err, fs.ErrNotExist) || !strings.Contains(err.Error(), "dictionaries.json") {
		t.Fatalf("missing canonical registry must reject despite old registry: %v", err)
	}
}

// LAY-025, 027, 028: registry identity participates in runtime provenance even
// when all old source bytes remain available to the selected provider.
func TestLibraryLayoutCompositionRegistryIdentity(t *testing.T) {
	files := layoutResources(t)
	manifest, _, err := distribution.LoadRuntimeComposition(files)
	if err != nil {
		t.Fatalf("valid composition control: %v", err)
	}
	if manifest.Sources["dictionaries.json"] == "" {
		t.Error("composition omits the selected registry fingerprint")
	}
	files["dictionaries.json"].Data = append(files["dictionaries.json"].Data, '\n')
	if _, _, err := distribution.LoadRuntimeComposition(files); err == nil || !strings.Contains(err.Error(), "dictionaries.json") || !strings.Contains(err.Error(), "stale") {
		t.Fatalf("changed registry retained stale composition: %v", err)
	}
}

func TestLibraryLayoutRegistryRetargetInvalidatesSupport(t *testing.T) {
	files := layoutResources(t)
	if _, _, err := distribution.LoadRuntimeComposition(files); err != nil {
		t.Fatal(err)
	}
	const old = "builtins/eng/storage/dictionary.l"
	const replacement = "builtins/eng/storage/selected.l"
	files[replacement] = &fstest.MapFile{Data: bytes.Clone(files[old].Data)}
	files["dictionaries.json"].Data = bytes.ReplaceAll(files["dictionaries.json"].Data, []byte(old), []byte(replacement))
	// The remapped dictionary is valid and the former sources still exist.
	if _, err := Lyger.NewLygerService(Lyger.Options{Resources: files}).Compile([]byte("Create \"Worker\" as a New Basic App.\n"), filepath.Join(t.TempDir(), "main.l")); err != nil {
		t.Fatalf("valid remapping control: %v", err)
	}
	if _, _, err := distribution.LoadRuntimeComposition(files); err == nil || !strings.Contains(err.Error(), "dictionaries.json") || !strings.Contains(err.Error(), "stale") {
		t.Fatalf("retargeted registry retained old composition: %v", err)
	}
}

// LAY-007, 025, 027: a manifest cannot name arbitrary provider files or retain
// the former canonical runtime root. The complete manifest is the control.
func TestLibraryLayoutManifestPathPolicy(t *testing.T) {
	for _, name := range []string{"../source.l", "/source.l", "runtime/generated/extra.l", "runtime/rust/extra.l", "builtins/runtime/eng/extra.l", "fixture/extra.l", "other.json", "runtime/eng/bad\\file.l"} {
		t.Run(name, func(t *testing.T) {
			files := layoutResources(t)
			manifest, _, err := distribution.LoadRuntimeComposition(files)
			if err != nil {
				t.Fatal(err)
			}
			manifest.Sources[name] = manifest.Sources["builtins/eng/core.l"]
			files[name] = &fstest.MapFile{Data: bytes.Clone(files["builtins/eng/core.l"].Data)}
			raw, err := json.Marshal(manifest)
			if err != nil {
				t.Fatal(err)
			}
			files["runtime/generated/support.json"] = &fstest.MapFile{Data: raw}
			if _, _, err := distribution.LoadRuntimeComposition(files); err == nil || !strings.Contains(err.Error(), "source path") {
				t.Fatalf("invalid manifest path accepted: %v", err)
			}
		})
	}
}

// LAY-008, 010, 012, 014: bootstrap and includes preserve their unit ownership;
// later explicit imports reuse the canonical dictionary within that compilation.
func TestLibraryLayoutBootstrapOwnership(t *testing.T) {
	for _, explicit := range []bool{false, true} {
		source := "Create \"Worker\" as a New Basic App.\n"
		if explicit {
			source = "Use Storage from `storage`.\n" + source
		}
		result, err := Lyger.NewLygerService(Lyger.Options{}).Compile([]byte(source), filepath.Join(t.TempDir(), "main.l"))
		if err != nil {
			t.Fatal(err)
		}
		builtins, storage, coreDeclarations := 0, 0, 0
		for _, unit := range result.Compilation.SourceUnits {
			if unit.Kind == Lyger.SourceBuiltin {
				builtins++
				if unit.Origin.Name != "builtins/eng/prelude.l" {
					t.Fatalf("builtin origin: %s", unit.Origin.Name)
				}
			}
			if unit.Dictionary == "Storage" {
				storage++
			}
		}
		for _, statement := range result.Compilation.SourceStatements {
			if strings.Contains(string(statement.Raw), `Define "Function|-s|-al|-ing"`) {
				coreDeclarations++
				if statement.Unit.Kind != Lyger.SourceBuiltin || statement.Origin.Name != "builtins/eng/core.l" {
					t.Fatalf("core definition lost unit/origin: %+v", statement.Origin)
				}
			}
		}
		if builtins != 1 || storage != 1 || coreDeclarations != 1 {
			t.Fatalf("builtin/storage/core counts=%d/%d/%d", builtins, storage, coreDeclarations)
		}
	}
}

func TestLibraryLayoutSelectedSourceMissing(t *testing.T) {
	for _, name := range []string{"builtins/eng/prelude.l", "builtins/eng/core.l", "builtins/eng/storage/dictionary.l"} {
		t.Run(name, func(t *testing.T) {
			files := layoutResources(t)
			source := []byte("Create \"Worker\" as a New Basic App.\n")
			compile := func() error {
				_, err := Lyger.NewLygerService(Lyger.Options{Resources: files}).Compile(source, filepath.Join(t.TempDir(), "main.l"))
				return err
			}
			if err := compile(); err != nil {
				t.Fatalf("complete provider: %v", err)
			}
			files["stdlib/eng/storage/dictionary.l"] = &fstest.MapFile{Data: append([]byte(nil), files["builtins/eng/storage/dictionary.l"].Data...)}
			delete(files, name)
			if err := compile(); !errors.Is(err, fs.ErrNotExist) || !strings.Contains(err.Error(), name) {
				t.Fatalf("missing selected %s: %v", name, err)
			}
		})
	}
}
