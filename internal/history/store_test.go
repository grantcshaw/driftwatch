package history_test

import (
	"os"
	"testing"
	"time"

	"github.com/your-org/driftwatch/internal/drift"
	"github.com/your-org/driftwatch/internal/history"
)

func makeDrifts() []drift.Drift {
	return []drift.Drift{
		{Key: "DB_HOST", BaselineValue: "prod-db", TargetValue: "staging-db"},
		{Key: "LOG_LEVEL", BaselineValue: "error", TargetValue: "debug"},
	}
}

func TestNewStore_CreatesDir(t *testing.T) {
	dir := t.TempDir()
	subDir := dir + "/history"

	_, err := history.NewStore(subDir)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if _, err := os.Stat(subDir); os.IsNotExist(err) {
		t.Fatal("expected store directory to be created")
	}
}

func TestStore_SaveAndLoadAll(t *testing.T) {
	dir := t.TempDir()
	store, err := history.NewStore(dir)
	if err != nil {
		t.Fatalf("NewStore: %v", err)
	}

	drifts := makeDrifts()
	if err := store.Save("staging", drifts); err != nil {
		t.Fatalf("Save: %v", err)
	}

	records, err := store.LoadAll("staging")
	if err != nil {
		t.Fatalf("LoadAll: %v", err)
	}

	if len(records) != 1 {
		t.Fatalf("expected 1 record, got %d", len(records))
	}

	rec := records[0]
	if rec.Environment != "staging" {
		t.Errorf("expected environment 'staging', got %q", rec.Environment)
	}
	if rec.DriftCount != 2 {
		t.Errorf("expected drift_count 2, got %d", rec.DriftCount)
	}
	if len(rec.Drifts) != 2 {
		t.Errorf("expected 2 drifts, got %d", len(rec.Drifts))
	}
	if rec.Timestamp.IsZero() {
		t.Error("expected non-zero timestamp")
	}
}

func TestStore_LoadAll_EmptyForUnknownEnv(t *testing.T) {
	dir := t.TempDir()
	store, err := history.NewStore(dir)
	if err != nil {
		t.Fatalf("NewStore: %v", err)
	}

	records, err := store.LoadAll("nonexistent")
	if err != nil {
		t.Fatalf("LoadAll: %v", err)
	}
	if len(records) != 0 {
		t.Errorf("expected 0 records, got %d", len(records))
	}
}

func TestStore_SaveMultiple_LoadAll(t *testing.T) {
	dir := t.TempDir()
	store, err := history.NewStore(dir)
	if err != nil {
		t.Fatalf("NewStore: %v", err)
	}

	for i := 0; i < 3; i++ {
		time.Sleep(time.Millisecond) // ensure distinct timestamps
		if err := store.Save("prod", makeDrifts()); err != nil {
			t.Fatalf("Save iteration %d: %v", i, err)
		}
	}

	records, err := store.LoadAll("prod")
	if err != nil {
		t.Fatalf("LoadAll: %v", err)
	}
	if len(records) != 3 {
		t.Errorf("expected 3 records, got %d", len(records))
	}
}
