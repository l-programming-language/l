//go:build darwin || linux

package main

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"syscall"
	"testing"
	"time"

	"github.com/l-programming-language/l/internal/testsupport/phase1"
)

func TestRunnableServiceStopWithoutJoining(t *testing.T) {
	project, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	phase1.CopyFile(t, filepath.Join("test", "runnable_service", "stop_pipe.l"), filepath.Join(project, "stop_pipe.l"), 0600)
	const source = "Use StopProbe's Print from `stop_pipe.l` as \"Print\".\n" +
		"Create \"Worker\" as a New Basic App.\nThe App Prints `worker`.\n" +
		"Create \"Gate\" as a New Basic App.\nThe App Prints `gate`.\n" +
		"Run Worker.\nRun Gate.\nAwait Gate.\nStop Worker.\n" +
		"Show `stop-returned|`.\nAwait Worker.\nShow Worker's Output.\nShow `joined|`.\n"
	if err := os.WriteFile(filepath.Join(project, "main.l"), []byte(source), 0600); err != nil {
		t.Fatal(err)
	}
	built := runnableBuildProject(t, project, []string{"Worker", "Gate"})
	executable, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	directory, err := os.MkdirTemp("/tmp", "l-stop-pipe-")
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := os.RemoveAll(directory); err != nil {
			t.Error(err)
		}
	}()
	socket := filepath.Join(directory, "s")
	listener, err := net.ListenUnix("unix", &net.UnixAddr{Name: socket, Net: "unix"})
	if err != nil {
		t.Fatal(err)
	}
	// This transfer shares the existing five-second runtime allowance. No timer
	// determines success: only the post-Stop bytes in finalized Output can do so.
	deadline := time.Now().Add(5 * time.Second)
	if err := listener.SetDeadline(deadline); err != nil {
		if closeErr := listener.Close(); closeErr != nil {
			t.Error(closeErr)
		}
		t.Fatal(err)
	}
	t.Setenv("L_SERVICE_PIPE_HELPER", executable)
	t.Setenv("L_SERVICE_PIPE_SOCKET", socket)
	t.Setenv("L_SERVICE_PIPE_DEADLINE", deadline.Format(time.RFC3339Nano))
	type transferResult struct {
		pipe *os.File
		err  error
	}
	transfer := make(chan transferResult, 1)
	go func() {
		pipe, receiveErr := runnableReceiveStopPipe(listener, deadline)
		transfer <- transferResult{pipe: pipe, err: receiveErr}
	}()
	var writer *os.File
	received := false
	defer func() {
		if err := listener.Close(); err != nil {
			t.Error(err)
		}
		if !received {
			// Closing the listener or the shared deadline settles the receiver,
			// including failures before Worker can announce readiness.
			result := <-transfer
			writer = result.pipe
			if result.err != nil {
				t.Errorf("Service pipe transfer: %v", result.err)
			}
		}
		if writer != nil {
			if err := writer.Close(); err != nil {
				t.Error(err)
			}
		}
	}()
	status, stdout, stderr := runnableGated(t, built, []string{"worker", "gate"}, func(ctx context.Context, controller *phase1.Controller, _ *exec.Cmd, output *phase1.Stream) {
		select {
		case result := <-transfer:
			received = true
			writer = result.pipe
			if result.err != nil {
				t.Fatal(result.err)
			}
		case <-ctx.Done():
			t.Fatal("Service stdout transfer did not finish before readiness")
		}
		if err := controller.Release("gate"); err != nil {
			t.Fatal(err)
		}
		if err := output.WaitContains(ctx, "stop-returned|"); err != nil {
			t.Fatal(err)
		}
		if err := controller.WaitDisconnected(ctx, "worker"); err != nil {
			t.Fatal(err)
		}
		// The bridge was reaped before readiness and Worker was never released.
		// Only this retained writer can keep its capture open after Stop. A Stop
		// that joins until the drain timeout cannot capture these later bytes.
		const late = "late-capture\n"
		if count, err := io.WriteString(writer, late); err != nil || count != len(late) {
			t.Fatalf("Stop returned after capture closed: wrote %d/%d bytes: %v", count, len(late), err)
		}
		closeErr := writer.Close()
		writer = nil
		if closeErr != nil {
			t.Fatal(closeErr)
		}
	})
	if status != 0 || !bytes.Equal(stdout, []byte("stop-returned|late-capture\njoined|")) || len(stderr) != 0 {
		t.Fatalf("Stop must return before finalized capture: status=%d stdout=%q stderr=%q", status, stdout, stderr)
	}
}

// The generated App synchronously reaps this descriptor-transfer helper
// before announcing readiness. It is neither an alternate App executable
// nor a runtime implementation, and it never writes application stdout.
func TestRunnableServicePipeBridgeHelper(t *testing.T) {
	if os.Getenv("L_SERVICE_PIPE_BRIDGE_CHILD") != "1" {
		return
	}
	if err := runnableSendStopPipe(); err != nil {
		if _, diagnosticErr := fmt.Fprintln(os.Stderr, err); diagnosticErr != nil {
			os.Exit(2)
		}
		os.Exit(1)
	}
	os.Exit(0) // Avoid the Go test harness's PASS line in the inherited pipe.
}

func runnableSendStopPipe() (resultErr error) {
	deadline, err := time.Parse(time.RFC3339Nano, os.Getenv("L_SERVICE_PIPE_DEADLINE"))
	if err != nil {
		return err
	}
	connection, err := net.DialUnix("unix", nil, &net.UnixAddr{Name: os.Getenv("L_SERVICE_PIPE_SOCKET"), Net: "unix"})
	if err != nil {
		return err
	}
	defer func() { resultErr = errors.Join(resultErr, connection.Close()) }()
	if err := connection.SetDeadline(deadline); err != nil {
		return err
	}
	rights := syscall.UnixRights(int(os.Stdout.Fd()))
	count, controlCount, err := connection.WriteMsgUnix([]byte{'P'}, rights, nil)
	if err != nil {
		return err
	}
	if count != 1 || controlCount != len(rights) {
		return fmt.Errorf("incomplete Service pipe transfer: %d/%d", count, controlCount)
	}
	ack := make([]byte, 1)
	if _, err := io.ReadFull(connection, ack); err != nil {
		return err
	}
	if ack[0] != 'A' {
		return fmt.Errorf("invalid Service pipe acknowledgment %q", ack)
	}
	return nil
}

func runnableReceiveStopPipe(listener *net.UnixListener, deadline time.Time) (pipe *os.File, resultErr error) {
	connection, err := listener.AcceptUnix()
	if err != nil {
		return nil, err
	}
	defer func() { resultErr = errors.Join(resultErr, connection.Close()) }()
	if err := connection.SetDeadline(deadline); err != nil {
		return nil, err
	}
	data, control := make([]byte, 2), make([]byte, syscall.CmsgSpace(2*4))
	count, controlCount, flags, _, readErr := connection.ReadMsgUnix(data, control)
	messages, parseErr := syscall.ParseSocketControlMessage(control[:controlCount])
	var descriptors []int
	defer func() {
		for _, descriptor := range descriptors {
			resultErr = errors.Join(resultErr, syscall.Close(descriptor))
		}
	}()
	for _, message := range messages {
		rights, err := syscall.ParseUnixRights(&message)
		if err != nil {
			parseErr = errors.Join(parseErr, err)
			continue
		}
		for _, descriptor := range rights {
			syscall.CloseOnExec(descriptor)
		}
		descriptors = append(descriptors, rights...)
	}
	if err := errors.Join(readErr, parseErr); err != nil {
		return nil, err
	}
	if flags&(syscall.MSG_CTRUNC|syscall.MSG_TRUNC) != 0 || count != 1 || data[0] != 'P' || len(messages) != 1 || len(descriptors) != 1 {
		return nil, fmt.Errorf("invalid Service pipe frame: bytes=%q flags=%d messages=%d descriptors=%d", data[:count], flags, len(messages), len(descriptors))
	}
	pipe = os.NewFile(uintptr(descriptors[0]), "Service stdout writer")
	descriptors = nil // The returned file now owns the received descriptor.
	if count, err := connection.Write([]byte{'A'}); err != nil || count != 1 {
		return pipe, fmt.Errorf("acknowledge Service pipe: bytes=%d error=%v", count, err)
	}
	return pipe, nil
}
