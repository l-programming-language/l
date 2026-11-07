package main

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"syscall"
	"testing"
	"time"

	"github.com/l-programming-language/l/contracts"
)

const contractsOfflineHelper = "L_CONTRACTS_OFFLINE_CAPABILITY_HELPER"

// The child runs real source-only commands with kernel-enforced process and
// socket denial. No PATH shim can stand in for this capability boundary.
func TestContractsOfflineNoProcessOrNetwork(t *testing.T) {
	runOfflineHelper(t)
}

func runOfflineHelper(t *testing.T, env ...string) {
	t.Helper()
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
	command.Env = append(append(os.Environ(), contractsOfflineHelper+"=1"), env...)
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

	if record := os.Getenv("L_CONTRACTS_BUILD_RECORD"); record != "" {
		source := os.Getenv("L_CONTRACTS_BUILD_SOURCE")
		report, _ := contractJSON(t, source, 0, "inspect", "--native-environment", record)
		if status, reason := contracts.EnvironmentApplicability(*report.NativeEnvironment, report.Inputs, report.Lowering.Fingerprints); status != contracts.Established ||
			!slices.Contains(report.Scope.Exclusions, reason) {
			t.Fatalf("offline build evidence: %s, %s, %v", status, reason, report.Scope.Exclusions)
		}
		changed := *report.NativeEnvironment
		token := "10"
		if changed.Configuration["companion-generation"] == token {
			token = "11"
		}
		changed.Configuration["companion-generation"] = token
		replaced := scopedWrite(t, t.TempDir(), "different-generation.json", contracts.SealNativeEnvironment(changed))
		other, _ := contractJSON(t, source, 0, "inspect", "--native-environment", replaced)
		if status, _ := contracts.EnvironmentApplicability(*other.NativeEnvironment, other.Inputs, other.Lowering.Fingerprints); status != contracts.Unresolved {
			t.Fatal("changed artifact placement retained build assurance")
		}
	}
	directory := t.TempDir()
	input := filepath.Join(directory, "main.l")
	writeContractFile(t, input, []byte("Create \"items\" as a New List of 1 and 2.\nPlace 4 in items at 0.\n"))
	report, _ := contractJSON(t, input, 0, "inspect")
	if len(report.Bindings) == 0 {
		t.Fatal("offline inspection returned no real bindings")
	}
	contractJSON(t, input, 0, "explain", "--binding", string(report.Bindings[0].Occurrence))
	contractJSON(t, input, 1, "check", "--profile", "NoBlocking")
	contractJSON(t, input, 2, "check", "--profile", "NoBlocking", "--native-environment", filepath.Join(directory, "missing.json"))

	marker := filepath.Join(directory, "historical-command-ran")
	exit := 0
	historical := contracts.SealExecutionRecord(contracts.ExecutionRecord{ID: "test-only:hostile-history", Kind: "environment", Command: []string{"/bin/sh", "-c", "touch " + marker}, WorkingDirectory: directory, Status: "passed", Started: "2026-11-03T00:00:00Z", Finished: "2026-11-03T00:00:01Z", ExitCode: &exit, StdoutFingerprint: "test-only:empty", StderrFingerprint: "test-only:empty"})
	environment := contracts.SealNativeEnvironment(contracts.NativeEnvironment{SchemaVersion: contracts.SchemaVersion, Kind: "native_environment", Producer: "test-only:offline-capability-test", Version: "1", Method: "test-only", Host: "test-only", Target: "test-only", Compiler: "test-only", PanicPolicy: "test-only", Runtime: "test-only", Allocator: "test-only", Inputs: report.Inputs, Dependencies: report.Lowering.Fingerprints, Executions: []contracts.ExecutionRecord{historical}, Limitations: []string{"Historical command is untrusted fixture data; never replay it"}})
	path := scopedWrite(t, directory, "environment.json", environment)
	contractJSON(t, input, 0, "inspect", "--native-environment", path)
	contractJSON(t, input, 0, "explain", "--binding", string(report.Bindings[0].Occurrence), "--native-environment", path)
	contractJSON(t, input, 1, "check", "--profile", "NoBlocking", "--native-environment", path)
	environment.Dependencies = map[string]string{"test-only:stale": "changed"}
	environment = contracts.SealNativeEnvironment(environment)
	contractJSON(t, input, 1, "check", "--profile", "NoBlocking", "--native-environment", scopedWrite(t, directory, "environment.json", environment))
	if _, err := os.Stat(marker); !os.IsNotExist(err) {
		t.Fatal("historical command was replayed")
	}
	if len(report.Executions) != 0 {
		t.Fatal("inspection fabricated execution evidence")
	}
	assertOfflineRuntimeResources(t)
	if _, err := fmt.Println("offline capability checks completed"); err != nil {
		t.Fatal(err)
	}
}
