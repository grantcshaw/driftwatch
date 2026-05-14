package notify

import (
	"errors"
	"testing"
	"time"

	"github.com/yourorg/driftwatch/internal/drift"
)

func makeDedupDrifts(keys ...string) []drift.Drift {
	out := make([]drift.Drift, 0, len(keys))
	for _, k := range keys {
		out = append(out, drift.Drift{
			Key:           k,
			BaselineValue: "old",
			CurrentValue:  "new",
		})
	}
	return out
}

func TestNewDedup_NilInner(t *testing.T) {
	_, err := NewDedup(nil, time.Minute)
	if err == nil {
		t.Fatal("expected error for nil inner sender")
	}
}

func TestNewDedup_ZeroWindow(t *testing.T) {
	_, err := NewDedup(&mockSender{}, 0)
	if err == nil {
		t.Fatal("expected error for zero window")
	}
}

func TestDedup_NoDrifts_Noop(t *testing.T) {
	inner := &mockSender{}
	d, _ := NewDedup(inner, time.Minute)
	if err := d.Send("prod", nil); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if inner.calls != 0 {
		t.Fatalf("expected 0 calls, got %d", inner.calls)
	}
}

func TestDedup_FirstSend_Passes(t *testing.T) {
	inner := &mockSender{}
	d, _ := NewDedup(inner, time.Minute)
	if err := d.Send("prod", makeDedupDrifts("KEY_A")); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if inner.calls != 1 {
		t.Fatalf("expected 1 call, got %d", inner.calls)
	}
}

func TestDedup_DuplicateWithinWindow_Suppressed(t *testing.T) {
	inner := &mockSender{}
	d, _ := NewDedup(inner, time.Hour)
	drifts := makeDedupDrifts("KEY_A")
	_ = d.Send("prod", drifts)
	_ = d.Send("prod", drifts)
	if inner.calls != 1 {
		t.Fatalf("expected 1 call (second suppressed), got %d", inner.calls)
	}
}

func TestDedup_DifferentEnv_NotSuppressed(t *testing.T) {
	inner := &mockSender{}
	d, _ := NewDedup(inner, time.Hour)
	drifts := makeDedupDrifts("KEY_A")
	_ = d.Send("prod", drifts)
	_ = d.Send("staging", drifts)
	if inner.calls != 2 {
		t.Fatalf("expected 2 calls for different envs, got %d", inner.calls)
	}
}

func TestDedup_DifferentDrifts_NotSuppressed(t *testing.T) {
	inner := &mockSender{}
	d, _ := NewDedup(inner, time.Hour)
	_ = d.Send("prod", makeDedupDrifts("KEY_A"))
	_ = d.Send("prod", makeDedupDrifts("KEY_B"))
	if inner.calls != 2 {
		t.Fatalf("expected 2 calls for different drift sets, got %d", inner.calls)
	}
}

func TestDedup_InnerError_Propagated(t *testing.T) {
	inner := &mockSender{err: errors.New("send failed")}
	d, _ := NewDedup(inner, time.Minute)
	err := d.Send("prod", makeDedupDrifts("KEY_A"))
	if err == nil {
		t.Fatal("expected error from inner sender")
	}
}
