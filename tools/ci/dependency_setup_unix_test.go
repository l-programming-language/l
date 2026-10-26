//go:build darwin || linux

package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestDependencySeedRetainsProductsAndDoesNotOverwriteWarmCaches(t *testing.T) {
	for _, parity := range []bool{false, true} {
		t.Run(map[bool]string{false: "CLI", true: "parity"}[parity], func(t *testing.T) {
			root := t.TempDir()
			source := filepath.Join(root, "source")
			product := filepath.Join("debug", "deps", "registry-product.rlib")
			original := filepath.Join(source, product)
			testWrite(t, original, []byte("registry bytes"))
			stamp := time.Unix(1600000000, 123456000)
			if err := os.Chmod(original, 0755); err != nil {
				t.Fatal(err)
			}
			if err := os.Chtimes(original, stamp, stamp); err != nil {
				t.Fatal(err)
			}
			target := filepath.Join(root, "destination", "target")
			lock := filepath.Join(filepath.Dir(target), "publication.lock")
			if parity {
				lock = filepath.Join(target, ".l-parity-publication.lock")
			}
			destination := dependencyCacheTarget{Target: target, Lock: lock}
			status, size, err := seedDependencyCache(source, destination)
			if err != nil || status != "seeded" || size != int64(len("registry bytes")) {
				t.Fatalf("seed: %s %d %v", status, size, err)
			}
			copied := filepath.Join(target, product)
			info, err := os.Stat(copied)
			if err != nil || !info.ModTime().Equal(stamp) || info.Mode().Perm() != 0755 {
				t.Fatalf("dependency product metadata changed: %v %v", info, err)
			}
			testWrite(t, original, []byte("changed seed"))
			data, err := os.ReadFile(copied)
			if err != nil || string(data) != "registry bytes" {
				t.Fatal("dependency product was linked to its writable seed", err)
			}
			status, size, err = seedDependencyCache(source, destination)
			if err != nil || status != "existing" || size != 0 {
				t.Fatalf("warm cache was disturbed: %s %d %v", status, size, err)
			}
			data, err = os.ReadFile(copied)
			if err != nil || string(data) != "registry bytes" {
				t.Fatal("warm cache bytes changed", err)
			}
		})
	}
}

func TestDependencySeedRespectsPublicationLock(t *testing.T) {
	root := t.TempDir()
	source, target := filepath.Join(root, "source"), filepath.Join(root, "target")
	testWrite(t, filepath.Join(source, "product"), []byte("registry"))
	lock := filepath.Join(root, "publication.lock")
	release, err := lockDependencyCache(lock)
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := release(); err != nil {
			t.Error(err)
		}
	}()
	if _, _, err := seedDependencyCache(source, dependencyCacheTarget{Target: target, Lock: lock}); err == nil || !strings.Contains(err.Error(), "busy") {
		t.Fatalf("active build lock did not block preparation: %v", err)
	}
	if _, err := os.Stat(target); !os.IsNotExist(err) {
		t.Fatalf("busy cache was mutated: %v", err)
	}
}

func TestDependencySeedRejectsSymlinksAndProjectState(t *testing.T) {
	root := t.TempDir()
	source := filepath.Join(root, "source")
	testWrite(t, filepath.Join(source, "product"), []byte("registry"))
	link := filepath.Join(root, "linked-source")
	if err := os.Symlink(source, link); err != nil {
		t.Fatal(err)
	}
	if _, err := copyDependencyCache(link, t.TempDir()); err == nil {
		t.Fatal("symlink cache root accepted")
	}
	if err := os.Symlink(filepath.Join(source, "product"), filepath.Join(source, "linked-product")); err != nil {
		t.Fatal(err)
	}
	if _, err := copyDependencyCache(source, t.TempDir()); err == nil {
		t.Fatal("symlink product accepted")
	}
	if err := os.Remove(filepath.Join(source, "linked-product")); err != nil {
		t.Fatal(err)
	}
	testWrite(t, filepath.Join(source, ".l-active-inputs"), []byte("project"))
	if _, err := copyDependencyCache(source, t.TempDir()); err == nil {
		t.Fatal("local project cache ownership state accepted")
	}
	if _, err := lockDependencyCache(link); err == nil {
		t.Fatal("symlink lock accepted")
	}
}
