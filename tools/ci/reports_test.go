package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/l-programming-language/l/tools/internal/testgate"
)

func testJSON(t *testing.T, value any) []byte {
	t.Helper()
	data, err := json.Marshal(value)
	if err != nil {
		t.Fatal(err)
	}
	return data
}
func testWrite(t *testing.T, path string, data []byte) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, data, 0644); err != nil {
		t.Fatal(err)
	}
}

func validateHarness(data []byte, command, root, path string, started time.Time) error {
	report, err := decodeHarnessReport(data, command, root, path, started)
	if err != nil {
		return err
	}
	return validateHarnessOutcome(report, command)
}

func successfulHarnessEvidence(report map[string]any, directory string) {
	report["environment"] = map[string]string{"test_case_parallelism": "4"}
	report["test_packages"] = []any{
		map[string]any{"package": "fixture", "phase": "tests", "has_tests": true},
		map[string]any{"package": "fixture/empty", "phase": "tests", "has_tests": false},
		map[string]any{"package": "fixture/helper", "phase": "tests", "has_tests": true},
		map[string]any{"package": "fixture/cmd/l", "phase": "tests-native", "has_tests": true},
	}
	commands := []any{
		map[string]any{"name": "build", "executable": "go", "args": []string{"build", "-mod=readonly", "./..."}, "timeout_ms": 180000, "exit_code": 0},
		map[string]any{"name": "packages", "stdout_path": filepath.Join(directory, "packages.stdout.log"), "executable": "go", "args": []string{"list", "-mod=readonly", "-json", "./..."}, "timeout_ms": 60000, "exit_code": 0},
	}
	for _, phase := range []string{"tests", "tests-native"} {
		args := []string{"test", "-p=3", "-parallel=4", "-mod=readonly", "-count=1", "-timeout=120s", "-json", "fixture", "fixture/empty", "fixture/helper"}
		packages := []testgate.Outcome{{Package: "fixture", Status: "pass"}, {Package: "fixture/empty", Status: "skip"}, {Package: "fixture/helper", Status: "pass"}}
		timeout := 360000
		if phase == "tests-native" {
			args[5], args, timeout = "-timeout=600s", args[:8], 900000
			args[7] = "fixture/cmd/l"
			packages = []testgate.Outcome{{Package: "fixture/cmd/l", Status: "pass"}}
		}
		commands = append(commands, map[string]any{"name": phase, "executable": "go", "args": args, "timeout_ms": timeout, "exit_code": 0, "events": map[string]any{"packages": packages, "tests": []testgate.Outcome{{Package: packages[0].Package, Name: "TestExample", Status: "pass"}}}})
	}
	report["commands"] = commands
}

func writeHarnessDiscovery(t *testing.T, directory string) string {
	t.Helper()
	path := filepath.Join(directory, "packages.stdout.log")
	testWrite(t, path, []byte(`{"ImportPath":"fixture","TestGoFiles":["unit_test.go"]}
{"ImportPath":"fixture/empty"}
{"ImportPath":"fixture/helper","TestGoFiles":["helper_test.go"]}
{"ImportPath":"fixture/cmd/l","XTestGoFiles":["native_test.go"]}`))
	return path
}

func TestHarnessRetentionRequiresCompleteDiscovery(t *testing.T) {
	for _, mode := range []string{"success", "omitted-native", "wrong-metadata", "missing", "malformed", "duplicate", "wrong-path", "symlink"} {
		t.Run(mode, func(t *testing.T) {
			r := runner{root: t.TempDir(), directory: t.TempDir(), report: stageReport{Revision: "fixture"}}
			directory := filepath.Join(r.root, ".harness", "runs", "fresh")
			path, started := filepath.Join(directory, "report.json"), time.Now()
			base := map[string]any{"schema_version": 2, "command": "verify", "started_at": started, "finished_at": time.Now(), "repository": r.root, "run_directory": directory, "current_revision": "fixture", "git_dirty": false, "acceptance_passed": true, "exit_code": 0, "integrity_before": map[string]bool{"ok": true}, "integrity_after": map[string]bool{"ok": true}}
			successfulHarnessEvidence(base, directory)
			report, err := decodeHarnessReport(testJSON(t, base), "verify", r.root, path, started)
			if err != nil {
				t.Fatal(err)
			}
			discovery := writeHarnessDiscovery(t, directory)
			switch mode {
			case "omitted-native":
				report.TestPackages, report.Commands = report.TestPackages[:3], report.Commands[:3]
			case "wrong-metadata":
				*report.TestPackages[2].HasTests = false
			case "missing":
				if err := os.Remove(discovery); err != nil {
					t.Fatal(err)
				}
			case "malformed":
				testWrite(t, discovery, []byte(`{"ImportPath":`))
			case "duplicate":
				testWrite(t, discovery, []byte(`{"ImportPath":"fixture","TestGoFiles":["unit_test.go"]} {"ImportPath":"fixture","TestGoFiles":["unit_test.go"]}`))
			case "wrong-path", "symlink":
				outside := writeHarnessDiscovery(t, t.TempDir())
				if mode == "wrong-path" {
					report.Commands[1].StdoutPath = outside
				} else {
					if err := os.Remove(discovery); err != nil {
						t.Fatal(err)
					}
					if err := os.Symlink(outside, discovery); err != nil {
						t.Fatal(err)
					}
				}
			}
			if err := validateHarnessOutcome(report, "verify"); err != nil {
				t.Fatal("fixture should otherwise be self-consistent", err)
			}
			testWrite(t, path, testJSON(t, report))
			err = r.retainHarnessReport([]byte("report: "+path), "verify", started)
			if (err == nil) != (mode == "success") {
				t.Fatalf("discovery validation: %v", err)
			}
			if _, err := os.ReadFile(filepath.Join(r.directory, "verify.report.json")); err != nil {
				t.Fatal("lost failure evidence", err)
			}
		})
	}
}

func TestHarnessReportRejectsStaleOrIncompleteSuccess(t *testing.T) {
	root := t.TempDir()
	path := filepath.Join(root, ".harness", "runs", "fresh", "report.json")
	started := time.Now().Add(-time.Second)
	finished := time.Now()
	base := map[string]any{"schema_version": 2, "command": "verify", "started_at": started, "finished_at": finished, "repository": root, "run_directory": filepath.Dir(path), "acceptance_passed": true, "exit_code": 0, "integrity_before": map[string]bool{"ok": true}, "integrity_after": map[string]bool{"ok": true}}
	successfulHarnessEvidence(base, filepath.Dir(path))
	if err := validateHarness(testJSON(t, base), "verify", root, path, started); err != nil {
		t.Fatal(err)
	}
	for name, change := range map[string]map[string]any{"stale": {"started_at": started.Add(-time.Hour)}, "wrong-command": {"command": "doctor"}, "wrong-repository": {"repository": root + "-other"}, "wrong-report-path": {"run_directory": root}, "integrity": {"integrity_after": map[string]bool{"ok": false}}, "no-acceptance": {"acceptance_passed": false}, "no-exit": {"exit_code": nil}, "failed-exit": {"exit_code": 1}, "bad-schema": {"schema_version": 1}, "unfinished": {"finished_at": time.Time{}}} {
		t.Run(name, func(t *testing.T) {
			candidate := map[string]any{}
			for k, v := range base {
				candidate[k] = v
			}
			for k, v := range change {
				candidate[k] = v
			}
			if err := validateHarness(testJSON(t, candidate), "verify", root, path, started); err == nil {
				t.Fatal("invalid harness success accepted")
			}
		})
	}
	for name, change := range map[string]func(*harnessReport){
		"missing-inventory":      func(r *harnessReport) { r.TestPackages = nil },
		"duplicate-inventory":    func(r *harnessReport) { r.TestPackages = append(r.TestPackages, r.TestPackages[0]) },
		"missing-test-metadata":  func(r *harnessReport) { r.TestPackages[0].HasTests = nil },
		"unknown-phase":          func(r *harnessReport) { r.TestPackages[0].Phase = "other" },
		"wrong-phase":            func(r *harnessReport) { r.TestPackages[0].Phase = "tests-native" },
		"omitted-phase":          func(r *harnessReport) { r.Commands = r.Commands[:3] },
		"empty-phase":            func(r *harnessReport) { r.Commands[3].Events = nil },
		"missing-named-evidence": func(r *harnessReport) { r.Commands[3].Events.Tests = nil },
		"no-named-outcomes": func(r *harnessReport) {
			r.Commands[2].Events.Tests, r.Commands[3].Events.Tests = []testgate.Outcome{}, []testgate.Outcome{}
		},
		"wrong-executable":           func(r *harnessReport) { r.Commands[1].Executable = "other" },
		"parse-error":                func(r *harnessReport) { r.Commands[3].Events.ParseError = "truncated JSON" },
		"failed-phase":               func(r *harnessReport) { *r.Commands[3].ExitCode = 1 },
		"timed-out":                  func(r *harnessReport) { r.Commands[3].TimedOut = true },
		"wrong-budget":               func(r *harnessReport) { r.Commands[3].TimeoutMS = 360000 },
		"filtered":                   func(r *harnessReport) { r.Commands[3].Args = append(r.Commands[3].Args, "-run=TestExample") },
		"wrong-discovery":            func(r *harnessReport) { r.Commands[1].Args[3] = "./cmd/l" },
		"wrong-concurrency":          func(r *harnessReport) { r.Commands[2].Args[1] = "-p=4" },
		"missing-package":            func(r *harnessReport) { r.Commands[2].Events.Packages = r.Commands[2].Events.Packages[:2] },
		"duplicate-package":          func(r *harnessReport) { r.Commands[2].Events.Packages[1] = r.Commands[2].Events.Packages[0] },
		"wrong-package":              func(r *harnessReport) { r.Commands[2].Events.Packages[0].Package = "fixture/cmd/l" },
		"failed-package":             func(r *harnessReport) { r.Commands[3].Events.Packages[0].Status = "fail" },
		"skipped-package-with-tests": func(r *harnessReport) { r.Commands[3].Events.Packages[0].Status = "skip" },
		"skipped-test":               func(r *harnessReport) { r.Commands[3].Events.Tests[0].Status = "skip" },
		"aborted-test":               func(r *harnessReport) { r.Commands[3].Events.Tests[0].Status = "aborted" },
		"wrong-test-package":         func(r *harnessReport) { r.Commands[3].Events.Tests[0].Package = "fixture" },
		"unnamed-test":               func(r *harnessReport) { r.Commands[3].Events.Tests[0].Name = "" },
		"duplicate-test": func(r *harnessReport) {
			r.Commands[3].Events.Tests = append(r.Commands[3].Events.Tests, r.Commands[3].Events.Tests[0])
		},
	} {
		t.Run(name, func(t *testing.T) {
			report, err := decodeHarnessReport(testJSON(t, base), "verify", root, path, started)
			if err != nil {
				t.Fatal(err)
			}
			change(&report)
			if err := validateHarnessOutcome(report, "verify"); err == nil {
				t.Fatal("invalid harness evidence accepted")
			}
		})
	}
	for _, phase := range []string{"tests", "tests-native"} {
		t.Run("empty-"+phase, func(t *testing.T) {
			report, err := decodeHarnessReport(testJSON(t, base), "verify", root, path, started)
			if err != nil {
				t.Fatal(err)
			}
			for i := len(report.TestPackages) - 1; i >= 0; i-- {
				if report.TestPackages[i].Phase == phase {
					report.TestPackages = append(report.TestPackages[:i], report.TestPackages[i+1:]...)
				}
			}
			index := 2
			if phase == "tests-native" {
				index = 3
			}
			report.Commands = append(report.Commands[:index], report.Commands[index+1:]...)
			if err := validateHarnessOutcome(report, "verify"); err != nil {
				t.Fatal(err)
			}
		})
	}
}

func goodNativeEvents(t *testing.T) []byte {
	t.Helper()
	var events []nativeEvent
	add := func(action, test string) {
		events = append(events, nativeEvent{Action: action, Package: "github.com/l-programming-language/l/logiceme", Test: test})
	}
	add("start", "")
	for _, root := range []string{"TestLucidNativeCompile", "TestNativeMetadataArrowsCompile"} {
		add("run", root)
		for _, phase := range []string{"Expected", "Generated"} {
			parent := root + "/" + phase
			add("run", parent)
			if root == "TestLucidNativeCompile" {
				add("run", parent+"/intentional-negative")
				add("pass", parent+"/intentional-negative")
			}
			add("pass", parent)
		}
		add("pass", root)
	}
	add("pass", "")
	var data []byte
	for _, event := range events {
		data = append(data, testJSON(t, event)...)
		data = append(data, '\n')
	}
	return data
}

func TestNativeEventsRequireRealCompletePairedRoots(t *testing.T) {
	good := goodNativeEvents(t)
	if _, err := nativeOutcomes(good); err != nil {
		t.Fatal(err)
	}
	for name, data := range map[string][]byte{
		"empty": nil, "truncated": good[:len(good)-15], "malformed": []byte("not json"),
		"wrong-package":          []byte(strings.ReplaceAll(string(good), "github.com/l-programming-language/l/logiceme", "example.test/other")),
		"unknown-action":         []byte(strings.Replace(string(good), `"Action":"start"`, `"Action":"invented"`, 1)),
		"skip":                   []byte(strings.Replace(string(good), `"Action":"pass"`, `"Action":"skip"`, 1)),
		"fail":                   []byte(strings.Replace(string(good), `"Action":"pass"`, `"Action":"fail"`, 1)),
		"missing-root":           []byte(strings.ReplaceAll(string(good), "TestNativeMetadataArrowsCompile", "UnrequestedTest")),
		"mismatched-cases":       []byte(strings.ReplaceAll(string(good), "Generated/intentional-negative", "Generated/different")),
		"extra-after-completion": append(append([]byte{}, good...), testJSON(t, nativeEvent{Action: "output", Package: "github.com/l-programming-language/l/logiceme"})...),
	} {
		t.Run(name, func(t *testing.T) {
			if _, err := nativeOutcomes(data); err == nil {
				t.Fatal("incomplete or invalid native stream accepted")
			}
		})
	}
}

func TestNativeParentCannotPassBeforeChild(t *testing.T) {
	good := string(goodNativeEvents(t))
	childPass := string(testJSON(t, nativeEvent{Action: "pass", Package: "github.com/l-programming-language/l/logiceme", Test: "TestLucidNativeCompile/Expected/intentional-negative"})) + "\n"
	parentPass := string(testJSON(t, nativeEvent{Action: "pass", Package: "github.com/l-programming-language/l/logiceme", Test: "TestLucidNativeCompile/Expected"})) + "\n"
	outOfOrder := strings.Replace(good, childPass+parentPass, parentPass+childPass, 1)
	if outOfOrder == good {
		t.Fatal("ordering fixture was not changed")
	}
	if _, err := nativeOutcomes([]byte(outOfOrder)); err == nil {
		t.Fatal("native parent completed before its child")
	}
}

func TestNativeRecordsAcceptExpectedCompilerRejectionsButNotMissingEvidence(t *testing.T) {
	root := t.TempDir()
	passed, err := nativeOutcomes(goodNativeEvents(t))
	if err != nil {
		t.Fatal(err)
	}
	for _, phase := range []string{"expected", "generated"} {
		testWrite(t, filepath.Join(root, "compile-"+phase, "intentional-negative", "result.json"), testJSON(t, map[string]any{"scenario": "intentional-negative", "stage": "compile-" + phase, "status": "pass", "exit_code": 101}))
	}
	if err := validateNativeRecords(root, passed); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(root, "compile-generated", "intentional-negative", "result.json")
	testWrite(t, path, testJSON(t, map[string]any{"scenario": "intentional-negative", "stage": "compile-generated", "status": "blocked", "exit_code": 101}))
	if err := validateNativeRecords(root, passed); err == nil {
		t.Fatal("blocked generation accepted")
	}
	if err := os.Remove(path); err != nil {
		t.Fatal(err)
	}
	if err := validateNativeRecords(root, passed); err == nil {
		t.Fatal("missing generation record accepted")
	}
}

func TestLintReportRejectsWarningsDespiteSuccessfulToolExit(t *testing.T) {
	root := t.TempDir()
	path := filepath.Join(root, "lint.json")
	stderr := filepath.Join(root, "stderr.log")
	testWrite(t, stderr, nil)
	testWrite(t, path, []byte(`{"Issues":[],"Report":{"Warnings":[],"Errors":[]}}`))
	if err := validateLintReport(path, stderr); err != nil {
		t.Fatal(err)
	}
	for _, data := range []string{`{}`, `{"Report":{"Warnings":["incomplete analysis"]}}`, `{"Report":{"Errors":["failure"]}}`, `{"Issues":[{}],"Report":{}}`, `{bad`} {
		testWrite(t, path, []byte(data))
		if err := validateLintReport(path, stderr); err == nil {
			t.Fatalf("accepted %s", data)
		}
	}
	testWrite(t, path, []byte(`{"Report":{}}`))
	testWrite(t, stderr, []byte("level=warning incomplete\n"))
	if err := validateLintReport(path, stderr); err == nil {
		t.Fatal("warning stderr accepted")
	}
}
