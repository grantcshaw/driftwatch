package notify

import (
	"context"
	"errors"
	"testing"

	"github.com/example/driftwatch/internal/drift"
)

func makeTapDrifts(n int) []drift.Drift {
	out := make([]drift.Drift, n)
	for i := range out {
		out[i] = drift.Drift{Key: "key", Baseline: "old", Current: "new"}
	}
	return out
}

func TestNewTap_NilInner(t *testing.T) {
	_, err := NewTap(nil)
	if err == nil {
		t.Fatal("expected error for nil inner")
	}
}

func TestTap_RecordsCall(t *testing.T) {
	inner := &mockSender{}
	tap, _ := NewTap(inner)
	_ = tap.Send(context.Background(), "staging", makeTapDrifts(2))
	if tap.Len() != 1 {
		t.Fatalf("expected 1 record, got %d", tap.Len())
	}
	if tap.Records[0].Env != "staging" {
		t.Fatalf("expected env staging, got %s", tap.Records[0].Env)
	}
	if len(tap.Records[0].Drifts) != 2 {
		t.Fatalf("expected 2 drifts, got %d", len(tap.Records[0].Drifts))
	}
}

func TestTap_DelegatesError(t *testing.T) {
	inner := &mockSender{err: errors.New("send failed")}
	tap, _ := NewTap(inner)
	err := tap.Send(context.Background(), "prod", makeTapDrifts(1))
	if err == nil {
		t.Fatal("expected error from inner")
	}
	if tap.Len() != 1 {
		t.Fatal("record should still be captured on error")
	}
}

func TestTap_Reset_ClearsRecords(t *testing.T) {
	inner := &mockSender{}
	tap, _ := NewTap(inner)
	_ = tap.Send(context.Background(), "prod", makeTapDrifts(1))
	_ = tap.Send(context.Background(), "prod", makeTapDrifts(1))
	tap.Reset()
	if tap.Len() != 0 {
		t.Fatalf("expected 0 after reset, got %d", tap.Len())
	}
}

func TestTap_MultipleEnvs(t *testing.T) {
	inner := &mockSender{}
	tap, _ := NewTap(inner)
	_ = tap.Send(context.Background(), "prod", makeTapDrifts(1))
	_ = tap.Send(context.Background(), "staging", makeTapDrifts(3))
	if tap.Len() != 2 {
		t.Fatalf("expected 2 records, got %d", tap.Len())
	}
	if tap.Records[1].Env != "staging" {
		t.Fatalf("expected staging, got %s", tap.Records[1].Env)
	}
}

func TestTap_TimestampSet(t *testing.T) {
	inner := &mockSender{}
	tap, _ := NewTap(inner)
	_ = tap.Send(context.Background(), "prod", makeTapDrifts(1))
	if tap.Records[0].At.IsZero() {
		t.Fatal("expected non-zero timestamp")
	}
}
