package native

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/l-programming-language/l/internal/timing"
)

// Retain diagnostic counters and the actual Cargo report while the cache slot
// is still owned. Neither is a build input or a substitute for artifact checks.
func recordCargoTiming(subject, target string, started time.Time, output []byte) {
	if !timing.Enabled() {
		return
	}
	fresh, rebuilt := 0, 0
	for _, line := range bytes.Split(output, []byte{'\n'}) {
		var event struct {
			Reason string
			Fresh  bool
		}
		if json.Unmarshal(line, &event) == nil && event.Reason == "compiler-artifact" {
			if event.Fresh {
				fresh++
			} else {
				rebuilt++
			}
		}
	}
	fields := map[string]any{"fresh": fresh, "rebuilt": rebuilt}
	report := filepath.Join(target, "cargo-timings", "cargo-timing.html")
	if info, err := os.Stat(report); err == nil && !info.ModTime().Before(started) {
		if data, err := os.ReadFile(report); err == nil {
			directory := filepath.Join(os.Getenv("L_HARNESS_TIMING_DIR"), "cargo-timings")
			if os.MkdirAll(directory, 0700) == nil {
				destination := filepath.Join(directory, fmt.Sprintf("%d-%d.html", os.Getpid(), started.UnixNano()))
				if os.WriteFile(destination, data, 0600) == nil {
					fields["report"] = destination
				}
			}
		}
	}
	timing.Record("cargo.artifacts", subject, fields)
}
