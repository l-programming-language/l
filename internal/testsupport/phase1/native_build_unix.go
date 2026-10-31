//go:build darwin || linux

package phase1

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"syscall"
	"testing"
	"time"

	nativetest "github.com/l-programming-language/l/internal/testsupport/native"
	"github.com/l-programming-language/l/internal/timing"
)

// Native lifecycle observers run after mandatory Cargo provenance.
// This adapter never translates returned AML or generated source.
var buildProbes sync.Map

func buildProbe(t testing.TB, project string) *nativetest.CLIRecorder {
	if p, ok := buildProbes.Load(project); ok {
		return p.(*nativetest.CLIRecorder) //nolint:errcheck // buildProbe is the sole writer and stores only CLIRecorder pointers.
	}
	p := nativetest.CLIProbe(t)
	buildProbes.Store(project, p)
	t.Cleanup(func() { buildProbes.Delete(project) })
	return p
}
func LoadFixture(t testing.TB, id string) Fixture {
	t.Helper()
	paths, err := filepath.Glob(filepath.Join(RepoRoot(t), "testdata", "phase1", strings.ToLower(id)+"-*", "case.json"))
	if err != nil || len(paths) != 1 {
		t.Fatalf("fixture %s: expected one manifest, got %v (%v)", id, paths, err)
	}
	data, err := os.ReadFile(paths[0])
	if err != nil {
		t.Fatal(err)
	}
	f, err := DecodeFixture(data)
	if err != nil {
		t.Fatalf("%s: %v", paths[0], err)
	}
	if f.ID != id {
		t.Fatalf("manifest ID %s != %s", f.ID, id)
	}
	f.Directory = filepath.Dir(paths[0])
	return f
}

func buildProgram(t testing.TB, project string, env []string) (string, []byte, []byte, error) {
	t.Helper()
	cli := CLI(t) // Build the Go-implemented compiler before denying child Go builds.
	probe := buildProbe(t, project)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	out, diag, err := controlledBuildCommand(ctx, cli, project, probe.Env(append(os.Environ(), env...)), "build", "main.l")
	saveBuildEvidence(t, project, out, diag, err)
	canonical, resolveErr := filepath.EvalSymlinks(project)
	if resolveErr != nil && err == nil {
		err = resolveErr
	}
	if canonical == "" {
		canonical = project
	}
	binary := filepath.Join(canonical, "main")
	if err == nil {
		probe.AssertPublished(t, binary)
	}
	return binary, out, diag, err
}

// env is the complete controlled tool environment. Re-adding os.Environ here
// would restore compiler overrides deliberately removed by CLIRecorder.Env.
// Match the original process-group supervision and watchdog exactly.
func controlledBuildCommand(ctx context.Context, executable, directory string, env []string, args ...string) (outBytes, diagnosticBytes []byte, err error) {
	finish := timing.Start("test.command.total", fmt.Sprintf("%s %q", executable, args))
	defer func() { finish(err) }()
	cmd := exec.CommandContext(ctx, executable, args...)
	cmd.Dir, cmd.Env = directory, env
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	cmd.Cancel = func() error { return syscall.Kill(-cmd.Process.Pid, syscall.SIGKILL) }
	cmd.WaitDelay = 2 * time.Second
	var out, diagnostic bytes.Buffer
	cmd.Stdout, cmd.Stderr = &out, &diagnostic
	err = cmd.Run()
	if ctx.Err() != nil {
		err = fmt.Errorf("command watchdog: %w (%v)", ctx.Err(), err)
	}
	return out.Bytes(), diagnostic.Bytes(), err
}

func RequireBuild(t testing.TB, project string) string {
	t.Helper()
	binary, out, diag, err := buildProgram(t, project, nil)
	if err != nil {
		t.Fatalf("compile/build prerequisite; runtime not exercised: %v\nstdout=%q\nstderr=%s", err, out, diag)
	}
	if string(out) != binary+"\n" || len(diag) != 0 {
		t.Fatalf("build output stdout=%q stderr=%q", out, diag)
	}
	return binary
}

func RunFixture(t *testing.T, id string) {
	t.Helper()
	f := LoadFixture(t, id)
	project := Materialize(t, f)
	if f.ExpectedCompileError != nil {
		runCompileNegative(t, f, project)
		return
	}
	if id == "F20" {
		// The source and native tool inputs are identical for both signals.
		// Retain one real CLI/Cargo build under this root, then install its exact
		// executable bytes into a fresh directory for every runtime profile.
		binary := RequireBuild(t, project)
		install := retainSignalProgram(t, binary)
		probe := buildProbe(t, project)
		for _, signal := range []struct {
			name  string
			value syscall.Signal
			code  int
		}{{"SIGINT", syscall.SIGINT, 130}, {"SIGTERM", syscall.SIGTERM, 143}} {
			t.Run(signal.name, func(t *testing.T) {
				executable := install(t)
				probe.AssertPublished(t, executable)
				runBuilt(t, f, executable, signal.value, signal.code)
			})
		}
		return
	}
	binary := RequireBuild(t, project)
	if id == "F07" {
		apps := companions(t, binary)
		if len(apps) != 1 {
			t.Fatalf("launch-fault fixture requires one companion, got %v", apps)
		}
		if err := os.Remove(apps[0]); err != nil {
			t.Fatal(err)
		}
	}
	var status int
	if err := json.Unmarshal(f.ExpectedRuntime.ExitStatus, &status); err != nil {
		t.Fatalf("invalid runtime status: %s", f.ExpectedRuntime.ExitStatus)
	}
	runBuilt(t, f, binary, 0, status)
}

// Retain only this build's published executable tree. Runtime signal profiles
// get separate files and controllers, while Cargo provenance stays with the
// parent build. No source, manifest, or executable bytes are rewritten.
func retainSignalProgram(t *testing.T, binary string) func(*testing.T) string {
	t.Helper()
	type retainedFile struct {
		data []byte
		mode os.FileMode
	}
	files := map[string]retainedFile{}
	retain := func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() {
			return nil
		}
		info, err := entry.Info()
		if err != nil {
			return err
		}
		if !info.Mode().IsRegular() {
			return fmt.Errorf("published runtime input is not regular: %s", path)
		}
		data, err := os.ReadFile(path)
		if err == nil {
			files[path] = retainedFile{data, info.Mode().Perm()}
		}
		return err
	}
	for _, path := range []string{binary, binary + ".apps"} {
		if err := filepath.WalkDir(path, retain); err != nil {
			t.Fatal(err)
		}
	}
	owner := t.Name()
	t.Cleanup(func() {
		for path, file := range files {
			got, err := os.ReadFile(path)
			if err != nil || !bytes.Equal(got, file.data) {
				t.Errorf("parent native build changed after signal profiles: %s: %v", path, err)
			}
		}
	})
	return func(t *testing.T) string {
		t.Helper()
		directory := t.TempDir()
		for path, file := range files {
			relative, err := filepath.Rel(filepath.Dir(binary), path)
			if err != nil || !filepath.IsLocal(relative) {
				t.Fatalf("invalid published runtime path: %s: %v", path, err)
			}
			destination := filepath.Join(directory, relative)
			if err := os.MkdirAll(filepath.Dir(destination), 0700); err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(destination, file.data, file.mode); err != nil {
				t.Fatal(err)
			}
			if got, err := os.ReadFile(destination); err != nil || !bytes.Equal(got, file.data) {
				t.Fatalf("runtime input differs from parent native build: %s: %v", relative, err)
			}
		}
		t.Logf("native build evidence: %s; orchestrator SHA256 %x", owner, sha256.Sum256(files[binary].data))
		return filepath.Join(directory, filepath.Base(binary))
	}
}

func runCompileNegative(t *testing.T, f Fixture, project string) {
	binary, out, diag, err := buildProgram(t, project, nil)
	if err == nil {
		t.Fatal("negative source was accepted")
	}
	if len(out) != 0 {
		t.Errorf("failed compilation polluted stdout: %q", out)
	}
	if _, err := os.Stat(binary); !errors.Is(err, os.ErrNotExist) {
		t.Errorf("negative compilation published an executable: %v", err)
	}
	// Require provenance of the offending main statement. A missing distribution
	// resource cannot satisfy a semantic rejection; frontend tests inspect causes.
	lines, err := os.ReadFile(filepath.Join(project, "main.l"))
	if err != nil {
		t.Fatal(err)
	}
	source := strings.TrimSuffix(string(lines), "\n")
	line := strings.Count(source, "\n") + 1
	location := fmt.Sprintf("main.l:%d:", line)
	message := strings.ToLower(string(diag))
	if !strings.Contains(string(diag), location) {
		t.Fatalf("wrong rejection provenance, want %s; intended %s assertion not exercised\n%s", location, f.ExpectedCompileError.Category, diag)
	}
	required := map[string][]string{"N01": {"await", "app"}, "N02": {"missing", "unresolved"}, "N03": {"grammar", "suspensive", "object"}, "N04": {"stop", "app"}, "N09": {"unsupported", "app"}, "N10": {"privateworker", "unresolved"}, "N11": {"pronoun", "it", "paragraph", "no preceding subject"}, "N12": {"conditional"}}[f.ID]
	for _, word := range required {
		if !strings.Contains(message, word) {
			t.Errorf("wrong rejection for %s: missing %q; diagnostic=%s", f.ID, word, diag)
		}
	}
	if strings.Contains(message, "unknown await") || strings.Contains(message, "unrecognized await") || strings.Contains(message, "unsupported instruction await") && f.ID != "N09" {
		t.Errorf("unrelated unsupported Await cannot satisfy %s: %s", f.ID, diag)
	}
}

func RunSource(t *testing.T, source, wantStdout string, localFiles ...string) {
	RunSourceOutcomes(t, source, []string{wantStdout}, localFiles...)
}

func RunSourceOutcomes(t *testing.T, source string, outputs []string, localFiles ...string) {
	t.Helper()
	project := t.TempDir()
	if err := os.WriteFile(filepath.Join(project, "main.l"), []byte(source), 0600); err != nil {
		t.Fatal(err)
	}
	for _, path := range localFiles {
		if filepath.Base(path) == "main.l" || filepath.Ext(path) != ".l" {
			t.Fatalf("invalid explicit local source dependency %q", path)
		}
		CopyFile(t, path, filepath.Join(project, filepath.Base(path)), 0600)
	}
	binary := RequireBuild(t, project)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	out, diag, err := command(ctx, binary, t.TempDir(), nil)
	if err != nil || len(diag) != 0 {
		t.Fatalf("native language runtime: %v stdout=%q stderr=%q", err, out, diag)
	}
	for _, want := range outputs {
		if string(out) == want {
			return
		}
	}
	t.Errorf("native stdout=%q; allowed exact outcomes=%q", out, outputs)
}
