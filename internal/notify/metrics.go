package notify

import (
	"fmt"
	"io"
	"os"
	"sync"
	"time"

	"github.com/example/driftwatch/internal/drift"
)

// SenderMetrics wraps a Sender and tracks send counts, failures, and latency.
type SenderMetrics struct {
	mu       sync.Mutex
	inner    Sender
	name     string
	out      io.Writer

	SendTotal   int
	SendErrors  int
	LastLatency time.Duration
	LastSentAt  time.Time
}

// NewSenderMetrics wraps inner with metrics tracking. name labels the sender
// in log output. If out is nil, os.Stdout is used.
func NewSenderMetrics(inner Sender, name string, out io.Writer) (*SenderMetrics, error) {
	if inner == nil {
		return nil, fmt.Errorf("notify/metrics: inner sender must not be nil")
	}
	if name == "" {
		return nil, fmt.Errorf("notify/metrics: name must not be empty")
	}
	if out == nil {
		out = os.Stdout
	}
	return &SenderMetrics{inner: inner, name: name, out: out}, nil
}

// Send forwards drifts to the inner sender and records timing and error stats.
func (m *SenderMetrics) Send(env string, drifts []drift.Drift) error {
	start := time.Now()
	err := m.inner.Send(env, drifts)
	elapsed := time.Since(start)

	m.mu.Lock()
	defer m.mu.Unlock()

	m.SendTotal++
	m.LastLatency = elapsed
	m.LastSentAt = time.Now()

	if err != nil {
		m.SendErrors++
		fmt.Fprintf(m.out, "[metrics] sender=%s env=%s drifts=%d latency=%s error=%v\n",
			m.name, env, len(drifts), elapsed, err)
		return err
	}

	fmt.Fprintf(m.out, "[metrics] sender=%s env=%s drifts=%d latency=%s ok\n",
		m.name, env, len(drifts), elapsed)
	return nil
}

// Summary returns a human-readable metrics snapshot.
func (m *SenderMetrics) Summary() string {
	m.mu.Lock()
	defer m.mu.Unlock()
	return fmt.Sprintf("sender=%s total=%d errors=%d last_latency=%s",
		m.name, m.SendTotal, m.SendErrors, m.LastLatency)
}
