package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/l-programming-language/l/contracts"
)

type nativeObservationFixture struct {
	report   contracts.Report
	observer nativeEnvironmentObserver
	calls    [][]string
	paths    map[string]string
	target   string
	failure  bool
}

func newNativeObservationFixture(t *testing.T) *nativeObservationFixture {
	t.Helper()
	root := t.TempDir()
	fixture := &nativeObservationFixture{report: contracts.NewReport(filepath.Join(root, "entry.l")), paths: make(map[string]string)}
	fixture.report.Inputs["source:entry"] = "source-fingerprint"
	fixture.report.Lowering.Fingerprints = map[string]string{"helper": "helper-fingerprint"}
	architecture := runtime.GOARCH
	if architecture == "arm64" {
		architecture = "aarch64"
	} else if architecture == "amd64" {
		architecture = "x86_64"
	} else if architecture == "386" {
		architecture = "i686"
	}
	fixture.target = architecture + "-unknown-linux-gnu"
	if runtime.GOOS == "darwin" {
		fixture.target = architecture + "-apple-darwin"
	}
	for _, name := range []string{"rustup", "cargo", "rustc", "cargo-launcher"} {
		path := filepath.Join(root, name)
		if err := os.WriteFile(path, []byte("test-only identity: "+name), 0700); err != nil {
			t.Fatal(err)
		}
		fixture.paths[name] = path
	}
	fixture.observer.lookup = func(name string) (string, error) {
		if name == "cargo" {
			return fixture.paths["cargo-launcher"], nil
		}
		return fixture.paths[name], nil
	}
	fixture.observer.query = func(_ context.Context, executable string, args []string, directory string, environment []string) ([]byte, []byte, int, error) {
		fixture.calls = append(fixture.calls, append([]string{executable}, args...))
		if directory != root || !containsNativeEnvironment(environment, "CARGO_NET_OFFLINE=true") || !containsNativeEnvironment(environment, "RUSTUP_AUTO_INSTALL=0") {
			t.Fatalf("capture lost explicit offline directory/environment: %q / %v", directory, environment)
		}
		if fixture.failure {
			return nil, []byte("test-only missing toolchain"), 1, errors.New("query failed")
		}
		name := filepath.Base(executable)
		switch {
		case name == "rustup" && reflect.DeepEqual(args, []string{"--version"}):
			return []byte("rustup 1.28.2 (test-only)\n"), nil, 0, nil
		case name == "rustup" && len(args) == 4 && args[0] == "which" && args[1] == "--toolchain" && args[2] == nativeToolchainVersion:
			return []byte(fixture.paths[args[3]] + "\n"), nil, 0, nil
		case name == "rustc" && reflect.DeepEqual(args, []string{"--version", "--verbose"}):
			return fmt.Appendf(nil, "rustc %s (test-only)\nhost: %s\n", nativeToolchainVersion, fixture.target), nil, 0, nil
		case name == "cargo" && reflect.DeepEqual(args, []string{"--version", "--verbose"}), name == "cargo-launcher" && reflect.DeepEqual(args, []string{"+" + nativeToolchainVersion, "--version", "--verbose"}):
			return fmt.Appendf(nil, "cargo %s (test-only)\n", nativeToolchainVersion), nil, 0, nil
		default:
			t.Fatalf("capture attempted a non-observation operation: %s %v", executable, args)
			return nil, nil, -1, errors.New("unexpected observation")
		}
	}
	return fixture
}

func containsNativeEnvironment(environment []string, want string) bool {
	for _, item := range environment {
		if item == want {
			return true
		}
	}
	return false
}

func TestNativeEnvironmentCaptureIsExplicitAndOffline(t *testing.T) {
	fixture := newNativeObservationFixture(t)
	environment, err := captureNativeEnvironmentWith(context.Background(), fixture.report, fixture.target, fixture.observer)
	if err != nil {
		t.Fatal(err)
	}
	if len(fixture.calls) != 6 || len(environment.Executions) != 6 || len(environment.Tools) != 4 || len(environment.Artifacts) != 0 || environment.Method != "capture" {
		t.Fatalf("capture did more than selected tool observations: %+v / %+v", fixture.calls, environment)
	}
	for _, record := range environment.Executions {
		if record.Kind != "environment" || record.Status != "passed" || record.Validate() != nil {
			t.Fatalf("observation lacks exact actual execution evidence: %+v", record)
		}
	}
	data, err := contracts.EncodeDocument(environment)
	if err != nil {
		t.Fatal(err)
	}
	loaded, err := contracts.DecodeDocument(data)
	if err != nil {
		t.Fatal(err)
	}
	decoded, ok := loaded.(contracts.NativeEnvironment)
	if !ok {
		t.Fatalf("wrong native environment document kind: %T", loaded)
	}
	if status, _ := contracts.EnvironmentApplicability(decoded, fixture.report.Inputs, fixture.report.Lowering.Fingerprints); status != contracts.Established || len(fixture.calls) != 6 {
		t.Fatalf("offline validation invoked a tool or lost input applicability: %+v", decoded)
	}
	fixture.report.Inputs["source:entry"] = "changed"
	if status, _ := contracts.EnvironmentApplicability(decoded, fixture.report.Inputs, fixture.report.Lowering.Fingerprints); status != contracts.Unresolved || len(fixture.calls) != 6 {
		t.Fatal("stale environment remained applicable or replayed commands")
	}
}

func TestNativeEnvironmentCaptureFailureAndTarget(t *testing.T) {
	for _, mode := range []string{"failure", "different-target", "cross-host"} {
		t.Run(mode, func(t *testing.T) {
			fixture := newNativeObservationFixture(t)
			target := ""
			switch mode {
			case "failure":
				fixture.failure = true
			case "different-target":
				target = "wasm32-unknown-unknown"
			case "cross-host":
				fixture.target = "wasm32-unknown-unknown"
			}
			environment, err := captureNativeEnvironmentWith(context.Background(), fixture.report, target, fixture.observer)
			if err == nil || environment.Fingerprint != "" {
				t.Fatalf("invalid environment became a sealed observation: %+v, %v", environment, err)
			}
			if mode == "failure" && (len(environment.Executions) != 1 || environment.Executions[0].Status != "failed") {
				t.Fatalf("failed tool query was hidden: %+v", environment.Executions)
			}
		})
	}
}

func TestNativeEvidenceBuildRecordingAndNoReplay(t *testing.T) {
	fixture := newNativeObservationFixture(t)
	environment, err := captureNativeEnvironmentWith(context.Background(), fixture.report, "", fixture.observer)
	if err != nil {
		t.Fatal(err)
	}
	environment.Method = "build"
	recorder := &nativeBuildRecorder{environment: environment}
	if nativeBuildEvidence(context.Background()) != nil || nativeBuildEvidence(context.WithValue(context.Background(), nativeBuildEvidenceKey{}, recorder)) != recorder {
		t.Fatal("native evidence recording is not per-invocation and opt-in")
	}
	command := &exec.Cmd{Path: fixture.paths["cargo"], Args: []string{fixture.paths["cargo"], "+" + nativeToolchainVersion, "build", "--frozen"}, Dir: filepath.Dir(fixture.report.Entry)}
	recorder.record(command, time.Now().UTC(), nil, []byte("failed before launch"), errors.New("test-only failed launch"))
	output := filepath.Join(filepath.Dir(fixture.report.Entry), "failed-build.json")
	if err := recorder.write(output, fixture.report.Entry); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(output)
	if err != nil {
		t.Fatal(err)
	}
	decoded, err := contracts.DecodeDocument(data)
	if err != nil {
		t.Fatal(err)
	}
	loaded, ok := decoded.(contracts.NativeEnvironment)
	if !ok || len(loaded.Executions) != 7 || loaded.Executions[6].Status != "failed" || len(loaded.Artifacts) != 0 || len(fixture.calls) != 6 {
		t.Fatalf("failed build evidence was favorable or replayed: %+v", decoded)
	}
	if err := os.WriteFile(fixture.paths["cargo"], []byte("changed tool"), 0700); err != nil {
		t.Fatal(err)
	}
	if err := recorder.verifyTools(); err == nil {
		t.Fatal("changed executable retained captured identity")
	}
}

func TestNativeEvidenceConfigurationAndWorkspaceIdentity(t *testing.T) {
	root := t.TempDir()
	if err := nativeConfiguration(make(map[string]string), root, []string{"CARGO_HOME=relative"}); err == nil {
		t.Fatal("relative Cargo home could fingerprint configuration in the wrong working directory")
	}
	environment := []string{"CARGO_HOME=" + root, "CARGO_NET_OFFLINE=true", "CARGO_REGISTRIES_PRIVATE_TOKEN=never-record-this-secret", "UNRELATED_SECRET=also-not-recorded"}
	first := make(map[string]string)
	if err := nativeConfiguration(first, root, environment); err != nil {
		t.Fatal(err)
	}
	if strings.Contains(fmt.Sprint(first), "never-record-this-secret") || strings.Contains(fmt.Sprint(first), "also-not-recorded") {
		t.Fatal("environment values leaked into native provenance")
	}
	if err := os.WriteFile(filepath.Join(root, "config.toml"), []byte("[build]\njobs=1\n"), 0600); err != nil {
		t.Fatal(err)
	}
	second := make(map[string]string)
	if err := nativeConfiguration(second, root, environment); err != nil {
		t.Fatal(err)
	}
	if reflect.DeepEqual(first, second) {
		t.Fatal("newly matching Cargo configuration did not change identity")
	}
	before, err := nativeWorkspaceFingerprint(root)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "main.rs"), []byte("fn main() {}\n"), 0600); err != nil {
		t.Fatal(err)
	}
	after, err := nativeWorkspaceFingerprint(root)
	if err != nil || before == after {
		t.Fatalf("actual workspace bytes did not affect build inputs: %q / %q, %v", before, after, err)
	}
}

func TestNativeEvidenceBuildOptionAndDestinations(t *testing.T) {
	command, err := parseCommand([]string{"build", "--native-evidence-output", "evidence.json", "entry.l"})
	if err != nil || command.NativeEvidenceOutput != "evidence.json" || command.CLI.mainPath != "entry.l" {
		t.Fatalf("native evidence option changed CLI projection: %+v, %v", command, err)
	}
	_, _, options, err := resolveBuildCommand(command)
	if err != nil || !filepath.IsAbs(options.NativeEvidenceOutput) {
		t.Fatalf("evidence path was not resolved: %+v, %v", options, err)
	}
	for _, args := range [][]string{{"build", "--native-evidence-output="}, {"build", "entry.l", "--native-evidence-output", "evidence.json"}} {
		if _, err := parseCommand(args); err == nil {
			t.Fatalf("invalid evidence invocation accepted: %v", args)
		}
	}
	root := t.TempDir()
	source, output := filepath.Join(root, "entry.l"), filepath.Join(root, "entry")
	for _, destination := range []string{source, output, filepath.Join(output+".apps", "evidence.json")} {
		if err := validateNativeEvidenceOutput(destination, []string{source}, output, output+".apps", ""); err == nil {
			t.Fatalf("evidence could overwrite protected build input/output %s", destination)
		}
	}
	if err := validateNativeEvidenceOutput(filepath.Join(root, "evidence.json"), []string{source}, output, output+".apps", ""); err != nil {
		t.Fatal(err)
	}
	alias := filepath.Join(t.TempDir(), "output-alias")
	if err := os.Symlink(root, alias); err != nil {
		t.Fatal(err)
	}
	if err := validateNativeEvidenceOutput(filepath.Join(alias, "entry"), []string{source}, output, output+".apps", ""); err == nil {
		t.Fatal("symlinked parent concealed future executable overwrite")
	}
}

func TestNativeEnvironmentOfflineConfigurationApplicability(t *testing.T) {
	fixture := newNativeObservationFixture(t)
	environment, err := captureNativeEnvironmentWith(context.Background(), fixture.report, "", fixture.observer)
	if err != nil {
		t.Fatal(err)
	}
	for _, path := range fixture.paths {
		if err := os.Remove(path); err != nil {
			t.Fatal(err)
		}
	}
	if status, _, err := validateNativeEnvironmentConfiguration(environment); err != nil || status != contracts.Established || len(fixture.calls) != 6 {
		t.Fatalf("offline check re-observed historical tools: %s, %v", status, err)
	}
	t.Run("environment", func(t *testing.T) {
		t.Setenv("CC", "test-only-changed-compiler")
		if status, _, err := validateNativeEnvironmentConfiguration(environment); err != nil || status != contracts.Unresolved {
			t.Fatalf("changed relevant environment remained applicable: %s, %v", status, err)
		}
	})
	t.Run("malformed-config-path", func(t *testing.T) {
		malformed := environment
		malformed.Configuration = cloneNativeStrings(environment.Configuration)
		malformed.Configuration["cargo-config:/not-an-allowed-location/secret"] = "hash"
		if _, _, err := validateNativeEnvironmentConfiguration(malformed); err == nil {
			t.Fatal("arbitrary recorded config path accepted")
		}
	})
	configDir := filepath.Join(filepath.Dir(fixture.report.Entry), ".cargo")
	if err := os.Mkdir(configDir, 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(configDir, "config.toml"), []byte("[build]\njobs=2\n"), 0600); err != nil {
		t.Fatal(err)
	}
	if status, _, err := validateNativeEnvironmentConfiguration(environment); err != nil || status != contracts.Unresolved || len(fixture.calls) != 6 {
		t.Fatalf("newly matching config was ignored or commands replayed: %s, %v", status, err)
	}
}
