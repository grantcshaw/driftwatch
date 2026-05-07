package notify

import (
	"errors"
	"sync/atomic"
	"testing"
	"time"
)

// stubSender counts how many times Send is called.
type stubSender struct {
	calls atomic.Int32
	err   error
}

func (s *stubSender) Send(_ string, _ []DriftEntry) error {
	s.calls.Add(1)
	return s.err
}

func makeThrottleDrifts() []DriftEntry {
	return []DriftEntry{{Key: "DB_HOST", BaselineValue: "prod-db", CurrentValue: "staging-db", Severity: "warning"}}
}

func TestNewThrottle_NilInner(t *testing.T) {
	_, err := NewThrottle(nil, ThrottleConfig{MinInterval: time.Second})
	if err == nil {
		t.Fatal("expected error for nil inner sender")
	}
}

func TestNewThrottle_ZeroInterval(t *testing.T) {
	_, err := NewThrottle(&stubSender{}, ThrottleConfig{MinInterval: 0})
	if err == nil {
		t.Fatal("expected error for zero interval")
	}
}

func TestThrottle_FirstSend_Passes(t *testing.T) {
	stub := &stubSender{}
	th, _ := NewThrottle(stub, ThrottleConfig{MinInterval: time.Minute})

	if err := th.Send("prod", makeThrottleDrifts()); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if stub.calls.Load() != 1 {
		t.Fatalf("expected 1 call, got %d", stub.calls.Load())
	}
}

func TestThrottle_SecondSend_Suppressed(t *testing.T) {
	stub := &stubSender{}
	th, _ := NewThrottle(stub, ThrottleConfig{MinInterval: time.Minute})

	th.Send("prod", makeThrottleDrifts()) //nolint:errcheck
	th.Send("prod", makeThrottleDrifts()) //nolint:errcheck

	if stub.calls.Load() != 1 {
		t.Fatalf("expected 1 call after throttle, got %d", stub.calls.Load())
	}
}

func TestThrottle_DifferentEnvs_IndependentLimits(t *testing.T) {
	stub := &stubSender{}
	th, _ := NewThrottle(stub, ThrottleConfig{MinInterval: time.Minute})

	th.Send("prod", makeThrottleDrifts())    //nolint:errcheck
	th.Send("staging", makeThrottleDrifts()) //nolint:errcheck

	if stub.calls.Load() != 2 {
		t.Fatalf("expected 2 calls for different envs, got %d", stub.calls.Load())
	}
}

func TestThrottle_Reset_AllowsResend(t *testing.T) {
	stub := &stubSender{}
	th, _ := NewThrottle(stub, ThrottleConfig{MinInterval: time.Minute})

	th.Send("prod", makeThrottleDrifts()) //nolint:errcheck
	th.Reset("prod")
	th.Send("prod", makeThrottleDrifts()) //nolint:errcheck

	if stub.calls.Load() != 2 {
		t.Fatalf("expected 2 calls after reset, got %d", stub.calls.Load())
	}
}

func TestThrottle_NoDrifts_SkipsSend(t *testing.T) {
	stub := &stubSender{}
	th, _ := NewThrottle(stub, ThrottleConfig{MinInterval: time.Minute})

	th.Send("prod", nil) //nolint:errcheck

	if stub.calls.Load() != 0 {
		t.Fatalf("expected 0 calls for empty drifts, got %d", stub.calls.Load())
	}
}

func TestThrottle_PropagatesInnerError(t *testing.T) {
	want := errors.New("send failed")
	stub := &stubSender{err: want}
	th, _ := NewThrottle(stub, ThrottleConfig{MinInterval: time.Minute})

	got := th.Send("prod", makeThrottleDrifts())
	if !errors.Is(got, want) {
		t.Fatalf("expected %v, got %v", want, got)
	}
}
