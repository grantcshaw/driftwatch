package notify

import (
	"errors"
	"fmt"
	"time"

	"github.com/org/driftwatch/internal/drift"
)

// Batch collects drifts over a time window or until a max size is reached,
// then forwards them to an inner Sender as a single call.
type Batch struct {
	inner    Sender
	window   time.Duration
	maxSize  int
	pending  []drift.Drift
	flushAt  time.Time
}

// NewBatch creates a Batch sender.
// window is the maximum time to hold drifts before flushing.
// maxSize is the maximum number of drifts to accumulate before an early flush.
func NewBatch(inner Sender, window time.Duration, maxSize int) (*Batch, error) {
	if inner == nil {
		return nil, errors.New("batch: inner sender must not be nil")
	}
	if window <= 0 {
		return nil, errors.New("batch: window must be positive")
	}
	if maxSize <= 0 {
		return nil, errors.New("batch: maxSize must be positive")
	}
	return &Batch{
		inner:   inner,
		window:  window,
		maxSize: maxSize,
		flushAt: time.Now().Add(window),
	}, nil
}

// Send accumulates drifts. It flushes to the inner sender when the window
// expires or the batch reaches maxSize.
func (b *Batch) Send(env string, drifts []drift.Drift) error {
	if len(drifts) == 0 {
		return nil
	}
	b.pending = append(b.pending, drifts...)
	if len(b.pending) >= b.maxSize || time.Now().After(b.flushAt) {
		return b.Flush(env)
	}
	return nil
}

// Flush forces immediate delivery of all pending drifts to the inner sender.
func (b *Batch) Flush(env string) error {
	if len(b.pending) == 0 {
		return nil
	}
	snapshot := make([]drift.Drift, len(b.pending))
	copy(snapshot, b.pending)
	b.pending = b.pending[:0]
	b.flushAt = time.Now().Add(b.window)
	if err := b.inner.Send(env, snapshot); err != nil {
		return fmt.Errorf("batch: flush failed: %w", err)
	}
	return nil
}

// Len returns the number of drifts currently pending in the batch.
func (b *Batch) Len() int {
	return len(b.pending)
}
