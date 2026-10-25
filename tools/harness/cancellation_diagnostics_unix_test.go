//go:build darwin || linux

package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"testing"
	"time"
)

// Decode the persisted shape so the regression also covers report publication.
func reportedCancellations(t *testing.T, result commandResult) []struct {
	Phase      string `json:"phase"`
	PID        int    `json:"pid"`
	PGID       int    `json:"pgid"`
	Signal     string `json:"signal"`
	DurationNS int64  `json:"duration_ns"`
	Outcome    string `json:"outcome"`
	Error      string `json:"error"`
} {
	t.Helper()
	data, err := json.Marshal(result)
	if err != nil {
		t.Fatal(err)
	}
	var report struct {
		Attempts []struct {
			Phase      string `json:"phase"`
			PID        int    `json:"pid"`
			PGID       int    `json:"pgid"`
			Signal     string `json:"signal"`
			DurationNS int64  `json:"duration_ns"`
			Outcome    string `json:"outcome"`
			Error      string `json:"error"`
		} `json:"cancellation_attempts"`
	}
	if err := json.Unmarshal(data, &report); err != nil {
		t.Fatal(err)
	}
	return report.Attempts
}

func TestCancellationSignalDiagnostics(t *testing.T) {
	for _, tc := range []struct {
		name, outcome string
		raw, returned error
	}{
		{"signaled", "signaled", nil, nil},
		{"disappeared", "absent", syscall.ESRCH, os.ErrProcessDone},
		{"permission denied", "failed", syscall.EPERM, syscall.EPERM},
	} {
		t.Run(tc.name, func(t *testing.T) {
			cmd := &exec.Cmd{Process: &os.Process{Pid: 12345}}
			var targets []int
			cleanup := configureCancellationWithSignal(cmd, func(pid int, signal syscall.Signal) error {
				targets = append(targets, pid)
				if signal != syscall.SIGKILL {
					t.Errorf("unexpected signal: %v", signal)
				}
				return tc.raw
			})
			for _, phase := range []string{"deadline", "post_wait"} {
				if err := cleanup.cancel(phase); !errors.Is(err, tc.returned) {
					t.Errorf("%s normalized %v to %v; want %v", phase, tc.raw, err, tc.returned)
				}
			}
			attempts := cleanup.snapshot()
			if len(attempts) != 2 || len(targets) != 2 {
				t.Fatalf("cleanup attempts were suppressed: %+v, targets %v", attempts, targets)
			}
			for index, attempt := range attempts {
				wantError := ""
				if tc.raw != nil {
					wantError = tc.raw.Error()
				}
				if targets[index] != -12345 || attempt.PID != 12345 || attempt.PGID != 12345 || attempt.Signal != "SIGKILL" || attempt.Outcome != tc.outcome || attempt.Error != wantError || attempt.DurationNS < 0 {
					t.Errorf("raw signal result or owned target lost: %+v, target %d", attempt, targets[index])
				}
			}
			if attempts[0].Phase != "deadline" || attempts[1].Phase != "post_wait" {
				t.Errorf("cancellation phases conflated: %+v", attempts)
			}
			attempts[0].Phase = "mutated snapshot"
			if cleanup.snapshot()[0].Phase != "deadline" {
				t.Fatal("report snapshot aliases live diagnostic storage")
			}
		})
	}
}

func TestCancellationBeforeStartDoesNotSignal(t *testing.T) {
	cleanup := configureCancellationWithSignal(&exec.Cmd{}, func(int, syscall.Signal) error {
		t.Fatal("unstarted command attempted signaling")
		return nil
	})
	if err := cleanup.cancel("context_cancellation"); !errors.Is(err, os.ErrProcessDone) {
		t.Fatalf("unstarted cancellation result: %v", err)
	}
	attempts := cleanup.snapshot()
	if len(attempts) != 1 || attempts[0].PID != 0 || attempts[0].PGID != 0 || attempts[0].Outcome != "absent" {
		t.Fatalf("unstarted cancellation evidence: %+v", attempts)
	}
}

func TestCancellationDiagnosticsConcurrentAttempts(t *testing.T) {
	entered := make(chan struct{}, 2)
	release := make(chan struct{})
	cleanup := configureCancellationWithSignal(&exec.Cmd{Process: &os.Process{Pid: 12345}}, func(int, syscall.Signal) error {
		entered <- struct{}{}
		<-release
		return syscall.EPERM
	})
	var waiting sync.WaitGroup
	for _, phase := range []string{"context_cancellation", "post_wait"} {
		phase := phase
		waiting.Add(1)
		go func() {
			defer waiting.Done()
			if err := cleanup.cancel(phase); !errors.Is(err, syscall.EPERM) {
				t.Errorf("concurrent %s lost permission error: %v", phase, err)
			}
		}()
	}
	<-entered
	<-entered
	if attempts := cleanup.snapshot(); len(attempts) != 0 {
		t.Errorf("unfinished signal reported complete: %+v", attempts)
	}
	close(release)
	waiting.Wait()
	attempts := cleanup.snapshot()
	counts := map[string]int{}
	for _, attempt := range attempts {
		counts[attempt.Phase]++
		if attempt.Outcome != "failed" || attempt.Error != syscall.EPERM.Error() {
			t.Errorf("concurrent result lost: %+v", attempt)
		}
	}
	if len(attempts) != 2 || counts["context_cancellation"] != 1 || counts["post_wait"] != 1 {
		t.Fatalf("concurrent attempts lost: %+v", attempts)
	}
}

func TestRunCommandRetainsCleanupPermissionFailure(t *testing.T) {
	for _, tc := range []struct {
		name, ending, primary string
		code                  int
		timedOut              bool
	}{
		{"failed parent", "exit 7", "exit status 7", 7, false},
		{"deadline", "exec /bin/sleep 30", "signal: killed", -1, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			dir := t.TempDir()
			result := runCommandWithCancellation(context.Background(), dir, commandSpec{
				Name: "permission", Executable: "/bin/sh",
				Args: []string{"-c", `printf 'raw stdout\000\n'; printf 'raw stderr\r\n' >&2; ` + tc.ending},
				Dir:  dir, Env: os.Environ(), Timeout: 500 * time.Millisecond,
			}, func(cmd *exec.Cmd) *commandCancellation {
				// Keep the injected denied group denied during read-only probes;
				// a disappearing group is covered separately with real processes.
				cmd.WaitDelay = 20 * time.Millisecond
				cleanup := configureCancellationWithSignal(cmd, func(target int, signal syscall.Signal) error {
					if signal == 0 {
						return syscall.EPERM
					}
					return syscall.Kill(target, signal)
				})
				cancel := cleanup.cancel
				cleanup.cancel = func(phase string) error {
					if phase != "post_wait" {
						return cancel(phase)
					}
					// The child has already been reaped, and its deadline signal
					// (when needed) was real. Inject only the final signal result.
					return cleanup.signal(phase, cmd.Process.Pid, cmd.Process.Pid, "SIGKILL", func() error { return syscall.EPERM }, nil)
				}
				return cleanup
			})
			if result.succeeded() || result.ExitCode != tc.code || result.TimedOut != tc.timedOut || !strings.HasPrefix(result.Error, tc.primary+"\n") || !strings.Contains(result.Error, "terminate remaining processes: "+syscall.EPERM.Error()) {
				t.Fatalf("cleanup permission failure changed the primary outcome: %+v", result)
			}
			for path, want := range map[string]string{result.StdoutPath: "raw stdout\x00\n", result.StderrPath: "raw stderr\r\n"} {
				data, err := os.ReadFile(path)
				if err != nil || string(data) != want {
					t.Errorf("cleanup failure changed stream %q: %q, %v; want %q", path, data, err, want)
				}
			}
			attempts := reportedCancellations(t, result)
			wantCount := 1
			if tc.timedOut {
				wantCount = 2
			}
			if len(attempts) != wantCount || attempts[len(attempts)-1].Phase != "post_wait" || attempts[len(attempts)-1].Error != syscall.EPERM.Error() || attempts[len(attempts)-1].Outcome != "failed" {
				t.Fatalf("cleanup failure diagnostics lost: %+v", attempts)
			}
			if tc.timedOut && (attempts[0].Phase != "deadline" || attempts[0].Outcome != "signaled" || attempts[0].Error != "") {
				t.Errorf("cleanup failure obscured successful deadline signaling: %+v", attempts)
			}
		})
	}
}

func TestRunCommandReportsDeadlineAndDescendantCleanup(t *testing.T) {
	dir := t.TempDir()
	lockPath := filepath.Join(dir, "descendant.lock")
	file, err := os.Create(lockPath)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := file.Close(); err != nil {
			t.Errorf("close lifetime lock: %v", err)
		}
	})
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
	exe, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	result := runCommand(context.Background(), dir, commandSpec{
		Name: "deadline-descendant", Executable: "/bin/sh",
		Args: []string{"-c", `"$1" -test.run=^TestWaitDelayDescendant$ & while ! test -s "$2"; do sleep 0.01; done; wait`, "deadline-parent", exe, lockPath},
		Dir:  dir, Env: append(os.Environ(), "L_HARNESS_DESCENDANT_LOCK="+lockPath),
		Timeout: time.Second, TestJSON: true,
	})
	data, err := os.ReadFile(lockPath)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := strconv.Atoi(string(data)); err != nil {
		t.Fatalf("descendant did not reach its owned lifetime: %q; %+v", data, result)
	}
	if !result.TimedOut || result.succeeded() || result.ExitCode != -1 || result.Error == "" {
		t.Fatalf("deadline result lost: %+v", result)
	}
	deadline := time.Now().Add(time.Second)
	for {
		err = syscall.Flock(int(file.Fd()), syscall.LOCK_EX|syscall.LOCK_NB)
		if err == nil {
			break
		}
		if !errors.Is(err, syscall.EWOULDBLOCK) || time.Now().After(deadline) {
			t.Fatalf("descendant lifetime survived deadline cleanup: %v", err)
		}
		time.Sleep(time.Millisecond)
	}
	attempts := reportedCancellations(t, result)
	if len(attempts) != 2 || attempts[0].Phase != "deadline" || attempts[1].Phase != "post_wait" {
		t.Fatalf("initial and post-wait cancellation evidence missing: %+v", attempts)
	}
	for _, attempt := range attempts {
		if attempt.PID <= 1 || attempt.PGID != attempt.PID || attempt.PID != attempts[0].PID || attempt.Signal != "SIGKILL" || attempt.DurationNS < 0 {
			t.Errorf("owned signal target or timing lost: %+v", attempt)
		}
	}
	if attempts[0].Outcome != "signaled" || attempts[0].Error != "" {
		t.Errorf("deadline signaling did not succeed: %+v", attempts[0])
	}
}

func TestCancellationCompletionFixture(t *testing.T) {
	dir := os.Getenv("L_HARNESS_COMPLETION_RACE")
	if dir == "" {
		return
	}
	lock, err := os.OpenFile(filepath.Join(dir, "lifetime"), os.O_RDWR, 0600)
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := lock.Close(); err != nil {
			t.Errorf("close lifetime lock: %v", err)
		}
	}()
	if err := syscall.Flock(int(lock.Fd()), syscall.LOCK_EX); err != nil {
		t.Fatal(err)
	}
	if _, err := fmt.Fprint(os.Stdout, "completed stdout\x00\n"); err != nil {
		t.Fatal(err)
	}
	if _, err := fmt.Fprint(os.Stderr, "completed stderr\r\n"); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "ready"), []byte(strconv.Itoa(os.Getpid())), 0600); err != nil {
		t.Fatal(err)
	}
	for {
		if _, err := os.Stat(filepath.Join(dir, "release")); err == nil {
			os.Exit(0)
		} else if !errors.Is(err, os.ErrNotExist) {
			t.Fatal(err)
		}
		time.Sleep(time.Millisecond)
	}
}

func TestRunCommandCancellationRacesNormalCompletion(t *testing.T) {
	dir := t.TempDir()
	lock, err := os.Create(filepath.Join(dir, "lifetime"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := lock.Close(); err != nil {
			t.Errorf("close lifetime lock: %v", err)
		}
	})
	exe, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	finished := make(chan commandResult, 1)
	go func() {
		finished <- runCommand(ctx, dir, commandSpec{
			Name: "completion-race", Executable: exe, Args: []string{"-test.run=^TestCancellationCompletionFixture$"},
			Dir: dir, Env: append(os.Environ(), "L_HARNESS_COMPLETION_RACE="+dir), Timeout: 5 * time.Second,
		})
	}()
	var pid int
	readyDeadline := time.Now().Add(4 * time.Second)
	for {
		data, err := os.ReadFile(filepath.Join(dir, "ready"))
		if err == nil {
			pid, err = strconv.Atoi(string(data))
			if err == nil {
				break
			}
		}
		if time.Now().After(readyDeadline) {
			cancel()
			t.Fatalf("completion fixture did not establish its barrier: %+v", <-finished)
		}
		time.Sleep(time.Millisecond)
	}
	// Both actions begin at the same barrier after all asserted bytes were
	// written. Either normal completion or cancellation may legitimately win.
	race := make(chan struct{})
	var racers sync.WaitGroup
	racers.Add(2)
	go func() {
		defer racers.Done()
		<-race
		cancel()
	}()
	go func() {
		defer racers.Done()
		<-race
		if err := os.WriteFile(filepath.Join(dir, "release"), nil, 0600); err != nil {
			t.Errorf("release normal completion: %v", err)
		}
	}()
	close(race)
	racers.Wait()
	result := <-finished
	if result.TimedOut || result.ExitCode != 0 && result.ExitCode != -1 {
		t.Fatalf("completion/cancellation lost the real command outcome: %+v", result)
	}
	if result.ExitCode == -1 && !strings.Contains(result.Error, "signal: killed") {
		t.Errorf("canceled process cause lost: %+v", result)
	}
	if result.ExitCode == 0 && result.Error != "" && result.Error != context.Canceled.Error() {
		t.Errorf("unexpected normal-exit cancellation cause: %+v", result)
	}
	for path, want := range map[string]string{result.StdoutPath: "completed stdout\x00\n", result.StderrPath: "completed stderr\r\n"} {
		data, err := os.ReadFile(path)
		if err != nil || string(data) != want {
			t.Errorf("completed stream prefix lost: %q, %v; want %q", data, err, want)
		}
	}
	attempts := reportedCancellations(t, result)
	if len(attempts) < 1 || len(attempts) > 2 || attempts[len(attempts)-1].Phase != "post_wait" {
		t.Fatalf("post-wait cleanup lost in completion race: %+v", attempts)
	}
	for _, attempt := range attempts {
		if attempt.PID != pid || attempt.PGID != pid || attempt.Outcome == "failed" || attempt.Phase != "context_cancellation" && attempt.Phase != "post_wait" {
			t.Errorf("unexpected owned cleanup result: %+v", attempt)
		}
	}
	if err := syscall.Flock(int(lock.Fd()), syscall.LOCK_EX|syscall.LOCK_NB); err != nil {
		t.Errorf("fixture lifetime survived completion/cancellation: %v", err)
	}
	if err := syscall.Kill(-pid, 0); !errors.Is(err, syscall.ESRCH) {
		t.Errorf("owned group survived completion/cancellation: %v", err)
	}
}
