package history

import (
	"testing"
	"time"
)

func TestPrune_OlderThan(t *testing.T) {
	dir := t.TempDir()
	s, _ := NewStore(dir)

	old := time.Now().Add(-3 * time.Hour)
	recent := time.Now().Add(-5 * time.Minute)

	_ = s.Save("prod", makeTimedDrifts(t, "warning", old))
	_ = s.Save("prod", makeTimedDrifts(t, "critical", recent))

	removed, err := s.Prune("prod", PruneOptions{OlderThan: time.Hour})
	if err != nil {
		t.Fatalf("prune error: %v", err)
	}
	if removed != 1 {
		t.Errorf("expected 1 removed, got %d", removed)
	}

	records, _ := s.LoadAll("prod")
	if len(records) != 1 {
		t.Errorf("expected 1 record remaining, got %d", len(records))
	}
}

func TestPrune_KeepLast(t *testing.T) {
	dir := t.TempDir()
	s, _ := NewStore(dir)

	for i := 0; i < 6; i++ {
		at := time.Now().Add(time.Duration(i) * time.Minute)
		_ = s.Save("prod", makeTimedDrifts(t, "warning", at))
	}

	removed, err := s.Prune("prod", PruneOptions{KeepLast: 3})
	if err != nil {
		t.Fatalf("prune error: %v", err)
	}
	if removed != 3 {
		t.Errorf("expected 3 removed, got %d", removed)
	}

	records, _ := s.LoadAll("prod")
	if len(records) != 3 {
		t.Errorf("expected 3 records remaining, got %d", len(records))
	}
}

func TestPrune_AllRemoved_DeletesFile(t *testing.T) {
	dir := t.TempDir()
	s, _ := NewStore(dir)

	old := time.Now().Add(-5 * time.Hour)
	_ = s.Save("prod", makeTimedDrifts(t, "warning", old))

	removed, err := s.Prune("prod", PruneOptions{OlderThan: time.Hour})
	if err != nil {
		t.Fatalf("prune error: %v", err)
	}
	if removed != 1 {
		t.Errorf("expected 1 removed, got %d", removed)
	}

	records, _ := s.LoadAll("prod")
	if len(records) != 0 {
		t.Errorf("expected 0 records, got %d", len(records))
	}
}

func TestPrune_UnknownEnv_NoError(t *testing.T) {
	dir := t.TempDir()
	s, _ := NewStore(dir)

	removed, err := s.Prune("ghost", PruneOptions{OlderThan: time.Hour})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if removed != 0 {
		t.Errorf("expected 0 removed, got %d", removed)
	}
}
