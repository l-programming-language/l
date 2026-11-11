package main

import (
	"bytes"
	"crypto/sha256"
	"io/fs"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	distributiontest "github.com/l-programming-language/l/internal/testsupport/distribution"
)

// The ordinary unfiltered gate regenerates AML using the current frontend and
// source resources. Matching hashes alone cannot establish this relationship.
func TestCanonicalRuntimeCompositionIsCurrent(t *testing.T) {
	if err := run(filepath.Join("..", ".."), true); err != nil {
		t.Fatal(err)
	}
}

func TestCompositionSelectedAssets(t *testing.T) {
	compositionSelectedAssets(t, distributiontest.CompleteOverride(t))
}

func TestCompositionAlternateAssets(t *testing.T) {
	root := t.TempDir()
	if err := os.CopyFS(root, distributiontest.AlternateRuntime(t)); err != nil {
		t.Fatal(err)
	}
	compositionSelectedAssets(t, root)
}

func compositionSelectedAssets(t *testing.T, root string) {
	t.Helper()
	outputs := []string{"stdlib/runtime/generated/support.aml", "stdlib/runtime/generated/support.json"}
	for _, name := range outputs {
		path := filepath.Join(root, filepath.FromSlash(name))
		if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte("stale"), 0600); err != nil {
			t.Fatal(err)
		}
	}
	before := compositionFiles(t, root)
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := run(root, true); err == nil || !strings.Contains(err.Error(), outputs[0]) {
		t.Fatalf("stale selected provider was not rejected: %v", err)
	}
	if !reflect.DeepEqual(before, compositionFiles(t, root)) {
		t.Fatal("check changed the selected provider")
	}
	var diagnostics bytes.Buffer
	if status := command([]string{"-assets", root}, &diagnostics); status != 0 {
		t.Fatalf("selected generation status=%d: %s", status, &diagnostics)
	}
	after := compositionFiles(t, root)
	for _, name := range outputs {
		if before[name] == after[name] {
			t.Fatalf("generation did not replace selected output %s", name)
		}
		before[name] = after[name]
	}
	if !reflect.DeepEqual(before, after) {
		t.Fatal("generation changed files outside the two selected support outputs")
	}
	if err := run(root, false); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(after, compositionFiles(t, root)) {
		t.Fatal("unchanged generation was not deterministic")
	}
	if status := command([]string{"-check", "-assets", root}, &diagnostics); status != 0 {
		t.Fatalf("selected check status=%d: %s", status, &diagnostics)
	}
	if current, err := os.Getwd(); err != nil || current != cwd {
		t.Fatalf("composition changed CWD: %q, %v", current, err)
	}
	if err := os.Remove(filepath.Join(root, "dictionaries.json")); err != nil {
		t.Fatal(err)
	}
	if err := run(root, false); err == nil || !strings.Contains(err.Error(), "dictionaries.json") {
		t.Fatalf("incomplete provider fell back to checkout resources: %v", err)
	}
}

func TestCompositionArguments(t *testing.T) {
	for _, tc := range []struct {
		name   string
		args   []string
		status int
	}{
		{"help", []string{"-help"}, 0},
		{"unknown", []string{"-unknown"}, 2},
		{"missing assets value", []string{"-assets"}, 2},
		{"positional", []string{"unexpected"}, 2},
		{"missing provider", []string{"-assets", filepath.Join(t.TempDir(), "missing"), "-check"}, 1},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var diagnostics bytes.Buffer
			if status := command(tc.args, &diagnostics); status != tc.status || diagnostics.Len() == 0 {
				t.Fatalf("status=%d, want %d; diagnostic=%q", status, tc.status, &diagnostics)
			}
		})
	}
}

func TestCompositionDefaultAssets(t *testing.T) {
	t.Chdir(distributiontest.RepoRoot(t))
	var diagnostics bytes.Buffer
	if status := command([]string{"-check"}, &diagnostics); status != 0 || diagnostics.Len() != 0 {
		t.Fatalf("default provider status=%d: %s", status, &diagnostics)
	}
}

func compositionFiles(t *testing.T, root string) map[string][sha256.Size]byte {
	t.Helper()
	files := make(map[string][sha256.Size]byte)
	provider := os.DirFS(root)
	if err := fs.WalkDir(provider, ".", func(name string, entry fs.DirEntry, err error) error {
		if err != nil || entry.IsDir() {
			return err
		}
		content, err := fs.ReadFile(provider, name)
		if err == nil {
			files[name] = sha256.Sum256(content)
		}
		return err
	}); err != nil {
		t.Fatal(err)
	}
	return files
}
