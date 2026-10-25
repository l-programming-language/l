//go:build darwin || linux

package main

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"syscall"
	"testing"
	"time"

	nativetest "github.com/l-programming-language/l/internal/testsupport/native"
	"github.com/l-programming-language/l/internal/testsupport/phase1"
	"github.com/l-programming-language/l/lerrors"
)

func TestCargoToolHelper(t *testing.T) {
	mode := os.Getenv("L_TEST_CARGO_MODE")
	if mode == "" {
		return
	}
	if os.Getenv(phase1.RuntimeRegistryEnv) != "" {
		lifetime, _, err := phase1.RegisterRuntimeHelper("controlled Cargo: " + mode)
		if err != nil {
			panic(err)
		}
		defer func() {
			if err := lifetime.Close(); err != nil {
				t.Errorf("close test resource: %v", err)
			}
		}()
	}
	switch mode {
	case "failure":
		if _, callErr := os.Stderr.WriteString("controlled compiler failure\n"); callErr != nil {
			t.Fatal(callErr)
		}
		os.Exit(7)
	case "descendant", "failed descendant", "closed descendant", "failed closed descendant", "cancel":
		executable, err := os.Executable()
		if err != nil {
			panic(err)
		}
		child := exec.Command(executable, "-test.run=^TestCargoToolHelper$")
		child.Env = append(os.Environ(), "L_TEST_CARGO_MODE=hold")
		if !strings.Contains(mode, "closed") {
			child.Stdout, child.Stderr = os.Stdout, os.Stderr
		}
		if err := child.Start(); err != nil {
			panic(err)
		}
		for i := 0; i < 500; i++ {
			if data, err := os.ReadFile(os.Getenv("L_TEST_CARGO_READY")); err == nil && string(data) == strconv.Itoa(child.Process.Pid) {
				if mode == "cancel" {
					time.Sleep(30 * time.Second)
					os.Exit(9)
				}
				if _, callErr := os.Stdout.WriteString("controlled descendant output\n\n"); callErr != nil {
					t.Fatal(callErr)
				}
				if _, callErr := os.Stderr.WriteString("controlled descendant error\n"); callErr != nil {
					t.Fatal(callErr)
				}
				if strings.HasPrefix(mode, "failed") {
					os.Exit(7)
				}
				if mode == "closed descendant" {
					if err := os.WriteFile(os.Getenv("L_TEST_CARGO_OUTPUT"), mustCargoArtifact(), 0700); err != nil {
						panic(err)
					}
					emitCargoArtifact()
				}
				os.Exit(0)
			}
			time.Sleep(10 * time.Millisecond)
		}
		panic("helper descendant did not start")
	case "hold":
		file, err := os.OpenFile(os.Getenv("L_TEST_CARGO_LOCK"), os.O_CREATE|os.O_RDWR, 0600)
		if err != nil {
			panic(err)
		}
		if err := syscall.Flock(int(file.Fd()), syscall.LOCK_EX); err != nil {
			panic(err)
		}
		if err := os.WriteFile(os.Getenv("L_TEST_CARGO_READY"), []byte(strconv.Itoa(os.Getpid())), 0600); err != nil {
			panic(err)
		}
		time.Sleep(30 * time.Second)
		if callErr := file.Close(); callErr != nil {
			t.Fatal(callErr)
		}
		os.Exit(0)
	case "success", "interleaved diagnostics", "nonexecutable", "empty":
		content, readErr := os.ReadFile(os.Getenv("L_TEST_CARGO_ARTIFACT"))
		if readErr != nil {
			panic(readErr)
		}
		if mode == "empty" {
			content = nil
		}
		permission := os.FileMode(0700)
		if mode == "nonexecutable" {
			permission = 0600
		}
		if err := os.WriteFile(os.Getenv("L_TEST_CARGO_OUTPUT"), content, permission); err != nil {
			panic(err)
		}
	}
	if mode == "success" || mode == "interleaved diagnostics" {
		emitCargoArtifact()
		cwd, callErr := os.Getwd()
		if callErr != nil {
			t.Fatal(callErr)
		}
		report := map[string]string{"cwd": cwd}
		for _, key := range []string{"RUSTFLAGS", "CARGO_ENCODED_RUSTFLAGS", "RUSTC_WRAPPER", "RUSTC_WORKSPACE_WRAPPER", "CARGO_BUILD_TARGET", "CARGO_NET_OFFLINE", "RUSTUP_TOOLCHAIN"} {
			report[key] = os.Getenv(key)
		}
		content, callErr := json.Marshal(report)
		if callErr != nil {
			t.Fatal(callErr)
		}
		if err := os.WriteFile(os.Getenv("L_TEST_CARGO_REPORT"), content, 0600); err != nil {
			panic(err)
		}
	}
	os.Exit(0)
}

func TestCargoBuildToolFailuresAndIsolation(t *testing.T) {
	const isolatedCase = "L_TEST_CARGO_ISOLATED_CASE"
	isolated := strings.HasPrefix(os.Getenv(isolatedCase), t.Name()+"/")
	var seed string
	if isolated {
		seed = os.Getenv("L_TEST_CARGO_ISOLATED_SEED")
		if seed == "" {
			t.Fatal("isolated tool case has no independent native seed")
		}
	} else {
		t.Parallel()
		seed = independentCargoArtifact(t)
	}
	for _, mode := range []string{"success", "interleaved diagnostics", "failure", "missing", "absent output", "nonexecutable", "empty", "pre-cancel", "cancel", "descendant", "failed descendant", "closed descendant", "failed closed descendant"} {
		mode := mode
		t.Run(mode, func(t *testing.T) {
			if !isolated {
				t.Parallel()
				buildCLICases <- struct{}{}
				defer func() { <-buildCLICases }()
				runIsolatedCargoCase(t, isolatedCase, seed)
				return
			}
			if os.Getenv(isolatedCase) != t.Name() {
				t.Fatal("isolated tool case selection changed")
			}
			dir := filepath.Join(t.TempDir(), "build tool with spaces")
			if err := os.Mkdir(dir, 0700); err != nil {
				t.Fatal(err)
			}
			dir, err := filepath.EvalSymlinks(dir)
			if err != nil {
				t.Fatal(err)
			}
			binary, err := os.Executable()
			if err != nil {
				t.Fatal(err)
			}
			t.Setenv("L_TEST_CARGO_BINARY", binary)
			t.Setenv("L_TEST_CARGO_ARTIFACT", seed)
			t.Setenv("L_TEST_CARGO_MODE", mode)
			output := filepath.Join(dir, "result")
			t.Setenv("L_TEST_CARGO_OUTPUT", filepath.Join(dir, "controlled-cargo-artifact"))
			reportPath := filepath.Join(dir, "report.json")
			t.Setenv("L_TEST_CARGO_REPORT", reportPath)
			argsPath := filepath.Join(dir, "args")
			t.Setenv("L_TEST_CARGO_ARGS", argsPath)
			ready := filepath.Join(dir, "ready")
			lockPath := filepath.Join(dir, "lock")
			t.Setenv("L_TEST_CARGO_READY", ready)
			t.Setenv("L_TEST_CARGO_LOCK", lockPath)
			childStopped := false
			t.Cleanup(func() {
				if childStopped {
					return
				}
				if data, err := os.ReadFile(ready); err == nil {
					if pid, err := strconv.Atoi(string(data)); err == nil && pid > 1 {
						_ = syscall.Kill(pid, syscall.SIGKILL) //nolint:errcheck // Fallback may race process exit; lifetime and Wait assertions remain authoritative.
					}
				}
			})
			if mode != "missing" {
				script := "#!/bin/sh\nif [ \"$2\" = clean ]; then exit 0; fi\nprintf '%s\\n' \"$@\" > \"$L_TEST_CARGO_ARGS\"\nexec \"$L_TEST_CARGO_BINARY\" -test.run='^TestCargoToolHelper$'\n"
				if err := os.WriteFile(filepath.Join(dir, "cargo"), []byte(script), 0700); err != nil {
					t.Fatal(err)
				}
			}
			t.Setenv("PATH", dir)
			for key, value := range map[string]string{"RUSTFLAGS": "--invalid", "CARGO_ENCODED_RUSTFLAGS": "--invalid", "RUSTC_WRAPPER": "/absent", "RUSTC_WORKSPACE_WRAPPER": "/absent", "CARGO_BUILD_TARGET": "invalid-target", "CARGO_NET_OFFLINE": "false", "RUSTUP_TOOLCHAIN": "invalid-toolchain"} {
				t.Setenv(key, value)
			}
			bundle, err := plainSourceBundle("fn main() {}\n")
			if err != nil {
				t.Fatal(err)
			}
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()
			if mode == "pre-cancel" {
				cancel()
			}
			cancelled := make(chan struct{})
			if mode == "cancel" {
				go func() {
					defer close(cancelled)
					for {
						if _, err := os.Stat(ready); err == nil {
							cancel()
							return
						}
						select {
						case <-ctx.Done():
							return
						case <-time.After(10 * time.Millisecond):
						}
					}
				}()
			} else {
				close(cancelled)
			}
			sources := filepath.Join(dir, "sources")
			started := time.Now()
			err = compileCargoWorkspace(ctx, nil, bundle, nil, nil, sources, nil, output)
			<-cancelled
			if strings.Contains(mode, "descendant") || mode == "cancel" {
				if time.Since(started) > 8*time.Second {
					t.Fatalf("supervised command did not return promptly: %v", err)
				}
				if strings.HasPrefix(mode, "failed") {
					var exit *exec.ExitError
					if !errors.As(err, &exit) || exit.ExitCode() != 7 {
						t.Fatalf("lost failed parent's exit status: %v", err)
					}
					for _, text := range []string{"controlled descendant output\n\n", "controlled descendant error\n"} {
						if !strings.Contains(err.Error(), text) {
							t.Fatalf("lost captured diagnostics %q: %v", text, err)
						}
					}
				}
				file, openErr := os.OpenFile(lockPath, os.O_RDWR, 0600)
				if openErr != nil {
					t.Fatal(openErr)
				}
				defer func() {
					if err := file.Close(); err != nil {
						t.Errorf("close test resource: %v", err)
					}
				}()
				deadline := time.Now().Add(3 * time.Second)
				for {
					if err := syscall.Flock(int(file.Fd()), syscall.LOCK_EX|syscall.LOCK_NB); err == nil {
						childStopped = true
						break
					}
					if time.Now().After(deadline) {
						t.Fatal("supervised descendant survived")
					}
					time.Sleep(10 * time.Millisecond)
				}
			}
			if mode == "closed descendant" {
				if err != nil {
					t.Fatalf("successful parent became a build failure: %v", err)
				}
				return
			}
			if mode != "success" && mode != "interleaved diagnostics" {
				if !errors.Is(err, lerrors.Tool) {
					t.Errorf("Cargo build failure lost Tool classification: %v", err)
				}
				if err == nil {
					t.Fatal("expected Cargo build failure")
				}
				if mode == "failure" && !strings.Contains(err.Error(), "controlled compiler failure") {
					t.Fatalf("lost diagnostics: %v", err)
				}
				if (mode == "cancel" || mode == "pre-cancel") && !errors.Is(err, context.Canceled) {
					t.Fatalf("lost cancellation: %v", err)
				}
				if mode == "cancel" {
					var exit *exec.ExitError
					if !errors.As(err, &exit) {
						t.Fatalf("lost interrupted child cause: %v", err)
					}
				}
				if mode == "pre-cancel" {
					if _, err := os.Stat(argsPath); !errors.Is(err, os.ErrNotExist) {
						t.Fatalf("already-canceled build started the tool: %v", err)
					}
				}
				if mode == "descendant" && !errors.Is(err, exec.ErrWaitDelay) {
					t.Fatalf("lost pipe timeout: %v", err)
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			encoded, err := os.ReadFile(reportPath)
			if err != nil {
				t.Fatal(err)
			}
			var report map[string]string
			if err := json.Unmarshal(encoded, &report); err != nil {
				t.Fatal(err)
			}
			for key, want := range map[string]string{"cwd": sources, "RUSTFLAGS": "", "CARGO_ENCODED_RUSTFLAGS": "", "RUSTC_WRAPPER": "", "RUSTC_WORKSPACE_WRAPPER": "", "CARGO_BUILD_TARGET": "", "CARGO_NET_OFFLINE": "true"} {
				if report[key] != want {
					t.Errorf("%s=%q; want %q", key, report[key], want)
				}
			}
			args, err := os.ReadFile(argsPath)
			if err != nil {
				t.Fatal(err)
			}
			words := strings.Split(strings.TrimSpace(string(args)), "\n")
			if len(words) > 0 && strings.HasPrefix(words[0], "+") {
				if words[0] != "+1.98.1" {
					t.Fatalf("wrong Cargo toolchain selector %q", words[0])
				}
				words = words[1:]
			}
			if len(words) == 0 || words[0] != "build" {
				t.Fatalf("Cargo command = %q", words)
			}
			for _, required := range []string{"--frozen", "--message-format=json"} {
				found := false
				for _, arg := range words {
					if arg == required {
						found = true
					}
				}
				if !found {
					t.Errorf("Cargo command missing %s: %q", required, words)
				}
			}
			if value := report["RUSTUP_TOOLCHAIN"]; value != "" && value != "1.98.1" {
				t.Errorf("unisolated Rust toolchain: %q", value)
			}
			nativetest.AssertNativeArtifact(t, output)
		})
	}
}

// Re-enter the same assertion body with process-local environment changes. The
// outer t.Run remains the reported scenario; no case result is reused. Native
// tool helpers register with the supervisor before they can start descendants.
func runIsolatedCargoCase(t *testing.T, marker, seed string) {
	t.Helper()
	binary, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	parts := strings.Split(t.Name(), "/")
	for index, part := range parts {
		parts[index] = "^" + regexp.QuoteMeta(part) + "$"
	}
	command := exec.Command(binary, "-test.run="+strings.Join(parts, "/"), "-test.v", "-test.count=1", "-test.timeout=20s")
	command.Env = append(os.Environ(), marker+"="+t.Name(), "L_TEST_CARGO_ISOLATED_SEED="+seed)
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	result := phase1.RunRuntimeDriver(ctx, command)
	if err := result.Failure(); err != nil {
		t.Fatalf("isolated tool assertions: %v\nstdout:\n%s\nstderr:\n%s", err, result.Stdout, result.Stderr)
	}
	output := string(result.Stdout)
	if !strings.Contains(output, "=== RUN   "+t.Name()+"\n") || !strings.Contains(output, "--- PASS: "+t.Name()+" (") {
		t.Fatalf("isolated tool process did not report the selected assertion body %q:\n%s", t.Name(), output)
	}
}

func mustCargoArtifact() []byte {
	b, e := os.ReadFile(os.Getenv("L_TEST_CARGO_ARTIFACT"))
	if e != nil {
		panic(e)
	}
	return b
}

func emitCargoArtifact() {
	cwd, callErr := os.Getwd()
	if callErr != nil {
		panic(callErr)
	}
	manifest := filepath.Join(cwd, "Cargo.toml")
	arguments, err := os.ReadFile(os.Getenv("L_TEST_CARGO_ARGS"))
	if err != nil {
		panic(err)
	}
	words := strings.Split(strings.TrimSpace(string(arguments)), "\n")
	for index, word := range words {
		if word == "--manifest-path" && index+1 < len(words) {
			manifest = words[index+1]
		}
	}
	manifest = filepath.Join(filepath.Dir(manifest), "orchestrator", "Cargo.toml")
	message, err := json.Marshal(map[string]any{"reason": "compiler-artifact", "package_id": "path+file:///controlled#cli_parity@0.0.0", "manifest_path": manifest, "target": map[string]any{"name": "cli_parity", "kind": []string{"bin"}}, "executable": os.Getenv("L_TEST_CARGO_OUTPUT")})
	if err != nil {
		panic(err)
	}
	if os.Getenv("L_TEST_CARGO_MODE") == "interleaved diagnostics" {
		if _, callErr := os.Stdout.Write(message[:len(message)/2]); callErr != nil {
			panic(callErr)
		}
		if _, callErr := os.Stderr.WriteString("controlled compiler progress\n"); callErr != nil {
			panic(callErr)
		}
		if _, callErr := os.Stdout.Write(append(message[len(message)/2:], '\n')); callErr != nil {
			panic(callErr)
		}
		return
	}
	if _, callErr := os.Stdout.Write(append(message, '\n')); callErr != nil {
		panic(callErr)
	}
}
