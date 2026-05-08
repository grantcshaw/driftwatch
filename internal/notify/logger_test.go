package notify

import (
	"bytes"
	"strings"
	"testing"

	"github.com/yourusername/driftwatch/internal/drift"
)

func makeLoggerDrifts() []drift.Drift {
	return []drift.Drift{
		{Key: "DB_HOST", BaselineValue: "prod-db", CurrentValue: "staging-db"},
		{Key: "REPLICAS", BaselineValue: "3", CurrentValue: "1"},
	}
}

func TestNewLoggerSender_NilWriter_UsesStdout(t *testing.T) {
	s, err := NewLoggerSender(nil, "INFO")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if s.out == nil {
		t.Fatal("expected non-nil writer")
	}
}

func TestLoggerSender_Send_NoDrifts_Noop(t *testing.T) {
	var buf bytes.Buffer
	s, _ := NewLoggerSender(&buf, "DRIFT")

	if err := s.Send("staging", nil); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if buf.Len() != 0 {
		t.Fatalf("expected no output, got: %q", buf.String())
	}
}

func TestLoggerSender_Send_WritesDrifts(t *testing.T) {
	var buf bytes.Buffer
	s, _ := NewLoggerSender(&buf, "DRIFT")

	drifts := makeLoggerDrifts()
	if err := s.Send("staging", drifts); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := buf.String()
	lines := strings.Split(strings.TrimSpace(output), "\n")
	if len(lines) != len(drifts) {
		t.Fatalf("expected %d lines, got %d", len(drifts), len(lines))
	}

	for _, line := range lines {
		if !strings.Contains(line, "env=staging") {
			t.Errorf("expected env=staging in line: %q", line)
		}
		if !strings.Contains(line, "[DRIFT]") {
			t.Errorf("expected prefix [DRIFT] in line: %q", line)
		}
	}
}

func TestLoggerSender_Send_ContainsKeyAndValues(t *testing.T) {
	var buf bytes.Buffer
	s, _ := NewLoggerSender(&buf, "ALERT")

	drifts := []drift.Drift{
		{Key: "TIMEOUT", BaselineValue: "30s", CurrentValue: "10s"},
	}
	if err := s.Send("prod", drifts); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	output := buf.String()
	for _, want := range []string{"TIMEOUT", "30s", "10s", "env=prod"} {
		if !strings.Contains(output, want) {
			t.Errorf("expected %q in output: %q", want, output)
		}
	}
}
