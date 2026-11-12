//go:build darwin || linux

package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	distributiontest "github.com/l-programming-language/l/internal/testsupport/distribution"
	nativetest "github.com/l-programming-language/l/internal/testsupport/native"
	"github.com/l-programming-language/l/internal/timing"
)

func TestCargoWorkspacePackageIdentityNative(t *testing.T) {
	const dependencies = "tokio = { version = \"=1.53.1\", features = [\"rt-multi-thread\", \"macros\", \"process\", \"io-util\", \"io-std\", \"signal\", \"sync\", \"time\"] }\nnix = { version = \"=0.31.3\", features = [\"signal\", \"process\"] }\n"
	for _, tc := range []struct {
		name, packageName, version, alias string
		edges                             []string
	}{
		{"direct", "tokio", "0.0.0", "local_tokio", []string{"tokio 0.0.0", "tokio 1.53.1"}},
		{"transitive", "bytes", "0.0.0", "local_bytes", []string{"bytes 0.0.0", "bytes 1.12.1"}},
		{"same version different source", "tokio", "1.53.1", "local_tokio", []string{"tokio 1.53.1", "tokio 1.53.1 (registry+https://github.com/rust-lang/crates.io-index)"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			entryManifest := "[package]\nname = \"identity_entry\"\nversion = \"0.0.0\"\nedition = \"2024\"\n\n[dependencies]\n" + dependencies + fmt.Sprintf("%s = { package = %q, path = \"../services/service-000\" }\n", tc.alias, tc.packageName)
			entrySource := fmt.Sprintf("fn main() { let _runtime = tokio::runtime::Builder::new_current_thread().enable_all().build().unwrap(); let _pid = nix::unistd::getpid(); let _: u8 = %s::local_marker(); }\n", tc.alias)
			libraryManifest := fmt.Sprintf("[package]\nname = %q\nversion = %q\nedition = \"2024\"\nautobins = false\n\n[lib]\npath = \"src/lib.rs\"\n", tc.packageName, tc.version)
			const librarySource = "pub fn local_marker() -> u8 { 7 }\n"
			// Retain independently authored inputs before planning or compilation.
			inputs := map[string][]byte{
				"entry/Cargo.toml":                []byte(entryManifest),
				"entry/src/main.rs":               []byte(entrySource),
				"services/service-000/Cargo.toml": []byte(libraryManifest),
				"services/service-000/src/lib.rs": []byte(librarySource),
			}
			entry, err := planSourceBundle([][]byte{[]byte(entryManifest), []byte(entrySource)}, []string{"Cargo.toml", "src/main.rs"})
			if err != nil {
				t.Fatal(err)
			}
			library, err := planSourceBundle([][]byte{[]byte(libraryManifest), []byte(librarySource)}, []string{"Cargo.toml", "src/lib.rs"})
			if err != nil {
				t.Fatal(err)
			}
			lock, err := cargoWorkspaceLock([]*cargoPackage{entry.packagePlan, library.packagePlan})
			if err != nil {
				t.Fatal(err)
			}
			inputs["Cargo.lock"] = bytes.Clone(lock)

			// These subtests are intentionally serial: the observer's environment
			// belongs to this in-process production compiler invocation.
			recorder := nativetest.CLIProbe(t)
			for _, item := range recorder.Env(nil) {
				key, value, _ := strings.Cut(item, "=")
				if current, present := os.LookupEnv(key); !present || current != value {
					t.Setenv(key, value)
				}
			}
			directory := t.TempDir()
			workspace, executable := filepath.Join(directory, "workspace"), filepath.Join(directory, "identity_entry")
			ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
			defer cancel()
			if err := compileCargoWorkspace(ctx, nil, entry, []*sourceBundle{library}, workspace, nil, executable); err != nil {
				t.Fatalf("native identity build: %v", err)
			}
			recorder.AssertPublished(t, executable)
			nativetest.AssertNativeArtifact(t, executable)
			observed, err := filepath.Glob(filepath.Join(recorder.Directory, "cargo.*.workspace"))
			if err != nil || len(observed) != 1 {
				t.Fatalf("Cargo workspace evidence = %v, %v", observed, err)
			}
			for _, directory := range []string{workspace, observed[0]} {
				for name, want := range inputs {
					got, err := os.ReadFile(filepath.Join(directory, filepath.FromSlash(name)))
					if err != nil || !bytes.Equal(got, want) {
						t.Errorf("native build changed %s/%s: %v", directory, name, err)
					}
				}
			}
			for _, edge := range tc.edges {
				if !strings.Contains(string(lock), fmt.Sprintf(" %q,\n", edge)) {
					t.Errorf("lock lacks literal dependency edge %q", edge)
				}
			}
			for _, name := range []string{tc.packageName, tc.alias} {
				if strings.Contains(string(lock), fmt.Sprintf(" %q,\n", name)) {
					t.Errorf("lock contains ambiguous package name or dependency alias %q", name)
				}
			}
		})
	}
}

func TestCLIReservedSourceNames(t *testing.T) {
	t.Parallel()
	root := distributiontest.RepoRoot(t)
	cli := nativetest.InstalledCLI(t)
	fixture, err := os.ReadFile(filepath.Join(root, "demo/basic/main.l"))
	if err != nil {
		t.Fatal(err)
	}
	for _, stem := range []string{"build", "deps"} {
		stem := stem
		t.Run(stem, func(t *testing.T) {
			t.Parallel()
			admitted := timing.Start("admission.wait", t.Name())
			buildCLICases <- struct{}{}
			admitted(nil)
			t.Cleanup(func() { <-buildCLICases })
			project := t.TempDir()
			copyBuildProject(t, root, project, stem+".l", "")
			project, err := filepath.EvalSymlinks(project)
			if err != nil {
				t.Fatal(err)
			}
			recorder := nativetest.CLIProbe(t)
			stdout, stderr, err := runBuildCommand(t, cli, project, recorder.Env(nil), "build", stem+".l")
			if err != nil {
				t.Fatalf("l build %s.l: %v\nstdout: %s\nstderr: %s", stem, err, stdout, stderr)
			}
			entry := filepath.Join(project, stem)
			if string(stdout) != entry+"\n" || len(stderr) != 0 {
				t.Fatalf("publication stdout=%q stderr=%q; want %q and no diagnostics", stdout, stderr, entry+"\n")
			}
			recorder.AssertPublished(t, entry)
			metadataFiles, err := filepath.Glob(filepath.Join(recorder.Directory, "cargo.*.metadata.json"))
			if err != nil || len(metadataFiles) != 1 {
				t.Fatalf("Cargo metadata = %v, %v", metadataFiles, err)
			}
			data, err := os.ReadFile(metadataFiles[0])
			if err != nil {
				t.Fatal(err)
			}
			var metadata struct {
				Packages []struct {
					Name         string
					ManifestPath string `json:"manifest_path"`
					Targets      []struct {
						Name string
						Kind []string
					}
				}
			}
			if err := json.Unmarshal(data, &metadata); err != nil {
				t.Fatal(err)
			}
			entries := 0
			for _, member := range metadata.Packages {
				if !strings.HasSuffix(filepath.ToSlash(member.ManifestPath), "/entry/Cargo.toml") {
					continue
				}
				entries++
				if member.Name != "Hello" || len(member.Targets) != 1 || member.Targets[0].Name != "Hello" || len(member.Targets[0].Kind) != 1 || member.Targets[0].Kind[0] != "bin" {
					t.Fatalf("reserved source filename changed the App target identity: %+v", member)
				}
			}
			if entries != 1 {
				t.Fatalf("Cargo has %d entry packages; want one", entries)
			}
			var generation generationManifest
			nativetest.ReadPublishedGeneration(t, entry, &generation)
			if len(generation.Apps) != 1 || generation.Apps[0].ID != "Hello" {
				t.Fatalf("published App records = %+v; want Hello", generation.Apps)
			}
			copied, err := os.ReadFile(filepath.Join(project, stem+".l"))
			if err != nil || !bytes.Equal(copied, fixture) {
				t.Fatalf("renamed immutable fixture changed: %v", err)
			}
			stdout, stderr, err = runRuntimeCommand(t, entry, t.TempDir(), nil)
			if err != nil || string(stdout) != "Hello World!\n" || len(stderr) != 0 {
				t.Fatalf("published entry: %v stdout=%q stderr=%q", err, stdout, stderr)
			}
		})
	}
}
