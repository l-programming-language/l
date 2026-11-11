// Package process provides test-owned subprocess supervision and lifetime tracking.
package process

import (
	"bytes"
	"sync"
)

// stream preserves exact output bytes and returns independent snapshots.
type stream struct {
	mu   sync.Mutex
	data bytes.Buffer
}

func (s *stream) Write(p []byte) (int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.data.Write(p)
}
func (s *stream) Bytes() []byte {
	s.mu.Lock()
	defer s.mu.Unlock()
	return append([]byte(nil), s.data.Bytes()...)
}
