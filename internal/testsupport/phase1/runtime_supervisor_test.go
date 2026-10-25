//go:build darwin || linux

package phase1

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"syscall"
	"testing"
	"time"
)

func TestRuntimeOwnership(t *testing.T) {
	t.Run("released-records-never-signal", func(t *testing.T) {
		executable, err := os.Executable()
		if err != nil {
			t.Fatal(err)
		}
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		cmd := exec.Command(executable, "-test.run=^TestRuntimeOwnershipHelper$")
		cmd.Env = append(os.Environ(), "P1_OWNERSHIP_SELFTEST=1")
		result := RunRuntimeDriver(ctx, cmd)
		if err := result.Failure(); err != nil {
			t.Fatalf("successful driver: %v\n%s\n%s", err, result.Stdout, result.Stderr)
		}
		if string(result.Stdout) != "out\x00\n" || string(result.Stderr) != "err\x00\n" || len(result.SignaledGroups) != 0 {
			t.Fatalf("streams or no-fallback invariant: %+v", result)
		}
	})
	t.Run("closed-and-missing-admission", func(t *testing.T) {
		dir := t.TempDir()
		t.Setenv(RuntimeRegistryEnv, dir)
		for _, state := range []string{"missing", "open"} {
			if state == "open" {
				if err := os.WriteFile(filepath.Join(dir, "admission"), []byte(state), 0600); err != nil {
					t.Fatal(err)
				}
				if err := closeRuntimeAdmission(dir); err != nil {
					t.Fatal(err)
				}
			}
			lock, _, err := RegisterRuntimeHelper("late")
			if lock != nil || !errors.Is(err, ErrRuntimeAdmissionClosed) {
				t.Fatalf("late helper admitted: %v", err)
			}
			paths, callErr := filepath.Glob(filepath.Join(dir, "alive-*"))
			if callErr != nil {
				t.Fatal(callErr)
			}
			if len(paths) != 0 {
				t.Fatalf("late helper created lifetime records: %v", paths)
			}
			executable, err := os.Executable()
			if err != nil {
				t.Fatal(err)
			}
			ctx, cancel := context.WithTimeout(context.Background(), time.Second)
			cmd := exec.CommandContext(ctx, executable, "-test.run=^TestRuntimeOwnershipHelper$")
			cmd.Env = append(os.Environ(), "P1_OWNERSHIP_SELFTEST=1")
			output, err := cmd.CombinedOutput()
			cancel()
			if err != nil || len(output) != 0 {
				t.Fatalf("late helper entered its output behavior: %q %v", output, err)
			}
		}
	})
	t.Run("lifetime-lock-and-record", func(t *testing.T) {
		dir := t.TempDir()
		t.Setenv(RuntimeRegistryEnv, dir)
		if err := os.WriteFile(filepath.Join(dir, "admission"), []byte("open"), 0600); err != nil {
			t.Fatal(err)
		}
		file, record, err := RegisterRuntimeHelper("owned")
		if err != nil {
			t.Fatal(err)
		}
		defer file.Close() //nolint:errcheck // Fallback release may close a lifetime descriptor already closed explicitly by this test.
		if err := closeRuntimeAdmission(dir); err != nil {
			t.Fatal(err)
		}
		held, err := runtimeLifetimes(dir)
		if err != nil || len(held) != 1 || held[0] != record || record.PID != os.Getpid() {
			t.Fatalf("incomplete ownership evidence: %+v %v", held, err)
		}
		if callErr := file.Close(); callErr != nil {
			t.Fatal(callErr)
		}
		held, err = runtimeLifetimes(dir)
		if err != nil || len(held) != 0 {
			t.Fatalf("released record remained live: %+v %v", held, err)
		}
	})
	t.Run("admission-deadline", func(t *testing.T) {
		dir := t.TempDir()
		file, err := os.OpenFile(filepath.Join(dir, "admission"), os.O_CREATE|os.O_RDWR, 0600)
		if err != nil {
			t.Fatal(err)
		}
		defer func() {
			if err := file.Close(); err != nil {
				t.Errorf("close admission lock: %v", err)
			}
		}()
		if err := syscall.Flock(int(file.Fd()), syscall.LOCK_SH); err != nil {
			t.Fatal(err)
		}
		if err := closeRuntimeAdmission(dir); err == nil {
			t.Fatal("exclusive admission ignored held shared lock")
		}
	})
}

func TestRuntimeOwnershipHelper(t *testing.T) {
	if os.Getenv("P1_OWNERSHIP_SELFTEST") == "" {
		return
	}
	file, _, err := RegisterRuntimeHelper("released")
	if errors.Is(err, ErrRuntimeAdmissionClosed) {
		os.Exit(0)
	}
	if err != nil {
		t.Fatal(err)
	}
	// Leave a real, unlocked lifetime record. A record's existence alone must
	// not produce a leak or authorize a signal, even while its PID is live.
	if callErr := file.Close(); callErr != nil {
		t.Fatal(callErr)
	}
	if _, callErr := fmt.Fprint(os.Stdout, "out\x00\n"); callErr != nil {
		t.Fatal(callErr)
	}
	fmt.Fprint(os.Stderr, "err\x00\n")
	os.Exit(0)
}
