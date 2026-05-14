package notify

import (
	"errors"
	"math/rand"
	"sync"

	"github.com/driftwatch/internal/drift"
)

// Sampler wraps a Sender and probabilistically forwards drift events,
// sending only a configured fraction of alerts. Useful for high-volume
// environments where full alerting would be noisy.
type Sampler struct {
	mu       sync.Mutex
	inner    Sender
	rate     float64 // 0.0 to 1.0 inclusive
	randfn   func() float64
}

// NewSampler creates a Sampler that forwards alerts with probability rate.
// rate must be in the range (0.0, 1.0].
func NewSampler(inner Sender, rate float64) (*Sampler, error) {
	if inner == nil {
		return nil, errors.New("sampler: inner sender must not be nil")
	}
	if rate <= 0.0 || rate > 1.0 {
		return nil, errors.New("sampler: rate must be in range (0.0, 1.0]")
	}
	return &Sampler{
		inner:  inner,
		rate:   rate,
		randfn: rand.Float64,
	}, nil
}

// Send forwards drifts to the inner sender with probability equal to the
// configured rate. If no drifts are provided, Send is a no-op.
func (s *Sampler) Send(env string, drifts []drift.Drift) error {
	if len(drifts) == 0 {
		return nil
	}
	s.mu.Lock()
	sample := s.randfn() < s.rate
	s.mu.Unlock()
	if !sample {
		return nil
	}
	return s.inner.Send(env, drifts)
}
