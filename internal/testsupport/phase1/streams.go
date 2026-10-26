package phase1

import (
	"bytes"
	"context"
	"fmt"
	"sync"
)

// Stream keeps raw bytes and a broadcast wakeup for progress assertions. The
// event protocol uses a separate socket and never contaminates either stream.
type Stream struct {
	mu      sync.Mutex
	data    bytes.Buffer
	changed chan struct{}
}

func (s *Stream) signal() {
	if s.changed != nil {
		close(s.changed)
	}
	s.changed = make(chan struct{})
}
func (s *Stream) Write(p []byte) (int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	n, err := s.data.Write(p)
	s.signal()
	return n, err
}
func (s *Stream) Bytes() []byte {
	s.mu.Lock()
	defer s.mu.Unlock()
	return append([]byte(nil), s.data.Bytes()...)
}
func (s *Stream) WaitContains(ctx context.Context, needle string) error {
	for {
		s.mu.Lock()
		if bytes.Contains(s.data.Bytes(), []byte(needle)) {
			s.mu.Unlock()
			return nil
		}
		if s.changed == nil {
			s.changed = make(chan struct{})
		}
		changed := s.changed
		s.mu.Unlock()
		select {
		case <-ctx.Done():
			return fmt.Errorf("missing orchestrator checkpoint %q: %w", needle, ctx.Err())
		case <-changed:
		}
	}
}
