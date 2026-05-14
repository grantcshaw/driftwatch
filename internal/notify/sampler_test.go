package notify

import (
	"errors"
	"testing"

	"github.com/driftwatch/internal/drift"
)

func makeSamplerDrifts() []drift.Drift {
	return []drift.Drift{
		{Key: "DB_HOST", BaselineValue: "prod-db", TargetValue: "staging-db"},
	}
}

func TestNewSampler_NilInner(t *testing.T) {
	_, err := NewSampler(nil, 0.5)
	if err == nil {
		t.Fatal("expected error for nil inner sender")
	}
}

func TestNewSampler_InvalidRate_Zero(t *testing.T) {
	mock := &mockSender{}
	_, err := NewSampler(mock, 0.0)
	if err == nil {
		t.Fatal("expected error for rate=0.0")
	}
}

func TestNewSampler_InvalidRate_Negative(t *testing.T) {
	mock := &mockSender{}
	_, err := NewSampler(mock, -0.1)
	if err == nil {
		t.Fatal("expected error for negative rate")
	}
}

func TestNewSampler_InvalidRate_AboveOne(t *testing.T) {
	mock := &mockSender{}
	_, err := NewSampler(mock, 1.1)
	if err == nil {
		t.Fatal("expected error for rate > 1.0")
	}
}

func TestSampler_NoDrifts_Noop(t *testing.T) {
	mock := &mockSender{}
	s, _ := NewSampler(mock, 1.0)
	if err := s.Send("env", nil); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if mock.calls != 0 {
		t.Fatalf("expected 0 calls, got %d", mock.calls)
	}
}

func TestSampler_AlwaysSend_RateOne(t *testing.T) {
	mock := &mockSender{}
	s, _ := NewSampler(mock, 1.0)
	// rand always returns < 1.0, so rate=1.0 always passes
	for i := 0; i < 5; i++ {
		if err := s.Send("env", makeSamplerDrifts()); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	}
	if mock.calls != 5 {
		t.Fatalf("expected 5 calls, got %d", mock.calls)
	}
}

func TestSampler_NeverSend_ControlledRng(t *testing.T) {
	mock := &mockSender{}
	s, _ := NewSampler(mock, 0.5)
	// override randfn to always return 0.9 (>= 0.5), so no sends
	s.randfn = func() float64 { return 0.9 }
	if err := s.Send("env", makeSamplerDrifts()); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if mock.calls != 0 {
		t.Fatalf("expected 0 calls, got %d", mock.calls)
	}
}

func TestSampler_ForwardsSendError(t *testing.T) {
	mock := &mockSender{err: errors.New("downstream failure")}
	s, _ := NewSampler(mock, 1.0)
	err := s.Send("env", makeSamplerDrifts())
	if err == nil {
		t.Fatal("expected error to be forwarded")
	}
}
