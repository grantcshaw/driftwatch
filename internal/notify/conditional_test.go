package notify

import (
	"errors"
	"testing"

	"github.com/your-org/driftwatch/internal/drift"
)

func makeConditionalDrifts(keys ...string) []drift.Drift {
	out := make([]drift.Drift, len(keys))
	for i, k := range keys {
		out[i] = drift.Drift{Key: k, BaselineValue: "x", CurrentValue: "y"}
	}
	return out
}

func TestNewConditional_NilInner(t *testing.T) {
	_, err := NewConditional(nil, func(_ string, _ []drift.Drift) bool { return true })
	if err == nil {
		t.Fatal("expected error for nil inner")
	}
}

func TestNewConditional_NilCond(t *testing.T) {
	_, err := NewConditional(&mockSender{}, nil)
	if err == nil {
		t.Fatal("expected error for nil condition")
	}
}

func TestConditional_NoDrifts_Noop(t *testing.T) {
	inner := &mockSender{}
	c, _ := NewConditional(inner, func(_ string, _ []drift.Drift) bool { return true })
	if err := c.Send("prod", nil); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if inner.calls != 0 {
		t.Fatalf("expected 0 calls, got %d", inner.calls)
	}
}

func TestConditional_ConditionFalse_Skips(t *testing.T) {
	inner := &mockSender{}
	c, _ := NewConditional(inner, func(_ string, _ []drift.Drift) bool { return false })
	if err := c.Send("prod", makeConditionalDrifts("k1")); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if inner.calls != 0 {
		t.Fatalf("expected 0 calls, got %d", inner.calls)
	}
}

func TestConditional_ConditionTrue_Forwards(t *testing.T) {
	inner := &mockSender{}
	c, _ := NewConditional(inner, func(_ string, _ []drift.Drift) bool { return true })
	if err := c.Send("prod", makeConditionalDrifts("k1")); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if inner.calls != 1 {
		t.Fatalf("expected 1 call, got %d", inner.calls)
	}
}

func TestConditional_PropagatesInnerError(t *testing.T) {
	inner := &mockSender{err: errors.New("boom")}
	c, _ := NewConditional(inner, func(_ string, _ []drift.Drift) bool { return true })
	if err := c.Send("prod", makeConditionalDrifts("k1")); err == nil {
		t.Fatal("expected error from inner sender")
	}
}

func TestMinDriftCount(t *testing.T) {
	cond := MinDriftCount(3)
	if cond("prod", makeConditionalDrifts("a", "b")) {
		t.Fatal("expected false for 2 drifts with min=3")
	}
	if !cond("prod", makeConditionalDrifts("a", "b", "c")) {
		t.Fatal("expected true for 3 drifts with min=3")
	}
}

func TestEnvMatches(t *testing.T) {
	cond := EnvMatches("prod", "staging")
	if !cond("prod", nil) {
		t.Fatal("expected match for prod")
	}
	if cond("dev", nil) {
		t.Fatal("expected no match for dev")
	}
}
