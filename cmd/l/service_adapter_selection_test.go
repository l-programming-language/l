//go:build darwin || linux

package main

import (
	"os"
	"testing"
)

func TestStandaloneServiceDoesNotAcquireBuildArtifacts(t *testing.T) {
	project, filename := runnableProject(t, "basic.l")
	const source = "Use ProbeIO's Print from `probe.l` as \"Print\".\n" +
		"Create \"Unused\" as a New Main Service.\n" +
		"Create \"Worker\" as a New Basic App.\nThe App Prints `chosen`.\n" +
		"Run Worker.\nAwait Worker.\nShow Worker's Output.\n"
	if err := os.WriteFile(filename, []byte(source), 0600); err != nil {
		t.Fatal(err)
	}
	// The real build helper checks the entire Cargo role inventory and every
	// published artifact identity. Unused must add no crate or companion.
	built := runnableBuildProject(t, project, []string{"Worker"})
	status, stdout, stderr := runnableRun(t, built.binary, t.TempDir(), nil)
	if status != 0 || string(stdout) != "chosen" || len(stderr) != 0 {
		t.Fatalf("structural Service changed App execution: %d/%q/%q", status, stdout, stderr)
	}
}
