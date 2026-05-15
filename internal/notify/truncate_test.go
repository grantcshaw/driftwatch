package notify

import (
	"testing"

	"github.com/driftwatch/internal/drift"
)

func makeTruncateDrifts(n int) []drift.Drift {
	out := make([]drift.Drift, n)
	for i := 0; i < n; i++ {
		out[i] = drift.Drift{
			Key:      fmt.Sprintf("key%d", i),
			Baseline: "old",
			Current:  "new",
		}
	}
	return out
}

func TestNewTruncate_NilInner(t *testing.T) {
	_, err := NewTruncate(nil, 5)
	if err == nil {
		t.Fatal("expected error for nil inner sender")
	}
}

func TestNewTruncate_ZeroMaxItems(t *testing.T) {
	mock := &mockSender{}
	_, err := NewTruncate(mock, 0)
	if err == nil {
		t.Fatal("expected error for maxItems=0")
	}
}

func TestNewTruncate_NegativeMaxItems(t *testing.T) {
	mock := &mockSender{}
	_, err := NewTruncate(mock, -3)
	if err == nil {
		t.Fatal("expected error for negative maxItems")
	}
}

func TestTruncate_NoDrifts_Noop(t *testing.T) {
	mock := &mockSender{}
	tr, _ := NewTruncate(mock, 5)
	if err := tr.Send("env", nil); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if mock.called {
		t.Fatal("inner sender should not be called with no drifts")
	}
}

func TestTruncate_BelowLimit_PassesAll(t *testing.T) {
	mock := &mockSender{}
	tr, _ := NewTruncate(mock, 10)
	drifts := makeTruncateDrifts(5)
	if err := tr.Send("env", drifts); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(mock.lastDrifts) != 5 {
		t.Fatalf("expected 5 drifts, got %d", len(mock.lastDrifts))
	}
}

func TestTruncate_AtLimit_PassesAll(t *testing.T) {
	mock := &mockSender{}
	tr, _ := NewTruncate(mock, 5)
	drifts := makeTruncateDrifts(5)
	if err := tr.Send("env", drifts); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(mock.lastDrifts) != 5 {
		t.Fatalf("expected 5 drifts, got %d", len(mock.lastDrifts))
	}
}

func TestTruncate_ExceedsLimit_TruncatesAndAppendsSummary(t *testing.T) {
	mock := &mockSender{}
	tr, _ := NewTruncate(mock, 3)
	drifts := makeTruncateDrifts(7)
	if err := tr.Send("env", drifts); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// 3 real + 1 summary
	if len(mock.lastDrifts) != 4 {
		t.Fatalf("expected 4 drifts (3+summary), got %d", len(mock.lastDrifts))
	}
	last := mock.lastDrifts[3]
	if last.Key != "__truncated__" {
		t.Errorf("expected summary key '__truncated__', got %q", last.Key)
	}
	if last.Current == "" {
		t.Error("expected summary message in Current field")
	}
}

func TestTruncate_DelegatesError(t *testing.T) {
	mock := &mockSender{err: fmt.Errorf("downstream failure")}
	tr, _ := NewTruncate(mock, 5)
	if err := tr.Send("env", makeTruncateDrifts(2)); err == nil {
		t.Fatal("expected error from inner sender")
	}
}
