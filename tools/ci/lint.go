package main

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

func (r *runner) lint() error {
	golangci, err := r.installedTool("golangci-lint")
	if err != nil {
		return err
	}
	actionlint, err := r.installedTool("actionlint")
	if err != nil {
		return err
	}
	files, err := repositoryFiles(r.root)
	if err != nil {
		return err
	}
	formatArgs := []string{"-l"}
	for _, file := range files {
		if strings.HasSuffix(file, ".go") {
			formatArgs = append(formatArgs, file)
		}
	}
	output, err := r.command("gofmt", time.Minute, nil, "gofmt", formatArgs...)
	if err != nil {
		return err
	}
	if len(output) != 0 {
		return fmt.Errorf("unformatted Go files (gofmt without simplification):\n%s", output)
	}
	if _, err := r.command("lint-config", time.Minute, nil, golangci, "config", "verify", "--config", filepath.Join(r.root, ".golangci.yml")); err != nil {
		return err
	}
	var findings []error
	for _, profile := range []string{"ordinary", "nativecompile"} {
		args := []string{"run", "--config", filepath.Join(r.root, ".golangci.yml"), "--issues-exit-code=1", "--max-issues-per-linter=0", "--max-same-issues=0", "--output.json.path=" + filepath.Join(r.directory, "golangci-"+profile+".json")}
		if profile == "nativecompile" {
			args = append(args, "--build-tags=nativecompile")
		}
		args = append(args, "./...")
		_, err := r.command("golangci-"+profile, 9*time.Minute, map[string]string{"GOLANGCI_LINT_CACHE": filepath.Join(r.root, ".harness", "cache", "golangci-lint")}, golangci, args...)
		if err == nil {
			err = validateLintReport(filepath.Join(r.directory, "golangci-"+profile+".json"), filepath.Join(r.directory, "golangci-"+profile+".stderr.log"))
		}
		findings = append(findings, err)
	}
	// External shellcheck/pyflakes versions are not among this project's pinned
	// tools. Keep actionlint's own workflow checks identical on every host.
	output, err = r.command("actionlint", time.Minute, nil, actionlint, "-shellcheck=", "-pyflakes=")
	if err == nil {
		stderr, readErr := os.ReadFile(filepath.Join(r.directory, "actionlint.stderr.log"))
		if readErr != nil {
			err = readErr
		} else if strings.TrimSpace(string(output)) != "" || strings.TrimSpace(string(stderr)) != "" {
			err = errors.New("actionlint emitted unexpected diagnostics despite successful exit")
		}
	}
	findings = append(findings, err)
	return errors.Join(findings...)
}
