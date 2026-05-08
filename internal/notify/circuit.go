package notify

import (
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/yourorg/driftwatch/internal/drift"
)

// CircuitState represents the current state of the circuit breaker.
type CircuitState int

const (
	CircuitClosed CircuitState = iota // normal operation
	CircuitOpen                       // blocking sends
	CircuitHalfOpen                   // probing for recovery
)

// CircuitBreaker wraps a Sender and stops forwarding when the downstream
// sender fails repeatedly, protecting it from overload.
type CircuitBreaker struct {
	mu           sync.Mutex
	inner        Sender
	maxFailures  int
	resetAfter   time.Duration
	failures     int
	state        CircuitState
	openedAt     time.Time
}

// NewCircuitBreaker creates a CircuitBreaker that opens after maxFailures
// consecutive errors and attempts recovery after resetAfter duration.
func NewCircuitBreaker(inner Sender, maxFailures int, resetAfter time.Duration) (*CircuitBreaker, error) {
	if inner == nil {
		return nil, errors.New("circuit breaker: inner sender must not be nil")
	}
	if maxFailures <= 0 {
		return nil, errors.New("circuit breaker: maxFailures must be greater than zero")
	}
	if resetAfter <= 0 {
		return nil, errors.New("circuit breaker: resetAfter must be greater than zero")
	}
	return &CircuitBreaker{
		inner:       inner,
		maxFailures: maxFailures,
		resetAfter:  resetAfter,
		state:       CircuitClosed,
	}, nil
}

// Send forwards drifts to the inner sender unless the circuit is open.
func (c *CircuitBreaker) Send(env string, drifts []drift.Drift) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	switch c.state {
	case CircuitOpen:
		if time.Since(c.openedAt) >= c.resetAfter {
			c.state = CircuitHalfOpen
		} else {
			return fmt.Errorf("circuit breaker open for env %q: too many consecutive failures", env)
		}
	case CircuitClosed, CircuitHalfOpen:
		// proceed
	}

	err := c.inner.Send(env, drifts)
	if err != nil {
		c.failures++
		if c.failures >= c.maxFailures {
			c.state = CircuitOpen
			c.openedAt = time.Now()
		}
		return err
	}

	// success — reset
	c.failures = 0
	c.state = CircuitClosed
	return nil
}

// State returns the current circuit state (safe for external inspection).
func (c *CircuitBreaker) State() CircuitState {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.state
}
