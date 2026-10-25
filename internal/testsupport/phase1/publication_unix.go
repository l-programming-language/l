//go:build darwin || linux

package phase1

import (
	"bytes"
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
	"testing"
	"time"
)

func RunPublication(t *testing.T, name string) {
	t.Helper()
	if name == "ImmutableCompanionGeneration" {
		runningGeneration(t)
		return
	}
	f := LoadFixture(t, "F01")
	if name == "BuildDoesNotExecute" {
		f = LoadFixture(t, "F08")
	}
	project := Materialize(t, f)
	if name == "BuildDoesNotExecute" {
		control := t.TempDir()
		binary, out, diag, err := buildProgram(t, project, []string{"L_PHASE1_DIRECTORY=" + control})
		if err != nil {
			t.Fatalf("compile/build prerequisite; runtime not exercised: %v stdout=%q stderr=%s", err, out, diag)
		}
		if _, err := os.Stat(filepath.Join(control, "count")); !os.IsNotExist(err) {
			t.Fatalf("build executed source App, count exists: %v", err)
		}
		if _, err := os.Stat(binary); err != nil {
			t.Fatal(err)
		}
		return
	}
	binary := RequireBuild(t, project)
	oldBinary, err := os.ReadFile(binary)
	if err != nil {
		t.Fatal(err)
	}
	switch name {
	case "RelocatedInstalledAssets":
		relocated := filepath.Join(t.TempDir(), "main")
		if err := os.Rename(binary, relocated); err != nil {
			t.Fatal(err)
		}
		if err := os.Rename(binary+".apps", relocated+".apps"); err != nil {
			t.Fatal(err)
		}
		if err := os.RemoveAll(project); err != nil {
			t.Fatal(err)
		}
		ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer cancel()
		out, diag, err := command(ctx, relocated, t.TempDir(), []string{"PATH=" + t.TempDir()})
		if err != nil || string(out) != "Hello\n" || len(diag) != 0 {
			t.Fatalf("relocated program depends on source/CWD: %v %q %q", err, out, diag)
		}
	case "FailedPublicationPreservesExisting":
		existing := companions(t, binary)
		before := map[string][]byte{}
		for _, app := range existing {
			before[app], err = os.ReadFile(app)
			if err != nil {
				t.Fatal(err)
			}
		}
		if err := os.WriteFile(filepath.Join(project, "main.l"), []byte("Unrecognized broken source.\n"), 0600); err != nil {
			t.Fatal(err)
		}
		_, out, diag, err := buildProgram(t, project, nil)
		if err == nil || len(out) != 0 || len(diag) == 0 {
			t.Fatalf("failed compilation state: %v %q %q", err, out, diag)
		}
		after, err := os.ReadFile(binary)
		if err != nil || !bytes.Equal(oldBinary, after) {
			t.Fatalf("failed publication changed previous orchestrator: %v", err)
		}
		for path, want := range before {
			got, err := os.ReadFile(path)
			if err != nil || !bytes.Equal(want, got) {
				t.Errorf("failed publication changed %s: %v", path, err)
			}
		}
		if got := companions(t, binary); len(got) != len(existing) {
			t.Errorf("failed publication added generation")
		}
		ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer cancel()
		out, diag, err = command(ctx, binary, t.TempDir(), nil)
		if err != nil || string(out) != "Hello\n" || len(diag) != 0 {
			t.Fatalf("previous executable stopped working: %v %q %q", err, out, diag)
		}
	default:
		t.Fatalf("unknown publication assertion %q", name)
	}
}

// A live old orchestrator remains behind A's barrier while a new generation is
// published. Its later launch of B must use its original companion generation.
func runningGeneration(t *testing.T) {
	f := LoadFixture(t, "F03")
	project := Materialize(t, f)
	source, err := os.ReadFile(filepath.Join(project, "main.l"))
	if err != nil {
		t.Fatal(err)
	}
	original := strings.Replace(string(source), "Start Second.\n", "", 1)
	original = strings.Replace(original, "Await First.\n", "Await First.\nStart Second.\n", 1)
	if err := os.WriteFile(filepath.Join(project, "main.l"), []byte(original), 0600); err != nil {
		t.Fatal(err)
	}
	binary := RequireBuild(t, project)
	old := map[string][]byte{}
	for _, app := range companions(t, binary) {
		old[app], err = os.ReadFile(app)
		if err != nil {
			t.Fatal(err)
		}
	}
	ctrl, err := NewController("A", "B")
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if leaks := ctrl.Leaks(); len(leaks) != 0 {
			t.Errorf("runtime leaked before fallback: %v", leaks)
			ctrl.Fallback()
		}
		if err := ctrl.Close(); err != nil {
			t.Errorf("close runtime controller: %v", err)
		}
	}()
	budget := publicationRuntimeBudget{remaining: 20 * time.Second, now: time.Now}
	ctx := budget.resume()
	defer budget.close()
	cmd := exec.Command(binary)
	cmd.Dir = t.TempDir()
	cmd.Env = append(os.Environ(), "L_PHASE1_SOCKET="+ctrl.Socket, "L_PHASE1_DIRECTORY="+ctrl.Directory)
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	cmd.WaitDelay = 2 * time.Second
	out, diag := &Stream{}, &Stream{}
	cmd.Stdout, cmd.Stderr = out, diag
	if err := cmd.Start(); err != nil {
		t.Fatal(err)
	}
	done := make(chan error, 1)
	go func() { done <- cmd.Wait() }()
	waited := false
	defer func() {
		if !waited {
			_ = syscall.Kill(-cmd.Process.Pid, syscall.SIGKILL) //nolint:errcheck // Fallback may race process exit; the owned Wait below still must complete.
			select {
			case <-done:
			case <-time.After(2 * time.Second):
				t.Error("old orchestrator fallback did not reap")
			}
		}
	}()
	if err := ctrl.WaitReady(ctx, "A"); err != nil {
		t.Fatal(err)
	}
	if err := ctrl.CheckAlive("A", cmd.Process.Pid); err != nil {
		t.Fatal(err)
	}
	if err := out.WaitContains(ctx, "launched|"); err != nil {
		t.Fatal(err)
	}
	updated := strings.Replace(original, "`B`", "`NEW`", 1)
	if err := os.WriteFile(filepath.Join(project, "main.l"), []byte(updated), 0600); err != nil {
		t.Fatal(err)
	}
	if err := ctrl.CheckAlive("A", cmd.Process.Pid); err != nil {
		t.Fatalf("old child was not alive before publication: %v", err)
	}
	select {
	case err := <-done:
		waited = true
		t.Fatalf("old orchestrator exited before publication: %v", err)
	default:
	}
	if err := budget.pause(); err != nil {
		t.Fatalf("old orchestrator runtime allowance exhausted before publication: %v", err)
	}
	// RequireBuild owns the separate build-prerequisite watchdog. Only this
	// compilation interval is excluded from the cumulative runtime allowance.
	RequireBuild(t, project)
	ctx = budget.resume()
	select {
	case err := <-done:
		waited = true
		t.Fatalf("old orchestrator exited during publication: %v", err)
	default:
	}
	if err := ctrl.CheckAlive("A", cmd.Process.Pid); err != nil {
		t.Fatalf("publication terminated active old child: %v", err)
	}
	for path, want := range old {
		got, err := os.ReadFile(path)
		if err != nil || !bytes.Equal(want, got) {
			t.Errorf("later publication mutated old companion %s: %v", path, err)
		}
	}
	manifests, err := filepath.Glob(filepath.Join(binary+".apps", "generations", "*", "manifest.json"))
	if err != nil || len(manifests) != 2 {
		t.Fatalf("old/new generation inventory: %v %v", manifests, err)
	}
	if err := ctrl.CheckAlive("A", cmd.Process.Pid); err != nil {
		t.Fatalf("publication terminated active old child: %v", err)
	}
	if err := ctrl.Release("A"); err != nil {
		t.Fatal(err)
	}
	if err := ctrl.WaitReady(ctx, "B"); err != nil {
		t.Fatalf("old orchestrator's later launch lost old companion binding: %v", err)
	}
	if err := ctrl.Release("B"); err != nil {
		t.Fatal(err)
	}
	select {
	case err = <-done:
		waited = true
	case <-ctx.Done():
		t.Fatal("old orchestrator did not complete after publication")
	}
	if err != nil || string(out.Bytes()) != "launched|A\nB\n" || len(diag.Bytes()) != 0 {
		t.Errorf("old invocation result after publication: %v %q %q", err, out.Bytes(), diag.Bytes())
	}
	cleanupCtx, cleanupCancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cleanupCancel()
	if err := ctrl.ObserveCleanup(cleanupCtx); err != nil {
		t.Errorf("runtime cleanup failed BEFORE fallback: %v", err)
	}
}
