//go:build darwin || linux

package main

import (
	"bufio"
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"runtime"
	"sort"
	"strings"
	"syscall"
	"testing"
	"time"

	distributiontest "github.com/l-programming-language/l/internal/testsupport/distribution"
	nativetest "github.com/l-programming-language/l/internal/testsupport/native"
	"github.com/l-programming-language/l/internal/testsupport/phase1"
)

type runnableBuildResult struct {
	project, binary, manifestPath, workspace string
	manifest                                 runnableManifest
	recorder                                 *nativetest.CLIRecorder
}

func runnableBuild(t *testing.T, source string, apps []string) runnableBuildResult {
	t.Helper()
	project, _ := runnableProject(t, source)
	return runnableBuildProject(t, project, apps)
}

func runnableBuildProject(t *testing.T, project string, apps []string) runnableBuildResult {
	t.Helper()
	before, err := filepath.Glob(filepath.Join(project, "main.apps", "generations", "*", "manifest.json"))
	if err != nil {
		t.Fatal(err)
	}
	recorder := nativetest.CLIProbe(t)
	out, diagnostics, err := runBuildCommand(t, phase1.CLI(t), project, runnableRecorderEnv(recorder), "build", "main.l")
	if err != nil {
		t.Fatalf("App installed-CLI build prerequisite; native generation/publication/runtime not reached: %v\nstdout=%s\nstderr=%s", err, out, diagnostics)
	}
	binary := filepath.Join(project, "main")
	if string(out) != binary+"\n" || len(diagnostics) != 0 {
		t.Fatalf("CLI streams = %q / %q; want only published path", out, diagnostics)
	}
	recorder.Workspace(t)
	nativetest.AssertNativeArtifact(t, binary)
	after, err := filepath.Glob(filepath.Join(project, "main.apps", "generations", "*", "manifest.json"))
	if err != nil {
		t.Fatal(err)
	}
	prior := make(map[string]bool)
	for _, name := range before {
		prior[name] = true
	}
	var fresh []string
	for _, name := range after {
		if !prior[name] {
			fresh = append(fresh, name)
		}
	}
	if len(fresh) != 1 {
		t.Fatalf("build must publish exactly one fresh immutable generation: %v", fresh)
	}
	result := runnableBuildResult{project: project, binary: binary, manifestPath: fresh[0], recorder: recorder}
	data, err := os.ReadFile(result.manifestPath)
	if err != nil {
		t.Fatal(err)
	}
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&result.manifest); err != nil {
		t.Fatalf("App generation schema: %v", err)
	}
	if result.manifest.Version != 1 || result.manifest.Entrypoint != "main" || result.manifest.GOOS != runtime.GOOS || result.manifest.GOARCH != runtime.GOARCH {
		t.Fatalf("App publication requires a v1 native manifest: %+v", result.manifest)
	}
	if len(result.manifest.Apps) != len(apps) {
		t.Fatalf("manifest App count=%d want=%d", len(result.manifest.Apps), len(apps))
	}
	for i, companion := range result.manifest.Apps {
		if companion.ID != apps[i] || companion.Executable != fmt.Sprintf("app-%03d", i) {
			t.Fatalf("canonical App companion %d=%+v", i, companion)
		}
		path := filepath.Join(filepath.Dir(result.manifestPath), companion.Executable)
		if hash := nativetest.NativeArtifactHash(t, path); hash != companion.SHA256 {
			t.Fatalf("published %s hash=%s want=%s", companion.ID, hash, companion.SHA256)
		}
	}
	metadata, err := filepath.Glob(filepath.Join(recorder.Directory, "cargo.*.metadata.json"))
	if err != nil || len(metadata) != 1 {
		t.Fatalf("one real Cargo workspace expected: %v %v", metadata, err)
	}
	result.workspace = strings.TrimSuffix(metadata[0], ".metadata.json") + ".workspace"
	assertRunnableCargoRoles(t, metadata[0], apps)
	assertRunnableIdentityArtifact(t, recorder.Directory, binary, "orchestrator")
	for i, companion := range result.manifest.Apps {
		assertRunnableIdentityArtifact(t, recorder.Directory, filepath.Join(filepath.Dir(result.manifestPath), companion.Executable), fmt.Sprintf("apps/app-%03d", i))
	}
	return result
}

// Detailed evidence binds immutable artifact bytes to Cargo's package/target
// identity. The plain recorder's hash-only inventory cannot prove this alone.
func runnableRecorderEnv(recorder *nativetest.CLIRecorder) []string {
	var env []string
	for _, entry := range os.Environ() {
		if !strings.HasPrefix(entry, "L_HARNESS_CARGO_EVIDENCE_DIR=") {
			env = append(env, entry)
		}
	}
	return recorder.Env(append(env, "L_HARNESS_CARGO_EVIDENCE_DIR="+filepath.Join(recorder.Directory, "identity-evidence")))
}

type runnableCargoTarget struct {
	Name   string
	Kind   []string
	Source string `json:"src_path"`
}

func assertRunnableIdentityArtifact(t testing.TB, recorder, published, location string) {
	t.Helper()
	metadata, err := filepath.Glob(filepath.Join(recorder, "cargo.*.metadata.json"))
	if err != nil || len(metadata) != 1 {
		t.Fatalf("one identity metadata capture required: %v %v", metadata, err)
	}
	data, err := os.ReadFile(metadata[0])
	if err != nil {
		t.Fatal(err)
	}
	var graph struct {
		Root     string `json:"workspace_root"`
		Packages []struct {
			ID       string
			Manifest string `json:"manifest_path"`
			Targets  []runnableCargoTarget
		}
	}
	if err := json.Unmarshal(data, &graph); err != nil {
		t.Fatal(err)
	}
	var packageID string
	var target runnableCargoTarget
	for _, p := range graph.Packages {
		rel, err := filepath.Rel(graph.Root, p.Manifest)
		if err != nil {
			t.Fatal(err)
		}
		if filepath.ToSlash(filepath.Dir(rel)) == location {
			if packageID != "" || len(p.Targets) != 1 || !reflect.DeepEqual(p.Targets[0].Kind, []string{"bin"}) {
				t.Fatalf("ambiguous native target at %s", location)
			}
			packageID, target = p.ID, p.Targets[0]
		}
	}
	if packageID == "" {
		t.Fatalf("missing native Cargo package at %s", location)
	}
	stream, err := os.Open(strings.TrimSuffix(metadata[0], ".metadata.json") + ".jsonl")
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := stream.Close(); err != nil {
			t.Errorf("close Cargo artifact evidence: %v", err)
		}
	}()
	scan := bufio.NewScanner(stream)
	scan.Buffer(make([]byte, 4096), 8*1024*1024)
	var executable string
	for scan.Scan() {
		var event struct {
			Reason, Executable string
			PackageID          string `json:"package_id"`
			Target             runnableCargoTarget
		}
		if json.Unmarshal(scan.Bytes(), &event) != nil || event.Reason != "compiler-artifact" || event.PackageID != packageID || !reflect.DeepEqual(event.Target, target) || event.Executable == "" {
			continue
		}
		if executable != "" {
			t.Fatalf("duplicate native artifact message for %s", location)
		}
		executable = event.Executable
	}
	if err := scan.Err(); err != nil {
		t.Fatal(err)
	}
	if executable == "" {
		t.Fatalf("no Cargo compiler-artifact for package/target %s/%s", packageID, target.Name)
	}
	want, err := os.ReadFile(published)
	if err != nil {
		t.Fatal(err)
	}
	hash := fmt.Sprintf("%x", sha256.Sum256(want))
	matched := 0
	err = filepath.WalkDir(filepath.Join(recorder, "identity-evidence"), func(path string, entry os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() || entry.Name() != "complete.json" {
			return nil
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		var report struct {
			Complete  bool
			Status    int
			Errors    []string
			Artifacts []struct {
				Source, SHA256 string
				Target         runnableCargoTarget
			}
		}
		if err := json.Unmarshal(data, &report); err != nil {
			return err
		}
		if !report.Complete || report.Status != 0 || len(report.Errors) != 0 {
			return fmt.Errorf("incomplete native identity evidence: %s: %+v", path, report.Errors)
		}
		for _, artifact := range report.Artifacts {
			if artifact.Source != executable || !reflect.DeepEqual(artifact.Target, target) {
				continue
			}
			if artifact.SHA256 != hash {
				return fmt.Errorf("published %s bytes came from the wrong Cargo package/target", location)
			}
			data, err := os.ReadFile(filepath.Join(filepath.Dir(path), "artifacts", hash))
			if err != nil {
				return err
			}
			if !bytes.Equal(data, want) {
				return fmt.Errorf("immutable identity snapshot differs for %s", location)
			}
			matched++
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if matched != 1 {
		t.Fatalf("matched %d immutable snapshots for %s; want exactly one", matched, location)
	}
}

func assertRunnableCargoRoles(t testing.TB, metadata string, apps []string) {
	t.Helper()
	data, err := os.ReadFile(metadata)
	if err != nil {
		t.Fatal(err)
	}
	var graph struct {
		Root     string `json:"workspace_root"`
		Packages []struct {
			Name         string
			ManifestPath string `json:"manifest_path"`
			Targets      []struct {
				Name string
				Kind []string
			}
		}
	}
	if err := json.Unmarshal(data, &graph); err != nil {
		t.Fatal(err)
	}
	roles := make(map[string]string)
	packageNames := make(map[string]bool)
	binaryNames := make(map[string]bool)
	for _, p := range graph.Packages {
		relative, err := filepath.Rel(graph.Root, p.ManifestPath)
		if err != nil || !filepath.IsLocal(relative) {
			t.Fatalf("workspace package is outside captured workspace: %s", p.ManifestPath)
		}
		location := filepath.ToSlash(filepath.Dir(relative))
		if len(p.Targets) != 1 || len(p.Targets[0].Kind) != 1 {
			t.Fatalf("package %s acquired additional target roles", location)
		}
		roles[location] = p.Targets[0].Kind[0]
		if p.Name == "" || p.Targets[0].Name == "" || packageNames[p.Name] {
			t.Fatalf("empty or colliding allocated Cargo package/target identity: %s", p.Name)
		}
		packageNames[p.Name] = true
		if p.Targets[0].Kind[0] == "bin" {
			key := strings.ToLower(p.Targets[0].Name)
			if binaryNames[key] {
				t.Fatalf("case-equivalent executable target collision: %s", p.Targets[0].Name)
			}
			binaryNames[key] = true
		}
	}
	want := map[string]string{"orchestrator": "bin", "support/runtime": "lib"}
	for i := range apps {
		want[fmt.Sprintf("apps/app-%03d", i)] = "bin"
	}
	if !reflect.DeepEqual(roles, want) {
		t.Fatalf("native Cargo roles = %v; want %v", roles, want)
	}
}

func runnableRun(t testing.TB, binary, directory string, environment []string) (int, []byte, []byte) {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	command := exec.CommandContext(ctx, binary)
	command.Dir = directory
	command.Env = append(os.Environ(), environment...)
	command.Env = append(command.Env, "PATH="+t.TempDir())
	command.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	command.Cancel = func() error { return syscall.Kill(-command.Process.Pid, syscall.SIGKILL) }
	command.WaitDelay = 2 * time.Second
	var stdout, stderr bytes.Buffer
	command.Stdout, command.Stderr = &stdout, &stderr
	err := command.Run()
	if ctx.Err() != nil {
		t.Fatalf("runtime watchdog: %v", ctx.Err())
	}
	if err == nil {
		return 0, stdout.Bytes(), stderr.Bytes()
	}
	if exit, ok := err.(*exec.ExitError); ok {
		return exit.ExitCode(), stdout.Bytes(), stderr.Bytes()
	}
	t.Fatalf("run actual native executable: %v", err)
	return -1, nil, nil
}

func TestRunnableServiceBuildAndRuntime(t *testing.T) {
	for _, tc := range runnableServiceCases {
		t.Run(tc.name, func(t *testing.T) {
			built := runnableBuild(t, tc.source, tc.apps)
			assertRunnableExactWorkspace(t, built, strings.TrimSuffix(tc.source, ".l"))
			status, stdout, stderr := runnableRun(t, built.binary, t.TempDir(), nil)
			if status != 0 || string(stdout) != tc.stdout || len(stderr) != 0 {
				t.Fatalf("App runtime status/stdout/stderr = %d/%q/%q; want 0/%q/empty", status, stdout, stderr, tc.stdout)
			}
		})
	}
}

func TestRunnableServiceRuntimeMisuse(t *testing.T) {
	for _, tc := range []struct {
		source, operation, message string
		line                       int
	}{
		{"await-before-run.l", "await", "Await requires an invocation", 5},
		{"output-before-run.l", "show", "Output is unavailable before completed invocation capture", 5},
		{"run-twice.l", "run", "Run conflicts with an outstanding invocation", 6},
	} {
		t.Run(tc.source, func(t *testing.T) {
			built := runnableBuild(t, tc.source, []string{"Worker"})
			status, stdout, stderr := runnableRun(t, built.binary, t.TempDir(), nil)
			want := runnableError(t, tc.line, tc.operation, tc.message, filepath.Join(built.project, "main.l"))
			if status != 1 || len(stdout) != 0 || string(stderr) != want {
				t.Fatalf("App misuse changed its process contract: %d/%q/%q; want 1/empty/%q", status, stdout, stderr, want)
			}
		})
	}
}

func TestRunnableServiceExactRustWorkspace(t *testing.T) {
	built := runnableBuild(t, "basic.l", []string{"Worker"})
	assertRunnableExactWorkspace(t, built, "basic")
}

func runnableExpectedWorkspaceFiles(t testing.TB, scenario, source, generation string) map[string][]byte {
	t.Helper()
	base := filepath.Join("test", "runnable_service", "expected-cases", scenario)
	files := make(map[string][]byte)
	err := filepath.WalkDir(base, func(path string, entry os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() {
			return nil
		}
		relative, err := filepath.Rel(base, path)
		if err != nil {
			return err
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		files[filepath.ToSlash(relative)] = runnableSubstituteExpected(t, data, source, generation)
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(files) == 0 {
		t.Fatal("empty independent Rust corpus")
	}
	// Shared immutable support is kept at its canonical resource location. No
	// scenario-specific expected Rust is extracted from production or snapshots.
	files["support/runtime/src/lib.rs"] = runnableRead(t, "expected/support/runtime/src/lib.rs")
	files["support/runtime/Cargo.toml"] = runnableRead(t, "expected/support/runtime/Cargo.toml")
	root := distributiontest.RepoRoot(t)
	for _, name := range []string{"runtime", "process_unix", "process_other"} {
		data, err := os.ReadFile(filepath.Join(root, "builtins", "templates", "runtime", name+".lu"))
		if err != nil {
			t.Fatal(err)
		}
		files["support/runtime/src/"+name+".rs"] = data
	}
	for name := range files {
		if strings.HasSuffix(name, "/Cargo.toml") && !strings.HasPrefix(name, "services/") && !strings.HasPrefix(name, "support/") {
			config := strings.TrimSuffix(name, "Cargo.toml") + "src/config.rs"
			if _, exists := files[config]; !exists {
				expectedConfig := "expected/config.rs"
				if strings.HasPrefix(name, "apps/") {
					expectedConfig = "expected/app-config.rs"
				}
				files[config] = runnableRead(t, expectedConfig)
			}
		}
	}
	return files
}

func assertRunnableExactWorkspace(t testing.TB, built runnableBuildResult, scenario string) {
	t.Helper()
	expected := runnableExpectedWorkspaceFiles(t, scenario, filepath.Join(built.project, "main.l"), filepath.Base(filepath.Dir(built.manifestPath)))
	actual := make(map[string][]byte)
	err := filepath.WalkDir(built.workspace, func(path string, entry os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() {
			return nil
		}
		relative, err := filepath.Rel(built.workspace, path)
		if err != nil {
			return err
		}
		relative = filepath.ToSlash(relative)
		if !strings.HasSuffix(relative, ".rs") && entry.Name() != "Cargo.toml" {
			return nil
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		actual[relative] = data
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(actual) != len(expected) {
		t.Errorf("%s raw Rust/manifest inventory=%d files; want %d", scenario, len(actual), len(expected))
	}
	for name, want := range expected {
		got, exists := actual[name]
		if !exists || !bytes.Equal(got, want) {
			t.Errorf("%s raw %s differs from complete independent oracle (present=%t)\nACTUAL:\n%s\nEXPECTED:\n%s", scenario, name, exists, got, want)
		}
		delete(actual, name)
	}
	if len(actual) != 0 {
		t.Errorf("unexpected generated Rust/manifest artifacts: %v", reflect.ValueOf(actual).MapKeys())
	}
}

func TestRunnableServiceLibraryConsumer(t *testing.T) {
	root := distributiontest.RepoRoot(t)
	c := newCompiler(&cliData{jobType: "build", mainPath: filepath.Join(root, "lyger", "test", "workspace", "services.l"), osType: "ubuntu"})
	c.initialize()
	var generated *generatedProgram
	if output := sourceStdout(t, func() { generated = c.process() }); c.err != nil || generated == nil || output != "" {
		t.Fatalf("App-owned library source prerequisite: %v/%q", c.err, output)
	}
	if len(generated.ApplicationIDs) != 2 || len(generated.LibraryIDs) != 4 {
		t.Fatalf("App-owned workspace dimensions: %d Apps, %d libraries", len(generated.ApplicationIDs), len(generated.LibraryIDs))
	}
	nativetest.RecordBundle(t, "app-owned-libraries", generated.LibrarySources)
	files := map[string][]byte{}
	for path, data := range runnableBundleFiles(t, generated.LibrarySources, 0) {
		files["library/"+strings.TrimPrefix(path, "/")] = data
	}
	files["Cargo.toml"] = []byte("[workspace]\nresolver = \"3\"\nmembers = [\"library\", \"consumer\"]\n")
	files["consumer/Cargo.toml"] = []byte(fmt.Sprintf("[package]\nname = \"service_consumer\"\nversion = \"0.0.0\"\nedition = \"2024\"\n[dependencies]\nleft = { package = %q, path = \"../library\" }\n", generated.LibraryIDs[0]))
	files["consumer/src/main.rs"] = []byte("fn main() { left::left::left(); }\n")
	files["Cargo.lock"] = []byte(fmt.Sprintf("version = 4\n\n[[package]]\nname = %q\nversion = \"0.0.0\"\n\n[[package]]\nname = \"service_consumer\"\nversion = \"0.0.0\"\ndependencies = [%q]\n", generated.LibraryIDs[0], generated.LibraryIDs[0]))
	project := nativetest.Materialize(t, files)
	compiled := nativetest.Build(t, project, nativetest.BuildOptions{Timeout: nativetest.BuildPrerequisiteTimeout})
	if compiled.Err != nil {
		t.Fatalf("actual unmodified App-owned library does not link independently: %v\n%s", compiled.Err, compiled.Stderr)
	}
	for _, artifact := range compiled.Artifacts {
		if artifact.TargetName == "service_consumer" {
			status, stdout, stderr := runnableRun(t, artifact.Executable, t.TempDir(), nil)
			if status != 0 || string(stdout) != "first-service-body\n" || len(stderr) != 0 {
				t.Fatalf("independent library consumer = %d/%q/%q", status, stdout, stderr)
			}
			return
		}
	}
	t.Fatal("Cargo did not return independent consumer artifact")
}

func TestRunnableServicePublicationPreservesGenerations(t *testing.T) {
	first := runnableBuild(t, "mixed.l", []string{"Host", "Worker"})
	oldBinary := filepath.Join(first.project, "retained-main")
	data, err := os.ReadFile(first.binary)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(oldBinary, data, 0700); err != nil {
		t.Fatal(err)
	}
	oldFiles := runnableGenerationSnapshot(t, filepath.Dir(first.manifestPath))
	updated := bytes.ReplaceAll(runnableRead(t, "mixed.l"), []byte("work completed"), []byte("second generation"))
	if err := os.WriteFile(filepath.Join(first.project, "main.l"), updated, 0600); err != nil {
		t.Fatal(err)
	}
	second := runnableBuildProject(t, first.project, []string{"Host", "Worker"})
	if first.manifestPath == second.manifestPath || !reflect.DeepEqual(oldFiles, runnableGenerationSnapshot(t, filepath.Dir(first.manifestPath))) {
		t.Fatal("new build reused or modified prior companion generation")
	}
	for _, tc := range []struct{ path, output string }{{oldBinary, "app\nwork completed\n"}, {second.binary, "app\nsecond generation\n"}} {
		status, stdout, stderr := runnableRun(t, tc.path, t.TempDir(), nil)
		if status != 0 || string(stdout) != tc.output || len(stderr) != 0 {
			t.Fatalf("retained/new generation mixed companions: %d/%q/%q", status, stdout, stderr)
		}
	}
	beforeBinary, err := os.ReadFile(second.binary)
	if err != nil {
		t.Fatal(err)
	}
	before := runnableGenerationSnapshot(t, second.binary+".apps")
	if err := os.WriteFile(filepath.Join(first.project, "main.l"), []byte("Use `absent.l`.\n"), 0600); err != nil {
		t.Fatal(err)
	}
	out, diagnostic, err := runBuildCommand(t, phase1.CLI(t), first.project, nil, "build", "main.l")
	if err == nil || len(out) != 0 || !bytes.Contains(diagnostic, []byte("absent.l")) {
		t.Fatalf("failed rebuild did not report actual source failure: %v/%q/%q", err, out, diagnostic)
	}
	afterBinary, err := os.ReadFile(second.binary)
	if err != nil || !bytes.Equal(beforeBinary, afterBinary) || !reflect.DeepEqual(before, runnableGenerationSnapshot(t, second.binary+".apps")) {
		t.Fatal("failed build changed the published mixed generation")
	}
}

func runnableGenerationSnapshot(t testing.TB, root string) map[string][32]byte {
	t.Helper()
	files := make(map[string][32]byte)
	err := filepath.WalkDir(root, func(path string, entry os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() {
			return nil
		}
		relative, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		files[relative] = sha256.Sum256(data)
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	return files
}

func TestRunnableServiceRelocation(t *testing.T) {
	built := runnableBuild(t, "mixed.l", []string{"Host", "Worker"})
	destination := t.TempDir()
	for _, name := range []string{"main", "main.apps"} {
		if err := os.Rename(filepath.Join(built.project, name), filepath.Join(destination, name)); err != nil {
			t.Fatal(err)
		}
	}
	status, stdout, stderr := runnableRun(t, filepath.Join(destination, "main"), t.TempDir(), nil)
	if status != 0 || string(stdout) != "app\nwork completed\n" || len(stderr) != 0 {
		t.Fatalf("relocated mixed generation = %d/%q/%q", status, stdout, stderr)
	}
}

func TestRunnableServiceExpectedRustCompiles(t *testing.T) {
	// Compile and execute an independently authored App body separately from
	// the complete expected workspace and every actual generated artifact.
	files := map[string][]byte{
		"Cargo.toml":    runnableRead(t, "expected-cases/basic/apps/app-000/Cargo.toml"),
		"src/main.rs":   runnableRead(t, "expected-cases/basic/apps/app-000/src/main.rs"),
		"src/config.rs": runnableRead(t, "expected/app-config.rs"),
		"Cargo.lock":    []byte("version = 4\n\n[[package]]\nname = \"Worker\"\nversion = \"0.0.0\"\n"),
	}
	project := nativetest.Materialize(t, files)
	result := nativetest.Build(t, project, nativetest.BuildOptions{Timeout: nativetest.BuildPrerequisiteTimeout})
	if result.Err != nil {
		t.Fatalf("independent expected Rust is invalid: %v\n%s", result.Err, result.Stderr)
	}
	var found []string
	for _, artifact := range result.Artifacts {
		found = append(found, artifact.TargetName)
		nativetest.AssertNativeArtifact(t, artifact.Executable)
		status, stdout, stderr := runnableRun(t, artifact.Executable, t.TempDir(), nil)
		if status != 0 || string(stdout) != "work completed\n" || len(stderr) != 0 {
			t.Fatalf("expected Rust contract = %d/%q/%q", status, stdout, stderr)
		}
	}
	sort.Strings(found)
	if !reflect.DeepEqual(found, []string{"Worker"}) {
		t.Fatalf("expected Rust native artifacts = %v", found)
	}
}

func TestRunnableServiceExpectedWorkspaceCorpusCompiles(t *testing.T) {
	for _, tc := range runnableServiceCases {
		t.Run(tc.name, func(t *testing.T) {
			files := runnableExpectedWorkspaceFiles(t, strings.TrimSuffix(tc.source, ".l"), "/expected/main.l", "expected")
			members := []string{"orchestrator"}
			for i := range tc.apps {
				members = append(members, fmt.Sprintf("apps/app-%03d", i))
			}
			members = append(members, "support/runtime")
			var manifest strings.Builder
			manifest.WriteString("[workspace]\nresolver = \"3\"\nmembers = [\n")
			var packages []*cargoPackage
			for _, member := range members {
				fmt.Fprintf(&manifest, " %q,\n", member)
				var paths []string
				for name := range files {
					if strings.HasPrefix(name, member+"/") {
						paths = append(paths, strings.TrimPrefix(name, member+"/"))
					}
				}
				sort.Strings(paths)
				var contents [][]byte
				for _, path := range paths {
					contents = append(contents, files[member+"/"+path])
				}
				bundle, err := planSourceBundle(contents, paths)
				if err != nil {
					t.Fatalf("independent %s manifest/source contract: %v", member, err)
				}
				packages = append(packages, bundle.packagePlan)
			}
			manifest.WriteString("]\n")
			if !bytes.Equal(files["Cargo.toml"], []byte(manifest.String())) {
				t.Fatal("independent root manifest does not match ordered expected package members")
			}
			// Assemble only local lock records around the retained pinned external
			// fragment. No dependency resolution, updates, or source generation.
			lock, err := cargoWorkspaceLock(packages)
			if err != nil {
				t.Fatal(err)
			}
			files["Cargo.lock"] = lock
			project := nativetest.Materialize(t, files)
			compiled := nativetest.Build(t, project, nativetest.BuildOptions{Timeout: nativetest.BuildPrerequisiteTimeout})
			if compiled.Err != nil {
				t.Fatalf("independent complete expected Rust does not compile: %v\n%s", compiled.Err, compiled.Stderr)
			}
			if len(compiled.Artifacts) != 1+len(tc.apps) {
				t.Fatalf("expected native binary count=%d; got %d", 1+len(tc.apps), len(compiled.Artifacts))
			}
		})
	}
}

func TestRunnableServiceAppOnlyManifestV1Control(t *testing.T) {
	project, _ := runnableProject(t, "basic-app-control.l")
	recorder := nativetest.CLIProbe(t)
	stdout, stderr, err := runBuildCommand(t, phase1.CLI(t), project, runnableRecorderEnv(recorder), "build", "main.l")
	if err != nil || len(stderr) != 0 || string(stdout) != filepath.Join(project, "main")+"\n" {
		t.Fatalf("App build = %v/%q/%q", err, stdout, stderr)
	}
	binary := filepath.Join(project, "main")
	recorder.AssertPublished(t, binary)
	assertRunnableIdentityArtifact(t, recorder.Directory, binary, "orchestrator")
	var raw map[string]json.RawMessage
	nativetest.ReadPublishedGeneration(t, binary, &raw)
	var keys []string
	for key := range raw {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	if !reflect.DeepEqual(keys, []string{"apps", "entrypoint", "goarch", "goos", "version"}) || string(raw["version"]) != "1" {
		t.Fatalf("App-only manifest wire shape changed: %v", raw)
	}
	status, out, diagnostics := runnableRun(t, binary, t.TempDir(), nil)
	if status != 0 || string(out) != "work completed\n" || len(diagnostics) != 0 {
		t.Fatalf("App-only behavior changed: %d/%q/%q", status, out, diagnostics)
	}
}

func TestRunnableServicePublicationProtectsDestinations(t *testing.T) {
	for _, kind := range []string{"unrelated output", "companion symlink", "generation symlink", "source input", "selected compiler assets"} {
		t.Run(kind, func(t *testing.T) {
			// Require valid source first: an unrelated compiler error
			// must not pass a protection test aimed at a later build boundary.
			_, _, project := runnableProcess(t, "basic.l")
			filename, destination, assets := filepath.Join(project, "main.l"), project, ""
			protected := project
			want := "unrelated"
			switch kind {
			case "unrelated output":
				if err := os.WriteFile(filepath.Join(project, "main"), []byte("owned by another program"), 0600); err != nil {
					t.Fatal(err)
				}
			case "companion symlink":
				protected = t.TempDir()
				want = "companion path is not an owned directory"
				if err := os.Symlink(protected, filepath.Join(project, "main.apps")); err != nil {
					t.Fatal(err)
				}
			case "generation symlink":
				protected = t.TempDir()
				want = "invalid generation directory"
				if err := os.Mkdir(filepath.Join(project, "main.apps"), 0700); err != nil {
					t.Fatal(err)
				}
				if err := os.WriteFile(filepath.Join(project, "main.apps", ".l-owner"), []byte(companionOwner+"main\n"), 0600); err != nil {
					t.Fatal(err)
				}
				if err := os.Symlink(protected, filepath.Join(project, "main.apps", "generations")); err != nil {
					t.Fatal(err)
				}
			case "source input":
				want = "source input"
				filename = filepath.Join(project, "target.l.l")
				if err := os.WriteFile(filepath.Join(project, "target.l"), []byte("\"imported\" is `protected`.\n"), 0600); err != nil {
					t.Fatal(err)
				}
				if err := os.WriteFile(filename, append([]byte("Use `target.l`.\n"), runnableRead(t, "basic.l")...), 0600); err != nil {
					t.Fatal(err)
				}
				if err := os.MkdirAll(filepath.Join(project, "target.l.apps", "generations"), 0700); err != nil {
					t.Fatal(err)
				}
				if err := os.WriteFile(filepath.Join(project, "target.l.apps", ".l-owner"), []byte(companionOwner+"target.l\n"), 0600); err != nil {
					t.Fatal(err)
				}
			case "selected compiler assets":
				assets = distributiontest.CompleteOverride(t)
				destination = filepath.Join(assets, "builtins")
				protected, want = assets, "compiler assets"
			}
			before := runnableGenerationSnapshot(t, protected)
			args := []string{"build", "--output-dir", destination}
			if assets != "" {
				args = append(args, "--assets", assets)
			}
			args = append(args, filename)
			stdout, stderr, err := runBuildCommand(t, phase1.CLI(t), project, nil, args...)
			if err == nil || len(stdout) != 0 || !strings.Contains(string(stderr), want) {
				t.Fatalf("wrong protected-destination boundary: %v/%q/%q; want %s", err, stdout, stderr, want)
			}
			if !reflect.DeepEqual(before, runnableGenerationSnapshot(t, protected)) {
				t.Fatal("rejected App build modified protected destination")
			}
		})
	}
}

func TestRunnableServicePackageNameCollisions(t *testing.T) {
	for _, app := range []string{"l_service_worker", "L_SERVICE_WORKER", "build", "l_runtime_support"} {
		t.Run(app, func(t *testing.T) {
			project, filename := runnableProject(t, "mixed.l")
			source := bytes.ReplaceAll(runnableRead(t, "mixed.l"), []byte("Host"), []byte(app))
			if err := os.WriteFile(filename, source, 0600); err != nil {
				t.Fatal(err)
			}
			built := runnableBuildProject(t, project, []string{app, "Worker"})
			// Cargo workspace metadata/provenance must demonstrate allocation;
			// successful execution ensures the intended Apps retain their identities.
			status, stdout, stderr := runnableRun(t, built.binary, t.TempDir(), nil)
			if status != 0 || string(stdout) != "app\nwork completed\n" || len(stderr) != 0 {
				t.Fatalf("colliding names changed targets: %d/%q/%q", status, stdout, stderr)
			}
		})
	}
}

func TestRunnableServiceConcurrentPublication(t *testing.T) {
	// The source files are distinct; only publication is shared. Both real
	// Cargo builds must complete before the selected generation is inspected.
	output := t.TempDir()
	cli := phase1.CLI(t)
	type job struct {
		project  string
		recorder *nativetest.CLIRecorder
		env      []string
	}
	var jobs []job
	for _, text := range []string{"first", "second"} {
		project, filename := runnableProject(t, "mixed.l")
		source := bytes.ReplaceAll(runnableRead(t, "mixed.l"), []byte("work completed"), []byte("service-"+text))
		source = bytes.ReplaceAll(source, []byte("`app\n`"), []byte("`app-"+text+"\n`"))
		if err := os.WriteFile(filename, source, 0600); err != nil {
			t.Fatal(err)
		}
		recorder := nativetest.CLIProbe(t)
		jobs = append(jobs, job{project, recorder, runnableRecorderEnv(recorder)})
	}
	type result struct {
		stdout, stderr []byte
		err            error
	}
	completed := make(chan result, len(jobs))
	for _, work := range jobs {
		go func(work job) {
			ctx, cancel := context.WithTimeout(context.Background(), nativetest.BuildPrerequisiteTimeout)
			defer cancel()
			cmd := exec.CommandContext(ctx, cli, "build", "--output-dir", output, "main.l")
			cmd.Dir, cmd.Env = work.project, work.env
			cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
			cmd.Cancel = func() error { return syscall.Kill(-cmd.Process.Pid, syscall.SIGKILL) }
			cmd.WaitDelay = 2 * time.Second
			var stdout, stderr bytes.Buffer
			cmd.Stdout, cmd.Stderr = &stdout, &stderr
			err := cmd.Run()
			completed <- result{stdout.Bytes(), stderr.Bytes(), err}
		}(work)
	}
	for range jobs {
		got := <-completed
		if got.err != nil || len(got.stderr) != 0 || string(got.stdout) != filepath.Join(output, "main")+"\n" {
			t.Errorf("concurrent App build = %v/%q/%q", got.err, got.stdout, got.stderr)
		}
	}
	if t.Failed() {
		return
	}
	manifests, err := filepath.Glob(filepath.Join(output, "main.apps", "generations", "*", "manifest.json"))
	if err != nil || len(manifests) != 2 {
		t.Fatalf("concurrent publication lost a generation: %v %v", manifests, err)
	}
	matchedRecorders := make(map[int]bool)
	for _, manifest := range manifests {
		data, err := os.ReadFile(manifest)
		if err != nil {
			t.Fatal(err)
		}
		var generation runnableManifest
		if err := json.Unmarshal(data, &generation); err != nil {
			t.Fatal(err)
		}
		if generation.Version != 1 || len(generation.Apps) != 2 {
			t.Fatalf("partial concurrent generation: %+v", generation)
		}
		if generation.Apps[0].ID != "Host" || generation.Apps[0].Executable != "app-000" || generation.Apps[1].ID != "Worker" || generation.Apps[1].Executable != "app-001" {
			t.Fatalf("concurrent manifest lost companion identity: %+v", generation)
		}
		for _, companion := range generation.Apps {
			if nativetest.NativeArtifactHash(t, filepath.Join(filepath.Dir(manifest), companion.Executable)) != companion.SHA256 {
				t.Fatal("concurrent generation hash mismatch")
			}
		}
		matched := -1
		for i, work := range jobs {
			all := true
			for _, companion := range generation.Apps {
				candidates, err := filepath.Glob(filepath.Join(work.recorder.Directory, "cargo.*.artifacts", companion.SHA256))
				if err != nil {
					t.Fatal(err)
				}
				all = all && len(candidates) == 1
			}
			if all {
				if matched >= 0 {
					t.Fatal("generation unexpectedly matches both distinct native builds")
				}
				matched = i
			}
		}
		if matched < 0 || matchedRecorders[matched] {
			t.Fatal("published generation mixed Cargo artifacts from independent builds")
		}
		matchedRecorders[matched] = true
		assertRunnableIdentityArtifact(t, jobs[matched].recorder.Directory, filepath.Join(filepath.Dir(manifest), generation.Apps[0].Executable), "apps/app-000")
		assertRunnableIdentityArtifact(t, jobs[matched].recorder.Directory, filepath.Join(filepath.Dir(manifest), generation.Apps[1].Executable), "apps/app-001")
	}
	status, stdout, stderr := runnableRun(t, filepath.Join(output, "main"), t.TempDir(), nil)
	if status != 0 || len(stderr) != 0 || string(stdout) != "app-first\nservice-first\n" && string(stdout) != "app-second\nservice-second\n" {
		t.Fatalf("concurrent publication mixed source generations: %d/%q/%q", status, stdout, stderr)
	}
	winner := 0
	if string(stdout) == "app-second\nservice-second\n" {
		winner = 1
	}
	assertRunnableIdentityArtifact(t, jobs[winner].recorder.Directory, filepath.Join(output, "main"), "orchestrator")
}
