//go:build !darwin && !linux

package main

import "os/exec"

func configureCommand(_ *exec.Cmd) {}
