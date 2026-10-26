package native

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

type helperCargoFixture struct{ root, project, target, artifact string }

// These commands expose the Build observer boundary, not native compilation.
// Their arbitrary artifact bytes are never passed off as real Rust provenance.
func newHelperCargoFixture(t *testing.T) helperCargoFixture {
	t.Helper()
	tool := controlledCargo(t)
	f := helperCargoFixture{root: t.TempDir()}
	f.project, f.target = filepath.Join(f.root, "project"), filepath.Join(f.root, "target")
	f.artifact = filepath.Join(f.root, "controlled-artifact")
	for name, data := range map[string]string{
		"project/Cargo.toml":  "[package]\nname = \"helper\"\nversion = \"0.0.0\"\n",
		"project/src/main.rs": "independent source A",
		"controlled-artifact": "independent observer bytes\x00tail",
	} {
		path := filepath.Join(f.root, name)
		if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte(data), 0700); err != nil {
			t.Fatal(err)
		}
	}
	bin := filepath.Join(f.root, "tools")
	if err := os.Mkdir(bin, 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(tool, filepath.Join(bin, "rustup")); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", bin+string(os.PathListSeparator)+os.Getenv("PATH"))
	t.Setenv("L_CONTROLLED_CARGO_MODE", "helper")
	t.Setenv("L_HELPER_WORKSPACE", f.project)
	t.Setenv("L_HELPER_ROOT", f.root)
	for _, name := range []string{"L_HELPER_LOCATE_STDOUT", "L_HELPER_LOCATE_STDERR", "L_HELPER_LOCATE_STATUS", "L_HELPER_BUILD_STDOUT", "L_HELPER_BUILD_STDERR", "L_HELPER_BUILD_STATUS", "L_HELPER_RELEASE"} {
		t.Setenv(name, "")
	}
	return f
}

func (f helperCargoFixture) stream(t *testing.T, name string, data []byte) {
	t.Helper()
	path := filepath.Join(f.root, name)
	if err := os.WriteFile(path, data, 0600); err != nil {
		t.Fatal(err)
	}
	t.Setenv(name, path)
}

func (f helperCargoFixture) artifactMessage(t *testing.T, name string, fresh bool) []byte {
	t.Helper()
	data, err := json.Marshal(map[string]any{"reason": "compiler-artifact", "package_id": "controlled#" + name, "executable": f.artifact, "target": map[string]any{"name": name, "kind": []string{"bin"}}, "fresh": fresh})
	if err != nil {
		t.Fatal(err)
	}
	return append(data, '\n')
}

func TestCargoBuildKeepsLineOrientedArtifacts(t *testing.T) {
	for _, name := range []string{"prefix", "between", "trailing", "malformed", "malformed-fields", "unrelated", "long-line", "no-final-newline", "duplicate", "empty", "no-executable"} {
		t.Run(name, func(t *testing.T) {
			f := newHelperCargoFixture(t)
			first, second := f.artifactMessage(t, "first", true), f.artifactMessage(t, "second", false)
			var output []byte
			want := []string{"first", "second"}
			switch name {
			case "prefix":
				output = append([]byte("native macro output\n"), append(first, second...)...)
			case "between":
				output = append(append(first, []byte("native macro output\n")...), second...)
			case "trailing":
				output = append(append(first, second...), []byte("native tail\x00\n")...)
			case "malformed":
				output = append(append(first, []byte("{incomplete JSON\n\n")...), second...)
			case "malformed-fields":
				output = append(first, []byte("unrelated native output\n")...)
				for _, field := range []string{"reason", "package_id", "executable", "target"} {
					event := map[string]any{"reason": "compiler-artifact", "package_id": "invalid", "executable": f.artifact, "target": map[string]any{"name": "invalid", "kind": []string{"bin"}}}
					event[field] = 42
					data, err := json.Marshal(event)
					if err != nil {
						t.Fatal(err)
					}
					output = append(append(output, data...), '\n')
				}
				output = append(output, second...)
			case "unrelated":
				output = append(append(first, []byte("{\"reason\":\"build-script-executed\"}\n")...), second...)
			case "long-line":
				output = append(append(append(first, bytes.Repeat([]byte{'x'}, 128*1024)...), '\n'), second...)
			case "no-final-newline":
				output = append(first, bytes.TrimSuffix(second, []byte{'\n'})...)
			case "duplicate":
				output = append(first, first...)
				want = []string{"first", "first"}
			case "empty":
				want = nil
			case "no-executable":
				output = []byte("{\"reason\":\"compiler-artifact\",\"executable\":null}\n")
				want = nil
			}
			before := bytes.Clone(output)
			f.stream(t, "L_HELPER_BUILD_STDOUT", output)
			f.stream(t, "L_HELPER_BUILD_STDERR", []byte("build stderr\x00tail\n"))
			result := Build(t, f.project, BuildOptions{TargetDir: f.target})
			if result.Err != nil {
				t.Fatalf("controlled build failed: %v; %s", result.Err, result.Stderr)
			}
			if !bytes.Equal(result.Stdout, before) || string(result.Stderr) != "build stderr\x00tail\n" || !bytes.Equal(output, before) {
				t.Fatalf("raw streams/input changed: %q / %q", result.Stdout, result.Stderr)
			}
			var names []string
			for _, artifact := range result.Artifacts {
				names = append(names, artifact.TargetName)
				if artifact.PackageID != "controlled#"+artifact.TargetName || !reflect.DeepEqual(artifact.TargetKinds, []string{"bin"}) {
					t.Fatalf("artifact identity changed: %+v", artifact)
				}
				if data, err := os.ReadFile(artifact.Executable); err != nil || string(data) != "independent observer bytes\x00tail" {
					t.Fatalf("snapshot bytes=%q %v", data, err)
				}
			}
			if !reflect.DeepEqual(names, want) {
				t.Fatalf("artifact order=%v want %v", names, want)
			}
			if name == "duplicate" && result.Artifacts[0].Executable == result.Artifacts[1].Executable {
				t.Fatal("duplicate records lost their separate immutable snapshots")
			}
		})
	}
}

func TestCargoWorkspaceDiscoveryCancellation(t *testing.T) {
	f := newHelperCargoFixture(t)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	environment := CleanEnvironment("controlled-rustc")
	before := append([]string(nil), environment...)
	r := locateCargoWorkspace(ctx, controlledCargo(t), f.project, environment)
	if !errors.Is(r.Err, context.Canceled) {
		t.Fatalf("discovery cancellation cause changed: %v", r.Err)
	}
	if r.Stdout != nil || r.Stderr != nil || len(r.Artifacts) != 0 {
		t.Fatalf("pre-canceled discovery produced output: %+v", r)
	}
	if len(r.Command) < 2 || r.Command[1] != "locate-project" || !reflect.DeepEqual(before, environment) {
		t.Fatalf("discovery command/environment changed: %+v", r)
	}
	if _, err := os.Stat(filepath.Join(f.root, "locate-entered")); !os.IsNotExist(err) {
		t.Fatalf("Cargo started despite canceled discovery: %v", err)
	}
}

func TestCargoBuildPreservesNativeFailure(t *testing.T) {
	f := newHelperCargoFixture(t)
	stdout := append([]byte("native diagnostic prefix\n"), f.artifactMessage(t, "before-failure", true)...)
	f.stream(t, "L_HELPER_BUILD_STDOUT", stdout)
	f.stream(t, "L_HELPER_BUILD_STDERR", []byte("native failure stderr\x00tail\n"))
	t.Setenv("L_HELPER_BUILD_STATUS", "7")
	r := Build(t, f.project, BuildOptions{TargetDir: f.target})
	var exit *exec.ExitError
	if !errors.As(r.Err, &exit) || exit.ExitCode() != 7 {
		t.Fatalf("native failure cause changed: %v", r.Err)
	}
	if !bytes.Equal(r.Stdout, stdout) || string(r.Stderr) != "native failure stderr\x00tail\n" {
		t.Fatalf("native failure streams changed: %q / %q", r.Stdout, r.Stderr)
	}
	if len(r.Artifacts) != 1 || r.Artifacts[0].TargetName != "before-failure" {
		t.Fatalf("prior native artifact lost after error: %+v", r.Artifacts)
	}
}

func TestCargoWorkspaceDiscoveryPreservesFailureStreams(t *testing.T) {
	f := newHelperCargoFixture(t)
	f.stream(t, "L_HELPER_LOCATE_STDOUT", []byte("discovery stdout\x00tail\n"))
	f.stream(t, "L_HELPER_LOCATE_STDERR", []byte("discovery stderr\r\n\x00tail"))
	t.Setenv("L_HELPER_LOCATE_STATUS", "7")
	result := Build(t, f.project, BuildOptions{TargetDir: f.target})
	var exit *exec.ExitError
	if !errors.As(result.Err, &exit) || exit.ExitCode() != 7 {
		t.Fatalf("discovery cause lost: %v", result.Err)
	}
	if string(result.Stdout) != "discovery stdout\x00tail\n" || string(result.Stderr) != "discovery stderr\r\n\x00tail" {
		t.Fatalf("discovery streams=%q / %q", result.Stdout, result.Stderr)
	}
	if len(result.Command) < 2 || result.Command[1] != "locate-project" || len(result.Artifacts) != 0 {
		t.Fatalf("wrong failure phase: %+v", result)
	}
	if _, err := os.Stat(filepath.Join(f.root, "build-entered")); !os.IsNotExist(err) {
		t.Fatalf("build ran after discovery failure: %v", err)
	}
}

func TestCargoEnvironmentFiltersAlternateOverrides(t *testing.T) {
	removed := []string{"CARGO_BUILD_RUSTFLAGS", "CARGO_BUILD_RUSTDOCFLAGS", "CARGO_ENCODED_RUSTDOCFLAGS", "CARGO_BUILD_BUILD_DIR", "CARGO_BUILD_TARGET_DIR", "RUSTFLAGS", "CARGO_ENCODED_RUSTFLAGS", "CARGO_BUILD_TARGET"}
	for _, name := range removed {
		t.Setenv(name, "must not reach Cargo")
	}
	preserved := map[string]string{"CARGO_LOG": "cargo::compiler::fingerprint=info", "CARGO_BUILD_JOBS": "3", "CARGO_HOME": "registry-cache", "RUSTUP_HOME": "toolchain-cache", "L_HARNESS_TIMING_DIR": "diagnostic-artifacts"}
	for name, value := range preserved {
		t.Setenv(name, value)
	}
	before := os.Environ()
	got := map[string]string{}
	for _, entry := range CleanEnvironment("selected-rustc") {
		key, value, _ := strings.Cut(entry, "=")
		got[key] = value
	}
	for _, name := range removed {
		if value, present := got[name]; present {
			t.Errorf("override %s survived: %q", name, value)
		}
	}
	for name, value := range preserved {
		if got[name] != value {
			t.Errorf("required environment %s=%q want%q", name, got[name], value)
		}
	}
	if got["PATH"] != os.Getenv("PATH") || got["RUSTC"] != "selected-rustc" || !reflect.DeepEqual(before, os.Environ()) {
		t.Error("environment preparation changed caller or pinned tool selection")
	}
}
