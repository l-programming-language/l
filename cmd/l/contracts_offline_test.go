package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"syscall"
	"testing"
	"time"

	"github.com/l-programming-language/l/contracts"
)

const contractsOfflineHelper = "L_CONTRACTS_OFFLINE_CAPABILITY_HELPER"

// The child runs real source-only commands with kernel-enforced process and
// socket denial. No PATH shim can stand in for this capability boundary.
func TestContractsOfflineNoProcessOrNetwork(t *testing.T) {
	executable, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	defer cancel()
	command, err := contractsOfflineCommand(ctx, executable)
	if err != nil {
		t.Fatal(err)
	}
	command.Env = append(os.Environ(), contractsOfflineHelper+"=1")
	output, err := command.CombinedOutput()
	if err != nil {
		t.Fatalf("offline capability subprocess failed: %v\n%s", err, output)
	}
	if !bytes.Contains(output, []byte("offline capability checks completed")) {
		t.Fatalf("helper did not complete capability assertions: %s", output)
	}
}

func TestContractsOfflineCapabilitiesHelper(t *testing.T) {
	if os.Getenv(contractsOfflineHelper) != "1" {
		return
	}
	if err := contractsDenyProcessAndNetwork(); err != nil {
		t.Fatal(err)
	}
	executable, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	err = exec.Command(executable, "-test.run=^$").Run()
	if !errors.Is(err, syscall.EACCES) && !errors.Is(err, syscall.EPERM) {
		t.Fatalf("process capability was not denied: %v", err)
	}
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if listener != nil {
		if closeErr := listener.Close(); closeErr != nil {
			t.Fatal(closeErr)
		}
	}
	if !errors.Is(err, syscall.EACCES) && !errors.Is(err, syscall.EPERM) {
		t.Fatalf("network capability was not denied: %v", err)
	}

	directory := t.TempDir()
	input := filepath.Join(directory, "main.l")
	if err := os.WriteFile(input, []byte("Create \"items\" as a New List of 1 and 2.\nPlace 4 in items at 0.\n"), 0600); err != nil {
		t.Fatal(err)
	}
	invoke := func(want int, args ...string) contracts.Report {
		t.Helper()
		var stdout, stderr bytes.Buffer
		args = append(append([]string{"contracts"}, args...), "--format", "json", input)
		if status := run(context.Background(), args, &stdout, &stderr); status != want {
			t.Fatalf("offline %v exit %d want %d: %s", args, status, want, &stderr)
		}
		var report contracts.Report
		if want != 2 {
			if err := json.Unmarshal(stdout.Bytes(), &report); err != nil {
				t.Fatal(err)
			}
		}
		return report
	}
	report := invoke(0, "inspect")
	if len(report.Bindings) == 0 {
		t.Fatal("offline inspection returned no real bindings")
	}
	invoke(0, "explain", "--binding", string(report.Bindings[0].Occurrence))
	invoke(1, "check", "--profile", "NoBlocking")
	invoke(2, "check", "--profile", "NoBlocking", "--native-environment", filepath.Join(directory, "missing.json"))

	marker := filepath.Join(directory, "historical-command-ran")
	exit := 0
	historical := contracts.SealExecutionRecord(contracts.ExecutionRecord{ID: "test-only:hostile-history", Kind: "environment", Command: []string{"/bin/sh", "-c", "touch " + marker}, WorkingDirectory: directory, Status: "passed", Started: "2026-11-03T00:00:00Z", Finished: "2026-11-03T00:00:01Z", ExitCode: &exit, StdoutFingerprint: "test-only:empty", StderrFingerprint: "test-only:empty"})
	environment := contracts.SealNativeEnvironment(contracts.NativeEnvironment{SchemaVersion: contracts.SchemaVersion, Kind: "native_environment", Producer: "test-only:offline-capability-test", Version: "1", Method: "test-only", Host: "test-only", Target: "test-only", Compiler: "test-only", PanicPolicy: "test-only", Runtime: "test-only", Allocator: "test-only", Inputs: report.Inputs, Dependencies: report.Lowering.Fingerprints, Executions: []contracts.ExecutionRecord{historical}, Limitations: []string{"Historical command is untrusted fixture data; never replay it"}})
	writeEnvironment := func() string {
		t.Helper()
		encoded, err := contracts.EncodeDocument(environment)
		if err != nil {
			t.Fatal(err)
		}
		path := filepath.Join(directory, "environment.json")
		if err := os.WriteFile(path, encoded, 0600); err != nil {
			t.Fatal(err)
		}
		return path
	}
	path := writeEnvironment()
	invoke(0, "inspect", "--native-environment", path)
	invoke(0, "explain", "--binding", string(report.Bindings[0].Occurrence), "--native-environment", path)
	invoke(1, "check", "--profile", "NoBlocking", "--native-environment", path)
	environment.Dependencies = map[string]string{"test-only:stale": "changed"}
	environment = contracts.SealNativeEnvironment(environment)
	invoke(1, "check", "--profile", "NoBlocking", "--native-environment", writeEnvironment())
	if _, err := os.Stat(marker); !os.IsNotExist(err) {
		t.Fatal("historical command was replayed")
	}
	if len(report.Executions) != 0 {
		t.Fatal("inspection fabricated execution evidence")
	}
	if _, err := fmt.Println("offline capability checks completed"); err != nil {
		t.Fatal(err)
	}
}
