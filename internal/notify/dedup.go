package notify

import (
	"crypto/sha256"
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/yourorg/driftwatch/internal/drift"
)

// Dedup wraps a Sender and suppresses duplicate drift notifications
// within a configurable window. Two drift sets are considered duplicates
// if they contain the same keys with the same values for a given environment.
type Dedup struct {
	inner  Sender
	window time.Duration

	mu   sync.Mutex
	seen map[string]time.Time // fingerprint -> last sent
}

// NewDedup creates a Dedup sender. window is the suppression period;
// identical drift fingerprints within that window are dropped.
func NewDedup(inner Sender, window time.Duration) (*Dedup, error) {
	if inner == nil {
		return nil, fmt.Errorf("dedup: inner sender must not be nil")
	}
	if window <= 0 {
		return nil, fmt.Errorf("dedup: window must be positive, got %s", window)
	}
	return &Dedup{
		inner:  inner,
		window: window,
		seen:   make(map[string]time.Time),
	}, nil
}

// Send forwards drifts to the inner sender only if the fingerprint has
// not been seen within the dedup window.
func (d *Dedup) Send(env string, drifts []drift.Drift) error {
	if len(drifts) == 0 {
		return nil
	}

	fp := fingerprint(env, drifts)

	d.mu.Lock()
	last, exists := d.seen[fp]
	if exists && time.Since(last) < d.window {
		d.mu.Unlock()
		return nil
	}
	d.seen[fp] = time.Now()
	d.mu.Unlock()

	return d.inner.Send(env, drifts)
}

// fingerprint builds a stable hash from the environment name and drift slice.
func fingerprint(env string, drifts []drift.Drift) string {
	keys := make([]string, 0, len(drifts))
	for _, d := range drifts {
		keys = append(keys, fmt.Sprintf("%s=%s->%s", d.Key, d.BaselineValue, d.CurrentValue))
	}
	sort.Strings(keys)
	h := sha256.Sum256([]byte(env + "\x00" + strings.Join(keys, "\x01")))
	return fmt.Sprintf("%x", h)
}
