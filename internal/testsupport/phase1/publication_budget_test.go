//go:build darwin || linux

package phase1

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestPublicationRuntimeBudgetPreservesRemainder(t *testing.T) {
	now := time.Unix(0, 0)
	budget := publicationRuntimeBudget{remaining: 20 * time.Second, now: func() time.Time { return now }}
	defer budget.close()
	first := budget.resume()
	now = now.Add(7 * time.Second)
	if budget.resume() != first {
		t.Fatal("resuming an active phase renewed its context")
	}
	if err := budget.pause(); err != nil {
		t.Fatal(err)
	}
	if first.Err() != context.Canceled || budget.remaining != 13*time.Second {
		t.Fatalf("first phase: context=%v remaining=%v", first.Err(), budget.remaining)
	}

	// A full build-prerequisite interval must neither consume runtime time nor
	// reset the allowance used by the already-running old orchestrator.
	now = now.Add(40 * time.Second)
	if err := budget.pause(); err != nil || budget.remaining != 13*time.Second {
		t.Fatalf("paused build: error=%v remaining=%v", err, budget.remaining)
	}
	second := budget.resume()
	if second == first || second.Err() != nil {
		t.Fatalf("resumed phase did not get a live context: %v", second.Err())
	}
	now = now.Add(5 * time.Second)
	if err := budget.pause(); err != nil || budget.remaining != 8*time.Second {
		t.Fatalf("second phase: error=%v remaining=%v", err, budget.remaining)
	}
	now = now.Add(40 * time.Second)
	budget.resume()
	now = now.Add(8 * time.Second)
	if err := budget.pause(); !errors.Is(err, context.DeadlineExceeded) || budget.remaining != 0 {
		t.Fatalf("cumulative limit: error=%v remaining=%v", err, budget.remaining)
	}
	if err := budget.resume().Err(); !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("exhausted allowance was renewed: %v", err)
	}
}

func TestPublicationRuntimeBudgetDoesNotReviveExpiredPhase(t *testing.T) {
	for _, elapsed := range []time.Duration{20 * time.Second, 21 * time.Second} {
		t.Run(elapsed.String(), func(t *testing.T) {
			now := time.Unix(0, 0)
			budget := publicationRuntimeBudget{remaining: 20 * time.Second, now: func() time.Time { return now }}
			defer budget.close()
			budget.resume()
			now = now.Add(elapsed)
			if err := budget.pause(); !errors.Is(err, context.DeadlineExceeded) {
				t.Fatalf("expired runtime was allowed to enter build phase: %v", err)
			}
			now = now.Add(40 * time.Second)
			if err := budget.resume().Err(); !errors.Is(err, context.DeadlineExceeded) {
				t.Fatalf("expired phase was renewed after build: %v", err)
			}
		})
	}
}
