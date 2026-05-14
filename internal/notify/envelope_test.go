package notify

import (
	"testing"

	"github.com/yourorg/driftwatch/internal/drift"
)

func makeEnvelopeDrifts(n int) []drift.Drift {
	out := make([]drift.Drift, n)
	for i := range out {
		out[i] = drift.Drift{Key: fmt.Sprintf("KEY_%d", i), BaselineValue: "a", CurrentValue: "b"}
	}
	return out
}

func TestNewEnvelope_EmptyEnv_Errors(t *testing.T) {
	_, err := NewEnvelope("", &mockSender{})
	if err == nil {
		t.Error("expected error for empty env")
	}
}

func TestNewEnvelope_NilInner_Errors(t *testing.T) {
	_, err := NewEnvelope("prod", nil)
	if err == nil {
		t.Error("expected error for nil inner sender")
	}
}

func TestNewEnvelope_NoDrifts_WarningDefault(t *testing.T) {
	recv := &mockSender{}
	e, _ := NewEnvelope("prod", recv)
	if err := e.Send("prod", nil); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if recv.called {
		t.Error("expected inner sender not to be called with no drifts")
	}
}

func TestNewEnvelope_SingleChange_Warning(t *testing.T) {
	recv := &mockSender{}
	e, _ := NewEnvelope("prod", recv)
	drifts := []drift.Drift{{Key: "X", BaselineValue: "a", CurrentValue: "b"}}
	_ = e.Send("prod", drifts)
	if !recv.called {
		t.Fatal("expected inner sender to be called")
	}
	for _, d := range recv.lastDrifts {
		if d.Severity != "warning" {
			t.Errorf("expected warning, got %q", d.Severity)
		}
	}
}

func TestNewEnvelope_MultipleChanges_Critical(t *testing.T) {
	recv := &mockSender{}
	e, _ := NewEnvelope("prod", recv)
	drifts := makeEnvelopeDrifts(3)
	_ = e.Send("prod", drifts)
	for _, d := range recv.lastDrifts {
		if d.Severity != "critical" {
			t.Errorf("expected critical, got %q", d.Severity)
		}
	}
}

func TestNewEnvelope_PreservesExistingSeverity(t *testing.T) {
	recv := &mockSender{}
	e, _ := NewEnvelope("prod", recv)
	drifts := []drift.Drift{{Key: "Y", BaselineValue: "1", CurrentValue: "2", Severity: "critical"}}
	_ = e.Send("prod", drifts)
	for _, d := range recv.lastDrifts {
		if d.Severity != "critical" {
			t.Errorf("expected preserved critical, got %q", d.Severity)
		}
	}
}
