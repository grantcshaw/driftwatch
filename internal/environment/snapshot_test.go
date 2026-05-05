package environment

import (
	"testing"
)

func TestNewSnapshot_Basic(t *testing.T) {
	values := map[string]string{
		"DB_HOST": "localhost",
		"DB_PORT": "5432",
	}

	snap, err := NewSnapshot("staging", values)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if snap.Environment != "staging" {
		t.Errorf("expected environment 'staging', got %q", snap.Environment)
	}
	if snap.Checksum == "" {
		t.Error("expected non-empty checksum")
	}
	if snap.CapturedAt.IsZero() {
		t.Error("expected non-zero CapturedAt")
	}
}

func TestNewSnapshot_EmptyEnvName(t *testing.T) {
	_, err := NewSnapshot("", map[string]string{"key": "val"})
	if err == nil {
		t.Fatal("expected error for empty environment name, got nil")
	}
}

func TestSnapshot_Equal_SameValues(t *testing.T) {
	values := map[string]string{"FOO": "bar"}

	s1, _ := NewSnapshot("prod", values)
	s2, _ := NewSnapshot("prod", values)

	if !s1.Equal(s2) {
		t.Error("expected snapshots with identical values to be equal")
	}
}

func TestSnapshot_Equal_DifferentValues(t *testing.T) {
	s1, _ := NewSnapshot("prod", map[string]string{"FOO": "bar"})
	s2, _ := NewSnapshot("prod", map[string]string{"FOO": "baz"})

	if s1.Equal(s2) {
		t.Error("expected snapshots with different values to be unequal")
	}
}

func TestSnapshot_DiffKeys(t *testing.T) {
	s1, _ := NewSnapshot("staging", map[string]string{
		"A": "1",
		"B": "2",
		"C": "3",
	})
	s2, _ := NewSnapshot("prod", map[string]string{
		"A": "1",
		"B": "changed",
		"D": "4",
	})

	diffs := s1.DiffKeys(s2)
	diffSet := make(map[string]bool, len(diffs))
	for _, k := range diffs {
		diffSet[k] = true
	}

	for _, expected := range []string{"B", "C", "D"} {
		if !diffSet[expected] {
			t.Errorf("expected key %q in diffs, got %v", expected, diffs)
		}
	}
	if diffSet["A"] {
		t.Error("key 'A' should not appear in diffs")
	}
}

func TestSnapshot_Equal_NilHandling(t *testing.T) {
	var s1 *Snapshot
	var s2 *Snapshot
	if !s1.Equal(s2) {
		t.Error("two nil snapshots should be equal")
	}

	s3, _ := NewSnapshot("prod", map[string]string{"k": "v"})
	if s3.Equal(nil) {
		t.Error("non-nil snapshot should not equal nil")
	}
}
