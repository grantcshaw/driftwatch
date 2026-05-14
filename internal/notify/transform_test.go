package notify

import (
	"errors"
	"strings"
	"testing"

	"github.com/yourorg/driftwatch/internal/drift"
)

func makeTransformDrifts() []drift.Drift {
	return []drift.Drift{
		{Key: "db_password", ExpectedValue: "secret", ActualValue: "hunter2", Severity: "warning"},
		{Key: "app_port", ExpectedValue: "8080", ActualValue: "9090", Severity: "warning"},
	}
}

func TestNewTransform_NilInner(t *testing.T) {
	_, err := NewTransform(nil, UpperCaseKey())
	if err == nil {
		t.Fatal("expected error for nil inner sender")
	}
}

func TestNewTransform_NoFuncs(t *testing.T) {
	_, err := NewTransform(&mockSender{}, )
	if err == nil {
		t.Fatal("expected error when no transform funcs provided")
	}
}

func TestTransform_NoDrifts_Noop(t *testing.T) {
	ms := &mockSender{}
	tx, err := NewTransform(ms, UpperCaseKey())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if err := tx.Send("prod", nil); err != nil {
		t.Fatalf("expected nil error for empty drifts, got %v", err)
	}
	if ms.called {
		t.Error("inner sender should not be called for empty drifts")
	}
}

func TestTransform_UpperCaseKey(t *testing.T) {
	ms := &mockSender{}
	tx, _ := NewTransform(ms, UpperCaseKey())
	drifts := makeTransformDrifts()
	if err := tx.Send("prod", drifts); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	for _, d := range ms.lastDrifts {
		if d.Key != strings.ToUpper(d.Key) {
			t.Errorf("expected upper-case key, got %q", d.Key)
		}
	}
}

func TestTransform_RedactValue(t *testing.T) {
	ms := &mockSender{}
	tx, _ := NewTransform(ms, RedactValue("db_"))
	drifts := makeTransformDrifts()
	if err := tx.Send("prod", drifts); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	for _, d := range ms.lastDrifts {
		if strings.HasPrefix(d.Key, "db_") {
			if d.ActualValue != "[REDACTED]" || d.ExpectedValue != "[REDACTED]" {
				t.Errorf("expected redacted values for key %q", d.Key)
			}
		} else {
			if d.ActualValue == "[REDACTED]" {
				t.Errorf("non-db key %q should not be redacted", d.Key)
			}
		}
	}
}

func TestTransform_MultipleTransforms_AppliedInOrder(t *testing.T) {
	ms := &mockSender{}
	// UpperCaseKey then Redact — after upper-case, prefix becomes "DB_"
	tx, _ := NewTransform(ms, UpperCaseKey(), RedactValue("DB_"))
	drifts := makeTransformDrifts()
	if err := tx.Send("prod", drifts); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	for _, d := range ms.lastDrifts {
		if strings.HasPrefix(d.Key, "DB_") {
			if d.ActualValue != "[REDACTED]" {
				t.Errorf("expected redacted after upper+redact pipeline for key %q", d.Key)
			}
		}
	}
}

func TestTransform_InnerError_Propagated(t *testing.T) {
	ms := &mockSender{err: errors.New("send failed")}
	tx, _ := NewTransform(ms, UpperCaseKey())
	err := tx.Send("prod", makeTransformDrifts())
	if err == nil || !strings.Contains(err.Error(), "send failed") {
		t.Errorf("expected inner error to propagate, got %v", err)
	}
}
