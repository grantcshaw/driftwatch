package drift_test

import (
	"testing"

	"github.com/yourusername/driftwatch/internal/drift"
	"github.com/yourusername/driftwatch/internal/environment"
)

func makeSnapshot(t *testing.T, name string, data map[string]string) *environment.Snapshot {
	t.Helper()
	s, err := environment.NewSnapshot(name, data)
	if err != nil {
		t.Fatalf("failed to create snapshot: %v", err)
	}
	return s
}

func TestDetector_NoDrift(t *testing.T) {
	d := drift.NewDetector()
	base := makeSnapshot(t, "staging", map[string]string{"DB_HOST": "localhost", "PORT": "5432"})
	target := makeSnapshot(t, "production", map[string]string{"DB_HOST": "localhost", "PORT": "5432"})

	result := d.Compare(base, target)

	if result.Drifted {
		t.Errorf("expected no drift, got: %s", result)
	}
}

func TestDetector_DriftOnChangedValue(t *testing.T) {
	d := drift.NewDetector()
	base := makeSnapshot(t, "staging", map[string]string{"DB_HOST": "localhost"})
	target := makeSnapshot(t, "production", map[string]string{"DB_HOST": "prod-db.internal"})

	result := d.Compare(base, target)

	if !result.Drifted {
		t.Fatal("expected drift but got none")
	}
	if len(result.DiffKeys) != 1 || result.DiffKeys[0] != "DB_HOST" {
		t.Errorf("unexpected DiffKeys: %v", result.DiffKeys)
	}
}

func TestDetector_MissingKeyInTarget(t *testing.T) {
	d := drift.NewDetector()
	base := makeSnapshot(t, "staging", map[string]string{"DB_HOST": "localhost", "FEATURE_X": "true"})
	target := makeSnapshot(t, "production", map[string]string{"DB_HOST": "localhost"})

	result := d.Compare(base, target)

	if !result.Drifted {
		t.Fatal("expected drift due to missing key")
	}
	if len(result.MissingKeys) != 1 || result.MissingKeys[0] != "FEATURE_X" {
		t.Errorf("unexpected MissingKeys: %v", result.MissingKeys)
	}
}

func TestDetector_ExtraKeyInTarget(t *testing.T) {
	d := drift.NewDetector()
	base := makeSnapshot(t, "staging", map[string]string{"DB_HOST": "localhost"})
	target := makeSnapshot(t, "production", map[string]string{"DB_HOST": "localhost", "NEW_KEY": "value"})

	result := d.Compare(base, target)

	if !result.Drifted {
		t.Fatal("expected drift due to extra key")
	}
	if len(result.ExtraKeys) != 1 || result.ExtraKeys[0] != "NEW_KEY" {
		t.Errorf("unexpected ExtraKeys: %v", result.ExtraKeys)
	}
}

func TestDetector_MultipleDriftTypes(t *testing.T) {
	d := drift.NewDetector()
	base := makeSnapshot(t, "staging", map[string]string{
		"DB_HOST":   "localhost",
		"FEATURE_X": "true",
	})
	target := makeSnapshot(t, "production", map[string]string{
		"DB_HOST": "prod-db.internal",
		"NEW_KEY": "value",
	})

	result := d.Compare(base, target)

	if !result.Drifted {
		t.Fatal("expected drift but got none")
	}
	if len(result.DiffKeys) != 1 || result.DiffKeys[0] != "DB_HOST" {
		t.Errorf("unexpected DiffKeys: %v", result.DiffKeys)
	}
	if len(result.MissingKeys) != 1 || result.MissingKeys[0] != "FEATURE_X" {
		t.Errorf("unexpected MissingKeys: %v", result.MissingKeys)
	}
	if len(result.ExtraKeys) != 1 || result.ExtraKeys[0] != "NEW_KEY" {
		t.Errorf("unexpected ExtraKeys: %v", result.ExtraKeys)
	}
}

func TestResult_String_NoDrift(t *testing.T) {
	r := drift.Result{BaseEnv: "staging", TargetEnv: "production", Drifted: false}
	got := r.String()
	if got == "" {
		t.Error("expected non-empty string")
	}
}
