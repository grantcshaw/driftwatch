package notify

import (
	"errors"
	"testing"

	"github.com/yourorg/driftwatch/internal/drift"
)

type stubSender struct {
	called bool
	err    error
	env    string
	drifts []drift.Drift
}

func (s *stubSender) Send(env string, drifts []drift.Drift) error {
	s.called = true
	s.env = env
	s.drifts = drifts
	return s.err
}

func TestMultiSender_Len(t *testing.T) {
	m := NewMultiSender(&stubSender{}, &stubSender{})
	if m.Len() != 2 {
		t.Fatalf("expected 2 senders, got %d", m.Len())
	}
}

func TestMultiSender_Add(t *testing.T) {
	m := NewMultiSender()
	m.Add(&stubSender{})
	if m.Len() != 1 {
		t.Fatalf("expected 1 sender after Add")
	}
}

func TestMultiSender_Send_AllCalled(t *testing.T) {
	a, b := &stubSender{}, &stubSender{}
	m := NewMultiSender(a, b)
	drifts := []drift.Drift{{Key: "K", BaselineValue: "v1", TargetValue: "v2"}}
	if err := m.Send("prod", drifts); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !a.called || !b.called {
		t.Error("expected both senders to be called")
	}
	if a.env != "prod" || b.env != "prod" {
		t.Error("expected env to be forwarded correctly")
	}
}

func TestMultiSender_Send_CollectsErrors(t *testing.T) {
	a := &stubSender{err: errors.New("smtp down")}
	b := &stubSender{err: errors.New("webhook timeout")}
	m := NewMultiSender(a, b)
	err := m.Send("staging", []drift.Drift{{Key: "X"}})
	if err == nil {
		t.Fatal("expected combined error")
	}
	if !errors.Is(err, a.err) && !errors.Is(err, b.err) {
		t.Errorf("expected wrapped errors, got: %v", err)
	}
}

func TestMultiSender_Send_ContinuesAfterError(t *testing.T) {
	a := &stubSender{err: errors.New("fail")}
	b := &stubSender{}
	m := NewMultiSender(a, b)
	_ = m.Send("staging", []drift.Drift{{Key: "Y"}})
	if !b.called {
		t.Error("expected second sender to be called even after first failed")
	}
}

func TestMultiSender_Send_NoDrifts_AllCalled(t *testing.T) {
	a, b := &stubSender{}, &stubSender{}
	m := NewMultiSender(a, b)
	_ = m.Send("prod", nil)
	if !a.called || !b.called {
		t.Error("expected senders called even with no drifts")
	}
}
