//go:build darwin || linux

package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestSourceBundleValidation(t *testing.T) {
	for _, tc := range []struct {
		name     string
		paths    []string
		contents []string
		want     string
	}{
		{"empty", nil, nil, "empty"},
		{"unequal dimensions", []string{"/main.rs"}, nil, "dimensions"},
		{"traversal", []string{"/../main.rs"}, []string{"fn main() {}"}, "path"},
		{"nested traversal", []string{"/a/../main.rs"}, []string{"fn main() {}"}, "path"},
		{"backslash", []string{`a\main.rs`}, []string{"fn main() {}"}, "path"},
		{"duplicate", []string{"/main.rs", "main.rs"}, []string{"fn main() {}", "fn main() {}"}, "duplicate"},
		{"case collision", []string{"/main.rs", "/MAIN.rs"}, []string{"fn main() {}", "fn main() {}"}, "duplicate"},
		{"directory conflict", []string{"/helper", "/helper/main.rs"}, []string{"data", "fn main() {}"}, "collision"},
		{"reserved module", []string{"/.cargo/config.toml", "/src/main.rs"}, []string{"[build]\nrustc=\"injected\"", "fn main() {}"}, "reserved"},
		{"reserved module case", []string{"/.cargo/config.TOML", "/src/main.rs"}, []string{"[build]\nrustc=\"injected\"", "fn main() {}"}, "reserved"},
		{"test source", []string{"/tests/main.rs"}, []string{"fn main() {}"}, "test"},
		{"invalid Rust", []string{"/main.rs"}, []string{"fn main( {"}, "Rust"},
		{"no entrypoint", []string{"/main.rs"}, []string{"fn worker() {}"}, "main"},
		{"external import", []string{"/Cargo.toml", "/src/main.rs"}, []string{cargoFixtureManifest + "\n[dependencies]\nexternal=\"999.0.0\"\n", "fn main() {}"}, "dependency"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var contents [][]byte
			for _, content := range tc.contents {
				contents = append(contents, []byte(content))
			}
			_, err := planSourceBundle(contents, tc.paths)
			if err == nil || !strings.Contains(err.Error(), tc.want) {
				t.Fatalf("error = %v; want %q", err, tc.want)
			}
		})
	}
}

func TestSourceBundleNeverOverwritesExistingFiles(t *testing.T) {
	bundle, err := plainSourceBundle("fn main() {}\n")
	if err != nil {
		t.Fatal(err)
	}
	dir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(dir, "src"), 0700); err != nil {
		t.Fatal(err)
	}
	existing := filepath.Join(dir, "src/main.rs")
	if err := os.WriteFile(existing, []byte("original"), 0600); err != nil {
		t.Fatal(err)
	}
	if err := bundle.write(dir); err == nil {
		t.Fatal("source writer overwrote an existing destination")
	}
	got, callErr := os.ReadFile(existing)
	if callErr != nil {
		t.Fatal(callErr)
	}
	if string(got) != "original" {
		t.Fatalf("existing bytes changed to %q", got)
	}
}

func TestSourceBundlePreservesNestedPaths(t *testing.T) {
	paths := []string{"/Cargo.toml", "/Cargo.lock", "/src/main.rs", "/src/helper.rs", "/start.sh"}
	contents := [][]byte{[]byte(cargoFixtureManifest), []byte(cargoFixtureLock), []byte("mod helper; fn main() { helper::work(); }\n"), []byte("pub fn work() {}\n"), []byte("this script must not execute\n")}
	bundle, err := planSourceBundle(contents, paths)
	if err != nil {
		t.Fatal(err)
	}
	dir := t.TempDir()
	if err := bundle.write(dir); err != nil {
		t.Fatal(err)
	}
	for i, name := range paths {
		got, err := os.ReadFile(filepath.Join(dir, strings.TrimPrefix(name, "/")))
		if err != nil || string(got) != string(contents[i]) {
			t.Fatalf("%s bytes=%q error=%v", name, got, err)
		}
	}
}
