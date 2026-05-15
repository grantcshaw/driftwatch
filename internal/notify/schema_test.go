package notify

import (
	"errors"
	"testing"

	"github.com/driftwatch/driftwatch/internal/drift"
)

func makeSchemaDrifts(keys ...string) []drift.Drift {
	result := make([]drift.Drift, 0, len(keys))
	for _, k := range keys {
		result = append(result, drift.Drift{
			Key:      k,
			Baseline: "old",
			Current:  "new",
		})
	}
	return result
}

func TestNewSchema_NilInner(t *testing.T) {
	_, err := NewSchema(nil, []string{"key"}, nil)
	if err == nil {
		t.Fatal("expected error for nil inner")
	}
}

func TestNewSchema_NoKeys(t *testing.T) {
	_, err := NewSchema(&mockSender{}, nil, nil)
	if err == nil {
		t.Fatal("expected error when no keys specified")
	}
}

func TestSchema_NoDrifts_Noop(t *testing.T) {
	inner := &mockSender{}
	s, err := NewSchema(inner, []string{"db_host"}, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if err := s.Send("prod", nil); err != nil {
		t.Fatalf("expected nil error for empty drifts, got %v", err)
	}
	if inner.calls != 0 {
		t.Errorf("expected 0 inner calls, got %d", inner.calls)
	}
}

func TestSchema_RequiredKeyPresent_Passes(t *testing.T) {
	inner := &mockSender{}
	s, err := NewSchema(inner, []string{"db_host"}, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	drifts := makeSchemaDrifts("db_host", "port")
	if err := s.Send("prod", drifts); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if inner.calls != 1 {
		t.Errorf("expected 1 inner call, got %d", inner.calls)
	}
}

func TestSchema_RequiredKeyMissing_Errors(t *testing.T) {
	inner := &mockSender{}
	s, err := NewSchema(inner, []string{"db_host"}, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	drifts := makeSchemaDrifts("port", "timeout")
	if err := s.Send("prod", drifts); err == nil {
		t.Fatal("expected error for missing required key")
	}
	if inner.calls != 0 {
		t.Errorf("inner should not be called on validation failure")
	}
}

func TestSchema_ForbiddenKeyPresent_Errors(t *testing.T) {
	inner := &mockSender{}
	s, err := NewSchema(inner, nil, []string{"secret"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	drifts := makeSchemaDrifts("secret", "host")
	if err := s.Send("prod", drifts); err == nil {
		t.Fatal("expected error for forbidden key")
	}
}

func TestSchema_ForbiddenKeyAbsent_Passes(t *testing.T) {
	inner := &mockSender{}
	s, err := NewSchema(inner, nil, []string{"secret"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	drifts := makeSchemaDrifts("host", "port")
	if err := s.Send("prod", drifts); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}

func TestSchema_InnerError_Propagates(t *testing.T) {
	inner := &mockSender{err: errors.New("send failed")}
	s, err := NewSchema(inner, nil, []string{"secret"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	drifts := makeSchemaDrifts("host")
	if err := s.Send("prod", drifts); err == nil {
		t.Fatal("expected inner error to propagate")
	}
}
