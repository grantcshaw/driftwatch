package notify

import (
	"errors"
	"testing"
	"time"

	"github.com/yourorg/driftwatch/internal/drift"
)

type countingSender struct {
	calls int
	err   error
}

func (s *countingSender) Send(_ string, _ []drift.Drift) error {
	s.calls++
	return s.err
}

func makeCircuitDrifts() []drift.Drift {
	return []drift.Drift{{Key: "PORT", BaselineValue: "8080", TargetValue: "9090"}}
}

func TestNewCircuitBreaker_NilInner(t *testing.T) {
	_, err := NewCircuitBreaker(nil, 3, time.Second)
	if err == nil {
		t.Fatal("expected error for nil inner sender")
	}
}

func TestNewCircuitBreaker_InvalidMaxFailures(t *testing.T) {
	s := &countingSender{}
	_, err := NewCircuitBreaker(s, 0, time.Second)
	if err == nil {
		t.Fatal("expected error for zero maxFailures")
	}
}

func TestNewCircuitBreaker_InvalidResetAfter(t *testing.T) {
	s := &countingSender{}
	_, err := NewCircuitBreaker(s, 3, 0)
	if err == nil {
		t.Fatal("expected error for zero resetAfter")
	}
}

func TestCircuitBreaker_ClosedOnSuccess(t *testing.T) {
	inner := &countingSender{}
	cb, _ := NewCircuitBreaker(inner, 3, time.Second)

	if err := cb.Send("prod", makeCircuitDrifts()); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cb.State() != CircuitClosed {
		t.Errorf("expected closed, got %v", cb.State())
	}
}

func TestCircuitBreaker_OpensAfterMaxFailures(t *testing.T) {
	inner := &countingSender{err: errors.New("downstream down")}
	cb, _ := NewCircuitBreaker(inner, 3, time.Second)
	drifts := makeCircuitDrifts()

	for i := 0; i < 3; i++ {
		_ = cb.Send("prod", drifts)
	}
	if cb.State() != CircuitOpen {
		t.Errorf("expected open after 3 failures, got %v", cb.State())
	}
}

func TestCircuitBreaker_BlocksWhenOpen(t *testing.T) {
	inner := &countingSender{err: errors.New("fail")}
	cb, _ := NewCircuitBreaker(inner, 2, time.Hour)
	drifts := makeCircuitDrifts()

	_ = cb.Send("prod", drifts)
	_ = cb.Send("prod", drifts)

	callsBefore := inner.calls
	err := cb.Send("prod", drifts)
	if err == nil {
		t.Fatal("expected error from open circuit")
	}
	if inner.calls != callsBefore {
		t.Error("inner sender should not be called when circuit is open")
	}
}

func TestCircuitBreaker_HalfOpenAfterReset(t *testing.T) {
	inner := &countingSender{err: errors.New("fail")}
	cb, _ := NewCircuitBreaker(inner, 2, 10*time.Millisecond)
	drifts := makeCircuitDrifts()

	_ = cb.Send("prod", drifts)
	_ = cb.Send("prod", drifts)

	time.Sleep(20 * time.Millisecond)

	// should attempt (half-open), fail, re-open
	err := cb.Send("prod", drifts)
	if err == nil {
		t.Fatal("expected error on half-open probe failure")
	}
	if cb.State() != CircuitOpen {
		t.Errorf("expected re-opened circuit, got %v", cb.State())
	}
}

func TestCircuitBreaker_RecoverOnSuccess(t *testing.T) {
	inner := &countingSender{err: errors.New("fail")}
	cb, _ := NewCircuitBreaker(inner, 2, 10*time.Millisecond)
	drifts := makeCircuitDrifts()

	_ = cb.Send("prod", drifts)
	_ = cb.Send("prod", drifts)

	time.Sleep(20 * time.Millisecond)
	inner.err = nil // downstream recovered

	if err := cb.Send("prod", drifts); err != nil {
		t.Fatalf("unexpected error after recovery: %v", err)
	}
	if cb.State() != CircuitClosed {
		t.Errorf("expected closed after recovery, got %v", cb.State())
	}
}
