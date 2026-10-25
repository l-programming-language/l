package timing

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
)

func TestDisabledTimingDoesNotWrite(t *testing.T) {
	t.Setenv("L_HARNESS_TIMING_DIR", "")
	if Enabled() {
		t.Fatal("timing enabled without its explicit directory")
	}
	Start("disabled", "subject")(errors.New("unchanged caller error"))
	Record("disabled", "subject", map[string]any{"count": 1})
}

func TestTimingRecordsOutcomesWithoutMixingConcurrentEvents(t *testing.T) {
	directory := t.TempDir()
	t.Setenv("L_HARNESS_TIMING_DIR", directory)
	if !Enabled() {
		t.Fatal("explicit timing directory ignored")
	}
	done := Start("failure", "source")
	done(errors.New("sentinel error"))
	done(nil) // completion is idempotent
	var workers sync.WaitGroup
	for index := 0; index < 24; index++ {
		workers.Add(1)
		go func() { defer workers.Done(); Start("work", "source")(nil) }()
	}
	workers.Wait()
	Record("cargo.artifacts", "source", map[string]any{"fresh": 3, "rebuilt": 2})
	paths, err := filepath.Glob(filepath.Join(directory, "*.jsonl"))
	if err != nil || len(paths) != 1 {
		t.Fatalf("per-process files=%v %v", paths, err)
	}
	data, err := os.ReadFile(paths[0])
	if err != nil {
		t.Fatal(err)
	}
	lines := strings.Split(strings.TrimSpace(string(data)), "\n")
	if len(lines) != 51 {
		t.Fatalf("events=%d want51", len(lines))
	}
	open := map[string]bool{}
	for i, line := range lines {
		var event struct {
			Event, Span, Phase, Subject, Outcome, Error string
			PID, PPID                                   int
			StartedAt                                   string `json:"started_at"`
			Duration                                    int64  `json:"duration_ns"`
			Fields                                      map[string]int
		}
		if err := json.Unmarshal([]byte(line), &event); err != nil {
			t.Fatal(err)
		}
		if event.PID != os.Getpid() || event.PPID != os.Getppid() || event.StartedAt == "" || event.Duration < 0 || event.Subject != "source" {
			t.Fatalf("invalid event %+v", event)
		}
		if event.Event == "start" {
			open[event.Span] = true
		}
		if event.Event == "finish" {
			if !open[event.Span] {
				t.Fatalf("finish without start %+v", event)
			}
			delete(open, event.Span)
		}
		if i == 1 && (event.Outcome != "error" || event.Error != "sentinel error") {
			t.Fatalf("lost error %+v", event)
		}
		if i == 50 && (event.Fields["fresh"] != 3 || event.Fields["rebuilt"] != 2) {
			t.Fatalf("lost metadata %+v", event)
		}
	}
	if len(open) != 0 {
		t.Fatalf("unclosed spans %v", open)
	}
}

func TestTimingFailureDoesNotChangeCallerBehavior(t *testing.T) {
	path := filepath.Join(t.TempDir(), "file")
	if err := os.WriteFile(path, []byte("preserve"), 0600); err != nil {
		t.Fatal(err)
	}
	t.Setenv("L_HARNESS_TIMING_DIR", path)
	Start("unwritable", "subject")(nil)
	Record("unwritable", "subject", nil)
	data, err := os.ReadFile(path)
	if err != nil || string(data) != "preserve" {
		t.Fatalf("diagnostic changed file=%q %v", data, err)
	}
}
