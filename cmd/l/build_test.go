package main

import (
	"bytes"
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	nativetest "github.com/l-programming-language/l/internal/testsupport/native"
)

// Exercise the installed CLI, not a second implementation of its lifecycle.

func copyBuildProject(t *testing.T, root, destination, filename, source string) {
	t.Helper()
	if err := os.MkdirAll(destination, 0700); err != nil {
		t.Fatal(err)
	}
	if source == "" {
		content, err := os.ReadFile(filepath.Join(root, "demo/basic/main.l"))
		if err != nil {
			t.Fatal(err)
		}
		source = string(content)
	}
	if err := os.WriteFile(filepath.Join(destination, filename), []byte(source), 0600); err != nil {
		t.Fatal(err)
	}
}

func runBuildCommand(t *testing.T, executable, directory string, environment []string, args ...string) ([]byte, []byte, error) {
	t.Helper()
	return runCommandWithTimeout(t, nativetest.BuildPrerequisite, executable, directory, environment, args...)
}

func runRuntimeCommand(t *testing.T, executable, directory string, environment []string, args ...string) ([]byte, []byte, error) {
	t.Helper()
	return runCommandWithTimeout(t, nativetest.RuntimeExecution, executable, directory, environment, args...)
}

func runCommandWithTimeout(t *testing.T, phase nativetest.CommandPhase, executable, directory string, environment []string, args ...string) ([]byte, []byte, error) {
	t.Helper()
	ctx, cancel := phase.Context(context.Background(), 20*time.Second)
	defer cancel()
	command := exec.CommandContext(ctx, executable, args...)
	command.Dir = directory
	command.Env = append(os.Environ(), environment...)
	var stdout, stderr bytes.Buffer
	command.Stdout, command.Stderr = &stdout, &stderr
	err := command.Run()
	if ctx.Err() != nil {
		t.Fatalf("%s command %v canceled on %s/%s: %v", phase, args, runtime.GOOS, runtime.GOARCH, ctx.Err())
	}
	return stdout.Bytes(), stderr.Bytes(), err
}
