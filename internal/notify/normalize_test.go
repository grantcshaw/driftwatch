package notify

import (
	"errors"
	"strings"
	"testing"

	"github.com/driftwatch/driftwatch/internal/drift"
)

func makeNormalizeDrifts() []drift.Drift {
	return []drift.Drift{
		{Key: "  DB_HOST  ", Baseline: "  localhost  ", Current: "  prod.db  "},
		{Key: "PORT", Baseline: "5432", Current: "5433"},
	}
}

func TestNewNormalize_NilInner(t *testing.T) {
	_, err := NewNormalize(nil, []NormalizeFunc{TrimSpace}, nil)
	if err == nil {
		t.Fatal("expected error for nil inner")
	}
}

func TestNewNormalize_NoFunctions(t *testing.T) {
	_, err := NewNormalize(&mockSender{}, nil, nil)
	if err == nil {
		t.Fatal("expected error when no functions provided")
	}
}

func TestNormalize_NoDrifts_Noop(t *testing.T) {
	inner := &mockSender{}
	n, err := NewNormalize(inner, []NormalizeFunc{TrimSpace}, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if err := n.Send("prod", nil); err != nil {
		t.Fatalf("expected nil for empty drifts, got %v", err)
	}
	if inner.calls != 0 {
		t.Errorf("expected 0 inner calls, got %d", inner.calls)
	}
}

func TestNormalize_TrimSpaceKey(t *testing.T) {
	var got []drift.Drift
	inner := &mockSender{}
	inner.fn = func(env string, d []drift.Drift) error { got = d; return nil }
	n, _ := NewNormalize(inner, []NormalizeFunc{TrimSpace}, nil)
	_ = n.Send("prod", makeNormalizeDrifts())
	if got[0].Key != "DB_HOST" {
		t.Errorf("expected trimmed key 'DB_HOST', got %q", got[0].Key)
	}
}

func TestNormalize_LowerCaseKey(t *testing.T) {
	var got []drift.Drift
	inner := &mockSender{}
	inner.fn = func(env string, d []drift.Drift) error { got = d; return nil }
	n, _ := NewNormalize(inner, []NormalizeFunc{LowerCase}, nil)
	_ = n.Send("prod", makeNormalizeDrifts())
	if got[1].Key != "port" {
		t.Errorf("expected 'port', got %q", got[1].Key)
	}
}

func TestNormalize_ValueFns_Applied(t *testing.T) {
	var got []drift.Drift
	inner := &mockSender{}
	inner.fn = func(env string, d []drift.Drift) error { got = d; return nil }
	n, _ := NewNormalize(inner, nil, []NormalizeFunc{TrimSpace, LowerCase})
	_ = n.Send("prod", makeNormalizeDrifts())
	if got[0].Baseline != "localhost" {
		t.Errorf("expected 'localhost', got %q", got[0].Baseline)
	}
	if got[0].Current != "prod.db" {
		t.Errorf("expected 'prod.db', got %q", got[0].Current)
	}
}

func TestNormalize_ChainedKeyFns(t *testing.T) {
	var got []drift.Drift
	inner := &mockSender{}
	inner.fn = func(env string, d []drift.Drift) error { got = d; return nil }
	addSuffix := func(s string) string { return s + "_norm" }
	n, _ := NewNormalize(inner, []NormalizeFunc{strings.ToLower, addSuffix}, nil)
	_ = n.Send("prod", []drift.Drift{{Key: "HOST", Baseline: "a", Current: "b"}})
	if got[0].Key != "host_norm" {
		t.Errorf("expected 'host_norm', got %q", got[0].Key)
	}
}

func TestNormalize_InnerError_Propagates(t *testing.T) {
	inner := &mockSender{err: errors.New("downstream failure")}
	n, _ := NewNormalize(inner, []NormalizeFunc{TrimSpace}, nil)
	if err := n.Send("prod", makeNormalizeDrifts()); err == nil {
		t.Fatal("expected inner error to propagate")
	}
}
