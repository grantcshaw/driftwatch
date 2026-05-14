package notify

import (
	"context"
	"errors"
	"testing"

	"github.com/example/driftwatch/internal/drift"
)

func makeChainDrifts(n int) []drift.Drift {
	out := make([]drift.Drift, n)
	for i := range out {
		out[i] = drift.Drift{Key: "k", Baseline: "a", Current: "b"}
	}
	return out
}

func TestNewChain_NoSenders(t *testing.T) {
	_, err := NewChain(false)
	if err == nil {
		t.Fatal("expected error for empty senders")
	}
}

func TestNewChain_NilSender(t *testing.T) {
	_, err := NewChain(false, nil)
	if err == nil {
		t.Fatal("expected error for nil sender")
	}
}

func TestChain_Add_NilSender(t *testing.T) {
	c, _ := NewChain(false, &mockSender{})
	if err := c.Add(nil); err == nil {
		t.Fatal("expected error adding nil sender")
	}
}

func TestChain_Len(t *testing.T) {
	c, _ := NewChain(false, &mockSender{}, &mockSender{})
	if c.Len() != 2 {
		t.Fatalf("expected 2, got %d", c.Len())
	}
	_ = c.Add(&mockSender{})
	if c.Len() != 3 {
		t.Fatalf("expected 3, got %d", c.Len())
	}
}

func TestChain_AllSendersCalled(t *testing.T) {
	a, b := &mockSender{}, &mockSender{}
	c, _ := NewChain(false, a, b)
	_ = c.Send(context.Background(), "prod", makeChainDrifts(1))
	if a.calls != 1 || b.calls != 1 {
		t.Fatalf("expected both called once, got a=%d b=%d", a.calls, b.calls)
	}
}

func TestChain_HaltOnError_StopsEarly(t *testing.T) {
	fail := &mockSender{err: errors.New("boom")}
	next := &mockSender{}
	c, _ := NewChain(true, fail, next)
	err := c.Send(context.Background(), "prod", makeChainDrifts(1))
	if err == nil {
		t.Fatal("expected error")
	}
	if next.calls != 0 {
		t.Fatalf("expected next not called, got %d", next.calls)
	}
}

func TestChain_NoHalt_CollectsErrors(t *testing.T) {
	a := &mockSender{err: errors.New("err-a")}
	b := &mockSender{err: errors.New("err-b")}
	c, _ := NewChain(false, a, b)
	err := c.Send(context.Background(), "prod", makeChainDrifts(1))
	if err == nil {
		t.Fatal("expected combined error")
	}
	if a.calls != 1 || b.calls != 1 {
		t.Fatalf("expected both called, got a=%d b=%d", a.calls, b.calls)
	}
}

func TestChain_NoDrifts_StillCallsSenders(t *testing.T) {
	s := &mockSender{}
	c, _ := NewChain(false, s)
	_ = c.Send(context.Background(), "prod", nil)
	if s.calls != 1 {
		t.Fatalf("expected 1 call, got %d", s.calls)
	}
}
