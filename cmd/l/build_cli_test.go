//go:build darwin || linux

package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	distributiontest "github.com/l-programming-language/l/internal/testsupport/distribution"
	nativetest "github.com/l-programming-language/l/internal/testsupport/native"
	"github.com/l-programming-language/l/internal/testsupport/phase1"
)

func TestBuildCLI(t *testing.T) {
	const source = "Create \"Hello\" as a New Basic App.\nThe App Prints `literal <DEF> 世界`.\n"
	cli := filepath.Join(t.TempDir(), "compiler installation", "bin", "l")
	phase1.CopyFile(t, phase1.CLI(t), cli, 0700)
	for _, tc := range []struct {
		name, filename          string
		defaultSource, override bool
	}{
		{"explicit source", "main.l", false, false},
		{"default source", "main.l", true, false},
		{"renamed source with spaces", "worker project.l", false, false},
		{"explicit paths and native environment", "-worker.l", false, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			project, err := filepath.EvalSymlinks(t.TempDir())
			if err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(filepath.Join(project, tc.filename), []byte(source), 0600); err != nil {
				t.Fatal(err)
			}
			args := []string{"build"}
			outputDirectory := project
			var environment []string
			if tc.override {
				outputDirectory = filepath.Join(project, "build output")
				args = append(args, "--assets", distributiontest.CompleteOverride(t), "--output-dir", "build output", "--")
				environment = []string{"RUSTFLAGS=--invalid-rust-flag", "CARGO_BUILD_TARGET=invalid-host", "RUSTC_WRAPPER=/missing/rustc-wrapper", "GOOS=windows", "GOARCH=386", "GO111MODULE=off", "GOFLAGS=-invalid-flag", "GOWORK=" + filepath.Join(project, "missing.work"), "GOTOOLCHAIN=go99.0.0"}
			}
			if !tc.defaultSource {
				args = append(args, tc.filename)
			}
			recorder := nativetest.CLIProbe(t)
			stdout, stderr, err := runBuildCommand(t, cli, project, recorder.Env(environment), args...)
			executable := filepath.Join(outputDirectory, strings.TrimSuffix(tc.filename, ".l"))
			if err != nil || string(stdout) != executable+"\n" || len(stderr) != 0 {
				t.Fatalf("CLI build: %v stdout=%q stderr=%q", err, stdout, stderr)
			}
			recorder.AssertPublished(t, executable)
			nativetest.AssertNativeArtifact(t, executable)
			stdout, stderr, err = runRuntimeCommand(t, executable, t.TempDir(), nil)
			if err != nil || string(stdout) != "literal <DEF> 世界\n" || len(stderr) != 0 {
				t.Fatalf("standalone execution: %v stdout=%q stderr=%q", err, stdout, stderr)
			}
			if tc.name == "renamed source with spaces" {
				moved := filepath.Join(t.TempDir(), "moved program")
				if err := os.Rename(executable, moved); err != nil {
					t.Fatal(err)
				}
				stdout, stderr, err = runRuntimeCommand(t, moved, t.TempDir(), nil)
				if err != nil || string(stdout) != "literal <DEF> 世界\n" || len(stderr) != 0 {
					t.Fatalf("relocated execution: %v stdout=%q stderr=%q", err, stdout, stderr)
				}
			}
		})
	}
}
