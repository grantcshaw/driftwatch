package notify

import (
	"context"
	"fmt"
	"sync"
)

// Route associates a label selector with a Sender. An envelope is
// dispatched to the route if all selector key/value pairs appear in
// the envelope's Labels map. An empty selector matches every envelope.
type Route struct {
	Selector map[string]string
	Sender   Sender
}

// EnvelopeRouter dispatches Envelopes to one or more Senders based on
// label selectors. Routes are evaluated in registration order; all
// matching routes receive the envelope (fan-out).
type EnvelopeRouter struct {
	mu     sync.RWMutex
	routes []Route
}

// NewEnvelopeRouter returns an empty EnvelopeRouter.
func NewEnvelopeRouter() *EnvelopeRouter {
	return &EnvelopeRouter{}
}

// AddRoute registers a route. Returns an error if the sender is nil.
func (r *EnvelopeRouter) AddRoute(selector map[string]string, s Sender) error {
	if s == nil {
		return fmt.Errorf("envelope_router: sender must not be nil")
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	r.routes = append(r.routes, Route{Selector: selector, Sender: s})
	return nil
}

// Dispatch sends the envelope to all matching routes. Errors from
// individual senders are collected and returned as a combined error.
func (r *EnvelopeRouter) Dispatch(ctx context.Context, env *Envelope) error {
	if env == nil {
		return fmt.Errorf("envelope_router: envelope must not be nil")
	}
	r.mu.RLock()
	routes := make([]Route, len(r.routes))
	copy(routes, r.routes)
	r.mu.RUnlock()

	var errs []error
	for _, route := range routes {
		if matchesSelector(route.Selector, env.Labels) {
			if err := route.Sender.Send(ctx, env.Drifts); err != nil {
				errs = append(errs, fmt.Errorf("route to %T: %w", route.Sender, err))
			}
		}
	}
	if len(errs) > 0 {
		return fmt.Errorf("envelope_router: %d sender(s) failed: %v", len(errs), errs)
	}
	return nil
}

// matchesSelector returns true when every key/value pair in selector
// exists with the same value in labels.
func matchesSelector(selector, labels map[string]string) bool {
	for k, v := range selector {
		if labels[k] != v {
			return false
		}
	}
	return true
}
