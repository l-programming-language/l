package cargocache

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

func TestShardUsesActualInputBytes(t *testing.T) {
	source := t.TempDir()
	if err := os.WriteFile(filepath.Join(source, "Cargo.toml"), []byte("manifest"), 0600); err != nil {
		t.Fatal(err)
	}
	seen := map[int]bool{}
	for value := 0; value < 64; value++ {
		if err := os.WriteFile(filepath.Join(source, "main.rs"), []byte(fmt.Sprintf("fn main() { let value = %d; }", value)), 0600); err != nil {
			t.Fatal(err)
		}
		first, err := Shard(source, "")
		if err != nil {
			t.Fatal(err)
		}
		second, err := Shard(source, "")
		if err != nil || first != second || first < 0 || first >= Shards {
			t.Fatalf("source identity is not stable: %d %d %v", first, second, err)
		}
		seen[first] = true
	}
	if len(seen) != Shards {
		t.Fatalf("distinct native inputs do not use the bounded cache shards: %v", seen)
	}
}

func TestStagePreservesBytesAndUnchangedInputTime(t *testing.T) {
	source, cache := t.TempDir(), t.TempDir()
	manifest := filepath.Join(source, "Cargo.toml")
	program := filepath.Join(source, "main.rs")
	for name, data := range map[string]string{manifest: "manifest", program: "first"} {
		if err := os.WriteFile(name, []byte(data), 0600); err != nil {
			t.Fatal(err)
		}
	}
	first, err := Stage(source, cache, "")
	if err != nil {
		t.Fatal(err)
	}
	before, err := os.Stat(filepath.Join(first, "Cargo.toml"))
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(program, []byte("second"), 0600); err != nil {
		t.Fatal(err)
	}
	second, err := Stage(source, cache, "")
	if err != nil || second != first {
		t.Fatalf("source-only change changed graph location: %q %v", second, err)
	}
	after, callErr := os.Stat(filepath.Join(second, "Cargo.toml"))
	if callErr != nil {
		t.Fatal(callErr)
	}
	data, callErr := os.ReadFile(filepath.Join(second, "main.rs"))
	if callErr != nil {
		t.Fatal(callErr)
	}
	if !before.ModTime().Equal(after.ModTime()) || string(data) != "second" {
		t.Fatal("input staging changed unchanged file time or lost changed bytes")
	}
	if err := os.WriteFile(manifest, []byte("other graph"), 0600); err != nil {
		t.Fatal(err)
	}
	third, err := Stage(source, cache, "")
	if err != nil || third == first {
		t.Fatalf("manifest change reused old graph location: %v", err)
	}
}

func TestStageRejectsSourceSymlink(t *testing.T) {
	source, cache := t.TempDir(), t.TempDir()
	if err := os.Symlink(filepath.Join(t.TempDir(), "outside"), filepath.Join(source, "main.rs")); err != nil {
		t.Fatal(err)
	}
	if _, err := Stage(source, cache, ""); err == nil {
		t.Fatal("Cargo source symlink accepted")
	}
}

func TestStageRejectsUninventoriedCachedSource(t *testing.T) {
	source, cache := t.TempDir(), t.TempDir()
	if err := os.WriteFile(filepath.Join(source, "Cargo.toml"), []byte("manifest"), 0600); err != nil {
		t.Fatal(err)
	}
	root, err := Stage(source, cache, "")
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "surprise.rs"), []byte("fn main() {}"), 0600); err != nil {
		t.Fatal(err)
	}
	if _, err := Stage(source, cache, ""); err == nil {
		t.Fatal("unplanned source was admitted to Cargo inputs")
	}
}
