package Logiceme

import (
	"bytes"
	"errors"
	"io/fs"
	"strings"
	"sync"
	"testing"

	"github.com/l-programming-language/l/lerrors"
)

func TestRuntimeCompositionCacheRereadsSelectedResources(t *testing.T) {
	for _, name := range []string{
		"builtins/runtime/support.json", "builtins/runtime/support.aml",
		"builtins/runtime/eng/number.l", "builtins/segments/vowels/calculation.lu",
		"builtins/templates/install/cargo.lu",
	} {
		t.Run(name, func(t *testing.T) {
			files := compositionResources(t, "7")
			original, err := loadSupportTemplates(files, true)
			if err != nil {
				t.Fatal(err)
			}
			selected := files[name]
			delete(files, name)
			if _, err := loadSupportTemplates(files, true); !errors.Is(err, fs.ErrNotExist) || !strings.Contains(err.Error(), name) {
				t.Fatalf("missing selected input = %v", err)
			}
			files[name] = selected
			before := selected.Data
			selected.Data = []byte("damaged selected input")
			if _, err := loadSupportTemplates(files, true); !errors.Is(err, lerrors.InvalidConfiguration) {
				t.Fatalf("damaged selected input = %v", err)
			}
			selected.Data = before
			restored, err := loadSupportTemplates(files, true)
			if err != nil || !bytes.Equal(restored["runtime/runtime.rs"], original["runtime/runtime.rs"]) {
				t.Fatalf("restored provider differs after errors: %v", err)
			}
		})
	}
}

func TestRuntimeCompositionCacheTracksSelectedContent(t *testing.T) {
	files := compositionResources(t, "7")
	first, err := loadSupportTemplates(files, true)
	if err != nil {
		t.Fatal(err)
	}
	first["runtime/runtime.rs"][0] = '!'
	second, err := loadSupportTemplates(files, true)
	if err != nil || second["runtime/runtime.rs"][0] == '!' {
		t.Fatalf("returned bytes alias cached output: %v", err)
	}
	const segment = "builtins/segments/vowels/calculation.lu"
	files[segment].Data = bytes.Replace(files[segment].Data, []byte("[|Meta.Statement|]"), []byte("/* selected segment */ [|Meta.Statement|]"), 1)
	changed, err := loadSupportTemplates(files, true)
	if err != nil || !bytes.Contains(changed["runtime/runtime.rs"], []byte("/* selected segment */")) {
		t.Fatalf("selected segment did not control cached generation: %v", err)
	}
	// A valid Lucid template can still fail support assembly. Such failures
	// must not be hidden behind a cached runtime with the same AML.
	const cargo = "builtins/templates/install/cargo.lu"
	before := files[cargo].Data
	files[cargo].Data = bytes.Replace(before, []byte("[[bin]]"), []byte("[[changed]]"), 1)
	if _, err := loadSupportTemplates(files, true); !errors.Is(err, lerrors.InvalidConfiguration) || !strings.Contains(err.Error(), "binary target stanza") {
		t.Fatalf("changed support template = %v", err)
	}
	files[cargo].Data = before
	replacement := compositionResources(t, "9")
	for _, name := range []string{"builtins/runtime/support.json", "builtins/runtime/support.aml", "builtins/runtime/eng/number.l"} {
		files[name] = replacement[name]
	}
	changed, err = loadSupportTemplates(files, true)
	if err != nil || !bytes.Contains(changed["runtime/runtime.rs"], []byte("return 9;")) {
		t.Fatalf("updated selected composition did not control generation: %v", err)
	}
}

func TestRuntimeCompositionCacheConcurrentResults(t *testing.T) {
	files := compositionResources(t, "103")
	var workers sync.WaitGroup
	for i := 0; i < 8; i++ {
		workers.Add(1)
		go func() {
			defer workers.Done()
			for j := 0; j < 2; j++ {
				result, err := loadSupportTemplates(files, true)
				if err != nil {
					t.Error(err)
					return
				}
				runtime := result["runtime/runtime.rs"]
				if runtime[0] == '!' || !bytes.Contains(runtime, []byte("return 103;")) {
					t.Error("concurrent generation shared mutable output")
				}
				runtime[0] = '!'
			}
		}()
	}
	workers.Wait()
}
