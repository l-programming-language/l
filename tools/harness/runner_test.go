package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"
	"testing"
	"testing/iotest"
	"time"

	"github.com/l-programming-language/l/tools/internal/testgate"
)

func TestPackageDiscoveryAndCompletion(t *testing.T) {
	const module = "example.test/l"
	var input bytes.Buffer
	for _, path := range []string{"/cmd/l", "/lyger", "/logiceme", "/internal/cargocache", "/internal/testsupport/native", "/acceptance/future", "/acceptance", "/future", ""} {
		if _, err := fmt.Fprintf(&input, `{"ImportPath":%q,"XTestGoFiles":["helpers_test.go"]}`, module+path); err != nil {
			t.Fatal(err)
		}
	}
	packages, err := discoverTestPackages(&input, module)
	if err != nil || len(packages) != 9 {
		t.Fatalf("discovery: %v, %+v", err, packages)
	}
	for i, pkg := range packages {
		want := "tests-native"
		if pkg.Package == module || pkg.Package == module+"/future" {
			want = "tests"
		}
		if pkg.Phase != want || !(*pkg.HasTests) || i > 0 && packages[i-1].Package >= pkg.Package {
			t.Fatalf("wrong assignment or order: %+v", packages)
		}
	}
	for _, data := range []string{"", "not-json", `{}`, `{"ImportPath":"elsewhere"}`, `{"ImportPath":"example.test/l"}{"ImportPath":"example.test/l"}`} {
		if _, err := discoverTestPackages(strings.NewReader(data), module); err == nil {
			t.Errorf("accepted invalid discovery %q", data)
		}
	}
	hasTests, noTests := true, false
	packages = []testgate.Package{{Package: module, Phase: "tests", HasTests: &hasTests}, {Package: module + "/none", Phase: "tests", HasTests: &noTests}, {Package: module + "/cmd/l", Phase: "tests-native", HasTests: &hasTests}}
	const complete = `{"Action":"pass","Package":"example.test/l"}{"Action":"skip","Package":"example.test/l/none"}`
	const run = `{"Action":"run","Package":"example.test/l","Test":"TestNamed"}`
	for _, tc := range []struct {
		name   string
		events string
		valid  bool
	}{
		{"helper-only and no-test packages", complete, true},
		{"named pass", run + `{"Action":"pass","Package":"example.test/l","Test":"TestNamed"}` + complete, true},
		{"empty", "", false},
		{"missing", `{"Action":"pass","Package":"example.test/l"}`, false},
		{"duplicate", complete + `{"Action":"pass","Package":"example.test/l"}`, false},
		{"wrong phase", complete + `{"Action":"pass","Package":"example.test/l/cmd/l"}`, false},
		{"skip package with tests", `{"Action":"skip","Package":"example.test/l"}`, false},
		{"skip named test", run + `{"Action":"skip","Package":"example.test/l","Test":"TestNamed"}` + complete, false},
		{"failed named test", run + `{"Action":"fail","Package":"example.test/l","Test":"TestNamed"}` + complete, false},
		{"named test in no-test package", strings.ReplaceAll(run+`{"Action":"pass","Package":"example.test/l","Test":"TestNamed"}`, module, module+"/none") + complete, false},
		{"named test in wrong phase", strings.ReplaceAll(run+`{"Action":"pass","Package":"example.test/l","Test":"TestNamed"}`, module, module+"/cmd/l") + complete, false},
		{"pass without run", `{"Action":"pass","Package":"example.test/l","Test":"TestNamed"}` + complete, false},
		{"aborted", run + complete, false},
		{"failed", `{"Action":"fail","Package":"example.test/l"}`, false},
		{"malformed", complete + "{", false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			summary := consumeTestEvents(strings.NewReader(tc.events), io.Discard)
			if err := testgate.CheckPhase(packages, "tests", &summary); (err == nil) != tc.valid {
				t.Fatalf("validation=%v events=%+v", err, summary)
			}
		})
	}
	for _, action := range []string{"run", "pass"} {
		event := fmt.Sprintf(`{"Action":%q,"Package":"example.test/l","Test":"TestNamed"}`, action)
		if summary := consumeTestEvents(strings.NewReader(run+event+event), io.Discard); summary.ParseError == "" {
			t.Errorf("duplicate %s accepted", action)
		}
	}
}

func TestTestEventsRequireEveryOriginalPass(t *testing.T) {
	c := contract{Module: "example.test/l", Tests: []testContract{
		{Package: "./cmd/l", Name: "TestCompilerCompile"},
		{Package: "./cmd/l", Name: "TestCompilerCompile/Default"},
	}}
	for _, tc := range []struct {
		name    string
		events  []testEvent
		want    string
		wantAll bool
	}{
		{"pass", []testEvent{{Action: "pass", Test: "TestCompilerCompile"}, {Action: "pass", Test: "TestCompilerCompile/Default"}}, "pass", true},
		{"skip", []testEvent{{Action: "pass", Test: "TestCompilerCompile"}, {Action: "skip", Test: "TestCompilerCompile/Default"}}, "skip", false},
		{"missing subtest", []testEvent{{Action: "pass", Test: "TestCompilerCompile"}}, "missing", false},
		{"panic", []testEvent{{Action: "run", Test: "TestCompilerCompile"}, {Action: "run", Test: "TestCompilerCompile/Default"}, {Action: "output", Output: "panic: no case matches\n"}, {Action: "fail"}}, "aborted", false},
		{"unfinished stream", []testEvent{{Action: "run", Test: "TestCompilerCompile/Default"}}, "aborted", false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var input bytes.Buffer
			for _, event := range tc.events {
				event.Package = "example.test/l/cmd/l"
				if event.Test != "" && (event.Action == "pass" || event.Action == "skip") {
					if err := json.NewEncoder(&input).Encode(testEvent{Action: "run", Package: event.Package, Test: event.Test}); err != nil {
						t.Fatal(err)
					}
				}
				if err := json.NewEncoder(&input).Encode(event); err != nil {
					t.Fatal(err)
				}
			}
			summary := consumeTestEvents(&input, io.Discard)
			outcomes, all := expectedOutcomes(c, summary.Tests)
			if all != tc.wantAll || outcomes[1].Status != tc.want {
				t.Fatalf("outcomes=%+v all=%t", outcomes, all)
			}
		})
	}
}

func TestEventDecoderHandlesLargeLinesAndMalformedJSON(t *testing.T) {
	message := strings.Repeat("λ", 100000)
	var input, readable bytes.Buffer
	if err := json.NewEncoder(&input).Encode(testEvent{Action: "output", Output: message}); err != nil {
		t.Fatal(err)
	}
	summary := consumeTestEvents(&input, &readable)
	if summary.ParseError != "" || readable.String() != message {
		t.Fatalf("large output was lost: error=%s length=%d", summary.ParseError, readable.Len())
	}
	summary = consumeTestEvents(strings.NewReader("not-json\n{\"Action\":\"pass\"}\n"), io.Discard)
	if summary.ParseError == "" {
		t.Fatal("malformed test JSON must not look successful")
	}
}

func TestEventDecoderRetainsFailuresAndCompletedOutcomes(t *testing.T) {
	const completed = "{\"Action\":\"run\",\"Package\":\"fixture\",\"Test\":\"TestDone\"}\n{\"Action\":\"pass\",\"Package\":\"fixture\",\"Test\":\"TestDone\"}\n"
	for _, name := range []string{"malformed JSON", "drain failure", "decoded output failure"} {
		t.Run(name, func(t *testing.T) {
			var output io.Writer = io.Discard
			prefix := completed + "not-json\n"
			wantErrors := []string{"invalid character"}
			wantOutcomes := 1
			if name == "decoded output failure" {
				file, err := os.Create(filepath.Join(t.TempDir(), "closed-output"))
				if err != nil {
					t.Fatal(err)
				}
				if err := file.Close(); err != nil {
					t.Fatal(err)
				}
				output = file
				prefix = completed + "{\"Action\":\"output\",\"Output\":\"message\"}\n"
				wantErrors = []string{"write decoded output", file.Name()}
				wantOutcomes = 2
			}
			tail := bytes.NewBufferString("{\"Action\":\"run\",\"Package\":\"fixture\",\"Test\":\"TestAfter\"}\n{\"Action\":\"pass\",\"Package\":\"fixture\",\"Test\":\"TestAfter\"}\n")
			var input io.Reader = io.MultiReader(strings.NewReader(prefix), tail)
			if name == "drain failure" {
				input = io.MultiReader(input, iotest.ErrReader(errors.New("drain fixture failure")))
				wantErrors = append(wantErrors, "drain test events: drain fixture failure")
			}
			summary := consumeTestEvents(input, output)
			for _, message := range wantErrors {
				if !strings.Contains(summary.ParseError, message) {
					t.Errorf("diagnostic %q lost %q", summary.ParseError, message)
				}
			}
			if tail.Len() != 0 {
				t.Errorf("failure left %d unread bytes", tail.Len())
			}
			outcomes, all := expectedOutcomes(contract{Module: "fixture", Tests: []testContract{{Package: ".", Name: "TestDone"}}}, summary.Tests)
			if !all || len(summary.Tests) != wantOutcomes {
				t.Errorf("completed outcomes lost: %+v; original: %+v", summary.Tests, outcomes)
			}
			if (commandResult{ExitCode: 0, Events: &summary}).succeeded() {
				t.Fatal("event I/O failure reported success")
			}
		})
	}
}

func TestRunCommandArtifactCreationFailure(t *testing.T) {
	exe, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	for _, suffix := range []string{"stdout.log", "stderr.log", "test-events.jsonl", "test-output.log"} {
		t.Run(suffix, func(t *testing.T) {
			dir := t.TempDir()
			blocked := filepath.Join(dir, "setup."+suffix)
			if err := os.Mkdir(blocked, 0700); err != nil {
				t.Fatal(err)
			}
			ready := filepath.Join(dir, "started")
			r := runCommand(context.Background(), dir, commandSpec{
				Name: "setup", Executable: exe, Args: []string{"-test.run=^TestCommandFixture$"}, Dir: dir,
				Env:      append(os.Environ(), "L_HARNESS_TEST_CHILD=wait", "L_HARNESS_TEST_READY="+ready),
				TestJSON: suffix != "stdout.log", Timeout: time.Second,
			})
			if r.succeeded() || r.ExitCode != -1 || r.TimedOut || !strings.Contains(r.Error, blocked) || r.Events != nil {
				t.Fatalf("artifact setup failure lost: %+v", r)
			}
			if _, err := os.Stat(ready); !errors.Is(err, os.ErrNotExist) {
				t.Fatalf("child started before artifacts were ready: %v", err)
			}
		})
	}
}

func TestCommandResultClosesArtifacts(t *testing.T) {
	for _, tc := range []struct {
		name     string
		failures int
		primary  string
		code     int
		timedOut bool
	}{
		{"successful close", 0, "", 0, false},
		{"original failure unchanged", 0, "exit status 7", 7, false},
		{"close failure", 1, "", 0, false},
		{"multiple close failures", 2, "", 0, false},
		{"original failure retained", 2, "exit status 7", 7, false},
		{"timeout retained", 1, "context deadline exceeded", -1, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var files []*os.File
			for index := 0; index <= tc.failures; index++ {
				file, err := os.Create(filepath.Join(t.TempDir(), fmt.Sprintf("artifact-%d", index)))
				if err != nil {
					t.Fatal(err)
				}
				if index < tc.failures {
					if err := file.Close(); err != nil {
						t.Fatal(err)
					}
				} else {
					t.Cleanup(func() {
						if err := file.Close(); err != nil && !errors.Is(err, os.ErrClosed) {
							t.Errorf("fallback close: %v", err)
						}
					})
				}
				files = append(files, file)
			}
			r := commandResult{ExitCode: tc.code, TimedOut: tc.timedOut, Error: tc.primary, StdoutPath: "stdout", StderrPath: "stderr"}
			before := r
			r.closeArtifacts(nil)
			if !reflect.DeepEqual(r, before) {
				t.Fatal("nil handle changed the command result")
			}
			r.closeArtifacts(append(files, nil)...)
			if tc.failures == 0 && r.Error != tc.primary {
				t.Errorf("successful cleanup changed original error: %q", r.Error)
			}
			if tc.primary != "" && !strings.HasPrefix(r.Error, tc.primary) {
				t.Errorf("original error lost: %q", r.Error)
			}
			previous := -1
			for _, file := range files[:tc.failures] {
				index := strings.Index(r.Error, "close artifact "+strconv.Quote(file.Name()))
				if index <= previous {
					t.Errorf("close errors missing or reordered: %q", r.Error)
				}
				previous = index
			}
			for _, file := range files {
				if _, err := file.Stat(); !errors.Is(err, os.ErrClosed) {
					t.Errorf("artifact %s still open: %v", file.Name(), err)
				}
			}
			if r.succeeded() != (tc.failures == 0 && tc.primary == "") {
				t.Errorf("incorrect success after cleanup: %+v", r)
			}
			r.Error = before.Error
			if !reflect.DeepEqual(r, before) {
				t.Errorf("cleanup changed command status or artifact paths: %+v", r)
			}
		})
	}
}

func TestHarnessEnvironmentAndFlagRejection(t *testing.T) {
	t.Setenv("GOFLAGS", "-run=Nothing")
	t.Setenv("GOWORK", "inherited-workspace")
	t.Setenv("GOENV", "inherited-goenv")
	for _, tc := range []struct {
		name, value string
		present     bool
	}{
		{name: "unset"},
		{name: "empty", present: true},
		{name: "explicit", value: "8", present: true},
		{name: "verbatim", value: " 04 ", present: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Setenv("GOMAXPROCS", tc.value)
			if !tc.present {
				if err := os.Unsetenv("GOMAXPROCS"); err != nil {
					t.Fatal(err)
				}
			}
			env := harnessEnvironment(t.TempDir(), "L_HARNESS_ARTIFACT_DIR=/temporary/artifacts")
			values := make(map[string]string)
			for _, entry := range env {
				parts := strings.SplitN(entry, "=", 2)
				if len(parts) == 2 {
					if _, exists := values[parts[0]]; exists {
						t.Fatalf("duplicate environment key %s", parts[0])
					}
					values[parts[0]] = parts[1]
				}
			}
			if values["GOFLAGS"] != "" || values["GOENV"] != "off" || values["GOWORK"] != "off" || values["GOTOOLCHAIN"] != "local" {
				t.Fatalf("test execution environment is not isolated: %+v", values)
			}
			value, present := values["GOMAXPROCS"]
			if value != tc.value || present != tc.present {
				t.Errorf("child GOMAXPROCS=(%q, %t), want inherited (%q, %t)", value, present, tc.value, tc.present)
			}
		})
	}
	for _, flags := range []string{"-run=Nothing", "-skip=TestCompilerCompile", "-overlay=/tmp/overlay.json", "-vet=off", "-count=0", "-short", "-exec=noop", "-tags=changed", `"-overlay=/tmp/path with spaces/overlay.json"`, `'-run=Nothing'`, "-n", "-C=elsewhere"} {
		if validateGoFlags(flags) == nil {
			t.Errorf("unsafe flags accepted: %s", flags)
		}
	}
	if err := validateGoFlags("-p=1"); err != nil {
		t.Fatal(err)
	}
}

func TestCommandFixture(t *testing.T) {
	switch os.Getenv("L_HARNESS_TEST_CHILD") {
	case "exit7":
		if _, err := fmt.Fprint(os.Stdout, "raw stdout\x00\n"); err != nil {
			t.Fatal(err)
		}
		if _, err := fmt.Fprint(os.Stderr, "raw stderr\r\n"); err != nil {
			t.Fatal(err)
		}
		os.Exit(7)
	case "wait":
		if ready := os.Getenv("L_HARNESS_TEST_READY"); ready != "" {
			if err := os.WriteFile(ready, []byte("ready"), 0600); err != nil {
				t.Fatal(err)
			}
		}
		for {
			time.Sleep(time.Second)
		}
	}
}

func TestRunCommandStartupAndCancellation(t *testing.T) {
	exe, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct {
		name string
		json bool
	}{
		{"missing executable", false}, {"canceled before start", false}, {"canceled while running", false},
		{"missing executable with test JSON", true}, {"canceled before start with test JSON", true}, {"canceled while running with test JSON", true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			scenario := strings.TrimSuffix(tc.name, " with test JSON")
			dir := t.TempDir()
			ready := filepath.Join(dir, "ready")
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			program := exe
			done := make(chan struct{})
			switch scenario {
			case "missing executable":
				program = filepath.Join(dir, "missing")
				close(done)
			case "canceled before start":
				cancel()
				close(done)
			case "canceled while running":
				go func() {
					defer close(done)
					ticker := time.NewTicker(5 * time.Millisecond)
					defer ticker.Stop()
					for {
						select {
						case <-ctx.Done():
							return
						case <-ticker.C:
							if _, err := os.Stat(ready); err == nil {
								cancel()
								return
							}
						}
					}
				}()
			}
			r := runCommand(ctx, dir, commandSpec{
				Name: "startup", Executable: program, Args: []string{"-test.run=^TestCommandFixture$"}, Dir: dir,
				Env: append(os.Environ(), "L_HARNESS_TEST_CHILD=wait", "L_HARNESS_TEST_READY="+ready), Timeout: 5 * time.Second, TestJSON: tc.json,
			})
			cancel()
			<-done
			if r.succeeded() || r.Error == "" || r.TimedOut || r.ExitCode != -1 {
				t.Fatalf("startup/cancellation result lost: %+v", r)
			}
			if tc.json && (r.Events == nil || r.Events.ParseError != "") {
				t.Fatalf("event decoder did not finish cleanly: %+v", r.Events)
			}
			if scenario == "canceled while running" {
				if _, err := os.Stat(ready); err != nil {
					t.Fatalf("cancellation did not exercise a running child: %v", err)
				}
			}
		})
	}
}

func TestRunCommandPreservesExitAndStreams(t *testing.T) {
	exe, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	dir := t.TempDir()
	r := runCommand(context.Background(), dir, commandSpec{
		Name: "exit", Executable: exe, Args: []string{"-test.run=^TestCommandFixture$"}, Dir: dir,
		Env: append(os.Environ(), "L_HARNESS_TEST_CHILD=exit7"), Timeout: time.Second * 5,
	})
	if r.ExitCode != 7 || r.succeeded() || r.TimedOut || r.TimeoutMS != 5000 {
		t.Fatalf("incorrect child status: %+v", r)
	}
	if r.StdoutPath != filepath.Join(dir, "exit.stdout.log") || r.StderrPath != filepath.Join(dir, "exit.stderr.log") || r.TestEventsPath != "" || r.TestOutputPath != "" {
		t.Fatalf("non-JSON artifact layout changed: %+v", r)
	}
	stdout, err := os.ReadFile(r.StdoutPath)
	if err != nil {
		t.Fatal(err)
	}
	stderr, err := os.ReadFile(r.StderrPath)
	if err != nil {
		t.Fatal(err)
	}
	if string(stdout) != "raw stdout\x00\n" || string(stderr) != "raw stderr\r\n" {
		t.Fatalf("streams changed: %q %q", stdout, stderr)
	}
	r = runCommand(context.Background(), dir, commandSpec{
		Name: "timeout", Executable: exe, Args: []string{"-test.run=^TestCommandFixture$"}, Dir: dir,
		Env: append(os.Environ(), "L_HARNESS_TEST_CHILD=wait"), Timeout: 50 * time.Millisecond,
	})
	if !r.TimedOut || r.succeeded() || r.TimeoutMS != 50 {
		t.Fatalf("timeout was treated as success: %+v", r)
	}
	r = runCommand(context.Background(), dir, commandSpec{
		Name: "default-timeout", Executable: exe, Args: []string{"-test.run=^TestCommandFixture$"}, Dir: dir,
		Env: append(os.Environ(), "L_HARNESS_TEST_CHILD=exit7"),
	})
	if r.ExitCode != 7 || r.TimeoutMS != 180000 {
		t.Fatalf("default timeout was not recorded: %+v", r)
	}
}

func TestProbeRequiresTestEventAndReport(t *testing.T) {
	dir := t.TempDir()
	r := commandResult{ExitCode: 0, Events: &testgate.Events{}}
	validateProbeResult(&r, dir, "TestHarnessCompilerProbe")
	if r.succeeded() || !strings.Contains(r.Error, "did not execute") || !strings.Contains(r.Error, "report missing") {
		t.Fatalf("empty diagnostic incorrectly accepted: %+v", r)
	}
	if err := os.WriteFile(filepath.Join(dir, "report.json"), []byte(`{"schema_version":1}`), 0644); err != nil {
		t.Fatal(err)
	}
	r = commandResult{ExitCode: 0, Events: &testgate.Events{Tests: []testgate.Outcome{{Name: "TestHarnessCompilerProbe", Status: "skip"}}}}
	validateProbeResult(&r, dir, "TestHarnessCompilerProbe")
	if r.succeeded() {
		t.Fatal("skipped diagnostic incorrectly accepted")
	}
	r = commandResult{ExitCode: 0, Events: &testgate.Events{Tests: []testgate.Outcome{{Name: "TestHarnessCompilerProbe", Status: "pass"}}}}
	validateProbeResult(&r, dir, "TestHarnessCompilerProbe")
	if !r.succeeded() {
		t.Fatalf("valid diagnostic rejected: %+v", r)
	}
}
