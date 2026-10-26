//go:build !darwin && !linux

package main

import "os/exec"

// The CLI rejects executable builds on unsupported hosts before this boundary.
func configureBuildCommand(command *exec.Cmd) {}
