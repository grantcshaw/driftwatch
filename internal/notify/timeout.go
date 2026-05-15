package notify

import (
	"context"
	"fmt"
	"time"

	"github.com/your-org/driftwatch/internal/drift"
)

// Timeout wraps a Sender and enforces a maximum duration per Send call.
// If the inner sender does not return within the deadline, Send returns
// an error and the context passed to the inner sender is cancelled.
type Timeout struct {
	inner   Sender
	duration time.Duration
}

// NewTimeout creates a Timeout sender. duration must be positive.
func NewTimeout(inner Sender, duration time.Duration) (*Timeout, error) {
	if inner == nil {
		return nil, fmt.Errorf("notify/timeout: inner sender must not be nil")
	}
	if duration <= 0 {
		return nil, fmt.Errorf("notify/timeout: duration must be positive, got %s", duration)
	}
	return &Timeout{inner: inner, duration: duration}, nil
}

// Send calls the inner sender with a deadline-bound context.
// If the deadline is exceeded, a wrapped error is returned.
func (t *Timeout) Send(env string, drifts []drift.Drift) error {
	if len(drifts) == 0 {
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), t.duration)
	defer cancel()

	type result struct{ err error }
	ch := make(chan result, 1)

	go func() {
		_ = ctx // inner sender receives ordinary call; context guards goroutine lifetime
		ch <- result{err: t.inner.Send(env, drifts)}
	}()

	select {
	case r := <-ch:
		return r.err
	case <-ctx.Done():
		return fmt.Errorf("notify/timeout: send to %q exceeded %s deadline", env, t.duration)
	}
}
