package notify

import (
	"fmt"
	"sync"
	"time"
)

// ThrottleConfig holds configuration for the Throttle sender wrapper.
type ThrottleConfig struct {
	// MinInterval is the minimum duration between sends for the same environment.
	MinInterval time.Duration
}

// Throttle wraps a Sender and suppresses sends that occur too frequently
// for a given environment key.
type Throttle struct {
	mu       sync.Mutex
	inner    Sender
	interval time.Duration
	lastSent map[string]time.Time
}

// NewThrottle creates a Throttle that forwards to inner no more than once per
// cfg.MinInterval for each environment.
func NewThrottle(inner Sender, cfg ThrottleConfig) (*Throttle, error) {
	if inner == nil {
		return nil, fmt.Errorf("throttle: inner sender must not be nil")
	}
	if cfg.MinInterval <= 0 {
		return nil, fmt.Errorf("throttle: MinInterval must be positive, got %s", cfg.MinInterval)
	}
	return &Throttle{
		inner:    inner,
		interval: cfg.MinInterval,
		lastSent: make(map[string]time.Time),
	}, nil
}

// Send forwards drifts to the inner Sender only if the minimum interval has
// elapsed since the last send for env. If suppressed, Send returns nil.
func (t *Throttle) Send(env string, drifts []DriftEntry) error {
	if len(drifts) == 0 {
		return nil
	}

	t.mu.Lock()
	last, seen := t.lastSent[env]
	if seen && time.Since(last) < t.interval {
		t.mu.Unlock()
		return nil
	}
	t.lastSent[env] = time.Now()
	t.mu.Unlock()

	return t.inner.Send(env, drifts)
}

// Reset clears the last-sent timestamp for env, allowing the next Send to
// proceed immediately regardless of the configured interval.
func (t *Throttle) Reset(env string) {
	t.mu.Lock()
	delete(t.lastSent, env)
	t.mu.Unlock()
}
