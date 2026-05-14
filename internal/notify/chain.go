package notify

import (
	"context"
	"errors"
	"fmt"

	"github.com/example/driftwatch/internal/drift"
)

// Chain executes a sequence of Senders in order. If any sender fails and
// haltOnError is true, the chain stops immediately. Otherwise all senders
// are attempted and errors are collected.
type Chain struct {
	senders      []Sender
	haltOnError  bool
}

// NewChain creates a Chain wrapping the provided senders.
// haltOnError controls whether the first failure aborts the chain.
func NewChain(haltOnError bool, senders ...Sender) (*Chain, error) {
	for i, s := range senders {
		if s == nil {
			return nil, fmt.Errorf("chain: sender at index %d is nil", i)
		}
	}
	if len(senders) == 0 {
		return nil, errors.New("chain: at least one sender is required")
	}
	return &Chain{
		senders:     senders,
		haltOnError: haltOnError,
	}, nil
}

// Add appends a Sender to the chain.
func (c *Chain) Add(s Sender) error {
	if s == nil {
		return errors.New("chain: cannot add nil sender")
	}
	c.senders = append(c.senders, s)
	return nil
}

// Len returns the number of senders in the chain.
func (c *Chain) Len() int {
	return len(c.senders)
}

// Send calls each sender in order. On error, behaviour depends on haltOnError.
func (c *Chain) Send(ctx context.Context, env string, drifts []drift.Drift) error {
	var errs []error
	for _, s := range c.senders {
		if err := s.Send(ctx, env, drifts); err != nil {
			if c.haltOnError {
				return fmt.Errorf("chain: halted: %w", err)
			}
			errs = append(errs, err)
		}
	}
	if len(errs) > 0 {
		return fmt.Errorf("chain: %d sender(s) failed: %w", len(errs), errors.Join(errs...))
	}
	return nil
}
