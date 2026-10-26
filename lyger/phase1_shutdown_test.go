//go:build darwin || linux

package Lyger

import (
	"bytes"
	"context"
	"crypto/sha256"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/l-programming-language/l/internal/testsupport/phase1"
)

// These cases exercise generated main unchanged, with the production runtime
// emitted by logeme. Instrumentation supplies only dependency barriers and real
// signal delivery; it does not replace main, Shutdown, or process exit.
func TestPhase1GeneratedShutdown(t *testing.T) {
	t.Parallel()
	const source = "Use IO's Print from `io` as \"Print\".\n" + `"message" is ` + "`child`" + `.

    Create "First" as a New Public Basic App.
The App Prints message.
Show ` + "`unexpected-first-continuation|`" + `.

    Create "Second" as a New Public Basic App.
The App Prints message.
Show ` + "`unexpected-second-continuation|`" + `.

Run First.
Run Second.
Show ` + "`root|`" + `.
`
	child := buildShutdownCompanion(t)
	for _, completion := range []string{"fallthrough", "bare-exit", "function-exit"} {
		t.Run(completion, func(t *testing.T) {
			input := source
			if completion == "bare-exit" {
				input += "Exit.\n"
			} else if completion == "function-exit" {
				input += "\n    Publicly create \"Finish\" as a New Function with No Inputs and No Outputs.\nExit.\n\n    Publicly create \"Relay\" as a New Function with No Inputs and No Outputs.\nCall Finish.\nShow `unexpected-function|`.\n\nCall Relay.\nShow `unexpected-root|`.\n"
			}
			result := phase1Require(t, input, nil)
			// Only the runtime mode changes below. Compile the actual generated
			// project and its observer once for this completion/source group.
			built, _ := buildOrchestratorForTest(t, result, orchestratorTestBuild{instrumentation: shutdownInstrumentation, race: false})
			executable, err := os.ReadFile(built)
			if err != nil {
				t.Fatal(err)
			}
			buildOwner := t.Name()
			t.Cleanup(func() {
				got, err := os.ReadFile(built)
				if err != nil || !bytes.Equal(got, executable) {
					t.Errorf("shared native build changed after runtime profiles: %v", err)
				}
			})

			for _, tc := range []struct {
				mode string
				code int
			}{
				{"sigint", 130}, {"sigterm", 143}, {"clean", 0},
				{"cleanup-error", 1}, {"panic", 101}, {"panic-nil", 101},
			} {
				t.Run(tc.mode, func(t *testing.T) {
					// Each process still owns a fresh executable/companion directory;
					// only immutable bytes from the parent Cargo build are reused.
					binary := filepath.Join(t.TempDir(), filepath.Base(built))
					if err := os.WriteFile(binary, executable, 0700); err != nil {
						t.Fatal(err)
					}
					if got, err := os.ReadFile(binary); err != nil || !bytes.Equal(got, executable) {
						t.Fatalf("runtime executable differs from parent Cargo artifact: %v", err)
					}
					t.Logf("native build evidence: %s; executable SHA256 %x", buildOwner, sha256.Sum256(executable))
					for _, artifact := range executableTestCompanions(result) {
						path := filepath.Join(filepath.Dir(binary), artifact.RelativeExecutable)
						if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
							t.Fatal(err)
						}
						if err := os.WriteFile(path, child, 0700); err != nil {
							t.Fatal(err)
						}
						if got, err := os.ReadFile(path); err != nil || !bytes.Equal(got, child) {
							t.Fatalf("runtime companion differs from parent Cargo artifact: %v", err)
						}
					}
					dir := t.TempDir()
					ctx, cancel := context.WithTimeout(context.Background(), 8*time.Second)
					defer cancel()
					command := exec.Command(binary)
					command.Env = append(os.Environ(), "P1_SHUTDOWN_DIRECTORY="+dir, "P1_SHUTDOWN_MODE="+tc.mode,
						"RUST_BACKTRACE=0")
					run := phase1.RunRuntimeDriver(ctx, command)
					err := run.Err
					stdout, stderr := string(run.Stdout), string(run.Stderr)
					if len(run.Leaks) != 0 || len(run.SignaledGroups) != 0 || run.CleanupErr != nil {
						t.Errorf("runtime leaked before outer fallback: %+v; groups=%v cleanup=%v", run.Leaks, run.SignaledGroups, run.CleanupErr)
					}

					for _, label := range []string{"first", "first.descendant", "second", "second.descendant"} {
						data, err := os.ReadFile(filepath.Join(dir, label+".ready"))
						if err != nil {
							t.Errorf("missing companion readiness %s: %v", label, err)
							continue
						}
						var pid, group int
						var record string
						if _, err := fmt.Sscanf(string(data), "%d %d %s", &pid, &group, &record); err != nil || pid <= 1 || group <= 1 || !strings.HasPrefix(record, "alive-") {
							t.Errorf("invalid companion ownership for %s: %q", label, data)
						}
					}
					if ctx.Err() != nil {
						t.Fatalf("shutdown watchdog: %v; stdout=%q stderr=%q", err, stdout, stderr)
					}
					code := 0
					if err != nil {
						var exit *exec.ExitError
						if !errors.As(err, &exit) {
							t.Fatalf("execute generated program: %v", err)
						}
						code = exit.ExitCode()
					}
					events, readErr := os.ReadFile(filepath.Join(dir, "events"))
					if readErr != nil {
						t.Fatal(readErr)
					}
					for _, label := range []string{"first", "second"} {
						if n := strings.Count(string(events), "wait "+label+"\n"); n != 1 {
							t.Errorf("OS Wait count for %s = %d, want one; events=%q", label, n, events)
						}
					}
					if strings.HasPrefix(tc.mode, "sig") && !strings.Contains(string(events), "accepted "+strconv.Itoa(tc.code)+"\n") {
						t.Fatalf("signal was not accepted before cleanup release: %q", events)
					}
					if stdout != "root|" {
						t.Errorf("application stdout/continuation escaped: %q", stdout)
					}
					switch tc.mode {
					case "cleanup-error":
						if stderr != "terminate application process group: injected cleanup failure\n" {
							t.Errorf("cleanup diagnostic = %q", stderr)
						}
					case "panic":
						if !strings.Contains(stderr, "injected source panic") {
							t.Errorf("panic was suppressed or replaced: %q", stderr)
						}
					case "panic-nil":
						if !strings.Contains(stderr, "Box<dyn Any>") {
							t.Errorf("non-string Rust panic was suppressed or replaced: %q", stderr)
						}
					default:
						if len(stderr) != 0 {
							t.Errorf("stderr = %q, want empty", stderr)
						}
					}
					if code != tc.code {
						t.Errorf("generated main exit status = %d, want %d; events=%q stdout=%q stderr=%q", code, tc.code, events, stdout, stderr)
					}
				})
			}
		})
	}
}

func buildShutdownCompanion(t *testing.T) []byte {
	t.Helper()
	binary := buildNativeHelper(t)
	data, err := os.ReadFile(binary)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("native companion build: %s; SHA256 %x", t.Name(), sha256.Sum256(data))
	t.Cleanup(func() {
		got, err := os.ReadFile(binary)
		if err != nil || !bytes.Equal(got, data) {
			t.Errorf("shared native companion changed after shutdown profiles: %v", err)
		}
	})
	return data
}

// Names the independently authored dependency/output adapter. It invokes the
// actual generated l_entry_with_runtime seam; it supplies no main/l_main body.
const shutdownInstrumentation = "native/observer/main.rs"
