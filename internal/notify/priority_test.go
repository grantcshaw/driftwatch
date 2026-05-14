package notify

import (
	"testing"

	"github.com/org/driftwatch/internal/drift"
)

func makePriorityDrifts() []drift.Drift {
	return []drift.Drift{
		{Key: "alpha", Severity: "warning"},
		{Key: "beta", Severity: "critical"},
		{Key: "gamma", Severity: "info"},
		{Key: "delta", Severity: "critical"},
	}
}

func TestNewPriority_NilInner(t *testing.T) {
	_, err := NewPriority(nil)
	if err == nil {
		t.Fatal("expected error for nil inner sender")
	}
}

func TestPriority_NoDrifts_Noop(t *testing.T) {
	m := &mockSender{}
	p, _ := NewPriority(m)
	if err := p.Send("env", nil); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if m.calls != 0 {
		t.Errorf("expected 0 inner calls, got %d", m.calls)
	}
}

func TestPriority_CriticalFirst(t *testing.T) {
	var received []drift.Drift
	capture := senderFunc(func(_ string, d []drift.Drift) error {
		received = d
		return nil
	})
	p, _ := NewPriority(capture)
	_ = p.Send("env", makePriorityDrifts())

	if received[0].Severity != "critical" || received[1].Severity != "critical" {
		t.Errorf("expected first two drifts to be critical, got %v %v",
			received[0].Severity, received[1].Severity)
	}
	if received[2].Severity != "warning" {
		t.Errorf("expected third drift to be warning, got %v", received[2].Severity)
	}
}

func TestPriority_PreservesKeys(t *testing.T) {
	var received []drift.Drift
	capture := senderFunc(func(_ string, d []drift.Drift) error {
		received = d
		return nil
	})
	p, _ := NewPriority(capture)
	input := makePriorityDrifts()
	_ = p.Send("env", input)

	if len(received) != len(input) {
		t.Fatalf("expected %d drifts, got %d", len(input), len(received))
	}
}

func TestPriority_DoesNotMutateOriginal(t *testing.T) {
	m := &mockSender{}
	p, _ := NewPriority(m)
	input := makePriorityDrifts()
	origFirst := input[0].Key
	_ = p.Send("env", input)
	if input[0].Key != origFirst {
		t.Error("original slice was mutated by Priority.Send")
	}
}

// senderFunc is a convenience adapter for inline Sender implementations.
type senderFunc func(string, []drift.Drift) error

func (f senderFunc) Send(env string, drifts []drift.Drift) error {
	return f(env, drifts)
}
