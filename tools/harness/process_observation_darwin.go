//go:build darwin

package main

import (
	"errors"
	"os/exec"
	"syscall"
	"time"
)

// Darwin can return EPERM for a group containing only unreaped zombies. After
// our successful deadline SIGKILL, observe disappearance without signaling it
// again. Reuse only the remaining WaitDelay budget; persistent errors stay errors.
func configureDisappearanceObservation(cleanup *commandCancellation, cmd *exec.Cmd, send func(int, syscall.Signal) error) {
	cleanup.observe = func(phase string, pid, pgid int, err error) (*cancellationObservation, bool) {
		if phase != "post_wait" || !errors.Is(err, syscall.EPERM) || cmd.WaitDelay <= 0 {
			return nil, false
		}
		cleanup.mu.Lock()
		prior := cleanup.deadline
		cleanup.mu.Unlock()
		if prior == nil || prior.pid != pid || prior.pgid != pgid || pgid <= 1 {
			return nil, false
		}
		started := time.Now()
		until := prior.started.Add(cmd.WaitDelay)
		budget := time.Until(until)
		if budget < 0 {
			budget = 0
		}
		observation := &cancellationObservation{BudgetNS: budget.Nanoseconds(), Outcome: "unresolved"}
		for {
			err = send(-pgid, 0)
			observation.Probes++
			observation.LastError = ""
			if err != nil {
				observation.LastError = err.Error()
			}
			if errors.Is(err, syscall.ESRCH) {
				observation.DurationNS = time.Since(started).Nanoseconds()
				observation.Outcome = "absent"
				return observation, true
			}
			remaining := time.Until(until)
			if remaining <= 0 || err != nil && !errors.Is(err, syscall.EPERM) {
				observation.DurationNS = time.Since(started).Nanoseconds()
				return observation, false
			}
			if remaining > 5*time.Millisecond {
				remaining = 5 * time.Millisecond
			}
			time.Sleep(remaining)
		}
	}
}
