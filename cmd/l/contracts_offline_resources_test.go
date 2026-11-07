package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	distributiontest "github.com/l-programming-language/l/internal/testsupport/distribution"
)

// The caller has already denied process creation and network access at the
// kernel boundary. Resource failures must remain diagnostics, never repair work.
func assertOfflineRuntimeResources(t *testing.T) {
	t.Helper()
	assets := distributiontest.CompleteOverride(t)
	input := writeContractFile(t, filepath.Join(t.TempDir(), "main.l"), []byte("Show `offline resources`.\n"))
	report, _ := contractJSON(t, input, 0, "inspect", "--assets", assets)
	if len(report.Bindings) == 0 || len(report.Executions) != 0 {
		t.Fatal("selected-provider inspection omitted bindings or invented executions")
	}
	binding := string(report.Bindings[0].Occurrence)
	commands := [][]string{
		{"inspect"},
		{"check", "--profile", "NoBlocking"},
		{"explain", "--binding", binding},
	}
	for _, command := range commands {
		status := 0
		if command[0] == "check" {
			status = 1
		}
		contractJSON(t, input, status, append(command, "--assets", assets)...)
	}
	for _, mutation := range []struct {
		name    string
		missing bool
	}{
		{"runtime/generated/support.json", true},
		{"runtime/rust/host.rs", true},
		{"runtime/generated/support.aml", false},
		{"builtins/eng/storage/dictionary.l", false},
	} {
		path := filepath.Join(assets, filepath.FromSlash(mutation.name))
		original := readContractFile(t, path)
		if mutation.missing {
			if err := os.Remove(path); err != nil {
				t.Fatal(err)
			}
		} else {
			writeContractFile(t, path, append(append([]byte(nil), original...), '\n'))
		}
		for _, command := range commands {
			_, diagnostic := contractOutput(t, 1, append(append([]string(nil), command...), "--assets", assets, input)...)
			if !strings.Contains(diagnostic, mutation.name) {
				t.Fatalf("offline %s failure lost selected input %s: %s", command[0], mutation.name, diagnostic)
			}
		}
		writeContractFile(t, path, original)
		contractJSON(t, input, 0, "inspect", "--assets", assets)
	}
}
