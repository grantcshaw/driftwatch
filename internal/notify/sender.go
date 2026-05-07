package notify

import "github.com/yourorg/driftwatch/internal/drift"

// Sender is the common interface implemented by all notification backends.
type Sender interface {
	Send(env string, drifts []drift.Drift) error
}
