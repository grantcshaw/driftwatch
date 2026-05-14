package notify

import (
	"errors"
	"testing"

	"github.com/yourorg/driftwatch/internal/drift"
)

func makeRouterDrifts(severity string) []drift.Drift {
	return []drift.Drift{
		{Key: "DB_HOST", BaselineValue: "old", CurrentValue: "new", Severity: severity},
	}
}

func TestEnvelopeRouter_NoDrifts_Noop(t *testing.T) {
	recv := &mockSender{}
	r := NewEnvelopeRouter(nil)
	_ = r.AddRoute(Selector{}, recv)
	if err := r.Send("prod", nil); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if recv.called {
		t.Error("expected sender not to be called")
	}
}

func TestEnvelopeRouter_AddRoute_NilSender(t *testing.T) {
	r := NewEnvelopeRouter(nil)
	if err := r.AddRoute(Selector{}, nil); err == nil {
		t.Error("expected error for nil sender")
	}
}

func TestEnvelopeRouter_AddRoute_InvalidSeverity(t *testing.T) {
	r := NewEnvelopeRouter(nil)
	if err := r.AddRoute(Selector{MinSeverity: "extreme"}, &mockSender{}); err == nil {
		t.Error("expected error for invalid severity")
	}
}

func TestEnvelopeRouter_MatchesEnv(t *testing.T) {
	prodSender := &mockSender{}
	stagingSender := &mockSender{}
	r := NewEnvelopeRouter(nil)
	_ = r.AddRoute(Selector{Env: "prod"}, prodSender)
	_ = r.AddRoute(Selector{Env: "staging"}, stagingSender)

	_ = r.Send("prod", makeRouterDrifts("warning"))
	if !prodSender.called {
		t.Error("expected prod sender to be called")
	}
	if stagingSender.called {
		t.Error("expected staging sender not to be called")
	}
}

func TestEnvelopeRouter_FallbackUsed_WhenNoMatch(t *testing.T) {
	fallback := &mockSender{}
	r := NewEnvelopeRouter(fallback)
	_ = r.AddRoute(Selector{Env: "staging"}, &mockSender{})

	_ = r.Send("prod", makeRouterDrifts("warning"))
	if !fallback.called {
		t.Error("expected fallback sender to be called")
	}
}

func TestEnvelopeRouter_CriticalFilter(t *testing.T) {
	critSender := &mockSender{}
	r := NewEnvelopeRouter(nil)
	_ = r.AddRoute(Selector{MinSeverity: "critical"}, critSender)

	_ = r.Send("prod", makeRouterDrifts("warning"))
	if critSender.called {
		t.Error("expected critical sender not to be called for warning drift")
	}

	_ = r.Send("prod", makeRouterDrifts("critical"))
	if !critSender.called {
		t.Error("expected critical sender to be called for critical drift")
	}
}

func TestEnvelopeRouter_CollectsErrors(t *testing.T) {
	failing := &mockSender{err: errors.New("send failed")}
	r := NewEnvelopeRouter(nil)
	_ = r.AddRoute(Selector{}, failing)

	if err := r.Send("prod", makeRouterDrifts("warning")); err == nil {
		t.Error("expected error from failing sender")
	}
}
