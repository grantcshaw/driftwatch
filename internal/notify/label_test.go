package notify

import (
	"errors"
	"testing"

	"github.com/yourorg/driftwatch/internal/drift"
)

func makeLabelDrifts(keys ...string) []drift.Drift {
	out := make([]drift.Drift, len(keys))
	for i, k := range keys {
		out[i] = drift.Drift{Key: k, BaselineValue: "a", CurrentValue: "b"}
	}
	return out
}

func TestNewLabelSender_NilInner(t *testing.T) {
	_, err := NewLabelSender(nil, map[string]string{"env": "prod"})
	if err == nil {
		t.Fatal("expected error for nil inner")
	}
}

func TestNewLabelSender_EmptyLabels(t *testing.T) {
	_, err := NewLabelSender(&mockSender{}, map[string]string{})
	if err == nil {
		t.Fatal("expected error for empty labels")
	}
}

func TestLabelSender_NoDrifts_Noop(t *testing.T) {
	m := &mockSender{}
	ls, _ := NewLabelSender(m, map[string]string{"region": "us-east-1"})
	if err := ls.Send("staging", nil); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if m.called {
		t.Fatal("inner should not be called for empty drifts")
	}
}

func TestLabelSender_InjectsLabels(t *testing.T) {
	m := &mockSender{}
	ls, _ := NewLabelSender(m, map[string]string{"region": "us-east-1", "team": "infra"})
	drifts := makeLabelDrifts("db.host")
	if err := ls.Send("prod", drifts); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	got := m.lastDrifts[0].Metadata
	if got["region"] != "us-east-1" {
		t.Errorf("expected region label, got %v", got)
	}
	if got["team"] != "infra" {
		t.Errorf("expected team label, got %v", got)
	}
}

func TestLabelSender_ExistingMetadata_NotOverwritten(t *testing.T) {
	m := &mockSender{}
	ls, _ := NewLabelSender(m, map[string]string{"source": "label"})
	d := drift.Drift{Key: "k", Metadata: map[string]string{"source": "original"}}
	if err := ls.Send("prod", []drift.Drift{d}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	got := m.lastDrifts[0].Metadata["source"]
	if got != "original" {
		t.Errorf("existing metadata should not be overwritten, got %q", got)
	}
}

func TestLabelSender_PropagatesError(t *testing.T) {
	expected := errors.New("send failed")
	m := &mockSender{err: expected}
	ls, _ := NewLabelSender(m, map[string]string{"x": "y"})
	err := ls.Send("prod", makeLabelDrifts("key"))
	if !errors.Is(err, expected) {
		t.Errorf("expected propagated error, got %v", err)
	}
}

func TestLabelSender_OriginalDrifts_Unchanged(t *testing.T) {
	m := &mockSender{}
	ls, _ := NewLabelSender(m, map[string]string{"injected": "yes"})
	orig := []drift.Drift{{Key: "k", Metadata: nil}}
	_ = ls.Send("prod", orig)
	if orig[0].Metadata != nil {
		t.Error("original drift slice should not be mutated")
	}
}
