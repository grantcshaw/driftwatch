package notify

import (
	"errors"
	"testing"

	"github.com/example/driftwatch/internal/drift"
)

func makeHeaderDrifts(keys ...string) []drift.Drift {
	out := make([]drift.Drift, len(keys))
	for i, k := range keys {
		out[i] = drift.Drift{
			Key:      k,
			OldValue: "a",
			NewValue: "b",
			Labels:   map[string]string{"existing": "yes"},
		}
	}
	return out
}

func TestNewHeaderSender_NilInner(t *testing.T) {
	_, err := NewHeaderSender(nil, map[string]string{"k": "v"})
	if err == nil {
		t.Fatal("expected error for nil inner")
	}
}

func TestNewHeaderSender_EmptyHeaders(t *testing.T) {
	mock := &mockSender{}
	_, err := NewHeaderSender(mock, map[string]string{})
	if err == nil {
		t.Fatal("expected error for empty headers")
	}
}

func TestHeaderSender_NoDrifts_Noop(t *testing.T) {
	mock := &mockSender{}
	s, err := NewHeaderSender(mock, map[string]string{"region": "us-east-1"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if err := s.Send("prod", nil); err != nil {
		t.Fatalf("expected nil error on empty drifts, got %v", err)
	}
	if mock.called {
		t.Fatal("inner sender should not be called for empty drifts")
	}
}

func TestHeaderSender_InjectsHeaders(t *testing.T) {
	mock := &mockSender{}
	headers := map[string]string{"region": "us-east-1", "cluster": "main"}
	s, err := NewHeaderSender(mock, headers)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	drifts := makeHeaderDrifts("cpu_limit")
	if err := s.Send("staging", drifts); err != nil {
		t.Fatalf("unexpected send error: %v", err)
	}
	if !mock.called {
		t.Fatal("expected inner sender to be called")
	}
	got := mock.lastDrifts[0].Labels
	if got["region"] != "us-east-1" {
		t.Errorf("expected region header, got %v", got)
	}
	if got["cluster"] != "main" {
		t.Errorf("expected cluster header, got %v", got)
	}
	// original label preserved
	if got["existing"] != "yes" {
		t.Errorf("expected existing label to be preserved, got %v", got)
	}
}

func TestHeaderSender_DoesNotMutateOriginal(t *testing.T) {
	mock := &mockSender{}
	s, _ := NewHeaderSender(mock, map[string]string{"env": "prod"})
	drifts := makeHeaderDrifts("mem_limit")
	orig := drifts[0].Labels
	_ = s.Send("prod", drifts)
	if _, ok := orig["env"]; ok {
		t.Error("original drift labels should not be mutated")
	}
}

func TestHeaderSender_PropagatesError(t *testing.T) {
	mock := &mockSender{err: errors.New("downstream failure")}
	s, _ := NewHeaderSender(mock, map[string]string{"x": "y"})
	err := s.Send("prod", makeHeaderDrifts("k"))
	if err == nil {
		t.Fatal("expected error to be propagated")
	}
}
