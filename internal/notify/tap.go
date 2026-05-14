package notify

import (
	"context"
	"errors"
	"time"

	"github.com/example/driftwatch/internal/drift"
)

// TapRecord captures a single invocation of a Tap sender.
type TapRecord struct {
	Env    string
	Drifts []drift.Drift
	At     time.Time
}

// Tap wraps a Sender and records every Send call for inspection.
// It is primarily useful in tests and debugging pipelines.
type Tap struct {
	inner   Sender
	Records []TapRecord
}

// NewTap creates a Tap around inner.
func NewTap(inner Sender) (*Tap, error) {
	if inner == nil {
		return nil, errors.New("tap: inner sender must not be nil")
	}
	return &Tap{inner: inner}, nil
}

// Send records the call then delegates to the inner sender.
func (t *Tap) Send(ctx context.Context, env string, drifts []drift.Drift) error {
	t.Records = append(t.Records, TapRecord{
		Env:    env,
		Drifts: drifts,
		At:     time.Now(),
	})
	return t.inner.Send(ctx, env, drifts)
}

// Len returns the number of recorded calls.
func (t *Tap) Len() int {
	return len(t.Records)
}

// Reset clears all recorded calls.
func (t *Tap) Reset() {
	t.Records = nil
}
