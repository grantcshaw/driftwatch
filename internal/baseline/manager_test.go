package baseline_test

import (
	"errors"
	"testing"

	"github.com/example/driftwatch/internal/baseline"
	"github.com/example/driftwatch/internal/environment"
)

func newTestManager(t *testing.T) (*baseline.Manager, *baseline.Store) {
	t.Helper()
	store, err := baseline.NewStore(t.TempDir())
	if err != nil {
		t.Fatalf("store: %v", err)
	}
	reg := environment.NewRegistry()
	mgr := baseline.NewManager(store, reg)
	return mgr, store
}

func TestManager_Promote_And_Current(t *testing.T) {
	mgr, _ := newTestManager(t)
	snap := makeSnap("dev")

	if err := mgr.Promote(snap); err != nil {
		t.Fatalf("promote: %v", err)
	}

	got, err := mgr.Current("dev")
	if err != nil {
		t.Fatalf("current: %v", err)
	}
	if got.Checksum != snap.Checksum {
		t.Errorf("checksum mismatch: got %s want %s", got.Checksum, snap.Checksum)
	}
}

func TestManager_Promote_Nil_Errors(t *testing.T) {
	mgr, _ := newTestManager(t)
	if err := mgr.Promote(nil); err == nil {
		t.Error("expected error for nil snapshot")
	}
}

func TestManager_Current_NotFound(t *testing.T) {
	mgr, _ := newTestManager(t)
	_, err := mgr.Current("missing")
	if !errors.Is(err, baseline.ErrNotFound) {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}

func TestManager_Age_NotFound(t *testing.T) {
	mgr, _ := newTestManager(t)
	_, err := mgr.Age("missing")
	if !errors.Is(err, baseline.ErrNotFound) {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}

func TestManager_Age_ReturnsPositiveDuration(t *testing.T) {
	mgr, _ := newTestManager(t)
	snap := makeSnap("qa")
	_ = mgr.Promote(snap)

	dur, err := mgr.Age("qa")
	if err != nil {
		t.Fatalf("age: %v", err)
	}
	if dur < 0 {
		t.Errorf("expected non-negative duration, got %v", dur)
	}
}
