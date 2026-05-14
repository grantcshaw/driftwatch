package notify

import (
	"fmt"
	"strings"

	"github.com/yourorg/driftwatch/internal/drift"
)

// EnvelopeRouter routes drift envelopes to different senders based on selectors.
type EnvelopeRouter struct {
	routes []routeEntry
	fallback Sender
}

type routeEntry struct {
	selector Selector
	sender   Sender
}

// Selector defines criteria for matching an envelope.
type Selector struct {
	// Env matches envelopes for this environment name (empty = any).
	Env string
	// MinSeverity is the minimum severity to match ("warning" or "critical").
	MinSeverity string
}

// NewEnvelopeRouter creates a router with an optional fallback sender.
func NewEnvelopeRouter(fallback Sender) *EnvelopeRouter {
	return &EnvelopeRouter{fallback: fallback}
}

// AddRoute registers a sender for drifts matching the given selector.
func (r *EnvelopeRouter) AddRoute(sel Selector, s Sender) error {
	if s == nil {
		return fmt.Errorf("envelope_router: sender must not be nil")
	}
	if sel.MinSeverity != "" && sel.MinSeverity != "warning" && sel.MinSeverity != "critical" {
		return fmt.Errorf("envelope_router: invalid min_severity %q", sel.MinSeverity)
	}
	r.routes = append(r.routes, routeEntry{selector: sel, sender: s})
	return nil
}

// Send routes the drifts to every matching sender; falls back if none match.
func (r *EnvelopeRouter) Send(env string, drifts []drift.Drift) error {
	if len(drifts) == 0 {
		return nil
	}
	var matched bool
	var errs []string
	for _, route := range r.routes {
		if !matchesSelector(route.selector, env, drifts) {
			continue
		}
		matched = true
		if err := route.sender.Send(env, drifts); err != nil {
			errs = append(errs, err.Error())
		}
	}
	if !matched && r.fallback != nil {
		if err := r.fallback.Send(env, drifts); err != nil {
			errs = append(errs, err.Error())
		}
	}
	if len(errs) > 0 {
		return fmt.Errorf("envelope_router: %s", strings.Join(errs, "; "))
	}
	return nil
}

func matchesSelector(sel Selector, env string, drifts []drift.Drift) bool {
	if sel.Env != "" && sel.Env != env {
		return false
	}
	if sel.MinSeverity == "critical" {
		for _, d := range drifts {
			if d.Severity == "critical" {
				return true
			}
		}
		return false
	}
	return true
}
