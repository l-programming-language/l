//go:build darwin || linux

package main

import (
	"os"
	"path/filepath"
	"testing"

	distributiontest "github.com/l-programming-language/l/internal/testsupport/distribution"
	nativetest "github.com/l-programming-language/l/internal/testsupport/native"
	"github.com/l-programming-language/l/internal/testsupport/phase1"
)

func TestAlternateRuntimeStopWithoutJoining(t *testing.T) {
	assets := t.TempDir()
	if err := os.CopyFS(assets, distributiontest.AlternateRuntime(t)); err != nil {
		t.Fatal(err)
	}
	composeScopedTestProvider(t, assets)
	runnableStopWithoutJoining(t, func(t *testing.T, project string, _ []string) runnableBuildResult {
		t.Helper()
		recorder := nativetest.CLIProbe(t)
		stdout, stderr, err := runBuildCommand(t, phase1.CLI(t), project, recorder.Env(nil), "build", "--assets", assets, "main.l")
		if err != nil {
			t.Fatalf("alternate Runtime build: %v\nstdout=%s\nstderr=%s", err, stdout, stderr)
		}
		binary := filepath.Join(project, "main")
		if string(stdout) != binary+"\n" || len(stderr) != 0 {
			t.Fatalf("CLI streams = %q / %q; want only published path", stdout, stderr)
		}
		recorder.AssertPublished(t, binary)
		return runnableBuildResult{project: project, binary: binary, recorder: recorder}
	})
}
