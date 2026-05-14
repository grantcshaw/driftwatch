package notify

import (
	"errors"
	"sort"

	"github.com/org/driftwatch/internal/drift"
)

// Priority wraps a Sender and reorders drifts before forwarding them,
// placing critical-severity drifts first so downstream senders and humans
// see the most important changes at the top.
type Priority struct {
	inner Sender
}

// NewPriority creates a Priority sender that sorts drifts by severity
// (critical before warning) before delegating to inner.
func NewPriority(inner Sender) (*Priority, error) {
	if inner == nil {
		return nil, errors.New("priority: inner sender must not be nil")
	}
	return &Priority{inner: inner}, nil
}

// Send sorts drifts so critical entries precede warnings, then forwards
// the sorted slice to the inner sender.
func (p *Priority) Send(env string, drifts []drift.Drift) error {
	if len(drifts) == 0 {
		return nil
	}
	sorted := make([]drift.Drift, len(drifts))
	copy(sorted, drifts)
	sort.SliceStable(sorted, func(i, j int) bool {
		return severityRank(sorted[i].Severity) > severityRank(sorted[j].Severity)
	})
	return p.inner.Send(env, sorted)
}

// severityRank maps a severity string to a numeric rank for ordering.
func severityRank(s string) int {
	switch s {
	case "critical":
		return 2
	case "warning":
		return 1
	default:
		return 0
	}
}
