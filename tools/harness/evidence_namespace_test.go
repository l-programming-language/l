package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// Each real harness command needs a new evidence identity even when a caller
// supplies the same artifact root and an inherited identity from another run.
func TestHarnessEvidenceNamespace(t *testing.T) {
	root := integrationFixture(t)
	t.Setenv("L_HARNESS_ROOT", root)
	t.Setenv("L_HARNESS_ORIGINAL_GOFLAGS", "")
	t.Setenv("L_HARNESS_RUN_ID", "inherited-run")
	evidenceRoot := filepath.Join(t.TempDir(), "evidence root")
	t.Setenv("L_HARNESS_EVIDENCE_DIR", evidenceRoot)
	source := `package fixture
import ("fmt"; "os"; "testing")
func TestEvidenceEnvironment(t *testing.T) {
 fmt.Printf("EVIDENCE_NAMESPACE=%s\nEVIDENCE_ROOT=%s\n", os.Getenv("L_HARNESS_RUN_ID"), os.Getenv("L_HARNESS_EVIDENCE_DIR"))
}
`
	if err := os.WriteFile(filepath.Join(root, "evidence_test.go"), []byte(source), 0600); err != nil {
		t.Fatal(err)
	}
	previous := ""
	for i := 0; i < 2; i++ {
		code, report := invokeIntegrationCLI(t, "verify")
		if code != 0 || report.AcceptancePassed == nil || !*report.AcceptancePassed {
			t.Fatalf("fixture gate failed: code=%d report=%+v", code, report)
		}
		id := report.Environment["evidence_run_id"]
		if id == "" || id == "inherited-run" || id == previous || id != filepath.Base(report.RunDirectory) {
			t.Fatalf("missing or reused evidence namespace: previous=%q current=%q run=%q", previous, id, report.RunDirectory)
		}
		previous = id
		for _, command := range report.Commands {
			if command.Name != "tests" {
				continue
			}
			data, err := os.ReadFile(command.TestOutputPath)
			if err != nil {
				t.Fatal(err)
			}
			for _, want := range []string{"EVIDENCE_NAMESPACE=" + id + "\n", "EVIDENCE_ROOT=" + evidenceRoot + "\n"} {
				if !strings.Contains(string(data), want) {
					t.Errorf("child command did not retain evidence context %q: %s", want, data)
				}
			}
		}
	}
}
