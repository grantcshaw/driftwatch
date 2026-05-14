package notify

import (
	"fmt"

	"github.com/yourorg/driftwatch/internal/drift"
)

// Envelope wraps a Sender and attaches computed severity metadata before
// forwarding, ensuring downstream senders always have severity populated.
type Envelope struct {
	env    string
	inner  Sender
}

// NewEnvelope creates an Envelope for the given environment name.
func NewEnvelope(env string, inner Sender) (*Envelope, error) {
	if env == "" {
		return nil, fmt.Errorf("envelope: env must not be empty")
	}
	if inner == nil {
		return nil, fmt.Errorf("envelope: inner sender must not be nil")
	}
	return &Envelope{env: env, inner: inner}, nil
}

// Send annotates each drift with a computed severity then forwards them.
func (e *Envelope) Send(env string, drifts []drift.Drift) error {
	if len(drifts) == 0 {
		return nil
	}
	annotated := make([]drift.Drift, len(drifts))
	for i, d := range drifts {
		if d.Severity == "" {
			d.Severity = computeSeverity(drifts)
		}
		annotated[i] = d
	}
	return e.inner.Send(env, annotated)
}

// computeSeverity returns "critical" when there are multiple drifts, else "warning".
func computeSeverity(drifts []drift.Drift) string {
	if len(drifts) > 1 {
		return "critical"
	}
	return "warning"
}
