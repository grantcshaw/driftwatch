package notify

import (
	"errors"
	"fmt"

	"github.com/driftwatch/internal/drift"
)

// Truncate is a Sender middleware that limits the number of drift entries
// forwarded to the inner sender. If the total number of drifts exceeds
// MaxItems, only the first MaxItems entries are forwarded and a summary
// entry is appended indicating how many were dropped.
type Truncate struct {
	inner    Sender
	maxItems int
}

// NewTruncate creates a Truncate sender wrapping inner.
// maxItems must be >= 1.
func NewTruncate(inner Sender, maxItems int) (*Truncate, error) {
	if inner == nil {
		return nil, errors.New("truncate: inner sender must not be nil")
	}
	if maxItems < 1 {
		return nil, errors.New("truncate: maxItems must be at least 1")
	}
	return &Truncate{inner: inner, maxItems: maxItems}, nil
}

// Send forwards at most maxItems drifts to the inner sender.
// If drifts are truncated a synthetic summary entry is appended.
func (t *Truncate) Send(env string, drifts []drift.Drift) error {
	if len(drifts) == 0 {
		return nil
	}
	if len(drifts) <= t.maxItems {
		return t.inner.Send(env, drifts)
	}

	truncated := make([]drift.Drift, t.maxItems)
	copy(truncated, drifts[:t.maxItems])

	dropped := len(drifts) - t.maxItems
	summary := drift.Drift{
		Key:      "__truncated__",
		Baseline: "",
		Current:  fmt.Sprintf("%d additional drift(s) suppressed", dropped),
	}
	truncated = append(truncated, summary)

	return t.inner.Send(env, truncated)
}
