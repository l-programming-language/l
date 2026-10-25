//go:build darwin || linux

package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"testing"
	"time"
)

// The lock is released on termination even if the orphan remains a zombie.
func TestWaitDelayDescendant(t *testing.T) {
	path := os.Getenv("L_HARNESS_DESCENDANT_LOCK")
	if path == "" {
		return
	}
	file, err := os.OpenFile(path, os.O_WRONLY, 0600)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := file.Close(); err != nil {
			t.Errorf("close descendant lock: %v", err)
		}
	})
	if err := syscall.Flock(int(file.Fd()), syscall.LOCK_EX); err != nil {
		t.Fatal(err)
	}
	if _, err := fmt.Fprint(file, os.Getpid()); err != nil {
		t.Fatal(err)
	}
	time.Sleep(30 * time.Second)
	os.Exit(0)
}

func TestRunCommandTerminatesDescendantsAfterWaitDelay(t *testing.T) {
	dir := t.TempDir()
	lockPath := filepath.Join(dir, "descendant.lock")
	file, err := os.Create(lockPath)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := file.Close(); err != nil {
			t.Errorf("close descendant lock: %v", err)
		}
	})
	exe, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	r := runCommand(context.Background(), dir, commandSpec{
		Name: "orphan", Executable: "/bin/sh",
		Args: []string{"-c", `"$1" -test.run=^TestWaitDelayDescendant$ & while ! test -s "$2"; do sleep 0.01; done`, "harness-parent", exe, lockPath},
		Dir:  dir, Env: append(os.Environ(), "L_HARNESS_DESCENDANT_LOCK="+lockPath),
		Timeout: 10 * time.Second, TestJSON: true,
	})
	data, err := os.ReadFile(lockPath)
	if err != nil {
		t.Fatal(err)
	}
	pid, err := strconv.Atoi(string(data))
	if err != nil {
		t.Fatalf("descendant did not acquire the lock: %q; command: %+v", data, r)
	}
	deadline := time.Now().Add(time.Second)
	for {
		err = syscall.Flock(int(file.Fd()), syscall.LOCK_EX|syscall.LOCK_NB)
		if err == nil {
			break
		}
		if !errors.Is(err, syscall.EWOULDBLOCK) || time.Now().After(deadline) {
			// Avoid leaving the fixture alive when the regression is present.
			_ = syscall.Kill(pid, syscall.SIGKILL) //nolint:errcheck // Fallback cleanup after lifetime assertions; the process may already have exited.
			t.Fatalf("descendant survived WaitDelay cleanup: %v", err)
		}
		time.Sleep(10 * time.Millisecond)
	}
	if r.ExitCode != 0 || r.TimedOut || r.succeeded() || !strings.Contains(r.Error, exec.ErrWaitDelay.Error()) {
		t.Fatalf("WaitDelay failure was not preserved: %+v", r)
	}
}

func TestRunCommandCleansCompletedProcessGroups(t *testing.T) {
	exe, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	const output = "{\"Action\":\"output\",\"Package\":\"fixture\",\"Output\":\"parent output\\n\"}\n"
	for _, tc := range []struct {
		name        string
		code        int
		json, close bool
	}{
		{"failed parent with held pipe", 7, true, false},
		{"successful parent with held pipe", 0, true, false},
		{"successful parent with closed pipes", 0, true, true},
		{"failed parent with closed pipes", 7, true, true},
		{"successful parent with direct files", 0, false, false},
		{"failed parent with direct files", 7, false, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			dir := t.TempDir()
			lockPath := filepath.Join(dir, "descendant.lock")
			file, err := os.Create(lockPath)
			if err != nil {
				t.Fatal(err)
			}
			t.Cleanup(func() {
				if err := file.Close(); err != nil {
					t.Errorf("close descendant lock: %v", err)
				}
			})
			// Install fallback cleanup before starting the parent, including when
			// a later stream/status assertion fails while the regression is present.
			t.Cleanup(func() {
				data, readErr := os.ReadFile(lockPath)
				if readErr != nil {
					t.Errorf("read lifetime lock for fallback cleanup: %v", readErr)
					return
				}
				if pid, err := strconv.Atoi(string(data)); err == nil {
					_ = syscall.Kill(pid, syscall.SIGKILL) //nolint:errcheck // Fallback cleanup after lifetime assertions; the process may already have exited.
				}
			})
			child := `"$1" -test.run=^TestWaitDelayDescendant$`
			if tc.close {
				child += ` >/dev/null 2>&1`
			}
			script := child + ` & while ! test -s "$2"; do sleep 0.01; done; printf '%s' "$3"; printf 'parent stderr\r\n' >&2; exit "$4"`
			r := runCommand(context.Background(), dir, commandSpec{
				Name: "completed", Executable: "/bin/sh",
				Args: []string{"-c", script, "harness-parent", exe, lockPath, output, strconv.Itoa(tc.code)},
				Dir:  dir, Env: append(os.Environ(), "L_HARNESS_DESCENDANT_LOCK="+lockPath),
				Timeout: 10 * time.Second, TestJSON: tc.json,
			})
			data, err := os.ReadFile(lockPath)
			if err != nil {
				t.Fatal(err)
			}
			if _, err := strconv.Atoi(string(data)); err != nil {
				t.Fatalf("descendant did not announce lock ownership: %q; result: %+v", data, r)
			}
			if r.ExitCode != tc.code || r.TimedOut || r.DurationMS >= 10000 {
				t.Errorf("parent status or bounded completion lost: %+v", r)
			}
			wantError := ""
			if tc.code != 0 {
				wantError = "exit status 7"
			} else if tc.json && !tc.close {
				wantError = exec.ErrWaitDelay.Error()
			}
			if r.Error != wantError || r.succeeded() != (wantError == "") {
				t.Errorf("completion result changed: %+v; want error %q", r, wantError)
			}
			for path, want := range map[string]string{r.StdoutPath: output, r.StderrPath: "parent stderr\r\n"} {
				got, err := os.ReadFile(path)
				if err != nil || string(got) != want {
					t.Errorf("captured stream %s = %q, %v; want %q", path, got, err, want)
				}
			}
			if tc.json && (r.Events == nil || r.Events.ParseError != "") {
				t.Errorf("valid test events were lost: %+v", r.Events)
			}
			deadline := time.Now().Add(time.Second)
			for {
				err = syscall.Flock(int(file.Fd()), syscall.LOCK_EX|syscall.LOCK_NB)
				if err == nil {
					break
				}
				if !errors.Is(err, syscall.EWOULDBLOCK) || time.Now().After(deadline) {
					t.Fatalf("descendant survived completed parent: %v", err)
				}
				time.Sleep(10 * time.Millisecond)
			}
		})
	}
}
