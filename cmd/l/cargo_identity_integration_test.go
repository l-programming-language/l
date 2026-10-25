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
	"github.com/l-programming-language/l/internal/testsupport/phase1"
	"github.com/l-programming-language/l/internal/timing"
)

func TestCLIReservedCompanionNames(t *testing.T) {
	parallelIndependentBuildRoot(t)
	root := distributiontest.RepoRoot(t)
	cli := phase1.CLI(t)
	// Compose flat, single-Service structural and multi-Service Apps in one
	// program. Later declarations occupy the first two companion fallbacks.
	source := "Use IO's Print from `io` as \"Print\".\n\n" +
		"\"Left\" is a New Adjective describing any Service, Object, or Function.\n" +
		"It describes a Generic `left` Function of any Process with No Inputs and No Outputs.\n" +
		"It describes a Generic `left` Object with a New Objective Left Function using any Process.\n" +
		"It describes a Generic `left` Service with a New Left Object using any Process.\n" +
		"\"Right\" is a New Adjective describing any Service, Object, or Function.\n" +
		"It describes a Generic `right` Function of any Process with No Inputs and No Outputs.\n" +
		"It describes a Generic `right` Object with a New Objective Right Function using any Process.\n" +
		"It describes a Generic `right` Service with a New Right Object using any Process.\n" +
		"\"Grouped\" describes a Generic App with a New Main Service, New Left Service, and New Right Service using any Process.\n\n" +
		"\"Leaf\" describes a Generic `leaf` Class with a New Main Object using any Process.\n" +
		"\"Middle\" describes a Generic `middle` Class with a New Leaf Class using any Process.\n" +
		"\"Outer\" describes a Generic `outer` Class with a New Middle Class using any Process.\n" +
		"\"Branched\" describes a Generic `main` Service with a New Outer Class using any Process.\n" +
		"\"Tree\" describes a Generic App with a New Branched Service using any Process.\n\n"
	ids := []string{"build", "deps", "examples", "incremental", "l_app", "L_APP_1", "l_entry"}
	kinds := []string{"Basic", "Grouped", "Tree", "Basic", "Basic", "Basic", "Basic"}
	aliases := []string{"l_app_2", "l_app_3", "l_app_4", "l_app_5", "l_app", "L_APP_1", "l_entry"}
	for i, id := range ids {
		source += fmt.Sprintf("Create %q as a New %s App.\nThe App Prints `%s`.\nExecute %s.\nShow %s's Output.\n\n", id, kinds[i], id, id, id)
	}
	project := t.TempDir()
	copyBuildProject(t, root, project, "l_app_2.l", source)
	project, err := filepath.EvalSymlinks(project)
	if err != nil {
		t.Fatal(err)
	}
	recorder := nativetest.CLIProbe(t)
	stdout, stderr, err := runBuildCommand(t, cli, project, recorder.Env(nil), "build", "l_app_2.l")
	if err != nil {
		t.Fatalf("reserved companion build: %v\nstdout: %s\nstderr: %s", err, stdout, stderr)
	}
	entry := filepath.Join(project, "l_app_2")
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
	wantTargets := map[string]string{"orchestrator": "l_entry_1"}
	for i, alias := range aliases {
		wantTargets[fmt.Sprintf("apps/app-%03d", i)] = alias
	}
	for _, member := range metadata.Packages {
		for path, name := range wantTargets {
			if !strings.HasSuffix(filepath.ToSlash(member.ManifestPath), "/"+path+"/Cargo.toml") {
				continue
			}
			if member.Name != name || len(member.Targets) != 1 || member.Targets[0].Name != name || len(member.Targets[0].Kind) != 1 || member.Targets[0].Kind[0] != "bin" {
				t.Fatalf("%s Cargo identity = %+v; want package and binary %q", path, member, name)
			}
			delete(wantTargets, path)
		}
	}
	if len(wantTargets) != 0 {
		t.Fatalf("missing Cargo targets: %v", wantTargets)
	}
	var generation generationManifest
	nativetest.ReadPublishedGeneration(t, entry, &generation)
	if len(generation.Apps) != len(ids) {
		t.Fatalf("published companions = %+v; want %v", generation.Apps, ids)
	}
	for i, app := range generation.Apps {
		if app.ID != ids[i] || app.Executable != fmt.Sprintf("app-%03d", i) {
			t.Errorf("companion %d = %+v; want canonical ID %q and ordinal executable", i, app, ids[i])
		}
	}
	copied, err := os.ReadFile(filepath.Join(project, "l_app_2.l"))
	if err != nil || !bytes.Equal(copied, []byte(source)) {
		t.Fatalf("input source changed: %v", err)
	}
	relocated := filepath.Join(t.TempDir(), "l_app_2")
	for _, suffix := range []string{"", ".apps"} {
		if err := os.Rename(entry+suffix, relocated+suffix); err != nil {
			t.Fatal(err)
		}
	}
	if err := os.RemoveAll(project); err != nil {
		t.Fatal(err)
	}
	// Runtime needs neither the source tree nor any compiler on PATH.
	stdout, stderr, err = runRuntimeCommand(t, relocated, t.TempDir(), []string{"PATH=" + t.TempDir()})
	if err != nil || string(stdout) != strings.Join(ids, "\n")+"\n" || len(stderr) != 0 {
		t.Fatalf("relocated entry: %v stdout=%q stderr=%q", err, stdout, stderr)
	}
}

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
				"orchestrator/Cargo.toml":         []byte(entryManifest),
				"orchestrator/src/main.rs":        []byte(entrySource),
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
			if err := compileCargoWorkspace(ctx, nil, entry, []*sourceBundle{library}, nil, workspace, nil, executable); err != nil {
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
	cli := phase1.CLI(t)
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
				if !strings.HasSuffix(filepath.ToSlash(member.ManifestPath), "/orchestrator/Cargo.toml") {
					continue
				}
				entries++
				if member.Name != "l_entry" || len(member.Targets) != 1 || member.Targets[0].Name != "l_entry" || len(member.Targets[0].Kind) != 1 || member.Targets[0].Kind[0] != "bin" {
					t.Fatalf("reserved source did not select the safe native entry target: %+v", member)
				}
			}
			if entries != 1 {
				t.Fatalf("Cargo has %d orchestrator packages; want one", entries)
			}
			var generation generationManifest
			nativetest.ReadPublishedGeneration(t, entry, &generation)
			if len(generation.Apps) != 1 || generation.Apps[0].ID != "Hello" {
				t.Fatalf("published companion Apps = %+v; want Hello", generation.Apps)
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
