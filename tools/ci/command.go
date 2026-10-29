package main

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"time"
)

type commandResult struct {
	Name       string    `json:"name"`
	Executable string    `json:"executable"`
	Args       []string  `json:"args"`
	StartedAt  time.Time `json:"started_at"`
	DurationMS int64     `json:"duration_ms"`
	TimeoutMS  int64     `json:"timeout_ms"`
	ExitCode   int       `json:"exit_code"`
	TimedOut   bool      `json:"timed_out"`
	StdoutPath string    `json:"stdout_path"`
	StderrPath string    `json:"stderr_path"`
}

func environment(root string, extra map[string]string) []string {
	values := map[string]string{}
	for _, item := range os.Environ() {
		key, value, ok := strings.Cut(item, "=")
		if ok {
			values[key] = value
		}
	}
	for _, key := range []string{"L_HARNESS_TIMING_DIR", "L_HARNESS_CARGO_EVIDENCE_DIR", "CARGO_LOG", "L_HARNESS_RUN_ID", "L_LUCID_ARTIFACT_DIR"} {
		delete(values, key)
	}
	values["GOENV"], values["GOFLAGS"], values["GOWORK"], values["GOTOOLCHAIN"] = "off", "", "off", "local"
	values["GOCACHE"] = filepath.Join(root, ".harness", "cache", "go-build")
	values["GOMAXPROCS"] = strconv.Itoa(ciParallelism(runtime.NumCPU()))
	values["PATH"] = filepath.Join(runtime.GOROOT(), "bin") + string(os.PathListSeparator) + values["PATH"]
	for key, value := range extra {
		values[key] = value
	}
	keys := make([]string, 0, len(values))
	for key := range values {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	env := make([]string, 0, len(keys))
	for _, key := range keys {
		env = append(env, key+"="+values[key])
	}
	return env
}

func ciParallelism(cpus int) int {
	if cpus < 4 {
		return cpus
	}
	return 4
}

func (r *runner) command(name string, timeout time.Duration, extra map[string]string, executable string, args ...string) ([]byte, error) {
	if executable == "go" || executable == "gofmt" {
		executable = filepath.Join(runtime.GOROOT(), "bin", executable)
	}
	result := commandResult{Name: name, Executable: executable, Args: args, StartedAt: time.Now().UTC(), TimeoutMS: timeout.Milliseconds(), ExitCode: -1, StdoutPath: filepath.Join(r.directory, name+".stdout.log"), StderrPath: filepath.Join(r.directory, name+".stderr.log")}
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	cmd := exec.CommandContext(ctx, executable, args...)
	cmd.Dir = r.root
	cmd.Env = environment(r.root, extra)
	configureCommand(cmd)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	result.DurationMS = time.Since(result.StartedAt).Milliseconds()
	result.TimedOut = ctx.Err() != nil
	if cmd.ProcessState != nil {
		result.ExitCode = cmd.ProcessState.ExitCode()
	}
	writeErr := errors.Join(os.WriteFile(result.StdoutPath, stdout.Bytes(), 0644), os.WriteFile(result.StderrPath, stderr.Bytes(), 0644))
	r.report.Commands = append(r.report.Commands, result)
	if err != nil {
		err = fmt.Errorf("%s failed (exit %d, timeout %t); logs: %s, %s: %w", name, result.ExitCode, result.TimedOut, result.StdoutPath, result.StderrPath, err)
	}
	return stdout.Bytes(), errors.Join(err, writeErr)
}
