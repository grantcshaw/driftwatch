package notify

import (
	"errors"
	"sync/atomic"
	"testing"

	"github.com/your-org/driftwatch/internal/drift"
)

func makeFanoutDrifts(n int) []drift.Drift {
	out := make([]drift.Drift, n)
	for i := range out {
		out[i] = drift.Drift{Key: "k", BaselineValue: "a", CurrentValue: "b"}
	}
	return out
}

func TestNewFanout_NoSenders(t *testing.T) {
	_, err := NewFanout()
	if err == nil {
		t.Fatal("expected error for zero senders")
	}
}

func TestNewFanout_NilSender(t *testing.T) {
	_, err := NewFanout(nil)
	if err == nil {
		t.Fatal("expected error for nil sender")
	}
}

func TestFanout_Add_NilSender(t *testing.T) {
	f, _ := NewFanout(&mockSender{})
	if err := f.Add(nil); err == nil {
		t.Fatal("expected error when adding nil sender")
	}
}

func TestFanout_Len(t *testing.T) {
	f, _ := NewFanout(&mockSender{}, &mockSender{})
	if f.Len() != 2 {
		t.Fatalf("expected 2, got %d", f.Len())
	}
	_ = f.Add(&mockSender{})
	if f.Len() != 3 {
		t.Fatalf("expected 3, got %d", f.Len())
	}
}

func TestFanout_NoDrifts_Noop(t *testing.T) {
	m := &mockSender{}
	f, _ := NewFanout(m)
	if err := f.Send("prod", nil); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if m.calls != 0 {
		t.Fatalf("expected 0 calls, got %d", m.calls)
	}
}

func TestFanout_AllCalled(t *testing.T) {
	var count atomic.Int32
	type counter struct{ mockSender }
	make3 := func() Sender {
		return &struct {
			mockSender
			cnt *atomic.Int32
		}{cnt: &count}
	}
	// Use plain mockSenders and verify via call count on each.
	m1, m2, m3 := &mockSender{}, &mockSender{}, &mockSender{}
	f, _ := NewFanout(m1, m2, m3)
	_ = make3 // suppress unused

	if err := f.Send("staging", makeFanoutDrifts(2)); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	for i, m := range []*mockSender{m1, m2, m3} {
		if m.calls != 1 {
			t.Errorf("sender %d: expected 1 call, got %d", i, m.calls)
		}
	}
}

func TestFanout_CollectsErrors(t *testing.T) {
	errA := errors.New("sender A failed")
	errB := errors.New("sender B failed")
	good := &mockSender{}
	badA := &mockSender{err: errA}
	badB := &mockSender{err: errB}

	f, _ := NewFanout(good, badA, badB)
	err := f.Send("prod", makeFanoutDrifts(1))
	if err == nil {
		t.Fatal("expected combined error")
	}
	if !errors.Is(err, errA) {
		t.Errorf("expected errA in combined error, got: %v", err)
	}
	if !errors.Is(err, errB) {
		t.Errorf("expected errB in combined error, got: %v", err)
	}
	if good.calls != 1 {
		t.Errorf("good sender should still have been called")
	}
}
