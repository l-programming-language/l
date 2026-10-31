//go:build darwin || linux

package cargocache_test

import (
	"bytes"
	"encoding/json"
	"path/filepath"
	"testing"

	nativetest "github.com/l-programming-language/l/internal/testsupport/native"
)

func TestActivationPreservesNextUnchangedCargoBuild(t *testing.T) {
	// Cargo compares source mtimes with the preceding compilation's start.
	// Activation must invalidate old products without making the first real
	// compilation stale before it has even started.
	project := nativetest.Materialize(t, map[string][]byte{
		"Cargo.toml": []byte("[package]\nname = \"activation_freshness\"\nversion = \"0.0.0\"\nedition = \"2024\"\n"),
		"Cargo.lock": []byte("version = 4\n\n[[package]]\nname = \"activation_freshness\"\nversion = \"0.0.0\"\n"),
		"src/lib.rs": []byte("pub fn value() -> u32 { 42 }\n"),
	})
	target := filepath.Join(t.TempDir(), "target")
	for index, wantFresh := range []bool{false, true} {
		built := nativetest.Build(t, project, nativetest.BuildOptions{TargetDir: target})
		if built.Err != nil {
			t.Fatalf("native build %d: %v\n%s", index, built.Err, built.Stderr)
		}
		found := false
		for _, line := range bytes.Split(built.Stdout, []byte{'\n'}) {
			var event struct {
				Reason string
				Fresh  bool
				Target struct{ Name string }
			}
			if json.Unmarshal(line, &event) != nil || event.Reason != "compiler-artifact" || event.Target.Name != "activation_freshness" {
				continue
			}
			if found || event.Fresh != wantFresh {
				t.Fatalf("native build %d library fresh=%v, duplicate=%v; want fresh=%v", index, event.Fresh, found, wantFresh)
			}
			found = true
		}
		if !found {
			t.Fatalf("native build %d omitted the library artifact\n%s", index, built.Stdout)
		}
	}
}
