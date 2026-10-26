//go:build darwin || linux

package phase1

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"go/parser"
	"go/token"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
	"testing"
	"time"
)

func TestController(t *testing.T) {
	t.Run("framing", func(t *testing.T) {
		for _, tc := range []struct {
			frame string
			valid bool
		}{{"ready A 123\n", true}, {"ready A 123", false}, {"ready A 123\nextra\n", false}, {"ready  A 123\n", false}, {"ready A 0\n", false}, {"ready A nope\n", false}, {"launch A 123\n", false}} {
			_, _, err := ParseReady(tc.frame)
			if (err == nil) != tc.valid {
				t.Errorf("frame %q error=%v", tc.frame, err)
			}
		}
	})
	t.Run("duplicate-events", func(t *testing.T) {
		c := controllerTest(t, "A")
		conn := dialReady(t, c, "A", os.Getpid())
		defer func() {
			if err := conn.Close(); err != nil {
				t.Errorf("close test resource: %v", err)
			}
		}()
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		if err := c.WaitReady(ctx, "A"); err != nil {
			t.Fatal(err)
		}
		other := dialReady(t, c, "A", os.Getpid()+1)
		defer func() {
			if err := other.Close(); err != nil {
				t.Errorf("close test resource: %v", err)
			}
		}()
		if err := c.WaitReady(ctx, "missing"); err == nil || !strings.Contains(err.Error(), "duplicate") {
			t.Fatalf("duplicate ready accepted: %v", err)
		}
	})
	t.Run("early-checkpoints", func(t *testing.T) {
		c := controllerTest(t, "A", "B")
		c.RequireReleaseBefore("B", "A")
		conn := dialReady(t, c, "B", os.Getpid())
		defer func() {
			if err := conn.Close(); err != nil {
				t.Errorf("close test resource: %v", err)
			}
		}()
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		if err := c.WaitReady(ctx, "B"); err == nil || !strings.Contains(err.Error(), "early checkpoint") {
			t.Fatalf("early checkpoint accepted: %v", err)
		}
	})
	t.Run("deadlines", func(t *testing.T) {
		c := controllerTest(t, "A")
		ctx, cancel := context.WithCancel(context.Background())
		cancel()
		if err := c.WaitReady(ctx, "A"); err == nil || !strings.Contains(err.Error(), "unmet prerequisite ready A") {
			t.Fatalf("deadline lacks unmet prerequisite: %v", err)
		}
	})
	t.Run("release-order", func(t *testing.T) {
		c := controllerTest(t, "A")
		if err := c.Release("A"); err == nil {
			t.Fatal("release before readiness accepted")
		}
		conn := dialReady(t, c, "A", os.Getpid())
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		if err := c.WaitReady(ctx, "A"); err != nil {
			t.Fatal(err)
		}
		if err := c.Release("A"); err != nil {
			t.Fatal(err)
		}
		if err := c.Release("A"); err == nil {
			t.Fatal("duplicate release accepted")
		}
		buf := make([]byte, len("release\n"))
		if _, err := conn.Read(buf); err != nil || string(buf) != "release\n" {
			t.Fatalf("release framing %q %v", buf, err)
		}
		if callErr := conn.Close(); callErr != nil {
			t.Fatal(callErr)
		}
		if err := c.WaitDisconnected(ctx, "A"); err != nil {
			t.Fatal(err)
		}
	})
	t.Run("fallback-cleanup", func(t *testing.T) {
		c := controllerTest(t, "selftest")
		exe, err := os.Executable()
		if err != nil {
			t.Fatal(err)
		}
		cmd := exec.Command(exe, "-test.run=^TestControllerHelper$")
		cmd.Env = append(os.Environ(), "L_PHASE1_SELFTEST_SOCKET="+c.Socket, "L_PHASE1_SELFTEST_DIRECTORY="+c.Directory)
		cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
		if err := cmd.Start(); err != nil {
			t.Fatal(err)
		}
		waited := false
		defer func() {
			if !waited {
				_ = cmd.Process.Kill() //nolint:errcheck // Fallback may race process exit; lifetime and Wait assertions remain authoritative.
				_ = cmd.Wait()         //nolint:errcheck // Reap the fallback-killed child; a signal exit is expected here.
			}
		}()
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		if err := c.WaitReady(ctx, "selftest"); err != nil {
			t.Fatal(err)
		}
		if err := c.CheckAlive("selftest", os.Getpid()); err != nil {
			t.Fatal(err)
		}
		expired, expire := context.WithCancel(context.Background())
		expire()
		if err := c.ObserveCleanup(expired); err == nil {
			t.Fatal("live process was incorrectly reported cleaned")
		}
		leaks := c.Fallback()
		if len(leaks) != 1 {
			t.Fatalf("fallback did not report pre-termination lifetime lock: %v", leaks)
		}
		if err := cmd.Wait(); err == nil {
			t.Error("fallback child unexpectedly succeeded")
		}
		waited = true
		if err := c.ObserveCleanup(ctx); err != nil {
			t.Fatalf("post-fallback cleanup observation: %v", err)
		}
		if cmd.ProcessState == nil {
			t.Fatal("fallback child not reaped")
		}
		if !cmd.ProcessState.Exited() {
			status, ok := cmd.ProcessState.Sys().(syscall.WaitStatus)
			if !ok || status.Signal() != syscall.SIGKILL {
				t.Fatal("fallback child not reaped")
			}
		}
	})
}
func controllerTest(t *testing.T, labels ...string) *Controller {
	t.Helper()
	c, err := NewController(labels...)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := c.Close(); err != nil {
			t.Errorf("close controller: %v", err)
		}
	})
	return c
}

type closeErrorConn struct {
	net.Conn
	err    error
	closed bool
}

func (c *closeErrorConn) Close() error {
	c.closed = true
	return c.err
}

func TestControllerCloseRetainsCleanupFailures(t *testing.T) {
	controller, err := NewController()
	if err != nil {
		t.Fatal(err)
	}
	directory := controller.Directory
	t.Cleanup(func() {
		if err := os.RemoveAll(directory); err != nil {
			t.Errorf("remove controller directory: %v", err)
		}
	})
	firstErr, secondErr := errors.New("first connection close"), errors.New("second connection close")
	first, second := &closeErrorConn{err: firstErr}, &closeErrorConn{err: secondErr}
	controller.connections = []net.Conn{first, second}
	// A NUL is invalid on both supported hosts and makes removal fail without
	// depending on the current user's filesystem privileges.
	controller.Directory = filepath.Join(directory, "invalid\x00")
	err = controller.Close()
	var pathErr *os.PathError
	if !errors.Is(err, firstErr) || !errors.Is(err, secondErr) || !errors.As(err, &pathErr) {
		t.Fatalf("cleanup lost a connection or directory failure: %v", err)
	}
	if !first.closed || !second.closed {
		t.Fatal("an earlier close failure prevented later connection cleanup")
	}
}

func dialReady(t *testing.T, c *Controller, label string, pid int) net.Conn {
	t.Helper()
	conn, err := net.Dial("unix", c.Socket)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := fmt.Fprintf(conn, "ready %s %d\n", label, pid); err != nil {
		t.Fatal(err)
	}
	return conn
}

// Only the explicit subprocess environment enters this host-control helper. The
// ordinary invocation is a valid assertion that this gate is absent.
func TestControllerHelper(t *testing.T) {
	socket := os.Getenv("L_PHASE1_SELFTEST_SOCKET")
	if socket == "" {
		return
	}
	file, err := os.OpenFile(filepath.Join(os.Getenv("L_PHASE1_SELFTEST_DIRECTORY"), fmt.Sprintf("alive-selftest-%d", os.Getpid())), os.O_CREATE|os.O_RDWR, 0600)
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := file.Close(); err != nil {
			t.Errorf("close test resource: %v", err)
		}
	}()
	if err := syscall.Flock(int(file.Fd()), syscall.LOCK_EX|syscall.LOCK_NB); err != nil {
		t.Fatal(err)
	}
	conn, err := net.Dial("unix", socket)
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := conn.Close(); err != nil {
			t.Errorf("close test resource: %v", err)
		}
	}()
	if _, callErr := fmt.Fprintf(conn, "ready selftest %d\n", os.Getpid()); callErr != nil {
		t.Fatal(callErr)
	}
	data := make([]byte, 1)
	_, _ = conn.Read(data) //nolint:errcheck // The helper waits for release or disconnection; either read outcome ends its lifetime.
}
func TestManifest(t *testing.T) {
	t.Run("all-fixtures", func(t *testing.T) {
		for _, prefix := range []struct {
			letter string
			count  int
		}{{"F", 24}, {"N", 12}} {
			for i := 1; i <= prefix.count; i++ {
				f := LoadFixture(t, fmt.Sprintf("%s%02d", prefix.letter, i))
				t.Run(f.ID, func(t *testing.T) {
					dir := Materialize(t, f)
					wantFiles := make(map[string]bool)
					for _, name := range f.Files {
						got, err := os.ReadFile(filepath.Join(dir, name))
						if err != nil {
							t.Fatal(err)
						}
						want, err := os.ReadFile(filepath.Join(f.Directory, name))
						if err != nil {
							t.Fatal(err)
						}
						if !bytes.Equal(got, want) {
							t.Errorf("materialization changed %s/%s", f.ID, name)
						}
						wantFiles[filepath.Clean(filepath.Join(dir, name))] = true
					}
					if err := filepath.WalkDir(filepath.Dir(dir), func(path string, entry os.DirEntry, err error) error {
						if err != nil {
							return err
						}
						if !entry.IsDir() && !wantFiles[path] {
							t.Errorf("undeclared materialized dependency %s", path)
						}
						return nil
					}); err != nil {
						t.Fatal(err)
					}
				})
			}
		}
	})
	t.Run("malformed-manifests", func(t *testing.T) {
		for _, data := range []string{`{}`, `{"id":"F01"} {}`, `{"id":"../","title":"bad","profile":"print","assertions":["x"]}`, `{"id":"F01","title":"bad","profile":"print","files":["main.l"],"assertions":["x"],"expected_runtime":{"stdout":17,"exit_status":0}}`} {
			if _, err := DecodeFixture([]byte(data)); err == nil {
				t.Errorf("accepted malformed manifest %s", data)
			}
		}
	})
	t.Run("dependency-boundaries", func(t *testing.T) {
		for _, dependency := range []string{"", "/tmp/escape.l", "../../escape.l", "../..", "a/../escape.l", "a\\escape.l", "main.l", "builtin.l", "../common/core.l", "payload.go", "orchestrator.go.want.txt"} {
			data, err := json.Marshal(map[string]interface{}{
				"id": "F01", "title": "dependency", "profile": "print", "assertions": []string{"x"},
				"files": []string{"main.l", dependency}, "expected_runtime": map[string]interface{}{"stdout": "", "exit_status": 0},
			})
			if err != nil {
				t.Fatal(err)
			}
			if _, err := DecodeFixture(data); err == nil {
				t.Errorf("accepted invalid dependency %q", dependency)
			}
		}
	})
	t.Run("byte-oracles", func(t *testing.T) {
		got, err := ExpectedBytes(json.RawMessage(`{"parts":[" \u0000\n",{"repeat":"x","count":3}]}`))
		if err != nil || !bytes.Equal(got, []byte{' ', 0, '\n', 'x', 'x', 'x'}) {
			t.Fatalf("byte oracle %q %v", got, err)
		}
		for _, raw := range []string{`{"repeat":"x","count":-1}`, `{"parts":[false]}`, `{"unexpected":1}`} {
			if _, err := ExpectedBytes(json.RawMessage(raw)); err == nil {
				t.Errorf("accepted invalid oracle %s", raw)
			}
		}
	})
}
func TestStreams(t *testing.T) {
	out, diagnostic := &Stream{}, &Stream{}
	first := []byte{' ', 0, '\n', 0xff}
	if _, callErr := out.Write(first); callErr != nil {
		t.Fatal(callErr)
	}
	if _, callErr := diagnostic.Write([]byte("err\n")); callErr != nil {
		t.Fatal(callErr)
	}
	snapshot := out.Bytes()
	if _, callErr := out.Write([]byte("tail")); callErr != nil {
		t.Fatal(callErr)
	}
	snapshot[0] = 'x'
	if !bytes.Equal(out.Bytes(), append(first, []byte("tail")...)) {
		t.Fatalf("stdout corrupted %q", out.Bytes())
	}
	if string(diagnostic.Bytes()) != "err\n" {
		t.Fatal("streams mixed")
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if err := out.WaitContains(ctx, "tail"); err != nil {
		t.Fatal(err)
	}
	if err := out.WaitContains(ctx, "absent"); err == nil {
		t.Fatal("missing checkpoint did not fail")
	}
}

func TestPhase1SupportIsTestOnly(t *testing.T) {
	r := RepoRoot(t)
	err := filepath.WalkDir(r, func(path string, entry os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() {
			if path != r && (strings.HasPrefix(entry.Name(), ".") || entry.Name() == "testdata" || entry.Name() == "docs") {
				return filepath.SkipDir
			}
			return nil
		}
		if !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
			return nil
		}
		parsed, err := parser.ParseFile(token.NewFileSet(), path, nil, parser.ImportsOnly)
		if err != nil {
			return err
		}
		for _, imp := range parsed.Imports {
			if strings.Contains(imp.Path.Value, "internal/testsupport/phase1") {
				t.Errorf("production source imports Phase1 support: %s", path)
			}
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
}
