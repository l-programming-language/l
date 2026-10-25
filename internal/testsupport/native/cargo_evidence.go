package native

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"os/exec"
	"strings"
	"sync"
	"syscall"
	"testing"
	"time"

	"github.com/l-programming-language/l/internal/timing"
)

// cargoEvidenceInput describes the actual invocation after staging and while
// the caller still owns its cache lock. Evidence never becomes a build input.
type cargoEvidenceInput struct {
	Route, Directory, Workspace, Manifest, Target, RunID                  string
	Test, OriginalWorkspace, PreferredSlot, SelectedSlot, Toolchain, Host string
	Command, Environment                                                  []string
	PID, PPID                                                             int
}

type cargoEvidence struct {
	once                sync.Once
	t                   testing.TB
	recorder, directory string
	environment         []string
}

// Two bounded recorder launches are diagnostic overhead only. Sharing the
// collector with the CLI observer keeps the snapshot and completion protocol
// identical; an ordinary build performs neither launch nor evidence I/O.
func beginCargoEvidence(t testing.TB, input cargoEvidenceInput) *cargoEvidence {
	if os.Getenv(cargoEvidenceEnvironment) == "" {
		return &cargoEvidence{}
	}
	var err error
	input.RunID, err = evidenceRunID()
	if err != nil {
		t.Logf("Cargo evidence incomplete: %v", err)
		return &cargoEvidence{}
	}
	if input.Test == "" {
		input.Test = t.Name()
	}
	recorder, err := artifactRecorderResult()
	if err != nil {
		t.Logf("Cargo evidence incomplete: %v", err)
		return &cargoEvidence{}
	}
	input.PID, input.PPID = os.Getpid(), os.Getppid()
	data, err := json.Marshal(input)
	if err != nil {
		t.Logf("Cargo evidence incomplete: %v", err)
		return &cargoEvidence{}
	}
	e := &cargoEvidence{t: t, recorder: recorder, environment: append([]string(nil), input.Environment...)}
	finish := timing.Start("evidence.inputs", input.Workspace)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, e.recorder, "__evidence_begin")
	superviseCompiler(cmd)
	cmd.Stdin = bytes.NewReader(data)
	cmd.Env = append(append([]string(nil), e.environment...), cargoEvidenceEnvironment+"="+os.Getenv(cargoEvidenceEnvironment), evidenceRunEnvironment+"="+input.RunID)
	var out, diag bytes.Buffer
	cmd.Stdout, cmd.Stderr = &out, &diag
	err = cmd.Run()
	finish(err)
	if err != nil {
		t.Logf("Cargo evidence incomplete: %v: %s", err, diag.Bytes())
		return &cargoEvidence{}
	}
	e.directory = strings.TrimSpace(out.String())
	return e
}

func (e *cargoEvidence) Finish(stdout, stderr []byte, cargoErr error) {
	if e == nil || e.directory == "" {
		return
	}
	e.once.Do(func() {
		result := struct {
			Directory      string
			Stdout, Stderr []byte
			Status         int
			Error          string
		}{Directory: e.directory, Stdout: stdout, Stderr: stderr}
		if cargoErr != nil {
			result.Status = -1
			result.Error = cargoErr.Error()
			if exit, ok := cargoErr.(*exec.ExitError); ok {
				result.Status = exit.ExitCode()
				if wait, ok := exit.Sys().(syscall.WaitStatus); ok && wait.Signaled() {
					result.Status = 128 + int(wait.Signal())
				}
			}
		}
		data, err := json.Marshal(result)
		if err != nil {
			e.t.Logf("Cargo evidence incomplete: %v", err)
			return
		}
		finish := timing.Start("evidence.complete", e.directory)
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		cmd := exec.CommandContext(ctx, e.recorder, "__evidence_finish")
		superviseCompiler(cmd)
		cmd.Stdin = bytes.NewReader(data)
		cmd.Env = e.environment
		out, err := cmd.CombinedOutput()
		finish(err)
		if err != nil {
			e.t.Logf("Cargo evidence incomplete: %v: %s", err, out)
			return
		}
		timing.Record("cargo.evidence", e.directory, map[string]any{"status": result.Status})
		e.t.Logf("immutable Cargo evidence: %s", e.directory)
	})
}
