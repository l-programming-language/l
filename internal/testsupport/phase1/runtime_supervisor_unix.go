//go:build darwin || linux

package phase1

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"syscall"
	"time"

	"github.com/l-programming-language/l/internal/timing"
)

type RuntimeDriverResult struct {
	Stdout, Stderr  []byte
	Err, CleanupErr error
	// Leaks is captured before fallback. Successful fallback never clears it.
	Leaks          []RuntimeLifetime
	SignaledGroups []int
	Registry       string
}

func (r RuntimeDriverResult) Failure() error {
	var leak error
	if len(r.Leaks) != 0 {
		leak = fmt.Errorf("runtime leaks before fallback: %+v", r.Leaks)
	}
	return errors.Join(r.Err, leak, r.CleanupErr)
}

// RunRuntimeDriver owns the driver and registry independently of driver cleanup.
// Pass an unstarted exec.Command, not CommandContext: there is one Wait owner,
// and cancellation must stop the entire driver group before draining its pipes.
func RunRuntimeDriver(ctx context.Context, cmd *exec.Cmd) (result RuntimeDriverResult) {
	subject := fmt.Sprintf("%s %q", cmd.Path, cmd.Args)
	finish := timing.Start("runtime.total", subject)
	defer func() { finish(result.Failure()) }()
	cleanup := func(error) {}
	defer func() { cleanup(result.CleanupErr) }()
	dir, err := os.MkdirTemp("", "p1-owned-")
	if err != nil {
		result.Err = err
		return
	}
	result.Registry = dir
	defer func() {
		if result.CleanupErr == nil {
			result.CleanupErr = os.RemoveAll(dir)
		}
	}()
	if err := os.WriteFile(filepath.Join(dir, "admission"), []byte("open"), 0600); err != nil {
		result.Err = err
		return
	}
	if cmd.Env == nil {
		cmd.Env = os.Environ()
	}
	cmd.Env = append(cmd.Env, RuntimeRegistryEnv+"="+dir)
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	cmd.WaitDelay = 2 * time.Second
	var out, diagnostic Stream
	cmd.Stdout, cmd.Stderr = &out, &diagnostic
	started := timing.Start("runtime.start", subject)
	startErr := cmd.Start()
	started(startErr)
	if err := startErr; err != nil {
		result.Err = err
		return
	}
	wait := make(chan error, 1)
	go func() {
		finish := timing.Start("runtime.wait", subject)
		err := cmd.Wait()
		finish(err)
		wait <- err
	}()
	finished := false
	select {
	case result.Err = <-wait:
		finished = true
	case <-ctx.Done():
		result.Err = fmt.Errorf("runtime driver watchdog: %w", ctx.Err())
	}
	// Seal registration before taking the inventory, including Start operations
	// whose driver died before recording successful launch.
	cleanup = timing.Start("runtime.cleanup", subject)
	result.CleanupErr = closeRuntimeAdmission(dir)
	result.Leaks, err = runtimeLifetimes(dir)
	result.CleanupErr = errors.Join(result.CleanupErr, err)
	groups := map[int]bool{}
	if !finished {
		groups[cmd.Process.Pid] = true
		if err := syscall.Kill(-cmd.Process.Pid, syscall.SIGKILL); err != nil && !errors.Is(err, syscall.ESRCH) {
			result.CleanupErr = errors.Join(result.CleanupErr, fmt.Errorf("driver group cleanup: %w", err))
		}
	}
	for _, record := range result.Leaks {
		// Revalidate immediately before signaling; released records are historical
		// evidence, never authority to kill a possibly reused PID or process group.
		held, err := LockHeld(filepath.Join(dir, record.Record))
		if err != nil {
			result.CleanupErr = errors.Join(result.CleanupErr, err)
			continue
		}
		if !held {
			continue
		}
		pgid, err := syscall.Getpgid(record.PID)
		if errors.Is(err, syscall.ESRCH) {
			continue
		}
		if err != nil || pgid != record.PGID || pgid <= 1 {
			result.CleanupErr = errors.Join(result.CleanupErr, fmt.Errorf("unverified process group for %+v: current=%d error=%v", record, pgid, err))
			continue
		}
		if groups[pgid] {
			continue
		}
		groups[pgid] = true
		result.SignaledGroups = append(result.SignaledGroups, pgid)
		if err := syscall.Kill(-pgid, syscall.SIGKILL); err != nil && !errors.Is(err, syscall.ESRCH) {
			result.CleanupErr = errors.Join(result.CleanupErr, fmt.Errorf("App group cleanup %d: %w", pgid, err))
		}
	}
	if !finished {
		select {
		case err := <-wait:
			result.Err = errors.Join(result.Err, err)
		case <-time.After(3 * time.Second):
			result.CleanupErr = errors.Join(result.CleanupErr, errors.New("driver Wait did not finish after fallback"))
		}
	}
	deadline := time.Now().Add(time.Second)
	for {
		held, err := runtimeLifetimes(dir)
		if err != nil {
			result.CleanupErr = errors.Join(result.CleanupErr, err)
			break
		}
		if len(held) == 0 {
			break
		}
		if time.Now().After(deadline) {
			result.CleanupErr = errors.Join(result.CleanupErr, fmt.Errorf("lifetime locks remain after fallback: %+v", held))
			break
		}
		time.Sleep(time.Millisecond)
	}
	result.Stdout, result.Stderr = out.Bytes(), diagnostic.Bytes()
	return
}

func closeRuntimeAdmission(dir string) (resultErr error) {
	file, err := os.OpenFile(filepath.Join(dir, "admission"), os.O_RDWR, 0)
	if err != nil {
		return err
	}
	defer func() { resultErr = errors.Join(resultErr, file.Close()) }()
	if err := ownershipLock(file, syscall.LOCK_EX, time.Now().Add(time.Second)); err != nil {
		return err
	}
	_, err = file.WriteAt([]byte("stop"), 0)
	return err
}

func runtimeLifetimes(dir string) ([]RuntimeLifetime, error) {
	paths, err := filepath.Glob(filepath.Join(dir, "alive-*"))
	if err != nil {
		return nil, err
	}
	var records []RuntimeLifetime
	var failures []error
	for _, path := range paths {
		held, err := LockHeld(path)
		if err != nil {
			failures = append(failures, err)
			continue
		}
		if !held {
			continue
		}
		data, err := os.ReadFile(path)
		if err != nil {
			failures = append(failures, err)
			continue
		}
		var record RuntimeLifetime
		if err := json.Unmarshal(data, &record); err != nil {
			failures = append(failures, err)
			continue
		}
		if record.Record != filepath.Base(path) || record.PID <= 1 || record.PGID <= 1 {
			failures = append(failures, fmt.Errorf("invalid lifetime record %s: %+v", path, record))
			continue
		}
		records = append(records, record)
	}
	return records, errors.Join(failures...)
}
