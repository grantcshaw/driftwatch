package notify

import (
	"errors"
	"testing"
	"time"

	"github.com/your-org/driftwatch/internal/drift"
)

func makeBackoffDrifts() []drift.Drift {
	return []drift.Drift{
		{Key: "PORT", BaselineValue: "8080", CurrentValue: "9090"},
	}
}

func TestNewBackoffSender_NilInner(t *testing.T) {
	_, err := NewBackoffSender(nil, 3, 10*time.Millisecond, 100*time.Millisecond)
	if err == nil {
		t.Fatal("expected error for nil inner")
	}
}

func TestNewBackoffSender_ZeroAttempts(t *testing.T) {
	inner := &mockSender{}
	_, err := NewBackoffSender(inner, 0, 10*time.Millisecond, 100*time.Millisecond)
	if err == nil {
		t.Fatal("expected error for zero attempts")
	}
}

func TestNewBackoffSender_InvalidDelays(t *testing.T) {
	inner := &mockSender{}
	_, err := NewBackoffSender(inner, 2, 0, 100*time.Millisecond)
	if err == nil {
		t.Fatal("expected error for non-positive baseDelay")
	}
	_, err = NewBackoffSender(inner, 2, 50*time.Millisecond, 10*time.Millisecond)
	if err == nil {
		t.Fatal("expected error when maxDelay < baseDelay")
	}
}

func TestBackoffSender_NoDrifts_Noop(t *testing.T) {
	inner := &mockSender{}
	b, _ := NewBackoffSender(inner, 3, 10*time.Millisecond, 100*time.Millisecond)
	noopSleep(b)
	if err := b.Send("prod", nil); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if inner.calls != 0 {
		t.Fatalf("expected 0 inner calls, got %d", inner.calls)
	}
}

func TestBackoffSender_SuccessOnFirstAttempt(t *testing.T) {
	inner := &mockSender{}
	b, _ := NewBackoffSender(inner, 3, 10*time.Millisecond, 100*time.Millisecond)
	noopSleep(b)
	if err := b.Send("prod", makeBackoffDrifts()); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if inner.calls != 1 {
		t.Fatalf("expected 1 call, got %d", inner.calls)
	}
}

func TestBackoffSender_RetriesOnFailure(t *testing.T) {
	inner := &mockSender{failTimes: 2, err: errors.New("transient")}
	b, _ := NewBackoffSender(inner, 3, 10*time.Millisecond, 100*time.Millisecond)
	sleptCount := 0
	b.sleep = func(d time.Duration) { sleptCount++ }
	if err := b.Send("prod", makeBackoffDrifts()); err != nil {
		t.Fatalf("unexpected error after eventual success: %v", err)
	}
	if inner.calls != 3 {
		t.Fatalf("expected 3 calls, got %d", inner.calls)
	}
	if sleptCount != 2 {
		t.Fatalf("expected 2 sleeps, got %d", sleptCount)
	}
}

func TestBackoffSender_AllAttemptsExhausted(t *testing.T) {
	sentinel := errors.New("permanent failure")
	inner := &mockSender{failTimes: 99, err: sentinel}
	b, _ := NewBackoffSender(inner, 3, 10*time.Millisecond, 100*time.Millisecond)
	noopSleep(b)
	err := b.Send("prod", makeBackoffDrifts())
	if !errors.Is(err, sentinel) {
		t.Fatalf("expected sentinel error, got %v", err)
	}
	if inner.calls != 3 {
		t.Fatalf("expected 3 calls, got %d", inner.calls)
	}
}

func TestBackoffSender_DelayCapAtMax(t *testing.T) {
	inner := &mockSender{failTimes: 99, err: errors.New("fail")}
	b, _ := NewBackoffSender(inner, 5, 10*time.Millisecond, 15*time.Millisecond)
	maxObserved := time.Duration(0)
	b.sleep = func(d time.Duration) {
		if d > maxObserved {
			maxObserved = d
		}
	}
	b.Send("prod", makeBackoffDrifts()) //nolint
	if maxObserved > 15*time.Millisecond {
		t.Fatalf("delay exceeded maxDelay: %v", maxObserved)
	}
}

// noopSleep replaces the sleep function with a no-op for fast tests.
func noopSleep(b *BackoffSender) {
	b.sleep = func(time.Duration) {}
}
