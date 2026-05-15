package notify

import (
	"errors"
	"strings"

	"github.com/driftwatch/driftwatch/internal/drift"
)

// NormalizeFunc transforms a single string value.
type NormalizeFunc func(string) string

// Normalize applies normalization functions to drift keys and values
// before forwarding to the inner Sender.
type Normalize struct {
	inner    Sender
	keyFns   []NormalizeFunc
	valueFns []NormalizeFunc
}

// NewNormalize creates a Normalize sender wrapper.
// keyFns are applied (in order) to each drift key.
// valueFns are applied (in order) to baseline and current values.
func NewNormalize(inner Sender, keyFns, valueFns []NormalizeFunc) (*Normalize, error) {
	if inner == nil {
		return nil, errors.New("notify/normalize: inner sender must not be nil")
	}
	if len(keyFns) == 0 && len(valueFns) == 0 {
		return nil, errors.New("notify/normalize: at least one key or value function must be provided")
	}
	return &Normalize{inner: inner, keyFns: keyFns, valueFns: valueFns}, nil
}

// Send applies normalization to all drift items and forwards to inner.
func (n *Normalize) Send(env string, drifts []drift.Drift) error {
	if len(drifts) == 0 {
		return nil
	}
	normalized := make([]drift.Drift, len(drifts))
	for i, d := range drifts {
		normalized[i] = drift.Drift{
			Key:      applyFns(d.Key, n.keyFns),
			Baseline: applyFns(d.Baseline, n.valueFns),
			Current:  applyFns(d.Current, n.valueFns),
		}
	}
	return n.inner.Send(env, normalized)
}

func applyFns(s string, fns []NormalizeFunc) string {
	for _, fn := range fns {
		s = fn(s)
	}
	return s
}

// TrimSpace is a NormalizeFunc that trims leading and trailing whitespace.
func TrimSpace(s string) string { return strings.TrimSpace(s) }

// LowerCase is a NormalizeFunc that lowercases a string.
func LowerCase(s string) string { return strings.ToLower(s) }
