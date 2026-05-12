package notify

import (
	"errors"
	"math"
	"time"

	"github.com/your-org/driftwatch/internal/drift"
)

// BackoffSender wraps a Sender and applies exponential backoff between retry
// attempts. Unlike Retry (which retries immediately), BackoffSender waits an
// increasing delay between each attempt.
type BackoffSender struct {
	inner      Sender
	attempts   int
	baseDelay  time.Duration
	maxDelay   time.Duration
	sleep      func(time.Duration) // injectable for tests
}

// NewBackoffSender creates a BackoffSender wrapping inner.
// attempts is the total number of tries (>= 1).
// baseDelay is the initial wait duration; it doubles on each failure up to maxDelay.
func NewBackoffSender(inner Sender, attempts int, baseDelay, maxDelay time.Duration) (*BackoffSender, error) {
	if inner == nil {
		return nil, errors.New("backoff: inner sender must not be nil")
	}
	if attempts < 1 {
		return nil, errors.New("backoff: attempts must be >= 1")
	}
	if baseDelay <= 0 {
		return nil, errors.New("backoff: baseDelay must be positive")
	}
	if maxDelay < baseDelay {
		return nil, errors.New("backoff: maxDelay must be >= baseDelay")
	}
	return &BackoffSender{
		inner:     inner,
		attempts:  attempts,
		baseDelay: baseDelay,
		maxDelay:  maxDelay,
		sleep:     time.Sleep,
	}, nil
}

// Send attempts delivery up to b.attempts times, sleeping an exponentially
// growing delay between failures. Returns nil on first success.
func (b *BackoffSender) Send(env string, drifts []drift.Drift) error {
	if len(drifts) == 0 {
		return nil
	}
	var lastErr error
	for i := 0; i < b.attempts; i++ {
		if err := b.inner.Send(env, drifts); err == nil {
			return nil
		} else {
			lastErr = err
		}
		if i < b.attempts-1 {
			delay := time.Duration(float64(b.baseDelay) * math.Pow(2, float64(i)))
			if delay > b.maxDelay {
				delay = b.maxDelay
			}
			b.sleep(delay)
		}
	}
	return lastErr
}
