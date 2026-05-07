package report_test

import (
	"strings"
	"testing"

	"github.com/driftwatch/internal/report"
)

func TestParseFormat_Valid(t *testing.T) {
	cases := []struct {
		input    string
		wantFmt  report.Format
	}{
		{"text", report.FormatText},
		{"TEXT", report.FormatText},
		{"", report.FormatText},
		{"json", report.FormatJSON},
		{"JSON", report.FormatJSON},
	}
	for _, tc := range cases {
		t.Run(tc.input, func(t *testing.T) {
			got, err := report.ParseFormat(tc.input)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tc.wantFmt {
				t.Errorf("ParseFormat(%q) = %q, want %q", tc.input, got, tc.wantFmt)
			}
		})
	}
}

func TestParseFormat_Invalid(t *testing.T) {
	_, err := report.ParseFormat("csv")
	if err == nil {
		t.Error("expected error for unsupported format, got nil")
	}
}

func TestReport_Summary_NoDrift(t *testing.T) {
	r := report.New("production", nil)
	summary := r.Summary()
	if !strings.Contains(summary, "OK") {
		t.Errorf("expected OK in summary, got: %s", summary)
	}
	if !strings.Contains(summary, "production") {
		t.Errorf("expected environment in summary, got: %s", summary)
	}
}

func TestReport_Summary_WithDrift(t *testing.T) {
	r := report.New("staging", makeDrifts("DB_HOST", "PORT"))
	summary := r.Summary()
	if !strings.Contains(summary, "DRIFT") {
		t.Errorf("expected DRIFT in summary, got: %s", summary)
	}
	if !strings.Contains(summary, "DB_HOST") || !strings.Contains(summary, "PORT") {
		t.Errorf("expected drift keys in summary, got: %s", summary)
	}
	if !strings.Contains(summary, "2") {
		t.Errorf("expected count in summary, got: %s", summary)
	}
}

// makeDrifts is a test helper that constructs a slice of DriftItem values
// from the provided key names, using placeholder expected/actual values.
func makeDrifts(keys ...string) []report.DriftItem {
	items := make([]report.DriftItem, 0, len(keys))
	for _, k := range keys {
		items = append(items, report.DriftItem{
			Key:      k,
			Expected: "expected-" + k,
			Actual:   "actual-" + k,
		})
	}
	return items
}
