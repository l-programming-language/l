package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/l-programming-language/l/tools/internal/testgate"
)

type harnessReport struct {
	SchemaVersion    int       `json:"schema_version"`
	Command          string    `json:"command"`
	StartedAt        time.Time `json:"started_at"`
	FinishedAt       time.Time `json:"finished_at"`
	RunDirectory     string    `json:"run_directory"`
	Repository       string    `json:"repository"`
	CurrentRevision  string    `json:"current_revision"`
	GitDirty         *bool     `json:"git_dirty"`
	AcceptancePassed *bool     `json:"acceptance_passed"`
	ExitCode         *int      `json:"exit_code"`
	IntegrityBefore  struct {
		OK bool `json:"ok"`
	} `json:"integrity_before"`
	IntegrityAfter struct {
		OK bool `json:"ok"`
	} `json:"integrity_after"`
	Environment  map[string]string  `json:"environment"`
	TestPackages []testgate.Package `json:"test_packages"`
	Commands     []struct {
		Name       string           `json:"name"`
		Executable string           `json:"executable"`
		Args       []string         `json:"args"`
		TimeoutMS  int64            `json:"timeout_ms"`
		StdoutPath string           `json:"stdout_path"`
		ExitCode   *int             `json:"exit_code"`
		TimedOut   bool             `json:"timed_out"`
		Events     *testgate.Events `json:"events"`
	} `json:"commands"`
}

func decodeHarnessReport(data []byte, command, root, path string, started time.Time) (harnessReport, error) {
	var report harnessReport
	if err := json.Unmarshal(data, &report); err != nil {
		return report, fmt.Errorf("decode harness report: %w", err)
	}
	if report.SchemaVersion != 2 || report.Command != command || report.Repository != root || report.RunDirectory != filepath.Dir(path) {
		return report, errors.New("harness report identity does not match this command")
	}
	if report.StartedAt.IsZero() || report.FinishedAt.IsZero() || report.StartedAt.Before(started) || report.FinishedAt.Before(report.StartedAt) || report.FinishedAt.After(time.Now().Add(time.Second)) {
		return report, errors.New("harness report is stale or has invalid timestamps")
	}
	if report.ExitCode == nil {
		return report, errors.New("harness report has no exit status")
	}
	return report, nil
}

func validateHarnessOutcome(report harnessReport, command string) error {
	if report.ExitCode == nil || *report.ExitCode != 0 || !report.IntegrityBefore.OK || !report.IntegrityAfter.OK {
		return errors.New("harness report did not pass integrity and execution")
	}
	if command == "verify" && (report.AcceptancePassed == nil || !*report.AcceptancePassed) {
		return errors.New("harness did not establish acceptance")
	}
	if command == "doctor" && report.AcceptancePassed != nil {
		return errors.New("doctor unexpectedly claims acceptance")
	}
	if command == "verify" {
		parallel, err := strconv.Atoi(report.Environment["test_case_parallelism"])
		if len(report.TestPackages) == 0 || err != nil || parallel < 1 {
			return errors.New("acceptance report lacks package inventory or scheduling evidence")
		}
		expected := testgate.Commands(report.TestPackages, parallel)
		if len(report.Commands) != len(expected) {
			return errors.New("acceptance report lacks required build, discovery, or test phases")
		}
		named := 0
		for i, cmd := range report.Commands {
			want := expected[i]
			if cmd.Name != want.Name || cmd.Executable != "go" || !reflect.DeepEqual(cmd.Args, want.Args) || cmd.TimeoutMS != want.Timeout.Milliseconds() || cmd.ExitCode == nil || *cmd.ExitCode != 0 || cmd.TimedOut {
				return errors.New("acceptance child has invalid arguments, budget, or completion")
			}
			if i < 2 {
				continue
			}
			if err := testgate.CheckPhase(report.TestPackages, cmd.Name, cmd.Events); err != nil {
				return err
			}
			named += len(cmd.Events.Tests)
		}
		if named == 0 {
			return errors.New("acceptance report lacks named outcomes")
		}
	} else if len(report.Commands) != 0 {
		return errors.New("doctor unexpectedly ran compiler commands")
	}
	return nil
}

func (r *runner) acceptance() error {
	for _, command := range []string{"doctor", "verify"} {
		started := time.Now().UTC()
		timeout := 7 * time.Minute
		if command == "verify" {
			timeout = 26 * time.Minute
		}
		output, commandErr := r.command(command, timeout, nil, filepath.Join(r.root, "scripts", "l-harness"), command)
		// Failure reports are evidence too. Retention cannot replace the original
		// process error or turn a failed harness outcome into success.
		if err := errors.Join(commandErr, r.retainHarnessReport(output, command, started)); err != nil {
			return err
		}
	}
	return nil
}

func (r *runner) retainHarnessReport(output []byte, command string, started time.Time) error {
	var path string
	for _, line := range strings.Split(string(output), "\n") {
		if strings.HasPrefix(line, "report: ") {
			if path != "" {
				return errors.New("harness printed more than one report")
			}
			path = strings.TrimPrefix(line, "report: ")
		}
	}
	if path == "" {
		return errors.New("harness printed no report")
	}
	runs := filepath.Join(r.root, ".harness", "runs")
	relative, err := filepath.Rel(runs, path)
	if err != nil || !filepath.IsAbs(path) || filepath.Clean(path) != path || !filepath.IsLocal(relative) || filepath.Base(path) != "report.json" || filepath.Dir(relative) == "." || filepath.Dir(filepath.Dir(relative)) != "." {
		return errors.New("harness report is not a canonical report in this checkout's run directory")
	}
	resolvedRuns, err := filepath.EvalSymlinks(runs)
	if err != nil {
		return fmt.Errorf("resolve harness run directory: %w", err)
	}
	resolvedPath, err := filepath.EvalSymlinks(path)
	if err != nil {
		return fmt.Errorf("resolve harness report: %w", err)
	}
	if resolvedPath != filepath.Join(resolvedRuns, relative) {
		return errors.New("harness report resolves outside its canonical run path")
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("read harness report: %w", err)
	}
	report, err := decodeHarnessReport(data, command, r.root, path, started)
	if err != nil {
		return err
	}
	if report.CurrentRevision != r.report.Revision || report.GitDirty == nil || *report.GitDirty != r.report.Dirty {
		return errors.New("harness report revision or dirty state differs from this stage")
	}
	writeErr := os.WriteFile(filepath.Join(r.directory, command+".report.json"), data, 0644)
	if writeErr != nil {
		writeErr = fmt.Errorf("retain harness report: %w", writeErr)
	}
	outcomeErr := validateHarnessOutcome(report, command)
	if outcomeErr == nil && command == "verify" {
		outcomeErr = validateHarnessDiscovery(report, filepath.Dir(resolvedPath))
	}
	return errors.Join(outcomeErr, writeErr)
}

func validateHarnessDiscovery(report harnessReport, resolvedDirectory string) error {
	path := report.Commands[1].StdoutPath
	resolved, err := filepath.EvalSymlinks(path)
	if err != nil || path != filepath.Join(report.RunDirectory, "packages.stdout.log") || resolved != filepath.Join(resolvedDirectory, "packages.stdout.log") {
		return errors.New("harness discovery artifact is missing or outside its canonical run path")
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("read harness discovery: %w", err)
	}
	discovered, err := testgate.ReadPackages(bytes.NewReader(data))
	if err != nil {
		return fmt.Errorf("decode harness discovery: %w", err)
	}
	inventory := make(map[string]bool, len(report.TestPackages))
	for _, pkg := range report.TestPackages {
		inventory[pkg.Package] = *pkg.HasTests
	}
	if !reflect.DeepEqual(discovered, inventory) {
		return errors.New("harness discovery and package inventory differ")
	}
	return nil
}

type nativeEvent struct{ Action, Package, Test string }

func nativeOutcomes(data []byte) (map[string]bool, error) {
	decoder := json.NewDecoder(bytes.NewReader(data))
	running := map[string]bool{}
	passed := map[string]bool{}
	packageStarted, packagePassed := false, false
	known := map[string]bool{"start": true, "run": true, "pause": true, "cont": true, "pass": true, "output": true, "bench": true, "skip": true, "fail": true, "build-output": true, "build-fail": true}
	for {
		var event nativeEvent
		err := decoder.Decode(&event)
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("invalid native test event stream: %w", err)
		}
		if event.Package != "github.com/l-programming-language/l/logiceme" || !known[event.Action] || packagePassed {
			return nil, errors.New("native event has unexpected package/action or follows package completion")
		}
		if event.Action == "start" {
			if packageStarted || event.Test != "" {
				return nil, errors.New("invalid native package start")
			}
			packageStarted = true
			continue
		}
		if !packageStarted {
			return nil, errors.New("native event precedes package start")
		}
		if event.Action == "fail" || event.Action == "build-fail" || event.Action == "skip" {
			return nil, fmt.Errorf("native outcome %s: %s/%s", event.Action, event.Package, event.Test)
		}
		if event.Test != "" {
			parts := strings.Split(event.Test, "/")
			if parts[0] != "TestLucidNativeCompile" && parts[0] != "TestNativeMetadataArrowsCompile" {
				return nil, fmt.Errorf("unexpected native root %s", parts[0])
			}
			if len(parts) > 3 || len(parts) > 1 && (parts[1] != "Expected" && parts[1] != "Generated") || parts[0] == "TestNativeMetadataArrowsCompile" && len(parts) > 2 {
				return nil, fmt.Errorf("unexpected native phase/case %s", event.Test)
			}
			switch event.Action {
			case "run":
				if running[event.Test] || passed[event.Test] {
					return nil, fmt.Errorf("duplicate native test %s", event.Test)
				}
				if len(parts) > 1 && !running[strings.Join(parts[:len(parts)-1], "/")] {
					return nil, fmt.Errorf("native child ran without parent: %s", event.Test)
				}
				running[event.Test] = true
			case "pass":
				if !running[event.Test] {
					return nil, fmt.Errorf("native pass without run: %s", event.Test)
				}
				for test := range running {
					if strings.HasPrefix(test, event.Test+"/") {
						return nil, fmt.Errorf("native parent passed before child: %s", test)
					}
				}
				delete(running, event.Test)
				passed[event.Test] = true
			}
		} else if event.Action == "pass" {
			if len(running) != 0 {
				return nil, errors.New("native package passed before its tests completed")
			}
			packagePassed = true
		}
	}
	for _, root := range []string{"TestLucidNativeCompile", "TestNativeMetadataArrowsCompile"} {
		for _, suffix := range []string{"", "/Expected", "/Generated"} {
			if !passed[root+suffix] {
				return nil, errors.New("native stage is incomplete or requested phases did not run")
			}
		}
	}
	if !packagePassed || len(running) != 0 {
		return nil, errors.New("native stage did not complete")
	}
	expected, generated := map[string]bool{}, map[string]bool{}
	for name := range passed {
		parts := strings.Split(name, "/")
		if len(parts) == 3 {
			if parts[1] == "Expected" {
				expected[parts[2]] = true
			} else {
				generated[parts[2]] = true
			}
		}
	}
	if len(expected) == 0 || !reflect.DeepEqual(expected, generated) {
		return nil, errors.New("native Expected/Generated case sets differ or are empty")
	}
	return passed, nil
}

func validateNativeRecords(directory string, passed map[string]bool) error {
	count := 0
	for name := range passed {
		parts := strings.Split(name, "/")
		if len(parts) != 3 || parts[0] != "TestLucidNativeCompile" {
			continue
		}
		if parts[1] != "Expected" && parts[1] != "Generated" {
			return fmt.Errorf("unexpected native phase %s", name)
		}
		if parts[2] == "" || filepath.Base(parts[2]) != parts[2] {
			return errors.New("invalid native scenario name")
		}
		path := filepath.Join(directory, "compile-"+strings.ToLower(parts[1]), parts[2], "result.json")
		var record struct {
			Scenario string   `json:"scenario"`
			Stage    string   `json:"stage"`
			Status   string   `json:"status"`
			Issues   []string `json:"issues"`
			ExitCode *int     `json:"exit_code"`
		}
		if err := readJSON(path, &record); err != nil {
			return err
		}
		if record.Scenario != parts[2] || record.Stage != "compile-"+strings.ToLower(parts[1]) || record.Status != "pass" || len(record.Issues) != 0 || record.ExitCode == nil {
			return fmt.Errorf("native artifact record did not pass: %s", path)
		}
		count++
	}
	if count == 0 {
		return errors.New("native stage produced no corpus records")
	}
	return nil
}

func (r *runner) native() error {
	artifacts := filepath.Join(r.directory, "artifacts")
	output, err := r.command("nativecompile", 31*time.Minute, map[string]string{"L_LUCID_ARTIFACT_DIR": artifacts, "L_HARNESS_EVIDENCE_DIR": filepath.Join(r.directory, "bundles")}, "go", "test", "-mod=readonly", "-count=1", "-tags=nativecompile", "-run", "^(TestLucidNativeCompile|TestNativeMetadataArrowsCompile)$", "-timeout=30m", "-json", "./logiceme")
	if err != nil {
		return err
	}
	passed, err := nativeOutcomes(output)
	if err != nil {
		return err
	}
	return validateNativeRecords(artifacts, passed)
}

func validateLintReport(path, stderrPath string) error {
	var report struct {
		Issues []json.RawMessage `json:"Issues"`
		Report *struct {
			Warnings []json.RawMessage `json:"Warnings"`
			Errors   []json.RawMessage `json:"Errors"`
		} `json:"Report"`
	}
	if err := readJSON(path, &report); err != nil {
		return err
	}
	if report.Report == nil {
		return errors.New("golangci-lint JSON is missing its report")
	}
	if len(report.Issues) != 0 || len(report.Report.Warnings) != 0 || len(report.Report.Errors) != 0 {
		return fmt.Errorf("golangci-lint reported issues, warnings or errors: %s", path)
	}
	stderr, err := os.ReadFile(stderrPath)
	if err != nil {
		return err
	}
	if strings.TrimSpace(string(stderr)) != "" {
		return fmt.Errorf("golangci-lint emitted diagnostic stderr: %s", stderrPath)
	}
	return nil
}
