// Package timing records optional acceptance diagnostics without changing command
// output or outcomes. Set L_HARNESS_TIMING_DIR to a private artifact directory;
// each process writes its own JSONL stream. Trace failures are nonfatal.
package timing

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"sync/atomic"
	"time"
)

var sequence uint64
var writes sync.Mutex

type event struct {
	Event           string         `json:"event"`
	Span            string         `json:"span,omitempty"`
	Phase           string         `json:"phase"`
	Subject         string         `json:"subject"`
	PID             int            `json:"pid"`
	PPID            int            `json:"ppid"`
	StartedAt       time.Time      `json:"started_at"`
	DurationNS      int64          `json:"duration_ns"`
	DurationSeconds float64        `json:"duration_seconds"`
	Outcome         string         `json:"outcome"`
	Error           string         `json:"error,omitempty"`
	Fields          map[string]any `json:"fields,omitempty"`
}

func Enabled() bool { return os.Getenv("L_HARNESS_TIMING_DIR") != "" }

// Start records admission to a phase immediately, so interrupted processes leave
// an open span. Its completion callback is idempotent and never changes err.
func Start(phase, subject string) func(error) {
	directory := os.Getenv("L_HARNESS_TIMING_DIR")
	if directory == "" {
		return func(error) {}
	}
	directory, err := filepath.Abs(directory)
	if err != nil {
		return func(error) {}
	}
	started := time.Now()
	base := event{Event: "start", Span: fmt.Sprintf("%d-%d", os.Getpid(), atomic.AddUint64(&sequence, 1)), Phase: phase, Subject: subject, PID: os.Getpid(), PPID: os.Getppid(), StartedAt: started.UTC(), Outcome: "running"}
	write(directory, base)
	var once sync.Once
	return func(err error) {
		once.Do(func() {
			elapsed := time.Since(started)
			base.Event = "finish"
			base.DurationNS = elapsed.Nanoseconds()
			base.DurationSeconds = elapsed.Seconds()
			base.Outcome = "ok"
			if err != nil {
				base.Outcome = "error"
				base.Error = err.Error()
			}
			write(directory, base)
		})
	}
}

// Record attaches read-only counters or artifact locations to the same stream.
func Record(phase, subject string, fields map[string]any) {
	directory := os.Getenv("L_HARNESS_TIMING_DIR")
	if directory == "" {
		return
	}
	write(directory, event{Event: "record", Phase: phase, Subject: subject, PID: os.Getpid(), PPID: os.Getppid(), StartedAt: time.Now().UTC(), Outcome: "ok", Fields: fields})
}

func write(directory string, value event) {
	data, err := json.Marshal(value)
	if err != nil {
		return
	}
	data = append(data, '\n')
	writes.Lock()
	defer writes.Unlock()
	if os.MkdirAll(directory, 0700) != nil {
		return
	}
	file, err := os.OpenFile(filepath.Join(directory, fmt.Sprintf("%d.jsonl", os.Getpid())), os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0600)
	if err != nil {
		return
	}
	_, _ = file.Write(data) //nolint:errcheck // Optional timing must never turn compiler or test success into failure.
	_ = file.Close()        //nolint:errcheck // Optional timing is best-effort, including flushing its diagnostic file.
}
