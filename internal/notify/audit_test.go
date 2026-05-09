package notify

import (
	"bufio"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/yourusername/driftwatch/internal/drift"
)

func makeAuditDrifts(keys ...string) []drift.Drift {
	out := make([]drift.Drift, 0, len(keys))
	for _, k := range keys {
		out = append(out, drift.Drift{Key: k, BaselineValue: "a", CurrentValue: "b", Severity: "warning"})
	}
	return out
}

type stubSender struct {
	called bool
	err    error
}

func (s *stubSender) Send(_ string, _ []drift.Drift) error {
	s.called = true
	return s.err
}

func TestNewAuditSender_NilInner(t *testing.T) {
	_, err := NewAuditSender(nil, t.TempDir())
	if err == nil {
		t.Fatal("expected error for nil inner")
	}
}

func TestNewAuditSender_EmptyDir(t *testing.T) {
	_, err := NewAuditSender(&stubSender{}, "")
	if err == nil {
		t.Fatal("expected error for empty dir")
	}
}

func TestAuditSender_Send_NoDrifts_Noop(t *testing.T) {
	stub := &stubSender{}
	a, _ := NewAuditSender(stub, t.TempDir())
	if err := a.Send("prod", nil); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if stub.called {
		t.Fatal("inner should not be called for empty drifts")
	}
}

func TestAuditSender_Send_WritesEntry(t *testing.T) {
	dir := t.TempDir()
	stub := &stubSender{}
	a, _ := NewAuditSender(stub, dir)

	drifts := makeAuditDrifts("DB_HOST", "API_KEY")
	if err := a.Send("staging", drifts); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	file := filepath.Join(dir, "audit-staging.jsonl")
	f, err := os.Open(file)
	if err != nil {
		t.Fatalf("audit file not created: %v", err)
	}
	defer f.Close()

	var entry AuditEntry
	if err := json.NewDecoder(bufio.NewReader(f)).Decode(&entry); err != nil {
		t.Fatalf("failed to decode entry: %v", err)
	}
	if entry.Env != "staging" {
		t.Errorf("expected env=staging, got %s", entry.Env)
	}
	if len(entry.DriftKeys) != 2 {
		t.Errorf("expected 2 drift keys, got %d", len(entry.DriftKeys))
	}
	if !entry.Success {
		t.Error("expected success=true")
	}
}

func TestAuditSender_Send_RecordsFailure(t *testing.T) {
	dir := t.TempDir()
	stub := &stubSender{err: errors.New("send failed")}
	a, _ := NewAuditSender(stub, dir)

	drifts := makeAuditDrifts("TIMEOUT")
	err := a.Send("prod", drifts)
	if err == nil {
		t.Fatal("expected error from inner sender")
	}

	file := filepath.Join(dir, "audit-prod.jsonl")
	f, _ := os.Open(file)
	defer f.Close()

	var entry AuditEntry
	json.NewDecoder(f).Decode(&entry) //nolint:errcheck
	if entry.Success {
		t.Error("expected success=false on inner error")
	}
	if entry.Error == "" {
		t.Error("expected error message to be recorded")
	}
}
