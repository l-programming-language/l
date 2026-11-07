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
		"runtime/generated/support.json", "runtime/generated/support.aml",
		"dictionaries.json",
		"stdlib/eng/format/dictionary.l", "builtins/segments/vowels/calculation.lu",
		"builtins/templates/install/cargo.lu",
	} {
		t.Run(name, func(t *testing.T) {
			files := compositionResources(t, "7")
			original := composedRuntime(t, files)
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
			if !bytes.Equal(composedRuntime(t, files), original) {
				t.Fatal("restored provider differs after errors")
			}
		})
	}
}

func TestRuntimeCompositionCacheTracksSelectedContent(t *testing.T) {
	files := compositionResources(t, "7")
	first := composedRuntime(t, files)
	first[0] = '!'
	if composedRuntime(t, files)[0] == '!' {
		t.Fatal("returned bytes alias cached output")
	}
	const segment = "builtins/segments/vowels/calculation.lu"
	files[segment].Data = bytes.Replace(files[segment].Data, []byte("[|Meta.Statement|]"), []byte("/* selected segment */ [|Meta.Statement|]"), 1)
	if !bytes.Contains(composedRuntime(t, files), []byte("/* selected segment */")) {
		t.Fatal("selected segment did not control cached generation")
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
	for _, name := range []string{"runtime/generated/support.json", "runtime/generated/support.aml", "runtime/eng/errors.l"} {
		files[name] = replacement[name]
	}
	if !bytes.Contains(composedRuntime(t, files), []byte("selected cancellation: 9")) {
		t.Fatal("updated selected composition did not control generation")
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
				if runtime[0] == '!' || !bytes.Contains(runtime, []byte("selected cancellation: 103")) {
					t.Error("concurrent generation shared mutable output")
				}
				runtime[0] = '!'
			}
		}()
	}
	workers.Wait()
}
