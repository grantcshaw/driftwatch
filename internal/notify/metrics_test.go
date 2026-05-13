package notify

import (
	"bytes"
	"errors"
	"strings"
	"testing"

	"github.com/example/driftwatch/internal/drift"
)

type mockMetricsSender struct {
	called int
	err    error
}

func (m *mockMetricsSender) Send(_ string, _ []drift.Drift) error {
	m.called++
	return m.err
}

func makeMetricsDrifts(n int) []drift.Drift {
	out := make([]drift.Drift, n)
	for i := range out {
		out[i] = drift.Drift{Key: "k", BaselineValue: "a", CurrentValue: "b"}
	}
	return out
}

func TestNewSenderMetrics_NilInner(t *testing.T) {
	_, err := NewSenderMetrics(nil, "test", nil)
	if err == nil {
		t.Fatal("expected error for nil inner")
	}
}

func TestNewSenderMetrics_EmptyName(t *testing.T) {
	_, err := NewSenderMetrics(&mockMetricsSender{}, "", nil)
	if err == nil {
		t.Fatal("expected error for empty name")
	}
}

func TestSenderMetrics_Send_Success(t *testing.T) {
	inner := &mockMetricsSender{}
	var buf bytes.Buffer
	m, err := NewSenderMetrics(inner, "webhook", &buf)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if err := m.Send("prod", makeMetricsDrifts(2)); err != nil {
		t.Fatalf("unexpected send error: %v", err)
	}

	if m.SendTotal != 1 {
		t.Errorf("expected SendTotal=1, got %d", m.SendTotal)
	}
	if m.SendErrors != 0 {
		t.Errorf("expected SendErrors=0, got %d", m.SendErrors)
	}
	if !strings.Contains(buf.String(), "ok") {
		t.Errorf("expected 'ok' in output, got: %s", buf.String())
	}
}

func TestSenderMetrics_Send_Error(t *testing.T) {
	inner := &mockMetricsSender{err: errors.New("send failed")}
	var buf bytes.Buffer
	m, _ := NewSenderMetrics(inner, "slack", &buf)

	err := m.Send("staging", makeMetricsDrifts(1))
	if err == nil {
		t.Fatal("expected error from inner sender")
	}

	if m.SendErrors != 1 {
		t.Errorf("expected SendErrors=1, got %d", m.SendErrors)
	}
	if !strings.Contains(buf.String(), "error=") {
		t.Errorf("expected error in output, got: %s", buf.String())
	}
}

func TestSenderMetrics_Send_MultipleCallsAccumulate(t *testing.T) {
	inner := &mockMetricsSender{}
	m, _ := NewSenderMetrics(inner, "webhook", nil)

	for i := 0; i < 3; i++ {
		if err := m.Send("prod", makeMetricsDrifts(1)); err != nil {
			t.Fatalf("unexpected error on call %d: %v", i, err)
		}
	}

	if m.SendTotal != 3 {
		t.Errorf("expected SendTotal=3 after three calls, got %d", m.SendTotal)
	}
	if m.SendErrors != 0 {
		t.Errorf("expected SendErrors=0, got %d", m.SendErrors)
	}
}

func TestSenderMetrics_Summary(t *testing.T) {
	inner := &mockMetricsSender{}
	m, _ := NewSenderMetrics(inner, "email", nil)
	_ = m.Send("prod", makeMetricsDrifts(3))

	summary := m.Summary()
	if !strings.Contains(summary, "sender=email") {
		t.Errorf("expected sender name in summary: %s", summary)
	}
	if !strings.Contains(summary, "total=1") {
		t.Errorf("expected total=1 in summary: %s", summary)
	}
}

func TestSenderMetrics_NilWriter_UsesStdout(t *testing.T) {
	inner := &mockMetricsSender{}
	m, err := NewSenderMetrics(inner, "test", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if m.out == nil {
		t.Error("expected non-nil writer")
	}
}
