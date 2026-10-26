package Logiceme

import (
	"bytes"
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"
)

func TestDistributionBackend(t *testing.T) {
	if id := os.Getenv("L_DISTRIBUTION_BACKEND_CHILD"); id != "" {
		data, readErr := os.ReadFile(filepath.Join(os.Getenv("L_DISTRIBUTION_CANONICAL_ROOT"), "logiceme/testdata/lucid/grouping.aml.txt"))
		if readErr != nil {
			t.Fatal(readErr)
		}
		switch id {
		case "FS13":
			bundle, err := NewLogicemeService(nil).Evaluate(data)
			if err != nil {
				t.Fatalf("embedded backend prerequisite failed; runtime not exercised: %v", err)
			}
			if len(bundle.ByteArrays) != 1 || len(bundle.ByteArrays[0]) == 0 {
				t.Fatal("empty backend source bundle")
			}
		case "FS15":
			if err := os.MkdirAll("programs", 0700); err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile("programs/external.aml", data, 0600); err != nil {
				t.Fatal(err)
			}
			scanner, err := LoadScanner("external.aml")
			if err != nil || !bytes.Equal(scanner.ProgramBytes, data) {
				t.Fatalf("explicit external AML read=%q error=%v", scanner.ProgramBytes, err)
			}
		default:
			t.Fatalf("unexpected backend case %s", id)
		}
		return
	}
	executable, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	root, err := filepath.Abs("..")
	if err != nil {
		t.Fatal(err)
	}
	for _, id := range []string{"FS13", "FS15"} {
		t.Run(id, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()
			cmd := exec.CommandContext(ctx, executable, "-test.run=^TestDistributionBackend$", "-test.v")
			cmd.Dir = t.TempDir()
			cmd.Env = append(os.Environ(), "L_DISTRIBUTION_BACKEND_CHILD="+id, "L_DISTRIBUTION_CANONICAL_ROOT="+root)
			cmd.WaitDelay = 2 * time.Second
			var out, diagnostic bytes.Buffer
			cmd.Stdout, cmd.Stderr = &out, &diagnostic
			if err := cmd.Run(); err != nil || diagnostic.Len() != 0 {
				t.Fatalf("backend %s: %v\nstdout=%s\nstderr=%s", id, err, out.Bytes(), diagnostic.Bytes())
			}
		})
	}
}
