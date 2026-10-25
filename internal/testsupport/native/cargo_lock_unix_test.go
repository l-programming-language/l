//go:build darwin || linux

package native

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
	"testing"
	"time"
)

func TestCargoCacheLockLifecycle(t *testing.T) {
	target := filepath.Join(t.TempDir(), "target")
	canceled, cancel := context.WithCancel(context.Background())
	cancel()
	if release, err := acquireBuildCache(canceled, target); !errors.Is(err, context.Canceled) || release != nil {
		t.Fatalf("pre-canceled lock: %v", err)
	}
	if _, err := os.Stat(target); !os.IsNotExist(err) {
		t.Fatalf("pre-canceled acquire created target: %v", err)
	}
	first, err := acquireBuildCache(context.Background(), target)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := first(); err != nil {
			t.Errorf("release test lock: %v", err)
		}
	})
	ctx, deadline := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer deadline()
	if release, err := acquireBuildCache(ctx, target); !errors.Is(err, context.DeadlineExceeded) || release != nil {
		t.Fatalf("contention deadline: %v", err)
	}
	if err := first(); err != nil {
		t.Fatal(err)
	}
	if err := first(); err != nil {
		t.Fatalf("idempotent release: %v", err)
	}
	path := filepath.Join(target, ".l-parity-publication.lock")
	before, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	second, err := acquireBuildCache(context.Background(), target)
	if err != nil {
		t.Fatal(err)
	}
	if err := second(); err != nil {
		t.Fatal(err)
	}
	after, err := os.Stat(path)
	if err != nil || !os.SameFile(before, after) {
		t.Fatalf("handoff replaced lock inode: %v", err)
	}
}

func TestCargoCacheLockRejectsUnsafePaths(t *testing.T) {
	for _, kind := range []string{"target-symlink", "lock-symlink", "directory", "fifo"} {
		t.Run(kind, func(t *testing.T) {
			root := t.TempDir()
			target := filepath.Join(root, "target")
			if err := os.Mkdir(target, 0700); err != nil {
				t.Fatal(err)
			}
			lock := filepath.Join(target, ".l-parity-publication.lock")
			switch kind {
			case "target-symlink":
				linked := filepath.Join(root, "linked")
				if err := os.Symlink(target, linked); err != nil {
					t.Fatal(err)
				}
				target = linked
			case "lock-symlink":
				original := filepath.Join(root, "original")
				if err := os.WriteFile(original, []byte("unaltered"), 0600); err != nil {
					t.Fatal(err)
				}
				if err := os.Symlink(original, lock); err != nil {
					t.Fatal(err)
				}
			case "directory":
				if err := os.Mkdir(lock, 0700); err != nil {
					t.Fatal(err)
				}
			case "fifo":
				if err := syscall.Mkfifo(lock, 0600); err != nil {
					t.Fatal(err)
				}
			}
			ctx, cancel := context.WithTimeout(context.Background(), time.Second)
			defer cancel()
			if release, err := acquireBuildCache(ctx, target); err == nil {
				if err := release(); err != nil {
					t.Errorf("release unexpectedly acquired lock: %v", err)
				}
				t.Fatal("unsafe cache ownership path accepted")
			}
			if kind == "lock-symlink" {
				if data, err := os.ReadFile(filepath.Join(root, "original")); err != nil || string(data) != "unaltered" {
					t.Fatalf("symlink referent changed: %q %v", data, err)
				}
			}
		})
	}
}

func TestCargoCacheLockOwnerProcess(t *testing.T) {
	if os.Getenv("L_HELPER_LOCK_PROCESS") != "yes" {
		return
	}
	release, err := acquireBuildCache(context.Background(), os.Getenv("L_HELPER_TARGET"))
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := release(); err != nil {
			t.Errorf("close test resource: %v", err)
		}
	}()
	if err := os.WriteFile(os.Getenv("L_HELPER_LOCK_READY"), []byte("owned"), 0600); err != nil {
		t.Fatal(err)
	}
	// Parent intentionally kills this owned leaf; this watchdog bounds failure
	// cleanup without introducing a subprocess whose lifetime could escape.
	time.Sleep(5 * time.Second)
}

func TestCargoCacheLockOwnerDeathReleases(t *testing.T) {
	root := t.TempDir()
	target := filepath.Join(root, "target")
	ready := filepath.Join(root, "ready")
	cmd := exec.Command(os.Args[0], "-test.run=^TestCargoCacheLockOwnerProcess$")
	cmd.Env = append(os.Environ(), "L_HELPER_LOCK_PROCESS=yes", "L_HELPER_TARGET="+target, "L_HELPER_LOCK_READY="+ready)
	var output bytes.Buffer
	cmd.Stdout, cmd.Stderr = &output, &output
	if err := cmd.Start(); err != nil {
		t.Fatal(err)
	}
	waited := false
	t.Cleanup(func() {
		if !waited {
			_ = cmd.Process.Kill() //nolint:errcheck // Fallback may race process exit; lifetime and Wait assertions remain authoritative.
			_ = cmd.Wait()         //nolint:errcheck // Reap the fallback-killed child; a signal exit is expected here.
		}
	})
	until := time.Now().Add(3 * time.Second)
	for {
		if _, err := os.Stat(ready); err == nil {
			break
		}
		if time.Now().After(until) {
			t.Fatal("owner never acquired lock")
		}
		time.Sleep(time.Millisecond)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	if release, err := acquireBuildCache(ctx, target); !errors.Is(err, context.DeadlineExceeded) || release != nil {
		cancel()
		t.Fatalf("living child did not own cache: %v", err)
	}
	cancel()
	if err := cmd.Process.Kill(); err != nil {
		t.Fatal(err)
	}
	err := cmd.Wait()
	waited = true
	var exit *exec.ExitError
	if !errors.As(err, &exit) {
		t.Fatalf("owned process kill: %v %s", err, output.Bytes())
	}
	ctx, cancel = context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	release, err := acquireBuildCache(ctx, target)
	if err != nil {
		t.Fatalf("dead owner leaked lock: %v", err)
	}
	if err := release(); err != nil {
		t.Fatal(err)
	}
}

type helperBuildOutcome struct {
	Error    string
	Deadline bool
}

// A separate test process crosses the real Build entry point. The controlled
// native command reports which staged source it actually read.
func TestCargoHelperBuildProcess(t *testing.T) {
	if os.Getenv("L_HELPER_BUILD_PROCESS") != "yes" {
		return
	}
	timeout, err := time.ParseDuration(os.Getenv("L_HELPER_TIMEOUT"))
	if err != nil {
		t.Fatal(err)
	}
	r := Build(t, os.Getenv("L_HELPER_WORKSPACE"), BuildOptions{TargetDir: os.Getenv("L_HELPER_TARGET"), Timeout: timeout})
	outcome := helperBuildOutcome{Deadline: errors.Is(r.Err, context.DeadlineExceeded)}
	if r.Err != nil {
		outcome.Error = r.Err.Error()
	}
	data, err := json.Marshal(outcome)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(os.Getenv("L_HELPER_ROOT"), "result.json"), data, 0600); err != nil {
		t.Fatal(err)
	}
}

func TestCargoBuildSerializesSharedInputsAcrossProcesses(t *testing.T) {
	f := newHelperCargoFixture(t)
	other := filepath.Join(f.root, "other")
	if err := os.MkdirAll(filepath.Join(other, "src"), 0700); err != nil {
		t.Fatal(err)
	}
	manifest, err := os.ReadFile(filepath.Join(f.project, "Cargo.toml"))
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(other, "Cargo.toml"), manifest, 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(other, "src/main.rs"), []byte("independent source B"), 0600); err != nil {
		t.Fatal(err)
	}
	firstRoot, secondRoot := filepath.Join(f.root, "first"), filepath.Join(f.root, "second")
	for _, dir := range []string{firstRoot, secondRoot} {
		if err := os.Mkdir(dir, 0700); err != nil {
			t.Fatal(err)
		}
	}
	release := filepath.Join(firstRoot, "release")
	// Release the causal barrier before any child watchdog/cleanup. Build owns
	// and reaps its native group; each parent Cmd has exactly one Wait below.
	t.Cleanup(func() {
		if err := os.WriteFile(release, []byte("release"), 0600); err != nil {
			t.Errorf("release causal barrier: %v", err)
		}
	})
	start := func(project, root, barrier, timeout string) (<-chan error, *bytes.Buffer) {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		t.Cleanup(cancel)
		cmd := exec.CommandContext(ctx, os.Args[0], "-test.run=^TestCargoHelperBuildProcess$", "-test.timeout=8s")
		replacements := map[string]string{"L_HELPER_BUILD_PROCESS": "yes", "L_HELPER_WORKSPACE": project, "L_HELPER_ROOT": root, "L_HELPER_TARGET": f.target, "L_HELPER_RELEASE": barrier, "L_HELPER_TIMEOUT": timeout}
		for _, entry := range os.Environ() {
			key, _, _ := strings.Cut(entry, "=")
			if _, replaced := replacements[key]; !replaced {
				cmd.Env = append(cmd.Env, entry)
			}
		}
		for key, value := range replacements {
			cmd.Env = append(cmd.Env, key+"="+value)
		}
		out := &bytes.Buffer{}
		cmd.Stdout, cmd.Stderr = out, out
		if err := cmd.Start(); err != nil {
			t.Fatal(err)
		}
		done := make(chan error, 1)
		go func() { done <- cmd.Wait(); close(done) }()
		t.Cleanup(func() {
			if err := os.WriteFile(release, []byte("release"), 0600); err != nil {
				t.Errorf("release causal barrier before child cleanup: %v", err)
			}
			select {
			case <-done:
			case <-time.After(7 * time.Second):
				cancel()
				<-done
			}
		})
		return done, out
	}
	first, firstOutput := start(f.project, firstRoot, release, "4s")
	until := time.Now().Add(5 * time.Second)
	for {
		if _, err := os.Stat(filepath.Join(firstRoot, "build-entered")); err == nil {
			break
		}
		if time.Now().After(until) {
			t.Fatal("first Cargo did not reach its readiness barrier")
		}
		time.Sleep(time.Millisecond)
	}
	second, secondOutput := start(other, secondRoot, "", "200ms")
	if err := <-second; err != nil {
		t.Fatalf("second observer process: %v\n%s", err, secondOutput.Bytes())
	}
	data, err := os.ReadFile(filepath.Join(secondRoot, "result.json"))
	if err != nil {
		t.Fatal(err)
	}
	var result helperBuildOutcome
	if err = json.Unmarshal(data, &result); err != nil {
		t.Fatal(err)
	}
	if !result.Deadline {
		t.Errorf("contending build did not consume its existing deadline while first build owns inputs: %+v", result)
	}
	if _, err := os.Stat(filepath.Join(secondRoot, "build-entered")); !os.IsNotExist(err) {
		t.Errorf("contending Cargo entered before ownership release: %v", err)
	}
	if err := os.WriteFile(release, []byte("release"), 0600); err != nil {
		t.Fatal(err)
	}
	if err := <-first; err != nil {
		t.Fatalf("first observer process: %v\n%s", err, firstOutput.Bytes())
	}
	if data, err := os.ReadFile(filepath.Join(firstRoot, "observed-source")); err != nil || string(data) != "independent source A" {
		t.Errorf("active Cargo input mutated across processes: %q %v", data, err)
	}
	data, err = os.ReadFile(filepath.Join(firstRoot, "result.json"))
	if err != nil {
		t.Fatal(err)
	}
	if err = json.Unmarshal(data, &result); err != nil {
		t.Fatal(err)
	}
	if result.Error != "" {
		t.Fatalf("owning build failed: %+v", result)
	}
	for path, want := range map[string]string{filepath.Join(f.project, "src/main.rs"): "independent source A", filepath.Join(other, "src/main.rs"): "independent source B"} {
		if data, err := os.ReadFile(path); err != nil || string(data) != want {
			t.Errorf("caller input changed: %s %q %v", path, data, err)
		}
	}
}
