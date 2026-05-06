package notify

import (
	"errors"
	"fmt"

	"github.com/yourorg/driftwatch/internal/drift"
)

// Sender is the common interface for all notification backends.
type Sender interface {
	Send(env string, drifts []drift.Drift) error
}

// MultiSender fans out drift notifications to multiple Sender implementations.
type MultiSender struct {
	senders []Sender
}

// NewMultiSender creates a MultiSender that dispatches to each provided Sender.
func NewMultiSender(senders ...Sender) *MultiSender {
	return &MultiSender{senders: senders}
}

// Send calls every registered Sender and collects all errors.
// All senders are attempted even if one fails.
func (m *MultiSender) Send(env string, drifts []drift.Drift) error {
	var errs []error
	for i, s := range m.senders {
		if err := s.Send(env, drifts); err != nil {
			errs = append(errs, fmt.Errorf("sender[%d]: %w", i, err))
		}
	}
	return errors.Join(errs...)
}

// Add appends a Sender to the MultiSender at runtime.
func (m *MultiSender) Add(s Sender) {
	m.senders = append(m.senders, s)
}

// Len returns the number of registered senders.
func (m *MultiSender) Len() int {
	return len(m.senders)
}
