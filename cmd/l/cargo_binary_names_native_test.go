//go:build darwin || linux

package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	nativetest "github.com/l-programming-language/l/internal/testsupport/native"
	"github.com/l-programming-language/l/lerrors"
)

func cargoBinaryNameTestBundle(t *testing.T, name, target, marker string) *sourceBundle {
	t.Helper()
	manifest := fmt.Sprintf("[package]\nname = %q\nversion = \"0.0.0\"\nedition = \"2024\"\n[[bin]]\nname = %q\npath = \"src/main.rs\"\n", name, target)
	source := fmt.Sprintf("fn main() { println!(%q); }\n", marker)
	bundle, err := planSourceBundle([][]byte{[]byte(manifest), []byte(source)}, []string{"Cargo.toml", "src/main.rs"})
	if err != nil {
		t.Fatal(err)
	}
	return bundle
}

func TestCargoWorkspaceRejectsCollidingBinaryTargets(t *testing.T) {
	for _, tc := range []struct {
		name, entry, first, second string
		packages, targets          []string
	}{
		{"same App targets", "entry", "worker", "worker", []string{"First", "Second"}, []string{"worker", "worker"}},
		{"case equivalent App targets", "entry", "Worker", "worker", []string{"First", "Second"}, []string{"Worker", "worker"}},
		{"entry target", "ENTRY", "entry", "other", []string{"entry_package", "First"}, []string{"ENTRY", "entry"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			directory := t.TempDir()
			workspace := filepath.Join(directory, "workspace")
			entry := cargoBinaryNameTestBundle(t, "entry_package", tc.entry, "entry")
			apps := []*sourceBundle{cargoBinaryNameTestBundle(t, "First", tc.first, "first"), cargoBinaryNameTestBundle(t, "Second", tc.second, "second")}
			destinations := []string{filepath.Join(directory, "first"), filepath.Join(directory, "second")}
			output := filepath.Join(directory, "entry")
			ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
			defer cancel()
			err := compileCargoWorkspace(ctx, apps, entry, nil, workspace, destinations, output)
			if !errors.Is(err, lerrors.InvalidInput) {
				t.Errorf("colliding binary targets require InvalidInput, got %v", err)
			} else {
				for _, name := range tc.packages {
					if !strings.Contains(err.Error(), name) {
						t.Errorf("collision diagnostic omitted package %q: %v", name, err)
					}
				}
				for _, name := range tc.targets {
					if !strings.Contains(err.Error(), fmt.Sprintf("%q", name)) {
						t.Errorf("collision diagnostic omitted binary target %q: %v", name, err)
					}
				}
			}
			for _, path := range append([]string{workspace, output}, destinations...) {
				if _, err := os.Lstat(path); !os.IsNotExist(err) {
					t.Errorf("invalid targets created %s before rejection: %v", path, err)
				}
			}
		})
	}
}

func TestCargoWorkspaceAllowsDistinctBinaryTargets(t *testing.T) {
	parallelIndependentBuildRoot(t)
	// Package identities stay case-sensitive; binary filenames do not equate
	// hyphens with underscores. Library-only members have no binary to reserve.
	apps := []*sourceBundle{cargoBinaryNameTestBundle(t, "Worker", "work-er", "upper"), cargoBinaryNameTestBundle(t, "worker", "work_er", "lower")}
	entry := cargoBinaryNameTestBundle(t, "entry", "entry", "entry")
	var libraries []*sourceBundle
	for _, name := range []string{"first_library", "second_library"} {
		manifest := fmt.Sprintf("[package]\nname = %q\nversion = \"0.0.0\"\nedition = \"2024\"\nautobins = false\n[lib]\npath = \"src/lib.rs\"\n", name)
		bundle, err := planSourceBundle([][]byte{[]byte(manifest), []byte("pub fn value() -> u8 { 7 }\n")}, []string{"Cargo.toml", "src/lib.rs"})
		if err != nil {
			t.Fatal(err)
		}
		libraries = append(libraries, bundle)
	}
	directory := t.TempDir()
	outputs := []string{filepath.Join(directory, "first"), filepath.Join(directory, "second")}
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()
	if err := compileCargoWorkspace(ctx, apps, entry, libraries, filepath.Join(directory, "workspace"), outputs, filepath.Join(directory, "entry")); err != nil {
		t.Fatal(err)
	}
	for index, want := range []string{"upper\n", "lower\n"} {
		nativetest.AssertNativeArtifact(t, outputs[index])
		stdout, stderr, err := runRuntimeCommand(t, outputs[index], directory, nil)
		if err != nil || string(stdout) != want || len(stderr) != 0 {
			t.Errorf("binary %d: stdout=%q stderr=%q error=%v; want %q", index, stdout, stderr, err, want)
		}
	}
}
