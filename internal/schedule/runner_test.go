package schedule_test

import (
	"context"
	"testing"
	"time"

	"github.com/example/driftwatch/internal/alert"
	"github.com/example/driftwatch/internal/drift"
	"github.com/example/driftwatch/internal/environment"
	"github.com/example/driftwatch/internal/report"
	"github.com/example/driftwatch/internal/schedule"
)

func buildRunner(t *testing.T, baseline string, targets []string) *schedule.Runner {
	t.Helper()

	reg := environment.NewRegistry()
	det := drift.NewDetector()
	not := alert.NewNotifier(3, nil)
	rep := report.New(report.FormatText, nil)

	return schedule.NewRunner(reg, det, not, rep, time.Second, baseline, targets)
}

func TestRunner_RunOnce_MissingBaseline(t *testing.T) {
	r := buildRunner(t, "production", []string{"staging"})
	err := r.RunOnce(context.Background())
	if err == nil {
		t.Fatal("expected error when baseline env is missing, got nil")
	}
}

func TestRunner_RunOnce_NoDrift(t *testing.T) {
	reg := environment.NewRegistry()

	// Register two identical file-based collectors using a temp env file.
	data := "KEY=value\nFOO=bar\n"
	file := writeTempEnvFile(t, data)

	col1, _ := environment.NewCollector("prod", "file", file, "")
	col2, _ := environment.NewCollector("staging", "file", file, "")
	reg.Register(col1)
	reg.Register(col2)

	det := drift.NewDetector()
	not := alert.NewNotifier(3, nil)
	rep := report.New(report.FormatText, nil)

	r := schedule.NewRunner(reg, det, not, rep, time.Second, "prod", []string{"staging"})
	err := r.RunOnce(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestRunner_Run_StopsOnContextCancel(t *testing.T) {
	r := buildRunner(t, "prod", nil)
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	// Use a long interval so the ticker never fires; only ctx cancellation matters.
	r2 := schedule.NewRunner(
		environment.NewRegistry(),
		drift.NewDetector(),
		alert.NewNotifier(3, nil),
		report.New(report.FormatText, nil),
		10*time.Second,
		"prod",
		nil,
	)
	_ = r

	err := r2.Run(ctx)
	if err == nil {
		t.Fatal("expected context error, got nil")
	}
	if err != context.DeadlineExceeded {
		t.Fatalf("expected DeadlineExceeded, got %v", err)
	}
}

// writeTempEnvFile creates a temp file with the given contents and returns its path.
func writeTempEnvFile(t *testing.T, content string) string {
	t.Helper()
	f, err := os.CreateTemp(t.TempDir(), "env-*.env")
	if err != nil {
		t.Fatalf("create temp file: %v", err)
	}
	if _, err := f.WriteString(content); err != nil {
		t.Fatalf("write temp file: %v", err)
	}
	f.Close()
	return f.Name()
}
