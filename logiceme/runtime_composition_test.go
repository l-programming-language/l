package Logiceme

import (
	"bytes"
	"crypto/sha256"
	"errors"
	"fmt"
	"io/fs"
	"strings"
	"testing"
	"testing/fstest"

	distribution "github.com/l-programming-language/l"
	distributiontest "github.com/l-programming-language/l/internal/testsupport/distribution"
	"github.com/l-programming-language/l/lerrors"
	Lyger "github.com/l-programming-language/l/lyger"
)

func compositionResources(t *testing.T, value string) fstest.MapFS {
	t.Helper()
	files := distributiontest.Files(t)
	const source = "stdlib/eng/runtime/errors.l"
	files[source].Data = bytes.ReplaceAll(files[source].Data, []byte("source execution stopped"), []byte("selected cancellation: "+value))
	aml, manifest, err := Lyger.ComposeRuntimeResources(files)
	if err != nil {
		t.Fatal(err)
	}
	files["stdlib/runtime/generated/support.aml"].Data = aml
	files["stdlib/runtime/generated/support.json"].Data = marshalTestJSON(t, manifest)
	return files
}

func composedRuntime(t *testing.T, files fs.FS) []byte {
	t.Helper()
	support, err := loadSupportTemplates(files, true)
	if err != nil {
		t.Fatal(err)
	}
	return support["runtime/runtime.rs"]
}

func TestRuntimeCompositionUsesSelectedResources(t *testing.T) {
	first := composedRuntime(t, compositionResources(t, "first cancellation"))
	second := composedRuntime(t, compositionResources(t, "second cancellation"))
	if !bytes.Contains(first, []byte("selected cancellation: first cancellation")) || !bytes.Contains(second, []byte("selected cancellation: second cancellation")) {
		t.Fatal("selected AML did not control composed support")
	}
	first[0] = '!'
	if second[0] == '!' {
		t.Fatal("generated runtime support aliases another provider")
	}
}

func TestRuntimeCompositionRejectsStaleInputs(t *testing.T) {
	for _, name := range []string{"stdlib/eng/format/dictionary.l", "stdlib/runtime/generated/support.aml"} {
		t.Run(name, func(t *testing.T) {
			files := compositionResources(t, "first cancellation")
			files[name].Data = append(files[name].Data, '\n')
			if _, aml, err := distribution.LoadRuntimeComposition(files); err == nil || aml != nil {
				t.Fatalf("invalid composition returned usable AML: %v", err)
			}
			if _, err := loadSupportTemplates(files, true); !errors.Is(err, lerrors.InvalidConfiguration) || !strings.Contains(err.Error(), name) || !strings.Contains(err.Error(), "stale") {
				t.Fatalf("stale selected input = %v", err)
			}
		})
	}
	files := compositionResources(t, "first cancellation")
	delete(files, "stdlib/runtime/generated/support.aml")
	if _, err := loadSupportTemplates(files, true); !errors.Is(err, fs.ErrNotExist) {
		t.Fatalf("missing selected composition = %v", err)
	}
	if _, err := loadSupportTemplates(files, false); err != nil {
		t.Fatalf("ordinary generation unexpectedly requires runtime: %v", err)
	}
}

func TestRuntimeCompositionRequiresLibraryTarget(t *testing.T) {
	files := compositionResources(t, "first cancellation")
	manifest, _, err := distribution.LoadRuntimeComposition(files)
	if err != nil {
		t.Fatal(err)
	}
	files["stdlib/runtime/generated/support.aml"].Data = bytes.Replace(files["stdlib/runtime/generated/support.aml"].Data, []byte(`"Kind":"lib"`), []byte(`"Kind":"bin"`), 1)
	manifest.SHA256, err = manifest.Fingerprint(files["stdlib/runtime/generated/support.aml"].Data)
	if err != nil {
		t.Fatal(err)
	}
	files["stdlib/runtime/generated/support.json"].Data = marshalTestJSON(t, manifest)
	if _, err := loadSupportTemplates(files, true); !errors.Is(err, lerrors.InvalidConfiguration) || !strings.Contains(err.Error(), "library target") {
		t.Fatalf("binary composition = %v", err)
	}
}

func TestRuntimeCompositionRejectsRehashedOrMissingSource(t *testing.T) {
	for _, warm := range []bool{false, true} {
		for _, mutation := range []string{"rehash", "omit", "selection", "fingerprint"} {
			t.Run(fmt.Sprintf("warm=%t/%s", warm, mutation), func(t *testing.T) {
				files := distributiontest.Files(t)
				result, err := Lyger.NewLygerService(Lyger.Options{Resources: files}).Compile([]byte("Show `provenance`.\n"), "provenance.l")
				if err != nil {
					t.Fatal(err)
				}
				entry, err := result.OrchestratorUnit("entry", nil)
				if err != nil {
					t.Fatal(err)
				}
				assemble := func() error {
					_, _, err := NewLogicemeService(files).EvaluateWorkspaceEntry(entry.BackendAML, WorkspaceRuntime{Package: "provenance_runtime", RelativeDirectory: "../support/runtime"})
					return err
				}
				if err := assemble(); err != nil {
					t.Fatalf("valid complete provider: %v", err)
				}
				if !warm {
					runtimeCompositionCache.Lock()
					runtimeCompositionCache.runtime = nil
					runtimeCompositionCache.Unlock()
				}
				manifest, _, err := distribution.LoadRuntimeComposition(files)
				if err != nil {
					t.Fatal(err)
				}
				const source = "stdlib/eng/runtime/errors.l"
				want := "stale"
				switch mutation {
				case "omit":
					delete(manifest.Sources, source)
					delete(files, source)
				case "selection":
					delete(manifest.Selection, "")
				case "fingerprint":
					manifest.SHA256, want = "malformed", "SHA-256"
				default:
					files[source].Data = bytes.ReplaceAll(files[source].Data, []byte("source execution stopped"), []byte("changed cancellation message"))
					manifest.Sources[source] = fmt.Sprintf("%x", sha256.Sum256(files[source].Data))
				}
				files["stdlib/runtime/generated/support.json"].Data = marshalTestJSON(t, manifest)
				if err := assemble(); !errors.Is(err, lerrors.InvalidConfiguration) || !strings.Contains(err.Error(), want) {
					t.Fatalf("invalid provider retained generated runtime: %v", err)
				}
			})
		}
	}
}
