package distribution

import (
	"io/fs"
	"reflect"
	"sort"
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
	for _, path := range []string{"lexicon/eng/grammar.json", "lexicon/eng/vocabulary.json", "builtins/eng/core.l", "dictionaries.json", "stdlib/eng/io/dictionary.l", "builtins/eng/storage/dictionary.l", "runtime/eng/foundations.l", "runtime/generated/support.aml", "runtime/generated/support.json", "runtime/rust/host.rs", "runtime/rust/process_unix.rs", "runtime/rust/process_other.rs"} {
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
	expected := Resources(t)
	sort.Strings(files)
	sort.Strings(expected)
	if !reflect.DeepEqual(files, expected) {
		t.Errorf("embedded resources = %v; want exact release inventory %v", files, expected)
	}
}
