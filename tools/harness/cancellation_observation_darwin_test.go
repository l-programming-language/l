//go:build darwin

package main

import (
	"context"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"syscall"
	"testing"
	"time"
)

func TestRunCommandObservesReapedZombieGroup(t *testing.T) {
	dir := t.TempDir()
	lockPath := filepath.Join(dir, "sibling.lock")
	lock, err := os.Create(lockPath)
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
	var sibling *exec.Cmd
	siblingReaped := false
	t.Cleanup(func() {
		if sibling != nil && sibling.Process != nil && !siblingReaped {
			_ = sibling.Process.Kill() //nolint:errcheck // Best-effort fallback when the main zombie-observation assertion has already failed.
			_ = sibling.Wait()         //nolint:errcheck // Reap the fallback-killed child; its failing exit status is expected.
		}
	})
	kills, observations := 0, 0
	result := runCommandWithCancellation(context.Background(), dir, commandSpec{
		Name: "zombie-group", Executable: "/bin/sh",
		Args: []string{"-c", `printf 'ready\n'; exec /bin/sleep 30`}, Dir: dir,
		Env: os.Environ(), Timeout: 500 * time.Millisecond,
	}, func(cmd *exec.Cmd) *commandCancellation {
		return configureCancellationWithSignal(cmd, func(target int, signal syscall.Signal) error {
			if signal == syscall.SIGKILL {
				kills++
				if kills == 1 {
					// Own a second child in precisely the command's group, but
					// retain its Wait ourselves so its zombie remains observable.
					sibling = exec.Command(exe, "-test.run=^TestWaitDelayDescendant$")
					sibling.Env = append(os.Environ(), "L_HARNESS_DESCENDANT_LOCK="+lockPath)
					sibling.SysProcAttr = &syscall.SysProcAttr{Setpgid: true, Pgid: -target}
					if err := sibling.Start(); err != nil {
						t.Errorf("start owned sibling: %v", err)
						return syscall.Kill(target, signal)
					}
					until := time.Now().Add(time.Second)
					for {
						data, readErr := os.ReadFile(lockPath)
						if readErr != nil {
							t.Errorf("read sibling lifetime lock: %v", readErr)
							break
						}
						if len(data) != 0 {
							break
						}
						if time.Now().After(until) {
							t.Error("owned sibling did not acquire its lifetime lock")
							break
						}
						time.Sleep(time.Millisecond)
					}
				} else {
					// The leader was reaped by cmd.Wait. Wait read-only until the
					// killed, unreaped sibling is the group's remaining zombie.
					until := time.Now().Add(time.Second)
					for !errors.Is(syscall.Kill(target, 0), syscall.EPERM) {
						if time.Now().After(until) {
							t.Error("owned zombie group did not expose Darwin EPERM")
							break
						}
						time.Sleep(time.Millisecond)
					}
				}
				return syscall.Kill(target, signal)
			}
			if signal != 0 {
				t.Errorf("unexpected signal: %v", signal)
			}
			observations++
			err := syscall.Kill(target, signal)
			if observations == 1 {
				if !errors.Is(err, syscall.EPERM) {
					t.Errorf("initial zombie observation: %v", err)
				}
				if err := sibling.Wait(); err == nil {
					t.Error("killed sibling unexpectedly exited successfully")
				}
				siblingReaped = true
			}
			return err
		})
	})
	if !result.TimedOut || result.ExitCode != -1 || result.Error != "signal: killed" || result.succeeded() {
		t.Fatalf("verified group disappearance did not retain only the original timeout failure: %+v", result)
	}
	if kills != 2 || observations < 2 || !siblingReaped {
		t.Fatalf("cleanup resignal or missing disappearance observation: kills=%d observations=%d reaped=%t", kills, observations, siblingReaped)
	}
	attempts := reportedCancellations(t, result)
	if len(attempts) != 2 || attempts[1].Error != syscall.EPERM.Error() || attempts[1].Outcome != "absent_after_observation" {
		t.Fatalf("raw transient EPERM was erased: %+v", attempts)
	}
	observation := result.Cancellations[1].Observation
	if observation == nil || observation.Outcome != "absent" || observation.Probes != observations || observation.LastError != syscall.ESRCH.Error() || observation.BudgetNS < 0 || observation.DurationNS < 0 {
		t.Errorf("verified disappearance diagnostics lost: %+v", observation)
	}
	if err := syscall.Flock(int(lock.Fd()), syscall.LOCK_EX|syscall.LOCK_NB); err != nil {
		t.Errorf("owned sibling lifetime survived cleanup: %v", err)
	}
	if err := syscall.Kill(-attempts[1].PGID, 0); !errors.Is(err, syscall.ESRCH) {
		t.Errorf("group still exists after both owned children were reaped: %v", err)
	}
	data, err := os.ReadFile(result.StdoutPath)
	if err != nil || string(data) != "ready\n" {
		t.Errorf("timeout changed completed output: %q, %v", data, err)
	}
}

func TestDisappearanceObservationRequiresSuccessfulOwnedDeadline(t *testing.T) {
	for _, tc := range []struct {
		name       string
		firstPhase string
		firstError error
		changePID  bool
		budget     time.Duration
		wantProbe  bool
	}{
		{"persistent permission failure", "deadline", nil, false, 15 * time.Millisecond, true},
		{"expired deadline budget", "deadline", nil, false, time.Nanosecond, true},
		{"failed deadline signal", "deadline", syscall.EPERM, false, time.Second, false},
		{"ordinary cancellation", "context_cancellation", nil, false, time.Second, false},
		{"different process", "deadline", nil, true, time.Second, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			cmd := &exec.Cmd{Process: &os.Process{Pid: 12345}, WaitDelay: tc.budget}
			kills, probes := 0, 0
			cleanup := configureCancellationWithSignal(cmd, func(target int, signal syscall.Signal) error {
				if signal == 0 {
					probes++
					return syscall.EPERM
				}
				kills++
				if kills == 1 {
					return tc.firstError
				}
				return syscall.EPERM
			})
			if err := cleanup.cancel(tc.firstPhase); !errors.Is(err, tc.firstError) {
				t.Fatalf("first signal result: %v", err)
			}
			if tc.changePID {
				cmd.Process = &os.Process{Pid: 12346}
			}
			if err := cleanup.cancel("post_wait"); !errors.Is(err, syscall.EPERM) {
				t.Fatalf("persistent permission failure normalized: %v", err)
			}
			attempts := cleanup.snapshot()
			if kills != 2 || len(attempts) != 2 || attempts[1].Error != syscall.EPERM.Error() || attempts[1].Outcome != "failed" {
				t.Fatalf("signal retries or erased failure: kills=%d attempts=%+v", kills, attempts)
			}
			if tc.wantProbe != (probes != 0) {
				t.Errorf("observation eligibility changed: probes=%d want=%t", probes, tc.wantProbe)
			}
			if tc.wantProbe {
				observation := attempts[1].Observation
				if observation == nil || observation.Outcome != "unresolved" || observation.Probes != probes || observation.LastError != syscall.EPERM.Error() || observation.BudgetNS > tc.budget.Nanoseconds() {
					t.Errorf("persistent observation diagnostics lost: %+v", observation)
				}
				observation.Outcome = "mutated"
				if cleanup.snapshot()[1].Observation.Outcome != "unresolved" {
					t.Error("observation snapshot aliases recorded state")
				}
				if tc.budget == time.Nanosecond && probes != 1 {
					t.Errorf("expired WaitDelay gained another observation window: %d probes", probes)
				}
			} else if attempts[1].Observation != nil {
				t.Errorf("ineligible attempt was observed: %+v", attempts[1])
			}
		})
	}
}
