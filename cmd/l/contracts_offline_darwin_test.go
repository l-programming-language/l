package main

import (
	"context"
	"os/exec"
	"path/filepath"
	"strconv"
)

func contractsOfflineCommand(ctx context.Context, executable string) (*exec.Cmd, error) {
	resolved, err := filepath.EvalSymlinks(executable)
	if err != nil {
		return nil, err
	}
	// sandbox-exec needs this initial exec; process-fork remains denied, so no
	// subsequently requested process can use the executable allow rule.
	profile := "(version 1)(allow default)(deny process-fork)(deny network*)(deny process-exec (require-not (literal " + strconv.Quote(resolved) + ")))"
	return exec.CommandContext(ctx, "/usr/bin/sandbox-exec", "-p", profile, resolved, "-test.run=^TestContractsOfflineCapabilitiesHelper$", "-test.v"), nil
}

func contractsDenyProcessAndNetwork() error { return nil }
