package notify

import (
	"errors"
	"testing"

	"github.com/example/driftwatch/internal/drift"
)

func makeShadowDrifts() []drift.Drift {
	return []drift.Drift{
		{Key: "DB_HOST", BaselineValue: "prod-db", TargetValue: "staging-db"},
	}
}

func TestNewShadowSender_NilPrimary(t *testing.T) {
	_, err := NewShadowSender(nil, &mockSender{})
	if err == nil {
		t.Fatal("expected error for nil primary")
	}
}

func TestNewShadowSender_NilShadow(t *testing.T) {
	_, err := NewShadowSender(&mockSender{}, nil)
	if err == nil {
		t.Fatal("expected error for nil shadow")
	}
}

func TestShadowSender_NoDrifts_Noop(t *testing.T) {
	prim := &mockSender{}
	shad := &mockSender{}
	s, _ := NewShadowSender(prim, shad)

	if err := s.Send("env", nil); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if prim.calls != 0 || shad.calls != 0 {
		t.Fatal("expected no calls for empty drifts")
	}
}

func TestShadowSender_BothSucceed_NoMismatch(t *testing.T) {
	prim := &mockSender{}
	shad := &mockSender{}
	s, _ := NewShadowSender(prim, shad)

	if err := s.Send("env", makeShadowDrifts()); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	results := s.Results()
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	if results[0].Mismatch {
		t.Error("expected no mismatch when both succeed")
	}
}

func TestShadowSender_PrimaryFails_ReturnsPrimaryError(t *testing.T) {
	primErr := errors.New("primary failed")
	prim := &mockSender{err: primErr}
	shad := &mockSender{}
	s, _ := NewShadowSender(prim, shad)

	err := s.Send("env", makeShadowDrifts())
	if !errors.Is(err, primErr) {
		t.Fatalf("expected primary error, got %v", err)
	}
}

func TestShadowSender_ShadowFails_PrimarySucceeds_Mismatch(t *testing.T) {
	prim := &mockSender{}
	shad := &mockSender{err: errors.New("shadow failed")}
	s, _ := NewShadowSender(prim, shad)

	if err := s.Send("env", makeShadowDrifts()); err != nil {
		t.Fatalf("unexpected primary error: %v", err)
	}
	results := s.Results()
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	if !results[0].Mismatch {
		t.Error("expected mismatch when shadow fails but primary succeeds")
	}
}

func TestShadowSender_Reset_ClearsResults(t *testing.T) {
	prim := &mockSender{}
	shad := &mockSender{}
	s, _ := NewShadowSender(prim, shad)

	_ = s.Send("env", makeShadowDrifts())
	s.Reset()
	if len(s.Results()) != 0 {
		t.Fatal("expected empty results after Reset")
	}
}

func TestShadowSender_MultipleRounds_AccumulatesResults(t *testing.T) {
	prim := &mockSender{}
	shad := &mockSender{}
	s, _ := NewShadowSender(prim, shad)

	_ = s.Send("env", makeShadowDrifts())
	_ = s.Send("env", makeShadowDrifts())

	if len(s.Results()) != 2 {
		t.Fatalf("expected 2 results, got %d", len(s.Results()))
	}
}
