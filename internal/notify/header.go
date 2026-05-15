package notify

import (
	"errors"

	"github.com/example/driftwatch/internal/drift"
)

// HeaderSender wraps an inner Sender and injects static metadata headers
// (as key/value label pairs) into every drift's Labels map before forwarding.
// This is useful for tagging notifications with deployment region, cluster,
// or pipeline identifiers without modifying upstream producers.
type HeaderSender struct {
	inner   Sender
	headers map[string]string
}

// NewHeaderSender creates a HeaderSender that injects the given headers into
// each drift before delegating to inner. Returns an error if inner is nil or
// headers is empty.
func NewHeaderSender(inner Sender, headers map[string]string) (*HeaderSender, error) {
	if inner == nil {
		return nil, errors.New("header: inner sender must not be nil")
	}
	if len(headers) == 0 {
		return nil, errors.New("header: headers map must not be empty")
	}
	// copy to avoid external mutation
	h := make(map[string]string, len(headers))
	for k, v := range headers {
		h[k] = v
	}
	return &HeaderSender{inner: inner, headers: h}, nil
}

// Send injects the configured headers into each drift's Labels map and
// forwards the modified slice to the inner sender. Original drift values are
// not mutated; a shallow copy of each drift is used.
func (s *HeaderSender) Send(env string, drifts []drift.Drift) error {
	if len(drifts) == 0 {
		return nil
	}
	tagged := make([]drift.Drift, len(drifts))
	for i, d := range drifts {
		merged := make(map[string]string, len(d.Labels)+len(s.headers))
		for k, v := range d.Labels {
			merged[k] = v
		}
		for k, v := range s.headers {
			merged[k] = v
		}
		copy := d
		copy.Labels = merged
		tagged[i] = copy
	}
	return s.inner.Send(env, tagged)
}
