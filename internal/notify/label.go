package notify

import (
	"errors"

	"github.com/yourorg/driftwatch/internal/drift"
)

// LabelSender wraps an inner Sender and attaches static metadata labels
// to every drift before forwarding. Labels are merged into each drift's
// Metadata map; existing keys are not overwritten.
type LabelSender struct {
	inner  Sender
	labels map[string]string
}

// NewLabelSender returns a LabelSender that injects the provided labels.
// Returns an error if inner is nil or labels is empty.
func NewLabelSender(inner Sender, labels map[string]string) (*LabelSender, error) {
	if inner == nil {
		return nil, errors.New("label: inner sender must not be nil")
	}
	if len(labels) == 0 {
		return nil, errors.New("label: labels map must not be empty")
	}
	copy := make(map[string]string, len(labels))
	for k, v := range labels {
		copy[k] = v
	}
	return &LabelSender{inner: inner, labels: copy}, nil
}

// Send attaches labels to each drift and forwards to the inner sender.
func (l *LabelSender) Send(env string, drifts []drift.Drift) error {
	if len(drifts) == 0 {
		return nil
	}
	labeled := make([]drift.Drift, len(drifts))
	for i, d := range drifts {
		meta := make(map[string]string, len(l.labels))
		for k, v := range l.labels {
			meta[k] = v
		}
		// Preserve any existing metadata, overwriting only missing keys.
		for k, v := range d.Metadata {
			meta[k] = v
		}
		d.Metadata = meta
		labeled[i] = d
	}
	return l.inner.Send(env, labeled)
}
