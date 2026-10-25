package native

import (
	"testing"
	"time"
)

func TestCommandPhaseSeparatesCompilationAndRuntimeBudgets(t *testing.T) {
	for _, runtimeTimeout := range []time.Duration{5 * time.Second, 12 * time.Second, 15 * time.Second, 20 * time.Second} {
		t.Run(runtimeTimeout.String(), func(t *testing.T) {
			if got := RuntimeExecution.Timeout(runtimeTimeout); got != runtimeTimeout {
				t.Fatalf("runtime watchdog changed from %s to %s", runtimeTimeout, got)
			}
			if got := BuildPrerequisite.Timeout(runtimeTimeout); got != 40*time.Second {
				t.Fatalf("compilation inherited the runtime watchdog: %s", got)
			}
		})
	}
}

func TestCommandPhaseRejectsUnclassifiedOrUnboundedRuntime(t *testing.T) {
	for _, tc := range []struct {
		name           string
		phase          CommandPhase
		runtimeTimeout time.Duration
	}{
		{"unclassified", "", 20 * time.Second},
		{"unknown", "future phase", 20 * time.Second},
		{"zero runtime", RuntimeExecution, 0},
		{"negative runtime", RuntimeExecution, -time.Second},
	} {
		t.Run(tc.name, func(t *testing.T) {
			defer func() {
				if recover() == nil {
					t.Fatal("invalid phase or watchdog silently accepted")
				}
			}()
			_ = tc.phase.Timeout(tc.runtimeTimeout)
		})
	}
}
