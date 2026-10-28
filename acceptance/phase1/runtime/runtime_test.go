//go:build darwin || linux

package phase1runtime

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	distribution "github.com/l-programming-language/l"
	nativetest "github.com/l-programming-language/l/internal/testsupport/native"
	"github.com/l-programming-language/l/internal/testsupport/phase1"
	Logiceme "github.com/l-programming-language/l/logiceme"
	Lyger "github.com/l-programming-language/l/lyger"
)

type runtimeCase struct {
	ID         string `json:"id"`
	Scenario   string `json:"scenario"`
	Assertions string `json:"assertions"`
}

// Each source remains inert until compiled beside actual production runtime
// bytes. Missing private hooks fail this prerequisite and every dependent ID.
func TestPhase1Runtime(t *testing.T) { runRuntimeCases(t, false) }

// Retain the six concurrency scenarios and deterministic barriers. This test
// does NOT claim race-detector, sanitizer or model-checking instrumentation.
func TestPhase1RuntimeConcurrent(t *testing.T) { runRuntimeCases(t, true) }

func runtimeProject(t testing.TB) (string, []runtimeCase, map[string][]byte) {
	t.Helper()
	root := phase1.RepoRoot(t)
	fixtures := filepath.Join(root, "testdata/phase1/runtime")
	data, e := os.ReadFile(filepath.Join(fixtures, "cases.json"))
	if e != nil {
		t.Fatal(e)
	}
	var cases []runtimeCase
	if e = json.Unmarshal(data, &cases); e != nil {
		t.Fatal(e)
	}
	if len(cases) != 98 {
		t.Fatalf("Rust runtime inventory=%d; want98", len(cases))
	}
	files := map[string][]byte{}
	for source, target := range map[string]string{"Cargo.toml.txt": "Cargo.toml", "Cargo.lock": "Cargo.lock", "driver.Cargo.toml.txt": "driver/Cargo.toml", "helper.Cargo.toml.txt": "companion/Cargo.toml", "driver.rs.txt": "driver/driver.rs", "checks.rs.txt": "driver/checks.rs", "support.rs.txt": "driver/support.rs", "api.rs.txt": "driver/runtime_api.rs", "fault.rs.txt": "driver/fault.rs", "helper.rs.txt": "companion/helper.rs"} {
		b, e := os.ReadFile(filepath.Join(fixtures, source))
		if e != nil {
			t.Fatal(e)
		}
		files[target] = b
	}
	ownership, e := os.ReadFile(filepath.Join(fixtures, "ownership.rs.txt"))
	if e != nil {
		t.Fatal(e)
	}
	files["driver/ownership.rs"] = ownership
	files["companion/ownership.rs"] = bytes.Clone(ownership)
	for source, target := range map[string]string{"builtins/runtime/support.aml": "runtime.aml", "builtins/templates/runtime/host.lu": "driver/runtime/host.rs", "builtins/templates/runtime/process_unix.lu": "driver/process.rs"} {
		b, e := fs.ReadFile(distribution.Files(), source)
		if e != nil {
			t.Fatal(e)
		}
		disk, e := os.ReadFile(filepath.Join(root, source))
		if e != nil || !bytes.Equal(b, disk) {
			t.Fatalf("production runtime provenance %s: %v", source, e)
		}
		files[target] = b
	}
	entry, e := Lyger.NewLygerService(Lyger.Options{Resources: distribution.Files()}).Compile([]byte("Show ``.\n"), "runtime-observer.l")
	if e != nil {
		t.Fatal(e)
	}
	unit, e := entry.OrchestratorUnit("runtime_observer", nil)
	if e != nil {
		t.Fatal(e)
	}
	_, bundle, e := Logiceme.NewLogicemeService(distribution.Files()).EvaluateWorkspaceEntry(unit.BackendAML, Logiceme.WorkspaceRuntime{Package: "runtime_support", RelativeDirectory: "../runtime_support"})
	if e != nil {
		t.Fatal(e)
	}
	for i, paths := range bundle.MapArrays {
		for j, path := range paths {
			if path == "/src/runtime.rs" {
				files["driver/runtime.rs"] = bundle.ByteArrays[i][j]
			}
		}
	}
	if len(files["driver/runtime.rs"]) == 0 {
		t.Fatal("canonical source composition supplied no runtime library")
	}
	files["driver/runtime.rs"] = append(files["driver/runtime.rs"], []byte("\n#[cfg(test)]\ninclude!(\"runtime_api.rs\");\n")...)
	return nativetest.Materialize(t, files), cases, files
}

func runtimeArtifact(r nativetest.BuildResult, name string) string {
	for _, a := range r.Artifacts {
		if a.TargetName == name {
			return a.Executable
		}
	}
	return ""
}
func runRuntimeCases(t *testing.T, concurrency bool) {
	project, cases, files := runtimeProject(t)
	helper := nativetest.Build(t, filepath.Join(project, "companion"), nativetest.BuildOptions{})
	helperPath := runtimeArtifact(helper, "helper")
	t.Run("helper-interface", func(t *testing.T) {
		if helper.Err != nil || helperPath == "" {
			t.Fatalf("native Rust helper build: %v\n%s\n%s", helper.Err, helper.Stdout, helper.Stderr)
		}
		nativetest.AssertNativeArtifact(t, helperPath)
	})
	driver := nativetest.Build(t, filepath.Join(project, "driver"), nativetest.BuildOptions{Tests: true})
	fault := nativetest.Build(t, filepath.Join(project, "driver"), nativetest.BuildOptions{Binary: "runtime-fault", TestConfiguration: true})
	driverPath := runtimeArtifact(driver, "l_runtime_parity_driver")
	faultPath := runtimeArtifact(fault, "runtime-fault")
	blocked := errors.Join(helper.Err, driver.Err, fault.Err)
	if blocked == nil && (driverPath == "" || faultPath == "" || helperPath == "") {
		blocked = fmt.Errorf("missing Cargo native artifact: driver=%q fault=%q helper=%q", driverPath, faultPath, helperPath)
	}
	diagnostic := fmt.Sprintf("driver: %v\n%s\n%s\nfault: %v\n%s\n%s", driver.Err, driver.Stdout, driver.Stderr, fault.Err, fault.Stdout, fault.Stderr)
	t.Run("runtime-interface", func(t *testing.T) {
		if blocked != nil {
			t.Fatalf("actual production Rust runtime does not satisfy private observer contract: %v\n%s\nruntime not exercised", blocked, diagnostic)
		}
		nativetest.AssertNativeArtifact(t, driverPath)
		nativetest.AssertNativeArtifact(t, faultPath)
	})
	if blocked == nil {
		dir := filepath.Dir(driverPath)
		phase1.CopyFile(t, helperPath, filepath.Join(dir, "companion/helper"), 0700)
		installedFault := filepath.Join(dir, "runtime-fault")
		phase1.CopyFile(t, faultPath, installedFault, 0700)
		faultPath = installedFault
	}
	if base := os.Getenv("L_HARNESS_EVIDENCE_DIR"); base != "" {
		dir := filepath.Join(base, strings.ReplaceAll(t.Name(), "/", "_"))
		for name, b := range files {
			path := filepath.Join(dir, "actual", filepath.FromSlash(name))
			if e := os.MkdirAll(filepath.Dir(path), 0700); e != nil {
				t.Fatal(e)
			}
			if e := os.WriteFile(path, b, 0600); e != nil {
				t.Fatal(e)
			}
		}
		if e := os.WriteFile(filepath.Join(dir, "cargo.txt"), []byte(diagnostic), 0600); e != nil {
			t.Fatal(e)
		}
	}
	selected := map[string]bool{"P1-AWAIT-15": true, "P1-AWAIT-18": true, "P1-STOP-08": true, "P1-RESULT-05": true, "P1-EXIT-09": true, "P1-EXIT-11": true}
	for _, item := range cases {
		tc := item
		if concurrency && !selected[tc.ID] {
			continue
		}
		t.Run(tc.ID, func(t *testing.T) {
			t.Logf("scenario: %s\nrequired assertions: %s", tc.Scenario, tc.Assertions)
			if blocked != nil {
				t.Fatalf("blocked by runtime-interface: %v; runtime not exercised", blocked)
			}
			ctx, cancel := context.WithTimeout(context.Background(), 8*time.Second)
			defer cancel()
			cmd := exec.Command(driverPath, "--exact", "checks::driver", "--nocapture", "--test-threads=1")
			cmd.Dir = t.TempDir()
			cmd.Env = append(os.Environ(), "P1_CASE="+tc.ID)
			result := phase1.RunRuntimeDriver(ctx, cmd)
			if e := result.Failure(); e != nil || !bytes.Contains(result.Stdout, []byte("RUST-PARITY-PASS "+tc.ID)) {
				t.Fatalf("actual Rust runtime assertion: %v\nstdout=%s\nstderr=%s", e, result.Stdout, result.Stderr)
			}
		})
	}
	if !concurrency {
		runtimeFaults(t, driverPath, faultPath, blocked)
	}
}

func runtimeFaults(t *testing.T, binary, faultBinary string, blocked error) {
	t.Run("fatal-controller-unwind", func(t *testing.T) {
		if blocked != nil {
			t.Fatalf("blocked by runtime-interface: %v; controller cleanup not exercised", blocked)
		}
		ctx, cancel := context.WithTimeout(context.Background(), 8*time.Second)
		defer cancel()
		cmd := exec.Command(binary, "--exact", "checks::fatal_controller_cleanup", "--nocapture", "--test-threads=1")
		cmd.Dir = t.TempDir()
		cmd.Env = append(os.Environ(), "P1_FATAL_CONTROLLER_FAULT="+filepath.Join(cmd.Dir, "fault-pid"))
		result := phase1.RunRuntimeDriver(ctx, cmd)
		if result.Failure() != nil || ctx.Err() != nil || len(result.SignaledGroups) != 0 {
			t.Fatalf("fatal controller cleanup regression: %+v\nstdout=%s\nstderr=%s", result, result.Stdout, result.Stderr)
		}
		if bytes.Count(result.Stdout, []byte("fatal child cleanup reaped once")) != 1 ||
			!bytes.Contains(result.Stdout, []byte("fatal controller direct child already reaped after failure")) ||
			!bytes.Contains(result.Stderr, []byte("intentional fatal controller failure before checkpoint release")) {
			t.Fatalf("controller ownership or original panic evidence lost: %s\n%s", result.Stdout, result.Stderr)
		}
	})
	for _, mode := range []string{"wait-watchdog", "wait-early"} {
		mode := mode
		t.Run("fatal-wait/"+mode, func(t *testing.T) {
			if blocked != nil {
				t.Fatalf("blocked by runtime-interface: %v; fatal-wait not exercised", blocked)
			}
			ctx, cancel := context.WithTimeout(context.Background(), 8*time.Second)
			defer cancel()
			cmd := exec.Command(binary, "--exact", "checks::fatal_wait_fault", "--nocapture", "--test-threads=1")
			cmd.Dir = t.TempDir()
			cmd.Env = append(os.Environ(), "P1_FATAL_WAIT_FAULT="+mode)
			result := phase1.RunRuntimeDriver(ctx, cmd)
			var exit *exec.ExitError
			if !errors.As(result.Err, &exit) || exit.ExitCode() != 101 || ctx.Err() != nil {
				t.Fatalf("expected Rust assertion panic: %v\n%s\n%s", result.Failure(), result.Stdout, result.Stderr)
			}
			if result.CleanupErr != nil || len(result.Leaks) != 0 || len(result.SignaledGroups) != 0 {
				t.Fatalf("fatal wait cleanup required fallback: %+v", result)
			}
			needle := "fatal-path watchdog: orchestrator did not exit"
			if mode == "wait-early" {
				needle = "intentional failure before fatal wait selection"
			}
			if !bytes.Contains(result.Stdout, []byte("fatal child cleanup reaped once")) || !bytes.Contains(result.Stdout, []byte("fatal wait fault armed: "+mode)) || !bytes.Contains(result.Stderr, []byte(needle)) {
				t.Fatalf("wrong Rust panic path: %s\n%s", result.Stdout, result.Stderr)
			}
		})
	}
	for _, mode := range []string{"rust-panic", "outer-timeout", "abrupt-start", "nested"} {
		mode := mode
		t.Run("supervision/"+mode, func(t *testing.T) {
			if blocked != nil {
				t.Fatalf("blocked by runtime-interface: %v; supervision not exercised", blocked)
			}
			ctx, cancel := context.WithTimeout(context.Background(), 8*time.Second)
			defer cancel()
			faultContext, cancelFault := context.WithCancel(ctx)
			defer cancelFault()
			cmd := exec.Command(faultBinary)
			cmd.Dir = t.TempDir()
			armed := filepath.Join(cmd.Dir, "fault-armed")
			cmd.Env = append(os.Environ(), "P1_SUPERVISION_FAULT="+mode, "P1_SUPERVISION_ARMED="+armed)
			watched := make(chan struct{})
			go func() {
				defer close(watched)
				for {
					if _, e := os.Stat(armed); e == nil {
						break
					}
					select {
					case <-faultContext.Done():
						return
					case <-time.After(time.Millisecond):
					}
				}
				limit := 2 * time.Second
				if mode == "outer-timeout" {
					limit = 500 * time.Millisecond
				}
				select {
				case <-faultContext.Done():
				case <-time.After(limit):
					cancelFault()
				}
			}()
			result := phase1.RunRuntimeDriver(faultContext, cmd)
			cancelFault()
			<-watched
			if ctx.Err() != nil {
				t.Fatalf("setup watchdog before intended fault: %v", result.Failure())
			}
			if _, e := os.Stat(armed); e != nil {
				t.Fatalf("fault not armed after real helper ownership: %v\n%s\n%s", e, result.Stdout, result.Stderr)
			}
			wantLeaks, wantGroups := 1, 1
			if mode == "nested" {
				wantLeaks, wantGroups = 3, 2
			}
			if result.Err == nil || result.CleanupErr != nil || len(result.Leaks) != wantLeaks || len(result.SignaledGroups) != wantGroups {
				t.Fatalf("original failure/ownership evidence lost: %+v", result)
			}
			if !bytes.Equal(result.Stdout, []byte("out\x00\n")) || !bytes.HasPrefix(result.Stderr, []byte("err\x00\n")) {
				t.Fatalf("fault streams changed: %q/%q", result.Stdout, result.Stderr)
			}
			if mode == "rust-panic" {
				if !bytes.Contains(result.Stderr, []byte("intentional Rust supervision panic")) {
					t.Fatalf("missing native panic: %s", result.Stderr)
				}
			} else if !bytes.Equal(result.Stderr, []byte("err\x00\n")) {
				t.Fatalf("unexpected diagnostics: %q", result.Stderr)
			}
			if _, e := os.Stat(result.Registry); !os.IsNotExist(e) {
				t.Fatalf("cleaned ownership registry remains: %v", e)
			}
		})
	}
}
