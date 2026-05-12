package notify

import (
	"github.com/your-org/driftwatch/internal/drift"
)

// mockSender is a test double for the Sender interface.
// It records calls and optionally returns an error for the first failTimes calls.
type mockSender struct {
	calls     int
	failTimes int
	err       error
	lastEnv   string
	lastDrift []drift.Drift
}

func (m *mockSender) Send(env string, drifts []drift.Drift) error {
	m.calls++
	m.lastEnv = env
	m.lastDrift = drifts
	if m.failTimes > 0 && m.calls <= m.failTimes {
		return m.err
	}
	return nil
}
