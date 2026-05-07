package notify

import (
	"bytes"
	"fmt"
	"sync"
	"time"

	"github.com/yourorg/driftwatch/internal/drift"
)

// Digest batches drift notifications over a time window and sends a single
// aggregated message to the inner Sender when the window elapses or Flush is
// called explicitly.
type Digest struct {
	mu       sync.Mutex
	inner    Sender
	window   time.Duration
	buf      map[string][]drift.Drift // keyed by environment name
	timer    *time.Timer
	stopped  bool
}

// NewDigest creates a Digest that accumulates drifts for the given window
// duration before forwarding them as a single batch to inner.
func NewDigest(inner Sender, window time.Duration) (*Digest, error) {
	if inner == nil {
		return nil, fmt.Errorf("digest: inner sender must not be nil")
	}
	if window <= 0 {
		return nil, fmt.Errorf("digest: window must be positive, got %s", window)
	}
	d := &Digest{
		inner:  inner,
		window: window,
		buf:    make(map[string][]drift.Drift),
	}
	d.resetTimer()
	return d, nil
}

// Send buffers the provided drifts. The actual notification is deferred until
// the digest window expires or Flush is called.
func (d *Digest) Send(env string, drifts []drift.Drift) error {
	if len(drifts) == 0 {
		return nil
	}
	d.mu.Lock()
	defer d.mu.Unlock()
	d.buf[env] = append(d.buf[env], drifts...)
	return nil
}

// Flush immediately sends all buffered drifts and resets the internal buffer.
func (d *Digest) Flush() error {
	d.mu.Lock()
	payload := d.buf
	d.buf = make(map[string][]drift.Drift)
	d.mu.Unlock()

	var errs []error
	for env, drifts := range payload {
		if err := d.inner.Send(env, drifts); err != nil {
			errs = append(errs, fmt.Errorf("env %s: %w", env, err))
		}
	}
	if len(errs) > 0 {
		return joinErrors(errs)
	}
	return nil
}

// Stop cancels the background flush timer.
func (d *Digest) Stop() {
	d.mu.Lock()
	defer d.mu.Unlock()
	if d.timer != nil {
		d.timer.Stop()
	}
	d.stopped = true
}

func (d *Digest) resetTimer() {
	if d.timer != nil {
		d.timer.Stop()
	}
	d.timer = time.AfterFunc(d.window, func() {
		_ = d.Flush()
		d.mu.Lock()
		stopped := d.stopped
		d.mu.Unlock()
		if !stopped {
			d.mu.Lock()
			d.resetTimer()
			d.mu.Unlock()
		}
	})
}

func joinErrors(errs []error) error {
	var buf bytes.Buffer
	for i, e := range errs {
		if i > 0 {
			buf.WriteString("; ")
		}
		buf.WriteString(e.Error())
	}
	return fmt.Errorf(buf.String())
}
