package notify

import (
	"fmt"
	"sync"

	"github.com/example/driftwatch/internal/drift"
)

// ShadowSender sends to a primary sender and a shadow sender in parallel,
// logging discrepancies between their outcomes without affecting the primary result.
type ShadowSender struct {
	mu      sync.Mutex
	primary Sender
	shadow  Sender
	log     []ShadowResult
}

// ShadowResult records the outcome of a single shadow comparison.
type ShadowResult struct {
	PrimaryErr error
	ShadowErr  error
	Mismatch   bool
}

// NewShadowSender creates a ShadowSender wrapping primary and shadow senders.
// Both primary and shadow must be non-nil.
func NewShadowSender(primary, shadow Sender) (*ShadowSender, error) {
	if primary == nil {
		return nil, fmt.Errorf("shadow: primary sender must not be nil")
	}
	if shadow == nil {
		return nil, fmt.Errorf("shadow: shadow sender must not be nil")
	}
	return &ShadowSender{primary: primary, shadow: shadow}, nil
}

// Send dispatches drifts to both primary and shadow senders concurrently.
// The primary result is returned; shadow errors are recorded internally.
func (s *ShadowSender) Send(env string, drifts []drift.Drift) error {
	if len(drifts) == 0 {
		return nil
	}

	type result struct {
		err error
	}

	primCh := make(chan result, 1)
	shadCh := make(chan result, 1)

	go func() { pErr := s.primary.Send(env, drifts); primCh <- result{pErr} }()
	go func() { sErr := s.shadow.Send(env, drifts); shadCh <- result{sErr} }()

	primRes := <-primCh
	shadRes := <-shadCh

	mismatch := (primRes.err == nil) != (shadRes.err == nil)

	s.mu.Lock()
	s.log = append(s.log, ShadowResult{
		PrimaryErr: primRes.err,
		ShadowErr:  shadRes.err,
		Mismatch:   mismatch,
	})
	s.mu.Unlock()

	return primRes.err
}

// Results returns a copy of all recorded shadow comparison results.
func (s *ShadowSender) Results() []ShadowResult {
	s.mu.Lock()
	defer s.mu.Unlock()
	out := make([]ShadowResult, len(s.log))
	copy(out, s.log)
	return out
}

// Reset clears the recorded shadow results.
func (s *ShadowSender) Reset() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.log = nil
}
