package history

import (
	"bytes"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/driftwatch/driftwatch/internal/drift"
)

func makeTimedDriftsForSummary(keys []string, t time.Time) []drift.Drift {
	var out []drift.Drift
	for _, k := range keys {
		out = append(out, drift.Drift{Key: k, BaselineValue: "a", TargetValue: "b"})
	}
	return out
}

func TestBuildEnvSummary_NoDrifts(t *testing.T) {
	dir, _ := os.MkdirTemp("", "summary-test")
	defer os.RemoveAll(dir)

	store, _ := NewStore(dir)
	s, err := BuildEnvSummary(store, "staging", time.Now().Add(-time.Hour))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if s.TotalDrifts != 0 {
		t.Errorf("expected 0 drifts, got %d", s.TotalDrifts)
	}
}

func TestBuildEnvSummary_WithDrifts(t *testing.T) {
	dir, _ := os.MkdirTemp("", "summary-test")
	defer os.RemoveAll(dir)

	store, _ := NewStore(dir)
	now := time.Now()

	_ = store.Save("staging", makeTimedDriftsForSummary([]string{"DB_HOST", "PORT"}, now.Add(-30*time.Minute)))
	_ = store.Save("staging", makeTimedDriftsForSummary([]string{"DB_HOST"}, now.Add(-10*time.Minute)))

	s, err := BuildEnvSummary(store, "staging", now.Add(-time.Hour))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if s.TotalDrifts != 2 {
		t.Errorf("expected 2 drift events, got %d", s.TotalDrifts)
	}
	if s.Environment != "staging" {
		t.Errorf("expected env staging, got %s", s.Environment)
	}
	if len(s.TopKeys) == 0 {
		t.Error("expected at least one top key")
	}
	if s.TopKeys[0] != "DB_HOST" {
		t.Errorf("expected DB_HOST as top key, got %s", s.TopKeys[0])
	}
}

func TestBuildEnvSummary_SinceFilter(t *testing.T) {
	dir, _ := os.MkdirTemp("", "summary-test")
	defer os.RemoveAll(dir)

	store, _ := NewStore(dir)
	now := time.Now()

	_ = store.Save("prod", makeTimedDriftsForSummary([]string{"OLD_KEY"}, now.Add(-2*time.Hour)))
	_ = store.Save("prod", makeTimedDriftsForSummary([]string{"NEW_KEY"}, now.Add(-5*time.Minute)))

	s, err := BuildEnvSummary(store, "prod", now.Add(-time.Hour))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if s.TotalDrifts != 1 {
		t.Errorf("expected 1 recent drift event, got %d", s.TotalDrifts)
	}
}

func TestWriteSummary_NoDrift(t *testing.T) {
	var buf bytes.Buffer
	WriteSummary(&buf, EnvSummary{Environment: "dev"})
	if !strings.Contains(buf.String(), "No drift") {
		t.Errorf("expected no-drift message, got: %s", buf.String())
	}
}

func TestWriteSummary_WithDrift(t *testing.T) {
	var buf bytes.Buffer
	now := time.Now()
	WriteSummary(&buf, EnvSummary{
		Environment: "staging",
		TotalDrifts: 3,
		FirstSeen:   now.Add(-time.Hour),
		LastSeen:    now,
		TopKeys:     []string{"DB_HOST", "PORT"},
	})
	out := buf.String()
	if !strings.Contains(out, "staging") {
		t.Error("expected environment name in output")
	}
	if !strings.Contains(out, "3") {
		t.Error("expected total drift count in output")
	}
	if !strings.Contains(out, "DB_HOST") {
		t.Error("expected top key in output")
	}
}
