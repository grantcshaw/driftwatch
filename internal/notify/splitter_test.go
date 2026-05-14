package notify

import (
	"errors"
	"strings"
	"testing"

	"github.com/example/driftwatch/internal/drift"
)

func makeSplitterDrifts() []drift.Drift {
	return []drift.Drift{
		{Key: "DB_HOST", BaselineValue: "a", TargetValue: "b", Severity: "critical"},
		{Key: "LOG_LEVEL", BaselineValue: "info", TargetValue: "debug", Severity: "warning"},
		{Key: "TIMEOUT", BaselineValue: "30", TargetValue: "60", Severity: "critical"},
	}
}

func bySeverity(d drift.Drift) string { return d.Severity }

func TestNewSplitter_NilFn(t *testing.T) {
	_, err := NewSplitter(nil)
	if err == nil {
		t.Fatal("expected error for nil split function")
	}
}

func TestSplitter_Register_EmptyBucket(t *testing.T) {
	s, _ := NewSplitter(bySeverity)
	if err := s.Register("", &mockSender{}); err == nil {
		t.Fatal("expected error for empty bucket")
	}
}

func TestSplitter_Register_NilSender(t *testing.T) {
	s, _ := NewSplitter(bySeverity)
	if err := s.Register("critical", nil); err == nil {
		t.Fatal("expected error for nil sender")
	}
}

func TestSplitter_NoDrifts_Noop(t *testing.T) {
	s, _ := NewSplitter(bySeverity)
	critSender := &mockSender{}
	_ = s.Register("critical", critSender)

	if err := s.Send("env", nil); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if critSender.calls != 0 {
		t.Fatal("expected no calls for empty drifts")
	}
}

func TestSplitter_RoutesCorrectly(t *testing.T) {
	s, _ := NewSplitter(bySeverity)
	critSender := &mockSender{}
	warnSender := &mockSender{}
	_ = s.Register("critical", critSender)
	_ = s.Register("warning", warnSender)

	if err := s.Send("env", makeSplitterDrifts()); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if critSender.calls != 1 {
		t.Errorf("expected 1 call to critical sender, got %d", critSender.calls)
	}
	if warnSender.calls != 1 {
		t.Errorf("expected 1 call to warning sender, got %d", warnSender.calls)
	}
}

func TestSplitter_UnmatchedUsesDefault(t *testing.T) {
	s, _ := NewSplitter(bySeverity)
	defaultSender := &mockSender{}
	_ = s.SetDefault(defaultSender)

	if err := s.Send("env", makeSplitterDrifts()); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if defaultSender.calls == 0 {
		t.Fatal("expected default sender to be called for unmatched buckets")
	}
}

func TestSplitter_SetDefault_NilSender(t *testing.T) {
	s, _ := NewSplitter(bySeverity)
	if err := s.SetDefault(nil); err == nil {
		t.Fatal("expected error for nil default sender")
	}
}

func TestSplitter_SenderError_Propagates(t *testing.T) {
	s, _ := NewSplitter(bySeverity)
	_ = s.Register("critical", &mockSender{err: errors.New("send failed")})

	err := s.Send("env", []drift.Drift{
		{Key: "K", Severity: "critical"},
	})
	if err == nil || !strings.Contains(err.Error(), "critical") {
		t.Fatalf("expected error referencing bucket, got %v", err)
	}
}
