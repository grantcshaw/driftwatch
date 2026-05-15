package notify

import (
	"errors"
	"sync"
	"time"

	"github.com/yourorg/driftwatch/internal/drift"
)

// Window collects drift events within a rolling time window and forwards
// only those that arrive after the window has accumulated at least MinCount
// distinct drift entries. This is useful for suppressing noise from
// transient flaps that self-correct quickly.
type Window struct {
	inner    Sender
	duration time.Duration
	minCount int

	mu      sync.Mutex
	bucket  []drift.Drift
	windowStart time.Time
}

// NewWindow creates a Window sender that buffers drifts for dur and only
// forwards them when at least minCount unique keys are drifting.
func NewWindow(inner Sender, dur time.Duration, minCount int) (*Window, error) {
	if inner == nil {
		return nil, errors.New("window: inner sender must not be nil")
	}
	if dur <= 0 {
		return nil, errors.New("window: duration must be positive")
	}
	if minCount < 1 {
		return nil, errors.New("window: minCount must be at least 1")
	}
	return &Window{
		inner:    inner,
		duration: dur,
		minCount: minCount,
	}, nil
}

// Send buffers drifts within the current window and forwards them when the
// threshold is met. The window resets after a successful forward.
func (w *Window) Send(env string, drifts []drift.Drift) error {
	if len(drifts) == 0 {
		return nil
	}

	w.mu.Lock()
	defer w.mu.Unlock()

	now := time.Now()
	if w.windowStart.IsZero() || now.Sub(w.windowStart) > w.duration {
		w.bucket = nil
		w.windowStart = now
	}

	w.bucket = mergeDrifts(w.bucket, drifts)

	if len(w.bucket) >= w.minCount {
		send := make([]drift.Drift, len(w.bucket))
		copy(send, w.bucket)
		w.bucket = nil
		w.windowStart = time.Time{}
		return w.inner.Send(env, send)
	}
	return nil
}

// mergeDrifts appends new drifts, deduplicating by key.
func mergeDrifts(existing, incoming []drift.Drift) []drift.Drift {
	seen := make(map[string]struct{}, len(existing))
	for _, d := range existing {
		seen[d.Key] = struct{}{}
	}
	result := append([]drift.Drift(nil), existing...)
	for _, d := range incoming {
		if _, ok := seen[d.Key]; !ok {
			seen[d.Key] = struct{}{}
			result = append(result, d)
		}
	}
	return result
}
