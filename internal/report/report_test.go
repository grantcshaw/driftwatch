package report_test

import (
	"strings"
	"testing"

	"github.com/driftwatch/internal/drift"
	"github.com/driftwatch/internal/report"
)

func makeDrifts(keys ...string) []drift.Drift {
	var drifts []drift.Drift
	for _, k := range keys {
		drifts = append(drifts, drift.Drift{
			Key:           k,
			BaselineValue: "val-base",
			TargetValue:   "val-target",
		})
	}
	return drifts
}

func TestReport_WriteText_NoDrift(t *testing.T) {
	r := report.New("production", nil)
	var sb strings.Builder
	if err := r.Write(&sb, report.FormatText); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	out := sb.String()
	if !strings.Contains(out, "no drift detected") {
		t.Errorf("expected no-drift message, got:\n%s", out)
	}
	if !strings.Contains(out, "production") {
		t.Errorf("expected environment name in output, got:\n%s", out)
	}
}

func TestReport_WriteText_WithDrift(t *testing.T) {
	r := report.New("staging", makeDrifts("DB_HOST", "API_KEY"))
	var sb strings.Builder
	if err := r.Write(&sb, report.FormatText); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	out := sb.String()
	for _, key := range []string{"DB_HOST", "API_KEY", "val-base", "val-target"} {
		if !strings.Contains(out, key) {
			t.Errorf("expected %q in text output, got:\n%s", key, out)
		}
	}
}

func TestReport_WriteJSON_NoDrift(t *testing.T) {
	r := report.New("production", nil)
	var sb strings.Builder
	if err := r.Write(&sb, report.FormatJSON); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	out := sb.String()
	if !strings.Contains(out, `"drift_count": 0`) {
		t.Errorf("expected drift_count 0 in JSON, got:\n%s", out)
	}
}

func TestReport_WriteJSON_WithDrift(t *testing.T) {
	r := report.New("staging", makeDrifts("TIMEOUT"))
	var sb strings.Builder
	if err := r.Write(&sb, report.FormatJSON); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	out := sb.String()
	for _, needle := range []string{`"TIMEOUT"`, `"drift_count": 1`, `"environment": "staging"`} {
		if !strings.Contains(out, needle) {
			t.Errorf("expected %q in JSON output, got:\n%s", needle, out)
		}
	}
}

func TestReport_Write_UnknownFormat(t *testing.T) {
	r := report.New("prod", nil)
	var sb strings.Builder
	if err := r.Write(&sb, report.Format("xml")); err == nil {
		t.Error("expected error for unknown format, got nil")
	}
}
