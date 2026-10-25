package native

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"testing"
)

const proxyConfigEnvironment = "L_HARNESS_CARGO_CONFIG"

type cliProxyFixture struct {
	probe     *CLIRecorder
	workspace string
	artifact  string
	output    string
	env       []string
}

// This controlled command checks the observer protocol only. Real parity tests
// still require Cargo-produced native executable identities and publication.
func newCLIProxyFixture(t *testing.T) cliProxyFixture {
	t.Helper()
	probe := CLIProbe(t)
	info, err := os.Lstat(filepath.Join(probe.Directory, "cargo"))
	if err != nil || info.Mode()&os.ModeSymlink == 0 {
		t.Fatalf("Cargo observer must reuse the compiled recorder through a probe-local symlink: %v %v", info, err)
	}
	workspace := filepath.Join(t.TempDir(), "workspace with spaces")
	if err := os.MkdirAll(filepath.Join(workspace, "src"), 0700); err != nil {
		t.Fatal(err)
	}
	workspace, err = filepath.EvalSymlinks(workspace)
	if err != nil {
		t.Fatal(err)
	}
	for path, content := range map[string]string{"Cargo.toml": "independent manifest", "Cargo.lock": "independent retained lock", "src/main.rs": "independent native source"} {
		if err := os.WriteFile(filepath.Join(workspace, path), []byte(content), 0600); err != nil {
			t.Fatal(err)
		}
	}
	artifact := filepath.Join(workspace, "native artifact")
	if err := os.WriteFile(artifact, []byte("observer protocol fixture\x00native bytes"), 0700); err != nil {
		t.Fatal(err)
	}
	message, err := json.Marshal(map[string]any{"reason": "compiler-artifact", "executable": artifact, "target": map[string]any{"kind": []string{"bin"}}, "fresh": false})
	if err != nil {
		t.Fatal(err)
	}
	cargo := controlledCargo(t)
	config, callErr := json.Marshal(map[string]string{"Directory": probe.Directory, "Cargo": cargo, "Test": t.Name(), "Host": probe.tools.Host, "RunID": os.Getenv(evidenceRunEnvironment)})
	if callErr != nil {
		t.Fatal(callErr)
	}
	if err := os.WriteFile(filepath.Join(probe.Directory, "cargo-proxy.json"), config, 0600); err != nil {
		t.Fatal(err)
	}
	env := probe.Env(append(os.Environ(), proxyConfigEnvironment+"=/wrong/earlier/config"))
	count := 0
	for _, entry := range env {
		if strings.HasPrefix(entry, proxyConfigEnvironment+"=") {
			count++
			if entry != proxyConfigEnvironment+"="+filepath.Join(probe.Directory, "cargo-proxy.json") {
				t.Fatalf("stale observer configuration: %q", entry)
			}
		}
	}
	if count != 1 {
		t.Fatalf("configuration count = %d, want 1", count)
	}
	env = append(env, "L_CONTROLLED_CARGO_MODE=proxy", "L_PROXY_ROOT="+probe.Directory, "L_PROXY_WORKSPACE="+workspace, "L_PROXY_ARTIFACT="+artifact, "L_PROXY_VALUE=literal spaces ' quotes $() ; value")
	return cliProxyFixture{probe: probe, workspace: workspace, artifact: artifact, output: string(message) + "\n", env: env}
}

func (f cliProxyFixture) invoke(args []string, stdin string, extra ...string) ([]byte, []byte, int) {
	cmd := exec.Command(filepath.Join(f.probe.Directory, "cargo"), args...)
	cmd.Dir, cmd.Env = f.workspace, append(append([]string(nil), f.env...), extra...)
	cmd.Stdin = strings.NewReader(stdin)
	var out, diag bytes.Buffer
	cmd.Stdout, cmd.Stderr = &out, &diag
	status := 0
	if err := cmd.Run(); err != nil {
		status = -1
		if exit, ok := err.(*exec.ExitError); ok {
			status = exit.ExitCode()
		}
	}
	return out.Bytes(), diag.Bytes(), status
}

func TestCLIProxyPreservesInvocationStreamsAndStatus(t *testing.T) {
	for _, timing := range []bool{false, true} {
		for _, outcome := range []string{"success", "exit", "signal"} {
			t.Run(fmt.Sprintf("timing_%t_%s", timing, outcome), func(t *testing.T) {
				trace := ""
				if timing {
					trace = filepath.Join(t.TempDir(), "trace")
				}
				t.Setenv("L_HARNESS_TIMING_DIR", trace)
				f := newCLIProxyFixture(t)
				manifest := filepath.Join(f.workspace, "Cargo.toml")
				args := []string{"+" + Version, "build", "--manifest-path", manifest, "literal ' quote $() ; argument"}
				if timing {
					args = []string{"+" + Version, "build", "--manifest-path=" + manifest, "literal ' quote $() ; argument"}
				}
				extra, wantStatus := []string{}, 0
				if outcome == "exit" {
					extra, wantStatus = []string{"L_PROXY_STATUS=7"}, 7
				} else if outcome == "signal" {
					extra, wantStatus = []string{"L_PROXY_SIGNAL=yes"}, 143
				}
				stdin := "stdin\x00with raw bytes\n"
				out, diag, status := f.invoke(args, stdin, extra...)
				if status != wantStatus || string(out) != f.output+stdin || string(diag) != "native stderr\x00tail\n" {
					t.Fatalf("status/streams = %d/%q/%q, want %d", status, out, diag, wantStatus)
				}
				wantNative := append([]string(nil), args[1:]...)
				if timing {
					wantNative = append(wantNative, "--timings")
				}
				data, err := os.ReadFile(filepath.Join(f.probe.Directory, "native.args"))
				if err != nil || string(data) != strings.Join(wantNative, "\x00")+"\x00" {
					t.Fatalf("native argv %q: %v", data, err)
				}
				for path, want := range map[string]string{"native.cwd": f.workspace + "\n", "native.env": "literal spaces ' quotes $() ; value", "metadata.args": "metadata\x00--manifest-path\x00" + manifest + "\x00--frozen\x00--offline\x00--no-deps\x00--format-version\x001\x00"} {
					if path == "metadata.args" && timing {
						want = "metadata\x00--manifest-path=" + manifest + "\x00--frozen\x00--offline\x00--no-deps\x00--format-version\x001\x00"
					}
					data, err := os.ReadFile(filepath.Join(f.probe.Directory, path))
					if err != nil || string(data) != want {
						t.Fatalf("%s = %q, want %q: %v", path, data, want, err)
					}
				}
				paths, callErr := filepath.Glob(filepath.Join(f.probe.Directory, "cargo.*.jsonl"))
				if callErr != nil {
					t.Fatal(callErr)
				}
				if len(paths) != 1 {
					t.Fatalf("records = %v", paths)
				}
				base := strings.TrimSuffix(paths[0], ".jsonl")
				for suffix, want := range map[string]string{".jsonl": f.output + stdin, ".args": strings.Join(args, "\n") + "\n", ".cwd": f.workspace + "\n", ".workspace/src/main.rs": "independent native source"} {
					data, err := os.ReadFile(base + suffix)
					if err != nil || string(data) != want {
						t.Fatalf("record %s = %q: %v", suffix, data, err)
					}
				}
				if len(f.probe.artifacts(t)) != 1 {
					t.Fatal("missing artifact snapshot")
				}
			})
		}
	}
}

func TestCLIProxyRejectsWrongSelectorAndPreservesErrorPrecedence(t *testing.T) {
	t.Setenv("L_HARNESS_TIMING_DIR", "")
	t.Run("selector", func(t *testing.T) {
		f := newCLIProxyFixture(t)
		out, diag, status := f.invoke([]string{"+wrong", "build"}, "")
		if status != 96 || len(out) != 0 || string(diag) != "wrong Cargo toolchain selector\n" {
			t.Fatalf("selector = %d/%q/%q", status, out, diag)
		}
		if _, err := os.Stat(filepath.Join(f.probe.Directory, "native.args")); !os.IsNotExist(err) {
			t.Fatalf("rejected selector invoked Cargo: %v", err)
		}
	})
	for _, cargoStatus := range []int{0, 7} {
		t.Run(fmt.Sprintf("observer_failure_cargo_%d", cargoStatus), func(t *testing.T) {
			f := newCLIProxyFixture(t)
			if err := os.Remove(f.artifact); err != nil {
				t.Fatal(err)
			}
			_, diag, status := f.invoke([]string{"build"}, "", fmt.Sprintf("L_PROXY_STATUS=%d", cargoStatus))
			want := cargoStatus
			if want == 0 {
				want = 96
			}
			if status != want || !bytes.Contains(diag, []byte("no such file")) {
				t.Fatalf("observer failure = %d/%q, want %d", status, diag, want)
			}
		})
	}
	t.Run("metadata_failure_retained", func(t *testing.T) {
		f := newCLIProxyFixture(t)
		_, _, status := f.invoke([]string{"build"}, "", "L_PROXY_METADATA_ERROR=yes")
		if status != 0 {
			t.Fatalf("changed existing metadata-failure status policy: %d", status)
		}
		paths, callErr := filepath.Glob(filepath.Join(f.probe.Directory, "cargo.*.metadata.stderr"))
		if callErr != nil {
			t.Fatal(callErr)
		}
		if len(paths) != 1 {
			t.Fatalf("metadata diagnostics = %v", paths)
		}
		data, callErr := os.ReadFile(paths[0])
		if callErr != nil {
			t.Fatal(callErr)
		}
		if string(data) != "metadata failure" {
			t.Fatalf("metadata diagnostic %q", data)
		}
	})
}

func TestCLIProxyConcurrentConfigIsolationAndFreshEvidence(t *testing.T) {
	t.Setenv("L_HARNESS_TIMING_DIR", "")
	a, b := newCLIProxyFixture(t), newCLIProxyFixture(t)
	fixtures := []cliProxyFixture{a, b, a}
	type result struct {
		out, diag []byte
		status    int
	}
	results := make([]result, len(fixtures))
	var wg sync.WaitGroup
	for i, f := range fixtures {
		wg.Add(1)
		go func(i int, f cliProxyFixture) {
			defer wg.Done()
			results[i].out, results[i].diag, results[i].status = f.invoke([]string{"+" + Version, "build"}, fmt.Sprintf("invocation-%d\n", i))
		}(i, f)
	}
	wg.Wait()
	for i, result := range results {
		if result.status != 0 || string(result.out) != fixtures[i].output+fmt.Sprintf("invocation-%d\n", i) {
			t.Fatalf("invocation %d = %+v", i, result)
		}
	}
	for i, f := range []cliProxyFixture{a, b} {
		paths, callErr := filepath.Glob(filepath.Join(f.probe.Directory, "cargo.*.jsonl"))
		if callErr != nil {
			t.Fatal(callErr)
		}
		want := 2 - i
		if len(paths) != want {
			t.Fatalf("probe %d records = %v, want %d", i, paths, want)
		}
		for _, path := range paths {
			data, callErr := os.ReadFile(path)
			if callErr != nil {
				t.Fatal(callErr)
			}
			if !bytes.HasPrefix(data, []byte(f.output)) {
				t.Fatalf("cross-probe evidence: %s", data)
			}
		}
	}
	old := map[string][]byte{}
	for _, path := range a.probe.artifacts(t) {
		data, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		old[path] = data
	}
	a.env = a.probe.Env(a.env)
	if err := os.WriteFile(a.artifact, []byte("second actual invocation artifact"), 0700); err != nil {
		t.Fatal(err)
	}
	_, _, status := a.invoke([]string{"build"}, "")
	if status != 0 || len(a.probe.freshArtifacts(t)) != 1 {
		t.Fatalf("fresh invocation = %d", status)
	}
	for path, before := range old {
		after, err := os.ReadFile(path)
		if err != nil || !bytes.Equal(before, after) {
			t.Fatalf("previous artifact snapshot mutated: %s %v", path, err)
		}
	}
}

func TestCLIProxyPreservesCargoFailureWhenMetadataFilesCannotOpen(t *testing.T) {
	t.Setenv("L_HARNESS_TIMING_DIR", "")
	for _, suffix := range []string{"metadata.json", "metadata.stderr"} {
		for _, cargoStatus := range []int{0, 7, 143} {
			t.Run(fmt.Sprintf("%s_cargo_%d", suffix, cargoStatus), func(t *testing.T) {
				f := newCLIProxyFixture(t)
				extra := []string{"L_PROXY_METADATA_DIRECTORY=" + suffix}
				if cargoStatus == 143 {
					extra = append(extra, "L_PROXY_SIGNAL=yes")
				} else {
					extra = append(extra, fmt.Sprintf("L_PROXY_STATUS=%d", cargoStatus))
				}
				out, diag, status := f.invoke([]string{"build"}, "", extra...)
				want := cargoStatus
				if want == 0 {
					want = 96
				}
				if status != want || string(out) != f.output || !bytes.HasPrefix(diag, []byte("native stderr\x00tail\n")) || !bytes.Contains(diag, []byte(suffix)) {
					t.Fatalf("metadata-open failure = %d/%q/%q, want status %d", status, out, diag, want)
				}
			})
		}
	}
}

func TestCLIProxyPreservesCargoLaunchFailureStatus(t *testing.T) {
	t.Setenv("L_HARNESS_TIMING_DIR", "")
	for _, status := range []int{127, 126} {
		t.Run(fmt.Sprint(status), func(t *testing.T) {
			f := newCLIProxyFixture(t)
			cargo := filepath.Join(t.TempDir(), "unavailable-cargo")
			if status == 126 {
				if err := os.WriteFile(cargo, []byte("#!/bin/sh\nexit 0\n"), 0600); err != nil {
					t.Fatal(err)
				}
			}
			config, callErr := json.Marshal(map[string]string{"Directory": f.probe.Directory, "Cargo": cargo})
			if callErr != nil {
				t.Fatal(callErr)
			}
			if err := os.WriteFile(filepath.Join(f.probe.Directory, "cargo-proxy.json"), config, 0600); err != nil {
				t.Fatal(err)
			}
			out, diag, got := f.invoke([]string{"build"}, "")
			if got != status || len(out) != 0 || !bytes.Contains(diag, []byte(cargo)) {
				t.Fatalf("Cargo launch failure = %d/%q/%q, want %d", got, out, diag, status)
			}
		})
	}
}
