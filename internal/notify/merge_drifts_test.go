package notify

import (
	"testing"

	"github.com/yourorg/driftwatch/internal/drift"
)

func TestMergeDrifts_EmptyExisting(t *testing.T) {
	incoming := []drift.Drift{
		{Key: "a", BaselineValue: "1", CurrentValue: "2"},
	}
	result := mergeDrifts(nil, incoming)
	if len(result) != 1 {
		t.Fatalf("expected 1, got %d", len(result))
	}
}

func TestMergeDrifts_EmptyIncoming(t *testing.T) {
	existing := []drift.Drift{
		{Key: "a", BaselineValue: "1", CurrentValue: "2"},
	}
	result := mergeDrifts(existing, nil)
	if len(result) != 1 {
		t.Fatalf("expected 1, got %d", len(result))
	}
}

func TestMergeDrifts_NoDuplicates(t *testing.T) {
	existing := []drift.Drift{{Key: "a"}, {Key: "b"}}
	incoming := []drift.Drift{{Key: "c"}, {Key: "d"}}
	result := mergeDrifts(existing, incoming)
	if len(result) != 4 {
		t.Fatalf("expected 4, got %d", len(result))
	}
}

func TestMergeDrifts_DeduplicatesOnKey(t *testing.T) {
	existing := []drift.Drift{{Key: "a"}, {Key: "b"}}
	incoming := []drift.Drift{{Key: "b", CurrentValue: "new"}, {Key: "c"}}
	result := mergeDrifts(existing, incoming)
	if len(result) != 3 {
		t.Fatalf("expected 3, got %d", len(result))
	}
	// Ensure original value for "b" is preserved (first-write wins)
	for _, d := range result {
		if d.Key == "b" && d.CurrentValue == "new" {
			t.Fatal("duplicate key 'b' should not overwrite existing entry")
		}
	}
}

func TestMergeDrifts_AllDuplicates(t *testing.T) {
	existing := []drift.Drift{{Key: "x"}, {Key: "y"}}
	incoming := []drift.Drift{{Key: "x"}, {Key: "y"}}
	result := mergeDrifts(existing, incoming)
	if len(result) != 2 {
		t.Fatalf("expected 2, got %d", len(result))
	}
}

func TestMergeDrifts_DoesNotMutateExisting(t *testing.T) {
	existing := []drift.Drift{{Key: "a"}}
	incoming := []drift.Drift{{Key: "b"}}
	_ = mergeDrifts(existing, incoming)
	if len(existing) != 1 {
		t.Fatal("mergeDrifts must not mutate the existing slice")
	}
}
