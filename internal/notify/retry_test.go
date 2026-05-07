package notify

import (
	"errors"
	"testing"
	"time"

	"github.com/org/driftwatch/internal/drift"
)

type countingSender struct {
	calls   int
	failFor int // fail the first N calls
}

func (c *countingSender) Send(_ string, _ []drift.Drift) error {
	c.calls++
	if c.calls <= c.failFor {
		return errors.New("transient error")
	}
	return nil
}

func makeRetryDrifts() []drift.Drift {
	return []drift.Drift{{Key: "PORT", BaselineValue: "8080", TargetValue: "9090"}}
}

func TestNewRetry_NilInner(t *testing.T) {
	_, err := NewRetry(nil, 3, 0)
	if err == nil {
		t.Fatal("expected error for nil inner sender")
	}
}

func TestNewRetry_ZeroAttempts(t *testing.T) {
	s := &countingSender{}
	_, err := NewRetry(s, 0, 0)
	if err == nil {
		t.Fatal("expected error for maxAttempts=0")
	}
}

func TestRetry_SuccessOnFirstAttempt(t *testing.T) {
	s := &countingSender{failFor: 0}
	r, _ := NewRetry(s, 3, 0)
	if err := r.Send("prod", makeRetryDrifts()); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if s.calls != 1 {
		t.Fatalf("expected 1 call, got %d", s.calls)
	}
}

func TestRetry_SuccessAfterRetries(t *testing.T) {
	s := &countingSender{failFor: 2}
	r, _ := NewRetry(s, 3, 0)
	if err := r.Send("prod", makeRetryDrifts()); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if s.calls != 3 {
		t.Fatalf("expected 3 calls, got %d", s.calls)
	}
}

func TestRetry_AllAttemptsFail(t *testing.T) {
	s := &countingSender{failFor: 10}
	r, _ := NewRetry(s, 3, 0)
	err := r.Send("prod", makeRetryDrifts())
	if err == nil {
		t.Fatal("expected error after all attempts fail")
	}
	if s.calls != 3 {
		t.Fatalf("expected 3 calls, got %d", s.calls)
	}
}

func TestRetry_RespectsDelay(t *testing.T) {
	s := &countingSender{failFor: 1}
	delay := 20 * time.Millisecond
	r, _ := NewRetry(s, 2, delay)
	start := time.Now()
	_ = r.Send("prod", makeRetryDrifts())
	elapsed := time.Since(start)
	if elapsed < delay {
		t.Fatalf("expected delay >= %v, got %v", delay, elapsed)
	}
}

func TestRetry_NoDrifts_DelegatesToInner(t *testing.T) {
	s := &countingSender{}
	r, _ := NewRetry(s, 2, 0)
	if err := r.Send("staging", nil); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if s.calls != 1 {
		t.Fatalf("expected 1 call, got %d", s.calls)
	}
}
