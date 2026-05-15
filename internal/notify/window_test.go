package notify

import (
	"errors"
	"testing"
	"time"

	"github.com/yourorg/driftwatch/internal/drift"
)

func makeWindowDrifts(keys ...string) []drift.Drift {
	out := make([]drift.Drift, len(keys))
	for i, k := range keys {
		out[i] = drift.Drift{Key: k, BaselineValue: "a", CurrentValue: "b"}
	}
	return out
}

func TestNewWindow_NilInner(t *testing.T) {
	_, err := NewWindow(nil, time.Second, 2)
	if err == nil {
		t.Fatal("expected error for nil inner")
	}
}

func TestNewWindow_ZeroDuration(t *testing.T) {
	mock := &mockSender{}
	_, err := NewWindow(mock, 0, 2)
	if err == nil {
		t.Fatal("expected error for zero duration")
	}
}

func TestNewWindow_ZeroMinCount(t *testing.T) {
	mock := &mockSender{}
	_, err := NewWindow(mock, time.Second, 0)
	if err == nil {
		t.Fatal("expected error for zero minCount")
	}
}

func TestWindow_NoDrifts_Noop(t *testing.T) {
	mock := &mockSender{}
	w, _ := NewWindow(mock, time.Second, 2)
	if err := w.Send("prod", nil); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if mock.calls != 0 {
		t.Fatalf("expected 0 calls, got %d", mock.calls)
	}
}

func TestWindow_BelowThreshold_DoesNotForward(t *testing.T) {
	mock := &mockSender{}
	w, _ := NewWindow(mock, time.Second, 3)
	_ = w.Send("prod", makeWindowDrifts("k1", "k2"))
	if mock.calls != 0 {
		t.Fatalf("expected 0 calls, got %d", mock.calls)
	}
}

func TestWindow_AtThreshold_Forwards(t *testing.T) {
	mock := &mockSender{}
	w, _ := NewWindow(mock, time.Second, 2)
	_ = w.Send("prod", makeWindowDrifts("k1"))
	_ = w.Send("prod", makeWindowDrifts("k2"))
	if mock.calls != 1 {
		t.Fatalf("expected 1 call, got %d", mock.calls)
	}
	if len(mock.lastDrifts) != 2 {
		t.Fatalf("expected 2 drifts forwarded, got %d", len(mock.lastDrifts))
	}
}

func TestWindow_ResetsAfterForward(t *testing.T) {
	mock := &mockSender{}
	w, _ := NewWindow(mock, time.Second, 2)
	_ = w.Send("prod", makeWindowDrifts("k1", "k2"))
	_ = w.Send("prod", makeWindowDrifts("k3", "k4"))
	if mock.calls != 2 {
		t.Fatalf("expected 2 calls, got %d", mock.calls)
	}
}

func TestWindow_DeduplicatesKeys(t *testing.T) {
	mock := &mockSender{}
	w, _ := NewWindow(mock, time.Second, 2)
	_ = w.Send("prod", makeWindowDrifts("k1"))
	_ = w.Send("prod", makeWindowDrifts("k1", "k2"))
	if mock.calls != 1 {
		t.Fatalf("expected 1 call, got %d", mock.calls)
	}
	if len(mock.lastDrifts) != 2 {
		t.Fatalf("expected 2 unique drifts, got %d", len(mock.lastDrifts))
	}
}

func TestWindow_PropagatesInnerError(t *testing.T) {
	mock := &mockSender{err: errors.New("send failed")}
	w, _ := NewWindow(mock, time.Second, 1)
	err := w.Send("prod", makeWindowDrifts("k1"))
	if err == nil {
		t.Fatal("expected error from inner sender")
	}
}
