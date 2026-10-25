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
	"sync"
	"syscall"
	"testing"
	"time"
)

// This exercises the real PATH proxy, not a substitute supervisor. Its Cargo
// and child remain in the caller's group and hold independent lifetime locks.
func TestCLIProxyRetainsOwnedProcessGroup(t *testing.T) {
	dir := t.TempDir()
	source := filepath.Join(dir, "owned.go")
	helper := filepath.Join(dir, "owned-cargo")
	if err := os.WriteFile(source, []byte(ownedCargoSource), 0600); err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), BuildPrerequisiteTimeout)
	defer cancel()
	build := exec.CommandContext(ctx, "go", "build", "-o", helper, source)
	build.Env = append(os.Environ(), "GOENV=off", "GOFLAGS=", "GOWORK=off", "GOTOOLCHAIN=local")
	if output, err := build.CombinedOutput(); err != nil {
		t.Fatalf("compile owned Cargo fixture: %v: %s", err, output)
	}
	probe := CLIProbe(t)
	config, err := json.Marshal(struct{ Directory, Cargo string }{probe.Directory, helper})
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(probe.Directory, "cargo-proxy.json"), config, 0600); err != nil {
		t.Fatal(err)
	}
	cmd := exec.Command(filepath.Join(probe.Directory, "cargo"), "+"+Version, "build", dir)
	cmd.Env = probe.Env(nil)
	cmd.Dir = dir
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	cmd.WaitDelay = 2 * time.Second
	var out, diagnostic ownedProxyStream
	cmd.Stdout, cmd.Stderr = &out, &diagnostic
	if err := cmd.Start(); err != nil {
		t.Fatal(err)
	}
	done := make(chan error, 1)
	go func() { done <- cmd.Wait() }()
	waited := false
	lifetimesReleased := false
	t.Cleanup(func() {
		// Reaping the proxy does not prove its Cargo/descendants exited. Keep
		// group cleanup separate from the exactly-once Wait bookkeeping.
		if !lifetimesReleased {
			_ = syscall.Kill(-cmd.Process.Pid, syscall.SIGKILL) //nolint:errcheck // Fallback may race process exit; lifetime and Wait assertions remain authoritative.
		}
		if !waited {
			select {
			case <-done:
			case <-time.After(2 * time.Second):
				t.Error("owned proxy did not reap during fallback cleanup")
			}
		}
	})
	deadline := time.Now().Add(5 * time.Second)
	var observed struct{ CargoPID, CargoPGID, ChildPID, ChildPGID int }
	for {
		data, readErr := os.ReadFile(filepath.Join(dir, "ready.json"))
		if readErr == nil && json.Unmarshal(data, &observed) == nil && out.String() == "cargo ready\n" && diagnostic.String() == "cargo diagnostic\x00tail\n" {
			break
		}
		select {
		case err := <-done:
			waited = true
			t.Fatalf("proxy exited before owned descendants ready: %v; stdout=%q stderr=%q", err, out.Bytes(), diagnostic.Bytes())
		default:
		}
		if time.Now().After(deadline) {
			t.Fatal("owned Cargo descendant did not become ready")
		}
		time.Sleep(time.Millisecond)
	}
	if observed.CargoPID <= 1 || observed.ChildPID <= 1 || observed.CargoPGID != cmd.Process.Pid || observed.ChildPGID != cmd.Process.Pid {
		t.Fatalf("proxy PGID %d, descendant identities %+v", cmd.Process.Pid, observed)
	}
	for _, name := range []string{"cargo.lock", "child.lock"} {
		if !ownedCargoLockHeld(t, filepath.Join(dir, name)) {
			t.Fatalf("%s not held before cancellation", name)
		}
	}
	if err := syscall.Kill(-cmd.Process.Pid, syscall.SIGKILL); err != nil {
		t.Fatal(err)
	}
	select {
	case err := <-done:
		waited = true
		var exit *exec.ExitError
		if !errors.As(err, &exit) {
			t.Fatalf("cancelled proxy status: %v", err)
		}
		status, ok := exit.Sys().(syscall.WaitStatus)
		if !ok || !status.Signaled() {
			t.Fatalf("cancelled proxy status: %v", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("cancelled proxy did not reap")
	}
	deadline = time.Now().Add(2 * time.Second)
	for ownedCargoLockHeld(t, filepath.Join(dir, "cargo.lock")) || ownedCargoLockHeld(t, filepath.Join(dir, "child.lock")) {
		if time.Now().After(deadline) {
			t.Fatal("owned descendant retained lifetime lock after group cancellation")
		}
		time.Sleep(time.Millisecond)
	}
	lifetimesReleased = true
	if out.String() != "cargo ready\n" || diagnostic.String() != "cargo diagnostic\x00tail\n" {
		t.Fatalf("cancelled proxy streams: %q / %q", out.Bytes(), diagnostic.Bytes())
	}
}

type ownedProxyStream struct {
	mu sync.Mutex
	b  bytes.Buffer
}

func (s *ownedProxyStream) Write(p []byte) (int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.b.Write(p)
}

func (s *ownedProxyStream) Bytes() []byte {
	s.mu.Lock()
	defer s.mu.Unlock()
	return bytes.Clone(s.b.Bytes())
}

func (s *ownedProxyStream) String() string { return string(s.Bytes()) }

func ownedCargoLockHeld(t *testing.T, path string) bool {
	t.Helper()
	f, err := os.OpenFile(path, os.O_RDWR, 0)
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := f.Close(); err != nil {
			t.Errorf("close test resource: %v", err)
		}
	}()
	err = syscall.Flock(int(f.Fd()), syscall.LOCK_EX|syscall.LOCK_NB)
	if errors.Is(err, syscall.EWOULDBLOCK) || errors.Is(err, syscall.EAGAIN) {
		return true
	}
	if err != nil {
		t.Fatal(err)
	}
	if err := syscall.Flock(int(f.Fd()), syscall.LOCK_UN); err != nil {
		t.Fatal(err)
	}
	return false
}

const ownedCargoSource = `package main
import("encoding/json";"fmt";"os";"os/exec";"path/filepath";"syscall";"time")
func lock(path string)*os.File{f,e:=os.OpenFile(path,os.O_CREATE|os.O_RDWR,0600);if e!=nil{panic(e)};if e=syscall.Flock(int(f.Fd()),syscall.LOCK_EX);e!=nil{panic(e)};return f}
func main(){
 dir:=os.Args[2]
 if os.Args[1]=="child"{f:=lock(filepath.Join(dir,"child.lock"));defer f.Close();ready:=os.NewFile(3,"ready");_,_=ready.Write([]byte{1});_=ready.Close();for{time.Sleep(time.Hour)}}
 f:=lock(filepath.Join(dir,"cargo.lock"));defer f.Close()
 r,w,e:=os.Pipe();if e!=nil{panic(e)}
 child:=exec.Command(os.Args[0],"child",dir);child.ExtraFiles=[]*os.File{w};if e=child.Start();e!=nil{panic(e)};_=w.Close();var b [1]byte;if _,e=r.Read(b[:]);e!=nil{panic(e)};_=r.Close()
 pgid,e:=syscall.Getpgid(0);if e!=nil{panic(e)};cpgid,e:=syscall.Getpgid(child.Process.Pid);if e!=nil{panic(e)}
 data,_:=json.Marshal(struct{CargoPID,CargoPGID,ChildPID,ChildPGID int}{os.Getpid(),pgid,child.Process.Pid,cpgid})
 fmt.Print("cargo ready\n");fmt.Fprint(os.Stderr,"cargo diagnostic\x00tail\n")
 if e=os.WriteFile(filepath.Join(dir,"ready.json"),data,0600);e!=nil{panic(e)}
 _=child.Wait()
}
`
