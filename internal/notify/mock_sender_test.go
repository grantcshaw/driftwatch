package notify

import "github.com/yourorg/driftwatch/internal/drift"

// mockSender is a test double for the Sender interface.
type mockSender struct {
	called     bool
	lastEnv    string
	lastDrifts []drift.Drift
	err        error
}

func (m *mockSender) Send(env string, drifts []drift.Drift) error {
	m.called = true
	m.lastEnv = env
	m.lastDrifts = append([]drift.Drift(nil), drifts...)
	return m.err
}
