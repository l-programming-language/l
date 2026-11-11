package Lyger

import (
	"bytes"
	nativetest "github.com/l-programming-language/l/internal/testsupport/native"
	Logiceme "github.com/l-programming-language/l/logiceme"

	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// Build actual returned source/manifests. The source-bundle API deliberately
// omits locks; a finite independently authored manifest/lock pair supplies frozen
// resolution only after the actual manifest matches its expectation exactly.
func buildActualBundle(t *testing.T, bundle Logiceme.Executable, name string, withRuntime bool) string {
	t.Helper()
	observeBundleSources(t, bundle)
	files := map[string][]byte{}
	for i, p := range bundle.MapArrays[0] {
		files[strings.TrimPrefix(p, "/")] = bytes.Clone(bundle.ByteArrays[0][i])
	}
	if len(files["Cargo.toml"]) == 0 {
		t.Fatalf("Rust workspace prerequisite: actual generated Cargo.toml for %s is missing; compilation/runtime blocked", name)
	}
	attachRetainedBuildLock(t, files, name, withRuntime)
	if withRuntime && len(files["src/runtime.rs"]) == 0 {
		t.Fatal("Rust runtime prerequisite: generated src/runtime.rs missing")
	}

	dir := nativetest.Materialize(t, files)
	built := nativetest.Build(t, dir, nativetest.BuildOptions{Timeout: 3 * time.Minute})
	phase1Evidence(t, "cargo.stdout.jsonl", built.Stdout)
	phase1Evidence(t, "cargo.stderr.txt", built.Stderr)
	if built.Err != nil {
		t.Fatalf("actual generated Rust Cargo build failed; runtime blocked: %v\n%s", built.Err, built.Stderr)
	}

	var executable string
	for _, a := range built.Artifacts {
		for _, kind := range a.TargetKinds {
			if kind == "bin" {
				if executable != "" {
					t.Fatal("generated application has multiple executable targets without primary selection")
				}
				executable = a.Executable
			}
		}
	}
	if executable == "" {
		t.Fatal("Cargo returned no native executable artifact")
	}
	nativetest.AssertNativeArtifact(t, executable)
	// Keep companions beside the observed executable exactly as AppArtifact requests.
	destination := filepath.Join(dir, "entry")
	data, err := os.ReadFile(executable)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(destination, data, 0700); err != nil {
		t.Fatal(err)
	}
	return destination
}

func attachRetainedBuildLock(t *testing.T, files map[string][]byte, name string, runtime bool) {
	t.Helper()
	allowed := map[string]bool{"Recursive": false, "Interleaved": false, "Worker": false, "Other": false, "First": false, "Second": false}
	wantRuntime, ok := allowed[name]
	if !ok || runtime != wantRuntime {
		t.Fatalf("test fixture has no independently reviewed native manifest/lock for %s (runtime=%t)", name, runtime)
	}
	base := filepath.Join("test", "builds", name)
	want, err := os.ReadFile(filepath.Join(base, "Cargo.toml"))
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(files["Cargo.toml"], want) {
		t.Fatalf("Rust generated manifest prerequisite for %s differs from the independent native expectation:\nactual:\n%s\nexpected:\n%s", name, files["Cargo.toml"], want)
	}
	lock, err := os.ReadFile(filepath.Join(base, "Cargo.lock"))
	if err != nil {
		t.Fatal(err)
	}
	if actual, exists := files["Cargo.lock"]; exists {
		if !bytes.Equal(actual, lock) {
			t.Fatal("returned lock differs from retained independent resolution")
		}
		return
	}
	files["Cargo.lock"] = lock
}

func recordBundle(t *testing.T, bundle Logiceme.Executable) {
	t.Helper()
	nativetest.RecordBundle(t, "lyger-generated", bundle)
}
