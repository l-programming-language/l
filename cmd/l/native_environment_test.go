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
	distributiontest "github.com/l-programming-language/l/internal/testsupport/distribution"
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
	data := readContractFile(t, path)
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
	writeContractFile(t, fixture.paths["cargo"], []byte("changed tool"))
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
	writeContractFile(t, filepath.Join(root, "config.toml"), []byte("[build]\njobs=1\n"))
	second := nativeTestConfiguration(t, root, environment)
	if reflect.DeepEqual(first, second) {
		t.Fatal("newly matching Cargo configuration did not change identity")
	}
	before, err := nativeWorkspaceFingerprint(root)
	if err != nil {
		t.Fatal(err)
	}
	writeContractFile(t, filepath.Join(root, "main.rs"), []byte("fn main() {}\n"))
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
	target := writeContractFile(t, filepath.Join(root, "target"), []byte("previous"))
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

func TestNativeEnvironmentOfflineConfigurationApplicability(t *testing.T) {
	fixture := newNativeObservationFixture(t)
	environment := fixture.capture(t, "")
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
		malformed.Configuration = maps.Clone(environment.Configuration)
		malformed.Configuration["cargo-config:/not-an-allowed-location/secret"] = "hash"
		if _, _, err := validateNativeEnvironmentConfiguration(malformed); err == nil {
			t.Fatal("arbitrary recorded config path accepted")
		}
	})
	for _, key := range []string{"CARGO_REGISTRY_TOKEN", "CARGO_REGISTRIES_PRIVATE_TOKEN", "RUSTC_SECRET", "UNRELATED_SECRET"} {
		t.Run("unlisted-"+key, func(t *testing.T) {
			malformed := environment
			malformed.Configuration = maps.Clone(environment.Configuration)
			malformed.Configuration["environment:"+key] = nativeBytesHash([]byte("not-even-a-credential-hash"))
			if _, _, err := validateNativeEnvironmentConfiguration(malformed); err == nil {
				t.Fatal("unlisted environment key accepted in supplied evidence")
			}
		})
	}
	configDir := filepath.Join(filepath.Dir(fixture.report.Entry), ".cargo")
	if err := os.Mkdir(configDir, 0700); err != nil {
		t.Fatal(err)
	}
	writeContractFile(t, filepath.Join(configDir, "config.toml"), []byte("[build]\njobs=2\n"))
	if status, _, err := validateNativeEnvironmentConfiguration(environment); err != nil || status != contracts.Unresolved || len(fixture.calls) != 6 {
		t.Fatalf("newly matching config was ignored or commands replayed: %s, %v", status, err)
	}
}

// Change an import and an asset between the member and final entry renders.
// Each recorded dependency must name the bytes that its own unit consumed.
func TestNativeBuildEvidenceUsesPreparedInputs(t *testing.T) {
	for _, drift := range []bool{false, true} {
		t.Run(fmt.Sprintf("changed-between-members-and-entry=%t", drift), func(t *testing.T) {
			source := contractCLIInput(t)
			helper := writeContractFile(t, filepath.Join(filepath.Dir(source), "helper.l"), nil)
			data := readContractFile(t, source)
			writeContractFile(t, source, append([]byte("Use `helper.l`.\n"), data...))
			assets := distributiontest.CompleteOverride(t)
			c, _ := pathTestCompiler(t, assets, source, t.TempDir())
			compiled, err := c.lyger.CompileForContracts(c.mainBytes, source)
			if err != nil {
				t.Fatal(err)
			}
			var inventory contracts.LoweringInventory
			generated, err := generatePackages(c.log, compiled, &inventory)
			if err != nil {
				t.Fatal(err)
			}
			template := filepath.Join(assets, "builtins/templates/install/cargo.lu")
			original := readContractFile(t, template)
			entryTemplate := original
			if drift {
				entryTemplate = append(append([]byte(nil), original...), []byte("\n# fixture save\n")...)
				writeContractFile(t, template, entryTemplate)
				writeContractFile(t, helper, []byte("\n"))
			}
			if _, err := generated.workspace(c.log, "main", "123", &inventory); err != nil {
				t.Fatal(err)
			}
			report := contracts.NewReport(source)
			bindings, lowering, err := compiled.RecordedInventory()
			if err != nil {
				t.Fatal(err)
			}
			report.Inputs, report.Lowering = bindings.Inputs, lowering
			mergeLowering(&report.Lowering, "backend", inventory)
			fixture := newNativeObservationFixture(t)
			fixture.report = report
			env, err := captureNativeEnvironmentWith(context.Background(), report, "", fixture.observer)
			if err != nil {
				t.Fatal(err)
			}
			env.Method, env.Version = "build", "2"
			env.Configuration["companion-generation"] = "123"
			key := contracts.StableID("backend-meaning", "install/cargo.lu")
			if env.Inputs["source:helper.l"] != nativeBytesHash(nil) || env.Dependencies["backend:unit:Worker:"+key] != contracts.StableID("backend-input", original) ||
				env.Dependencies["backend:orchestrator:"+key] != contracts.StableID("backend-input", entryTemplate) {
				t.Fatal("evidence did not bind each phase to its consumed inputs")
			}
			env = contracts.SealNativeEnvironment(env)
			file := scopedWrite(t, t.TempDir(), "environment.json", env)
			fresh, _ := contractJSON(t, source, 0, "inspect", "--assets", assets, "--native-environment", file)
			want := contracts.Established
			if drift {
				want = contracts.Unresolved
			}
			if status, reason := contracts.EnvironmentApplicability(*fresh.NativeEnvironment, fresh.Inputs, fresh.Lowering.Fingerprints); status != want {
				t.Fatalf("offline build identity %s, want %s: %s", status, want, reason)
			}
			if status, _ := contracts.EnvironmentApplicability(env, report.Inputs, fresh.Lowering.Fingerprints); status != want {
				t.Fatal("changed provider retained build assurance independently of source changes")
			}
			for _, token := range []string{"", "../escape", "/abs"} {
				env.Configuration["companion-generation"] = token
				if err := contracts.SealNativeEnvironment(env).Validate(); err == nil {
					t.Fatalf("accepted invalid companion placement %q", token)
				}
			}
			origins := compiled.Compilation.SourceOrigins
			second := origins[0]
			second.SHA256 = nativeBytesHash([]byte("different import bytes"))
			compiled.Compilation.SourceOrigins = append(origins, second)
			if _, _, err := compiled.RecordedInventory(); err == nil {
				t.Fatal("two digests for one logical source certified only the last read")
			}
		})
	}
}
