package cargocache

import (
	"encoding/json"
	"io/fs"
	"os"
	"path/filepath"
	"reflect"
	"testing"
	"time"
)

func currentWrite(t *testing.T, root string, files map[string]string) {
	t.Helper()
	for name, data := range files {
		path := filepath.Join(root, name)
		if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte(data), 0600); err != nil {
			t.Fatal(err)
		}
	}
}

func currentSnapshot(t *testing.T, root string) map[string]string {
	t.Helper()
	result := map[string]string{}
	if err := filepath.WalkDir(root, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		name, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}
		if entry.IsDir() {
			result[name] = "directory"
			return nil
		}
		if entry.Type()&os.ModeSymlink != 0 {
			value, err := os.Readlink(path)
			result[name] = "symlink:" + value
			return err
		}
		data, err := os.ReadFile(path)
		result[name] = "file:" + string(data)
		return err
	}); err != nil {
		t.Fatal(err)
	}
	return result
}

func TestStageCurrentReconcilesGraphsWithoutChangingPackagePaths(t *testing.T) {
	stageCurrent := StageCurrent
	cache := t.TempDir()
	var first string
	stamp := time.Unix(123456789, 0)
	for index, graph := range []string{"A", "B", "A"} {
		source := t.TempDir()
		currentWrite(t, source, map[string]string{
			"Cargo.toml":                         "workspace " + graph,
			"support/runtime/Cargo.toml":         "same runtime package",
			"support/runtime/src/lib.rs":         "same runtime bytes",
			"packages/" + graph + "/src/main.rs": "current " + graph,
		})
		root, err := stageCurrent(source, cache, "")
		if err != nil {
			t.Fatal(err)
		}
		if index == 0 {
			first = root
			if err := os.Chtimes(filepath.Join(root, "support/runtime/src/lib.rs"), stamp, stamp); err != nil {
				t.Fatal(err)
			}
		} else if root != first {
			t.Fatalf("graph %s changed Cargo package paths: %q != %q", graph, root, first)
		}
		data, err := os.ReadFile(filepath.Join(root, "Cargo.toml"))
		if err != nil || string(data) != "workspace "+graph {
			t.Fatalf("current manifest: %q %v", data, err)
		}
		info, err := os.Stat(filepath.Join(root, "support/runtime/src/lib.rs"))
		if err != nil || !info.ModTime().Equal(stamp) {
			t.Fatalf("unchanged runtime mtime: %v %v", info, err)
		}
		other := "A"
		if graph == "A" {
			other = "B"
		}
		if _, err := os.Lstat(filepath.Join(root, "packages", other)); !os.IsNotExist(err) {
			t.Fatalf("obsolete owned package remains: %v", err)
		}
		want := currentSnapshot(t, source)
		if got := currentSnapshot(t, root); !reflect.DeepEqual(got, want) {
			t.Fatalf("staged inventory differs: got %v want %v", got, want)
		}
	}
}

func TestStageCurrentRejectsBeforeMutating(t *testing.T) {
	stageCurrent := StageCurrent
	for _, fault := range []string{"unknown-file", "unknown-directory", "cached-symlink", "source-symlink", "duplicate-inventory", "escaping-inventory", "missing-owned-file"} {
		t.Run(fault, func(t *testing.T) {
			source, cache := t.TempDir(), t.TempDir()
			currentWrite(t, source, map[string]string{"Cargo.toml": "same manifest", "src/main.rs": "before"})
			root, err := stageCurrent(source, cache, "")
			if err != nil {
				t.Fatal(err)
			}
			currentWrite(t, source, map[string]string{"src/main.rs": "after"})
			switch fault {
			case "unknown-file":
				currentWrite(t, root, map[string]string{"extra.rs": "unexpected"})
			case "unknown-directory":
				if err := os.Mkdir(filepath.Join(root, "extra"), 0700); err != nil {
					t.Fatal(err)
				}
			case "cached-symlink":
				if err := os.Symlink(t.TempDir(), filepath.Join(root, "extra")); err != nil {
					t.Fatal(err)
				}
			case "source-symlink":
				if err := os.Symlink(t.TempDir(), filepath.Join(source, "extra")); err != nil {
					t.Fatal(err)
				}
			case "duplicate-inventory", "escaping-inventory":
				names := []string{"Cargo.toml", "src/main.rs", "src/main.rs"}
				if fault == "escaping-inventory" {
					names = []string{"../escape", "Cargo.toml", "src/main.rs"}
				}
				data, callErr := json.Marshal(names)
				if callErr != nil {
					t.Fatal(callErr)
				}
				if err := os.WriteFile(filepath.Join(filepath.Dir(root), "inputs.json"), data, 0600); err != nil {
					t.Fatal(err)
				}
			case "missing-owned-file":
				if err := os.Remove(filepath.Join(root, "src/main.rs")); err != nil {
					t.Fatal(err)
				}
			}
			beforeSource, beforeCache := currentSnapshot(t, source), currentSnapshot(t, cache)
			if _, err := stageCurrent(source, cache, ""); err == nil {
				t.Fatal("invalid input/cache state accepted")
			}
			if got := currentSnapshot(t, source); !reflect.DeepEqual(got, beforeSource) {
				t.Fatal("rejected operation mutated supplied source")
			}
			if got := currentSnapshot(t, cache); !reflect.DeepEqual(got, beforeCache) {
				t.Fatal("rejected operation mutated cached inputs")
			}
		})
	}
}

func TestStageCurrentReconcilesOwnedFileDirectoryTransitions(t *testing.T) {
	stageCurrent := StageCurrent
	cache := t.TempDir()
	var root string
	for _, name := range []string{"item", "item/nested.rs", "item"} {
		source := t.TempDir()
		currentWrite(t, source, map[string]string{"Cargo.toml": "same manifest", name: "current bytes"})
		current, err := stageCurrent(source, cache, "")
		if err != nil {
			t.Fatal(err)
		}
		if root != "" && current != root {
			t.Fatal("file/directory transition changed source path")
		}
		root = current
		if !reflect.DeepEqual(currentSnapshot(t, source), currentSnapshot(t, root)) {
			t.Fatal("owned transition left stale source")
		}
	}
}

func TestStageCurrentRejectsPhysicalSourceCacheOverlap(t *testing.T) {
	for _, existing := range []bool{false, true} {
		name := "missing-cache"
		if existing {
			name = "existing-cache"
		}
		t.Run(name, func(t *testing.T) {
			base := t.TempDir()
			real := filepath.Join(base, "real")
			source := filepath.Join(real, "source")
			currentWrite(t, source, map[string]string{"Cargo.toml": "supplied manifest"})
			if existing {
				if err := os.Mkdir(filepath.Join(source, "cache"), 0700); err != nil {
					t.Fatal(err)
				}
			}
			alias := filepath.Join(base, "alias")
			if err := os.Symlink(real, alias); err != nil {
				t.Fatal(err)
			}
			before := currentSnapshot(t, base)
			if _, err := StageCurrent(source, filepath.Join(alias, "source", "cache"), ""); err == nil {
				t.Fatal("physical cache inside supplied source accepted through symlink ancestor")
			}
			if !reflect.DeepEqual(before, currentSnapshot(t, base)) {
				t.Fatal("rejected physical overlap mutated source or cache")
			}
		})
	}
}
