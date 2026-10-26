//go:build linux

package main

import (
	"os/exec"
	"syscall"
)

func configureDisappearanceObservation(*commandCancellation, *exec.Cmd, func(int, syscall.Signal) error) {
}
