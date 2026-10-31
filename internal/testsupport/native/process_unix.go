//go:build darwin || linux

package native

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"syscall"
	"testing"
	"time"

	"github.com/l-programming-language/l/internal/timing"
)

// Admission bounds concurrent native cases before their watchdogs start.
type Admission chan struct{}

// Parallel admits a parallel test until its cleanup completes.
func (cases Admission) Parallel(t *testing.T) {
	t.Helper()
	t.Parallel()
	admitted := timing.Start("admission.wait", t.Name())
	cases <- struct{}{}
	admitted(nil)
	t.Cleanup(func() { <-cases })
}

// The test observer owns the complete command process group on cancellation.
func superviseCompiler(cmd *exec.Cmd) {
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	cmd.Cancel = func() error { return syscall.Kill(-cmd.Process.Pid, syscall.SIGKILL) }
	cmd.WaitDelay = 2 * time.Second
}

// Run supervises distribution commands with their existing 20-second runtime
// watchdog; compilation remains cancelable without a short wrapper deadline.
func (phase CommandPhase) Run(t testing.TB, executable, directory string, environment []string, args ...string) (outBytes, diagnosticBytes []byte, err error) {
	t.Helper()
	finish := timing.Start("test.command.total", fmt.Sprintf("%s: %s %q", t.Name(), executable, args))
	defer func() { finish(err) }()
	ctx, cancel := phase.Context(context.Background(), 20*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, executable, args...)
	cmd.Dir, cmd.Env = directory, append(os.Environ(), environment...)
	superviseCompiler(cmd)
	var stdout, stderr bytes.Buffer
	cmd.Stdout, cmd.Stderr = &stdout, &stderr
	err = cmd.Run()
	if ctx.Err() != nil {
		t.Fatalf("%s command canceled: %v; stdout=%q stderr=%q", phase, ctx.Err(), stdout.String(), stderr.String())
	}
	return stdout.Bytes(), stderr.Bytes(), err
}
