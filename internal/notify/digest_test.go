package notify

import (
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/yourorg/driftwatch/internal/drift"
)

// capturesSender records every Send call for assertions.
type captureSender struct {
	mu    sync.Mutex
	calls []captureCall
	err   error
}

type captureCall struct {
	env    string
	drifts []drift.Drift
}

func (c *captureSender) Send(env string, drifts []drift.Drift) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.calls = append(c.calls, captureCall{env: env, drifts: drifts})
	return c.err
}

func makeDigestDrifts(n int) []drift.Drift {
	var out []drift.Drift
	for i := 0; i < n; i++ {
		out = append(out, drift.Drift{
			Key:      fmt.Sprintf("key%d", i),
			Baseline: "a",
			Current:  "b",
		})
	}
	return out
}

func TestNewDigest_NilInner(t *testing.T) {
	_, err := NewDigest(nil, time.Second)
	if err == nil {
		t.Fatal("expected error for nil inner sender")
	}
}

func TestNewDigest_ZeroWindow(t *testing.T) {
	cs := &captureSender{}
	_, err := NewDigest(cs, 0)
	if err == nil {
		t.Fatal("expected error for zero window")
	}
}

func TestDigest_Send_NoDrifts_Noop(t *testing.T) {
	cs := &captureSender{}
	d, _ := NewDigest(cs, time.Hour)
	defer d.Stop()

	if err := d.Send("prod", nil); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if err := d.Flush(); err != nil {
		t.Fatalf("unexpected flush error: %v", err)
	}
	cs.mu.Lock()
	defer cs.mu.Unlock()
	if len(cs.calls) != 0 {
		t.Fatalf("expected no calls, got %d", len(cs.calls))
	}
}

func TestDigest_Flush_SendsBuffered(t *testing.T) {
	cs := &captureSender{}
	d, _ := NewDigest(cs, time.Hour)
	defer d.Stop()

	_ = d.Send("staging", makeDigestDrifts(2))
	_ = d.Send("staging", makeDigestDrifts(1))

	if err := d.Flush(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	cs.mu.Lock()
	defer cs.mu.Unlock()
	if len(cs.calls) != 1 {
		t.Fatalf("expected 1 call, got %d", len(cs.calls))
	}
	if len(cs.calls[0].drifts) != 3 {
		t.Fatalf("expected 3 drifts, got %d", len(cs.calls[0].drifts))
	}
}

func TestDigest_Flush_ClearsBuffer(t *testing.T) {
	cs := &captureSender{}
	d, _ := NewDigest(cs, time.Hour)
	defer d.Stop()

	_ = d.Send("prod", makeDigestDrifts(1))
	_ = d.Flush()
	_ = d.Flush() // second flush should send nothing

	cs.mu.Lock()
	defer cs.mu.Unlock()
	if len(cs.calls) != 1 {
		t.Fatalf("expected 1 call after double flush, got %d", len(cs.calls))
	}
}

func TestDigest_WindowTriggers_Flush(t *testing.T) {
	cs := &captureSender{}
	d, _ := NewDigest(cs, 50*time.Millisecond)
	defer d.Stop()

	_ = d.Send("dev", makeDigestDrifts(2))
	time.Sleep(150 * time.Millisecond)

	cs.mu.Lock()
	defer cs.mu.Unlock()
	if len(cs.calls) == 0 {
		t.Fatal("expected automatic flush after window elapsed")
	}
}
