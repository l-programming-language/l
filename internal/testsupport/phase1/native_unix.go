//go:build darwin || linux

package phase1

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"syscall"
	"testing"
	"time"

	nativetest "github.com/l-programming-language/l/internal/testsupport/native"
	"github.com/l-programming-language/l/internal/timing"
)

var installed struct {
	once                  sync.Once
	directory, executable string
	err                   error
}

func CLI(t testing.TB) string {
	t.Helper()
	installed.once.Do(func() {
		finish := timing.Start("test.cli.setup", t.Name())
		defer func() { finish(installed.err) }()
		installed.directory, installed.err = os.MkdirTemp("", "l-phase1-installed-")
		if installed.err != nil {
			return
		}
		installed.executable = filepath.Join(installed.directory, "l")
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		cmd := exec.CommandContext(ctx, "go", "build", "-mod=readonly", "-o", installed.executable, "./cmd/l")
		cmd.Dir = RepoRoot(t)
		output, err := cmd.CombinedOutput()
		if err != nil {
			installed.err = fmt.Errorf("installed CLI build: %w\n%s", err, output)
			return
		}
	})
	if installed.err != nil {
		t.Fatalf("installed-cli prerequisite; runtime not exercised: %v", installed.err)
	}
	return installed.executable
}

// Main removes immutable package-owned CLI build products after the package.
func Main(m *testing.M) {
	code := m.Run()
	if installed.directory != "" {
		_ = os.RemoveAll(installed.directory) //nolint:errcheck // Post-suite temporary executable cleanup must preserve the test process exit status.
	}
	os.Exit(code)
}

func command(ctx context.Context, executable, directory string, env []string, args ...string) (outBytes, diagnosticBytes []byte, err error) {
	finish := timing.Start("test.command.total", fmt.Sprintf("%s %q", executable, args))
	defer func() { finish(err) }()
	cmd := exec.CommandContext(ctx, executable, args...)
	cmd.Dir = directory
	cmd.Env = append(os.Environ(), env...)
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

type generation struct {
	Apps []struct{ ID, Executable, SHA256 string } `json:"apps"`
}

func companions(t testing.TB, binary string) []string {
	t.Helper()
	var m generation
	manifestPath := nativetest.ReadPublishedGeneration(t, binary, &m)
	out := []string{}
	for _, app := range m.Apps {
		path := filepath.Join(filepath.Dir(manifestPath), app.Executable)
		content, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		hash := sha256.Sum256(content)
		if hex.EncodeToString(hash[:]) != app.SHA256 {
			t.Errorf("immutable companion hash mismatch for %s", app.ID)
		}
		out = append(out, path)
	}
	return out
}

func labels(id string) []string {
	switch id {
	case "F02", "F09", "F10", "F20", "N06":
		return []string{"A"}
	case "F03", "F11", "F12", "F13", "F14":
		return []string{"A", "B"}
	case "F05":
		return []string{"worker", "other", "gate"}
	case "F17", "F18", "F19":
		return []string{"worker", "gate"}
	case "F22":
		return []string{"x", "gate"}
	}
	return nil
}
func runBuilt(t *testing.T, f Fixture, binary string, signal syscall.Signal, status int) {
	t.Helper()
	finish := timing.Start("runtime.total", t.Name()+" "+binary)
	defer func() {
		var err error
		if t.Failed() {
			err = errors.New("runtime assertions failed")
		}
		finish(err)
	}()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	ctrl, err := NewController(labels(f.ID)...)
	if err != nil {
		t.Fatal(err)
	}
	// This cleanup is installed before launch. It records held resources before
	// fallback signals or controller closure can hide a leak.
	defer func() {
		finish := timing.Start("runtime.fallback_cleanup", t.Name())
		var cleanupErr error
		defer func() { finish(cleanupErr) }()
		if leaks := ctrl.Leaks(); len(leaks) != 0 {
			cleanupErr = fmt.Errorf("runtime leak observed before fallback: %v", leaks)
			t.Errorf("fallback cleanup required; runtime leak observed before fallback: %v", leaks)
			ctrl.Fallback()
		}
		cleanupErr = errors.Join(cleanupErr, ctrl.Close())
	}()
	if f.ID == "F14" {
		ctrl.RequireReleaseBefore("B", "A")
	}
	cmd := exec.Command(binary)
	cmd.Dir = t.TempDir()
	cmd.Env = append(os.Environ(), "L_PHASE1_SOCKET="+ctrl.Socket, "L_PHASE1_DIRECTORY="+ctrl.Directory)
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	cmd.WaitDelay = 2 * time.Second
	out, diagnostic := &Stream{}, &Stream{}
	defer func() { saveRuntimeEvidence(t, out.Bytes(), diagnostic.Bytes(), ctrl.Events()) }()
	cmd.Stdout, cmd.Stderr = out, diagnostic
	started := timing.Start("runtime.start", t.Name()+" "+binary)
	startErr := cmd.Start()
	started(startErr)
	if err := startErr; err != nil {
		t.Fatal(err)
	}
	done := make(chan error, 1)
	go func() {
		finish := timing.Start("runtime.wait", t.Name()+" "+binary)
		err := cmd.Wait()
		finish(err)
		done <- err
	}()
	complete := false
	defer func() {
		if !complete {
			_ = syscall.Kill(-cmd.Process.Pid, syscall.SIGKILL) //nolint:errcheck // Best-effort fallback may race process exit; the owned Wait below still must complete.
			select {
			case <-done:
			case <-time.After(2 * time.Second):
				t.Error("fallback orchestrator Wait did not finish")
			}
		}
	}()
	ready := func(names ...string) (err error) {
		finish := timing.Start("runtime.readiness", t.Name()+" "+strings.Join(names, ","))
		defer func() { finish(err) }()
		if err := ctrl.WaitReady(ctx, names...); err != nil {
			return err
		}
		for _, label := range names {
			if err := ctrl.CheckAlive(label, cmd.Process.Pid); err != nil {
				return err
			}
		}
		return nil
	}
	checkpoint := func(s string) error {
		finish := timing.Start("runtime.output_wait", t.Name())
		err := out.WaitContains(ctx, s)
		finish(err)
		return err
	}
	release := func(names ...string) error {
		for _, label := range names {
			if err := ctrl.Release(label); err != nil {
				return err
			}
		}
		return nil
	}
	var protocol error
	switch f.ID {
	case "F02", "F10":
		if protocol = ready("A"); protocol == nil {
			protocol = checkpoint("launched|")
		}
		if protocol == nil {
			protocol = release("A")
		}
	case "F03", "F13":
		if protocol = ready("A", "B"); protocol == nil {
			protocol = checkpoint("launched|")
		}
		if protocol == nil {
			protocol = release("B")
		}
		if protocol == nil {
			finish := timing.Start("runtime.disconnect_wait", t.Name()+" B")
			protocol = ctrl.WaitDisconnected(ctx, "B")
			finish(protocol)
		}
		if protocol == nil {
			protocol = release("A")
		}
	case "F05":
		if protocol = ready("worker", "other", "gate"); protocol == nil {
			protocol = release("gate")
		}
		if protocol == nil {
			protocol = checkpoint("worker-stopped|")
		}
		if protocol == nil {
			protocol = ctrl.CheckAlive("other", cmd.Process.Pid)
		}
		if protocol == nil {
			protocol = release("other")
		}
	case "F11", "F12":
		if protocol = ready("A", "B"); protocol == nil {
			protocol = checkpoint("launched|")
		}
		if protocol == nil {
			protocol = release("A")
		}
	case "F14":
		if protocol = ready("A"); protocol == nil {
			protocol = release("A")
		}
		if protocol == nil {
			protocol = ready("B")
		}
		if protocol == nil {
			protocol = release("B")
		}
	case "F17", "F18", "F19":
		if protocol = ready("worker", "gate"); protocol == nil {
			protocol = release("gate")
		}
	case "F20":
		if protocol = ready("A"); protocol == nil {
			protocol = cmd.Process.Signal(signal)
		}
	case "F22":
		if protocol = ready("x", "gate"); protocol == nil {
			protocol = release("gate")
		}
		if protocol == nil {
			protocol = checkpoint("drained|")
		}
		if protocol == nil {
			protocol = release("x")
		}
	}
	if protocol != nil {
		t.Errorf("controller failed: %v; stdout=%q stderr=%q events=%v", protocol, out.Bytes(), diagnostic.Bytes(), ctrl.Events())
		return
	}
	select {
	case err = <-done:
		complete = true
	case <-ctx.Done():
		t.Errorf("orchestrator deadline: %v; stdout=%q stderr=%q events=%v", ctx.Err(), out.Bytes(), diagnostic.Bytes(), ctrl.Events())
		return
	}
	code := 0
	if err != nil {
		var exit *exec.ExitError
		if !errors.As(err, &exit) {
			t.Errorf("orchestrator wait: %v", err)
		} else {
			code = exit.ExitCode()
		}
	}
	if code != status {
		t.Errorf("exit status=%d want=%d; stderr=%q", code, status, diagnostic.Bytes())
	}
	want, err := ExpectedBytes(f.ExpectedRuntime.Stdout)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(out.Bytes(), want) {
		t.Errorf("raw stdout=%q want=%q", out.Bytes(), want)
	}
	if f.ExpectedRuntime.Stderr != nil && string(diagnostic.Bytes()) != *f.ExpectedRuntime.Stderr {
		t.Errorf("raw stderr=%q want=%q", diagnostic.Bytes(), *f.ExpectedRuntime.Stderr)
	}
	for _, word := range f.ExpectedRuntime.StderrContains {
		if !bytes.Contains(diagnostic.Bytes(), []byte(word)) {
			t.Errorf("stderr %q lacks %q", diagnostic.Bytes(), word)
		}
	}
	cleanupCtx, cleanupCancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cleanupCancel()
	cleanup := timing.Start("runtime.cleanup", t.Name())
	cleanupErr := ctrl.ObserveCleanup(cleanupCtx)
	cleanup(cleanupErr)
	if err := cleanupErr; err != nil {
		t.Errorf("runtime cleanup failed BEFORE fallback: %v; events=%v", err, ctrl.Events())
	}
	if f.ID == "F08" {
		count, err := os.ReadFile(filepath.Join(ctrl.Directory, "count"))
		if err != nil || string(count) != "2" {
			t.Errorf("rerun invocation counter=%q %v", count, err)
		}
	}
	t.Logf("native evidence: status=%d stdout_bytes=%d stderr_bytes=%d controller=%v", code, len(out.Bytes()), len(diagnostic.Bytes()), ctrl.Events())
}

func saveBuildEvidence(t testing.TB, project string, out, diag []byte, buildError error) {
	t.Helper()
	safe := regexp.MustCompile(`[^a-zA-Z0-9_.-]+`).ReplaceAllString(t.Name(), "_")
	dir := filepath.Join(RepoRoot(t), ".harness/phase1/cli", safe)
	if err := os.MkdirAll(dir, 0700); err != nil {
		t.Errorf("save build evidence: %v", err)
		return
	}
	for name, data := range map[string][]byte{"build.stdout": out, "build.stderr": diag} {
		if err := os.WriteFile(filepath.Join(dir, name), data, 0600); err != nil {
			t.Errorf("save evidence: %v", err)
		}
	}
	if err := filepath.WalkDir(project, func(path string, entry os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".l") {
			return nil
		}
		relative, err := filepath.Rel(project, path)
		if err != nil {
			return err
		}
		CopyFile(t, path, filepath.Join(dir, "sources", relative), 0600)
		return nil
	}); err != nil {
		t.Errorf("save source evidence: %v", err)
	}
	result := map[string]interface{}{"test": t.Name(), "runtime_exercised": false, "stage": "native-build", "command": []string{CLI(t), "build", "main.l"}}
	if buildError != nil {
		result["error"] = buildError.Error()
	}
	data, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		t.Errorf("encode build evidence: %v", err)
		return
	}
	if err := os.WriteFile(filepath.Join(dir, "build.json"), data, 0600); err != nil {
		t.Errorf("save build evidence: %v", err)
	}
	t.Logf("build evidence: %s", dir)
}

func saveRuntimeEvidence(t testing.TB, out, diag []byte, events []string) {
	t.Helper()
	safe := regexp.MustCompile(`[^a-zA-Z0-9_.-]+`).ReplaceAllString(t.Name(), "_")
	dir := filepath.Join(RepoRoot(t), ".harness/phase1/cli", safe)
	if err := os.MkdirAll(dir, 0700); err != nil {
		t.Errorf("runtime evidence: %v", err)
		return
	}
	for name, data := range map[string][]byte{"runtime.stdout": out, "runtime.stderr": diag} {
		if err := os.WriteFile(filepath.Join(dir, name), data, 0600); err != nil {
			t.Errorf("runtime evidence: %v", err)
		}
	}
	data, err := json.MarshalIndent(map[string]interface{}{"test": t.Name(), "runtime_exercised": true, "controller_events": events}, "", "  ")
	if err != nil {
		t.Errorf("encode runtime evidence: %v", err)
		return
	}
	if err := os.WriteFile(filepath.Join(dir, "runtime.json"), data, 0600); err != nil {
		t.Errorf("save runtime evidence: %v", err)
	}
}
