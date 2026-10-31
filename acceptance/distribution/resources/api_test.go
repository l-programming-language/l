package resources_test

import (
	"bytes"
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"

	distributiontest "github.com/l-programming-language/l/internal/testsupport/distribution"
)

// Only the inert probe references new production APIs. A missing API fails a
// named test with real compiler diagnostics, not compilation of this package.
func TestDistributionAPI(t *testing.T) {
	root := distributiontest.RepoRoot(t)
	directory := t.TempDir()
	for source, target := range map[string]string{"testdata/distribution/api/probe_test.go.txt": "probe_test.go", "go.sum": "go.sum"} {
		data, err := os.ReadFile(filepath.Join(root, source))
		if err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(directory, target), data, 0600); err != nil {
			t.Fatal(err)
		}
	}
	module := "module github.com/l-programming-language/l/distributionchecks\n\ngo 1.20\n\nrequire (\n github.com/l-programming-language/l v0.0.0\n github.com/gofrs/uuid v4.4.0+incompatible // indirect\n)\nreplace github.com/l-programming-language/l => " + strconv.Quote(root) + "\n"
	if err := os.WriteFile(filepath.Join(directory, "go.mod"), []byte(module), 0600); err != nil {
		t.Fatal(err)
	}
	binary := filepath.Join(directory, "probe")
	ctx, cancel := context.WithCancel(context.Background())
	cmd := exec.CommandContext(ctx, "go", "test", "-mod=readonly", "-race", "-c", "-o", binary, ".")
	cmd.Dir = directory
	cmd.Env = append(os.Environ(), "GOPROXY=off", "GOWORK=off", "GOTOOLCHAIN=local")
	cmd.WaitDelay = 2 * time.Second
	buildOutput, buildErr := cmd.CombinedOutput()
	cancel()
	for _, id := range []string{"FS05", "FS06", "FS07", "FS08", "FS09", "FS12", "FS13", "FS14", "FS15", "MG10", "ReleaseInventory"} {
		t.Run(id, func(t *testing.T) {
			if buildErr != nil {
				t.Fatalf("distribution-api prerequisite failed: %v\n%s\nintended assertion not exercised; runtime not exercised", buildErr, buildOutput)
			}
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			cmd := exec.CommandContext(ctx, binary, "-test.run=^TestDistributionAPI$/^"+id+"$", "-test.v", "-test.timeout=0")
			cmd.Dir = t.TempDir()
			cmd.Env = append(os.Environ(), "L_DISTRIBUTION_ROOT="+root, "GORACE=atexit_sleep_ms=0")
			cmd.WaitDelay = 2 * time.Second
			var stdout, stderr bytes.Buffer
			cmd.Stdout, cmd.Stderr = &stdout, &stderr
			err := cmd.Run()
			if err != nil || stderr.Len() != 0 || !strings.Contains(stdout.String(), "--- PASS: TestDistributionAPI/"+id) {
				t.Fatalf("real distribution API assertion failed: %v\nstdout=%s\nstderr=%s", err, stdout.Bytes(), stderr.Bytes())
			}
			t.Logf("actual race-instrumented API probe: %s", stdout.Bytes())
		})
	}
}
