package notify

import (
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/org/driftwatch/internal/drift"
)

func makeBatchDrifts(n int) []drift.Drift {
	out := make([]drift.Drift, n)
	for i := range out {
		out[i] = drift.Drift{Key: fmt.Sprintf("key%d", i), BaselineValue: "a", CurrentValue: "b"}
	}
	return out
}

func TestNewBatch_NilInner(t *testing.T) {
	_, err := NewBatch(nil, time.Second, 10)
	if err == nil {
		t.Fatal("expected error for nil inner sender")
	}
}

func TestNewBatch_ZeroWindow(t *testing.T) {
	m := &mockSender{}
	_, err := NewBatch(m, 0, 10)
	if err == nil {
		t.Fatal("expected error for zero window")
	}
}

func TestNewBatch_ZeroMaxSize(t *testing.T) {
	m := &mockSender{}
	_, err := NewBatch(m, time.Second, 0)
	if err == nil {
		t.Fatal("expected error for zero maxSize")
	}
}

func TestBatch_NoDrifts_Noop(t *testing.T) {
	m := &mockSender{}
	b, _ := NewBatch(m, time.Second, 5)
	if err := b.Send("env", nil); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if m.calls != 0 {
		t.Errorf("expected 0 inner calls, got %d", m.calls)
	}
}

func TestBatch_AccumulatesUntilMaxSize(t *testing.T) {
	m := &mockSender{}
	b, _ := NewBatch(m, time.Hour, 3)

	_ = b.Send("env", makeBatchDrifts(2))
	if m.calls != 0 {
		t.Errorf("expected no flush yet, got %d calls", m.calls)
	}
	if b.Len() != 2 {
		t.Errorf("expected 2 pending, got %d", b.Len())
	}

	_ = b.Send("env", makeBatchDrifts(1))
	if m.calls != 1 {
		t.Errorf("expected 1 flush, got %d calls", m.calls)
	}
	if b.Len() != 0 {
		t.Errorf("expected 0 pending after flush, got %d", b.Len())
	}
}

func TestBatch_FlushOnExpiredWindow(t *testing.T) {
	m := &mockSender{}
	b, _ := NewBatch(m, time.Millisecond, 100)
	_ = b.Send("env", makeBatchDrifts(1))
	time.Sleep(5 * time.Millisecond)
	_ = b.Send("env", makeBatchDrifts(1))
	if m.calls != 1 {
		t.Errorf("expected 1 flush after window expiry, got %d", m.calls)
	}
}

func TestBatch_Flush_EmptyNoop(t *testing.T) {
	m := &mockSender{}
	b, _ := NewBatch(m, time.Hour, 10)
	if err := b.Flush("env"); err != nil {
		t.Fatalf("unexpected error flushing empty batch: %v", err)
	}
	if m.calls != 0 {
		t.Errorf("expected 0 inner calls, got %d", m.calls)
	}
}

func TestBatch_Flush_ForwardsError(t *testing.T) {
	m := &mockSender{err: errors.New("send failed")}
	b, _ := NewBatch(m, time.Hour, 10)
	_ = b.Send("env", makeBatchDrifts(9))
	if err := b.Flush("env"); err == nil {
		t.Fatal("expected error from inner sender")
	}
}
