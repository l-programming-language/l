//go:build darwin || linux

package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/l-programming-language/l/internal/cargocache"
	nativetest "github.com/l-programming-language/l/internal/testsupport/native"
)

func TestOwnedRuntimeCachePreservesABAInputs(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()
	cache, target := t.TempDir(), t.TempDir()
	const support = "support/runtime"
	supportManifest := "[package]\nname = \"l_runtime_support\"\nversion = \"0.0.0\"\nedition = \"2024\"\nautobins = false\n[lib]\nname = \"l_runtime_support\"\npath = \"src/lib.rs\"\n"
	library, err := readCargoPackage([]byte(supportManifest))
	if err != nil {
		t.Fatal(err)
	}
	var retained []struct{ executable, output string }
	var firstInputs string
	for index, step := range []struct {
		graph, value, marker string
	}{{"A", "one", "first"}, {"A", "one", "changed-entry"}, {"B", "one", "second"}, {"A", "one", "first"}, {"B", "two", "changed-runtime"}, {"A", "one", "restored"}} {
		source := t.TempDir()
		member := "entry_" + strings.ToLower(step.graph)
		entryManifest := fmt.Sprintf("[package]\nname = %q\nversion = \"0.0.0\"\nedition = \"2024\"\n[dependencies]\nl_runtime_support = { path = \"../support/runtime\" }\n", "runtime_cache_"+member)
		entry, err := readCargoPackage([]byte(entryManifest))
		if err != nil {
			t.Fatal(err)
		}
		lock, err := cargoWorkspaceLock([]*cargoPackage{entry, library})
		if err != nil {
			t.Fatal(err)
		}
		files := map[string][]byte{
			"Cargo.toml":            []byte(fmt.Sprintf("[workspace]\nresolver = \"3\"\nmembers = [%q, \"support/runtime\"]\n", member)),
			"Cargo.lock":            lock,
			member + "/Cargo.toml":  []byte(entryManifest),
			member + "/src/main.rs": []byte(fmt.Sprintf("fn main() { println!(\"%s:{}:{}\", l_runtime_support::value(), l_runtime_support::origin()); }\n", step.marker)),
			support + "/Cargo.toml": []byte(supportManifest),
			support + "/src/lib.rs": []byte(fmt.Sprintf("pub fn value() -> &'static str { %q }\npub fn origin() -> &'static str { env!(\"CARGO_MANIFEST_DIR\") }\n", step.value)),
		}
		for name, data := range files {
			path := filepath.Join(source, name)
			if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(path, data, 0600); err != nil {
				t.Fatal(err)
			}
		}
		inputs, err := cargocache.StageCurrent(source, cache, "")
		if err != nil {
			t.Fatal(err)
		}
		if firstInputs == "" {
			firstInputs = inputs
		} else if inputs != firstInputs {
			t.Fatalf("graph changed canonical runtime path: %q != %q", inputs, firstInputs)
		}
		other := "entry_a"
		if member == other {
			other = "entry_b"
		}
		if _, err := os.Lstat(filepath.Join(inputs, other)); !os.IsNotExist(err) {
			t.Fatalf("unexpected native member in current graph: %s: %v", other, err)
		}
		if err := cargocache.Activate(inputs, target); err != nil {
			t.Fatal(err)
		}
		command := exec.CommandContext(ctx, "cargo", "+"+nativeToolchainVersion, "build", "--workspace", "--frozen", "--message-format=json", "--manifest-path", filepath.Join(inputs, "Cargo.toml"), "--target", "host-tuple", "--target-dir", target)
		command.Env = cargoBuildEnvironment()
		var stdout, stderr bytes.Buffer
		command.Stdout, command.Stderr = &stdout, &stderr
		if err := command.Run(); err != nil {
			t.Fatalf("build %d: %v\n%s\n%s", index, err, stdout.Bytes(), stderr.Bytes())
		}
		artifacts, failed := cargoMessages(stdout.Bytes())
		if failed || len(artifacts) != 1 {
			t.Fatalf("native build %d artifacts = %v", index, artifacts)
		}
		found, fresh := false, false
		for _, line := range bytes.Split(stdout.Bytes(), []byte{'\n'}) {
			var event struct {
				Reason string
				Fresh  bool
				Target struct {
					Name string
					Kind []string
				}
			}
			if json.Unmarshal(line, &event) == nil && event.Reason == "compiler-artifact" && event.Target.Name == "l_runtime_support" {
				found, fresh = true, event.Fresh
			}
		}
		if !found {
			t.Fatalf("build %d omitted the actual runtime library artifact", index)
		}
		t.Logf("actual runtime library build %d: Cargo fresh=%v", index, fresh)
		runtimeExecutable := filepath.Join(t.TempDir(), "entry")
		if err := copyCargoArtifact(artifacts[0].Executable, runtimeExecutable); err != nil {
			t.Fatal(err)
		}
		nativetest.AssertNativeArtifact(t, runtimeExecutable)
		output, err := exec.CommandContext(ctx, runtimeExecutable).CombinedOutput()
		// Cargo preserves the spelling of the explicitly selected manifest
		// directory in this macro, including macOS's /var path alias.
		origin := filepath.Join(inputs, support)
		want := step.marker + ":" + step.value + ":" + origin + "\n"
		if err != nil || string(output) != want {
			t.Fatalf("build %d used stale source/runtime: %q %v", index, output, err)
		}
		retained = append(retained, struct{ executable, output string }{runtimeExecutable, want})
	}
	for index, snapshot := range retained {
		output, err := exec.CommandContext(ctx, snapshot.executable).CombinedOutput()
		if err != nil || string(output) != snapshot.output {
			t.Fatalf("earlier snapshot %d changed after later builds: %q %v", index, output, err)
		}
	}
}
