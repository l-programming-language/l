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
)

// These fingerprints record the native inputs immediately before their
// relocation. Moving native support must not change its implementation bytes.
func TestNativeRuntimeResourceLayout(t *testing.T) {
	resources := distribution.Files()
	for name, digest := range map[string]string{
		"main":          "d8d5d7c7fd7ac577dcdb69547924861b1e55bab6c89cc60822b6b372a64cf9b4",
		"host":          "3cd54bfc39c3389a655115ffd3fa01a319475f57601afedbd86931b757e2377b",
		"process_unix":  "fcfcb0f60693ad5a0a09d7120c6a8a52d3a1ba4968684e5845de3729d811fd48",
		"process_other": "27b43fd571a52a2283ffe40517e0b4cb7ec6a1c315bb91f3a2b5aa5fb4290c98",
	} {
		t.Run(name, func(t *testing.T) {
			path := "runtime/rust/" + name + ".rs"
			content, err := fs.ReadFile(resources, path)
			if err != nil {
				t.Fatalf("required native resource %s: %v", path, err)
			}
			if actual := fmt.Sprintf("%x", sha256.Sum256(content)); actual != digest {
				t.Fatalf("native relocation changed %s: SHA-256 %s, want %s", path, actual, digest)
			}
			if _, err := fs.Stat(resources, "builtins/templates/runtime/"+name+".lu"); !errors.Is(err, fs.ErrNotExist) {
				t.Fatalf("obsolete native resource remains: %s: %v", name, err)
			}
		})
	}
}

func TestNativeRuntimeSelectedProviderIsolation(t *testing.T) {
	for _, name := range []string{"main", "host", "process_unix", "process_other"} {
		t.Run(name, func(t *testing.T) {
			files := compositionResources(t, "7")
			path := "runtime/rust/" + name + ".rs"
			original, err := fs.ReadFile(files, path)
			if err != nil {
				t.Fatal(err)
			}
			if _, err := loadSupportTemplates(files, true); err != nil {
				t.Fatalf("complete selected-provider control: %v", err)
			}
			// A usable obsolete asset must not repair a missing selected input.
			files["builtins/templates/runtime/"+name+".lu"] = &fstest.MapFile{Data: bytes.Clone(original)}
			delete(files, path)
			if _, err := loadSupportTemplates(files, true); !errors.Is(err, fs.ErrNotExist) || !strings.Contains(err.Error(), path) {
				t.Fatalf("missing selected native resource used fallback or lost cause: %v", err)
			}
			if _, err := loadSupportTemplates(files, false); err != nil {
				t.Fatalf("ordinary generation read optional native support: %v", err)
			}
			files[path] = &fstest.MapFile{Data: append(bytes.Clone(original), []byte("\r\n// selected native provider \xe9\x9b\xaa")...)}
			selected, err := loadSupportTemplates(files, true)
			if err != nil || !bytes.Equal(selected["runtime/"+name+".rs"], files[path].Data) {
				t.Fatalf("warm loader lost selected native bytes: %v", err)
			}
			selected["runtime/"+name+".rs"][0] = '!'
			if files[path].Data[0] == '!' {
				t.Fatal("returned native support aliases provider bytes")
			}
		})
	}
}
