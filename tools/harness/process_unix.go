//go:build darwin || linux

package main

import (
	"errors"
	"os"
	"os/exec"
	"syscall"
)

// Cancel the whole group, including go test's child test binary.
func configureCancellation(cmd *exec.Cmd) *commandCancellation {
	return configureCancellationWithSignal(cmd, syscall.Kill)
}

// The dependency is local to this command; tests can expose signal failures
// without changing a global hook or signaling a process they do not own.
func configureCancellationWithSignal(cmd *exec.Cmd, send func(int, syscall.Signal) error) *commandCancellation {
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	cleanup := &commandCancellation{}
	configureDisappearanceObservation(cleanup, cmd, send)
	cleanup.cancel = func(phase string) error {
		pid := 0
		if cmd.Process != nil {
			pid = cmd.Process.Pid
		}
		// Setpgid creates this command's group with its PID as the PGID. Capture
		// that owned target before signaling; querying the reaped leader cannot
		// establish whether descendants remain in its group.
		return cleanup.signal(phase, pid, pid, "SIGKILL", func() error {
			if cmd.Process == nil {
				return os.ErrProcessDone
			}
			return send(-pid, syscall.SIGKILL)
		}, func(err error) bool { return errors.Is(err, syscall.ESRCH) })
	}
	return cleanup
}
