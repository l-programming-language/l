//go:build darwin || linux

package phase1

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"
)

type Peer struct {
	Label                  string
	PID                    int
	connection             net.Conn
	Released, Disconnected bool
}

// Controller owns only the control endpoint. Cleanup observation must happen
// before Close or Fallback: closing a socket can terminate a fixture itself.
type Controller struct {
	Directory, Socket string
	listener          net.Listener
	mu                sync.Mutex
	peers             map[string]*Peer
	connections       []net.Conn
	expected          map[string]bool
	precedes          map[string]string
	events            []string
	protocolErr       error
	changed           chan struct{}
	closed            bool
}

func NewController(labels ...string) (*Controller, error) {
	dir, err := os.MkdirTemp("/tmp", "lp1-")
	if err != nil {
		return nil, err
	}
	socket := filepath.Join(dir, "s")
	ln, err := net.Listen("unix", socket)
	if err != nil {
		return nil, errors.Join(err, os.RemoveAll(dir))
	}
	c := &Controller{Directory: dir, Socket: socket, listener: ln, peers: map[string]*Peer{}, expected: map[string]bool{}, precedes: map[string]string{}, changed: make(chan struct{})}
	for _, label := range labels {
		c.expected[label] = true
	}
	go c.accept()
	return c, nil
}
func ParseReady(line string) (string, int, error) {
	if !strings.HasSuffix(line, "\n") || strings.Contains(strings.TrimSuffix(line, "\n"), "\n") {
		return "", 0, fmt.Errorf("incomplete or multiple readiness frames %q", line)
	}
	fields := strings.Fields(strings.TrimSuffix(line, "\n"))
	if len(fields) != 3 || fields[0] != "ready" || strings.Join(fields, " ")+"\n" != line {
		return "", 0, fmt.Errorf("invalid readiness frame %q", line)
	}
	pid, err := strconv.Atoi(fields[2])
	if err != nil || pid <= 1 {
		return "", 0, fmt.Errorf("invalid child PID %q", fields[2])
	}
	return fields[1], pid, nil
}
func (c *Controller) change() { close(c.changed); c.changed = make(chan struct{}) }
func (c *Controller) fail(err error) {
	if c.protocolErr == nil {
		c.protocolErr = err
	}
	c.events = append(c.events, "violation: "+err.Error())
	c.change()
}
func (c *Controller) accept() {
	for {
		conn, err := c.listener.Accept()
		if err != nil {
			return
		}
		c.mu.Lock()
		if c.closed {
			c.mu.Unlock()
			conn.Close() //nolint:errcheck // This connection arrived after shutdown and has no pending controller output.
			return
		}
		c.connections = append(c.connections, conn)
		c.mu.Unlock()
		go c.read(conn)
	}
}
func (c *Controller) read(conn net.Conn) {
	reader := bufio.NewReader(conn)
	line, err := reader.ReadString('\n')
	if err != nil {
		c.mu.Lock()
		if !c.closed {
			c.fail(fmt.Errorf("read readiness: %w", err))
		}
		c.mu.Unlock()
		return
	}
	label, pid, err := ParseReady(line)
	c.mu.Lock()
	if err != nil {
		c.fail(err)
		c.mu.Unlock()
		return
	}
	if !c.expected[label] {
		c.fail(fmt.Errorf("unexpected ready label %s", label))
		c.mu.Unlock()
		return
	}
	if _, ok := c.peers[label]; ok {
		c.fail(fmt.Errorf("duplicate ready label %s", label))
		c.mu.Unlock()
		return
	}
	for _, other := range c.peers {
		if other.PID == pid {
			c.fail(fmt.Errorf("two Apps share PID %d", pid))
			c.mu.Unlock()
			return
		}
	}
	if prerequisite := c.precedes[label]; prerequisite != "" {
		p := c.peers[prerequisite]
		if p == nil || !p.Released {
			c.fail(fmt.Errorf("early checkpoint %s before release %s", label, prerequisite))
		}
	}
	p := &Peer{Label: label, PID: pid, connection: conn}
	c.peers[label] = p
	c.events = append(c.events, fmt.Sprintf("ready %s %d", label, pid))
	c.change()
	c.mu.Unlock()
	// An EOF is a child observation, not something synthesized by controller Close.
	b := make([]byte, 1)
	n, err := reader.Read(b)
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.closed {
		return
	}
	if n != 0 || !errors.Is(err, io.EOF) {
		c.fail(fmt.Errorf("unexpected post-ready frame %s: %q %v", label, b[:n], err))
		return
	}
	p.Disconnected = true
	c.events = append(c.events, "disconnected "+label)
	c.change()
}
func (c *Controller) RequireReleaseBefore(later, earlier string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.precedes[later] = earlier
}
func (c *Controller) wait(ctx context.Context, description string, ready func() bool) error {
	for {
		c.mu.Lock()
		if c.protocolErr != nil {
			err := c.protocolErr
			c.mu.Unlock()
			return err
		}
		if ready() {
			c.mu.Unlock()
			return nil
		}
		changed := c.changed
		c.mu.Unlock()
		select {
		case <-ctx.Done():
			return fmt.Errorf("unmet prerequisite %s: %w; events=%v", description, ctx.Err(), c.Events())
		case <-changed:
		}
	}
}
func (c *Controller) WaitReady(ctx context.Context, labels ...string) error {
	return c.wait(ctx, "ready "+strings.Join(labels, ","), func() bool {
		for _, label := range labels {
			if c.peers[label] == nil {
				return false
			}
		}
		return true
	})
}
func (c *Controller) WaitDisconnected(ctx context.Context, label string) error {
	return c.wait(ctx, "disconnected "+label, func() bool { return c.peers[label] != nil && c.peers[label].Disconnected })
}
func (c *Controller) Release(label string) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	p := c.peers[label]
	if p == nil {
		return fmt.Errorf("release before readiness: %s", label)
	}
	if p.Released {
		return fmt.Errorf("duplicate release: %s", label)
	}
	if _, err := io.WriteString(p.connection, "release\n"); err != nil {
		return err
	}
	p.Released = true
	c.events = append(c.events, "release "+label)
	c.change()
	return nil
}
func (c *Controller) Events() []string {
	c.mu.Lock()
	defer c.mu.Unlock()
	return append([]string(nil), c.events...)
}
func (c *Controller) Peers() []Peer {
	c.mu.Lock()
	defer c.mu.Unlock()
	out := []Peer{}
	for _, p := range c.peers {
		out = append(out, *p)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Label < out[j].Label })
	return out
}
func LockHeld(path string) (held bool, resultErr error) {
	f, err := os.OpenFile(path, os.O_RDWR, 0600)
	if err != nil {
		return false, err
	}
	defer func() { resultErr = errors.Join(resultErr, f.Close()) }()
	err = syscall.Flock(int(f.Fd()), syscall.LOCK_EX|syscall.LOCK_NB)
	if errors.Is(err, syscall.EWOULDBLOCK) || errors.Is(err, syscall.EAGAIN) {
		return true, nil
	}
	if err != nil {
		return false, err
	}
	return false, syscall.Flock(int(f.Fd()), syscall.LOCK_UN)
}
func (c *Controller) CheckAlive(label string, orchestratorPID int) error {
	c.mu.Lock()
	p := c.peers[label]
	if p == nil {
		c.mu.Unlock()
		return fmt.Errorf("missing peer %s", label)
	}
	copy := *p
	c.mu.Unlock()
	if copy.PID == orchestratorPID || copy.Disconnected {
		return fmt.Errorf("App %s is not a distinct live child", label)
	}
	held, err := LockHeld(filepath.Join(c.Directory, fmt.Sprintf("alive-%s-%d", label, copy.PID)))
	if err != nil || !held {
		return fmt.Errorf("App %s lifetime lock not held: %v", label, err)
	}
	return nil
}
func (c *Controller) ObserveCleanup(ctx context.Context) error {
	for _, p := range c.Peers() {
		if err := c.WaitDisconnected(ctx, p.Label); err != nil {
			return err
		}
	}
	ticker := time.NewTicker(5 * time.Millisecond)
	defer ticker.Stop()
	for {
		files, err := filepath.Glob(filepath.Join(c.Directory, "alive-*"))
		if err != nil {
			return err
		}
		var heldPaths []string
		for _, file := range files {
			held, err := LockHeld(file)
			if err != nil {
				return err
			}
			if held {
				heldPaths = append(heldPaths, file)
			}
		}
		if len(heldPaths) == 0 {
			return nil
		}
		select {
		case <-ctx.Done():
			return fmt.Errorf("runtime leak: held lifetime locks %v: %w", heldPaths, ctx.Err())
		case <-ticker.C:
		}
	}
}
func (c *Controller) Leaks() []string {
	var leaks []string
	files, _ := filepath.Glob(filepath.Join(c.Directory, "alive-*")) //nolint:errcheck // NewController creates a /tmp/lp1-* directory, so this fixed pattern is always valid.
	for _, file := range files {
		held, err := LockHeld(file)
		if err == nil && held {
			leaks = append(leaks, file)
		}
	}
	return leaks
}

// Fallback returns the pre-termination leak inventory. Tests must fail on that
// inventory before calling it; successful fallback does not establish cleanup.
func (c *Controller) Fallback() []string {
	leaks := []string{}
	files, _ := filepath.Glob(filepath.Join(c.Directory, "alive-*")) //nolint:errcheck // NewController creates a /tmp/lp1-* directory, so this fixed pattern is always valid.
	for _, file := range files {
		held, err := LockHeld(file)
		if err != nil || !held {
			continue
		}
		parts := strings.Split(filepath.Base(file), "-")
		pid, err := strconv.Atoi(parts[len(parts)-1])
		if err != nil || pid <= 1 || pid == os.Getpid() {
			continue
		}
		leaks = append(leaks, file)
		if group, e := syscall.Getpgid(pid); e == nil && group == pid {
			_ = syscall.Kill(-pid, syscall.SIGKILL) //nolint:errcheck // Fallback can race exit; the returned pre-termination leak remains a test failure.
		} else {
			_ = syscall.Kill(pid, syscall.SIGKILL) //nolint:errcheck // Fallback can race exit; the returned pre-termination leak remains a test failure.
		}
	}
	return leaks
}
func (c *Controller) Close() error {
	c.mu.Lock()
	c.closed = true
	c.change()
	connections := append([]net.Conn(nil), c.connections...)
	c.mu.Unlock()
	err := c.listener.Close()
	for _, conn := range connections {
		err = errors.Join(err, conn.Close())
	}
	return errors.Join(err, os.RemoveAll(c.Directory))
}
