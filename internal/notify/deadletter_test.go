package notify

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/org/driftwatch/internal/drift"
)

type alwaysFailSender struct{ err error }

func (a *alwaysFailSender) Send(_ string, _ []drift.Drift) error { return a.err }

type alwaysOkSender struct{ called bool }

func (a *alwaysOkSender) Send(_ string, _ []drift.Drift) error { a.called = true; return nil }

func makeDeadLetterDrifts() []drift.Drift {
	return []drift.Drift{{Key: "DB_HOST", BaselineValue: "db-prod", TargetValue: "db-staging"}}
}

func TestNewDeadLetter_NilInner(t *testing.T) {
	_, err := NewDeadLetter(nil, t.TempDir())
	if err == nil {
		t.Fatal("expected error for nil inner")
	}
}

func TestNewDeadLetter_EmptyDir(t *testing.T) {
	_, err := NewDeadLetter(&alwaysOkSender{}, "")
	if err == nil {
		t.Fatal("expected error for empty dir")
	}
}

func TestDeadLetter_Send_Success_NoFile(t *testing.T) {
	dir := t.TempDir()
	ok := &alwaysOkSender{}
	dl, _ := NewDeadLetter(ok, dir)
	if err := dl.Send("prod", makeDeadLetterDrifts()); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	entries, _ := os.ReadDir(dir)
	if len(entries) != 0 {
		t.Fatalf("expected no dead-letter files, got %d", len(entries))
	}
}

func TestDeadLetter_Send_Failure_WritesFile(t *testing.T) {
	dir := t.TempDir()
	fail := &alwaysFailSender{err: errors.New("connection refused")}
	dl, _ := NewDeadLetter(fail, dir)
	err := dl.Send("staging", makeDeadLetterDrifts())
	if err == nil {
		t.Fatal("expected error from failing sender")
	}
	entries, _ := os.ReadDir(dir)
	if len(entries) != 1 {
		t.Fatalf("expected 1 dead-letter file, got %d", len(entries))
	}
	data, _ := os.ReadFile(filepath.Join(dir, entries[0].Name()))
	var entry DeadLetterEntry
	if jsonErr := json.Unmarshal(data, &entry); jsonErr != nil {
		t.Fatalf("failed to parse dead-letter file: %v", jsonErr)
	}
	if entry.Env != "staging" {
		t.Errorf("expected env=staging, got %s", entry.Env)
	}
	if entry.Error != "connection refused" {
		t.Errorf("unexpected error string: %s", entry.Error)
	}
	if len(entry.Drifts) != 1 {
		t.Errorf("expected 1 drift in entry, got %d", len(entry.Drifts))
	}
}

func TestDeadLetter_CreatesDir(t *testing.T) {
	base := t.TempDir()
	dir := filepath.Join(base, "nested", "deadletter")
	_, err := NewDeadLetter(&alwaysOkSender{}, dir)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, statErr := os.Stat(dir); os.IsNotExist(statErr) {
		t.Fatal("expected directory to be created")
	}
}
