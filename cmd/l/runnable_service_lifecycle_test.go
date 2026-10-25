//go:build darwin || linux

package main

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"strings"
	"syscall"
	"testing"
	"time"

	"github.com/l-programming-language/l/internal/testsupport/phase1"
)

// Reuse the unchanged test-owned readiness/lifetime-lock protocol. Only L
// programs below are new: no supervisor implementation or executable is supplied
// in place of generated App code.
func runnableLifecycleBuild(t *testing.T, source string, apps []string) runnableBuildResult {
	t.Helper()
	project, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"gate.l", "stop_gate.l", "raw.l", "failure.l", "counter.l", "actions.l"} {
		phase1.CopyFile(t, filepath.Join(phase1.RepoRoot(t), "testdata", "phase1", "companions", name), filepath.Join(project, name), 0600)
	}
	if err := os.WriteFile(filepath.Join(project, "main.l"), []byte(source), 0600); err != nil {
		t.Fatal(err)
	}
	return runnableBuildProject(t, project, apps)
}

func TestRunnableServiceRunningGenerationSurvivesPublication(t *testing.T) {
	source := runnableGatePrefix + "Create \"Later\" as a New Basic App.\nThe App Prints `B`.\nCreate \"Host\" as a New Basic App.\nThe App Prints `H`.\nRun Worker.\nRun Host.\nShow `launched|`.\nAwait Worker.\nShow Worker's Output.\nRun Later.\nAwait Later.\nShow Later's Output.\nAwait Host.\nShow Host's Output.\n"
	old := runnableLifecycleBuild(t, source, []string{"Worker", "Later", "Host"})
	originalFiles := runnableGenerationSnapshot(t, filepath.Dir(old.manifestPath))
	controller, err := phase1.NewController("A", "B", "H")
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if leaks := controller.Leaks(); len(leaks) != 0 {
			t.Errorf("publication leaked old generation children before fallback: %v", leaks)
			controller.Fallback()
		}
		if err := controller.Close(); err != nil {
			t.Error(err)
		}
	}()
	// Preserve the retained publication observer's cumulative 20-second runtime
	// allowance. Only the separately bounded compilation interval is excluded;
	// resuming does not grant a new full runtime allowance.
	remaining := 20 * time.Second
	started := time.Now()
	ctx, cancel := context.WithTimeout(context.Background(), remaining)
	defer cancel()
	command := exec.Command(old.binary)
	command.Dir = t.TempDir()
	command.Env = append(os.Environ(), "PATH="+t.TempDir(), "L_PHASE1_SOCKET="+controller.Socket, "L_PHASE1_DIRECTORY="+controller.Directory)
	command.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	command.WaitDelay = 2 * time.Second
	stdout, stderr := &phase1.Stream{}, &phase1.Stream{}
	command.Stdout, command.Stderr = stdout, stderr
	if err := command.Start(); err != nil {
		t.Fatal(err)
	}
	done := make(chan error, 1)
	go func() { done <- command.Wait() }()
	complete := false
	defer func() {
		if !complete {
			if leaks := controller.Leaks(); len(leaks) != 0 {
				t.Errorf("owned old-generation lifetimes before fallback: %v", leaks)
				controller.Fallback()
			}
			if err := syscall.Kill(-command.Process.Pid, syscall.SIGKILL); err != nil && err != syscall.ESRCH {
				t.Error(err)
			}
			select {
			case <-done:
			case <-time.After(2 * time.Second):
				t.Error("old orchestrator was not reaped")
			}
		}
	}()
	if err := controller.WaitReady(ctx, "A", "H"); err != nil {
		t.Fatal(err)
	}
	if err := stdout.WaitContains(ctx, "launched|"); err != nil {
		t.Fatal(err)
	}
	checkAlive := func() {
		for _, label := range []string{"A", "H"} {
			if err := controller.CheckAlive(label, command.Process.Pid); err != nil {
				t.Fatal(err)
			}
		}
		select {
		case err := <-done:
			complete = true
			t.Fatalf("old orchestrator exited across publication: %v", err)
		default:
		}
	}
	checkAlive()
	updated := strings.Replace(source, "Prints `B`", "Prints `NEW`", 1)
	if updated == source {
		t.Fatal("publication fixture failed to change later App entry")
	}
	if err := os.WriteFile(filepath.Join(old.project, "main.l"), []byte(updated), 0600); err != nil {
		t.Fatal(err)
	}
	remaining -= time.Since(started)
	if ctx.Err() != nil || remaining <= 0 {
		t.Fatal("old-generation runtime allowance exhausted before rebuild")
	}
	cancel()
	newGeneration := runnableBuildProject(t, old.project, []string{"Worker", "Later", "Host"})
	ctx, cancel = context.WithTimeout(context.Background(), remaining)
	defer cancel()
	checkAlive()
	if newGeneration.manifestPath == old.manifestPath || !reflect.DeepEqual(originalFiles, runnableGenerationSnapshot(t, filepath.Dir(old.manifestPath))) {
		t.Fatal("replacement reused or mutated the running immutable generation")
	}
	if err := controller.Release("A"); err != nil {
		t.Fatal(err)
	}
	// The new generation's corresponding App says NEW. The old running
	// orchestrator must still launch its old B companion after publication.
	if err := controller.WaitReady(ctx, "B"); err != nil {
		t.Fatalf("late App launch lost old-generation binding: %v", err)
	}
	for _, label := range []string{"B", "H"} {
		if err := controller.Release(label); err != nil {
			t.Fatal(err)
		}
	}
	select {
	case err = <-done:
		complete = true
	case <-ctx.Done():
		t.Fatal("old generation did not finish within remaining runtime allowance")
	}
	if err != nil || !bytes.Equal(stdout.Bytes(), []byte("launched|A\nB\nH\n")) || len(stderr.Bytes()) != 0 {
		t.Fatalf("old-generation result after replacement: %v %q %q", err, stdout.Bytes(), stderr.Bytes())
	}
	if err := controller.ObserveCleanup(ctx); err != nil {
		t.Error(err)
	}
}

type runnableGateAction func(context.Context, *phase1.Controller, *exec.Cmd, *phase1.Stream)

func runnableGated(t *testing.T, built runnableBuildResult, labels []string, action runnableGateAction) (int, []byte, []byte) {
	t.Helper()
	controller, err := phase1.NewController(labels...)
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if leaks := controller.Leaks(); len(leaks) != 0 {
			t.Errorf("Service lifecycle leaked owned children before fallback: %v", leaks)
			controller.Fallback()
		}
		if err := controller.Close(); err != nil {
			t.Errorf("close Service readiness controller: %v", err)
		}
	}()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	command := exec.Command(built.binary)
	command.Dir = t.TempDir()
	command.Env = append(os.Environ(), "PATH="+t.TempDir(), "L_PHASE1_SOCKET="+controller.Socket, "L_PHASE1_DIRECTORY="+controller.Directory)
	command.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	command.WaitDelay = 2 * time.Second
	stdout, stderr := &phase1.Stream{}, &phase1.Stream{}
	command.Stdout, command.Stderr = stdout, stderr
	if err := command.Start(); err != nil {
		t.Fatal(err)
	}
	done := make(chan error, 1)
	go func() { done <- command.Wait() }()
	complete := false
	defer func() {
		if !complete {
			if leaks := controller.Leaks(); len(leaks) != 0 {
				t.Errorf("owned lifetimes before fallback termination: %v", leaks)
				controller.Fallback()
			}
			if err := syscall.Kill(-command.Process.Pid, syscall.SIGKILL); err != nil && err != syscall.ESRCH {
				t.Errorf("fallback signal: %v", err)
			}
			select {
			case <-done:
			case <-time.After(2 * time.Second):
				t.Error("Service orchestrator not reaped after fallback")
			}
		}
	}()
	if err := controller.WaitReady(ctx, labels...); err != nil {
		t.Fatal(err)
	}
	for _, label := range labels {
		if err := controller.CheckAlive(label, command.Process.Pid); err != nil {
			t.Fatal(err)
		}
	}
	seen := make(map[int]bool)
	for _, peer := range controller.Peers() {
		if seen[peer.PID] {
			t.Fatal("independent App declarations share a process")
		}
		seen[peer.PID] = true
	}
	action(ctx, controller, command, stdout)
	select {
	case err = <-done:
		complete = true
	case <-ctx.Done():
		t.Fatalf("Service runtime watchdog: %v; stdout=%q stderr=%q", ctx.Err(), stdout.Bytes(), stderr.Bytes())
	}
	if cleanup := controller.ObserveCleanup(ctx); cleanup != nil {
		t.Errorf("Service runtime failed cleanup before controller closure: %v", cleanup)
	}
	status := 0
	if err != nil {
		if exit, ok := err.(*exec.ExitError); ok {
			status = exit.ExitCode()
		} else {
			t.Fatalf("native Service Wait failed: %v", err)
		}
	}
	return status, stdout.Bytes(), stderr.Bytes()
}

const runnableGatePrefix = "Use ProbeIO's Print from `gate.l` as \"Print\".\n" +
	"Create \"Worker\" as a New Basic App.\nThe App Prints `A`.\n"

func TestRunnableServiceControlledLifecycle(t *testing.T) {
	t.Run("Apps launch without joining", func(t *testing.T) {
		source := runnableGatePrefix + "Create \"Other\" as a New Basic App.\nThe App Prints `B`.\nRun Worker.\nRun Other.\nShow `launched|`.\nAwait Worker.\nShow Worker's Output.\nAwait Other.\nShow Other's Output.\n"
		built := runnableLifecycleBuild(t, source, []string{"Worker", "Other"})
		status, stdout, stderr := runnableGated(t, built, []string{"A", "B"}, func(ctx context.Context, controller *phase1.Controller, _ *exec.Cmd, out *phase1.Stream) {
			if err := out.WaitContains(ctx, "launched|"); err != nil {
				t.Fatal(err)
			}
			// Neither ready child has been released. Reaching the checkpoint
			// proves both Run calls returned while both invocations are live.
			for _, label := range []string{"B", "A"} {
				if err := controller.Release(label); err != nil {
					t.Fatal(err)
				}
			}
		})
		if status != 0 || string(stdout) != "launched|A\nB\n" || len(stderr) != 0 {
			t.Fatalf("mixed process overlap = %d/%q/%q", status, stdout, stderr)
		}
	})
	t.Run("Stop owns one invocation and preserves other App", func(t *testing.T) {
		source := runnableGatePrefix + "Create \"Other\" as a New Basic App.\nThe App Prints `B`.\nCreate \"Gate\" as a New Basic App.\nThe App Prints `G`.\nRun Worker.\nRun Other.\nRun Gate.\nAwait Gate.\nStop Worker.\nShow `stop-returned|`.\nAwait Worker.\nStop Worker.\nIf Worker has an Error, Show `stopped|`.\nAwait Other.\nShow Other's Output.\n"
		built := runnableLifecycleBuild(t, source, []string{"Worker", "Other", "Gate"})
		status, stdout, stderr := runnableGated(t, built, []string{"A", "B", "G"}, func(ctx context.Context, controller *phase1.Controller, command *exec.Cmd, out *phase1.Stream) {
			if err := controller.Release("G"); err != nil {
				t.Fatal(err)
			}
			if err := out.WaitContains(ctx, "stopped|"); err != nil {
				t.Fatal(err)
			}
			if err := controller.WaitDisconnected(ctx, "A"); err != nil {
				t.Fatal(err)
			}
			if err := controller.CheckAlive("B", command.Process.Pid); err != nil {
				t.Fatal("Stop killed the wrong Service: ", err)
			}
			if err := controller.Release("B"); err != nil {
				t.Fatal(err)
			}
		})
		if status != 0 || string(stdout) != "stop-returned|stopped|B\n" || len(stderr) != 0 {
			t.Fatalf("Service Stop isolation = %d/%q/%q", status, stdout, stderr)
		}
	})
	for _, tc := range []struct {
		name, tail, output, diagnostic string
		status                         int
	}{
		{"normal return cleans unawaited", "Show `finished`.\n", "finished", "", 0},
		{"bare Exit cleans unawaited", "Exit.\nShow `unreachable`.\n", "", "", 0},
		{"error Exit cleans unawaited", "Exit with an error of `requested failure`.\n", "", "requested failure", 1},
		{"Output before completion is misuse", "Show Worker's Output.\n", "", "Output is unavailable before completed invocation capture", 1},
		{"Build while outstanding is misuse", "Build Worker.\n", "", "Build conflicts with an outstanding invocation", 1},
	} {
		t.Run(tc.name, func(t *testing.T) {
			source := runnableGatePrefix + "Create \"Gate\" as a New Basic App.\nThe App Prints `G`.\nRun Worker.\nRun Gate.\nAwait Gate.\n" + tc.tail
			built := runnableLifecycleBuild(t, source, []string{"Worker", "Gate"})
			status, stdout, stderr := runnableGated(t, built, []string{"A", "G"}, func(_ context.Context, controller *phase1.Controller, _ *exec.Cmd, _ *phase1.Stream) {
				if err := controller.Release("G"); err != nil {
					t.Fatal(err)
				}
			})
			if status != tc.status || string(stdout) != tc.output || (tc.diagnostic == "" && len(stderr) != 0) || !strings.Contains(string(stderr), tc.diagnostic) {
				t.Fatalf("controlled lifecycle = %d/%q/%q; want %d/%q/containing %q", status, stdout, stderr, tc.status, tc.output, tc.diagnostic)
			}
		})
	}
	for _, signal := range []syscall.Signal{syscall.SIGINT, syscall.SIGTERM} {
		t.Run(fmt.Sprintf("signal %s cleans pending Await", signal), func(t *testing.T) {
			built := runnableLifecycleBuild(t, runnableGatePrefix+"Run Worker.\nAwait Worker.\nShow `unreachable`.\n", []string{"Worker"})
			status, stdout, stderr := runnableGated(t, built, []string{"A"}, func(_ context.Context, _ *phase1.Controller, command *exec.Cmd, _ *phase1.Stream) {
				if err := command.Process.Signal(signal); err != nil {
					t.Fatal(err)
				}
			})
			if status != 128+int(signal) || len(stdout) != 0 || len(stderr) != 0 {
				t.Fatalf("signal Service cleanup = %d/%q/%q", status, stdout, stderr)
			}
		})
	}
}

func TestRunnableServiceInvocationResults(t *testing.T) {
	t.Run("Build validates without creating an invocation", func(t *testing.T) {
		built := runnableLifecycleBuild(t, "Create \"Worker\" as a New Basic App.\nBuild Worker.\nAwait Worker.\n", []string{"Worker"})
		status, stdout, stderr := runnableRun(t, built.binary, t.TempDir(), nil)
		if status != 1 || len(stdout) != 0 || !strings.Contains(string(stderr), "Await requires an invocation") || !strings.Contains(string(stderr), "main.l:3:1") {
			t.Fatalf("Build created an invocation or compiled at runtime: %d/%q/%q", status, stdout, stderr)
		}
	})
	t.Run("snapshots retain earlier distinct invocation bytes", func(t *testing.T) {
		source := "Use ProbeIO's Print from `counter.l` as \"Print\".\nCreate \"Worker\" as a New Basic App.\nThe App Prints `value`.\nRun Worker.\nAwait Worker.\nDefine \"saved\" as Worker's Output.\nRun Worker.\nAwait Worker.\nShow saved.\nShow Worker's Output.\n"
		built := runnableLifecycleBuild(t, source, []string{"Worker"})
		state := t.TempDir()
		status, stdout, stderr := runnableRun(t, built.binary, t.TempDir(), []string{"L_PHASE1_DIRECTORY=" + state})
		if status != 0 || string(stdout) != "value#1\nvalue#2\n" || len(stderr) != 0 {
			t.Fatalf("Service snapshot was refreshed/reused across rerun: %d/%q/%q", status, stdout, stderr)
		}
		count, err := os.ReadFile(filepath.Join(state, "count"))
		if err != nil || string(count) != "2" {
			t.Fatalf("expected two real entry invocations, got %q: %v", count, err)
		}
	})
	t.Run("nested and cached joins keep independent continuations", func(t *testing.T) {
		source := "    Create \"Worker\" as a New Public Basic App.\nAwait Other.\nShow `worker|`.\n\n    Create \"Other\" as a New Public Basic App.\nAwait Other.\nShow `other|`.\n\nRun Worker.\nRun Other.\nAwait Worker.\nAwait Other.\nShow `done`.\n"
		built := runnableLifecycleBuild(t, source, []string{"Worker", "Other"})
		status, stdout, stderr := runnableRun(t, built.binary, t.TempDir(), nil)
		if status != 0 || string(stdout) != "other|worker|done" || len(stderr) != 0 {
			t.Fatalf("nested/cached Service continuation = %d/%q/%q", status, stdout, stderr)
		}
	})
	t.Run("failed launch stays awaitable", func(t *testing.T) {
		source := "Create \"Worker\" as a New Basic App.\nRun Worker.\nAwait Worker.\nIf Worker has an Error, Show `failed|`.\nShow Worker's Output.\nAwait Worker.\nShow `joined`.\n"
		built := runnableLifecycleBuild(t, source, []string{"Worker"})
		if len(built.manifest.Apps) != 1 {
			t.Fatal("missing App companion before launch-failure setup")
		}
		if err := os.Remove(filepath.Join(filepath.Dir(built.manifestPath), built.manifest.Apps[0].Executable)); err != nil {
			t.Fatal(err)
		}
		status, stdout, stderr := runnableRun(t, built.binary, t.TempDir(), nil)
		if status != 0 || string(stdout) != "failed|joined" || len(stderr) != 0 {
			t.Fatalf("failed launch could not be joined: %d/%q/%q", status, stdout, stderr)
		}
	})
	for _, tc := range []struct{ name, dictionary, output string }{
		{"raw stdout excludes stderr", "raw.l", "Snow 雪 \x00\n\n"},
		{"nonzero child is stored Error", "failure.l", "Snow 雪failed|"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			source := "Use ProbeIO's Print from `" + tc.dictionary + "` as \"Print\".\nCreate \"Worker\" as a New Basic App.\nThe App Prints `Snow 雪`.\nRun Worker.\nAwait Worker.\nShow Worker's Output.\nIf Worker has an Error, Show `failed|`.\n"
			built := runnableLifecycleBuild(t, source, []string{"Worker"})
			status, stdout, stderr := runnableRun(t, built.binary, t.TempDir(), nil)
			if status != 0 || string(stdout) != tc.output || len(stderr) != 0 {
				t.Fatalf("stored Service result = %d/%q/%q; want %q", status, stdout, stderr, tc.output)
			}
		})
	}
	t.Run("Stop before Run is source misuse", func(t *testing.T) {
		built := runnableLifecycleBuild(t, "Create \"Worker\" as a New Basic App.\nStop Worker.\n", []string{"Worker"})
		status, stdout, stderr := runnableRun(t, built.binary, t.TempDir(), nil)
		if status != 1 || len(stdout) != 0 || !strings.Contains(string(stderr), "Stop requires an invocation") || !strings.Contains(string(stderr), "main.l:2:1") {
			t.Fatalf("Stop-before-Run = %d/%q/%q", status, stdout, stderr)
		}
	})
	t.Run("continuation once and own rerun rejected", func(t *testing.T) {
		for _, rerun := range []bool{false, true} {
			t.Run(fmt.Sprint(rerun), func(t *testing.T) {
				body := "Show `continued|`.\n"
				if rerun {
					body += "Run Worker.\n"
				}
				source := "    Create \"Worker\" as a New Public Basic App.\n" + body + "\nRun Worker.\nAwait Worker.\nAwait Worker.\nShow `done`.\n"
				built := runnableLifecycleBuild(t, source, []string{"Worker"})
				status, stdout, stderr := runnableRun(t, built.binary, t.TempDir(), nil)
				if !rerun && (status != 0 || string(stdout) != "continued|done" || len(stderr) != 0) {
					t.Fatalf("once-only Service continuation = %d/%q/%q", status, stdout, stderr)
				}
				if rerun && (status != 1 || string(stdout) != "continued|" || !strings.Contains(string(stderr), "outstanding invocation")) {
					t.Fatalf("own continuation rerun = %d/%q/%q", status, stdout, stderr)
				}
			})
		}
	})
}
