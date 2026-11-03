//go:build darwin || linux

package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/l-programming-language/l/contracts"
	"github.com/l-programming-language/l/internal/testsupport/phase1"
)

func TestNativeEvidenceRecordsActualBuildWithoutRunningApp(t *testing.T) {
	project := t.TempDir()
	source := filepath.Join(project, "entry.l")
	if err := os.WriteFile(source, []byte("Exit.\n"), 0600); err != nil {
		t.Fatal(err)
	}
	output := filepath.Join(project, "native-evidence.json")
	stdout, stderr, err := runBuildCommand(t, phase1.CLI(t), project, nil, "build", "--native-evidence-output", output, source)
	if err != nil {
		t.Fatalf("evidence-producing native build failed: %v\n%s\n%s", err, stdout, stderr)
	}
	data, err := os.ReadFile(output)
	if err != nil {
		t.Fatal(err)
	}
	document, err := contracts.DecodeDocument(data)
	if err != nil {
		t.Fatal(err)
	}
	environment, ok := document.(contracts.NativeEnvironment)
	if !ok || environment.Method != "build" || len(environment.Artifacts) == 0 {
		t.Fatalf("actual compilation has no build/artifact evidence: %+v", document)
	}
	binary, err := os.ReadFile(strings.TrimSpace(string(stdout)))
	if err != nil {
		t.Fatal(err)
	}
	matched := false
	for _, artifact := range environment.Artifacts {
		matched = matched || artifact.SHA256 == nativeBytesHash(binary)
	}
	if !matched {
		t.Fatal("recorded artifacts do not contain the published orchestrator bytes")
	}
	builds := 0
	for _, execution := range environment.Executions {
		if execution.Kind == "build" {
			builds++
			if execution.Status != "passed" || !containsNativeEnvironment(execution.Command, "--frozen") || execution.Validate() != nil {
				t.Fatalf("build evidence is not the actual frozen invocation: %+v", execution)
			}
		} else if execution.Kind != "environment" {
			t.Fatalf("evidence collection executed an unexpected operation: %+v", execution)
		}
	}
	if builds != 1 {
		t.Fatalf("evidence caused %d Cargo builds, want one", builds)
	}
}

func TestNativeEvidenceRetainsFailedCargoInvocation(t *testing.T) {
	cargo, err := exec.LookPath("cargo")
	if err != nil {
		t.Fatal(err)
	}
	cargo, err = filepath.Abs(cargo)
	if err != nil {
		t.Fatal(err)
	}
	project := t.TempDir()
	tools := filepath.Join(project, "tools")
	if err := os.Mkdir(tools, 0700); err != nil {
		t.Fatal(err)
	}
	// Only the explicit build fails. Tool identity/version observations still
	// traverse this launcher, so the failure occurs at the recorded Cargo boundary.
	script := "#!/bin/sh\nfor arg in \"$@\"; do\n  if [ \"$arg\" = build ]; then echo 'test-only controlled build failure' >&2; exit 7; fi\ndone\nexec '" + strings.ReplaceAll(cargo, "'", "'\\''") + "' \"$@\"\n"
	if err := os.WriteFile(filepath.Join(tools, "cargo"), []byte(script), 0700); err != nil {
		t.Fatal(err)
	}
	source := filepath.Join(project, "entry.l")
	if err := os.WriteFile(source, []byte("Exit.\n"), 0600); err != nil {
		t.Fatal(err)
	}
	output := filepath.Join(project, "failed-native-evidence.json")
	stdout, stderr, buildErr := runBuildCommand(t, phase1.CLI(t), project, []string{"PATH=" + tools + string(os.PathListSeparator) + os.Getenv("PATH")}, "build", "--native-evidence-output", output, source)
	if buildErr == nil || len(stdout) != 0 || !strings.Contains(string(stderr), "test-only controlled build failure") {
		t.Fatalf("controlled Cargo failure was not retained: %v\n%s\n%s", buildErr, stdout, stderr)
	}
	data, err := os.ReadFile(output)
	if err != nil {
		t.Fatal(err)
	}
	document, err := contracts.DecodeDocument(data)
	if err != nil {
		t.Fatal(err)
	}
	environment, ok := document.(contracts.NativeEnvironment)
	if !ok || environment.Method != "build" || len(environment.Artifacts) != 0 {
		t.Fatalf("failed compilation advertised artifacts: %+v", document)
	}
	last := environment.Executions[len(environment.Executions)-1]
	if last.Kind != "build" || last.Status != "failed" || last.ExitCode == nil || *last.ExitCode != 7 {
		t.Fatalf("failed build execution missing or rewritten: %+v", last)
	}
}
