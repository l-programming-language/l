//go:build darwin || linux

package phase1

import (
	"context"
	"time"
)

// publicationRuntimeBudget charges every active runtime phase to one allowance.
// A paused build has its own watchdog; resuming never renews spent runtime time.
// Calls belong to the single publication-test goroutine.
type publicationRuntimeBudget struct {
	remaining time.Duration
	now       func() time.Time
	started   time.Time
	context   context.Context
	cancel    context.CancelFunc
}

func (b *publicationRuntimeBudget) resume() context.Context {
	if b.cancel != nil {
		return b.context
	}
	b.started = b.now()
	b.context, b.cancel = context.WithTimeout(context.Background(), b.remaining)
	return b.context
}

func (b *publicationRuntimeBudget) pause() error {
	if b.cancel != nil {
		b.remaining -= b.now().Sub(b.started)
		err := b.context.Err()
		b.cancel()
		b.cancel = nil
		if err != nil {
			b.remaining = 0
			return err
		}
	}
	if b.remaining <= 0 {
		b.remaining = 0
		return context.DeadlineExceeded
	}
	return nil
}

func (b *publicationRuntimeBudget) close() {
	if b.cancel != nil {
		b.cancel()
	}
}
