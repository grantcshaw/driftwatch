package baseline_test

import (
	"errors"
	"os"
	"testing"
	"time"

	"github.com/example/driftwatch/internal/baseline"
	"github.com/example/driftwatch/internal/environment"
)

func makeSnap(name string) *environment.Snapshot {
	return &environment.Snapshot{
		EnvName:   name,
		Collected: time.Now().UTC().Truncate(time.Second),
		Values:    map[string]string{"KEY": "val"},
		Checksum:  "abc123",
	}
}

func TestNewStore_CreatesDir(t *testing.T) {
	dir := t.TempDir() + "/baselines"
	_, err := baseline.NewStore(dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, err := os.Stat(dir); err != nil {
		t.Fatalf("directory not created: %v", err)
	}
}

func TestStore_SaveAndLoad(t *testing.T) {
	dir := t.TempDir()
	store, _ := baseline.NewStore(dir)
	snap := makeSnap("staging")

	if err := store.Save(snap); err != nil {
		t.Fatalf("save: %v", err)
	}

	loaded, err := store.Load("staging")
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	if loaded.EnvName != snap.EnvName {
		t.Errorf("env name: got %q want %q", loaded.EnvName, snap.EnvName)
	}
	if loaded.Values["KEY"] != "val" {
		t.Errorf("value mismatch")
	}
}

func TestStore_Load_NotFound(t *testing.T) {
	dir := t.TempDir()
	store, _ := baseline.NewStore(dir)

	_, err := store.Load("nonexistent")
	if !errors.Is(err, baseline.ErrNotFound) {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}

func TestStore_Delete(t *testing.T) {
	dir := t.TempDir()
	store, _ := baseline.NewStore(dir)
	snap := makeSnap("prod")
	_ = store.Save(snap)

	if err := store.Delete("prod"); err != nil {
		t.Fatalf("delete: %v", err)
	}
	_, err := store.Load("prod")
	if !errors.Is(err, baseline.ErrNotFound) {
		t.Errorf("expected ErrNotFound after delete, got %v", err)
	}
}

func TestStore_Delete_UnknownEnv_NoError(t *testing.T) {
	dir := t.TempDir()
	store, _ := baseline.NewStore(dir)
	if err := store.Delete("ghost"); err != nil {
		t.Errorf("expected nil, got %v", err)
	}
}
