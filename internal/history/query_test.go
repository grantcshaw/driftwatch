package history

import (
	"testing"
	"time"

	"github.com/your-org/driftwatch/internal/drift"
)

func makeTimedDrifts(t *testing.T, severity string, at time.Time) []drift.Drift {
	t.Helper()
	return []drift.Drift{
		{Key: "k", BaselineValue: "a", TargetValue: "b", Severity: severity, DetectedAt: at},
	}
}

func TestQuery_Since(t *testing.T) {
	dir := t.TempDir()
	s, _ := NewStore(dir)

	old := time.Now().Add(-2 * time.Hour)
	recent := time.Now().Add(-10 * time.Minute)

	_ = s.Save("prod", makeTimedDrifts(t, "warning", old))
	_ = s.Save("prod", makeTimedDrifts(t, "critical", recent))

	results, err := s.Query(QueryOptions{Environment: "prod", Since: time.Now().Add(-30 * time.Minute)})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	if results[0].Severity != "critical" {
		t.Errorf("expected critical, got %s", results[0].Severity)
	}
}

func TestQuery_Limit(t *testing.T) {
	dir := t.TempDir()
	s, _ := NewStore(dir)

	for i := 0; i < 5; i++ {
		at := time.Now().Add(time.Duration(i) * time.Minute)
		_ = s.Save("prod", makeTimedDrifts(t, "warning", at))
	}

	results, err := s.Query(QueryOptions{Environment: "prod", Limit: 2})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 2 {
		t.Fatalf("expected 2 results, got %d", len(results))
	}
}

func TestQuery_SortedNewestFirst(t *testing.T) {
	dir := t.TempDir()
	s, _ := NewStore(dir)

	t1 := time.Now().Add(-1 * time.Hour)
	t2 := time.Now()
	_ = s.Save("prod", makeTimedDrifts(t, "warning", t1))
	_ = s.Save("prod", makeTimedDrifts(t, "critical", t2))

	results, _ := s.Query(QueryOptions{Environment: "prod"})
	if results[0].DetectedAt.Before(results[1].DetectedAt) {
		t.Error("results not sorted newest first")
	}
}

func TestSummarize_Counts(t *testing.T) {
	dir := t.TempDir()
	s, _ := NewStore(dir)

	now := time.Now()
	_ = s.Save("staging", makeTimedDrifts(t, "critical", now))
	_ = s.Save("staging", makeTimedDrifts(t, "warning", now.Add(time.Minute)))
	_ = s.Save("staging", makeTimedDrifts(t, "critical", now.Add(2*time.Minute)))

	sum, err := s.Summarize("staging")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if sum.Total != 3 {
		t.Errorf("expected total 3, got %d", sum.Total)
	}
	if sum.Critical != 2 {
		t.Errorf("expected 2 critical, got %d", sum.Critical)
	}
	if sum.Warning != 1 {
		t.Errorf("expected 1 warning, got %d", sum.Warning)
	}
}

func TestSummarize_UnknownEnv(t *testing.T) {
	dir := t.TempDir()
	s, _ := NewStore(dir)

	sum, err := s.Summarize("ghost")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if sum.Total != 0 {
		t.Errorf("expected 0 total, got %d", sum.Total)
	}
}
