package alert_test

import (
	"strings"
	"testing"
	"time"

	"github.com/yourorg/driftwatch/internal/alert"
	"github.com/yourorg/driftwatch/internal/drift"
)

func makeDrifts(keys ...string) []drift.DriftResult {
	results := make([]drift.DriftResult, 0, len(keys))
	for _, k := range keys {
		results = append(results, drift.DriftResult{
			Key:           k,
			BaselineValue: "old",
			TargetValue:   "new",
		})
	}
	return results
}

func TestNotify_NoDrifts_ReturnsNil(t *testing.T) {
	var buf strings.Builder
	n := alert.NewNotifier(&buf, 5)
	a, err := n.Notify("staging", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if a != nil {
		t.Errorf("expected nil alert for empty drifts, got %+v", a)
	}
	if buf.Len() != 0 {
		t.Errorf("expected no output for empty drifts, got %q", buf.String())
	}
}

func TestNotify_Warning_BelowThreshold(t *testing.T) {
	var buf strings.Builder
	n := alert.NewNotifier(&buf, 5)
	drifts := makeDrifts("DB_HOST", "API_KEY")
	a, err := n.Notify("staging", drifts)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if a == nil {
		t.Fatal("expected non-nil alert")
	}
	if a.Severity != alert.SeverityWarning {
		t.Errorf("expected WARNING severity, got %s", a.Severity)
	}
	if !strings.Contains(buf.String(), "WARNING") {
		t.Errorf("expected WARNING in output, got: %s", buf.String())
	}
}

func TestNotify_Critical_AtThreshold(t *testing.T) {
	var buf strings.Builder
	n := alert.NewNotifier(&buf, 3)
	drifts := makeDrifts("K1", "K2", "K3")
	a, err := n.Notify("production", drifts)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if a.Severity != alert.SeverityCritical {
		t.Errorf("expected CRITICAL severity, got %s", a.Severity)
	}
}

func TestNotify_OutputContainsKeyDetails(t *testing.T) {
	var buf strings.Builder
	n := alert.NewNotifier(&buf, 5)
	drifts := makeDrifts("TIMEOUT")
	_, err := n.Notify("dev", drifts)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	out := buf.String()
	if !strings.Contains(out, "TIMEOUT") {
		t.Errorf("expected key name in output, got: %s", out)
	}
	if !strings.Contains(out, "dev") {
		t.Errorf("expected environment name in output, got: %s", out)
	}
}

func TestNotify_AlertTimestampIsRecent(t *testing.T) {
	var buf strings.Builder
	n := alert.NewNotifier(&buf, 5)
	before := time.Now().UTC().Add(-time.Second)
	a, _ := n.Notify("prod", makeDrifts("X"))
	after := time.Now().UTC().Add(time.Second)
	if a.Timestamp.Before(before) || a.Timestamp.After(after) {
		t.Errorf("alert timestamp %v out of expected range [%v, %v]", a.Timestamp, before, after)
	}
}
