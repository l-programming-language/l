package native

import (
	"context"
	"testing"
	"time"
)

func TestCommandPhaseSeparatesCompilationAndRuntimeBudgets(t *testing.T) {
	for _, runtimeTimeout := range []time.Duration{5 * time.Second, 12 * time.Second, 15 * time.Second, 20 * time.Second} {
		t.Run(runtimeTimeout.String(), func(t *testing.T) {
			started := time.Now()
			ctx, cancel := RuntimeExecution.Context(context.Background(), runtimeTimeout)
			deadline, bounded := ctx.Deadline()
			if !bounded || deadline.Before(started.Add(runtimeTimeout)) || deadline.After(time.Now().Add(runtimeTimeout)) {
				t.Fatalf("runtime watchdog changed from %s: %v, %v", runtimeTimeout, deadline, bounded)
			}
			cancel()
			ctx, cancel = BuildPrerequisite.Context(context.Background(), runtimeTimeout)
			if deadline, bounded := ctx.Deadline(); bounded || ctx.Err() != nil {
				t.Fatalf("compilation has a watchdog or canceled context: %v, %v", deadline, ctx.Err())
			}
			cancel()
			if ctx.Err() != context.Canceled {
				t.Fatalf("compilation no longer supports cancellation: %v", ctx.Err())
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
			_, cancel := tc.phase.Context(context.Background(), tc.runtimeTimeout)
			cancel()
		})
	}
}

func TestCommandPhasePreservesParentCancellation(t *testing.T) {
	for _, phase := range []CommandPhase{BuildPrerequisite, RuntimeExecution} {
		parent, cancelParent := context.WithCancel(context.Background())
		ctx, cancel := phase.Context(parent, time.Second)
		cancelParent()
		if ctx.Err() != context.Canceled {
			t.Fatalf("%s lost parent cancellation: %v", phase, ctx.Err())
		}
		cancel()
	}
}
