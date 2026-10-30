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
)

func TestCargoWorkspaceBuildsLibraryAndBinaries(t *testing.T) {
	makeBundle := func(name, kind, source, dependencies string) *sourceBundle {
		t.Helper()
		manifest := fmt.Sprintf("[package]\nname = %q\nversion = \"0.0.0\"\nedition = \"2024\"\n", name)
		file := "src/main.rs"
		if kind == "lib" {
			manifest += "\n[lib]\npath = \"src/lib.rs\"\n"
			file = "src/lib.rs"
		}
		manifest += dependencies
		bundle, err := planSourceBundle([][]byte{[]byte(manifest), []byte(source)}, []string{"Cargo.toml", file})
		if err != nil {
			t.Fatal(err)
		}
		return bundle
	}
	app := makeBundle("application", "bin", "fn main() { assert_eq!(service::value(), 7); }\n", "\n[dependencies]\nservice = { path = \"../../services/service-000\" }\n")
	service := makeBundle("service", "lib", "pub fn value() -> u8 { 7 }\n", "")
	entry := makeBundle("entry", "bin", "fn main() {}\n", "")
	directory := t.TempDir()
	workspace := filepath.Join(directory, "workspace")
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()
	if err := compileCargoWorkspace(ctx, []*sourceBundle{app}, entry, []*sourceBundle{service}, nil, workspace, []string{filepath.Join(directory, "app")}, filepath.Join(directory, "entry")); err != nil {
		t.Fatal(err)
	}
	for _, member := range []struct {
		prefix string
		bundle *sourceBundle
	}{{"apps/app-000", app}, {"orchestrator", entry}, {"services/service-000", service}} {
		for _, file := range member.bundle.files {
			got, err := os.ReadFile(filepath.Join(workspace, member.prefix, file.name))
			if err != nil || !bytes.Equal(got, file.content) {
				t.Fatalf("workspace changed %s/%s: %v", member.prefix, file.name, err)
			}
		}
	}
	for _, name := range []string{"app", "entry"} {
		info, err := os.Stat(filepath.Join(directory, name))
		if err != nil || !info.Mode().IsRegular() || info.Mode().Perm()&0111 == 0 {
			t.Fatalf("missing Cargo artifact %s: %v", name, err)
		}
	}
}

func TestCargoWorkspacePinnedRuntimeDependencies(t *testing.T) {
	manifest := "[package]\nname = \"entry\"\nversion = \"0.0.0\"\nedition = \"2024\"\n\n[dependencies]\ntokio = { version = \"=1.53.1\", features = [\"rt-multi-thread\", \"macros\", \"process\", \"io-util\", \"io-std\", \"signal\", \"sync\", \"time\"] }\nnix = { version = \"=0.31.3\", features = [\"signal\", \"process\"] }\n"
	source := "fn main() { let _runtime = tokio::runtime::Builder::new_current_thread().enable_all().build().unwrap(); let _pid = nix::unistd::getpid(); }\n"
	bundle, err := planSourceBundle([][]byte{[]byte(manifest), []byte(source)}, []string{"Cargo.toml", "src/main.rs"})
	if err != nil {
		t.Fatal(err)
	}
	directory := t.TempDir()
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()
	if err := compileCargoWorkspace(ctx, nil, bundle, nil, nil, filepath.Join(directory, "workspace"), nil, filepath.Join(directory, "entry")); err != nil {
		t.Fatal(err)
	}
	target, release, err := cargoBuildCache(ctx, filepath.Join(directory, "workspace"))
	if err != nil {
		t.Fatal(err)
	}
	dependencies := map[string]time.Time{}
	err = filepath.WalkDir(target, func(path string, entry os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if strings.HasPrefix(entry.Name(), "libtokio-") && strings.HasSuffix(entry.Name(), ".rlib") {
			info, err := entry.Info()
			if err != nil {
				return err
			}
			dependencies[path] = info.ModTime()
		}
		return nil
	})
	release()
	if err != nil || len(dependencies) == 0 {
		t.Fatalf("no retained Tokio dependency artifacts: %v", err)
	}
	if err := compileCargoWorkspace(ctx, nil, bundle, nil, nil, filepath.Join(directory, "workspace-again"), nil, filepath.Join(directory, "entry-again")); err != nil {
		t.Fatal(err)
	}
	for path, modified := range dependencies {
		info, err := os.Stat(path)
		if err != nil || !info.ModTime().Equal(modified) {
			t.Fatalf("Cargo rebuilt a cached dependency %s: %v", path, err)
		}
	}
}

func TestRuntimePackageAvoidsNativeNamesAndAliases(t *testing.T) {
	packages := []*sourceBundle{{packagePlan: &cargoPackage{name: "l-runtime-support", dependencies: []cargoDependency{{name: "l_runtime_support_1"}}}}}
	if got := runtimePackageName("l_runtime_support_2", packages); got != "l_runtime_support_3" {
		t.Fatalf("runtime package collides with native crate or alias: %q", got)
	}
}

// The runtime is a real Cargo library linked into the published native entry,
// not another installed executable or a replacement source observer.
func TestCLICompilesRuntimeAsWorkspaceLibrary(t *testing.T) {
	t.Parallel()
	root := distributiontest.RepoRoot(t)
	project := t.TempDir()
	copyBuildProject(t, root, project, "main.l", "")
	recorder := nativetest.CLIProbe(t)
	out, diagnostics, err := runBuildCommand(t, phase1.CLI(t), project, recorder.Env(nil), "build", "main.l")
	if err != nil || len(diagnostics) != 0 {
		t.Fatalf("build: %v\n%s\n%s", err, out, diagnostics)
	}
	entry := filepath.Join(project, "main")
	recorder.AssertPublished(t, entry)
	files, err := filepath.Glob(filepath.Join(recorder.Directory, "cargo.*.metadata.json"))
	if err != nil || len(files) != 1 {
		t.Fatalf("native workspace metadata: %v %v", files, err)
	}
	data, err := os.ReadFile(files[0])
	if err != nil {
		t.Fatal(err)
	}
	var metadata struct {
		Packages []struct {
			Name         string
			ManifestPath string `json:"manifest_path"`
			Targets      []struct{ Kind []string }
			Dependencies []struct {
				Name string
				Path *string
			}
		}
	}
	if err := json.Unmarshal(data, &metadata); err != nil {
		t.Fatal(err)
	}
	found, linked := false, false
	for _, p := range metadata.Packages {
		if p.Name == "l_runtime_support" {
			if len(p.Targets) != 1 || len(p.Targets[0].Kind) != 1 || p.Targets[0].Kind[0] != "lib" || !strings.HasSuffix(filepath.ToSlash(p.ManifestPath), "/support/runtime/Cargo.toml") {
				t.Fatalf("runtime is not the compiler-owned library target: %+v", p)
			}
			found = true
		}
		if p.Name == "main" {
			for _, dependency := range p.Dependencies {
				if dependency.Name == "l_runtime_support" && dependency.Path != nil && strings.HasSuffix(filepath.ToSlash(*dependency.Path), "/support/runtime") {
					linked = true
				}
			}
		}
	}
	if !found || !linked {
		t.Fatalf("Cargo did not build and link the dedicated runtime library: library=%v dependency=%v", found, linked)
	}
	workspace := strings.TrimSuffix(files[0], ".metadata.json") + ".workspace"
	const libraryRoot = "#![forbid(unsafe_code)]\n\npub mod runtime;\n#[cfg(any(target_os = \"linux\", target_os = \"macos\"))]\n#[path = \"process_unix.rs\"]\nmod process;\n#[cfg(not(any(target_os = \"linux\", target_os = \"macos\")))]\n#[path = \"process_other.rs\"]\nmod process;\n"
	actual, err := os.ReadFile(filepath.Join(workspace, "support/runtime/src/lib.rs"))
	if err != nil || string(actual) != libraryRoot {
		t.Fatalf("runtime library root = %q, %v", actual, err)
	}
	for name, asset := range map[string]string{"runtime/host.rs": "host", "process_unix.rs": "process_unix", "process_other.rs": "process_other"} {
		want, err := os.ReadFile(filepath.Join(root, "builtins/templates/runtime", asset+".lu"))
		if err != nil {
			t.Fatal(err)
		}
		got, err := os.ReadFile(filepath.Join(workspace, "support/runtime/src", name))
		if err != nil || !bytes.Equal(got, want) {
			t.Fatalf("shared runtime changed selected %s bytes: %v", name, err)
		}
		if _, err := os.Stat(filepath.Join(workspace, "orchestrator/src", name)); !os.IsNotExist(err) {
			t.Fatalf("orchestrator retained inline runtime %s: %v", name, err)
		}
	}
	if _, err := os.Stat(filepath.Join(workspace, "orchestrator/src/runtime.rs")); !os.IsNotExist(err) {
		t.Fatalf("orchestrator retained inline runtime: %v", err)
	}
	out, diagnostics, err = runRuntimeCommand(t, entry, project, nil)
	if err != nil || string(out) != "Hello World!\n" || len(diagnostics) != 0 {
		t.Fatalf("standalone entry: %q/%q %v", out, diagnostics, err)
	}
}
