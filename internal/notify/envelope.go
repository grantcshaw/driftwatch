package notify

import (
	"fmt"
	"time"

	"github.com/yourorg/driftwatch/internal/drift"
)

// Severity represents the urgency level of a drift notification.
type Severity string

const (
	SeverityWarning  Severity = "warning"
	SeverityCritical Severity = "critical"
)

// Envelope wraps a set of drifts with routing metadata used by
// pipeline stages (filter, throttle, priority, etc.).
type Envelope struct {
	// Environment the drifts belong to.
	Environment string

	// Drifts is the list of detected drift items.
	Drifts []drift.Drift

	// Severity is the highest severity across all drifts.
	Severity Severity

	// DetectedAt is when the drift was first observed.
	DetectedAt time.Time

	// Labels are arbitrary key/value pairs for routing or filtering.
	Labels map[string]string
}

// NewEnvelope creates an Envelope, computing the aggregate severity
// from the provided drifts.
func NewEnvelope(env string, drifts []drift.Drift) (*Envelope, error) {
	if env == "" {
		return nil, fmt.Errorf("envelope: environment name must not be empty")
	}
	sev := computeSeverity(drifts)
	return &Envelope{
		Environment: env,
		Drifts:      drifts,
		Severity:    sev,
		DetectedAt:  time.Now().UTC(),
		Labels:      make(map[string]string),
	}, nil
}

// WithLabel returns the envelope with an additional label set.
func (e *Envelope) WithLabel(key, value string) *Envelope {
	e.Labels[key] = value
	return e
}

// IsCritical reports whether the envelope carries critical severity.
func (e *Envelope) IsCritical() bool {
	return e.Severity == SeverityCritical
}

// computeSeverity returns critical if any drift has more than one changed
// key, otherwise warning. An empty slice returns warning.
func computeSeverity(drifts []drift.Drift) Severity {
	for _, d := range drifts {
		if len(d.Changes) > 1 {
			return SeverityCritical
		}
	}
	return SeverityWarning
}
