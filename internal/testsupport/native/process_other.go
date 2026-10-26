//go:build !darwin && !linux

package native

import (
	"os/exec"
	"time"
)

func superviseCompiler(cmd *exec.Cmd) { cmd.WaitDelay = 2 * time.Second }
