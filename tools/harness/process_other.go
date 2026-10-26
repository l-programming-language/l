//go:build !darwin && !linux

package main

import (
	"os"
	"os/exec"
)

// CommandContext supplies direct-process cancellation on other platforms.
// Go's -timeout also bounds the independently launched test executable.
func configureCancellation(cmd *exec.Cmd) *commandCancellation {
	cancel := cmd.Cancel
	cleanup := &commandCancellation{}
	cleanup.cancel = func(phase string) error {
		pid := 0
		if cmd.Process != nil {
			pid = cmd.Process.Pid
		}
		return cleanup.signal(phase, pid, 0, "Kill", func() error {
			if cancel == nil || cmd.Process == nil {
				return os.ErrProcessDone
			}
			return cancel()
		}, nil)
	}
	return cleanup
}
