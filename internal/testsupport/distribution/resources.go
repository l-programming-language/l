// Package distribution contains test-owned distribution acceptance utilities.
// It supplies no language definitions, resource fallback, or compiler behavior.
package distribution

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"testing/fstest"

	release "github.com/l-programming-language/l"
)

func RepoRoot(t testing.TB) string {
	t.Helper()
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("cannot locate test support")
	}
	return filepath.Clean(filepath.Join(filepath.Dir(file), "../../.."))
}

// Resources is a test-owned, closed release inventory, not an asset provider.
func Resources(t testing.TB) []string {
	t.Helper()
	data, err := os.ReadFile(filepath.Join(RepoRoot(t), "testdata/distribution/resources.json"))
	if err != nil {
		t.Fatal(err)
	}
	var paths []string
	if err := json.Unmarshal(data, &paths); err != nil {
		t.Fatal(err)
	}
	return paths
}

// Files returns an independently mutable copy of the actual embedded resources.
func Files(t testing.TB) fstest.MapFS {
	t.Helper()
	files := fstest.MapFS{}
	if err := fs.WalkDir(release.Files(), ".", func(name string, entry fs.DirEntry, err error) error {
		if err != nil || entry.IsDir() {
			return err
		}
		data, err := fs.ReadFile(release.Files(), name)
		files[name] = &fstest.MapFile{Data: data}
		return err
	}); err != nil {
		t.Fatal(err)
	}
	return files
}

// CompleteOverride copies only inventoried production resources to a fresh root.
func CompleteOverride(t testing.TB) string {
	t.Helper()
	root := RepoRoot(t)
	files := make(map[string][]byte)
	var missing []string
	for _, path := range Resources(t) {
		data, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(path)))
		if err != nil {
			missing = append(missing, fmt.Sprintf("%s: %v", path, err))
			continue
		}
		files[path] = data
	}
	if len(missing) != 0 {
		t.Fatalf("distribution-assets prerequisite failed; intended assertion not exercised; runtime not exercised:\n%s", strings.Join(missing, "\n"))
	}
	directory := t.TempDir()
	for path, data := range files {
		target := filepath.Join(directory, filepath.FromSlash(path))
		if err := os.MkdirAll(filepath.Dir(target), 0700); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(target, data, 0600); err != nil {
			t.Fatal(err)
		}
	}
	return directory
}
