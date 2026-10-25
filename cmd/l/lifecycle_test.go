package main

import (
	"context"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// These tests are deliberately serial while observing process stdout. A file
// avoids pipe-buffer deadlock if a regression emits a large compiler trace.
func lifecycleStdout(t *testing.T, action func()) string {
	t.Helper()
	capture, err := os.CreateTemp(t.TempDir(), "stdout-")
	if err != nil {
		t.Fatal(err)
	}
	previous := os.Stdout
	func() {
		defer func() {
			os.Stdout = previous
			if callErr := capture.Close(); callErr != nil {
				t.Fatal(callErr)
			}
		}()
		os.Stdout = capture
		action()
	}()
	content, err := os.ReadFile(capture.Name())
	if err != nil {
		t.Fatal(err)
	}
	return string(content)
}

func TestCompilerLifecycleGuards(t *testing.T) {
	for _, operation := range []string{"compile", "process"} {
		for _, cli := range []*cliData{nil, {jobType: "build", mainPath: "uninitialized.l"}} {
			name := operation + "/without_input"
			if cli != nil {
				name = operation + "/without_initialization"
			}
			t.Run(name, func(t *testing.T) {
				c := newCompiler(cli)
				returned := false
				output := lifecycleStdout(t, func() {
					if operation == "compile" {
						returned = c.compile() != nil
					} else {
						returned = c.process() != nil
					}
				})
				if returned || c.err == nil || !strings.Contains(c.err.Error(), "not initialized") {
					t.Fatalf("%s returned=%v, err=%v; expected initialization error", operation, returned, c.err)
				}
				if output != "" {
					t.Fatalf("%s wrote to stdout: %q", operation, output)
				}
			})
		}
	}
}

func TestCompilerPreservesInitializationError(t *testing.T) {
	directorySource := filepath.Join(t.TempDir(), "directory.l")
	if err := os.Mkdir(directorySource, 0700); err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct {
		name string
		cli  *cliData
		want string
	}{
		{"missing input", nil, "input is required"},
		{"missing file", &cliData{mainPath: filepath.Join(t.TempDir(), "missing.l")}, "does not exist"},
		{"invalid extension", &cliData{mainPath: filepath.Join(t.TempDir(), "source.txt")}, "not L"},
		{"source read error", &cliData{mainPath: directorySource}, "read main source"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			c := newCompiler(tc.cli)
			var original error
			var compiled, processed bool
			output := lifecycleStdout(t, func() {
				c.initialize(canonicalCompilerOptions(t))
				original = c.err
				compiled = c.compile() != nil
				processed = c.process() != nil
			})
			if original == nil || !strings.Contains(original.Error(), tc.want) {
				t.Fatalf("initialization error = %v; want %q", original, tc.want)
			}
			if compiled || processed || c.err != original {
				t.Fatalf("lifecycle lost its initialization error: compile=%v, process=%v, original=%v, final=%v", compiled, processed, original, c.err)
			}
			if output != "" {
				t.Fatalf("failed lifecycle wrote to stdout: %q", output)
			}
		})
	}
}

func TestCompilerPropagatesSourceErrors(t *testing.T) {
	// Keep accidental exits or unbounded frontend recursion in a failing child
	// test, so a missing import cannot terminate the parent package test run.
	const childEnvironment = "L_COMPILER_LIFECYCLE_SOURCE_CHILD"
	if os.Getenv(childEnvironment) != "1" {
		binary, err := os.Executable()
		if err != nil {
			t.Fatal(err)
		}
		ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer cancel()
		command := exec.CommandContext(ctx, binary, "-test.run=^TestCompilerPropagatesSourceErrors$")
		command.Env = append(os.Environ(), childEnvironment+"=1")
		if output, err := command.CombinedOutput(); err != nil {
			t.Fatalf("source error lifecycle test exited instead of returning a usable error: %v\n%s", err, output)
		}
		return
	}
	for _, operation := range []string{"compile", "process"} {
		for _, tc := range []struct {
			name, source, want string
			missing            bool
		}{
			{"unterminated literal", "\"message\" is `unfinished.", "unterminated", false},
			{"missing import", "Use `absent.l`.\n", "absent.l", true},
		} {
			t.Run(operation+"/"+tc.name, func(t *testing.T) {
				directory := t.TempDir()
				path := filepath.Join(directory, "source.l")
				if err := os.WriteFile(path, []byte(tc.source), 0600); err != nil {
					t.Fatal(err)
				}
				c := newCompiler(&cliData{jobType: "build", mainPath: path})
				c.initialize(canonicalCompilerOptions(t))
				if c.err != nil {
					t.Fatalf("prepare compiler: %v", c.err)
				}
				returned := false
				output := lifecycleStdout(t, func() {
					if operation == "compile" {
						returned = c.compile() != nil
					} else {
						returned = c.process() != nil
					}
				})
				if returned || c.err == nil {
					t.Fatalf("%s returned=%v, error=%v; source errors must prevent a result", operation, returned, c.err)
				}
				if !strings.Contains(c.err.Error(), "compile:") || !strings.Contains(c.err.Error(), tc.want) || !strings.Contains(c.err.Error(), path) {
					t.Fatalf("source error lost phase, cause, or source path: %v", c.err)
				}
				if tc.missing && !errors.Is(c.err, os.ErrNotExist) {
					t.Fatalf("import error lost wrapped filesystem cause: %v", c.err)
				}
				if c.compiledBytes != nil || c.asmBytes != nil {
					t.Fatal("failed compilation published output bytes")
				}
				if output != "" {
					t.Fatalf("%s error path wrote to stdout: %q", operation, output)
				}
			})
		}
	}
}
