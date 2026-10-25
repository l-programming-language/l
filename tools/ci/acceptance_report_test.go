package main

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// Exercise report handling through a real failing subprocess without running the
// compiler or the protected acceptance suite.
func TestAcceptanceRetainsOnlyCurrentCanonicalFailureReports(t *testing.T) {
	for _, tc := range []struct {
		name   string
		retain bool
	}{
		{"failed", true}, {"success-report-failed-child", true},
		{"malformed", false}, {"stale", false}, {"zero-start", false}, {"unfinished", false}, {"future-finish", false},
		{"wrong-command", false}, {"wrong-root", false}, {"wrong-revision", false}, {"wrong-dirty", false},
		{"wrong-directory", false}, {"outside", false}, {"symlink", false}, {"wrong-filename", false},
		{"duplicate", false}, {"missing", false}, {"missing-exit", false}, {"no-report", false},
		{"write-error", false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			r := acceptanceFixtureRunner(t, tc.name)
			retained := filepath.Join(r.directory, "verify.report.json")
			if tc.name == "write-error" {
				if err := os.Mkdir(retained, 0755); err != nil {
					t.Fatal(err)
				}
			}
			err := r.acceptance()
			var child *exec.ExitError
			if !errors.As(err, &child) || child.ExitCode() != 7 {
				t.Fatalf("lost subprocess failure: %v", err)
			}
			if len(r.report.Commands) != 2 || r.report.Commands[0].TimeoutMS != 420000 || r.report.Commands[1].TimeoutMS != 1560000 {
				t.Fatalf("incorrect doctor/verify wrapper budgets: %+v", r.report.Commands)
			}
			if _, err := os.ReadFile(filepath.Join(r.directory, "doctor.report.json")); err != nil {
				t.Fatalf("successful doctor report was not retained: %v", err)
			}
			got, readErr := os.ReadFile(retained)
			if tc.retain {
				original, originalErr := os.ReadFile(filepath.Join(r.root, "verify-original.json"))
				if readErr != nil || originalErr != nil || !bytes.Equal(got, original) {
					t.Fatalf("failed report bytes were not preserved: %v, %v", readErr, originalErr)
				}
			} else if readErr == nil {
				t.Fatal("invalid or unwritable report was retained")
			}
			if tc.name != "success-report-failed-child" && !strings.Contains(err.Error(), "harness") {
				t.Fatalf("report validation/write failure was lost: %v", err)
			}
			if tc.name == "write-error" && !strings.Contains(err.Error(), retained) {
				t.Fatalf("retention write error was lost: %v", err)
			}
		})
	}
}

func TestAcceptanceRetainsFailedDoctorAndStopsBeforeVerify(t *testing.T) {
	r := acceptanceFixtureRunner(t, "doctor-failure")
	err := r.acceptance()
	var child *exec.ExitError
	if !errors.As(err, &child) || child.ExitCode() != 7 || len(r.report.Commands) != 1 {
		t.Fatalf("failed doctor did not stop acceptance: %v %+v", err, r.report.Commands)
	}
	if _, err := os.ReadFile(filepath.Join(r.directory, "doctor.report.json")); err != nil {
		t.Fatal("failed doctor evidence was lost", err)
	}
	if _, err := os.Stat(filepath.Join(r.directory, "verify.report.json")); !errors.Is(err, os.ErrNotExist) {
		t.Fatal("verify evidence exists after a failed doctor", err)
	}
}

func acceptanceFixtureRunner(t *testing.T, mode string) runner {
	t.Helper()
	root := t.TempDir()
	executable, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(root, "scripts", "l-harness")
	script := "#!/bin/sh\nexec '" + strings.ReplaceAll(executable, "'", "'\"'\"'") + "' -test.run='^TestAcceptanceHarnessFixture$' -- \"$@\"\n"
	testWrite(t, path, []byte(script))
	if err := os.Chmod(path, 0755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("L_CI_REPORT_FIXTURE_ROOT", root)
	t.Setenv("L_CI_REPORT_FIXTURE_MODE", mode)
	return runner{root: root, directory: t.TempDir(), report: stageReport{Revision: "fixture-revision"}}
}

func TestAcceptanceHarnessFixture(t *testing.T) {
	root := os.Getenv("L_CI_REPORT_FIXTURE_ROOT")
	if root == "" {
		return
	}
	command := os.Args[len(os.Args)-1]
	mode := os.Getenv("L_CI_REPORT_FIXTURE_MODE")
	started := time.Now().UTC()
	directory := filepath.Join(root, ".harness", "runs", command+"-fresh")
	path := filepath.Join(directory, "report.json")
	report := map[string]any{
		"schema_version": 2, "command": command, "repository": root, "run_directory": directory,
		"current_revision": "fixture-revision", "git_dirty": false,
		"started_at": started, "finished_at": time.Now().UTC(), "exit_code": 0,
		"integrity_before": map[string]bool{"ok": true}, "integrity_after": map[string]bool{"ok": true},
		"commands": []any{},
	}
	exitCode := 0
	if command == "verify" {
		exitCode = 7
		report["exit_code"], report["acceptance_passed"] = 7, false
		report["commands"] = []any{map[string]any{"name": "build", "exit_code": 0}, map[string]any{"name": "tests", "exit_code": 1, "events": map[string]any{"tests": []any{map[string]string{"status": "fail"}}}}}
		switch mode {
		case "success-report-failed-child":
			report["exit_code"], report["acceptance_passed"] = 0, true
			successfulHarnessEvidence(report, directory)
			writeHarnessDiscovery(t, directory)
		case "stale":
			report["started_at"] = started.Add(-time.Hour)
		case "zero-start":
			report["started_at"] = time.Time{}
		case "unfinished":
			report["finished_at"] = time.Time{}
		case "future-finish":
			report["finished_at"] = started.Add(time.Hour)
		case "wrong-command":
			report["command"] = "doctor"
		case "wrong-root":
			report["repository"] = root + "-other"
		case "wrong-revision":
			report["current_revision"] = "other-revision"
		case "wrong-dirty":
			report["git_dirty"] = true
		case "wrong-directory":
			report["run_directory"] = root
		case "outside":
			path = filepath.Join(root, "outside", "report.json")
			report["run_directory"] = filepath.Dir(path)
		case "wrong-filename":
			path = filepath.Join(directory, "other.json")
		case "missing-exit":
			delete(report, "exit_code")
		}
	} else if mode == "doctor-failure" {
		report["exit_code"], exitCode = 7, 7
	}
	data := testJSON(t, report)
	if command == "verify" && mode == "malformed" {
		data = []byte("{invalid")
	}
	testWrite(t, path, data)
	testWrite(t, filepath.Join(root, command+"-original.json"), data)
	if command == "verify" {
		switch mode {
		case "symlink":
			outside := filepath.Join(root, "outside", "report.json")
			testWrite(t, outside, data)
			if err := os.Remove(path); err != nil {
				t.Fatal(err)
			}
			if err := os.Symlink(outside, path); err != nil {
				t.Fatal(err)
			}
		case "missing":
			if err := os.Remove(path); err != nil {
				t.Fatal(err)
			}
		case "no-report":
			os.Exit(exitCode)
		case "duplicate":
			if _, err := fmt.Printf("report: %s\n", path); err != nil {
				t.Fatal(err)
			}
		}
	}
	if _, err := fmt.Printf("report: %s\n", path); err != nil {
		t.Fatal(err)
	}
	os.Exit(exitCode)
}
