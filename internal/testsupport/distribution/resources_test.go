package distribution

import (
	"io/fs"
	"strings"
	"testing"

	release "github.com/l-programming-language/l"
)

func TestReleaseInventory(t *testing.T) {
	seen := make(map[string]bool)
	for _, path := range Resources(t) {
		if !fs.ValidPath(path) || seen[path] {
			t.Fatalf("invalid or duplicate release path %q", path)
		}
		seen[path] = true
		if strings.Contains(path, "testdata") || strings.Contains(path, "draft") || strings.HasSuffix(path, "_test.go") {
			t.Errorf("non-release resource %q", path)
		}
	}
	for _, path := range []string{"lexicon/eng/grammar.json", "lexicon/eng/vocabulary.json", "builtins/eng/core.l", "stdlib/index.json", "stdlib/eng/io/dictionary.l", "builtins/templates/runtime/runtime.lu", "builtins/templates/runtime/process_unix.lu", "builtins/templates/runtime/process_other.lu"} {
		if !seen[path] {
			t.Errorf("missing required release path %s", path)
		}
	}
}

func TestReleaseSourceResources(t *testing.T) {
	var files []string
	err := fs.WalkDir(release.Files(), ".", func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() {
			return nil
		}
		files = append(files, path)
		for _, excluded := range []string{"builtins/go/", "builtins/eng/go/", "stdlib/eng/go/"} {
			if strings.HasPrefix(path, excluded) {
				t.Errorf("unsupported generated-Go resource is embedded: %s", path)
			}
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(files) != 36 {
		t.Errorf("embedded resources = %d; want 36", len(files))
	}
}
