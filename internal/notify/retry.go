package notify

import (
	"fmt"
	"time"

	"github.com/org/driftwatch/internal/drift"
)

// Retry wraps a Sender and retries on failure up to MaxAttempts times.
type Retry struct {
	inner       Sender
	maxAttempts int
	delay       time.Duration
}

// NewRetry creates a Retry sender. maxAttempts must be >= 1; delay is the
// pause between attempts (0 means no pause).
func NewRetry(inner Sender, maxAttempts int, delay time.Duration) (*Retry, error) {
	if inner == nil {
		return nil, fmt.Errorf("retry: inner sender must not be nil")
	}
	if maxAttempts < 1 {
		return nil, fmt.Errorf("retry: maxAttempts must be >= 1, got %d", maxAttempts)
	}
	return &Retry{
		inner:       inner,
		maxAttempts: maxAttempts,
		delay:       delay,
	}, nil
}

// Send attempts to deliver drifts, retrying up to maxAttempts times on error.
// It returns the last error if all attempts fail, or nil on success.
func (r *Retry) Send(env string, drifts []drift.Drift) error {
	var lastErr error
	for attempt := 1; attempt <= r.maxAttempts; attempt++ {
		if err := r.inner.Send(env, drifts); err == nil {
			return nil
		} else {
			lastErr = err
		}
		if attempt < r.maxAttempts && r.delay > 0 {
			time.Sleep(r.delay)
		}
	}
	return fmt.Errorf("retry: all %d attempts failed for env %q: %w", r.maxAttempts, env, lastErr)
}
