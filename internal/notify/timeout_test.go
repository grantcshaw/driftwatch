package notify

import (
	"errors"
	"testing"
	"time"

	"github.com/your-org/driftwatch/internal/drift"
)

func makeTimeoutDrifts(keys ...string) []drift.Drift {
	out := make([]drift.Drift, len(keys))
	for i, k := range keys {
		out[i] = drift.Drift{Key: k, BaselineValue: "a", CurrentValue: "b"}
	}
	return out
}

func TestNewTimeout_NilInner(t *testing.T) {
	_, err := NewTimeout(nil, time.Second)
	if err == nil {
		t.Fatal("expected error for nil inner sender")
	}
}

func TestNewTimeout_ZeroDuration(t *testing.T) {
	_, err := NewTimeout(&mockSender{}, 0)
	if err == nil {
		t.Fatal("expected error for zero duration")
	}
}

func TestNewTimeout_NegativeDuration(t *testing.T) {
	_, err := NewTimeout(&mockSender{}, -time.Second)
	if err == nil {
		t.Fatal("expected error for negative duration")
	}
}

func TestTimeout_NoDrifts_Noop(t *testing.T) {
	inner := &mockSender{}
	to, _ := NewTimeout(inner, time.Second)
	if err := to.Send("prod", nil); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if inner.calls != 0 {
		t.Fatalf("expected 0 inner calls, got %d", inner.calls)
	}
}

func TestTimeout_SuccessWithinDeadline(t *testing.T) {
	inner := &mockSender{}
	to, _ := NewTimeout(inner, time.Second)
	if err := to.Send("prod", makeTimeoutDrifts("k1")); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if inner.calls != 1 {
		t.Fatalf("expected 1 inner call, got %d", inner.calls)
	}
}

func TestTimeout_PropagatesInnerError(t *testing.T) {
	inner := &mockSender{err: errors.New("downstream failure")}
	to, _ := NewTimeout(inner, time.Second)
	err := to.Send("prod", makeTimeoutDrifts("k1"))
	if err == nil {
		t.Fatal("expected error from inner sender")
	}
}

func TestTimeout_ExceedsDeadline(t *testing.T) {
	// slow sender that blocks longer than the timeout
	slow := &slowSender{delay: 200 * time.Millisecond}
	to, _ := NewTimeout(slow, 20*time.Millisecond)
	err := to.Send("prod", makeTimeoutDrifts("k1"))
	if err == nil {
		t.Fatal("expected timeout error")
	}
}

// slowSender blocks for delay before returning.
type slowSender struct {
	delay time.Duration
}

func (s *slowSender) Send(_ string, _ []drift.Drift) error {
	time.Sleep(s.delay)
	return nil
}
