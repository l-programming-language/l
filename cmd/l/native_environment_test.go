package main

import (
	"context"
	"errors"
	"fmt"
	"maps"

	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"runtime"
	"slices"
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
		if directory != filepath.Dir(fixture.report.Entry) || !slices.Contains(environment, "CARGO_NET_OFFLINE=true") || !slices.Contains(environment, "RUSTUP_AUTO_INSTALL=0") {
			t.Fatalf("capture lost the expected offline settings or directory %q", directory)
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

func readBuildEnvironment(t *testing.T, path string) contracts.NativeEnvironment {
	t.Helper()
	data := readNativeEvidenceFile(t, path)
	document, err := contracts.DecodeDocument(data)
	if err != nil {
		t.Fatal(err)
	}
	value, ok := document.(contracts.NativeEnvironment)
	if !ok || value.Method != "build" {
		t.Fatalf("unexpected build record: %T", document)
	}
	return value
}

func (fixture *nativeObservationFixture) capture(t *testing.T, target string) contracts.NativeEnvironment {
	t.Helper()
	environment, err := captureNativeEnvironmentWith(context.Background(), fixture.report, target, fixture.observer)
	if err != nil {
		t.Fatal(err)
	}
	return environment
}

func TestNativeEnvironmentCaptureIsExplicitAndOffline(t *testing.T) {
	fixture := newNativeObservationFixture(t)
	environment := fixture.capture(t, fixture.target)
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

func TestNativeEnvironmentExactSourceDependencies(t *testing.T) {
	fixture := newNativeObservationFixture(t)
	environment := fixture.capture(t, "")
	if !maps.Equal(environment.Inputs, fixture.report.Inputs) || !maps.Equal(environment.Dependencies, fixture.report.Lowering.Fingerprints) {
		t.Fatalf("capture lost selected source or implementation identities: %+v", environment)
	}
	for _, mutation := range []string{"unchanged", "changed-source", "missing-source", "extra-source", "changed-dependency", "missing-dependency", "extra-dependency"} {
		t.Run(mutation, func(t *testing.T) {
			inputs, dependencies := maps.Clone(fixture.report.Inputs), maps.Clone(fixture.report.Lowering.Fingerprints)
			want := contracts.Unresolved
			switch mutation {
			case "unchanged":
				want = contracts.Established
			case "changed-source":
				inputs["source:entry"] = "changed"
			case "missing-source":
				delete(inputs, "source:entry")
			case "extra-source":
				inputs["source:extra"] = "extra"
			case "changed-dependency":
				dependencies["helper"] = "changed"
			case "missing-dependency":
				delete(dependencies, "helper")
			case "extra-dependency":
				dependencies["extra"] = "extra"
			}
			if got, reason := contracts.EnvironmentApplicability(environment, inputs, dependencies); got != want {
				t.Fatalf("want %s, got %s: %s", want, got, reason)
			}
			if len(fixture.calls) != 6 {
				t.Fatal("offline applicability replayed tool observations")
			}
		})
	}
}

func TestNativeEvidenceBuildRecordingAndNoReplay(t *testing.T) {
	fixture := newNativeObservationFixture(t)
	environment := fixture.capture(t, "")
	environment.Method, environment.Version = "build", "2"
	environment.Configuration["companion-generation"] = "1"
	recorder := &nativeBuildRecorder{environment: environment}
	if nativeBuildEvidence(context.Background()) != nil || nativeBuildEvidence(context.WithValue(context.Background(), nativeBuildEvidenceKey{}, recorder)) != recorder {
		t.Fatal("native evidence recording is not per-invocation and opt-in")
	}
	command := &exec.Cmd{Path: fixture.paths["cargo"], Args: []string{fixture.paths["cargo"], "+" + nativeToolchainVersion, "build", "--frozen"}, Dir: filepath.Dir(fixture.report.Entry)}
	recorder.record(command, time.Now().UTC(), nil, []byte("failed before launch"), errors.New("test-only failed launch"))
	output := filepath.Join(filepath.Dir(fixture.report.Entry), "failed-build.json")
	if err := recorder.write(output); err != nil {
		t.Fatal(err)
	}
	loaded := readBuildEnvironment(t, output)
	if len(loaded.Executions) != 7 || loaded.Executions[6].Status != "failed" || len(loaded.Artifacts) != 0 || len(fixture.calls) != 6 {
		t.Fatalf("failed build evidence was favorable or replayed: %+v", loaded)
	}
	writeNativeEvidenceFile(t, fixture.paths["cargo"], []byte("changed tool"))
	if err := recorder.verifyTools(); err == nil {
		t.Fatal("changed executable retained captured identity")
	}
}

func nativeTestConfiguration(t *testing.T, root string, environment []string) map[string]string {
	t.Helper()
	configuration := make(map[string]string)
	if err := nativeConfiguration(configuration, root, environment); err != nil {
		t.Fatal(err)
	}
	return configuration
}

func TestNativeEvidenceConfigurationAndWorkspaceIdentity(t *testing.T) {
	root := t.TempDir()
	if err := nativeConfiguration(make(map[string]string), root, []string{"CARGO_HOME=relative"}); err == nil {
		t.Fatal("relative Cargo home could fingerprint configuration in the wrong working directory")
	}
	environment := []string{"CARGO_HOME=" + root, "CARGO_NET_OFFLINE=true", "CARGO_REGISTRIES_PRIVATE_TOKEN=never-record-this-secret", "UNRELATED_SECRET=also-not-recorded"}
	first := nativeTestConfiguration(t, root, environment)
	if strings.Contains(fmt.Sprint(first), "never-record-this-secret") || strings.Contains(fmt.Sprint(first), "also-not-recorded") {
		t.Fatal("environment values leaked into native provenance")
	}
	for _, key := range []string{"CARGO_REGISTRIES_PRIVATE_TOKEN", "UNRELATED_SECRET"} {
		if _, recorded := first["environment:"+key]; recorded {
			t.Fatalf("unlisted environment variable retained even as a hash: %s", key)
		}
	}
	writeNativeEvidenceFile(t, filepath.Join(root, "config.toml"), []byte("[build]\njobs=1\n"))
	second := nativeTestConfiguration(t, root, environment)
	if reflect.DeepEqual(first, second) {
		t.Fatal("newly matching Cargo configuration did not change identity")
	}
	before, err := nativeWorkspaceFingerprint(root)
	if err != nil {
		t.Fatal(err)
	}
	writeNativeEvidenceFile(t, filepath.Join(root, "main.rs"), []byte("fn main() {}\n"))
	after, err := nativeWorkspaceFingerprint(root)
	if err != nil || before == after {
		t.Fatalf("actual workspace bytes did not affect build inputs: %q / %q, %v", before, after, err)
	}
}

func TestNativeEnvironmentConfigurationAllowlist(t *testing.T) {
	root := t.TempDir()
	keys := []string{"PATH", "CARGO_HOME", "RUSTUP_HOME", "RUSTUP_TOOLCHAIN", "RUSTFLAGS", "RUSTDOCFLAGS", "CC", "CXX", "AR", "CARGO_NET_OFFLINE", "CARGO_ENCODED_RUSTFLAGS", "CARGO_TERM_COLOR", "CARGO_PROFILE_DEV_DEBUG", "CARGO_INCREMENTAL", "RUSTC_WRAPPER", "RUSTC_WORKSPACE_WRAPPER", "RUSTC_BOOTSTRAP", "RUSTC_FORCE_INCREMENTAL"}
	var environment []string
	for _, key := range keys {
		value := "selected-" + key
		if key == "CARGO_HOME" {
			value = root
		}
		environment = append(environment, key+"="+value)
	}
	baseline := nativeTestConfiguration(t, root, environment)
	for index, key := range keys {
		t.Run(key, func(t *testing.T) {
			if baseline["environment:"+key] == "" {
				t.Fatal("documented configuration setting was omitted")
			}
			changed := append([]string(nil), environment...)
			changed[index] = key + "=changed-setting"
			if key == "CARGO_HOME" {
				changed[index] = key + "=" + filepath.Join(root, "other")
			}
			actual := nativeTestConfiguration(t, root, changed)
			if actual["environment:"+key] == baseline["environment:"+key] {
				t.Fatal("changed relevant setting retained its configuration identity")
			}
		})
	}
	unlisted := []string{"CARGO_REGISTRY_TOKEN", "CARGO_REGISTRIES_PRIVATE_TOKEN", "CARGO_CREDENTIAL_PROVIDER", "CARGO_UNRELATED", "RUSTC_SECRET", "RUSTC_UNRELATED", "UNRELATED_SECRET"}
	for _, key := range unlisted {
		environment = append(environment, key+"=never-record-even-a-hash")
	}
	actual := nativeTestConfiguration(t, root, environment)
	if !reflect.DeepEqual(baseline, actual) {
		t.Fatal("credential or unrelated variables changed the captured configuration")
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
	for _, destination := range []string{root, source, output, filepath.Join(output+".apps", "evidence.json")} {
		if _, err := resolveNativeEvidenceOutput(destination, []string{source}, output, output+".apps", ""); err == nil {
			t.Fatalf("evidence could overwrite protected build input/output %s", destination)
		}
	}
	if _, err := resolveNativeEvidenceOutput(filepath.Join(root, "evidence.json"), []string{source}, output, output+".apps", ""); err != nil {
		t.Fatal(err)
	}
	alias := filepath.Join(t.TempDir(), "output-alias")
	if err := os.Symlink(root, alias); err != nil {
		t.Fatal(err)
	}
	if _, err := resolveNativeEvidenceOutput(filepath.Join(alias, "entry"), []string{source}, output, output+".apps", ""); err == nil {
		t.Fatal("symlinked parent concealed future executable overwrite")
	}
	target := writeNativeEvidenceFile(t, filepath.Join(root, "target"), []byte("previous"))
	link := filepath.Join(root, "record.json")
	if err := os.Symlink(target, link); err != nil {
		t.Fatal(err)
	}
	physical, err := resolveNativeEvidenceOutput(link, nil, output, output+".apps", "")
	if err != nil {
		t.Fatal(err)
	}
	if err := replaceFile(physical, []byte("next")); err != nil {
		t.Fatal(err)
	}
	if actual, err := os.Readlink(link); err != nil || actual != target {
		t.Fatalf("replaced a valid evidence symlink: %q, %v", actual, err)
	}
	data, err := os.ReadFile(target)
	if err != nil || string(data) != "next" {
		t.Fatalf("validated evidence target was not updated: %q, %v", data, err)
	}

}

func readNativeEvidenceFile(t *testing.T, path string) []byte {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return data
}

func writeNativeEvidenceFile(t *testing.T, path string, data []byte) string {
	t.Helper()
	if err := os.WriteFile(path, data, 0600); err != nil {
		t.Fatal(err)
	}
	return path
}

func nativeEvidenceSource(t *testing.T) string {
	t.Helper()
	source := "Create \"Worker\" as a New Basic App.\nThe App Prints `hello`.\n"
	return writeNativeEvidenceFile(t, filepath.Join(t.TempDir(), "main.l"), []byte(source))
}
