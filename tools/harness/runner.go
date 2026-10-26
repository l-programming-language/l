package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/l-programming-language/l/tools/internal/testgate"
)

type testEvent struct {
	Time    time.Time `json:"Time"`
	Action  string    `json:"Action"`
	Package string    `json:"Package"`
	Test    string    `json:"Test"`
	Output  string    `json:"Output"`
}

func discoverTestPackages(input io.Reader, module string) ([]testgate.Package, error) {
	metadata, err := testgate.ReadPackages(input)
	if err != nil {
		return nil, err
	}
	packages := make([]testgate.Package, 0, len(metadata))
	for name, hasTests := range metadata {
		if name != module && !strings.HasPrefix(name, module+"/") {
			return nil, fmt.Errorf("foreign discovered package %q", name)
		}
		path := strings.TrimPrefix(name, module+"/")
		phase := "tests"
		switch path {
		case "acceptance", "cmd/l", "lyger", "logiceme", "internal/cargocache", "internal/testsupport/native":
			phase = "tests-native"
		default:
			if strings.HasPrefix(path, "acceptance/") {
				phase = "tests-native"
			}
		}
		hasTests := hasTests
		packages = append(packages, testgate.Package{Package: name, Phase: phase, HasTests: &hasTests})
	}
	sort.Slice(packages, func(i, j int) bool { return packages[i].Package < packages[j].Package })
	return packages, nil
}

// consumeTestEvents decodes incrementally: debug output is not retained in memory.
func consumeTestEvents(input io.Reader, readable io.Writer) testgate.Events {
	decoder := json.NewDecoder(input)
	states := make(map[string]testgate.Outcome)
	summary := testgate.Events{Tests: []testgate.Outcome{}, Packages: []testgate.Outcome{}}
	for {
		var event testEvent
		err := decoder.Decode(&event)
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			summary.ParseError = err.Error()
			// Keep draining so a malformed stream cannot deadlock the child process.
			if _, drainErr := io.Copy(io.Discard, input); drainErr != nil {
				summary.ParseError = errors.Join(err, fmt.Errorf("drain test events: %w", drainErr)).Error()
			}
			break
		}
		if event.Output != "" {
			if _, err = io.WriteString(readable, event.Output); err != nil && summary.ParseError == "" {
				summary.ParseError = "write decoded output: " + err.Error()
			}
		}
		if event.Test != "" {
			key := event.Package + "\x00" + event.Test
			state, exists := states[key]
			if !exists {
				state = testgate.Outcome{Package: event.Package, Name: event.Test, Status: "missing"}
			}
			switch event.Action {
			case "run":
				if state.Status != "missing" && summary.ParseError == "" {
					summary.ParseError = "duplicate test run: " + key
				}
				state.Status = "running"
			case "pass", "fail", "skip":
				if state.Status != "running" && summary.ParseError == "" {
					summary.ParseError = "test outcome without an active run: " + key
				}
				state.Status = event.Action
			}
			states[key] = state
		}
		if event.Test == "" && (event.Action == "pass" || event.Action == "fail" || event.Action == "skip" || event.Action == "build-fail") {
			summary.Packages = append(summary.Packages, testgate.Outcome{Package: event.Package, Status: event.Action})
		}
		if event.Test == "" && (event.Action == "fail" || event.Action == "build-fail") {
			for key, state := range states {
				if state.Package == event.Package && state.Status == "running" {
					state.Status = "aborted"
					states[key] = state
				}
			}
		}
	}
	for _, state := range states {
		if state.Status == "running" {
			state.Status = "aborted"
		}
		summary.Tests = append(summary.Tests, state)
	}
	sort.Slice(summary.Tests, func(i, j int) bool {
		a, b := summary.Tests[i], summary.Tests[j]
		return a.Package < b.Package || a.Package == b.Package && a.Name < b.Name
	})
	return summary
}

func expectedOutcomes(c contract, actual []testgate.Outcome) ([]testgate.Outcome, bool) {
	states := make(map[string]string)
	for _, test := range actual {
		states[test.Package+"\x00"+test.Name] = test.Status
	}
	results := make([]testgate.Outcome, 0, len(c.Tests))
	ok := true
	for _, test := range c.Tests {
		packageName := c.Module
		if test.Package != "." {
			packageName += "/" + strings.TrimPrefix(test.Package, "./")
		}
		status := states[packageName+"\x00"+test.Name]
		if status == "" {
			status = "missing"
		}
		if status != "pass" {
			ok = false
		}
		results = append(results, testgate.Outcome{Package: test.Package, Name: test.Name, Status: status})
	}
	return results, ok
}

type commandSpec struct {
	Name       string
	Executable string
	Args       []string
	Dir        string
	Env        []string
	Timeout    time.Duration
	TestJSON   bool
}

type commandResult struct {
	Name           string                `json:"name"`
	Executable     string                `json:"executable"`
	Args           []string              `json:"args"`
	Directory      string                `json:"directory"`
	StartedAt      time.Time             `json:"started_at"`
	DurationMS     int64                 `json:"duration_ms"`
	TimeoutMS      int64                 `json:"timeout_ms"`
	ExitCode       int                   `json:"exit_code"`
	TimedOut       bool                  `json:"timed_out"`
	Error          string                `json:"error,omitempty"`
	StdoutPath     string                `json:"stdout_path"`
	StderrPath     string                `json:"stderr_path"`
	TestEventsPath string                `json:"test_events_path,omitempty"`
	TestOutputPath string                `json:"test_output_path,omitempty"`
	Events         *testgate.Events      `json:"events,omitempty"`
	Cancellations  []cancellationAttempt `json:"cancellation_attempts,omitempty"`
}

// Cancellation diagnostics preserve the raw signal result separately from the
// command's exit status. In particular, a post-wait error does not describe the
// earlier deadline signal, and an absent group is not a permissions failure.
type cancellationAttempt struct {
	Phase       string                   `json:"phase"`
	PID         int                      `json:"pid"`
	PGID        int                      `json:"pgid,omitempty"`
	Signal      string                   `json:"signal"`
	DurationNS  int64                    `json:"duration_ns"`
	Outcome     string                   `json:"outcome"`
	Error       string                   `json:"error,omitempty"`
	Observation *cancellationObservation `json:"disappearance_observation,omitempty"`
}

type cancellationObservation struct {
	DurationNS int64  `json:"duration_ns"`
	BudgetNS   int64  `json:"remaining_wait_delay_ns"`
	Probes     int    `json:"signal_zero_probes"`
	Outcome    string `json:"outcome"`
	LastError  string `json:"last_error,omitempty"`
}

type deadlineSignal struct {
	pid, pgid int
	started   time.Time
}

type commandCancellation struct {
	mu       sync.Mutex
	attempts []cancellationAttempt
	cancel   func(string) error
	deadline *deadlineSignal
	observe  func(string, int, int, error) (*cancellationObservation, bool)
}

func (c *commandCancellation) signal(phase string, pid, pgid int, signal string, send func() error, absent func(error) bool) error {
	started := time.Now()
	err := send()
	attempt := cancellationAttempt{Phase: phase, PID: pid, PGID: pgid, Signal: signal,
		DurationNS: time.Since(started).Nanoseconds(), Outcome: "signaled"}
	if phase == "deadline" && err == nil {
		c.mu.Lock()
		if c.deadline == nil {
			c.deadline = &deadlineSignal{pid: pid, pgid: pgid, started: started}
		}
		c.mu.Unlock()
	}
	if err != nil {
		attempt.Error = err.Error()
		attempt.Outcome = "failed"
		if errors.Is(err, os.ErrProcessDone) || absent != nil && absent(err) {
			attempt.Outcome = "absent"
			err = os.ErrProcessDone
		}
		if c.observe != nil {
			var absent bool
			attempt.Observation, absent = c.observe(phase, pid, pgid, err)
			if absent {
				// Preserve the raw signal error in the report. Only a subsequent
				// ESRCH observation establishes that no group remains to clean.
				attempt.Outcome = "absent_after_observation"
				err = os.ErrProcessDone
			}
		}
	}
	c.mu.Lock()
	c.attempts = append(c.attempts, attempt)
	c.mu.Unlock()
	return err
}

func (c *commandCancellation) snapshot() []cancellationAttempt {
	c.mu.Lock()
	defer c.mu.Unlock()
	attempts := append([]cancellationAttempt(nil), c.attempts...)
	for i := range attempts {
		if attempts[i].Observation != nil {
			observation := *attempts[i].Observation
			attempts[i].Observation = &observation
		}
	}
	return attempts
}

func (r commandResult) succeeded() bool {
	return r.ExitCode == 0 && !r.TimedOut && r.Error == "" && (r.Events == nil || r.Events.ParseError == "")
}

func (r *commandResult) closeArtifacts(files ...*os.File) {
	for _, file := range files {
		if file == nil {
			continue
		}
		if err := file.Close(); err != nil {
			if r.Error != "" {
				r.Error += "\n"
			}
			r.Error += fmt.Sprintf("close artifact %q: %s", file.Name(), err)
		}
	}
}

func runCommand(ctx context.Context, outputDir string, spec commandSpec) (r commandResult) {
	return runCommandWithCancellation(ctx, outputDir, spec, configureCancellation)
}

// Keep cancellation injection local to one invocation so cleanup failures can
// be tested through the real command, stream, deadline and reporting paths.
func runCommandWithCancellation(ctx context.Context, outputDir string, spec commandSpec, configure func(*exec.Cmd) *commandCancellation) (r commandResult) {
	started := time.Now()
	r = commandResult{Name: spec.Name, Executable: spec.Executable, Args: spec.Args, Directory: spec.Dir, ExitCode: -1, StartedAt: started.UTC()}
	if spec.Timeout <= 0 {
		spec.Timeout = testgate.BuildTimeout
	}
	r.TimeoutMS = spec.Timeout.Milliseconds()
	ctx, cancel := context.WithTimeout(ctx, spec.Timeout)
	defer cancel()
	var stdout, stderr, readable *os.File
	defer func() {
		r.closeArtifacts(readable, stderr, stdout)
	}()
	r.StdoutPath = filepath.Join(outputDir, spec.Name+".stdout.log")
	if spec.TestJSON {
		r.TestEventsPath = filepath.Join(outputDir, spec.Name+".test-events.jsonl")
		r.StdoutPath = r.TestEventsPath
	}
	r.StderrPath = filepath.Join(outputDir, spec.Name+".stderr.log")
	var err error
	stdout, err = os.Create(r.StdoutPath)
	if err != nil {
		r.Error = err.Error()
		return r
	}
	stderr, err = os.Create(r.StderrPath)
	if err != nil {
		r.Error = err.Error()
		return r
	}
	cmd := exec.CommandContext(ctx, spec.Executable, spec.Args...)
	cmd.Dir, cmd.Env = spec.Dir, spec.Env
	cmd.Stderr = stderr
	cmd.WaitDelay = 2 * time.Second
	cleanup := configure(cmd)
	cmd.Cancel = func() error {
		phase := "context_cancellation"
		if errors.Is(ctx.Err(), context.DeadlineExceeded) {
			phase = "deadline"
		}
		return cleanup.cancel(phase)
	}
	var decoderDone sync.WaitGroup
	var pipeWriter *io.PipeWriter
	if spec.TestJSON {
		r.TestOutputPath = filepath.Join(outputDir, spec.Name+".test-output.log")
		readable, err = os.Create(r.TestOutputPath)
		if err != nil {
			r.Error = err.Error()
			return r
		}
		reader, writer := io.Pipe()
		pipeWriter = writer
		cmd.Stdout = io.MultiWriter(stdout, writer)
		decoderDone.Add(1)
		go func() {
			defer decoderDone.Done()
			// io.Pipe closes always return nil.
			defer func() { _ = reader.Close() }() //nolint:errcheck // io.PipeReader.Close always returns nil.
			summary := consumeTestEvents(reader, readable)
			r.Events = &summary
		}()
	} else {
		cmd.Stdout = stdout
	}
	err = cmd.Run()
	if cmd.Process != nil {
		// Descendants can outlive any completed parent, even after closing its pipes.
		if cancelErr := cleanup.cancel("post_wait"); cancelErr != nil && !errors.Is(cancelErr, os.ErrProcessDone) {
			err = errors.Join(err, fmt.Errorf("terminate remaining processes: %w", cancelErr))
		}
	}
	r.Cancellations = cleanup.snapshot()
	if pipeWriter != nil {
		_ = pipeWriter.Close() //nolint:errcheck // Always nil; sends EOF before waiting for the decoder.
		decoderDone.Wait()
	}
	r.DurationMS = time.Since(started).Milliseconds()
	r.TimedOut = errors.Is(ctx.Err(), context.DeadlineExceeded)
	if cmd.ProcessState != nil {
		r.ExitCode = cmd.ProcessState.ExitCode()
	}
	if err != nil {
		r.Error = err.Error()
	}
	return r
}

func harnessEnvironment(root string, extra ...string) []string {
	// GOFLAGS could inject -run, -skip, -overlay, or -vet=off into the gate.
	overrides := []string{"GOFLAGS=", "GOENV=off", "GOWORK=off", "GOTOOLCHAIN=local", "GOCACHE=" + filepath.Join(root, ".harness", "cache", "go-build")}
	overrides = append(overrides, extra...)
	keys := make(map[string]bool)
	for _, entry := range overrides {
		keys[strings.SplitN(entry, "=", 2)[0]] = true
	}
	env := make([]string, 0, len(os.Environ())+len(overrides))
	for _, entry := range os.Environ() {
		if !keys[strings.SplitN(entry, "=", 2)[0]] {
			env = append(env, entry)
		}
	}
	return append(env, overrides...)
}

func originalGoFlags() string {
	if flags, ok := os.LookupEnv("L_HARNESS_ORIGINAL_GOFLAGS"); ok {
		return flags
	}
	if flags := os.Getenv("GOFLAGS"); flags != "" {
		return flags
	}
	// An empty environment value falls back to persisted `go env -w` settings.
	// Reading this one setting does not apply it to any compiler invocation.
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, "go", "env", "GOFLAGS")
	cmd.Env = append(os.Environ(), "GOTOOLCHAIN=local", "GOWORK=off")
	if output, err := cmd.Output(); err == nil {
		return strings.TrimSpace(string(output))
	}
	return ""
}

func validateGoFlags(flags string) error {
	blocked := map[string]bool{
		"-run": true, "-skip": true, "-list": true, "-overlay": true,
		"-vet": true, "-tags": true, "-exec": true, "-c": true,
		"-count": true, "-failfast": true, "-short": true, "-shuffle": true,
		"-modfile": true, "-toolexec": true, "-n": true, "-C": true,
	}
	for _, field := range strings.Fields(flags) {
		// Go accepts a quoted entire flag token, including paths with spaces.
		// Only the flag name is needed here; all flags are cleared for execution.
		name := strings.Trim(strings.SplitN(field, "=", 2)[0], "\"'")
		if blocked[name] || strings.HasPrefix(name, "-test.") {
			return fmt.Errorf("GOFLAGS contains %s, which can change test selection or execution; clear it before running the harness", name)
		}
	}
	return nil
}

func requireCommand(result commandResult) string {
	if result.succeeded() {
		return "pass"
	}
	if result.TimedOut {
		return "timeout"
	}
	if result.Events != nil && result.Events.ParseError != "" {
		return "invalid test event stream"
	}
	return fmt.Sprintf("failed (exit %d)", result.ExitCode)
}
