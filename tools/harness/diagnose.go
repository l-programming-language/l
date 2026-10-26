package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/l-programming-language/l/tools/internal/testgate"
)

func runFrontendDiagnostics(ctx context.Context, root, runDir string) ([]commandResult, error) {
	var results []commandResult
	for _, probe := range []struct {
		name, template, pkg, test string
	}{
		{"compiler", "compiler.go.txt", "cmd/l", "TestHarnessCompilerProbe"},
		{"imports", "imports.go.txt", "lyger", "TestHarnessImportProbe"},
	} {
		artifactDir := filepath.Join(runDir, "frontend", probe.name)
		if err := os.MkdirAll(artifactDir, 0755); err != nil {
			return results, err
		}
		virtualPath := filepath.Join(root, filepath.FromSlash(probe.pkg), "zz_harness_probe_test.go")
		if _, err := os.Lstat(virtualPath); !os.IsNotExist(err) {
			return results, fmt.Errorf("refusing to overlay existing path %s", virtualPath)
		}
		content, err := os.ReadFile(filepath.Join(root, "tools", "harness", "probes", probe.template))
		if err != nil {
			return results, err
		}
		probePath := filepath.Join(artifactDir, "probe_test.go.txt")
		if err = os.WriteFile(probePath, content, 0644); err != nil {
			return results, err
		}
		overlayPath := filepath.Join(artifactDir, "overlay.json")
		overlay := struct {
			Replace map[string]string `json:"Replace"`
		}{Replace: map[string]string{virtualPath: probePath}}
		if err = testgate.WriteJSON(overlayPath, overlay); err != nil {
			return results, err
		}
		spec := commandSpec{
			Name: "frontend-" + probe.name, Executable: "go", Dir: root,
			Args: []string{"test", "-mod=readonly", "-overlay=" + overlayPath, "-count=1", "-timeout=60s", "-json", "-run", "^" + probe.test + "$", "./" + probe.pkg},
			Env:  harnessEnvironment(root, "L_HARNESS_ARTIFACT_DIR="+artifactDir, "L_HARNESS_RUN_ID="+filepath.Base(runDir)), Timeout: 180 * time.Second, TestJSON: true,
		}
		result := runCommand(ctx, runDir, spec)
		validateProbeResult(&result, artifactDir, probe.test)
		results = append(results, result)
	}
	return results, nil
}

func validateProbeResult(result *commandResult, artifactDir, expectedTest string) {
	addError := func(message string) {
		if result.Error != "" {
			result.Error += "; "
		}
		result.Error += message
	}
	found, passed := false, false
	if result.Events != nil {
		for _, test := range result.Events.Tests {
			if test.Name == expectedTest {
				found, passed = true, test.Status == "pass"
			}
		}
	}
	if !found {
		addError("expected diagnostic test did not execute: " + expectedTest)
	} else if !passed && result.ExitCode == 0 {
		addError("diagnostic test did not pass: " + expectedTest)
	}
	data, err := os.ReadFile(filepath.Join(artifactDir, "report.json"))
	if err != nil {
		addError("diagnostic report missing: " + err.Error())
	} else if !json.Valid(data) {
		addError("diagnostic report is not valid JSON")
	}
}
