package notify

import (
	"errors"
	"strings"

	"github.com/yourorg/driftwatch/internal/drift"
)

// TransformFunc is a function that mutates a drift entry before sending.
type TransformFunc func(d drift.Drift) drift.Drift

// Transform wraps a Sender and applies one or more transformation functions
// to each drift entry before forwarding to the inner sender.
type Transform struct {
	inner Sender
	funcs []TransformFunc
}

// NewTransform creates a Transform sender wrapping inner with the given
// transformation functions applied in order. Returns an error if inner is
// nil or no transform functions are provided.
func NewTransform(inner Sender, fns ...TransformFunc) (*Transform, error) {
	if inner == nil {
		return nil, errors.New("transform: inner sender must not be nil")
	}
	if len(fns) == 0 {
		return nil, errors.New("transform: at least one TransformFunc is required")
	}
	return &Transform{inner: inner, funcs: fns}, nil
}

// Send applies each TransformFunc to every drift entry and forwards the
// transformed slice to the inner sender. If drifts is empty, Send returns nil
// without calling the inner sender.
func (t *Transform) Send(env string, drifts []drift.Drift) error {
	if len(drifts) == 0 {
		return nil
	}
	out := make([]drift.Drift, len(drifts))
	for i, d := range drifts {
		for _, fn := range t.fns {
			d = fn(d)
		}
		out[i] = d
	}
	return t.inner.Send(env, out)
}

// RedactValue returns a TransformFunc that replaces the ActualValue and
// ExpectedValue of any drift whose key matches one of the provided key
// prefixes with the string "[REDACTED]".
func RedactValue(keyPrefixes ...string) TransformFunc {
	return func(d drift.Drift) drift.Drift {
		for _, prefix := range keyPrefixes {
			if strings.HasPrefix(d.Key, prefix) {
				d.ActualValue = "[REDACTED]"
				d.ExpectedValue = "[REDACTED]"
				return d
			}
		}
		return d
	}
}

// UpperCaseKey returns a TransformFunc that converts every drift key to
// upper-case.
func UpperCaseKey() TransformFunc {
	return func(d drift.Drift) drift.Drift {
		d.Key = strings.ToUpper(d.Key)
		return d
	}
}
