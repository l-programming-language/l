//go:build darwin || linux

package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	nativetest "github.com/l-programming-language/l/internal/testsupport/native"
	Logiceme "github.com/l-programming-language/l/logiceme"
)

// The historical and Rust roots share a bound while each case retains its own
// project, provenance recorder, process group and ordered publication checks.
var buildCLICases = make(chan struct{}, 4)

// Independent compiler/resource roots retain their ordered assertions while
// sharing the same admission bound as native CLI cases.
func parallelIndependentBuildRoot(t *testing.T) {
	t.Helper()
	t.Parallel()
	buildCLICases <- struct{}{}
	t.Cleanup(func() { <-buildCLICases })
}

// These are independent input projects, never replacements for returned source.
const cargoFixtureManifest = "[package]\nname = \"cli_parity\"\nversion = \"0.0.0\"\nedition = \"2024\"\n"

const cargoFixtureLock = "version = 4\n\n[[package]]\nname = \"cli_parity\"\nversion = \"0.0.0\"\n"

func plainSourceBundle(source string) (*sourceBundle, error) {
	return planSourceBundle([][]byte{[]byte(cargoFixtureManifest), []byte(cargoFixtureLock), []byte(source)}, []string{"/Cargo.toml", "/Cargo.lock", "/src/main.rs"})
}

func buildReturnedBundle(t *testing.T, b Logiceme.Executable) {
	t.Helper()
	observeBundleSources(t, b)
	files := map[string][]byte{}
	for i, p := range b.MapArrays[0] {
		files[strings.TrimPrefix(p, "/")] = bytes.Clone(b.ByteArrays[0][i])
	}
	if len(files["Cargo.toml"]) == 0 {
		t.Fatal("actual returned Rust project has no Cargo.toml")
	}
	attachRetainedCargoLock(t, files)
	directory := nativetest.Materialize(t, files)
	built := nativetest.Build(t, directory, nativetest.BuildOptions{Timeout: 3 * time.Minute})
	if built.Err != nil {
		t.Fatalf("actual generated Rust build failed: %v\n%s", built.Err, built.Stderr)
	}
	binaries := 0
	for _, a := range built.Artifacts {
		for _, kind := range a.TargetKinds {
			if kind == "bin" {
				nativetest.AssertNativeArtifact(t, a.Executable)
				binaries++
			}
		}
	}
	if binaries != 1 {
		t.Fatalf("Cargo emitted %d executable artifacts, want one", binaries)
	}
	for p, want := range files {
		got, e := os.ReadFile(filepath.Join(directory, p))
		if e != nil || !bytes.Equal(got, want) {
			t.Fatalf("Cargo source changed: %s %v", p, e)
		}
	}
}

func independentCargoArtifact(t *testing.T) string {
	t.Helper()
	dir := nativetest.Materialize(t, map[string][]byte{"Cargo.toml": []byte(cargoFixtureManifest), "Cargo.lock": []byte(cargoFixtureLock), "src/main.rs": []byte("fn main() {}\n")})
	built := nativetest.Build(t, dir, nativetest.BuildOptions{Timeout: 3 * time.Minute})
	if built.Err != nil {
		t.Fatalf("independent tool-supervision control: %v %s", built.Err, built.Stderr)
	}
	for _, a := range built.Artifacts {
		if a.Executable != "" {
			nativetest.AssertNativeArtifact(t, a.Executable)
			return a.Executable
		}
	}
	t.Fatal("independent Cargo produced no artifact")
	return ""
}

// Finite independently authored manifest/lock pairs. Actual source/manifests
// are never replaced. An unknown manifest is a failing prerequisite.
func attachRetainedCargoLock(t *testing.T, files map[string][]byte) {
	t.Helper()
	matched := false
	entries, e := os.ReadDir("test/builds")
	if e != nil {
		t.Fatal(e)
	}
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		root := filepath.Join("test/builds", entry.Name())
		manifest, e := os.ReadFile(filepath.Join(root, "Cargo.toml"))
		if e != nil {
			t.Fatal(e)
		}
		if !bytes.Equal(files["Cargo.toml"], manifest) {
			continue
		}
		if matched {
			t.Fatal("duplicate independent manifest")
		}
		matched = true
		lock, e := os.ReadFile(filepath.Join(root, "Cargo.lock"))
		if e != nil {
			t.Fatal(e)
		}
		if previous, ok := files["Cargo.lock"]; ok && !bytes.Equal(previous, lock) {
			t.Fatal("actual retained lock differs from independent input")
		}
		files["Cargo.lock"] = bytes.Clone(lock)
	}
	if !matched {
		t.Fatalf("actual Cargo manifest has no finite independent expectation:\n%s", files["Cargo.toml"])
	}
}
