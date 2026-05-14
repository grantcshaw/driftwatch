package notify

import (
	"testing"
	"time"

	"github.com/yourorg/driftwatch/internal/drift"
)

func makeEnvelopeDrifts(numChanges int) []drift.Drift {
	changes := make([]drift.Change, numChanges)
	for i := range changes {
		changes[i] = drift.Change{Key: fmt.Sprintf("key%d", i), Old: "a", New: "b"}
	}
	return []drift.Drift{{Environment: "prod", Changes: changes}}
}

func TestNewEnvelope_EmptyEnv_Errors(t *testing.T) {
	_, err := NewEnvelope("", nil)
	if err == nil {
		t.Fatal("expected error for empty environment name")
	}
}

func TestNewEnvelope_NoDrifts_WarningDefault(t *testing.T) {
	e, err := NewEnvelope("staging", []drift.Drift{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if e.Severity != SeverityWarning {
		t.Errorf("expected warning, got %s", e.Severity)
	}
}

func TestNewEnvelope_SingleChange_Warning(t *testing.T) {
	drifts := makeEnvelopeDrifts(1)
	e, err := NewEnvelope("prod", drifts)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if e.Severity != SeverityWarning {
		t.Errorf("expected warning, got %s", e.Severity)
	}
}

func TestNewEnvelope_MultipleChanges_Critical(t *testing.T) {
	drifts := makeEnvelopeDrifts(3)
	e, err := NewEnvelope("prod", drifts)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if e.Severity != SeverityCritical {
		t.Errorf("expected critical, got %s", e.Severity)
	}
}

func TestNewEnvelope_SetsDetectedAt(t *testing.T) {
	before := time.Now().UTC()
	e, _ := NewEnvelope("dev", nil)
	after := time.Now().UTC()
	if e.DetectedAt.Before(before) || e.DetectedAt.After(after) {
		t.Errorf("DetectedAt %v not in expected range [%v, %v]", e.DetectedAt, before, after)
	}
}

func TestEnvelope_WithLabel(t *testing.T) {
	e, _ := NewEnvelope("dev", nil)
	e.WithLabel("team", "platform").WithLabel("region", "us-east-1")
	if e.Labels["team"] != "platform" {
		t.Errorf("expected label team=platform, got %q", e.Labels["team"])
	}
	if e.Labels["region"] != "us-east-1" {
		t.Errorf("expected label region=us-east-1, got %q", e.Labels["region"])
	}
}

func TestEnvelope_IsCritical(t *testing.T) {
	e, _ := NewEnvelope("prod", makeEnvelopeDrifts(2))
	if !e.IsCritical() {
		t.Error("expected IsCritical to return true")
	}
	w, _ := NewEnvelope("prod", makeEnvelopeDrifts(1))
	if w.IsCritical() {
		t.Error("expected IsCritical to return false for warning envelope")
	}
}
