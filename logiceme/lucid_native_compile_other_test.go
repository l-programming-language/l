//go:build nativecompile && !darwin && !linux

package Logiceme

import (
	"os/exec"
	"time"
)

func lucidBoundCompiler(cmd *exec.Cmd) { cmd.WaitDelay = time.Second }
