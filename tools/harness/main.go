// Command harness checks current acceptance inputs and runs the ordinary gate.
package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/l-programming-language/l/tools/internal/testgate"
)

type runReport struct {
	SchemaVersion    int                `json:"schema_version"`
	Command          string             `json:"command"`
	StartedAt        time.Time          `json:"started_at"`
	FinishedAt       time.Time          `json:"finished_at"`
	RunDirectory     string             `json:"run_directory"`
	Repository       string             `json:"repository"`
	CurrentRevision  string             `json:"current_revision,omitempty"`
	GitDirty         *bool              `json:"git_dirty,omitempty"`
	Environment      map[string]string  `json:"environment"`
	IntegrityBefore  integrityReport    `json:"integrity_before"`
	IntegrityAfter   integrityReport    `json:"integrity_after"`
	Commands         []commandResult    `json:"commands"`
	TestPackages     []testgate.Package `json:"test_packages,omitempty"`
	OriginalTests    []testgate.Outcome `json:"original_tests,omitempty"`
	Difference       *byteDifference    `json:"difference,omitempty"`
	AcceptancePassed *bool              `json:"acceptance_passed,omitempty"`
	ExitCode         int                `json:"exit_code"`
	Error            string             `json:"error,omitempty"`
}

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	code := runCLI(ctx, os.Args[1:], os.Stdout, os.Stderr)
	stop()
	os.Exit(code)
}

func findRoot() (string, error) {
	if root := os.Getenv("L_HARNESS_ROOT"); root != "" {
		return filepath.Abs(root)
	}
	current, err := os.Getwd()
	if err != nil {
		return "", err
	}
	for ; ; current = filepath.Dir(current) {
		if _, err = os.Stat(filepath.Join(current, "harness", "contract.json")); err == nil {
			return current, nil
		}
		if filepath.Dir(current) == current {
			return "", fmt.Errorf("cannot find harness/contract.json; use scripts/l-harness from the repository")
		}
	}
}

func runCLI(ctx context.Context, args []string, stdout, stderr io.Writer) int {
	fail := func(message any) int {
		_, _ = fmt.Fprintln(stderr, message) //nolint:errcheck // Console diagnostics are best-effort; checked reports and command status determine acceptance.
		return 2
	}
	usage := "usage: l-harness doctor | verify | diagnose [--frontend] | diff EXPECTED ACTUAL"
	var frontend bool
	if len(args) > 0 && args[0] == "diagnose" {
		flags := flag.NewFlagSet("diagnose", flag.ContinueOnError)
		flags.SetOutput(stderr)
		flags.BoolVar(&frontend, "frontend", false, "add isolated compiler and import probe overlays")
		if err := flags.Parse(args[1:]); err != nil || flags.NArg() != 0 {
			return fail(usage)
		}
	} else if len(args) == 0 || len(args) != map[string]int{"doctor": 1, "verify": 1, "diff": 3}[args[0]] {
		return fail(usage)
	}
	command := args[0]
	root, err := findRoot()
	if err != nil {
		return fail(err)
	}
	c, contractHash, err := readContract(root)
	if err != nil {
		return fail(err)
	}
	runs := filepath.Join(root, ".harness", "runs")
	for _, directory := range []string{runs, filepath.Join(root, ".harness", "cache", "go-build")} {
		if err = os.MkdirAll(directory, 0755); err != nil {
			return fail(err)
		}
	}
	started := time.Now().UTC()
	runDir, err := os.MkdirTemp(runs, started.Format("20060102T150405.000000000Z")+"-")
	if err != nil {
		return fail(err)
	}
	effectiveGOMAXPROCS := runtime.GOMAXPROCS(0)
	r := runReport{
		SchemaVersion: 2, Command: command, StartedAt: started, RunDirectory: runDir, Repository: root,
		Commands: []commandResult{},
		Environment: map[string]string{"go_version": runtime.Version(), "goos": runtime.GOOS, "goarch": runtime.GOARCH,
			"evidence_run_id":  filepath.Base(runDir),
			"original_goflags": originalGoFlags(), "original_goenv": os.Getenv("GOENV"),
			"effective_goflags": "", "goenv": "off", "gowork": "off", "gotoolchain": "local",
			"effective_gomaxprocs": fmt.Sprint(effectiveGOMAXPROCS), "num_cpu": fmt.Sprint(runtime.NumCPU()),
			"gocache": filepath.Join(root, ".harness", "cache", "go-build")},
	}
	if requested, present := os.LookupEnv("GOMAXPROCS"); present {
		r.Environment["original_gomaxprocs"] = requested
	}
	if command == "verify" {
		packageParallelism := effectiveGOMAXPROCS
		if packageParallelism > 3 {
			packageParallelism = 3
		}
		r.Environment["test_package_parallelism"] = fmt.Sprint(packageParallelism)
		// Keep Go's existing parallel-test budget explicit in the report and command.
		r.Environment["test_case_parallelism"] = fmt.Sprint(effectiveGOMAXPROCS)
		r.Environment["build_command_timeout_ms"] = fmt.Sprint(testgate.BuildTimeout.Milliseconds())
		r.Environment["test_package_timeout_ms"] = fmt.Sprint(testgate.PackageTimeout.Milliseconds())
		r.Environment["test_command_timeout_ms"] = fmt.Sprint(testgate.PhaseTimeout.Milliseconds())
		r.Environment["package_discovery_timeout_ms"] = fmt.Sprint(testgate.DiscoveryTimeout.Milliseconds())
		r.Environment["native_test_package_timeout_ms"] = fmt.Sprint(testgate.NativePackageTimeout.Milliseconds())
		r.Environment["native_test_command_timeout_ms"] = fmt.Sprint(testgate.NativePhaseTimeout.Milliseconds())
	}
	r.CurrentRevision, r.GitDirty = repositoryState(ctx, root)
	if goPath, lookupErr := exec.LookPath("go"); lookupErr == nil {
		if absolute, absErr := filepath.Abs(goPath); absErr == nil {
			goPath = absolute
		}
		r.Environment["go_executable"] = goPath
	}
	r.IntegrityBefore = inspectIntegrity(root, c, contractHash)
	code := 0
	if !r.IntegrityBefore.OK {
		code = 2
		r.Error = "integrity checks failed; no build, test, or probe commands were run"
	} else {
		appendCommand := func(name string, args []string, testJSON bool, timeout time.Duration) commandResult {
			result := runCommand(ctx, runDir, commandSpec{Name: name, Executable: "go", Args: args, Dir: root,
				Env: harnessEnvironment(root, "L_HARNESS_RUN_ID="+r.Environment["evidence_run_id"]), Timeout: timeout, TestJSON: testJSON})
			r.Commands = append(r.Commands, result)
			if result.ExitCode == -1 && !result.TimedOut {
				code = 2
				r.Error = "command could not start: " + result.Error
			} else if !result.succeeded() && code == 0 {
				code = 1
			}
			_, _ = fmt.Fprintf(stdout, "%s: %s\n", name, requireCommand(result)) //nolint:errcheck // Console diagnostics are best-effort; checked reports and command status determine acceptance.
			return result
		}
		switch command {
		case "verify":
			commands := testgate.Commands(nil, effectiveGOMAXPROCS)
			appendCommand(commands[0].Name, commands[0].Args, false, commands[0].Timeout)
			discovery := appendCommand(commands[1].Name, commands[1].Args, false, commands[1].Timeout)
			if !discovery.succeeded() {
				break
			}
			data, readErr := os.ReadFile(discovery.StdoutPath)
			if readErr == nil {
				r.TestPackages, readErr = discoverTestPackages(strings.NewReader(string(data)), c.Module)
			}
			if readErr != nil {
				r.Error, code = readErr.Error(), 2
				break
			}
			commands = testgate.Commands(r.TestPackages, effectiveGOMAXPROCS)
			var actual []testgate.Outcome
			for _, phase := range commands[2:] {
				result := appendCommand(phase.Name, phase.Args, true, phase.Timeout)
				if eventErr := testgate.CheckPhase(r.TestPackages, phase.Name, result.Events); eventErr != nil {
					if r.Error == "" {
						r.Error = eventErr.Error()
					}
					if code == 0 {
						code = 1
					}
				}
				if result.Events != nil {
					actual = append(actual, result.Events.Tests...)
				}
			}
			var originalsPassed bool
			r.OriginalTests, originalsPassed = expectedOutcomes(c, actual)
			if !originalsPassed && code == 0 {
				code = 1
			}
		case "diagnose":
			for _, item := range []struct{ name, pkg, test string }{
				{"initialize", "./cmd/l", "TestCompilerInitialize"},
				{"compile", "./cmd/l", "TestCompilerCompile"},
				{"imports", "./lyger", "Test_importScan"},
			} {
				appendCommand(item.name, []string{"test", "-mod=readonly", "-count=1", "-timeout=60s", "-json", "-run", "^" + item.test + "$", item.pkg}, true, testgate.BuildTimeout)
			}
			appendCommand("vet", []string{"vet", "-mod=readonly", "./..."}, false, testgate.BuildTimeout)
			if frontend {
				results, probeErr := runFrontendDiagnostics(ctx, root, runDir)
				r.Commands = append(r.Commands, results...)
				for _, result := range results {
					_, _ = fmt.Fprintf(stdout, "%s: %s\n", result.Name, requireCommand(result)) //nolint:errcheck // Console diagnostics are best-effort; checked reports and command status determine acceptance.
					if !result.succeeded() && code == 0 {
						code = 1
					}
				}
				if probeErr != nil {
					r.Error, code = probeErr.Error(), 2
				}
			}
		case "diff":
			caller := os.Getenv("L_HARNESS_CALLER_DIR")
			if caller == "" {
				caller, err = os.Getwd()
				if err != nil {
					r.Error, code = err.Error(), 2
					break
				}
			}
			paths := make([]string, 2)
			for i, path := range args[1:] {
				if !filepath.IsAbs(path) {
					path = filepath.Join(caller, path)
				}
				paths[i] = path
			}
			difference, diffErr := compareFiles(paths[0], paths[1])
			if diffErr != nil {
				r.Error, code = diffErr.Error(), 2
			} else {
				r.Difference = &difference
				if !difference.Equal {
					code = 1
				}
				encoded, encodeErr := json.MarshalIndent(difference, "", "  ")
				if encodeErr != nil {
					r.Error, code = encodeErr.Error(), 2
					break
				}
				_, _ = fmt.Fprintln(stdout, string(encoded)) //nolint:errcheck // Console diagnostics are best-effort; checked reports and command status determine acceptance.
			}
		}
	}
	r.IntegrityAfter = inspectIntegrity(root, c, contractHash)
	if !r.IntegrityAfter.OK {
		code = 2
		if r.Error == "" {
			r.Error = "integrity checks failed after execution"
		}
	}
	if command == "verify" {
		passed := code == 0
		r.AcceptancePassed = &passed
	}
	r.FinishedAt, r.ExitCode = time.Now().UTC(), code
	reportPath := filepath.Join(runDir, "report.json")
	if err = testgate.WriteJSON(reportPath, r); err != nil {
		return fail(fmt.Errorf("write report: %w", err))
	}
	if code == 2 {
		_, _ = fmt.Fprintln(stderr, r.Error) //nolint:errcheck // Console diagnostics are best-effort; checked reports and command status determine acceptance.
		for _, check := range r.IntegrityAfter.Checks {
			if !check.OK {
				_, _ = fmt.Fprintf(stderr, "%s %s: %s\n", check.Kind, check.Path, check.Message) //nolint:errcheck // Console diagnostics are best-effort; checked reports and command status determine acceptance.
			}
		}
	}
	if command == "doctor" && code == 0 {
		_, _ = fmt.Fprintf(stdout, "doctor: pass (%d checks)\n", len(r.IntegrityAfter.Checks)) //nolint:errcheck // Console diagnostics are best-effort; checked reports and command status determine acceptance.
	} else if command == "verify" {
		_, _ = fmt.Fprintf(stdout, "acceptance_passed: %t\n", *r.AcceptancePassed) //nolint:errcheck // Console diagnostics are best-effort; checked reports and command status determine acceptance.
	} else if command == "diagnose" {
		_, _ = fmt.Fprintln(stdout, "diagnostics only; acceptance was not evaluated") //nolint:errcheck // Console diagnostics are best-effort; checked reports and command status determine acceptance.
	}
	_, _ = fmt.Fprintln(stdout, "report:", reportPath) //nolint:errcheck // Console diagnostics are best-effort; checked reports and command status determine acceptance.
	return code
}

// Record the checkout identity when Git metadata is available.
// Non-Git fixture directories are valid harness inputs, so unavailable Git is not fatal.
func repositoryState(ctx context.Context, root string) (string, *bool) {
	readGit := func(args ...string) ([]byte, error) {
		commandCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
		defer cancel()
		cmd := exec.CommandContext(commandCtx, "git", args...)
		cmd.Dir = root
		cmd.Env = append(os.Environ(), "GIT_OPTIONAL_LOCKS=0")
		return cmd.Output()
	}
	head, err := readGit("rev-parse", "--verify", "HEAD")
	if err != nil {
		return "", nil
	}
	status, err := readGit("status", "--porcelain")
	if err != nil {
		return strings.TrimSpace(string(head)), nil
	}
	dirty := len(status) != 0
	return strings.TrimSpace(string(head)), &dirty
}
