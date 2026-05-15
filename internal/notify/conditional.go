package notify

import (
	"fmt"

	"github.com/your-org/driftwatch/internal/drift"
)

// ConditionFn decides whether a set of drifts should be forwarded to
// the inner Sender. It receives the environment name and the drift slice.
type ConditionFn func(env string, drifts []drift.Drift) bool

// Conditional forwards Send calls to the inner Sender only when the
// supplied ConditionFn returns true. When the condition is false the
// call is silently dropped and nil is returned.
type Conditional struct {
	inner Sender
	cond  ConditionFn
}

// NewConditional creates a Conditional sender.
func NewConditional(inner Sender, cond ConditionFn) (*Conditional, error) {
	if inner == nil {
		return nil, fmt.Errorf("notify/conditional: inner sender must not be nil")
	}
	if cond == nil {
		return nil, fmt.Errorf("notify/conditional: condition function must not be nil")
	}
	return &Conditional{inner: inner, cond: cond}, nil
}

// Send forwards drifts to the inner sender when the condition is met.
func (c *Conditional) Send(env string, drifts []drift.Drift) error {
	if len(drifts) == 0 {
		return nil
	}
	if !c.cond(env, drifts) {
		return nil
	}
	return c.inner.Send(env, drifts)
}

// MinDriftCount returns a ConditionFn that is true when at least n drifts
// are present.
func MinDriftCount(n int) ConditionFn {
	return func(_ string, drifts []drift.Drift) bool {
		return len(drifts) >= n
	}
}

// EnvMatches returns a ConditionFn that is true when the environment name
// equals one of the supplied names.
func EnvMatches(names ...string) ConditionFn {
	set := make(map[string]struct{}, len(names))
	for _, n := range names {
		set[n] = struct{}{}
	}
	return func(env string, _ []drift.Drift) bool {
		_, ok := set[env]
		return ok
	}
}
